package goje

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Postgres array columns map to these slice types in generated models.
// database/sql delivers pg arrays in their text format (`{a,b,"c"}`), so
// every type implements sql.Scanner and driver.Valuer on that format.
//
// NULL arrays scan into a nil slice; NULL elements become zero values
// (a limitation shared with lib/pq's array support).

// StringArray is a postgres text[]/varchar[]/uuid[] column
type StringArray []string

// Int16Array is a postgres smallint[] column
type Int16Array []int16

// Int32Array is a postgres integer[] column
type Int32Array []int32

// Int64Array is a postgres bigint[] column
type Int64Array []int64

// Float32Array is a postgres real[] column
type Float32Array []float32

// Float64Array is a postgres double precision[] column
type Float64Array []float64

// BoolArray is a postgres boolean[] column
type BoolArray []bool

// ---- sql.Scanner / driver.Valuer implementations ----

func (a *StringArray) Scan(src any) error {
	if src == nil {
		*a = nil
		return nil
	}
	elems, err := scanArray(src)
	if err != nil {
		return err
	}
	out := make(StringArray, len(elems))
	copy(out, elems)
	*a = out
	return nil
}

func (a StringArray) Value() (driver.Value, error) {
	return formatArray(len(a), func(i int) string { return arrayQuote(a[i]) })
}

func (a *Int16Array) Scan(src any) error {
	return scanArrayInto(src, func(elems []string) error {
		out := make(Int16Array, len(elems))
		for i, e := range elems {
			n, err := strconv.ParseInt(orZero(e), 10, 16)
			if err != nil {
				return err
			}
			out[i] = int16(n)
		}
		*a = out
		return nil
	})
}

func (a Int16Array) Value() (driver.Value, error) {
	return formatArray(len(a), func(i int) string { return strconv.FormatInt(int64(a[i]), 10) })
}

func (a *Int32Array) Scan(src any) error {
	return scanArrayInto(src, func(elems []string) error {
		out := make(Int32Array, len(elems))
		for i, e := range elems {
			n, err := strconv.ParseInt(orZero(e), 10, 32)
			if err != nil {
				return err
			}
			out[i] = int32(n)
		}
		*a = out
		return nil
	})
}

func (a Int32Array) Value() (driver.Value, error) {
	return formatArray(len(a), func(i int) string { return strconv.FormatInt(int64(a[i]), 10) })
}

func (a *Int64Array) Scan(src any) error {
	return scanArrayInto(src, func(elems []string) error {
		out := make(Int64Array, len(elems))
		for i, e := range elems {
			n, err := strconv.ParseInt(orZero(e), 10, 64)
			if err != nil {
				return err
			}
			out[i] = n
		}
		*a = out
		return nil
	})
}

func (a Int64Array) Value() (driver.Value, error) {
	return formatArray(len(a), func(i int) string { return strconv.FormatInt(a[i], 10) })
}

func (a *Float32Array) Scan(src any) error {
	return scanArrayInto(src, func(elems []string) error {
		out := make(Float32Array, len(elems))
		for i, e := range elems {
			f, err := strconv.ParseFloat(orZero(e), 32)
			if err != nil {
				return err
			}
			out[i] = float32(f)
		}
		*a = out
		return nil
	})
}

func (a Float32Array) Value() (driver.Value, error) {
	return formatArray(len(a), func(i int) string { return strconv.FormatFloat(float64(a[i]), 'g', -1, 32) })
}

func (a *Float64Array) Scan(src any) error {
	return scanArrayInto(src, func(elems []string) error {
		out := make(Float64Array, len(elems))
		for i, e := range elems {
			f, err := strconv.ParseFloat(orZero(e), 64)
			if err != nil {
				return err
			}
			out[i] = f
		}
		*a = out
		return nil
	})
}

func (a Float64Array) Value() (driver.Value, error) {
	return formatArray(len(a), func(i int) string { return strconv.FormatFloat(a[i], 'g', -1, 64) })
}

