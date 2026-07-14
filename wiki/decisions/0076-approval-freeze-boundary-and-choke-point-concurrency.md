# ADR 0076 — Approval Freeze Boundary + Choke-Point Concurrency

> **Status:** Accepted
> **Date:** 2026-07-07
> **Scope:** Content freeze boundary (review layer → approval layer), instance-scoped
> unresolved-comments gate, choke-point concurrency resolution, markup-gate bounded defer.
> **Milestone:** `docs/superpowers/milestones/approval-remediation/milestone-2b-approval-kernel-backend/f5-freeze-boundary/`
> **Key files:**
> - `internal/modules/documents/approval/application/freeze.go` — freeze executor
> - `internal/modules/documents/approval/application/markup_gate.go` — standalone markup scan
> - `internal/modules/documents/approval/application/review_verdict_service.go` — freeze call site 1
> - `internal/modules/documents/approval/application/submit_service.go` — freeze call site 2
> - `internal/modules/documents/approval/application/decision_service.go` — W10 gate removed
> - `internal/modules/documents/approval/infrastructure/approval_repository.go` — `PinFrozenHash`,
>   `HasUnresolvedInstanceComments`

## Context

The design spec (`docs/superpowers/specs/2026-07-07-approval-remediation-design.md` §2.2, "W2 core")
identified a gap: nothing in the approval module marks the point after which a document's content
becomes immutable for the remainder of its approval lifecycle. Two consequences followed from this
gap (W2/W9/W10):

1. **No freeze boundary.** A document could, in principle, keep accumulating edits through the
   approval-kind (signoff) stages with no recorded "this is the content that was actually approved"
   hash — `approval_instances.frozen_content_hash` existed as a column (F1, migration `0286`) but
   nothing ever wrote to it.
2. **W10: the unresolved-comments gate ran at the wrong moment, against the wrong scope.**
   `DecisionService.RecordSignoff` and `ReviewVerdictService.RecordVerdict` (F4) each independently
   called `HasUnresolvedComments` — a **document-wide** check with no time bound — immediately before
   collapsing the instance to `InstanceApproved`. A single unresolved comment from years ago, on an
   unrelated prior revision, would permanently block every future approval of that document. The
   check also ran at final-approve, not at the actual point the content stops changing (which, for
   multi-stage routes, is earlier — when the last *review*-kind stage completes, not when the last
   *approval*-kind stage does).

## Decision

### 1. Freeze fires at the review→approval boundary, not at final-approve

`executeFreeze(ctx, tx, repo, tenantID, instance)` (`application/freeze.go`) is the single freeze
executor. It is idempotent (`instance.FrozenContentHash != nil` → no-op), checks an
**instance-scoped** comment gate, and pins `instance.ContentHashAtSubmit` (the already-computed,
already-canonicalized hash from `ComputeContentHash` at submit time — see §3) via a CAS UPDATE.

Two call sites, both inside the existing stage-transition tx:

- **`ReviewVerdictService.RecordVerdict`** (the `ready` path): the completing stage's `Kind` is
  captured *before* `instance.AdvanceStage()` mutates the in-memory stage slice. After advancing, if
  the completing stage was `StageKindReview` **and** (there is no next active stage, or the next
  active stage is `StageKindApproval`), the transition just crossed the freeze boundary —
  `executeFreeze` runs before any further instance-status handling.
- **`SubmitService.SubmitRevisionForReview`**: for approval-only routes (no `StageKindReview` stage
  anywhere in the route), there is no review→approval transition to hook into — freeze fires at
  submit time instead, using the same in-memory `inst` value already built for `InsertInstance`.

`DecisionService.RecordSignoff` (final-approve, approval-kind stages) gets **no new freeze call
site**. By construction — review stages are ordered before approval stages in every route (`Route.
Validate()`) — an approval-kind stage can only ever become active after a `ready` verdict crossed the
freeze boundary via `RecordVerdict`'s stage-advance path, or the route had no review stage at all and
froze at submit. `RecordSignoff` only ever completes an *already-frozen* instance.

### 2. W10 fixed: instance-scoped comment gate, freeze is now the sole enforcement point

`HasUnresolvedInstanceComments(ctx, tx, tenantID, documentID, since time.Time)` — a new repository
method — scopes the same query `HasUnresolvedComments` ran, plus `AND created_at >= $3`, bound to
`instance.SubmittedAt`. `executeFreeze` calls it; a hit returns `ErrFreezeBlockedByUnresolvedComments`
(new sentinel, distinct from the pre-existing `ErrApprovalBlockedByUnresolvedComments`).

