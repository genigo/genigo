package goje

import (
	"errors"
	"strings"
)

const (
	QueryTypeLimit      = "limit"
	QueryTypeOffset     = "offset"
	QueryTypeOrder      = "order"
	QueryTypeHaving     = "having"
	QueryTypeGroup      = "group"
	QueryTypeWhere      = "where"
	QueryTypeWhereIn    = "where in"
	QueryTypeWhereNotIn = "where not in"
	QueryTypeJoin       = "join"
	QueryTypeOR         = "or"
)

type QueryInterface interface {
	GetType() string
	GetQuery() string
	GetArgs() []any
}

const (
	Select string = "SELECT"
	Insert string = "INSERT"
	Update string = "UPDATE"
	Delete string = "DELETE"

	Left    string = "LEFT"
	Right   string = "RIGHT"
	Inner   string = "INNER"
	Outer   string = "OUTER"
	Natural string = "NATURAL"
)

// make a select ... FROM query
func SelectQueryBuilder(Tablename string, Columns []string, Queries []QueryInterface) (string, []any, error) {
	return ArgumentLessQueryBuilder(Select, Tablename, Columns, Queries)
}

// ArgumentLessQueryBuilder (Select, Delete) query builder
func ArgumentLessQueryBuilder(Action, Tablename string, Columns []string, Queries []QueryInterface) (string, []any, error) {

	if Action != Select && Action != Delete {
		return "", nil, errors.New("this function dosen't support: " + Action)
	}

	query := Action

	Columns = columnsFilter(Columns)
	if Action == Select {
		query += " " + strings.Join(Columns, ",") + " "
	}

	query += " FROM " + qouteColumn(Tablename)

	conditions, args, err := SQLConditionBuilder(Queries)

	return query + conditions, args, err
}

// SQLConditionBuilder [JOIN WHERE LIMIT OFFSET] ...builder
func SQLConditionBuilder(Queries []QueryInterface) (string, []any, error) {
	query := " "
	var args []any
	var where []string
	//Produce Joins
	for _, q := range Queries {
		if q.GetType() == QueryTypeJoin {
			if strings.Count(q.GetQuery(), "?") != len(q.GetArgs()) {
				return "", nil, errors.New(q.GetQuery() + "; args dosen't match with binds `?`")
			}
			query += q.GetQuery()
			args = append(args, q.GetArgs()...)
		}
	}

	//Produce Where condition
	for _, q := range Queries {
		if q.GetType() == QueryTypeWhere ||
			q.GetType() == QueryTypeOR ||
			q.GetType() == QueryTypeWhereIn ||
			q.GetType() == QueryTypeWhereNotIn {
			if strings.Count(q.GetQuery(), "?") != len(q.GetArgs()) {
				return "", nil, errors.New(q.GetQuery() + "; args dosen't match with binds `?`")
			}
			where = append(where, "("+q.GetQuery()+")")
			args = append(args, q.GetArgs()...)
		}
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	//Add group by
	var groupbys []string
	for _, q := range Queries {
		if q.GetType() == QueryTypeGroup {
			if strings.Count(q.GetQuery(), "?") != len(q.GetArgs()) {
				return "", nil, errors.New(q.GetQuery() + "; args dosen't match with binds `?`")
			}
			groupbys = append(groupbys, qouteColumn(q.GetQuery()))
			args = append(args, q.GetArgs()...)
		}

	}
	if len(groupbys) > 0 {
		query += " GROUP BY " + strings.Join(groupbys, ",")
	}

	//Add having
	var havings []string
	for _, q := range Queries {
		if q.GetType() == QueryTypeHaving {
			if strings.Count(q.GetQuery(), "?") != len(q.GetArgs()) {
				return "", nil, errors.New(q.GetQuery() + "; args dosen't match with binds `?`")
			}
			havings = append(havings, qouteColumn(q.GetQuery()))
			args = append(args, q.GetArgs()...)
		}
	}

	if len(havings) > 0 {
		query += " HAVING " + strings.Join(havings, " AND ")
	}

	//Add orders
	for _, q := range Queries {
		if q.GetType() == QueryTypeOrder {

			if strings.Count(q.GetQuery(), "?") != len(q.GetArgs()) {
				return "", nil, errors.New(q.GetQuery() + "; args dosen't match with binds `?`")
			}

			query += " ORDER BY " + q.GetQuery()
			args = append(args, q.GetArgs()...)
		}
	}

	//Add limitations
	for _, q := range Queries {
		if q.GetType() == QueryTypeLimit || q.GetType() == QueryTypeOffset {
			query += " " + q.GetQuery()
			args = append(args, q.GetArgs()...)
		}
	}

	return query, args, nil
}

// filter multiple columns
func columnsFilter(in []string) []string {
	for i := range in {
		in[i] = qouteColumn(in[i])
	}
	return in
}

// filter column: add backtik if needed
func qouteColumn(input string) string {
	if strings.Contains(input, "`") ||
		strings.Contains(input, " ") ||
		strings.Contains(input, "(") ||
		strings.Contains(input, ":") ||
		strings.Contains(input, "+") ||
		strings.Contains(input, "-") ||
		strings.Contains(input, "^") ||
		strings.Contains(input, "=") ||
		strings.Contains(input, "'") ||
		strings.Contains(input, "\"") ||
		strings.Contains(input, "*") ||
		strings.Contains(input, "/") ||
		strings.Contains(input, "%") {
		return input
	}
	// backtick all parts
	if strings.Contains(input, ",") {
		values := strings.Split(input, ",")
		for i := range values {
			values[i] = qouteColumn(values[i])
		}
		return strings.Join(values, ",")
	}

	// backtick all column parts
	if strings.Contains(input, ".") {
		values := strings.Split(input, ".")
		for i := range values {
			values[i] = qouteColumn(values[i])
		}
		return strings.Join(values, ".")
	}

	return "`" + input + "`"
}
