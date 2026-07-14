# Phase 4 — Single pnpm Workspace Consolidation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Collapse the repo's dual npm+pnpm setup into ONE unified pnpm workspace — one lockfile, one `node_modules` store, `pnpm -r` orchestration, all CI on pnpm.

**Architecture:** Add a root `pnpm-workspace.yaml` that makes `frontend/apps/web`, `packages/*`, `apps/*`, and `shared/*` members of a single workspace. Internal `@metaldocs/*` deps become `workspace:*`. All npm lockfiles + the stray root pnpm stub are deleted and regenerated as one root `pnpm-lock.yaml` from a clean install (which also root-causes the recurring junction/version drift). Local scripts and all three CI workflows are converted to pnpm.

**Tech Stack:** pnpm 9 (via corepack), Node 20.11.0, TypeScript, Vite, Vitest, Playwright, GitHub Actions. Governing spec: `docs/superpowers/specs/2026-06-23-pnpm-workspace-consolidation-design.md`.

---

## Conventions & Cautions (read before starting)

- **Caveman/PowerShell environment.** Primary shell is PowerShell on Windows; a Bash tool is also available. Commands below are given in a shell-neutral form — translate to the tool you use. Use `pnpm`, never `npm`/`yarn`.
- **No version bumps.** This is a pure consolidation. Do not upgrade any dependency. If a version conflict surfaces during install, STOP and report it — do not "fix" it by bumping.
- **Destructive steps are gated.** Task 6 deletes lockfiles and all `node_modules`. Commit the preceding tasks first so the working tree is recoverable.
- **Push requires explicit permission.** CI can only be verified on a real run (Task 13). Per project rules, do NOT push until the user explicitly authorizes it. Commit freely; push only on the user's go-ahead.
- **The web package name is `@metaldocs/web`.** All `--filter` targets use that exact name.
- **Verification-driven, not TDD.** This is infra config; most tasks verify via a command + expected output rather than a unit test. Where a guard can be expressed as a check, it is.

---

## File Structure

**Created:**
- `pnpm-workspace.yaml` — workspace member globs.

**Modified:**
- `package.json` (root) — drop `workspaces`, add `packageManager`, swap script impls to `pnpm -r`.
- `frontend/apps/web/package.json` — `file:` deps → `workspace:*`; fix `build` script.
- Any other member `package.json` with intra-repo `file:`/path deps (audited in Task 1).
- `Makefile` — `npx vitest` → `pnpm --filter @metaldocs/web exec vitest`.
- `scripts/dev-api-web.ps1`, `scripts/run_metaldocs.ps1`, `scripts/dev-local.ps1` — `npm run dev` → `pnpm`.
- `.github/workflows/ci.yml`, `.github/workflows/api-contract.yml`, `.github/workflows/e2e-coverage-gate.yml`.

**Deleted:**
- `package-lock.json` (root), `pnpm-lock.yaml` (root 114-byte stub), `frontend/apps/web/pnpm-lock.yaml`, `packages/editor-ui/pnpm-lock.yaml`, `shared/schemas/package-lock.json`, plus any others found in Task 1.

**Regenerated:**
- `pnpm-lock.yaml` (root) — the single lockfile.

---

## Task 1: Pre-flight audit + pin pnpm version

**Files:** none modified (read-only discovery + corepack).

- [ ] **Step 1: Enable corepack and pin the exact pnpm 9 version**

Run:
```
corepack enable
corepack prepare pnpm@latest-9 --activate
pnpm --version
```
Record the printed version (e.g. `9.15.9`). Call it `<PNPM_VERSION>` for the rest of the plan — use the REAL printed value, never a placeholder.

- [ ] **Step 2: Enumerate every lockfile outside node_modules**

Run:
```
git ls-files | grep -E 'package-lock\.json$|pnpm-lock\.yaml$|yarn\.lock$|npm-shrinkwrap\.json$'
```
Expected to include at least: `package-lock.json`, `pnpm-lock.yaml`, `frontend/apps/web/pnpm-lock.yaml`, `packages/editor-ui/pnpm-lock.yaml`, `shared/schemas/package-lock.json`. Record the FULL list — every entry here gets deleted in Task 6.

