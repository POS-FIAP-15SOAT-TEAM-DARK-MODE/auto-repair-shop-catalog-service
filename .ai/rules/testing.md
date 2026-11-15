# Testing

## Coverage
- **CI gate** (`coverage` job in `.github/workflows/ci.yml` and `make coverage`): total statement coverage ≥ **80%**, counted across all packages with `-coverpkg` and excluding generated mocks (`/mocks/`).
- Measure before claiming a number. The current measured value is in `README.md` (section *Tests*).
- No per-domain thresholds are enforced by tooling today.

## Commands
```bash
make test                                   # all tests with the race detector (LOG_LEVEL=PANIC silences logs)
make coverage                               # CI coverage gate (coverage.out, mocks excluded)
go test ./internal/reservation/... -v -race
make test-integration                       # build tag integration; needs PostgreSQL + MiniStack
```

## Unit Tests — In-Memory Implementations
- Repositories: `repository.NewInMemory()` in each domain's `repository/` (`work`, `supply`, `reservation`).
- UoW: `uow.NewInMemoryUoW()` from `internal/pkg/uow`.
- Mocks (Mockery, excluded from coverage): `internal/supply/interfaces/mocks`, `internal/work/interfaces/mocks`, `internal/pkg/uow/mocks`.

Pattern:
```go
// Arrange
repo := repository.NewInMemory()
svc := NewService(uow.NewInMemoryUoW(), repo, ...)
// Act
result, err := svc.Create(ctx, input)
// Assert
assert.NoError(t, err)
```

## Integration Tests
- Build tag `integration`, package `internal/integration` (`flow_test.go`, `supply_lock_test.go`, `setup_test.go`).
- Run the real HTTP API and the SQS consumer against PostgreSQL (migrations applied) and MiniStack on `:4566`.
- They do not roll back: each test creates its own ids, and the consumer test deletes its queue on cleanup.

## Test Structure Rules
- Arrange / Act / Assert; tests must be fully independent.
- Naming: `TestXxx` or `TestType_Method`; use a descriptive name that states the condition and the expected outcome.
- Group cases with table tests and sub-tests (`t.Run`).
- Never share mutable state between tests.

## What to Cover
| Area | Critical cases |
|---|---|
| Reservation | Reserve success, already reserved (`409`), insufficient stock (`409`), unknown supply (`404`), quantity ≤ 0 (`422`), release idempotency |
| ReleaseStock | Releases all reservations of the order; duplicate `message_id` restores stock once |
| Supply | Validation (`400`), optimistic lock (`409`), missing or zero version (`400`), list filters and pagination |
| Work | Validation (`400`), status values, internal lookup (`active` flag), 404 |
| RBAC | `403` for out-of-scope roles, `401` for missing or invalid JWT, `401` for the internal token |
| Error mapping | `internal/pkg/web` status and body shape (`code`, `error` or `errors`) |
