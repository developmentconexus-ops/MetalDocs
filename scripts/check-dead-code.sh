#!/usr/bin/env bash
# A2.2 — dead-code ratchet, primary gate (issue #91).
#
# knip (the tool issue #91 names) has no built-in baseline/suppression
# mechanism the way ESLint 10's `--suppressions-location` does (see
# check-eslint-suppression-expiry.sh's header for that mechanism). This
# script is knip's missing baseline half: it runs knip scoped to exactly the
# two issue types A2.2 owns — `files` (dead files) and `exports` (unused
# exports) — and compares the result against the committed
# dead-code-baseline.json. Pre-existing debt captured in that file passes;
# anything NOT in it is new debt and fails the build. Unused *types*,
# duplicate exports, unused dependencies, unresolved imports etc. are
# reported by plain knip but intentionally out of scope for THIS gate (A2.2's
# acceptance is "dead files + unused exports", not the rest of knip's report
# — clone detection and component size are separate slices, A2.3/A2.4).
#
# Scope: frontend/apps/web/design-source/**, scripts/perf/**, and
# tools/perfbench/** are excluded via knip.json's top-level "ignore". They
# are not part of any package's dependency graph — design-source/ is
# reference/mockup material compared against by hand during screen review
# (never imported by the app), and scripts/perf + tools/perfbench are
# standalone k6/node scripts invoked directly, not through an entry knip can
# resolve. Everything else, including frontend/apps/web/src/ (the 29-and-
# growing dead-file count issue #91 cites, and its abandoned
# documents/canvas/ subtree) and every workspace's src/, stays in scope.
#
# Identity: a dead file is identified by its repo-relative path. An unused
# export is identified by (file, export name) — not by line number, so a
# file edited elsewhere that shifts an already-baselined export's line does
# not read as new debt. This mirrors eslint-suppressions.json's own (file,
# rule) keying, not (file, rule, line).
#
# Usage:
#   bash scripts/check-dead-code.sh
#     Runs knip live, compares to dead-code-baseline.json. Exit 0 = no new
#     dead file or unused export beyond the baseline. Exit 1 = new debt
#     found, or the baseline/knip output could not be trusted (see below).
#
#   bash scripts/check-dead-code.sh --write
#     Dev convenience (not run by CI): regenerates dead-code-baseline.json
#     from a live knip run, sorted and pretty-printed. The equivalent of
#     ESLint's `--prune-suppressions` for this gate — run after fixing dead
#     code so the baseline shrinks with it, or after deliberately, reviewedly
#     accepting new debt.
#
#   bash scripts/check-dead-code.sh --current-json FILE --baseline FILE
#     Test-only overrides (used by the guard-fixture harness, see
#     tools/verify/fixtures.go's ArgvOverride doc): skip the live knip run
#     and/or the repo's own baseline file, and compare two files directly.
#     FILE must already be in knip's `--reporter json` shape for
#     --current-json (an {"issues":[...]} object), or in this script's own
#     baseline shape for --baseline (an {"files":[...],"exports":[...]}
#     object).
#
# Fail-closed posture (A2.1 review caught two classes of this repo's guards
# failing OPEN; both are deliberately closed here):
#   - Malformed baseline shape (not an object, "files"/"exports" missing or
#     not arrays, an exports entry missing a string "file" or "name") is a
#     FAILURE, never treated as "nothing baselined, clean". A baseline that
#     cannot be trusted must not silently pass everything through it.
#   - knip producing output that is not parseable JSON (crash, config error,
#     missing binary, wrong pnpm/node on PATH) is a FAILURE with the raw
#     output surfaced, never conflated with "knip ran and found zero issues"
#     (which is also empty-ish output but valid JSON with an empty `issues`
#     array — the two are distinguished by JSON validity, not by emptiness).
set -euo pipefail

