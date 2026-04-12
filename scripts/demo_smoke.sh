#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

PORT="${PORT:-$((18000 + RANDOM % 1000))}"
DEMO_DB_PATH="${DEMO_DB_PATH:-tmp/demo/smoke-admin.db}"
BASE="http://127.0.0.1:${PORT}"
COOKIE_JAR="$(mktemp)"
LOG_FILE="$(mktemp)"

cleanup() {
  rm -f "$COOKIE_JAR"
  rm -f "$LOG_FILE"
  if [[ -n "${SERVER_PID:-}" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT

npm run build:css >/dev/null

DEMO_DB_PATH="$DEMO_DB_PATH" DEMO_RESET=1 PORT="$PORT" go run ./cmd/demo >"$LOG_FILE" 2>&1 &
SERVER_PID=$!

ready=0
for _ in $(seq 1 30); do
  if curl -fsS "$BASE/admin/login" >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 1
done

if [[ "$ready" != "1" ]]; then
  echo "demo failed to start" >&2
  cat "$LOG_FILE" >&2
  exit 1
fi

curl -fsS -c "$COOKIE_JAR" -b "$COOKIE_JAR" -L \
  -d 'username=admin&password=admin' \
  -X POST "$BASE/admin/login" >/tmp/goadmin-dashboard.html

grep -q 'Quick Access' /tmp/goadmin-dashboard.html
grep -q 'Module Guide' /tmp/goadmin-dashboard.html
grep -q 'Suggested Walkthrough' /tmp/goadmin-dashboard.html
grep -q 'Framework Coverage' /tmp/goadmin-dashboard.html
grep -q 'Demo Helpers' /tmp/goadmin-dashboard.html
grep -q 'Capability Matrix' /tmp/goadmin-dashboard.html
curl -fsS -L -c "$COOKIE_JAR" -b "$COOKIE_JAR" "$BASE/admin/demo-center" | grep -q 'Export JSON'
curl -fsS -L -c "$COOKIE_JAR" -b "$COOKIE_JAR" "$BASE/admin/demo-export.json" | grep -q '"reports"'
curl -fsS -L -c "$COOKIE_JAR" -b "$COOKIE_JAR" "$BASE/admin/articles" | grep -q '2 links'
curl -fsS -L -c "$COOKIE_JAR" -b "$COOKIE_JAR" "$BASE/admin/articles/1" | grep -q 'Why Go?'
curl -fsS -L -c "$COOKIE_JAR" -b "$COOKIE_JAR" "$BASE/admin/projects" | grep -q 'Framework Launch'
curl -fsS -L -c "$COOKIE_JAR" -b "$COOKIE_JAR" "$BASE/admin/projects/1" | grep -q 'Core modules done'
curl -fsS -L -c "$COOKIE_JAR" -b "$COOKIE_JAR" "$BASE/admin/audits" | grep -q 'Audit Logs'
curl -fsS -L -c "$COOKIE_JAR" -b "$COOKIE_JAR" "$BASE/admin/audits/1" | grep -q 'article'
curl -fsS -L -c "$COOKIE_JAR" -b "$COOKIE_JAR" "$BASE/admin/tickets" | grep -q 'Polish dashboard spacing'
curl -fsS -L -c "$COOKIE_JAR" -b "$COOKIE_JAR" "$BASE/admin/tickets/1" | grep -q 'Framework Launch'
curl -fsS -L -c "$COOKIE_JAR" -b "$COOKIE_JAR" "$BASE/admin/reports" | grep -q 'Weekly Signups'
curl -fsS -L -c "$COOKIE_JAR" -b "$COOKIE_JAR" "$BASE/admin/reports/1" | grep -q 'signups'
curl -fsS -L -c "$COOKIE_JAR" -b "$COOKIE_JAR" "$BASE/admin/categories/tree" | grep -q 'Getting Started'
curl -fsS -L -c "$COOKIE_JAR" -b "$COOKIE_JAR" "$BASE/admin/menus/tree" | grep -q 'Menu Tree'

echo "demo-smoke-ok"
