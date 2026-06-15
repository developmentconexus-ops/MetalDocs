# Feature F3.5 — Plan (the "how")

> **Milestone:** 3 · **Feature:** `f3.5-deleteobject-error-log`. Contract in `spec.md`.
> Execution: subagent-driven — fresh implementer (TDD) → spec-compliance reviewer → code-quality
> reviewer; fix by root-cause family until both pass.

## Boundary

One site: `internal/modules/documents/application/service.go:534` (orphan cleanup on hash mismatch in
`CommitAutosave`). Observability only — no behavior change, no signature change, no contract change.

## Steps

1. **Confirm the site set (fail-closed).** `grep .DeleteObject(` across `internal/modules` → exactly
   one swallow (`service.go:534`); the milestone's `:740`/`:331` are stale (file is 710 lines; the
   others aren't DeleteObject). Documented in `spec.md` under the milestone's "or documented" clause.
   → verify: single hit, broadened pattern confirms no other storage-cleanup swallow.
2. **TDD (red first).** Add `TestCommitAutosave_LogsCleanupFailureOnHashMismatch`
   (`application_test`): install a capturing `slog` handler over a buffer (restore via `t.Cleanup`);
   drive `CommitAutosave` with `hashReturn:"WRONG"` (forces mismatch→delete) + `deleteErr` (forces the
   WARN); assert the call still returns `ErrContentHashMismatch`, `deleteCalls==1`, and the buffer
   contains every logged attribute (`level=WARN`, `storage_key`+value, `document_id`+value,
   `err`+value). → verify: red against the blank-assign (empty buffer).
3. **Impl.** Replace `_ = s.presigner.DeleteObject(...)` with an `if err != nil` guard emitting
   `slog.WarnContext(ctx, "commit autosave: orphaned object cleanup failed after content-hash mismatch",
   "storage_key", …, "document_id", …, "err", err)`; add `"log/slog"`. Keep the
   `return nil, domain.ErrContentHashMismatch`. → verify: test green.
4. **Verify boundary.** `go build ./...`; `go test ./internal/modules/documents/... -count=1`; redocly
   lint clean; grep shows the call now inside the guard.
5. **Two-stage review.** Spec-compliance (PASS) → code-quality (APPROVE). Code-quality IMPORTANT
   (test didn't assert `document_id`/`err` keys) folded back in step 2's want-list and re-verified.

## Risk / rollback

Single-line behavior-preserving guard + one new test. Rollback = revert the two edits. No migration,
no contract, no deploy coordination.
