#!/usr/bin/env bash
# idempotency-identity-scope-guard (#90/A3.5a follow-up).
#
# The rule is implemented by the adjacent Go AST/type-resolution scanner.
# Keeping this shell entrypoint preserves the command used by CI and by the
# negative-fixture harness while avoiding a second, weaker parser in awk.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"
exec go run ./scripts/check-idempotency-identity-scope.go
