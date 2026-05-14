# Migration Baseline and Legacy Recovery Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish one trusted full-replay reference database, generate a baseline path for fresh local environments, and preserve the legacy migration chain as the official upgrade path for existing databases.

**Architecture:** This work is split into three phases. Phase 1 performs a clean legacy recovery and captures runtime/schema truth. Phase 2 creates a canonical baseline bootstrap path plus validation tooling. Phase 3 adds governance docs and archive policy without deleting historical migrations. The initial rollout is dual-path: legacy replay for upgrades, baseline bootstrap for fresh local installs.

**Tech Stack:** PowerShell, Postgres, `psql`, Docker Compose, SQL migrations, MetalDocs runtime gates, Markdown runbooks.

---

## Execution Rules

- Do not modify or delete historical migration files unless the task explicitly says to create an archive manifest or documentation entry.
- Follow ADR-0007 additive-first policy. This project changes bootstrap behavior and documentation first; it does not rewrite upgrade history.
- The local DB must be fully rebuilt and replayed once before any baseline SQL is generated.
- Keep production/shared upgrade flow pointed at the legacy chain for this first rollout.
- Use parallel workers only where file ownership is disjoint and the phase preconditions are already satisfied.

## Parallel Agent Strategy

Parallel-safe zones:

- Read-only inventory after the recovery gate passes.
- Baseline script/runbook work can run in parallel with migration classification documentation if files are disjoint.
- Validation tooling and archive/governance docs can run in parallel after the baseline bootstrap path exists.

Unsafe parallelism:

- Do not generate baseline SQL before the full legacy replay and runtime validation complete.
- Do not let two workers edit the same PowerShell bootstrap script.
- Do not let one worker change runtime bootstrap while another changes baseline validation commands in the same files.

Recommended worker split after Phase 1:

1. Worker A: baseline bootstrap scripts + baseline SQL artifact ownership
2. Worker B: validation/equivalence tooling ownership
3. Worker C: governance/archive docs ownership

Every worker prompt must include:

```text
You are not alone in this codebase. Do not revert or overwrite edits outside your owned files. Adapt to existing changes.
```

## Mandatory Preflight

- [ ] **Step 1: Record current git/worktree state**

Run:

```powershell
git status --short --branch
```

Expected: current branch and dirty state are understood; existing feature work is not reverted.

- [ ] **Step 2: Confirm Docker/Postgres local infra is available**

Run:

