# Multi-stage Dockerfile for transfer-system
# - Builder compiles a static Linux binary
# - Runner is a minimal, non-root distroless image

##########
# Builder
##########
FROM golang:1.22 as builder

# Enable automatic toolchain download to satisfy go.mod's newer Go version if needed
ENV CGO_ENABLED=0 GOOS=linux GOTOOLCHAIN=auto

WORKDIR /src

# Cache deps
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of the source
COPY . .

# Build the binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o /out/transfer-system ./cmd/transfer-system

##########
# Runner
##########
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy binary
COPY --from=builder /out/transfer-system /app/transfer-system

# Copy migrations to a path that matches the app's relative lookup (file://../../migrations from /app → /migrations)
COPY --from=builder /src/migrations /migrations

# Copy API docs for serving OpenAPI spec
COPY --from=builder /src/docs /docs

# Set sensible defaults; override as needed at runtime
ENV PORT=9000 \
    ENV=production

EXPOSE 9000

USER nonroot:nonroot

ENTRYPOINT ["/app/transfer-system"]

