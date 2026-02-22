# Pfötchen API

`erdbeerpfoetchen` (Pfötchen API) is a Go-based REST API service designed for managing "lamimi" collections. It provides a simple interface to store and retrieve JSON data with authentication and database persistence.

## Features

- **RESTful API**: Endpoints for reading and writing collections.
- **CLI-driven Configuration**: Flexible configuration via command-line arguments and environment variables using `kong`.
- **Database Persistence**: PostgreSQL integration using `sqlc` for type-safe queries.
- **Graceful Shutdown**: Handles OS signals (`SIGINT`, `SIGTERM`) for clean termination.
- **Authentication**: Bearer token middleware.

## Requirements

- **Go**: 1.26 or higher
- **PostgreSQL**: Used for data storage.
- **SQLC**: For generating database code from SQL queries.
- **Goose**: For running database migrations.
- **Podman/Docker**: (Optional) For running the local development database script.

## Project Structure

```text
.
├── cmd/
│   └── api/                # Application entry point (main.go)
├── docs/                   # API documentation (OpenAPI spec)
├── internal/
│   ├── cli/                # CLI and configuration logic (using Kong)
│   ├── collection/         # Domain models and services
│   ├── database/           # Database connection and repository implementations
│   │   └── model/          # Generated sqlc models
│   ├── handle/             # HTTP request handlers
│   └── middleware/         # HTTP middlewares (Auth)
├── schema/                 # Database schema and migrations
│   ├── migrations/         # Goose migrations
│   └── queries.sql         # SQLC queries
├── script/
│   └── postgres.sh         # Script to run local PostgreSQL via Podman
├── .env                    # Environment variables file (ignored by VCS)
├── go.mod                  # Go module definition
├── sqlc.yaml               # SQLC configuration
└── README.md
```

## Setup & Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/ripmav/erdbeerpfoetchen.git
   cd erdbeerpfoetchen
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Database Setup**:
   
   You can start a local PostgreSQL instance using the provided script (requires Podman):
   ```bash
   ./script/postgres.sh
   ```
   
   Then run migrations using `goose`:
   ```bash
   goose -dir schema/migrations postgres "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" up
   ```

## Configuration

The application is configured using command-line flags or environment variables (prefixed with `PFOETCHEN_`).

| Flag | Environment Variable | Default | Description |
|------|----------------------|---------|-------------|
| `--debug` | `PFOETCHEN_DEBUG` | `false` | Enable debug mode and logging |
| `--server.listen` | `PFOETCHEN_SERVER_LISTEN` | `:8080` | Server listen address |
| `--server.unsafe` | `PFOETCHEN_SERVER_UNSAFE` | `false` | Allow insecure connections (HTTP) |
| `--server.tls.cert` | `PFOETCHEN_SERVER_TLS_CERT_PATH` | - | Path to TLS certificate file |
| `--server.tls.key` | `PFOETCHEN_SERVER_TLS_KEY_PATH` | - | Path to TLS key file |
| `--database.uri` | `PFOETCHEN_DATABASE_URI` | - | PostgreSQL connection URI |

## Running the API

To start the API server:

```bash
go run ./cmd/api api --database.uri "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
```

Or using environment variables:

```bash
export PFOETCHEN_DATABASE_URI="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
go run ./cmd/api api
```

## API Endpoints

The full API specification is available in the [OpenAPI spec](docs/openapi.yaml).

All endpoints require an `Authorization: Bearer <token>` header.

> [!NOTE]
> Current valid token is hardcoded to `my-super-secret-key` (See [TODO](TODO.md)).

### 1. Store a Collection
- **URL**: `POST /api/v1/collection/{streamer_name}`
- **Headers**:
  - `X-COLLECTION-KEY`: Unique collection identifier (required)
  - `X-USER-KEY`: Unique user identifier (required)
- **Body**: JSON object
- **Response**: `202 Accepted` on success.

### 2. Retrieve a Collection
- **URL**: `GET /api/v1/collection/{streamer_name}`
- **Headers**:
  - `X-COLLECTION-KEY`: Unique collection identifier (required)
  - `X-USER-KEY`: Unique user identifier (required)
- **Response**: `200 OK` with JSON body.

## Development

### Code Generation
This project uses `sqlc` to generate database code. If you modify `schema/queries.sql` or the schema in `schema/migrations`, regenerate the code:

```bash
sqlc generate
```

### Running Tests
To run available tests:

```bash
go test ./...
```

## Scripts

- `script/postgres.sh`: Starts a Percona PostgreSQL container using Podman for local development.

## License

This project is licensed under [The Unlicense](LICENSE) - see the LICENSE file for details.
