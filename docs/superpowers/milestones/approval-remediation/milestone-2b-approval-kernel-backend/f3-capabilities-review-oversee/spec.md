# Feature F3 — Spec

> **Milestone:** 2b — Approval Kernel Backend  ·  **Folder:** `f3-capabilities-review-oversee`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-07 — pre-answered by the ratified governing spec §6.1/§6.2 and
> `plan.md` F3 section; no new interview needed (runtime-truth grounding below closes the only open
> question: which routes are actually still on the generic fallback).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Plan/milestone.md F3 row says "explicit tier-1 row per runtime approval verb" — which verbs are actually still falling through the generic `/api/v1/approval/` prefix block (`permissions.go:250-253`) today, vs already explicit? | Runtime-truth grep of `permissions.go` (this session): `submit`/`signoff`/`publish`/`schedule-publish`/`supersede`/`obsolete`/`cancel` are **already explicit** under `/api/v1/documents/*/{suffix}` rows (lines 170-179) — route-admin CRUD is also already explicit (BE-9, lines 245-247). Only **4** routes still resolve via the generic `/api/v1/approval/` block: `POST .../instances/{id}/stages/{sid}/signoffs` (stage signoff — tier-2 is `CapDocumentSignoff`, but falls through the generic POST row to `CapDocumentSubmit` — **mismatch**), `POST .../instances/{id}/cancel` (tier-2 is `CapDocumentEdit`, falls through to `CapDocumentSubmit` — **mismatch**), `GET .../instances/{id}` (tier-2 is `CapDocumentView`, falls through to `CapDocumentView` — matches, but must still get its own explicit row so the generic block can be deleted), `GET /approval/inbox` (actor-scoped by construction, no tier-2 capability check inside `ListInboxItems`/`ListInboxItemsWithTotal` today — tier-1 stays `CapDocumentView`). The stray generic PUT/DELETE rows (lines 252-253) have no surviving route under `/api/v1/approval/` (routes-admin and documents-scoped verbs own their own explicit rows) — dead once the 4 real GET/POST rows above are added; deleted along with the block. |
| 2 | Does F3 also change `cancel`'s tier-2 capability to `approval.oversee` (per spec §6.1's "cancel with reason" grant), or leave it `CapDocumentEdit`? | Leave it `CapDocumentEdit` — `approval.oversee`'s "cancel with reason" grant is the **required-reason** cancel contract change, which is F4/F8 scope (`plan.md` F4: `POST .../cancel` gains required `reason`). F3 only fixes the tier-1↔tier-2 **mismatch** (stage-signoff and instance-cancel tier-1 rows pointing at the wrong capability), it does not redesign which capability gates cancel. |
| 3 | Does F3 wire `CapApprovalReview` into a live route? | No — `POST .../stages/{stageId}/review-verdict` does not exist yet (F4 creates it). F3 registers the capability through all 10 touchpoints (const, scope classification, catalog, seed grants, DB tripwire arm, guard tests, registry size) so F4 can gate the new route on day one without its own capability-wiring feature; touchpoint #3 (tier-1 route→cap) is satisfied for `CapApprovalReview` by F4, not F3 — recorded as a bounded defer below, not a gap. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** tier-1 middleware (`iam_authz` chain link) for every runtime approval route; `ReadService.LoadInstance`/`LoadActiveInstanceByDocument` callers (get-instance handler, get-instance-by-document handler); F4 (consumes `CapApprovalReview` for the new review-verdict route); F8 (consumes `CapApprovalOversee` for the worklist `?scope=oversee` variant).
- **Contract:**
  - Two new capabilities registered end-to-end: `approval.review` (`CapApprovalReview`, area-scoped — acts on a stage in an area) and `approval.oversee` (`CapApprovalOversee`, tenant-scoped read — oversight of any instance in tenant). Neither collides with the existing `document.review` (`CapDocumentReview`, ADR 0069 periodic mark-reviewed workflow) or `document.signoff` (`CapDocumentSignoff`, unchanged, continues to gate stage signoffs).
  - `permissions.go`'s generic `/api/v1/approval/` prefix block (lines 250-253) is deleted. Every route it used to catch gets its own explicit tier-1 row matching its real tier-2 requirement: `POST .../instances/{id}/stages/{sid}/signoffs` → `CapDocumentSignoff` (fixes the mismatch); `POST .../instances/{id}/cancel` → `CapDocumentEdit` (fixes the mismatch); `GET .../instances/{id}` → `CapDocumentView`; `GET /approval/inbox` → `CapDocumentView`.
  - `ReadService.LoadInstance` and `LoadActiveInstanceByDocument` (get-instance paths) change their tier-2 gate from a single `authz.Require(CapDocumentView, "tenant")` call to an explicit two-capability check: pass if the actor holds `CapDocumentView` **or** `CapApprovalOversee` (never a role check; both capabilities independently sufficient).
  - Migration `0288_approval_caps_seed_tripwire.sql`: seed grants (quality-manager profile → `approval.oversee`; reviewer pools → `approval.review`, dev-seed parity) + regenerated DB tripwire arms from the Go registry (M2-generation pipeline) covering the two new capabilities wherever the registry's tripwire generator scopes them.
  - `TestCapabilityRegistrySize` bumps by exactly 2 (38 → 40).
