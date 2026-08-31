MYSQL_DSN ?= root:genigo@tcp(127.0.0.1:3307)/genigo_test
PG_DSN    ?= postgres://genigo:genigo@127.0.0.1:5433/genigo_test?sslmode=disable
GOJE_PATH ?= ../goje

.PHONY: build test unit golden-update compose-up compose-down sync-goje integration

build:
	go build -o genigo .

# database-free tests (statement builders, type mappings)
unit:
	go test ./internal/generator/ -run 'TestPg|TestNormalize|TestNullable'

# full suite: unit + golden files against the docker databases
test: compose-up
	TEST_MYSQL_DSN=$(MYSQL_DSN) TEST_POSTGRES_DSN=$(PG_DSN) go test ./...

# refresh testdata/golden after an intentional template change
golden-update: compose-up
	TEST_MYSQL_DSN=$(MYSQL_DSN) TEST_POSTGRES_DSN=$(PG_DSN) GOLDEN_UPDATE=1 \
		go test ./internal/generator/ -run TestGenerateGolden

compose-up:
	docker compose up -d --wait

compose-down:
	docker compose down -v

# copy the vendored goje package into the standalone goje repository and run
# its test suite there (`go test` cannot see packages inside vendor/)
sync-goje:
	cp vendor/github.com/genigo/goje/*.go $(GOJE_PATH)/
	cd $(GOJE_PATH) && go test ./...

# end to end: compile the generated models and exercise them on both engines
integration: compose-up build sync-goje
	cd integration && \
		../genigo -c genigo-mysql.yaml && \
		../genigo -c genigo-postgres.yaml && \
		TEST_MYSQL_DSN='root:genigo@tcp(127.0.0.1:3307)/genigo_test' \
		TEST_POSTGRES_DSN='postgres://genigo:genigo@127.0.0.1:5433/genigo_test?sslmode=disable' \
		go test -v ./...
