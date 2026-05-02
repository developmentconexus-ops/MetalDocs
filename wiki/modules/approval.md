# Module: approval

> **Last verified:** 2026-05-02
> **Scope:** Approval routes, signoffs, ISO segregation enforcement, freeze trigger.
> **Out of scope:** Freeze pipeline mechanics (see `workflows/freeze-and-fanout.md`).
> **Key files:**
> - `internal/modules/documents_v2/approval/` — backend approval logic
> - `internal/modules/documents_v2/approval/application/decision_service.go:284` — RecordSignoff reject path (D4 gap)
> - `internal/modules/documents_v2/approval/http/errors.go:174` — looksLikeValidationError (E4 gap)
> - `internal/modules/render/fanout/pdf_dispatcher.go:27` — PDFDispatcher.Dispatch (outbox idempotency_key bug)
> - `internal/modules/documents_v2/repository/repository.go:44` — CreateDocument INSERT (revision_number gap)
> - `internal/modules/documents_v2/approval/http/handler.go:63` — `NewHandler` — accepts `signoffIdempStore` positional param
> - `internal/modules/documents_v2/approval/http/doc_approval_handler.go:51` — `SignoffByDocumentHandler` with idempotency replay
> - `internal/modules/documents_v2/approval/infrastructure/postgres_signoff_idemp_store.go:1` — `PostgresSignoffIdempStore`
> - `frontend/apps/web/src/features/approval/pages/InboxPage.tsx` — Caixa de Entrada
> - `frontend/apps/web/src/features/approval/pages/RouteAdminPage.tsx` — route admin
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

Rejecting any required signoff bumps the document back to `draft` and notifies the author.

**Gap D4 (smoke test 2026-05-01):** The actual implementation does NOT transition the document back to `draft`. `RecordSignoff` at `decision_service.go:284-295` only marks the approval instance as `rejected` (`InstanceRejected = true`) and updates the stage/instance statuses. The document row's `status` column remains `under_review`. The prose above describes the intended behaviour, not the current one.

## Known implementation gaps (as of 2026-05-01)

### D4 — Rejection does not return document to draft

**File:** `internal/modules/documents_v2/approval/application/decision_service.go:284`

The `QuorumRejectedStage` branch in `RecordSignoff` sets `result.InstanceRejected = true` and marks the approval instance + stage as rejected, but never issues an `UPDATE documents SET status = 'draft'`. The document remains in `under_review` indefinitely after a rejection. Callers must issue the status transition separately or via a DB patch.

**Workaround (smoke testing):**

```sql
UPDATE documents SET status = 'draft' WHERE id = '<doc_id>';
```

---

### E4 — Re-signoff on completed instance returns 500 instead of 409

**File:** `internal/modules/documents_v2/approval/http/errors.go:174`

When a user attempts a second signoff on an already-approved document, the domain returns an error containing `"no active stage in this approval instance"`. The HTTP error mapper at `errors.go:174` calls `looksLikeValidationError`, which only matches substrings `" is required"`, `" must be "`, and `" must not be "`. The phrase `"no active stage"` does not match, so the error falls through to the 500 path and is returned as `internal.unknown`. The correct HTTP response should be `409 state.instance_completed`.

**Fix needed:** Either add `"no active stage"` (or a canonical sentinel error) to `looksLikeValidationError`, or switch to typed error matching in the HTTP layer.

---

### Outbox idempotency_key bug — PDFDispatcher silently no-ops on repeated dispatches

**File:** `internal/modules/render/fanout/pdf_dispatcher.go:27`

`PDFDispatcher.Dispatch` publishes a `messaging.Event` with no `IdempotencyKey` field set. If the platform messaging layer deduplicates by `IdempotencyKey` and the first call produced an empty key, all subsequent dispatches for different documents may be treated as duplicates and silently dropped. Symptom: freeze succeeds but no PDF is ever generated for documents after the first one.

**Fix needed:** Populate `IdempotencyKey` in `Dispatch`, e.g. `in.RevisionID` or a deterministic hash of `(tenant_id, revision_id)`.

---

### Nova Revisão fails with `ux_documents_v2_cd_revision` unique constraint

**File:** `internal/modules/documents_v2/repository/repository.go:44`

`POST /api/v2/documents` calls `CreateDocument`, whose INSERT at line 44 does not include `revision_number`. The column defaults to `1`. If a controlled document already has a document at `revision_number = 1` (even a rejected or archived one), the unique index `ux_documents_v2_cd_revision ON documents(controlled_document_id, revision_number)` (migration `0131`) blocks the insert with a constraint violation. The service does not compute `MAX(revision_number) + 1`.

**Workaround (smoke testing):** Use a fresh controlled document per test run, or manually set `revision_number` via direct DB insert:

```sql
INSERT INTO documents (tenant_id, template_version_id, name, status, form_data_json,
                       created_by, controlled_document_id, profile_code_snapshot,
                       process_area_code_snapshot, code, revision_number)
VALUES (..., 2);
```

## See also

- [workflows/user-onboarding.md](../workflows/user-onboarding.md) — Step 8
- [workflows/approval.md](../workflows/approval.md)
- [concepts/iso-segregation.md](../concepts/iso-segregation.md)
- [workflows/freeze-and-fanout.md](../workflows/freeze-and-fanout.md)
