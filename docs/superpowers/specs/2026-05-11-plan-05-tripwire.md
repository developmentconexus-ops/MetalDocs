# Plan 5 — Tier-2 `authz.Require` + Postgres Tripwire on Regulated Tables

> **For agentic workers:** REQUIRED SUB-SKILL: Use `nexus:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.
>
> **Agent routing (per project memory):** Use **Codex** for implementation tasks (editing Go/SQL source). Use **Haiku or Sonnet** for writing tests and creating git commits.

**Goal:** Extend the approval-module three-layer defense-in-depth pattern (tier-1 capability middleware + tier-2 `authz.Require` + Postgres `enforce_capability_asserted` trigger) to every regulated mutating table that currently lacks it.

**Architecture:** Every mutation on a regulated table must (1) pass the tier-1 HTTP path/method capability resolver in `apps/api/cmd/metaldocs-api/permissions.go`, (2) call `authz.Require(ctx, tx, cap, area)` inside a transaction before executing DML — which sets `metaldocs.asserted_caps` GUC — and (3) have the `enforce_capability_asserted` BEFORE INSERT/UPDATE trigger reject DML when the GUC is absent. The tripwire is the backstop: it fires even for out-of-band DB access (direct psql, alternate import paths, future test harnesses that forget to set up authz).

**Tech Stack:** Go 1.22, `database/sql`, `internal/modules/iam/authz` package (`authz.Require`), `internal/modules/iam/domain` typed `Capability` consts, PostgreSQL BEFORE-trigger pattern from `migrations/0142b_role_capabilities_v2_enforce.sql`.

**Prerequisite:** Plan 4 done (typed `iamdomain.Capability` constants stable in `internal/modules/iam/domain/model.go:16`).

---

## Key files reference

| File | Role |
|------|------|
| `internal/modules/iam/domain/model.go` | Typed `Capability` consts — add `CapRegistryObsolete`, `CapRegistrySupersede` here |
| `apps/api/cmd/metaldocs-api/permissions.go:174,190` | Tier-1 path-method resolver — add `MethodPatch` for families; swap `CapDocumentEdit` → new caps for obsolete/supersede |
| `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:33,72` | IAM mutation sites — add `authz.Require` inside existing internal txs |
| `internal/modules/iam/infrastructure/postgres/user_area_repository.go:51,75,90` | IAM area membership mutations — wrap in tx + `authz.Require` |
| `internal/modules/documents/repository/repository.go:73,216,428,1071,1082` | Document mutation sites |
| `internal/modules/registry/application/service.go:293,297,309` | Registry lifecycle — add tx + `authz.Require` in `changeStatus` |
| `internal/modules/registry/infrastructure/repository.go:133,137,184,208,239` | Registry mutations |
| `internal/modules/taxonomy/infrastructure/family_repository.go` | Taxonomy family mutations |
| `internal/modules/taxonomy/infrastructure/repository.go` | Taxonomy profile + area mutations |
| `apps/api/cmd/metaldocs-api/permissions.go:174` | Taxonomy PATCH families gap |
| `internal/modules/templates_v2/application/service.go` | Add `db *sql.DB` field |
| `internal/modules/templates_v2/application/lifecycle.go:265` | T-004 SoD + content_hash gate |
| `internal/modules/templates_v2/application/create.go:126` | T-002 cross-tenant tenant check |
| `apps/api/cmd/metaldocs-api/main.go:329` | Wire real `AuthzFunc` for templates_v2 |
| `migrations/0187_registry_lifecycle_caps_seed.sql` | New — rename `doc.supersede` + seed `registry.obsolete` + `registry.supersede` |
| `migrations/0188_tripwire_extend.sql` | New — extend trigger function + attach to all new tables + `document_families` code-immutability trigger |

---

## Task 1 — Typed capability constants for registry lifecycle

**Files:**
- Modify: `internal/modules/iam/domain/model.go:32-36`

- [ ] **Step 1.1: Add two new typed Capability constants**

In `internal/modules/iam/domain/model.go`, after the existing `CapRegistryCreate` line, add:

```go
// existing:
CapRegistryCreate   Capability = "registry.create"
// add:
CapRegistryObsolete  Capability = "registry.obsolete"
CapRegistrySupersede Capability = "registry.supersede"
```

Full updated const block (lines 15–36):

```go
const (
	CapDocumentView    Capability = "document.view"
	CapDocumentCreate  Capability = "document.create"
	CapDocumentEdit    Capability = "document.edit"
	CapDocumentSubmit  Capability = "document.submit"
	CapDocumentSignoff Capability = "document.signoff"
	CapWorkflowReview  Capability = "workflow.review"
	CapWorkflowApprove Capability = "workflow.approve"

	CapTemplateView    Capability = "template.view"
	CapTemplateCreate  Capability = "template.create"
	CapTemplateEdit    Capability = "template.edit"
	CapTemplateSubmit  Capability = "template.submit"
	CapTemplateApprove Capability = "template.approve"
	CapTemplatePublish Capability = "template.publish"

	CapRegistryCreate    Capability = "registry.create"
	CapRegistryObsolete  Capability = "registry.obsolete"
	CapRegistrySupersede Capability = "registry.supersede"
	CapTaxonomyManage    Capability = "taxonomy.manage"
	CapMembershipManage  Capability = "membership.manage"
	CapRouteManage       Capability = "route.manage"
	CapUserManage        Capability = "user.manage"
)
```

- [ ] **Step 1.2: Verify compile**

```powershell
go build ./internal/modules/iam/...
```

Expected: no errors.

- [ ] **Step 1.3: Commit**

```
git add internal/modules/iam/domain/model.go
git commit -m "feat(iam): add CapRegistryObsolete + CapRegistrySupersede typed capability constants"
```

---

## Task 2 — Critical: Fix taxonomy PATCH /families capability bypass (T-003)

**Files:**
- Modify: `apps/api/cmd/metaldocs-api/permissions.go:174-180`

This is a Critical security fix (unauthenticated user can mutate global catalog). Do this before any other change.

- [ ] **Step 2.1: Write a failing test for the PATCH gap**

In `apps/api/cmd/metaldocs-api/permissions_test.go` (create if absent), add:

```go
package main

