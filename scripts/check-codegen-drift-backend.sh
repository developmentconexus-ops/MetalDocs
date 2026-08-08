#!/usr/bin/env bash
set -euo pipefail

# backend-codegen-drift — extracted verbatim from ci.yml:lint-contract's
# former inline "API codegen drift check (backend)" step (Task R2). Registered
# as tools/verify check "codegen-drift-backend".
go generate ./...
git diff --exit-code -- '**/api.gen.go' '**/httpsurface_gen.go' '**/httpsurface_e2e_gen.go' || (echo "::error::Run 'go generate ./...' and commit"; exit 1)
