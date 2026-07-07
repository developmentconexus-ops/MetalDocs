# Feature F5 — evidence.md

Program: `approval-remediation` · Milestone: `M2b approval-kernel-backend` · Feature: `F5 freeze-boundary`
ADR: `wiki/decisions/0076-approval-freeze-boundary-and-choke-point-concurrency.md`

## 1. Implementation summary (by directory)

### `internal/modules/documents/approval/application/`
- **`freeze.go`** (new) — `executeFreeze(ctx, tx, repo, tenantID, instance)`: idempotent freeze
  executor. No-op if `instance.FrozenContentHash != nil`. Checks
  `HasUnresolvedInstanceComments(..., instance.SubmittedAt)` → blocks with
  `ErrFreezeBlockedByUnresolvedComments` if a comment was created at/after submit time. Else pins
  `instance.ContentHashAtSubmit` via `PinFrozenHash` (CAS). Lost-race CAS (`won==false`) is treated as
  success, not error.
- **`markup_gate.go`** (new) — `ScanForUnresolvedMarkup(docxBytes []byte) error`. Unzips buffer,
  reads `word/document.xml`, token-scans for `w:ins`/`w:del`/`w:pPrChange` (tracked changes) and
  `w:commentRangeStart`/`w:commentReference` (comment marks), matched on `xml.Name.Local`
  (namespace-agnostic). Returns `ErrUnresolvedTrackedChanges` / `ErrUnstrippedComments` /
  `ErrInvalidDocxBuffer`. **Standalone — no live call site in this feature** (bounded defer, see §4).
- **`review_verdict_service.go`** (edited) — `RecordVerdict`'s `ready` path: captures
  `completingStageKind` before `AdvanceStage()`; after advancing, if the completed stage was
  `StageKindReview` and the next active stage is nil or `StageKindApproval`, calls `executeFreeze`
  before any instance-status handling. The old `HasUnresolvedComments` gate (F4-introduced) is
  removed.
- **`submit_service.go`** (edited) — `SubmitRevisionForReview`: after `InsertStageInstances`, if the
  route has no `StageKindReview` stage anywhere, calls `executeFreeze` on the in-memory `inst` (using
  its already-set `SubmittedAt`/`ContentHashAtSubmit`).
- **`decision_service.go`** (edited) — `RecordSignoff`'s final-approve branch: removed the
  `HasUnresolvedComments` gate (W10). No new freeze call site added (approval-kind stages only ever
  activate post-freeze; see ADR 0076 §1).
- **`freeze_test.go`** (new, TDD) — 6 unit tests against a fake repo: idempotent re-entry, blocked by
  unresolved comments, comments-check error propagation, clean-path pin, lost-race-still-success,
  pin-error propagation.
- **`markup_gate_test.go`** (new, TDD) — table-driven tests building real minimal docx zips in memory
  via `archive/zip`: clean body, `w:ins`, `w:del`, `w:pPrChange`, `w:commentRangeStart`,
  `w:commentReference`, missing `word/document.xml`, corrupt non-zip bytes.
- **`decision_otel_test.go`**, **`phase5_integration_test.go`**, **`submit_service_test.go`**,
  **`decision_service_test.go`**, **`decision_service_freeze_test.go`** (all edited) — fake-repo
  interface-satisfaction fixes (3 nil-pointer-panic fixes) + 2 stale-assertion test rewrites (see §5).

### `internal/modules/documents/approval/infrastructure/`
- **`approval_repository.go`** (interface, edited) — added `PinFrozenHash` and
  `HasUnresolvedInstanceComments` to `ApprovalRepository`.
- **`postgres_approval_repository.go`** (impl, edited) — `PinFrozenHash`: CAS `UPDATE ... WHERE
  id=$1 AND tenant_id=$2 AND frozen_content_hash IS NULL`, `RowsAffected()` determines `won`.
  `HasUnresolvedInstanceComments`: `HasUnresolvedComments`'s query plus `AND created_at >= $3`.

### `tests/integration/approval/`
- **`freeze_integration_test.go`** (new, `testdb` factory, build tag `integration`) — 5 tests against
  real Postgres schema shapes (compiled/vetted, not run — bounded defer, see §4):
  review→approval transition pins hash; pre-instance comment does not block; during-instance comment
  blocks (whole tx rolled back, hash stays NULL); approval-only route freezes at submit; idempotent
  re-entry via direct `PinFrozenHash` call leaves hash unchanged.

