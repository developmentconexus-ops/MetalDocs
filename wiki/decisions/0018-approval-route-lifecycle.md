# ADR 0018 — Approval Route Lifecycle

> **Status:** accepted 2026-06-01 (drafted as the architectural baseline for the PR-4 frontend rewrite of the route admin page; codifies the lifecycle, OCC, in-use guard, capability pinning, and audit semantics already shipped by Approval Route Admin PR-2)
> **Last verified:** 2026-06-01
> **Scope:** Per-tenant approval route catalogue at `public.approval_routes` + `public.approval_route_stages`. State machine, version OCC, in-use deactivate guard, required-capability pin per stage, reason audit, deferred Tier-1 cap split.
> **Out of scope:** Per-instance runtime (stage instances, signoff state machine — see `wiki/modules/approval.md` §6). Eligibility drift policies — listed in stage config but enforced at instance runtime, owned by `domain/drift.go`. New `GET /api/v1/iam/roles` endpoint — proposed below, deferred to PR-4 or its own micro-PR.
> **Key files:**
> - `internal/modules/documents/approval/domain/route.go:39` — `Route` aggregate, `Stage` value object, `Validate`
> - `internal/modules/documents/approval/application/route_admin_service.go:343` — `Deactivate` (reason required, OCC, governance event)
> - `internal/modules/documents/approval/application/route_admin_service.go:44` — `ErrRouteDeactivateReasonRequired`
> - `internal/modules/documents/approval/http/route_admin_handler.go:79,134` — `If-Match` precondition parse on Update/Deactivate
> - `internal/modules/documents/approval/repository/errors.go:23` — `ErrRouteInUse`
> - `internal/modules/documents/approval/repository/errors.go:65` — `enforce_route_immutable` trigger error mapping (migration 0145)
> - `db/baseline/0001_current_schema.sql:1865` — `approval_routes` table (`active` boolean, `version` int)
> - `wiki/decisions/0007-two-tier-authz.md` — capability tier model that the required-cap pin participates in
> - `wiki/decisions/0016-view-grade-capabilities.md` — view-grade caps; reason for deferred Tier-1 split discussed below
> - `wiki/decisions/0017-signoff-idempotency-fingerprint.md` — client-stable fingerprint rule that route admin idempotency now follows

## Context

The approval route catalogue is the per-tenant configuration that drives every approval instance: a route is selected at submit time, its stages are snapshotted into stage instances, and the snapshot is what the runtime evaluates. PR-1 published the OpenAPI surface; PR-2 hardened the backend (idempotency persisted via `PostgresRouteAdminIdempStore`, `If-Match` precondition wired, `Reason` on deactivate required and persisted, single-tx List with `route.admin` authz). PR-4 will rewrite the frontend page to consume these contracts.

PR-4 needs an unambiguous, single-source written model of:
1. What lifecycle transitions are legal — and which are intentionally absent.
2. How concurrent edits are reconciled (OCC vs last-write-wins).
3. What blocks deactivation today, and what error code the frontend must map.
4. Where stage-level required capability is frozen, and when.
5. Where the deactivate reason lives in the audit trail.
6. Where role labels come from for the stage editor — replacing the hard-coded `STAGE_ROLES` literal at `frontend/apps/web/src/features/approval/pages/RouteAdminPage.tsx:10`.

Today this knowledge is split across `wiki/modules/approval.md` §12 (terms only), the route admin service implementation, and the schema. Codifying it here lets PR-4 reference one durable doc instead of re-reading scattered code.

## Decision

### 1. Route lifecycle is a two-state, terminal-on-deactivate machine

```
                  ┌────────────┐                    ┌──────────────┐
   create  ───►   │ active=true│  ── deactivate ──► │ active=false │  (terminal)
                  └────────────┘                    └──────────────┘
```

