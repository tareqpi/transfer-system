#!/usr/bin/env bash
set -euo pipefail

# Simple launcher: builds and starts the full stack via docker compose

APP_SERVICES=${APP_SERVICES:-api pgadmin}
LOG_SERVICE=${LOG_SERVICE:-api}

echo "[start.sh] Building and starting services via docker compose (${APP_SERVICES}) ..."
docker compose up -d --build ${APP_SERVICES}

echo "[start.sh] Tailing logs for '${LOG_SERVICE}' (press Ctrl+C to stop) ..."
docker compose logs -f ${LOG_SERVICE} | sed -e "s/^/[${LOG_SERVICE}] /"


