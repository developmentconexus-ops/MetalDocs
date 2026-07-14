# ADR 0074 — Approval Route Versioning + Supersession

> **Status:** Accepted
> **Date:** 2026-07-07
> **Scope:** `public.approval_routes` mutation model — supersedes ADR 0018 §1 ("Update is in-place
> only while active=true") and §3 ("Deactivate is blocked while any instance still references the
> route", as it pertains to Update — Deactivate's own guard is unchanged, see Consequences).
> **Milestone:** `docs/superpowers/milestones/approval-remediation/milestone-2b-approval-kernel-backend/f2-route-versioning-pool-validation/`
> **Key files:**
> - `db/migrations/0287_approval_route_versioning.sql` — schema change
> - `internal/modules/documents/approval/application/route_admin_service.go` — `updateTx` /
>   `updateInPlaceOrSupersede`
> - `tests/integration/approval/route_versioning_test.go` — live-DB proof

## Context

ADR 0018 §1 established: once a route is referenced by any `approval_instances` row, the
`enforce_route_immutable` trigger (migration 0145) rejects **any** UPDATE to it — permanently. There
was no path to ever edit a route again after its first real-world use. In practice this means a typo
in a stage's `required_role`, or any correction after go-live, requires standing up a brand-new
`profile_code` and re-pointing every consumer at it — disproportionate cost for a routine edit.

Runtime-truth verification (this feature, cross-checked against `route_admin_service.go` and
`db/baseline/0001_current_schema.sql`) also found the originally-planned fix ("add `version`/
`is_active`/`superseded_at` columns from scratch") was moot: `version` and `active` already exist and
`Update()` already increments `version` in place. The actual gap was narrower and purely at the
trigger layer.

## Decision

### 1. In-use routes are superseded, not frozen

`PUT /api/v1/approval/routes/{id}` keeps its existing path/method/headers. Behavior now branches on
whether the target route is referenced by any `approval_instances` row:

- **Not in use** (unchanged, cheap path): mutate the same row in place — `route_id` in the response
  equals the path `{id}`, `new_version` = old version + 1.
- **In use:** the service creates a **new** `approval_routes` row (new `id`, `version` = old version +
  1, `active = true`, same `tenant_id`/`profile_code`, new `name`/stages), then marks the **old** row
  `active = false, superseded_at = now()` in the same transaction, and inserts the new stage set under
  the new route id. Response: `route_id` = the new row's id (differs from the path `{id}`),
  `new_version` = the new version. In-flight `approval_instances` keep resolving stages from their
  original `route_id` / `RouteVersionSnapshot` — untouched, still fully readable (now superseded but
  never deleted).

The branch is inferred automatically from route state — the caller does not declare "in use"; the
existing `RouteResponse` contract (`route_id` + `new_version`) already covers "response id may differ
from the path id" with no schema change.

### 2. `enforce_route_immutable` gains a column-scoped exception

The trigger's in-use/inactive block is bypassed **only** when the UPDATE touches exclusively
`active`/`superseded_at` — `name`, `profile_code`, `version`, `tenant_id`, `created_at`, `created_by`
must all be unchanged (`IS NOT DISTINCT FROM`). Any UPDATE that also touches `version` (as
`Deactivate()`'s own UPDATE does — see Consequences) or any other definition column still hits the
original in-use/inactive guard and raises `P0001` (`ErrRouteInUse:` prefix, unchanged mapping).

This is the minimum column-scoped carve-out needed for the supersede step's own "mark old row
inactive" UPDATE — it is not a general relaxation of the freeze.

### 3. Unique constraints move from single-row to partial-active + version-history

- `approval_routes_tenant_profile_key UNIQUE(tenant_id, profile_code)` is dropped (it assumed exactly
  one row per profile per tenant, ever — incompatible with keeping superseded rows).
- Replaced by `approval_routes_active_profile_uq` — `UNIQUE (tenant_id, profile_code) WHERE active` —
  exactly one *active* row per profile per tenant, matching the pre-existing invariant callers depend
  on (route resolution by profile always finds ≤1 row).
- Added `approval_routes_profile_version_uq` — `UNIQUE (tenant_id, profile_code, version)` — full
  version history stays unique, so no two rows in the same profile lineage can share a version number.

### 4. Empty eligible-actor pool at submit is a 422, not a silent stuck instance

`submit_service.go` now checks, per stage after `ResolveEligibleActors`: if the resolved pool is empty,
return the existing sentinel `domain.ErrEmptyEligiblePool` (already defined, already used by
`decision_service.go`'s quorum evaluation) rather than proceeding to create a stage instance no one can
ever sign off. `http/errors.go` gains one mapping entry (`domain.ErrEmptyEligiblePool` → 422) — this
also closes a **pre-existing** gap: the sentinel had no mapping before this feature, so
`decision_service.go`'s quorum path fell through to a generic 500 on the same condition. No new
sentinel name is introduced (avoids two names for the same concept). The whole submit transaction rolls
back on this error — no partial instance/stage rows persist.

## Consequences

- **Positive:** a route can now be corrected after real-world use without abandoning its
  `profile_code` — the common case (fix a typo, adjust a stage role) no longer forces a new profile.
- **Positive:** in-flight approvals are never disrupted by a correction — they keep resolving the
  exact stage definition they were submitted against, via the now-superseded (but still-readable) row.
- **Positive:** the empty-pool check turns a silent "stuck forever, no eligible signer" instance into
  an explicit, actionable 422 at submit time.
- **Neutral — Deactivate() is unaffected.** `Deactivate()`'s own UPDATE sets `active = FALSE` **and**
  `version = version + 1` in the same statement. Because the trigger's exemption requires `version`
  unchanged, Deactivate's UPDATE never qualifies for the new column-scoped exception and still hits the
  original in-use/inactive guard — ADR 0018 §3 ("Deactivate is blocked while any instance still
  references the route") holds exactly as before. This was verified live: `TestRouteAdminDeactivate_*`
  unit tests are unmodified by this feature and remain green.
- **Negative:** `RouteResponse.route_id` returned from `PUT` can now legitimately differ from the path
  `{id}` the caller sent. Consumers that assumed echo-identity must read the response body's `route_id`
  as authoritative — this is a documented (not a schema) contract clarification added to
  `api/openapi/v1/openapi.yaml`'s `updateApprovalRoute` operation description.
- **Open:** M2c (frontend) must surface the "your edit created a new route version" case distinctly
  from "edited in place" — out of scope for this backend-only milestone.

## Alternatives Considered

| Option | Verdict | Reason |
|---|---|---|
| Require the caller to declare "in use" via a request flag | Rejected | Forces every consumer to first probe route usage before editing — extra round trip for no benefit; the service can determine this itself inside the same transaction. |
| Relax the trigger unconditionally for in-use routes (allow any column change) | Rejected | Reintroduces exactly the risk ADR 0018 built the freeze to prevent — an in-flight instance's stage definition could silently drift out from under a running approval. |
| Add a `reactivate`-style "unfreeze" verb | Rejected | Same rationale as ADR 0018's original rejection of reactivate — breaks the immutable-historical-record guarantee the audit trail depends on. |
| New sentinel `ErrEmptyStagePool` for the empty-pool check | Rejected | `domain.ErrEmptyEligiblePool` already exists and already means exactly this condition (used by `decision_service.go`) — two names for one concept is worse than mapping the existing one. |

## Rollback

Migration 0287 is additive/mechanical (new column, replaced constraints, replaced trigger function
body) — reversible by a follow-up migration restoring the single unique constraint and the
unconditional freeze trigger body, provided no superseded rows exist yet that would violate the
restored single-row-per-profile constraint.

## References

- `wiki/decisions/0018-approval-route-lifecycle.md` §1 (superseded by this ADR), §3 (Deactivate
  guard — unchanged, see Consequences)
- `docs/superpowers/specs/2026-07-07-approval-remediation-design.md` §3 (W1) / §7
- `docs/superpowers/milestones/approval-remediation/milestone-2b-approval-kernel-backend/f2-route-versioning-pool-validation/spec.md`
- `tests/integration/approval/route_versioning_test.go` — live-DB proof of the trigger exemption,
  in-use supersede sequence, and old-row stage immutability
