# Finish Database Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Finish the MetalDocs database foundation so fresh curated bootstrap, post-baseline migrations, runtime startup, and verification gates are trustworthy enough to unblock screen implementation.

**Architecture:** Keep the curated baseline model. Repair the post-baseline migration runner so the baseline ledger marker cannot block future migrations, remove default runtime dependence on legacy data backfill, and harden verification so the DB can prove both dev-seeded and no-dev-seed startup paths. Historical migrations remain evidence/recovery material only.

**Tech Stack:** Go `database/sql`, PowerShell scripts, PostgreSQL, Docker Compose, MetalDocs DB wiki/skills.

---

## Parallel Execution Map

These lanes are safe to run in parallel after a clean status check:

| Lane | Tasks | Write ownership | Dependency |
|---|---|---|---|
| A | Task 1 | `internal/platform/migrate/` | none |
| B | Task 2 | `apps/api/cmd/metaldocs-api/main.go`, `internal/modules/registry/` | none |
| C | Task 3 | `scripts/check-baseline-equivalence.ps1`, `scripts/check-db-bootstrap.ps1` | Task 1 for final fixture gate |
| D | Task 4 | docs/wiki/skill plan alignment | Task 1 and Task 2 decisions |
| E | Task 5 | verification only | Tasks 1-4 |

Do not run two workers against the same write ownership.

## File Structure

- `internal/platform/migrate/migrate.go` - post-baseline forward migration runner.
- `internal/platform/migrate/migrate_test.go` - unit tests proving baseline markers do not block numeric migrations.
- `apps/api/cmd/metaldocs-api/main.go` - API startup wiring; remove unconditional legacy maintenance from normal startup.
- `internal/modules/registry/module.go` - registry module maintenance entry points and naming.
- `internal/modules/registry/application/migration.go` - legacy backfill implementation remains available for explicit recovery only.
- `scripts/check-baseline-equivalence.ps1` - fail fast when comparing a DB to itself unless explicitly allowed.
- `scripts/check-db-bootstrap.ps1` - verify baseline marker and post-baseline tail behavior.
- `scripts/dev-local.ps1` - print the correct dev-seeded fresh bootstrap command.
- `.claude/skills/metaldocs-database/SKILL.md` - canonical DB workflow and definition of done.
- `.agents/skills/metaldocs-database/SKILL.md` - bridge skill; read-only unless bridge wording must change.
- `docs/runbooks/*.md`, `wiki/database/*.md`, `wiki/architecture/api-contract.md`, `wiki/modules/registry.md` - align supported workflow and verification commands.

---

### Task 0: Preflight and Preserve Current Skill Update

**Files:**
- Modify: `.claude/skills/metaldocs-database/SKILL.md`

- [ ] **Step 1: Inspect status**

Run:

```powershell
git status --short --branch
```

Expected: only known database skill edits are uncommitted before implementation begins.

- [ ] **Step 2: Validate skill**

Run:

```powershell
python C:\Users\leandro.theodoro.MN-NTB-LEANDROT\.codex\skills\.system\skill-creator\scripts\quick_validate.py .claude/skills/metaldocs-database
```

Expected: `Skill is valid!`

- [ ] **Step 3: Commit skill guardrails**

Run:

```powershell
git add .claude/skills/metaldocs-database/SKILL.md
git commit -m "docs(db): tighten database completion gates"
```

Expected: commit succeeds.

---

### Task 1: Fix Post-Baseline Migration Ledger Ordering

**Files:**
- Modify: `internal/platform/migrate/migrate.go`
- Create: `internal/platform/migrate/migrate_test.go`

- [ ] **Step 1: Add failing migration-runner test**

Create `internal/platform/migrate/migrate_test.go`:

