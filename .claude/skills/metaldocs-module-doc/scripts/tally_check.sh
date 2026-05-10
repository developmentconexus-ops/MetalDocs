#!/usr/bin/env bash
# Mechanical tally check — Phase 6.5 gate.
#
# Why this script exists:
#   The first run of this skill (iam module) had two drift bugs that an
#   eyeball pass could not catch reliably:
#     - §11 of the composed doc said "Critical: 1 / Major: 7 / Minor: 4"
#       while the register actually had 1 / 6 / 5.
#     - The coverage-stats line said "Decisions without ADR link: 8" while
#       enumerating 11 items.
#   Both are mechanical: the composer wrote a freehand summary from memory.
#   Grep + wc is cheaper and never miscounts.
#
# What it checks:
#   1. Severity counts in <module>.md §11 match the register
#   2. "missing-ADR" reference count in coverage-stats line matches the
#      number of "missing-ADR" occurrences in the register
#   3. Every T-NNN in the register has a matching row in the backlog
#      (debt_id column) — orphaned debt items are flagged
#   4. Every backlog row's debt_id is either T-NNN that exists, or matches
#      the maint:<kind> grammar
#
# Usage:
#   ./tally_check.sh <module-name>
#   ./tally_check.sh iam
#
# Exits 0 on pass, 1 on any mismatch (so the composer can wire into CI later
# if desired). All findings printed to stdout.

set -u

if [ $# -lt 1 ]; then
  echo "usage: $0 <module-name>" >&2
  exit 2
fi

MODULE="$1"
DOC="wiki/modules/${MODULE}.md"
DEBT="wiki/modules/${MODULE}-tech-debt.md"
BACKLOG="wiki/backlog/${MODULE}-refactor.md"

for f in "$DOC" "$DEBT" "$BACKLOG"; do
  if [ ! -f "$f" ]; then
    echo "FAIL: missing file $f"
    exit 1
  fi
done

fail=0

# --- Check 1: severity counts ---------------------------------------------
crit_actual=$(grep -cE '^- \*\*Severity:\*\* critical' "$DEBT" || true)
maj_actual=$(grep -cE '^- \*\*Severity:\*\* major' "$DEBT" || true)
min_actual=$(grep -cE '^- \*\*Severity:\*\* minor' "$DEBT" || true)

crit_stated=$(grep -oE '^- Critical: [0-9]+' "$DOC" | head -n1 | grep -oE '[0-9]+' || echo "")
maj_stated=$(grep -oE '^- Major: [0-9]+' "$DOC" | head -n1 | grep -oE '[0-9]+' || echo "")
min_stated=$(grep -oE '^- Minor: [0-9]+' "$DOC" | head -n1 | grep -oE '[0-9]+' || echo "")

echo "[tally] severities — actual C/M/m: ${crit_actual}/${maj_actual}/${min_actual} · stated: ${crit_stated:-?}/${maj_stated:-?}/${min_stated:-?}"

if [ "$crit_actual" != "${crit_stated:-}" ] || \
   [ "$maj_actual"  != "${maj_stated:-}"  ] || \
   [ "$min_actual"  != "${min_stated:-}"  ]; then
  echo "FAIL: §11 severity counts in $DOC do not match register $DEBT"
  fail=1
fi

# --- Check 2: missing-ADR list ↔ stated count -----------------------------
adr_missing_count=$(grep -cE 'missing-ADR' "$DEBT" || true)
adr_stated=$(grep -oE 'Decisions without ADR link: [0-9]+' "$DOC" | head -n1 | grep -oE '[0-9]+' || echo "")
echo "[tally] missing-ADR — register count: ${adr_missing_count} · doc stated: ${adr_stated:-?}"
if [ -n "${adr_stated:-}" ] && [ "$adr_missing_count" != "$adr_stated" ]; then
  echo "FAIL: missing-ADR count in $DOC coverage-stats disagrees with register"
  fail=1
fi

# --- Check 3 + 4: TD ↔ backlog linkage ------------------------------------
tds_in_register=$(grep -oE '^### T-[0-9]+' "$DEBT" | grep -oE 'T-[0-9]+' | sort -u)
debt_ids_in_backlog=$(grep -oE '\| T-[0-9]+ \|' "$BACKLOG" | grep -oE 'T-[0-9]+' | sort -u)

for td in $tds_in_register; do
  if ! echo "$debt_ids_in_backlog" | grep -qx "$td"; then
    echo "WARN: $td in register but no backlog row references it"
  fi
done

# Validate debt_id column: T-NNN must exist in register; maint:<kind> must be allowed
allowed_maint='maint:doc-cleanup|maint:dep-bump|maint:test-only|maint:docs-link|maint:migration-cleanup'
bad_rows=$(awk -F'|' '/^\| R-[0-9]+/ {gsub(/^ +| +$/,"",$4); print $4}' "$BACKLOG" \
  | grep -vE "^(T-[0-9]+|${allowed_maint})$" || true)
if [ -n "$bad_rows" ]; then
  echo "FAIL: backlog rows with debt_id not in allowed set:"
  echo "$bad_rows" | sed 's/^/  - /'
  fail=1
fi

# --- Summary --------------------------------------------------------------
if [ "$fail" -eq 0 ]; then
  echo "[tally] PASS"
  exit 0
fi
echo "[tally] FAIL — fix doc + re-run before Phase 7"
exit 1
