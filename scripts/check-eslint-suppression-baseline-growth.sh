#!/usr/bin/env bash
# A2.1 review round 2 (R1) — ratchet MONOTONICITY gate (issue #91).
#
# eslint-suppression-expiry (check-eslint-suppression-expiry.sh) proves every
# baselined rule eventually gets revisited. It proves nothing about direction:
# `eslint . --suppressions-location eslint-suppressions.json --suppress-all`
# silently absorbs a brand-new violation into eslint-suppressions.json, and a
# subsequent plain `pnpm run lint` then passes clean — proven live in the cold
# review of this PR. That is a baseline, not a ratchet. This script is the
# missing half: it fails when eslint-suppressions.json grows relative to the
# merge base with origin/main.
#
# Comparison point is computed live with `git merge-base`, never a second
# checked-in copy of the baseline — a hand-synced enumeration duplicating
# eslint-suppressions.json is exactly the defect class this repo keeps
# hitting (see the rule-list comment in check-eslint-suppression-expiry.sh).
#
# Rule: for every (file, rule) pair present in the CURRENT eslint-suppressions.json,
# its count must be <= the count that same pair had at the merge base (0 if the
# pair did not exist there). A pair that shrinks, or disappears entirely, always
# passes — burn-down must never be blocked.
#
# Two known, accepted gaps (not bugs — see eslint.config.mjs's R2 comment
# block for the full write-up):
#   - Renaming a file that carries suppressions looks like "old pair
#     disappeared (passes) + new pair appeared (fails)" here, because this
#     script has no move detection. A genuine rename needs the waiver below.
#   - ESLint's suppression key is (file, rule), not individual finding
#     identity, so swapping one baselined finding for a different one in the
#     same file at the same count is invisible to both this script and to
#     ESLint's own Suppressions feature — that is inherent to ESLint 10, not
#     something this script could detect from the JSON alone.
#
# Escape hatch for a genuine, reviewed debt increase: the same diff-based
# waiver channel check-governance.ps1 uses (scripts/check-governance-waivers.txt).
# Add a line whose rule id is "eslint-suppression-baseline-growth"; validity is
# diff-based (the ADDED line must appear in this PR's own diff of that file),
# exactly like every other rule in that file. No env-var override, no flag.
#
# Usage: bash scripts/check-eslint-suppression-baseline-growth.sh
# Exit 0 = no (file, rule) pair grew relative to the merge base (or the growth
#          is covered by a waiver line added in this diff).
# Exit 1 = at least one pair grew, unwaived — or the comparison point itself
#          could not be established (shallow clone, unresolvable base ref,
#          unparsable JSON). Every one of those is treated as failure, not
#          silent success: a ratchet that cannot prove monotonicity is not
#          a ratchet.
set -euo pipefail

SUPPRESSIONS="eslint-suppressions.json"
WAIVERS="scripts/check-governance-waivers.txt"
RULE_ID="eslint-suppression-baseline-growth"
NAME="eslint-suppression-baseline-growth"

if [[ ! -f "$SUPPRESSIONS" ]]; then
  echo "$NAME: no $SUPPRESSIONS — nothing baselined, clean."
  exit 0
fi

# Same base-ref resolution as check-governance.ps1 / check-migration-gapless.sh:
# GITHUB_BASE_REF is what Actions sets on pull_request events; "main" is the
# laptop/push default, matching origin/main.
BASE=${GITHUB_BASE_REF:-main}

if ! git rev-parse --verify -q "origin/$BASE" >/dev/null 2>&1; then
  echo "$NAME: origin/$BASE does not resolve locally (no remote-tracking ref)."
  echo "$NAME: FAIL-CLOSED — this check cannot prove the baseline did not grow without a comparison point, and treating an unknown comparison as a pass would silently defeat the ratchet. (On CI this ref is populated by the 'fetch base ref' prereq step; on a laptop, run \`git fetch origin $BASE\` first.)"
  exit 1
fi

