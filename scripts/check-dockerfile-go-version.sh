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
mod_version="$(grep -E '^go [0-9]+\.[0-9]+(\.[0-9]+)?$' "$gomod" | head -1 | awk '{print $2}')"
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
  from_line="$(grep -E '^FROM[[:space:]]+golang:' "$df" | head -1 || true)"
  [[ -z "$from_line" ]] && continue
  checked=$((checked + 1))
  df_version="$(printf '%s\n' "$from_line" | sed -E 's/^FROM[[:space:]]+golang:([0-9]+(\.[0-9]+){0,2}).*/\1/')"
  if [[ -z "$df_version" || "$df_version" == "$from_line" ]]; then
    echo "DOCKERFILE-GO-VERSION-DRIFT: $df: could not parse a numeric golang version from: $from_line" >&2
    fail=1
    continue
  fi
  if ! version_ge "$df_version" "$mod_version"; then
    echo "DOCKERFILE-GO-VERSION-DRIFT: $df pins golang:$df_version but go.mod requires go >= $mod_version" >&2
    fail=1
  fi
done

if (( checked == 0 )); then
  echo "check-dockerfile-go-version: no tracked Dockerfile declares a 'FROM golang:' builder stage" >&2
  exit 1
fi

if (( fail == 1 )); then
  exit 1
fi

echo "check-dockerfile-go-version: $checked Dockerfile(s) checked against go.mod's go $mod_version -- all OK"
