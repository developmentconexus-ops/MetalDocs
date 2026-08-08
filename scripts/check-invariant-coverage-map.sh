#!/usr/bin/env bash
set -euo pipefail

# invariant-coverage-map — extracted verbatim from ci.yml:governance's former
# inline "Check all invariants have ≥1 spec ID" step (Task R2), itself copied
# verbatim from e2e-coverage-gate.yml:coverage-map-check. Registered as
# tools/verify check "invariant-coverage-map".
COVERAGE=frontend/apps/web/e2e/COVERAGE.md
echo "Scanning $COVERAGE for unmapped invariants..."

# Find all invariant rows (lines with | Pn-Ixx | or | Fn | pattern)
UNMAPPED=$(grep -E '^\|[[:space:]]*(P[0-9]+-I[0-9]+|F[0-9]+)[[:space:]]*\|' "$COVERAGE" \
  | grep '❌' \
  | awk -F'|' '{print $2}' | xargs)

if [ -n "$UNMAPPED" ]; then
  echo "❌ Unmapped invariants found: $UNMAPPED"
  echo "Add ≥1 spec ID for each before merging."
  exit 1
fi

echo "✅ All invariants are mapped."