import (
	"net/http"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

func TestPermissionResolver_TaxonomyFamiliesPATCH_RequiresTaxonomyManage(t *testing.T) {
	r := newPermissionResolver()
	cap, ok := r.Resolve(http.MethodPatch, "/api/v2/taxonomy/families/PROC")
	if !ok {
		t.Fatal("PATCH /taxonomy/families/{code}: resolver returned ok=false, want true")
	}
	if cap != iamdomain.CapTaxonomyManage {
		t.Fatalf("PATCH /taxonomy/families: cap = %v, want CapTaxonomyManage", cap)
	}
}
```

- [ ] **Step 2.2: Run test — expect FAIL**

```powershell
go test ./apps/api/cmd/metaldocs-api/... -run TestPermissionResolver_TaxonomyFamiliesPATCH -v
```

Expected: FAIL — resolver returns `("", false)` for PATCH today.

- [ ] **Step 2.3: Add `http.MethodPatch` to the families branch**

In `apps/api/cmd/metaldocs-api/permissions.go`, change lines 174–181 from:

```go
if strings.HasPrefix(path, "/api/v2/taxonomy/families") {
    switch method {
    case http.MethodGet:
        return iamdomain.CapDocumentView, true
    case http.MethodPost, http.MethodPut, http.MethodDelete:
        return iamdomain.CapTaxonomyManage, true
    }
}
```

to:

```go
if strings.HasPrefix(path, "/api/v2/taxonomy/families") {
    switch method {
    case http.MethodGet:
        return iamdomain.CapDocumentView, true
    case http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete:
        return iamdomain.CapTaxonomyManage, true
    }
}
```

Also apply the same fix to the areas branch (line ~166–173) to add `http.MethodPatch` for symmetry:

```go
if strings.HasPrefix(path, "/api/v2/taxonomy/areas") {
    switch method {
    case http.MethodGet:
        return iamdomain.CapDocumentView, true
    case http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete:
        return iamdomain.CapTaxonomyManage, true
    }
}
```

- [ ] **Step 2.4: Run test — expect PASS**

```powershell
go test ./apps/api/cmd/metaldocs-api/... -run TestPermissionResolver_TaxonomyFamiliesPATCH -v
```

Expected: PASS.

- [ ] **Step 2.5: Run full test suite**

```powershell
go test ./...
```

Expected: all pass.

- [ ] **Step 2.6: Commit**

```
git add apps/api/cmd/metaldocs-api/permissions.go apps/api/cmd/metaldocs-api/permissions_test.go
git commit -m "fix(taxonomy): add MethodPatch to capability dispatcher for families + areas — closes T-003/R-003"
```

---

## Task 3 — Registry: swap CapDocumentEdit → specific caps for obsolete/supersede routes

**Files:**
- Modify: `apps/api/cmd/metaldocs-api/permissions.go:190-193`

- [ ] **Step 3.1: Write failing test**

In `apps/api/cmd/metaldocs-api/permissions_test.go`, add:

```go
func TestPermissionResolver_RegistryObsolete_RequiresRegistryObsoleteCap(t *testing.T) {
	r := newPermissionResolver()
	cap, ok := r.Resolve(http.MethodPut, "/api/v2/controlled-documents/some-id/obsolete")
	if !ok {
		t.Fatal("PUT .../obsolete: resolver returned ok=false")
	}
	if cap != iamdomain.CapRegistryObsolete {
		t.Fatalf("PUT .../obsolete: cap = %v, want CapRegistryObsolete", cap)
	}
}

func TestPermissionResolver_RegistrySupersede_RequiresRegistrySupersedeCAP(t *testing.T) {
	r := newPermissionResolver()
	cap, ok := r.Resolve(http.MethodPut, "/api/v2/controlled-documents/some-id/supersede")
	if !ok {
		t.Fatal("PUT .../supersede: resolver returned ok=false")
	}
	if cap != iamdomain.CapRegistrySupersede {
		t.Fatalf("PUT .../supersede: cap = %v, want CapRegistrySupersede", cap)
	}
}
```

- [ ] **Step 3.2: Run tests — expect FAIL**

```powershell
go test ./apps/api/cmd/metaldocs-api/... -run TestPermissionResolver_Registry -v
```

Expected: FAIL — still returns `CapDocumentEdit`.

- [ ] **Step 3.3: Update permissions.go obsolete/supersede cases**

Change:

```go
case method == http.MethodPut && strings.HasSuffix(path, "/obsolete"):
    return iamdomain.CapDocumentEdit, true
case method == http.MethodPut && strings.HasSuffix(path, "/supersede"):
    return iamdomain.CapDocumentEdit, true
```

to:

```go
case method == http.MethodPut && strings.HasSuffix(path, "/obsolete"):
    return iamdomain.CapRegistryObsolete, true
case method == http.MethodPut && strings.HasSuffix(path, "/supersede"):
    return iamdomain.CapRegistrySupersede, true
```

- [ ] **Step 3.4: Run tests — expect PASS**

```powershell
go test ./apps/api/cmd/metaldocs-api/... -run TestPermissionResolver_Registry -v
```

- [ ] **Step 3.5: Full suite**

```powershell
go test ./...
```

- [ ] **Step 3.6: Commit**

```
git add apps/api/cmd/metaldocs-api/permissions.go apps/api/cmd/metaldocs-api/permissions_test.go
git commit -m "fix(registry): route obsolete/supersede to CapRegistryObsolete/CapRegistrySupersede — closes T-001/R-001"
```

---

## Task 4 — Seed migration: `registry.obsolete` + `registry.supersede` + rename `doc.supersede`

**Files:**
- Create: `migrations/0187_registry_lifecycle_caps_seed.sql`

- [ ] **Step 4.1: Create migration**

```sql
-- migrations/0187_registry_lifecycle_caps_seed.sql
-- Plan 5: rename doc.supersede → registry.supersede (aligns with typed Capability namespace from Plan 4)
-- and seed registry.obsolete for the same regulated roles.

BEGIN;

-- Rename doc.supersede rows seeded by 0177 to match typed namespace.
UPDATE metaldocs.role_capabilities
   SET capability = 'registry.supersede'
 WHERE capability = 'doc.supersede';

-- Seed registry.obsolete for the same roles that may supersede.
INSERT INTO metaldocs.role_capabilities (role, capability) VALUES
  ('area_admin',   'registry.obsolete'),
  ('qms_admin',    'registry.obsolete'),
  ('system_admin', 'registry.obsolete')
ON CONFLICT (role, capability) DO NOTHING;

INSERT INTO public.schema_migrations (version, description)
VALUES ('0187', 'Plan 5: rename doc.supersede → registry.supersede, seed registry.obsolete')
ON CONFLICT (version) DO NOTHING;

COMMIT;
```

- [ ] **Step 4.2: Note**

This migration is applied against the running DB in the integration-test phase (Task 14). No Go code to compile yet.

- [ ] **Step 4.3: Commit**

```
git add migrations/0187_registry_lifecycle_caps_seed.sql
git commit -m "feat(registry): seed registry.obsolete + rename doc.supersede → registry.supersede (migration 0187)"
```

---

## Task 5 — IAM tier-2: `authz.Require` in `RoleAdminRepository` mutations (T-004)

**Files:**
- Modify: `internal/modules/iam/infrastructure/postgres/role_admin_repository.go`

The two methods `UpsertUserAndAssignRole` (line 33) and `ReplaceUserRoles` (line 72) already start their own internal `*sql.Tx`. We add `authz.Require` on that tx before any DML executes.

`authz.Require` signature: `func Require(ctx context.Context, tx *sql.Tx, capability, areaCode string) error`

The `areaCode` here is `"tenant"` — the tier-1 middleware already scope-checks the area; tier-2 here confirms the cap is asserted inside the tx, which sets the GUC needed by the tripwire trigger.

- [ ] **Step 5.1: Add import for authz + iamdomain**

At top of `role_admin_repository.go`, add imports:

```go
import (
    // existing imports...
    "metaldocs/internal/modules/iam/authz"
    iamdomain "metaldocs/internal/modules/iam/domain"
)
```

- [ ] **Step 5.2: Add `authz.Require` in `UpsertUserAndAssignRole`**

After `tx, err := r.db.BeginTx(ctx, nil)` (around line 34), add:

```go
if err := authz.Require(ctx, tx, string(iamdomain.CapUserManage), "tenant"); err != nil {
    return fmt.Errorf("iam: authz check UpsertUserAndAssignRole: %w", err)
}
```

Full method start becomes:

```go
func (r *RoleAdminRepository) UpsertUserAndAssignRole(ctx context.Context, userID, displayName, tenantID string, role domain.Role, assignedBy string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin iam tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapUserManage), "tenant"); err != nil {
		return fmt.Errorf("iam: authz check UpsertUserAndAssignRole: %w", err)
	}

	const upsertUser = `...` // unchanged
```

- [ ] **Step 5.3: Add `authz.Require` in `ReplaceUserRoles`**

Same pattern — after `BeginTx`, before any ExecContext:

```go
func (r *RoleAdminRepository) ReplaceUserRoles(ctx context.Context, userID, displayName, tenantID string, roles []domain.Role, assignedBy string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin iam replace tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapUserManage), "tenant"); err != nil {
		return fmt.Errorf("iam: authz check ReplaceUserRoles: %w", err)
	}

	const upsertUser = `...` // unchanged
