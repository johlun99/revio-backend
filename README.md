# Revio API

The backend API for Revio — a review management platform for ecommerce stores.

Built with Go (Chi), PostgreSQL (pgx), and Docker.

## Prerequisites

- Docker & Docker Compose
- Go 1.24+ (for local dev without Docker)

## Running with Docker Compose

```bash
cd backend/
cp .env.example .env
# Edit .env — set POSTGRES_PASSWORD and JWT_SECRET (min 32 chars)

docker compose up
```

The API starts on `http://localhost:8080` with Air hot-reload. Changes to `.go` files trigger an automatic rebuild.

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `JWT_SECRET` | Yes | — | Min 32 chars. Used to sign JWT tokens |
| `APP_ENV` | No | `development` | `development` or `production` |
| `APP_PORT` | No | `8080` | HTTP listen port |
| `CORS_ALLOWED_ORIGINS` | No | `http://localhost:5173` | Comma-separated list of allowed origins |
| `POSTGRES_USER` | No | `revio` | DB username (used by Compose) |
| `POSTGRES_PASSWORD` | Yes | — | DB password (used by Compose) |
| `POSTGRES_DB` | No | `revio_dev` | DB name (used by Compose) |

## Migrations

Migrations run automatically on startup via Goose. SQL files live in `internal/db/migrations/`.

To run manually:
```bash
docker compose exec api go run ./cmd/api  # triggers migration on boot
```

Or connect directly and inspect:
```bash
docker compose exec db psql -U revio -d revio_dev -c "\dt"
```

## Health Check

```bash
curl http://localhost:8080/health
# {"status":"ok","timestamp":"...","database":"ok"}
```
