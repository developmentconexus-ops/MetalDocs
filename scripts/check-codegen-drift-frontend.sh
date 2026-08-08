#!/usr/bin/env bash
set -euo pipefail

# frontend-codegen-drift — extracted verbatim from ci.yml:lint-contract's
# former inline "API codegen drift check (frontend)" step (Task R2).
# Registered as tools/verify check "codegen-drift-frontend". Needs pnpm on
# PATH with workspace deps already installed (`pnpm install --frozen-lockfile`
# stays a workflow prerequisite step, not part of this check).
cd frontend/apps/web
pnpm run gen:api
git diff --exit-code -- src/lib/api-types/ || (echo "::error::Run 'pnpm run gen:api' in frontend/apps/web and commit"; exit 1)
