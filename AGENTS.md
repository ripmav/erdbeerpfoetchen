# AGENTS.md

This file provides guidance to Claude Code (claude.ai/code) and all other AI Coding Agents when working with code in this repository.

## Commands

```bash
# Build
go build ./cmd/api

# Run (config file recommended)
go run ./cmd/api --config config.yaml api

# Run (flags only)
go run ./cmd/api api --server.unsafe --database.uri "postgres://user:pass@host:5432/db?sslmode=disable"

# Tests
go test -v -race ./...

# Local dev database (Podman-based PostgreSQL container)
./script/postgres.sh

# Database migrations (via binary)
go run ./cmd/api --config config.yaml migrate

# Regenerate database code after schema/query changes
sqlc generate
```

Configuration is via CLI flags, environment variables prefixed with `PFOETCHEN_` (e.g. `--database.uri` → `PFOETCHEN_DATABASE_URI`), or a YAML config file passed with `--config` / `-c`.

## Architecture

**Pfötchen API** stores and retrieves JSON collections keyed by streamer + collection key + viewer, backed by PostgreSQL with bearer token auth.

### Layer stack

```
CLI (kong)  →  HTTP handlers  →  Services  →  Repositories  →  PostgreSQL
```

- **`cmd/api/main.go`** — entry point; Kong CLI wiring, graceful shutdown on SIGINT/SIGTERM
- **`internal/cli/`** — config struct (`Server`, `Database`, debug flag) and HTTP server lifecycle with optional TLS
- **`internal/handle/collection.go`** — handlers for `POST /api/v1/collection/{streamer_name}` and `GET /api/v1/collection/{streamer_name}`; reads `X-USER-KEY`, `X-COLLECTION-KEY`, and `Authorization: Bearer` headers
- **`internal/middleware/auth.go`** — bearer token auth; validates the Bearer UUID against the stored `api_token` in the DB for the given streamer
- **`internal/middleware/ratelimit.go`** — per-streamer token bucket rate limiter; fetches `rate_limit_per_minute` from the DB via `RateLimitProvider`; falls back to 100 req/min on error; returns 429 when exceeded
- **`internal/collection/`** and **`internal/user/`** — domain services behind Repository interfaces
- **`internal/database/`** — concrete repository implementations; custom `DB` wrapper with `Update()` (write transactions) and `Read()` (read transactions); all queries are **sqlc**-generated from `schema/queries.sql` into `internal/database/model/`
- **`schema/migrations/`** — Goose migration files (embedded in binary, applied with `WithAllowMissing`); schema lives in the `streaming` PostgreSQL schema

### Key patterns

- Services depend on Repository interfaces, not concrete DB types — keeps domain logic testable without a real database.
- All SQL is written in `schema/queries.sql` and must be regenerated via `sqlc generate` after any change.
- Collections are stored as PostgreSQL `JSONB` and handled as `json.RawMessage` in Go — callers pass arbitrary JSON blobs.
- `log/slog` is used throughout; `--debug` enables verbose output.
- The rate limiter reads `rate_limit_per_minute` from `streaming.user` on every request via the `RateLimitProvider` interface. The `user.Service` satisfies this interface and is wired in `internal/cli/api.go`.

### API

Full spec in `docs/openapi.yaml`. Both endpoints require `Authorization: Bearer <token>`:

| Method | Path | Request headers | Success |
|--------|------|-----------------|---------|
| `POST` | `/api/v1/collection/{streamer_name}` | `X-COLLECTION-KEY`, `X-USER-KEY`, JSON body | 202 |
| `GET`  | `/api/v1/collection/{streamer_name}` | `X-COLLECTION-KEY`, `X-USER-KEY` | 200 + JSON |
