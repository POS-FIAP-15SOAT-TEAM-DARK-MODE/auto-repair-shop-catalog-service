# AI Rules — Master Index

Single source of truth for all AI agent rules in this project.
**Never duplicate content** in agent pointer files — those files only contain a routing table and 3 always-apply rules.

## Folder Map

```
.ai/
  rules/               # topic-scoped rule files — open on demand
    architecture.md        layers, controller/service/repository, DI root, UoW, QueryBuilder
    domain.md              ubiquitous language, validation, stock reservation rules, RBAC
    go-conventions.md      Go idioms, decimal, UUID/ULID, no framework leakage, Mockery, config keys
    database.md            PostgreSQL conventions, schema, migrations, indexes, FK
    api.md                 endpoint reference, error format, Swagger
    testing.md             coverage gate, unit/integration patterns
    security.md            JWT, RBAC enforcement, internal token, secrets, PII
    observability.md       Zap, log levels, event names, redaction
    workflow.md            Conventional Commits, pre-commit hooks, dev commands
  commands/            # reusable playbooks (also Claude Code slash commands via symlinks)
    new-domain.md          scaffold a new domain end-to-end
    add-migration.md       create and apply a versioned migration
  hooks/               # pre-commit hook scripts
    check-ai-rules.sh      enforces .ai/ as the only rule source
```

## Topic Index

| Topic | File | Key contents |
|---|---|---|
| Architecture | `rules/architecture.md` | Layer diagram, layer contracts, DI root, UoW, QueryBuilder |
| Domain & Business Rules | `rules/domain.md` | Ubiquitous language, validation, stock reservation, RBAC |
| Go Conventions | `rules/go-conventions.md` | Tech stack, UUID/ULID, decimal, no framework leakage, Mockery, config |
| Database | `rules/database.md` | NUMERIC(10,2), schema, migrations, indexes, FK |
| API | `rules/api.md` | Endpoints with roles, error format, Swagger |
| Testing | `rules/testing.md` | 80% CI coverage gate, unit vs integration patterns |
| Security | `rules/security.md` | JWT, RBAC, internal token, secrets, PII rules |
| Observability | `rules/observability.md` | Zap, log levels, event names, redaction |
| Workflow | `rules/workflow.md` | Conventional Commits, pre-commit, dev commands |

## Task Routing — Required Reading by Task Type

| Task type | Files to read |
|---|---|
| New feature or domain change | `architecture.md` · `domain.md` · `go-conventions.md` |
| Database schema or migration | `database.md` · `go-conventions.md` |
| API endpoint | `api.md` · `architecture.md` · `security.md` |
| Auth, RBAC, or security | `security.md` · `domain.md` |
| Tests | `testing.md` · `architecture.md` |
| Logging or observability | `observability.md` · `security.md` |
| Git workflow | `workflow.md` |
| Scaffold new domain (command) | `commands/new-domain.md` |
| Add migration (command) | `commands/add-migration.md` |
