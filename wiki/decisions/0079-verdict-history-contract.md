# ADR 0079 — Verdict-History Contract + Immutable Actor-Name Snapshot (No Read Fallback, DB-Enforced)

> **Status:** Accepted
> **Date:** 2026-07-09
> **Scope:** The approval instance-by-document view DTO exposes a chronological `verdicts[]` history;
> the actor display name on every verdict and signoff is the value snapshotted at the moment of the
> signed action — immutable, DB-enforced NOT-NULL, no read-time fallback.
> **Milestone:** `docs/superpowers/milestones/approval-remediation/milestone-2d-workflow-coherence-fe/f2-verdicts-display-names/`
> **Governing spec:** `docs/superpowers/specs/2026-07-08-approval-workflow-coherence-design.md` §2 (D2), §8.1
> **Related:** ADR 0078 (viewer-facts contract — sibling read projection), ADR 0035 (generated DTOs),
> ADR 0029 (`UserDisplayNameReader` port — the write-path name source), ADR 0022 (capabilities)
> **Key files:**
> - `api/openapi/v1/openapi.yaml` — `ApprovalInstanceByDocumentResponse.verdicts`
> - `internal/modules/documents/approval/http/contracts/instance_read.go` — `VerdictRecord`
> - `internal/modules/documents/approval/http/get_instance_handler.go` — verdict projection
> - `internal/modules/documents/approval/application/read_service.go` — `LoadInstanceByDocumentForViewWithViewer` (extended to load verdicts)
> - `internal/modules/documents/approval/infrastructure/postgres_approval_repository.go` — `LoadInstanceVerdicts`, `InsertVerdict`, `InsertSignoff` (unconditional snapshot bind)
> - `db/migrations/0294_*.sql` — backfill + `SET NOT NULL` + `CHECK (<> '')` on both snapshot columns

## Context

M2d converges the single relationship question ("what is the viewer's standing on this instance?")
onto one server-derived contract. ADR 0078 delivered the *forward-looking* half (`viewer` block —
what the viewer may do NOW). This ADR delivers the *backward-looking* half: the **verdict history** —
who has already reviewed this instance, what they decided, and why — so the single FE screen (F2d.5)
renders a complete audit trail from one DTO with zero client-side derivation.

Verdict history already persists (`approval_review_verdicts`, migration 0288). What did **not** exist
was a read projection onto the instance DTO. Building it surfaced a latent defect of the eQMS
audit-truth class: `actor_display_name_snapshot` is **nullable**, both insert paths bind it
**conditionally** (`if name != "" { … }` → NULL on empty), and both read paths **fall back**
(`coalesce(actor_display_name_snapshot, '')`; the signoff mapper additionally falls back to a *live*
name lookup when the snapshot is empty). A live fallback means a later rename **rewrites history** —
the signed record would display a name the actor did not hold when they signed. That is forbidden in a
regulated eQMS (Veeva/Qualio model: the actor's identity is fixed at the instant of the signed action).

Root cause, established on-disk: the fallback is **structurally unnecessary**. The write-path name
source is `iam_users.display_name`, which is `NOT NULL` (`db/baseline/0001_current_schema.sql:978`).
Every snapshot is written from a non-null source off-tx at the moment of the action
(`review_verdict_service.go:81` → `:179`). The nullable column + conditional bind + read fallback are a
**local-maximum workaround** masking a hole that the upstream invariant already closes transitively.
Per the operator's standing principle — *a robust system with no redundancies does not need a fallback;
a fallback is a weak bound that masks a bad implementation; fix the root cause, we don't care about
refactoring* — the fallbacks are removed and the invariant is made explicit at the storage layer.

## Decision

**1 — `verdicts[]` history contract.** `ApprovalInstanceByDocumentResponse` gains a **required**
`verdicts` array (empty ⇒ `[]`), spanning **all** stages of the instance, ordered chronological
ascending by `verdict_at`. Each element:

```yaml
verdict:
  type: object
  required: [id, stage_instance_id, actor_user_id, actor_display_name, verdict, verdict_at]
  properties:
    id:                 { type: string }
    stage_instance_id:  { type: string }
    actor_user_id:      { type: string }
    actor_display_name: { type: string }   # snapshot at action time — immutable, never re-resolved
    verdict:            { type: string, enum: [ready, request_changes] }
    reason:             { type: string, nullable: true }   # verdict comment
    verdict_at:         { type: string, format: date-time } # RFC3339 UTC
```

Defined in OpenAPI, consumed only via generated types (ADR 0035); no hand-written `body.data.verdicts`
reader.

**2 — Immutable actor name, DB-enforced, no read fallback.** The name displayed for any verdict or
signoff is the value cast into `actor_display_name_snapshot` at the moment of the action, and only that:

- **Storage invariant (D1).** Migration backfills any NULL/`''` snapshot from `iam_users.display_name`,
  then `SET NOT NULL` + `CHECK (actor_display_name_snapshot <> '')` on **both**
  `approval_review_verdicts` **and** `approval_signoffs`. The DB is the last line (DB-enforces-invariants
  rule); the column can no longer be empty.
- **Write binding (D1).** `InsertVerdict` and `InsertSignoff` bind the snapshot **unconditionally**
  (the `if name != ""` omission is deleted). A signed action with an empty name is now rejected at
  insert — fail closed, not silently NULL. No decision-logic changes on the write path.
- **Read projection (D2).** `LoadInstanceVerdicts` (new; mirrors `LoadStageVerdicts` with
  `WHERE approval_instance_id = $1`) selects the snapshot **directly** — the `coalesce(…, '')` is
  removed. The signoff mapper's snapshot-else-live branch collapses to **snapshot-only**; the live
  `displayNameReader` fallback on this path is deleted.

The delegator name for `on_behalf_of` verdicts (D3) needs a separate snapshot column and is **deferred**
to a later feature; it is out of scope here.

## Consequences

- History is audit-honest: a rename after a signed action never rewrites what a completed verdict/signoff
  displays. The name is the identity at action time, immutable.
- One fewer read fallback and one fewer live-lookup path on the view; the name is a pure projection of
  stored truth. Read code no longer branches on emptiness that can no longer occur.
- The instance DTO carries the full forward (`viewer`) + backward (`verdicts[]`) picture — the F2d.5
  screen and F2d.3 selector consume one shape; no client history derivation.
- Consciously overrides the M2d "F2d.2 needs NO migration" line via a recorded HS-2 clearance
  (milestone.md Amendment 2026-07-09). The migration is additive and forward-only; backfill is
  idempotent (non-empty snapshots untouched).

## Alternatives rejected

- **Keep the read fallback / live re-resolution (Option B).** Read-only, no migration — but leaves the
  audit-truth hole open (rename rewrites history) and locks in the local-maximum workaround. Rejected by
  the operator's no-fallback principle (chose Option A, 2026-07-09).
- **Keep the column nullable, enforce only in app code.** App checks are the friendly first line, not the
  last; a future write path could reintroduce a NULL. The invariant belongs in the DB (DB-enforces-invariants).
- **Snapshot the name at read time from `iam_users`.** That *is* the live fallback — the exact defect this
  ADR closes.
