package goje

import (
	"fmt"
	"strings"
)

// Supported database drivers
const (
	DriverMysql    = "mysql"
	DriverPostgres = "postgres"
)

// Dialect keeps the SQL generation rules of a database driver:
// identifier quoting and argument placeholders.
type Dialect struct {
	Name string // driver name in DBConfig: mysql | postgres
	// Quote is the identifier quote character: '`' for mysql, '"' for postgres
	Quote rune
	pg    bool
}

var (
	// MysqlDialect quotes identifiers with backticks and binds arguments with `?`
	MysqlDialect = Dialect{Name: DriverMysql, Quote: '`'}

	// PostgresDialect quotes identifiers with double quotes and binds arguments with $1, $2, ...
	PostgresDialect = Dialect{Name: DriverPostgres, Quote: '"', pg: true}

	// dialect is the package level active dialect.
	// It defaults to mysql and is set by NewDBConnection (InitDB) from DBConfig.Driver.
	dialect = MysqlDialect
)

// CurrentDialect returns the active dialect
func CurrentDialect() Dialect {
	return dialect
}

// SetDialect overrides the active dialect.
// NewDBConnection sets it automatically from DBConfig.Driver;
// use this only before building queries without a DBConfig connection.
func SetDialect(d Dialect) {
	dialect = d
}

// IsPostgres reports whether the active dialect is postgres
func IsPostgres() bool {
	return dialect.pg
}

// Bind returns the n-th (1 based) placeholder of the dialect: `?` or `$n`
func (d Dialect) Bind(n int) string {
	if d.pg {
		return fmt.Sprintf("$%d", n)
	}
	return "?"
}

// Placeholders returns count placeholders starting at from (1 based), joined by comma
func (d Dialect) Placeholders(from, count int) string {
	out := make([]string, count)
	for i := 0; i < count; i++ {
		out[i] = d.Bind(from + i)
	}
	return strings.Join(out, ",")
}

// QuoteIdentifier wraps a bare identifier with the dialect quote character
func (d Dialect) QuoteIdentifier(name string) string {
	return string(d.Quote) + name + string(d.Quote)
}

// rebind rewrites `?` placeholders of a whole statement to $1..$n on the
// postgres dialect. On other dialects the statement returns unchanged.
// Attention: a literal `?` inside a string constant would be rewritten too.
func rebind(query string) string {
	if !dialect.pg {
		return query
	}
	var b strings.Builder
	b.Grow(len(query) + 8)
	n := 0
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			n++
			b.WriteString(dialect.Bind(n))
			continue
		}
		b.WriteByte(query[i])
	}
	return b.String()
}
