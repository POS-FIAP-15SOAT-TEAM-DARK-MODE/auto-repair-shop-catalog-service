# Workflow

## Conventional Commits

Format:
```
<type>(<scope>): <short description>

[optional body]

[optional footer(s)]
```

Types: `feat` · `fix` · `refactor` · `test` · `chore` · `docs` · `perf` · `style` · `revert`

Scopes: `catalog` (whole service) · `work` · `supply` · `reservation` · `messaging` · `auth` · `web` · `middleware` · `routing` · `api` · `db` · `infra` · `k8s` · `integration` · `ci` · `deps`

Rules:
- Subject max 72 chars, lowercase, no trailing period.
- Imperative mood: "add feature" not "added feature".
- Breaking changes: `feat(api)!: ...` or `BREAKING CHANGE:` footer.
- Reference issues in footer: `Closes #123`.
- Branch flow: `feature → develop → main` (ADR 0001).

## Pre-commit Hooks
Configured in `.pre-commit-config.yaml`:
- `trailing-whitespace`, `end-of-file-fixer`, `check-yaml`, `check-added-large-files`, `check-merge-conflict`
- `golangci-lint` (v2.9.0)
- `go fmt ./...`
- `go build`, `go-unit-tests`
- `go vet ./...`
- `check-ai-rules`: enforces `.ai/` as the only source of truth for rule files

## Development Commands
```bash
make run              # start the API on PORT (default 8082)
make test             # all tests with the race detector
make coverage         # CI coverage gate (mocks excluded)
make test-integration # integration tests (needs PostgreSQL + MiniStack)
make up / make down   # docker compose: postgres, migrations, MiniStack, API (down removes volumes)
make migrate-up       # apply pending migrations
make migrate-down     # roll back the last migration
make mockgen          # regenerate mocks via go generate
make token ROLE=ATTENDANT   # development JWT for local calls
```
