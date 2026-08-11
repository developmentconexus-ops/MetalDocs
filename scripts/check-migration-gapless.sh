#!/usr/bin/env bash
set -euo pipefail

# migration-gapless — extracted verbatim from ci.yml:governance's former
# inline "Check gapless sequence + no historical edits" step (Task R2),
# itself copied verbatim from invariants.yml:migration-gapless. Registered as
# tools/verify check "migration-gapless".
#
# KNOWN LIMITATION (named, not fixed here — separate slice): the
# historical-edit half of this check only sees committed changes in the
# merge-base-to-HEAD comparison at the moment it runs. If a developer commits an
# edit to an already-applied migration and then, in that SAME commit or
# working session, runs this check locally BEFORE committing the edit,
# `git diff` cannot see a commit that does not exist yet — the check passes.
# The instant the edit IS committed, the same check (run again, or run in
# CI) fails on it. A check a commit can satisfy at run time and violate by
# the act of committing has a blind spot by construction: it can only ever
# catch the violation on a LATER run, never on the run that introduces it.
# This PR narrows *what* counts as a violation (the base-existence
# precondition below); it does not change *when* the check observes one,
# and does not close this gap.
MIGRATION_DIR="db/migrations"

# Accumulate-then-exit, not exit-on-first-failure: this script has TWO
# independent properties (gapless sequence, no historical edit), and a
# script that stops at its first failure can only ever demonstrate ONE of
# them in a given run — a real defect, found in review, because it makes
# the two properties mutually unprovable: no single fixture run can assert
# both `Want`s, so one of them inevitably has no negative-fixture coverage
# at all. Same pattern check-dockerfile-go-version.sh already uses on this
# repo (`fail=1` set per-violation inside its loop, one `exit 1` at the
# bottom) — established precedent, not a novel structure. Every violation
# this script finds is still printed with its own diagnostic; only WHEN it
# exits changed, not what it reports or how loudly.
fail=0

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
  gap_found=0
  for n in $(seq "$MIN" "$MAX"); do
    padded=$(printf "%04d" "$n")
    hits=("$MIGRATION_DIR"/${padded}_*.sql)
    if [ ${#hits[@]} -eq 0 ]; then
      echo "❌ Gap: migration $padded missing"
      gap_found=1
      fail=1
    fi
  done

  if [ "$gap_found" -eq 0 ]; then
    echo "✅ Gapless $MIN..$MAX"
  fi
fi

# Check no historical migrations were modified after merge
# (only check files that exist in the merge-base tree). The pathspec is fully
# quoted so GIT expands it — with nullglob an unquoted `*.sql` would vanish
# from the argv and turn this into a repo-wide query.
#
# This guard asserts the net effect of this branch, not the shape of its
# commit history. A `git log --diff-filter=M` range can report a migration
# added by an earlier branch commit and edited by a later branch commit even
# though it never existed on the base and was never applied. The merge-base
# tree is the precondition, and `git diff --diff-filter=M` naturally applies
# it: a branch-added file is net Added, while a base file changed by this
# branch is net Modified.
#
# Confirmed live on PR #113: db/migrations/0318_capability_bindings_schema_
# backfill.sql was ADDED by that branch (`git cat-file -e
# origin/main:db/migrations/0318_capability_bindings_schema_backfill.sql`
# fails — the file does not exist on origin/main) and edited twice afterward.
# The old history predicate fired on it because it never asked whether the
# file existed on the base branch at all.
#
# MERGE_BASE is resolved BEFORE the diff below and used as the LOWER end of
# an asymmetric `$MERGE_BASE..HEAD` range, not `origin/$BASE...HEAD`'s
# symmetric one. The comparison point is the merge base with origin/$BASE,
# not origin/$BASE's current tip — origin/$BASE keeps moving while a PR is
# open, and the question is "did this exist when the branch forked", not
# "does it exist on main right now". Unresolvable fails closed — a comparison
# this check cannot prove must not silently pass.
#
# Second false-positive shape, found in review (CodeRabbit, PR #123): a
# SYMMETRIC range (`origin/$BASE...HEAD`, three dots) walks commits
# reachable from EITHER side but not both — that includes commits that
# landed on origin/$BASE itself after this branch forked. If main advances
# with its own commit editing a migration that already existed at the
# merge base, that commit shows up as a Modified candidate too, and the
# base-existence check above does not save it: the file genuinely DOES exist
# at the merge base, so it passes that filter and gets reported as a violation
# on a branch that never touched it — the identical defect in the identical
# line, just the other lineage. The asymmetric range below walks only commits
# reachable from HEAD and not from the merge base, so a post-fork main commit
# never enters the candidate set.
BASE=${GITHUB_BASE_REF:-main}
if ! MERGE_BASE=$(git merge-base "origin/$BASE" HEAD 2>&1); then
  echo "❌ git merge-base origin/$BASE HEAD failed while checking for historical migration edits (bad ref origin/$BASE?, shallow clone?, unrelated histories?): $MERGE_BASE"
  exit 1
fi

# An empty diff is a legitimate clean outcome, but `git diff` can also fail
# (bad ref, shallow clone missing `origin/$BASE`, detached HEAD with no such
# ref). Capture its status explicitly so a broken comparison cannot silently
# pass as "no historical migration edits".
if ! MODIFIED=$(git diff --name-only --diff-filter=M \
  "$MERGE_BASE..HEAD" -- "$MIGRATION_DIR/*.sql" 2>&1); then
  echo "❌ git diff failed while checking for historical migration edits (bad ref origin/$BASE?, shallow clone?)"
  exit 1
fi
if [ -n "$MODIFIED" ]; then
  echo "❌ Historical migrations modified: $MODIFIED"
  fail=1
else
  echo "✅ No historical migration edits"
fi

if [ "$fail" -eq 1 ]; then
  exit 1
fi