- **Source of truth for the contract:** `docs/superpowers/specs/2026-07-07-approval-remediation-design.md` §6.1/§6.2; `docs/superpowers/plans/2026-07-07-m2b-approval-kernel-backend.md` F3 section; `.claude/skills/developing-new-work/references/capability-wiring.md` (10-touchpoint checklist, canonically instantiated at `docs/superpowers/milestones/global-maximum-remediation/milestone-6-eqms-review-reason/validation-contract.md` §3); runtime-truth grep in this Interview record.

## What this feature implements

- `internal/modules/iam/domain/model.go`: `CapApprovalReview Capability = "approval.review"`, `CapApprovalOversee Capability = "approval.oversee"` added to the const block + `validCapabilities`; scope classification (`CapApprovalReview` → area-graded, `CapApprovalOversee` → tenant-graded) in `capability_scope.go`; capability catalog description entries.
- `apps/api/cmd/metaldocs-api/permissions.go`: delete lines 250-253 (generic `/approval/` fallback); add 4 explicit rows (signoff, cancel, get-instance, inbox) per the Consumer contract above.
- `internal/modules/documents/approval/application/read_service.go`: `LoadInstance`/`LoadActiveInstanceByDocument` gate on `CapDocumentView` OR `CapApprovalOversee` (explicit two-capability check, not a role check).
- `db/migrations/0288_approval_caps_seed_tripwire.sql`: seed grants + regenerated tripwire arms (M2 pipeline).
- `internal/modules/iam/domain/model_test.go`: `TestCapabilityRegistrySize` want 38 → 40.
- `internal/modules/iam/tripwire_caps_test.go`, `apps/api/cmd/metaldocs-api/permissions_test.go`: extended/updated for the new rows and capabilities.
- ADR 3 (`approval.oversee` + visibility model) — durable decision, written this feature (visibility-model *behavior* beyond the tier-2 read gate above is F8; this ADR records the capability's existence and grant contract now, F8 extends it).

## Non-goals (mandatory)

- No new routes (`review-verdict` route creation is F4; `CapApprovalReview` gets no live tier-1 row in this feature).
- No change to `cancel`'s tier-2 capability (`CapDocumentEdit` unchanged) or to its request contract (`reason` field is F4/F8).
- No worklist `?scope=oversee` filter implementation (F8).
- No touch to route-admin tier-1/tier-2 rows (already explicit, BE-9, 2026-07-02).
- No touch to `submit`/`publish`/`schedule-publish`/`supersede`/`obsolete` tier-1 rows (already explicit under `/documents/*` prefixes, no mismatch found).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|---|---|---|
| `TestCapabilityRegistrySize` count exactly 38 → 40 | `go test ./internal/modules/iam/... -run TestCapabilityRegistrySize -v` | real |
| `CapApprovalReview`/`CapApprovalOversee` correctly scope-classified | `go test ./internal/modules/iam/... -run TestEveryCapabilityClassified -v`, `-run TestAreaGradeCapabilitySet -v` | real |
| Generic `/approval/` prefix block fully removed | grep-zero: `grep -n "pathPrefix: \"/api/v1/approval/\"" apps/api/cmd/metaldocs-api/permissions.go` → no match | real |
| Every ex-generic route has an explicit, correct tier-1 row | `apps/api/cmd/metaldocs-api/permissions_test.go` (extended: signoff → `CapDocumentSignoff`, cancel → `CapDocumentEdit`, get-instance → `CapDocumentView`, inbox → `CapDocumentView`) | real |
| Get-instance read paths accept `CapDocumentView` OR `CapApprovalOversee` | `internal/modules/documents/approval/application/read_service_test.go` (extended: oversee-only actor succeeds, no-cap actor denied) | real (fixture-driven fake-conn) |
| Both authz drift/parity lints green | the M2-generation drift lint + parity lint commands (per `wiki/decisions/0022-authz-capability-coherence.md` / M2 CI lints) | real |
| DB tripwire arms regenerated, no manual `TEXT[]` edits | `internal/modules/iam/tripwire_caps_test.go` (extended) + M2 generation pipeline re-run, diff-clean | real |
| No regression | `go build ./...` clean; `go test ./internal/modules/iam/... ./apps/api/... ./internal/modules/documents/approval/...` green | real |

## ADR needed?

- [x] Yes — durable decision: `approval.oversee` capability + its role in the visibility model.
  Written this feature (`wiki/decisions/00XX-approval-oversee-visibility.md`), records the capability's
  grant contract; F8 extends it with the full read-visibility predicate.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|--------------------------|
| `CapApprovalReview` has no live tier-1 route yet | The route it gates (`review-verdict`) does not exist until F4 creates it; wiring the capability now (const/scope/seed/tripwire/tests) lets F4 gate on day one with zero capability-wiring work of its own | F4's `spec.md`/`plan.md` step that adds the tier-1 row for `POST .../review-verdict` |
| Worklist `?scope=oversee` variant not implemented | Full visibility-gating + worklist filter model is F8's named scope per `plan.md` | F8's `spec.md` |
