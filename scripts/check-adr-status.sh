#!/usr/bin/env bash
#
# check-adr-status.sh — ADR status-field budget gate (F-R2, M10 Dim-10).
#
# Enforces the permanent rule in wiki/standards/documentation-governance.md
# ("ADR status field"): each ADR's `> **Status:**` block — from the Status line
# through any following `>`-continuation lines up to (not including) the next
# `> **Field:**` marker — MUST be <= 3 physical lines AND <= 400 characters.
# This is the "mega-status anti-pattern" guard (finding 105: ADR 0022's status
# field had grown to a 2757-char, 13-phase changelog).
#
# This script is the SINGLE SOURCE of the sweep: both CI (governance-check.yml)
# and the manual "repeatable sweep" in the governance doc invoke it, so the
# convention is enforced by a gate, not by discipline.
#
# Usage:   scripts/check-adr-status.sh [dir]
#   dir    directory of ADR files to scan (default: wiki/decisions)
# Exit:    0 = clean; 1 = one or more offenders (names + counts printed); 2 = usage/dir error.

set -euo pipefail

dir="${1:-wiki/decisions}"

if [ ! -d "$dir" ]; then
  echo "adr-status: directory not found: $dir" >&2
  exit 2
fi

offenders=""
shopt -s nullglob
for f in "$dir"/[0-9]*.md; do
  out=$(awk -v fn="$(basename "$f")" '
    /^> \*\*Status:\*\*/ { inb=1; total=0; lines=0 }
    inb {
      if (!/^>/)                                    { inb=0 }
      else if (lines>0 && /^> \*\*[A-Za-z ]+:\*\*/) { inb=0 }
      else                                          { total+=length($0); lines++ }
    }
    END { if (total>400 || lines>3) print fn": "lines" lines, "total" chars" }
  ' "$f")
  if [ -n "$out" ]; then
    offenders="${offenders}${out}"$'\n'
  fi
done
shopt -u nullglob

if [ -n "$offenders" ]; then
  echo "::error::ADR status-field budget exceeded (>3 physical lines OR >400 chars):" >&2
  printf '%s' "$offenders" >&2
  echo "Fix: move execution history / amendment narrative OUT of the status block into" >&2
  echo "the ADR body (## Status history) or a companion NNNN-execution-history.md." >&2
  echo "See wiki/standards/documentation-governance.md (ADR status field)." >&2
  exit 1
fi

echo "adr-status: clean"
