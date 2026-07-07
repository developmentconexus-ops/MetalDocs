# Feature F7 — Evidence

> **Milestone:** 2b — Approval Kernel Backend  ·  **Feature:** `f7-signature-meaning-sod-unification`
> **Closed:** 2026-07-07
> **Contract:** `spec.md` / `plan.md`

## What was implemented

### SoD unification (W7)

- `internal/modules/documents/approval/application/review_verdict_service.go`: replaced the F4-era
  inline self-verdict check (`if instance.SubmittedBy == req.ActorUserID { return
  domain.ErrAuthorCannotSign }`) with a call to the SAME shared predicate `decision_service.go`
  already calls: `domain.CheckSoD(instance.SubmittedBy, req.ActorUserID, nil)`. `priorSignoffs` is
  `nil` because the cross-stage-reuse clause needs a `[]Signoff`-shaped prior-record source that
  doesn't apply to review verdicts (different table, own `UNIQUE(stage_instance_id, actor_user_id)`
  + replay/conflict handling already covers same-stage re-recording) — spec.md Interview #2.
- `internal/modules/documents/approval/domain/sod.go`: doc comment on `CheckSoD` updated to state
  explicitly it is now the single shared SoD predicate called by both `DecisionService.RecordSignoff`
  and `ReviewVerdictService.RecordVerdict`. No signature change, no rename (`CheckSoD` kept as-is —
  spec.md Interview #4: renaming to plan.md's illustrative `ViolatesSoD(authorID, actorID,
  stageKind)` would be a pure-refactor, out-of-proportion blast radius for no functional gain, and
  `stageKind` would be dead weight since the rule text is stage-kind-agnostic).
- `db/migrations/0290_sod_unified_trigger.sql` (new): extracts `enforce_signoff_sod()`'s exact rule
  body into a shared, table-agnostic `enforce_approval_sod()` function (verbatim rule text/error
  message/ERRCODE, verified against `db/baseline/0001_current_schema.sql:682-701`), re-points the
  existing `trg_signoff_sod` trigger on `approval_signoffs` to it (zero behavior change), and adds a
  NEW `trg_review_verdict_sod BEFORE INSERT ON approval_review_verdicts` trigger — closing the
  previously-missing DB-tripwire-last enforcement gap on the F4 table (spec.md Interview #3: prior to
  this migration, `approval_review_verdicts` had zero DB-level SoD enforcement; only the app-level
  check existed). Idempotent: `CREATE OR REPLACE FUNCTION`, `DROP TRIGGER IF EXISTS` before `CREATE
  TRIGGER`. Ledger row inserted into `schema_migrations` (`'0290'`, `ON CONFLICT DO NOTHING`).
- `tests/integration/approval/review_verdict_integration_test.go`: added
  `TestReviewVerdictSoDTrigger_RejectsAuthorSelfInsert` — direct SQL `INSERT` into
  `approval_review_verdicts` with `actor_user_id` = document author, asserting the new
  `trg_review_verdict_sod` trigger rejects it with the mirrored error message (`"SoD: author cannot
  sign own revision"`). Added `"strings"` import for the message-substring assertion.

### Signature meaning (W8, 21 CFR 11.50(a)(3))

- `internal/modules/documents/approval/application/decision_service.go`: `RecordSignoff` now derives
  `signatureMeaning` from `req.Decision` immediately before constructing the domain `Signoff`
  (`DecisionApprove → "approval"`, `DecisionReject → "rejection"`), passed as
  `SignatureMeaning: signatureMeaning` in `domain.SignoffParams{}`. This closes the one missing wire
  identified in spec.md Interview #5: every other layer (domain validation/getter, repository
  insert/read paths) already round-tripped `signature_meaning` correctly — but `RecordSignoff` never
  set it, so `NewSignoff`'s empty-field default (`"approval"`) silently mis-attested every REJECT
  signoff as an approval. This was a live, pre-existing 21 CFR Part 11 defect, not a hypothetical gap.
- `internal/modules/documents/approval/http/contracts/instance_read.go`: `SignoffRecord` gains
  `SignatureMeaning string \`json:"signature_meaning"\`` — the live, hand-maintained DTO actually
  returned by `GetInstanceHandler`/`GetInstanceByDocumentHandler` (confirmed via spec.md Interview #6:
  the openapi-generated `ApprovalSignoffRecordResponse` schema has zero live Go call sites
  constructing it — a pre-existing generated-vs-hand-rolled split, out of scope to reconcile here).