The old gate — `s.repo.HasUnresolvedComments(...)` immediately before `InstanceApproved` — is
**deleted from both call sites** that had it: `DecisionService.RecordSignoff` (line ~381) and
`ReviewVerdictService.RecordVerdict` (line ~219, introduced in the same milestone by F4). Both were
checked and fixed consistently rather than migrating one and leaving the other stale.
`HasUnresolvedComments` (the document-wide method) itself is left in place, unused by any production
call site after this change — a deliberate, documented non-deletion (§6).

### 3. Freeze pins the existing submit-time content hash, not a new docx-byte hash

The literal reading of "freeze the content" could mean hashing the actual clean docx bytes at the
moment of freeze. That capability does not exist in this codebase: `ComputeContentHash` hashes
canonicalized JSON form data, not OOXML bytes, and there is no live path for the approval module to
read `document_revisions.storage_key` bytes (see §4). `executeFreeze` instead re-pins
`instance.ContentHashAtSubmit` — the same canonicalization already computed and stored at submit
time. This is a strict superset of current production behavior (the value existed but was never
independently re-verified/frozen) without claiming a capability ("content changed mid-review, hash
reflects post-review state") the codebase cannot yet exercise. Upgradeable once a real mid-review-edit
path exists (M2c).

### 4. Markup gate ships standalone, not wired to a live call site (bounded defer)

`ScanForUnresolvedMarkup(docxBytes []byte) error` (`application/markup_gate.go`) unzips a docx buffer,
reads `word/document.xml`, and token-scans for `w:ins` / `w:del` / `w:pPrChange` (unresolved tracked
changes) or `w:commentRangeStart` / `w:commentReference` (unstripped comment marks), matching on
`xml.Name.Local` so it is namespace-prefix-agnostic. It is fully implemented and table-driven tested
against real in-memory docx zips built via `archive/zip` (no external fixtures).

It is **not called from `executeFreeze`** in this feature. Investigation (background research, this
session) found:

- Real docx bytes exist in MinIO (`document_revisions.storage_key`), confirmed via
  `internal/platform/storage/minio/store.go` and its use in Gotenberg PDF conversion.
- The write path is entirely client-driven (`PresignAutosave` → `CommitAutosave`): the FE PUTs bytes
  directly to MinIO via a presigned URL. The Go backend never reads or parses those bytes in-process
  — it only verifies a hash on confirm. No `DocumentContentReader`/`BlobReader`/content port exists
  for the approval module to read them.
- Zero source-level call sites for eigenpal's tracked-changes/comment-mark APIs
  (`extractTrackedChanges`, `acceptChangeById`, `removeCommentMark`) exist in
  `frontend/apps/web/src/**` — they appear only in the built `dist/` bundle. M2c, the feature that
  would ever *produce* a docx containing unresolved `w:ins`/`w:del`/comment marks, has not been built.

Wiring the gate to a live byte source now would mean inventing an unused code path against a shape
nobody can confirm yet. The gate ships real, tested, and ready — the future call site only needs to
supply bytes once M2c exists.

### 5. Choke-point concurrency: reuse the existing OCC/lock, no new primitive

The spec's "freeze race" scenario — two concurrent stage-transition attempts, one should win, the
other should get a typed 409 rather than silently overwriting — is already fully resolved by
machinery every stage transition in this module uses: `LoadInstance` takes a `FOR UPDATE` row lock on
the instance (and its stage rows) before any mutation, and every request carries
`ExpectedRevisionVersion` checked against `instance.RevisionVersion` (OCC). A concurrent attempt
blocks on the row lock, then — once the winner commits and bumps `revision_version` — fails its own
OCC check on read, surfacing as the existing `infrastructure.ErrStaleRevision` (mapped to 409). No new
concurrency primitive, lock, or error type is introduced.

`PinFrozenHash`'s CAS (`UPDATE ... WHERE frozen_content_hash IS NULL`) is defense-in-depth on top of
that, not the actual race-resolution mechanism — by the time it runs, the caller already holds the
row lock. A lost CAS (`won=false`) is treated as success, not an error: it means a concurrent freeze
already happened, which given the row lock is only reachable in exotic edge cases (e.g. an external
direct-SQL write bypassing the service layer entirely).

### 6. `PinFrozenHash` needs a tripwire allow-list entry, not a new authz call

`PinFrozenHash` is a single-purpose CAS UPDATE, always invoked from inside `executeFreeze`, always
inside a tx where an `authz.Require` call already ran earlier (in `RecordVerdict`/
`SubmitRevisionForReview`, before the freeze boundary is ever reached). This is the exact same shape
already allow-listed for `UpdateInstanceStatus` / `UpdateStageStatus` in
`scripts/api-lint/tripwire-allowlist.txt` — authz is enforced one layer up, which the tripwire-pairing
lint's single-file scan cannot see. `PinFrozenHash` is added to that same allow-list rather than
duplicating an authz check the tx already performed.

## Consequences

- **Positive:** the document is provably immutable from the freeze boundary onward, at the correct
  moment (last review stage, or submit for approval-only routes) instead of at final-approve.
- **Positive:** W10 is fixed at both call sites that had it (F4's `RecordVerdict` and the pre-existing
  `RecordSignoff`) — no drift between the two review-layer surfaces.
- **Positive:** no new concurrency primitive — the freeze race is resolved by mechanism already
  proven correct for every other stage transition in this module.
- **Neutral:** the markup gate is real and tested but unwired — a bounded defer with a written
  trigger (M2c, or a dedicated task confirming the upload-time content shape with the eigenpal
  integration owner), not a silent scope cut.
- **Neutral:** `HasUnresolvedComments` (document-wide) is left in place with no remaining production
  call site — not deleted in this feature (see Alternatives Considered).
- **Negative:** freeze re-pins the submit-time hash rather than a true "content as it stood right
  before freeze" hash; if content genuinely changes mid-review (once M2c exists), this will need
  revisiting — documented, not hidden.

## Alternatives Considered

| Option | Verdict | Reason |
|---|---|---|
| Wire the markup gate to a live MinIO byte-fetch in this feature | Rejected | No existing Go port reads `document_revisions.storage_key` in-process; would be an unused, unconfirmed-shape code path invented against a producer (M2c) that doesn't exist yet. |
| Hash real docx bytes at freeze time instead of reusing `ContentHashAtSubmit` | Rejected | Same missing-byte-path problem; also a new hashing concept alongside `ComputeContentHash`'s JSON canonicalization, adding a second hash semantics with no current consumer needing the distinction. |
| Add a freeze call site inside `DecisionService.RecordSignoff` for defense-in-depth | Rejected | By construction an approval-kind stage never activates before freeze already fired; adding a redundant call site would let a future refactor accidentally treat it as load-bearing instead of provably unreachable. |
| Delete `HasUnresolvedComments` (document-wide) now that it has no call site | Rejected (deferred) | Cheap to keep; a defensive delete-on-sight risks removing a method some other bounded-context caller might still reference outside this feature's visibility. Left as a documented bounded defer. |
| Add a new 409-mapped error type for the freeze race | Rejected | `infrastructure.ErrStaleRevision` already covers it by construction (row lock + OCC); a second error type for the same HTTP outcome would be redundant. |
| Add an explicit `authz.Require` call inside `PinFrozenHash` | Rejected | Would duplicate the check the enclosing tx already performed; the established pattern for this exact situation (`UpdateInstanceStatus`/`UpdateStageStatus`) is an allow-list entry, not a redundant in-repo check. |

## Rollback

Additive at the schema level (no new migration — `frozen_content_hash`/`cancel_reason` already exist
from F1/`0286`). Reversible by: reverting the two `executeFreeze` call sites (freeze stops firing),
restoring the two `HasUnresolvedComments` gates in `decision_service.go`/`review_verdict_service.go`,
and removing `PinFrozenHash`/`HasUnresolvedInstanceComments` from the repository interface — provided
no code has come to depend on `frozen_content_hash` being populated by then. `markup_gate.go` is
purely additive and has no call sites to roll back.

## References

- `docs/superpowers/specs/2026-07-07-approval-remediation-design.md` §2.2 ("Freeze boundary (W2
  core)")
- `docs/superpowers/milestones/approval-remediation/milestone-2b-approval-kernel-backend/f5-freeze-boundary/spec.md`
- `db/migrations/0286_approval_stage_kinds_expand.sql` (F1 — `frozen_content_hash`/`cancel_reason`
  columns, purely additive)
- `wiki/decisions/0075-approval-oversee-visibility.md` (structure template; F4/F3 milestone siblings)