```

- [ ] **Step 5.4: Verify compile**

```powershell
go build ./internal/modules/iam/...
```

- [ ] **Step 5.5: Run IAM tests**

```powershell
go test ./internal/modules/iam/... -v
```

Expected: all pass (authz.Require will return `ErrActorContextMissing` for tests that don't set GUCs — verify existing tests already mock/bypass this or add `authz.WithCapCache` context).

> **Note:** If tests fail with `ErrActorContextMissing`, it means they are calling the repo directly without setting auth context. Add `authz.WithCapCache(ctx)` + a mock tx to the test setup, or use the `bypass_authz = 'scheduler'` GUC path via `authz.BypassSystem`.

- [ ] **Step 5.6: Commit**

```
git add internal/modules/iam/infrastructure/postgres/role_admin_repository.go
git commit -m "feat(iam): add authz.Require tier-2 to RoleAdminRepository mutations — T-004"
```

---

## Task 6 — IAM tier-2: `authz.Require` in `UserAreaRepository` mutations (T-004)

**Files:**
- Modify: `internal/modules/iam/infrastructure/postgres/user_area_repository.go`

Three methods: `Insert` (line 51), `CloseActive` (line 75), `GrantAtomic` (line 90).

`GrantAtomic` already opens a tx. `Insert` and `CloseActive` use `r.db.ExecContext` directly — wrap them in an internal tx.

- [ ] **Step 6.1: Add import for authz + iamdomain**

```go
import (
    // existing...
    "metaldocs/internal/modules/iam/authz"
    iamdomain "metaldocs/internal/modules/iam/domain"
)
```

- [ ] **Step 6.2: Wrap `Insert` in tx + add `authz.Require`**

Replace:

```go
func (r *UserAreaRepository) Insert(ctx context.Context, membership domain.UserProcessArea) error {
	const q = `INSERT INTO user_process_areas ...`
	_, err := r.db.ExecContext(ctx, q, membership.UserID, ...)
	if err != nil {
		return fmt.Errorf("insert user process area: %w", err)
	}
	return nil
}
```

With:

```go
func (r *UserAreaRepository) Insert(ctx context.Context, membership domain.UserProcessArea) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin insert area tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapMembershipManage), "tenant"); err != nil {
		return fmt.Errorf("iam: authz check Insert area: %w", err)
	}

	const q = `
INSERT INTO user_process_areas
  (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by)
VALUES
  ($1, $2::uuid, $3, $4, $5, $6, $7)
`
	if _, err := tx.ExecContext(ctx, q,
		membership.UserID,
		membership.TenantID,
		membership.AreaCode,
		string(membership.Role),
		membership.EffectiveFrom,
		membership.EffectiveTo,
		membership.GrantedBy,
	); err != nil {
		return fmt.Errorf("insert user process area: %w", err)
	}
	return tx.Commit()
}
```

- [ ] **Step 6.3: Wrap `CloseActive` in tx + add `authz.Require`**

Replace:

```go
func (r *UserAreaRepository) CloseActive(ctx context.Context, userID, tenantID, areaCode string, effectiveTo time.Time) error {
	const q = `UPDATE user_process_areas SET effective_to = $4 ...`
	if _, err := r.db.ExecContext(ctx, q, userID, tenantID, areaCode, effectiveTo); err != nil {
		return fmt.Errorf("close active user process area: %w", err)
	}
	return nil
}
```

With:

```go
func (r *UserAreaRepository) CloseActive(ctx context.Context, userID, tenantID, areaCode string, effectiveTo time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin close area tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapMembershipManage), "tenant"); err != nil {
		return fmt.Errorf("iam: authz check CloseActive area: %w", err)
	}

	const q = `
UPDATE user_process_areas
SET effective_to = $4
WHERE user_id = $1
  AND tenant_id::text = $2
  AND area_code = $3
  AND effective_to IS NULL
`
	if _, err := tx.ExecContext(ctx, q, userID, tenantID, areaCode, effectiveTo); err != nil {
		return fmt.Errorf("close active user process area: %w", err)
	}
	return tx.Commit()
}
```

- [ ] **Step 6.4: Add `authz.Require` in `GrantAtomic`**

`GrantAtomic` already opens a tx. Add authz.Require before any DML:

```go
func (r *UserAreaRepository) GrantAtomic(ctx context.Context, oldMembership, newMembership domain.UserProcessArea) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin grant transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := authz.Require(ctx, tx, string(iamdomain.CapMembershipManage), "tenant"); err != nil {
		return fmt.Errorf("iam: authz check GrantAtomic: %w", err)
	}

	// rest of method unchanged (CloseActive query + Insert query using tx)
