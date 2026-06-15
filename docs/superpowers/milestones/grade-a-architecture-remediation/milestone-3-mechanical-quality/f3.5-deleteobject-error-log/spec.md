# Feature F3.5 — Surface silently-discarded DeleteObject errors as WARN logs

> **Milestone:** 3 · **Feature:** `f3.5-deleteobject-error-log`
> **Skill:** `metaldocs-backend-api` · **Approach:** wrap the swallowed cleanup error in a structured
> `slog.WarnContext` (observability only, no behavior change).

## Site reconciliation (milestone said 3 sites; current tree has 1)

The milestone row named three sites in documents `service.go` (`:537`, `:740`, `:331-334`) and
explicitly required "confirm each during the feature (or any site found not to apply is documented)".
Verification against the current tree (2026-06-14):

- **`.DeleteObject(` call sites in all of `internal/modules` (non-test):** exactly **one** —
  `application/service.go:534`, `_ = s.presigner.DeleteObject(ctx, meta.StorageKey)` (orphan cleanup
  after a content-hash mismatch in `CommitAutosave`).
- **`:740`** — the file is **710 lines**; that line does not exist. Stale.
- **`:331-334`** — not a `DeleteObject` site in the current tree. Stale.
- Broadened search `_ = .*\.(DeleteObject|AdoptTempObject|Delete)\(` across `internal/modules`
  (non-test) returns the **same single** hit.

Conclusion: the "three sites" was stale audit data (prior waves already removed/relocated the others).
**One real swallow site remains.** This is within the milestone's documented contingency — **not an
HS-6 full stop** (the spec authored the "or documented" escape valve); recorded here per that clause.

## Consumer contract (read the consumer first)

The **consumer** of this change is the operator / log-aggregation pipeline that reads backend logs.
Contract for the one site:

- On the orphan-cleanup failure path (content-hash mismatch → best-effort `DeleteObject` fails), emit
  **one** structured WARN record via the module's existing logging convention:
  `slog.WarnContext(ctx, "<message>", "storage_key", …, "document_id", …, "err", err)`.
  - Convention is the package-level `log/slog` already used in this module
    (`documents/application/context_builder.go:63`:
    `slog.WarnContext(ctx, "approval instance lookup failed", "revision_id", …, "err", err)`).
  - Keys: at minimum `storage_key` (what failed to delete) and `err`; include `document_id` for
    correlation. Lower-snake_case keys, matching the existing `slog` call sites.
- **Behavior is unchanged:** `CommitAutosave` still returns `domain.ErrContentHashMismatch` on this
  path; the cleanup remains best-effort (a failed delete does **not** become a returned error — the
  caller already failed for a different reason, the object is orphaned regardless). The WARN only adds
  observability so orphaned objects are diagnosable.

## What to implement

In `internal/modules/documents/application/service.go`, replace the swallow at `:534`:

```go
_ = s.presigner.DeleteObject(ctx, meta.StorageKey)
```
with:
```go
if err := s.presigner.DeleteObject(ctx, meta.StorageKey); err != nil {
    slog.WarnContext(ctx, "commit autosave: orphaned object cleanup failed after content-hash mismatch",
        "storage_key", meta.StorageKey, "document_id", cmd.DocumentID, "err", err)
}
```
Add the `"log/slog"` import. No other site, no signature change, no new field on `Service`.

## Non-goals

- No behavior change beyond the log line: `CommitAutosave` still returns `ErrContentHashMismatch`; the
  delete stays best-effort (failure is **not** promoted to a returned error).
- No logger field added to `Service` (the package-level `slog` convention is already used here).
- No change to the other two stale-named sites (they don't exist).
- No new log lines on happy paths or other error paths.
- No OpenAPI/contract/FE change.

## Validation Gate

| # | Acceptance | Proof |
|---|------------|-------|
| G1 | TDD: drive `CommitAutosave` with a hash mismatch **and** a failing `DeleteObject`; assert a single WARN record is emitted carrying `storage_key` + `err`, and that the call still returns `ErrContentHashMismatch` (behavior unchanged). Red before impl (no record captured), green after. | `go test ./internal/modules/documents/application/ -run TestCommitAutosave -count=1` |
| G2 | Best-effort preserved: a failing `DeleteObject` does **not** change the returned error or the `deleteCalls` count vs today. | assertion in the new test |
| G3 | No other swallow site introduced/changed; `grep` still shows the single `DeleteObject` call now inside an `if err != nil`. | `grep -nE '\.DeleteObject\(' service.go` |
| G4 | Build + full documents module tests green. | `go build ./...`; `go test ./internal/modules/documents/... -count=1` |
| G5 | No contract drift. | `npx @redocly/cli lint api/openapi/v1/openapi.yaml` clean (no spec edit) |

## Interview record

No operator interview required — the contract (structured WARN, best-effort preserved) is fully
determined by the existing in-module `slog` convention and the non-goal that behavior must not change.
The only ambiguity (three named sites vs one real site) was resolved by code verification under the
milestone's own "or documented" clause, not by guessing.

## Approval

Approved for implementation: 2026-06-14 (consumer contract = structured WARN via existing `slog`
convention; behavior-preserving; one verified site).
