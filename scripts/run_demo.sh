#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

PORT="${PORT:-8091}"
DEMO_DB_PATH="${DEMO_DB_PATH:-tmp/demo/fresh-admin.db}"
DEMO_RESET="${DEMO_RESET:-1}"

echo "==> building tailwind css"
npm run build:css

echo "==> starting demo"
echo "    PORT=$PORT"
echo "    DEMO_DB_PATH=$DEMO_DB_PATH"
echo "    DEMO_RESET=$DEMO_RESET"

DEMO_DB_PATH="$DEMO_DB_PATH" DEMO_RESET="$DEMO_RESET" PORT="$PORT" go run ./examples/demo/cmd/demo
