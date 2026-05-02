# Pfötchen API

`erdbeerpfoetchen` (Pfötchen API) is a Go REST API for storing and retrieving per-viewer JSON collections, keyed by streamer and collection type, backed by PostgreSQL.

## Features

- **RESTful API** — endpoints for reading and writing JSON collections
- **Bearer token auth** — middleware-enforced on every request
- **YAML configuration** — all settings loadable from a config file; CLI flags and env vars override file values
- **Auto-migration** — migrations are embedded in the binary and applied on startup; no external tooling required
- **Docker Compose** — one-command local setup with optional bundled PostgreSQL
- **Graceful shutdown** — handles `SIGINT` / `SIGTERM`

## Quick start (Docker Compose)

```bash
cp .env.example .env             # enable bundled postgres (COMPOSE_PROFILES=local-db)
cp config.yaml.example config.yaml
docker compose up
```

The `db` service starts first, the `migrate` service applies all pending migrations, then `api` starts on port `8080`.

To use an **external** PostgreSQL instance, edit `database.uri` in `config.yaml`, leave `COMPOSE_PROFILES` empty in `.env`, and run `docker compose up`.

## Requirements

- **Go** 1.26.2+
- **PostgreSQL** (bundled via Docker Compose, or external)
- **sqlc** — only needed when modifying `schema/queries.sql`

## Project structure

```
.
├── cmd/api/                    # entry point
├── docs/                       # OpenAPI spec
├── internal/
│   ├── cli/                    # Config struct, ApiCommand, MigrateCommand
│   ├── config/                 # YAML file resolver for kong
│   ├── collection/             # domain service + repository interface
│   ├── database/               # PostgreSQL repositories, migrations
│   │   └── model/              # sqlc-generated types
│   ├── handle/                 # HTTP handlers
│   ├── middleware/             # auth, rate-limit
│   └── user/                   # user domain service + repository interface
├── schema/
│   ├── migrations/             # Goose SQL migration files (embedded in binary)
│   └── queries.sql             # sqlc query definitions
├── script/postgres.sh          # spin up a local Podman PostgreSQL container
├── config.yaml.example         # annotated config template
├── docker-compose.yml
├── .env.example
└── go.mod
```

## Configuration

Settings are resolved in this order (highest priority first):

1. CLI flags
2. Environment variables (`PFOETCHEN_*`)
3. YAML config file (`--config` / `-c`)
4. Built-in defaults

### YAML config file

```bash
cp config.yaml.example config.yaml
# edit config.yaml, then:
./api --config config.yaml api
```

`config.yaml.example`:

```yaml
debug: false

server:
  listen: ":8080"
  unsafe: true   # required when no TLS certificate is configured
  tls:
    cert: ""
    key: ""

database:
  uri: postgres://postgres:postgres@db:5432/postgres?sslmode=disable
```

The config file path can also be set via `PFOETCHEN_CONFIG`.

### All flags and env vars

| Flag | Environment variable | Default | Description |
|---|---|---|---|
| `--debug` | `PFOETCHEN_DEBUG` | `false` | Verbose logging |
| `--server.listen` | `PFOETCHEN_SERVER_LISTEN` | `:8080` | Listen address |
| `--server.unsafe` | `PFOETCHEN_SERVER_UNSAFE` | `false` | Allow HTTP (no TLS) |
| `--server.tls.cert` | `PFOETCHEN_SERVER_TLS_CERT_PATH` | — | TLS certificate path |
| `--server.tls.key` | `PFOETCHEN_SERVER_TLS_KEY_PATH` | — | TLS key path |
| `--database.uri` | `PFOETCHEN_DATABASE_URI` | — | PostgreSQL connection URI |
| `--config` / `-c` | `PFOETCHEN_CONFIG` | — | Path to YAML config file |

## Commands

```bash
# Run the API server (auto-applies any pending migrations on startup)
./api --config config.yaml api

# Apply database migrations and exit without starting the server
./api --config config.yaml migrate

# Build
go build ./cmd/api

# Run tests
go test -race ./...

# Regenerate database code after schema/query changes
sqlc generate

# Start a local PostgreSQL container (Podman)
./script/postgres.sh
```

## API

Full spec in [`docs/openapi.yaml`](docs/openapi.yaml). All endpoints require `Authorization: Bearer <token>`.

> [!NOTE]
> The bearer token is currently hardcoded to `my-super-secret-key` — see [TODO.md](TODO.md).

| Method | Path | Request headers | Success |
|---|---|---|---|
| `POST` | `/api/v1/collection/{streamer_name}` | `X-COLLECTION-KEY`, `X-USER-KEY`, JSON body | `202` |
| `GET` | `/api/v1/collection/{streamer_name}` | `X-COLLECTION-KEY`, `X-USER-KEY` | `200` + JSON |

## Development

### Local database (without Docker Compose)

```bash
./script/postgres.sh   # starts postgres via Podman on :5432
./api --config config.yaml migrate
./api --config config.yaml api
```

### Code generation

After modifying `schema/queries.sql` or adding a migration:

```bash
sqlc generate
```

## License

[The Unlicense](LICENSE)
