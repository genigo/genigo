package goje

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Default Database connection that fill by
var DefatultDB *sql.DB

// Default slow query log threshold = 5s
var SlowQueryLogTimeout = time.Second * 5

// InitDB Connect default database
func InitDB(conn *DBConfig) error {
	db, err := NewDBConnection(conn)
	if err != nil {
		return err
	}
	DefatultDB = db
	return nil
}

// NewDBConnection Connect to database and return database
// It also activates the matching dialect (identifier quoting and
// placeholder style) for the package level query builders.
func NewDBConnection(conn *DBConfig) (*sql.DB, error) {
	driverName, err := resolveDriver(conn.Driver)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(driverName, conn.String())

	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(conn.MaxIdleTime)
	db.SetMaxIdleConns(conn.MaxIdleConns)
	db.SetMaxOpenConns(conn.MaxOpenConns)
	db.SetConnMaxLifetime(conn.ConnMaxLifetime)

	return db, nil
}

// resolveDriver maps a DBConfig.Driver value to the database/sql driver
// name and activates the matching package dialect.
func resolveDriver(driver string) (string, error) {
	switch driver {
	case "", DriverMysql:
		dialect = MysqlDialect
		return "mysql", nil
	case DriverPostgres:
		dialect = PostgresDialect
		// pgx/v5 stdlib registers itself under the "pgx" name
		return "pgx", nil
	}
	return "", ErrUnknownDBDriver
}

// GetHandler make a handler from default database and a TODO context
func GetHandler() *Context {
	return &Context{
		Ctx: context.TODO(),
		DB:  DefatultDB,
	}
}

// H is a shortcut for GetHandler
func H() *Context {
	return GetHandler()
}

// MakeHandler make a handler from default database
func MakeHandler(ctx context.Context) *Context {
	return &Context{
		Ctx: ctx,
		DB:  DefatultDB,
	}
}

// DefaultHandler make a handler from default database
func MakeTxHandler(ctx context.Context, options *sql.TxOptions) (*Context, error) {
	tx, err := DefatultDB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}

	return &Context{
		Ctx: ctx,
		DB:  tx,
		Tx:  true,
	}, nil
}

// MakeHandler make a handler from the database connection
func MakeHandlerDB(ctx context.Context, db *sql.DB) *Context {
	return &Context{
		Ctx: ctx,
		DB:  db,
	}
}

// DefaultHandler make a handler from the database connection
func MakeTxHandlerDB(ctx context.Context, db *sql.DB, options *sql.TxOptions) (*Context, error) {
	tx, err := db.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}

	return &Context{
		Ctx: ctx,
		DB:  tx,
		Tx:  true,
	}, nil
}
