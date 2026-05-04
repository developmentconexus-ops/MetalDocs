# IAM / RBAC Unification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the two parallel authorization systems (StaticAuthorizer + authz.Require) with one unified DB-driven IAM layer, add document-level visibility, and fix the admin bootstrap catch-22.

**Architecture:** Single `CapabilityService.CanDo(ctx, userID, tenantID, capability)` backed by `iam_user_roles + role_capabilities` tables (plus group membership). The HTTP middleware delegates all permission checks to this service. `authz.Require` gains a `system_admin` bypass so the catch-22 is resolved.

**Tech Stack:** Go 1.22, PostgreSQL 16, `database/sql`, `github.com/jackc/pgx/v5`

**Spec:** `docs/superpowers/specs/2026-05-02-iam-rbac-unification.md`

**Model recommendations per task:**
| Task | Model | Parallel with |
|---|---|---|
| 1 — Migrations | codex | Task 2 |
| 2 — Domain layer | codex | Task 1 |
| 3 — CapabilityService | codex | Task 4 |
| 4 — authz bypass | sonnet | Task 3 |
| 5 — HTTP layer | codex | — (needs Task 3) |
| 6 — Role callers | haiku | Task 3+4 |
| 7 — main.go wiring | sonnet | — (needs Tasks 3+5+6) |
| 8 — Cleanup dead code | haiku | — (needs Task 7) |
| 9 — Wiki | wiki-curator agent | after Task 8 |

**Code style:** /simplify (no unnecessary abstractions). **Prompts:** /caveman (terse, direct).

---

## File Map

### Created
| File | Purpose |
|---|---|
| `migrations/0162_iam_user_roles_tenant_id.sql` | Add tenant_id column |
| `migrations/0163_iam_groups.sql` | Groups tables |
| `migrations/0164_documents_v2_visibility.sql` | Visibility column |
| `migrations/0165_role_capabilities_reseed.sql` | Full capability matrix |
| `migrations/0166_role_rename_reviewer_migration.sql` | Rename admin→system_admin, reviewer→approver |
| `internal/modules/iam/domain/capabilities.go` | 16 Cap* string constants |
| `internal/modules/iam/application/capability_service.go` | CanDo DB check |
| `tests/integration/iam/capability_service_test.go` | Integration tests |

### Modified
| File | Change |
|---|---|
| `internal/modules/iam/domain/model.go` | Add RoleSystemAdmin, RoleAuthor (Task 2); remove RoleAdmin, RoleReviewer, Permission type, Perm* (Task 8) |
| `internal/modules/iam/domain/role_capabilities.go` | Add new roles, bump version |
| `internal/modules/iam/domain/port.go` | Remove Authorizer interface (Task 8) |
| `internal/modules/iam/authz/authz.go` | system_admin bypass before capability check |
| `internal/modules/iam/delivery/http/middleware.go` | Replace authorizer→CapabilityService, resolver returns string |
| `internal/modules/iam/delivery/http/middleware_test.go` | Update NewMiddleware call |
| `internal/modules/iam/delivery/http/admin_handler.go` | Role allowlist update |
| `internal/modules/iam/application/area_membership_test.go` | RoleReviewer→RoleApprover |
| `internal/modules/auth/application/service.go` | RoleAdmin→RoleSystemAdmin |
| `internal/modules/auth/infrastructure/memory/repository.go` | RoleAdmin→RoleSystemAdmin |
| `internal/modules/notifications/delivery/http/handler.go` | RoleAdmin→RoleSystemAdmin |
| `internal/platform/authn/config.go` | RoleAdmin→RoleSystemAdmin |
| `apps/api/cmd/metaldocs-api/permissions.go` | Cap* strings, resolver returns string |
| `apps/api/cmd/metaldocs-api/permissions_test.go` | Cap* expected values |
| `apps/api/cmd/metaldocs-api/main.go` | Wire CapabilityService |
| `apps/api/cmd/metaldocs-e2e-seed/main.go` | RoleAdmin→RoleSystemAdmin |

### Deleted (Task 8)
| File | Reason |
|---|---|
| `internal/modules/iam/application/authorizer.go` | StaticAuthorizer replaced by CapabilityService |
| `internal/modules/iam/application/authorizer_test.go` | Tests deleted files |

---

## Task 1: SQL Migrations (0162–0166)

**Model:** codex ⚡ Parallel with Task 2

**Files:**
- Create: `migrations/0162_iam_user_roles_tenant_id.sql`
- Create: `migrations/0163_iam_groups.sql`
- Create: `migrations/0164_documents_v2_visibility.sql`
- Create: `migrations/0165_role_capabilities_reseed.sql`
- Create: `migrations/0166_role_rename_reviewer_migration.sql`

- [ ] **Step 1: Create 0162**

```sql
-- migrations/0162_iam_user_roles_tenant_id.sql
ALTER TABLE metaldocs.iam_user_roles
  ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL
    DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff';
```

- [ ] **Step 2: Create 0163**

```sql
-- migrations/0163_iam_groups.sql
CREATE TABLE IF NOT EXISTS metaldocs.iam_groups (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID NOT NULL DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff',
  name        TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, name)
);

CREATE TABLE IF NOT EXISTS metaldocs.iam_group_members (
  group_id   UUID NOT NULL REFERENCES metaldocs.iam_groups(id) ON DELETE CASCADE,
  user_id    TEXT NOT NULL,
  tenant_id  UUID NOT NULL,
  granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  granted_by TEXT,
  PRIMARY KEY (group_id, user_id)
);

CREATE TABLE IF NOT EXISTS metaldocs.iam_group_roles (
  group_id UUID NOT NULL REFERENCES metaldocs.iam_groups(id) ON DELETE CASCADE,
  role     TEXT NOT NULL,
  PRIMARY KEY (group_id, role)
);
```

- [ ] **Step 3: Create 0164**

```sql
-- migrations/0164_documents_v2_visibility.sql
ALTER TABLE public.documents_v2
  ADD COLUMN IF NOT EXISTS visibility TEXT NOT NULL DEFAULT 'area'
  CONSTRAINT documents_v2_visibility_check
    CHECK (visibility IN ('public', 'area', 'restricted'));
```

- [ ] **Step 4: Create 0165**

```sql
-- migrations/0165_role_capabilities_reseed.sql
TRUNCATE metaldocs.role_capabilities;

INSERT INTO metaldocs.role_capabilities (role, capability, description) VALUES
  ('viewer',       'doc.view',           'View documents'),
  ('viewer',       'template.view',      'View templates'),
  ('editor',       'doc.view',           'View documents'),
  ('editor',       'doc.create',         'Create document versions'),
  ('editor',       'doc.edit',           'Edit document drafts'),
  ('editor',       'template.view',      'View templates'),
  ('editor',       'template.edit',      'Edit template drafts'),
  ('editor',       'registry.create',    'Register controlled documents'),
  ('author',       'doc.view',           'View documents'),
  ('author',       'doc.create',         'Create document versions'),
  ('author',       'doc.edit',           'Edit document drafts'),
  ('author',       'doc.submit',         'Submit documents for approval'),
  ('author',       'template.view',      'View templates'),
  ('author',       'template.create',    'Create templates'),
  ('author',       'template.edit',      'Edit template drafts'),
  ('author',       'template.submit',    'Submit templates for approval'),
  ('author',       'registry.create',    'Register controlled documents'),
  ('approver',     'doc.view',           'View documents'),
  ('approver',     'doc.create',         'Create document versions'),
  ('approver',     'doc.edit',           'Edit document drafts'),
  ('approver',     'doc.submit',         'Submit documents for approval'),
  ('approver',     'doc.signoff',        'Sign off document approvals'),
  ('approver',     'template.view',      'View templates'),
  ('approver',     'template.approve',   'Approve template versions'),
  ('system_admin', 'doc.view',           'View documents'),
  ('system_admin', 'doc.create',         'Create document versions'),
  ('system_admin', 'doc.edit',           'Edit document drafts'),
  ('system_admin', 'doc.submit',         'Submit documents for approval'),
  ('system_admin', 'doc.signoff',        'Sign off document approvals'),
  ('system_admin', 'template.view',      'View templates'),
  ('system_admin', 'template.create',    'Create templates'),
  ('system_admin', 'template.edit',      'Edit template drafts'),
  ('system_admin', 'template.submit',    'Submit templates for approval'),
  ('system_admin', 'template.approve',   'Approve template versions'),
  ('system_admin', 'template.publish',   'Publish template versions'),
  ('system_admin', 'registry.create',    'Register controlled documents'),
  ('system_admin', 'taxonomy.manage',    'Manage taxonomy'),
  ('system_admin', 'membership.manage',  'Manage memberships'),
  ('system_admin', 'route.manage',       'Manage approval routes'),
  ('system_admin', 'user.manage',        'Manage users');
```

