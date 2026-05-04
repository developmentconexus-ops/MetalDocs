# Workflow: Approval

> **Last verified:** 2026-05-04
> **Scope:** Submit → route assignment → signoffs → approval condition met → freeze trigger.
> **Out of scope:** Freeze pipeline (see `workflows/freeze-and-fanout.md`), route admin (see `modules/approval.md`).
> **Key files:**
> - `internal/modules/documents/delivery/http/handler.go:259` — `finalizeDocument` — atomic finalize+submit handler
> - `internal/modules/documents/approval/application/submit_service.go:140` — `resolveEligibleActors` call inside stage-instance loop
> - `internal/modules/documents/approval/application/submit_service.go:299` — `resolveEligibleActors` implementation (queries `metaldocs.user_process_areas`)
> - `internal/modules/documents/approval/http/doc_approval_handler.go:51` — `SignoffByDocumentHandler` with idempotency replay
> - `internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go:1` — `PostgresSignoffIdempStore`

## Quick summary

1. Author clicks **Finalizar** in the editor → `POST /api/v2/documents/{id}/finalize` fires.
2. Handler resolves the active approval route for the document's profile, then calls `SubmitRevisionForReview` in a single transaction: document status moves `draft → under_review` **and** an `approval_instance` + `approval_stage_instances` rows are created atomically.
3. Each approver opens **Caixa de Entrada de Aprovação** and signs off (with password confirm).
4. ISO segregation blocks the submitter from approving their own version.
5. When the route's quorum condition is met, the version moves to `approved` and freeze fires automatically (same DB transaction).

See [workflows/user-onboarding.md](user-onboarding.md) for the full step-by-step click path.

## Finalize → submit atomicity (fixed 2026-05-02)

Previously `POST /api/v2/documents/{id}/finalize` only updated the document status to `under_review`. No approval instance was created, so the inbox was always empty after finalize.

Now `finalizeDocument` at `handler.go:259`:

1. Reads `revision_version` and `controlled_document_id` from the document row (errors with 409 if not in `draft`).
2. Reads `profile_code` from `controlled_documents`.
3. Queries `approval_routes` for the most-recent active route for that profile (errors with 409 if none exists).
4. Calls `SubmitRevisionForReview` — this single transaction creates the approval instance, stage instances, and transitions the document to `under_review`.
5. Returns HTTP 201 with `{"instanceId": "<uuid>"}`.

`NewHandlerWithSubmit` at `handler.go:73` is required to wire `db` + `submitSvc`; if only `NewHandler` is used the legacy status-only path runs as fallback.

## eligible_actor_ids populated at submit time (fixed 2026-05-02)

**File:** `internal/modules/documents/approval/application/submit_service.go:299`

`resolveEligibleActors` now queries `metaldocs.user_process_areas` to find all users holding `required_role` in `area_code` as of `now()`. The result is stored in `approval_stage_instances.eligible_actor_ids` so the inbox filter can match the calling approver's user ID.

No DB patch is needed for new submissions. Existing stage instances created before this fix still have `[]` and may need manual patching (see historic workaround below) or re-submission.

**Historic workaround (pre-fix submissions only):**

```sql
-- Find the stage instance
SELECT asi.id, asi.eligible_actor_ids, asi.status
FROM approval_stage_instances asi
JOIN approval_instances ai ON ai.id = asi.approval_instance_id
WHERE ai.document_id = '<document_id>'
ORDER BY asi.stage_order;

-- Patch in the approver's user ID
UPDATE approval_stage_instances
SET eligible_actor_ids = '["<approver-user-id>"]'::jsonb
WHERE id = '<stage_id>';
```

## Signoff idempotency (fixed 2026-05-02)

**Files:**
- `internal/modules/documents/approval/http/doc_approval_handler.go:51` — `SignoffByDocumentHandler`
- `internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go:1` — `PostgresSignoffIdempStore`

`SignoffByDocumentHandler` now requires an `Idempotency-Key` request header. Before calling the domain:

1. Checks `idempotency_keys` table via `PostgresSignoffIdempStore.CheckReplay`.
2. If a completed record exists, returns HTTP 200 with `{"was_replay": true, "outcome": "<prior outcome>"}` — **no duplicate domain call**.
3. On fresh signoff: records the outcome in `idempotency_keys` (keyed by `(tenant_id, actor_user_id, route_template, key)`; expires 24 h).

Previously a duplicate POST after instance close returned HTTP 500. Now it returns `was_replay:true`.

`NewHandler` in `internal/modules/documents/approval/http/handler.go:65` accepts `signoffIdempStore` as a positional parameter. Pass `nil` to disable idempotency (not recommended outside tests).

Migration `0160` grants `metaldocs_app` SELECT/INSERT/UPDATE on `metaldocs.idempotency_keys`. Without this migration the store returns a Postgres permission error.

## Idempotency

- Signoff: keyed by `Idempotency-Key` header + `(tenant_id, actor_user_id, route_template)`. Replays return `was_replay:true`.
- Re-clicking Finalizar: `SubmitRevisionForReview` fails with `ErrDuplicateSubmission` if an in-progress instance already exists for the document.

## Reject path (fixed 2977ef96 — B4)

- Any rejection in any required step → document returns to `draft` within the same DB transaction (cancel GUC + status UPDATE). The approval instance retains `rejected` for audit.
- Author can immediately open the document, edit, and call finalize again → new approval round (prior signoff history preserved on the old instance).
- Author notification mechanism TBD (email? in-app?).

**File:** `internal/modules/documents/approval/application/decision_service.go:284` — `QuorumRejectedStage` branch.

## Edge cases (TBD)

- Approver leaves org mid-route — reassignment policy.
- Route changes while document is `under_review` — does the in-flight approval use the old or new route? (Likely snapshot — verify.)
- Concurrent signoffs on the same step — last-write-wins or first-write-wins?

## See also

- [modules/approval.md](../modules/approval.md)
- [concepts/iso-segregation.md](../concepts/iso-segregation.md)
- [concepts/error-ux.md](../concepts/error-ux.md) — SoD error dialog states (E2), finalize toast (E3), global 401 auth-bus (E4)
- [workflows/freeze-and-fanout.md](freeze-and-fanout.md)
