# Phase 4 — Single pnpm Workspace Consolidation (Design)

**Date:** 2026-06-23
**Status:** Approved
**Topic:** Consolidate the repo onto ONE package manager (pnpm) as a single unified workspace.

## Problem

The repo runs **two package managers simultaneously**:

- **npm** — root (`package-lock.json` + npm `workspaces: ["apps/docx-renderer", "packages/*"]`), `.github/workflows/ci.yml` (`npm ci --include-workspace-root`), `.github/workflows/api-contract.yml` (`npm ci` in frontend), `shared/schemas` (`package-lock.json`).
- **pnpm** — `frontend/apps/web` (own `pnpm-lock.yaml`, NOT a member of the root workspace, links `@metaldocs/editor-ui` via a `file:` path), `packages/editor-ui` (`pnpm-lock.yaml`), `.github/workflows/e2e-coverage-gate.yml` (`pnpm/action-setup`).

Consequences:
- No `packageManager` field and no `pnpm-workspace.yaml` → the canonical PM is ambiguous.
- Multiple unsynced lockfiles (root `package-lock.json` + a 114-byte stray root `pnpm-lock.yaml` + `frontend/apps/web/pnpm-lock.yaml` + `packages/editor-ui/pnpm-lock.yaml` + `shared/schemas/package-lock.json`).
- `frontend/apps/web` is isolated from the workspace, so internal packages are consumed by `file:` link rather than a workspace link — the source of the recurring junction/version drift that breaks vitest+vite (see memory `fe-node-modules-junction-drift`).
- A latent CI bug: `e2e-coverage-gate.yml:125` sets `cache-dependency-path: frontend/pnpm-lock.yaml`, a path that does not exist (the real lockfile is `frontend/apps/web/pnpm-lock.yaml`) → cache misses.

## Goal

A single pnpm workspace rooted at the repo root: one package manager, one lockfile, one `node_modules` store, `pnpm -r` orchestration. A clean install from scratch root-causes the junction drift. CI runs entirely on pnpm.

## Non-Goals

- No dependency version upgrades. Consolidation only; versions stay as resolved today unless a conflict forces a pin (documented if so).
- No source-code refactoring. Only package-manager config, scripts, lockfiles, and CI.
- No change to what each package builds/tests — only how the orchestrator invokes them.

## Target End-State

### 1. Workspace config

- **New `pnpm-workspace.yaml`** at repo root:
  ```yaml
  packages:
    - 'frontend/apps/web'
    - 'packages/*'
    - 'apps/*'
    - 'shared/*'
  ```
- **Root `package.json`:**
  - Remove the npm `workspaces` field (pnpm reads `pnpm-workspace.yaml`).
  - Add `"packageManager": "pnpm@9.x"` (exact 9.x pin matching the version corepack/CI uses).
  - Scripts keep their NAMES (CI and `CLAUDE.md` reference them) but change implementation:
    - `build:docx-v2`: `npm -ws --if-present run build` → `pnpm -r run build`
    - `test:docx-v2`: `npm -ws --if-present run test` → `pnpm -r run test`
    - `typecheck:docx-v2`: `npm -ws --if-present run typecheck` → `pnpm -r run typecheck`

> Note on `pnpm -r` vs `--if-present`: `pnpm -r run <script>` already skips workspace members that lack the script, so the `--if-present` semantics are preserved without a flag.

### 2. Internal link

- `frontend/apps/web/package.json`: `"@metaldocs/editor-ui": "file:../../../packages/editor-ui"` → `"@metaldocs/editor-ui": "workspace:*"`.
- Any other intra-repo `@metaldocs/*` dependency expressed as a `file:`/version path is converted to `workspace:*` in the same pass (audit all member `package.json` during the plan).

### 3. Lockfile consolidation

- **Delete:** root `package-lock.json`, root stray `pnpm-lock.yaml`, `frontend/apps/web/pnpm-lock.yaml`, `packages/editor-ui/pnpm-lock.yaml`, `shared/schemas/package-lock.json` (and any other `*-lock.json` / nested `pnpm-lock.yaml` found outside `node_modules` during the plan audit).
- **Nuke** all `node_modules` directories (root + every member).
- **`pnpm install`** from scratch → exactly one root `pnpm-lock.yaml`.

