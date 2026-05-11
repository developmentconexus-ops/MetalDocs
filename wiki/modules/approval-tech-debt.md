# Tech Debt Register — approval

> Companion to `wiki/modules/approval.md`. Lists known gaps, smells, and missing-ADR items. **Debt only — no fix prescriptions.** Fixes belong in `wiki/backlog/approval-refactor.md`.

**Last verified:** 2026-05-10

## Severity scale

Source: `.claude/skills/metaldocs-module-doc/templates/tech-debt-register.md`. Use the trigger list. When in doubt and the bug is on a regulated path (ISO 9001 sign-off chain), escalate one level.

## Items

### T-001 · RFC 9457 envelope absent across approval HTTP surface
- **Severity:** critical
- **Surface:** `internal/modules/documents/approval/http/contracts/errors.go:3-12`, `internal/modules/documents/approval/http/errors.go:147` (`WriteError`)
- **Observation:** All non-2xx responses return legacy `{error:{code,message,details,trace_id}}` envelope. `wiki/decisions/0007-two-tier-authz.md` and the documents-module doc both target RFC 9457 Problem+JSON. Frontend `ApprovalError extends ApiError` (`frontend/apps/web/src/features/approval/api/mutationClient.ts:9`) parses the legacy shape; switching server-side breaks the parser.
- **Evidence:** `_artifacts/02-flow-inbox.md` §5; `_artifacts/02-flow-signoff.md` §5; `_artifacts/05-industry.md` IP-001 row.
- **Linked backlog row:** `backlog/approval-refactor.md#R-001`
- **Linked ADR:** missing-ADR

### T-002 · Signoff & cancel document-scoped routes absent from OpenAPI
- **Severity:** critical
- **Surface:** `api/openapi/spec2.yaml` (no entry for `POST /api/v2/documents/{id}/signoff`, `/signoffs`, `/cancel`); routes wired manually in `internal/modules/documents/approval/http/router.go:6-30`
- **Observation:** Per Phase 2 trace, `SignoffByDocumentHandler`, `CancelByDocumentHandler`, and several v2 doc-action POSTs have no OpenAPI operationId or response schema. Frontend cannot codegen types; `ApprovalError` and `signoffOutcome` shapes are hand-maintained. Contract drift downstream consumers rely on.
- **Evidence:** `_artifacts/02-flow-signoff.md` §1+§5; `_artifacts/02-flow-submit.md` §5; surface scan §3.
- **Linked backlog row:** `backlog/approval-refactor.md#R-002`
- **Linked ADR:** missing-ADR

### T-003 · `looksLikeValidationError` substring classifier — completed-instance signoff returns 500
- **Severity:** major
- **Surface:** `internal/modules/documents/approval/http/errors.go:181`
- **Observation:** Validation errors are detected by substring match on `" must be "`, `" is required"`, `" must not be "`. Domain error `"no active stage in this approval instance"` (re-signoff after instance approved) does not match → 500 `internal.unknown` instead of 409 `state.instance_completed`. Pre-existing E4 finding from prior stub still stands.
- **Evidence:** prior stub `wiki/modules/approval.md` §E4; `_artifacts/02-flow-inbox.md` §5 confirms current code path unchanged.
- **Linked backlog row:** `backlog/approval-refactor.md#R-003`
- **Linked ADR:** missing-ADR

### T-004 · Deprecated post-commit PDF dispatcher path still compiled
- **Severity:** major
- **Surface:** `internal/modules/documents/approval/application/decision_service.go:43,60` (`NewDecisionService` accepts `PDFDispatchInvoker`; `WithPDFOutbox` overlays optional outbox); `internal/modules/render/fanout/pdf_dispatcher.go:27`
- **Observation:** When a caller constructs `DecisionService` without `WithPDFOutbox`, the legacy post-commit `PDFDispatchInvoker.Dispatch` path runs. That path emits `messaging.Event` with empty `IdempotencyKey`; if the messaging bus dedupes by `IdempotencyKey`, dispatches after the first one are silently dropped. Composition root `apps/api/cmd/metaldocs-api/main.go:313` does wire `WithPDFOutbox(pdfOutboxRepo)`, so live binary uses the outbox path — but the deprecated surface remains constructible (test fixtures, future callers).
- **Evidence:** prior stub "Outbox idempotency_key bug"; cross-deps artifact §3 main.go:313.
- **Linked backlog row:** `backlog/approval-refactor.md#R-004`
- **Linked ADR:** missing-ADR

### T-005 · Inbox uses two-query `LIMIT/OFFSET + COUNT` with snapshot drift
- **Severity:** major
- **Surface:** `internal/modules/documents/approval/application/read_service.go:153` (`ListInboxItems`) + `:224` (`CountPendingForActor`)
- **Observation:** Both SELECTs use raw `db.QueryContext`, no enclosing tx. A signoff committed between the two queries can produce `Total < len(Items) + 1` or vice versa. Pagination via `LIMIT/OFFSET` only — no cursor (IP-003 drift). Same shape as documents list per `wiki/architecture/data-model.md` "two-query LIMIT/OFFSET+COUNT pattern".
- **Evidence:** `_artifacts/02-flow-inbox.md` §2 + §6; `_artifacts/05-industry.md` IP-003 not-applicable row.
- **Linked backlog row:** `backlog/approval-refactor.md#R-005`
- **Linked ADR:** missing-ADR

