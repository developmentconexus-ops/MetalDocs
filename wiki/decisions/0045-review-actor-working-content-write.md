# ADR 0045 — Review-Actor Working-Content Write Boundary (single-writer by status + eligible-approver eligibility)

> **Status:** Accepted — decision implemented in Task 5 (Eigenpal cockpit service gate). Tied to ADR 0022 (capability coherence) and ADR 0037 (membership temporal model).
> **Scope:** Authorization for document revisions authored during the review phase. Establishes the governance rule that allows an eligible approver to persist tracking-change suggestions into the working-content revision chain while preserving the single-writer invariant via status-gated transitions.
> **Out of scope:** Authentication / sessions (tier-0); the approval state machine (ADR 0018); reject→draft state transitions; the wider publish/freeze/materialize lifecycle (ADR 0015).
> **Key files:**
> - `internal/modules/documents/application/autosave_service.go` — tier-2 `CanWriteWorkingContent` gate + service composition
> - `internal/modules/documents/delivery/http/routes_documents.go` — session autosave + checkpoint routes
> - `internal/modules/iam/domain/model.go:50` — capability registry
> - `apps/api/cmd/metaldocs-api/permissions.go:82` — tier-1 route table

## Context

The Eigenpal cockpit mounts the editor in suggesting mode, allowing the assigned approver of the active approval stage to write tracking-change suggestions (review comments, strikethrough edits) directly into the document. These suggestions must persist into the working-content `document_revisions` chain.

The autosave and checkpoint write paths previously gated all document mutations with a hard `status == draft` check — author-only writes. Bringing the approver into the write surface requires a boundary rule: an approver may write working-content revisions only while the document is in `under_review` status and only if they hold an active eligibility in the current approval stage.

The single-writer invariant (one actor writing at a time) is preserved not by exclusive locks but by status gating: the author cannot write while `under_review` (rejected in a tier-2 check), and the approver cannot write while `draft` (rejected in a tier-2 check). Transitions between draft and under_review are the document's state machine; neither the author nor the approver can author a revision during a transition.

A new revision is not minted on every autosave or checkpoint. The `revision_number` advances only on governance events (submit for approval, signoff, publish) and rejection, following the existing `document_revisions` lifecycle. Approver edits are working-content changes to the existing in-flight revision.

### Finding — eligibility bounds the authority

From ADR 0022 Decision 4: administration is area-scoped and backed by active membership. For approvals, the "area" is the active approval stage's governance: eligibility reads from `approval_stage_instances.eligible_actor_ids` (set at stage entry, immutable during the stage). An eligible approver is one whose `user_id` appears in that snapshot.

Reading eligibility must not hold any lock or participate in the document's authz transaction (per `advisory-lock-deadlock-constraint` in memory). The eligibility is a snapshot fact — it can be read outside the write tx that checks it.

## Decision

Working-content writes are governed by a new `CanWriteWorkingContent(status, isOwner, isEligibleApprover)` gate:

| `status` | `isOwner` | `isEligibleApprover` | **Decision** |
|---|---|---|---|
| `draft` | true | false | ✓ ALLOW (author) |
| `draft` | true | true | ✓ ALLOW (author) |
| `draft` | false | false | ✗ DENY |
| `draft` | false | true | ✗ DENY |
| `under_review` | true | false | ✗ DENY (author cannot write while under review) |
| `under_review` | true | true | ✗ DENY (owner-edge case: author is never eligible approver) |
| `under_review` | false | true | ✓ ALLOW (eligible approver of the active stage) |
| `under_review` | false | false | ✗ DENY |
| `rejected` | true | false | ✓ ALLOW (author, back in draft-like mode) |
| `rejected` | true | true | ✓ ALLOW (author) |
| `rejected` | false | false | ✗ DENY |
| `rejected` | false | true | ✗ DENY |
| `published`, `superseded`, `obsoleted` | any | any | ✗ DENY (immutable) |

**Expressed in code:**
- draft or rejected → isOwner
- under_review → (not isOwner) AND isEligibleApprover
- otherwise → deny

**No new `Revision` is minted.** The `revision_number` remains unchanged. An autosave or checkpoint by an approver advances `document_revisions.content_hash` and timestamps, but leaves `revision_number` on the in-flight revision. The revision counter advances only on governance events (submit, signoff, publish, reject) — the existing behavior.

**Single-writer is preserved by status.** The author and approver never write concurrently because the document is either `draft` (author-writable, approver denied) or `under_review` (author denied, approver-writable) or `published`/`superseded`/`obsoleted` (both denied, immutable). Transitions are the state machine's responsibility.

## Consequences

