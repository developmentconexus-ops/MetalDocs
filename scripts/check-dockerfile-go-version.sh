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
# see the comment above. scripts/testdata/guard-fixtures/ is excluded: it
# deliberately contains drifted-on-purpose fixture Dockerfiles (suffixed
# .txt, but '*Dockerfile*' still matches the substring), and this check's
# own negative fixture must never make a clean checkout of the real repo
# fail against itself.
# vendor/ is excluded for the same reason check-gofmt.sh excludes it: it is
# upstream code this repo does not author and cannot edit -- `go mod vendor`
# overwrites any local change on the next dependency bump. Today
# vendor/go.opentelemetry.io/otel/dependencies.Dockerfile matches the discovery
# pattern and happens to pin no golang stage, so nothing fires; if upstream ever
# adds one below our go.mod floor, this check would fail an unrelated PR with a
# diagnostic nobody can act on. The fixture tree carries that file in the form
# that WOULD fire, and the check's NotWant entry keeps this exclusion honest.
mapfile -t dockerfiles < <(git ls-files -- '*.Dockerfile' '*Dockerfile*' ':!:scripts/testdata/guard-fixtures/**' ':!:vendor/**' | sort -u)

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
  [[ -f "$df" ]] || continue

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

    printf '%s\n' "$from_line" | grep -qiE 'golang:' || continue
    checked=$((checked + 1))
    # `|| true` for the same reason as mod_version above: `FROM golang:latest`
    # (or any non-numeric tag) matches the outer discovery grep but not this
    # one, and without the rescue the script would abort mid-loop instead of
    # reporting the unparseable line -- silently, and before checking any
    # remaining stage or Dockerfile.
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

echo "check-dockerfile-go-version: $checked Dockerfile(s) checked against go.mod's go $mod_version -- all OK"