if ! MERGE_BASE=$(git merge-base "origin/$BASE" HEAD 2>&1); then
  echo "$NAME: git merge-base origin/$BASE HEAD failed (shallow clone, or unrelated histories): $MERGE_BASE"
  echo "$NAME: FAIL-CLOSED — see above."
  exit 1
fi

# Old state: the file as it stood at the merge base. If it did not exist
# there at all, this PR is the one INTRODUCING the ratchet — there is no
# existing baseline for it to have grown, so it passes wholesale rather than
# being judged pair-by-pair against an empty object (which would otherwise
# flag every baselined finding as "new debt" on the very PR that establishes
# the mechanism). This is different from a new (file, rule) PAIR appearing
# inside an already-existing eslint-suppressions.json, which is real growth
# and is still caught below.
if ! git cat-file -e "$MERGE_BASE:$SUPPRESSIONS" 2>/dev/null; then
  echo "$NAME: $SUPPRESSIONS did not exist at merge base $MERGE_BASE — first introduction of the ratchet, nothing to have grown from."
  exit 0
fi
OLD_JSON=$(git show "$MERGE_BASE:$SUPPRESSIONS")

# New state: the working tree's current content, not a second `git show
# HEAD:...` read. In CI (a clean post-checkout tree) the two are identical;
# reading the working tree also means a local, uncommitted `--suppress-all`
# run is caught before it is even committed.
if ! NEW_JSON=$(cat "$SUPPRESSIONS"); then
  echo "$NAME: could not read $SUPPRESSIONS"
  exit 1
fi

# Flatten a suppressions JSON blob to sorted "file\trule\tcount" lines.
# Malformed JSON makes jq exit non-zero; combined with `set -euo pipefail`
# that aborts the script (fail-closed) rather than comparing against garbage.
flatten() {
  jq -r '
    to_entries[]? as $f
    | ($f.value // {}) | to_entries[]
    | "\($f.key)\t\(.key)\t\(.value.count // 0)"
  ' <<<"$1" | tr -d '\r' | sort
}

OLD_FLAT=$(flatten "$OLD_JSON")
NEW_FLAT=$(flatten "$NEW_JSON")

violations=0
report=""
while IFS=$'\t' read -r file rule count; do
  [[ -z "$file" ]] && continue
  old_count=$(awk -F'\t' -v f="$file" -v r="$rule" '$1==f && $2==r {print $3; exit}' <<<"$OLD_FLAT")
  old_count=${old_count:-0}
  if (( count > old_count )); then
    if [[ "$old_count" == "0" ]]; then
      line="$NAME: GREW — new suppressed pair ($file, $rule) count=$count (absent at merge base $MERGE_BASE)"
    else
      line="$NAME: GREW — ($file, $rule) count $old_count -> $count (merge base $MERGE_BASE)"
    fi
    echo "$line"
    report+="$line"$'\n'
    violations=$((violations + 1))
  fi
done <<<"$NEW_FLAT"

if (( violations == 0 )); then
  echo "$NAME: clean — no (file, rule) pair grew relative to merge base $MERGE_BASE"
  exit 0
fi

# Diff-based waiver, same channel and same semantics as check-governance.ps1's
# Test-Waived: valid ONLY for the line ADDED in this diff (never a merged
# entry), so the reviewer sees the claim in the PR diff and can contest it.
waived_lines=$(git diff -w "$MERGE_BASE" -- "$WAIVERS" 2>/dev/null \
  | grep -E '^\+[^+]' \
  | sed 's/^\+//' \
  | grep -E "^[[:space:]]*${RULE_ID}[[:space:]]*\|" || true)

if [[ -n "$waived_lines" ]]; then
  echo "$NAME: $violations pair(s) grew, but waived by an entry added in this diff to $WAIVERS:"
  echo "$waived_lines"
  echo "$NAME: PASS (waived)"
  exit 0
fi

echo "$NAME: $violations pair(s) grew — see GREW lines above."
echo "$NAME: if this growth is deliberate and reviewed, add a justified waiver line (rule id '$RULE_ID') to $WAIVERS. See that file's header for the format."
exit 1