```

- [ ] **Step 6.5: Compile + test**

```powershell
go build ./internal/modules/iam/...
go test ./internal/modules/iam/... -v
```

- [ ] **Step 6.6: Commit**

```
git add internal/modules/iam/infrastructure/postgres/user_area_repository.go
git commit -m "feat(iam): add authz.Require tier-2 to UserAreaRepository mutations — T-004"
```

---

## Task 7 — Documents tier-2: `authz.Require` in repository mutations (T-003)

**Files:**
- Modify: `internal/modules/documents/repository/repository.go`

Five mutation methods: `CreateDocumentTx` (already has tx, line 73), `UpdateDocumentName` (line 216), `UpdateDocumentStatus` (line 428), `MarkArchived` (line 1071), `Unarchive` (line 1082).

**Strategy:**
- `CreateDocumentTx`: inject `authz.Require(ctx, tx, "document.create", "tenant")` early in the existing tx, before DML.
- Other four: wrap in an internal tx, call `authz.Require`, execute DML, commit.

- [ ] **Step 7.1: Add imports**

```go
import (
    // existing...
    "metaldocs/internal/modules/iam/authz"
    iamdomain "metaldocs/internal/modules/iam/domain"
)
```

- [ ] **Step 7.2: Add `authz.Require` in `CreateDocumentTx`**

Inside `CreateDocumentTx`, after the advisory lock (`pg_advisory_xact_lock`) and before the first INSERT, add:

```go
if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentCreate), "tenant"); err != nil {
    return "", "", "", fmt.Errorf("create document: authz check: %w", err)
}
```

- [ ] **Step 7.3: Wrap `UpdateDocumentName` in tx + add `authz.Require`**

Replace:

```go
func (r *Repository) UpdateDocumentName(ctx context.Context, tenantID, docID, name string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE documents SET name=$2, updated_at=now() WHERE id=$1 AND tenant_id=$3`,
		docID, name, tenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
```

With:

```go
func (r *Repository) UpdateDocumentName(ctx context.Context, tenantID, docID, name string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), "tenant"); err != nil {
		return fmt.Errorf("update document name: authz check: %w", err)
	}

	res, err := tx.ExecContext(ctx,
		`UPDATE documents SET name=$2, updated_at=now() WHERE id=$1 AND tenant_id=$3`,
		docID, name, tenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return tx.Commit()
}
```

- [ ] **Step 7.4: Wrap `UpdateDocumentStatus` in tx + add `authz.Require`**

Replace the `r.db.ExecContext` call with an internal tx. The cap is `document.edit` (status mutations are edit-level ops). The `col` local variable and `fmt.Sprintf` call remain.

```go
func (r *Repository) UpdateDocumentStatus(ctx context.Context, tenantID, id string, cur, next domain.DocumentStatus, stampTime bool) error {
	col := ""
	if stampTime {
		if next == domain.DocStatusArchived {
			col = "archived_at  = now(),"
		}
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), "tenant"); err != nil {
		return fmt.Errorf("update document status: authz check: %w", err)
	}

	res, err := tx.ExecContext(ctx,
		fmt.Sprintf(`UPDATE documents SET status=$1, %s updated_at=now() WHERE id=$2 AND tenant_id=$3 AND status=$4`, col),
		next, id, tenantID, cur)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return tx.Commit()
}
```

> **Note:** Check the actual lines around 428–445 for the full `RowsAffected` / `ErrNotFound` path and replicate it exactly inside the tx block.

- [ ] **Step 7.5: Wrap `MarkArchived` in tx + add `authz.Require`**

```go
func (r *Repository) MarkArchived(ctx context.Context, tenantID, docID, actorID string) error {
	_ = actorID
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), "tenant"); err != nil {
		return fmt.Errorf("mark archived: authz check: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE public.documents
		   SET archived_at = now(), updated_at = now()
		 WHERE tenant_id = $1 AND id = $2 AND archived_at IS NULL`,
		tenantID, docID)
	if err != nil {
		return err
	}
	return tx.Commit()
}
```

- [ ] **Step 7.6: Wrap `Unarchive` in tx + add `authz.Require`**

```go
func (r *Repository) Unarchive(ctx context.Context, tenantID, docID, actorID string) error {
	_ = actorID
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentEdit), "tenant"); err != nil {
		return fmt.Errorf("unarchive: authz check: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE public.documents
		   SET archived_at = NULL, updated_at = now()
		 WHERE tenant_id = $1 AND id = $2 AND archived_at IS NOT NULL`,
		tenantID, docID)
	if err != nil {
		return err
	}
	return tx.Commit()
}
```

- [ ] **Step 7.7: Compile + test**

```powershell
go build ./internal/modules/documents/...
go test ./internal/modules/documents/... -v
```

- [ ] **Step 7.8: Full suite**

```powershell
go test ./...
```

- [ ] **Step 7.9: Commit**

```
git add internal/modules/documents/repository/repository.go
git commit -m "feat(documents): add authz.Require tier-2 to all document table mutations — T-003/R-003"
```

---

## Task 8 — Registry tier-2: `authz.Require` in service + repository (T-001, T-004)

**Files:**
- Modify: `internal/modules/registry/application/service.go`
- Modify: `internal/modules/registry/infrastructure/repository.go`
- Modify: `internal/modules/registry/domain/repository.go` (if `ControlledDocumentRepository` interface needs tx variant)

**Strategy:**
- `Create` (non-tx): repo already has internal path via `createWithQueryer`. Wrap in tx + authz.Require.
- `CreateTx`: already has tx — add authz.Require.
- `UpdateStatus`: called by `changeStatus` which has no tx. Modify `changeStatus` to start a tx, call authz.Require with the appropriate lifecycle cap, call a new `updateStatusTx` private method.
- `EnsureCounter` / `NextAndIncrement` (cd_sequence_counters): wrap EnsureCounter + the tx path of NextAndIncrement.

### 8a — Registry infrastructure repository

- [ ] **Step 8.1: Add imports to `registry/infrastructure/repository.go`**

```go
import (
    // existing...
    "metaldocs/internal/modules/iam/authz"
    iamdomain "metaldocs/internal/modules/iam/domain"
)
```

- [ ] **Step 8.2: Add `authz.Require` in `Create` (non-tx path)**

In `PostgresControlledDocumentRepository.Create`:

```go
func (r *PostgresControlledDocumentRepository) Create(ctx context.Context, doc *registrydomain.ControlledDocument) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapRegistryCreate), "tenant"); err != nil {
		return fmt.Errorf("registry create: authz check: %w", err)
	}

	if err := r.createWithQueryer(ctx, tx, doc); err != nil {
		return err
	}
	return tx.Commit()
}
```

- [ ] **Step 8.3: Add `authz.Require` in `CreateTx`**

`CreateTx` already receives a `*sql.Tx`. Add authz.Require before calling `createWithQueryer`:

```go
func (r *PostgresControlledDocumentRepository) CreateTx(ctx context.Context, tx *sql.Tx, doc *registrydomain.ControlledDocument) error {
	if tx == nil {
		return errors.New("nil transaction")
	}

	if err := authz.Require(ctx, tx, string(iamdomain.CapRegistryCreate), "tenant"); err != nil {
		return fmt.Errorf("registry createTx: authz check: %w", err)
	}

	return r.createWithQueryer(ctx, tx, doc)
}
```

- [ ] **Step 8.4: Add `authz.Require` in `EnsureCounter`**

`EnsureCounter` uses `a.db` directly. Wrap in tx:

```go
func (a *PostgresSequenceAllocator) EnsureCounter(ctx context.Context, tenantID, profileCode, areaCode string) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapRegistryCreate), "tenant"); err != nil {
		return fmt.Errorf("ensure counter: authz check: %w", err)
	}

	if err := a.ensureCounterViaExec(ctx, tx, tenantID, profileCode, areaCode); err != nil {
		return err
	}
	return tx.Commit()
}
```

### 8b — Registry service changeStatus

- [ ] **Step 8.5: Add imports to `registry/application/service.go`**

```go
import (
    // existing...
    "metaldocs/internal/modules/iam/authz"
    iamdomain "metaldocs/internal/modules/iam/domain"
)
```

- [ ] **Step 8.6: Add `UpdateStatusTx` to `ControlledDocumentRepository` interface**

In `internal/modules/registry/domain/repository.go`, add:

```go
type ControlledDocumentRepository interface {
    // existing methods...
    UpdateStatusTx(ctx context.Context, tx *sql.Tx, tenantID, id string, status CDStatus, updatedAt time.Time) error
}
```

Also check what imports this file needs (`database/sql`).

- [ ] **Step 8.7: Implement `UpdateStatusTx` in `infrastructure/repository.go`**

```go
func (r *PostgresControlledDocumentRepository) UpdateStatusTx(ctx context.Context, tx *sql.Tx, tenantID, id string, status registrydomain.CDStatus, updatedAt time.Time) error {
	res, err := tx.ExecContext(ctx,
		`UPDATE controlled_documents SET status = $1, updated_at = $2 WHERE tenant_id = $3 AND id = $4`,
		status, updatedAt, tenantID, id,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return registrydomain.ErrCDNotFound
	}
	return nil
}
```

- [ ] **Step 8.8: Modify `changeStatus` in service to use tx + authz.Require**

Replace:

```go
func (s *RegistryService) changeStatus(ctx context.Context, tenantID, controlledDocumentID string, next registrydomain.CDStatus) error {
	doc, err := s.docs.GetByID(ctx, tenantID, controlledDocumentID)
	if err != nil {
		return err
	}
	if !doc.IsActive() {
		return registrydomain.ErrCDNotActive
	}
	return s.docs.UpdateStatus(ctx, tenantID, controlledDocumentID, next, s.now().UTC())
}
```

With (note `changeStatus` gets a new `cap string` parameter, called only from `Obsolete` and `Supersede`):

```go
func (s *RegistryService) Obsolete(ctx context.Context, tenantID, controlledDocumentID string) error {
	return s.changeStatus(ctx, tenantID, controlledDocumentID, registrydomain.CDStatusObsolete, string(iamdomain.CapRegistryObsolete))
}

func (s *RegistryService) Supersede(ctx context.Context, tenantID, controlledDocumentID string) error {
	return s.changeStatus(ctx, tenantID, controlledDocumentID, registrydomain.CDStatusSuperseded, string(iamdomain.CapRegistrySupersede))
}

func (s *RegistryService) changeStatus(ctx context.Context, tenantID, controlledDocumentID string, next registrydomain.CDStatus, cap string) error {
	doc, err := s.docs.GetByID(ctx, tenantID, controlledDocumentID)
	if err != nil {
		return err
	}
	if !doc.IsActive() {
		return registrydomain.ErrCDNotActive
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("registry changeStatus: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, cap, "tenant"); err != nil {
		return fmt.Errorf("registry changeStatus: authz check: %w", err)
	}

	if err := s.docs.UpdateStatusTx(ctx, tx, tenantID, controlledDocumentID, next, s.now().UTC()); err != nil {
		return err
	}
	return tx.Commit()
}
```

- [ ] **Step 8.9: Update mock/fake for `ControlledDocumentRepository` in tests**

If any test file has a fake/spy implementing `ControlledDocumentRepository`, add a stub `UpdateStatusTx` method:

```go
func (f *fakeRepo) UpdateStatusTx(ctx context.Context, tx *sql.Tx, tenantID, id string, status registrydomain.CDStatus, updatedAt time.Time) error {
	return f.UpdateStatus(ctx, tenantID, id, status, updatedAt) // delegate to existing fake
}
```

Find these fakes by running:

```powershell
grep -rn "ControlledDocumentRepository" internal/modules/registry/ --include="*_test.go"
```

- [ ] **Step 8.10: Compile + test**

```powershell
go build ./internal/modules/registry/...
go test ./internal/modules/registry/... -v
```

- [ ] **Step 8.11: Full suite**

```powershell
go test ./...
```

- [ ] **Step 8.12: Commit**

```
git add internal/modules/registry/
git commit -m "feat(registry): add authz.Require tier-2 to all registry mutations + UpdateStatusTx — T-001/T-004"
```

---

## Task 9 — Taxonomy tier-2: `authz.Require` in infra repo mutations (T-006)

**Files:**
- Modify: `internal/modules/taxonomy/infrastructure/family_repository.go`
- Modify: `internal/modules/taxonomy/infrastructure/repository.go` (ProfileRepository + AreaRepository)

**Strategy:** Wrap each mutating method (`Create`, `Update`, and `HasActiveProfiles`-dependent `Deactivate` flow) in an internal tx, call `authz.Require` before DML.

- [ ] **Step 9.1: Add imports to `family_repository.go`**

```go
import (
    // existing...
    "metaldocs/internal/modules/iam/authz"
    iamdomain "metaldocs/internal/modules/iam/domain"
)
```

- [ ] **Step 9.2: Wrap `FamilyRepository.Create` in tx + authz.Require**

Check existing implementation at line 67; it calls `r.db.ExecContext`. Replace with:

```go
func (r *FamilyRepository) Create(ctx context.Context, f *domain.DocumentFamily) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy family create: authz: %w", err)
	}

	// copy existing INSERT query here, using tx.ExecContext instead of r.db.ExecContext
	_, err = tx.ExecContext(ctx, `INSERT INTO document_families (code, name, description, is_active) VALUES ($1, $2, $3, $4)`,
		f.Code, f.Name, f.Description, f.IsActive)
	if err != nil {
		return err
	}
	return tx.Commit()
}
```

> Read the actual INSERT query from the existing `Create` method before writing this step — copy it exactly.

- [ ] **Step 9.3: Wrap `FamilyRepository.Update` in tx + authz.Require**

Same pattern. Read existing UPDATE query from line 75 and replicate inside a tx with authz.Require.

```go
func (r *FamilyRepository) Update(ctx context.Context, f *domain.DocumentFamily) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy family update: authz: %w", err)
	}

	// copy existing UPDATE query here, using tx.ExecContext
	_, err = tx.ExecContext(ctx, `UPDATE document_families SET name=$2, description=$3, is_active=$4 WHERE code=$1`,
		f.Code, f.Name, f.Description, f.IsActive)
	if err != nil {
		return err
	}
	return tx.Commit()
}
```

- [ ] **Step 9.4: Add imports + wrap `ProfileRepository.Create` and `Update` in `repository.go`**

```go
import (
    // existing...
    "metaldocs/internal/modules/iam/authz"
    iamdomain "metaldocs/internal/modules/iam/domain"
)
```

Wrap `ProfileRepository.Create` (line ~102):

```go
func (r *ProfileRepository) Create(ctx context.Context, p *domain.DocumentProfile) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy profile create: authz: %w", err)
	}

	// copy existing INSERT body here, using tx.ExecContext
	// read from lines 102..126 in infrastructure/repository.go
	// ... (keep all existing columns and params)
	return tx.Commit()
}
```

Wrap `ProfileRepository.Update` (line ~127):

```go
func (r *ProfileRepository) Update(ctx context.Context, p *domain.DocumentProfile) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy profile update: authz: %w", err)
	}

	// copy existing UPDATE body here, using tx.ExecContext
	// read from lines 127..165 in infrastructure/repository.go
	return tx.Commit()
}
```

- [ ] **Step 9.5: Wrap `AreaRepository.Create` and `Update`**

Same pattern for `AreaRepository` (lines ~253, ~275 in `repository.go`).

```go
func (r *AreaRepository) Create(ctx context.Context, a *domain.ProcessArea) error {
	tx, err := r.db.BeginTx(ctx, nil)
	// ... same pattern
	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy area create: authz: %w", err)
	}
	// copy existing INSERT using tx.ExecContext
	return tx.Commit()
}

func (r *AreaRepository) Update(ctx context.Context, a *domain.ProcessArea) error {
	tx, err := r.db.BeginTx(ctx, nil)
	// ...
	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy area update: authz: %w", err)
	}
	// copy existing UPDATE using tx.ExecContext
	return tx.Commit()
}
```

- [ ] **Step 9.6: Wrap `ProfileRepository.Archive` + `AreaRepository.Archive` (if they exist as Update calls)**

> **Why required:** Migration 0188 attaches `enforce_capability_asserted` to `BEFORE UPDATE` on both tables. Any UPDATE without `metaldocs.asserted_caps` set will be rejected at the DB layer with P0001 — including archive operations. Not wrapping these breaks archive in production the moment 0188 is applied.

Check whether archive is a separate method or reuses `Update` by running:

```powershell
grep -n "Archive\|archive\|is_active" internal/modules/taxonomy/infrastructure/repository.go
```

If `Archive` calls `r.db.ExecContext` directly (separate UPDATE), wrap it the same way as `Create`/`Update` above:

```go
func (r *ProfileRepository) Archive(ctx context.Context, tenantID, code string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, string(iamdomain.CapTaxonomyManage), "tenant"); err != nil {
		return fmt.Errorf("taxonomy profile archive: authz: %w", err)
	}

	// copy existing UPDATE body from ProfileRepository.Archive, using tx.ExecContext
	return tx.Commit()
}
```

Apply same pattern to `AreaRepository.Archive`.

If `Archive` delegates to `Update` (reuses the same method), no change needed — `Update` is already wrapped above and will carry the authz check automatically.

- [ ] **Step 9.7: Compile + test**

```powershell
go build ./internal/modules/taxonomy/...
go test ./internal/modules/taxonomy/... -v
```

- [ ] **Step 9.8: Full suite**

```powershell
go test ./...
```

- [ ] **Step 9.9: Commit**

```
git add internal/modules/taxonomy/infrastructure/
git commit -m "feat(taxonomy): add authz.Require tier-2 to family/profile/area infra mutations incl. archive — T-006/R-006"
```

---

## Task 10 — Templates_v2: wire real AuthzFunc + add `db` field + service-level authz (T-001, T-002, T-004)

**Files:**
- Modify: `internal/modules/templates_v2/application/service.go`
- Modify: `internal/modules/templates_v2/application/lifecycle.go`
- Modify: `internal/modules/templates_v2/application/create.go`
- Modify: `apps/api/cmd/metaldocs-api/main.go:329`

### 10a — Add `db *sql.DB` to `Service`

- [ ] **Step 10.1: Extend `Service` struct and constructor**

In `internal/modules/templates_v2/application/service.go`:

```go
import (
    "context"
    "database/sql"

    "metaldocs/internal/modules/iam/authz"
    iamdomain "metaldocs/internal/modules/iam/domain"
)

type Service struct {
	repo      Repository
	presign   Presigner
	clock     Clock
	uuid      UUIDGen
	resolvers ResolverRegistryReader
	db        *sql.DB
}

func New(repo Repository, presign Presigner, clock Clock, uuid UUIDGen, resolvers ...ResolverRegistryReader) *Service {
	var registry ResolverRegistryReader
	if len(resolvers) > 0 {
		registry = resolvers[0]
	}
	return &Service{repo: repo, presign: presign, clock: clock, uuid: uuid, resolvers: registry}
}

// WithDB injects the *sql.DB required for tier-2 authz.Require calls inside
// service mutations. Must be called before any mutation method is invoked.
func (s *Service) WithDB(db *sql.DB) *Service {
	s.db = db
	return s
}
```

> `WithDB` uses a builder pattern consistent with `RegistryService.WithDocumentInitializer`.

- [ ] **Step 10.2: Wire `WithDB` in `main.go`**

Change line 329 from:

```go
tv2http.New(tv2Svc, nil).Register(mux)
```

to:

```go
tv2Svc.WithDB(deps.SQLDB)
authzFn := func(r *http.Request, tenantID, _ string, action string) error {
    userID := iamdomain.UserIDFromContext(r.Context())
    return capabilityService.CanDo(r.Context(), userID, tenantID, action)
}
tv2http.New(tv2Svc, authzFn).Register(mux)
```

Add imports to `main.go` if needed:
- `iamdomain "metaldocs/internal/modules/iam/domain"`

### 10b — Add `authz.Require` in lifecycle mutations

- [ ] **Step 10.3: Add `authz.Require` in `Service.CreateTemplate`**

In `internal/modules/templates_v2/application/create.go`, inside `CreateTemplate`:

```go
func (s *Service) CreateTemplate(ctx context.Context, cmd CreateTemplateCmd) (*domain.Template, error) {
	if s.db != nil {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("templates create: begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback() }()
		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateCreate), "tenant"); err != nil {
			return nil, fmt.Errorf("templates create: authz: %w", err)
		}
		_ = tx.Commit() // commit sets the GUC before the repo insert
	}
	// existing logic continues unchanged...
```

> The pattern here opens a tx, calls authz.Require (which sets `metaldocs.asserted_caps` GUC), commits, and THEN calls the repo method. The GUC is tx-local; this works because the repo method immediately opens its own tx that the tripwire trigger fires against. Wait — this does NOT work because the GUC resets on `tx.Commit()`.
>
> **Correct approach:** the authz.Require tx and the repo DML must be in the **same** transaction. For `CreateTemplate`, the repo `CreateTemplate` method uses `r.db.ExecContext` (not a tx). We need to either:
>   a) Provide a tx-accepting variant in the Repository interface, or
>   b) Wrap the entire `CreateTemplate` call inside a tx started here.
>
> Given the constraint "don't refactor repository signatures beyond adding tx param", we add `CreateTemplateTx(ctx, tx, t)` to the `Repository` interface.

