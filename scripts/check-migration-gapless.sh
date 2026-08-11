#!/usr/bin/env bash
set -euo pipefail

# migration-gapless — extracted verbatim from ci.yml:governance's former
# inline "Check gapless sequence + no historical edits" step (Task R2),
# itself copied verbatim from invariants.yml:migration-gapless. Registered as
# tools/verify check "migration-gapless".
#
# KNOWN LIMITATION (named, not fixed here — separate slice): the
# historical-edit half of this check only sees commits already in
# `origin/$BASE...HEAD` at the moment it runs. If a developer commits an
# edit to an already-applied migration and then, in that SAME commit or
# working session, runs this check locally BEFORE committing the edit,
# `git log` cannot see a commit that does not exist yet — the check passes.
# The instant the edit IS committed, the same check (run again, or run in
# CI) fails on it. A check a commit can satisfy at run time and violate by
# the act of committing has a blind spot by construction: it can only ever
# catch the violation on a LATER run, never on the run that introduces it.
# This PR narrows *what* counts as a violation (the base-existence
# precondition below); it does not change *when* the check observes one,
# and does not close this gap.
MIGRATION_DIR="db/migrations"

# After the 2026-07-29 fold (migrations 0257-0315 squashed into
# db/baseline, files moved to
# archive/migrations/post-baseline-2026-07-fold/) the directory holds
# only its README. Under `set -euo pipefail` an unmatched `ls` glob
# would abort the step, so the empty case is guarded explicitly: skip
# the gapless sequence check (there is no sequence), but still run the
# historical-edit check below — that one is a git-history question,
# not a filesystem one, and stays meaningful with an empty directory.
shopt -s nullglob
files=("$MIGRATION_DIR"/*.sql)
if [ ${#files[@]} -eq 0 ]; then
  echo "✅ No forward migrations (post-fold baseline) — gapless check skipped"
else
  # Discover min and max from filesystem (never hardcoded). Match only
  # the leading 4-digit prefix of each basename — `grep -oE '[0-9]+'` on
  # the full path also captured digits inside descriptions (e.g. phase8),
  # making MIN collapse to 2.
  nums=$(printf '%s\n' "${files[@]}" \
    | xargs -r -n1 basename | grep -oE '^[0-9]{4}' | sort -n)
  MIN=$(echo "$nums" | head -1)
  MAX=$(echo "$nums" | tail -1)

  echo "Checking migrations $MIN..$MAX for gapless sequence..."
  for n in $(seq "$MIN" "$MAX"); do
    padded=$(printf "%04d" "$n")
    hits=("$MIGRATION_DIR"/${padded}_*.sql)
    if [ ${#hits[@]} -eq 0 ]; then
      echo "❌ Gap: migration $padded missing"
      exit 1
    fi
  done

  echo "✅ Gapless $MIN..$MAX"
fi

# Check no historical migrations were modified after merge
# (only check files that exist in main already). The pathspec is fully
# quoted so GIT expands it — with nullglob an unquoted `*.sql` would
# vanish from the argv and turn this into a repo-wide query.
#
# `git log` and the `grep` that filters its output are split into two
# steps on purpose. `git log` finding no matching commits is a real,
# legitimate outcome (exit 0, empty output) — but `git log` itself can
# also fail (bad ref, shallow clone missing `origin/$BASE`, detached
# HEAD with no such ref), which is exit != 0 with output on stderr. The
# previous form piped both through `2>/dev/null | grep ... || true`,
# which swallowed a `git log` failure and a `grep` "no match" into the
# same "no historical edits" pass — so a broken git command reported the
# same ✅ as a clean history. Checking `git log`'s own exit status here
# means only *that* command's failure fails the check; the downstream
# `grep` finding nothing is still allowed to be empty.
BASE=${GITHUB_BASE_REF:-main}
if ! GIT_LOG_OUTPUT=$(git log --diff-filter=M --follow --name-only \
  "origin/$BASE...HEAD" -- "$MIGRATION_DIR/*.sql"); then
  echo "❌ git log failed while checking for historical migration edits (bad ref origin/$BASE?, shallow clone?)"
  exit 1
fi
MODIFIED_CANDIDATES=$(printf '%s\n' "$GIT_LOG_OUTPUT" | grep '\.sql$' || true)

# Base-existence precondition. The invariant this check protects is "a
# migration that has already been APPLIED must never change" — but
# `--diff-filter=M` above is not that predicate. `git log` over a commit
# RANGE classifies each commit's own diff against ITS OWN PARENT, one
# commit at a time. A file added by an earlier commit on THIS branch and
# edited by a later commit on THIS branch shows up as Modified between
# those two branch-only commits, even though the file has never touched
# the base branch and nothing has ever "applied" it. The comment two lines
# above this block ("only check files that exist in main already") already
# named the intended invariant — it was just never enforced in code.
#
# Confirmed live on PR #113: db/migrations/0318_capability_bindings_schema_
# backfill.sql was ADDED by that branch (`git cat-file -e
# origin/main:db/migrations/0318_capability_bindings_schema_backfill.sql`
# fails — the file does not exist on origin/main) and edited twice
# afterward. The old predicate fired on it anyway, because it never asked
# whether the file existed on the base branch at all — only whether some
# pair of branch commits disagreed about its content.
#
# A migration is subject to the immutability rule ONLY if it exists on the
# base branch tree. The comparison point is the merge base with
# origin/$BASE, not origin/$BASE's current tip — same resolution
# check-eslint-suppression-baseline-growth.sh uses and for the same reason:
# origin/$BASE keeps moving while a PR is open, and the question is "did
# this exist when the branch forked", not "does it exist on main right
# now". Unresolvable fails closed, matching the `git log` failure handling
# above — a comparison this check cannot prove must not silently pass.
if ! MERGE_BASE=$(git merge-base "origin/$BASE" HEAD 2>&1); then
  echo "❌ git merge-base origin/$BASE HEAD failed while checking for historical migration edits (bad ref origin/$BASE?, shallow clone?, unrelated histories?): $MERGE_BASE"
  exit 1
fi

MODIFIED=""
while IFS= read -r f; do
  [ -z "$f" ] && continue
  if git cat-file -e "$MERGE_BASE:$f" 2>/dev/null; then
    MODIFIED="$MODIFIED$f
"
  fi
done <<<"$MODIFIED_CANDIDATES"

if [ -n "$MODIFIED" ]; then
  echo "❌ Historical migrations modified: $MODIFIED"
  exit 1
fi
echo "✅ No historical migration edits"