- [ ] **Step 5: Create 0166**

```sql
-- migrations/0166_role_rename_reviewer_migration.sql
BEGIN;

-- Drop constraint first (constraint does not allow 'system_admin' yet)
ALTER TABLE metaldocs.iam_user_roles
  DROP CONSTRAINT IF EXISTS chk_iam_user_roles_role_code;

-- Rename rows BEFORE adding new constraint (old values admin/reviewer not in new constraint)
UPDATE metaldocs.iam_user_roles SET role_code = 'system_admin' WHERE role_code = 'admin';
UPDATE metaldocs.iam_user_roles SET role_code = 'approver'     WHERE role_code = 'reviewer';

-- Add new constraint (all rows now valid)
ALTER TABLE metaldocs.iam_user_roles
  ADD CONSTRAINT chk_iam_user_roles_role_code
    CHECK (role_code IN ('system_admin', 'approver', 'author', 'editor', 'viewer'));

-- Enforce one role per user per tenant (clean duplicates first in dev)
DELETE FROM metaldocs.iam_user_roles a
  USING metaldocs.iam_user_roles b
  WHERE a.ctid < b.ctid
    AND a.user_id   = b.user_id
    AND a.tenant_id = b.tenant_id;

ALTER TABLE metaldocs.iam_user_roles
  DROP CONSTRAINT IF EXISTS uq_iam_user_roles_user_tenant;

ALTER TABLE metaldocs.iam_user_roles
  ADD CONSTRAINT uq_iam_user_roles_user_tenant UNIQUE (tenant_id, user_id);

COMMIT;
```

- [ ] **Step 6: Run migrations**

```powershell
.\scripts\dev-migrate.ps1
```

Expected: no errors. Verify:
```sql
-- psql
SELECT role_code, COUNT(*) FROM metaldocs.iam_user_roles GROUP BY role_code;
-- should show system_admin (not admin), no reviewer rows
SELECT COUNT(*) FROM metaldocs.role_capabilities;
-- should be 40
SELECT COUNT(*) FROM information_schema.columns
  WHERE table_name = 'iam_user_roles' AND column_name = 'tenant_id';
-- should be 1
SELECT COUNT(*) FROM information_schema.tables
  WHERE table_name IN ('iam_groups','iam_group_members','iam_group_roles');
-- should be 3
SELECT COUNT(*) FROM information_schema.columns
  WHERE table_name = 'documents_v2' AND column_name = 'visibility';
-- should be 1
```

- [ ] **Step 7: Commit**

```bash
git add migrations/0162_iam_user_roles_tenant_id.sql \
        migrations/0163_iam_groups.sql \
        migrations/0164_documents_v2_visibility.sql \
        migrations/0165_role_capabilities_reseed.sql \
        migrations/0166_role_rename_reviewer_migration.sql
git commit -m "feat(iam): add IAM unification migrations 0162-0166"
```

---

## Task 2: Domain Layer — New Symbols

**Model:** codex ⚡ Parallel with Task 1

**Files:**
- Create: `internal/modules/iam/domain/capabilities.go`
- Modify: `internal/modules/iam/domain/model.go`
- Modify: `internal/modules/iam/domain/role_capabilities.go`

> Note: Do NOT remove `RoleAdmin`, `RoleReviewer`, `Permission` type or `Perm*` constants yet. Removal is Task 8. Only ADD new symbols to avoid breaking callers.

- [ ] **Step 1: Create capabilities.go**

```go
// internal/modules/iam/domain/capabilities.go
package domain

const (
	CapDocView          = "doc.view"
	CapDocCreate        = "doc.create"
	CapDocEdit          = "doc.edit"
	CapDocSubmit        = "doc.submit"
	CapDocSignoff       = "doc.signoff"
	CapTemplateView     = "template.view"
	CapTemplateCreate   = "template.create"
	CapTemplateEdit     = "template.edit"
	CapTemplateSubmit   = "template.submit"
	CapTemplateApprove  = "template.approve"
	CapTemplatePublish  = "template.publish"
	CapRegistryCreate   = "registry.create"
	CapTaxonomyManage   = "taxonomy.manage"
	CapMembershipManage = "membership.manage"
	CapRouteManage      = "route.manage"
	CapUserManage       = "user.manage"
)
```

- [ ] **Step 2: Add new role constants to model.go**

Add these two lines to the `Role` constant block in `internal/modules/iam/domain/model.go` (after `RoleViewer`). Keep existing constants untouched:

```go
const (
	RoleAdmin       Role = "admin"       // keep — removed in Task 8
	RoleApprover    Role = "approver"
	RoleAuthor      Role = "author"      // new
	RoleEditor      Role = "editor"
	RoleReviewer    Role = "reviewer"    // keep — removed in Task 8
	RoleSystemAdmin Role = "system_admin" // new
	RoleViewer      Role = "viewer"
)
```

- [ ] **Step 3: Update role_capabilities.go — add new roles, bump version**

Full replacement of `internal/modules/iam/domain/role_capabilities.go`:

```go
package domain

const RoleCapabilitiesVersion = 2

var RoleCapabilities = map[Role][]Capability{
	RoleViewer: {
		CapDocumentView,
		CapTemplateView,
	},
	RoleEditor: {
		CapDocumentView,
		CapDocumentCreate,
		CapDocumentEdit,
		CapTemplateView,
	},
	RoleReviewer: { // keep for backward compat during transition
		CapDocumentView,
		CapDocumentEdit,
		CapWorkflowReview,
		CapTemplateView,
	},
	RoleApprover: {
		CapDocumentView,
		CapWorkflowApprove,
		CapTemplateView,
		CapTemplatePublish,
	},
	RoleAuthor: {
		CapDocumentView,
		CapDocumentCreate,
		CapDocumentEdit,
	},
	RoleSystemAdmin: {
		CapDocumentView,
		CapDocumentCreate,
		CapDocumentEdit,
		CapTemplateView,
		CapTemplatePublish,
		CapWorkflowApprove,
		CapRegistryCreate,
	},
}
```

- [ ] **Step 4: Build to confirm no compile errors**

```bash
go build ./internal/modules/iam/...
```

Expected: PASS (zero errors — we only added symbols)

- [ ] **Step 5: Commit**

```bash
git add internal/modules/iam/domain/capabilities.go \
        internal/modules/iam/domain/model.go \
        internal/modules/iam/domain/role_capabilities.go
git commit -m "feat(iam): add new role/capability constants (system_admin, author, Cap*)"
```

---

## Task 3: CapabilityService

**Model:** codex ⚡ Parallel with Task 4 (and Task 6 — they touch different files)

**Files:**
- Create: `internal/modules/iam/application/capability_service.go`
- Create: `tests/integration/iam/capability_service_test.go`

- [ ] **Step 1: Write the failing integration test**

```go
// tests/integration/iam/capability_service_test.go
//go:build integration

package iam_test

import (
	"context"
	"testing"

	"metaldocs/internal/modules/iam/application"
	"metaldocs/tests/integration/testdb"
)

const devTenant = "ffffffff-ffff-ffff-ffff-ffffffffffff"

// Use testdb.DeterministicID to avoid conflicts across parallel runs.
func TestCapabilityService_SystemAdmin_AllowsAnyCapability(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()
	adminID := testdb.DeterministicID(t, "admin")

	if _, err := db.ExecContext(ctx, `
		INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code)
		VALUES ($1, $2::uuid, 'system_admin')
	`, adminID, devTenant); err != nil {
		t.Fatalf("setup system_admin: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), `DELETE FROM metaldocs.iam_user_roles WHERE user_id = $1`, adminID) //nolint:errcheck
	})

	svc := application.NewCapabilityService(db)
	caps := []string{"doc.view", "doc.create", "doc.submit", "doc.signoff",
		"user.manage", "taxonomy.manage", "membership.manage"}

	for _, cap := range caps {
		if err := svc.CanDo(ctx, adminID, devTenant, cap); err != nil {
			t.Errorf("system_admin denied %q: %v", cap, err)
		}
	}
}

func TestCapabilityService_Viewer_AllowedDocView_DeniedDocCreate(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()
	viewerID := testdb.DeterministicID(t, "viewer")

	_, err := db.ExecContext(ctx, `
		INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code)
		VALUES ($1, $2::uuid, 'viewer')
	`, viewerID, devTenant)
	if err != nil {
		t.Fatalf("setup viewer: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), `DELETE FROM metaldocs.iam_user_roles WHERE user_id = $1`, viewerID)
	})

	svc := application.NewCapabilityService(db)

	if err := svc.CanDo(ctx, viewerID, devTenant, "doc.view"); err != nil {
		t.Errorf("viewer should see docs: %v", err)
	}
	if err := svc.CanDo(ctx, viewerID, devTenant, "doc.create"); err == nil {
		t.Error("viewer should not create docs")
	}
}

func TestCapabilityService_UnknownUser_Denied(t *testing.T) {
	db, _ := testdb.Open(t)
	svc := application.NewCapabilityService(db)
	unknownID := testdb.DeterministicID(t, "nobody")

	if err := svc.CanDo(context.Background(), unknownID, devTenant, "doc.view"); err == nil {
		t.Error("unknown user should be denied")
	}
}
```

