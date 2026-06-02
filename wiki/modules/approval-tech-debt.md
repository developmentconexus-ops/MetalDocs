# Tech Debt Register — approval

> Companion to `wiki/modules/approval.md`. Lists known gaps, smells, and missing-ADR items. **Debt only — no fix prescriptions.** Fixes belong in `wiki/backlog/approval-refactor.md`.

**Last verified:** 2026-06-02 (Approval route admin PR-5 — pure wiki/backlog sync: composite tech-debt row T-015 registered closing FE-1..FE-20 + C-2 + BE-3..BE-12 across PR-1..PR-4; BE-1 marked resolved under T-002 (route-admin OpenAPI coverage shipped PR-1, signoff/cancel doc-scoped routes still pending); BE-9 explicitly carried over to F-001 follow-up (Tier-1 `route.view` split / `CapRouteView`); backlog row `R-RouteAdmin-Rewrite` added; bugs-audit K2 flipped `wont-fix` → `fixed`; approval module summary risks tally recomputed. Prior PR-4 verification preserved below.

**Prior verification — 2026-06-02 (PR-4):** Approval route admin PR-4 — FE rewrite of `RouteAdminPage` on canonical patterns closes the legacy FE scan items FE-1..FE-20 + C-2: TanStack Query for reads, optimistic+rollback mutations on writes; thin `routeAdminApi.ts` consumes `components['schemas']` codegen; native `<dialog>` primitive at `components/ui/Dialog.tsx` (no `@radix-ui` dep); stage rows keyed by stable uuid; `useAreasQuery` replaces silent `fetchAreas().catch(()=>{})`; 412/409/422/403 mapped to distinct PT-BR copy via `messageForRouteError`; status badge visually distinct; cause-based edit-disabled tooltip; role select fed by `useIamRolesQuery` (ADR 0018 Option B fallback); legacy `getJSON` + 4 route fns + 9 route-admin DTOs deleted; `ControlledDocumentDetailPanel.tsx` migrated to `routeAdminApi.listRoutes` + `RouteSummary`. FE-side debt rows to be authored in PR-5 will reference this PR. Prior PR-2 verification preserved below.

**Prior verification — 2026-06-01:** Approval route admin PR-2 backend hardening — closes route-admin scan items BE-3 (missing idempotency persistence), BE-4 (untyped list envelope), BE-5 (two-tx authz + read TOCTOU), BE-6 (missing governance reason on deactivate), BE-8 (asymmetric stage normalization), BE-10 (inconsistent error wrap), BE-11 (dead `Members` field on `ListStageItem`), BE-12 (route admin service unaware of `Idempotency-Key`). Implementation: `internal/modules/documents/approval/application/route_admin_service.go` + `internal/modules/documents/approval/infrastructure/postgres_route_admin_idemp_store.go` + `internal/modules/documents/approval/http/route_admin_handler.go`; tests under `application/route_admin_service_test.go` (`TestRouteAdminCreate_ReplayReturnsPriorResponse`, `TestRouteAdminCreate_IdempotencyKeyConflict`, `TestRouteAdminDeactivate_RejectsEmptyReason`, `TestRouteAdminDeactivate_ReasonInGovernancePayload`, `TestRouteAdminList_RunsUnderTenantGUC`) and `http/route_admin_handler_test.go` (`TestDeactivateRoute_RejectsEmptyReason`, `TestListRoutes_PassesTenantAndActor`). Prior verification — T-013 signoff content-pin canonicalization drift closed.)

## Severity scale

Source: `.claude/skills/metaldocs-module-doc/templates/tech-debt-register.md`. Use the trigger list. When in doubt and the bug is on a regulated path (ISO 9001 sign-off chain), escalate one level.

## Items

### T-001 · RFC 9457 envelope absent across approval HTTP surface — CLOSED 2026-05-12 (Plan 7)
- **Severity:** critical (closed)
- **Surface (resolved):** `internal/modules/documents/approval/http/errors.go:23` (`MapErrorToResponse`) returns `*problem.Problem`; `:136-141` (`WriteError`) calls `problem.Write(w, prob)`. `internal/modules/documents/approval/http/contracts/errors.go` — emptied to `package contracts` (error sentinels remain in `errors.go`). Frontend `frontend/apps/web/src/features/approval/api/mutationClient.ts:55-66` — `parseProblem(res.clone())` is now tried first in all non-2xx branches; legacy `{error:{code,message}}` fallback retained for incremental rollout.
- **Observation (original):** All non-2xx responses returned legacy `{error:{code,message,details,trace_id}}` envelope. Frontend `ApprovalError` parsed the legacy shape; switching server-side required a coordinated frontend update.
- **Evidence:** `_artifacts/02-flow-inbox.md` §5; `_artifacts/02-flow-signoff.md` §5; `_artifacts/05-industry.md` IP-001 row.
- **Linked backlog row:** `backlog/approval-refactor.md#R-001` (merged Plan 7 2026-05-11, commit `b8747d6a` + `c4f5535f`)
- **Linked ADR:** `wiki/architecture/api-design-system.md`

