# Curated Database Baseline and Dictionary Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace MetalDocs' fragile historical fresh-install DB bootstrap with a curated current-state baseline, separated product reference data/dev seeds, a trustworthy post-baseline ledger, and a wiki database dictionary.

**Architecture:** Work starts with parallel read-only research so the baseline is derived from runtime truth, migration archaeology, live schema, seed classification, and governance requirements. Implementation then proceeds through narrow ownership lanes: DB artifact layout, bootstrap scripts, startup migration policy, dictionary/wiki docs, skills/AGENTS guidance, and verification gates. Historical migrations remain available as legacy evidence and upgrade/debug material until a separate archive decision is approved.

**Tech Stack:** PowerShell, PostgreSQL 16, Docker Compose, SQL, Go migration runner, Markdown wiki/runbooks, MetalDocs skill system.

---

## Required Skills and Rules

- Use `runtime-contract-prereq` for the whole effort because DB bootstrap and startup truth are unreliable.
- Use `superpowers:subagent-driven-development` if executing this plan with subagents.
- Use `superpowers:verification-before-completion` before claiming the migration/bootstrap work is complete.
- Use `skill-creator` before creating or updating `.claude/skills/metaldocs-database/SKILL.md` or `.agents/skills/metaldocs-database/SKILL.md`.
- Do not continue feature work while the DB prerequisite remains failing.
- Do not patch historical migrations just to get startup passing unless a task explicitly says to preserve the legacy replay path.
- Do not create fake runtime data or mocked backend behavior.
- Do not delete or move historical migrations in this plan.

## Success Gates

The implementation is successful only when:

- Fresh DB bootstrap starts from empty Postgres without Docker entrypoint auto-running `migrations/`.
- Curated baseline, product reference data, and optional dev seeds apply through explicit scripts.
- API startup succeeds against the curated path.
- Auth/session runtime gate passes.
- Representative module route checks pass.
- Legacy replay remains intentionally runnable for evidence/debugging.
- `schema_migrations` records a baseline marker and post-baseline migrations consistently.
- Every table included in the curated baseline has a wiki dictionary entry or an explicit exception.
- `AGENTS.md`, `CLAUDE.md`, and MetalDocs skills point to canonical DB workflow rules without duplicating them.

## Parallel Agent Strategy

### Parallel-safe phase: read-only research

Run these agents at the same time. They must not edit files.

1. Runtime Usage Agent
   - Read-only ownership: Go code, SQL strings, repositories, startup/bootstrap code.
   - Output file after coordinator integration: `docs/db-research/runtime-usage-inventory.md`.

2. Migration Archaeology Agent
   - Read-only ownership: `migrations/*.sql`, `migrations_baseline/*.sql`, existing migration docs.
   - Output file after coordinator integration: `docs/db-research/migration-archaeology.md`.

3. Live Schema Agent
   - Read-only ownership: live Postgres catalog inspection after trusted legacy replay.
   - Output file after coordinator integration: `docs/db-research/live-schema-inventory.md`.

4. Seed/Auth Agent
   - Read-only ownership: auth/IAM/capability/system-template/dev seed classification.
   - Output file after coordinator integration: `docs/db-research/seed-reference-data-classification.md`.

5. Governance Agent
   - Read-only ownership: `AGENTS.md`, `CLAUDE.md`, `.agents/skills`, `.claude/skills`, runbooks, wiki structure.
   - Output file after coordinator integration: `docs/db-research/governance-update-plan.md`.

Every agent prompt must include:

```text
You are not alone in this codebase. Do not edit files. Do not revert or overwrite edits outside your owned research area. Return concise evidence with exact file paths, line references when available, commands used, and unresolved questions.
```

### Parallel-safe implementation lanes after research approval

After the research package and baseline inclusion catalog are approved, implementation can split into disjoint write lanes:

1. DB artifact lane
   - Owns `db/prerequisites/`, `db/baseline/`, `db/reference-data/`, `db/dev-seeds/`, `db/migrations/`.

2. Script/runtime lane
   - Owns `deploy/compose/docker-compose.yml`, `scripts/dev-db-reset.ps1`, `scripts/dev-bootstrap-baseline.ps1`, `scripts/dev-migrate.ps1`, `scripts/dev-local.ps1`, `scripts/check-baseline-equivalence.ps1`, and migration runner changes.

3. Dictionary/wiki lane
   - Owns `wiki/database/**` and dictionary coverage artifacts.

4. Governance/skills lane
   - Owns `AGENTS.md`, `CLAUDE.md`, `.agents/skills/metaldocs-database/`, `.claude/skills/metaldocs-database/`, and concise references from existing skills.

5. Verification lane
   - Owns test/check scripts under `scripts/` that are not already owned by the script/runtime lane, plus runbook verification command updates.

Do not run implementation lanes before Tasks 1-9 are complete.

## File Map

### Create

- `docs/db-research/runtime-usage-inventory.md` - runtime SQL/table/function usage evidence.
- `docs/db-research/migration-archaeology.md` - classification of historical migrations.
- `docs/db-research/live-schema-inventory.md` - live DB object inventory.
- `docs/db-research/seed-reference-data-classification.md` - product reference vs dev/test seed classification.
- `docs/db-research/governance-update-plan.md` - concise DB workflow guidance changes.
- `docs/db-research/curated-baseline-inclusion-catalog.md` - final inclusion/exclusion catalog for baseline writing.
- `db/prerequisites/0001_extensions.sql` - required extensions and schema bootstrap.
- `db/baseline/0001_current_schema.sql` - curated current-state schema.
- `db/reference-data/0001_product_reference_data.sql` - required product reference data.
- `db/dev-seeds/0001_local_dev_seed.sql` - local-only dev users/tenant/workflow convenience data.
- `db/migrations/README.md` - post-baseline forward migration rules.
- `.claude/skills/metaldocs-database/SKILL.md` - canonical DB workflow skill.
- `.agents/skills/metaldocs-database/SKILL.md` - Codex bridge to canonical skill.
- `wiki/database/overview.md`
- `wiki/database/schemas.md`
- `wiki/database/relationships.md`
- `wiki/database/reference-data.md`
- `wiki/database/migration-policy.md`
- `wiki/database/dictionary-index.md`
- `wiki/database/tables/*.md` - one table dictionary entry per curated table.
- `docs/runbooks/database-bootstrap.md` - fresh baseline, legacy replay, seed handling, verification.
- `scripts/check-db-dictionary-coverage.ps1` - verifies baseline tables have dictionary entries.
- `scripts/check-db-bootstrap.ps1` - orchestrates fresh baseline bootstrap verification.

### Modify

- `deploy/compose/docker-compose.yml` - remove historical migration mount from Postgres entrypoint.
- `scripts/dev-db-reset.ps1` - keep reset behavior aligned with empty Postgres.
- `scripts/dev-bootstrap-baseline.ps1` - apply curated baseline artifacts instead of skipping after entrypoint replay.
- `scripts/dev-migrate.ps1` - make legacy replay explicit and independent from entrypoint paths.
- `scripts/dev-local.ps1` - call baseline bootstrap for normal fresh local setup.
- `scripts/check-baseline-equivalence.ps1` - compare broader runtime-used objects.
- `internal/platform/migrate/migrate.go` - make post-baseline ledger behavior explicit and testable.
- `apps/api/cmd/metaldocs-api/main.go` - keep startup migrations scoped to the policy chosen by this plan.
- `internal/modules/registry/module.go` - decide whether durable startup backfill remains runtime job or moves into governed DB artifacts.
- `AGENTS.md` - add concise database workflow pointer.
- `CLAUDE.md` - mirror concise database workflow pointer.
- `docs/runbooks/migration-governance.md` - replace raw baseline wording with curated baseline policy.
- `docs/runbooks/migration-baseline-local.md` - point to curated baseline bootstrap.
- `docs/runbooks/migration-legacy-replay.md` - preserve intentional legacy replay path.
- `docs/adr/0007-schema-migration-policy.md` - link to curated baseline/database dictionary model.

---

## Phase 0: Preflight and Branch Hygiene

### Task 1: Confirm Starting State

**Files:** none

- [ ] **Step 1: Confirm branch and cleanliness**

Run:

```powershell
git status --short --branch
```

Expected:

```text
## main...origin/main [ahead N]
```

There must be no unstaged edits before research starts.

- [ ] **Step 2: Confirm design spec exists**

Run:

```powershell
Test-Path docs/superpowers/specs/2026-05-14-curated-database-baseline-and-dictionary-design.md
```

Expected:

```text
True
```

- [ ] **Step 3: Confirm DB-related working files from the previous checkpoint exist**

Run:

```powershell
Test-Path scripts/dev-db-reset.ps1
Test-Path scripts/dev-bootstrap-baseline.ps1
Test-Path migrations_baseline/0001_baseline_2026_05.sql
Test-Path deploy/compose/docker-compose.yml
```

