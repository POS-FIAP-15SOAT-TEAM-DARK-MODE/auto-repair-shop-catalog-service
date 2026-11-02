# Playbook: Scaffold New Domain

Use this playbook when adding a new domain next to `work`, `supply` and `reservation` (e.g., `vehicle-part`).
Follow the layout of an existing domain: `internal/supply/`.

## 1. Domain — `internal/<name>/domain/`
- Struct with a `google/uuid` id (or ULID when the id is ordered by creation time) and validation in `Validate() error`.
- Domain errors as `var ErrXxx = errors.New("...")` in `<name>_errors.go`.
- No imports from Gin, `lib/pq` or any package outside `domain/`.

## 2. Ports — `internal/<name>/interfaces/`
- `<name>_repository.go`, `<name>_service.go`, `<name>_controller.go`.
- Add `//go:generate go run github.com/vektra/mockery/v2@latest --name=<Name>Repository --with-expecter` above each interface a test will mock.

## 3. Repository — `internal/<name>/repository/`
- `<name>_postgres.go`: stateless struct; connection from `uow.GetTransaction(ctx)` (mutations) or `uow.GetOneTimeTransaction(ctx)` (reads). Raw SQL constants, no ORM.
- `<name>_in_memory.go`: `NewInMemory()` for unit tests only.

## 4. DTOs — `internal/<name>/adapters/`
- Request and response structs with JSON tags, and `DomainToResponse` mappers.

## 5. Service — `internal/<name>/service.go`
- Constructor takes `uow.Executor` and the repository interfaces.
- All business rules here; wrap every mutation in `unitOfWork.Execute(ctx, func(txCtx context.Context) error { ... })`.
- No SQL, no HTTP types.

## 6. Controller — `internal/<name>/controller.go`
- Parse path, query and body (`internal/pkg/json`); call the service; return the response DTO.
- Zero business logic.

## 7. Wire
- `internal/<name>/di.go`: `NewHTTPController(db *sql.DB)` builds repository → service → controller.
- `internal/app/container/handlers.go`: add the field to `HandlersWrapper`.
- `internal/app/bootstrap/bootstrap.go`: fill the field in `httpContainer`.
- `internal/app/routing/routing.go`: register routes with `middleware.Auth(<roles>...)`.
- `internal/pkg/web/errors.go`: add the domain errors to the right status list.

## 8. Migration
Follow `add-migration.md`.

## 9. Tests
- Unit: service with `repository.NewInMemory()` and `uow.NewInMemoryUoW()`.
- Integration: add a flow in `internal/integration/` (build tag `integration`).
- Update `docs/swagger.yaml` and `.ai/rules/api.md` for every new route.

## 10. Regenerate Mocks
```bash
make mockgen
```
