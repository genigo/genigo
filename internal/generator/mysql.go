package generator

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/genigo/genigo/internal/config"
	"github.com/genigo/genigo/internal/repo"
	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

// GetTables introspects the configured database (mysql or postgres)
func GetTables() map[string]Table {
	if config.Conf.DB.Driver == "postgres" {
		return GetTablesPostgres()
	}
	return GetTablesMysql()
}

func GetTablesMysql() map[string]Table {

	rows, err := repo.DB.Query(`SELECT 
	TABLES.TABLE_NAME,
	JSON_ARRAYAGG(JSON_OBJECT('name',COLUMNS.COLUMN_NAME,'data_type',COLUMNS.DATA_TYPE,'column_type',COLUMNS.COLUMN_TYPE,'nullable',IF(COLUMNS.IS_NULLABLE = 'NO',0,1),'length',IFNULL(COLUMNS.CHARACTER_MAXIMUM_LENGTH,-1),'default',COLUMNS.COLUMN_DEFAULT,'extra',COLUMNS.EXTRA)) AS cols
	FROM
	TABLES
	INNER JOIN COLUMNS ON COLUMNS.TABLE_SCHEMA = TABLES.TABLE_SCHEMA AND COLUMNS.TABLE_NAME = TABLES.TABLE_NAME
	WHERE TABLES.TABLE_SCHEMA=?
	GROUP BY TABLES.TABLE_SCHEMA, TABLES.TABLE_NAME`, config.Conf.DB.Schema)
	if err != nil {
		panic(err)
	}

	var out = make(map[string]Table)
	p := pluralize.NewClient()

	for rows.Next() {
		var ct Table
		var cols string
		rows.Scan(&ct.Name, &cols)
		ct.SingularName = ct.Name

		ct.SingularName = strcase.ToCamel(p.Singular(ct.Name))

		err = json.Unmarshal([]byte(cols), &ct.Columns)
		if err != nil {
			panic(err)
		}

		//Parse ENUMs
		for i := range ct.Columns {
			if ct.Columns[i].DataType == "enum" {
				str := strings.ReplaceAll(ct.Columns[i].ColumnType, "enum(", "")
				str = strings.ReplaceAll(str, ")", "")
				str = strings.ReplaceAll(str, "'", "")
				ct.Columns[i].Enum = strings.Split(str, ",")
			}

			if strings.Contains(ct.Columns[i].Extra, "auto_increment") {
				ct.Columns[i].AI = true
				ct.AICol = &ct.Columns[i]
				ct.AIColName = ct.Columns[i].Name
			}
			ct.SetGoDataType(i)
		}
		//get relations
		ct.Getchilds()
		//get parent relations
		ct.GetParents()
		//get indexes
		ct.GetIndexes()

		out[ct.Name] = ct
	}

	//put relation table addresses
	linkRelations(out)
	return out
}

func (t *Table) SetGoDataType(i int) {
	dt := goType(t.Columns[i].DataType, strings.Contains(t.Columns[i].ColumnType, "unsigned"))
	if dt == "decimal.Decimal" {
		t.Imports = append(t.Imports, "github.com/shopspring/decimal")
	}
	if t.Columns[i].Nullable == 1 && needsNullWrapper(dt) {
		t.Imports = append(t.Imports, "database/sql")
		dt = makeNullable(dt)
	}

	t.Columns[i].GoDataType = dt
}

// needsNullWrapper reports whether a nullable column needs a sql.Null* wrapper.
// []byte, slices, interface{} and the goje array types carry NULL natively
// (nil) and stay bare; on postgres this covers json/jsonb and array columns.
func needsNullWrapper(dt string) bool {
	if dt == "[]byte" || dt == "interface{}" {
		return false
	}
	if isPostgres() && (strings.HasPrefix(dt, "[]") || strings.HasPrefix(dt, "goje.")) {
		return false
	}
	return true
}

func (t *Table) GetIndexes() {
	rows, err := repo.DB.Query(`SELECT DISTINCT INDEX_NAME,COLUMN_NAME FROM INFORMATION_SCHEMA.STATISTICS WHERE  NON_UNIQUE=0 AND TABLE_SCHEMA = ? AND TABLE_NAME=?`,
		config.Conf.DB.Schema,
		t.Name)
	if err != nil {
		log.Printf("error on read indexes of %v: %+v", t.Name, err)
		return
	}

	for rows.Next() {
		var indexName, columnName string
		rows.Scan(&indexName, &columnName)
		if indexName == "PRIMARY" {
			for _, col := range t.Columns {
				if col.Name == columnName {
					t.Primaries = append(t.Primaries, col)
					break
				}
			}
		} else {
			if t.UniqIndexes == nil {
				t.UniqIndexes = make(map[string][]Column)
			}
			for _, col := range t.Columns {
				if col.Name == columnName {
					if _, ok := t.UniqIndexes[indexName]; ok {
						t.UniqIndexes[indexName] = append(t.UniqIndexes[indexName], col)
					} else {
						t.UniqIndexes[indexName] = []Column{}
						t.UniqIndexes[indexName] = append(t.UniqIndexes[indexName], col)
					}

				}
			}
		}
	}

	if len(t.Primaries) == 1 {
		t.Primary = t.Primaries[0].Name
		t.PrimaryColType = t.Primaries[0].GoDataType
	}
}

func (t *Table) Getchilds() {
	rows, err := repo.DB.Query(`SELECT DISTINCT
	column_name AS column_name,
	referenced_table_name AS referenced_table_name,
	referenced_column_name AS referenced_column_name
	from INFORMATION_SCHEMA.KEY_COLUMN_USAGE where table_schema = ? AND REFERENCED_TABLE_NAME IS NOT NULL AND TABLE_NAME = ?`,
		config.Conf.DB.Schema,
		t.Name)
	if err != nil {
		log.Printf("error on read childs of %v: %+v", t.Name, err)
		return
	}

	for rows.Next() {
		var rel Relation
		rows.Scan(&rel.FromCol, &rel.RefTable, &rel.ToCol)
		t.LRelations = append(t.LRelations, rel)
	}
}

func (t *Table) GetParents() {
	rows, err := repo.DB.Query(`SELECT DISTINCT
	column_name AS column_name,
	TABLE_NAME,
	referenced_column_name AS referenced_column_name
	from INFORMATION_SCHEMA.KEY_COLUMN_USAGE where REFERENCED_TABLE_SCHEMA = ? AND referenced_table_name = ? AND TABLE_NAME != ?`,
		config.Conf.DB.Schema,
		t.Name,
		t.Name)
	if err != nil {
		log.Printf("error on read parents of %v: %+v", t.Name, err)
		return
	}

	for rows.Next() {
		var rel Relation
		rows.Scan(&rel.FromCol, &rel.RefTable, &rel.ToCol)
		t.RRelations = append(t.RRelations, rel)
	}
}

func makeNullable(ctype string) string {
	// postgres numerics keep their decimal type when nullable
	if isPostgres() && ctype == "decimal.Decimal" {
		return "decimal.NullDecimal"
	}
	switch ctype {
	case "string":
		return "sql.NullString"
	case "[]byte":
		return "[]byte"
	case "time.Time":
		return "sql.NullTime"
	case "uint", "uint32", "uint8", "uint16":
		return "sql.NullInt32"
	case "int", "int32":
		return "sql.NullInt32"
	case "int8", "int16":
		return "sql.NullInt16"
	case "int64", "uint64":
		return "sql.NullInt64"
	case "float32", "float64":
		return "sql.NullFloat64"
	case "bool":
		return "sql.NullBool"
	}
	return "sql.NullString"
}