```go
package migrate

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestApplyRunsNumericMigrationAfterBaselineMarker(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	dir := t.TempDir()
	migrationPath := filepath.Join(dir, "0202_after_baseline.sql")
	if err := os.WriteFile(migrationPath, []byte("SELECT 1;"), 0o600); err != nil {
		t.Fatalf("write migration: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT version FROM public.schema_migrations")).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).
			AddRow("baseline-2026-05-14"))
	mock.ExpectExec(regexp.QuoteMeta("SELECT 1;")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := Apply(context.Background(), db, dir, slog.Default()); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestApplySkipsAlreadyAppliedNumericMigration(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	dir := t.TempDir()
	migrationPath := filepath.Join(dir, "0202_after_baseline.sql")
	if err := os.WriteFile(migrationPath, []byte("SELECT 1;"), 0o600); err != nil {
		t.Fatalf("write migration: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT version FROM public.schema_migrations")).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).
			AddRow("baseline-2026-05-14").
			AddRow("0202"))

	if err := Apply(context.Background(), db, dir, slog.Default()); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}
```

- [ ] **Step 2: Run test and confirm failure**

Run:

```powershell
go test ./internal/platform/migrate/... -run TestApplyRunsNumericMigrationAfterBaselineMarker -count=1
```

Expected before implementation: FAIL because `SELECT 1;` is not executed.

- [ ] **Step 3: Replace mixed-ledger high-water logic**

Modify `internal/platform/migrate/migrate.go` so the apply loop relies on exact ledger membership for the target migration version:

```go
	skipped, ran := 0, 0
	for _, f := range files {
		if applied[f.version] {
			skipped++
			continue
		}
		body, err := os.ReadFile(f.path)
		if err != nil {
			return fmt.Errorf("read %s: %w", f.name, err)
		}
		log.Info("applying migration", "version", f.version, "file", f.name)
		if _, err := db.ExecContext(ctx, string(body)); err != nil {
			return fmt.Errorf("apply %s: %w", f.name, err)
		}
		ran++
	}
```

Also remove the obsolete `maxApplied` block and its comment.

- [ ] **Step 4: Run migrate tests**

Run:

```powershell
go test ./internal/platform/migrate/... -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit migration runner fix**

Run:

```powershell
git add internal/platform/migrate/migrate.go internal/platform/migrate/migrate_test.go
git commit -m "fix(db): apply forward migrations after baseline marker"
```

Expected: commit succeeds.

---

### Task 2: Remove Legacy Maintenance From Default API Startup

**Files:**
- Modify: `apps/api/cmd/metaldocs-api/main.go`
- Modify: `internal/modules/registry/module.go`
- Modify: `wiki/modules/registry.md`

- [ ] **Step 1: Add explicit env gate in API startup**

Modify `apps/api/cmd/metaldocs-api/main.go` around the registry startup section:

```go
	registryModule.RegisterRoutes(mux)
	if deps.SQLDB != nil && strings.EqualFold(strings.TrimSpace(os.Getenv("METALDOCS_RUN_LEGACY_REGISTRY_MAINTENANCE")), "true") {
		if err := registryModule.RunLegacyMaintenance(context.Background(), deps.SQLDB, slog.Default()); err != nil {
			log.Printf("registry legacy maintenance failed: %v", err)
		}
	}
```

Expected: no default call to `RunStartupMigrations`.

- [ ] **Step 2: Remove startup-migration alias**

Modify `internal/modules/registry/module.go` by deleting:

```go
func (m *Module) RunStartupMigrations(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	return m.RunLegacyMaintenance(ctx, db, logger)
}
```

Keep `RunLegacyMaintenance` available for explicit recovery.

- [ ] **Step 3: Update registry wiki runtime section**

In `wiki/modules/registry.md`, add a concise note in the runtime or persistence section:

```markdown
Legacy registry maintenance is not part of normal API startup. It may run only when `METALDOCS_RUN_LEGACY_REGISTRY_MAINTENANCE=true` is intentionally set for recovery on older databases.
```

- [ ] **Step 4: Search for removed startup name**

Run:

```powershell
rg -n "RunStartupMigrations|METALDOCS_RUN_LEGACY_REGISTRY_MAINTENANCE|BackfillLegacyDocuments" apps internal wiki/modules/registry.md
```

Expected: `RunStartupMigrations` has no matches; explicit env gate and backfill references remain.

- [ ] **Step 5: Run registry tests**

Run:

```powershell
go test ./internal/modules/registry/... -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit startup maintenance change**

Run:

```powershell
git add apps/api/cmd/metaldocs-api/main.go internal/modules/registry/module.go wiki/modules/registry.md
git commit -m "fix(db): gate registry legacy maintenance"
```

Expected: commit succeeds.

---