**Wins.**
- Approver suggestions (tracking changes) persist in the document history (`document_revisions`) and are available for author review and export.
- Single-writer invariant is preserved without exclusive locks — the status enum is the boundary.
- Eligibility is a snapshot fact of the approval-stage governance, decoupled from the write transaction (no deadlock risk per `advisory-lock-deadlock-constraint`).
- Aligns with ADR 0022 Decision 4: the stage's `eligible_actor_ids` is the area/scope authority, not an ad-hoc owner check.

**Costs.**
- Approver writes can race with state transitions (e.g., an approver writes a suggestion just before rejection is recorded). The write is lost, and the user is shown an error on autosave return. Retry is the user-facing recovery; no automatic merge or conflict resolution. **Acceptable because:** the window is tiny (microseconds around rejection), the common case is the approver completing their review *before* signoff, and the Eigenpal UI surfaces the error (not a silent drop).
- Eligibility reads are not part of the gated write transaction. The eligibility snapshot can grow stale by the time the write executes (someone was added to `eligible_actor_ids` and later removed, but their write still succeeds). **Mitigated by:** eligibility is immutable during a stage (the snapshot is set at stage entry and checked only at stage-boundary transitions to the next stage). Within-stage eligibility volatility is out of scope (a future product decision).

## Implementation — gate mechanics

### Tier-1 route table
- `PATCH /api/v1/documents/{id}/autosave` → `CapDocumentEdit` (existing, no change)
- `POST /api/v1/documents/{id}/checkpoints` → `CapDocumentEdit` (existing, no change)

Both routes gate on the capability and the document area (tier-2). The new boundary is inside the service, not the route.

### Tier-2 `CanWriteWorkingContent`
**Function signature:**
```go
CanWriteWorkingContent(ctx context.Context, tx *sql.Tx, docID uuid.UUID, status DocumentStatus, isOwner bool, isEligibleApprover bool) error
```

**Call site:** `autosave_service.go` RecordAutosave + RecordCheckpoint, after loading the document but before writing the revision.

**Eligibility resolution (outside the write tx):**
```go
eligibleActors, err := approvalService.ActiveStageEligibleActors(ctx, docID) // off-tx read
isEligible := contains(eligibleActors, authn.UserIDFromContext(ctx))
```

The approvalService provides a read-only method that fetches `approval_stage_instances.eligible_actor_ids` for the active stage of the document. This read is **not** part of the autosave transaction.

**Gate logic (inside `CanWriteWorkingContent`):**
```go
switch status {
case Draft, Rejected:
    if !isOwner {
        return authz.ErrCapDenied{Cap: CapDocumentEdit}
    }
case UnderReview:
    if isOwner || !isEligibleApprover {
        return authz.ErrCapDenied{Cap: CapDocumentEdit}
    }
case Published, Superseded, Obsoleted:
    return authz.ErrCapDenied{Cap: CapDocumentEdit}
}
return nil
```

**Consequence for the revision:** the `revision_number` is not incremented. The autosave or checkpoint writes `document_revisions.content_hash`, timestamps, and tracking changes to the existing in-flight revision row.

## Alternatives considered

| Option | Verdict | Reason |
|---|---|---|
| Approver writes create a new revision | Rejected | Breaks the contract that `revision_number` advances only on governance events; complicates the author's review (which version is "the" submission?); rejected-then-resubmitted documents would have orphaned approver revisions. |
| Lock-based mutual exclusion (row lock on document, approver waits for author release) | Rejected | Expensive (row lock holds during the edit session); deadlock risk if both actors are in flight (author autosaving while approver's tx waits); violates `advisory-lock-deadlock-constraint`. |
| Eligibility read inside the autosave transaction | Rejected | Adds a second table (`approval_stage_instances`) to the write tx; increases deadlock surface; eligibility is immutable during a stage, so reading off-tx is safe and simpler. |
| Author can edit while under_review (collaborative mode) | Rejected | Violates single-writer invariant; author and approver could clobber each other; outside the current approval contract (sequential review). |

## Amendment — Authorization framework alignment (2026-06-23)

This ADR builds on ADR 0022 (capability coherence), ADR 0037 (membership temporal model — eligibility is an active-time fact), and the documented `advisory-lock-deadlock-constraint` in the project memory. The decision to read eligibility off-transaction follows H-PRE-1 (prevent circular dependencies in authorization reads, per ADR 0029/0031 pattern) and reuses the existing `approval_stage_instances.eligible_actor_ids` field, avoiding a new scope table or bit-indexed grant matrix.

No new capability is minted (reuses `CapDocumentEdit` from tier-1). No new schema is required. The write boundary is purely behavioral — a service-layer gate on existing document status and approval-stage data.

> **Current reality (2026-06-23):** The Eigenpal cockpit service gate is implemented in Task 5. The `CanWriteWorkingContent` mechanics are in `autosave_service.go` and wired into the session route handlers. Tier-2 eligibility checks are proven off-transaction per the `advisory-lock-deadlock-constraint` discipline.