func (a *BoolArray) Scan(src any) error {
	return scanArrayInto(src, func(elems []string) error {
		out := make(BoolArray, len(elems))
		for i, e := range elems {
			if e == "" {
				continue
			}
			b, err := strconv.ParseBool(pgBool(e))
			if err != nil {
				return err
			}
			out[i] = b
		}
		*a = out
		return nil
	})
}

func (a BoolArray) Value() (driver.Value, error) {
	return formatArray(len(a), func(i int) string { return strconv.FormatBool(a[i]) })
}

// ---- shared helpers ----

// scanArray parses a pg array text literal into raw string elements
func scanArray(src any) ([]string, error) {
	switch s := src.(type) {
	case nil:
		return nil, nil
	case string:
		return parseArrayLiteral(s)
	case []byte:
		return parseArrayLiteral(string(s))
	}
	return nil, fmt.Errorf("goje: cannot scan %T into an array type", src)
}

func scanArrayInto(src any, set func([]string) error) error {
	if src == nil {
		return nil
	}
	elems, err := scanArray(src)
	if err != nil {
		return err
	}
	return set(elems)
}

// parseArrayLiteral splits `{a,"b c",NULL}` into its raw elements,
// honoring double quotes and backslash escapes
func parseArrayLiteral(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("goje: empty array literal")
	}
	if s[0] != '{' || s[len(s)-1] != '}' {
		return nil, fmt.Errorf("goje: invalid array literal %q", s)
	}
	inner := s[1 : len(s)-1]
	if strings.TrimSpace(inner) == "" {
		return []string{}, nil
	}

	var out []string
	var cur strings.Builder
	quoted := false
	inQuote := false
	flush := func() {
		elem := cur.String()
		if !quoted && elem == "NULL" {
			elem = ""
		}
		out = append(out, elem)
		cur.Reset()
		quoted = false
	}

	for i := 0; i < len(inner); i++ {
		c := inner[i]
		switch {
		case inQuote:
			if c == '\\' && i+1 < len(inner) {
				cur.WriteByte(inner[i+1])
				i++
				continue
			}
			if c == '"' {
				inQuote = false
				continue
			}
			cur.WriteByte(c)
		case c == '"':
			inQuote = true
			quoted = true
		case c == ',':
			flush()
		default:
			cur.WriteByte(c)
		}
	}
	flush()

	return out, nil
}

// formatArray builds a pg array text literal for a Value() call
func formatArray(n int, elem func(i int) string) (driver.Value, error) {
	if n == 0 {
		return "{}", nil
	}
	parts := make([]string, n)
	for i := 0; i < n; i++ {
		parts[i] = elem(i)
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}

// arrayQuote quotes an element when the literal would be ambiguous
func arrayQuote(s string) string {
	if s == "" || s == "NULL" || strings.ContainsAny(s, `{}\," `) {
		return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
	}
	return s
}

// orZero maps a NULL element (already blanked by the parser) to "0"
func orZero(elem string) string {
	if elem == "" {
		return "0"
	}
	return elem
}

// pgBool accepts the short postgres boolean literals
func pgBool(elem string) string {
	switch elem {
	case "t":
		return "true"
	case "f":
		return "false"
	}
	return elem
}

// ensure interface satisfaction at compile time
var (
	_ sql.Scanner   = (*StringArray)(nil)
	_ driver.Valuer = StringArray{}
	_ sql.Scanner   = (*Int16Array)(nil)
	_ driver.Valuer = Int16Array{}
	_ sql.Scanner   = (*Int32Array)(nil)
	_ driver.Valuer = Int32Array{}
	_ sql.Scanner   = (*Int64Array)(nil)
	_ driver.Valuer = Int64Array{}
	_ sql.Scanner   = (*Float32Array)(nil)
	_ driver.Valuer = Float32Array{}
	_ sql.Scanner   = (*Float64Array)(nil)
	_ driver.Valuer = Float64Array{}
	_ sql.Scanner   = (*BoolArray)(nil)
	_ driver.Valuer = BoolArray{}
)