Expected: four `True` lines.

### Task 2: Create Research Output Directory

**Files:**
- Create: `docs/db-research/README.md`

- [ ] **Step 1: Create research README**

Create `docs/db-research/README.md`:

```markdown
# Database Research Evidence

This directory stores reviewed evidence for the curated database baseline effort.

Rules:

- Read-only research findings are recorded here before baseline implementation.
- Evidence must cite runtime files, migration files, live schema commands, or wiki/runbook sources.
- Do not place raw database dumps in this directory.
- Do not treat these files as final database policy; canonical policy lives in `wiki/database/` after implementation.
```

- [ ] **Step 2: Commit research directory marker**

Run:

```powershell
git add docs/db-research/README.md
git commit -m "docs(db): add database research evidence directory"
```

Expected: commit succeeds.

---

## Phase 1: Parallel Read-Only Research

### Task 3: Runtime Usage Inventory `[parallel-safe, read-only]`

**Files:**
- Create: `docs/db-research/runtime-usage-inventory.md`

- [ ] **Step 1: Search runtime SQL usage**

Run:

```powershell
rg -n "FROM |JOIN |INSERT INTO|UPDATE |DELETE FROM|CALL |SELECT .* FROM|ExecContext|QueryContext|QueryRowContext" apps internal -S > non_git/runtime-sql-usage.txt
```

Expected: command completes and writes `non_git/runtime-sql-usage.txt`.

- [ ] **Step 2: Search startup and migration mutation paths**

Run:

```powershell
rg -n "schema_migrations|migrate.Apply|RunStartupMigrations|Backfill|METALDOCS_SKIP_STARTUP_MIGRATIONS|METALDOCS_MIGRATIONS_DIR" apps internal scripts deploy -S > non_git/runtime-migration-paths.txt
```

Expected: command completes and writes `non_git/runtime-migration-paths.txt`.

- [ ] **Step 3: Write runtime usage inventory**

Create `docs/db-research/runtime-usage-inventory.md` with this structure and fill it from the two evidence files:

````markdown
# Runtime Usage Inventory

## Commands

```powershell
rg -n "FROM |JOIN |INSERT INTO|UPDATE |DELETE FROM|CALL |SELECT .* FROM|ExecContext|QueryContext|QueryRowContext" apps internal -S > non_git/runtime-sql-usage.txt
rg -n "schema_migrations|migrate.Apply|RunStartupMigrations|Backfill|METALDOCS_SKIP_STARTUP_MIGRATIONS|METALDOCS_MIGRATIONS_DIR" apps internal scripts deploy -S > non_git/runtime-migration-paths.txt
```

## Startup and Migration Mutation Paths

| Path | Owner | Behavior | Baseline Decision |
|---|---|---|---|
| `deploy/compose/docker-compose.yml` | Docker infra | Postgres currently mounts historical migrations into entrypoint. | Remove from fresh bootstrap path. |
| `scripts/dev-migrate.ps1` | Dev tooling | Legacy replay script. | Keep explicit legacy mode. |
| `internal/platform/migrate/migrate.go` | API startup | Applies SQL files based on `public.schema_migrations`. | Restrict to post-baseline forward migrations or validation mode. |
| `internal/modules/registry/module.go` | Registry module | Runs startup backfill. | Classify as durable migration or documented runtime maintenance job. |

## Runtime Table Usage By Area

| Area | Tables/Functions Observed | Read/Write | Evidence Files | Baseline Decision |
|---|---|---|---|---|
| auth | `metaldocs.auth_identities`, `metaldocs.auth_sessions`, `metaldocs.iam_users`, `metaldocs.iam_user_roles` | read/write | `internal/modules/auth/**`, `internal/modules/iam/**` | Include if still used. |
| audit | `metaldocs.audit_events` | read/write | `internal/modules/audit/**`, `apps/api/cmd/metaldocs-api/main.go` | Include if still used. |
| registry | `controlled_documents`, `profile_sequence_counters`, grant tables | read/write | `internal/modules/registry/**` | Include if still used. |
| templates | `templates_v2_template`, `templates_v2_template_version`, `templates_v2_approval_config`, `templates_v2_audit_log` | read/write | `internal/modules/templates/**` | Include if still used. |
| documents | `documents`, `document_revisions`, `document_placeholder_values`, snapshots/fill-ins/comments tables | read/write | `internal/modules/documents/**` | Include if still used. |
| approval | approval route/instance/signoff tables | read/write | `internal/modules/documents/approval/**` | Include if still used. |
| taxonomy | process areas, profiles, departments, subjects | read/write | `internal/modules/taxonomy/**` | Include if still used. |
| jobs/outbox | outbox, leases, PDF dispatch, idempotency keys | read/write | `internal/modules/jobs/**`, `internal/modules/render/**` | Include if still used. |

## Candidate Exclusions

Record runtime-unreferenced tables/functions here only when the evidence search finds no current reader or writer.

## Open Decisions For Coordinator

- Registry startup backfill: convert into governed migration or retain as documented runtime maintenance.
- API startup migrations: validate-only locally or apply post-baseline tail.
````

- [ ] **Step 4: Commit runtime inventory**

Run:

```powershell
git add docs/db-research/runtime-usage-inventory.md
git commit -m "docs(db): inventory runtime database usage"
```

Expected: commit succeeds.

### Task 4: Migration Archaeology Inventory `[parallel-safe, read-only]`

**Files:**
- Create: `docs/db-research/migration-archaeology.md`

- [ ] **Step 1: Count migrations and ledger participation**

Run:

```powershell
$files = Get-ChildItem migrations -Filter *.sql | Sort-Object Name
"total=$($files.Count)"
"ledger_insert_files=$(( $files | Select-String -Pattern 'INSERT INTO public\.schema_migrations|INSERT INTO schema_migrations' -List ).Count)"
"begin_files=$(( $files | Select-String -Pattern '^BEGIN;' -List ).Count)"
```

Expected: output includes `total=181`.

- [ ] **Step 2: Produce migration keyword evidence**

Run:

```powershell
rg -n "CREATE TABLE|ALTER TABLE|DROP TABLE|CREATE FUNCTION|CREATE TRIGGER|CREATE EXTENSION|GRANT |REVOKE |INSERT INTO|DELETE FROM|schema_migrations|dev|seed|repair|legacy|destroy|drop" migrations -S > non_git/migration-keyword-inventory.txt
```

Expected: command completes and writes `non_git/migration-keyword-inventory.txt`.

- [ ] **Step 3: Write migration archaeology inventory**

Create `docs/db-research/migration-archaeology.md`:

````markdown
# Migration Archaeology

## Commands

```powershell
$files = Get-ChildItem migrations -Filter *.sql | Sort-Object Name
rg -n "CREATE TABLE|ALTER TABLE|DROP TABLE|CREATE FUNCTION|CREATE TRIGGER|CREATE EXTENSION|GRANT |REVOKE |INSERT INTO|DELETE FROM|schema_migrations|dev|seed|repair|legacy|destroy|drop" migrations -S > non_git/migration-keyword-inventory.txt
```

## Migration Eras

| Range | Era | Main Purpose | Baseline Treatment |
|---|---|---|---|
| `0001-0077` | early metaldocs schema and document model | initial schema, taxonomy, templates, old document flows | include only runtime-used current objects |
| `0101-0118` | docx v2 transition | replacement editor/document tables and destructive cutover | classify carefully; exclude superseded objects |
| `0120-0160` | templates v2, registry, approval, IAM hardening | current product surface plus dev seeds/grants | include product objects; split dev seeds and grants |
| `0167-0188` | bridge/state/auth hardening | ledger-aware fixes and runtime hardening | include current invariants only |
| `0189-0201` | recent capabilities, audit, visibility, blank template | current feature deltas and ledger repairs | include product deltas; split system/reference data |

## Classification Rules

| Classification | Meaning | Curated Baseline Treatment |
|---|---|---|
| current schema | Creates or modifies runtime-used object. | Fold into clean schema. |
| product reference data | Required roles, capabilities, or system records. | Move to `db/reference-data/`. |
| local dev seed | Local users, demo tenant, dev-only workflow records. | Move to `db/dev-seeds/`. |
| grant/security | Roles, grants, security definer functions. | Keep only if current runtime/security model requires it. |
| repair-only | One-time data/schema repair for historical states. | Exclude from fresh baseline unless final state still matters. |
| destructive cleanup | Drops old era objects. | Exclude; fresh baseline never creates removed objects. |
| superseded | Adds object later removed or unused. | Exclude. |

## File Classification Table

Use this table format for every migration file. The coordinator must complete all rows before writing the curated baseline.

