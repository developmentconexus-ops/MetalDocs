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

**ISO segregation note:** `doc.submit` and `doc.signoff` are held by different roles intentionally. The existing domain-layer SoD enforcement (submitter ≠ signoff actor on same document) is unchanged.

### Document visibility

Set once at document creation. Controls who can see the document at all.

| Value | Who can see |
|---|---|
| `public` | All authenticated users in the tenant |
| `area` | Users whose area assignment matches the document's area (default) |
| `restricted` | Only explicitly named users or groups (Phase 2+) |

Visibility is stored as a column on `documents_v2`. Existing documents default to `area`.

**Phase 1 enforcement:** `public` and `area` are enforced. `restricted` is stored but treated as `area` until Phase 2 implements explicit user/group lists.

**`area` enforcement in Phase 1:** A user can see an `area` document if they have any active row in `user_process_areas` for that area OR if they are `system_admin`. Since `user_process_areas` is currently empty for most users, admins must grant area memberships (via the existing membership API) for users to access area-restricted documents. This is intentional — documents should not be globally visible by default.

**Visibility check is separate from capability check.** A user can have `doc.view` capability but still be blocked from a specific document by its visibility setting.

### Groups (data model only — UI in Phase 2)

Groups are tenant-scoped collections of users that share a role assignment. Adding a user to a group grants them the group's role automatically.

```
iam_groups        id, tenant_id, name, description
iam_group_members group_id, user_id, tenant_id
iam_group_roles   group_id, role
```

A user's effective capabilities = union of direct role capabilities + all group role capabilities. Groups only add capabilities, never restrict. A viewer assigned directly who belongs to an "editors" group gets editor capabilities.

---

## Data Model Changes

### New tables

```sql
CREATE TABLE metaldocs.iam_groups (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   UUID NOT NULL DEFAULT 'ffffffff-ffff-ffff-ffff-ffffffffffff',
  name        TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, name)
);

CREATE TABLE metaldocs.iam_group_members (
  group_id   UUID NOT NULL REFERENCES metaldocs.iam_groups(id) ON DELETE CASCADE,
  user_id    TEXT NOT NULL,
  tenant_id  UUID NOT NULL,
  granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  granted_by TEXT,
  PRIMARY KEY (group_id, user_id)
);

CREATE TABLE metaldocs.iam_group_roles (
  group_id UUID NOT NULL REFERENCES metaldocs.iam_groups(id) ON DELETE CASCADE,
  role     TEXT NOT NULL,
  PRIMARY KEY (group_id, role)
);
```

### Modified tables

```sql
-- Document visibility
ALTER TABLE public.documents_v2
  ADD COLUMN visibility TEXT NOT NULL DEFAULT 'area'
  CHECK (visibility IN ('public', 'area', 'restricted'));

-- Existing rows get 'area' via DEFAULT
```

### Seeded data (migration)

Truncate and re-seed `metaldocs.role_capabilities` with the capability matrix above.
Re-seed `metaldocs.role_capabilities` roles: `viewer`, `editor`, `author`, `approver`, `system_admin`.

Rename existing `admin` role in `iam_user_roles` to `system_admin` for `admin-local`.

---

## Go Code Changes

### Deleted

- `internal/modules/iam/application/authorizer.go` — `StaticAuthorizer`, `NewStaticAuthorizer`
- `internal/modules/iam/domain/model.go` — `Permission` type and all `Perm*` constants
- Static permission map in `internal/modules/iam/delivery/http/middleware.go`

### New: unified authz service

**File:** `internal/modules/iam/application/capability_service.go`

```go
type CapabilityService struct { db *sql.DB }

// CanDo returns nil if userID has capability in tenantID, ErrCapabilityDenied otherwise.
// system_admin bypasses all checks.
func (s *CapabilityService) CanDo(ctx context.Context, userID, tenantID, capability string) error
```