```powershell
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

Expected: local `metaldocs-postgres` container is visible or infra can be started by `scripts/dev-local.ps1`.

- [ ] **Step 3: Confirm baseline design doc exists**

Run:

```powershell
Test-Path docs/superpowers/specs/2026-05-14-migration-baseline-and-legacy-recovery-design.md
```

Expected: `True`

---

## File Map

### Create

- `scripts/dev-db-reset.ps1` - explicit local DB reset for trusted recovery
- `scripts/export-schema-baseline.ps1` - exports canonical baseline SQL from the trusted reference DB
- `scripts/dev-bootstrap-baseline.ps1` - fresh local bootstrap using baseline + tail
- `scripts/check-baseline-equivalence.ps1` - compares trusted reference DB vs baseline-built DB for runtime-used schema objects
- `migrations_baseline/0001_baseline_2026_05.sql` - canonical baseline schema artifact
- `docs/runbooks/migration-baseline-local.md` - operator runbook for local baseline bootstrap
- `docs/runbooks/migration-legacy-replay.md` - operator runbook for full legacy replay and recovery
- `docs/runbooks/migration-governance.md` - policy for baseline, legacy replay, tail migrations, and future archive cutovers
- `docs/runbooks/migration-archive-policy.md` - classification rules for historical migrations
- `non_git/db/reference-schema/` - ignored evidence directory for reference schema dumps and comparison artifacts

### Modify

- `.gitignore` - ignore `non_git/db/reference-schema/`
- `scripts/dev-local.ps1` - call out the dual-path bootstrap and point standard fresh-install flow to baseline path
- `scripts/dev-migrate.ps1` - keep legacy replay intact; add a mode guard or messaging clarifying that it is the full legacy path
- `docs/runbooks/dev-setup.md` - document default baseline bootstrap and explicit legacy replay mode
- `docs/adr/0007-schema-migration-policy.md` - add note linking to the new dual-path operational model
- `docs/tasks/NEXT_CYCLE_BACKLOG.md` - record follow-up archive/deletion work as deferred until post-validation

---

## Phase 1: Trusted Recovery and Reference Capture

### Task 1: Add Explicit Local DB Reset Script `[seq]`

**Files:**
- Create: `scripts/dev-db-reset.ps1`
- Modify: `.gitignore`

- [ ] **Step 1: Create ignored evidence directory entry**

Update `.gitignore` with:

```gitignore
non_git/db/reference-schema/
```

- [ ] **Step 2: Create reset script**

Create `scripts/dev-db-reset.ps1`:

```powershell
param(
  [string]$ComposeFile = "deploy/compose/docker-compose.yml",
  [string]$EnvFile = ".env"
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (-not (Test-Path $EnvFile)) {
  throw "$EnvFile not found. Copy .env.example to .env before resetting the local DB."
}

Write-Host "[dev-db-reset] Stopping app containers..."
docker compose -f $ComposeFile --env-file $EnvFile stop api web gateway worker | Out-Host

Write-Host "[dev-db-reset] Removing Postgres container and volume..."
docker compose -f $ComposeFile --env-file $EnvFile down -v postgres | Out-Host

Write-Host "[dev-db-reset] Starting infra containers again..."
docker compose -f $ComposeFile --env-file $EnvFile up -d postgres redis minio | Out-Host

Write-Host "[dev-db-reset] Local database reset complete."
```

- [ ] **Step 3: Run reset script**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-db-reset.ps1
```

Expected: Postgres is recreated from empty state; no migration replay has happened yet.

- [ ] **Step 4: Commit**

Run:

```powershell
git add .gitignore scripts/dev-db-reset.ps1
git commit -m "chore(dev): add explicit local db reset flow"
```

### Task 2: Trusted Legacy Replay Gate `[seq]`

**Files:**
- Modify: `docs/runbooks/migration-legacy-replay.md`

- [ ] **Step 1: Create legacy replay runbook**

Create `docs/runbooks/migration-legacy-replay.md`:

```markdown
# Runbook: Legacy Migration Replay

Use this runbook only when a full trusted historical replay is required.

## Commands

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-db-reset.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-migrate.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/start-api.ps1 -Build -NoWorker
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/controlled-documents
```
```

- [ ] **Step 2: Run full legacy replay**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-migrate.ps1
```

Expected: all SQL files in `migrations/` are applied without interruption.

- [ ] **Step 3: Run runtime validation gate**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/start-api.ps1 -Build -NoWorker
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/controlled-documents
```

Expected: login/auth/target-route checks pass.

- [ ] **Step 4: Record applied migration count**

Run:

```powershell
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "SELECT count(*) FROM schema_migrations;"
```

Expected: a concrete count is captured in the task notes and matches the reference DB state used for baseline generation.

- [ ] **Step 5: Commit runbook**

Run:

```powershell
git add docs/runbooks/migration-legacy-replay.md
git commit -m "docs(db): add legacy replay recovery runbook"
```

### Task 3: Post-Recovery Inventory Pack `[parallel-safe after Task 2]`

**Files:**
- Create: `non_git/db/reference-schema/README.md`

- [ ] **Step 1: Create evidence folder README**

Create `non_git/db/reference-schema/README.md`:

```markdown
# Reference Schema Evidence

This directory stores local-only schema dumps and comparison artifacts produced after a trusted legacy replay.

Do not commit generated SQL dumps from this directory.
```

- [ ] **Step 2: Run reference inventory commands**

Run:

```powershell
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "\dt"
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "\df"
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "\d public.templates_v2_template_version"
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "SELECT schemaname, tablename FROM pg_tables ORDER BY schemaname, tablename;"
```

Expected: reference DB object inventory is captured in the task notes or redirected to local evidence files during execution.

- [ ] **Step 3: Commit README**

Run:

```powershell
git add non_git/db/reference-schema/README.md
git commit -m "docs(db): add reference schema evidence directory"
```

---

## Phase 2: Baseline Path for Fresh Environments

### Task 4: Export Canonical Baseline SQL `[seq after Task 3]`

**Files:**
- Create: `scripts/export-schema-baseline.ps1`
- Create: `migrations_baseline/0001_baseline_2026_05.sql`

- [ ] **Step 1: Create export script**

Create `scripts/export-schema-baseline.ps1`:

```powershell
param(
  [string]$OutputFile = "migrations_baseline/0001_baseline_2026_05.sql"
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$outDir = Split-Path -Parent $OutputFile
if (-not (Test-Path $outDir)) {
  New-Item -ItemType Directory -Force -Path $outDir | Out-Null
}

docker exec metaldocs-postgres pg_dump `
  -U metaldocs_app `
  -d metaldocs `
  --schema-only `
  --no-owner `
  --no-privileges `
  > $OutputFile

Write-Host "[export-schema-baseline] Wrote baseline schema to $OutputFile"
```

- [ ] **Step 2: Run export script**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/export-schema-baseline.ps1
```

Expected: `migrations_baseline/0001_baseline_2026_05.sql` exists and is non-empty.

- [ ] **Step 3: Inspect baseline header**

Run:

```powershell
Get-Content migrations_baseline/0001_baseline_2026_05.sql -TotalCount 40
```

Expected: schema-only dump content is present; no data rows are included.

- [ ] **Step 4: Commit**

Run:

```powershell
git add scripts/export-schema-baseline.ps1 migrations_baseline/0001_baseline_2026_05.sql
git commit -m "feat(db): add canonical schema baseline export"
```

### Task 5: Baseline Bootstrap Script `[parallel-safe after Task 4; owns bootstrap scripts]`

**Files:**
- Create: `scripts/dev-bootstrap-baseline.ps1`
- Modify: `scripts/dev-local.ps1`
- Modify: `scripts/dev-migrate.ps1`

- [ ] **Step 1: Create baseline bootstrap script**

Create `scripts/dev-bootstrap-baseline.ps1`:

```powershell
param(
  [string]$ComposeFile = "deploy/compose/docker-compose.yml",
  [string]$EnvFile = ".env",
  [string]$BaselineFile = "migrations_baseline/0001_baseline_2026_05.sql"
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (-not (Test-Path $BaselineFile)) {
  throw "Baseline file not found: $BaselineFile"
}

powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-db-reset.ps1 -ComposeFile $ComposeFile -EnvFile $EnvFile | Out-Host

Write-Host "[dev-bootstrap-baseline] Applying baseline..."
docker compose -f $ComposeFile --env-file $EnvFile exec -T postgres `
  psql -v ON_ERROR_STOP=1 -U $env:POSTGRES_USER -d $env:POSTGRES_DB -f /workspace/$BaselineFile | Out-Host

Write-Host "[dev-bootstrap-baseline] Applying tail migrations (if any)..."
Write-Host "[dev-bootstrap-baseline] Phase 1 rollout keeps tail handling manual until cutoff is defined."
```

- [ ] **Step 2: Clarify legacy script intent**

At the top of `scripts/dev-migrate.ps1`, add:

```powershell
Write-Host "[dev-migrate] Legacy full replay mode: applying the complete migrations/ chain."
```

- [ ] **Step 3: Clarify default local bootstrap**

In `scripts/dev-local.ps1`, update the usage text to include:

```powershell
Write-Host "  0. Fresh install (preferred): powershell -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1"
Write-Host "     Legacy recovery: powershell -ExecutionPolicy Bypass -File scripts/dev-migrate.ps1"
```

- [ ] **Step 4: Commit**

Run:

```powershell
git add scripts/dev-bootstrap-baseline.ps1 scripts/dev-local.ps1 scripts/dev-migrate.ps1
git commit -m "feat(dev): add baseline bootstrap path"
```

### Task 6: Baseline Equivalence Validation Tooling `[parallel-safe after Task 4; owns validation tooling]`

**Files:**
- Create: `scripts/check-baseline-equivalence.ps1`

- [ ] **Step 1: Create validation script**

Create `scripts/check-baseline-equivalence.ps1`:

```powershell
param(
  [string]$ReferenceDb = "metaldocs",
  [string]$CandidateDb = "metaldocs"
)

$ErrorActionPreference = "Stop"

Write-Host "[check-baseline-equivalence] Comparing runtime-used schema objects..."

docker exec metaldocs-postgres psql -U metaldocs_app -d $ReferenceDb -tAc `
  "SELECT table_schema, table_name, column_name, data_type FROM information_schema.columns ORDER BY 1,2,3;" `
  > non_git/db/reference-schema/reference-columns.txt

docker exec metaldocs-postgres psql -U metaldocs_app -d $CandidateDb -tAc `
  "SELECT table_schema, table_name, column_name, data_type FROM information_schema.columns ORDER BY 1,2,3;" `
  > non_git/db/reference-schema/candidate-columns.txt

fc.exe non_git\db\reference-schema\reference-columns.txt non_git\db\reference-schema\candidate-columns.txt
```

- [ ] **Step 2: Run validation script**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-baseline-equivalence.ps1
```

Expected: no column-level drift for the compared DBs, or a concrete diff artifact to review.

- [ ] **Step 3: Commit**

Run:

```powershell
git add scripts/check-baseline-equivalence.ps1
git commit -m "feat(db): add baseline equivalence validation tooling"
```

### Task 7: Baseline Local Runbook `[seq after Tasks 5-6]`

**Files:**
- Create: `docs/runbooks/migration-baseline-local.md`
- Modify: `docs/runbooks/dev-setup.md`

- [ ] **Step 1: Create baseline runbook**

Create `docs/runbooks/migration-baseline-local.md`:

```markdown
# Runbook: Baseline Local Bootstrap

Use this runbook for fresh local environments after the baseline file has been generated and validated.

## Commands

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/start-api.ps1 -Build -NoWorker
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/controlled-documents
```
```

- [ ] **Step 2: Update dev setup runbook**

In `docs/runbooks/dev-setup.md`, add a short section:

```markdown
## Baseline bootstrap (preferred for fresh local setups)

Use `scripts/dev-bootstrap-baseline.ps1` for fresh local environments.

Use `scripts/dev-migrate.ps1` only when full historical replay is required for recovery or migration debugging.
```

- [ ] **Step 3: Commit**

Run:

```powershell
git add docs/runbooks/migration-baseline-local.md docs/runbooks/dev-setup.md
git commit -m "docs(dev): add baseline bootstrap runbook"
```

---

## Phase 3: Governance Hardening and Archive Policy

### Task 8: Migration Governance Runbook `[parallel-safe after Task 7; owns docs only]`

**Files:**
- Create: `docs/runbooks/migration-governance.md`
- Modify: `docs/adr/0007-schema-migration-policy.md`

- [ ] **Step 1: Create governance runbook**

Create `docs/runbooks/migration-governance.md`:

```markdown
# Runbook: Migration Governance

## Current model

- `migrations/` remains the official upgrade path for existing databases.
- `migrations_baseline/` is the canonical fresh-install path for new local environments.
- New schema changes after the baseline cutoff remain forward-only.

## Rules

1. Do not delete historical migrations during the first baseline rollout.
2. Generate a new baseline only from a trusted fully replayed reference DB.
3. Validate baseline-built DBs against runtime gates before changing default bootstrap behavior.
4. Archive policy applies only after validation and explicit review.
```

- [ ] **Step 2: Link ADR-0007 to the dual-path model**

Append to `docs/adr/0007-schema-migration-policy.md`:

```markdown
## Operational note (2026-05 baseline rollout)

MetalDocs temporarily operates a dual-path migration model:
- legacy chronological replay for existing database upgrades
- canonical baseline bootstrap for fresh local environments

This preserves additive-first upgrade safety while reducing bootstrap cost for new local installs.
```

- [ ] **Step 3: Commit**

Run:

```powershell
git add docs/runbooks/migration-governance.md docs/adr/0007-schema-migration-policy.md
git commit -m "docs(db): define migration governance model"
```

### Task 9: Archive Policy and Deferred Follow-ups `[parallel-safe after Task 7; owns docs only]`

**Files:**
- Create: `docs/runbooks/migration-archive-policy.md`
- Modify: `docs/tasks/NEXT_CYCLE_BACKLOG.md`

- [ ] **Step 1: Create archive policy**

Create `docs/runbooks/migration-archive-policy.md`:

```markdown
# Runbook: Migration Archive Policy

Historical migrations may be classified into:

- required for upgrade history
- historical-only
- destructive/dev-reset artifacts

No file moves or deletions happen until:

1. baseline bootstrap is validated
2. legacy replay remains available
3. archive candidates are reviewed explicitly
```

- [ ] **Step 2: Add backlog follow-up**

Append to `docs/tasks/NEXT_CYCLE_BACKLOG.md`:

```markdown
## Task - Migration archive follow-up

- classify `migrations/` into upgrade-critical vs historical-only
- identify destructive or dev-reset files that should leave the primary operational path
- move only after baseline bootstrap has been validated against runtime truth
```

- [ ] **Step 3: Commit**

Run:

```powershell
git add docs/runbooks/migration-archive-policy.md docs/tasks/NEXT_CYCLE_BACKLOG.md
git commit -m "docs(db): add migration archive policy follow-up"
```

---

## Final Verification

### Task 10: Baseline Program Verification `[seq final]`

**Files:** none unless verification exposes a bug in task-owned files.

- [ ] **Step 1: Re-run trusted legacy gate**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-db-reset.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-migrate.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/controlled-documents
```

Expected: full replay and runtime gate pass.

- [ ] **Step 2: Build fresh baseline environment**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/controlled-documents
```

Expected: fresh baseline bootstrap succeeds and runtime gate passes.

- [ ] **Step 3: Run contract/runtime checks**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module documents
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module registry
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module templates
```

Expected: all pass.

- [ ] **Step 4: Run equivalence validation**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-baseline-equivalence.ps1
```

Expected: no unexpected drift, or concrete diff files captured for follow-up.

- [ ] **Step 5: Scope check**

Run:

```powershell
git status --short
git diff --stat HEAD~10..HEAD
```

Expected: changes are limited to scripts, baseline SQL, docs/runbooks, and bootstrap/governance files.

## Execution Handoff

Recommended execution option: **Subagent-Driven with recovery-first gating**.

Run Task 1 and Task 2 sequentially. After Task 2 passes, split into parallel workers:

- Worker A: Task 4 + Task 5
- Worker B: Task 6
- Worker C: Task 8 + Task 9

Keep Task 7 and Task 10 with the coordinator because they integrate the new paths and perform the final gates.