| File | Classification | Objects/Data Touched | Current Runtime Evidence | Curated Treatment |
|---|---|---|---|---|
| `migrations/0001_init_documents.sql` | current schema or superseded | `metaldocs.documents`, `metaldocs.document_versions` | compare with runtime usage inventory | include only surviving columns/tables |
| `migrations/0112_docx_v2_schema_migrations_ledger.sql` | ledger bootstrap | `public.schema_migrations` | migration runner reads ledger | replace with curated ledger definition |
| `migrations/0151_seed_dev_tenant_approval_data.sql` | mixed schema/reference/dev seed | `public.user_process_areas`, capabilities, dev users | auth/approval runtime evidence | split into schema/reference/dev seed |
| `migrations/0159_seed_dev_approver_user.sql` | local dev seed plus extension prerequisite | `pgcrypto`, dev approver user | dev-only credential comment | move extension to prerequisites and user to dev seed |
| `migrations/0193_audit_events_hash_chain.sql` | current schema plus extension prerequisite | `metaldocs.audit_events`, `pgcrypto` | audit runtime evidence | fold schema/function into baseline, extension into prerequisites |
| `migrations/0197_repair_recent_migration_ledger.sql` | repair-only plus current data corrections | ledger rows and recent fixes | compare with final state | exclude repair mechanics; include final required state |
````

- [ ] **Step 4: Commit archaeology inventory**

Run:

```powershell
git add docs/db-research/migration-archaeology.md
git commit -m "docs(db): classify historical migration archaeology"
```

Expected: commit succeeds.

### Task 5: Live Schema Inventory `[parallel-safe after local DB is available]`

**Files:**
- Create: `docs/db-research/live-schema-inventory.md`

- [ ] **Step 1: Confirm Postgres container availability**

Run:

```powershell
docker ps --format "table {{.Names}}\t{{.Status}}" | Select-String "metaldocs-postgres"
```

Expected: `metaldocs-postgres` appears. If it does not, run:

```powershell
docker compose -f deploy/compose/docker-compose.yml --env-file .env up -d postgres
```

- [ ] **Step 2: Capture live object inventory**

Run:

```powershell
if (-not (Test-Path non_git/db/catalog)) { New-Item -ItemType Directory -Force non_git/db/catalog | Out-Null }
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "SELECT table_schema, table_name FROM information_schema.tables WHERE table_type='BASE TABLE' AND table_schema IN ('public','metaldocs') ORDER BY 1,2;" > non_git/db/catalog/tables.txt
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "SELECT table_schema, table_name, column_name, data_type, is_nullable, column_default FROM information_schema.columns WHERE table_schema IN ('public','metaldocs') ORDER BY 1,2,ordinal_position;" > non_git/db/catalog/columns.txt
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "SELECT schemaname, tablename, indexname, indexdef FROM pg_indexes WHERE schemaname IN ('public','metaldocs') ORDER BY 1,2,3;" > non_git/db/catalog/indexes.txt
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "SELECT n.nspname, c.relname, con.conname, con.contype, pg_get_constraintdef(con.oid) FROM pg_constraint con JOIN pg_class c ON c.oid = con.conrelid JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname IN ('public','metaldocs') ORDER BY 1,2,3;" > non_git/db/catalog/constraints.txt
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "SELECT trigger_schema, event_object_table, trigger_name, action_timing, event_manipulation FROM information_schema.triggers WHERE trigger_schema IN ('public','metaldocs') ORDER BY 1,2,3;" > non_git/db/catalog/triggers.txt
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "SELECT n.nspname, p.proname, pg_get_function_identity_arguments(p.oid) FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace WHERE n.nspname IN ('public','metaldocs') ORDER BY 1,2,3;" > non_git/db/catalog/functions.txt
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "SELECT extname, extnamespace::regnamespace::text FROM pg_extension ORDER BY extname;" > non_git/db/catalog/extensions.txt
```

Expected: all files are created under `non_git/db/catalog/`.

- [ ] **Step 3: Write live schema inventory**

Create `docs/db-research/live-schema-inventory.md`:

```markdown
# Live Schema Inventory

## Commands

Catalog evidence was captured into `non_git/db/catalog/` using `psql` against `metaldocs-postgres`.

## Schemas

| Schema | Role In Current DB | Baseline Treatment |
|---|---|---|
| `metaldocs` | Application-owned platform/business schema. | Preferred home for new curated objects. |
| `public` | Legacy and current tables/functions created by historical unqualified migrations. | Keep only runtime-required objects; mark future move candidates. |

## Required Extensions

| Extension | Evidence | Baseline Treatment |
|---|---|---|
| `pgcrypto` | Used by password hash seed and audit hash chain. | Declare in `db/prerequisites/0001_extensions.sql` before schema/reference data. |

## Object Inventory Files

| Evidence File | Contents |
|---|---|
| `non_git/db/catalog/tables.txt` | live tables |
| `non_git/db/catalog/columns.txt` | live columns |
| `non_git/db/catalog/indexes.txt` | live indexes |
| `non_git/db/catalog/constraints.txt` | live constraints |
| `non_git/db/catalog/triggers.txt` | live triggers |
| `non_git/db/catalog/functions.txt` | live functions |
| `non_git/db/catalog/extensions.txt` | live extensions |

## Curated Inclusion Notes

The coordinator must compare these live objects against runtime usage and migration archaeology. Objects with no runtime usage and no product justification should be excluded from the curated baseline.
```

- [ ] **Step 4: Commit live schema inventory**

Run:

```powershell
git add docs/db-research/live-schema-inventory.md
git commit -m "docs(db): capture live schema inventory"
```

Expected: commit succeeds.

### Task 6: Seed and Reference Data Classification `[parallel-safe, read-only]`

**Files:**
- Create: `docs/db-research/seed-reference-data-classification.md`

- [ ] **Step 1: Search seed/data migrations**

Run:

```powershell
rg -n "INSERT INTO|ON CONFLICT|password|tenant|approver|admin-local|reviewer-1|system blank|__system_blank__|role_capabilities|iam_users|auth_identities|capability" migrations internal apps -S > non_git/seed-reference-evidence.txt
```

Expected: command completes and writes `non_git/seed-reference-evidence.txt`.

- [ ] **Step 2: Write seed classification**

Create `docs/db-research/seed-reference-data-classification.md`:

````markdown
# Seed and Reference Data Classification

## Commands

```powershell
rg -n "INSERT INTO|ON CONFLICT|password|tenant|approver|admin-local|reviewer-1|system blank|__system_blank__|role_capabilities|iam_users|auth_identities|capability" migrations internal apps -S > non_git/seed-reference-evidence.txt
```

## Classification Rules

| Type | Definition | Target Location |
|---|---|---|
| product reference data | Required for the app to work in every environment. | `db/reference-data/0001_product_reference_data.sql` |
| local dev seed | Local users, passwords, demo tenant, convenience approval actors. | `db/dev-seeds/0001_local_dev_seed.sql` |
| automated test fixture | Data only needed by tests. | test setup code or test fixtures |
| historical repair data | One-time fix for previous DB state. | legacy migrations only |

## Known Product Reference Candidates

| Data | Evidence | Target Decision |
|---|---|---|
| role capabilities | `metaldocs.role_capabilities` seed migrations and IAM runtime checks | product reference data |
| system blank template record | `migrations/0199_system_blank_template.sql` and document creation flow | product reference data if runtime requires it |
| baseline ledger marker | migration runner policy | product reference data or baseline schema setup |

## Known Dev Seed Candidates

| Data | Evidence | Target Decision |
|---|---|---|
| `approver` user and password | `migrations/0159_seed_dev_approver_user.sql` | dev seed |
| `reviewer-1` user | `migrations/0151_seed_dev_tenant_approval_data.sql` | dev seed unless product workflow requires seeded user |
| `admin-local` demo tenant membership | `migrations/0151_seed_dev_tenant_approval_data.sql` | dev seed |
| default local tenant ID `ffffffff-ffff-ffff-ffff-ffffffffffff` | multiple local/demo seed references | dev seed unless app config declares it required |

## Product/Dev Boundary Rule

The curated baseline must allow product schema initialization without local-only users or demo workflow data.
````

- [ ] **Step 3: Commit seed classification**

Run:

```powershell
git add docs/db-research/seed-reference-data-classification.md
git commit -m "docs(db): classify reference and dev seed data"
```

Expected: commit succeeds.

### Task 7: Governance Update Inventory `[parallel-safe, read-only]`

**Files:**
- Create: `docs/db-research/governance-update-plan.md`

- [ ] **Step 1: Inspect governance files**

Run:

```powershell
Get-Content -Raw AGENTS.md > non_git/agents-current.txt
Get-Content -Raw CLAUDE.md > non_git/claude-current.txt
Get-ChildItem .agents/skills -Directory | Select-Object -ExpandProperty Name > non_git/agents-skills.txt
Get-ChildItem .claude/skills -Directory | Select-Object -ExpandProperty Name > non_git/claude-skills.txt
```

Expected: four evidence files are created under `non_git/`.

- [ ] **Step 2: Write governance update plan**

