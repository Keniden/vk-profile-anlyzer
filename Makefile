GOOSE_DRIVER=postgres
GOOSE_DBSTRING=postgres://postgres:postgres@localhost:5432/cruddb?sslmode=disable
MIGRATIONS_DIR=./internal/pkg/db/migrations

migrate-up:
	goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" status

migrate-create:
	goose -dir $(MIGRATIONS_DIR) create ${name} sql
