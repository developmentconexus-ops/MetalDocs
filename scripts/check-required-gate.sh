#!/usr/bin/env bash
# Exercises the jq expression from ci.yml's `required` job against fixtures.
# The gate is one expression; if it is wrong, every check behind it is
# decoration. This runs the real expression, extracted from the workflow, so
# a copy cannot drift from what CI executes.
set -euo pipefail
cd "$(dirname "$0")/.."

EXPR=$(python3 - <<'PY'
import re, yaml
with open('.github/workflows/ci.yml', encoding='utf-8') as fh:
    wf = yaml.safe_load(fh)
run = wf['jobs']['required']['steps'][0]['run']
m = re.search(r"if ! jq -e '(.*?)'\s*<<<", run, re.S)
print(m.group(1))
PY
)

fail=0
for f in scripts/testdata/required-gate/*.json; do
  want=$(basename "$f" .json | cut -d- -f1)   # pass-*.json / fail-*.json
  if jq -e "$EXPR" "$f" >/dev/null 2>&1; then got=pass; else got=fail; fi
  if [ "$got" != "$want" ]; then
    echo "FAIL $(basename "$f"): expected $want, got $got"
    fail=1
  fi
done
[ "$fail" -eq 0 ] && echo "[required-gate] OK"
exit "$fail"
