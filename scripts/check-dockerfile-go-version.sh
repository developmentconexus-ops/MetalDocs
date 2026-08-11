#!/usr/bin/env bash
set -euo pipefail

# check-dockerfile-go-version — registered as tools/verify check
# "dockerfile-go-version-drift".
#
# go.mod's `go` directive is the single source of truth for the minimum Go
# toolchain this repo requires (see the comment above it, which explains why
# it is the `go` directive and not `toolchain` -- GO-2026-5856). Every
# Dockerfile that builds a Go binary hardcodes a `FROM golang:<version>`
# builder base, and nothing forced the two to agree: deploy/docker/*.Dockerfile
# drifted to golang:1.25 while go.mod moved to go 1.26.5, and no CI job builds
# container images, so nothing caught it before the image build itself failed
# (docs/engineering/defect-class-catalog.md Class 2, Hand-Synced Enumerations).
#
# This check does not hand-list the Dockerfiles that matter -- that would be
# the same defect class one layer up. It discovers every tracked Dockerfile
# and checks any that declare a golang builder stage.
#
# Parsing limit, deliberately imposed (2026-08-11, CodeRabbit review on #114,
# thread at line ~113): Docker applies line-continuation to `FROM` like any
# other instruction, and a `# escape=` parser directive can swap the active
# continuation character from `\` to a backtick -- so
#   FROM --platform=$BUILDPLATFORM \
#     golang:1.19 AS builder
# is one logical FROM instruction whose `golang:` reference never appears on
# the physical line starting with FROM at all. A scanner that reads physical
# lines, as this one does, cannot see it -- a stale Go builder using either
# continuation form silently bypassed this check before this comment was
# written. The correct-looking fix is to join continuation lines before
# matching, honouring whatever escape character is active. That fix was
# rejected: a Dockerfile line-joiner is more code than this check, and it
# would carry its own untested edge cases (a directive is only honoured
# before the first instruction; only `\` and `` ` `` are legal; a directive
# appearing later is inert) -- new silent-wrong surface serving a drift
# check that should not need to know Dockerfile grammar at all.
# Instead: this check requires `FROM golang:<version>` on ONE physical
# line, full stop, and refuses -- loudly, exit 1, named file:line -- any
# Dockerfile that uses a `# escape=` parser directive or a continued FROM
# instruction, rather than silently skip a builder stage it cannot see. That
# is a real constraint on this repo's Dockerfile style, written down here.
#
# "Is this a golang stage?" and "is its version parseable?" are two
# different questions (independent review on #114, found live: a compliant
# first stage plus a digest-pinned second stage -- `FROM golang@sha256:<hex>
# AS builder`, normal supply-chain practice -- reported "all OK" and exited
# 0). A single `grep -qiE 'golang:'` test used to answer both at once, which
# dumped "golang, but no parseable tag" into the same bucket as "not golang
# at all" and skipped it silently. This check now decides (a) first, from
# the image reference's own repository component (the last `/`-segment,
# digest and tag stripped) compared case-insensitively to `golang` -- and
# only once that is yes does a missing/non-numeric tag become the existing
# loud failure. This is deliberately over-inclusive in the same spirit as
# the escape-directive and trailing-whitespace handling above: a vendored
# `myorg/golang:1.26` image, or a local multi-stage alias literally named
# `golang`, gets checked too. That costs an occasional false-positive
# diagnostic; the alternative -- a narrower match -- costs a silent miss,
# which is the one failure mode this check exists to close.
#
# Discovery itself had the same shape of bug, one layer up (independent
# review on #114): the `git ls-files` pathspec matched the string
# "dockerfile" anywhere in a file's full path rather than testing whether
# the file IS a Dockerfile, so a directory merely named `*-dockerfiles/`
# swept in unrelated prose files, and this script's own filename matched
# its own guard. Discovery is now basename-scoped at path-component
# boundaries (`git ls-files -- ':(glob,icase)**/*.Dockerfile'
# ':(glob,icase)**/Dockerfile' ...`) -- see the comment above the
# `mapfile` call below for the full account.
#
# Two more silent-skip holes were found the same way, in a second independent
# review on #114, both fixed the same way -- fail loud instead of a bare
# `continue`:
#   1. A discovered path that is not a readable regular file (a submodule
#      gitlink, a symlink with a missing target) used to vanish with no
#      diagnostic at the very top of the per-file loop, before any content
#      was even read. See the comment above the `[[ ! -f "$df" ]]` check
#      further down for the reproduction and why the fixture proving it uses
#      a gitlink, not a symlink.
#   2. `FROM $VAR` / `FROM ${VAR}` -- a build variable standing in for the
#      WHOLE image reference, not just its tag -- resolved to a literal
#      `$BASE_IMAGE`-shaped token that never matches the golang repo-component
#      test, so a stage like `ARG BASE_IMAGE=golang:1.10` + `FROM $BASE_IMAGE`
#      passed with zero diagnostic and no `checked` effect. See the comment
#      above the whole-reference-indirection check further down for the
#      contrast with `FROM golang:${GO_VERSION}` (a variable TAG, still
#      caught correctly) and why this matters now: three of this repo's five
#      Dockerfiles carry comments describing a planned consolidation onto
#      exactly the `ARG BASE_IMAGE` / `FROM $BASE_IMAGE` shape.
#
# `checked` counts golang builder STAGES (and stage-shaped positions this
# check refused to evaluate and therefore failed closed), not files -- a
# single multi-stage Dockerfile with two `FROM golang:` lines increments it
# twice. The final success message used to say "N Dockerfile(s) checked",
# which reads as a file count and is wrong whenever a file has more than one
# golang stage; it now says "N golang builder stage(s) checked" so the number
# means what it says.
#
# Static: parses go.mod and tracked Dockerfiles. No Docker daemon, no network.

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