### 10b (revised) — Repository interface Tx variants for templates_v2

- [ ] **Step 10.3 (revised): Add Tx variants to `application.Repository` interface**

In `internal/modules/templates_v2/application/repository.go` (or wherever the `Repository` interface is defined — check with `grep -rn "type Repository interface" internal/modules/templates_v2/`), add:

```go
CreateTemplateTx(ctx context.Context, tx *sql.Tx, t *domain.Template) error
UpdateTemplateTx(ctx context.Context, tx *sql.Tx, t *domain.Template) error
UpdateVersionTx(ctx context.Context, tx *sql.Tx, v *domain.TemplateVersion) error
```

- [ ] **Step 10.4: Implement Tx variants in `repository/postgres.go`**

```go
func (r *Repository) CreateTemplateTx(ctx context.Context, tx *sql.Tx, t *domain.Template) error {
	const q = `INSERT INTO templates_v2_template (...) VALUES (...)`
	// same as CreateTemplate but using tx.ExecContext
}

func (r *Repository) UpdateTemplateTx(ctx context.Context, tx *sql.Tx, t *domain.Template) error {
	// same as UpdateTemplate but using tx.ExecContext
}

func (r *Repository) UpdateVersionTx(ctx context.Context, tx *sql.Tx, v *domain.TemplateVersion) error {
	// same as UpdateVersion but using tx.ExecContext
}
```

