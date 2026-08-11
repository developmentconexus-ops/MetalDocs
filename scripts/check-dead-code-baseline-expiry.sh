#!/usr/bin/env bash
# A2.2 — dead-code ratchet, EXPIRY gate (issue #91).
#
# knip-dead-code (check-dead-code.sh) proves the tree never has MORE dead
# files/unused exports than dead-code-baseline.json; dead-code-baseline-growth
# proves the baseline file itself never grows unreviewed. Neither has a
# forcing function to make anyone ever revisit the baseline once it is
# written — it would otherwise sit forever, same gap A2.1 had before
# check-eslint-suppression-expiry.sh. This script is that forcing function,
# structurally the same fixed-date-not-silent mechanism, adapted for a
# baseline that has no natural per-rule grouping (dead-code-baseline.json is
# one flat ratchet, not N independently-enabled ESLint rules) — so this gate
# is keyed by the whole baseline, not per-entry.
#
# To renew: bump "expires" in dead-code-baseline.expiry.json in a reviewed PR
# that states why (still burning down / deliberately re-scoped).
# To retire once the baseline is empty:
#   bash scripts/check-dead-code.sh --write   # confirms {"files":[],"exports":[]}
# then delete dead-code-baseline.expiry.json once dead-code-baseline.json
# records nothing.
#
# Usage: bash scripts/check-dead-code-baseline-expiry.sh
# Exit 0 = no dead-code-baseline.json (nothing to expire), or it exists with a
#          live (non-expired), well-formed expiry entry.
# Exit 1 = dead-code-baseline.json exists but the expiry file is missing, or
#          expiry is malformed, or expiry has passed.
set -euo pipefail

BASELINE="dead-code-baseline.json"
EXPIRY="dead-code-baseline.expiry.json"

if [[ ! -f "$BASELINE" ]]; then
  echo "dead-code-baseline-expiry: no $BASELINE — nothing baselined, clean."
  exit 0
fi

if ! jq empty "$BASELINE" 2>/dev/null; then
  echo "dead-code-baseline-expiry: FAIL — $BASELINE is not valid JSON. A baseline that cannot be parsed cannot be trusted; this is a failure, not a clean pass."
  exit 1
fi

# An empty baseline (ratchet fully burned down) needs no expiry to track.
BASELINE_EMPTY=$(jq -r 'if ((.files // []) | length) == 0 and ((.exports // []) | length) == 0 then "yes" else "no" end' "$BASELINE" 2>/dev/null || echo "no")
if [[ "$BASELINE_EMPTY" == "yes" ]]; then
  echo "dead-code-baseline-expiry: $BASELINE is empty (files and exports both []) — nothing to expire, clean."
  exit 0
fi

if [[ ! -f "$EXPIRY" ]]; then
  echo "dead-code-baseline-expiry: $BASELINE has entries but $EXPIRY is missing."
  echo "dead-code-baseline-expiry: EXPIRED — a non-empty baseline needs a recorded expiry date."
  exit 1
fi

if ! jq empty "$EXPIRY" 2>/dev/null; then
  echo "dead-code-baseline-expiry: EXPIRED — $EXPIRY is not valid JSON. A malformed expiry file is a failure, never a silent pass."
  exit 1
fi

today=$(date -u +%F)

# Same reasoning as check-eslint-suppression-expiry.sh: fail closed on a
# missing/malformed "expires" value rather than string-comparing garbage.
# The zero-padded YYYY-MM-DD regex plus `date -d` validation together reject
# both non-date shapes and shapes that merely LOOK like dates but sort after
# today lexicographically without being real calendar dates (e.g.
# "9999-99-99", "2026-13-01").
expires=$(jq -r '.expires // empty' "$EXPIRY" | tr -d '\r')
if [[ -z "$expires" ]]; then
  echo "dead-code-baseline-expiry: EXPIRED — $EXPIRY has no \"expires\" field."
  exit 1
fi
if [[ ! "$expires" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]]; then
  echo "dead-code-baseline-expiry: EXPIRED — malformed expiry '$expires' in $EXPIRY (must be exact zero-padded YYYY-MM-DD)."
  exit 1
fi
if ! date -u -d "$expires" +%F >/dev/null 2>&1; then
  echo "dead-code-baseline-expiry: EXPIRED — invalid calendar date '$expires' in $EXPIRY."
  exit 1
fi
if [[ "$expires" < "$today" ]]; then
  echo "dead-code-baseline-expiry: EXPIRED — baseline expired on $expires (today: $today). Burn it down or renew the date with a reviewed edit to $EXPIRY."
  exit 1
fi

echo "dead-code-baseline-expiry: clean — $BASELINE expiry ($expires) not yet reached (today: $today)"