Create `docs/db-research/governance-update-plan.md`:

```markdown
# Governance Update Plan

## Canonical Rule

Detailed DB workflow rules should live in `.claude/skills/metaldocs-database/SKILL.md` and `wiki/database/migration-policy.md`.

`AGENTS.md` and `CLAUDE.md` should contain concise pointers only.

## Required Skill Additions

| File | Purpose |
|---|---|
| `.claude/skills/metaldocs-database/SKILL.md` | canonical database migration/bootstrap/dictionary workflow |
| `.agents/skills/metaldocs-database/SKILL.md` | Codex bridge that points to the canonical skill |

## Required Top-Level Instruction Additions

| File | Change |
|---|---|
| `AGENTS.md` | Add DB section and skill map pointer for database migrations/bootstrap/dictionary work. |
| `CLAUDE.md` | Mirror the concise DB section. |

## Required Wiki Additions

| File | Purpose |
|---|---|
| `wiki/database/overview.md` | database ownership and lifecycle overview |
| `wiki/database/migration-policy.md` | curated baseline, legacy replay, forward migration, seed rules |
| `wiki/database/dictionary-index.md` | index of table dictionary pages |

## Concision Rule

Do not duplicate detailed migration rules in `AGENTS.md`, `CLAUDE.md`, and skills. The top-level files point to the canonical skill and wiki.
```

- [ ] **Step 3: Commit governance inventory**

Run:

```powershell
git add docs/db-research/governance-update-plan.md
git commit -m "docs(db): plan database workflow governance updates"
```

Expected: commit succeeds.

---

## Phase 2: Coordinator Synthesis and Approval Gate

### Task 8: Write Curated Baseline Inclusion Catalog

**Files:**
- Create: `docs/db-research/curated-baseline-inclusion-catalog.md`

- [ ] **Step 1: Review all research files**

Run:

```powershell
Get-Content docs/db-research/runtime-usage-inventory.md
Get-Content docs/db-research/migration-archaeology.md
Get-Content docs/db-research/live-schema-inventory.md
Get-Content docs/db-research/seed-reference-data-classification.md
Get-Content docs/db-research/governance-update-plan.md
```

Expected: all five files are readable.

- [ ] **Step 2: Create inclusion catalog**

Create `docs/db-research/curated-baseline-inclusion-catalog.md`:

```markdown
# Curated Baseline Inclusion Catalog

## Decision Rules

- Include objects used by current runtime code.
- Include product reference data required in every environment.
- Include extensions before any object/data uses them.
- Exclude objects created and later dropped by historical migrations.
- Exclude local users, passwords, demo tenants, and workflow convenience data from product schema/reference data.
- Mark legacy `public` objects as intentionally retained, move candidate, or excluded.

## Schema Objects

| Object | Type | Schema | Owner | Evidence | Curated Decision | Dictionary Page |
|---|---|---|---|---|---|---|
| `public.schema_migrations` | table | `public` | platform/db tooling | migration runner reads it | include with baseline marker policy | `wiki/database/tables/schema_migrations.md` |
| `metaldocs.auth_identities` | table | `metaldocs` | auth | auth runtime usage | include if current repository uses it | `wiki/database/tables/auth_identities.md` |
| `metaldocs.auth_sessions` | table | `metaldocs` | auth | auth runtime usage | include if current repository uses it | `wiki/database/tables/auth_sessions.md` |
| `metaldocs.iam_users` | table | `metaldocs` | IAM | IAM/auth runtime usage | include if current repository uses it | `wiki/database/tables/iam_users.md` |
| `metaldocs.iam_user_roles` | table | `metaldocs` | IAM | IAM/auth runtime usage | include if current repository uses it | `wiki/database/tables/iam_user_roles.md` |
| `metaldocs.role_capabilities` | table | `metaldocs` | IAM | capability checks | include if current repository uses it | `wiki/database/tables/role_capabilities.md` |
| `metaldocs.audit_events` | table | `metaldocs` | audit | audit writer/reader | include if current repository uses it | `wiki/database/tables/audit_events.md` |
| `public.templates_v2_template` | table | `public` | templates | templates runtime usage | include if current repository uses it; mark legacy public | `wiki/database/tables/templates_v2_template.md` |
| `public.templates_v2_template_version` | table | `public` | templates | templates/documents runtime usage | include if current repository uses it; mark legacy public | `wiki/database/tables/templates_v2_template_version.md` |
| `public.controlled_documents` | table | `public` | registry | registry runtime usage | include if current repository uses it; mark legacy public | `wiki/database/tables/controlled_documents.md` |

## Reference Data

| Data | Target File | Evidence | Curated Decision |
|---|---|---|---|
| baseline marker | `db/reference-data/0001_product_reference_data.sql` | migration policy | include |
| product capabilities | `db/reference-data/0001_product_reference_data.sql` | IAM capability runtime | include current set |
| system blank template | `db/reference-data/0001_product_reference_data.sql` | document creation runtime | include if runtime requires non-null default template |

## Dev Seeds

| Data | Target File | Evidence | Curated Decision |
|---|---|---|---|
| `admin-local` credential/user/membership | `db/dev-seeds/0001_local_dev_seed.sql` | local auth/dev setup | include as optional dev seed |
| `approver` credential/user/role | `db/dev-seeds/0001_local_dev_seed.sql` | local approval flow | include as optional dev seed |
| `reviewer-1` credential/user/membership | `db/dev-seeds/0001_local_dev_seed.sql` | local approval flow | include as optional dev seed if login flow needs reviewer actor |

## Exclusions

| Object/Data | Evidence | Reason |
|---|---|---|
| CK5/MDDM tables dropped by legacy destructive cutover | historical migration archaeology | historical-only and not part of current fresh schema |
| repair-only ledger inserts | `0197` and related repairs | not needed for fresh curated baseline |
| local-only passwords in product migrations | seed classification | dev seed only |
```

- [ ] **Step 3: Commit inclusion catalog**

Run:

```powershell
git add docs/db-research/curated-baseline-inclusion-catalog.md
git commit -m "docs(db): define curated baseline inclusion catalog"
```

Expected: commit succeeds.

### Task 9: Approval Gate Before Implementation

**Files:** none

- [ ] **Step 1: Summarize decisions for user review**

Prepare a short review note with:

```text
Research package complete.

Please review:
- docs/db-research/runtime-usage-inventory.md
- docs/db-research/migration-archaeology.md
- docs/db-research/live-schema-inventory.md
- docs/db-research/seed-reference-data-classification.md
- docs/db-research/governance-update-plan.md
- docs/db-research/curated-baseline-inclusion-catalog.md

Implementation will not edit DB bootstrap/runtime files until the inclusion catalog is approved.
```

- [ ] **Step 2: Wait for approval**

Expected: user approves the inclusion catalog or requests changes. If changes are requested, update the research files and catalog before continuing.

---

## Phase 3: Curated DB Artifact Layout

### Task 10: Create New DB Artifact Directories

**Files:**
- Create: `db/prerequisites/0001_extensions.sql`
- Create: `db/baseline/0001_current_schema.sql`
- Create: `db/reference-data/0001_product_reference_data.sql`
- Create: `db/dev-seeds/0001_local_dev_seed.sql`
- Create: `db/migrations/README.md`

- [ ] **Step 1: Create prerequisite extensions file**

Create `db/prerequisites/0001_extensions.sql`:

```sql
-- MetalDocs database prerequisites.
-- This file is applied before product schema, reference data, dev seeds, and tail migrations.

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;
```

- [ ] **Step 2: Create baseline schema skeleton**

Create `db/baseline/0001_current_schema.sql`:

```sql
-- MetalDocs curated current-state schema baseline.
-- This file contains product schema only. Product reference data belongs in
-- db/reference-data/0001_product_reference_data.sql. Local-only seed data belongs
-- in db/dev-seeds/0001_local_dev_seed.sql.

BEGIN;

CREATE SCHEMA IF NOT EXISTS metaldocs;

CREATE TABLE IF NOT EXISTS public.schema_migrations (
  version TEXT PRIMARY KEY,
  description TEXT NOT NULL DEFAULT '',
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMIT;
```

- [ ] **Step 3: Create product reference data skeleton**

Create `db/reference-data/0001_product_reference_data.sql`:

```sql
-- MetalDocs product reference data.
-- Data in this file is required for every environment.

BEGIN;

INSERT INTO public.schema_migrations (version, description)
VALUES ('baseline-2026-05-14', 'curated current-state database baseline')
ON CONFLICT (version) DO NOTHING;

COMMIT;
```

- [ ] **Step 4: Create local dev seed skeleton**

Create `db/dev-seeds/0001_local_dev_seed.sql`:

```sql
-- MetalDocs local development seed data.
-- This file is optional and must not be required for product schema bootstrap.

BEGIN;

-- Local dev users, demo tenant, and approval convenience records are inserted here
-- after the seed classification catalog approves their final shape.

COMMIT;
```

