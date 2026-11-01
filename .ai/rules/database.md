# Database

## PostgreSQL 15 Conventions
- Monetary columns: `NUMERIC(10,2)` — never `FLOAT` or `DOUBLE`.
- `stock_reservation.status`: `CHECK` constraint (`RESERVED`, `RELEASED`).
- IDs: `VARCHAR(36)` UUID; `stock_reservation.service_order_id` is a ULID (`VARCHAR(26)`) received from the Service Order service.
- Timestamps: `TIMESTAMPTZ`, UTC, ISO 8601.
- No foreign keys across databases: the Service Order and other services only hold ids.

## Migration Strategy
All schema changes via versioned migration files only:
```
migrations/NNNNNN_description.up.sql
migrations/NNNNNN_description.down.sql
```
Never alter the schema manually. Never drop or modify existing columns — add new, deprecate old.

Commands:
```bash
make migrate-up      # apply pending
make migrate-down    # roll back last
```

## Transaction Context Pattern
- `postgres.GetTransaction(ctx)` — gets `*sql.Tx` for mutating ops (called inside a UoW step).
- `postgres.GetOneTimeTransaction(ctx)` — gets a `*sql.DB` connection for read-only ops.

See `architecture.md` for the full UoW pattern.

## Required Indexes
`stock_reservation.service_order_id`, `processed_message` (primary key).

## Required Foreign Keys
`stock_reservation.supply_id → supply.id`.
`service_order_work→(service_order, work)`, `service_order_supply→(service_order, supply)`,
`customer→user`.
