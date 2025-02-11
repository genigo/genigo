package consts

const StructTemplate = `// Autogenerated by genigo({{version}})
// Please don't edit this file
// If you want add some helpers/triggers you should open a different file such as {{.Name}}.helper.go
package {{pkg}}

import(
	{{if not (NonPrimaryCols .)}}"errors"{{end}}
	{{imports .Imports}}
	"github.com/genigo/goje"
)

type {{.SingularName}} struct{
	{{range .Columns}}{{ if eq $.Primary .Name }}
	//Primary Key {{ if eq .AI true}}Auto Increment{{end}}{{end}}{{if and (eq .DataType "enum") (eq .Nullable 0)}}

	//{{camel .Name}}: Enum acceptable values({{join .Enum ","}}){{end}}
	{{camel .Name}} 	{{if and (eq .DataType "enum") (eq .Nullable 0)}}{{$.SingularName}}{{camel .Name}}{{else}}{{.GoDataType}}{{end}} 	{{structTag .Name}}{{end}}
	{{range .LRelations}}
	//relation one to one, {{camel .Table.Name}} [{{.FromCol}} -> {{.Table.Name}}.{{.ToCol}}]
	{{camel .FromCol}}{{camel (singular .RefTable)}} *{{.Table.SingularName}} {{structTagExcept ( (print .FromCol "_" (singular .Table.Name)) ) "db"}}{{end}}
	{{range .RRelations}}
	//relation one to many {{camel .RefTable}} [{{.FromCol}} -> {{.ToCol}}]
	{{camel .RefTable}}By{{camel .FromCol}}Rel *[]{{camel (singular .RefTable)}} {{structTagExcept (print .RefTable "_by_" .FromCol)  "db"}}{{end}}
	//internal context, db handler
	ctx		*goje.Context {{structTag "-"}}
	parent  goje.Entity {{structTag "-"}}
}

var BeforeInsert{{.SingularName}} func(*goje.Context, *{{.SingularName}}) error
var BeforeUpdate{{.SingularName}} func(*goje.Context, *{{.SingularName}}) error
var BeforeDelete{{.SingularName}} func(*goje.Context, *{{.SingularName}}) error
var AfterInsert{{.SingularName}} func(*goje.Context, *{{.SingularName}})
var AfterUpdate{{.SingularName}} func(*goje.Context, *{{.SingularName}})
var AfterDelete{{.SingularName}} func(*goje.Context, *{{.SingularName}})

const {{.SingularName}}TableName = "{{.Name}}"
{{if .Primaries}}const {{.SingularName}}PrimaryKey = "{{joinCols .Primaries "%s" ","}}"{{end}}
var {{.SingularName}}Columns = []string{
	{{range $ir,$col := .Columns}}"{{$col.Name}}",{{end}}
}

{{range $ir,$col := .Columns}}{{if and (eq $col.DataType "enum") (eq $col.Nullable 0)}}
type {{$.SingularName}}{{camel $col.Name}} {{$col.GoDataType}}

const (
	{{range $ic,$en := $col.Enum}}
	{{$.SingularName}}{{camel $col.Name}}{{camelDefault $en $ic}} {{$.SingularName}}{{camel $col.Name}} = "{{$en}}"{{end}}
)
{{end}}{{end}}


//Tablename returns real table name in the database
//a helper method for goje.Entity interface
func (opt {{.SingularName}}) GetTableName() string {
	return {{.SingularName}}TableName
}

//GetCtx returns current context handler
func (opt {{.SingularName}}) GetCtx() *goje.Context {
	return opt.ctx
}

//SetCtx set context handler for current object
func (opt *{{.SingularName}}) SetCtx(ctx *goje.Context) {
	opt.ctx = ctx
}

// GetParent returns parent object
func (opt {{.SingularName}}) GetParent() *goje.Entity {
	return &opt.parent
}

//Columns returns list of table columns
//a helper method for goje.Entity interface
func (opt {{.SingularName}}) GetColumns() []string {
	return {{.SingularName}}Columns
}

// Delete multiple {{.SingularName}} by queries
// This method dosen't support After,Before Triggers ...
func Delete{{.SingularName}}Raw(handler *goje.Context, Queries ...goje.QueryInterface) (rowsAffected int64, err error) {
	return handler.RawDelete("{{.Name}}", Queries)
}

// Insert multiple {{.SingularName}}
// This method dosen't support After,Before Triggers ...
func BulkInsert{{.SingularName}}(handler *goje.Context, entities []{{.SingularName}}) (rowsInserted int64, err error) {
	rows := make([]map[string]interface{}, len(entities))

	for i := 0; i < len(entities); i++{
		rows[i] = make(map[string]interface{})
		{{range .Columns}}{{if ne .Default "CURRENT_TIMESTAMP"}}rows[i]["{{.Name}}"] = entities[i].{{camel .Name}}
		{{end}}{{end}}}
	return handler.RawBulkInsert("{{.Name}}", rows)
}

{{if .Primary }}
//Delete delete an entry
func (opt *{{.SingularName}}) Delete(handler *goje.Context) error {
	if handler == nil {
		handler = opt.ctx
	}
	if handler == nil {
		return goje.ErrHandlerIsNil
	}

	if BeforeDelete{{.SingularName}} != nil {
		err := BeforeDelete{{.SingularName}}(handler, opt)
		if err != nil {
			return err
		}
	}
	_, err := handler.DB.ExecContext(handler.Ctx, ` + "`" + `DELETE FROM {{.Name}} WHERE {{.Primary }}=?` + "`" + `, opt.{{camel .Primary }})
	if err != nil {
		return err
	}

	if AfterDelete{{.SingularName}} != nil {
		AfterDelete{{.SingularName}}(handler, opt)
	}

	return nil

}
	{{end}}

//Save update an entry when primary key is set or create an entry when primary key isn't set
func (opt *{{.SingularName}}) Save(handler *goje.Context) error {
	if handler == nil {
		handler = opt.ctx
	}
	if handler == nil {
		return goje.ErrHandlerIsNil
	}

	{{if ne .Primary ""}}
	if opt.{{camel .Primary}} == {{if eq .PrimaryColType "string"}}""{{else}}0{{end}} {
		return opt.Insert(handler)
	}

	return opt.Update(handler)
	{{else}}
	return opt.Insert(handler)
	{{end}}

}

//Insert create an entry
func (opt *{{.SingularName}}) Insert(handler *goje.Context) error {
	if handler == nil {
		handler = opt.ctx
	}
	if handler == nil {
		return goje.ErrHandlerIsNil
	}
	
	if BeforeInsert{{.SingularName}} != nil {
		err := BeforeInsert{{.SingularName}}(handler, opt)
		if err != nil {
			return err
		}
	}
	{{range .Columns}}
	{{if eq .Default "CURRENT_TIMESTAMP"}}
	if opt.{{camel .Name}}.IsZero(){
		opt.{{camel .Name}} = time.Now()
	}
	{{end}}
	{{end}}

	{{if .Primary}}
	result, err := handler.DB.ExecContext(handler.Ctx, ` + "`" + `INSERT INTO {{.Name}}({{joinCols (NonAICols .) "%s" ","}}) VALUES({{joinCols (NonAICols .) "?" ","}})` + "`" + `,{{joinCamelCols (NonAICols .) "opt.%s" ","}})
	{{else}}
	_, err := handler.DB.ExecContext(handler.Ctx, ` + "`" + `INSERT INTO {{.Name}}({{joinCols .Columns "%s" ","}}) VALUES({{joinCols .Columns "?" ","}})` + "`" + `,{{joinCamelCols .Columns "opt.%s" ","}})
	{{end}}
	if err != nil {
		return err
	}
	{{if ne .Primary ""}}
	lastId, err := result.LastInsertId() 
	if err == nil{
		opt.{{camel .Primary}} = {{if eq .PrimaryColType "string"}}strconv.Itoa(int(lastId)){{else}}{{.PrimaryColType}}(lastId){{end}}
	}{{end}}

	if AfterInsert{{.SingularName}} != nil {
		AfterInsert{{.SingularName}}(handler, opt)
	}
	return nil
}

//Update an entry
func (opt *{{.SingularName}}) Update(handler *goje.Context) error {
	{{if (NonPrimaryCols .)}}
	if handler == nil {
		handler = opt.ctx
	}

	if handler == nil {
		return goje.ErrHandlerIsNil
	}

	if BeforeUpdate{{.SingularName}} != nil {
		err := BeforeUpdate{{.SingularName}}(handler, opt)
		if err != nil {
			return err
		}
	}

	_, err := handler.DB.ExecContext(handler.Ctx, "UPDATE {{.Name}} SET {{joinCols (NonPrimaryCols .) "%s=?" ","}} WHERE {{joinCols .Primaries "%s=?" " AND "}}",{{joinCamelCols (NonPrimaryCols .) "opt.%s" ","}},{{joinCamelCols .Primaries "opt.%s" ","}})
	if err != nil {
		return err
	}

	if AfterUpdate{{.SingularName}} != nil {
		AfterUpdate{{.SingularName}}(handler, opt)
	}
	return nil
	{{else}}
	return errors.New("{{.SingularName}} doesn't have any non primary keys")
	{{end}}
}

{{if .Primaries}}
//Get{{.SingularName}}By{{joinCamelCols .Primaries "%s" ","}} get one object of {{.SingularName}} by primary keys
func Get{{.SingularName}}By{{joinCamelCols .Primaries "%s" "And"}}(handler *goje.Context, {{joinCamelCols .Primaries "%s" ","}} interface{}) (*{{.SingularName}}, error) {

	if handler == nil {
		return nil, goje.ErrHandlerIsNil
	}

	row := handler.DB.QueryRowContext(handler.Ctx, "SELECT {{range $ir,$col := .Columns}}{{if ne $ir 0}},{{end}}{{$col.Name}}{{end}} FROM {{.Name}} WHERE {{joinCols .Primaries "%s=?" " AND "}}", {{joinCamelCols .Primaries "%s" ","}})

	if row.Err() != nil {
		return nil, row.Err()
	}

	var out {{.SingularName}}
	out.ctx = handler
	err := row.Scan({{range $ir,$col := .Columns}}{{if ne $ir 0}},{{end}}&out.{{camel $col.Name}}{{end}})
	return &out, err
}
{{end}}

{{range $indexName,$cols := .UniqIndexes}}
{{if ne $indexName "PRIMARY"}}
//Get data attention by {{$indexName}} index
func Get{{$.SingularName}}By{{joinCamelCols $cols "%s" "And"}}(handler *goje.Context, {{joinCamelCols $cols "%sArg interface{}" ","}}) (*{{$.SingularName}}, error) {
	if handler == nil {
		return nil, goje.ErrHandlerIsNil
	}

	row := handler.DB.QueryRowContext(handler.Ctx, "SELECT {{range $ir,$col := $.Columns}}{{if ne $ir 0}},{{end}}{{$col.Name}}{{end}} FROM {{$.Name}} WHERE {{range $ic,$col := $cols}}{{if ne $ic 0}} AND {{end}}{{$col.Name}}=?{{end}}"{{range $cols}}, {{camel .Name}}Arg{{end}})

	if row.Err() != nil {
		return nil, row.Err()
	}

	var out {{$.SingularName}}
	out.ctx = handler

	err := row.Scan({{range $ir,$col := $.Columns}}{{if ne $ir 0}},{{end}}&out.{{camel $col.Name}}{{end}})
	return &out, err
}
{{end}}
{{end}}

//Get{{camel .Name}} Get list of entities by query
func Get{{camel .Name}}(handler *goje.Context, Queries ...goje.QueryInterface) ([]{{.SingularName}}, error) {
	if handler == nil {
		return nil, goje.ErrHandlerIsNil
	}
	
	query, args, err := goje.SelectQueryBuilder("{{.Name}}", {{.SingularName}}Columns, Queries)
	if err != nil {
		return nil, err
	}

	rows, err := handler.DB.QueryContext(handler.Ctx,query ,args...)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()
	
	var out []{{.SingularName}}
	for rows.Next(){
		var c {{.SingularName}}
		c.ctx = handler
		err = rows.Scan({{range $ir,$col := .Columns}}{{if ne $ir 0}},{{end}}&c.{{camel $col.Name}}{{end}})
		if err != nil{
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

//Count{{camel .Name}} count list of entities
func Count{{camel .Name}}(handler *goje.Context, Queries ...goje.QueryInterface) (int, error) {
	if handler == nil {
		return -1, goje.ErrHandlerIsNil
	}
	
	query, args, err := goje.SelectQueryBuilder("{{.Name}}", []string{"COUNT(*) AS total"}, Queries)
	if err != nil {
		return -1, err
	}

	row := handler.DB.QueryRowContext(handler.Ctx,query ,args...)

	if row.Err() != nil {
		return -1, row.Err()
	}
	
	var total int
	err = row.Scan(&total)

	return total, err
}

	
{{range .LRelations}}
// Load {{camel .Table.SingularName}} child by {{.FromCol}}
func (opt *{{$.SingularName}}) Load{{camel .FromCol}}{{camel (singular .RefTable)}}() error {
	
	if opt.ctx == nil {
		return goje.ErrHandlerIsNil
	}

	if opt.parent != nil && opt.parent.GetTableName() == "{{.Table.Name}}" {
		return goje.ErrRecursiveLoad
	}

	rows, err := Get{{camel .Table.Name}}(opt.ctx, goje.Where("{{.ToCol}} = ?", opt.{{camel .FromCol}}))

	if err != nil {
		return err
	}

	if len(rows) > 0{
		obj := rows[0]
		obj.parent = *opt
		opt.{{camel .FromCol}}{{camel (singular .RefTable)}} =  &obj
	}
	
	return nil
}
{{end}}

{{range .RRelations}}
//Load {{camel .RefTable}} by {{.FromCol}}
func (opt *{{$.SingularName}}) Load{{camel .RefTable}}By{{camel .FromCol}}(Queries ...goje.QueryInterface) error {
	
	if opt.ctx == nil {
		return goje.ErrHandlerIsNil
	}

	if opt.parent != nil && opt.parent.GetTableName() == "{{.RefTable}}" {
		return goje.ErrRecursiveLoad
	}
	Queries = append(Queries, goje.Where("{{.FromCol}} = ?", opt.{{camel .ToCol}}))

	rows, err := Get{{camel .RefTable}}(opt.ctx, Queries...)

	if err != nil {
		return err
	}

	rowsWithParent := make([]{{camel (singular .RefTable)}}, len(rows))
	for i := 0; i < len(rows); i++ {
		obj := rows[i]
		obj.parent = *opt
		rowsWithParent[i] =  obj
	}

	opt.{{camel .RefTable}}By{{camel .FromCol}}Rel = &rowsWithParent

	return nil
}
{{end}}

`
