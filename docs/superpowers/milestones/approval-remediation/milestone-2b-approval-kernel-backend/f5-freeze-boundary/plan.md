# Feature F5 — plan.md

Base: `docs/superpowers/plans/2026-07-07-m2b-approval-kernel-backend.md` F5 section, corrected per
spec.md's Interview record:

- No new migration — `frozen_content_hash` (F1/`0286`) already exists on `approval_instances`; no
  migration number consumed by F5.
- Markup gate ships as a standalone, fully-tested pure function; it is NOT wired to a live docx-byte
  source in this feature (bounded defer — no existing Go port reads `document_revisions.storage_key`
  bytes, and M2c, the first producer of real tracked-changes/comment markup, doesn't exist yet).
- Freeze re-pins `instance.ContentHashAtSubmit` (existing `ComputeContentHash` JSON-canonicalization
  value) rather than a hash over "clean post-review content" — the latter needs the same not-yet-built
  docx-byte path (bounded defer).
- Only TWO new freeze call sites: `review_verdict_service.go` (`ready` path) and `submit_service.go`
  (approval-only routes). `decision_service.go` gets NO new freeze call site — by construction, an
  approval-kind stage only ever activates after freeze has already fired (review→approval transitions
  only ever happen via `RecordVerdict`'s stage-advance, never via `RecordSignoff`). `decision_service.go`
  only loses its `HasUnresolvedComments` call (W10).
- Freeze concurrency ("race: two concurrent advance attempts, one wins CAS") is resolved by the SAME
  `FOR UPDATE` row lock + `ExpectedRevisionVersion` OCC every stage transition in this module already
  uses — no new 409 error type; `infrastructure.ErrStaleRevision` already covers it.

## Tasks

1. **Domain/sentinels.** `domain/errors.go`: no new domain errors needed (freeze sentinels live in
   `application` package, matching where `ErrContentHashMismatch`/`ErrApprovalBlockedByUnresolvedComments`
   already live). `application/services.go` or a new small file: add
   `ErrFreezeBlockedByUnresolvedComments`, `ErrUnresolvedTrackedChanges`, `ErrUnstrippedComments`,
   `ErrInvalidDocxBuffer`.

2. **Markup gate (TDD, failing test first).** `application/markup_gate.go` +
   `application/markup_gate_test.go`:
   - `ScanForUnresolvedMarkup(docxBytes []byte) error`: `archive/zip.NewReader` over
     `bytes.NewReader(docxBytes)`, find `word/document.xml`, `encoding/xml.NewDecoder` token-scan for
     start-elements with `Name.Local` in `{ins, del, pPrChange, commentRangeStart, commentReference}`.
   - Table-driven tests build minimal valid docx zips programmatically (a zip containing a
     `word/document.xml` entry with hand-written minimal OOXML body — no external fixture files):
     clean body → nil error; body containing `<w:ins>...</w:ins>` → `ErrUnresolvedTrackedChanges`;
     `<w:del>` → same; `<w:pPrChange/>` → same; `<w:commentRangeStart w:id="1"/>` →
     `ErrUnstrippedComments`; `<w:commentReference w:id="1"/>` → same; zip missing
     `word/document.xml` → `ErrInvalidDocxBuffer`; corrupt (non-zip) bytes → `ErrInvalidDocxBuffer`.

3. **Repository additions.** `infrastructure/approval_repository.go` (interface) +
   `infrastructure/postgres_approval_repository.go` (impl):
   - `PinFrozenHash(ctx, tx, tenantID, instanceID, hash string) (won bool, err error)` — CAS UPDATE
     `WHERE id=$2 AND tenant_id=$3 AND frozen_content_hash IS NULL`; `RowsAffected()==1` → `won=true`;
     `==0` → `won=false, nil` (never an error — freeze idempotency is a first-class outcome, not an
     edge case).
   - `HasUnresolvedInstanceComments(ctx, tx, tenantID, documentID string, since time.Time) (bool, error)`
     — mirrors `HasUnresolvedComments` plus `AND created_at >= $3`.
   - `HasUnresolvedComments` (document-wide) stays untouched (bounded defer — not deleted).

4. **Freeze executor (TDD, failing test first).** `application/freeze.go` +
   `application/freeze_test.go` (unit, fake repo) + `tests/integration/approval/freeze_integration_test.go`
   (testdb factory, REQUIRED framework):
   - `type freezer struct { repo infrastructure.ApprovalRepository }` (or a package-level function
     taking `repo` as a param — match whichever style keeps `RecordVerdict`/`SubmitRevisionForReview`
     simplest to call from; prefer a plain function `executeFreeze(ctx, tx, repo, tenantID, instance
     *domain.Instance) error` over a new stateful service type, since freeze has no independent
     lifecycle/dependencies beyond `repo` — avoids a `Services.Freeze` field nobody else calls
     directly).
   - Logic: if `instance.FrozenContentHash != nil` → return nil (idempotent no-op). Else:
     `HasUnresolvedInstanceComments(..., instance.SubmittedAt)` → if true, return
     `ErrFreezeBlockedByUnresolvedComments`. Else: `PinFrozenHash(..., instance.ContentHashAtSubmit)`
     → on `won==true`, set `instance.FrozenContentHash = &instance.ContentHashAtSubmit` in-memory;
     on `won==false` (lost race to a concurrent freeze that beat us to it, extremely unlikely given
     the row lock but handled defensively), still treat as success (no-op) since SOME freeze
     happened.
   - Unit tests (fake repo): idempotent re-entry no-ops; unresolved instance-scoped comment blocks;
     clean path pins successfully.
   - Integration tests (real Postgres): freeze fires and persists `frozen_content_hash`; a
     pre-instance comment (created before `submitted_at`) does NOT block; a during-instance comment
     DOES block; two sequential `executeFreeze` calls are idempotent (hash unchanged, no error).

5. **Wire the two call sites.**
   - `review_verdict_service.go`: capture `activeStage.Kind` and `activeStage.ID` BEFORE calling
     `instance.AdvanceStage()` (which mutates the in-memory stage slice). After `AdvanceStage()`
     succeeds, if `capturedKind == domain.StageKindReview` AND (`instance.Active() == nil` OR
     `instance.Active().Kind == domain.StageKindApproval`) → call `executeFreeze(...)` BEFORE the
     existing `instance.Status == domain.InstanceApproved` branch's unresolved-comments check (which
     is being deleted in task 6) — freeze must run whether or not the instance is already
     `InstanceApproved`, since "next stage is approval-kind" can be true while the instance is still
     `InstanceInProgress` (next stage just activated).
   - `submit_service.go`: after `s.repo.InsertStageInstances(...)` succeeds, check whether every
     `route.Stages[i].Kind == domain.StageKindApproval` (no review stage anywhere in the route) →
     call `executeFreeze(ctx, tx, s.repo, req.TenantID, &inst)` (the in-memory `inst` built earlier in
     this function; `inst.SubmittedAt = now` and `inst.ContentHashAtSubmit = contentHash` are already
     set at this point in the function, satisfying freeze's inputs with no new reads).

6. **W10: remove the unresolved-comments gate from final-approve, both call sites.**
   - `decision_service.go:381-388`: delete the `s.repo.HasUnresolvedComments(...)` call and its
     `if blocked { return ErrApprovalBlockedByUnresolvedComments }` guard entirely — freeze (now
     fired earlier, from `review_verdict_service.go` or `submit_service.go`, never from
     `decision_service.go` itself per the "no new call site" finding) is the sole gate for this
     concern by the time any signoff completes an approval-kind stage.
   - `review_verdict_service.go:219-226`: same deletion — freeze already fired for this exact
     instance moments earlier in this SAME function call (task 5's hook runs before this point).
   - Regression test: an instance with a comment created BEFORE `submitted_at` (old document-wide
     check would have blocked it forever; new instance-scoped check correctly ignores it) reaches
     `InstanceApproved`/stage-completed successfully.

7. **Grep-zero / consistency check:** confirm no remaining call site references
   `ErrApprovalBlockedByUnresolvedComments` outside its own declaration and the now-updated tests (the
   error type itself can stay declared — cheap, matches the "don't delete unless forced" judgment
   from Interview #8 applied consistently — but zero live call sites should return it after this
   feature).

8. **ADR 0076.** `wiki/decisions/0076-approval-freeze-boundary-and-choke-point-concurrency.md` —
   Context/Decision/Consequences/Alternatives Considered/Rollback/References, mirroring ADR 0075's
   structure exactly. Must document: the freeze boundary design, the choke-point concurrency
   resolution (existing OCC/lock reused, no new primitive), AND the markup-gate bounded-defer decision
   (this is a real architecture call worth a permanent record, not just a milestone evidence.md note).

9. **Full verification sweep:** `go build ./...`; `go build -tags integration ./...`;
   `go vet -tags integration ./tests/integration/approval/...`; targeted approval-module suite
   (`-count=1`); full `go test -count=1 ./...` (grep-zero `FAIL`); both authz lints
   (`go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` → `0 violation(s)`);
   `go test ./scripts/api-lint/...`.

10. **evidence.md**, then commit `feat(approval): F5 freeze boundary — hash pin + instance-scoped
    comment gate + markup gate (standalone) (W2/W9/W10)`. NOT pushed.

## ADR note

ADR 0076 is F5's due ADR per the milestone's 4-ADR ledger (spec.md §9 item 2: "Content freeze boundary
+ review layer... choke-point concurrency"). Unlike F4 (which explicitly deferred this ADR to F5), F5
is where it gets written — the freeze-boundary decision (including the markup-gate defer) is now fully
settled.
