# auto-repair-shop-catalog-service

Catalog service of the Auto Repair Shop: works (billable services) and supplies
(parts with stock), plus the stock reservations taken by the Service Order saga.

- Port `8082`, PostgreSQL database `catalog`, secret `auto-repair-shop-<env>/catalog`.
- Contracts: `docs/fase4/contracts/` in the platform repository (resources, REST, messaging, DDL).

## Endpoints

Public (JWT, role-based):

| Method | Path | Roles |
|---|---|---|
| GET | `/v1/works` | ATTENDANT, MECHANIC, ADMIN |
| POST / PUT / DELETE | `/v1/works[/:id]` | ATTENDANT, ADMIN |
| GET | `/v1/supplies` | ATTENDANT, MECHANIC, ADMIN |
| POST / PUT / DELETE | `/v1/supplies[/:id]` | ATTENDANT, MECHANIC, ADMIN |

Internal (header `X-Internal-Token`, never routed by the gateway):

| Method | Path | Result |
|---|---|---|
| GET | `/internal/works/{id}` | `200` `{id,name,description,unit_price,active}` · `404` |
| POST | `/internal/reservations` | `201` `{reservation_id,supply_id,name,description,unit_price,version}` · `409` insufficient stock or already reserved · `404` supply · `422` quantity ≤ 0 |
| DELETE | `/internal/reservations/{id}` | `204` (idempotent) · `404` unknown id |

Response examples (same as `docs/fase4/contracts/internal-rest.md`):

```json
GET /internal/works/9b2d6a0e-3c0f-4a0b-8d57-2a5f6a1c9e11
{"id":"9b2d6a0e-3c0f-4a0b-8d57-2a5f6a1c9e11","name":"Troca de óleo","description":"Troca de óleo do motor com filtro incluído","unit_price":"150.00","active":true}

POST /internal/reservations  {"service_order_id":"01JD8Z4R3Q8M2V7K5N9XWABCDE","supply_id":"…","quantity":2}
201 {"reservation_id":"…","supply_id":"…","name":"Filtro","description":"Filtro de óleo","unit_price":"35.50","version":3}
```

`version` is the supply's `version` after the stock decrement.

Operational: `GET /ping`, `GET /metrics`, `GET /swagger/index.html`.

## Messaging

Consumes `ReleaseStock` from `auto-repair-shop-<env>-catalog-queue` (SNS topic
`saga-commands`, filter `target=catalog`). Releases every reserved stock of the
command's service order. Idempotent by `message_id`. Undecodable or unknown messages
are deleted with a log line; handler errors keep the message for redelivery (DLQ after 5).
The consumer runs only when `CATALOG_QUEUE_URL` is set.

## Running locally

```sh
cp .env.example .env          # fill in local values; never commit .env
make up                       # postgres, migrations, MiniStack (SQS/SNS), API on :8082
make token ROLE=ATTENDANT     # prints a development JWT (uses JWT_SECRET)
curl -H "Authorization: Bearer $(make -s token ROLE=ATTENDANT)" localhost:8082/v1/works
make down                     # removes containers and volumes
```

`SEED_ENABLED=true` loads the five works and five supplies with the fixed ids of the
contracts. Without Docker, `make run` needs a reachable PostgreSQL (`POSTGRES_*`) and
`make migrate-up` needs the `migrate` CLI.

## Tests

```sh
make test                     # unit tests with the race detector
make coverage                 # coverage gate used by CI (>= 80%, generated mocks excluded)
make test-integration         # needs PostgreSQL with migrations applied and MiniStack on :4566
```

Integration tests (`internal/integration`, build tag `integration`) run the real
HTTP API and consumer against PostgreSQL and MiniStack.

## Layout

```
cmd/service          HTTP API + SQS consumer entrypoint
cmd/devtoken         development JWT generator
internal/app         bootstrap, routing, middleware, seed, DB connection
internal/work        works domain (public API and internal lookup)
internal/supply     supplies domain (public API)
internal/reservation stock reservations (internal API and ReleaseStock handler)
internal/pkg/messaging  SQS consumer, envelope, dedupe
migrations           golang-migrate SQL (catalog schema)
k8s/manifests        Kustomize base and lab/stg/prd overlays
local/init-aws.sh    MiniStack topics and queues for local runs
docs/adr             architecture decision records
```

Decisions: [ADR 0002 - stock reservation semantics](docs/adr/0002-stock-reservation-semantics.md).
