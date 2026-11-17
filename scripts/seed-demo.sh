#!/usr/bin/env bash
# Builds a pre-populated demo database at data/demo/optimanager.db from private CSV exports.
# The output lives under data/ (git-ignored). NEVER commit CSVs or the generated .db (personal data).
#
# Usage: scripts/seed-demo.sh <clientes.csv> <servicos.csv>
set -euo pipefail

if [ $# -ne 2 ]; then
  echo "Usage: $0 <clientes.csv> <servicos.csv>" >&2
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

CLIENTES="$(realpath "$1")"
SERVICOS="$(realpath "$2")"
DEMO_DIR="$ROOT_DIR/data/demo"
PORT=8686

rm -rf "$DEMO_DIR"
mkdir -p "$DEMO_DIR" "$ROOT_DIR/bin"

go build -o "$ROOT_DIR/bin/optimanager-demo" ./cmd/server

DB_PATH="$DEMO_DIR/optimanager.db" APP_ENV=dev PORT=$PORT "$ROOT_DIR/bin/optimanager-demo" >"$DEMO_DIR/server.log" 2>&1 &
SERVER_PID=$!
trap 'kill $SERVER_PID 2>/dev/null || true; wait $SERVER_PID 2>/dev/null || true' EXIT

for _ in $(seq 1 30); do
  curl -fsS "http://localhost:$PORT/api/health" >/dev/null 2>&1 && break
  sleep 0.5
done

python3 "$ROOT_DIR/scripts/seed_demo.py" --base-url "http://localhost:$PORT" --clientes "$CLIENTES" --servicos "$SERVICOS"

kill $SERVER_PID; wait $SERVER_PID 2>/dev/null || true
trap - EXIT
rm -f "$DEMO_DIR/server.log" "$DEMO_DIR"/*-wal "$DEMO_DIR"/*-shm
echo "Demo DB ready: $DEMO_DIR/optimanager.db"
echo "Run with: DB_PATH=$DEMO_DIR/optimanager.db APP_ENV=dev go run ./cmd/server"