### Task 3: Harden Verification Scripts

**Files:**
- Modify: `scripts/check-baseline-equivalence.ps1`
- Modify: `scripts/check-db-bootstrap.ps1`
- Modify: `scripts/dev-local.ps1`

- [ ] **Step 1: Prevent accidental self-compare**

Modify `scripts/check-baseline-equivalence.ps1` params:

```powershell
param(
  [string]$ReferenceDb = "metaldocs_reference",
  [string]$CandidateDb = "metaldocs",
  [switch]$AllowSameDatabase
)
```

Add after `Set-Location $root`:

```powershell
if ($ReferenceDb -eq $CandidateDb -and -not $AllowSameDatabase) {
  throw "[check-baseline-equivalence] ReferenceDb and CandidateDb are both '$ReferenceDb'. Pass distinct DB names or -AllowSameDatabase for a smoke check."
}
```

- [ ] **Step 2: Add tail-migration fixture option to bootstrap check**

Modify `scripts/check-db-bootstrap.ps1` params:

```powershell
param(
  [switch]$WithDevSeed,
  [switch]$VerifyForwardMigration
)
```

Add after the baseline ledger marker check:

```powershell
if ($VerifyForwardMigration) {
  Write-Host "[check-db-bootstrap] Checking post-baseline forward migration execution..."
  $fixtureVersion = "9999"
  $fixtureFile = Join-Path $root "db/migrations/9999_check_forward_migration_fixture.sql"
  @"
BEGIN;
CREATE TABLE IF NOT EXISTS metaldocs.forward_migration_fixture (
  id text PRIMARY KEY
);
INSERT INTO public.schema_migrations(version, description)
VALUES ('9999', 'check forward migration fixture')
ON CONFLICT (version) DO NOTHING;
COMMIT;
"@ | Set-Content -Path $fixtureFile -Encoding UTF8
  try {
    powershell -NoProfile -ExecutionPolicy Bypass -File scripts/start-api.ps1 -Build -NoWorker
  } finally {
    Remove-Item -Force $fixtureFile -ErrorAction SilentlyContinue
    Get-Process -Name metaldocs-api -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
  }
  $fixtureLedger = docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -tAc "SELECT version FROM public.schema_migrations WHERE version = '$fixtureVersion';"
  if (($fixtureLedger | Out-String).Trim() -ne $fixtureVersion) {
    throw "[check-db-bootstrap] forward migration fixture ledger marker missing"
  }
}
```

If this foreground start blocks in practice, replace the inner startup with `scripts/check-system-runnable.ps1 -StartApi -TargetRoute /api/v1/health/ready` and keep the process cleanup.

- [ ] **Step 3: Fix local dev printed command**

Modify `scripts/dev-local.ps1` output:

```powershell
Write-Host "  0. Fresh local install: powershell -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1 -WithDevSeed"
Write-Host "     Product schema only: powershell -ExecutionPolicy Bypass -File scripts/dev-bootstrap-baseline.ps1"
Write-Host "     Historical recovery: powershell -ExecutionPolicy Bypass -File scripts/dev-migrate.ps1"
```

- [ ] **Step 4: Parse touched PowerShell scripts**

Run:

```powershell
& { $files = @('scripts/check-baseline-equivalence.ps1','scripts/check-db-bootstrap.ps1','scripts/dev-local.ps1'); foreach ($f in $files) { [scriptblock]::Create((Get-Content -Raw $f)) > $null; Write-Host "parsed $f" } }
```

Expected: all files parse.