gomod="go.mod"
if [[ ! -f "$gomod" ]]; then
  echo "check-dockerfile-go-version: $gomod not found" >&2
  exit 1
fi

# The `go` directive line is exactly "go X.Y" or "go X.Y.Z" -- see the
# comment in go.mod: deliberately not `toolchain`.
# `|| true` is load-bearing under `set -euo pipefail`: a grep that matches
# nothing exits 1, pipefail propagates that through the pipeline, and the
# failing command substitution would abort the script right here -- making
# the emptiness check below unreachable and turning a diagnosable parse
# failure into a silent non-zero exit with no output at all. Rescue the
# pipeline so the message below is the thing that actually reports.
mod_version="$(grep -E '^go [0-9]+\.[0-9]+(\.[0-9]+)?$' "$gomod" | head -1 | awk '{print $2}' || true)"
if [[ -z "$mod_version" ]]; then
  echo "check-dockerfile-go-version: could not find a 'go' directive in $gomod" >&2
  exit 1
fi

# Every tracked Dockerfile, not a hand-kept list of "the ones that matter" --
# see the comment above.
#
# Discovery is BASENAME-scoped, matched at path-component boundaries,
# case-insensitively: `:(glob,icase)**/*.Dockerfile` and
# `:(glob,icase)**/Dockerfile` -- this repo's two Dockerfile conventions,
# bare `Dockerfile` and `<name>.Dockerfile`. The `glob` magic is load-bearing:
# under it `*` stops matching `/`, so a wildcard can no longer span path
# components, and a leading `**/` matches at any depth including the repo
# root. This replaced a plain `'*Dockerfile*'` pathspec, which had TWO
# stacked defects, found together on #114 (independent review):
#   1. It was case-sensitive (confirmed even with core.ignorecase=true, and
#      CI runs on Linux regardless), so a legitimately-spelled
#      `worker.dockerfile` was never returned by `git ls-files` at all --
#      not skipped, not excluded, simply invisible to this tool.
#   2. Making it case-insensitive with `:(icase)` alone did not fix that --
#      it widened a second, deeper defect instead. A no-slash glob pathspec
#      matches the FULL PATH, not the basename, so `'*Dockerfile*'` had
#      always meant "the string 'dockerfile' appears somewhere in this
#      path", not "this file IS a Dockerfile". Case-insensitive, that
#      predicate swept in doc/prose files whose PARENT DIRECTORY merely
#      contained the word (e.g. a `*-dockerfiles/` folder pulled in every
#      unrelated file inside it) and even this check's own script (its
#      filename contains "dockerfile"). Scoping to the basename, at a `/`
#      boundary, is the actual fix; matching on it case-insensitively is a
#      separate, smaller correctness fix on top -- same lesson as the
#      golang-stage split above, one layer up: match what the predicate
#      claims to test, not a proxy for it.
# This check already treats Dockerfile *instruction* syntax
# case-insensitively (`grep -inE '^[[:space:]]*from'`); the *filenames* it
# discovers get the same care now, correctly scoped.
#
# scripts/testdata/guard-fixtures/ is excluded: it deliberately contains
# drifted-on-purpose fixture Dockerfiles. Every fixture file in that tree is
# suffixed `.txt` on disk (the fixture harness's copyTree strips exactly one
# trailing `.txt` when it stages the sandbox), so under basename-scoped
# matching none of them are discovered by this pathspec in the first place
# -- this exclusion is now defense in depth against a future fixture added
# without that suffix, not load-bearing for the fixtures that exist today.
# Kept anyway: the failure mode it guards is "this check's own negative
# fixture makes a clean checkout of the real repo fail against itself",
# which is cheap to keep excluded and expensive to debug if it ever regresses.
#
# vendor/ is excluded for the same reason check-gofmt.sh excludes it: it is
# upstream code this repo does not author and cannot edit -- `go mod vendor`
# overwrites any local change on the next dependency bump. Today
# vendor/go.opentelemetry.io/otel/dependencies.Dockerfile matches the discovery
# pattern (its basename IS a real `<name>.Dockerfile`) and happens to pin no
# golang stage, so nothing fires; if upstream ever adds one below our go.mod
# floor, this check would fail an unrelated PR with a diagnostic nobody can
# act on. The fixture tree carries that file in the form that WOULD fire,
# and the check's NotWant entry keeps this exclusion honest.
mapfile -t dockerfiles < <(git ls-files -- ':(glob,icase)**/*.Dockerfile' ':(glob,icase)**/Dockerfile' ':!:scripts/testdata/guard-fixtures/**' ':!:vendor/**' | sort -u)