- [ ] **Step 3: Enumerate every intra-repo dependency expressed as a `file:` path**

Run:
```
git ls-files '*package.json' | grep -v node_modules | xargs grep -nE '"@metaldocs/[^"]+": *"(file:|\.\.)' 
```
Expected known hits: `frontend/apps/web/package.json` → `@metaldocs/editor-ui` and `@metaldocs/shared-tokens`. Record any ADDITIONAL hits (other members linking each other) — each gets converted in Task 4.

- [ ] **Step 4: List all workspace member package.json + their `name` fields**

Run:
```
git ls-files '*package.json' | grep -v node_modules | xargs grep -H '"name"' | grep -v '"name": *"metaldocs-monorepo"'
```
Confirm members match the glob set (`packages/*`, `apps/docx-renderer`, `shared/schemas`, `shared/mddm-layout-tokens`, `shared/mddm-pagination-types`, `frontend/apps/web`). Note any member with NO buildable scripts (expected: the two `mddm-*` shared packages) — `pnpm -r` skips them safely.

- [ ] **Step 5: Record findings**

No commit (read-only). Carry the recorded lists (lockfiles, file: deps, members, `<PNPM_VERSION>`) into the following tasks.

---

## Task 2: Create the workspace manifest

**Files:**
- Create: `pnpm-workspace.yaml`

- [ ] **Step 1: Write `pnpm-workspace.yaml`**

```yaml
packages:
  - 'frontend/apps/web'
  - 'packages/*'
  - 'apps/*'
  - 'shared/*'
```

- [ ] **Step 2: Commit**

```bash
git add pnpm-workspace.yaml
git commit -m "build(workspace): add pnpm-workspace.yaml defining unified members"
```

---

## Task 3: Convert root package.json to pnpm orchestration

**Files:**
- Modify: `package.json` (root)

- [ ] **Step 1: Edit root `package.json`**

Remove the `workspaces` array (lines 5-8). Add a `packageManager` field with the REAL version from Task 1. Swap the three script implementations. Result:

```json
{
  "name": "metaldocs-monorepo",
  "private": true,
  "version": "0.0.0",
  "packageManager": "pnpm@<PNPM_VERSION>",
  "scripts": {
    "build:docx-v2": "pnpm -r run build",
    "test:docx-v2": "pnpm -r run test",
    "typecheck:docx-v2": "pnpm -r run typecheck"
  },
  "engines": {
    "node": ">=20.11.0"
  }
}
```

> `pnpm -r run <script>` already skips members lacking the script, preserving the old `--if-present` behavior. Script NAMES are unchanged because CI and `CLAUDE.md` reference them.

- [ ] **Step 2: Commit**

```bash
git add package.json
git commit -m "build(workspace): root scripts to pnpm -r, pin packageManager, drop npm workspaces"
```

---

## Task 4: Convert internal deps to `workspace:*`

**Files:**
- Modify: `frontend/apps/web/package.json` (and any additional members found in Task 1, Step 3)

- [ ] **Step 1: Edit `frontend/apps/web/package.json` dependencies**

Change:
```json
    "@metaldocs/editor-ui": "file:../../../packages/editor-ui",
    "@metaldocs/shared-tokens": "file:../../../packages/shared-tokens",
```
to:
```json
    "@metaldocs/editor-ui": "workspace:*",
    "@metaldocs/shared-tokens": "workspace:*",
```

- [ ] **Step 2: Convert any other intra-repo `file:` deps found in Task 1**

For EACH additional hit from Task 1 Step 3, change its `"file:..."` (or relative path) value to `"workspace:*"`. If there were none, skip.

- [ ] **Step 3: Verify no `file:` intra-repo deps remain**

Run:
```
git ls-files '*package.json' | grep -v node_modules | xargs grep -nE '"@metaldocs/[^"]+": *"(file:|\.\.)' 
```
Expected: empty output.

- [ ] **Step 4: Commit**

```bash
git add -A '*package.json'
git commit -m "build(workspace): link internal @metaldocs deps via workspace:* protocol"
```

---

## Task 5: Fix the web build script for cross-platform / workspace use