### T-002 · Signoff & cancel document-scoped routes absent from OpenAPI
- **Severity:** critical
- **Surface:** `api/openapi/spec2.yaml` (no entry for `POST /api/v1/documents/{id}/signoff`, `/signoffs`, `/cancel`); routes wired manually in `internal/modules/documents/approval/http/router.go:6-30`
- **Observation:** Per Phase 2 trace, `SignoffByDocumentHandler`, `CancelByDocumentHandler`, and several v2 doc-action POSTs have no OpenAPI operationId or response schema. Frontend cannot codegen types; `ApprovalError` and `signoffOutcome` shapes are hand-maintained. Contract drift downstream consumers rely on.
- **Evidence:** `_artifacts/02-flow-signoff.md` §1+§5; `_artifacts/02-flow-submit.md` §5; surface scan §3.
- **Linked backlog row:** `backlog/approval-refactor.md#R-002`
- **Linked ADR:** missing-ADR
- **Sub-item progress 2026-06-02:** BE-1 (route-admin `/api/v1/approval/routes` × 4 absent from OpenAPI) — **resolved** by PR-1 (commit `f57d5525e`). Component schemas `CreateRouteRequest`, `UpdateRouteRequest`, `DeactivateRouteRequest`, `RouteResponse`, `ListRoutesResponse`, `RouteSummary`, `StageRequest`, `StageSummary`, `QuorumKind`, `DriftPolicy` shipped; `api.gen.go` + FE `lib/api-types/index.d.ts` regenerated. T-002 stays open for the remaining signoff/cancel/v2 doc-action POSTs called out above.

### T-003 · `looksLikeValidationError` substring classifier — completed-instance signoff returns 500 — CLOSED 2026-05-12 (Plan 7)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/documents/approval/http/errors.go:82-84` — `domain.ErrNoActiveStage` now has an explicit `errors.Is` branch: 409 `state.instance_completed`. The substring fallback (`looksLikeValidationError` at `:127`) is retained for genuine validation strings but is no longer the path for `ErrNoActiveStage`.
- **Observation (original):** `domain.ErrNoActiveStage` ("no active stage in this approval instance") did not match the substring classifier → 500 `internal.unknown` instead of the correct 409. E4 finding from prior stub.
- **Evidence:** prior stub `wiki/modules/approval.md` §E4; `_artifacts/02-flow-inbox.md` §5.
- **Linked backlog row:** `backlog/approval-refactor.md#R-003` (merged Plan 7 2026-05-11, commit `b8747d6a`)
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
- **Resolution sync 2026-05-25:** CLOSED by 5c High hardening. `cancel_service.go` now sets authz/RLS context before instance load, requires `workflow.instance.cancel` for user cancels, keeps system bypass behind `SystemCancelInstance`, tenant-scopes stage cancellation through `approval_instances`, and checks cancel event `json.Marshal` errors. `cutover_service.go` now counts legacy statuses inside a transaction with `authz.BypassSystem`, so RLS cannot hide rows during preflight. Verified with `go test ./internal/modules/documents/approval/application ./internal/modules/documents/approval/repository -count=1`.

### T-007 · `infra/signature/` vs `infrastructure/` package naming inconsistency
- **Severity:** minor
- **Surface:** `internal/modules/documents/approval/infra/signature/` (signature providers) vs `internal/modules/documents/approval/infrastructure/` (signoff idempotency store)
- **Observation:** Two adjacent package directories with near-identical purpose (infrastructure adapters) but different names. Forces consumers to memorise which adapter lives where.
- **Evidence:** surface scan §1 (file tree).
- **Linked backlog row:** `backlog/approval-refactor.md#R-007`
- **Linked ADR:** missing-ADR

### T-008 · `approval_instances.document_v2_id` column retains `_v2` suffix post-cutover — CLOSED 2026-05-15 (0194)
- **Severity:** minor (closed)
- **Surface (resolved):** `migrations/0194_*` migration evidence; `internal/modules/documents/approval/repository/postgres_approval_repository.go` now reads `document_id`.
- **Observation (original):** Cutover from `documents_v2` → `documents` completed, but `approval_instances` still used legacy `document_v2_id`.
- **Evidence:** migration `0194` + repository surface aligned to `document_id`; persistence map sync updated.
- **Linked backlog row:** `backlog/approval-refactor.md#R-008` (merged 2026-05-15 with 0194 evidence)
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

