# Database

## PostgreSQL 15 Conventions
- Monetary columns: `NUMERIC(10,2)` — never `FLOAT` or `DOUBLE`.
- `work.status`: `BOOLEAN` (`TRUE` = ACTIVE); exposed as `ACTIVE`/`INACTIVE` by the domain.
- `stock_reservation.status`: `CHECK` constraint (`RESERVED`, `RELEASED`).
- IDs: `VARCHAR(36)` UUID for `work`, `supply`, `stock_reservation`; `stock_reservation.service_order_id` is a ULID (`VARCHAR(26)`) received from the Service Order service; `processed_message.message_id` is `UUID`.
- Timestamps: `TIMESTAMPTZ`, UTC, ISO 8601.
- No foreign keys across databases: the Service Order and other services only hold ids.

## Schema (`migrations/000001_init_schema`)

| Table | Notes |
|---|---|
| `work` | `unit_price NUMERIC(10,2)`, `status BOOLEAN` |
| `supply` | `unit_price NUMERIC(10,2)`, `stock_quantity INT ≥ 0` (app rule), `version INT` (optimistic lock) |
| `stock_reservation` | `UNIQUE (service_order_id, supply_id)`, `quantity CHECK > 0`, `supply_id → supply(id)`, `released_at` |
| `processed_message` | dedupe of consumed messages (primary key `message_id`) |

## Migration Strategy
All schema changes via versioned migration files only:
```
migrations/NNNNNN_description.up.sql
migrations/NNNNNN_description.down.sql
```
Never alter the schema manually. Never drop or modify existing columns — add new, deprecate old.
Migrations already applied are never edited.

Commands:
```bash
make migrate-up      # apply pending (needs the golang-migrate CLI)
make migrate-down    # roll back the last one
```

## Transaction Context Pattern
- `uow.GetTransaction(ctx)` — `*sql.Tx` for mutating ops (called inside a UoW step).
- `uow.GetOneTimeTransaction(ctx)` — shared `*sql.DB` for read-only ops.

See `architecture.md` for the full UoW pattern.

## Required Indexes
- `stock_reservation.service_order_id` (`idx_stock_reservation_service_order_id`).
- `stock_reservation (service_order_id, supply_id)` is covered by its `UNIQUE` constraint.
- `processed_message` primary key.

## Required Foreign Keys
- `stock_reservation.supply_id → supply.id`.
