# Group B — Authz Cleanup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the 6 follow-up authz bugs (B1-B6) from the deep audit by formalising a two-tier authz contract, propagating `tenant_id` through `RoleProvider` and `RoleAdminRepository`, fixing the ON CONFLICT mismatch, and adding typed errors for missing GUC context.

**Architecture:** Two-tier authz contract: `iam_user_roles` answers tenant-level capability checks (HTTP middleware via `CapabilityService`); `user_process_areas` answers area-scoped checks (in-tx via `authz.Require`). Both consult `role_capabilities`. `system_admin` bypasses both. ADR documents the boundary. All IAM queries gain `tenant_id` filter; repo writes use DELETE-then-INSERT (consistent with the new `UNIQUE(tenant_id, user_id)` constraint).

**Tech Stack:** Go 1.22, PostgreSQL 16, `database/sql`, sqlmock for unit tests, integration tests via `//go:build integration` tag.

**Spec:** `docs/superpowers/specs/2026-05-03-group-b-authz-cleanup-design.md`

**Code style:** /simplify (no unnecessary abstractions). **Prompts:** /caveman (terse, direct).

---

## Model strategy

| Phase / Task | Model | Reason | Parallel with |
|---|---|---|---|
| Phase 0.1 worktree | sonnet | mechanical | — |
| Phase 0.2 codex plan validation | codex | independent review | — |
| Phase 0.3 wiki-curator agent verify | sonnet | trivial existence check | — |
| B4 helpers | codex | new public API, careful typing | B3, B2 |
| B3 migration | sonnet | one-line idempotent SQL | B4, B2 |
| B2 ADR + doc comments | sonnet | prose | B4, B3 |
| B5 RolesByUserID propagation | codex | wide-cascade signature change | — |
| B6 HasAnyRole propagation | codex | wide-cascade, builds on B5 | — |
| B1 repo writes + tenant_id | codex | most invasive (8 callers) | — |
| Phase 5 verify | sonnet | run commands | — |
| Phase 5 codex audit | codex | independent verification | — |
| Wiki-curator after each phase | wiki-curator agent (haiku) | doc-only | — |
| Phase reviews | opus | architectural eye | — |

---

## File map

### Created
| File | Purpose |
|---|---|
| `migrations/0170_dev_approver_role_correction.sql` | Flip dev `approver` user from `system_admin` back to `approver` |
| `wiki/decisions/0007-two-tier-authz.md` | ADR documenting tenant-tier vs area-tier authz |
| `internal/modules/iam/authz/context.go` | `MustActorID`/`MustTenantID` helpers + typed errors |
| `internal/modules/iam/authz/context_test.go` | Unit tests for helpers |
| `internal/modules/iam/infrastructure/postgres/role_admin_repository_test.go` | sqlmock unit tests for B1/B6 |
| `internal/modules/iam/infrastructure/postgres/role_provider_test.go` | sqlmock unit tests for B5 |
| `tests/integration/iam/tenant_isolation_test.go` | Cross-tenant isolation integration tests |
| `tests/integration/iam/migration_0170_test.go` | Migration 0170 integration test |

### Modified
| File | Change |
|---|---|
| `internal/modules/iam/domain/port.go` | Add `tenantID` param to `RolesByUserID`, `HasAnyRole`, `UpsertUserAndAssignRole`, `ReplaceUserRoles` |
| `internal/modules/iam/authz/authz.go` | Use `MustActorID`/`MustTenantID` helpers; remove inline `current_setting(.., false)` |
| `internal/modules/iam/application/capability_service.go` | Doc comment cross-references ADR |
| `internal/modules/iam/infrastructure/postgres/role_provider.go` | Add `tenant_id` filter, signature update |
| `internal/modules/iam/application/cached_role_provider.go` | Cache key keyed on `userID|tenantID`, signature update |
| `internal/modules/iam/application/dev_role_provider.go` | Signature update (no behaviour change in dev) |
| `internal/modules/iam/infrastructure/postgres/role_admin_repository.go` | DELETE-then-INSERT, all writes pass tenant_id |
| `internal/modules/iam/infrastructure/memory/role_admin_repository.go` | Mirror new signatures |
| `internal/modules/iam/application/admin_service.go` | Pass tenant_id from request context |
| `internal/modules/iam/delivery/http/admin_handler.go` | Resolve tenant_id from header/default, pass through |
| `internal/modules/iam/delivery/http/middleware.go` | Pass tenant_id to `RolesByUserID` |
| `internal/modules/auth/infrastructure/memory/repository.go` | Mirror new signatures |
| `internal/modules/auth/application/service.go` | Pass tenant_id at three call sites |
| `internal/platform/bootstrap/api.go` | Pass bootstrap tenant_id (default) |
| `apps/api/cmd/metaldocs-e2e-seed/main.go` | Pass seed tenant_id |
| `wiki/concepts/authz-tiers.md` | Light prose explaining the two-tier model (cross-link from ADR) |
| `wiki/references/dev-credentials.md` | Note approver user role is `approver`, not `system_admin` |
| `wiki/bugs/audit-2026-05-03.md` | Mark B1-B6 fixed with commit SHAs (final phase) |

### Deleted
None.

---

## Phase 0: Setup, Validation, Wiki-Curator

### Task 0.1: Create worktree

**Model:** sonnet

- [ ] **Step 1: Create worktree off main**

```bash
cd C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs
git worktree add ../MetalDocs-group-b -b group-b-authz-cleanup main
```

Expected: worktree created at `../MetalDocs-group-b`, branch `group-b-authz-cleanup` checked out.

- [ ] **Step 2: Verify clean state**

```bash
cd ../MetalDocs-group-b
git status
go build ./...
```

Expected: `nothing to commit, working tree clean`. `go build` succeeds.

---

### Task 0.2: Codex plan validation

**Model:** codex

- [ ] **Step 1: Dispatch codex agent**

Dispatch `codex:codex-rescue` agent with this prompt (terse, /caveman style):

> Validate plan: `docs/superpowers/plans/2026-05-03-group-b-authz-cleanup.md`. Spec: `docs/superpowers/specs/2026-05-03-group-b-authz-cleanup-design.md`. Audit doc: `wiki/bugs/audit-2026-05-03.md` lines 111-136.
>
> Goal: confirm 6 bugs (B1-B6) covered by plan, signature changes consistent across cascade, no missed callers, no contradictions with two-tier architecture from spec.
>
> For each bug: state PASS/FAIL with evidence (file:line of plan task that fixes it). Flag any: missed callers (grep `RolesByUserID`, `HasAnyRole`, `UpsertUserAndAssignRole`, `ReplaceUserRoles`), inconsistent signatures between port.go and impl, missing migration verification.
>
> Output: 6 lines (one per bug) + flag list. No fluff.

- [ ] **Step 2: Apply feedback inline**

If codex flags missed callers or inconsistencies, edit plan tasks. Re-dispatch only if structural rework needed.

- [ ] **Step 3: Commit plan**

```bash
git add docs/superpowers/plans/2026-05-03-group-b-authz-cleanup.md
git commit -m "plan: validated Group B authz cleanup plan via codex"
```

---

### Task 0.3: Verify wiki-curator subagent exists

**Model:** sonnet

- [ ] **Step 1: Check agent file**

```bash
ls .claude/agents/wiki-curator.md
```

Expected: file exists (created during Group A).

- [ ] **Step 2: If missing, recreate**

If the file does not exist, create `.claude/agents/wiki-curator.md` with the standard frontmatter (model: haiku, tools: Read/Write/Edit/Glob/Grep/Bash) and body matching the description in `CLAUDE.md` ("After refactors / new implementations, dispatch the wiki-curator agent..."). Same definition as Group A.

---

## Phase 1: Foundation (parallel — B4, B3, B2)

These three tasks touch zero overlapping files. Dispatch all three in parallel as separate subagents.

### Task B4: GUC strict-mode helpers + authz.Require integration

**Model:** codex

**Files:**
- Create: `internal/modules/iam/authz/context.go`
- Create: `internal/modules/iam/authz/context_test.go`
- Modify: `internal/modules/iam/authz/authz.go`

- [ ] **Step 1: Write failing test for `MustActorID`**

Create `internal/modules/iam/authz/context_test.go`:

