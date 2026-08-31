package generator

import (
	"log"
	"regexp"
	"strings"

	"github.com/genigo/genigo/internal/repo"
	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

// postgres schema that genigo introspects; multiple schemas beyond
// `public` are out of the v1 postgres scope
const pgSchemaName = "public"

// GetTablesPostgres introspects the public schema of the connected
// postgres database through pg_catalog and fills the same Tables
// structure as the mysql introspector.
func GetTablesPostgres() map[string]Table {
	enums := loadPGEnums()

	rows, err := repo.DB.Query(`
SELECT
	c.relname,
	a.attname,
	format_type(a.atttypid, a.atttypmod),
	NOT a.attnotnull,
	COALESCE(pg_get_expr(d.adbin, d.adrelid), ''),
	a.attidentity,
	t.typtype
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum > 0 AND NOT a.attisdropped
JOIN pg_type t ON t.oid = a.atttypid
LEFT JOIN pg_attrdef d ON d.adrelid = c.oid AND d.adnum = a.attnum
WHERE n.nspname = $1 AND c.relkind IN ('r','p')
ORDER BY c.relname, a.attnum`, pgSchemaName)
	if err != nil {
		panic(err)
	}

	var out = make(map[string]Table)
	p := pluralize.NewClient()

	var ordered []Table
	for rows.Next() {
		var tableName, colName, columnType, defaultExpr, identity, typtype string
		var nullable bool
		err = rows.Scan(&tableName, &colName, &columnType, &nullable, &defaultExpr, &identity, &typtype)
		if err != nil {
			panic(err)
		}

		if len(ordered) == 0 || ordered[len(ordered)-1].Name != tableName {
			ordered = append(ordered, Table{
				Name:         tableName,
				SingularName: strcase.ToCamel(p.Singular(tableName)),
			})
		}
		t := &ordered[len(ordered)-1]

		col := Column{
			Name:       colName,
			ColumnType: columnType,
			Nullable:   boolToUint8(nullable),
			Default:    normalizePGDefault(defaultExpr),
		}

		// enum columns carry their acceptable values like mysql enums
		if typtype == "e" {
			col.DataType = "enum"
			col.Enum = enums[columnType]
		} else {
			col.DataType = normalizePGType(columnType)
		}

		t.Columns = append(t.Columns, col)

		// identity columns (`GENERATED ... AS IDENTITY`) and serial style
		// nextval() defaults behave like mysql AUTO_INCREMENT
		if identity != "" || strings.HasPrefix(defaultExpr, "nextval(") {
			t.Columns[len(t.Columns)-1].AI = true
			t.AICol = &t.Columns[len(t.Columns)-1]
			t.AIColName = col.Name
		}
	}

	for i := range ordered {
		t := &ordered[i]
		for c := range t.Columns {
			t.SetGoDataType(c)
		}
		t.GetPGChilds()
		t.GetPGParents()
		t.GetPGIndexes()

		out[t.Name] = *t
	}

	linkRelations(out)
	return out
}

// loadPGEnums reads `CREATE TYPE ... AS ENUM` values of the public schema
func loadPGEnums() map[string][]string {
	out := map[string][]string{}

	rows, err := repo.DB.Query(`
SELECT t.typname, e.enumlabel
FROM pg_type t
JOIN pg_enum e ON e.enumtypid = t.oid
JOIN pg_namespace n ON n.oid = t.typnamespace
WHERE n.nspname = $1
ORDER BY t.typname, e.enumsortorder`, pgSchemaName)
	if err != nil {
		log.Printf("error on read enums of %v: %+v", pgSchemaName, err)
		return out
	}

	for rows.Next() {
		var typeName, label string
		if err := rows.Scan(&typeName, &label); err != nil {
			panic(err)
		}
		out[typeName] = append(out[typeName], label)
	}
	return out
}

// pgSizeModifier strips typmod sizes: `character varying(255)` -> `character varying`
var pgSizeModifier = regexp.MustCompile(`\([0-9]+(,[0-9]+)?\)`)

func normalizePGType(formatType string) string {
	return pgSizeModifier.ReplaceAllString(formatType, "")
}

// normalizePGDefault maps server generated defaults (now(), gen_random_uuid(),
// ...) to the shared `CURRENT_TIMESTAMP` marker that the codegen template
// understands: the column is omitted on insert and read back via RETURNING.
// Literal defaults (quoted strings, numbers) are kept as is.
func normalizePGDefault(expr string) string {
	if expr == "" {
		return ""
	}
	if !strings.HasPrefix(expr, "'") && strings.Contains(expr, "(") {
		return "CURRENT_TIMESTAMP"
	}
	return expr
}

func boolToUint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

// GetPGIndexes reads primary keys and unique indexes of a table
func (t *Table) GetPGIndexes() {
	rows, err := repo.DB.Query(`
SELECT ci.relname, a.attname, i.indisprimary
FROM pg_index i
JOIN pg_class ci ON ci.oid = i.indexrelid
JOIN pg_class ct ON ct.oid = i.indrelid
JOIN pg_namespace n ON n.oid = ct.relnamespace
JOIN pg_attribute a ON a.attrelid = ct.oid AND a.attnum = ANY(i.indkey)
WHERE n.nspname = $1 AND ct.relname = $2 AND (i.indisprimary OR i.indisunique)
ORDER BY ci.relname, a.attnum`, pgSchemaName, t.Name)
	if err != nil {
		log.Printf("error on read indexes of %v: %+v", t.Name, err)
		return
	}

	for rows.Next() {
		var indexName, columnName string
		var isPrimary bool
		if err := rows.Scan(&indexName, &columnName, &isPrimary); err != nil {
			panic(err)
		}

		if isPrimary {
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
					t.UniqIndexes[indexName] = append(t.UniqIndexes[indexName], col)
				}
			}
		}
	}

	if len(t.Primaries) == 1 {
		t.Primary = t.Primaries[0].Name
		t.PrimaryColType = t.Primaries[0].GoDataType
	}
}

