#!/usr/bin/env bash
# idempotency-identity-scope-guard (#90/A3.5a follow-up).
#
# What this enforces: internal/platform/idempotency/identity.go's
# TenantActorFromContext is documented as the ONLY sanctioned place that
# resolves (tenant, actor) identity for idempotency.Require's actorFromCtx
# parameter (func(context.Context) (string, string, error)). That claim used
# to be enforced by nothing but the docstring's own "grep for every call
# site" instruction — a comment is not a firing mechanism, and the defect
# this file exists to close (three hand-copied extractions, one of which
# silently discarded tenant.FromContext's error — see identity.go's history
# and #108/66cfb664) is exactly the class of bug that reproduces the moment
# a sixth handler writes its own two-line extractor instead of calling the
# shared function.
#
# This check fails when a Go function or closure matching actorFromCtx's
# shape — signature (ctx context.Context) (string, string, error) — calls
# tenant.FromContext anywhere in its body, in any tracked *.go file OTHER
# than internal/platform/idempotency/identity.go. It does not matter whether
# the error is propagated or discarded: any second implementation of this
# resolution is the drift this guard exists to catch, propagated correctly
# today or not, because a second correct copy is just a first copy that
# hasn't drifted yet.
#
# ---------------------------------------------------------------------------
# DETECTION STRATEGY — a text/regex + brace-depth scan, not an AST.
#
# The neighbouring checks in this registry (test-conventions, module-imports)
# are the same shape: grep/sed/awk over tracked source, not a compiler
# front end. This check follows that precedent rather than reaching for a Go
# analyzer, because the boundary of this change is scripts/check-*.sh plus
# the registry — not tools/cilint. tools/cilint/internal/analyzers already
# has an AST-based precedent for a near-identical problem (actorextraction.go,
# A3.3), including alias- and dot-import-following that a text scan cannot
# fully match; if this guard's honest gaps below ever bite in practice, that
# analyzer is the pattern to extend, not this script.
#
# One awk process reads every tracked *.go file (excluding vendor/ — upstream
# code this repo does not author — and identity.go itself, the one legal
# home), file by file, single pass, top to bottom:
#   1. Import-alias resolution: the first line in the file containing the
#      literal import path "metaldocs/internal/platform/tenant" fixes this
#      file's alias for the rest of the scan — default "tenant", a custom
#      alias (`platformtenant "metaldocs/.../tenant"`), or dot
#      (`. "metaldocs/.../tenant"`). A blank import (`_ "..."`) exposes no
#      callable symbol, so the file is left unusable and nothing downstream
#      can fire. Go requires imports before any declaration, so resolving
#      this once, on first sight, before function-scanning starts, is sound
#      for every valid Go file — there is no second import line to miss.
#   2. A file with no tenant import at all short-circuits to the next file
#      immediately (no header/body scan cost).
#   3. Header scan: a function or closure signature — named (`func NAME(ctx
#      context.Context) (string, string, error)`) or anonymous (`func(ctx
#      context.Context) (string, string, error)`) — opens a brace-depth
#      tracker. Depth is counted by `{`/`}` characters per line until it
#      returns to zero (the function's closing brace); Go requires the
#      opening `{` on the signature line, so this never has to guess where a
#      body starts.
#   4. Within that span, a call to the resolved tenant accessor —
#      `<alias>.FromContext(`, or a bare `FromContext(` for a dot import —
#      is reported against this file's path and line.
#
# One awk process rather than one subprocess per file: an earlier version of
# this script spawned `grep`+`sed`+`awk` per tracked file (~1200 files) and
# measured over a minute wall-clock on this machine, almost all of it process
# start-up overhead rather than work. A single awk reading every file as an
# argv entry removes essentially all of that. Files are still chunked through
# `xargs -n 200`, matching scripts/check-gofmt.sh's own documented reason:
# a single argv line long enough to hold every tracked Go file's path can
# exceed this platform's argument-list limit and silently run on zero files.
#
# STATED LIMITS (what this cannot catch, and why the gap is left open):
#   - Brace counting is character-level, not token-level: a `{` or `}`
#     inside a string, rune, or comment literal desyncs the depth count for
#     the rest of that scan. Rare in this codebase's identity-resolution
#     code, which is why the neighbouring checks accept the same class of
#     risk (test-conventions' own header documents an equivalent comment-
#     stripping heuristic and its blind spots).
#   - Method receivers are not matched: `func (h *Handler) x(ctx
#     context.Context) (string, string, error)` does not match the header
#     regex. Every resolver in this codebase today — TenantActorFromContext
#     and every call site it replaced — is a free function or closure, so
#     this is a stated gap, not a known-missed case.
#   - Named result lists are not matched: `func(ctx context.Context)
#     (tenantID, actorID string, err error)` is the SAME Go function type as
#     the unnamed form (parameter/result names do not affect type identity),
#     so it is just as assignable to Require's actorFromCtx parameter, but
#     this regex does not recognise it. Closing this honestly needs type
#     identity, not spelling — i.e. the AST alternative above, since named
#     results have too many equivalent groupings to enumerate in regex.
#   - A multi-line function signature (the `context.Context` parameter or
#     the `(string, string, error)` result wrapped onto its own line) is not
#     matched, because the header regex is evaluated per line. gofmt never
#     wraps a signature this short, so this is theoretical today.
#   - The tenant-package alias is resolved once per file, from whichever line
#     is seen FIRST containing the import path text. A line that merely
#     mentions the path inside a comment or string ahead of the real import
#     (contrived, not seen anywhere in this repo) would resolve the wrong
#     token; this is a text scan, not an import-declaration parser.
# ---------------------------------------------------------------------------
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

