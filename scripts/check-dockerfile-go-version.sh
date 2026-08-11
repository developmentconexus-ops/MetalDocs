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
mapfile -t dockerfiles < <(git ls-files -- '*.Dockerfile' '*Dockerfile*' ':!:scripts/testdata/guard-fixtures/**' | sort -u)

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
  # Every Go builder stage matters, not just the first: a multi-stage
  # Dockerfile can carry more than one `FROM golang:` line (e.g. a stale
  # intermediate builder left behind a refactor), and a stage checked only
  # once at the top of the file would let a later one drift undetected.
  # Match case-insensitively (Docker's own instructions are not
  # case-sensitive -- `from`/`From`/`FROM` are all valid), allow leading and
  # intervening whitespace, and allow a `--platform=...` flag between `FROM`
  # and the image reference (`FROM [--platform=<platform>] <image>`, per the
  # Dockerfile reference).
  while IFS= read -r matched; do
    [[ -z "$matched" ]] && continue
    lineno="${matched%%:*}"
    from_line="${matched#*:}"
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
  done < <(grep -inE '^[[:space:]]*from[[:space:]]+(--platform=[^[:space:]]+[[:space:]]+)?golang:' "$df" || true)
done

if (( checked == 0 )); then
  echo "check-dockerfile-go-version: no tracked Dockerfile declares a 'FROM golang:' builder stage" >&2
  exit 1
fi

if (( fail == 1 )); then
  exit 1
fi

echo "check-dockerfile-go-version: $checked Dockerfile(s) checked against go.mod's go $mod_version -- all OK"
