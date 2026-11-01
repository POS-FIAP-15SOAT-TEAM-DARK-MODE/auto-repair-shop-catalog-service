-include .env
export

DB_URL = postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/$(or $(POSTGRES_DB),catalog)?sslmode=disable

run:
	go run cmd/service/main.go

test:
	LOG_LEVEL=PANIC go test ./... --race -v

test-integration:
	BCRYPT_COST=4 go test -tags=integration ./internal/integration/... -v

coverage:
	LOG_LEVEL=PANIC go test ./... --coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1

# Signed development JWT for local calls, e.g. `make token ROLE=MECHANIC`.
token:
	@go run ./cmd/devtoken -role $(or $(ROLE),ATTENDANT)

up:
	docker compose up --build

down:
	docker compose down -v

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down

mockgen:
	go generate ./...

.PHONY: run test test-integration coverage token up down migrate-up migrate-down mockgen
