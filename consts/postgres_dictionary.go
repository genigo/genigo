package consts

// PostgresDicMp maps normalized postgres type names to go types.
// The postgres introspector feeds format_type() output stripped of
// size modifiers: `character varying`, `text[]`, `timestamp with time zone`, ...
var PostgresDicMp = map[string]string{
	// integers
	"smallint": "int16",
	"integer":  "int32",
	"int":      "int32",
	"int4":     "int32",
	"bigint":   "int64",
	"int8":     "int64",
	"serial":   "int32",
	"bigserial": "int64",

	// boolean
	"boolean": "bool",
	"bool":    "bool",

	// strings
	"character varying": "string",
	"varchar":           "string",
	"character":         "string",
	"char":              "string",
	"text":              "string",
	"name":              "string",
	"citext":            "string",
	"uuid":              "string",
	"xml":               "string",

	// json
	"json":  "interface{}",
	"jsonb": "interface{}",

	// numeric
	"numeric":          "decimal.Decimal",
	"money":            "decimal.Decimal",
	"real":             "float32",
	"double precision": "float64",
	"float":            "float64",
	"float4":           "float32",
	"float8":           "float64",

	// date & time
	"date":                        "time.Time",
	"timestamp without time zone": "time.Time",
	"timestamp with time zone":    "time.Time",
	"timestamp":                   "time.Time",
	"timestamptz":                 "time.Time",
	"time without time zone":      "time.Time",
	"time with time zone":         "time.Time",
	"time":                        "time.Time",
	"interval":                    "string",

	// binary & network
	"bytea":   "[]byte",
	"inet":    "string",
	"cidr":    "string",
	"macaddr": "string",

	// enums are introspected into Column.Enum with DataType "enum"
	"enum": "string",

	// arrays: goje slice types decode the postgres {a,b} text format
	// through database/sql (timestamps[] map to interface{} for now)
	"text[]":              "goje.StringArray",
	"character varying[]": "goje.StringArray",
	"varchar[]":           "goje.StringArray",
	"uuid[]":              "goje.StringArray",
	"smallint[]":          "goje.Int16Array",
	"integer[]":           "goje.Int32Array",
	"int[]":               "goje.Int32Array",
	"int4[]":              "goje.Int32Array",
	"bigint[]":            "goje.Int64Array",
	"int8[]":              "goje.Int64Array",
	"real[]":              "goje.Float32Array",
	"float4[]":            "goje.Float32Array",
	"double precision[]":  "goje.Float64Array",
	"float8[]":            "goje.Float64Array",
	"boolean[]":           "goje.BoolArray",
}
