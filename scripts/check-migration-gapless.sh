#!/usr/bin/env bash
set -euo pipefail

# migration-gapless — extracted verbatim from ci.yml:governance's former
# inline "Check gapless sequence + no historical edits" step (Task R2),
# itself copied verbatim from invariants.yml:migration-gapless. Registered as
# tools/verify check "migration-gapless".
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
BASE=${GITHUB_BASE_REF:-main}
MODIFIED=$(git log --diff-filter=M --follow --name-only \
  "origin/$BASE...HEAD" -- "$MIGRATION_DIR/*.sql" 2>/dev/null | grep '\.sql$' || true)

if [ -n "$MODIFIED" ]; then
  echo "❌ Historical migrations modified: $MODIFIED"
  exit 1
fi
echo "✅ No historical migration edits"
