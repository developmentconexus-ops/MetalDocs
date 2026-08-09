#!/usr/bin/env bash
# Exercises the jq expression from ci.yml's `required` job against fixtures.
# The gate is one expression; if it is wrong, every check behind it is
# decoration. The expression lives in exactly one file, scripts/required-gate.jq,
# and both this selftest and ci.yml's `required` job read it directly — a
# single source, not an extraction, so a copy cannot drift from what CI
# executes.
set -euo pipefail
cd "$(dirname "$0")/.."

fail=0
for f in scripts/testdata/required-gate/*.json; do
  want=$(basename "$f" .json | cut -d- -f1)   # pass-*.json / fail-*.json
  if jq -e -f scripts/required-gate.jq "$f" >/dev/null 2>&1; then got=pass; else got=fail; fi
  if [ "$got" != "$want" ]; then
    echo "FAIL $(basename "$f"): expected $want, got $got"
    fail=1
  fi
done
[ "$fail" -eq 0 ] && echo "[required-gate] OK"
exit "$fail"
