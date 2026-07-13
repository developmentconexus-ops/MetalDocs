#!/usr/bin/env bash
# healthz.sh — API health smoke against the real v1 health surface.
# Routes: /api/v1/health/live, /api/v1/health/ready (see internal/platform/observability/health.go).
# Usage: BASE_URL=http://localhost:80 bash ops/smoke/healthz.sh
set -euo pipefail

BASE_URL="${BASE_URL:?BASE_URL is required (e.g. http://localhost:80) — no default environment exists}"
TIMEOUT="${TIMEOUT:-5}"

fail() { echo "FAIL: $*" >&2; exit 1; }
ok()   { echo "OK: $*"; }

check_endpoint() {
  local url="$1" label="$2"
  local status
  status=$(curl -sf --max-time "$TIMEOUT" -o /tmp/smoke_health.json -w "%{http_code}" "$url" 2>/dev/null || echo "000")
  if [ "$status" = "200" ]; then
    ok "$label ($status)"
  else
    fail "$label returned $status (expected 200)"
  fi
}

echo "Smoke: $BASE_URL"

check_endpoint "$BASE_URL/api/v1/health/live"  "health/live"
check_endpoint "$BASE_URL/api/v1/health/ready" "health/ready"

# Readiness payload sanity: {"status": "...", "checks": [...]}
READY_STATUS=$(python3 -c "import sys,json; d=json.load(open('/tmp/smoke_health.json')); print(d.get('status','unknown'))" 2>/dev/null || echo "unknown")
if [ "$READY_STATUS" = "ready" ]; then
  ok "ready payload status=$READY_STATUS"
else
  fail "ready payload status=$READY_STATUS (expected ready)"
fi

echo "All health checks passed"
