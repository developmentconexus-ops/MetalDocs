# Module: approval

> **Last verified:** 2026-05-03
> **Scope:** Approval routes, signoffs, ISO segregation enforcement, freeze trigger.
> **Out of scope:** Freeze pipeline mechanics (see `workflows/freeze-and-fanout.md`).
> **Key files:**
> - `internal/modules/documents/approval/` — backend approval logic
> - `internal/modules/documents/approval/application/decision_service.go:284` — RecordSignoff QuorumRejectedStage branch (sets cancel GUC + transitions doc to draft)
> - `internal/modules/documents/approval/application/read_service.go:152` — ListInboxItems (JOIN against documents + signoff-count subquery)
> - `internal/modules/documents/approval/application/read_service.go:222` — CountPendingForActor (global count for pagination)
> - `internal/modules/documents/approval/http/inbox_handler.go:15` — InboxHandler (calls ListInboxItems + CountPendingForActor)
> - `internal/modules/documents/approval/http/handler.go:26` — readService interface (widened to include ListInboxItems + CountPendingForActor)
> - `internal/modules/documents/approval/http/errors.go:174` — looksLikeValidationError (E4 gap)
> - `internal/modules/render/fanout/pdf_dispatcher.go:27` — PDFDispatcher.Dispatch (outbox idempotency_key bug)
> - `internal/modules/documents/repository/repository.go:37` — CreateDocument INSERT with MAX(revision_number)+1
> - `internal/modules/documents/approval/http/handler.go:65` — `NewHandler` — accepts `signoffIdempStore` positional param
> - `internal/modules/documents/approval/http/doc_approval_handler.go:51` — `SignoffByDocumentHandler` with idempotency replay
> - `internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go:1` — `PostgresSignoffIdempStore`
> - `frontend/apps/web/src/features/approval/pages/InboxPage.tsx` — Caixa de Entrada
> - `frontend/apps/web/src/features/approval/pages/RouteAdminPage.tsx:7` — `StageDraft` interface; `toDraft` at :49 maps existing stage fields
> - `frontend/apps/web/src/features/approval/api/approvalTypes.ts:16` — `RouteStage` (includes `required_role`, `required_capability`, `area_code`)
> - `internal/modules/documents/approval/http/contracts/route.go:119` — `ListStageItem` (includes `RequiredRole`, `RequiredCapability`, `AreaCode`)
> - `internal/modules/documents/approval/http/route_admin_handler.go:207` — `ListRoutesHandler` SQL (selects all stage fields)
> - `frontend/apps/web/src/features/approval/components/SignoffDialog.tsx` — signoff dialog with password confirm

## Concepts

### Route

Sequence of stages. Each stage has steps. A step holds the quorum rule for that stage.

### Quorum kinds

- `any_1` — one signoff in the stage advances it.
- `m_of_n` — m of n named approvers.
- (others — verify against domain code)

### Signoff

A user's decision (approve / reject) recorded against a step. Stored with timestamp, user id, optional comment, password-confirmation flag.

## Domain rules

- **ISO segregation**: the document submitter cannot record a signoff on the same document version. Enforced at API, hidden in UI.
- **Idempotency**: re-submitting the same signoff with the same `Idempotency-Key` header returns `{"was_replay":true}` (HTTP 200). Backed by `PostgresSignoffIdempStore` which reads/writes `metaldocs.idempotency_keys`. Keys expire after 24 h.
- **Freeze trigger**: when the final stage's quorum is met, `decision_service.go` calls the freeze service inside the same transaction.

## Reject path

Rejecting any required signoff returns the document to `draft` so the author can edit and resubmit immediately. The `approval_instance` retains `rejected` status for audit; only the document row reverts.

**Implementation (fixed commit 2977ef96 — B4):** `QuorumRejectedStage` in `RecordSignoff` (`decision_service.go:284`):