**Files:**
- Modify: `frontend/apps/web/package.json`

- [ ] **Step 1: Replace the Windows-only `pnpm.cmd` invocation**

The `build` script currently is:
```json
    "build": "pnpm.cmd tsc --noEmit -p tsconfig.build.json && vite build",
```
`pnpm.cmd` is a Windows-only shim and will fail when `pnpm -r run build` invokes it on Linux CI. Scripts run with `node_modules/.bin` already on PATH, so call `tsc` directly:
```json
    "build": "tsc --noEmit -p tsconfig.build.json && vite build",
```

- [ ] **Step 2: Commit**

```bash
git add frontend/apps/web/package.json
git commit -m "build(web): call tsc directly in build script (drop Windows-only pnpm.cmd shim)"
```

---

## Task 6: Delete all lockfiles + node_modules; clean install

**Files:**
- Delete: every lockfile from Task 1 Step 2.

> Tasks 2-5 must be committed before this step — the tree should be clean apart from the deletions/regeneration here.

- [ ] **Step 1: Delete all old lockfiles**

Remove each path recorded in Task 1 Step 2. At minimum:
```
git rm -f package-lock.json pnpm-lock.yaml frontend/apps/web/pnpm-lock.yaml packages/editor-ui/pnpm-lock.yaml shared/schemas/package-lock.json
```
(Add any additional lockfiles from the Task 1 list.)

- [ ] **Step 2: Nuke all node_modules**

PowerShell:
```
Get-ChildItem -Path . -Recurse -Directory -Filter node_modules -ErrorAction SilentlyContinue | Where-Object { $_.FullName -notlike '*\.claude\worktrees\*' } | Remove-Item -Recurse -Force
```
(Or the Bash equivalent. Do NOT touch `.claude/worktrees`.)

- [ ] **Step 3: Clean install from scratch**

Run:
```
pnpm install
```
Expected: completes without error and creates a single root `pnpm-lock.yaml`. If it fails on a version conflict, STOP and report (do not bump versions).

- [ ] **Step 4: Verify exactly one lockfile exists**

Run:
```
git status --porcelain
git ls-files --others --exclude-standard | grep -E 'lock' ; ls pnpm-lock.yaml
```
Then stage and confirm only the root lock is tracked:
```
git add pnpm-lock.yaml
git ls-files | grep -E 'package-lock\.json$|pnpm-lock\.yaml$'
```
Expected final grep output: exactly `pnpm-lock.yaml` (root) and nothing else.

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "build(workspace): consolidate to single root pnpm-lock.yaml; remove npm + stray locks"
```

---

## Task 7: Local build/typecheck/test gates

**Files:** none (verification; `.npmrc` only if fallback needed).

- [ ] **Step 1: Typecheck all members**

Run:
```
pnpm -r run typecheck
```
Expected: PASS for every member that defines `typecheck`.

- [ ] **Step 2: Build all members**

Run:
```
pnpm -r run build
```
Expected: PASS; pnpm builds in topological order (shared/adapter/renderer before web).

- [ ] **Step 3: Test all members**

Run:
```
pnpm -r run test
```
Expected: PASS. NOTE: `shared/schemas` was previously a standalone npm project; its `test` now runs under the workspace. If it surfaces a pre-existing failure unrelated to packaging, record it and report — do not silently skip.

- [ ] **Step 4: Fallback ONLY if vite/vitest resolution breaks**

If Step 1-3 fail specifically with module-resolution errors under pnpm's symlinked `node_modules` (the documented junction-drift failure mode), create root `.npmrc`:
```
shamefully-hoist=true
```
Then re-run `pnpm install` and repeat Steps 1-3. Only do this if the symlinked layout genuinely fails — do not add it pre-emptively.

- [ ] **Step 5: Commit (only if `.npmrc` was added or lockfile changed)**

```bash
git add -A
git commit -m "build(workspace): hoist node_modules for vite/vitest resolution"
```
(If nothing changed in this task, skip the commit.)

---

## Task 8: Convert local scripts to pnpm

**Files:**
- Modify: `Makefile`, `scripts/dev-api-web.ps1`, `scripts/run_metaldocs.ps1`, `scripts/dev-local.ps1`

- [ ] **Step 1: Edit `Makefile`**

Replace the `test` and `test-watch` recipes (keep the explanatory comment block, update its `npx vitest run` reference too):
```make
test:
	pnpm --filter @metaldocs/web exec vitest run