BASELINE="dead-code-baseline.json"
CURRENT_JSON_OVERRIDE=""
BASELINE_OVERRIDE=""
WRITE=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --current-json)
      CURRENT_JSON_OVERRIDE="$2"
      shift 2
      ;;
    --baseline)
      BASELINE_OVERRIDE="$2"
      shift 2
      ;;
    --write)
      WRITE=1
      shift
      ;;
    *)
      echo "check-dead-code: unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [[ -n "$BASELINE_OVERRIDE" ]]; then
  BASELINE="$BASELINE_OVERRIDE"
fi

WORKDIR=$(mktemp -d)
trap 'rm -rf "$WORKDIR"' EXIT

# ---- Obtain current knip output (live run, or a test override) -----------
if [[ -n "$CURRENT_JSON_OVERRIDE" ]]; then
  if [[ ! -f "$CURRENT_JSON_OVERRIDE" ]]; then
    echo "check-dead-code: FAIL — --current-json file not found: $CURRENT_JSON_OVERRIDE"
    exit 1
  fi
  CURRENT_JSON="$CURRENT_JSON_OVERRIDE"
else
  CURRENT_JSON="$WORKDIR/current.json"
  KNIP_STDERR="$WORKDIR/knip.stderr"
  set +e
  pnpm exec knip --include files,exports --reporter json --no-progress >"$CURRENT_JSON" 2>"$KNIP_STDERR"
  set -e
  # knip's own exit code is not the signal we branch on: it exits non-zero
  # whenever it finds ANY issue (that is the expected, common case — the
  # baseline is exactly for that), and 0 only on a fully clean tree. What we
  # actually need to distinguish is "knip ran and produced a JSON report"
  # (branch below, via `jq empty`) from "knip could not run at all" (crash,
  # missing binary, config error) — that distinction is JSON validity, not
  # exit code.
  if [[ ! -s "$CURRENT_JSON" ]] || ! jq empty "$CURRENT_JSON" 2>/dev/null; then
    echo "check-dead-code: FAIL — knip did not produce parseable JSON output (broken environment or config error, not a clean tree)."
    echo "--- knip stderr ---"
    cat "$KNIP_STDERR" 2>/dev/null || true
    echo "--- knip stdout ---"
    cat "$CURRENT_JSON" 2>/dev/null || true
    exit 1
  fi
fi

if ! jq empty "$CURRENT_JSON" 2>/dev/null; then
  echo "check-dead-code: FAIL — --current-json is not valid JSON: $CURRENT_JSON"
  exit 1
fi
if [[ "$(jq -r 'if type == "object" and (.issues | type) == "array" then "ok" else "bad" end' "$CURRENT_JSON" 2>/dev/null)" != "ok" ]]; then
  echo "check-dead-code: FAIL — current JSON is not knip's {\"issues\":[...]} shape: $CURRENT_JSON"
  exit 1
fi

CURRENT_FILES="$WORKDIR/current-files.txt"
CURRENT_EXPORTS="$WORKDIR/current-exports.txt"
jq -r '[.issues[] | select((.files // []) | length > 0) | .file] | sort | .[]' "$CURRENT_JSON" >"$CURRENT_FILES"
jq -r '[.issues[] | .file as $f | (.exports // [])[] | "\($f)\t\(.name)"] | sort | .[]' "$CURRENT_JSON" >"$CURRENT_EXPORTS"

# ---- --write: regenerate the baseline from the current run and stop ------
if [[ "$WRITE" -eq 1 ]]; then
  FILES_JSON="$WORKDIR/files.json"
  EXPORTS_JSON="$WORKDIR/exports.json"
  jq -R . "$CURRENT_FILES" | jq -s 'sort' >"$FILES_JSON"
  jq -R 'split("\t") | {file: .[0], name: .[1]}' "$CURRENT_EXPORTS" | jq -s 'sort_by(.file, .name)' >"$EXPORTS_JSON"
  jq -n --slurpfile files "$FILES_JSON" --slurpfile exports "$EXPORTS_JSON" \
    '{files: $files[0], exports: $exports[0]}' >"$BASELINE"
  echo "check-dead-code: wrote $BASELINE ($(wc -l <"$CURRENT_FILES" | tr -d ' ') file(s), $(wc -l <"$CURRENT_EXPORTS" | tr -d ' ') export(s))"
  exit 0
