# Feature F6 — `no-fallback-hash-chain` — plan.md

Ref: `spec.md` (this folder), spec §11 locked no-fallback principle, plan.md `docs/superpowers/plans/2026-07-07-m2b-approval-kernel-backend.md` F6 section.

## Task 1 — Repository: rename + no-fallback `LoadFrozenContentHash`

- [ ] Edit `internal/modules/documents/approval/infrastructure/approval_repository.go`: rename interface method `LoadActiveDocumentContentHash(ctx, tx, tenantID, documentID string) (string, error)` → `LoadFrozenContentHash(ctx, tx, tenantID, instanceID string) (string, error)`; update doc comment.
- [ ] Edit `internal/modules/documents/approval/infrastructure/postgres_approval_repository.go`: rename impl, replace body with `SELECT frozen_content_hash FROM approval_instances WHERE id=$1 AND tenant_id=$2`; NULL or `sql.ErrNoRows` → `ErrNoActiveContentHash`; update doc comment (no more "mirrors the COALESCE").
- [ ] Update `infrastructure/errors.go`'s `ErrNoActiveContentHash` doc comment to describe the new source.
- [ ] Write failing unit test first (TDD) in a new/extended fake-repo test proving the SQL string contains no `COALESCE` and no `document_revisions` reference (string-inspection harness, mirrors F5's `phase5_integration_test.go` style query capture) — or, simpler given existing patterns, a fake-DB row test asserting NULL → `ErrNoActiveContentHash`.

## Task 2 — decision_service.go call-site update

- [ ] Edit `decision_service.go`: change `s.repo.LoadActiveDocumentContentHash(ctx, tx, req.TenantID, instance.DocumentID)` → `s.repo.LoadFrozenContentHash(ctx, tx, req.TenantID, instance.ID)`. Update surrounding comment (currently describes the old COALESCE-mirroring rationale) to describe the new frozen-pin-only rationale.
- [ ] Update every fake repo implementing the OLD method name (`decision_service_test.go`, `phase5_integration_test.go`, `submit_service_test.go`, `decision_otel_test.go`) to implement the NEW method name/signature. `fakeSubmitRepo`'s panic-stub stays a panic stub (submit tests never call this method — same as today). `phase5Repo`/`fakeDecisionRepo`/`stubApprovalRepo` need real behavior matching their existing test scenarios (frozen hash present → matches echoed hash; adjust any fixture that previously relied on `documents.content_hash_at_submit` value to instead set `frozen_content_hash` in the fake's returned instance / fake DB row).
- [ ] Write/extend failing test: `TestRecordSignoff_NullFrozenHash_FailsClosed` (or extend an existing decision_service_test.go table) — instance with `FrozenContentHash == nil` → `RecordSignoff` returns `ErrContentHashMismatch`, never a head-hash-based match.

## Task 3 — publish_service.go fail-closed pin check