test-watch:
	pnpm --filter @metaldocs/web exec vitest
```

- [ ] **Step 2: Edit the dev scripts**

In `scripts/dev-api-web.ps1` and `scripts/run_metaldocs.ps1`, replace the `npm run dev` invocation (run inside `frontend/apps/web`) with the workspace-filtered form so it works from repo root:
```
pnpm --filter @metaldocs/web run dev
```
In `scripts/dev-local.ps1`, update the `npm run dev` reference (comment/doc line) to the same `pnpm` form. Preserve each script's surrounding behavior (cwd handling, env loading) — change only the package-manager invocation.

- [ ] **Step 3: Verify no npm/npx remain in these files**

Run:
```
grep -nE 'npm |npx ' Makefile scripts/dev-api-web.ps1 scripts/run_metaldocs.ps1 scripts/dev-local.ps1
```
Expected: empty.

- [ ] **Step 4: Commit**

```bash
git add Makefile scripts/dev-api-web.ps1 scripts/run_metaldocs.ps1 scripts/dev-local.ps1
git commit -m "build(scripts): convert Makefile + dev scripts to pnpm"
```

---

## Task 9: Verify the app boots + make test green

**Files:** none (verification).

- [ ] **Step 1: `make test`**

Run: `make test`
Expected: vitest runs from the web package via pnpm and passes (same suite as before consolidation).

- [ ] **Step 2: Dev server boots**

Run: `pnpm --filter @metaldocs/web run dev`
Expected: Vite starts and serves without module-resolution errors. Stop it after confirming startup.

- [ ] **Step 3: No commit** (verification only).

---

## Task 10: Migrate `ci.yml` to pnpm

**Files:**
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Rewrite the `node` job steps**

Replace the `node` job's steps (lines 19-26) with the pnpm equivalent. Add `pnpm/action-setup` BEFORE `setup-node`, switch the cache, and use `pnpm install --frozen-lockfile`:
```yaml
  node:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v3
        with: { version: 9 }
      - uses: actions/setup-node@v4
        with: { node-version: 20.11.0, cache: pnpm }
      - run: pnpm install --frozen-lockfile
      - run: pnpm run typecheck:docx-v2
      - run: pnpm run test:docx-v2
      - run: pnpm run build:docx-v2
```
Leave the `go` job untouched.

- [ ] **Step 2: Sanity-check YAML**

Run (if available): `npx --yes yaml-lint .github/workflows/ci.yml` OR visually confirm indentation. (Real validation happens on the run in Task 13.)

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: run docx-v2 node job on pnpm workspace"
```

---

## Task 11: Migrate `api-contract.yml` to pnpm

**Files:**
- Modify: `.github/workflows/api-contract.yml`

- [ ] **Step 1: Update the path triggers**

In the `on.pull_request.paths` list, replace the now-removed lockfile path:
```yaml
      - 'frontend/apps/web/package-lock.json'
```
with the root lockfile:
```yaml
      - 'pnpm-lock.yaml'
```

- [ ] **Step 2: Rewrite the `frontend-codegen-drift` job**

Replace its steps (lines 38-48) so it installs the workspace at root via pnpm and runs `gen:api` through the filter:
```yaml
  frontend-codegen-drift:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v3
        with: { version: 9 }
      - uses: actions/setup-node@v4
        with: { node-version: 20.11.0, cache: pnpm }
      - run: pnpm install --frozen-lockfile
      - name: API codegen drift check (frontend)
        run: |
          pnpm --filter @metaldocs/web run gen:api
          git diff --exit-code -- frontend/apps/web/src/lib/api-types/ || (echo "::error::Run 'pnpm --filter @metaldocs/web run gen:api' and commit"; exit 1)
```

