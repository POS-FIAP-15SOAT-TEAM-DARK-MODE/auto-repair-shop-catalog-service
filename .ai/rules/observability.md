# Observability

## Logger
Structured JSON via Zap, initialized in `internal/pkg/logger/logger.go` (service field `SERVICE_NAME`, level `LOG_LEVEL`). Sensitive keys are redacted there (see `security.md`). Logs carry the request id and correlation fields (`message_id`, `saga_id`, `service_order_id`) when present.

## Log Levels
| Level | Use |
|---|---|
| `INFO` | Business events and lifecycle (reservation released, message processed, server started) |
| `WARN` | Expected rejections and database errors mapped by `internal/pkg/db/postgres` |
| `ERROR` | Unexpected failures: 5xx responses, infrastructure 409/422, consumer failures, panics |

## Event Names in Use
Reuse these names; do not invent new ones for the same operation.

| Event | Where | Fields |
|---|---|---|
| `request.started` / `request.completed` | `internal/app/middleware/logger.go` | request id, method, route, status |
| `reservation.released` | `internal/reservation/release_stock_handler.go` | correlation fields of the envelope |
| `message.processed` | `internal/pkg/messaging/consumer.go` | correlation fields |
| `message.duplicate` | consumer (deduplicated, not re-run) | correlation fields |
| `message.discarded` | consumer | `reason`: `undecodable` or `unknown_type` |
| `server.started` / `server.stopping` | `internal/app/server.go` | `name`, `addr` |
| `seed: …` | `internal/app/seed.go` | seed progress |

Operation logs in services and controllers use `operation` (for example `create_supply`, `list_works`) at `DEBUG`.

## Sensitive Field Redaction
- Prefer explicit fields (`zap.String("supply_id", id)`) over `zap.Any("supply", obj)`.
- Never log raw request bodies or DTOs; log errors with `zap.Error(err)` only.

## Errors and Read Operations
- Keep success logs of read operations minimal (debug or omit).
- Database errors are logged with their PostgreSQL code by `internal/pkg/db/postgres` (`WARN`); responses that are 5xx or infrastructure 409/422 are logged at `ERROR` with the status by `routing.respondError`.
