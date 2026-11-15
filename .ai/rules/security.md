# Security

## JWT
- Protected endpoints (`/v1/*`) require `Authorization: Bearer <JWT>`; missing, malformed, invalid or expired tokens return `401`.
- Signing: HS256 with `JWT_SECRET`. The catalog only validates tokens; it does not issue them.
- Claims (`internal/pkg/auth/jwt.go`, `UserClaims`): `user_id` and `roles` (array), plus the registered `iat`/`exp`.
- Public endpoints (no auth): `GET /ping`, `GET /metrics`, `GET /swagger/*`.
- Auth middleware must not make extra DB calls — rely entirely on the JWT payload.

## RBAC Enforcement
- Map roles from the JWT payload; a token without any role of the route's group gets `403`.
- Role constants and groups live in `internal/pkg/auth/roles.go`: `ADMIN`, `ATTENDANT`, `MECHANIC`, `CUSTOMER`; groups `AttendantRoles` (ATTENDANT, ADMIN) and `AttendantAndMechanicRoles` (ATTENDANT, MECHANIC, ADMIN). ADMIN is included in every group.
- `HasRequiredRoles` accepts a token that has at least one of the route's roles.

## Internal Endpoints
- `/internal/*` requires `X-Internal-Token` matching `INTERNAL_TOKEN` (constant-time comparison). An empty configured token rejects every request.
- The gateway never routes `/internal/*`; keep it that way in `gateway.tf` (platform repository).

## Secrets
- `JWT_SECRET`, `INTERNAL_TOKEN` and `POSTGRES_PASSWORD` come from the environment (Secrets Manager in AWS). Never commit a real value; `.env.example` holds placeholders only.
- The development default `JWT_SECRET` in code is for local runs only; production must set it.
- Never log or return the DSN or any secret.

## Sensitive Data — Never Expose or Log
The catalog holds no personal data (no CPF, CNPJ, e-mail or password). Still:
- Tokens, the internal token and database credentials must never appear in responses, headers, log entries or error messages.
- The logger redacts keys named `password`, `token`, `jwt`, `authorization`, `x-internal-token`, `internal_token`, `jwt_secret`, `cpf`, `cnpj`, `document`, `documento` and any DSN password (`internal/pkg/logger/logger.go`).
- Avoid `zap.Any` for domain or request objects; log explicit fields (`zap.String("supply_id", id)`).

## Input Sanitization
- Decode bodies with `internal/pkg/json` in the controller; malformed JSON is a `400` whose `errors` list carries only the sentinel text (for example `invalid JSON syntax`).
- Validate with the domain methods (`Work.Validate()`, `Supply.Validate()`, `domain.NewReservation`) — never trust raw strings in services or domain.
- Never expose stack traces or internal error details in production responses.