### T-006 · Tripwire pairing audit not extended to cancel & cutover service paths
- **Severity:** major
- **Surface:** `internal/modules/documents/approval/application/cancel_service.go:47` (`CancelInstance` writes `metaldocs.cancel_in_progress` but pairing with `authz.Require` on the cancel call site not verified by Phase 4 audit); `internal/modules/documents/approval/application/cutover_service.go:39`
- **Observation:** Phase 4 statement-local audit (`_artifacts/04-persistence.md` §6) flagged 3 FAIL rows on `InsertInstance` / `InsertSignoff` / `UpdateInstanceStatus` — these are repository methods called from `submit_service`/`decision_service`/`obsolete_service` where `authz.Require` does precede in the calling function (service-call-graph pairing OK). Not audited: cancel and cutover service call sites. Defense-in-depth gap if either path skips `authz.Require`.
- **Evidence:** `_artifacts/04-persistence.md` §6 (statement-local FAILs); cross-deps §1 (cancel_service.go:12 imports `authz`).
- **Linked backlog row:** `backlog/approval-refactor.md#R-006`
- **Linked ADR:** `wiki/decisions/0007-two-tier-authz.md` (covers tier-2 model)

### T-007 · `infra/signature/` vs `infrastructure/` package naming inconsistency
- **Severity:** minor
- **Surface:** `internal/modules/documents/approval/infra/signature/` (signature providers) vs `internal/modules/documents/approval/infrastructure/` (signoff idempotency store)
- **Observation:** Two adjacent package directories with near-identical purpose (infrastructure adapters) but different names. Forces consumers to memorise which adapter lives where.
- **Evidence:** surface scan §1 (file tree).
- **Linked backlog row:** `backlog/approval-refactor.md#R-007`
- **Linked ADR:** missing-ADR

### T-008 · `approval_instances.document_v2_id` column retains `_v2` suffix post-cutover
- **Severity:** minor
- **Surface:** `migrations/0135_approval_instances.sql:12`; `internal/modules/documents/approval/repository/postgres_approval_repository.go:36-37`
- **Observation:** Cutover from `documents_v2` → `documents` is complete (`migrations/0167`); the column on `approval_instances` still uses the migration-era name `document_v2_id`. UNIQUE constraint `(document_v2_id, idempotency_key)` carries the legacy name forward.
- **Evidence:** persistence map §1.
- **Linked backlog row:** `backlog/approval-refactor.md#R-008`
- **Linked ADR:** missing-ADR

### T-009 · `NOT VALID` FKs on tenant-scoped iam_users joins
- **Severity:** minor
- **Surface:** `migrations/0135_approval_instances.sql:28-31` (`approval_instances.submitted_by` → `metaldocs.iam_users`); `:95-98` (`approval_signoffs.actor_user_id`)
- **Observation:** Both FKs declared `NOT VALID` and never validated by a follow-up migration. Latent integrity gap — orphaned `submitted_by` / `actor_user_id` rows would not be caught at write time.
- **Evidence:** persistence map §1.
- **Linked backlog row:** `backlog/approval-refactor.md#R-009`
- **Linked ADR:** missing-ADR

### T-010 · ~110 exported symbols across domain/application/http undocumented
- **Severity:** minor
- **Surface:** `internal/modules/documents/approval/{domain,application,http}/**/*.go`
- **Observation:** Surface scan flagged every top-level symbol as `(undocumented)` except `WithMembershipContext` and the eligibility sentinel. Includes load-bearing types like `Instance`, `Signoff`, `EvaluateQuorum`, `RecordSignoff`, `SubmitRevisionForReview`.
- **Evidence:** surface scan §2.
- **Linked backlog row:** `backlog/approval-refactor.md#R-010`
- **Linked ADR:** missing-ADR

### T-011 · `WithMembershipContext` and `setAuthzGUC` both write authz GUCs
- **Severity:** minor
- **Surface:** `internal/modules/documents/approval/application/membership_tx.go:22` and `internal/modules/documents/approval/application/authz_guc.go:11`
- **Observation:** Two helpers set `metaldocs.tenant_id` + `metaldocs.actor_id` on the tx. `setAuthzGUC` is the call used by both `submit_service` and `decision_service`; `WithMembershipContext` predates it and is now only referenced internally. Risk: future caller picks the wrong helper.
- **Evidence:** surface scan §2; `_artifacts/02-flow-signoff.md` step 9.
- **Linked backlog row:** `backlog/approval-refactor.md#R-011`
- **Linked ADR:** missing-ADR

### T-012 · iam `AuthorizationService` / `ResourceCtx` / `ErrSoDViolation` unwired (cross-module)
- **Severity:** minor
- **Surface:** `internal/modules/iam/application/authorization.go` (not imported by approval)
- **Observation:** Per `wiki/modules/iam-tech-debt.md` T-003, an alternate authz surface exists in iam but is unwired. Approval module enforces SoD via its own pure-domain `domain/sod.go:15` `CheckSoD` and does not consume the iam service. Recorded here only as cross-module pointer; ownership stays in iam-tech-debt.
- **Evidence:** `_artifacts/00-context.md` Decision A.
- **Linked backlog row:** `backlog/approval-refactor.md#R-012`
- **Linked ADR:** `wiki/decisions/0007-two-tier-authz.md`

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: ~110 / ~110
- Operations missing C4 placement: 0 / 16
- Cross-deps missing in §5/§8: 0 / 16 (7 OUT + 9 IN)
- State transitions missing in §6: 0 / 6
- Decisions without ADR link: 10
