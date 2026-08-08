#!/usr/bin/env bash
# Self-test for check-test-discipline.sh. Proves the checker reads code and
# ignores Go line comments. Registered in tools/verify as
# `test-discipline-selftest` (tools/verify/registry.go), in profiles fast/pr/full.
set -uo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$here/.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

fail=0
expect() { # expect <label> <expected-count> <actual-count>
  if [[ "$2" != "$3" ]]; then
    echo "SELFTEST FAIL: $1 — expected $2 finding(s), got $3"
    fail=1
  else
    echo "selftest ok: $1"
  fi
}

mkdir -p "$tmp/tests/integration/selftest"
cp "$repo/scripts/testdata/test-discipline/comment_only_fixture.go.txt" \
   "$tmp/tests/integration/selftest/comment_only_integration_test.go"
cp "$repo/scripts/testdata/test-discipline/code_violation_fixture.go.txt" \
   "$tmp/tests/integration/selftest/code_violation_integration_test.go"
# comment_only/code_violation above only exercise branch 1 of the sed (whole-line
# comments). trailing_comment_fixture exercises branch 2: a real code line
# followed by a `//` comment that itself carries an R4 pattern. If branch 2
# regresses (e.g. stops truncating trailing comments), this line's "JOIN
# documents" text survives into the scan and is reported as a violation.
cp "$repo/scripts/testdata/test-discipline/trailing_comment_fixture.go.txt" \
   "$tmp/tests/integration/selftest/trailing_comment_integration_test.go"

# R3 reads DEV_TENANT_UUID from internal/platform/tenant/const.go relative to
# cwd. Stub that path inside the throwaway repo so R3 is exercised too — the
# literal must match the fixtures' hardcoded UUID, not the real repo's value.
mkdir -p "$tmp/internal/platform/tenant"
cat > "$tmp/internal/platform/tenant/const.go" <<'EOF'
package tenant

const DevTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
EOF

cd "$tmp"
# The checker discovers files via `git ls-files`, which requires a repo and
# only lists tracked files — a plain mktemp dir has neither. Make $tmp a
# throwaway repo and track the fixtures so discovery matches real usage.
git init -q
git add -A
out="$(bash "$repo/scripts/check-test-discipline.sh" 2>&1 || true)"

comment_hits="$(grep -c 'comment_only_integration_test.go' <<<"$out" || true)"
code_hits="$(grep -c 'code_violation_integration_test.go' <<<"$out" || true)"
trailing_hits="$(grep -c 'trailing_comment_integration_test.go' <<<"$out" || true)"

expect "comment lines are not violations" 0 "$comment_hits"
expect "code lines are violations" 3 "$code_hits"
expect "trailing comment lines are not violations" 0 "$trailing_hits"

exit "$fail"
