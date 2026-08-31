# :tomato: Genigo
Lightweight SQL to struct generator / ORM toolkit for `Golang`

`Default template uses `[Goje](https://github.com/genigo/goje)` as a helper dependency.`

## Steps
1. Database &rarr; Go struct:
	* init configs `genigo init` (or `genigo init --driver postgres`)
	* run `genigo` to build database models from the target database

2. Use the generated Go structs in your project.

## Install Genigo (Generator command line tool)

``` Bash
go install github.com/genigo/genigo@latest
```

### Supported Databases
- [x] MySQL
- [x] PostgreSQL

genigo connects to a **live** database, reads its schema and emits one
`<table>.generated.go` per table. On postgres the `schema` config value is the
**database** name and tables are introspected from the `public` schema
(v1 scope).

## Configuration (genigo.yaml)

``` YAML
db:
  driver: postgres        # mysql | postgres
  host: 127.0.0.1
  port: 5432
  user: postgres
  password: secret
  schema: mydb
  flags:                  # driver specific DSN parameters
    sslmode: require      # e.g. sslmode, application_name, ...
tags: [db, json]          # struct tags of generated fields
pkg: models               # package name of generated files
dir: ./models             # output directory
replace: true             # regenerate over existing files
```

## Example of using a generated model

``` Go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/genigo/goje"
	models "__PATH_TO_MODELS_DIR__"
)

func init() {
	// init a database connection; this also activates the matching dialect
	// (identifier quoting and placeholder style) for goje query builders
	err := goje.InitDB(&goje.DBConfig{
		Driver:      "postgres", // or "mysql"
		Host:        "127.0.0.1",
		Port:        5432,
		User:        "postgres",
		Password:    "***",
		Schema:      "mydb",
		MaxIdleTime: time.Second * 30,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	handler := goje.MakeHandler(context.Background()) // or goje.MakeTxHandler for transactions

	user := models.User{
		Name:   "Mahmoud",
		Status: models.UserStatusActive,
	}

	// INSERT with the generated key coming back:
	// mysql: LastInsertId | postgres: INSERT ... RETURNING
	err := user.Save(handler)
	if err != nil {
		log.Print(err)
	}
	log.Printf("user saved with id: %d", user.Id)

	// on postgres, server generated defaults (now(), gen_random_uuid(), ...)
	// are read back too: user.CreatedAt / user.Ref are filled after Save

	// change some props and update
	user.Balance = 10000
	user.Update(handler) // or user.Save(handler)

	// relations (foreign keys) become loaders
	user.LoadPostsByUserId(goje.Limit(10), goje.Order("id DESC"))

	// query builders write `?` placeholders; goje translates them to
	// $1, $2, ... automatically on postgres
	users, err := models.GetUsers(handler,
		goje.Where("name LIKE ?", "Mahmoud"),
		goje.WhereIn("id", 1, 2, 3, 100),
		goje.Limit(5),
	)
	fmt.Println(users, err)

	// triggers
	models.BeforeDeleteUser = func(ch *goje.Context, u *models.User) error {
		if u.Id == 100 {
			return fmt.Errorf("you are unable to delete user(id=100) because of this trigger :)")
		}
		return nil
	}
	models.AfterDeleteUser = func(ch *goje.Context, u *models.User) {
		log.Printf("the user(%+v) has been deleted", u)
	}

	// delete
	user.Delete(handler)

	// helpers
	models.GetUserById(handler, 100)
	models.GetUserByEmail(handler, "mahmoud@example.com") // by unique index
	countUsers, _ := models.CountUsers(handler, goje.Where("status = ?", "active"))

	// bulk insert; BulkInsertIgnoreUser skips duplicate keys
	// mysql: INSERT IGNORE | postgres: ON CONFLICT DO NOTHING
	_, _ = models.BulkInsertIgnoreUser(handler, []models.User{{Name: "A"}, {Name: "B"}})
}
```

## Postgres type mapping

| postgres type | go type |
|---|---|
| `smallint` / `integer` / `bigint` | `int16` / `int32` / `int64` |
| `boolean` | `bool` |
| `varchar` / `text` / `uuid` / `inet` / `interval` | `string` |
| `json` / `jsonb` | `interface{}` |
| `numeric` | `decimal.Decimal` ([shopspring](https://github.com/shopspring/decimal)) |
| `timestamp` / `timestamptz` / `date` | `time.Time` |
| `bytea` | `[]byte` |
| `text[]` / `int4[]` / `bool[]` / ... | `goje.StringArray` / `goje.Int32Array` / `goje.BoolArray` / ... |
| `CREATE TYPE ... AS ENUM` | generated enum type + constants (like mysql `ENUM`) |
| identity / `serial` columns | auto increment (excluded from inserts, read back via `RETURNING`) |

Nullable columns map to `sql.Null*` types; slices, `interface{}` and `[]byte`
carry `NULL` natively. Nullable `numeric` maps to `decimal.NullDecimal`.

## Known divergences between drivers

- **`INSERT IGNORE`**: mysql also downgrades data errors (truncation, invalid
  values) to warnings; postgres `ON CONFLICT DO NOTHING` skips unique
  violations only and errors loudly on bad data — usually what you want.
- **Arrays of timestamps** (`timestamptz[]`) map to `interface{}` for now.
- **`FIND_IN_SET`**: the goje helper emits `= ANY(string_to_array(col, ','))`
  on postgres.
- **goje placeholders**: keep writing `?` in `Where(...)` fragments — goje
  renumbers them to `$1..$n` per statement on postgres. A literal `?` inside a
  string constant of a hand-written fragment would be renumbered too.

## Development

``` Bash
make test          # unit + golden file tests against docker mysql/postgres
make integration   # compile generated models + full CRUD suite on both engines
make golden-update # refresh testdata/golden after an intentional change
make sync-goje     # copy vendor/github.com/genigo/goje into ../goje and test it there
```

goje is developed inside `genigo/vendor/github.com/genigo/goje` and synced to
its own repository (`make sync-goje`, direction: genigo &rarr; goje). Never
run `go mod vendor` with uncommitted goje edits — it would restore the
published goje version and drop them.