- **Only forward transition:** `active=true` → `active=false`. No `active=false` → `active=true` path exists in the service, the HTTP surface, or the database trigger. Adding one is out of scope.
- **No reactivate:** by design. A deactivated route is an immutable historical record. To restore behavior, **create a new route**; that new route is a new row with `id`/`version=1` and accumulates its own audit trail.
- **Update is in-place only while `active=true`.** `PUT /api/v1/approval/routes/{id}` rejects deactivated routes (handled at the trigger layer via `enforce_route_immutable` → `ErrRouteInUse` mapping; see also rationale below).
- **Stages are not separately addressable.** Stages are owned by their route. Update replaces the full stage set atomically; deactivate snapshots the existing stages in place.

### 2. Version is monotonic OCC; clients pin via `If-Match`

- `approval_routes.version` starts at 1 (Create) and increments on every successful Update and Deactivate.
- `PUT` and `DELETE` (deactivate) require `If-Match: "<version>"`. The header is parsed at `route_admin_handler.go:79,134`.
- The service writes the row with `WHERE id = $1 AND version = $expected`; zero rows affected → `repository.ErrStaleRevision` → `409 state.stale_revision` per the approval error mapper.
- Frontend must surface the conflict and refetch; auto-merge is **not** offered.
- A missing or malformed `If-Match` is rejected at the handler boundary before any DB work.

### 3. Deactivate is blocked while any instance still references the route