### T-014 · Idempotency store wrapper duplication
- **Severity:** minor
- **Surface:** `internal/modules/documents/approval/infrastructure/postgres_route_admin_idemp_store.go` and `internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go`
- **Observation:** ~95% wrapper code is duplicated across the two stores around `internal/platform/idempotency.Store` — `ReplayHandle` adapter, `beginReplay` helper, JSON envelope marshalling. ~80 LOC duplicated. Promote `TypedStore[T any]` generic helper into `internal/platform/idempotency/` when the 3rd module-local store appears (taxonomy admin / cd admin / templates). YAGNI today.
- **Resolution:** Defer. Revisit at 3rd-store boundary.
- **Linked backlog row:** n/a
- **Linked ADR:** [`0017-signoff-idempotency-fingerprint.md`](../decisions/0017-signoff-idempotency-fingerprint.md)

### T-013 · Signoff content-pin canonicalization drift — CLOSED 2026-06-01
- **Severity:** critical (closed)
- **Surface (resolved):** `internal/modules/documents/approval/application/decision_service.go:162-176` — signoff now loads the content hash via the same COALESCE used by `/api/v1/controlled-documents/{cd}/active-document` (`internal/modules/controlleddocuments/delivery/http/routes.go:320-343`): `documents.content_hash_at_submit` with fallback to the latest `document_revisions.content_hash`. Compared against the FE-echoed `content_hash` body field; persisted as `approval_signoffs.content_hash`.
- **Observation (original):** `RecordSignoff` recomputed a content hash with three diverging knobs vs the value the FE received from the active-document endpoint: `DocumentID=req.InstanceID` (was instance id, not document id), `RevisionNumber=0` (vs submit's actual revision), and `FormData=loadCurrentDocumentFormData(...)` (vs the revision-time form snapshot the FE saw). Every `POST /api/v1/documents/{id}/signoff` returned 412 `precondition.content_hash_mismatch` — approvals fully blocked past the submit stage. The two stale comments at `decision_service.go:168-170` ("keyed on instance for signoff hashing" / "signoff hash does not embed revision") were drift, not a documented requirement (no ADR; original commit `911e61f344` adds them without rationale).
- **Resolution:** Replaced the recompute with a direct lookup of the same value `active-document` exposes; deleted the now-unused `loadCurrentDocumentFormData` helper. Persisted `approval_signoffs.content_hash` now equals the revision hash the FE pinned to, which is the semantically correct value for the regulated audit trail. Hardened per ultrareview: `loadActiveDocumentContentHash` now returns `ErrContentHashMismatch` when the row is missing (`sql.ErrNoRows`) or the COALESCE resolves NULL (no `content_hash_at_submit` and no revisions), and `RecordSignoff` rejects a request that omits `_content_hash` entirely — closing the defense-in-depth bypass for programmatic callers that skip the HTTP boundary's 64-hex validation.
- **Evidence:** Live chain `create → submit → signoff (approver) → approved` verified `2026-06-01`; instance `8ec1c370-bcfc-4f7f-b5e8-8d697dd2150c` reached `approved` (first iteration), then `a296dd2f-840c-4f7e-9c9e-58409c4cf503` reached `approved` (post-hardening) with `outcome=approved` HTTP 200. Signoff persisted with hash `ba20ff68…` matching `document_revisions.content_hash`. Unit suite `go test ./internal/modules/documents/approval/... -count=1` green after hardening.
- **Linked backlog row:** n/a (fix landed direct)
- **Linked ADR:** missing-ADR (signoff content-pin source should be promoted to ADR once finalized)

### T-012 · iam `AuthorizationService` / `ResourceCtx` / `ErrSoDViolation` unwired (cross-module) — CLOSED Plan 4
- **Severity:** minor
- **Surface:** `internal/modules/iam/application/authorization.go` (DELETED Plan 4 — iam T-003 closed; approval never imported it)
- **Observation:** iam `AuthorizationService` was deleted in Plan 4 (2026-05-11). Approval module enforces SoD via its own pure-domain `domain/sod.go:15` `CheckSoD` and never consumed the iam service. No adoption path needed; cross-module pointer moot. T-012 closed.
- **Resolution:** Deletion confirmed via Plan 4 commit b17a09e8. No approval-side change required.
- **Evidence:** `_artifacts/00-context.md` Decision A.
- **Linked backlog row:** `backlog/approval-refactor.md#R-012` (merged)
- **Linked ADR:** `wiki/decisions/0007-two-tier-authz.md`

### T-015 · RouteAdminPage legacy FE/BE drift (composite) — CLOSED by PR-1..PR-4
- **Severity:** high (composite — closed)
- **Scope:** Composite tracking row for the Route Admin canonical conversion sprint. Aggregates the pre-conversion scan items (FE-1..FE-20 + C-2 on the frontend; BE-1..BE-12 on the backend). Detailed per-row narrative lives in the prior verification stamps at the top of this file and in `wiki/modules/approval.md` PR-1..PR-4 stamps; this row is the closure index so future scans do not re-flag the same items.
- **Closing PRs:**
  - PR-1 — OpenAPI describe (commit `f57d5525e`, #47): BE-1.
  - PR-2 — backend hardening (commit `0d61b7165`, #48): BE-3 (idempotency persistence), BE-4 (untyped list envelope → `ListRoutesResponse`), BE-5 (two-tx authz + read TOCTOU collapsed into one tx), BE-6 (governance reason mandatory on deactivate), BE-8 (asymmetric stage normalization), BE-10 (inconsistent error wrap), BE-11 (dead `Members []string` field), BE-12 (route admin service unaware of `Idempotency-Key`).
  - PR-3 — ADR 0018 + concept doc (commit `ab7c29539`, #49): codifies route lifecycle (state machine, version OCC, in-use guard, capability pin, reason audit) and the deferred Tier-1 `CapRouteView` split. No code change closure here; lights the path for BE-9.
  - PR-3 follow-up — review fixes (commit `e4e0be399`, #50): BE-7 (decode/validate inconsistency on `DeactivateRouteHandler` → canonical `contracts.Decode` + `DeactivateRouteRequest.Validate`); also: stage INSERT batching, no-op stage diff skip, replay-before-validation ordering for Deactivate, slog of unreleased-slot errors, truthful `ListRoutesResponse.Total`.
  - PR-4 — FE rewrite (commit `5794e4d4e`, merge `4d2b9bce2`, #51): FE-1 (TanStack Query for reads), FE-2 (optimistic + rollback mutations), FE-3 (codegen `components['schemas']`), FE-4 (412/409/422/403 PT-BR copy via `messageForRouteError`), FE-5 (stable uuid stage row keys), FE-6 (`useAreasQuery` over silent `.catch(()=>{})`), FE-7 (cause-based edit-disabled tooltip), FE-8 (status badge visually distinct), FE-9 (`useIamRolesQuery` over frozen import), FE-10 (legacy `getJSON` deletion), FE-11 (route-admin DTO deletion from `approvalTypes.ts`), FE-12 (`ControlledDocumentDetailPanel.tsx` migration to `routeAdminApi.listRoutes` + `RouteSummary`), FE-13 (`mutationClient.mutate()` shared transport for `Idempotency-Key` + `If-Match`), FE-14 (`QK.approval.routes.{list,detail}` + `QK.iam.roles` query-key registration), FE-15 (cache invalidation on write), FE-16 (component decomposition into `RouteListTable` / `RouteEditorDialog` / `DeactivateRouteDialog`), FE-17 (deactivate reason gated input ≥4 chars), FE-18 (validation cascade label→role→area), FE-19 (m_of_n M=0 rejection), FE-20 (Esc closes editor + focus restore via native `<dialog>` primitive); C-2 (no `@radix-ui` dep added — `components/ui/Dialog` native `<dialog>` + `showModal()`).
- **Sub-items NOT closed by this composite:**
  - BE-9 — Tier-1 `route.view` capability split / `CapRouteView` adoption on `GET /api/v1/approval/routes` and route-detail reads. **Carried over to F-001 follow-up** per [ADR 0018](../decisions/0018-approval-route-lifecycle.md) §6 and `wiki/concepts/authz-tiers.md` §"Tier-1 rule authoring rules" rule 4. Tracked alongside the F-001 view-grade cap sweep.
- **Resolution:** No further work on this composite. Future regressions on the same scan items should re-open the specific FE-x / BE-x sub-item rather than reusing T-015.
- **Linked backlog row:** `backlog/approval-refactor.md#R-RouteAdmin-Rewrite` (merged 2026-06-02)
- **Linked ADR:** [`0018-approval-route-lifecycle.md`](../decisions/0018-approval-route-lifecycle.md)

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: ~110 / ~110
- Operations missing C4 placement: 0 / 16
- Cross-deps missing in §5/§8: 0 / 16 (7 OUT + 9 IN)
- State transitions missing in §6: 0 / 6
- Decisions without ADR link: 10
