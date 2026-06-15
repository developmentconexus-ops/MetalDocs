#!/usr/bin/env bash
# F4c.4 — Test-discipline grep guard.
# Enforces four harness rules on integration test files.
# See docs/superpowers/milestones/.../f4c.4-ci-grep-guards/spec.md for rule rationale.
#
# Usage: bash scripts/check-test-discipline.sh
# Exit 0 = clean. Exit 1 = violations found (each printed as path:line: <rule>: <text>).
set -euo pipefail

# ---------------------------------------------------------------------------
# Known-debt allowlist — pre-F4c.4 violations tracked for future cleanup.
# This list MUST only shrink over time; additions require operator approval + reason.
# ---------------------------------------------------------------------------
# R3 allowlist (hardcoded DevTenantID literal — literal const value, not the Go name):
R3_ALLOWLIST=(
  "internal/modules/auth/infrastructure/postgres/sessions_admin_integration_test.go"
  "internal/modules/iam/infrastructure/postgres/role_provider_integration_test.go"
  "tests/integration/approval/eligibility_test.go"
)
# R4 allowlist (bare unqualified `documents` table reference):
R4_ALLOWLIST=(
  "internal/modules/iam/integration_test.go"
  "internal/platform/migrate/revision_number_zero_based_integration_test.go"
  "tests/integration/approval/eligibility_test.go"
)

in_allowlist() {
  local file="$1"; shift
  local item
  for item in "$@"; do [[ "$file" == "$item" ]] && return 0; done
  return 1
}

# ---------------------------------------------------------------------------
# Discover scope: integration test files with //go:build integration first line,
# excluding the testdb framework directory (it owns the primitives these rules guard).
# ---------------------------------------------------------------------------
mapfile -t FILES < <(
  git ls-files '*_test.go' \
    | grep -v '^tests/integration/testdb/' \
    | while IFS= read -r f; do
        head -1 "$f" 2>/dev/null | grep -q '^//go:build integration' && echo "$f"
      done
)

# ---------------------------------------------------------------------------
# Extract DevTenantID const value for R3 (reads from canonical source).
# ---------------------------------------------------------------------------
DEV_TENANT_UUID=$(grep -E "DevTenantID\s*=" internal/platform/tenant/const.go 2>/dev/null \
  | grep -oE '[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}' \
  | head -1 || true)

violations=0
report() { echo "$1:$2: $3: $4"; violations=$((violations+1)); }

for f in "${FILES[@]}"; do
  # R1 — no inline metaldocs.asserted_caps set_config literal in test files.
  # Sanctioned path: testdb.SetCapsOnDB / testdb.SeedWithCaps / testdb.SetCapsOnTx.
  while IFS= read -r hit; do
    [[ -z "$hit" ]] && continue
    line=$(echo "$hit" | cut -d: -f1)
    text=$(echo "$hit" | cut -d: -f2-)
    report "$f" "$line" "R1" "$text"
  done < <(grep -nE "set_config\('metaldocs\.asserted_caps'" "$f" 2>/dev/null || true)

  # R2 — no is_local=false in a set_config call.
  # F4c.1 invariant: tripwire writes must be tx-local (is_local=true) to avoid
  # leaking session state across the pool. Use testdb.SetCapsOnDB only when the
  # SUT takes *sql.DB directly and is known safe (MaxOpenConns=1, isolated DB).
  while IFS= read -r hit; do
    [[ -z "$hit" ]] && continue
    line=$(echo "$hit" | cut -d: -f1)
    text=$(echo "$hit" | cut -d: -f2-)
    report "$f" "$line" "R2" "$text"
  done < <(grep -nE "set_config\([^)]*,[[:space:]]*false[[:space:]]*\)" "$f" 2>/dev/null || true)

  # R3 — no hardcoded DevTenantID literal (ffffffff-...).
  # Tests must use factory.NewTenant(...) or tenant.DevTenantID constant, not the raw UUID.
  if [[ -n "${DEV_TENANT_UUID}" ]] && ! in_allowlist "$f" "${R3_ALLOWLIST[@]}"; then
    while IFS= read -r hit; do
      [[ -z "$hit" ]] && continue
      line=$(echo "$hit" | cut -d: -f1)
      text=$(echo "$hit" | cut -d: -f2-)
      report "$f" "$line" "R3" "$text"
    done < <(grep -nF "\"${DEV_TENANT_UUID}\"" "$f" 2>/dev/null || true)
  fi

  # R4 — no bare unqualified `documents` table reference in SQL.
  # M4b legacy-schema teardown: bare `documents` may resolve to dead legacy table.
  # Use testdb.Qualified(schema, "documents") or factory-returned identifiers.
  if ! in_allowlist "$f" "${R4_ALLOWLIST[@]}"; then
    while IFS= read -r hit; do
      [[ -z "$hit" ]] && continue
      line=$(echo "$hit" | cut -d: -f1)
      text=$(echo "$hit" | cut -d: -f2-)
      report "$f" "$line" "R4" "$text"
    done < <(grep -nE "(FROM|JOIN|INTO|UPDATE)[[:space:]]+documents([^_a-zA-Z]|$)" "$f" 2>/dev/null || true)
  fi
done

if (( violations > 0 )); then
  echo "test-discipline: $violations violation(s) found — see output above"
  exit 1
fi
echo "test-discipline: clean (${#FILES[@]} integration test files checked)"