EXCLUDE_FILE="internal/platform/idempotency/identity.go"
TENANT_IMPORT_PATH='metaldocs/internal/platform/tenant'

mapfile -t FILES < <(git ls-files '*.go' | grep -v '^vendor/' | grep -vF "$EXCLUDE_FILE")

if [ "${#FILES[@]}" -eq 0 ]; then
  echo "FAIL: found no tracked Go files to check (excluding vendor/ and $EXCLUDE_FILE); refusing to report success on an empty sweep"
  exit 1
fi

# Bracket expressions ([(], [)], [.]) rather than backslash escapes: awk's
# dynamic-regexp path (a string handed to `~` via -v, compiled at run time)
# treats a backslash before an ERE metacharacter as a string escape first and
# a regex escape second — harmless, but gawk warns on every match attempt. A
# bracket expression matches the same literal character without being
# ambiguous between the two escaping passes, so it says the same thing with
# no warning.
HEADER_RE='func[ ]?[A-Za-z0-9_]*[(][ 	]*[A-Za-z_][A-Za-z0-9_]*[ 	]+context[.]Context[ 	]*[)][ 	]*[(][ 	]*string[ 	]*,[ 	]*string[ 	]*,[ 	]*error[ 	]*[)]'

AWK_PROG="$(mktemp)"
report_file="$(mktemp)"
trap 'rm -f "$AWK_PROG" "$report_file"' EXIT

cat > "$AWK_PROG" <<'AWKEOF'
FNR == 1 {
  fname = FILENAME
  resolved = 0
  usable = 0
  alias = ""
  dotimp = 0
  indepth = 0
  depth = 0
  funcline = 0
}
{
  line = $0

  # Import-alias resolution: Go requires every import before any
  # declaration, so the first line naming the tenant import path fixes this
  # file's alias for everything after it.
  if (!resolved && index(line, tenantimport) > 0) {
    tok = line
    sub(/[ 	]*"[^"]*"[ 	]*(\/\/.*)?$/, "", tok)
    gsub(/^[ 	]+/, "", tok)
    gsub(/[ 	]+$/, "", tok)
    resolved = 1
    if (tok == "_") {
      usable = 0                 # blank import: no callable symbol
    } else if (tok == ".") {
      usable = 1; dotimp = 1
    } else if (tok == "") {
      usable = 1; alias = "tenant"
    } else {
      usable = 1; alias = tok
    }
  }

  if (!usable) next

  if (!indepth) {
    if (line ~ hdr) {
      indepth = 1
      funcline = FNR
      depth = 0
    } else {
      next
    }
  }

  o = gsub(/\{/, "{", line)
  c = gsub(/\}/, "}", line)
  depth += (o - c)

  hit = 0
  if (dotimp == 1) {
    if (line ~ /(^|[^.A-Za-z0-9_])FromContext[ 	]*[(]/) hit = 1
  } else if (alias != "") {
    pat = alias "[.]FromContext[(]"
    if (line ~ pat) hit = 1
  }
  if (hit) {
    callee = (dotimp == 1) ? "the dot-imported tenant package's FromContext" : (alias ".FromContext")
    printf "%s:%d: actorFromCtx-shaped function (opened %s:%d) calls %s directly — outside identity.go this re-implements idempotency.TenantActorFromContext instead of calling it\n", fname, FNR, fname, funcline, callee
  }

  if (depth <= 0) { indepth = 0 }
}
AWKEOF

printf '%s\n' "${FILES[@]}" \
  | xargs -n 200 awk -v hdr="$HEADER_RE" -v tenantimport="\"${TENANT_IMPORT_PATH}\"" -f "$AWK_PROG" \
  > "$report_file"

violations=$(wc -l < "$report_file" | tr -d '[:space:]')

if [ "$violations" -gt 0 ]; then
  echo "idempotency-identity-scope-guard: $violations violation(s) found — a second actorFromCtx-shaped resolver is calling tenant.FromContext outside identity.go"
  cat "$report_file"
  echo
  echo "Fix: delete the hand-rolled closure and call idempotency.TenantActorFromContext instead."
  exit 1
fi

echo "idempotency-identity-scope-guard: clean (${#FILES[@]} Go files scanned, none outside identity.go re-implement actorFromCtx over tenant.FromContext)"
