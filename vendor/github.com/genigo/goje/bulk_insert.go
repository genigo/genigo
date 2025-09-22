package goje

import (
	"reflect"
)

// BulkInsert insert multiple mixed items, INSERT [INTO,IGNORE]
func BulkInsert(ctx *Context, ignore bool, entities []Entity) (int64, []error) {

	// rows: [table_name][column_name]value
	rows := map[string][]map[string]any{}

	for _, entity := range entities {
		if _, ok := rows[entity.GetTableName()]; !ok {
			rows[entity.GetTableName()] = []map[string]any{}
		}

		v := reflect.ValueOf(entity)
		t := reflect.TypeOf(entity)
		if t.Len() == 0 {
			continue
		}

		currentItem := make(map[string]any, t.Len())

		// loop over struct fields and compare `db` tag
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			dbTag := field.Tag.Get("db")
			if dbTag != "-" && dbTag != "" {
				currentItem[dbTag] = v.Field(i).Interface()
			}
		}

		rows[entity.GetTableName()] = append(rows[entity.GetTableName()], currentItem)
	}

	if len(rows) == 0 {
		return 0, nil
	}

	var inserted int64
	var errorList []error
	for table, items := range rows {
		if len(items) == 0 {
			continue
		}

		r, err := RawBulkInsert(ctx, ignore, table, items)
		inserted += r
		errorList = append(errorList, err)
	}

	return inserted, errorList
}
