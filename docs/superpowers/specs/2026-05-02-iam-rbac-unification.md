# IAM / RBAC Unification Design

> **Status:** Approved for implementation
> **Date:** 2026-05-02
> **Author:** leandrotcawork + Claude
> **Scope:** Replace two parallel authorization systems with one unified, DB-driven IAM layer. Add document-level visibility. Lay groundwork for groups and future area-scoped roles.
> **Out of scope:** Groups UI (Phase 2), area-scoped roles (Phase 3), authentication mechanism (sessions/login unchanged).

---

## Problem

MetalDocs grew two separate authorization systems:

**System 1 — IAM Middleware (legacy)**
- Static Go map: `StaticAuthorizer` in `internal/modules/iam/application/authorizer.go`
- Global roles (`admin`, `editor`, etc.) from `metaldocs.iam_user_roles`
- Permissions checked at the HTTP middleware layer
- Permission names: `document.view`, `iam:manage_roles`, etc.

**System 2 — `authz.Require` (newer)**
- DB-driven: `metaldocs.role_capabilities` JOIN `metaldocs.user_process_areas`
- Area-scoped roles (`qms_admin`, `reviewer`, etc.) from `public.user_process_areas`
- Capability checked inside service methods
- Capability names: `route.admin`, `doc.submit`, `membership.grant`, etc.

**Consequences:**
- Two role sets with overlapping names but different semantics
- `admin-local` is global admin (System 1) but blocked by area checks (System 2) → catch-22
- `role_capabilities` DB table is incomplete/stale
- `user_process_areas` seeded with hardcoded area codes that may not exist
- No document-level visibility control

---

## Decision

Replace both systems with one unified IAM layer. Single source of truth. One capability check function called everywhere.

---

## Architecture

### Two-factor access model

Every protected action answers two independent questions:

```
1. CAN YOU DO THIS ACTION?
   → User's global role has the required capability

2. CAN YOU SEE THIS DOCUMENT?
   → Document's visibility setting permits this user
```

`system_admin` bypasses both checks unconditionally.

### Role hierarchy

```
system_admin   Global. Transcends all areas. Manages taxonomy, users,
               approval routes, memberships. Bypasses all capability checks.

approver       Can create, edit, submit, and sign off documents.
               Can approve template versions.

author         Can create, edit, submit documents and templates.
               Primary document lifecycle role.

editor         Can edit documents and templates. Cannot submit.
               Collaboration / review-without-commitment role.

viewer         Read-only. Can view documents and templates.
```

**`reviewer` role is retired.** Existing `reviewer` users are migrated to `approver` (see migration 0166). `reviewer` capabilities (doc.view, doc.edit, workflow.review) are a subset of `approver` — no capability is lost.

One user has exactly one global role (stored in `metaldocs.iam_user_roles`).

### Capability matrix

| Capability | viewer | editor | author | approver | system_admin |
|---|:---:|:---:|:---:|:---:|:---:|
| `doc.view` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `doc.create` | | ✓ | ✓ | ✓ | ✓ |
| `doc.edit` | | ✓ | ✓ | ✓ | ✓ |
| `doc.submit` | | | ✓ | ✓ | ✓ |
| `doc.signoff` | | | | ✓ | ✓ |
| `template.view` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `template.create` | | | ✓ | | ✓ |
| `template.edit` | | ✓ | ✓ | | ✓ |
| `template.submit` | | | ✓ | | ✓ |
| `template.approve` | | | | ✓ | ✓ |
| `template.publish` | | | | | ✓ |
| `registry.create` | | ✓ | ✓ | | ✓ |
| `taxonomy.manage` | | | | | ✓ |
| `membership.manage` | | | | | ✓ |
| `route.manage` | | | | | ✓ |
| `user.manage` | | | | | ✓ |

**ISO segregation note:** `doc.submit` and `doc.signoff` are held by different roles intentionally. The existing domain-layer SoD enforcement (submitter ≠ signoff actor on same document) is unchanged. `system_admin` can hold both capabilities but is still blocked by SoD at the domain layer.

### Document visibility

Set once at document creation. Controls who can see the document at all.

| Value | Who can see | Phase 1 enforcement |
|---|---|---|
| `public` | All authenticated users in the tenant | ✓ enforced |
| `area` | Users with any active row in `user_process_areas` for the document's area (default) | ✓ enforced via `user_process_areas` |
| `restricted` | Only explicitly named users or groups | stored, treated as `area` until Phase 2 |

Visibility is stored as a column on `documents_v2`. Existing documents default to `area`.

**`area` enforcement in Phase 1** reads `public.user_process_areas` directly — the same table used by Phase 3 for area-scoped roles. It is NOT unused in Phase 1. Admins must grant users area memberships (via the existing membership API, `POST /api/v2/iam/area-memberships`) for users to see area-restricted documents. `system_admin` always bypasses visibility checks.