- [ ] **Step 2: Run test — expect FAIL (package doesn't exist)**

```bash
go test -tags integration ./tests/integration/iam/... -v
```

Expected: compile error — `application.NewCapabilityService` undefined

- [ ] **Step 3: Implement CapabilityService**

```go
// internal/modules/iam/application/capability_service.go
package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrCapabilityDenied = errors.New("capability denied")

type CapabilityService struct {
	db *sql.DB
}

func NewCapabilityService(db *sql.DB) *CapabilityService {
	return &CapabilityService{db: db}
}

// CanDo returns nil if userID holds capability in tenant.
// system_admin role bypasses all capability checks.
// Checks direct role then group membership (OR logic).
func (s *CapabilityService) CanDo(ctx context.Context, userID, tenantID, capability string) error {
	if s.db == nil {
		return ErrCapabilityDenied
	}
	var allowed bool
	err := s.db.QueryRowContext(ctx, `
SELECT EXISTS (
  SELECT 1 FROM metaldocs.iam_user_roles ur
  JOIN metaldocs.role_capabilities rc ON rc.role = ur.role_code
  WHERE ur.user_id   = $1
    AND ur.tenant_id = $2::uuid
    AND (ur.role_code = 'system_admin' OR rc.capability = $3)
) OR EXISTS (
  SELECT 1 FROM metaldocs.iam_group_members gm
  JOIN metaldocs.iam_group_roles gr ON gr.group_id = gm.group_id
  JOIN metaldocs.role_capabilities rc ON rc.role = gr.role
  WHERE gm.user_id   = $1
    AND gm.tenant_id = $2::uuid
    AND rc.capability = $3
)`, userID, tenantID, capability).Scan(&allowed)
	if err != nil {
		return fmt.Errorf("capability check: %w", err)
	}
	if !allowed {
		return ErrCapabilityDenied
	}
	return nil
}
```

- [ ] **Step 4: Run tests — expect PASS**

```bash
go test -tags integration ./tests/integration/iam/... -v
```

Expected: 3 tests PASS

- [ ] **Step 5: Build full module**

```bash
go build ./internal/modules/iam/...
go build ./apps/api/cmd/metaldocs-api/...
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/modules/iam/application/capability_service.go \
        tests/integration/iam/capability_service_test.go
git commit -m "feat(iam): add CapabilityService with DB-backed CanDo"
```

---

## Task 4: authz.Require — system_admin Bypass

**Model:** sonnet ⚡ Parallel with Task 3 (different file)

**Files:**
- Modify: `internal/modules/iam/authz/authz.go`

- [ ] **Step 1: Write the failing test for bypass**

In `internal/modules/iam/authz/authz_test.go` (create if not exists — check with `ls internal/modules/iam/authz/`):

```go
// internal/modules/iam/authz/authz_bypass_test.go
//go:build integration

package authz_test

import (
	"context"
	"testing"

	"metaldocs/tests/integration/testdb"
	"metaldocs/internal/modules/iam/authz"
)

func TestRequire_SystemAdmin_Bypasses(t *testing.T) {
	db, schema := testdb.Open(t)
	_ = schema
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	// Set session vars for admin-local user
	_, err = tx.ExecContext(ctx,
		"SELECT set_config('metaldocs.actor_id', 'admin-local', true), set_config('metaldocs.tenant_id', 'ffffffff-ffff-ffff-ffff-ffffffffffff', true)")
	if err != nil {
		t.Fatalf("set session vars: %v", err)
	}

	// system_admin should bypass even a non-existent capability
	ctx = authz.WithCapCache(ctx)
	if err := authz.Require(ctx, tx, "does.not.exist", "any-area"); err != nil {
		t.Fatalf("system_admin bypass failed: %v", err)
	}
}
```

- [ ] **Step 2: Run test — expect FAIL (no bypass yet)**

```bash
go test -tags integration ./internal/modules/iam/authz/... -run TestRequire_SystemAdmin_Bypasses -v
```

Expected: FAIL — `authz: capability "does.not.exist" denied for actor "admin-local"`

- [ ] **Step 3: Add system_admin bypass to Require**

Edit `internal/modules/iam/authz/authz.go`. Insert BEFORE the existing `tx.QueryRowContext` that checks role_capabilities. The new `Require` function:

```go
func Require(ctx context.Context, tx *sql.Tx, capability, areaCode string) error {
	if cacheGranted(ctx, capability, areaCode) {
		return appendAssertedCap(ctx, tx, capability, areaCode)
	}

	// system_admin bypass — must check before capability query
	var isAdmin bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		  SELECT 1 FROM metaldocs.iam_user_roles
		   WHERE user_id   = current_setting('metaldocs.actor_id', false)
		     AND tenant_id = current_setting('metaldocs.tenant_id', false)::uuid
		     AND role_code = 'system_admin'
		)
	`).Scan(&isAdmin); err != nil {
		return fmt.Errorf("authz: system_admin check: %w", err)
	}
	if isAdmin {
		storeGranted(ctx, capability, areaCode)
		return appendAssertedCap(ctx, tx, capability, areaCode)
	}

	var granted bool
	err := tx.QueryRowContext(ctx, `
SELECT EXISTS (
  SELECT 1
  FROM metaldocs.role_capabilities rc
  JOIN metaldocs.user_process_areas upa
    ON upa.role = rc.role
   AND upa.tenant_id = current_setting('metaldocs.tenant_id', false)::uuid
   AND upa.user_id   = current_setting('metaldocs.actor_id', false)
   AND upa.effective_to IS NULL
  WHERE rc.capability = $1
    AND ($2 = 'tenant' OR upa.area_code = $2)
)`,
		capability, areaCode,
	).Scan(&granted)
	if err != nil {
		return err
	}

	if !granted {
		actorID, err := actorIDFromTx(ctx, tx)
		if err != nil {
			return err
		}
		return ErrCapabilityDenied{
			Capability: capability,
			AreaCode:   areaCode,
			ActorID:    actorID,
		}
	}

	storeGranted(ctx, capability, areaCode)
	return appendAssertedCap(ctx, tx, capability, areaCode)
}
```

Also add `"fmt"` to imports if not present.

- [ ] **Step 4: Run test — expect PASS**

```bash
go test -tags integration ./internal/modules/iam/authz/... -run TestRequire_SystemAdmin_Bypasses -v
```

Expected: PASS

- [ ] **Step 5: Build**

```bash
go build ./internal/modules/iam/authz/...
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/modules/iam/authz/authz.go \
        internal/modules/iam/authz/authz_bypass_test.go
git commit -m "fix(authz): add system_admin bypass to Require — fixes bootstrap catch-22"
```

---

## Task 5: HTTP Layer — Middleware + Permissions

**Model:** codex (depends on Tasks 2+3)

**Files:**
- Modify: `internal/modules/iam/delivery/http/middleware.go`
- Modify: `internal/modules/iam/delivery/http/middleware_test.go`
- Modify: `apps/api/cmd/metaldocs-api/permissions.go`
- Modify: `apps/api/cmd/metaldocs-api/permissions_test.go`

> Context: `middleware.go` currently holds `authorizer iamdomain.Authorizer` and a `requiredPermission` fallback. `permissions.go` holds `newPermissionResolver()` which maps routes to `iamdomain.Permission` values. Both change in this task.

- [ ] **Step 1: Update middleware.go**

Full replacement of `internal/modules/iam/delivery/http/middleware.go`:

```go
package httpdelivery

import (
	"errors"
	"net/http"
	"strings"

	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	authdomain "metaldocs/internal/modules/auth/domain"
	"metaldocs/internal/platform/httpresponse"
)

type ctxKeyCapability struct{}
type ctxKeyAreaCode struct{}
type ctxKeyResourceID struct{}

var writeJSON = httpresponse.WriteJSON

