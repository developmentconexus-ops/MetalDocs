# Feature F2 — Spec

> **Milestone:** 2b — Approval Kernel Backend  ·  **Folder:** `f2-route-versioning-pool-validation`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-07 — operator (via ratified governing spec §3/§7 W1/W6; this
> feature's contract is grounded in verified runtime-truth corrections below)

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Plan (`2026-07-07-m2b-approval-kernel-backend.md` F2) assumed `approval_routes` needs new `version`/`is_active`/`superseded_at` columns from scratch. Runtime-truth verification (this session, cross-checked directly against `route_admin_service.go` and `db/baseline/0001_current_schema.sql`, not the plan's assumption) found `version` and `active` (not `is_active`) **already exist** and `Update()` already mutates the SAME row in place, bumping `version` in place. The actual gap: `enforce_route_immutable` (baseline:636) blocks **any** UPDATE on a route once ANY `approval_instance` references it (`route_id = OLD.id`) — with no column-level exception. Once a route is used even once, it is **permanently frozen**; there is no path to edit it again. That is the real defect W1 must close, not "add version tracking from scratch". | Resolved by direct source read; see plan.md for the exact fix shape. |
| 2 | Should Update() branch behavior (in-place vs new-row) be exposed as a new field/flag in the request, or inferred automatically from whether the route is in use? | Inferred automatically — the consumer (route-admin UI/API) does not need to know or declare whether a route is "in use"; the service transparently does the cheapest safe thing. This keeps the existing `UpdateRouteRequest`/`RouteResponse` contract stable (already carries `route_id` + `new_version` — spec'd generically enough it already covers "response id may differ from path id" without a schema change). |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** Route-admin API consumers (FE/tests) calling `PUT /approval/routes/{id}`; F4 (review-verdict service resolves stage pool via same route); F9 (delegation extends `ResolveEligibleActors`, must see the same `ErrEmptyStagePool` sentinel).
- **Contract:**
  - `PUT /approval/routes/{id}` (existing route, unchanged path/method/headers): if the target route is **not** referenced by any `approval_instances` row, behavior is unchanged — mutate the same row in place, `RouteResponse.route_id == {id}`, `new_version` = old version + 1.
  - If the target route **is** referenced by at least one `approval_instances` row (i.e. the current `enforce_route_immutable` trigger would reject an in-place UPDATE): the service instead creates a **new** `approval_routes` row (new `id`, `version` = old version + 1, `active = true`, same `tenant_id`/`profile_code`, new `name`/stages from the request), then marks the OLD row `active = false, superseded_at = now()`, and inserts the new stage set under the new route id. Response: `RouteResponse.route_id` = the **new** row's id (differs from the path `{id}`), `new_version` = the new version. Existing in-flight `approval_instances` keep resolving stages from their `RouteVersionSnapshot`/original `route_id` — untouched, still pointing at the old (now-superseded) row, which remains fully readable.
  - `enforce_route_immutable` gains a column-scoped exception: it must still allow `UPDATE ... SET active = FALSE, superseded_at = now()` (and nothing else) on an in-use row (needed for the supersede step above), while continuing to block any update that touches `name`, `profile_code`, or `version` on an in-use OR already-inactive row.
  - Unique constraint replaces `approval_routes_tenant_profile_key UNIQUE(tenant_id, profile_code)` with a partial unique `UNIQUE (tenant_id, profile_code) WHERE active` (exactly one active row per profile per tenant) plus `UNIQUE (tenant_id, profile_code, version)` (full version history stays unique).
  - `submit_service.go`: after `ResolveEligibleActors` per stage, an empty pool (`len(eligibleIDs) == 0`) returns the **existing** sentinel `domain.ErrEmptyEligiblePool` (already defined in `domain/errors.go`, already used by `decision_service.go`'s quorum evaluation but — runtime-truth check found — **not currently mapped** in `http/errors.go`, so it falls through to a generic 500 today). F2 adds ONE mapping entry (`domain.ErrEmptyEligiblePool` → 422) that fixes both this pre-existing unmapped-error gap AND the new submit-time check in one place — no new sentinel is introduced (avoids two names for the same concept). No instance/stage rows are inserted when this fires at submit (whole submit tx rolls back).
  - `Deactivate()` behavior is unchanged (still flips `active=false` in place on the current active row via the OCC path) — it does not create a new version, since deactivating retires the entire profile_code line rather than creating a successor.
- **Source of truth for the contract:** `docs/superpowers/specs/2026-07-07-approval-remediation-design.md` §3 (W1) / §7; runtime-truth corrections in this Interview record and `../milestone.md`.

## What this feature implements

- Migration `0287_approval_route_versioning.sql`: drop old unique constraint, add the two replacement unique constraints (partial-active + version-history), replace `enforce_route_immutable` function body with the column-scoped exception, add `superseded_at timestamptz` column (`version`/`active` already exist — not re-added).
- `route_admin_service.go` `updateTx`: branch — attempt in-place UPDATE first (cheap path, current behavior); on `ErrRouteInUse`, fall back to the supersede-with-new-row path in the same transaction.
- `domain/errors.go`: add `ErrEmptyStagePool` sentinel.
- `submit_service.go`: empty-pool check per stage.
- ADR 1 (route versioning) under `wiki/decisions/`, marking ADR 0018 §1/§3 superseded.
- Contract: no `openapi.yaml` schema shape change needed (`RouteResponse.route_id`/`new_version` already generic enough) — only description text clarifying "route_id may differ from the path id when the route was in use" is added to the PUT operation.

## Non-goals (mandatory)

- No changes to `Deactivate()` semantics (still in-place, single row, no versioning branch).
- No capability/authz changes (F3).
- No stage-kind runtime behavior wiring beyond what F1 already added (F4/F5).
- No hash-chain, signature-meaning, SLA, or delegation changes (F6/F7/F8/F9).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|---|---|---|
| Route not in use: Update() still mutates in place, same id, version+1 | `route_versioning_integration_test.go::TestUpdateRoute_NotInUse_MutatesInPlace` | real |
| Route in use: Update() creates new row, old row superseded (active=false, superseded_at set), old row's stages/definition untouched, new row id ≠ path id | `route_versioning_integration_test.go::TestUpdateRoute_InUse_CreatesNewVersion` | real |
| In-flight instance still resolves stages from its original (now-superseded) route row | same test, assert via `LoadRoute`/stage load on the old id | real |
| Direct SQL UPDATE of `name`/`profile_code`/`version` on an in-use or inactive row → P0001 tripwire | same test file, direct `tx.Exec` against the trigger | real |
| Direct SQL UPDATE of only `active`/`superseded_at` on an in-use row → succeeds | same test file | real |
| Empty stage pool at submit → `ErrEmptyStagePool` → 422, no instance/stage rows persisted | `submit_service_test.go::TestSubmit_EmptyStagePool_Returns422` + integration counterpart | real |
| No regression | `go build ./...` clean; existing route-admin + submit test suites green | real |

## ADR needed?

- [x] Yes — durable decision: route versioning + supersession model. Written this feature, marks
  ADR 0018 §1/§3 superseded (`wiki/decisions/00XX-approval-route-versioning.md`).
