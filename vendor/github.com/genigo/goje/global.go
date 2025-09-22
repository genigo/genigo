package goje

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// Default Database connection that fill by
var DefatultDB *sql.DB

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
func NewDBConnection(conn *DBConfig) (*sql.DB, error) {
	if conn.Driver != "mysql" {
		return nil, ErrUnknownDBDriver
	}

	db, err := sql.Open(conn.Driver, conn.String())

	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(conn.MaxIdleTime)
	db.SetMaxIdleConns(conn.MaxIdleConns)
	db.SetMaxOpenConns(conn.MaxOpenConns)
	db.SetConnMaxLifetime(conn.ConnMaxLifetime)

	return db, nil
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
