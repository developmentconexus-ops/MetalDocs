#!/usr/bin/env bash
# Fails if any tracked Go source file is not gofmt-clean.
#
# Why this exists: nothing enforced gofmt, so 96 files had drifted by the time
# A1 looked. golangci-lint's enabled set does not include a formatter, and
# `go vet` does not care about layout. An unenforced convention is not a
# convention, it is a preference that some files happen to follow.
#
# Run it through the pinned toolchain, not whatever gofmt is on PATH: gofmt
# output differs between Go minor versions, so a sweep with the wrong one
# reports drift that CI does not see, or misses drift that it does.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

# --others --exclude-standard includes files that are new and not yet staged,
# while still honouring .gitignore. Without it this checks only TRACKED files,
# so a brand-new unformatted file passes on the laptop and fails in CI the
# moment it is committed — the precise "green locally, red in CI" split this
# tool exists to remove. The negative fixture caught that on its first run.
#
# vendor/ is upstream code formatted by its own authors, and *.gen.go is
# oapi-codegen output whose layout is the generator's business.
mapfile -t files < <(git ls-files --cached --others --exclude-standard '*.go' \
    | grep -v '^vendor/' | grep -v '\.gen\.go$')

# Fail closed. An empty list would make this script exit 0 while checking
# nothing — the vacuous-gate failure mode this whole axis exists to remove.
# (It is also the exact bug that made the first version of this sweep report
# "0 unformatted": 1164 paths on one command line blew the argv limit, gofmt
# never ran, and stderr was suppressed. Hence xargs below, and this guard.)
if [ "${#files[@]}" -eq 0 ]; then
    echo "FAIL: found no tracked Go files to check; refusing to report success on an empty sweep"
    exit 1
fi

unformatted=$(printf '%s\n' "${files[@]}" | xargs -n 200 gofmt -l)

if [ -n "$unformatted" ]; then
    echo "FAIL: these files are not gofmt-clean:"
    echo "$unformatted" | sed 's/^/  /'
    echo
    echo "Fix with: gofmt -w <file>   (use the toolchain go.mod pins, not a newer local one)"
    exit 1
fi

echo "OK: ${#files[@]} Go files are gofmt-clean"