- [ ] **Step 5: Create forward migration README**

Create `db/migrations/README.md`:

```markdown
# Post-Baseline Forward Migrations

Use this directory for forward-only database changes after the curated baseline cutoff.

Rules:

- Do not add local dev users or demo data here.
- Every migration must insert one `public.schema_migrations` row.
- Every migration must be safe to run exactly once.
- Destructive changes require ADR approval, rollback notes, and a maintenance-window plan.
- Update `wiki/database/` when a table, column, function, trigger, reference data rule, or ownership boundary changes.
```

- [ ] **Step 6: Commit DB artifact layout**

Run:

```powershell
git add db/prerequisites/0001_extensions.sql db/baseline/0001_current_schema.sql db/reference-data/0001_product_reference_data.sql db/dev-seeds/0001_local_dev_seed.sql db/migrations/README.md
git commit -m "feat(db): add curated database artifact layout"
```

Expected: commit succeeds.

### Task 11: Populate Curated Baseline Schema

**Files:**
- Modify: `db/baseline/0001_current_schema.sql`
- Modify: `docs/db-research/curated-baseline-inclusion-catalog.md`

- [ ] **Step 1: Build table definitions from approved inclusion catalog**

For every included table in `docs/db-research/curated-baseline-inclusion-catalog.md`, add the final `CREATE TABLE`, primary key, foreign key, unique, check, index, trigger, and function definitions to `db/baseline/0001_current_schema.sql`.

Required SQL structure:

```sql
BEGIN;

CREATE SCHEMA IF NOT EXISTS metaldocs;

CREATE TABLE IF NOT EXISTS public.schema_migrations (
  version TEXT PRIMARY KEY,
  description TEXT NOT NULL DEFAULT '',
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Add curated product schema objects below in dependency order:
-- 1. schemas and enum types
-- 2. base tables
-- 3. dependent tables
-- 4. functions
-- 5. triggers
-- 6. indexes
-- 7. constraints that require both sides to exist

COMMIT;
```

Dependency order must be visible in the file. Avoid creating tables that are later dropped or altered only to reach the final shape.

- [ ] **Step 2: Check for forbidden dev seed markers in baseline**

Run:

```powershell
rg -n "ApproverMetalDocs|admin-local|reviewer-1|approver|ffffffff-ffff-ffff-ffff-ffffffffffff|demo|dev seed" db/baseline/0001_current_schema.sql
```

Expected: no matches, except comments that explicitly say dev seeds are forbidden.

- [ ] **Step 3: Check baseline contains no raw dump noise**

Run:

```powershell
rg -n "Dumped from database|Dumped by pg_dump|\\restrict|\\unrestrict|OWNER TO|ACL|COPY " db/baseline/0001_current_schema.sql
```

Expected: no matches.

- [ ] **Step 4: Commit curated schema**

Run:

```powershell
git add db/baseline/0001_current_schema.sql docs/db-research/curated-baseline-inclusion-catalog.md
git commit -m "feat(db): add curated current-state schema baseline"
```

Expected: commit succeeds.

### Task 12: Populate Product Reference Data

**Files:**
- Modify: `db/reference-data/0001_product_reference_data.sql`

- [ ] **Step 1: Add required product reference inserts**

Move only approved product reference data from `docs/db-research/curated-baseline-inclusion-catalog.md` into `db/reference-data/0001_product_reference_data.sql`.

Every insert must be idempotent:

```sql
INSERT INTO metaldocs.role_capabilities (role, capability, description)
VALUES ('system_admin', 'audit.read', 'Read audit events')
ON CONFLICT (role, capability) DO NOTHING;
```

If the final schema uses a different unique key, use that key in the `ON CONFLICT` clause.

- [ ] **Step 2: Reject local-only data from product reference file**

Run:

```powershell
rg -n "ApproverMetalDocs|admin-local|reviewer-1|approver|demo|local dev|password" db/reference-data/0001_product_reference_data.sql
```

Expected: no matches, unless the line is a comment explaining that local dev data is forbidden.

- [ ] **Step 3: Commit product reference data**

Run:

```powershell
git add db/reference-data/0001_product_reference_data.sql
git commit -m "feat(db): add product reference data baseline"
```

Expected: commit succeeds.

### Task 13: Populate Optional Local Dev Seeds

**Files:**
- Modify: `db/dev-seeds/0001_local_dev_seed.sql`

- [ ] **Step 1: Add optional local seed data**

Move approved local-only data into `db/dev-seeds/0001_local_dev_seed.sql`.

Required header:

```sql
-- MetalDocs local development seed data.
-- Optional. Never apply this file in production/shared environments.
-- Credentials in this file are local-only and exist to support manual developer flows.
```

Required transaction pattern:

```sql
BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;

-- Insert local-only users, demo tenant memberships, and approval convenience data here.
-- Each insert must be idempotent.

COMMIT;
```

- [ ] **Step 2: Confirm product files do not depend on dev seed users**

Run:

```powershell
rg -n "admin-local|reviewer-1|approver" db/baseline db/reference-data
```

Expected: no matches.

- [ ] **Step 3: Commit dev seed data**

Run:

```powershell
git add db/dev-seeds/0001_local_dev_seed.sql
git commit -m "feat(db): add optional local dev seed data"
```

Expected: commit succeeds.

---

## Phase 4: Bootstrap Scripts and Runtime Policy

### Task 14: Stop Docker Entrypoint Auto-Running Historical Migrations

**Files:**
- Modify: `deploy/compose/docker-compose.yml`

- [ ] **Step 1: Remove migration entrypoint mount**

In `deploy/compose/docker-compose.yml`, remove this Postgres volume line:

```yaml
      - ../../migrations:/docker-entrypoint-initdb.d:ro
```

Keep the data volume:

```yaml
      - metaldocs_postgres_data:/var/lib/postgresql/data
```

- [ ] **Step 2: Verify compose parses**

Run:

```powershell
docker compose -f deploy/compose/docker-compose.yml --env-file .env config > non_git/compose-config-after-db-bootstrap-change.yml
```

Expected: command succeeds and generated config has no `/docker-entrypoint-initdb.d` mount.

- [ ] **Step 3: Commit compose change**

Run:

```powershell
git add deploy/compose/docker-compose.yml
git commit -m "chore(db): stop postgres entrypoint migration replay"
```

Expected: commit succeeds.

### Task 15: Rewrite Baseline Bootstrap Script

**Files:**
- Modify: `scripts/dev-bootstrap-baseline.ps1`

- [ ] **Step 1: Replace script behavior**

Update `scripts/dev-bootstrap-baseline.ps1` so it:

- loads `.env`
- resets DB using `scripts/dev-db-reset.ps1`
- waits for empty Postgres
- applies `db/prerequisites/0001_extensions.sql`
- applies `db/baseline/0001_current_schema.sql`
- applies `db/reference-data/0001_product_reference_data.sql`
- applies `db/dev-seeds/0001_local_dev_seed.sql` only when `-WithDevSeed` is passed
- applies files under `db/migrations/*.sql` in lexical order

Required parameter block:

```powershell
param(
  [string]$ComposeFile = "deploy/compose/docker-compose.yml",
  [string]$EnvFile = ".env",
  [switch]$WithDevSeed
)
```

Required apply helper:

```powershell
function Invoke-DbSqlFile {
  param([string]$Path)

  if (-not (Test-Path $Path)) {
    throw "SQL file not found: $Path"
  }

  Write-Host "[dev-bootstrap-baseline] Applying $Path"
  Get-Content -Raw $Path | docker compose -f $ComposeFile --env-file $EnvFile exec -T postgres `
    psql -v ON_ERROR_STOP=1 -U $env:POSTGRES_USER -d $env:POSTGRES_DB | Out-Host

  if ($LASTEXITCODE -ne 0) {
    throw "[dev-bootstrap-baseline] SQL apply failed: $Path"
  }
}
```

- [ ] **Step 2: Remove skip-on-ledger behavior**

Delete the branch that says:

```powershell
Container bootstrap already initialized schema; skipping baseline apply.
```

The new script must fail or apply the curated path; it must not silently accept entrypoint-created legacy state.

- [ ] **Step 3: Commit baseline script**

Run:

```powershell
git add scripts/dev-bootstrap-baseline.ps1
git commit -m "feat(db): apply curated baseline bootstrap explicitly"
```

Expected: commit succeeds.

### Task 16: Make Legacy Replay Explicit

**Files:**
- Modify: `scripts/dev-migrate.ps1`
- Modify: `docs/runbooks/migration-legacy-replay.md`

- [ ] **Step 1: Update script messaging**

At the top of `scripts/dev-migrate.ps1`, after env loading and validation, ensure output clearly says:

```powershell
Write-Host "[dev-migrate] LEGACY REPLAY MODE."
Write-Host "[dev-migrate] This script applies the historical migrations/ chain for recovery/debugging."
Write-Host "[dev-migrate] Normal fresh local setup uses scripts/dev-bootstrap-baseline.ps1."
```

- [ ] **Step 2: Remove dependency on Docker entrypoint path**

Change each migration apply from container path:

```powershell
$containerPath = "/docker-entrypoint-initdb.d/$($migration.Name)"
```

to streaming local file content:

```powershell
Get-Content -Raw $migration.FullName | docker compose -f $ComposeFile --env-file $EnvFile exec -T postgres `
  psql -v ON_ERROR_STOP=1 -U $env:POSTGRES_USER -d $env:POSTGRES_DB | Out-Host
```

