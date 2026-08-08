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
# TST-11 (2026-07-02): all 3 entries reviewed; none migrated to testdb.Open — see
# per-entry reason below. Kept as justified, not migrate-on-touch debt.
R3_ALLOWLIST=(
  # internal/modules/** importing tests/integration/testdb risks an import cycle:
  # testdb/factory.go imports internal/modules/controlleddocuments/{domain,infrastructure};
  # same constraint documented inline in role_provider_integration_test.go below.
  "internal/modules/auth/infrastructure/postgres/sessions_admin_integration_test.go"
  # Live-DB ambient probe: the dev-tenant literal IS the probe's subject (asserts
  # the pre-seeded admin is active in that exact tenant); factory data can't
  # substitute. (Historical import-cycle claim corrected 2026-07-02: the cycle
  # applies only to controlleddocuments/infrastructure; this package now imports
  # testdb and delegates seedWithCapsIAM to testdb.SeedWithCaps.)
  "internal/modules/iam/infrastructure/postgres/role_provider_integration_test.go"
  # Deliberately reuses pre-seeded dev-tenant reference data (profile_code 'dc')
  # across a hand-rolled FK chain with savepoint rollback; testdb.Open's isolated
  # per-test clone has no such ambient data — migrating is a behavioral redesign,
  # not a mechanical swap.
  "tests/integration/approval/eligibility_test.go"
)
# R4 allowlist (bare unqualified `documents` table reference):
# TST-11 (2026-07-02): all 3 entries reviewed; none migrated to testdb.Qualified —
# see per-entry reason below. Kept as justified, not migrate-on-touch debt.
R4_ALLOWLIST=(
  # Same internal/modules/** import-cycle constraint as the R3 entry above; also
  # depends on ambient fixture rows in the live dev DB ("requires fixture document
  # row" skips), incompatible with testdb's isolated per-test clone model.
  "internal/modules/iam/integration_test.go"
  # R4 false positive: the `documents` here is a `CREATE TEMP TABLE documents`
  # (migration-shift-logic probe), not the real table — TEMP TABLE names cannot be
  # schema-qualified. Also internal/platform/migrate importing tests/ is backwards
  # layering (tests/ depends on internal/, not vice versa).
  "internal/platform/migrate/revision_number_zero_based_integration_test.go"
  # Same ambient-fixture-data rationale as the R3 entry above.
  "tests/integration/approval/eligibility_test.go"
)
# R2 allowlist (is_local=false set_config — session-level GUC required by design):
# Both entries reviewed 2026-07-02; each carries a full in-file justification
# comment. These are NOT tripwire asserted_caps writes (R2's rationale) — they
# set tenant_id / bypass_authz GUCs where tx-local scoping is structurally
# impossible. This list MUST only shrink; additions require operator approval.
R2_ALLOWLIST=(
  # RLS probe: SET ROLE + tenant_id GUC + policy-filtered SELECTs must share ONE
  # pinned *sql.Conn across multiple statements with no wrapping tx — a tx-local
  # GUC would be discarded before the reads the test exists to make.
  # Path reconciled 2026-07-06 (M10 F-R3): F9.5 renamed templates/repository/ →
  # templates/infrastructure/; this is a path correction of an existing entry,
  # NOT a new allowlist widening.
  "internal/modules/templates/infrastructure/tenant_id_rls_integration_test.go"
  # applyMigrationWithBypass: the migration script under test carries its own
  # BEGIN/COMMIT, so a tx-local GUC set from a wrapping Go tx cannot survive
  # into it; bypass is set session-locally on a pinned conn and explicitly
  # cleared afterwards (see in-file doc comment).
  "tests/integration/iam/migration_0170_test.go"
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
  # Rules R1/R3/R4 describe CODE. A Go line comment that merely *describes* SQL
  # is not a violation, and rewording prose to satisfy a grep would be bending
  # the file to fit the rule instead of fixing the rule. Blank out everything
  # from an unquoted `//` to end of line, preserving the line count so reported
  # line numbers still point at the real file.
  scan="$(mktemp)"
  sed -E 's@^[[:space:]]*//.*$@@; s@([^:"'"'"'`])//[^"'"'"'`]*$@\1@' "$f" > "$scan"

  # R1 — no inline metaldocs.asserted_caps set_config literal in test files.
  # Sanctioned path: testdb.SetCapsOnDB / testdb.SeedWithCaps / testdb.SetCapsOnTx.
  while IFS= read -r hit; do
    [[ -z "$hit" ]] && continue
    line=$(echo "$hit" | cut -d: -f1)
    text=$(echo "$hit" | cut -d: -f2-)
    report "$f" "$line" "R1" "$text"
  done < <(grep -nE "set_config\('metaldocs\.asserted_caps'" "$scan" 2>/dev/null || true)

  # R2 — no is_local=false in a set_config call.
  # F4c.1 invariant: tripwire writes must be tx-local (is_local=true) to avoid
  # leaking session state across the pool. Use testdb.SetCapsOnDB only when the
  # SUT takes *sql.DB directly and is known safe (MaxOpenConns=1, isolated DB).
  if ! in_allowlist "$f" "${R2_ALLOWLIST[@]}"; then
    while IFS= read -r hit; do
      [[ -z "$hit" ]] && continue
      line=$(echo "$hit" | cut -d: -f1)
      text=$(echo "$hit" | cut -d: -f2-)
      report "$f" "$line" "R2" "$text"
    done < <(grep -nE "set_config\([^)]*,[[:space:]]*false[[:space:]]*\)" "$f" 2>/dev/null || true)
  fi

  # R3 — no hardcoded DevTenantID literal (ffffffff-...).
  # Tests must use factory.NewTenant(...) or tenant.DevTenantID constant, not the raw UUID.
  if [[ -n "${DEV_TENANT_UUID}" ]] && ! in_allowlist "$f" "${R3_ALLOWLIST[@]}"; then
    while IFS= read -r hit; do
      [[ -z "$hit" ]] && continue
      line=$(echo "$hit" | cut -d: -f1)
      text=$(echo "$hit" | cut -d: -f2-)
      report "$f" "$line" "R3" "$text"
    done < <(grep -nF "\"${DEV_TENANT_UUID}\"" "$scan" 2>/dev/null || true)
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
    done < <(grep -nE "(FROM|JOIN|INTO|UPDATE)[[:space:]]+documents([^_a-zA-Z]|$)" "$scan" 2>/dev/null || true)
  fi

  rm -f "$scan"
done

if (( violations > 0 )); then
  echo "test-discipline: $violations violation(s) found — see output above"
  exit 1
fi
echo "test-discipline: clean (${#FILES[@]} integration test files checked)"
