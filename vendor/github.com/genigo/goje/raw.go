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
	query := Update + " " + Tablename + " SET "
	var args []any
	var items []string
	for key, val := range Cols {
		items = append(items, Tablename+"."+key+" = ?")
		args = append(args, val)
	}

	conditions, cargs, err := SQLConditionBuilder(Queries)
	if err != nil {
		return -1, err
	}
	args = append(args, cargs...)
	start := time.Now()

	query = query + strings.Join(items, ",") + conditions
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
	if Ignore {
		strict = " IGNORE "
	}

	query := Insert + strict + Tablename
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

	start := time.Now()
	res, err := handler.DB.ExecContext(handler.Ctx, query+"("+strings.Join(columnNames, ",")+") VALUES "+values, args...)

	elapsed := time.Since(start)
	if SlowQueryLogTimeout > 0 && elapsed > SlowQueryLogTimeout {
		log.Printf("[SLOW QUERY] took=%s method=RawBulkInsert(Tablename:%s) query=%s\n", elapsed, Tablename, query+"("+strings.Join(columnNames, ",")+") VALUES ...")
	}

	if err != nil {
		return -1, err
	}

	return res.RowsAffected()
}
