module github.com/genigo/genigo/integration

go 1.24.0

require (
	github.com/genigo/goje v0.2.0
	github.com/shopspring/decimal v1.4.0
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/go-sql-driver/mysql v1.9.3
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.4 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

// tests run against the local development copy of goje
// (../../goje resolves to the goje repository checked out beside genigo)
replace github.com/genigo/goje => ../../goje