```go
package authz_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/iam/authz"
)

func TestMustActorID_ReturnsErrWhenGUCMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.actor_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(""))

	_, err = authz.MustActorID(context.Background(), tx)
	if !errors.Is(err, authz.ErrActorContextMissing) {
		t.Fatalf("expected ErrActorContextMissing, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestMustActorID_ReturnsValueWhenSet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.actor_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow("user-123"))

	got, err := authz.MustActorID(context.Background(), tx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "user-123" {
		t.Fatalf("got %q, want user-123", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestMustTenantID_ReturnsErrWhenGUCMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT current_setting('metaldocs.tenant_id', true)")).
		WillReturnRows(sqlmock.NewRows([]string{"current_setting"}).AddRow(""))

	_, err = authz.MustTenantID(context.Background(), tx)
	if !errors.Is(err, authz.ErrTenantContextMissing) {
		t.Fatalf("expected ErrTenantContextMissing, got %v", err)
	}
}
```

- [ ] **Step 2: Run test — expect FAIL (helpers undefined)**

```bash
go test -mod=mod ./internal/modules/iam/authz/ -run TestMustActorID -v
```

Expected: compile error, `authz.MustActorID undefined`.

- [ ] **Step 3: Implement helpers**

Create `internal/modules/iam/authz/context.go`:

```go
package authz

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrActorContextMissing indicates the metaldocs.actor_id GUC was not set on the
// current transaction. Callers must `SET LOCAL metaldocs.actor_id = '<userID>'`
// before invoking authz functions.
var ErrActorContextMissing = errors.New("authz: metaldocs.actor_id GUC not set on transaction")

// ErrTenantContextMissing indicates the metaldocs.tenant_id GUC was not set on
// the current transaction.
var ErrTenantContextMissing = errors.New("authz: metaldocs.tenant_id GUC not set on transaction")

// MustActorID returns the metaldocs.actor_id GUC value for the given transaction.
// Returns ErrActorContextMissing if the GUC is unset or empty.
func MustActorID(ctx context.Context, tx *sql.Tx) (string, error) {
	var v string
	if err := tx.QueryRowContext(ctx, "SELECT current_setting('metaldocs.actor_id', true)").Scan(&v); err != nil {
		return "", fmt.Errorf("read actor_id GUC: %w", err)
	}
	if v == "" {
		return "", ErrActorContextMissing
	}
	return v, nil
}

// MustTenantID returns the metaldocs.tenant_id GUC value for the given transaction.
// Returns ErrTenantContextMissing if the GUC is unset or empty.
func MustTenantID(ctx context.Context, tx *sql.Tx) (string, error) {
	var v string
	if err := tx.QueryRowContext(ctx, "SELECT current_setting('metaldocs.tenant_id', true)").Scan(&v); err != nil {
		return "", fmt.Errorf("read tenant_id GUC: %w", err)
	}
	if v == "" {
		return "", ErrTenantContextMissing
	}
	return v, nil
}
```

- [ ] **Step 4: Run test — expect PASS**

```bash
go test -mod=mod ./internal/modules/iam/authz/ -run TestMustActorID -v
go test -mod=mod ./internal/modules/iam/authz/ -run TestMustTenantID -v
```

Expected: PASS.

- [ ] **Step 5: Refactor `authz.Require` to use helpers**

Modify `internal/modules/iam/authz/authz.go`. Replace the existing `Require` function body. Move the inline `current_setting('metaldocs.actor_id', false)` calls in SQL to use `$3`/`$4` parameters bound from the helpers:

```go
func Require(ctx context.Context, tx *sql.Tx, capability, areaCode string) error {
	if cacheGranted(ctx, capability, areaCode) {
		return appendAssertedCap(ctx, tx, capability, areaCode)
	}

	actorID, err := MustActorID(ctx, tx)
	if err != nil {
		return err
	}
	tenantID, err := MustTenantID(ctx, tx)
	if err != nil {
		return err
	}

	// system_admin bypass — check before capability query
	var isAdmin bool
	if err := tx.QueryRowContext(ctx, `
SELECT EXISTS (
  SELECT 1 FROM metaldocs.iam_user_roles
   WHERE user_id   = $1
     AND tenant_id = $2::uuid
     AND role_code = 'system_admin'
)`, actorID, tenantID).Scan(&isAdmin); err != nil {
		return fmt.Errorf("authz: system_admin check: %w", err)
	}
	if isAdmin {
		storeGranted(ctx, capability, areaCode)
		return appendAssertedCap(ctx, tx, capability, areaCode)
	}

	var granted bool
	err = tx.QueryRowContext(ctx, `
SELECT EXISTS (
  SELECT 1
  FROM metaldocs.role_capabilities rc
  JOIN metaldocs.user_process_areas upa
    ON upa.role = rc.role
   AND upa.tenant_id = $4::uuid
   AND upa.user_id   = $3
   AND upa.effective_to IS NULL
  WHERE rc.capability = $1
    AND ($2 = 'tenant' OR upa.area_code = $2)
)`, capability, areaCode, actorID, tenantID).Scan(&granted)
	if err != nil {
		return err
	}

	if !granted {
		return ErrCapabilityDenied{Capability: capability, AreaCode: areaCode, ActorID: actorID}
	}

	storeGranted(ctx, capability, areaCode)
	return appendAssertedCap(ctx, tx, capability, areaCode)
}
```

Also update `actorIDFromTx` (line 124) — delete it; callers inside this file now use `MustActorID`. Update the existing `ErrCapabilityDenied` path that called it.

- [ ] **Step 6: Update `BypassSystem` and `appendAssertedCap` if they use removed helper**

Search file for `actorIDFromTx`. Replace any internal use with `MustActorID`. Re-run tests.

```bash
go test -mod=mod ./internal/modules/iam/authz/... -v
```

Expected: PASS.

- [ ] **Step 7: Add doc comment cross-referencing ADR (B2 prep)**

Add doc comment block above `Require`:

```go
// Require enforces a tier-2 area-scoped authz check inside the caller's transaction.
// See wiki/decisions/0007-two-tier-authz.md for the boundary between tier-1
// (CapabilityService.CanDo, HTTP middleware) and tier-2 (this function).
//
// Callers MUST set the metaldocs.actor_id and metaldocs.tenant_id GUCs on the
// transaction before invoking. Use MustActorID/MustTenantID helpers if reading
// them from elsewhere.
//
// Pass areaCode = "tenant" to skip the area filter (degenerates to a tier-1
// equivalent inside the transaction).
```

- [ ] **Step 8: Commit**

```bash
git add internal/modules/iam/authz/context.go \
        internal/modules/iam/authz/context_test.go \
        internal/modules/iam/authz/authz.go
git commit -m "fix(B4): typed GUC errors via MustActorID/MustTenantID helpers"
```

---

### Task B3: Migration 0170 + dev creds wiki

**Model:** sonnet

**Files:**
- Create: `migrations/0170_dev_approver_role_correction.sql`
- Create: `tests/integration/iam/migration_0170_test.go`
- Modify: `wiki/references/dev-credentials.md`

- [ ] **Step 1: Write failing integration test**

Create `tests/integration/iam/migration_0170_test.go`:

```go
//go:build integration

package iam_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"metaldocs/tests/integration/testdb"
)

func TestMigration0170_FlipsApproverFromSystemAdminToApprover(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()

	const tenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"

	// Seed: approver user with system_admin role (simulates pre-0170 state)
	if _, err := db.ExecContext(ctx, `
DELETE FROM metaldocs.iam_user_roles WHERE user_id = 'approver';
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code)
VALUES ('approver', $1::uuid, 'system_admin')
`, tenantID); err != nil {
		t.Fatalf("seed: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(),
			`DELETE FROM metaldocs.iam_user_roles WHERE user_id = 'approver'`) //nolint:errcheck
	})

	sqlBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "migrations", "0170_dev_approver_role_correction.sql"))
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if _, err := db.ExecContext(ctx, string(sqlBytes)); err != nil {
		t.Fatalf("apply 0170: %v", err)
	}

	var role string
	if err := db.QueryRowContext(ctx,
		`SELECT role_code FROM metaldocs.iam_user_roles WHERE user_id = 'approver'`).
		Scan(&role); err != nil {
		t.Fatalf("query role: %v", err)
	}
	if role != "approver" {
		t.Fatalf("got role %q, want approver", role)
	}
}

func TestMigration0170_Idempotent(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()
	const tenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"

	if _, err := db.ExecContext(ctx, `
