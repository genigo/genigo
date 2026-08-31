package generator

import (
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

// This file assembles the postgres flavored SQL statements that the codegen
// template bakes into generated models. On mysql the template keeps its
// historical inline SQL, byte for byte.

// pgQ quotes an identifier for postgres
func pgQ(name string) string {
	return `"` + name + `"`
}

// pgInsertCols are the columns the generated postgres Insert() sends:
// identity columns and server generated defaults are omitted,
// the database fills them and RETURNING brings them back.
func pgInsertCols(t Table) []Column {
	out := []Column{}
	for _, col := range t.Columns {
		if col.AI || col.Default == "CURRENT_TIMESTAMP" {
			continue
		}
		out = append(out, col)
	}
	return out
}

// pgReturningCols are the columns RETURNING gives back: the identity
// key plus every server generated default column
func pgReturningCols(t Table) []Column {
	out := []Column{}
	for _, col := range t.Columns {
		if col.AI || col.Default == "CURRENT_TIMESTAMP" {
			out = append(out, col)
		}
	}
	return out
}

// pgInsertCall builds the QueryRowContext call of the generated Insert()
func pgInsertCall(t Table) string {
	cols := pgInsertCols(t)
	query := "INSERT INTO " + pgQ(t.Name)

	if len(cols) == 0 {
		// every column is identity or server defaulted
		query += " DEFAULT VALUES"
	} else {
		names := make([]string, len(cols))
		binds := make([]string, len(cols))
		for i, col := range cols {
			names[i] = pgQ(col.Name)
			binds[i] = "$" + strconv.Itoa(i+1)
		}
		query += "(" + strings.Join(names, ",") + ") VALUES(" + strings.Join(binds, ",") + ")"
	}

	if ret := pgReturningCols(t); len(ret) > 0 {
		names := make([]string, len(ret))
		for i, col := range ret {
			names[i] = pgQ(col.Name)
		}
		query += " RETURNING " + strings.Join(names, ",")
	}

	args := make([]string, len(cols))
	for i, col := range cols {
		args[i] = "opt." + strcase.ToCamel(col.Name)
	}

	call := "handler.DB.QueryRowContext(handler.Ctx, `" + query + "`"
	if len(args) > 0 {
		call += ", " + strings.Join(args, ",")
	}
	return call + ")"
}

// pgInsertScan reads the RETURNING values back into the struct
func pgInsertScan(t Table) string {
	ret := pgReturningCols(t)
	if len(ret) == 0 {
		return "err := row.Err()"
	}
	targets := make([]string, len(ret))
	for i, col := range ret {
		targets[i] = "&opt." + strcase.ToCamel(col.Name)
	}
	return "err := row.Scan(" + strings.Join(targets, ",") + ")"
}

// pgUpdateStmt builds the UPDATE statement (query + args) of the generated Update()
func pgUpdateStmt(t Table) string {
	cols := NonPrimaryCols(t)

	names := make([]string, len(cols))
	args := make([]string, len(cols))
	for i, col := range cols {
		names[i] = pgQ(col.Name) + " = $" + strconv.Itoa(i+1)
		args[i] = "opt." + strcase.ToCamel(col.Name)
	}

	where := pgWhereExpr(t, len(cols))
	for _, pk := range t.Primaries {
		args = append(args, "opt."+strcase.ToCamel(pk.Name))
	}

	return "`UPDATE " + pgQ(t.Name) + " SET " + strings.Join(names, ",") + " WHERE " + where + "`, " + strings.Join(args, ",")
}

// pgDeleteStmt builds the DELETE statement (query + args) of the generated Delete()
func pgDeleteStmt(t Table) string {
	args := make([]string, len(t.Primaries))
	for i, pk := range t.Primaries {
		args[i] = "opt." + strcase.ToCamel(pk.Name)
	}
	return "`DELETE FROM " + pgQ(t.Name) + " WHERE " + pgWhereExpr(t, 0) + "`, " + strings.Join(args, ",")
}

// pgSelectByPKStmt builds the SELECT ... WHERE primary-key statement of GetXByPK()
// args reference the camel-cased function parameters
func pgSelectByPKStmt(t Table) string {
	return pgSelectBy(t, t.Primaries, func(col Column) string {
		return strcase.ToCamel(col.Name)
	})
}

// pgSelectByUniqueStmt builds the SELECT ... WHERE unique-index statement of
// GetXBy<Index>(); args reference the `<col>Arg` function parameters
func pgSelectByUniqueStmt(t Table, cols []Column) string {
	return pgSelectBy(t, cols, func(col Column) string {
		return strcase.ToCamel(col.Name) + "Arg"
	})
}

func pgSelectBy(t Table, cols []Column, argName func(Column) string) string {
	names := make([]string, len(t.Columns))
	for i, col := range t.Columns {
		names[i] = pgQ(col.Name)
	}

	args := make([]string, len(cols))
	for i, col := range cols {
		args[i] = argName(col)
	}

	return "`SELECT " + strings.Join(names, ",") + " FROM " + pgQ(t.Name) + " WHERE " + pgWhereCols(cols, 0) + "`, " + strings.Join(args, ",")
}

// pgWhereExpr builds `"pk"=$n AND ...` for the primary keys,
// placeholders numbered from `from`
func pgWhereExpr(t Table, from int) string {
	return pgWhereCols(t.Primaries, from)
}

func pgWhereCols(cols []Column, from int) string {
	parts := make([]string, len(cols))
	for i, col := range cols {
		parts[i] = pgQ(col.Name) + " = $" + strconv.Itoa(from+i+1)
	}
	return strings.Join(parts, " AND ")
}

// pgSkipCol reports whether the generated BulkInsert helpers should skip a
// column on postgres: identity columns must stay out of the column list,
// otherwise `id=0` would be inserted literally instead of being generated
func pgSkipCol(col Column) bool {
	return isPostgres() && col.AI
}