- `internal/modules/documents/approval/http/get_instance_handler.go`: `mapInstanceResponse`'s
  per-signoff loop adds `SignatureMeaning: sig.SignatureMeaning()` — the manifest now renders the
  actual attestation meaning per spec.md §2.1 ("rendered in the signature manifest and audit trail").
- `api/openapi/v1/openapi.yaml`: `ApprovalSignoffRecordResponse` schema gains `signature_meaning:
  {type: string, enum: [approval, rejection]}` (now in `required`), for contract-surface completeness
  even though this specific endpoint's live handler doesn't construct the generated type (Interview
  #6/#7 — server-derived only, no client-writable request field added).
- `internal/modules/documents/approval/api/api.gen.go` regenerated via `go generate
  ./internal/modules/documents/approval/api/...` — new `ApprovalSignoffRecordResponseSignatureMeaning`
  enum type + `.Valid()`, new struct field, embedded base64 spec blob re-encoded (expected churn, not
  hand-edited).

### Tests

- `internal/modules/documents/approval/application/decision_service_test.go`: added
  `TestRecordSignoff_ApproveSetsSignatureMeaningApproval` and
  `TestRecordSignoff_RejectSetsSignatureMeaningRejection`, both asserting
  `repo.insertedSignoff.SignatureMeaning()` against the fake repo's captured insert. The reject test
  is the TDD red-state proof: it failed pre-fix with `SignatureMeaning() = "approval"; want
  "rejection"`, confirming the real defect, then passed after the `decision_service.go` fix.
- `internal/modules/documents/approval/http/get_instance_handler_test.go`: added
  `TestMapInstanceResponse_RendersSignatureMeaning` — builds one approve signoff and one reject
  signoff, maps them through `mapInstanceResponse`, asserts the JSON-facing
  `contracts.SignoffRecord.SignatureMeaning` matches by signoff ID for both. Red-state proof: first
  run failed to compile (`rec.SignatureMeaning undefined`) before the DTO field was added.

## Verification

| Check | Command | Result |
|-------|---------|--------|
| Build | `go build ./...` | clean, exit 0 |
| Integration-tag build | `go build -tags integration ./...` | clean, exit 0 |
| Approval module suite | `go test -count=1 ./internal/modules/documents/approval/...` | all subpackages PASS (application/domain/http/http-contracts/infrastructure/idempotency/signature/jobs) |
| Full regression | `go test -count=1 ./...` | all packages PASS — zero `FAIL` lines |
| api-lint strict | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)` |
| api-lint unit suite | `go test ./scripts/api-lint/...` | PASS |
| Grep-zero (spec's exact Validation Gate wording) | `grep -rn "instance.SubmittedBy ==" --include=*.go internal/modules/documents/approval` | zero hits |
| Grep-zero (plan.md's stated pattern) | `grep -rn "SubmittedBy == req.ActorUserID" --include=*.go .` | zero hits |
| Both SoD call sites confirmed unified | `grep -rn "CheckSoD" --include=*.go internal/modules/documents/approval` | exactly 2 call sites (`decision_service.go:301`, `review_verdict_service.go:153`), both calling `domain.CheckSoD`; no other production hits |
| New integration test build/vet | `go vet -tags integration ./tests/integration/approval/...` | clean, exit 0 |
| New integration test **live run** | — | **not run** — see Bounded defers (no `DATABASE_URL` obtainable without reading `.env`, forbidden; same precedent as F3-F6) |
| Migration 0290 SQL hand-review | compared against `db/baseline/0001_current_schema.sql:682-701` (`enforce_signoff_sod`) | verbatim rule-text/error-message/ERRCODE match; new trigger attaches the identical function to `approval_review_verdicts` |

## Judgment calls

1. **Kept `CheckSoD`'s name/signature, did not rename to `ViolatesSoD` or add `stageKind`** — spec.md
   Interview #4: plan.md's naming is illustrative; a rename would break `errors.Is`-based call sites
   and existing tests for zero functional gain. The "single exported predicate" requirement is
   satisfied by `CheckSoD` already being that predicate — the defect was a missing call site, not a
   naming mismatch.
2. **Passed `nil` for `priorSignoffs` from the verdict call site** rather than fabricating a
   `[]Signoff` adapter slice from `ReviewVerdict` rows — spec.md Interview #2. `CheckSoD`'s loop over
   `priorSignoffs` is a no-op on `nil`; the cross-stage-reuse rule doesn't apply to verdicts per
   spec.md §5, and `InsertVerdict`'s existing conflict handling already covers same-stage re-recording.
3. **Updated the openapi-generated `ApprovalSignoffRecordResponse` schema for contract-surface
   completeness, but wired `signature_meaning` into the actually-live `contracts.SignoffRecord` DTO**
   — spec.md Interview #6. These are two different types; only the hand-maintained one is read by any
   client today. The pre-existing generated-vs-hand-rolled split for this endpoint predates F7 and is
   explicitly out of scope to reconcile (see Non-goals / Bounded defers below).
4. **`enforce_approval_sod()` left the old `enforce_signoff_sod()` function in place, un-dropped** —
   avoids any risk of an object-not-found error on stale references (pg_dump snapshots, migration
   re-apply); it is simply no longer referenced by any trigger after 0290 runs.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|--------------------------|
| Live-DB run of migration 0290 and the new/updated integration tests (`TestReviewVerdictSoDTrigger_RejectsAuthorSelfInsert`, plus a re-run of F4's existing `TestReviewVerdict_SoDBlocksSelfVerdict` to confirm the call-site change didn't regress it) | No `DATABASE_URL`/`METALDOCS_DATABASE_URL` obtainable without reading `.env` (forbidden by standing rule); identical precedent to F3-F6's own defers | Run `.\scripts\start-api.ps1` (or the project's normal PowerShell local-dev bootstrap) then `go test -tags integration -run 'SoD' -v ./tests/integration/approval/...`. Owner: next session with authorized local Postgres access. |
| Generated-vs-hand-rolled DTO split for the approval-instance read endpoints (`ApprovalSignoffRecordResponse` vs `contracts.SignoffRecord`) | Pre-existing architectural split, out of F7's file-list scope (spec.md Non-goals) | Trigger: a future dedicated cleanup task reconciling the read-path DTO strategy for `GET /approval/instances/{id}` and `GET /documents/{id}/approval-instance`. Owner: unassigned, low priority. |

## Self-review (per task instructions)

- **Grep-zero on old duplicate inline SoD pattern**: confirmed twice, both spec.md's literal wording
  and plan.md's — zero hits outside `domain/sod.go`.
- **Column/field reused across two unrelated semantic purposes** (F4's caught-bug class): checked —
  `signature_meaning` is a single-purpose column (attestation meaning only); the new
  `enforce_approval_sod()` function reuses `approval_instance_id` (already present on both tables for
  its existing FK purpose) purely as a lookup key, not a repurposed semantic field. No reuse found.
- **Silently swallowed error path**: checked `enforce_approval_sod()` — a missing
  instance/document row yields `author_id = NULL`, and `NEW.actor_user_id = NULL` is never true in
  SQL, so the check fails open on a missing-row case exactly like the pre-existing
  `enforce_signoff_sod()` did (verbatim-copied behavior, not a new risk). Checked `signatureMeaning`
  derivation in `decision_service.go` — two exhaustive branches over the two-value `Decision` enum,
  no default/swallow; `NewSignoff` itself still validates the value and errors on anything outside
  `{"approval","rejection"}`.

## Acceptance vs spec Validation Gate

| Gate item | Met? | Evidence |
|-----------|------|----------|
| `review_verdict_service.go` calls `domain.CheckSoD`, not inline comparison | yes | code change + grep-zero above |
| Grep-zero: no duplicate inline SoD logic anywhere in the approval module | yes | both grep commands above, zero hits |
| One shared SQL function backs SoD enforcement on both tables | yes (schema) / **not live-verified** | migration 0290 written + hand-reviewed against baseline; new integration test written/vetted but not run against Postgres (see defer) |
| `RecordSignoff` derives `signature_meaning` from `Decision`, never defaults silently | yes | `decision_service.go` fix + both new unit tests PASS (reject test proved the defect pre-fix) |
| `signature_meaning` rendered in the read-path manifest | yes | `TestMapInstanceResponse_RendersSignatureMeaning` PASS |
| openapi `ApprovalSignoffRecordResponse` gains `signature_meaning` enum; regen clean | yes | schema edit + `go generate` regen + `go build ./...` clean |
| No regression | yes | full `go test -count=1 ./...` PASS, api-lint strict 0 violations |