DELETE FROM metaldocs.iam_user_roles WHERE user_id = 'approver';
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code)
VALUES ('approver', $1::uuid, 'system_admin')
`, tenantID); err != nil {
		t.Fatalf("seed: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(),
			`DELETE FROM metaldocs.iam_user_roles WHERE user_id = 'approver'`) //nolint:errcheck
	})

	sqlBytes, _ := os.ReadFile(filepath.Join("..", "..", "..", "migrations", "0170_dev_approver_role_correction.sql"))
	if _, err := db.ExecContext(ctx, string(sqlBytes)); err != nil {
		t.Fatalf("first apply: %v", err)
	}
	if _, err := db.ExecContext(ctx, string(sqlBytes)); err != nil {
		t.Fatalf("second apply (idempotency): %v", err)
	}

	var role string
	db.QueryRowContext(ctx,
		`SELECT role_code FROM metaldocs.iam_user_roles WHERE user_id = 'approver'`).Scan(&role)
	if role != "approver" {
		t.Fatalf("got role %q, want approver", role)
	}
}
```

- [ ] **Step 2: Run test — expect FAIL (migration file does not exist)**

```bash
go test -mod=mod -tags=integration ./tests/integration/iam/ -run TestMigration0170 -v
```

Expected: FAIL with file-not-found error from `os.ReadFile`.

- [ ] **Step 3: Create migration**

Create `migrations/0170_dev_approver_role_correction.sql`:

```sql
-- 0170_dev_approver_role_correction.sql
-- Rolls back migration 0166's blanket admin->system_admin rename for the dev
-- approver seed user (introduced by 0159). Restores SoD by ensuring approver
-- and admin-local are distinct roles in dev environments.
--
-- Idempotent: only flips rows that are still incorrectly system_admin.

UPDATE metaldocs.iam_user_roles
   SET role_code = 'approver'
 WHERE user_id = 'approver'
   AND role_code = 'system_admin';
```

- [ ] **Step 4: Run test — expect PASS**

```bash
go test -mod=mod -tags=integration ./tests/integration/iam/ -run TestMigration0170 -v
```

Expected: PASS.

- [ ] **Step 5: Update dev creds wiki**

Edit `wiki/references/dev-credentials.md`. Find the approver row. Update the role column to read `approver` (not `admin` or `system_admin`). Add a note: "Migration 0170 enforces SoD by keeping the approver dev user at role `approver`. Migration 0166's blanket rename caught this user incorrectly; 0170 corrects it."

If the file does not exist, create it minimally:

```markdown
# Dev Credentials

> **Last verified:** 2026-05-03

| User | Password | Role | Notes |
|---|---|---|---|
| admin | AdminMetalDocs123! | system_admin | Bootstrap admin (admin-local) |
| approver | ApproverMetalDocs123! | approver | Used for SoD testing — distinct from admin |

Migration 0170 ensures the approver user keeps role `approver` after migration 0166's blanket admin→system_admin rename.
```

- [ ] **Step 6: Commit**

```bash
git add migrations/0170_dev_approver_role_correction.sql \
        tests/integration/iam/migration_0170_test.go \
        wiki/references/dev-credentials.md
git commit -m "fix(B3): correct dev approver role via migration 0170"
```

---

### Task B2: ADR + doc comment

**Model:** sonnet

**Files:**
- Create: `wiki/decisions/0007-two-tier-authz.md`
- Create: `wiki/concepts/authz-tiers.md`
- Modify: `internal/modules/iam/application/capability_service.go`

- [ ] **Step 1: Write ADR**

Create `wiki/decisions/0007-two-tier-authz.md`:

```markdown
# ADR 0007 — Two-Tier Authorization

> **Status:** accepted 2026-05-03
> **Last verified:** 2026-05-03

## Context

MetalDocs has two authz services that each consult `role_capabilities`:

1. `CapabilityService.CanDo(ctx, userID, tenantID, capability)` — reads `iam_user_roles`. Used by HTTP middleware to gate route access.
2. `authz.Require(ctx, tx, capability, areaCode)` — reads `user_process_areas`. Used inside service-layer transactions to gate area-scoped writes (signoff, area-bound state changes).

The 2026-05-02 IAM unification plan attempted to consolidate both. It unified the middleware path (`StaticAuthorizer` → `CapabilityService`) but left `authz.Require` reading `user_process_areas`. This produced apparent dual systems and confused engineers.

## Decision

Treat the two services as **distinct tiers** with explicit responsibilities, not as a unification gap.

| Tier | Service | Table | Used by | Question answered |
|---|---|---|---|---|
| 1 — Tenant | `CapabilityService.CanDo` | `iam_user_roles` JOIN `role_capabilities` | HTTP middleware | "Can user X do `doc.create` in tenant T?" |
| 2 — Area | `authz.Require` | `user_process_areas` JOIN `role_capabilities` | Service layer (in-tx) | "Can user X sign for area QA-01?" |

`role_capabilities` is shared: both tiers map a role code to a capability set. Roles in `iam_user_roles` are tenant-scoped; roles in `user_process_areas` are area-scoped.

`system_admin` bypasses both tiers.

`authz.Require` accepts `areaCode = "tenant"` as a degenerate form: skips the area filter, behaves like a tier-1 check inside the transaction.

## Consequences

- **Positive:** matches the QMS domain (ISO 9001 segregation requires per-area approver grants); clear contract; no schema migration needed.
- **Negative:** IAM admin UI must distinguish "tenant role" assignment from "area membership" assignment — separate flows.
- **Open:** future area-membership service consolidation if area grants diverge from `user_process_areas`.

## References

- `internal/modules/iam/application/capability_service.go` — tier 1
- `internal/modules/iam/authz/authz.go` — tier 2
- Migration 0162 — added `tenant_id` to `iam_user_roles`
- Migration 0165 — reseeded `role_capabilities`
- Audit 2026-05-03 — bugs B1-B6 (`wiki/bugs/audit-2026-05-03.md` lines 111-136)
```

- [ ] **Step 2: Write concept doc**

Create `wiki/concepts/authz-tiers.md`:

```markdown
# Authz Tiers

> **Last verified:** 2026-05-03
> See ADR `wiki/decisions/0007-two-tier-authz.md` for the decision rationale.

MetalDocs has **two authorization tiers**.

## Tier 1 — Tenant Capability

- **Where:** HTTP middleware (`internal/modules/iam/delivery/http/middleware.go`)
- **Service:** `CapabilityService.CanDo(ctx, userID, tenantID, capability)`
- **Tables:** `iam_user_roles` JOIN `role_capabilities`
- **Use:** "Can user X do `doc.create` in tenant T?"
- **Bypass:** `system_admin` role in tenant T

## Tier 2 — Area Grant

- **Where:** service layer, inside transactions (`internal/modules/iam/authz/authz.go`)
- **Service:** `authz.Require(ctx, tx, capability, areaCode)`
- **Tables:** `user_process_areas` JOIN `role_capabilities`
- **Use:** "Can user X sign for area QA-01?"
- **Special:** pass `areaCode = "tenant"` to skip area filter
- **Bypass:** `system_admin` role for the user

## When to use which

- **Route guards** (entry into HTTP handler): tier 1
- **Signoff, approval, area-scoped writes** (inside DB tx): tier 2
- **Both required** for area-scoped actions: middleware passes tier 1, then service layer enforces tier 2

## Common pitfalls

- Forgetting to set `metaldocs.actor_id`/`metaldocs.tenant_id` GUCs before calling `authz.Require` → `ErrActorContextMissing` (typed). Set via `SET LOCAL metaldocs.actor_id = '<userID>'` at start of tx.
- Assigning a tenant role via IAM admin UI does NOT grant area access. Area grants live in `user_process_areas`.
```

- [ ] **Step 3: Add doc comment to CapabilityService**

Modify `internal/modules/iam/application/capability_service.go`. Add doc comment above `CanDo`:

