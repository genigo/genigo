package goje

import (
	"log"
	"strings"
	"time"
)

// RawDelete Deletes entries with standard query
// This method dosen't support After,Before Triggers ...
func (handler *Context) RawDelete(Tablename string, Queries []QueryInterface) (int64, error) {
	query, args, err := ArgumentLessQueryBuilder(Delete, Tablename, nil, Queries)
	if err != nil {
		return -1, err
	}

	start := time.Now()
	// run query
	res, err := handler.DB.ExecContext(handler.Ctx, query, args...)
	// log slow queries
	elapsed := time.Since(start)
	if SlowQueryLogTimeout > 0 && elapsed > SlowQueryLogTimeout {
		log.Printf("[SLOW QUERY] took=%s method=RawDelete(Tablename:%s) query=%s\n", elapsed, Tablename, query)
	}

	if err != nil {
		return -1, err
	}

	return res.RowsAffected()
}

// RawUpdate update entries by map
// This method dosen't support After,Before Triggers ...
func (handler *Context) RawUpdate(Tablename string, Cols map[string]any, Queries ...QueryInterface) (int64, error) {
	if len(Cols) == 0 {
		return -1, ErrNoColsSetForUpdate
	}
	query := Update + " " + insertTable(Tablename) + " SET "
	var args []any
	var items []string
	for key, val := range Cols {
		items = append(items, updateColExpr(Tablename, key))
		args = append(args, val)
	}

	// conditions come with `?` binds; the whole statement is renumbered at once
	// so SET binds ($1..) precede WHERE binds in postgres style
	conditions, cargs, err := buildConditions(Queries)
	if err != nil {
		return -1, err
	}
	args = append(args, cargs...)
	start := time.Now()

	query = rebind(query + strings.Join(items, ",") + conditions)
	res, err := handler.DB.ExecContext(handler.Ctx, query, args...)
	// log slow queries
	elapsed := time.Since(start)
	if SlowQueryLogTimeout > 0 && elapsed > SlowQueryLogTimeout {
		log.Printf("[SLOW QUERY] took=%s method=RawUpdate(Tablename:%s) query=%s\n", elapsed, Tablename, query)
	}
	if err != nil {
		return -1, err
	}

	return res.RowsAffected()
}

// RawBulkInsert insert multiple entries by []map[column name]value
// This method dosen't support After,Before Triggers ...
func (handler *Context) RawBulkInsert(Tablename string, Rows []map[string]any) (int64, error) {
	return RawBulkInsert(handler, false, Tablename, Rows)
}

// RawBulkInsertIgnore insert ignore errors multiple entries by []map[column name]value
// mysql: INSERT IGNORE INTO ... | postgres: INSERT INTO ... ON CONFLICT DO NOTHING
// This method dosen't support After,Before Triggers ...
func (handler *Context) RawBulkInsertIgnore(Tablename string, Rows []map[string]any) (int64, error) {
	return RawBulkInsert(handler, true, Tablename, Rows)
}

// RawBulkInsert blank arguments
func RawBulkInsert(handler *Context, Ignore bool, Tablename string, Rows []map[string]any) (int64, error) {
	if len(Rows) == 0 {
		return -1, ErrNoRowsForInsert
	}

	strict := " INTO "
	tail := ""
	if Ignore {
		if IsPostgres() {
			tail = " ON CONFLICT DO NOTHING"
		} else {
			strict = " IGNORE "
		}
	}

	query := Insert + strict + insertTable(Tablename)
	var args []any
	var columnNames []string

	for index, row := range Rows {
		//use first index as column name index
		if index == 0 {
			for colName := range row {
				columnNames = append(columnNames, colName)
			}
			if len(columnNames) == 0 {
				return -1, ErrNoRowsColsForInsert
			}
		}

		//put arguments attiontion to column names that fetched from index 0
		for _, colName := range columnNames {
			if arg, ok := row[colName]; ok {
				args = append(args, arg)
			} else {
				args = append(args, nil)
			}

		}
	}

	eachRowArgs := strings.Repeat(",?", len(columnNames))
	eachRowArgs = ",(" + eachRowArgs[1:] + ")"
	values := strings.Repeat(eachRowArgs, len(Rows))
	values = values[1:]

	query = query + "(" + strings.Join(insertColumns(columnNames), ",") + ") VALUES " + insertValues(len(columnNames), len(Rows)) + tail

	start := time.Now()
	res, err := handler.DB.ExecContext(handler.Ctx, query, args...)

	elapsed := time.Since(start)
	if SlowQueryLogTimeout > 0 && elapsed > SlowQueryLogTimeout {
		log.Printf("[SLOW QUERY] took=%s method=RawBulkInsert(Tablename:%s) query=%s\n", elapsed, Tablename, query)
	}

	if err != nil {
		return -1, err
	}

	return res.RowsAffected()
}

// insertTable quotes a table name on postgres
// (mysql keeps the historical unquoted form)
func insertTable(table string) string {
	if IsPostgres() {
		return qouteColumn(table)
	}
	return table
}

// updateColExpr builds `table.col = ?` for RawUpdate.
// mysql keeps the historical qualified raw form; postgres only allows
// bare quoted column targets in a SET clause.
func updateColExpr(table, col string) string {
	if IsPostgres() {
		return qouteColumn(col) + " = ?"
	}
	return table + "." + col + " = ?"
}

// insertColumns quotes bulk insert column names on postgres
// (mysql keeps the historical unquoted form)
func insertColumns(columnNames []string) []string {
	if !IsPostgres() {
		return columnNames
	}
	out := make([]string, len(columnNames))
	for i, name := range columnNames {
		out[i] = qouteColumn(name)
	}
	return out
}

// insertValues builds the VALUES placeholder list of a bulk insert:
// mysql `(?,?),(?,?)` | postgres `($1,$2),($3,$4)`
func insertValues(colCount, rowCount int) string {
	if !IsPostgres() {
		eachRowArgs := strings.Repeat(",?", colCount)
		eachRowArgs = ",(" + eachRowArgs[1:] + ")"
		values := strings.Repeat(eachRowArgs, rowCount)
		return values[1:]
	}

	chunks := make([]string, rowCount)
	n := 1
	for i := range chunks {
		chunks[i] = "(" + dialect.Placeholders(n, colCount) + ")"
		n += colCount
	}
	return strings.Join(chunks, ",")
}