### `scripts/api-lint/`
- **`tripwire-allowlist.txt`** (edited) — added
  `internal/modules/documents/approval/infrastructure/postgres_approval_repository.go|PinFrozenHash`,
  same justification class as the file's 5 pre-existing entries (authz enforced one layer up, in the
  tx-owning caller).

### `wiki/decisions/`
- **`0076-approval-freeze-boundary-and-choke-point-concurrency.md`** (new) — freeze boundary design,
  W10 fix, submit-time-hash-reuse rationale, markup-gate bounded defer, choke-point concurrency reuse
  (`ErrStaleRevision`), tripwire allow-list rationale. Alternatives Considered table with 6 rejected
  options.

No new migration — `frozen_content_hash` (`approval_instances`) already existed from F1/migration
`0286`.

## 2. Verification table

| # | Command | Result |
|---|---|---|
| 1 | `go build ./...` | clean, no output |
| 2 | `go build -tags integration ./...` | clean, no output |
| 3 | `go vet -tags integration ./tests/integration/approval/...` | clean, no output |
| 4 | `go test -count=1 ./internal/modules/documents/approval/...` | `ok` × 8 packages (application 3.9s, domain 1.09s, http 4.79s, http/contracts 1.26s, infrastructure 2.55s, infrastructure/idempotency 3.18s, infrastructure/signature 2.41s, jobs 3.20s); `api` package has no test files |
| 5 | `go test -count=1 ./...` | zero `FAIL` across entire repo (grep-zero confirmed) |
| 6 | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)` |
| 7 | `go test ./scripts/api-lint/...` | `ok` (7.33s) |
| 8 | grep `ErrApprovalBlockedByUnresolvedComments` across approval module | only declaration (`decision_service.go:30`), doc comments, and dead HTTP error-mapping code (`http/errors.go:276`, `http/errors_test.go`) remain — zero live production call sites return it (task 7 grep-zero check, PASS) |

## 3. Bounded defers

| # | Defer | Why bounded | Trigger / rerun |
|---|---|---|---|
| 1 | `freeze_integration_test.go` not executed against a live Postgres in this session | Same class as F4's defer — no live DB available in this sandboxed session (standing constraint: no `docker inspect`/`printenv` on DB credentials). File compiles and vets clean under `-tags integration`. | Rerun: `go test -tags integration -run TestFreeze -v ./tests/integration/approval/...` against a running local Postgres (`.\scripts\start-api.ps1` stack or equivalent test DB). Owner: next session with live DB access, before this feature is considered fully closed at the milestone level. |
| 2 | Markup gate (`ScanForUnresolvedMarkup`) built and unit-tested but not wired to any live call site | No existing Go port reads real docx bytes (`document_revisions.storage_key`) in-process — autosave is client→MinIO direct via presigned URL, backend never parses the bytes. M2c (the first feature that would *produce* unresolved tracked-changes/comment markup) does not exist yet. Wiring now would invent an unconfirmed-shape code path against a non-existent producer. Documented as a permanent architecture decision in ADR 0076 §4, not just here. | Trigger: M2c design, or a dedicated task to add a `DocumentContentReader` port for the approval module. Owner: M2c's implementer. |
| 3 | `HasUnresolvedComments` (document-wide, pre-existing) left in place with zero remaining production call sites | Cheap to keep; deleting speculatively risks removing a method another bounded-context caller might reference outside this feature's visibility (not confirmed absent via full-repo grep beyond the approval module's own tests/http mapping). | Trigger: a future cleanup pass with full cross-module call-site confirmation. Owner: unassigned, low priority. |
| 4 | Freeze pins `instance.ContentHashAtSubmit` rather than a hash computed over "clean post-review content" | Same missing-byte-path problem as defer #2 — no capability exists yet to read/hash real docx bytes at the freeze moment. Superset of current behavior (value now provably pinned, previously never verified), not a regression. | Trigger: same as #2 (M2c). Owner: M2c's implementer, revisit alongside the markup-gate wiring. |

## 4. Review disposition (self-review — judgment calls)

- **No dual-purpose columns/fields.** `frozen_content_hash` is written by exactly one path
  (`PinFrozenHash`, called only from `executeFreeze`) and read nowhere new in this feature. No field
  reuse across unrelated semantics found.
- **No silently-swallowed errors.** `executeFreeze` wraps every error from its two repo calls with
  `fmt.Errorf(...: %w)` — nothing is discarded. The one "error-shaped" outcome that is intentionally
  treated as success (`PinFrozenHash` returning `won==false`) is documented inline in `freeze.go` and
  in ADR 0076 §5 as a deliberate idempotency/lost-race semantic, not a swallowed error — `err` from
  that call is still checked and propagated separately.
- **Fake-repo panic-on-embed pattern (3 fixes).** `phase5Repo`, `fakeSubmitRepo`, and
  `stubApprovalRepo` all embed a nil `infrastructure.ApprovalRepository` as a catch-all; widening the
  interface with `PinFrozenHash`/`HasUnresolvedInstanceComments` made previously-dormant embed
  fallthroughs reachable, causing nil-pointer panics in `TestPhase5_FullApprovalAndPublish` and
  `TestSubmitRevisionForReview_HappyPath`. Fixed by adding real (non-panicking) overrides on exactly
  those two fakes, matching each test's existing query-driven-override style. Left `fakeSubmitRepo`'s
  pre-existing deliberately-panicking `HasUnresolvedComments` stub untouched — it remains genuinely
  unreachable in that test's scenario (same reasoning applied, not blanket-patched).
- **Two stale assertion tests rewritten, not silently changed.** `TestRecordSignoff_
  FinalApprovalBlockedByUnresolvedComments` and `TestRecordSignoff_UnresolvedComments_
  RollsBackBeforeApprove` asserted the OLD (now-removed) blocking behavior. Per plan.md task 6's
  explicit regression-test instruction, both were renamed
  (`TestRecordSignoff_FinalApprove_NoLongerGatedByUnresolvedComments`,
  `TestRecordSignoff_UnresolvedComments_NoLongerBlocksApprove`) and rewritten to assert the correct
  post-W10 behavior, with doc comments citing the W10/F5 change so a future reader isn't confused by
  the name flip.
- **api-lint allow-list addition is a real, reviewed false positive**, not a suppressed real
  violation — same shape as 5 pre-existing entries in the same file, confirmed via both
  `-strict` (0 violations) and the lint's own test suite (`ok`).
- **No adapter/hack reusing another type's shape.** `freezeRepo` (the narrow interface `executeFreeze`
  depends on) is a genuinely new, minimal 2-method interface — not a cast or reuse of an unrelated
  type. The pre-existing "PinInvoker"/"freeze" naming in `decision_service_freeze_test.go` (ADR
  0015's async PDF-materialization pinning) is a coincidental term reuse, confirmed unrelated to F5's
  content-freeze concept — no code shares state between the two.

## 5. Commands run (this session, chronological tail)

```
go vet ./internal/modules/documents/approval/...          # decision_otel_test.go interface fix
go test ./internal/modules/documents/approval/application/...  # freeze_test.go RED then GREEN
go build ./...
go build -tags integration ./...
go vet -tags integration ./tests/integration/approval/...
go test -count=1 ./internal/modules/documents/approval/...
go test -count=1 ./...
go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .
go test ./scripts/api-lint/...
```

All green. See §2 for the final consolidated re-run of the same sweep.

## 6. Scope confirmation

Touched paths (verified via `git status --porcelain`, scoped `git add` only — no `git add -A`):

```
internal/modules/documents/approval/application/decision_otel_test.go
internal/modules/documents/approval/application/decision_service.go
internal/modules/documents/approval/application/decision_service_freeze_test.go
internal/modules/documents/approval/application/decision_service_test.go
internal/modules/documents/approval/application/freeze.go
internal/modules/documents/approval/application/freeze_test.go
internal/modules/documents/approval/application/markup_gate.go
internal/modules/documents/approval/application/markup_gate_test.go
internal/modules/documents/approval/application/phase5_integration_test.go
internal/modules/documents/approval/application/review_verdict_service.go
internal/modules/documents/approval/application/submit_service.go
internal/modules/documents/approval/application/submit_service_test.go
internal/modules/documents/approval/infrastructure/approval_repository.go
internal/modules/documents/approval/infrastructure/postgres_approval_repository.go
scripts/api-lint/tripwire-allowlist.txt
tests/integration/approval/freeze_integration_test.go
wiki/decisions/0076-approval-freeze-boundary-and-choke-point-concurrency.md
docs/superpowers/milestones/approval-remediation/milestone-2b-approval-kernel-backend/f5-freeze-boundary/evidence.md
docs/superpowers/milestones/approval-remediation/milestone-2b-approval-kernel-backend/f5-freeze-boundary/plan.md
docs/superpowers/milestones/approval-remediation/milestone-2b-approval-kernel-backend/f5-freeze-boundary/spec.md
```

Unrelated pending repo state (pre-existing deletions in `.claude/PRPs/`, untracked
`docs/release/`, `frontend/apps/web/PRODUCT.md`, `scratch_qa/`) is explicitly **not** staged.