```go
// CanDo enforces a tier-1 tenant-level capability check.
// See wiki/decisions/0007-two-tier-authz.md for the boundary between tier-1
// (this function, HTTP middleware) and tier-2 (authz.Require, in-transaction).
//
// Returns nil if the user holds the capability in the given tenant via either:
//   - direct role in iam_user_roles
//   - membership in iam_groups + iam_group_roles
//
// system_admin role bypasses all capability checks.
//
// Returns ErrCapabilityDenied otherwise.
func (s *CapabilityService) CanDo(ctx context.Context, userID, tenantID, capability string) error {
```

- [ ] **Step 4: Commit**

```bash
git add wiki/decisions/0007-two-tier-authz.md \
        wiki/concepts/authz-tiers.md \
        internal/modules/iam/application/capability_service.go
git commit -m "docs(B2): ADR 0007 + concept doc + CapabilityService doc comment"
```

---

### Phase 1 review

**Model:** opus

- [ ] Dispatch opus reviewer with this prompt:

> Phase 1 of Group B authz cleanup landed three commits:
> - B4: GUC strict-mode helpers + authz.Require integration (typed errors)
> - B3: migration 0170 + dev creds wiki update
> - B2: ADR 0007 two-tier authz + concept doc + CapabilityService doc comment
>
> Read each commit (`git log -3`) and the spec at `docs/superpowers/specs/2026-05-03-group-b-authz-cleanup-design.md`. Verify:
> 1. B4: helpers correctly typed, `authz.Require` no longer uses `current_setting(.., false)` strict mode in inline SQL
> 2. B3: migration is idempotent + dev wiki updated + integration test exists
> 3. B2: ADR clearly distinguishes tier 1 vs tier 2 + doc comments added at both services
>
> Flag anything missed. 200 words max.

Apply feedback inline before Phase 2.

- [ ] **Dispatch wiki-curator**

Dispatch the `wiki-curator` agent to refresh `Last verified:` stamps on:
- `wiki/decisions/0007-two-tier-authz.md` (new file, no stamp refresh needed)
- `wiki/concepts/authz-tiers.md`
- `wiki/references/dev-credentials.md`
- `wiki/README.md` (add ADR 0007 + new concept doc to index)

---

## Phase 2: B5 — RolesByUserID propagation

**Model:** codex
**Sequential:** must precede B6 (port.go conflict).

**Files:**
- Modify: `internal/modules/iam/domain/port.go`
- Modify: `internal/modules/iam/infrastructure/postgres/role_provider.go`
- Modify: `internal/modules/iam/application/cached_role_provider.go`
- Modify: `internal/modules/iam/application/dev_role_provider.go`
- Modify: `internal/modules/iam/delivery/http/middleware.go`
- Modify: `internal/modules/auth/infrastructure/memory/repository.go`
- Modify: `internal/modules/auth/application/service.go`
- Create: `internal/modules/iam/infrastructure/postgres/role_provider_test.go`
- Create: `tests/integration/iam/tenant_isolation_test.go`

### Task B5.1: Failing sqlmock test for postgres `RoleProvider`

- [ ] **Step 1: Write failing test**

Create `internal/modules/iam/infrastructure/postgres/role_provider_test.go`:

```go
package postgres_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/iam/infrastructure/postgres"
)

const testTenant = "ffffffff-ffff-ffff-ffff-ffffffffffff"

func TestRolesByUserID_FiltersByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT is_active
FROM metaldocs.iam_users
WHERE user_id = $1
`)).WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"is_active"}).AddRow(true))

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT role_code
FROM metaldocs.iam_user_roles
WHERE user_id = $1
  AND tenant_id = $2::uuid
ORDER BY role_code ASC
`)).WithArgs("alice", testTenant).
		WillReturnRows(sqlmock.NewRows([]string{"role_code"}).AddRow("author"))

	provider := postgres.NewRoleProvider(db)
	roles, err := provider.RolesByUserID(context.Background(), "alice", testTenant)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roles) != 1 || string(roles[0]) != "author" {
		t.Fatalf("got %v, want [author]", roles)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
```

- [ ] **Step 2: Run test — expect FAIL**

```bash
go test -mod=mod ./internal/modules/iam/infrastructure/postgres/ -run TestRolesByUserID -v
```

Expected: FAIL — `RolesByUserID` signature mismatch.

### Task B5.2: Update domain port

- [ ] **Step 3: Modify `internal/modules/iam/domain/port.go`**

Replace:

```go
package domain

import "context"

// RoleProvider resolves effective roles for a given user identity within a tenant.
type RoleProvider interface {
	RolesByUserID(ctx context.Context, userID, tenantID string) ([]Role, error)
}

// RoleAdminRepository writes IAM user and role assignments.
type RoleAdminRepository interface {
	HasAnyRole(ctx context.Context, role Role, tenantID string) (bool, error)
	UpsertUserAndAssignRole(ctx context.Context, userID, displayName, tenantID string, role Role, assignedBy string) error
	ReplaceUserRoles(ctx context.Context, userID, displayName, tenantID string, roles []Role, assignedBy string) error
}
```

### Task B5.3: Update postgres `RoleProvider`

- [ ] **Step 4: Modify `internal/modules/iam/infrastructure/postgres/role_provider.go`**

Replace `RolesByUserID`:

```go
func (p *RoleProvider) RolesByUserID(ctx context.Context, userID, tenantID string) ([]domain.Role, error) {
	const checkUserSQL = `
SELECT is_active
FROM metaldocs.iam_users
WHERE user_id = $1
`
	var isActive bool
	if err := p.db.QueryRowContext(ctx, checkUserSQL, userID).Scan(&isActive); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("check iam user: %w", err)
	}
	if !isActive {
		return nil, domain.ErrUserInactive
	}

	const rolesSQL = `
SELECT role_code
FROM metaldocs.iam_user_roles
WHERE user_id = $1
  AND tenant_id = $2::uuid
ORDER BY role_code ASC
`
	rows, err := p.db.QueryContext(ctx, rolesSQL, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query iam roles: %w", err)
	}
	defer rows.Close()

	roles := make([]domain.Role, 0, 4)
	for rows.Next() {
		var roleCode string
		if err := rows.Scan(&roleCode); err != nil {
			return nil, fmt.Errorf("scan iam role: %w", err)
		}
		roles = append(roles, domain.Role(roleCode))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate iam roles: %w", err)
	}
	if len(roles) == 0 {
		return nil, domain.ErrUserNotFound
	}

	return roles, nil
}
```

### Task B5.4: Update cached + dev providers

- [ ] **Step 5: Modify `internal/modules/iam/application/cached_role_provider.go`**

Update cache to key on `userID|tenantID`:

```go
func cacheKey(userID, tenantID string) string {
	return userID + "|" + tenantID
}

func (c *CachedRoleProvider) RolesByUserID(ctx context.Context, userID, tenantID string) ([]domain.Role, error) {
	key := cacheKey(userID, tenantID)
	now := time.Now().UTC()

	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()

	if ok && now.Before(entry.expiresAt) {
		return cloneRoles(entry.roles), nil
	}

	roles, err := c.base.RolesByUserID(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.items[key] = cacheEntry{roles: cloneRoles(roles), expiresAt: now.Add(c.ttl)}
	c.mu.Unlock()

	return roles, nil
}

// InvalidateUser invalidates cache entries for a user across all tenants.
func (c *CachedRoleProvider) InvalidateUser(userID string) {
	c.mu.Lock()
	for k := range c.items {
		if strings.HasPrefix(k, userID+"|") {
			delete(c.items, k)
		}
	}
	c.mu.Unlock()
}
```

Note: add `"strings"` to imports.

- [ ] **Step 6: Modify `internal/modules/iam/application/dev_role_provider.go`**

```go
func (p *DevRoleProvider) RolesByUserID(_ context.Context, userID, _ string) ([]domain.Role, error) {
	id := strings.TrimSpace(userID)
	if id == "" {
		return nil, domain.ErrUserNotFound
	}
	roles, ok := p.rolesByUser[id]
	if !ok || len(roles) == 0 {
		return nil, domain.ErrUserNotFound
	}
	out := make([]domain.Role, len(roles))
	copy(out, roles)
	return out, nil
}
```

(tenantID intentionally ignored — DevRoleProvider is single-tenant test scaffold.)

### Task B5.5: Update HTTP middleware

- [ ] **Step 7: Modify `internal/modules/iam/delivery/http/middleware.go:91`**

Change:

