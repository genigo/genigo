package goje

// Contains Query: A helper for `column LIKE '%Phrase%'`
func Contains(columnName string, argument string) QueryWhere {
	return QueryWhere{
		query: qouteColumn(columnName) + " LIKE ?",
		args:  []any{"%" + argument + "%"},
	}
}

// Find Query: A helper for `column LIKE ?`
func Find(columnName string, argument string) QueryWhere {
	return QueryWhere{
		query: qouteColumn(columnName) + " LIKE ?",
		args:  []any{argument},
	}
}

// StartsWith Query: A helper for `column LIKE 'Phrase%'`
func StartsWith(columnName string, argument string) QueryWhere {
	return QueryWhere{
		query: qouteColumn(columnName) + " LIKE ?",
		args:  []any{argument + "%"},
	}
}

// EndsWith Query: A helper for `column LIKE '%Phrase'`
func EndsWith(columnName string, argument string) QueryWhere {
	return QueryWhere{
		query: qouteColumn(columnName) + " LIKE ?",
		args:  []any{"%" + argument},
	}
}

// Eq: A helper for `column =?`
func Eq(columnName string, argument string) QueryWhere {
	return QueryWhere{
		query: qouteColumn(columnName) + " = ?",
		args:  []any{argument},
	}
}

// Not: A helper for `column !=?`
func Not(columnName string, argument any) QueryWhere {
	return QueryWhere{
		query: qouteColumn(columnName) + " != ?",
		args:  []any{argument},
	}
}

// FindInSet: A helper for `FIND_IN_SET(?, column) > 0 `
func FindInSet(columnName string, argument any) QueryWhere {
	return QueryWhere{
		query: " FIND_IN_SET(?, " + qouteColumn(columnName) + ") > 0",
		args:  []any{argument},
	}
}

// Gt: (greater than) A helper for `column > ?`
func Gt(columnName string, argument any) QueryWhere {
	return QueryWhere{
		query: qouteColumn(columnName) + " > ?",
		args:  []any{argument},
	}
}

// Gte: (greater equal than) A helper for `column >= ?`
func Gte(columnName string, argument any) QueryWhere {
	return QueryWhere{
		query: qouteColumn(columnName) + " >= ?",
		args:  []any{argument},
	}
}

// Lt: (lower than) A helper for `column < ?`
func Lt(columnName string, argument any) QueryWhere {
	return QueryWhere{
		query: qouteColumn(columnName) + " < ?",
		args:  []any{argument},
	}
}

// Lte: (lower equal than) A helper for `column <= ?`
func Lte(columnName string, argument any) QueryWhere {
	return QueryWhere{
		query: qouteColumn(columnName) + " <= ?",
		args:  []any{argument},
	}
}
