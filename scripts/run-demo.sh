#!/usr/bin/env bash
# Runs OptiManager against the demo database (data/demo/optimanager.db) on port 8686,
# so it does not clash with the Docker container on 8585.
# Generate the demo DB first with: scripts/seed-demo.sh <clientes.csv> <servicos.csv>
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

DB="$ROOT_DIR/data/demo/optimanager.db"
PORT="${PORT:-8686}"

if [ ! -f "$DB" ]; then
  echo "Demo DB not found at $DB. Run scripts/seed-demo.sh first." >&2
  exit 1
fi

if (echo > "/dev/tcp/127.0.0.1/$PORT") >/dev/null 2>&1; then
  echo "Port $PORT is already in use. Run with another port: PORT=8787 $0" >&2
  exit 1
fi

echo "OptiManager demo: http://localhost:$PORT  (login: admin@optimanager.local / demo1234)"
DB_PATH="$DB" APP_ENV=dev PORT="$PORT" exec go run ./cmd/server
