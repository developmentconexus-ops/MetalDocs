#!/usr/bin/env bash
set -euo pipefail

# invariant-coverage-map — extracted verbatim from ci.yml:governance's former
# inline "Check all invariants have ≥1 spec ID" step (Task R2), itself copied
# verbatim from e2e-coverage-gate.yml:coverage-map-check. Registered as
# tools/verify check "invariant-coverage-map".
COVERAGE=frontend/apps/web/e2e/COVERAGE.md
echo "Scanning $COVERAGE for unmapped invariants..."

# Find all invariant rows (lines with | Pn-Ixx | or | Fn | pattern) that
# are marked unmapped (❌).
#
# This used to be a `grep | grep | awk | xargs` pipeline. Under `pipefail`
# (GitHub Actions' default `bash -eo pipefail`), any `grep` stage that
# matches zero lines exits 1, which aborts the whole pipeline via `set -e`
# before the assignment even completes — including in the all-mapped case,
# the one case this check exists to report ✅ for. A single awk pass
# replaces the two chained greps: awk exits 0 whenever it can read the
# file, whether or not any line matches the pattern, so "zero unmapped
# invariants" and "the file is unreadable" are no longer conflated the way
# "grep found nothing" and "the pipeline broke" were.
UNMAPPED=$(awk -F'|' '
  /^\|[[:space:]]*(P[0-9]+-I[0-9]+|F[0-9]+)[[:space:]]*\|/ && /❌/ {
    gsub(/^[ \t]+|[ \t]+$/, "", $2)
    print $2
  }
' "$COVERAGE" | xargs)

if [ -n "$UNMAPPED" ]; then
  echo "❌ Unmapped invariants found: $UNMAPPED"
  echo "Add ≥1 spec ID for each before merging."
  exit 1
fi

echo "✅ All invariants are mapped."