Read existing `CreateTemplate`, `UpdateTemplate`, `UpdateVersion` bodies from `repository/postgres.go` and replicate with `tx.ExecContext`.

- [ ] **Step 10.5: Add `authz.Require` in `Service.CreateTemplate` using tx**

```go
func (s *Service) CreateTemplate(ctx context.Context, cmd CreateTemplateCmd) (*domain.Template, error) {
	// ... validation logic unchanged ...

	t := &domain.Template{ /* ... */ }

	if s.db != nil {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("templates create: begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback() }()

		if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateCreate), "tenant"); err != nil {
			return nil, err
		}
		if err := s.repo.CreateTemplateTx(ctx, tx, t); err != nil {
			return nil, err
		}
		return t, tx.Commit()
	}

	return t, s.repo.CreateTemplate(ctx, t)
}
```

- [ ] **Step 10.6: Add `authz.Require` in `Service.Submit`, `Service.Review`, `Service.Approve`**

For each lifecycle state-transition method that mutates `templates_v2_template_version`:

```go
// Inside Submit (find the method — likely in lifecycle.go):
if s.db != nil {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil { return nil, err }
    defer func() { _ = tx.Rollback() }()
    if err := authz.Require(ctx, tx, string(iamdomain.CapTemplateSubmit), "tenant"); err != nil {
        return nil, err
    }
    if err := s.repo.UpdateVersionTx(ctx, tx, version); err != nil { return nil, err }
    if err := s.repo.AppendAudit(ctx, auditEvent); err != nil { return nil, err }
    return version, tx.Commit()
}
// fallback (s.db == nil, e.g. in tests)
```

Apply similar pattern for `Review` (`CapTemplateEdit` or the reviewer-role cap), `Approve` (`CapTemplateApprove`), and `PublishTemplateVersion` (`CapTemplatePublish`).

### 10c — T-002 cross-tenant fix in `CreateNextVersion`

- [ ] **Step 10.7: Add tenant guard in `CreateNextVersion`**

In `internal/modules/templates_v2/application/create.go:126`, after calling `s.repo.GetVersionByID(ctx, *template.PublishedVersionID)`:

```go
source, err = s.repo.GetVersionByID(ctx, *template.PublishedVersionID)
if err != nil {
    return nil, err
}
// T-002 cross-tenant guard: GetVersionByID has no tenant predicate.
// The version must belong to the same template we already gated via GetTemplate(tenantID, ...).
if source.TemplateID != template.ID {
    return nil, domain.ErrNotFound // version belongs to another template — reject
}
```

- [ ] **Step 10.8: Run templates_v2 tests**

```powershell
go test ./internal/modules/templates_v2/... -v
```

### 10d — T-004: `PublishTemplateVersion` SoD + content_hash gate

- [ ] **Step 10.9: Add SoD and content_hash guard in `PublishTemplateVersion`**

In `lifecycle.go:265`, after `GetVersion` returns, before any state mutations:

```go
// T-004: content_hash gate — presigned docx must have been committed.
if version.ContentHash == "" {
    return nil, domain.ErrContentHashMismatch
}

// T-004: SoD — publisher must not be the author or the reviewer.
if err := domain.CheckSegregation("approver", cmd.ActorUserID, version.AuthorID, version.ReviewerID); err != nil {
    return nil, err
}
```

Add these two guards after the `version.Status != domain.VersionStatusDraft` check.

- [ ] **Step 10.10: Compile + test**

```powershell
go build ./internal/modules/templates_v2/...
go test ./internal/modules/templates_v2/... -v
```

- [ ] **Step 10.11: Full suite**

```powershell
go test ./...
```

- [ ] **Step 10.12: Commit**

```
git add internal/modules/templates_v2/ apps/api/cmd/metaldocs-api/main.go
git commit -m "feat(templates_v2): wire real AuthzFunc, add db+authz.Require, fix T-002 cross-tenant, fix T-004 SoD+content_hash — T-001/T-002/T-004"
```

---

## Task 11 — Migration: extend tripwire trigger + attach to all new tables + `document_families` code-immutability

**Files:**
- Create: `migrations/0188_tripwire_extend.sql`