fi

# ---- Load and validate the baseline ---------------------------------------
if [[ ! -f "$BASELINE" ]]; then
  if [[ ! -s "$CURRENT_FILES" && ! -s "$CURRENT_EXPORTS" ]]; then
    echo "check-dead-code: no $BASELINE and knip reports nothing — nothing baselined, clean."
    exit 0
  fi
  echo "check-dead-code: FAIL — $BASELINE is missing but knip found dead files/unused exports."
  echo "check-dead-code: run 'bash scripts/check-dead-code.sh --write' to record them as the reviewed baseline, or fix them."
  exit 1
fi

if ! jq empty "$BASELINE" 2>/dev/null; then
  echo "check-dead-code: FAIL — $BASELINE is not valid JSON. A baseline that cannot be parsed cannot be trusted; this is a failure, not a clean pass."
  exit 1
fi

BASELINE_SHAPE_OK=$(jq -r '
  if type != "object" then "bad: not an object"
  elif (.files | type) != "array" then "bad: .files is not an array"
  elif (.exports | type) != "array" then "bad: .exports is not an array"
  elif ([.files[] | select(type != "string")] | length) > 0 then "bad: .files has a non-string entry"
  elif ([.exports[] | select(type != "object" or (.file | type) != "string" or (.name | type) != "string")] | length) > 0 then "bad: .exports has an entry missing string file/name"
  else "ok"
  end
' "$BASELINE")
if [[ "$BASELINE_SHAPE_OK" != "ok" ]]; then
  echo "check-dead-code: FAIL — $BASELINE has the wrong shape ($BASELINE_SHAPE_OK). MALFORMED baseline is a failure, never a silent pass."
  exit 1
fi

BASELINE_FILES="$WORKDIR/baseline-files.txt"
BASELINE_EXPORTS="$WORKDIR/baseline-exports.txt"
jq -r '.files | sort | .[]' "$BASELINE" >"$BASELINE_FILES"
jq -r '[.exports[] | "\(.file)\t\(.name)"] | sort | .[]' "$BASELINE" >"$BASELINE_EXPORTS"

# ---- Diff: anything CURRENT has that BASELINE does not is new debt -------
# comm -23: lines only in file 1 (current) — i.e. present now, absent from
# the baseline. Shrinkage (a baseline entry no longer reported by knip) is
# never flagged: burn-down must never be blocked, same policy as
# check-eslint-suppression-baseline-growth.sh.
NEW_FILES=$(comm -23 "$CURRENT_FILES" "$BASELINE_FILES" || true)
NEW_EXPORTS=$(comm -23 "$CURRENT_EXPORTS" "$BASELINE_EXPORTS" || true)

violations=0
if [[ -n "$NEW_FILES" ]]; then
  while IFS= read -r f; do
    [[ -z "$f" ]] && continue
    echo "check-dead-code: NEW DEAD FILE: $f (not in $BASELINE)"
    violations=$((violations + 1))
  done <<<"$NEW_FILES"
fi
if [[ -n "$NEW_EXPORTS" ]]; then
  while IFS=$'\t' read -r f n; do
    [[ -z "$f" ]] && continue
    echo "check-dead-code: NEW UNUSED EXPORT: $f :: $n (not in $BASELINE)"
    violations=$((violations + 1))
  done <<<"$NEW_EXPORTS"
fi

if (( violations > 0 )); then
  echo "check-dead-code: $violations new dead-code finding(s) beyond the recorded baseline — see above."
  echo "check-dead-code: fix them, or if reviewed and deliberate, run 'bash scripts/check-dead-code.sh --write' and include the baseline diff in the PR."
  exit 1
fi

echo "check-dead-code: clean — no dead file or unused export beyond $BASELINE ($(wc -l <"$BASELINE_FILES" | tr -d ' ') file(s), $(wc -l <"$BASELINE_EXPORTS" | tr -d ' ') export(s) baselined)"
