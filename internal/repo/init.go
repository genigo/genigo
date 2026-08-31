package repo

import (
	"database/sql"

	"github.com/genigo/genigo/internal/config"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func Connect() error {
	dbc := config.Conf.DB
	driver := "mysql"

	if dbc.Driver == "postgres" {
		// postgres: connect to the target database itself (Schema = database name);
		// the introspector filters tables by the `public` namespace
		driver = "pgx"
	} else {
		// mysql: introspection reads the server wide INFORMATION_SCHEMA,
		// connect to it instead of the target schema
		dbc.Schema = "information_schema"
	}

	db, err := sql.Open(driver, dbc.String())

	if err != nil {
		return err
	}

	DB = db
	return db.Ping()
}
