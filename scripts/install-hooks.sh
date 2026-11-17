#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT=$(git rev-parse --show-toplevel)
chmod +x "$REPO_ROOT/scripts/check-forbidden-files.sh" "$REPO_ROOT/.githooks/pre-commit"
git config core.hooksPath .githooks
echo "Git hooks installed successfully (core.hooksPath = .githooks)."
