# ADR 0002 - Stock reservation semantics

Status: accepted

## Context

The Service Order saga must hold supply units while a budget is being approved and
return them when the order is cancelled, rejected or fails. Commands can be delivered
more than once (SQS is at-least-once), and the Service Order service retries nothing
on its own. The catalog is the only owner of supply stock.

## Decision

- A reservation is identified by `(service_order_id, supply_id)` while `RESERVED`.
  A second active reservation for the same pair is a conflict (`409`).
- Reserving decrements stock with a conditional update (`stock_quantity >= qty`).
  Zero affected rows means insufficient stock (`409`). The reservation row and the
  decrement share one transaction.
- Releasing (by id or by service order) flips `RESERVED` to `RELEASED` and returns the
  stock in the same transaction. A reservation already `RELEASED` is a no-op
  (`204`); an unknown id is `404`.
- A released reservation can be reserved again. The row is reactivated in place, so
  the reservation id is stable for the pair.
- `ReleaseStock` commands are consumed with per-message deduplication
  (`processed_message`). The dedupe insert and the release run in one transaction:
  a failed release leaves no dedupe record, so SQS redelivers it, and a duplicate is
  acknowledged without touching stock.
- Stock is restored at most once per reservation, guaranteed by the status
  transition (`WHERE status = 'RESERVED'`), not only by deduplication.

- Supply updates (`PUT /v1/supplies/:id`) use an optimistic lock: the caller sends the
  `version` it read, and the write is `WHERE id = $1 AND version = $6`. A stale version
  returns `409`, so a stock value read before a reservation cannot overwrite it.

## Consequences

- Concurrent reservations of the same supply serialize on the conditional update.
- Compensation is safe to repeat: releasing twice, or releasing an order with nothing
  reserved, changes nothing.
- The Service Order service must treat `409` on reserve as a business failure of the
  step, not as a transient error, because retries of `POST` are not automatic.
