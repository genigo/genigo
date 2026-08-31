package generator

import (
	"testing"

	"github.com/genigo/genigo/internal/config"
)

func setDriver(driver string) func() {
	prev := config.Conf.DB.Driver
	config.Conf.DB.Driver = driver
	return func() { config.Conf.DB.Driver = prev }
}

func sampleTable() Table {
	return Table{
		Name:         "users",
		SingularName: "User",
		Primary:      "id",
		PrimaryColType: "int64",
		Primaries:    []Column{{Name: "id", GoDataType: "int64", AI: true}},
		Columns: []Column{
			{Name: "id", AI: true, GoDataType: "int64"},
			{Name: "name", GoDataType: "string"},
			{Name: "email", GoDataType: "string"},
			{Name: "created_at", Default: "CURRENT_TIMESTAMP", GoDataType: "time.Time"},
		},
	}
}

func TestPgInsertCall(t *testing.T) {
	defer setDriver("postgres")()

	want := "handler.DB.QueryRowContext(handler.Ctx, `INSERT INTO \"users\"(\"name\",\"email\") VALUES($1,$2) RETURNING \"id\",\"created_at\"`, opt.Name,opt.Email)"
	if got := pgInsertCall(sampleTable()); got != want {
		t.Errorf("pgInsertCall:\n got: %s\nwant: %s", got, want)
	}
}

func TestPgInsertCallAllDefaulted(t *testing.T) {
	defer setDriver("postgres")()

	tbl := Table{
		Name:  "counters",
		Columns: []Column{
			{Name: "id", AI: true, GoDataType: "int64"},
			{Name: "hits", GoDataType: "int32"},
		},
	}
	want := "handler.DB.QueryRowContext(handler.Ctx, `INSERT INTO \"counters\"(\"hits\") VALUES($1) RETURNING \"id\"`, opt.Hits)"
	if got := pgInsertCall(tbl); got != want {
		t.Errorf("pgInsertCall:\n got: %s\nwant: %s", got, want)
	}
}

func TestPgUpdateStmt(t *testing.T) {
	defer setDriver("postgres")()

	want := "`UPDATE \"users\" SET \"name\" = $1,\"email\" = $2,\"created_at\" = $3 WHERE \"id\" = $4`, opt.Name,opt.Email,opt.CreatedAt,opt.Id"
	if got := pgUpdateStmt(sampleTable()); got != want {
		t.Errorf("pgUpdateStmt:\n got: %s\nwant: %s", got, want)
	}
}

func TestPgDeleteStmt(t *testing.T) {
	defer setDriver("postgres")()

	want := "`DELETE FROM \"users\" WHERE \"id\" = $1`, opt.Id"
	if got := pgDeleteStmt(sampleTable()); got != want {
		t.Errorf("pgDeleteStmt:\n got: %s\nwant: %s", got, want)
	}
}

func TestPgSelectStmts(t *testing.T) {
	defer setDriver("postgres")()

	want := "`SELECT \"id\",\"name\",\"email\",\"created_at\" FROM \"users\" WHERE \"id\" = $1`, Id"
	if got := pgSelectByPKStmt(sampleTable()); got != want {
		t.Errorf("pgSelectByPKStmt:\n got: %s\nwant: %s", got, want)
	}

	want = "`SELECT \"id\",\"name\",\"email\",\"created_at\" FROM \"users\" WHERE \"email\" = $1`, EmailArg"
	if got := pgSelectByUniqueStmt(sampleTable(), []Column{{Name: "email"}}); got != want {
		t.Errorf("pgSelectByUniqueStmt:\n got: %s\nwant: %s", got, want)
	}
}

func TestNullableTypes(t *testing.T) {
	defer setDriver("postgres")()

	if needsNullWrapper("[]string") {
		t.Error("[]string carries NULL natively and should stay bare on postgres")
	}
	if needsNullWrapper("interface{}") {
		t.Error("interface{} carries NULL natively and should stay bare")
	}
	if !needsNullWrapper("string") {
		t.Error("string needs a sql.NullString wrapper")
	}
	if got := makeNullable("decimal.Decimal"); got != "decimal.NullDecimal" {
		t.Errorf("nullable decimal on postgres = %s, want decimal.NullDecimal", got)
	}
}

func TestNormalizePGDefault(t *testing.T) {
	cases := map[string]string{
		"":                        "",
		"now()":                   "CURRENT_TIMESTAMP",
		"CURRENT_TIMESTAMP":       "CURRENT_TIMESTAMP",
		"gen_random_uuid()":       "CURRENT_TIMESTAMP",
		"nextval('s'::regclass)":  "CURRENT_TIMESTAMP",
		"'active'::user_status":   "'active'::user_status",
		"0":                       "0",
		"true":                    "true",
		"'a(b)'::text":            "'a(b)'::text",
	}
	for in, want := range cases {
		if got := normalizePGDefault(in); got != want {
			t.Errorf("normalizePGDefault(%q) = %q, want %q", in, got, want)
		}
	}
}