This migration:
1. Replaces `enforce_capability_asserted()` with an extended version handling all new table names.
2. Attaches the trigger to every new regulated table.
3. Adds a `reject_code_update` BEFORE UPDATE trigger on `document_families.code` (mirrors migrations 0122/0123 for profiles/areas).

The new trigger function extends the `ELSIF` chain. For tables where multiple caps can satisfy the check (e.g. `controlled_documents` UPDATE can assert either `registry.obsolete` or `registry.supersede`), the function uses a helper that scans for any cap in a given prefix set.

- [ ] **Step 11.1: Create migration file**

```sql
-- migrations/0188_tripwire_extend.sql
-- Plan 5: extend enforce_capability_asserted trigger to all regulated tables.
-- Also adds document_families.code immutability trigger (mirrors migrations/0122, 0123).
--
-- Idempotent: DROP TRIGGER IF EXISTS before CREATE; CREATE OR REPLACE FUNCTION.

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Extend the tripwire function with new table/cap entries.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION public.enforce_capability_asserted()
  RETURNS trigger
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, pg_temp
AS $$
DECLARE
  v_bypass        TEXT;
  v_asserted_raw  TEXT;
  v_asserted      JSONB;
  v_required_caps TEXT[];   -- one or more acceptable caps for this table/op
  v_tenant_id     UUID;
  v_cap_found     BOOLEAN := FALSE;
  v_element       JSONB;
BEGIN
  -- ---- Determine required capability set for this table/operation. --------
  CASE
    WHEN TG_TABLE_NAME = 'approval_instances' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['document.submit'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'approval_signoffs' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['document.signoff'];
      v_tenant_id     := NEW.actor_tenant_id;

    WHEN TG_TABLE_NAME = 'iam_user_roles' THEN
      v_required_caps := ARRAY['user.manage'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'user_process_areas' THEN
      v_required_caps := ARRAY['membership.manage'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'documents' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['document.create'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'documents' AND TG_OP = 'UPDATE' THEN
      v_required_caps := ARRAY['document.edit'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'controlled_documents' AND TG_OP = 'INSERT' THEN
      v_required_caps := ARRAY['registry.create'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'controlled_documents' AND TG_OP = 'UPDATE' THEN
      -- Either lifecycle cap is acceptable (obsolete or supersede).
      v_required_caps := ARRAY['registry.obsolete', 'registry.supersede'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'cd_sequence_counters' THEN
      v_required_caps := ARRAY['registry.create'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'document_profiles' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'document_process_areas' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'document_families' THEN
      v_required_caps := ARRAY['taxonomy.manage'];
      -- document_families has no tenant_id; use NULL sentinel (no bypass logging context).
      v_tenant_id     := NULL;

    WHEN TG_TABLE_NAME = 'templates_v2_template' THEN
      v_required_caps := ARRAY['template.create', 'template.edit', 'template.submit',
                                'template.approve', 'template.publish'];
      v_tenant_id     := NEW.tenant_id;

    WHEN TG_TABLE_NAME = 'templates_v2_template_version' THEN
      v_required_caps := ARRAY['template.create', 'template.edit', 'template.submit',
                                'template.approve', 'template.publish'];
      -- template_version has no tenant_id column; derive from asserted_caps context.
      v_tenant_id     := NULL;

    ELSE
      -- Unknown table — conservative pass-through. Do not block unknown tables.
      RETURN NEW;
  END CASE;

  -- ---- Bypass path. -------------------------------------------------------
  v_bypass := pg_catalog.current_setting('metaldocs.bypass_authz', true);
  IF v_bypass IS NOT NULL AND v_bypass <> '' THEN
    IF v_bypass = 'scheduler' THEN
      BEGIN
        INSERT INTO public.governance_events
          (tenant_id, event_type, actor_user_id, resource_type, resource_id, reason, payload_json)
        VALUES (
          v_tenant_id,
          'authz.bypass_used',
          'system:scheduler',
          TG_TABLE_NAME,
          COALESCE(NEW.id::TEXT, 'unknown'),
          'scheduler bypass for ' || v_required_caps[1],
          pg_catalog.to_jsonb(jsonb_build_object(
            'required_caps', to_jsonb(v_required_caps),
            'bypass_token',  v_bypass,
            'table',         TG_TABLE_NAME,
            'op',            TG_OP
          ))
        );
      EXCEPTION WHEN others THEN
        RAISE NOTICE 'enforce_capability_asserted: governance_events insert failed: %', SQLERRM;
      END;
      RETURN NEW;
    ELSE
      RAISE EXCEPTION 'ErrCapabilityNotAsserted: unrecognised bypass token; caps % required on %',
                      v_required_caps, TG_TABLE_NAME
        USING ERRCODE = 'P0001';
    END IF;
  END IF;

  -- ---- Read asserted_caps GUC. --------------------------------------------
  v_asserted_raw := pg_catalog.current_setting('metaldocs.asserted_caps', true);
  IF v_asserted_raw IS NULL OR v_asserted_raw = '' THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: one of % required but metaldocs.asserted_caps is not set on %',
                    v_required_caps, TG_TABLE_NAME
      USING ERRCODE = 'P0001';
  END IF;

  BEGIN
    v_asserted := v_asserted_raw::JSONB;
  EXCEPTION WHEN invalid_text_representation OR others THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: metaldocs.asserted_caps is not valid JSONB (caps % required)',
                    v_required_caps
      USING ERRCODE = 'P0001';
  END;

  IF jsonb_typeof(v_asserted) <> 'array' THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: metaldocs.asserted_caps must be a JSONB array (caps % required)',
                    v_required_caps
      USING ERRCODE = 'P0001';
  END IF;

  -- ---- Scan for any required cap. ----------------------------------------
  FOR v_element IN SELECT * FROM jsonb_array_elements(v_asserted) LOOP
    IF (v_element->>'cap') = ANY(v_required_caps) THEN
      v_cap_found := TRUE;
      EXIT;
    END IF;
  END LOOP;

  IF NOT v_cap_found THEN
    RAISE EXCEPTION 'ErrCapabilityNotAsserted: none of % present in asserted_caps on %',
                    v_required_caps, TG_TABLE_NAME
      USING ERRCODE = 'P0001';
  END IF;

  RETURN NEW;
END;
$$;

-- Ownership / revoke (same as 0142b pattern).
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'metaldocs_security_owner') THEN
    EXECUTE 'ALTER FUNCTION public.enforce_capability_asserted() OWNER TO metaldocs_security_owner';
  END IF;
END $$;
REVOKE EXECUTE ON FUNCTION public.enforce_capability_asserted() FROM PUBLIC;

-- ---------------------------------------------------------------------------
-- 2. Attach trigger to new tables.
--    approval_instances + approval_signoffs already have triggers (0142b).
-- ---------------------------------------------------------------------------

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON metaldocs.iam_user_roles;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE OR DELETE ON metaldocs.iam_user_roles
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON metaldocs.user_process_areas;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE OR DELETE ON metaldocs.user_process_areas
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON public.documents;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE ON public.documents
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON public.controlled_documents;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE ON public.controlled_documents
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON public.cd_sequence_counters;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE ON public.cd_sequence_counters
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON public.document_profiles;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE OR DELETE ON public.document_profiles
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON public.document_process_areas;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE OR DELETE ON public.document_process_areas
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON public.document_families;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE OR DELETE ON public.document_families
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON public.templates_v2_template;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE OR DELETE ON public.templates_v2_template
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

DROP TRIGGER IF EXISTS trg_require_cap_asserted ON public.templates_v2_template_version;
CREATE TRIGGER trg_require_cap_asserted
  BEFORE INSERT OR UPDATE OR DELETE ON public.templates_v2_template_version
  FOR EACH ROW EXECUTE FUNCTION public.enforce_capability_asserted();

-- ---------------------------------------------------------------------------
-- 3. document_families.code immutability trigger (mirrors 0122/0123 pattern).
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION public.reject_families_code_update()
  RETURNS trigger
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, pg_temp
AS $$
BEGIN
  IF NEW.code IS DISTINCT FROM OLD.code THEN
    RAISE EXCEPTION 'document_families.code is immutable'
      USING ERRCODE = 'P0002';
  END IF;
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_reject_families_code_update ON public.document_families;
CREATE TRIGGER trg_reject_families_code_update
  BEFORE UPDATE ON public.document_families
  FOR EACH ROW EXECUTE FUNCTION public.reject_families_code_update();

INSERT INTO public.schema_migrations (version, description)
VALUES ('0188', 'Plan 5: extend tripwire to all regulated tables + document_families code immutability')
ON CONFLICT (version) DO NOTHING;

COMMIT;
```

