# Concept: Approval Routes

> **Last verified:** 2026-06-01
> **Status:** Canonical concept doc.
> **Scope:** Plain-language explanation of approval routes, stages, quorum kinds, drift policies, and the route lifecycle. Reader audience: anyone who needs to *understand* the route catalogue without reading service code.
> **Out of scope:** Per-instance signoff runtime (see `wiki/modules/approval.md` §6). Quorum evaluation math edge cases (see `internal/modules/documents/approval/domain/quorum.go` and `drift.go`). HTTP contract (see [`wiki/decisions/0018-approval-route-lifecycle.md`](../decisions/0018-approval-route-lifecycle.md) and OpenAPI).
> **Key files:**
> - `internal/modules/documents/approval/domain/route.go:39` — `Route` aggregate, `Stage` value object, `Validate`
> - `internal/modules/documents/approval/domain/quorum.go` — quorum evaluation
> - `internal/modules/documents/approval/domain/drift.go:17` — `ApplyEligibilityDrift`
> - `db/baseline/0001_current_schema.sql:1865` — `approval_routes` table
> - [`wiki/decisions/0018-approval-route-lifecycle.md`](../decisions/0018-approval-route-lifecycle.md) — lifecycle ADR
> - [`wiki/decisions/0007-two-tier-authz.md`](../decisions/0007-two-tier-authz.md) — authz model the required-capability pin participates in

## Why routes exist

A **route** is the per-tenant *plan* for how a controlled document gets approved. Different document profiles ("operational procedure", "engineering drawing", "safety procedure") need different approval chains — different reviewers, different ordering, different quorum rules. Hard-coding one chain doesn't scale across tenants or document classes.

A route is selected at submit time by the document's profile, snapshotted into stage instances, and **frozen** — once a document enters approval, its chain is locked. The runtime then drives signoffs against the snapshot.