Query:
```sql
SELECT EXISTS (
  SELECT 1 FROM metaldocs.iam_user_roles ur
  JOIN metaldocs.role_capabilities rc ON rc.role = ur.role_code
  WHERE ur.user_id   = $1
    AND ur.tenant_id = $2
    AND (ur.role_code = 'system_admin' OR rc.capability = $3)
)
```

Groups path (also checked):
```sql
SELECT EXISTS (
  SELECT 1 FROM metaldocs.iam_group_members gm
  JOIN metaldocs.iam_group_roles gr ON gr.group_id = gm.group_id
  JOIN metaldocs.role_capabilities rc ON rc.role = gr.role
  WHERE gm.user_id  = $1
    AND gm.tenant_id = $2
    AND rc.capability = $3
)
```

### Updated: `authz.Require`

**File:** `internal/modules/iam/authz/authz.go`

Add system_admin bypass at the top of `Require`:

```go
// Short-circuit for system_admin — bypass area + capability check entirely.
var isAdmin bool
_ = tx.QueryRowContext(ctx, `
  SELECT EXISTS (
    SELECT 1 FROM metaldocs.iam_user_roles
    WHERE user_id    = current_setting('metaldocs.actor_id', false)
      AND tenant_id  = current_setting('metaldocs.tenant_id', false)::uuid
      AND role_code  = 'system_admin'
  )
`).Scan(&isAdmin)
if isAdmin {
  return appendAssertedCap(ctx, tx, capability, areaCode)
}
```

### Updated: IAM middleware

Replace `StaticAuthorizer.Can(role, permission)` call with `CapabilityService.CanDo(userID, tenantID, capability)`. Map old `Permission` constants to new `Capability` strings in `permissions.go`.

### Updated: `permissions.go`

Replace `iamdomain.PermDocumentCreate` etc. with string capability constants:
```go
const (
  CapDocView        = "doc.view"
  CapDocCreate      = "doc.create"
  CapDocEdit        = "doc.edit"
  CapDocSubmit      = "doc.submit"
  CapDocSignoff     = "doc.signoff"
  CapTemplateView   = "template.view"
  CapTemplateCreate = "template.create"
  CapTemplateEdit   = "template.edit"
  CapTemplateSubmit = "template.submit"
  CapTemplateApprove = "template.approve"
  CapTemplatePublish = "template.publish"
  CapRegistryCreate = "registry.create"
  CapTaxonomyManage = "taxonomy.manage"
  CapMembershipManage = "membership.manage"
  CapRouteManage    = "route.manage"
  CapUserManage     = "user.manage"
)
```

---

## Migrations

| Migration | Description |
|---|---|
| `0162` | Create `iam_groups`, `iam_group_members`, `iam_group_roles` tables |
| `0163` | Add `visibility` column to `documents_v2` (default `'area'`) |
| `0164` | Truncate + re-seed `role_capabilities` with correct matrix |
| `0165` | Rename `admin` → `system_admin` in `iam_user_roles` for existing rows |

All idempotent. Safe to re-run.

---

## Invariants

- One user = one direct role. Group membership can add capabilities but not override direct role.
- `doc.submit` and `doc.signoff` on same document by same user: blocked at domain layer (ISO SoD, unchanged).
- `system_admin` bypass applies to capability checks only — SoD still enforced for system_admin.
- Document visibility defaults to `area` — documents are NOT public by default.
- `user_process_areas` table is preserved but not used by this system. Phase 3 will layer area-scoped roles on top without breaking this design.

---

## What Is NOT Changed

- Authentication (sessions, login, password) — unchanged
- `user_process_areas` table — preserved for Phase 3
- Approval routing logic — unchanged (routes still tied to profile/area)
- ISO SoD enforcement in domain layer — unchanged
- Frontend routing / page access control — updated to use new capability strings, logic unchanged

---

## Phase Roadmap

| Phase | Scope | Spec |
|---|---|---|
| **1 (this spec)** | Unified IAM backend + document visibility field | This document |
| **2** | Groups UI: create groups, add members, assign roles | Separate spec |
| **3** | Area-scoped roles: scope a role to specific areas via `user_process_areas` | Separate spec |