```go
resolvedRoles, err := m.roleProvider.RolesByUserID(r.Context(), userID, tenantID)
```

`tenantID` was already resolved at line 77-80 (above the `m.caps.CanDo` block). Reuse the same variable; lift it out of the `if m.caps != nil` block if needed:

```go
tenantID := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
if tenantID == "" {
	tenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
}

if m.caps != nil {
	if err := m.caps.CanDo(r.Context(), userID, tenantID, capability); err != nil {
		writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", traceID)
		return
	}
}

ctx := r.Context()
if _, ok := authdomain.CurrentUserFromContext(ctx); !ok {
	var roles []iamdomain.Role
	if m.roleProvider != nil {
		resolvedRoles, err := m.roleProvider.RolesByUserID(r.Context(), userID, tenantID)
		// ... existing err handling
```

### Task B5.6: Update auth memory repo + service

- [ ] **Step 8: Modify `internal/modules/auth/infrastructure/memory/repository.go:315`**

Update signature:

```go
func (r *Repository) RolesByUserID(_ context.Context, userID, _ string) ([]iamdomain.Role, error) {
	// existing body unchanged — memory repo single-tenant
}
```

- [ ] **Step 9: Modify `internal/modules/auth/application/service.go:252,400`**

At line 252 (paginated user list, inside loop):

```go
roles, roleErr := s.roleProvider.RolesByUserID(ctx, items[i].UserID, tenantID)
```

`tenantID` is the active tenant for the request — find the variable in scope (likely passed into the function). If not in scope, add a `tenantID string` parameter to the enclosing function.

At line 400 (single user lookup):

```go
roles, err := s.roleProvider.RolesByUserID(ctx, userID, tenantID)
```

Same — add `tenantID` parameter to the enclosing function. Cascade caller at the handler layer.

### Task B5.7: Run sqlmock test — expect PASS

- [ ] **Step 10:**

```bash
go test -mod=mod ./internal/modules/iam/infrastructure/postgres/ -run TestRolesByUserID -v
```

Expected: PASS.

### Task B5.8: Cross-tenant integration test

- [ ] **Step 11: Create `tests/integration/iam/tenant_isolation_test.go`**

```go
//go:build integration

package iam_test

import (
	"context"
	"testing"

	"metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/tests/integration/testdb"
)

func TestRolesByUserID_RejectsCrossTenantBleed(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()

	tenantA := "11111111-1111-1111-1111-111111111111"
	tenantB := "22222222-2222-2222-2222-222222222222"
	userID := testdb.DeterministicID(t, "alice")

	// Seed: alice has author role in tenant A
	_, err := db.ExecContext(ctx, `
INSERT INTO metaldocs.iam_users (user_id, display_name, is_active)
VALUES ($1, 'Alice', TRUE)
ON CONFLICT (user_id) DO NOTHING;
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code)
VALUES ($1, $2::uuid, 'author')
`, userID, tenantA)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(),
			`DELETE FROM metaldocs.iam_user_roles WHERE user_id = $1`, userID) //nolint:errcheck
		db.ExecContext(context.Background(),
			`DELETE FROM metaldocs.iam_users WHERE user_id = $1`, userID) //nolint:errcheck
	})

	provider := postgres.NewRoleProvider(db)

	rolesA, err := provider.RolesByUserID(ctx, userID, tenantA)
	if err != nil {
		t.Fatalf("tenant A query: %v", err)
	}
	if len(rolesA) != 1 || string(rolesA[0]) != "author" {
		t.Fatalf("tenant A roles: got %v, want [author]", rolesA)
	}

	_, err = provider.RolesByUserID(ctx, userID, tenantB)
	if err == nil {
		t.Fatalf("tenant B query should have returned ErrUserNotFound, got nil")
	}
}
```

- [ ] **Step 12: Run integration test**

```bash
go test -mod=mod -tags=integration ./tests/integration/iam/ -run TestRolesByUserID_RejectsCrossTenantBleed -v
```

Expected: PASS (after dev DB has migrations applied).

### Task B5.9: Build the world

- [ ] **Step 13: Build everything to catch any missed callers**

```bash
go build -mod=mod ./...
```

Expected: PASS. If FAIL, the error names a caller of `RolesByUserID` not yet updated. Fix it (add `tenantID` param), repeat.

- [ ] **Step 14: Run full test suite**

```bash
go test -mod=mod ./...
```

Expected: PASS.

- [ ] **Step 15: Commit**

```bash
git add internal/modules/iam/domain/port.go \
        internal/modules/iam/infrastructure/postgres/role_provider.go \
        internal/modules/iam/infrastructure/postgres/role_provider_test.go \
        internal/modules/iam/application/cached_role_provider.go \
        internal/modules/iam/application/dev_role_provider.go \
        internal/modules/iam/delivery/http/middleware.go \
        internal/modules/auth/infrastructure/memory/repository.go \
        internal/modules/auth/application/service.go \
        tests/integration/iam/tenant_isolation_test.go
git commit -m "fix(B5): propagate tenant_id through RolesByUserID, prevent cross-tenant bleed"
```

---

## Phase 3: B6 — HasAnyRole propagation

**Model:** codex
**Sequential:** after B5 (shares port.go).

**Files:**
- Modify: `internal/modules/iam/domain/port.go` (already changed in B5 — verify HasAnyRole signature lands)
- Modify: `internal/modules/iam/infrastructure/postgres/role_admin_repository.go`
- Modify: `internal/modules/iam/infrastructure/memory/role_admin_repository.go`
- Modify: `internal/modules/auth/infrastructure/memory/repository.go`
- Modify: `internal/modules/auth/application/service.go:63`
- Create test entries in `internal/modules/iam/infrastructure/postgres/role_admin_repository_test.go` (new file — also used by B1)

### Task B6.1: Failing sqlmock test for HasAnyRole

- [ ] **Step 1: Create test file**

Create `internal/modules/iam/infrastructure/postgres/role_admin_repository_test.go`:

```go
package postgres_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/modules/iam/infrastructure/postgres"
)

func TestHasAnyRole_FiltersByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT COUNT(*)
FROM metaldocs.iam_user_roles
WHERE role_code = $1
  AND tenant_id = $2::uuid
`)).WithArgs("system_admin", testTenant).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	repo := postgres.NewRoleAdminRepository(db)
	got, err := repo.HasAnyRole(context.Background(), domain.Role("system_admin"), testTenant)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Fatalf("expected true, got false")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
