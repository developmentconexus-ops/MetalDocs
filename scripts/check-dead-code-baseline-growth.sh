#!/usr/bin/env bash
# A2.2 — dead-code ratchet, MONOTONICITY gate (issue #91).
#
# check-dead-code.sh proves the working tree never has MORE dead files or
# unused exports than dead-code-baseline.json records. It proves nothing
# about the baseline file itself: nothing stops a PR from hand-editing
# dead-code-baseline.json to add a new dead file or export straight into the
# "accepted" list, at which point check-dead-code.sh reports clean even
# though real new debt just landed unreviewed. This script is that missing
# half, structurally the same fix A2.1 needed for
# eslint-suppressions.json — see check-eslint-suppression-baseline-growth.sh
# for the full precedent this is modeled on, including why the comparison
# point is `git merge-base` computed live and never a second checked-in copy
# of the baseline (a hand-synced enumeration duplicating
# dead-code-baseline.json would be exactly the defect class this repo keeps
# hitting).
#
# "CURRENT" is the COMMITTED content at HEAD (`git show HEAD:...`), never the
# working tree — same reasoning as eslint-suppression-baseline-growth's R3
# fix: an in-run, uncommitted mutation (e.g. a local, uncommitted
# `--write` run) is invisible to a git-sourced read categorically, so the
# race that check closed by switching off the working tree cannot reappear
# here by construction.
#
# Rule: every (dead file) path and every (file, export-name) pair present in
# the COMMITTED dead-code-baseline.json at HEAD must already have been
# present at the merge base with origin/main. A path/pair that disappears
# (shrinks) always passes — burn-down must never be blocked, only growth is
# gated.
#
# Known, accepted gap (same class as eslint-suppression-baseline-growth's):
# renaming a file that carries a baselined entry looks like "old path
# disappeared (passes) + new path appeared (fails)" here, because this
# script has no move detection. A genuine rename needs the waiver below.
#
# Escape hatch for a genuine, reviewed debt increase: the same diff-based
# waiver channel check-governance.ps1 and eslint-suppression-baseline-growth
# use (scripts/check-governance-waivers.txt). Add a line whose rule id is
# "dead-code-baseline-growth"; validity is diff-based (the ADDED line must
# appear in this PR's own diff of that file), exactly like every other rule
# in that file. No env-var override, no flag.
#
# Usage: bash scripts/check-dead-code-baseline-growth.sh
# Exit 0 = no dead file or (file, export) pair grew relative to the merge
#          base with origin/main (or the growth is covered by a waiver line
#          added in this diff).
# Exit 1 = at least one entry grew, unwaived — or the comparison point itself
#          could not be established (shallow clone, unresolvable base ref,
#          unparsable JSON). Every one of those is treated as failure, not
#          silent success: a ratchet that cannot prove monotonicity is not a
#          ratchet.
set -euo pipefail

BASELINE="dead-code-baseline.json"
WAIVERS="scripts/check-governance-waivers.txt"
RULE_ID="dead-code-baseline-growth"
NAME="dead-code-baseline-growth"

# Resolve HEAD explicitly BEFORE asking whether $BASELINE exists there. A
# broken environment (unborn branch, detached-and-broken, corrupt repo) must
# not be indistinguishable from "nothing baselined, clean" — same guard as
# eslint-suppression-baseline-growth's HEAD check.
if ! git rev-parse --verify -q 'HEAD^{commit}' >/dev/null 2>&1; then
  echo "$NAME: could not resolve HEAD to a commit — FAIL-CLOSED (broken or unborn checkout, not a legitimate empty-baseline case)."
  exit 1
fi

if ! git cat-file -e "HEAD:$BASELINE" 2>/dev/null; then
  echo "$NAME: no $BASELINE at HEAD — nothing baselined, clean."
  exit 0
fi

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