- [ ] Edit `publish_service.go`'s `PublishApproved`: immediately after the `instance.Status != domain.InstanceApproved` check, add `if instance.FrozenContentHash == nil { return infrastructure.ErrNoActiveContentHash }`.
- [ ] Confirm `domain.Instance` exposes `FrozenContentHash *string` already (F1) — reuse, no new field.
- [ ] Write failing test `TestPublishApproved_NullFrozenHash_FailsClosed` in `publish_service_test.go`: fake/instance with `Status=InstanceApproved`, `FrozenContentHash=nil` → expect `ErrNoActiveContentHash`, document status unchanged (no UPDATE issued — assert via fake repo call tracking or a spy `TxRunner`).
- [ ] Check `http/errors.go` maps `infrastructure.ErrNoActiveContentHash` (or its usual wrap) to a sane RFC 9457 response for the publish handler's error path — `publish_handler.go`'s existing error mapping likely already funnels through the shared `MapErrorToResponse`/`errors.go` table; confirm `errors.Is(err, infrastructure.ErrNoActiveContentHash)` is reachable there (today only reachable via `application.ErrContentHashMismatch` from decision_service — publish_service returning the raw infra sentinel needs its own `case` OR needs to be wrapped as `application.ErrContentHashMismatch` for consistency). **Decision (make now, document in evidence.md if it deviates from this default):** wrap as `application.ErrContentHashMismatch` at the publish_service.go call site (matches decision_service.go's existing pattern exactly, avoids adding a new case to `http/errors.go` for a sentinel that already has a home) rather than returning the infra sentinel raw.

## Task 4 — active_instance_reader.go split query

- [ ] Edit `internal/modules/documents/infrastructure/active_instance_reader.go`: remove the `COALESCE(active.content_hash_at_submit, (SELECT ...))` expression and the `active.content_hash_at_submit` selection from the main FULL OUTER JOIN query; drop `content_hash_at_submit` from the `active` subquery's selected columns too (no longer needed there).
- [ ] After the FULL OUTER JOIN resolves `view.DocumentID`/`view.Status`, extend the EXISTING `under_review` in-progress-instance query (the one populating `view.ApprovalInstanceID`) to also `SELECT id::text, frozen_content_hash FROM approval_instances WHERE ...` (add the one column) so both values come from the single existing round-trip.
- [ ] If `frozen_content_hash` from that row is non-NULL, set `view.ContentHash` from it.
- [ ] Else (no under_review row, or row's `frozen_content_hash` is NULL, or status isn't `under_review` at all), run a second explicit query: `SELECT content_hash FROM document_revisions WHERE document_id=$1 ORDER BY created_at DESC LIMIT 1` scoped to `*view.DocumentID` — only when `view.DocumentID != nil`.
- [ ] Write failing unit/integration tests first: extend or create a test file covering the 3-4 state matrix (draft/no-instance, under_review+frozen, under_review+not-yet-frozen, published-only/no active doc) before making the edit.

## Task 5 — Parity test update

- [ ] Edit `tests/integration/controlleddocuments/active_instance_parity_test.go`: update `rawGetActiveInstance` to replay the new two-step logic (mirrors Task 4's SQL exactly — this is a raw-SQL parity baseline, not calling the port).
- [ ] Add a new sub-test `under_review_with_frozen_instance` (freeze the seeded instance via testdb factory helper or direct `PinFrozenHash`-equivalent SQL, per existing `testdb.NewApprovalInstance`/`testdb.WithStatus` helpers) asserting `ContentHash` equals the pinned hash, not the head revision hash.
- [ ] Add a sub-test (or extend `under_review_with_in_progress_approval`) for the not-yet-frozen case to confirm it still resolves to head-revision hash (regression guard for the correct fallthrough, not a re-introduced defect).

## Task 6 — Full verification sweep

- [ ] `go build ./...`
- [ ] `go build -tags integration ./...`
- [ ] `go test -count=1 ./internal/modules/documents/approval/...`
- [ ] `go test -count=1 ./internal/modules/documents/...`
- [ ] `go test -count=1 ./...` (grep zero FAIL)
- [ ] `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` (0 violations — no contract change expected, this is a regression check)
- [ ] `go test ./scripts/api-lint/...`
- [ ] `grep -rn "COALESCE(d.content_hash_at_submit\|COALESCE(active.content_hash_at_submit" internal/` → zero hits

## Task 7 — Self-review pass

- [ ] Re-grep for any remaining COALESCE-style fallback in the same read path (Interview #7 already did a first pass; repeat post-edit).
- [ ] Confirm no error path silently swallowed (every new `if ... == nil { return Err... }` is a real return, not a logged-and-continue).
- [ ] Confirm no adapter reuses `documents.content_hash_at_submit` for a second unrelated purpose anywhere in the diff (F4's caught-bug class).

## Task 8 — Evidence + commit

- [ ] Write `evidence.md` (implementation summary, verification table, judgment calls, bounded defers, scope confirmation).
- [ ] `git add` explicit touched paths only (never `-A`).
- [ ] Commit: `feat(approval): F6 no-fallback hash chain — frozen pin only, fail closed (W9, locked principle)`.
