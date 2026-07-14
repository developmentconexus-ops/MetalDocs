# Phase 0 — Context (approval module)

**Date:** 2026-05-10
**Module path:** `internal/modules/documents/approval/`
**Existing stub:** `wiki/modules/approval.md` (Last verified 2026-05-08) — covers J1/E4/route-edit fixes + Caixa de Aprovação inbox UI; not Arc42-shaped.

## Inputs read

- `wiki/README.md` — module catalogue.
- `wiki/modules/approval.md` — predecessor stub.
- `wiki/decisions/0007-two-tier-authz.md` — tier-1/tier-2 authz model + amendment for tripwire (`migrations/0142b:200-209` attaches `enforce_capability_asserted` to `approval_instances` + `approval_signoffs`).
- `wiki/concepts/iso-segregation.md` — submitter-cannot-approve rule; UI+API enforcement points.
- `wiki/concepts/authz-tiers.md` — GUC pitfalls; tenant-degenerate area.
- `wiki/workflows/approval.md` — finalize→submit atomicity; signoff idempotency; reject path.
- `wiki/modules/iam-tech-debt.md` — T-003 (`AuthorizationService`/`ResourceCtx`/`ErrSoDViolation` unwired); T-004 (tripwire pairing on approval tables only).
- `wiki/modules/documents.md` (header) — IN-edge consumer; `CreateDocumentTx`, finalize trace.
- `wiki/modules/auth.md` — upstream session middleware; injects `iamdomain.WithAuthContext`.

## Cross-module ownership decisions (Phase 0 calls)

### Decision A — `AuthorizationService` (iam T-003) does NOT belong to approval module
Rationale: `application/authorization.go` lives in `internal/modules/iam/application/`. No approval-module file imports it. Approval enforces SoD via its own `domain/sod.go` (`CheckSoD` pure function) called from `decision_service.go`. Approval does not import `iamapp.AuthorizationService` or `ResourceCtx` or `ErrSoDViolation`.
**Conclusion:** T-003 stays in iam-tech-debt. Approval doc cross-references it as "third authz surface, unused — approval owns its own SoD path; not a substitute for the unwired iam service".

### Decision B — Tripwire pairing (iam T-004) IS the approval-module showcase
Rationale: migration `0142b_role_capabilities_v2_enforce.sql:200-209` attaches `enforce_capability_asserted` triggers exclusively to `public.approval_instances` and `public.approval_signoffs`. Approval is the sole module today whose mutating SQL is paired with the Postgres tripwire (defense-in-depth IP-004). Document the SECURITY DEFINER + GUC pattern thoroughly in §4 (persistence) and §8.1 (authz cross-cutting).

## IN-edges (consumers) confirmed at Phase 0
- `internal/modules/documents/delivery/http/handler.go:259` — `finalizeDocument` → `SubmitRevisionForReview`.
- `internal/modules/documents/application/...` — finalize trace (documents §6).
- Frontend `frontend/apps/web/src/features/approval/` — Caixa de Aprovação UI (existing stub already covers).

## OUT-edges expected
- `internal/modules/iam/authz` — `Require`, GUC helpers (`MustActorID`, `MustTenantID`).
- `internal/modules/iam/application` — `CapabilityService` (likely via injected port; verify Phase 3).
- `internal/modules/audit` (governance writer) — signoff governance events.
- `internal/modules/documents/repository` — document status transitions inside signoff tx.
- `internal/modules/render/fanout` — PDF dispatcher post-freeze (existing stub flagged outbox idempotency_key bug).
- `metaldocs.idempotency_keys` via `infrastructure/postgres_signoff_idemp_store.go`.
- `eligibility` + `sod` pure-domain modules.

## Phase 2 op picks (planned — confirm against §1 before dispatching)
1. **Read:** `GET /api/v1/approvals/inbox` → `InboxHandler` → `ListInboxItems` + `CountPendingForActor` (already mapped in stub at `read_service.go:152,222`).
2. **Write (state-transitioning):** `POST /api/v1/documents/{id}/signoffs` → `SignoffByDocumentHandler` → `RecordSignoff` → eligibility/sod/quorum/freeze (J1 + 0180 trigger story).
3. **State-transition:** `POST /api/v1/documents/{id}/finalize` → `SubmitRevisionForReview` (instance lifecycle: nothing → pending → in-review). Captures route resolution + `eligible_actor_ids` snapshot + `under_review` document transition.

## Open questions (deferred to tech-debt unless answered in artifacts)

1. Does `RecordSignoff` invoke `authz.Require` AND rely on the tripwire, or does the tripwire alone gate inserts? (Phase 4 §5 will answer.)
2. RFC 9457 envelope status of approval HTTP handlers — same drift as iam/documents? (Phase 1/2.)
3. Reject path GUC `metaldocs.cancel_in_progress` — is it whitelisted by a state-transition trigger separate from the tripwire? (Phase 4 §3.)
4. Does `PostgresSignoffIdempStore` body-hash the request per Stripe IP-002, or only key-hash? (Phase 2.)
5. `SubmitRevisionForReview` path for `eligible_actor_ids` — single tx with FOR UPDATE on what rows?

## Constraints carried into composition

- One module this session — no scope creep into `documents` or `iam` rewrites.
- Severity rubric strict: SoD bypass = Critical; regulated audit-trail gap = Critical (ISO 9001 sign-off chain is THE auditor target); tripwire bypass = Critical.
- `maint:<kind>` debt_id only for the 5 enumerated kinds.
- RFC 9457 envelope: record actual handler state (legacy or 9457), do not assume.
- Single doc-only commit, no Co-Authored-By.