# Old state: the baseline as it stood at the merge base. If it did not exist
# there, this PR is the one INTRODUCING the ratchet — nothing for it to have
# grown from, so it passes wholesale rather than being judged against an
# empty baseline (which would flag every entry as "new" on the very PR that
# establishes the mechanism).
if ! git cat-file -e "$MERGE_BASE:$BASELINE" 2>/dev/null; then
  echo "$NAME: $BASELINE did not exist at merge base $MERGE_BASE — first introduction of the ratchet, nothing to have grown from."
  exit 0
fi
OLD_JSON=$(git show "$MERGE_BASE:$BASELINE")

if ! NEW_JSON=$(git show "HEAD:$BASELINE"); then
  echo "$NAME: could not read HEAD:$BASELINE"
  exit 1
fi

# Malformed JSON at either point makes jq exit non-zero; combined with
# `set -euo pipefail` that aborts the script (fail-closed) rather than
# comparing against garbage.
validate_shape() {
  jq -e '
    type == "object"
    and (.files | type) == "array"
    and (.exports | type) == "array"
    and ([.files[] | select(type != "string")] | length) == 0
    and ([.exports[] | select(type != "object" or (.file | type) != "string" or (.name | type) != "string")] | length) == 0
  ' >/dev/null
}
if ! echo "$OLD_JSON" | validate_shape; then
  echo "$NAME: FAIL-CLOSED — $BASELINE at merge base $MERGE_BASE has the wrong shape (malformed input, never a silent pass)."
  exit 1
fi
if ! echo "$NEW_JSON" | validate_shape; then
  echo "$NAME: FAIL-CLOSED — $BASELINE at HEAD has the wrong shape (malformed input, never a silent pass)."
  exit 1
fi

flatten_files() {
  jq -r '.files | sort | .[]' <<<"$1"
}
flatten_exports() {
  jq -r '[.exports[] | "\(.file)\t\(.name)"] | sort | .[]' <<<"$1"
}

OLD_FILES=$(flatten_files "$OLD_JSON")
NEW_FILES=$(flatten_files "$NEW_JSON")
OLD_EXPORTS=$(flatten_exports "$OLD_JSON")
NEW_EXPORTS=$(flatten_exports "$NEW_JSON")

violations=0
report=""

while IFS= read -r f; do
  [[ -z "$f" ]] && continue
  if ! grep -Fxq "$f" <<<"$OLD_FILES"; then
    line="$NAME: GREW — new dead file ($f) absent at merge base $MERGE_BASE"
    echo "$line"
    report+="$line"$'\n'
    violations=$((violations + 1))
  fi
done <<<"$NEW_FILES"

while IFS=$'\t' read -r f n; do
  [[ -z "$f" ]] && continue
  if ! grep -Fxq "$f"$'\t'"$n" <<<"$OLD_EXPORTS"; then
    line="$NAME: GREW — new unused export ($f, $n) absent at merge base $MERGE_BASE"
    echo "$line"
    report+="$line"$'\n'
    violations=$((violations + 1))
  fi
done <<<"$NEW_EXPORTS"

if (( violations == 0 )); then
  echo "$NAME: clean — no dead file or (file, export) pair grew relative to merge base $MERGE_BASE"
  exit 0
fi

waived_lines=$(git diff -w "$MERGE_BASE" HEAD -- "$WAIVERS" 2>/dev/null \
  | grep -E '^\+[^+]' \
  | sed 's/^\+//' \
  | grep -E "^[[:space:]]*${RULE_ID}[[:space:]]*\|" || true)

if [[ -n "$waived_lines" ]]; then
  echo "$NAME: $violations entry(ies) grew, but waived by an entry added in this diff to $WAIVERS:"
  echo "$waived_lines"
  echo "$NAME: PASS (waived)"
  exit 0
fi

echo "$NAME: $violations entry(ies) grew — see GREW lines above."
echo "$NAME: if this growth is deliberate and reviewed, add a justified waiver line (rule id '$RULE_ID') to $WAIVERS. See that file's header for the format."
exit 1
