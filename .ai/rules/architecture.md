# Architecture

## Layer Map

```
HTTP → routing (internal/app/routing) → controller (internal/<domain>/controller.go)
     → service (internal/<domain>/service.go) → repository (internal/<domain>/repository/)
     → PostgreSQL
                    ↕
        domain (internal/<domain>/domain/)   ports (internal/<domain>/interfaces/)
```

Domains: `work`, `supply`, `reservation`. Each one follows the same folder shape:
`adapters/` (DTOs and mapping), `domain/` (entities, validation, errors), `interfaces/` (ports),
`repository/` (Postgres and in-memory), `controller.go`, `service.go`, `di.go`.
`reservation/release_stock_handler.go` is the SQS handler for `ReleaseStock`.

## Layer Contracts
- **Controller**: parse query, path and body into DTOs, call the service, return the response. Zero business logic.
- **Service**: business rules and validation orchestration; wraps every mutation in `uow.Execute`.
- **Repository**: all SQL (raw, `database/sql`). No SQL outside `repository/`.
- **Domain**: framework-free structs, validation methods and errors. No Gin, no `lib/pq`.
- **Shared packages** (`internal/pkg/*`) must not import a domain package.

## DI Root
- `internal/app/bootstrap/bootstrap.go`: `httpContainer(db)` builds the controllers through each domain's `NewHTTPController(db)` / `NewInternalController(...)` (`di.go`) and passes them to `container.HandlersWrapper`.
- `internal/app/routing/routing.go`: registers the routes and the role middleware (`middleware.Auth`, `middleware.InternalAuth`).
- `internal/app/bootstrap` also builds the SQS consumer (only when `CATALOG_QUEUE_URL` is set).

## Repository DB Connection Pattern
Repositories hold no `*sql.DB`. The connection comes from the context, injected by the UoW:
- **Mutating ops** (inside a UoW step): `uow.GetTransaction(ctx)` returns the `*sql.Tx` opened by the UoW.
- **Read-only ops**: `uow.GetOneTimeTransaction(ctx)` returns the shared `*sql.DB` handle.

## Unit of Work
```go
unitOfWork := uow.NewTransactionalUoW(db)   // internal/pkg/uow
unitOfWork.Execute(ctx, step1, step2, ...)
```
`Execute` begins a transaction, runs the steps in order, then commits or rolls back.
Calls nested inside a transaction that is already in the context join it (`joinableUoW`):
the consumer uses this to run the dedupe insert and the handler in one transaction.
`uow.NewInMemoryUoW()` runs the steps without a database and is used by unit tests.

## Dynamic SQL
Use `db.QueryBuilder(baseQuery)` from `internal/pkg/db/` (used by the work and supply list queries):
```go
qb := db.QueryBuilder(base)
qb.Add("s.version =", value)        // skipped when value is nil or ""
qb.AddAny("s.id", ids)              // = ANY($N); skipped when empty
qb.OrderBy(field, db.ASC)
qb.AddPagination(limit, offset)
query, args := qb.Build()
```