This concept is what separates "an approval flow" (declarative, editable, per-tenant config) from "an approval instance" (a specific document's live execution of one chain).

## Anatomy

```
Route
├── tenant_id          ─── owned by exactly one tenant
├── profile_code       ─── which document profile this route applies to
├── name               ─── human-readable label ("CD Approval — Engineering")
├── version            ─── monotonic OCC counter; bumps on Update + Deactivate
├── active             ─── boolean; terminal-on-deactivate (see lifecycle below)
└── Stages (ordered, dense from 1)
    ├── order                  ─── 1, 2, 3, … (no gaps)
    ├── name                   ─── unique within route ("Review", "Approval")
    ├── required_role          ─── e.g. "approver" (used for eligibility snapshot)
    ├── required_capability    ─── e.g. "document.signoff" (pinned at instance time)
    ├── area_code              ─── optional area scope ("QA-01") or "tenant"
    ├── quorum                 ─── any_1_of / all_of / m_of_n
    ├── quorum_m               ─── required only when quorum = m_of_n
    └── on_eligibility_drift   ─── reduce_quorum / fail_stage / keep_snapshot
```

Constraint: stage `order` must be dense and start at 1; stage `name` must be unique within the route. Both are enforced by `Route.Validate` at `domain/route.go:48`.

## Quorum kinds

Quorum is *how many signoffs satisfy a stage*. Three kinds:

| Kind | Meaning | Example |
|---|---|---|
| `any_1_of` | One eligible signoff approves the stage. | "Any approver on the QA-01 list signs." |
| `all_of` | Every eligible actor in the snapshot must approve. | "Both the safety lead and the engineering lead sign." |
| `m_of_n` | At least `m` out of the snapshot's `n` actors approve. Carries a `quorum_m` integer. | "At least 2 of 3 architects sign." |

A reject from any eligible actor short-circuits the stage to `rejected_here` regardless of kind — the document returns to draft (`wiki/workflows/approval.md`).

## Drift policies

Between submit and final signoff, the underlying *who is eligible* can change: a user gains or loses a role, an area membership is granted or revoked. The snapshot taken at submit (`eligible_actor_ids`) is the truth for that instance, but the live set can drift. **Drift policy** says how the stage reacts.

| Policy | Behavior when live eligible set shrinks vs snapshot |
|---|---|
| `reduce_quorum` | Denominator reduces with the live set. A 2-of-3 stage becomes 2-of-2 if one actor lost eligibility. Pending unless reduced denominator already satisfied. |
| `fail_stage` | Any departure from the snapshot fails the stage immediately. Most conservative; used where the *identity* of approvers is regulated, not just the count. |
| `keep_snapshot` | Snapshot is sticky; live changes ignored. The original `n` is the denominator; a departed actor can no longer sign but the quorum target does not move. |

Drift is *only* applied at signoff-time evaluation; it does not retroactively re-eligible anyone. See `domain/drift.go:17` for the exact algorithm; the unknown-policy fallback returns a diagnostic reason and treats the policy as `keep_snapshot`.

## Required-capability pin

Each stage carries `required_capability` (e.g. `document.signoff`, `document.review`). At submit time the runtime **copies** the stage's `required_capability` into the stage instance row. The pin is then **frozen** — later edits to the route stage do not retroactively change in-flight stage instances.

This is what makes the runtime check deterministic: `authz.Require(ctx, tx, <pinned cap>, <area>)` is evaluated against a value that cannot move under it. The route admin **declares** the cap; the runtime **enforces** it. See [`wiki/decisions/0007-two-tier-authz.md`](../decisions/0007-two-tier-authz.md) for the two-tier authz model this participates in.

Consequence for route admins: editing `required_capability` on an active route is safe — open instances keep their pin. Deactivating a route with open instances is **not** safe and is blocked by the in-use guard (next section).

## Lifecycle

```
   create  ──►  active=true  ──── deactivate ────►  active=false   (terminal)
```

- **Only forward.** No reactivate. Deactivation is terminal; the row is now immutable historical record.
- **To "restore" a deactivated route, create a new one.** The new row gets a fresh `id`, `version=1`, and accumulates its own audit trail.
- **Update is in-place** while `active=true`. Stages are replaced atomically per Update; partial stage edits are not supported.
- **Deactivate is blocked while in use.** The trigger `enforce_route_immutable` (migration 0145) raises `P0001` if any `approval_instances` row references the route in a non-terminal state. The HTTP layer surfaces this as `409 route.in_use`. Remediation: wait for in-flight instances to complete, or cancel them, before deactivating.
- **Version OCC.** Every successful Update and Deactivate bumps `version`. Clients send `If-Match: "<version>"`; mismatch → `409 state.stale_revision`.
- **Reason audit.** Deactivate requires a non-empty `Reason`; the reason is persisted in `governance_events.payload_json` under `event_type = "route.config.deactivated"` in the same tx as the `active=false` update. The audit trail is the source of truth — there is no `deactivated_reason` column on the route row.

Full lifecycle rationale + alternatives considered: [`wiki/decisions/0018-approval-route-lifecycle.md`](../decisions/0018-approval-route-lifecycle.md).

## Selection at submit time

When a document is submitted for review, the runtime looks up the *single active* route for `(tenant_id, profile_code)`. Uniqueness is enforced by `approval_routes_tenant_profile_key` UNIQUE constraint (`db/baseline/0001_current_schema.sql:2890`). At most one active route per tenant+profile is allowed; deactivated routes are excluded by `WHERE active = true`.

If no active route exists for the profile, submit fails fast (no fallback chain, no implicit default).

## Worked example

A tenant configures a route for `profile_code = "engineering-procedure"`:

| Stage | Name | Role | Cap | Area | Quorum | Drift |
|---|---|---|---|---|---|---|
| 1 | Technical Review | `editor` | `document.review` | `ENG-01` | `any_1_of` | `reduce_quorum` |
| 2 | Approval | `approver` | `document.signoff` | `ENG-01` | `m_of_n` (m=2) | `fail_stage` |

- A document of that profile submits. Eligibility snapshot for stage 1 = all `editor` role-holders in `ENG-01`; for stage 2 = all `approver` role-holders in `ENG-01`. Both caps are pinned to the stage instances.
- Any one editor signs → stage 1 completes; stage 2 opens.
- Two approvers sign → m_of_n satisfied → final approve path runs (freeze, document → `approved`, governance event).
- If between submit and stage-2 evaluation an approver loses `ENG-01` membership, the `fail_stage` policy aborts stage 2 immediately — the document returns to draft and the author resubmits (with the corrected route or after the membership is restored).

## Relationship to other concepts

- **Eligibility snapshot (J1):** routes *declare* who is eligible by role + area + cap; the snapshot at submit *materializes* the actor list. See `wiki/workflows/approval.md` §J1.
- **SoD:** the submitter cannot sign any stage of their own submission, independent of the route's role/cap config. See [`iso-segregation.md`](iso-segregation.md).
- **Authz tiers:** route admin operations require `route.admin` (manage-grade) cap today; a deferred Tier-1 split will introduce `route.view` for read paths — see [`authz-tiers.md`](authz-tiers.md) "Tier-1 rule authoring rules" and [`wiki/decisions/0018-approval-route-lifecycle.md`](../decisions/0018-approval-route-lifecycle.md) §6.

## See also

- [`wiki/decisions/0018-approval-route-lifecycle.md`](../decisions/0018-approval-route-lifecycle.md) — lifecycle ADR (formal decision record)
- [`wiki/modules/approval.md`](../modules/approval.md) — module doc (Arc42 + C4)
- [`wiki/workflows/approval.md`](../workflows/approval.md) — runtime workflow (submit, signoff, freeze)
- [`wiki/concepts/iso-segregation.md`](iso-segregation.md) — SoD rule (orthogonal to route config)
- [`wiki/concepts/authz-tiers.md`](authz-tiers.md) — two-tier authz the cap pin participates in