1. Marks stage as `rejected_here` and approval instance as `rejected`.
2. Sets `SET LOCAL metaldocs.cancel_in_progress = '<instance_id>'` — the same GUC used by the cancel flow, which authorises the DB trigger to permit the `under_review → draft` transition.
3. Issues `UPDATE documents SET status = 'draft', revision_version = revision_version + 1` within the same transaction.

The approval instance record keeps `status = 'rejected'` for the audit trail. Author can edit the document and call finalize again for a new approval round.

## Route edit — stage config preserved (fixed 41ca209d)

`ListStageItem` previously only returned `label`, `members`, `quorum_kind`, `drift_policy`. Opening a route for editing caused `toDraft` to call `defaultStage()` for every stage, silently wiping `required_role`, `required_capability`, and `area_code`.

**Fixed (F3):**

1. `route_admin_handler.go:207` SQL extended: `SELECT …, required_capability, area_code, …`.
2. `ListStageItem` gains `RequiredRole`, `RequiredCapability`, `AreaCode` JSON fields.
3. `RouteStage` (frontend type) gains `required_role`, `required_capability`, `area_code`.
4. `StageDraft` gains `requiredRole`, `requiredCapability`, `areaCode`.
5. `toDraft` at `RouteAdminPage.tsx:49` maps each existing stage's fields instead of substituting `defaultStage()`.
6. `toRouteStages` at `RouteAdminPage.tsx:118` writes all three fields back on save.

## Known implementation gaps (as of 2026-05-03)

---

### E4 — Re-signoff on completed instance returns 500 instead of 409

**File:** `internal/modules/documents/approval/http/errors.go:174`

When a user attempts a second signoff on an already-approved document, the domain returns an error containing `"no active stage in this approval instance"`. The HTTP error mapper at `errors.go:174` calls `looksLikeValidationError`, which only matches substrings `" is required"`, `" must be "`, and `" must not be "`. The phrase `"no active stage"` does not match, so the error falls through to the 500 path and is returned as `internal.unknown`. The correct HTTP response should be `409 state.instance_completed`.

**Fix needed:** Either add `"no active stage"` (or a canonical sentinel error) to `looksLikeValidationError`, or switch to typed error matching in the HTTP layer.

---

### Outbox idempotency_key bug — PDFDispatcher silently no-ops on repeated dispatches

**File:** `internal/modules/render/fanout/pdf_dispatcher.go:27`

`PDFDispatcher.Dispatch` publishes a `messaging.Event` with no `IdempotencyKey` field set. If the platform messaging layer deduplicates by `IdempotencyKey` and the first call produced an empty key, all subsequent dispatches for different documents may be treated as duplicates and silently dropped. Symptom: freeze succeeds but no PDF is ever generated for documents after the first one.

**Fix needed:** Populate `IdempotencyKey` in `Dispatch`, e.g. `in.RevisionID` or a deterministic hash of `(tenant_id, revision_id)`.

---

### Nova Revisão — revision_number gap (FIXED in migration 0167)

**File:** `internal/modules/documents/repository/repository.go:37`

Previously `CreateDocument` did not include `revision_number` in its INSERT (defaulted to `1`), causing a `ux_documents_v2_cd_revision` unique-constraint violation on any controlled document that already had a document at `revision_number = 1`.

**Fixed by migration 0167:** `controlled_document_id` is now present on `public.documents`, and `CreateDocument` at `repository.go:37` computes `COALESCE(MAX(revision_number), 0) + 1` in the same INSERT via a subquery. The unique index `ux_documents_v2_cd_revision ON documents(controlled_document_id, revision_number)` (migration `0131`) now resolves correctly.

## See also

- [workflows/user-onboarding.md](../workflows/user-onboarding.md) — Step 8
- [workflows/approval.md](../workflows/approval.md)
- [concepts/iso-segregation.md](../concepts/iso-segregation.md)
- [workflows/freeze-and-fanout.md](../workflows/freeze-and-fanout.md)
