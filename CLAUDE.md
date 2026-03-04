# Revio API — Claude Context

## Module
`github.com/johlun99/revio` (repo: `github.com/johlun99/revio-api`)

## Running Locally
```bash
cp .env.example .env   # fill POSTGRES_PASSWORD, JWT_SECRET
docker compose up      # postgres + api with Air hot-reload
curl localhost:8080/health
```

## Structure
```
cmd/api/main.go              # entry point: config → db → migrate → router → serve
internal/config/config.go    # env loading + validation
internal/db/conn.go          # pgxpool setup + goose migrations
internal/db/migrations/      # .sql migration files
internal/api/router.go       # chi router + CORS + middleware
internal/api/handler/        # HTTP handlers
```

## Conventions

### Errors
Always wrap with context:
```go
return fmt.Errorf("describe what failed: %w", err)
```

### Logging
Use `slog` (stdlib only, no third-party loggers):
```go
slog.Info("message", "key", value)
slog.Error("failed to do X", "error", err)
```

### HTTP Handlers
- Handlers are structs with dependencies injected via constructor (`NewXHandler(pool)`)
- Always set `Content-Type: application/json` before writing
- Return errors as `{"error": "message"}` JSON with appropriate HTTP status

### Database
- Use `pgxpool.Pool` passed from `main.go`
- All queries go in `internal/repository/` (sqlc-generated in Phase 2)
- Use `pgx/v5` native types where possible (e.g. `pgtype.UUID`)

### Migrations
- Plain SQL files in `internal/db/migrations/`
- Named `NNN_description.sql` (e.g. `002_add_tenants_index.sql`)
- Always include both `-- +goose Up` and `-- +goose Down` sections

### Commits
- Conventional Commits: `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`
- Small, focused commits — one logical change per commit
- Always show proposed message and get approval before committing
- Never push — user pushes manually
- No co-author lines

## Key Decisions
- Chi v5 for routing (stdlib-compatible, no context lock-in)
- pgx/v5 + pgxpool for PostgreSQL (binary protocol, pooling)
- Goose v3 for migrations (programmatic, plain SQL)
- sqlc for query layer (Phase 2 — not yet wired)
- scratch runtime image (CGO_ENABLED=0, static binary)
- slog for logging (stdlib, JSON in prod / text in dev)
