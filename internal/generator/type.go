package generator

type Table struct {
	Name         string
	SingularName string
	Imports      []string
	Columns      []Column

	LRelations []Relation
	RRelations []Relation

	UniqIndexes map[string][]Column

	Primary        string
	PrimaryColType string
	Primaries      []Column
	AICol          *Column
	AIColName      string
}

type Column struct {
	Name               string `json:"name"`
	DataType           string `json:"data_type"`
	GoDataType         string `json:"go_data_type"`
	GoDataTypeUnsigned bool   `json:"go_data_type_unsigned"`
	ColumnType         string `json:"column_type"`
	Type               string `json:"type"`
	Nullable           uint8  `json:"nullable"`
	Default            string `json:"default"`
	Extra              string `json:"extra"`
	Length             int    `json:"length"`
	AI                 bool   `json:"ai"`
	Enum               []string
}

type Relation struct {
	FromCol  string
	ToCol    string
	RefTable string
	Table    *Table
}
