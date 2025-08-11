## Transfer System

A simple money transfer HTTP API built with Go, Gin, and PostgreSQL. It supports:

- Create account
- Get account balance
- Transfer money between accounts with transactional safety

After starting the API, you can view the interactive docs at [http://localhost:9000/docs](http://localhost:9000/docs).

### Project layout

- `cmd/transfer-system/main.go`: application entrypoint
- `internal/api`: Gin router, handlers, middleware
- `internal/config`: env config loader
- `internal/repository`: PostgreSQL access and migrations
- `internal/service`: domain logic
- `internal/domain`: domain models
- `migrations`: SQL migrations (applied automatically on startup)
- `docker-compose.yml`: local Postgres and PgAdmin

### Requirements

- Docker and Docker Compose

Optional for local dev without Docker:
- Go (1.22+) and a running PostgreSQL instance

### One-command start/stop scripts

For convenience, you can use the helper scripts:

```bash
# Build the image and start the full stack (api + postgres + pgadmin) via docker compose
./start.sh

# Stop the app container and bring down the compose stack
./stop.sh
```

- `APP_SERVICES`: which compose services to build/start (default: `"api pgadmin"`)
- `LOG_SERVICE`: which service logs to follow (default: `api`)


### Configuration

The app reads environment variables (a `.env` file is optional). Required variables:

- `PORT`: HTTP port the API listens on (default in container: `8080`)
- `DATABASE_URL`: PostgreSQL connection URL, e.g. `postgres://postgres:password@localhost:5432/transfer-system?sslmode=disable`
- `ENV`: `development` or `production` (default: `production`)


### Run the app and dependencies with Docker Compose

This repo includes a `docker-compose.yml` that starts the API, Postgres and PgAdmin:

```bash
docker compose up -d --build api pgadmin
```

This will:
- build the application image from `Dockerfile`
- start the API on `http://localhost:9000`
- expose Postgres on `localhost:5432`
- expose PgAdmin on `http://localhost:8080`

Connection string used by the API in compose (service-to-service via network):

```
DATABASE_URL=postgres://postgres:tareq123@postgres:5432/transfer-system?sslmode=disable
```

### Build and run the app manually with Docker (optional)

1) Build the image:

```bash
docker build -t transfer-system:latest .
```

2) Run the container (ensure Postgres is up first):

```bash
docker run --rm -p 8080:8080 \
  -e PORT=8080 \
  -e ENV=production \
  -e DATABASE_URL="postgres://postgres:tareq123@host.docker.internal:5432/transfer-system?sslmode=disable" \
  --name transfer-system \
  transfer-system:latest
```

### Run the app container on the same network as docker-compose

If you prefer the app to connect to the `postgres` service by name:

```bash
# Ensure compose is up
docker compose up -d

# Find the compose network name (here it's `transfer-system_local-network` by default)
NETWORK=$(docker network ls --format '{{.Name}}' | grep local-network)

docker run --rm -p 9000:8080 \
  --network "$NETWORK" \
  -e PORT=8080 \
  -e ENV=production \
  -e DATABASE_URL="postgres://postgres:tareq123@postgres:5432/transfer-system?sslmode=disable" \
  --name transfer-system \
  transfer-system:latest
```

### Local development without Docker

1) Ensure Postgres is running locally and create the database `transfer-system`.

2) Set env vars (example for macOS/Linux):

```bash
export PORT=9000
export ENV=development
export DATABASE_URL="postgres://postgres:tareq123@localhost:5432/transfer-system?sslmode=disable"
```

3) Run:

```bash
go run ./cmd/transfer-system
```

Migrations run automatically on startup.

### Run tests

```bash
go test ./... -v
```

### API quickstart

Assuming the app is listening on `http://localhost:9000`.

- Create account

```bash
curl -X POST http://localhost:9000/api/v1/accounts \
  -H 'Content-Type: application/json' \
  -d '{"account_id": 1, "initial_balance": "100.00"}' -i
```

- Get account

```bash
curl http://localhost:9000/api/v1/accounts/1 -i
```

- Transfer money

```bash
curl -X POST http://localhost:9000/api/v1/transactions \
  -H 'Content-Type: application/json' \
  -d '{"source_account_id": 1, "destination_account_id": 2, "amount": "25.50"}' -i
```

### Notes on migrations and paths

The app uses `golang-migrate` with a file source set to `file://../../migrations` from the executing binary. In the container, migrations are copied to `/migrations`, which matches that relative path from `/app`.

### Troubleshooting

- If the server exits with `PORT is not set` or `DATABASE_URL is not set`, ensure these env vars are passed to the container or exported in your shell.