- [ ] **Step 11.2: Commit migration**

```
git add migrations/0188_tripwire_extend.sql
git commit -m "feat(tripwire): extend enforce_capability_asserted to all regulated tables — Plan 5 (migration 0188)"
```

---

## Task 12 — Apply migrations + run tests

- [ ] **Step 12.1: Start local API + apply migrations**

```powershell
.\scripts\start-api.ps1 -Build
```

This runs pending migrations (0187, 0188) on startup if the migration runner is in the startup path. If migrations are applied separately:

```powershell
# Check how migrations run in this project — look for a goose/migrate/flyway invocation
grep -rn "migrate\|goose\|flyway" scripts/ Makefile 2>/dev/null | head -10
```

- [ ] **Step 12.2: Run full test suite**

```powershell
go test ./...
```

Expected: all pass. If tests fail because of new authz.Require checks blocking existing test-only DB writes:
- Tests that call repo methods directly in integration tests need to call `authz.BypassSystem(ctx, tx)` (sets `metaldocs.bypass_authz = 'scheduler'`) before the DML.
- OR: start the test tx, set actor/tenant GUCs via `SET LOCAL`, then let `authz.Require` pass normally.

- [ ] **Step 12.3: Login and exercise a protected mutation**

```powershell
$token = (Invoke-RestMethod -Uri "http://localhost:8081/api/v1/auth/login" `
  -Method Post -ContentType "application/json" `
  -Body '{"identifier":"admin","password":"AdminMetalDocs123!"}').token
```

Test create document (should pass — admin has `document.create`):

```powershell
Invoke-RestMethod -Uri "http://localhost:8081/api/v2/documents" `
  -Method Post -Headers @{Authorization="Bearer $token"} `
  -ContentType "application/json" `
  -Body '{"name":"Plan5Test","profile_code":"PO","area_code":"QA"}'
```

Test PATCH taxonomy family as unauthenticated (should get 401/403):

```powershell
Invoke-RestMethod -Uri "http://localhost:8081/api/v2/taxonomy/families/PROC" `
  -Method Patch -ContentType "application/json" `
  -Body '{"name":"should fail"}'
```

Expected: 401 Unauthorized (no token) — confirms PATCH families gap is closed.

- [ ] **Step 12.4: Confirm tripwire fires on direct DB access**

Connect to DB with psql (or via e2e test), attempt INSERT into `controlled_documents` without setting GUCs:

```sql
-- Run in psql WITHOUT setting metaldocs.asserted_caps:
INSERT INTO controlled_documents (tenant_id, profile_code, ...)
VALUES ('...', '...', ...);
-- Expected: ERROR P0001 ErrCapabilityNotAsserted
```

---

## Task 13 — Update roadmap + dispatch wiki-curator

- [ ] **Step 13.1: Mark Plan 5 done in roadmap**

In `wiki/backlog/roadmap.md`:
- Change Plan 5 status to `done 2026-05-11` (or current date).
- Add `**PRs:**` line with commit SHA range.
- Update `Last verified` at top of file.

- [ ] **Step 13.2: Close tech-debt rows**

For each Closes item (iam T-004/R-004, documents T-003/R-003, registry T-001/R-001 + T-004/R-004, taxonomy T-003/R-003 + T-006/R-006 + T-013/R-013, templates_v2 T-001/R-001 + T-002/R-002 + T-004/R-004):
- Set status to `closed 2026-05-11` in the relevant `wiki/modules/<m>-tech-debt.md` and `wiki/backlog/<m>-refactor.md` rows.

- [ ] **Step 13.3: Dispatch wiki-curator agent**

Invoke the `wiki-curator` subagent (`metaldocs-module-doc-sync` skill) with:
> "Plan 5 is complete. Update all wiki docs for files touched: iam role_admin_repository + user_area_repository, documents repository, registry service + infrastructure, taxonomy infrastructure family/profile/area repos, templates_v2 service + lifecycle + create, permissions.go. Bump Last verified stamps, refresh Key files line anchors, update wiki/backlog/roadmap.md Closes rows."

- [ ] **Step 13.4: Final commit**

```
git add wiki/ docs/superpowers/specs/
git commit -m "docs: mark Plan 5 done — tripwire + authz.Require coverage complete"
```

---

## Self-Review

### Spec coverage check

| Spec requirement | Task that implements it |
|---|---|
| IAM `iam_user_roles` + `user_process_areas` tier-2 + tripwire | Tasks 5, 6, 11 |
| documents `public.documents` tier-2 + tripwire | Tasks 7, 11 |
| registry `controlled_documents` + `cd_sequence_counters` tier-2 + tripwire | Tasks 3, 4, 8, 11 |
| registry Obsolete/Supersede specific caps | Tasks 1, 3, 4, 8 |
| taxonomy PATCH families gap | Task 2 |
| taxonomy `document_profiles` + `document_process_areas` + `document_families` tier-2 + tripwire (incl. Archive) | Tasks 9, 11 |
| taxonomy `document_families.code` immutability trigger | Task 11 |
| templates_v2 wire real AuthzFunc | Task 10 |
| templates_v2 authz.Require in service mutations | Task 10 |
| templates_v2 T-002 cross-tenant `GetVersionByID` | Task 10 step 10.7 |
| templates_v2 T-004 SoD + content_hash gate | Task 10 step 10.9 |
| templates_v2 + IAM tables tripwire | Task 11 |
| Migration reversibility | Migration uses `DROP TRIGGER IF EXISTS` + `CREATE OR REPLACE FUNCTION` — idempotent, no destructive DDL beyond trigger add |

### Type / naming consistency check

- `iamdomain.CapRegistryObsolete` defined in Task 1, used in Tasks 3, 8. ✓
- `iamdomain.CapRegistrySupersede` defined in Task 1, used in Tasks 3, 8. ✓
- `authz.Require(ctx, tx, string(cap), "tenant")` — consistent signature across all tasks. ✓
- `UpdateStatusTx` defined in Task 8 step 8.6 (interface) and 8.7 (implementation). ✓
- `WithDB(*sql.DB)` defined in Task 10 step 10.1, called in main.go step 10.2. ✓
- `domain.CheckSegregation("approver", ...)` used in Task 10 step 10.9 — signature matches `domain/approval.go:17`. ✓
- `domain.ErrContentHashMismatch` used in Task 10 step 10.9 — defined at `domain/version.go:70`. ✓

### Placeholder scan

- Step 9.2, 9.3, 9.4, 9.5: SQL bodies say "copy existing INSERT/UPDATE query" — implementer must read the actual file before editing. Flagged with `> Read the actual query...` reminders. Acceptable — the concrete SQL is already in the repo; copying it is a 30-second read.
- Step 10.3 revised: notes the tx-scope issue and corrects it inline. ✓
- Step 10.6: says "similar pattern for Review/Approve" — Add explicit cap per method:
  - `Submit` → `CapTemplateSubmit`
  - `Review` (if exists as separate step) → `CapTemplateEdit` (reviewer is editing the review decision)
  - `Approve` → `CapTemplateApprove`
  - `PublishTemplateVersion` → `CapTemplatePublish`