- Trigger `enforce_route_immutable` (migration 0145) raises `P0001` with prefix `ErrRouteInUse:` when an attempt to deactivate (or otherwise mutate a route's instance-critical columns) would orphan in-flight signoffs.
- `repository.MapPgError` at `repository/errors.go:65-71` matches the prefix and returns `repository.ErrRouteInUse`.
- HTTP layer maps to `409 route.in_use` via `approvalCodeRouteInUse` (`http/errors.go:34,102`).
- "In use" means: at least one `approval_instances` row references this route in a non-terminal state. Terminal instances do **not** block — they are historical and pin their own snapshot independently.
- Frontend remediation: the route admin page must surface this 409 with a copy that names what to do ("Wait for in-flight approvals to complete, or cancel them, before deactivating."). PR-4 owns the copy.

### 4. Required capability pin per stage is frozen at submit time

- The route stage carries `required_capability` (e.g. `document.signoff`, `document.review`) and optional `required_role` / `area_code`.
- At Submit time, the runtime copies the stage's `required_capability` into the stage instance row. From that point the capability is **frozen**; later mutation of the route stage (within an `active=true` route, before any new submit) does not retroactively change in-flight stage instances.
- This is the route-admin half of the two-tier authz model: the route admin **declares** the capability; the runtime **enforces** it in-tx via `authz.Require` (`wiki/decisions/0007-two-tier-authz.md`). The pin is what makes the runtime check deterministic — the runtime never re-resolves capability from a moving target.
- Therefore: editing `required_capability` on an active route is safe and does not retro-affect open instances; deactivating a route with open instances is unsafe and is blocked by §3 above.

### 5. Deactivate reason is mandatory and lands in `governance_events`

- `DeactivateRouteInput.Reason` is required. Empty / whitespace-only rejected with `ErrRouteDeactivateReasonRequired` → `400 validation.required` at the handler.
- On success the service emits one `governance_events` row in the same tx as the `UPDATE approval_routes SET active=false`:
  - `event_type = "route.config.deactivated"`
  - `resource_type = "approval_route"`, `resource_id = <route_id>`
  - `payload_json = {"route_id": "...", "active": false, "reason": "..."}` (`route_admin_service.go:442-446`)
- The reason is therefore part of the durable, append-only governance trail. There is no separate "reason" column on `approval_routes` — the payload is the source of truth.
- Why this matters for PR-4: a deactivated route in the list view should be displayable with its reason. PR-4 may need a future "history" affordance; that is out of scope for the rewrite but is unblocked because the data is already captured.

### 6. Tier-1 cap split for route admin reads is deferred to F-001 follow-up

The Tier-1 declarative table at `apps/api/cmd/metaldocs-api/permissions.go` currently gates `GET /api/v1/approval/routes` and the `POST`/`PUT`/`DELETE` route admin verbs with the same `route.admin` (manage-grade) cap. This conflates read and write per the F-001 audit pattern (see [ADR 0016](0016-view-grade-capabilities.md) for the four view-grade caps that unblocked the F-001 split elsewhere).

The principled fix is to:
- Introduce `CapRouteView` (string `route.view`) in `internal/modules/iam/domain/model.go`.
- Add `route.view` to the role-capability seed for any role that should read the route catalogue (minimum: every role that needs to render the route picker — `approver`, `author`, `editor`, `system_admin`, `area_admin`; `viewer` if/when the picker is shown read-only).
- Add a `{method: GET, path: "/api/v1/approval/routes", cap: CapRouteView}` row above the existing `route.admin` row in `permissions.go`.
- Update the Tier-2 read path in `route_admin_service.go:486` to require `CapRouteView` instead of `route.admin` on the List path; keep `route.admin` on Create/Update/Deactivate.

**This change is deferred** because:
- The route admin page is admin-only today (only `system_admin` reaches it). The over-grant has no real-world effect yet.
- Splitting read/write here is symmetric with the F-001 work and should land in that follow-up PR alongside the other Tier-1 split rows, not piecemeal.
- The deferred work has its own row in the F-001 follow-up plan and is cross-linked from `wiki/concepts/authz-tiers.md` §"Tier-1 rule authoring rules".

PR-4 should not block on this. It consumes the existing endpoint with the existing cap.

## Consequences

- **Positive:** a single doc PR-4 can cite for every lifecycle question. No more re-reading service + handler + schema + trigger.
- **Positive:** the "no reactivate" rule keeps the governance trail clean — a route's history is a single, monotonically-progressing record, not a yo-yo of active/inactive.
- **Positive:** OCC + `If-Match` makes concurrent edits in the admin UI deterministic and resolvable client-side.
- **Positive:** the in-use guard means deactivate can never silently orphan an in-flight signoff chain. The cost is a UX rough edge (admin must wait or cancel) that PR-4 must surface clearly.
- **Negative:** deferred Tier-1 split means `route.admin` (manage-grade) still gates the read path. Audit-trail granularity ("user X exercised route.admin on a GET") is imprecise until the follow-up lands.
- **Negative:** the immutability of deactivated routes means a typo in `name` on a heavily-used route requires a new route + re-pointing of profile usage. Acceptable tradeoff for governance clarity; revisit if real-world usage shows pain.
- **Open:** future "soft archive with restore window" pattern (analogous to `wiki/decisions/0010-soft-archive-via-timestamp.md` for documents) — currently rejected for routes because the immutability guarantee is load-bearing for the in-instance pin; revisit if/when the pin design changes.

## IAM roles source — spike result (for PR-4 consumer)

**Status: endpoint does not exist.** Searched `internal/modules/iam/delivery/http/` and the OpenAPI spec — there is no `GET /api/v1/iam/roles` (or any equivalent). The role catalogue is defined entirely in code at `internal/modules/iam/domain/model.go:10-16`:

| Constant | String |
|---|---|
| `RoleApprover` | `approver` |
| `RoleAuthor` | `author` |
| `RoleEditor` | `editor` |
| `RoleSystemAdmin` | `system_admin` |
| `RoleViewer` | `viewer` |

The frontend page hard-codes the same list at `frontend/apps/web/src/features/approval/pages/RouteAdminPage.tsx:10` as `STAGE_ROLES`. This drifts the moment a role is added.

**Proposal — `GET /api/v1/iam/roles`** (deferred to PR-4, or its own micro-PR if non-trivial):

- **Path:** `GET /api/v1/iam/roles`
- **Capability:** `CapMembershipView` (`membership.view`). Per [ADR 0016](0016-view-grade-capabilities.md) the matrix grants `membership.view` to every role, which matches the use case (the route admin page UI needs the labels; broader read access is harmless because the catalogue is non-secret reference data).
- **Tier-2:** not required. The catalogue is tenant-invariant; no area scope.
- **Response shape:**
  ```json
  {
    "roles": [
      {"code": "approver",     "label": "Aprovador"},
      {"code": "author",       "label": "Autor"},
      {"code": "editor",       "label": "Editor"},
      {"code": "system_admin", "label": "Administrador do sistema"},
      {"code": "viewer",       "label": "Visualizador"}
    ]
  }
  ```
- **Implementation sketch:** new handler at `internal/modules/iam/delivery/http/admin_handler.go` (or sibling file) that iterates a deterministic-ordered list of `domain.Role` consts and emits a static i18n-friendly label map. Labels live in code for now; promote to DB-backed only if/when a tenant-overridable label requirement appears.
- **OpenAPI:** new operation `listIamRoles`, component schemas `IamRole` and `IamRolesResponse`.
- **Out of this ADR:** the actual endpoint implementation. This ADR records the spike result and the proposal; PR-4 (or a micro-PR before it) implements.

Until that endpoint lands, PR-4 should:
- **Option A (preferred if endpoint lands first):** consume `GET /api/v1/iam/roles` via the generated TanStack Query hook.
- **Option B (fallback):** move the literal into a shared `frontend/apps/web/src/lib/iam/roles.ts` constants module with the same shape as the proposed response, so the eventual swap to the API is a one-line change.

`wiki/modules/iam.md` row addition + cross-link to this ADR is part of the deliverables of this PR.

## Alternatives Considered

| Option | Verdict | Reason |
|---|---|---|
| Add a `reactivate` verb (`POST /api/v1/approval/routes/{id}/reactivate`) | Rejected | Breaks the immutability guarantee that the `enforce_route_immutable` trigger and the audit trail depend on. A reactivated route would need a separate version-history concept; new-route-instead is simpler and matches how QMS auditors think about controlled artifacts. |
| Drop `If-Match` and last-write-wins on Update | Rejected | Concurrent admin sessions are rare but real (multi-admin tenants). Silent overwrite of a peer's edit on a route that drives every approval is exactly the class of bug OCC exists for. |
| Make `Reason` optional on Deactivate | Rejected | The reason is what auditors read in `governance_events`. Optional → empty in practice. Mandatory keeps the trail useful with no UX cost (the UI can require a single sentence). |
| Add `CapRouteView` in this PR | Rejected for scope | Belongs in the F-001 Tier-1 split follow-up, not this docs PR. Deferring keeps PRs single-axis. |
| Implement `GET /api/v1/iam/roles` in this PR | Rejected for scope | This is a doc-only PR by design (see SCOPE). The endpoint is a small spec change + handler + codegen rerun; lands in PR-4 or its own micro-PR. |
| Store `deactivated_reason` as a column on `approval_routes` | Rejected | The reason belongs in the immutable governance trail, not in the row that is otherwise stable. Reading "current reason" off the row is a feature nobody asked for; reading the audit log is what auditors do. |

## Rollback

- This ADR is doc-only. Rollback = revert this file + the concept doc + the module wiki cross-links. No code or schema change is gated on this ADR landing.
- The behavior described here was shipped by PR-1 + PR-2; this ADR codifies it. If a future change reverses any of these rules (e.g. introduces reactivate), open a new ADR superseding §X of this one and update the state-machine diagram.

## References

- `wiki/modules/approval.md` §12 — Glossary terms (Route, Stage, Quorum, J1, SoD)
- `wiki/decisions/0007-two-tier-authz.md` — two-tier authz model the per-stage cap pin participates in
- `wiki/decisions/0016-view-grade-capabilities.md` — view-grade cap pattern (deferred `CapRouteView` will follow this shape)
- `wiki/decisions/0017-signoff-idempotency-fingerprint.md` — client-stable fingerprint rule that route admin idempotency now follows
- `wiki/concepts/approval-routes.md` — plain-language explanation of routes for non-engineer readers
- `wiki/concepts/authz-tiers.md` — Tier-1 rule authoring rules; cross-references the deferred split
- Migration 0145 — `enforce_route_immutable` trigger (raises `P0001` with `ErrRouteInUse:` prefix)
- `internal/modules/documents/approval/repository/errors.go:23,65` — `ErrRouteInUse` + prefix-match mapping
- `internal/modules/documents/approval/application/route_admin_service.go:343,442` — Deactivate + governance event emit
- `internal/modules/iam/domain/model.go:10-16` — Role catalogue (source of truth for the proposed `GET /api/v1/iam/roles`)
