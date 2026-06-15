# Micro-task — Test-Infra Rebaseline (e2e_seed) — Evidence

> **Milestone:** 1 (Reach-A Blockers) · **Type:** bounded micro-task (operator HS-1 condition on M1 close, option 2)
> **Closed:** 2026-06-14 · **Boundary:** `internal/test/e2e_seed.go` ONLY (+ throwaway proof scripts under `scripts/tmp/`)

## Why this micro-task

M1 was approved with one condition: before opening M2, discharge the F1.3 **AC5
full-HTTP-E2E defer** by repairing post-v1-rebaseline drift in the e2e HTTP seed so a
real `seed → finalize → signoff` workflow runs and the
`approval_signoffs.actor_display_name_snapshot` can be read back end-to-end. Scope was
explicitly bounded to the **e2e_seed HTTP path** — the separate `tests/integration/testdb`
curated-bootstrap drift is a different harness and stays out of scope.

## Drift found and fixed (all in `internal/test/e2e_seed.go`)

The seed produced a `draft` document that real `finalize` could not advance. Three
post-rebaseline gaps, each surfaced by driving the live route and reading the failing
constraint, then fixed at the seed boundary:

1. **No linked controlled document.** `GetFinalizePrereqs` Step 2 reads `profile_code`
   from `controlled_documents` via `documents.controlled_document_id` (not the draft's
   `profile_code_snapshot`); an unlinked draft → `ErrProfileNotConfigured` (400). Fix:
   new `ensureControlledDocument` helper creates/reuses a CD (`profile_code`,
   `process_area_code`, `owner_user_id`, `status='active'`, `visibility_scope='company'`)
   keyed `(tenant_id, profile_code, 'E2E-DOC')`; `documents.controlled_document_id` now
   links it.
2. **Capability tripwire gap.** `controlled_documents` INSERT is guarded by
   `enforce_capability_asserted`, which requires `controlled_documents.create` in the
   tx-local `metaldocs.asserted_caps`. Fix: added `controlled_documents.create` to
   `e2eAssertedCaps`.
3. **Missing `under_review` snapshot columns.** `enforce_snapshot_on_submit` fires on the
   `draft → under_review` transition that `submit_service` performs, requiring six
   non-NULL columns (`placeholder_schema_snapshot`/`_hash`,
   `composition_config_snapshot`/`_hash`, `body_docx_snapshot_s3_key`, `body_docx_hash`).
   The submit UPDATE touches only status/title/version, so they must pre-exist on the
   draft. Fix: seed them on insert — jsonb snapshots `'{}'`, the three `_hash` columns as
   **32-byte** zero bytea (`decode(repeat('00',32),'hex')`, per `*_hash_len` CHECKs),
   `body_docx_snapshot_s3_key='seed/body.docx'`.

Also seeded `documents.content_hash_at_submit` to a deterministic, docID-derived 64-hex
(returned in the seed response as `content_hash`) so the signoff's
`LoadActiveDocumentContentHash` resolves to a value the proof can echo back; added a
`controlled_documents` cleanup row to the reset handler.

No production code changed (a temporary `slog.Error` added to `finalizeDocument` to
surface the unmapped 500 was reverted; `git diff` on `handler.go` is empty).

## Runtime proof

`scripts/tmp/e2e-relaunch-seed.ps1` (start e2e binary `:8081`, admin login, seed) then
`scripts/tmp/f13-e2e-proof.ps1` (login → seed → finalize as author → GET instance →
POST signoff as approver → DB read-back).

| Step | Result |
|------|--------|
| Build | `go build -tags integration -o metaldocs-api-e2e.exe ./apps/api/cmd/metaldocs-api/` → exit 0; `go build ./...` → exit 0 |
| Seed | `200`, returns users/cookies + `content_hash` |
| Finalize (author) | `POST /api/v1/documents/{id}/finalize` → **201** `{"instance_id":"3fb28610-…"}` |
| Get instance | `in_progress`, 2 stages (Review active, Approval pending), `etag "v1"` |
| Signoff (approver) | `POST /api/v1/approval/instances/{id}/stages/{stage}/signoffs` → **200** `{"outcome":"stage_completed"}` |
| **DB read-back** | `approval_signoffs.actor_display_name_snapshot = "E2E Approver"` == `iam_users.display_name = "E2E Approver"` → **`matches = t`** |

This is the real HTTP workflow chrome that F1.3 AC5 had bounded-deferred: the off-tx
display-name read is now proven to persist into the signoff snapshot through the live
finalize→signoff route under real Postgres/RLS.

## Disposition

- F1.3 AC5 defer (full HTTP E2E) — **discharged**.
- F1.3 test-infra-drift defer — **discharged for the e2e_seed HTTP path**; the distinct
  `tests/integration/testdb` curated-bootstrap drift remains an explicit, owned defer
  (`metaldocs-database`), unrelated to this proof path.
- M1 close condition (HS-1, option 2) satisfied → M2 may be opened **on operator approval**.

## Bounded defers (carried forward)

| Defer | Trigger / owner |
|-------|-----------------|
| `tests/integration/testdb` curated-bootstrap drift (`0001_product_reference_data.sql` vs current `templates_template.visibility` NOT NULL) | Before any tag-gated integration suite that uses the curated bootstrap. Owner: `metaldocs-database`. |
| HS-2 — FE eigenpal `file:` path defer | Before any FE `pnpm install` / M2 start. Owner: M2 prep. |