**Visibility SQL for document list / detail access check:**

```sql
-- Returns true if actor can see the document
SELECT CASE d.visibility
  WHEN 'public' THEN true
  WHEN 'area'   THEN (
    -- system_admin bypass
    EXISTS (
      SELECT 1 FROM metaldocs.iam_user_roles
       WHERE user_id = $actor_id AND role_code = 'system_admin'
    )
    OR
    -- has any active membership in the document's area
    EXISTS (
      SELECT 1 FROM public.user_process_areas upa
       WHERE upa.user_id  = $actor_id
         AND upa.area_code = d.process_area_code
         AND upa.effective_to IS NULL
    )
  )
  ELSE false  -- 'restricted': denied until Phase 2
END
FROM public.documents_v2 d
WHERE d.id = $doc_id
```

Add index: `CREATE INDEX ix_upa_user_area ON public.user_process_areas(user_id, area_code) WHERE effective_to IS NULL;`

**Visibility check is separate from capability check.** A user can have `doc.view` capability but still be blocked from a specific document by its visibility setting.

### Groups (data model only — UI in Phase 2)

Groups are tenant-scoped collections of users that share a role. Adding a user to a group grants them the group's role capabilities in addition to their direct role.

```
iam_groups        id, tenant_id, name, description
iam_group_members group_id, user_id, tenant_id
iam_group_roles   group_id, role
```

A user's effective capabilities = union of direct role capabilities + all group role capabilities. Groups only ADD capabilities, never restrict. No FK from `iam_group_members.user_id` to `iam_users` — intentional, keeps the constraint light and avoids cascade complexity. Both tables live in `metaldocs` schema.

---

## Data Model Changes

### Migration 0162 — Add `tenant_id` to `iam_user_roles`

`metaldocs.iam_user_roles` currently has no `tenant_id` column. Required for multi-tenant-ready design and for the `CanDo` query. This migration must run before 0163–0166.

```sql
ALTER TABLE metaldocs.iam_user_roles
  ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL
    DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff';

-- Backfill existing rows (all are dev tenant)
UPDATE metaldocs.iam_user_roles
   SET tenant_id = 'ffffffff-ffff-ffff-ffff-ffffffffffff'
 WHERE tenant_id = 'ffffffff-ffff-ffff-ffff-ffffffffffff'; -- no-op but explicit
```

### Migration 0163 — Create groups tables

```sql
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

### Migration 0164 — Add `visibility` to `documents_v2`

```sql
ALTER TABLE public.documents_v2
  ADD COLUMN IF NOT EXISTS visibility TEXT NOT NULL DEFAULT 'area'
  CONSTRAINT documents_v2_visibility_check
    CHECK (visibility IN ('public', 'area', 'restricted'));
```

### Migration 0165 — Re-seed `role_capabilities`

```sql
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

### Migration 0166 — Role rename + reviewer migration

```sql
BEGIN;

-- 1. Widen CHECK constraint to allow system_admin; remove reviewer and admin
ALTER TABLE metaldocs.iam_user_roles
  DROP CONSTRAINT IF EXISTS chk_iam_user_roles_role_code;

ALTER TABLE metaldocs.iam_user_roles
  ADD CONSTRAINT chk_iam_user_roles_role_code
    CHECK (role_code IN ('system_admin', 'approver', 'author', 'editor', 'viewer'));

-- 2. Rename admin → system_admin
UPDATE metaldocs.iam_user_roles SET role_code = 'system_admin' WHERE role_code = 'admin';

-- 3. Migrate reviewer → approver (reviewer capabilities ⊆ approver)
UPDATE metaldocs.iam_user_roles SET role_code = 'approver' WHERE role_code = 'reviewer';

COMMIT;
```

---

## Go Code Changes

### Deleted files / symbols

| File | What is removed |
|---|---|
| `internal/modules/iam/application/authorizer.go` | `StaticAuthorizer`, `NewStaticAuthorizer` |
| `internal/modules/iam/application/authorizer_test.go` | entire file (5 tests for StaticAuthorizer) |
| `internal/modules/iam/domain/model.go` | `Permission` type + all `Perm*` constants (keep `Role`, `Capability`, `UserProcessArea`) |

### New: unified capability service

**File:** `internal/modules/iam/application/capability_service.go`

```go
type CapabilityService struct { db *sql.DB }

func NewCapabilityService(db *sql.DB) *CapabilityService

// CanDo returns nil if userID has capability, ErrCapabilityDenied otherwise.
// system_admin role bypasses capability check unconditionally.
func (s *CapabilityService) CanDo(ctx context.Context, userID, tenantID, capability string) error
```

**Direct role query:**
```sql
SELECT EXISTS (
  SELECT 1 FROM metaldocs.iam_user_roles ur
  JOIN metaldocs.role_capabilities rc ON rc.role = ur.role_code
  WHERE ur.user_id   = $1
    AND ur.tenant_id = $2
    AND (ur.role_code = 'system_admin' OR rc.capability = $3)
)
```

