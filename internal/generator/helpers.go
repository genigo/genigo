package generator

import (
	"fmt"

	"github.com/genigo/genigo/consts"
	"github.com/genigo/genigo/internal/config"
	"github.com/iancoleman/strcase"
)

func structTag(Column string) string {
	out := ""
	for i := range config.Conf.Tags {
		if i > 0 {
			out += " "
		}
		out += config.Conf.Tags[i] + ":\"" + Column + "\""
	}

	if out != "" {
		out = "`" + out + "`"
	}
	return out
}

func structTagExcept(Column string, Except string) string {
	out := ""
	for i := range config.Conf.Tags {
		if out != "" {
			out += " "
		}
		col := Column
		if config.Conf.Tags[i] == Except {
			col = "-"
		}
		out += config.Conf.Tags[i] + ":\"" + col + "\""
	}

	if out != "" {
		out = "`" + out + "`"
	}
	return out
}

// dictionary returns the active column-type dictionary of the configured driver
func dictionary() map[string]string {
	if config.Conf.DB.Driver == "postgres" {
		return consts.PostgresDicMp
	}
	return consts.MysqlDicMp
}

// isPostgres reports whether generation targets a postgres database
func isPostgres() bool {
	return config.Conf.DB.Driver == "postgres"
}

func goType(ColumnType string, unsigned bool) string {
	dic := dictionary()
	if unsigned {
		if t, ok := dic[ColumnType+" unsigned"]; ok {
			return t
		}
	}
	if t, ok := dic[ColumnType]; ok {
		return t
	}
	fmt.Print(ColumnType, ",")
	return "interface{}"
}

func camelDefault(s string, def interface{}) string {
	camel := strcase.ToCamel(s)
	if camel == "" {
		return fmt.Sprintf("%v", def)
	}
	return camel
}

func SetImports(imports []string) string {
	if len(imports) == 0 {
		return ""
	}
	if len(imports) == 1 {
		return `"` + imports[0] + `"`
	}

	out := ""
	visited := make(map[string]bool)
	for _, imp := range imports {
		if ok := visited[imp]; ok {
			continue
		}
		visited[imp] = true
		out += "	\"" + imp + "\"\n"
	}
	return out
}

func NonPrimaryCols(t Table) []Column {
	out := []Column{}
	for _, col := range t.Columns {
		if COlContains(t.Primaries, col) < 0 {
			out = append(out, col)
		}
	}
	return out
}
func NonAICols(t Table) []Column {
	out := []Column{}
	for _, col := range t.Columns {
		if !col.AI {
			out = append(out, col)
		}
	}
	return out
}

func (t *Table) Prepare() {
	// postgres reads generated keys via RETURNING, no strconv conversion needed
	if t.PrimaryColType == "string" && !isPostgres() {
		t.Imports = append(t.Imports, "strconv")
	}
}

// COlContains check heyStack contains needle and return needle`s position
func COlContains(heyStack []Column, needle Column) int {
	for i, item := range heyStack {
		if item.Name == needle.Name {
			return i
		}
	}
	return -1
}

// StringContains check heyStack contains needle and return needle`s position
func StringContains(heyStack []string, needle string) int {
	for i, item := range heyStack {
		if item == needle {
			return i
		}
	}
	return -1
}

// linkRelations resolves LRelations.Table pointers against the full table
// list and removes duplicate relations (shared by both introspectors)
func linkRelations(out map[string]Table) {
	for tableName, table := range out {
		visited := []string{}
		for relId := 0; relId < len(table.LRelations); relId++ {
			rel := table.LRelations[relId]
			key := rel.FromCol + rel.ToCol + rel.RefTable
			if StringContains(visited, key) > -1 {
				table.LRelations = RemoveItem(table.LRelations, relId)
				out[tableName] = table
				relId--
				continue
			}
			visited = append(visited, key)
			RelTable, ok := out[rel.RefTable]
			if !ok {
				panic(rel.RefTable + " dosen't exists in table list that is needed for relation of " + tableName)
			}
			out[tableName].LRelations[relId].Table = &RelTable
		}
	}
}

func RemoveItem(s []Relation, i int) []Relation {
	if i < 0 || i > len(s)-1 {
		return s
	}

	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}