// PermissionResolver maps an HTTP method+path to a capability string.
// Returns ("", false) for unguarded routes.
type PermissionResolver func(method, path string) (string, bool)

type Middleware struct {
	caps         *iamapp.CapabilityService
	roleProvider iamdomain.RoleProvider
	enabled      bool
	legacyHeader bool
	resolver     PermissionResolver
}

func NewMiddleware(caps *iamapp.CapabilityService, roleProvider iamdomain.RoleProvider, enabled bool, legacyHeader ...bool) *Middleware {
	allowLegacy := false
	if len(legacyHeader) > 0 {
		allowLegacy = legacyHeader[0]
	}
	return &Middleware{
		caps:         caps,
		roleProvider: roleProvider,
		enabled:      enabled,
		legacyHeader: allowLegacy,
	}
}

func (m *Middleware) WithPermissionResolver(resolver PermissionResolver) *Middleware {
	if resolver != nil {
		m.resolver = resolver
	}
	return m
}

func (m *Middleware) Wrap(next http.Handler) http.Handler {
	if !m.enabled {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always strip — security invariant regardless of whether route is guarded.
		r.Header.Del("X-User-ID")

		if m.resolver == nil {
			next.ServeHTTP(w, r)
			return
		}

		capability, guarded := m.resolver(r.Method, r.URL.Path)
		if !guarded {
			next.ServeHTTP(w, r)
			return
		}

		traceID := requestTraceID(r)
		userID := iamdomain.UserIDFromContext(r.Context())
		if userID == "" && m.legacyHeader {
			userID = strings.TrimSpace(r.Header.Get("X-User-Id"))
		}
		if userID == "" {
			writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "Authentication required", traceID)
			return
		}

		if m.caps != nil {
			tenantID := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
			if tenantID == "" {
				tenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
			}
			if err := m.caps.CanDo(r.Context(), userID, tenantID, capability); err != nil {
				writeAPIError(w, http.StatusForbidden, "AUTH_FORBIDDEN", "Insufficient permissions", traceID)
				return
			}
		}

		ctx := r.Context()
		if _, ok := authdomain.CurrentUserFromContext(ctx); !ok {
			var roles []iamdomain.Role
			if m.roleProvider != nil {
				resolvedRoles, err := m.roleProvider.RolesByUserID(r.Context(), userID)
				if err != nil {
					if errors.Is(err, iamdomain.ErrUserNotFound) || errors.Is(err, iamdomain.ErrUserInactive) {
						writeAPIError(w, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "User is not authorized", traceID)
						return
					}
					writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Authorization lookup failed", traceID)
					return
				}
				roles = resolvedRoles
			}
			ctx = iamdomain.WithAuthContext(ctx, userID, roles)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type apiErrorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
	TraceID string         `json:"trace_id"`
}

func requestTraceID(r *http.Request) string {
	if traceID := strings.TrimSpace(r.Header.Get("X-Trace-Id")); traceID != "" {
		return traceID
	}
	return "trace-local"
}

func writeAPIError(w http.ResponseWriter, status int, code, message, traceID string) {
	httpresponse.WriteJSON(w, status, apiErrorEnvelope{
		Error: apiError{
			Code:    code,
			Message: message,
			Details: map[string]any{},
			TraceID: traceID,
		},
	})
}

func NewV2AuthzMiddleware(service *iamapp.AuthorizationService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if service == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCap := r.Context().Value(ctxKeyCapability{})
			if rawCap == nil {
				next.ServeHTTP(w, r)
				return
			}

			var capability iamdomain.Capability
			switch value := rawCap.(type) {
			case iamdomain.Capability:
				capability = value
			case string:
				capability = iamdomain.Capability(strings.TrimSpace(value))
			}
			if capability == "" {
				next.ServeHTTP(w, r)
				return
			}

			areaCode, _ := r.Context().Value(ctxKeyAreaCode{}).(string)
			areaCode = strings.TrimSpace(areaCode)
			if areaCode == "" {
				areaCode = strings.TrimSpace(r.Header.Get("X-Area-Code"))
			}

			resourceID, _ := r.Context().Value(ctxKeyResourceID{}).(string)
			userID := strings.TrimSpace(iamdomain.UserIDFromContext(r.Context()))
			if userID == "" {
				userID = strings.TrimSpace(r.Header.Get("X-User-Id"))
			}
			tenantID := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
			if tenantID == "" {
				httpresponse.WriteJSON(w, http.StatusBadRequest, map[string]any{
					"code":                "missing_tenant",
					"required_capability": capability,
					"area_code":           areaCode,
				})
				return
			}

			err := service.Check(r.Context(), userID, tenantID, capability, iamapp.ResourceCtx{
				AreaCode:   areaCode,
				ResourceID: strings.TrimSpace(resourceID),
			})
			if err != nil {
				httpresponse.WriteJSON(w, http.StatusForbidden, map[string]any{
					"code":                err.Error(),
					"required_capability": capability,
					"area_code":           areaCode,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
```

- [ ] **Step 2: Update middleware_test.go**

```go
// internal/modules/iam/delivery/http/middleware_test.go
package httpdelivery

import (
	"net/http"
	"net/http/httptest"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

func TestMiddlewareStripsUserIDHeaderAfterAuthContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	req.Header.Set("X-User-ID", "attacker")
	req = req.WithContext(iamdomain.WithAuthContext(req.Context(), "real-user", []iamdomain.Role{iamdomain.RoleSystemAdmin}))

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if got := iamdomain.UserIDFromContext(r.Context()); got != "real-user" {
			t.Fatalf("UserIDFromContext() = %q, want %q", got, "real-user")
		}
		if got := r.Header.Get("X-User-ID"); got != "" {
			t.Fatalf("X-User-ID header = %q, want empty", got)
		}
	})

	// nil caps and nil resolver: header still stripped, route unguarded, next called
	middleware := NewMiddleware(nil, nil, true)
	middleware.Wrap(next).ServeHTTP(httptest.NewRecorder(), req)

	if !called {
		t.Fatal("next handler was not called")
	}
}
```

- [ ] **Step 3: Run middleware test**

```bash
go test ./internal/modules/iam/delivery/http/... -v
```

Expected: PASS

- [ ] **Step 4: Update permissions.go**

Full replacement of `apps/api/cmd/metaldocs-api/permissions.go`:

```go
package main

import (
	"net/http"
	"strings"

	authdelivery "metaldocs/internal/modules/auth/delivery/http"
	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

func newPermissionResolver() iamdelivery.PermissionResolver {
	return func(method, path string) (string, bool) {
		// Public: health, auth, feature-flags
		if path == "/api/v1/health/live" || path == "/api/v1/health/ready" {
			return "", false
		}
		if strings.HasPrefix(path, "/api/v1/auth/") {
			return "", false
		}
		if method == http.MethodGet && path == "/api/v1/feature-flags" {
			return "", false
		}

		// Metrics — admin only
		if path == "/api/v1/metrics" {
			return iamdomain.CapUserManage, true
		}

		// V1 documents
		if method == http.MethodPost && path == "/api/v1/documents" {
			return iamdomain.CapDocCreate, true
		}
		if method == http.MethodPost && strings.HasPrefix(path, "/api/v1/documents/") && strings.HasSuffix(path, "/attachments") {
			return iamdomain.CapDocEdit, true
		}
		if method == http.MethodGet && path == "/api/v1/documents" {
			return iamdomain.CapDocView, true
		}
		if method == http.MethodGet && path == "/api/v1/document-types" {
			return iamdomain.CapDocView, true
		}
		if method == http.MethodGet && path == "/api/v1/document-templates" {
			return iamdomain.CapDocView, true
		}
		if method == http.MethodGet && strings.HasPrefix(path, "/api/v1/documents/") && !strings.HasSuffix(path, "/versions") {
			return iamdomain.CapDocView, true
		}
		if method == http.MethodPost && strings.HasPrefix(path, "/api/v1/documents/") && strings.HasSuffix(path, "/submit-for-approval") {
			return iamdomain.CapDocSubmit, true
		}
		if method == http.MethodPut && strings.HasPrefix(path, "/api/v1/documents/") && strings.HasSuffix(path, "/content") {
			return iamdomain.CapDocEdit, true
		}
		if method == http.MethodPost && strings.HasPrefix(path, "/api/v1/documents/") && strings.HasSuffix(path, "/content/browser") {
			return iamdomain.CapDocEdit, true
		}
		if method == http.MethodPut && strings.HasPrefix(path, "/api/v1/documents/") && strings.HasSuffix(path, "/template-assignment") {
			return iamdomain.CapDocEdit, true
		}
		if method == http.MethodPost && strings.HasPrefix(path, "/api/v1/documents/") && strings.HasSuffix(path, "/versions") {
			return iamdomain.CapDocEdit, true
		}
		if method == http.MethodPost && strings.HasPrefix(path, "/api/v1/documents/") && strings.HasSuffix(path, "/export/docx") {
			return iamdomain.CapDocView, true
		}
		if method == http.MethodGet && path == "/api/v1/search/documents" {
			return iamdomain.CapDocView, true
		}
		if method == http.MethodGet && path == "/api/v1/notifications" {
			return iamdomain.CapDocView, true
		}
		if method == http.MethodPost && strings.HasPrefix(path, "/api/v1/notifications/") && strings.HasSuffix(path, "/read") {
			return iamdomain.CapDocView, true
		}
		if method == http.MethodGet && strings.HasPrefix(path, "/api/v1/documents/") && strings.Contains(path, "/versions/diff") {
			return iamdomain.CapDocView, true
		}
		if (method == http.MethodGet || method == http.MethodPut) && path == "/api/v1/access-policies" {
			return iamdomain.CapMembershipManage, true
		}
		if method == http.MethodGet && strings.HasPrefix(path, "/api/v1/documents/") && strings.HasSuffix(path, "/versions") {
			return iamdomain.CapDocView, true
		}
		if method == http.MethodPost && strings.HasPrefix(path, "/api/v1/workflow/documents/") && strings.HasSuffix(path, "/transitions") {
			return iamdomain.CapDocSubmit, true
		}
		if method == http.MethodGet && strings.HasPrefix(path, "/api/v1/workflow/documents/") && strings.HasSuffix(path, "/approvals") {
			return iamdomain.CapDocView, true
		}

		// V1 IAM users
		if method == http.MethodPost && path == "/api/v1/iam/users" {
			return iamdomain.CapUserManage, true
		}
		if method == http.MethodGet && path == "/api/v1/iam/users" {
			return iamdomain.CapUserManage, true
		}
		if method == http.MethodPatch && strings.HasPrefix(path, "/api/v1/iam/users/") && !strings.HasSuffix(path, "/roles") {
			return iamdomain.CapUserManage, true
		}
		if method == http.MethodPost && strings.HasPrefix(path, "/api/v1/iam/users/") && strings.HasSuffix(path, "/roles") {
			return iamdomain.CapUserManage, true
		}
		if method == http.MethodPut && strings.HasPrefix(path, "/api/v1/iam/users/") && strings.HasSuffix(path, "/roles") {
			return iamdomain.CapUserManage, true
		}
		if method == http.MethodPost && strings.HasPrefix(path, "/api/v1/iam/users/") && strings.HasSuffix(path, "/reset-password") {
			return iamdomain.CapUserManage, true
		}
		if method == http.MethodPost && strings.HasPrefix(path, "/api/v1/iam/users/") && strings.HasSuffix(path, "/unlock") {
			return iamdomain.CapUserManage, true
		}
		if method == http.MethodGet && path == "/api/v1/iam/admin/overview" {
			return iamdomain.CapUserManage, true
		}

		// V1 taxonomy (profiles, areas, subjects, document-profiles)
		if strings.HasPrefix(path, "/api/v1/document-profiles") {
			switch method {
			case http.MethodGet:
				return iamdomain.CapDocView, true
			case http.MethodPost, http.MethodPut, http.MethodDelete:
				return iamdomain.CapTaxonomyManage, true
			}
		}
		if strings.HasPrefix(path, "/api/v1/process-areas") {
			switch method {
			case http.MethodGet:
				return iamdomain.CapDocView, true
			case http.MethodPost, http.MethodPut, http.MethodDelete:
				return iamdomain.CapTaxonomyManage, true
			}
		}
		if strings.HasPrefix(path, "/api/v1/document-subjects") {
			switch method {
			case http.MethodGet:
				return iamdomain.CapDocView, true
			case http.MethodPost, http.MethodPut, http.MethodDelete:
				return iamdomain.CapTaxonomyManage, true
			}
		}

		// V1 templates (fine-grained RBAC in service layer)
		if path == "/api/v1/templates" || strings.HasPrefix(path, "/api/v1/templates/") {
			return iamdomain.CapTemplateView, true
		}

		// V2 templates
		if strings.HasPrefix(path, "/api/v2/templates") {
			switch {
			case method == http.MethodGet:
				return iamdomain.CapTemplateView, true
			case method == http.MethodPost && path == "/api/v2/templates":
				return iamdomain.CapTemplateEdit, true
			case method == http.MethodPut && strings.HasSuffix(path, "/draft"):
				return iamdomain.CapTemplateEdit, true
			case method == http.MethodPost && strings.HasSuffix(path, "/publish"):
				return iamdomain.CapTemplatePublish, true
			case method == http.MethodPost && strings.HasSuffix(path, "/docx-upload-url"):
				return iamdomain.CapTemplateEdit, true
			case method == http.MethodPost && strings.HasSuffix(path, "/schema-upload-url"):
				return iamdomain.CapTemplateEdit, true
			case method == http.MethodPost && strings.HasSuffix(path, "/submit"):
				return iamdomain.CapTemplateSubmit, true
			case method == http.MethodPost && strings.HasSuffix(path, "/review"):
				return iamdomain.CapTemplateApprove, true
			case method == http.MethodPost && strings.HasSuffix(path, "/approve"):
				return iamdomain.CapTemplateApprove, true
			case method == http.MethodPost && strings.HasSuffix(path, "/approval-config"):
				return iamdomain.CapTemplateEdit, true
			}
		}

		// V2 documents
		if strings.HasPrefix(path, "/api/v2/documents") {
			switch {
			case method == http.MethodGet:
				return iamdomain.CapDocView, true
			case method == http.MethodPost && path == "/api/v2/documents":
				return iamdomain.CapDocCreate, true
			case method == http.MethodPost && strings.HasSuffix(path, "/finalize"):
				return iamdomain.CapDocSignoff, true
			case method == http.MethodPost && strings.HasSuffix(path, "/archive"):
				return iamdomain.CapDocEdit, true
			case method == http.MethodPost && strings.Contains(path, "/session/force-release"):
				return iamdomain.CapMembershipManage, true
			case method == http.MethodPost && strings.Contains(path, "/session/"):
				return iamdomain.CapDocEdit, true
			case method == http.MethodPost && strings.Contains(path, "/autosave/"):
				return iamdomain.CapDocEdit, true
			case method == http.MethodPost && strings.Contains(path, "/checkpoints/") && strings.HasSuffix(path, "/restore"):
				return iamdomain.CapDocEdit, true
			case method == http.MethodPost && strings.Contains(path, "/checkpoints"):
				return iamdomain.CapDocEdit, true
			case method == http.MethodPost && strings.HasSuffix(path, "/export/pdf"):
				return iamdomain.CapDocView, true
			case method == http.MethodPut && strings.Contains(path, "/placeholders/"):
				return iamdomain.CapDocEdit, true
			case method == http.MethodPatch:
				return iamdomain.CapDocEdit, true
			case method == http.MethodPost && strings.HasSuffix(path, "/submit"):
				return iamdomain.CapDocSubmit, true
			case method == http.MethodPost && strings.HasSuffix(path, "/signoff"):
				return iamdomain.CapDocSignoff, true
			case method == http.MethodPost && strings.HasSuffix(path, "/cancel"):
				return iamdomain.CapDocEdit, true
			case method == http.MethodGet && strings.HasSuffix(path, "/approval-instance"):
				return iamdomain.CapDocView, true
			case method == http.MethodPost && strings.HasSuffix(path, "/reconstruct"):
				return iamdomain.CapDocEdit, true
			}
		}

		// V2 taxonomy
		if strings.HasPrefix(path, "/api/v2/taxonomy/profiles") {
			switch method {
			case http.MethodGet:
				return iamdomain.CapDocView, true
			case http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete:
				return iamdomain.CapTaxonomyManage, true
			}
		}
		if strings.HasPrefix(path, "/api/v2/taxonomy/areas") {
			switch method {
			case http.MethodGet:
				return iamdomain.CapDocView, true
			case http.MethodPost, http.MethodPut, http.MethodDelete:
				return iamdomain.CapTaxonomyManage, true
			}
		}
		if strings.HasPrefix(path, "/api/v2/taxonomy/families") {
			switch method {
			case http.MethodGet:
				return iamdomain.CapDocView, true
			case http.MethodPost, http.MethodPut, http.MethodDelete:
				return iamdomain.CapTaxonomyManage, true
			}
		}

		// V2 controlled documents
		if strings.HasPrefix(path, "/api/v2/controlled-documents") {
			switch {
			case method == http.MethodGet:
				return iamdomain.CapDocView, true
			case method == http.MethodPost && path == "/api/v2/controlled-documents":
				return iamdomain.CapRegistryCreate, true
			case method == http.MethodPut && strings.HasSuffix(path, "/obsolete"):
				return iamdomain.CapDocEdit, true
			case method == http.MethodPut && strings.HasSuffix(path, "/supersede"):
				return iamdomain.CapDocEdit, true
			}
		}

		// V2 area memberships
		if strings.HasPrefix(path, "/api/v2/iam/area-memberships") {
			return iamdomain.CapMembershipManage, true
		}

		// V2 signed downloads
		if method == http.MethodGet && path == "/api/v2/signed" {
			return iamdomain.CapTemplateView, true
		}

		// V2 approval routes
		if strings.HasPrefix(path, "/api/v2/approval/") {
			switch {
			case method == http.MethodGet:
				return iamdomain.CapDocView, true
			case method == http.MethodPost, method == http.MethodPut, method == http.MethodDelete:
				return iamdomain.CapDocSubmit, true
			}
		}

		return "", false
	}
}

func newPublicPathChecker(resolver iamdelivery.PermissionResolver) authdelivery.PublicPathChecker {
	return func(method, path string) bool {
		if requiresSessionButNoPermission(method, path) {
			return false
		}
		_, guarded := resolver(method, path)
		return !guarded
	}
}

func requiresSessionButNoPermission(method, path string) bool {
	if method == http.MethodGet && path == "/api/v1/auth/me" {
		return true
	}
	if method == http.MethodPost && path == "/api/v1/auth/change-password" {
		return true
	}
	if strings.HasPrefix(path, "/api/v2/") {
		return true
	}
	return false
}
```

- [ ] **Step 5: Update permissions_test.go**

```go
// apps/api/cmd/metaldocs-api/permissions_test.go
package main

import (
	"net/http"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

func TestPermissionResolver(t *testing.T) {
	t.Parallel()

	resolver := newPermissionResolver()

	testCases := []struct {
		name      string
		method    string
		path      string
		wantCap   string
		wantGuard bool
	}{
		{name: "health live unguarded", method: http.MethodGet, path: "/api/v1/health/live", wantCap: "", wantGuard: false},
		{name: "auth login unguarded", method: http.MethodPost, path: "/api/v1/auth/login", wantCap: "", wantGuard: false},
		{name: "feature flags unguarded", method: http.MethodGet, path: "/api/v1/feature-flags", wantCap: "", wantGuard: false},
		{name: "unknown endpoint unguarded", method: http.MethodGet, path: "/api/v1/unknown", wantCap: "", wantGuard: false},
		{name: "documents create", method: http.MethodPost, path: "/api/v1/documents", wantCap: iamdomain.CapDocCreate, wantGuard: true},
		{name: "documents list", method: http.MethodGet, path: "/api/v1/documents", wantCap: iamdomain.CapDocView, wantGuard: true},
		{name: "document detail", method: http.MethodGet, path: "/api/v1/documents/doc-1", wantCap: iamdomain.CapDocView, wantGuard: true},
		{name: "document browser content save", method: http.MethodPost, path: "/api/v1/documents/doc-1/content/browser", wantCap: iamdomain.CapDocEdit, wantGuard: true},
		{name: "document versions list", method: http.MethodGet, path: "/api/v1/documents/doc-1/versions", wantCap: iamdomain.CapDocView, wantGuard: true},
		{name: "document version create", method: http.MethodPost, path: "/api/v1/documents/doc-1/versions", wantCap: iamdomain.CapDocEdit, wantGuard: true},
		{name: "workflow transition", method: http.MethodPost, path: "/api/v1/workflow/documents/doc-1/transitions", wantCap: iamdomain.CapDocSubmit, wantGuard: true},
		{name: "iam users list", method: http.MethodGet, path: "/api/v1/iam/users", wantCap: iamdomain.CapUserManage, wantGuard: true},
		{name: "iam roles update", method: http.MethodPut, path: "/api/v1/iam/users/u-1/roles", wantCap: iamdomain.CapUserManage, wantGuard: true},
		{name: "template list", method: http.MethodGet, path: "/api/v1/templates", wantCap: iamdomain.CapTemplateView, wantGuard: true},
		{name: "template create", method: http.MethodPost, path: "/api/v1/templates", wantCap: iamdomain.CapTemplateView, wantGuard: true},
		{name: "template draft sub-route", method: http.MethodGet, path: "/api/v1/templates/my-key/draft", wantCap: iamdomain.CapTemplateView, wantGuard: true},
		{name: "template publish sub-route", method: http.MethodPost, path: "/api/v1/templates/my-key/publish", wantCap: iamdomain.CapTemplateView, wantGuard: true},
		{name: "v2 templates list", method: http.MethodGet, path: "/api/v2/templates", wantCap: iamdomain.CapTemplateView, wantGuard: true},
		{name: "v2 templates create", method: http.MethodPost, path: "/api/v2/templates", wantCap: iamdomain.CapTemplateEdit, wantGuard: true},
		{name: "v2 templates version draft", method: http.MethodPut, path: "/api/v2/templates/t1/versions/1/draft", wantCap: iamdomain.CapTemplateEdit, wantGuard: true},
		{name: "v2 templates publish", method: http.MethodPost, path: "/api/v2/templates/t1/versions/1/publish", wantCap: iamdomain.CapTemplatePublish, wantGuard: true},
		{name: "v2 docx-upload-url", method: http.MethodPost, path: "/api/v2/templates/t1/versions/1/docx-upload-url", wantCap: iamdomain.CapTemplateEdit, wantGuard: true},
		{name: "v2 schema-upload-url", method: http.MethodPost, path: "/api/v2/templates/t1/versions/1/schema-upload-url", wantCap: iamdomain.CapTemplateEdit, wantGuard: true},
		{name: "v2 signed download", method: http.MethodGet, path: "/api/v2/signed", wantCap: iamdomain.CapTemplateView, wantGuard: true},
		{name: "v2 doc submit", method: http.MethodPost, path: "/api/v2/documents/d1/submit", wantCap: iamdomain.CapDocSubmit, wantGuard: true},
		{name: "v2 doc signoff", method: http.MethodPost, path: "/api/v2/documents/d1/signoff", wantCap: iamdomain.CapDocSignoff, wantGuard: true},
		{name: "v2 taxonomy families list", method: http.MethodGet, path: "/api/v2/taxonomy/families", wantCap: iamdomain.CapDocView, wantGuard: true},
		{name: "v2 taxonomy families create", method: http.MethodPost, path: "/api/v2/taxonomy/families", wantCap: iamdomain.CapTaxonomyManage, wantGuard: true},
		{name: "v2 controlled-documents create", method: http.MethodPost, path: "/api/v2/controlled-documents", wantCap: iamdomain.CapRegistryCreate, wantGuard: true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotCap, gotGuard := resolver(tc.method, tc.path)
			if gotGuard != tc.wantGuard {
				t.Fatalf("guard mismatch for %s %s: got %v want %v", tc.method, tc.path, gotGuard, tc.wantGuard)
			}
			if gotCap != tc.wantCap {
				t.Fatalf("cap mismatch for %s %s: got %q want %q", tc.method, tc.path, gotCap, tc.wantCap)
			}
		})
	}
}
```

- [ ] **Step 6: Run permissions tests**

```bash
go test ./apps/api/cmd/metaldocs-api/... -run TestPermissionResolver -v
```

Expected: all test cases PASS

- [ ] **Step 7: Build IAM packages (not binary — main.go is fixed in Task 7)**

```bash
go build ./internal/modules/iam/...
go vet ./internal/modules/iam/...
```

Expected: PASS. The API binary (`apps/api/cmd/metaldocs-api/...`) will fail to compile until Task 7 updates `main.go` — that is expected and correct. Do not run `go test` on the `metaldocs-api` package here; it won't compile until `main.go` is wired in Task 7.

- [ ] **Step 8: Commit**

```bash
git add internal/modules/iam/delivery/http/middleware.go \
        internal/modules/iam/delivery/http/middleware_test.go \
        apps/api/cmd/metaldocs-api/permissions.go \
        apps/api/cmd/metaldocs-api/permissions_test.go
git commit -m "feat(iam): replace Middleware.authorizer with CapabilityService, resolver returns string"
```

---

## Task 6: Role Callers — Cascade RoleAdmin → RoleSystemAdmin

**Model:** haiku ⚡ Parallel with Tasks 3+4 (depends on Task 2 only)

**Files:**
- Modify: `internal/modules/iam/delivery/http/admin_handler.go`
- Modify: `internal/modules/iam/application/area_membership_test.go`
- Modify: `internal/modules/auth/application/service.go`
- Modify: `internal/modules/auth/infrastructure/memory/repository.go`
- Modify: `internal/modules/notifications/delivery/http/handler.go`
- Modify: `internal/platform/authn/config.go`
- Modify: `apps/api/cmd/metaldocs-e2e-seed/main.go`

- [ ] **Step 1: Update admin_handler.go role allowlists**

In `internal/modules/iam/delivery/http/admin_handler.go`, find the two switch statements that validate roles and replace them:

Find line ~322:
```go
// OLD:
case iamdomain.RoleAdmin, iamdomain.RoleEditor, iamdomain.RoleReviewer, iamdomain.RoleViewer:
// NEW:
case iamdomain.RoleSystemAdmin, iamdomain.RoleApprover, iamdomain.RoleAuthor, iamdomain.RoleEditor, iamdomain.RoleViewer:
```

Find line ~454 (in `parseRoles`):
```go
// OLD:
case iamdomain.RoleAdmin, iamdomain.RoleEditor, iamdomain.RoleReviewer, iamdomain.RoleViewer:
// NEW:
case iamdomain.RoleSystemAdmin, iamdomain.RoleApprover, iamdomain.RoleAuthor, iamdomain.RoleEditor, iamdomain.RoleViewer:
```

- [ ] **Step 2: Update area_membership_test.go**

In `internal/modules/iam/application/area_membership_test.go`, replace `domain.RoleReviewer` with `domain.RoleApprover` (2 occurrences):

```go
// OLD:
err := service.Grant(context.Background(), "u1", "t1", "A1", domain.RoleReviewer, "admin")
// ...
if repo.atomicNew.Role != domain.RoleReviewer {

// NEW:
err := service.Grant(context.Background(), "u1", "t1", "A1", domain.RoleApprover, "admin")
// ...
if repo.atomicNew.Role != domain.RoleApprover {
```

- [ ] **Step 3: Update auth/application/service.go**

In `internal/modules/auth/application/service.go`:
- Line ~63: `HasAnyRole(ctx, iamdomain.RoleAdmin)` → `HasAnyRole(ctx, iamdomain.RoleSystemAdmin)`
- Line ~92: `iamdomain.RoleAdmin,` → `iamdomain.RoleSystemAdmin,`

- [ ] **Step 4: Update auth/infrastructure/memory/repository.go**

In `internal/modules/auth/infrastructure/memory/repository.go`:
- Line ~289: `if role == iamdomain.RoleAdmin {` → `if role == iamdomain.RoleSystemAdmin {`
- Line ~308: `Roles: []iamdomain.Role{iamdomain.RoleAdmin},` → `Roles: []iamdomain.Role{iamdomain.RoleSystemAdmin},`

- [ ] **Step 5: Update notifications handler**

In `internal/modules/notifications/delivery/http/handler.go`:
```go
// OLD:
if role == iamdomain.RoleAdmin {
// NEW:
if role == iamdomain.RoleSystemAdmin {
```

- [ ] **Step 6: Update authn/config.go**

In `internal/platform/authn/config.go`:
```go
// Line ~130 OLD:
return map[string][]iamdomain.Role{"admin-local": {iamdomain.RoleAdmin}}
// NEW:
return map[string][]iamdomain.Role{"admin-local": {iamdomain.RoleSystemAdmin}}

// Line ~163 OLD:
out["admin-local"] = []iamdomain.Role{iamdomain.RoleAdmin}
// NEW:
out["admin-local"] = []iamdomain.Role{iamdomain.RoleSystemAdmin}
```

- [ ] **Step 7: Update e2e-seed/main.go**

In `apps/api/cmd/metaldocs-e2e-seed/main.go`:
```go
// OLD (2 occurrences):
iamdomain.RoleAdmin
// NEW:
iamdomain.RoleSystemAdmin
```

- [ ] **Step 8: Build all changed packages**

```bash
go build ./internal/modules/iam/...
go build ./internal/modules/auth/...
go build ./internal/modules/notifications/...
go build ./internal/platform/authn/...
go build ./apps/api/cmd/metaldocs-e2e-seed/...
```

Expected: PASS for all

- [ ] **Step 9: Run unit tests in modified packages**

```bash
go test ./internal/modules/iam/application/...
go test ./internal/modules/auth/application/...
```

Expected: PASS

- [ ] **Step 10: Commit**

```bash
git add internal/modules/iam/delivery/http/admin_handler.go \
        internal/modules/iam/application/area_membership_test.go \
        internal/modules/auth/application/service.go \
        internal/modules/auth/infrastructure/memory/repository.go \
        internal/modules/notifications/delivery/http/handler.go \
        internal/platform/authn/config.go \
        apps/api/cmd/metaldocs-e2e-seed/main.go
git commit -m "feat(iam): migrate callers from RoleAdmin/RoleReviewer to RoleSystemAdmin/RoleApprover"
```

---

## Task 7: main.go Wiring

**Model:** sonnet (depends on Tasks 3+5+6)

**Files:**
- Modify: `apps/api/cmd/metaldocs-api/main.go`

- [ ] **Step 1: Write failing build test**

```bash
go build ./apps/api/cmd/metaldocs-api/...
```

Expected: FAIL — `NewMiddleware` arg mismatch (still passes `authorizer`)

- [ ] **Step 2: Update main.go**

Find the wiring block around lines 153-163 in `apps/api/cmd/metaldocs-api/main.go`. Replace:

```go
// OLD:
authorizer := iamapp.NewStaticAuthorizer()
cachedProvider := iamapp.NewCachedRoleProvider(deps.RoleProvider, authn.CacheTTL())
// ...
iamMiddleware := iamdelivery.NewMiddleware(authorizer, cachedProvider, authn.Enabled(), authCfg.LegacyHeaderEnabled).

// NEW:
capService := iamapp.NewCapabilityService(deps.SQLDB)
cachedProvider := iamapp.NewCachedRoleProvider(deps.RoleProvider, authn.CacheTTL())
// ...
iamMiddleware := iamdelivery.NewMiddleware(capService, cachedProvider, authn.Enabled(), authCfg.LegacyHeaderEnabled).
```

Also remove the import of `authorizer` if it becomes unused. The `iamapp` import remains (used for other things like `NewAreaMembershipService`, `NewAdminService`).

- [ ] **Step 3: Build binary**

```bash
go build ./apps/api/cmd/metaldocs-api/...
```

Expected: PASS — binary compiles successfully

- [ ] **Step 4: Start API and verify login works**

```powershell
.\scripts\start-api.ps1 -Build
```

In another terminal:
```powershell
$r = Invoke-RestMethod -Method POST -Uri "http://localhost:8081/api/v1/auth/login" `
     -ContentType "application/json" `
     -Body '{"identifier":"admin","password":"AdminMetalDocs123!"}'
$token = $r.token
Write-Host "Login OK, token: $($token.Substring(0,10))..."
```

Expected: HTTP 200, token returned

- [ ] **Step 5: Verify admin can hit a guarded endpoint**

```powershell
$headers = @{ Authorization = "Bearer $token"; "X-Tenant-ID" = "ffffffff-ffff-ffff-ffff-ffffffffffff" }
Invoke-RestMethod -Uri "http://localhost:8081/api/v1/iam/users" -Headers $headers
```

Expected: HTTP 200, user list (not 403)

- [ ] **Step 6: Commit**

```bash
git add apps/api/cmd/metaldocs-api/main.go
git commit -m "feat(iam): wire CapabilityService into API — replace StaticAuthorizer"
```

---

## Task 8: Cleanup Dead Code

**Model:** haiku (depends on Task 7 — all callers updated)

**Files:**
- Delete: `internal/modules/iam/application/authorizer.go`
- Delete: `internal/modules/iam/application/authorizer_test.go`
- Modify: `internal/modules/iam/domain/model.go` (remove dead constants)
- Modify: `internal/modules/iam/domain/role_capabilities.go` (remove old roles)
- Modify: `internal/modules/iam/domain/port.go` (remove Authorizer interface)

- [ ] **Step 1: Verify no remaining callers of dead symbols**

```bash
grep -r "NewStaticAuthorizer\|StaticAuthorizer\|RoleAdmin\b\|RoleReviewer\b\|iamdomain\.Perm\|domain\.Perm" \
     internal/ apps/ tests/ --include="*.go" | grep -v "vendor"
```

Expected: zero matches (all callers were updated in Tasks 5+6+7)

If any matches remain, fix them before proceeding.

- [ ] **Step 2: Delete authorizer files**

```bash
rm internal/modules/iam/application/authorizer.go
rm internal/modules/iam/application/authorizer_test.go
```

- [ ] **Step 3: Clean model.go — remove dead symbols**

In `internal/modules/iam/domain/model.go`, remove:
1. `RoleAdmin Role = "admin"` line
2. `RoleReviewer Role = "reviewer"` line
3. The entire `Permission` type declaration: `type Permission string`
4. All `Perm*` constants block (13 constants from `PermDocumentCreate` to `PermTemplateExport`)
5. The old `Capability` type block (8 constants: `CapDocumentView`, `CapDocumentCreate`, `CapDocumentEdit`, `CapTemplateView`, `CapTemplatePublish`, `CapWorkflowReview`, `CapWorkflowApprove`, `CapRegistryCreate`)

After cleanup, `model.go` should contain ONLY:
```go
package domain

type Role string

const (
	RoleApprover    Role = "approver"
	RoleAuthor      Role = "author"
	RoleEditor      Role = "editor"
	RoleSystemAdmin Role = "system_admin"
	RoleViewer      Role = "viewer"
)
```

- [ ] **Step 4: Clean role_capabilities.go — remove old roles**

Replace `internal/modules/iam/domain/role_capabilities.go` with:

```go
package domain

const RoleCapabilitiesVersion = 2

// RoleCapabilities maps roles to their in-memory capability set.
// The authoritative capability matrix lives in the DB (role_capabilities table).
// This map is used by AuthorizationService for area-scoped legacy checks.
var RoleCapabilities = map[Role][]Capability{
	RoleViewer: {
		CapDocumentView,
		CapTemplateView,
	},
	RoleEditor: {
		CapDocumentView,
		CapDocumentCreate,
		CapDocumentEdit,
		CapTemplateView,
	},
	RoleApprover: {
		CapDocumentView,
		CapWorkflowApprove,
		CapTemplateView,
		CapTemplatePublish,
	},
	RoleAuthor: {
		CapDocumentView,
		CapDocumentCreate,
		CapDocumentEdit,
	},
	RoleSystemAdmin: {
		CapDocumentView,
		CapDocumentCreate,
		CapDocumentEdit,
		CapTemplateView,
		CapTemplatePublish,
		CapWorkflowApprove,
		CapRegistryCreate,
	},
}
```

Note: `RoleCapabilities` still uses old `Capability` type constants (`CapDocumentView` etc.) from the OLD `Capability` block. But we just deleted those in Step 3. We need to either:
- Keep the old `Capability` type + constants (don't delete them in Step 3), OR
- Use the new string constants instead

Decision: The `AuthorizationService` uses `domain.RoleCapabilities` with `domain.Capability` type. These are separate from the new `Cap*` string constants. Keep the old `Capability` type for `AuthorizationService` backward compatibility. Only remove `Permission` type + `Perm*` and the old role constants.

Revised Step 3 — keep Capability type, only remove:
1. `RoleAdmin Role = "admin"`
2. `RoleReviewer Role = "reviewer"`
3. `type Permission string` + all `Perm*` constants

After cleanup, `model.go` contains:
```go
package domain

type Role string

const (
	RoleApprover    Role = "approver"
	RoleAuthor      Role = "author"
	RoleEditor      Role = "editor"
	RoleSystemAdmin Role = "system_admin"
	RoleViewer      Role = "viewer"
)

type Capability string

const (
	CapDocumentView    Capability = "document.view"
	CapDocumentCreate  Capability = "document.create"
	CapDocumentEdit    Capability = "document.edit"
	CapTemplateView    Capability = "template.view"
	CapTemplatePublish Capability = "template.publish"
	CapWorkflowReview  Capability = "workflow.review"
	CapWorkflowApprove Capability = "workflow.approve"
	CapRegistryCreate  Capability = "registry.create"
)
```

- [ ] **Step 5: Remove Authorizer interface from port.go**

In `internal/modules/iam/domain/port.go`, remove the `Authorizer` interface:

```go
// DELETE these 4 lines:
// Authorizer is the stable IAM contract for authorization decisions.
type Authorizer interface {
    Can(role Role, permission Permission) bool
}
```

After cleanup `port.go` contains only `RoleProvider` and `RoleAdminRepository`.

- [ ] **Step 6: Build and test**

```bash
go build ./...
go test ./internal/modules/iam/... ./internal/modules/auth/... ./apps/api/...
```

Expected: all PASS (no references to deleted symbols)

- [ ] **Step 7: Start API one more time**

```powershell
.\scripts\start-api.ps1 -Build
```

Login and verify guarded endpoint still returns 200.

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "chore(iam): remove StaticAuthorizer, Perm* constants, Authorizer interface — cleanup"
```

---

## Task 9: Wiki

**Model:** wiki-curator agent (after Task 8)

Dispatch the wiki-curator agent with this prompt:

> IAM/RBAC unification just landed on main. Changes since last wiki update:
> - 5 new migrations (0162-0166): tenant_id on iam_user_roles, groups tables, documents_v2.visibility, role_capabilities reseed, role rename
> - New domain constants: RoleSystemAdmin, RoleAuthor, Cap* (16 capability strings) in `internal/modules/iam/domain/`
> - New `CapabilityService` in `internal/modules/iam/application/capability_service.go`
> - `authz.Require` now has system_admin bypass in `internal/modules/iam/authz/authz.go`
> - `Middleware` replaced `StaticAuthorizer` with `CapabilityService` in `internal/modules/iam/delivery/http/middleware.go`
> - `PermissionResolver` now returns `(string, bool)` — capability strings replace Permission type
> - Roles renamed: admin→system_admin, reviewer→approver. `author` role added.
> - Deleted: `application/authorizer.go`, `Perm*` constants, `Authorizer` interface
>
> Update `wiki/modules/iam-rbac.md` (Key files, role list, capability matrix, auth flow description). Refresh Last verified stamps on any wiki doc referencing IAM/auth. Update `wiki/README.md` index if new sections are added.

---

## Note: Document Visibility Enforcement (Phase 1b — Separate Task)

The visibility column is created by migration 0164 and defaults to `'area'`. However, the actual query filtering (blocking users from seeing documents they can't access) is NOT in this plan. It requires modifying document list/detail queries to join against `public.user_process_areas`.

The spec provides the SQL for this check:

```sql
SELECT CASE d.visibility
  WHEN 'public' THEN true
  WHEN 'area'   THEN (
    EXISTS (SELECT 1 FROM metaldocs.iam_user_roles WHERE user_id = $actor_id AND role_code = 'system_admin')
    OR
    EXISTS (SELECT 1 FROM public.user_process_areas upa
             WHERE upa.user_id = $actor_id AND upa.area_code = d.process_area_code AND upa.effective_to IS NULL)
  )
  ELSE false  -- restricted: denied until Phase 2
END
FROM public.documents_v2 d WHERE d.id = $doc_id
```

This enforcement is a Phase 1 requirement per spec but is deferred to a follow-up plan targeting `documents_v2` repository layer. The critical path (bootstrap catch-22 fix) does not depend on it.

---

## Phase Review Gate (After Task 8, Before Wiki)

**Model:** Opus

Before dispatching the wiki agent, run a final Opus review:

> Review the IAM/RBAC unification implementation. Spec: `docs/superpowers/specs/2026-05-02-iam-rbac-unification.md`. Check:
> 1. All spec requirements have corresponding code
> 2. No dead imports or symbol references left
> 3. `admin-local` can log in and hit guarded endpoints (system_admin bypass works)
> 4. Migration 0166 correctly renamed admin→system_admin in DB
> 5. `CapabilityService.CanDo` combined query is correct SQL
> 6. Documents visibility column exists with correct constraint
> 7. Groups tables created with correct schema

---

## Running Smoke Tests

After Task 7 (before cleanup), run the existing smoke tests to verify the bootstrap is fixed:

```powershell
# Start API
.\scripts\start-api.ps1 -Build

# Run Routine A (Bootstrap + admin login)
# Expected: Bootstrap PASS, admin-local can hit /api/v2/iam/area-memberships without 403
```

The catch-22 fix: `admin-local` has `role_code = 'system_admin'` (after migration 0166). `authz.Require` now checks for `system_admin` before the area check → bypass → no more 403.

---

## Invariants to Verify

- `doc.submit` and `doc.signoff` in same document by same actor: still blocked at domain layer (ISO SoD, unrelated to this plan)
- Document visibility defaults to `area` — not public
- `system_admin` bypasses both `CapabilityService.CanDo` and `authz.Require`
- One user = one direct role (enforced by `iam_user_roles` CHECK constraint after 0166)
