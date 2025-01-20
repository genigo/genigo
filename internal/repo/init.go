package repo

import (
	"database/sql"

	"github.com/genigo/genigo/internal/config"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Connect() error {
	dbc := config.Conf.DB
	dbc.Schema = "information_schema"

	db, err := sql.Open("mysql", dbc.String())

	if err != nil {
		return err
	}

	DB = db
	return db.Ping()
}
