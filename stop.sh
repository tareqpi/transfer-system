#!/usr/bin/env bash
set -euo pipefail

APP_CONTAINER=${APP_CONTAINER:-transfer-system}

echo "[stop.sh] Stopping and removing app container '${APP_CONTAINER}' (if exists) ..."
docker rm -f "${APP_CONTAINER}" >/dev/null 2>&1 || true

echo "[stop.sh] Bringing down docker compose stack ..."
docker compose down

echo "[stop.sh] Done."