- [ ] **Step 5: Run targeted script checks**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-baseline-equivalence.ps1 -AllowSameDatabase
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-bootstrap.ps1 -WithDevSeed
```

Expected: equivalence smoke passes; bootstrap passes.

- [ ] **Step 6: Commit verification hardening**

Run:

```powershell
git add scripts/check-baseline-equivalence.ps1 scripts/check-db-bootstrap.ps1 scripts/dev-local.ps1
git commit -m "test(db): harden baseline verification gates"
```

Expected: commit succeeds.

---

### Task 4: Align Docs, Skill, and Plan Truth

**Files:**
- Modify: `.claude/skills/metaldocs-database/SKILL.md`
- Modify: `docs/runbooks/database-bootstrap.md`
- Modify: `docs/runbooks/migration-baseline-local.md`
- Modify: `docs/runbooks/migration-legacy-replay.md`
- Modify: `wiki/database/migration-policy.md`
- Modify: `wiki/database/overview.md`
- Modify: `wiki/architecture/api-contract.md`

- [ ] **Step 1: Update database runbook verification**

In `docs/runbooks/database-bootstrap.md`, replace foreground two-command verification with:

```markdown
```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -StartApi -TargetRoute /api/v1/controlled-documents
```
```

- [ ] **Step 2: Update baseline local runbook**

In `docs/runbooks/migration-baseline-local.md`, make the command block:

```markdown
```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-bootstrap.ps1 -WithDevSeed
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -StartApi -TargetRoute /api/v1/controlled-documents
```
```

- [ ] **Step 3: Ensure historical replay language is recovery-only**

Search:

```powershell
rg -n "legacy replay|Legacy replay|Legacy Migration|normal fresh|/api/v2/controlled-documents" docs wiki .claude/skills .agents/skills
```

Expected: no supported normal workflow uses legacy replay; no `/api/v2/controlled-documents` verification remains.

- [ ] **Step 4: Validate database skill**

Run:

```powershell
python C:\Users\leandro.theodoro.MN-NTB-LEANDROT\.codex\skills\.system\skill-creator\scripts\quick_validate.py .claude/skills/metaldocs-database
```

Expected: `Skill is valid!`

- [ ] **Step 5: Commit docs and skill alignment**

Run:

```powershell
git add .claude/skills/metaldocs-database/SKILL.md docs/runbooks/database-bootstrap.md docs/runbooks/migration-baseline-local.md docs/runbooks/migration-legacy-replay.md wiki/database/migration-policy.md wiki/database/overview.md wiki/architecture/api-contract.md
git commit -m "docs(db): align database completion workflow"
```

Expected: commit succeeds.

---

### Task 5: Final DB Completion Gates

**Files:** none unless verification exposes a defect in task-owned files.

- [ ] **Step 1: Fresh curated bootstrap with dev seed**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-bootstrap.ps1 -WithDevSeed -VerifyForwardMigration
```

Expected: bootstrap, dictionary, baseline marker, and forward migration fixture checks pass.

- [ ] **Step 2: Product schema without dev seed**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-bootstrap.ps1
```

Expected: bootstrap succeeds without local users.

- [ ] **Step 3: Runtime auth/session and route gate**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-bootstrap.ps1 -WithDevSeed
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -StartApi -TargetRoute /api/v1/controlled-documents
```

Expected: login, session, `/api/v1/auth/me`, and `/api/v1/controlled-documents` pass.

- [ ] **Step 4: Contract and module gates**

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module documents
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module registry
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module templates
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-db-dictionary-coverage.ps1
go test ./internal/platform/migrate/... ./internal/modules/registry/... -count=1
```

Expected: all pass.

- [ ] **Step 5: Final scope review**

Run:

```powershell
git status --short --branch
git diff --stat origin/main..HEAD
```

Expected: no unstaged changes; diff is limited to DB artifacts, migration runner, scripts, docs/wiki, skills, and narrowly scoped startup policy.

- [ ] **Step 6: Request code review**

Use `superpowers:requesting-code-review` against the new fix range.

Expected: no Critical or Important findings remain before moving to screen implementation.

---

## Self-Review

Spec coverage:
- Forward migration ledger correctness is covered by Task 1 and Task 5.
- No default legacy startup mutation is covered by Task 2 and Task 5.
- Dev-seed and no-dev-seed bootstrap are covered by Task 3 and Task 5.
- Documentation and skill truth are covered by Task 4.
- Final review before screens is covered by Task 5.

Placeholder scan:
- No `TBD`, `TODO`, or unspecified test steps are present.

Type and command consistency:
- All paths match current repository paths.
- Runtime route uses `/api/v1/controlled-documents`.
- Historical migration language remains recovery/evidence-only.

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-05-15-finish-database-foundation.md`.

Two execution options:

1. **Subagent-Driven (recommended)** - dispatch separate workers for Task 1 and Task 2 in parallel, then Task 3 and Task 4 with review checkpoints.
2. **Inline Execution** - execute in this session with checkpoints after each task.
