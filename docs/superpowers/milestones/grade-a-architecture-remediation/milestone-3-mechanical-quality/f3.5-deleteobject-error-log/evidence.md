# Feature F3.5 — Evidence

> **Milestone:** 3 · **Feature:** `f3.5-deleteobject-error-log` · **Closed:** 2026-06-14
> **Contract:** `spec.md` (consumer = operator / log pipeline; structured WARN via existing `slog`
> convention; behavior-preserving).

## What was implemented

`internal/modules/documents/application/service.go` — in `Service.CommitAutosave`, the
content-hash-mismatch orphan-cleanup error (previously `_ = s.presigner.DeleteObject(...)`) is now
surfaced as a structured WARN and the cleanup stays best-effort:

```go
if err := s.presigner.DeleteObject(ctx, meta.StorageKey); err != nil {
    slog.WarnContext(ctx, "commit autosave: orphaned object cleanup failed after content-hash mismatch",
        "storage_key", meta.StorageKey, "document_id", cmd.DocumentID, "err", err)
}
return nil, domain.ErrContentHashMismatch
```

Added `"log/slog"`. Matches the in-module convention (`context_builder.go:63`). No `Service` struct
change, no signature change, no other site touched.

## Site reconciliation (milestone named 3, tree has 1)

The milestone row named `:537/:740/:331-334`. Verified against the current tree: **one** `.DeleteObject(`
call exists in all of `internal/modules` (non-test) — `service.go:534`. The file is **710 lines** (so
`:740` cannot exist); `:331` is not a DeleteObject site. Broadened search
`_ = .*\.(DeleteObject|AdoptTempObject|Delete)\(` returns the same single hit. Resolved under the
milestone's documented contingency ("any site found not to apply is documented") — **not an HS-6 full
stop**.

## Verification

| Gate | Command | Result (real output) |
|------|---------|----------------------|
| G1 (TDD red→green) | `go test ./internal/modules/documents/application/ -run TestCommitAutosave -count=1` | **RED before impl:** `service_test.go:554: expected WARN log output, got empty buffer` → `FAIL`. **GREEN after impl:** `ok metaldocs/internal/modules/documents/application 1.992s`. Test asserts the WARN carries every promised attribute (`level=WARN`, `storage_key`+value, `document_id`+`doc_1`, `err`+`s3 down`). |
| G2 (best-effort preserved) | same test | asserts `errors.Is(err, ErrContentHashMismatch)` (return unchanged) and `deleteCalls == 1` (cleanup still attempted, failure NOT promoted to a returned error). |
| G3 (single site, now guarded) | `grep -nE '\.DeleteObject\(' internal/modules/documents/application/service.go` | `535: if err := s.presigner.DeleteObject(ctx, meta.StorageKey); err != nil {` — the only call, now inside the guard. |
| G4 (build + module tests) | `go build ./...`; `go test ./internal/modules/documents/... -count=1` | build clean; all documents packages `ok`. |
| G5 (no contract drift) | `npx @redocly/cli lint api/openapi/v1/openapi.yaml` | `Your API description is valid. 🎉` (no spec edit). |

Real vs fixture: in-process `fakePresigner`/`fakeRepo` (no S3, no DB) — labeled fixture. The WARN
emission is real production `slog` output captured via a test `slog.TextHandler` over a buffer
(restored with `t.Cleanup`, no global-logger leak).

## Acceptance vs spec Validation Gate

All gates (G1–G5) met. No behavior change → **HS-2 did not trip** (observability-only).

## Review disposition (two-stage, subagent-driven)

- **Spec-compliance reviewer → PASS.** Exactly one site; still returns `ErrContentHashMismatch`; no
  `Service` logger field; keys `storage_key`/`document_id`/`err` present; no scope creep. Flagged the
  test didn't assert `document_id`/`err`.
- **Code-quality reviewer → APPROVE.** Zero critical. IMPORTANT: tighten the test want-list to cover
  `document_id` + `err` (a regression dropping either would otherwise pass). **Fixed** — the want-list
  now asserts all four attributes + values; re-ran G1/G4/G5 green. Confirmed the `slog.SetDefault`
  global-state mutation is properly restored via `t.Cleanup` (no parallel leak).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `fakePresigner.deleteCalls` is a non-atomic `int` (code-quality MINOR). | Single-goroutine test; no concurrency on this counter today. | If a concurrent `CommitAutosave` test is ever added, guard the counter; owner: whoever adds it. |

## Memory / cross-refs

Observability hardening on the documents commit path. Related:
`[[backend-target-architecture-governs-reviews]]`.
