#!/usr/bin/env bash
# Verifies that no forbidden database, legacy xBase, archive, or private data files
# are tracked or staged in the Git repository.
set -euo pipefail

MODE="${1:-all}"

if [ "$MODE" = "--staged" ]; then
    FILES=$(git diff --cached --name-only --diff-filter=ACMR || true)
else
    FILES=$(git ls-files || true)
fi

if [ -z "$FILES" ]; then
    exit 0
fi

FORBIDDEN_FOUND=0

while IFS= read -r file; do
    [ -z "$file" ] && continue
    lower=$(printf '%s' "$file" | tr '[:upper:]' '[:lower:]')

    case "$lower" in
        *.dbf|*.dbt|*.fpt|*.cdx|*.ntx|*.mdx|*.idx|*.bak|*.db|*.db-shm|*.db-wal|*.sqlite|*.sqlite3|*.csv|*.zip|.env)
            echo "[SECURITY ERROR] Forbidden file extension detected in git: $file" >&2
            FORBIDDEN_FOUND=1
            ;;
        data/*|legacy-data/*|testdata-private/*)
            case "$lower" in
                data/.gitkeep) ;;
                *)
                    echo "[SECURITY ERROR] File inside forbidden data directory detected in git: $file" >&2
                    FORBIDDEN_FOUND=1
                    ;;
            esac
            ;;
    esac
done <<< "$FILES"

if [ "$FORBIDDEN_FOUND" -ne 0 ]; then
    echo "[SECURITY ERROR] Commit/check aborted to prevent leaking customer or database files." >&2
    exit 1
fi

exit 0