**Group membership query (OR'd with above):**
```sql
SELECT EXISTS (
  SELECT 1 FROM metaldocs.iam_group_members gm
  JOIN metaldocs.iam_group_roles gr ON gr.group_id = gm.group_id
  JOIN metaldocs.role_capabilities rc ON rc.role = gr.role
  WHERE gm.user_id   = $1
    AND gm.tenant_id = $2
    AND rc.capability = $3
)
```

If either query returns true → allowed. If both false → `ErrCapabilityDenied`.

### Updated: IAM Middleware

**File:** `internal/modules/iam/delivery/http/middleware.go`

Remove the `Authorizer` interface field from `Middleware` struct. Replace with `*CapabilityService`. Update `NewMiddleware` signature accordingly.

Old:
```go
type Middleware struct {
    authorizer iamdomain.Authorizer
    ...
}
func NewMiddleware(authorizer iamdomain.Authorizer, ...) *Middleware
```

New:
```go
type Middleware struct {
    caps *application.CapabilityService
    ...
}
func NewMiddleware(caps *application.CapabilityService, ...) *Middleware
```

Replace `hasPermission(authorizer, roles, permission)` call with:
```go
if err := caps.CanDo(r.Context(), userID, tenantID, capability); err != nil {
    writeAPIError(w, http.StatusForbidden, ...)
    return
}
```

**File:** `internal/modules/iam/delivery/http/middleware_test.go`

Update test setup: replace `iamapp.NewStaticAuthorizer()` with `iamapp.NewCapabilityService(testDB)` or a mock. All 5 test functions need updating.

### Updated: `permissions.go`

**File:** `apps/api/cmd/metaldocs-api/permissions.go`

Replace all `iamdomain.Perm*` constants with string capability constants. Define them in a new file `internal/modules/iam/domain/capabilities.go`:

```go
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

### Updated: `main.go` wiring

**File:** `apps/api/cmd/metaldocs-api/main.go`

Replace:
```go
authorizer := iamapp.NewStaticAuthorizer()
...
iamMiddleware := iamdelivery.NewMiddleware(authorizer, cachedProvider, ...)
```

With:
```go
capService := iamapp.NewCapabilityService(deps.SQLDB)
...
iamMiddleware := iamdelivery.NewMiddleware(capService, cachedProvider, ...)
```

### Updated: `admin_handler.go` role allowlist

**File:** `internal/modules/iam/delivery/http/admin_handler.go`

Remove `reviewer` and `admin` from the role allowlist. New allowed values: `system_admin`, `approver`, `author`, `editor`, `viewer`.

### Updated: `authz.Require` — system_admin bypass

**File:** `internal/modules/iam/authz/authz.go`

Insert at top of `Require` function, BEFORE the existing capability query. Propagate error instead of discarding:

```go
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
    return appendAssertedCap(ctx, tx, capability, areaCode)
}
```

---

## Migration Order

| Migration | Description | Dependency |
|---|---|---|
| `0162` | Add `tenant_id` to `metaldocs.iam_user_roles` | Must run before 0163–0166 |
| `0163` | Create `iam_groups`, `iam_group_members`, `iam_group_roles` | After 0162 |
| `0164` | Add `visibility` column to `documents_v2` | None |
| `0165` | Truncate + re-seed `role_capabilities` | After 0162 |
| `0166` | Widen CHECK constraint, rename `admin` → `system_admin`, migrate `reviewer` → `approver` | After 0162 |

All idempotent. Safe to re-run via `dev-migrate.ps1`.

---

## Invariants

- One user = one direct role. Group membership adds capabilities but cannot exceed `system_admin`.
- `doc.submit` and `doc.signoff` on same document by same user: blocked at domain layer (ISO SoD, unchanged). Applies to `system_admin` too.
- Document visibility defaults to `area` — documents are NOT public by default.
- `user_process_areas` IS used for `area` visibility enforcement in Phase 1. It is NOT used for capability checks (that is `iam_user_roles` + `role_capabilities`). Phase 3 adds area-scoped capability checks on top.
- `restricted` visibility stores the value but grants no access until Phase 2.

---

## What Is NOT Changed

- Authentication (sessions, login, password) — unchanged
- `user_process_areas` table structure — unchanged
- Approval routing logic — unchanged
- ISO SoD enforcement in domain layer — unchanged
- Frontend page routing / sidebar visibility — updated to use new capability strings, logic unchanged

---

## Phase Roadmap

| Phase | Scope | Spec |
|---|---|---|
| **1 (this spec)** | Unified IAM backend + document visibility field | This document |
| **2** | Groups UI: create groups, add members, assign roles | Separate spec |
| **3** | Area-scoped roles: scope capabilities to specific areas via `user_process_areas` | Separate spec |