// GetPGChilds reads foreign keys that point from this table to others
func (t *Table) GetPGChilds() {
	rows, err := repo.DB.Query(`
SELECT a.attname, rc.relname, ra.attname
FROM pg_constraint con
JOIN pg_class c ON c.oid = con.conrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
JOIN pg_class rc ON rc.oid = con.confrelid
JOIN pg_namespace rn ON rn.oid = rc.relnamespace
JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = ANY(con.conkey)
JOIN pg_attribute ra ON ra.attrelid = rc.oid AND ra.attnum = ANY(con.confkey)
WHERE con.contype = 'f' AND n.nspname = $1 AND rn.nspname = $1 AND c.relname = $2`, pgSchemaName, t.Name)
	if err != nil {
		log.Printf("error on read childs of %v: %+v", t.Name, err)
		return
	}

	for rows.Next() {
		var rel Relation
		if err := rows.Scan(&rel.FromCol, &rel.RefTable, &rel.ToCol); err != nil {
			panic(err)
		}
		t.LRelations = append(t.LRelations, rel)
	}
}

// GetPGParents reads foreign keys of other tables that point to this table
func (t *Table) GetPGParents() {
	rows, err := repo.DB.Query(`
SELECT a.attname, c.relname, ra.attname
FROM pg_constraint con
JOIN pg_class c ON c.oid = con.conrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
JOIN pg_class rc ON rc.oid = con.confrelid
JOIN pg_namespace rn ON rn.oid = rc.relnamespace
JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = ANY(con.conkey)
JOIN pg_attribute ra ON ra.attrelid = rc.oid AND ra.attnum = ANY(con.confkey)
WHERE con.contype = 'f' AND n.nspname = $1 AND rn.nspname = $1 AND rc.relname = $2 AND c.relname != rc.relname`, pgSchemaName, t.Name)
	if err != nil {
		log.Printf("error on read parents of %v: %+v", t.Name, err)
		return
	}

	for rows.Next() {
		var rel Relation
		if err := rows.Scan(&rel.FromCol, &rel.RefTable, &rel.ToCol); err != nil {
			panic(err)
		}
		t.RRelations = append(t.RRelations, rel)
	}
}
