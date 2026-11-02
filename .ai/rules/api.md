# API

## Versioning
Public endpoints under `/v1/`. Swagger UI at `http://localhost:8082/swagger/index.html`, OpenAPI document at `/swagger.yaml`.
Service-to-service endpoints under `/internal/` (never routed by the gateway; require `X-Internal-Token`).
Operational endpoints: `GET /ping`, `GET /metrics`.

## Endpoints (exactly the routes registered in `internal/app/routing/routing.go`)

```
GET    /v1/works                    ATTENDANT, MECHANIC, ADMIN
POST   /v1/works                    ATTENDANT, ADMIN                  201
PUT    /v1/works/:id                ATTENDANT, ADMIN                  200
DELETE /v1/works/:id                ATTENDANT, ADMIN                  204
GET    /v1/supplies                 ATTENDANT, MECHANIC, ADMIN
POST   /v1/supplies                 ATTENDANT, MECHANIC, ADMIN        201
PUT    /v1/supplies/:id             ATTENDANT, MECHANIC, ADMIN        200 (body: version)
DELETE /v1/supplies/:id             ATTENDANT, MECHANIC, ADMIN        204

GET    /internal/works/:id          X-Internal-Token                  200 · 404
POST   /internal/reservations       X-Internal-Token                  201 · 400/404/409/422
DELETE /internal/reservations/:id   X-Internal-Token                  204 (idempotent) · 404
```

`/ping`, `/metrics`, `/swagger/index.html` and `/swagger.yaml` need no authentication.
The catalog issues no tokens: JWTs come from the authentication service (`cmd/devtoken` makes development tokens).

## Request/Response Conventions
- All bodies: JSON. Money: decimal strings (`"150.00"`).
- Timestamps: UTC ISO 8601.
- IDs: UUID strings; `service_order_id` is a ULID string received from the Service Order service.
- List endpoints: `?page=N&pageSize=N` (defaults 1 and 10). Invalid or non-positive values silently fall back to the defaults. Lists answer `{items, totalItems, totalPages, pageSize, page}`.
- `GET /v1/works?status=` accepts `ACTIVE` or `INACTIVE`; any other value is ignored.

## Error Format
Two shapes, both always carry `code` (the HTTP status):

```json
{ "code": 404, "error": "supply not found" }
{ "code": 400, "errors": ["supply name must be at least 3 characters long"] }
```

- `errors` (list): validation failures (400). One message per failed sentinel; messages are domain-safe texts.
- `error` (string): every other error (401, 403, 404, 409, 422, 500). Infrastructure failures always map to generic text (`conflict`, `invalid data`, `internal server error`); the cause is only logged.
- Middleware answers (`internal/app/middleware`): `401` `{"code":401,"error":"Unauthorized"}` (or `"Token expired"`), `403` `{"code":403,"error":"Forbidden"}`, internal token `401` `{"code":401,"error":"unauthorized"}`.

| HTTP | Meaning |
|---|---|
| 400 | Validation error (`errors` list) |
| 401 | Missing or invalid JWT / internal token |
| 403 | Forbidden (role not allowed) |
| 404 | Not found |
| 409 | Conflict (version conflict, already reserved, insufficient stock) |
| 422 | Business rule violation (reservation quantity ≤ 0) |
| 500 | Unexpected error |

## Swagger
`docs/swagger.yaml` documents every route above, with summary, request and response schemas, the error schemas, the role requirements and the 401/403 responses of protected endpoints. A new or changed route must update it in the same change.