> The `openapi-lint` job uses `npx --yes @redocly/cli` with no install — leave it as-is (it is a one-off tool fetch, not a package-manager dependency). It does not block the one-PM goal.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/api-contract.yml
git commit -m "ci: frontend codegen-drift job on pnpm workspace; fix lockfile trigger path"
```

---

## Task 12: Migrate `e2e-coverage-gate.yml` to pnpm workspace

**Files:**
- Modify: `.github/workflows/e2e-coverage-gate.yml`

- [ ] **Step 1: Fix install wiring in the `e2e-smoke` job**

Replace the cache path + install step (lines 121-129) so it uses the root lockfile and installs the whole workspace at repo root:
```yaml
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'pnpm'
          cache-dependency-path: pnpm-lock.yaml

      - name: Install workspace deps
        run: pnpm install --frozen-lockfile
```
(Remove the `working-directory: frontend` from the install step — it installs at repo root now.)

- [ ] **Step 2: Keep the Playwright + test steps scoped to the web app**

The `Install Playwright browsers` (line 131-133) and `Run E2E approval flows` (line 150-156) steps keep `working-directory: frontend/apps/web` and their `pnpm exec playwright ...` commands — those resolve correctly now that the workspace is installed. No change needed beyond confirming they still read.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/e2e-coverage-gate.yml
git commit -m "ci: e2e job installs root pnpm workspace; fix stale lockfile cache path"
```

---

## Task 13: Final verification + CI watch

**Files:** none (verification + gated push).

- [ ] **Step 1: Definition-of-Done greps**

Run:
```
git ls-files | grep -E 'package-lock\.json$|pnpm-lock\.yaml$'
```
Expected: only `pnpm-lock.yaml` (root).

Run:
```
git ls-files | grep -v -E 'docs/superpowers/(specs|plans)/2026-06-23-pnpm' | xargs grep -nE 'npm ci|npm -ws|npm run|npx ' 2>/dev/null
```
Expected: empty EXCEPT the intentional `npx --yes @redocly/cli` line in `api-contract.yml` (one-off tool fetch). Confirm that is the only hit; anything else is a leak to fix.

- [ ] **Step 2: Full local gate sweep**

Run, expecting all green:
```
pnpm -r run typecheck
pnpm -r run build
pnpm -r run test
make test
```

- [ ] **Step 3: Request push permission**

STOP and ask the user for explicit permission to push (project rule: never push without asking). Do not proceed to Step 4 without it.

- [ ] **Step 4: Push and watch the real CI runs**

After authorization, push the branch and open a PR (or push to the PR branch). Watch all three workflows:
- `docx-renderer CI` (`ci.yml`) — node job green on pnpm.
- `api-contract` (`api-contract.yml`) — frontend-codegen-drift green on pnpm.
- `E2E Coverage Gate` (`e2e-coverage-gate.yml`) — e2e-smoke installs + runs green.

Use `gh run watch` / `gh run list` to confirm. If any workflow fails, debug from the run logs (systematic-debugging), fix, recommit, re-watch.

- [ ] **Step 5: Report closure**

Report: commands run + outcomes, the two DoD greps, all three CI run conclusions, and confirm exactly one lockfile remains. Phase 4 done when all local gates green, all three CI runs green, and both greps clean.

---

## Self-Review Notes (author checklist — already applied)

- **Spec coverage:** every spec section maps to a task — workspace config (T2/T3), internal link (T4), build-script fix surfaced during file read (T5, not in spec but required for the spec's `pnpm -r build` DoD on Linux CI), lockfile consolidation (T6), local scripts (T8), CI ×3 (T10/T11/T12), DoD greps + CI watch (T13). The `.npmrc` fallback risk is T7 Step 4.
- **Placeholder scan:** `<PNPM_VERSION>` is the only token and is explicitly derived (Task 1 Step 1) from real `pnpm --version` output, not left as a literal — the worker substitutes the printed value. No TODO/TBD.
- **Type/name consistency:** `@metaldocs/web` filter target used identically across T8/T9/T11/T13. `pnpm -r run {build,test,typecheck}` script names match the root scripts defined in T3 and the spec DoD. Lockfile delete list in T6 matches the enumeration in T1.
- **Destructive + push gating** explicitly called out (T6 prerequisite, T13 Step 3 push permission) per project rules.
