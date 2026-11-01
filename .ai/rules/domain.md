# Domain

Catalog service: owns the works (billable services) and supplies (parts with stock),
and the stock reservations that the Service Order saga takes and releases.

## Ubiquitous Language

| Code name | Business concept |
|---|---|
| `Work` | Billable service: name, description, unit price, ACTIVE/INACTIVE |
| `Supply` | Part with stock quantity, unit price, optimistic-lock `Version` |
| `Reservation` | Stock held for one supply on one service order (`RESERVED` → `RELEASED`) |

## Stock Reservation Rules

- One active (`RESERVED`) reservation per (service order, supply). Releasing and reserving again reactivates the same row.
- Reserving decrements stock and inserting the reservation happen in the same transaction; insufficient stock → `409`.
- Release (by id or by service order) returns the stock once; releasing an already released reservation is a no-op.
- `ReleaseStock` commands are deduplicated by `message_id` in `processed_message`, in the same transaction as the effect.

## RBAC (public `/v1` API)

| Route group | Roles |
|---|---|
| `/v1/works` read | ATTENDANT, MECHANIC, ADMIN |
| `/v1/works` write | ATTENDANT, ADMIN |
| `/v1/supplies` | ATTENDANT, MECHANIC, ADMIN |

`/internal/*` is not role-based: it requires the `X-Internal-Token` header (service-to-service).
