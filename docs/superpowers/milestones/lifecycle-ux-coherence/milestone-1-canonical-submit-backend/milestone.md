# Milestone 1 — Canonical Submit Backend

> **Program:** lifecycle-ux-coherence  ·  **Governing spec:** `docs/superpowers/specs/2026-07-06-lifecycle-ux-coherence-design.md`
> **Status:** Spec approved
> **Authored:** 2026-07-06 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front: **what** M1 is, **which features** it
> contains, **what each implements**, and **what gets validated**. No execution steps —
> the "how" lives in each feature's `plan.md`. The close gate (`qa/milestone-qa.md`)
> validates the milestone against *this* document.

## Objective

An author's **fresh-draft (`rev_version=0`) AND `REV≥1`** submit succeeds against the
canonical `POST /documents/{id}/submit` with **zero client-supplied governance data**
(no `route_id`, no `content_hash`) — the server resolves route, profile, content hash,
and governed revision number **inside the submit transaction** (closes the wrapper-era
TOCTOU). The deprecated `/finalize` chain is **deleted from Go**. The idempotency map is
**complete** (mark-reviewed + any Idempotency-Key-bearing template mutations).

Bar moved: ADR 0073 fully realized in Go; Grade-A invariants (capabilities ADR 0022,
contract-first, module boundaries ADR 0072, RFC 9457, tenant scoping) preserved. Closes
gap-register findings **1, 2, 3, 4, 5, 16, 17**.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F1 | `f1-intx-resolution` | `SubmitService` resolves `route_id`, `profile_code`, head `content_hash` in-tx when the client omits them: new narrow `SubmitDefaultsResolver` port (`LoadActiveRouteByProfile` mirroring `repository.go:1801-1817`, head content hash mirroring `1819-1828`) + `cdRead.ProfileCode` (tx-capable). Plain non-recording SELECTs (HS-PRE-1). Explicit client values keep prior semantics. | Integration (testdb): REV0 empty-body submit succeeds → 201 + ETag `"v1"`; explicit `route_id`/`content_hash` still honored; no active route → 409; profile absent → 400. Route/hash resolved off no client input. |
| F2 | `f2-submit-contract` | `contracts/submit.go`: `route_id` + `content_hash` optional (format-checked only when present); add `revision_title`. Align **handwritten** contract to the already-regenerated OpenAPI/gen on disk — do not re-edit the spec. | Empty body decodes+validates clean; malformed `route_id`/`content_hash` when present → 400; `revision_title` threaded to service. |
| F3 | `f3-error-mapping` | `approvalhttp/errors.go`: map `ErrRevisionTitleRequired`→422; `ErrDocumentNotDraft`→409; `ErrProfileNotConfigured`→400; `ErrApprovalRouteMissing`→409 (were finalize-only). RFC 9457 `problem+json`. | Each sentinel returns its mapped status+typed code, never 500. REV≥1 missing title → 422 typed. |
| F4 | `f4-delete-finalize` | Delete from Go: `finalizeDocument` handler, `parseFinalizeIfMatch`, bespoke finalize idempotency store wiring, `skippingMux` finalize exception, `GetFinalizePrereqs` chain (handler→service→repo→`domain.FinalizePrereqs`). **Keep** domain sentinels (now returned by submit). Fix `parseIfMatch` v0 test (v0 accepted). Delete finalize tests; migrate any invariant guards to submit-path. | `grep -ri finalize` over Go = zero in documents/approval submit path (domain sentinels + unrelated freeze/outbox "finalize" excluded). `go build ./...` + `go test ./...` green. |
| F5 | `f5-idempotency-map` | `mark-reviewed` present in approval `idempotentRoutes` (finding 16). Templates `archive` + `approval-config` (finding 17): add to templates idempotent map **iff** the OpenAPI spec declares `Idempotency-Key` for them; else spec+regen first, or bounded defer with written trigger. | mark-reviewed replay returns same result under repeated Idempotency-Key. Finding 17 disposition recorded with evidence (mapped, or deferred with trigger). |

## Milestone validation definition

Run by the **`milestone-validator` subagent** (separation of powers — it judges and writes
`qa/milestone-qa.md`; main session flips status only on its PASS), per the binding C1–C7
checklist in `.claude/skills/milestone/references/milestone-end-validation.md`.

1. **Per-feature acceptance** — every feature meets its "what to validate"; each `spec.md`
   consumer contract honored (producer matches consumer).
2. **Workflow-class QA** — `wiki/quality/backend-api-checklist.md` (route/contract/error
   discipline) + `wiki/quality/test-discipline.md` (testdb factory for DB integration).
3. **Regression** — `go build ./...` + `go test ./...` green; prior milestones/GMR gates intact.
4. **Root-cause check** — ADR 0073: no submit entrypoint rejects the true `v0` OCC state;
   in-tx resolution replaces off-tx TOCTOU; finalize wrapper gone, not shimmed.
5. **Live QA** — rebuild api image + docker stack + real `curl` submit of a fresh draft → 201.
6. **No unplanned scope** — YAGNI §4 refusals honored (no ETag sweep, no idemp-store
   consolidation, no template OCC); anything extra recorded with rationale.

## Dependencies & constraints

- Depends on: dirty-tree pre-work (OpenAPI `/finalize` removal + `/submit` extension,
  oapi-codegen + FE api-types regen, `parseIfMatch` v0 fix, ADR 0073) — **absorbed, never reverted**.
- FE edits in the dirty tree (`documents.ts`, `DocumentEditorPage`, tests) are **M2 scope** — left as-is.
- Constraints: capabilities-not-roles (ADR 0022; submit = `document.submit`+`document.edit`
  @area, in a **writable** tx); contract-first (spec is route truth); module boundaries
  (`CDFieldReader` port is the only cross-module read; ADR 0072); `tenant_id` on every query;
  RFC 9457 errors; resolution reads are non-recording SELECTs (HS-PRE-1).

## Applicable hard-stops

- **HS-1** (milestone boundary): operator review gate after validator PASS; no M2, no push without approval.
- **HS-2** (redesign outside boundary): a fix implying redesign beyond M1 (e.g. changing
  `content_hash_at_submit` semantics, cross-module auth model, storage) → stop, report boundary + minimum plan.
- **HS-3** (prereq boundary fails): build/runnable/auth-session/route/contract-truth failure → repair prereq, rerun, resume.
- **HS-4** (validator FAIL): open the named fix feature, re-run its lifecycle, re-dispatch validator.
- **HS-6** (scope drift): off-plan discovery mid-milestone → stop, surface, replan.
