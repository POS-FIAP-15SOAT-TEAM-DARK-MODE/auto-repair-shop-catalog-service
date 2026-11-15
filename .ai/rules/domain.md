# Domain

Catalog service: owns the works (billable services) and supplies (parts with stock),
and the stock reservations that the Service Order saga takes and releases.

## Ubiquitous Language

| Code name | Business concept |
|---|---|
| `Work` | Billable service: name, description, unit price, `ACTIVE` / `INACTIVE` |
| `Supply` | Part with name, description, unit price, stock quantity and optimistic-lock `version` |
| `Reservation` | Stock held for one supply on one service order (`RESERVED` → `RELEASED`) |

## Validation Rules

| Entity | Rule | Error (HTTP) |
|---|---|---|
| Work | name ≥ 3 chars; description ≥ 10 chars; price > 0; status `ACTIVE`/`INACTIVE` | 400 |
| Supply | name ≥ 3 chars; description ≥ 10 chars; unit price > 0; stock ≥ 0; version ≥ 0 | 400 |
| Supply update | `version` required and ≥ 1; stored version must still match | 400 / 409 |
| Reservation | service order id required; quantity > 0 | 400 / 422 |

## Stock Reservation Rules (see `docs/adr/0002-stock-reservation-semantics.md`)

- One active (`RESERVED`) reservation per (service order, supply). A second one is `409`.
- A released reservation can be reserved again: the row is reactivated in place, so its id is stable for the pair.
- Reserving checks stock, inserts the reservation and decrements stock in the same transaction. The decrement is conditional (`stock_quantity >= qty`); insufficient stock is `409`; unknown supply is `404`.
- Release (by reservation id or by service order) returns the stock once. Releasing an already released reservation is a no-op (`204` by id); an unknown id is `404`.
- `ReleaseStock` commands are deduplicated by `message_id` in `processed_message`, in the same transaction as the release.
- Supply updates (`PUT /v1/supplies/:id`) are optimistic: the write matches the `version` the client read. A stale version is `409` and nothing is written, so a stock value read before a reservation cannot overwrite it.

## RBAC (public `/v1` API)

| Route group | Roles |
|---|---|
| `GET /v1/works` | ATTENDANT, MECHANIC, ADMIN |
| `POST/PUT/DELETE /v1/works[/:id]` | ATTENDANT, ADMIN |
| `/v1/supplies` (all methods) | ATTENDANT, MECHANIC, ADMIN |

`/internal/*` is not role-based: it requires the `X-Internal-Token` header (service-to-service).
Role groups live in `internal/pkg/auth/roles.go`.