- [ ] **Step 3: Update legacy replay runbook**

In `docs/runbooks/migration-legacy-replay.md`, add:

```markdown
Legacy replay streams local `migrations/*.sql` files through `psql`. Docker Postgres no longer auto-runs this directory through `/docker-entrypoint-initdb.d`.
```

- [ ] **Step 4: Commit legacy replay update**

Run:

```powershell
git add scripts/dev-migrate.ps1 docs/runbooks/migration-legacy-replay.md
git commit -m "chore(db): make legacy migration replay explicit"
```

Expected: commit succeeds.

### Task 17: Update Local Dev Startup Flow

**Files:**
- Modify: `scripts/dev-local.ps1`
- Modify: `docs/runbooks/migration-baseline-local.md`
- Create or Modify: `docs/runbooks/database-bootstrap.md`

- [ ] **Step 1: Change dev-local to use baseline bootstrap**

Replace:

```powershell
Write-Host "[dev-local] Applying migrations (idempotent) ..."
powershell -ExecutionPolicy Bypass -File scripts/dev-migrate.ps1 | Out-Host
```

with:

```powershell
Write-Host "[dev-local] Applying curated baseline bootstrap with local dev seed..."
powershell -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1 -WithDevSeed | Out-Host
```

- [ ] **Step 2: Update baseline local runbook**

Set `docs/runbooks/migration-baseline-local.md` commands to:

````markdown
```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1 -WithDevSeed
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/start-api.ps1 -Build -NoWorker
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v2/controlled-documents
```
````

- [ ] **Step 3: Create database bootstrap runbook**

Create `docs/runbooks/database-bootstrap.md`:

````markdown
# Runbook: Database Bootstrap

## Fresh local setup

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1 -WithDevSeed
```

## Fresh product schema without dev seed

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1
```

## Legacy replay for recovery/debugging

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-db-reset.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-migrate.ps1
```

## Verification

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/start-api.ps1 -Build -NoWorker
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v2/controlled-documents
```
````

- [ ] **Step 4: Commit local dev flow**

Run:

```powershell
git add scripts/dev-local.ps1 docs/runbooks/migration-baseline-local.md docs/runbooks/database-bootstrap.md
git commit -m "docs(db): make curated baseline the local bootstrap path"
```

Expected: commit succeeds.

### Task 18: Scope API Startup Migrations

**Files:**
- Modify: `internal/platform/migrate/migrate.go`
- Modify: `apps/api/cmd/metaldocs-api/main.go`

- [ ] **Step 1: Add explicit migration directory expectation**

In `apps/api/cmd/metaldocs-api/main.go`, change the default migration directory from:

```go
migrationsDir = "migrations"
```

to:

```go
migrationsDir = "db/migrations"
```

- [ ] **Step 2: Keep skip env behavior**

Preserve:

```go
METALDOCS_SKIP_STARTUP_MIGRATIONS
```

so local verification can choose validation-only startup.

- [ ] **Step 3: Update migration runner package comment**

In `internal/platform/migrate/migrate.go`, update the package comment to say:

```go
// Package migrate applies post-baseline forward SQL files from a configured
// migrations directory. It is not responsible for fresh database bootstrap;
// curated baseline bootstrap is owned by scripts/dev-bootstrap-baseline.ps1.
```

- [ ] **Step 4: Commit startup migration policy**

Run:

```powershell
git add internal/platform/migrate/migrate.go apps/api/cmd/metaldocs-api/main.go
git commit -m "chore(db): scope API startup migrations to post-baseline tail"
```

Expected: commit succeeds.

### Task 19: Classify Registry Startup Backfill

**Files:**
- Modify: `internal/modules/registry/module.go`
- Modify: `docs/db-research/curated-baseline-inclusion-catalog.md`

- [ ] **Step 1: Inspect backfill implementation**

Run:

```powershell
rg -n "func BackfillLegacyDocuments|BackfillLegacyDocuments" internal/modules/registry internal/modules/documents -S
```

Expected: exact implementation file is identified.

- [ ] **Step 2: Decide backfill treatment**

Choose one of these two concrete paths:

Path A: If the backfill changes durable product state needed by fresh DBs, move its final state into `db/baseline/` or `db/reference-data/` and remove startup invocation.

Path B: If it is a true runtime maintenance job for old data only, keep it but rename/document it as legacy maintenance and make it no-op on fresh baseline.

- [ ] **Step 3: Record decision in catalog**

Add a section to `docs/db-research/curated-baseline-inclusion-catalog.md`:

```markdown
## Runtime Backfill Decision

| Backfill | Decision | Reason | Verification |
|---|---|---|---|
| `registry.BackfillLegacyDocuments` | `governed-migration` or `documented-legacy-maintenance-job` | cite the implementation file and runtime usage evidence that proves the chosen classification | fresh baseline startup passes without unintended durable mutation |
```

- [ ] **Step 4: Commit backfill classification**

Run:

```powershell
git add internal/modules/registry/module.go docs/db-research/curated-baseline-inclusion-catalog.md
git commit -m "chore(db): classify registry startup backfill"
```

Expected: commit succeeds.

---

## Phase 5: Database Dictionary and Wiki Policy

### Task 20: Create Database Wiki Skeleton

**Files:**
- Create: `wiki/database/overview.md`
- Create: `wiki/database/schemas.md`
- Create: `wiki/database/relationships.md`
- Create: `wiki/database/reference-data.md`
- Create: `wiki/database/migration-policy.md`
- Create: `wiki/database/dictionary-index.md`

- [ ] **Step 1: Create overview**

Create `wiki/database/overview.md`:

```markdown
# Database Overview

MetalDocs uses PostgreSQL as the source of truth for product state.

The database is governed by:

- curated baseline for fresh environments
- product reference data for required roles/capabilities/system records
- optional local dev seeds for developer workflows
- post-baseline forward migrations for new changes
- legacy historical migrations for recovery/debugging until explicit archive approval

Runtime truth comes first. A database object belongs in the curated baseline only when current runtime code, product behavior, or an explicit documented debt note justifies it.
```

- [ ] **Step 2: Create schemas page**

Create `wiki/database/schemas.md`:

```markdown
# Database Schemas

## `metaldocs`

Preferred schema for application-owned business and platform objects.

## `public`

Contains legacy and current objects created by historical unqualified migrations. New objects should not be added here unless a migration plan explicitly requires it.

Every retained `public` object must be marked in its table dictionary page as:

- intentionally public for now
- candidate for future move
- historical/unused candidate for exclusion
```

- [ ] **Step 3: Create relationships page**

Create `wiki/database/relationships.md`:

```markdown
# Database Relationships

This page summarizes cross-module relationships that matter for runtime behavior.

Detailed primary keys, foreign keys, checks, indexes, triggers, and functions live in each table dictionary page under `wiki/database/tables/`.
```

- [ ] **Step 4: Create reference data page**

Create `wiki/database/reference-data.md`:

```markdown
# Product Reference Data and Dev Seeds

Product reference data is required in every environment and lives in `db/reference-data/`.

Local dev seed data is optional and lives in `db/dev-seeds/`.

Production schema migrations must not create local-only users, demo tenants, or convenience passwords.
```

- [ ] **Step 5: Create migration policy page**

Create `wiki/database/migration-policy.md`:

```markdown
# Database Migration Policy

## Fresh environments

Use the curated baseline path:

1. `db/prerequisites/`
2. `db/baseline/`
3. `db/reference-data/`
4. optional `db/dev-seeds/`
5. `db/migrations/`

## Existing environments

Use the legacy migration chain until a reviewed upgrade path replaces it.

## New changes

New schema changes after the baseline cutoff are forward-only migrations under `db/migrations/`.

Every post-baseline migration must insert a `public.schema_migrations` row.

Destructive changes require ADR approval, rollback notes, and maintenance-window planning.
```

- [ ] **Step 6: Create dictionary index**

Create `wiki/database/dictionary-index.md`:

```markdown
# Database Dictionary Index

Every table in `db/baseline/0001_current_schema.sql` must have a dictionary page here or an explicit exception.

## Tables

| Table | Schema | Owner | Dictionary |
|---|---|---|---|
| `schema_migrations` | `public` | platform/db tooling | `wiki/database/tables/schema_migrations.md` |
```

- [ ] **Step 7: Commit wiki skeleton**

Run:

```powershell
git add wiki/database
git commit -m "docs(db): add database wiki skeleton"
```

Expected: commit succeeds.

### Task 21: Add Table Dictionary Pages

**Files:**
- Create: `wiki/database/tables/*.md`
- Modify: `wiki/database/dictionary-index.md`

- [ ] **Step 1: Create table dictionary template**

For each included table in `docs/db-research/curated-baseline-inclusion-catalog.md`, create `wiki/database/tables/<table>.md` using:

```markdown
# `<schema>.<table>`

## Owner

Module or platform owner.

## Purpose

One paragraph explaining why the table exists in product terms.

## Row Lifecycle

Explain how rows are created, changed, archived, or deleted.

## Columns

| Column | Type | Nullable | Meaning |
|---|---|---|---|

## Keys and Constraints

| Name | Type | Definition | Reason |
|---|---|---|---|

## Indexes

| Name | Definition | Reason |
|---|---|---|

## Triggers and Functions

| Name | Purpose |
|---|---|

## Runtime Usage

| Reader/Writer | File | Behavior |
|---|---|---|

## Seed or Reference Data

State whether the table has product reference data, dev seed data, tenant data, audit data, or runtime-generated data.

## Notes and Debt

Document legacy schema placement, move candidates, or known cleanup work.
```

- [ ] **Step 2: Update dictionary index**

Add every page to `wiki/database/dictionary-index.md`:

```markdown
| `<table>` | `<schema>` | `<owner>` | `wiki/database/tables/<table>.md` |
```

- [ ] **Step 3: Commit table dictionary**

Run:

```powershell
git add wiki/database/tables wiki/database/dictionary-index.md
git commit -m "docs(db): add database table dictionary"
```

Expected: commit succeeds.

### Task 22: Add Dictionary Coverage Check

**Files:**
- Create: `scripts/check-db-dictionary-coverage.ps1`

- [ ] **Step 1: Create coverage script**

Create `scripts/check-db-dictionary-coverage.ps1`:

```powershell
param(
  [string]$BaselineFile = "db/baseline/0001_current_schema.sql",
  [string]$DictionaryDir = "wiki/database/tables"
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

if (-not (Test-Path $BaselineFile)) {
  throw "Baseline file not found: $BaselineFile"
}

if (-not (Test-Path $DictionaryDir)) {
  throw "Dictionary directory not found: $DictionaryDir"
}

$matches = Select-String -Path $BaselineFile -Pattern 'CREATE TABLE IF NOT EXISTS\s+([a-zA-Z0-9_]+\.)?([a-zA-Z0-9_]+)' -AllMatches
$tables = @()
foreach ($match in $matches) {
  foreach ($m in $match.Matches) {
    $tables += $m.Groups[2].Value
  }
}
$tables = $tables | Sort-Object -Unique

$missing = @()
foreach ($table in $tables) {
  $page = Join-Path $DictionaryDir "$table.md"
  if (-not (Test-Path $page)) {
    $missing += $table
  }
}

if ($missing.Count -gt 0) {
  Write-Host "[check-db-dictionary-coverage] Missing dictionary pages:"
  $missing | ForEach-Object { Write-Host "  - $_" }
  exit 1
}

Write-Host "[check-db-dictionary-coverage] Dictionary coverage OK for $($tables.Count) table(s)."
```

- [ ] **Step 2: Run coverage script**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-dictionary-coverage.ps1
```

Expected: coverage passes or prints exact missing table pages. Add missing pages before continuing.

- [ ] **Step 3: Commit coverage script**

Run:

```powershell
git add scripts/check-db-dictionary-coverage.ps1
git commit -m "test(db): add dictionary coverage check"
```

Expected: commit succeeds.

---

## Phase 6: Governance and Skill Updates

### Task 23: Add Canonical Database Skill With `skill-creator`

**Files:**
- Create: `.claude/skills/metaldocs-database/SKILL.md`
- Create: `.agents/skills/metaldocs-database/SKILL.md`

- [ ] **Step 1: Invoke and follow `skill-creator`**

Before writing either skill file, invoke the `skill-creator` skill and follow its workflow for creating/updating effective skills.

Required `skill-creator` conclusions for this task:

```text
Skill name: metaldocs-database
Skill purpose: guide MetalDocs database migration, bootstrap, curated baseline, reference data, dev seed, schema ownership, database dictionary, schema_migrations, extension, grants, triggers, functions, and runtime DB startup drift work.
Canonical skill location: .claude/skills/metaldocs-database/SKILL.md
Codex bridge location: .agents/skills/metaldocs-database/SKILL.md
Resources: no extra references/assets/scripts in the skill folder for the first version; link to wiki/database and existing runbooks instead.
Style: concise frontmatter and body; detailed rules live in wiki/database to avoid duplication.
```

If the skill creation helper scripts are available in the active environment, use them to initialize the skill folder. If they are not available for repo-local `.claude`/`.agents` skill paths, create the two `SKILL.md` files manually using the `skill-creator` anatomy and validation rules.

- [ ] **Step 2: Create canonical Claude skill**

Create `.claude/skills/metaldocs-database/SKILL.md`:

````markdown
---
name: metaldocs-database
description: Use for any MetalDocs database work touching migrations, bootstrap, curated baseline, reference data, dev seeds, schema ownership, database dictionary, schema_migrations, Postgres extensions, grants, triggers, functions, or runtime DB startup drift.
---

# MetalDocs Database Workflow

Use this skill for database migration/bootstrap/dictionary work.

## Required Reads

- `wiki/database/overview.md`
- `wiki/database/migration-policy.md`
- `wiki/database/dictionary-index.md`
- `docs/superpowers/specs/2026-05-14-curated-database-baseline-and-dictionary-design.md`

## Classification

Classify the task before editing:

- runtime prerequisite
- workflow/tooling gap
- schema baseline change
- product reference data change
- local dev seed change
- post-baseline forward migration
- dictionary/wiki update
- legacy replay/debugging

## Rules

- Do not patch historical migrations to hide bootstrap drift.
- Fresh local setup uses curated baseline artifacts, not Docker entrypoint migration replay.
- Product schema, product reference data, and local dev seeds stay separated.
- New post-baseline migrations must write `public.schema_migrations`.
- Required extensions are declared before use.
- Every baseline table needs a wiki dictionary page or explicit exception.
- Runtime truth wins over migration archaeology when deciding whether an object is still used.
- Stop and surface contradictions involving startup path, schema ownership, migration ledger, seed scope, or verification expectations.

## Verification

Use the smallest applicable gate:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1 -WithDevSeed
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-dictionary-coverage.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/start-api.ps1 -Build -NoWorker
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v2/controlled-documents
```
````

- [ ] **Step 3: Create Codex bridge skill**

Create `.agents/skills/metaldocs-database/SKILL.md`:

```markdown
---
name: metaldocs-database
description: Use for any MetalDocs database work touching migrations, bootstrap, curated baseline, reference data, dev seeds, schema ownership, database dictionary, schema_migrations, Postgres extensions, grants, triggers, functions, or runtime DB startup drift.
---

# MetalDocs Database Workflow

Read and follow `.claude/skills/metaldocs-database/SKILL.md`.

This bridge exists so Codex sessions can discover the canonical database workflow.

Stop if canonical guidance is missing or conflicts with `wiki/database/`.
```

- [ ] **Step 4: Validate skill files**

Run these checks:

```powershell
Test-Path .claude/skills/metaldocs-database/SKILL.md
Test-Path .agents/skills/metaldocs-database/SKILL.md
Select-String -Path .claude/skills/metaldocs-database/SKILL.md -Pattern "^name: metaldocs-database$"
Select-String -Path .claude/skills/metaldocs-database/SKILL.md -Pattern "^description: Use for any MetalDocs database work"
Select-String -Path .agents/skills/metaldocs-database/SKILL.md -Pattern "Read and follow `.claude/skills/metaldocs-database/SKILL.md`"
```

Expected: both files exist and all `Select-String` commands return one matching line.

If the `skill-creator` validation script is available, also run:

```powershell
python C:\Users\leandro.theodoro.MN-NTB-LEANDROT\.codex\skills\.system\skill-creator\scripts\quick_validate.py .claude/skills/metaldocs-database
python C:\Users\leandro.theodoro.MN-NTB-LEANDROT\.codex\skills\.system\skill-creator\scripts\quick_validate.py .agents/skills/metaldocs-database
```

Expected: validation passes. If those script paths do not exist, record that validation was limited to the explicit file/frontmatter checks above.

- [ ] **Step 5: Commit database skill**

Run:

```powershell
git add .claude/skills/metaldocs-database/SKILL.md .agents/skills/metaldocs-database/SKILL.md
git commit -m "docs(db): add MetalDocs database workflow skill"
```

Expected: commit succeeds.

### Task 24: Update AGENTS and CLAUDE Pointers

**Files:**
- Modify: `AGENTS.md`
- Modify: `CLAUDE.md`

- [ ] **Step 1: Add concise DB section to AGENTS.md**

Add this section after the Backend/API section or Core Skill Map introduction:

```markdown
## Database

For ANY work on MetalDocs database migrations, bootstrap, curated baseline, reference data, dev seeds, schema ownership, Postgres extensions, grants, triggers, functions, `schema_migrations`, or database dictionary/wiki pages, use the `metaldocs-database` skill at `.agents/skills/metaldocs-database/SKILL.md`.

The database wiki under `wiki/database/` is the source of truth for schema ownership, dictionary entries, migration policy, reference data, and bootstrap rules. Do not duplicate those rules here.
```

Add to the Core Skill Map:

```markdown
- database migrations / bootstrap / curated baseline / seeds / dictionary -> `metaldocs-database`
```

- [ ] **Step 2: Mirror concise DB section in CLAUDE.md**

Add the same short database section, pointing to `.claude/skills/metaldocs-database/SKILL.md`.

- [ ] **Step 3: Commit instruction pointers**

Run:

```powershell
git add AGENTS.md CLAUDE.md
git commit -m "docs(db): add database workflow pointers"
```

Expected: commit succeeds.

### Task 25: Update Existing Migration Docs and ADR

**Files:**
- Modify: `docs/runbooks/migration-governance.md`
- Modify: `docs/runbooks/migration-archive-policy.md`
- Modify: `docs/adr/0007-schema-migration-policy.md`

- [ ] **Step 1: Update governance runbook**

Replace raw baseline language in `docs/runbooks/migration-governance.md` with:

```markdown
# Runbook: Migration Governance

MetalDocs uses a curated current-state baseline for fresh environments and preserves historical migrations for legacy replay/debugging until archive approval.

Canonical policy lives in `wiki/database/migration-policy.md`.

Rules:

1. Do not use Docker Postgres entrypoint to auto-run historical migrations.
2. Do not patch historical migrations to hide bootstrap drift.
3. Keep product schema, product reference data, and local dev seeds separated.
4. Every post-baseline forward migration must record `public.schema_migrations`.
5. Every baseline table must have a database dictionary page or explicit exception.
6. Historical migration archive/deletion requires explicit review.
```

- [ ] **Step 2: Update archive policy**

Add:

```markdown
Historical migrations remain evidence until the curated baseline and legacy replay gates pass. Archive classification must not remove the ability to debug existing DB upgrade history.
```

- [ ] **Step 3: Update ADR-0007**

Append:

```markdown
## Operational note (curated baseline)

Fresh environments use a curated current-state baseline plus product reference data and optional dev seeds. Historical migrations remain available for legacy replay/debugging until an explicit archive decision is approved.

The canonical operational policy is `wiki/database/migration-policy.md`.
```

- [ ] **Step 4: Commit docs**

Run:

```powershell
git add docs/runbooks/migration-governance.md docs/runbooks/migration-archive-policy.md docs/adr/0007-schema-migration-policy.md
git commit -m "docs(db): align migration governance with curated baseline"
```

Expected: commit succeeds.

---

## Phase 7: Verification Tooling

### Task 26: Add Fresh Bootstrap Verification Script

**Files:**
- Create: `scripts/check-db-bootstrap.ps1`

- [ ] **Step 1: Create script**

Create `scripts/check-db-bootstrap.ps1`:

```powershell
param(
  [switch]$WithDevSeed
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$argsList = @()
if ($WithDevSeed) {
  $argsList += "-WithDevSeed"
}

Write-Host "[check-db-bootstrap] Bootstrapping curated database..."
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1 @argsList | Out-Host

Write-Host "[check-db-bootstrap] Checking dictionary coverage..."
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-dictionary-coverage.ps1 | Out-Host

Write-Host "[check-db-bootstrap] Checking migration ledger..."
$ledger = docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "SELECT version FROM public.schema_migrations WHERE version = 'baseline-2026-05-14';"
if (($ledger | Out-String).Trim() -ne "baseline-2026-05-14") {
  throw "[check-db-bootstrap] baseline ledger marker missing"
}

Write-Host "[check-db-bootstrap] Curated DB bootstrap checks passed."
```

- [ ] **Step 2: Run script**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-bootstrap.ps1 -WithDevSeed
```

Expected: script passes.

- [ ] **Step 3: Commit script**

Run:

```powershell
git add scripts/check-db-bootstrap.ps1
git commit -m "test(db): add curated bootstrap verification"
```

Expected: commit succeeds.

### Task 27: Expand Baseline Equivalence Check

**Files:**
- Modify: `scripts/check-baseline-equivalence.ps1`

- [ ] **Step 1: Add constraints/indexes/functions checks**

Extend `scripts/check-baseline-equivalence.ps1` to write and compare:

```powershell
columns.txt
constraints.txt
indexes.txt
triggers.txt
functions.txt
extensions.txt
```

Use the catalog queries from Task 5 for both reference and candidate DBs.

- [ ] **Step 2: Run equivalence check against current DB**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-baseline-equivalence.ps1
```

Expected: comparing the same DB to itself passes.

- [ ] **Step 3: Commit equivalence update**

Run:

```powershell
git add scripts/check-baseline-equivalence.ps1
git commit -m "test(db): compare runtime-used schema objects"
```

Expected: commit succeeds.

---

## Phase 8: Final Verification

### Task 28: Fresh Curated Bootstrap Gate

**Files:** none unless verification exposes a defect in task-owned files.

- [ ] **Step 1: Run fresh baseline bootstrap with dev seeds**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-bootstrap.ps1 -WithDevSeed
```

Expected: bootstrap checks pass.

- [ ] **Step 2: Start API**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/start-api.ps1 -Build -NoWorker
```

Expected: API starts without migration/bootstrap failure.

- [ ] **Step 3: Run runtime gate**

In a separate shell if the API process is foregrounded, run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v2/controlled-documents
```

Expected: auth/session and target route checks pass.

### Task 29: Product Schema Without Dev Seed Gate

**Files:** none unless verification exposes a defect in task-owned files.

- [ ] **Step 1: Bootstrap without dev seeds**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-bootstrap.ps1
```

Expected: bootstrap succeeds and product schema/reference data initialize without local users.

- [ ] **Step 2: Start API in no-dev-seed mode**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/start-api.ps1 -Build -NoWorker
```

Expected: API starts. Auth login may require configured bootstrap admin, but schema startup must not depend on dev seed users.

### Task 30: Legacy Replay Gate

**Files:** none unless verification exposes a defect in task-owned files.

- [ ] **Step 1: Run intentional legacy replay**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-db-reset.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev-migrate.ps1
```

Expected: historical replay remains available for recovery/debugging.

- [ ] **Step 2: Confirm legacy replay is not Docker entrypoint-driven**

Run:

```powershell
docker inspect metaldocs-postgres --format '{{json .Mounts}}' | Select-String "docker-entrypoint-initdb.d"
```

Expected: no match.

### Task 31: Contract and Module Runtime Gates

**Files:** none unless verification exposes a defect in task-owned files.

- [ ] **Step 1: Run module contract sync checks**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module documents
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module registry
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module templates
```

Expected: all pass.

- [ ] **Step 2: Run dictionary coverage**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-dictionary-coverage.ps1
```

Expected: coverage passes.

- [ ] **Step 3: Run Go tests for touched migration/startup code**

Run:

```powershell
go test ./internal/platform/migrate/... ./internal/modules/registry/... -count=1
```

Expected: tests pass.

### Task 32: Final Scope Review

**Files:** none

- [ ] **Step 1: Inspect git status**

Run:

```powershell
git status --short --branch
```

Expected: no unstaged changes unless final verification generated ignored evidence files.

- [ ] **Step 2: Inspect changed file summary**

Run:

```powershell
git diff --stat origin/main..HEAD
```

Expected: changes are limited to DB artifacts, scripts, docs/wiki, skills/instructions, and narrowly scoped migration runner/startup policy.

- [ ] **Step 3: Write final implementation summary**

Summarize:

```text
- research evidence completed
- curated baseline files created
- Docker entrypoint migration replay removed
- explicit fresh bootstrap and legacy replay paths documented
- database dictionary created
- skills and top-level instructions updated
- verification gates run with results
```

---

## Execution Handoff

Recommended execution option: **Subagent-Driven**.

Execution order:

1. Coordinator runs Tasks 1-2.
2. Dispatch five read-only research agents for Tasks 3-7 in parallel.
3. Coordinator integrates research in Task 8.
4. Stop for user approval at Task 9.
5. After approval, run implementation lanes in parallel where write ownership is disjoint.
6. Coordinator owns final verification Tasks 28-32.

Do not skip the Task 9 approval gate. The inclusion catalog is the point where MetalDocs decides what belongs in the clean DB.
