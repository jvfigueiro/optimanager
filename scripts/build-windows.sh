#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "[OptiManager] Building frontend..."
npm --prefix ui run build

mkdir -p dist
echo "[OptiManager] Cross-compiling Windows amd64 portable binary..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/optimanager-windows-amd64.exe ./cmd/server

echo "[OptiManager] Done: dist/optimanager-windows-amd64.exe"