# version_ge A B: true (exit 0) if dotted-numeric version A >= B, comparing
# component-wise with a missing trailing component treated as 0 (so "1.26"
# is treated as satisfying a requirement of "1.26.0" but not "1.26.5").
version_ge() {
  local a="$1" b="$2"
  local IFS=.
  local -a av=($a) bv=($b)
  local len=${#av[@]}
  (( ${#bv[@]} > len )) && len=${#bv[@]}
  local i ai bi
  for (( i = 0; i < len; i++ )); do
    ai=${av[i]:-0}
    bi=${bv[i]:-0}
    if (( ai > bi )); then return 0; fi
    if (( ai < bi )); then return 1; fi
  done
  return 0
}

fail=0
checked=0
for df in "${dockerfiles[@]}"; do
  # `git ls-files` returns a tracked PATH, not a promise that the checkout
  # holds a readable regular file at it -- a submodule gitlink (mode 160000)
  # or a symlink whose target is missing both satisfy the discovery pathspec
  # while leaving nothing this check can open. A bare `continue` here used to
  # drop such a path with no diagnostic and no `checked` effect: the same
  # silent-skip shape as the digest-pinned and case-sensitivity bugs above,
  # one layer earlier, in discovery itself rather than in parsing (found
  # live, 2026-08-11, independent review on #114, proved with a gitlink
  # entry -- a real dangling symlink could not be constructed to prove this
  # on every checkout this check runs on, since a Windows checkout with
  # `core.symlinks=false` never materializes one; the gitlink reproduces the
  # identical failure at the `[[ -f ]]` test, platform-independently, and is
  # the fixture below). Fail loud instead of skipping.
  if [[ ! -f "$df" ]]; then
    echo "DOCKERFILE-GO-VERSION-DRIFT: $df: git tracks this path but it is not a readable regular file in this checkout (a submodule gitlink, a symlink with a missing target, or another non-regular entry) -- this check cannot read it to look for a golang builder stage, and treats \"cannot read\" as \"must not silently pass\"" >&2
    fail=1
    checked=$((checked + 1))
    continue
  fi

  # A `# escape=` parser directive is only ever honoured by Docker in the
  # preamble before the first instruction -- anything found after that is
  # just an inert comment, and flagging it would fail files that mention
  # "escape=" in later prose for no reason. So: find the first FROM, and
  # look for the directive only above it. (Docker's real rule is narrower
  # still -- a blank line or a non-directive comment also ends the directive
  # search -- but a narrower detector here would only mean fewer files fail
  # closed, never more, so this stays the safe over-inclusive direction.)
  from_lineno="$(grep -inE '^[[:space:]]*from[[:space:]]+' "$df" | head -1 | cut -d: -f1 || true)"
  if [[ -n "$from_lineno" ]]; then
    preamble_lines=$((from_lineno - 1))
  else
    preamble_lines="$(wc -l < "$df" || echo 0)"
  fi
  escape_match=""
  if (( preamble_lines > 0 )); then
    escape_match="$(head -n "$preamble_lines" "$df" | grep -inE '^[[:space:]]*#[[:space:]]*escape[[:space:]]*=' | head -1 || true)"
  fi
  if [[ -n "$escape_match" ]]; then
    escape_lineno="${escape_match%%:*}"
    escape_text="${escape_match#*:}"
    echo "DOCKERFILE-GO-VERSION-DRIFT: $df:$escape_lineno: found a '# escape=' parser directive ($escape_text); this check reads Dockerfiles one physical line at a time and cannot honour a custom escape character, so it cannot safely parse any FROM instruction in this file -- put FROM golang:<version> on a single physical line and drop the escape directive" >&2
    fail=1
    checked=$((checked + 1))
    continue
  fi

  # Every Go builder stage matters, not just the first: a multi-stage
  # Dockerfile can carry more than one `FROM golang:` line (e.g. a stale
  # intermediate builder left behind a refactor), and a stage checked only
  # once at the top of the file would let a later one drift undetected.
  # Match case-insensitively (Docker's own instructions are not
  # case-sensitive -- `from`/`From`/`FROM` are all valid), allow leading and
  # intervening whitespace, and allow a `--platform=...` flag between `FROM`
  # and the image reference (`FROM [--platform=<platform>] <image>`, per the
  # Dockerfile reference). This matches every FROM instruction now, not only
  # ones with `golang:` on the same physical line -- a continued FROM's
  # image reference can be on the NEXT physical line, invisible to a
  # same-line match, and the continuation check below needs to see the FROM
  # line itself to catch that case.
  while IFS= read -r matched; do
    [[ -z "$matched" ]] && continue
    lineno="${matched%%:*}"
    from_line="${matched#*:}"

    # Trim a trailing \r (CRLF checkout) and trailing whitespace before
    # testing for the continuation character. Docker's own parser is
    # actually stricter than this -- trailing whitespace AFTER the `\`
    # breaks continuation for Docker itself -- but treating it as a
    # continuation anyway is the safe direction: over-flagging a line Docker
    # would treat as complete costs a false alarm, under-flagging a real
    # continuation ships the exact bypass this detection exists for.
    trimmed="${from_line%$'\r'}"
    trimmed="${trimmed%"${trimmed##*[![:space:]]}"}"
    if [[ "$trimmed" == *'\' ]]; then
      echo "DOCKERFILE-GO-VERSION-DRIFT: $df:$lineno: this check cannot parse a continued FROM instruction (line ends with a line-continuation character); put the image reference on the same physical line as FROM -- offending line: $from_line" >&2
      fail=1
      checked=$((checked + 1))
      continue
    fi

    # "Is this a golang stage?" and "is its version parseable?" are two
    # different questions, and they used to be one `grep -qiE 'golang:'`
    # test -- which meant a digest-pinned stage (`FROM golang@sha256:<hex>
    # AS builder`, normal supply-chain practice) contained no `golang:`
    # substring at all and fell into the bare `continue` below with no
    # diagnostic and no `checked` increment: a stale Go builder shipped
    # silently, exit 0 (found live, 2026-08-11 review on #114). Answer (a)
    # first, on the image reference alone; only once (a) is yes does an
    # unparseable version become the existing loud failure instead of a
    # silent skip.
    #
    # Extract the image reference: the first whitespace-delimited token
    # after FROM and its optional `--platform=...` flag, i.e. everything up
    # to (but not including) an ` AS <alias>`.
    image_ref=""
    if [[ "$from_line" =~ ^[[:space:]]*[Ff][Rr][Oo][Mm][[:space:]]+(--platform=[^[:space:]]+[[:space:]]+)?([^[:space:]]+) ]]; then
      image_ref="${BASH_REMATCH[2]}"
    fi
    [[ -z "$image_ref" ]] && continue

    # Whole-reference indirection: `FROM $VAR` or `FROM ${VAR}` is a build
    # variable standing in for the ENTIRE image reference, not just its tag
    # or digest. Resolving it would mean evaluating the ARG's default (or
    # whatever `--build-arg` overrides it at build time) -- i.e. writing a
    # Dockerfile/build interpreter, which this check has already refused to
    # become (see the escape-directive and continued-FROM refusals above).
    # Contrast: `FROM golang:${GO_VERSION}` substitutes only the TAG
    # position -- the repository component is still the literal string
    # `golang`, so the ordinary match two blocks down catches it correctly
    # (and then, since the tag itself is non-numeric text, the existing
    # "could not parse a numeric golang version" failure below fires -- no
    # special-case needed for that shape). Only a substitution of the WHOLE
    # reference is invisible to the repo-component check, because no literal
    # `golang` is left anywhere on the line to match against (found live,
    # 2026-08-11, independent review on #114: `ARG BASE_IMAGE=golang:1.10` +
    # `FROM $BASE_IMAGE` reproduced silently -- zero diagnostic, `checked`
    # untouched, exit 0 -- in a tree where deploy/docker/*.Dockerfile's own
    # comments describe a planned consolidation onto exactly this pattern,
    # A7.4/A7.5, threading GO_VERSION from go.mod through an ARG BASE_IMAGE).
    # Fail loud: this check cannot resolve ARG values, so it cannot tell
    # whether the image behind the indirection is golang at all, and treats
    # "cannot tell" as "must not silently pass" rather than skipping.
    if [[ "$image_ref" =~ ^\$\{?[A-Za-z_][A-Za-z0-9_]*\}?$ ]]; then
      echo "DOCKERFILE-GO-VERSION-DRIFT: $df:$lineno: FROM's image reference is a build variable standing in for the WHOLE image ($image_ref), not just a tag -- this check cannot resolve ARG values to tell whether the resolved image is golang, so it cannot safely skip this stage; pin the golang version literally, or keep the repository name literal and vary only the tag (FROM golang:\${GO_VERSION} is fine and already checked) -- offending line: $from_line" >&2
      fail=1
      checked=$((checked + 1))
      continue
    fi

    # Repository component: the LAST `/`-separated path segment FIRST --
    # otherwise a registry host with a port (`registry:5000/golang@sha256:x`)
    # would have its `:5000` mistaken for a tag separator -- THEN strip an
    # `@sha256:...` digest suffix, THEN strip a trailing `:tag`. Order
    # matters; doing the tag/digest strip before the path split does not.
    repo_component="${image_ref##*/}"
    repo_component="${repo_component%%@*}"
    repo_component="${repo_component%%:*}"
    # Case-insensitive on purpose and deliberately over-inclusive: this
    # matches `myorg/golang:1.26` (a vendored image merely named "golang",
    # not the upstream one) and a local multi-stage alias literally named
    # `golang` (`FROM previous AS golang` makes a later `FROM golang`
    # resolve to that stage, not an image) just as readily as the real
    # thing. Both get checked anyway -- same reasoning the escape-directive
    # detector above already uses: a narrower match here only means fewer
    # files fail closed, never more, so over-inclusive is the safe
    # direction. A false positive costs one wasted diagnostic; a false
    # negative ships the exact silent-bypass this check exists to prevent.
    if [[ "${repo_component,,}" != "golang" ]]; then
      continue
    fi
    checked=$((checked + 1))
    # `|| true` for the same reason as mod_version above: `FROM golang:latest`
    # (or any non-numeric tag, or no tag at all -- digest-pinned or a bare
    # `golang` stage alias) matches here but not this pattern, and without
    # the rescue the script would abort mid-loop instead of reporting the
    # unparseable line -- silently, and before checking any remaining stage
    # or Dockerfile.
    df_version="$(printf '%s\n' "$from_line" | grep -oiE 'golang:[0-9]+(\.[0-9]+){0,2}' | head -1 | cut -d: -f2 || true)"
    if [[ -z "$df_version" ]]; then
      echo "DOCKERFILE-GO-VERSION-DRIFT: $df:$lineno: could not parse a numeric golang version from: $from_line" >&2
      fail=1
      continue
    fi
    if ! version_ge "$df_version" "$mod_version"; then
      echo "DOCKERFILE-GO-VERSION-DRIFT: $df pins golang:$df_version but go.mod requires go >= $mod_version (line $lineno)" >&2
      fail=1
    fi
  done < <(grep -inE '^[[:space:]]*from[[:space:]]+(--platform=[^[:space:]]+[[:space:]]+)?' "$df" || true)
done

if (( checked == 0 )); then
  echo "check-dockerfile-go-version: no tracked Dockerfile declares a 'FROM golang:' builder stage" >&2
  exit 1
fi

if (( fail == 1 )); then
  exit 1
fi

echo "check-dockerfile-go-version: $checked golang builder stage(s) checked against go.mod's go $mod_version -- all OK"
