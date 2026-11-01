# API

## Versioning
Public endpoints under `/v1/` (Swagger UI at `http://localhost:8082/swagger/index.html`). Service-to-service endpoints under `/internal/` (never routed by the gateway; require `X-Internal-Token`).

## Endpoints

```
GET    /v1/works                                   ATTENDANT, MECHANIC, ADMIN
POST   /v1/works                                   ATTENDANT, ADMIN
PUT    /v1/works/:id                               ATTENDANT, ADMIN
DELETE /v1/works/:id                               ATTENDANT, ADMIN
GET    /v1/supplies                                ATTENDANT, MECHANIC, ADMIN
POST   /v1/supplies                                ATTENDANT, MECHANIC, ADMIN
PUT    /v1/supplies/:id                            ATTENDANT, MECHANIC, ADMIN
DELETE /v1/supplies/:id                            ATTENDANT, MECHANIC, ADMIN

GET    /internal/works/:id                         X-Internal-Token
POST   /internal/reservations                      X-Internal-Token   (201 | 409 | 404 | 422)
DELETE /internal/reservations/:id                  X-Internal-Token   (204, idempotent)
```

## Request/Response Conventions
- All bodies: JSON.
- Timestamps: UTC ISO 8601.
- IDs: UUID strings (ULID strings for service order IDs).
- Pagination (list endpoints): `?page=N&pageSize=N`.

## Error Format
```json
{ "error": "Human-readable message", "code": "MACHINE_READABLE_CODE" }
```

| HTTP | Meaning |
|---|---|
| 400 | Validation error |
| 401 | Missing or invalid auth |
| 403 | Forbidden (wrong role) |
| 404 | Not found |
| 409 | Conflict |
| 422 | Business rule violation |
| 500 | Unexpected error |

## Swagger
Every endpoint must have: summary, description, request body schema, response schemas (success + errors), 401/403 documented on all protected endpoints, and role requirements noted.