```

- [ ] **Step 2: Run test — expect FAIL (signature mismatch)**

```bash
go test -mod=mod ./internal/modules/iam/infrastructure/postgres/ -run TestHasAnyRole -v
```

Expected: FAIL.

### Task B6.2: Update postgres impl

- [ ] **Step 3: Modify `internal/modules/iam/infrastructure/postgres/role_admin_repository.go`**

Replace `HasAnyRole`:

```go
func (r *RoleAdminRepository) HasAnyRole(ctx context.Context, role domain.Role, tenantID string) (bool, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM metaldocs.iam_user_roles
WHERE role_code = $1
  AND tenant_id = $2::uuid
`, string(role), tenantID).Scan(&count); err != nil {
		return false, fmt.Errorf("count role assignments: %w", err)
	}
	return count > 0, nil
}
```

### Task B6.3: Update memory impl

- [ ] **Step 4: Modify `internal/modules/iam/infrastructure/memory/role_admin_repository.go:24`**

```go
func (r *RoleAdminRepository) HasAnyRole(_ context.Context, role domain.Role, _ string) (bool, error) {
	// existing body — single-tenant memory store
}
```

- [ ] **Step 5: Modify `internal/modules/auth/infrastructure/memory/repository.go:360`**

```go
func (r *Repository) HasAnyRole(_ context.Context, role iamdomain.Role, _ string) (bool, error) {
	// existing body
}
```

### Task B6.4: Update bootstrap caller

- [ ] **Step 6: Modify `internal/modules/auth/application/service.go:63`**

Find `s.roleAdmin.HasAnyRole(ctx, iamdomain.RoleSystemAdmin)`. Change to:

```go
hasAdmin, err := s.roleAdmin.HasAnyRole(ctx, iamdomain.RoleSystemAdmin, tenantID)
```

`tenantID` must be in scope. If not, add it as a parameter to the enclosing function (`bootstrapAdmin` or similar — find via grep on call site). Cascade caller — likely `internal/platform/bootstrap/api.go` already has the tenant; pass it through.

If the bootstrap function has no tenant context, use the dev default `"ffffffff-ffff-ffff-ffff-ffffffffffff"` and add a comment:

```go
// Bootstrap runs in single-tenant dev mode. Use the default tenant.
const bootstrapTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
hasAdmin, err := s.roleAdmin.HasAnyRole(ctx, iamdomain.RoleSystemAdmin, bootstrapTenantID)
```

### Task B6.5: Build + test

- [ ] **Step 7: Build**

```bash
go build -mod=mod ./...
```

Expected: PASS.

- [ ] **Step 8: Run sqlmock test**

```bash
go test -mod=mod ./internal/modules/iam/infrastructure/postgres/ -run TestHasAnyRole -v
```

Expected: PASS.

### Task B6.6: Cross-tenant integration test

- [ ] **Step 9: Append to `tests/integration/iam/tenant_isolation_test.go`**

```go
func TestHasAnyRole_TenantIsolation(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()

	tenantA := "33333333-3333-3333-3333-333333333333"
	tenantB := "44444444-4444-4444-4444-444444444444"
	userID := testdb.DeterministicID(t, "alice-admin")

	_, err := db.ExecContext(ctx, `
INSERT INTO metaldocs.iam_users (user_id, display_name, is_active)
VALUES ($1, 'Alice Admin', TRUE)
ON CONFLICT (user_id) DO NOTHING;
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code)
VALUES ($1, $2::uuid, 'system_admin')
`, userID, tenantA)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(),
			`DELETE FROM metaldocs.iam_user_roles WHERE user_id = $1`, userID) //nolint:errcheck
		db.ExecContext(context.Background(),
			`DELETE FROM metaldocs.iam_users WHERE user_id = $1`, userID) //nolint:errcheck
	})

	repo := postgres.NewRoleAdminRepository(db)

	hasA, err := repo.HasAnyRole(ctx, "system_admin", tenantA)
	if err != nil {
		t.Fatalf("tenant A: %v", err)
	}
	if !hasA {
		t.Fatalf("tenant A: expected true, got false")
	}

	hasB, err := repo.HasAnyRole(ctx, "system_admin", tenantB)
	if err != nil {
		t.Fatalf("tenant B: %v", err)
	}
	if hasB {
		t.Fatalf("tenant B: expected false (cross-tenant bleed), got true")
	}
}
```

- [ ] **Step 10: Run integration test**

```bash
go test -mod=mod -tags=integration ./tests/integration/iam/ -run TestHasAnyRole_TenantIsolation -v
```

Expected: PASS.

- [ ] **Step 11: Commit**

```bash
git add internal/modules/iam/domain/port.go \
        internal/modules/iam/infrastructure/postgres/role_admin_repository.go \
        internal/modules/iam/infrastructure/postgres/role_admin_repository_test.go \
        internal/modules/iam/infrastructure/memory/role_admin_repository.go \
        internal/modules/auth/infrastructure/memory/repository.go \
        internal/modules/auth/application/service.go \
        tests/integration/iam/tenant_isolation_test.go
git commit -m "fix(B6): propagate tenant_id through HasAnyRole, isolate bootstrap admin per tenant"
```

---

## Phase 4: B1 — Repo writes (DELETE-then-INSERT, tenant_id everywhere)

**Model:** codex
**Sequential:** after B6 (shares port.go and same repo file).

**Files:**
- Modify: `internal/modules/iam/domain/port.go` (already changed in B5 — verify writes signatures land)
- Modify: `internal/modules/iam/infrastructure/postgres/role_admin_repository.go`
- Modify: `internal/modules/iam/infrastructure/memory/role_admin_repository.go`
- Modify: `internal/modules/iam/application/admin_service.go`
- Modify: `internal/modules/iam/delivery/http/admin_handler.go`
- Modify: `internal/modules/auth/infrastructure/memory/repository.go`
- Modify: `internal/modules/auth/application/service.go:88,315`
- Modify: `internal/platform/bootstrap/api.go:114,118`
- Modify: `apps/api/cmd/metaldocs-e2e-seed/main.go:86`
- Append tests to: `internal/modules/iam/infrastructure/postgres/role_admin_repository_test.go`

### Task B1.1: Failing sqlmock tests

- [ ] **Step 1: Append to `role_admin_repository_test.go`**