### 4. Local scripts

- `scripts/dev-api-web.ps1`, `scripts/run_metaldocs.ps1`, `scripts/dev-local.ps1`: `npm run dev` (in `frontend/apps/web`) → `pnpm --filter @metaldocs/web run dev` (or `pnpm -C frontend/apps/web run dev` — whichever the existing script structure makes cleanest; preserve behavior).
- `Makefile`:
  - `test` target: `cd frontend/apps/web && npx vitest run` → `pnpm --filter @metaldocs/web exec vitest run`.
  - `test-watch` target: `cd frontend/apps/web && npx vitest` → `pnpm --filter @metaldocs/web exec vitest`.
  - Target names (`make test`, `make test-watch`) unchanged.

### 5. CI workflows (sequenced LAST; verified on the real run)

- **`ci.yml`:** replace `cache: npm` + `npm ci --include-workspace-root` with `pnpm/action-setup@v3` (version 9) + `actions/setup-node` `cache: pnpm` + `pnpm install --frozen-lockfile`; `npm run typecheck:docx-v2` / `test:docx-v2` / `build:docx-v2` → `pnpm run …` (same script names).
- **`api-contract.yml`:** `npm ci` (frontend) → root `pnpm install --frozen-lockfile`; `npm run gen:api` → `pnpm --filter @metaldocs/web run gen:api`.
- **`e2e-coverage-gate.yml`:** fix `cache-dependency-path` → root `pnpm-lock.yaml`; install at repo root (`pnpm install --frozen-lockfile`); Playwright steps run via `pnpm --filter @metaldocs/web exec …` (or correct working-directory).

## Risks & Mitigations

| Risk | Mitigation |
|---|---|
| vite/vitest fail to resolve deps under pnpm's symlinked `node_modules` | Fallback only if it breaks: add root `.npmrc` with `shamefully-hoist=true`. NOT applied by default — try the clean symlinked install first. |
| CI cannot be fully verified on the local Windows machine | CI tasks are the LAST tasks in the plan; each includes an explicit "push and watch the actual workflow run go green" verification step. |
| `pnpm -r` build order across workspace deps | pnpm topologically sorts by workspace dependency graph (builds `eigenpal-adapter`/`docx-renderer`/shared packages before `web`); expected correct. Verify via a clean `pnpm -r run build`. |
| Hidden intra-repo `file:`/version deps missed | Plan audits every member `package.json` for `@metaldocs/*` deps and converts all to `workspace:*` before the first install. |
| Junction/version drift persists | A from-scratch nuke + reinstall is the root-cause fix (the drift came from an incomplete install, not a bad linker). |

## Definition of Done

1. Exactly ONE lockfile in the repo: root `pnpm-lock.yaml`. Zero `package-lock.json`; zero other `pnpm-lock.yaml` outside `node_modules`.
2. Clean `pnpm install` from scratch succeeds.
3. `pnpm -r run typecheck`, `pnpm -r run build`, `pnpm -r run test` all green (i.e. `typecheck:docx-v2` / `build:docx-v2` / `test:docx-v2`).
4. `make test` green (web vitest via pnpm).
5. `pnpm --filter @metaldocs/web run dev` boots the app.
6. Grep over tracked files: no `npm ci`, no `npm -ws`, no `npm run`, no `npx` remain (corepack-enable lines excepted).
7. All three CI workflows (`ci.yml`, `api-contract.yml`, `e2e-coverage-gate.yml`) green on a real run after push.

## Verification Commands

- `git ls-files | grep -E 'package-lock.json|pnpm-lock.yaml'` → expect only `pnpm-lock.yaml` (root).
- `pnpm install` (from clean tree).
- `pnpm -r run typecheck && pnpm -r run build && pnpm -r run test`
- `make test`
- `pnpm --filter @metaldocs/web run dev` (boots, then stop).
- `git ls-files -z | xargs -0 grep -nE 'npm ci|npm -ws|npm run|npx ' -- ` (excluding this spec/plan) → expect empty.
- Post-push: GitHub Actions runs for all three workflows green.