```go
func TestUpsertUserAndAssignRole_PassesTenantID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO metaldocs.iam_users`)).
		WithArgs("alice", "Alice").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// DELETE existing role rows for (tenant_id, user_id), then INSERT new
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM metaldocs.iam_user_roles WHERE tenant_id = $1::uuid AND user_id = $2`)).
		WithArgs(testTenant, "alice").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code, assigned_at, assigned_by)`)).
		WithArgs("alice", testTenant, "author", "admin").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	repo := postgres.NewRoleAdminRepository(db)
	err = repo.UpsertUserAndAssignRole(context.Background(), "alice", "Alice", testTenant, domain.Role("author"), "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestReplaceUserRoles_DeleteThenInsert_OnlyLastSurvives(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO metaldocs.iam_users`)).
		WithArgs("alice", "Alice").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM metaldocs.iam_user_roles WHERE tenant_id = $1::uuid AND user_id = $2`)).
		WithArgs(testTenant, "alice").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Only last role in slice is inserted (UNIQUE(tenant_id, user_id) enforces single row)
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code, assigned_at, assigned_by)`)).
		WithArgs("alice", testTenant, "approver", "admin").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	repo := postgres.NewRoleAdminRepository(db)
	err = repo.ReplaceUserRoles(context.Background(), "alice", "Alice", testTenant,
		[]domain.Role{domain.Role("author"), domain.Role("approver")}, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
```

- [ ] **Step 2: Run tests — expect FAIL (signature mismatch)**

```bash
go test -mod=mod ./internal/modules/iam/infrastructure/postgres/ -run TestUpsertUserAndAssignRole_PassesTenantID -v
go test -mod=mod ./internal/modules/iam/infrastructure/postgres/ -run TestReplaceUserRoles -v
```

Expected: FAIL.

### Task B1.2: Rewrite postgres repo

- [ ] **Step 3: Replace `UpsertUserAndAssignRole` and `ReplaceUserRoles` in `role_admin_repository.go`**

```go
func (r *RoleAdminRepository) UpsertUserAndAssignRole(ctx context.Context, userID, displayName, tenantID string, role domain.Role, assignedBy string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin iam tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const upsertUser = `
INSERT INTO metaldocs.iam_users (user_id, display_name, is_active, updated_at)
VALUES ($1, $2, TRUE, NOW())
ON CONFLICT (user_id)
DO UPDATE SET display_name = EXCLUDED.display_name, is_active = TRUE, updated_at = NOW()
`
	if _, err := tx.ExecContext(ctx, upsertUser, userID, displayName); err != nil {
		return fmt.Errorf("upsert iam user: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM metaldocs.iam_user_roles WHERE tenant_id = $1::uuid AND user_id = $2`,
		tenantID, userID); err != nil {
		return fmt.Errorf("delete prior iam roles: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code, assigned_at, assigned_by)
VALUES ($1, $2::uuid, $3, NOW(), $4)
`, userID, tenantID, string(role), assignedBy); err != nil {
		return fmt.Errorf("insert iam role: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit iam tx: %w", err)
	}
	return nil
}

// ReplaceUserRoles writes the user+role assignment. Note: the schema constraint
// UNIQUE(tenant_id, user_id) means at most ONE role row per user per tenant.
// If the input slice has multiple roles, only the last one is written (caller
// is expected to pass a single-element slice; multi-element slices accepted
// for backward compat with pre-0166 callers).
func (r *RoleAdminRepository) ReplaceUserRoles(ctx context.Context, userID, displayName, tenantID string, roles []domain.Role, assignedBy string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin iam replace tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const upsertUser = `
INSERT INTO metaldocs.iam_users (user_id, display_name, is_active, updated_at)
VALUES ($1, $2, TRUE, NOW())
ON CONFLICT (user_id)
DO UPDATE SET display_name = EXCLUDED.display_name, updated_at = NOW()
`
	if _, err := tx.ExecContext(ctx, upsertUser, userID, displayName); err != nil {
		return fmt.Errorf("upsert iam user: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM metaldocs.iam_user_roles WHERE tenant_id = $1::uuid AND user_id = $2`,
		tenantID, userID); err != nil {
		return fmt.Errorf("delete prior iam roles: %w", err)
	}

	// Find the last non-empty role
	var lastRole string
	for _, role := range roles {
		code := strings.TrimSpace(string(role))
		if code != "" {
			lastRole = code
		}
	}
	if lastRole == "" {
		// Caller passed empty slice — leave user with no role rows
		return tx.Commit()
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code, assigned_at, assigned_by)
VALUES ($1, $2::uuid, $3, NOW(), $4)
`, userID, tenantID, lastRole, assignedBy); err != nil {
		return fmt.Errorf("insert iam role: %w", err)
	}

	return tx.Commit()
}
```

Remove the old `existing/desired` map logic — no longer needed (constraint enforces single row).

### Task B1.3: Cascade callers

- [ ] **Step 4: Update memory impl**

`internal/modules/iam/infrastructure/memory/role_admin_repository.go` — add `tenantID string` parameter to both methods, ignore it (single-tenant memory).

```go
func (r *RoleAdminRepository) UpsertUserAndAssignRole(_ context.Context, userID, displayName, _ string, role domain.Role, _ string) error {
	// existing body unchanged
}

func (r *RoleAdminRepository) ReplaceUserRoles(_ context.Context, userID, displayName, _ string, roles []domain.Role, _ string) error {
	// existing body unchanged
}
```

- [ ] **Step 5: Update auth memory repo**

`internal/modules/auth/infrastructure/memory/repository.go:331,374` — same pattern, add `tenantID` parameter, ignore.

- [ ] **Step 6: Update `admin_service.go`**

Modify `internal/modules/iam/application/admin_service.go:23-65`:

```go
func (s *AdminService) UpsertUserAndAssignRole(ctx context.Context, userID, displayName, tenantID string, role domain.Role, assignedBy string) error {
	// validation
	if err := s.repo.UpsertUserAndAssignRole(ctx, userID, displayName, tenantID, role, assignedBy); err != nil {
		return err
	}
	// invalidations etc
}

func (s *AdminService) ReplaceUserRoles(ctx context.Context, userID, displayName, tenantID string, roles []domain.Role, assignedBy string) error {
	if err := s.repo.ReplaceUserRoles(ctx, userID, displayName, tenantID, roles, assignedBy); err != nil {
		return err
	}
}
```

- [ ] **Step 7: Update `admin_handler.go`**

Modify `internal/modules/iam/delivery/http/admin_handler.go:333,363`. Resolve tenant_id from `X-Tenant-ID` header (default UUID if empty), pass through:

```go
tenantID := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
if tenantID == "" {
	tenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
}

if err := h.service.UpsertUserAndAssignRole(r.Context(), userID, req.DisplayName, tenantID, role, assignedBy); err != nil {
	// existing err handling
}

// in handleReplaceUserRoles:
if err := h.service.ReplaceUserRoles(r.Context(), userID, req.DisplayName, tenantID, roles, assignedBy); err != nil {
	// existing err handling
}
```

- [ ] **Step 8: Update `auth/application/service.go:88,315`**

Find `s.roleAdmin.UpsertUserAndAssignRole(...)` and `s.roleAdmin.ReplaceUserRoles(...)`. Add `tenantID` argument from the enclosing function's scope. If the function has no tenantID parameter, add one and cascade up.

- [ ] **Step 9: Update `bootstrap/api.go:114,118`**

```go
const bootstrapTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
if err := authRepo.UpsertUserAndAssignRole(ctx, userID, userID, bootstrapTenantID, userRoles[0], "bootstrap"); err != nil {
	// existing
}
// ...
if err := authRepo.UpsertUserAndAssignRole(ctx, userID, userID, bootstrapTenantID, role, "bootstrap"); err != nil {
	// existing
}
```

- [ ] **Step 10: Update `metaldocs-e2e-seed/main.go:86`**

```go
const seedTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
if err := iamAdmin.UpsertUserAndAssignRole(ctx, seed.UserID, seed.DisplayName, seedTenantID, iamdomain.RoleSystemAdmin, "e2e-seed"); err != nil {
	// existing
}
```

### Task B1.4: Build + test

- [ ] **Step 11: Build**

```bash
go build -mod=mod ./...
```

Expected: PASS. If FAIL, missed caller — fix and repeat.

- [ ] **Step 12: Run sqlmock tests**

```bash
go test -mod=mod ./internal/modules/iam/infrastructure/postgres/ -v
```

Expected: PASS.

- [ ] **Step 13: Run full suite**

```bash
go test -mod=mod ./...
```

Expected: PASS.

### Task B1.5: Integration test for second-call same-tenant

- [ ] **Step 14: Append to `tests/integration/iam/tenant_isolation_test.go`**

```go
func TestReplaceUserRoles_TwoCallsSameTenantNoError(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()

	tenant := "55555555-5555-5555-5555-555555555555"
	userID := testdb.DeterministicID(t, "alice-replace")

	t.Cleanup(func() {
		db.ExecContext(context.Background(),
			`DELETE FROM metaldocs.iam_user_roles WHERE user_id = $1`, userID) //nolint:errcheck
		db.ExecContext(context.Background(),
			`DELETE FROM metaldocs.iam_users WHERE user_id = $1`, userID) //nolint:errcheck
	})

	repo := postgres.NewRoleAdminRepository(db)

	if err := repo.UpsertUserAndAssignRole(ctx, userID, "Alice", tenant, "author", "test"); err != nil {
		t.Fatalf("first call: %v", err)
	}
	// Second call — should overwrite cleanly, no UNIQUE violation
	if err := repo.UpsertUserAndAssignRole(ctx, userID, "Alice", tenant, "approver", "test"); err != nil {
		t.Fatalf("second call (overwrite): %v", err)
	}

	var role string
	if err := db.QueryRowContext(ctx, `
SELECT role_code FROM metaldocs.iam_user_roles
WHERE user_id = $1 AND tenant_id = $2::uuid
`, userID, tenant).Scan(&role); err != nil {
		t.Fatalf("query: %v", err)
	}
	if role != "approver" {
		t.Fatalf("got role %q, want approver", role)
	}
}
```

- [ ] **Step 15: Run integration**

```bash
go test -mod=mod -tags=integration ./tests/integration/iam/ -run TestReplaceUserRoles_TwoCallsSameTenantNoError -v
```

Expected: PASS.

- [ ] **Step 16: Commit**

```bash
git add internal/modules/iam/domain/port.go \
        internal/modules/iam/infrastructure/postgres/role_admin_repository.go \
        internal/modules/iam/infrastructure/postgres/role_admin_repository_test.go \
        internal/modules/iam/infrastructure/memory/role_admin_repository.go \
        internal/modules/iam/application/admin_service.go \
        internal/modules/iam/delivery/http/admin_handler.go \
        internal/modules/auth/infrastructure/memory/repository.go \
        internal/modules/auth/application/service.go \
        internal/platform/bootstrap/api.go \
        apps/api/cmd/metaldocs-e2e-seed/main.go \
        tests/integration/iam/tenant_isolation_test.go
git commit -m "fix(B1): tenant_id in iam writes, DELETE-then-INSERT for UNIQUE(tenant_id,user_id)"
```

### Phase 2-4 review

**Model:** opus

- [ ] Dispatch opus reviewer:

> Review commits B5/B6/B1 against `docs/superpowers/specs/2026-05-03-group-b-authz-cleanup-design.md`. For each, verify:
> - signature change consistently applied across cascade
> - tests cover happy path + cross-tenant isolation
> - no missed callers (search for old signature)
> - DELETE-then-INSERT pattern in B1 is atomic (inside tx)
>
> Read each commit (`git show <sha>`) and check the resulting file. Output: per-bug PASS/FAIL with file:line evidence. 200 words.

Apply feedback inline.

- [ ] Dispatch wiki-curator to refresh stamps on touched modules.

---

## Phase 5: Verify

### Task 5.1: Full Go test suite

- [ ] **Step 1: Run unit tests**

```bash
go test -mod=mod ./...
```

Expected: ALL PASS.

- [ ] **Step 2: Run integration tests**

```bash
go test -mod=mod -tags=integration ./tests/integration/iam/...
```

Expected: ALL PASS.

### Task 5.2: Codex independent audit

**Model:** codex

- [ ] **Step 1: Dispatch audit**

Dispatch `codex:codex-rescue` agent:

> Independent audit of Group B authz cleanup. Branch: group-b-authz-cleanup. 6 commits to verify (use `git log -10 --oneline`):
> - B4: fix(B4): typed GUC errors via MustActorID/MustTenantID helpers
> - B3: fix(B3): correct dev approver role via migration 0170
> - B2: docs(B2): ADR 0007 two-tier authz + concept doc + CapabilityService doc comment
> - B5: fix(B5): propagate tenant_id through RolesByUserID, prevent cross-tenant bleed
> - B6: fix(B6): propagate tenant_id through HasAnyRole, isolate bootstrap admin per tenant
> - B1: fix(B1): tenant_id in iam writes, DELETE-then-INSERT for UNIQUE(tenant_id,user_id)
>
> For each: read the diff (`git show <sha>`), open the resulting file, confirm:
> 1. Code matches intent (not just diff applied)
> 2. No missed callers (grep old signature)
> 3. Tests exist and assert correct behaviour
>
> Output: 6 lines (one per bug) PASS/FAIL with file:line evidence. Brief, no fluff.

Expected: 6/6 PASS.

### Task 5.3: Smoke test

- [ ] **Step 1: Bootstrap fresh dev DB**

```powershell
.\scripts\dev-migrate.ps1
```

Expected: migrations 0162-0170 apply cleanly.

- [ ] **Step 2: Verify dev approver role**

```sql
-- psql
SELECT user_id, role_code FROM metaldocs.iam_user_roles WHERE user_id = 'approver';
```

Expected: `role_code = 'approver'`.

- [ ] **Step 3: Login as approver, perform signoff**

```bash
.\scripts\start-api.ps1
```

Use Postman/curl:
- POST `/api/v1/auth/login` with `{"identifier":"approver","password":"ApproverMetalDocs123!"}`
- POST `/api/v1/approval/instances/<id>/signoff` (against existing instance from prior smoke)

Expected:
- Login: 200 with token
- Signoff: 200 (was 403 before B2 docs and B3 fix)
- Verify SoD: approver cannot approve a document they authored — should get 403 with clear error

### Task 5.4: Frontend regression

- [ ] **Step 1: Run vitest**

```bash
cd frontend/apps/web
npx vitest run
```

Expected: PASS.

---

## Phase 6: Merge & finalise

### Task 6.1: Update audit doc

- [ ] **Step 1: Modify `wiki/bugs/audit-2026-05-03.md`**

In the Group B table, fill the Status and Fix commit columns:

```markdown
| B1 | `ReplaceUserRoles` 500s ... | 🟠 high | fixed | <B1 sha> |
| B2 | Two authz models not synced ... | 🟠 high | fixed | <B2 sha> |
| B3 | Dev approver seed user has system_admin ... | 🟠 high | fixed | <B3 sha> |
| B4 | authz.Require GUC dependency ... | 🟠 high | fixed | <B4 sha> |
| B5 | RoleProvider.RolesByUserID ignores tenant_id ... | 🟠 high | fixed | <B5 sha> |
| B6 | HasAnyRole ignores tenant ... | 🟠 high | fixed | <B6 sha> |
```

Update the progress summary table: Group B → Fixed: 6, Open: 0. Total fixed: 16/40.

Add a session history row:

```markdown
| 2026-05-03 | Group B authz cleanup | All 6 Group B bugs fixed. Two-tier authz contract formalised (ADR 0007). Tenant_id propagated through RoleProvider/HasAnyRole/role_admin_repository. Migration 0170 corrects dev approver role. Typed GUC errors via MustActorID/MustTenantID. go test ./... + integration tests pass. Codex audit 6/6 PASS. |
```

- [ ] **Step 2: Commit**

```bash
git add wiki/bugs/audit-2026-05-03.md
git commit -m "docs(audit): mark Group B bugs B1-B6 fixed"
```

### Task 6.2: Wiki-curator full pass

- [ ] **Step 1: Dispatch wiki-curator agent**

Prompt:

> Group B authz cleanup landed on branch group-b-authz-cleanup. Refresh `Last verified: 2026-05-03` stamps on:
> - `wiki/concepts/authz-tiers.md` (new)
> - `wiki/decisions/0007-two-tier-authz.md` (new)
> - `wiki/references/dev-credentials.md`
> - `wiki/bugs/audit-2026-05-03.md`
> - `wiki/modules/iam-*.md` (any iam module docs)
> - `wiki/README.md` — add new ADR + concept doc to index
>
> For each touched Go file in this branch, find any wiki doc that references it via Key files anchors (file:line), update those anchors if line numbers shifted. Use `git diff main...HEAD --name-only` to see touched files.

### Task 6.3: Final Opus review

**Model:** opus

- [ ] **Step 1: Dispatch reviewer**

> Final review of Group B authz cleanup branch. Read:
> - Spec: `docs/superpowers/specs/2026-05-03-group-b-authz-cleanup-design.md`
> - Plan: `docs/superpowers/plans/2026-05-03-group-b-authz-cleanup.md`
> - All commits on branch (`git log main..HEAD`)
> - Audit doc: `wiki/bugs/audit-2026-05-03.md`
>
> Verify:
> 1. All 6 bugs (B1-B6) have a fix commit + test
> 2. Two-tier architecture from ADR 0007 is faithfully implemented (no logic changes to authz.Require beyond GUC helpers; no new schema)
> 3. Cross-tenant isolation tests cover RolesByUserID and HasAnyRole
> 4. Smoke test results recorded in audit doc
> 5. No regressions in Go test suite
> 6. No new lint warnings
>
> Output: ready-to-merge YES/NO + concerns list. 250 words.

Apply feedback if any concerns. Then proceed to merge.

### Task 6.4: Finishing

- [ ] **Step 1: Use finishing skill**

Invoke `superpowers:finishing-a-development-branch`. Choose Option 1 (merge locally) or Option 2 (PR) per user preference.

---

## Self-Review (run before handing off plan)

**1. Spec coverage:**
- B1 → Phase 4 Tasks B1.1-B1.5 ✓
- B2 → Phase 1 Task B2 (ADR + doc comments) ✓
- B3 → Phase 1 Task B3 (migration 0170) ✓
- B4 → Phase 1 Task B4 (helpers) ✓
- B5 → Phase 2 Tasks B5.1-B5.9 ✓
- B6 → Phase 3 Tasks B6.1-B6.6 ✓
- ADR 0007 → Phase 1 Task B2 Step 1 ✓
- Concept doc `authz-tiers.md` → Phase 1 Task B2 Step 2 ✓
- Cross-tenant isolation tests → Phase 2 Step 11, Phase 3 Step 9 ✓

**2. Placeholder scan:** none. All steps contain code or exact commands.

**3. Type consistency:**
- `RolesByUserID(ctx, userID, tenantID string)` — used identically in port, postgres, cached, dev, middleware, service ✓
- `HasAnyRole(ctx, role domain.Role, tenantID string)` — identical across port, postgres, memory, auth-memory ✓
- `UpsertUserAndAssignRole(ctx, userID, displayName, tenantID string, role domain.Role, assignedBy string)` — identical across port, postgres, memory, admin-service, auth-memory, bootstrap, e2e-seed ✓
- `ReplaceUserRoles(ctx, userID, displayName, tenantID string, roles []domain.Role, assignedBy string)` — identical ✓
- `MustActorID(ctx, tx) (string, error)` and `MustTenantID(ctx, tx) (string, error)` — used identically in `authz.Require` and tests ✓
- `ErrActorContextMissing`, `ErrTenantContextMissing` — typed errors used in helpers + tests ✓

---

## Done definition

- All 6 bugs B1-B6 marked fixed in `wiki/bugs/audit-2026-05-03.md` with commit SHAs
- ADR 0007 committed; concept doc `authz-tiers.md` committed
- `go test -mod=mod ./...` — all packages PASS
- `go test -mod=mod -tags=integration ./tests/integration/iam/...` — all PASS
- Codex audit returns 6/6 PASS
- Smoke: approver dev user has role `approver`, can sign off, SoD enforced
- Branch merged or PR created
