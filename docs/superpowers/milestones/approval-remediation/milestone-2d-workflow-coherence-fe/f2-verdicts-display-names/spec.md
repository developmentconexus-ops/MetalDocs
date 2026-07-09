# Feature F2d.2 — Spec

> **Milestone:** 2d — Workflow Coherence FE + Viewer Contract  ·  **Folder:** `f2-verdicts-display-names`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-09 / operator (milestone pre-approved HS-1 2026-07-08 + interview
> answer below on name source + **HS-2 scope override, Option A**) — *no implementation begins until this line is filled.*
> **HS-2 cleared (operator 2026-07-09):** the global-maximum root-cause fix (DB-enforced snapshot
> invariant + removal of the pre-existing signoff read fallback) is IN SCOPE for F2d.2, consciously
> overriding the milestone's "F2d.2 needs NO migration" line. One migration + a minimal
> insert-binding + read-fallback removal are authorized. See milestone.md amendment 2026-07-09.

> Feature contract, approved **before any code**. Defines what the feature must do and how it is
> proven — not how it is built (that is `plan.md`). Validator judges against this file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | `verdicts[]` actor display name — source: `actor_display_name_snapshot` (as-cast) vs live off-tx resolution? | **Snapshot, NO fallback** (operator 2026-07-09): *"a professional system … global maximum … no redundancies, does not need a fallback (unless a fail-safe); fallback is a weak bound, a workaround that can mask a bad implementation (local maximum) … analyse the root cause … we don't care about refactoring."* → verdict actor name = `actor_display_name_snapshot`, read as a **pure projection**. No read-side `coalesce`, no live `displayNameReader` fallback. This is the eQMS audit-truth model (Veeva/Qualio snapshot the signer name at the moment of the signed action; a later rename must NOT rewrite history). |
| 2 | Surface the delegator (`on_behalf_of_user_id`) in `verdicts[]` ("X on behalf of Y")? | **No preference (operator)** → **defer** (YAGNI + no-migration boundary). The delegator name is NOT snapshotted; showing it as-cast would need a new snapshot column (migration, out of F2d.2 scope) and live resolution would reintroduce a fallback. The sidebar timeline shows who cast each verdict; the delegation chain is deferred (trigger below). `on_behalf_of_user_id` is not projected in this feature. |

## Root-cause grounding (why no fallback is correct here, not a shortcut)

The no-fallback contract is only sound if the snapshot is a guaranteed non-empty invariant. It is — transitively, already, with no new enforcement:

- A verdict actor has passed tier-2 authz to cast the verdict ⇒ they have a `metaldocs.iam_users`
  row in the tenant.
- `iam_users.display_name` is **`NOT NULL`** (`db/baseline/0001_current_schema.sql:978`).
- The write path resolves the name off-tx via `LoadActorDisplayName` → `DisplayName`
  (`iam/.../user_display_name_repository.go:32`) and snapshots it at verdict time
  (`review_verdict_service.go:81,179`). For a real actor the row exists and the name is non-empty, so
  the snapshot is non-empty.

Under Option A (operator HS-2 override, 2026-07-09) the invariant is not left "true transitively" — it
is **made the DB's last line**, and every read becomes a pure projection with no fallback anywhere:

- **D1 (DB-enforces-invariants, root cause).** A migration hardens both snapshot columns
  (`approval_review_verdicts.actor_display_name_snapshot`, `approval_signoffs.actor_display_name_snapshot`):
  backfill any legacy `NULL`/`''` from `iam_users.display_name`, then `SET NOT NULL` + `CHECK (… <> '')`.
  The DB is now the single enforcement point (no redundant app-level guard — matches the operator's
  "no redundancies"). The insert path binds the column **unconditionally** (removing the current
  `if name != ""` omission in `InsertVerdict`/`InsertSignoff`, `postgres_approval_repository.go:163,1226`)
  so a genuinely-empty name surfaces as a `CHECK` violation (fail-closed at write) instead of a silent
  `NULL`. This is the only write-side touch — verdict/signoff **decision logic is unchanged**.
- **D2 (read is a pure projection).** With the column `NOT NULL`, the read-side `coalesce(…,'')` is
  dead and removed (`LoadStageVerdicts` + the new by-instance loader scan a plain non-null string), and
  the pre-existing signoff read's snapshot-**else-live** branch (`get_instance_handler.go:147→171`) —
  which under the as-cast audit lens is a latent defect (an actor who HAS signed would render a
  *current* name, not the as-signed one) — collapses to **snapshot-only**. No `displayNameReader`
  fallback survives on any snapshotted-action read path.

The verdict projection therefore reads `actor_display_name_snapshot` directly; the DB guarantees it is
present and non-empty. (The `on_behalf_of` delegator name remains the one field with no snapshot — see
Non-goals / D3.)

## Consumer contract (FIRST — before any producer)

- **Consumer:** the instance sidebar **timeline** on the single working screen (F2d.5) — a
  reverse/forward-chronological list of review decisions across the instance's stages.
- **Contract:** `ApprovalInstanceByDocumentResponse` (the by-document view DTO — same DTO F2d.1's
  `viewer` block lives on; the by-id `ApprovalInstanceResponse` does NOT carry verdicts, matching the
  F2d.1 split) gains a **required** `verdicts` array (empty array when the instance has no review
  verdicts yet — present, never omitted):

  ```yaml
  verdicts:
    type: array
    items:
      type: object
      required: [id, stage_instance_id, actor_user_id, actor_display_name, verdict, verdict_at]
      properties:
        id:                { type: string, format: uuid }
        stage_instance_id: { type: string, format: uuid }
        actor_user_id:     { type: string }
        actor_display_name:{ type: string }   # actor_display_name_snapshot, as-cast (immutable audit truth)
        verdict:
          type: string
          enum: [ready, request_changes]      # domain.Verdict values
        reason:            { type: string }    # the verdict comment; "" when none
        verdict_at:        { type: string, format: date-time }  # RFC3339 UTC
  ```
  Semantics:
  - Ordered **chronological ascending** by `verdict_at` (the timeline renders oldest→newest; matches
    the existing `LoadStageVerdicts ORDER BY verdict_at ASC`).
  - Spans **all** stages of the instance (by-instance read), not just the active stage.
  - `actor_display_name` = `actor_display_name_snapshot`, projected directly (no fallback, per §Interview).
  - `reason` = the verdict `comment` (empty string when the verdict carried none).
- **Source of truth for the contract:** `api/openapi/v1/openapi.yaml` → `oapi-codegen` Go + TS types.
  No hand-written `body.data.verdicts` consumer (ADR 0035).

- **Stage-actor display names** (the milestone row's second clause) are **already** projected on
  `StageActor.DisplayName` via `resolveEligibleActorNames` (off-tx `displayNameReader`,
  `get_instance_handler.go:171`) — untouched here (F2d.1 non-goal reaffirmed). This feature adds only
  `verdicts[]`.

## What this feature implements

- **D1 — migration + insert hardening (root cause).** New migration: backfill any
  `NULL`/`''` `actor_display_name_snapshot` from `iam_users.display_name`, then `SET NOT NULL` +
  `CHECK (actor_display_name_snapshot <> '')` on **both** `approval_review_verdicts` and
  `approval_signoffs`. Bind the column unconditionally in `InsertVerdict` + `InsertSignoff` (remove the
  `if name != ""` omission). No decision-logic change.
- **D2 — read-fallback removal.** Drop `coalesce(actor_display_name_snapshot,'')` in `LoadStageVerdicts`
  and the new by-instance loader (scan a plain non-null string, guaranteed by D1). Collapse the signoff
  read's snapshot-else-live branch (`get_instance_handler.go:147→171`) to snapshot-only.
- **New by-instance repository read** `LoadInstanceVerdicts(ctx, tx, tenantID, instanceID)` — mirrors
  `LoadStageVerdicts` (`postgres_approval_repository.go:1326`) but `WHERE v.approval_instance_id = $1`
  (still tenant-joined + `actor_tenant_id = ai.tenant_id`), `ORDER BY v.verdict_at ASC`, snapshot read
  as a plain string (post-D1, no `coalesce`).
- **App read** extends `LoadInstanceByDocumentForViewWithViewer` (F2d.1) to also load the verdicts in
  the SAME view tx (plain read-only SELECT, no lock, snapshot names are inline ⇒ **no off-tx join
  needed at all** — H-PRE-1 trivially satisfied) and return them alongside the instance + viewer facts.
- **Handler** projects `[]domain.ReviewVerdict` → `[]contracts.VerdictRecord` in `mapInstanceResponse`
  (by-document path; by-id passes none). Display name = `v.ActorDisplayNameSnapshot()` directly.
- **OpenAPI edit** + `oapi-codegen` regen (Go + TS). New required `verdicts` array as above.

## Non-goals (mandatory)

- **No verdict/signoff DECISION-logic change.** The write-side touch is limited to binding the snapshot
  column unconditionally in the INSERTs (D1); eligibility/SoD/quorum/transition logic is untouched.
- **No `on_behalf_of` projection** (interview #2 defer → D3).
- **No stage-actor name change** — already exists via `resolveEligibleActorNames` (the D2 signoff-name
  change removes a fallback branch; it does not change the resolution for not-yet-acted actors, which
  legitimately use the live name).
- **No new capability, no route change** (the endpoint already exists; only its response schema grows).
- FE timeline component is **F2d.5**, not here — this feature ships the typed contract it consumes.
- *(Migration IS in scope under the HS-2 override — D1. This reverses the milestone row's "No
  migration"; recorded in milestone.md amendment 2026-07-09.)*

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| verdict history shape after a **`ready`** verdict (single-stage) | real-DB `TestInstanceVerdicts_Ready` (testdb factory) — asserts one entry, correct actor/verdict/reason/verdict_at | real |
| verdict history after a **`request_changes`** verdict | `TestInstanceVerdicts_RequestChanges` — asserts entry with `verdict=request_changes`, reason present | real |
| verdicts span **multiple stages**, chronological asc | `TestInstanceVerdicts_MultiStageOrdered` — two stages, asserts order by `verdict_at` | real |
| display names **resolve for seeded users** (snapshot, as-cast) | asserted in the above (seed users with display names, cast verdicts, assert `actor_display_name` = snapshot) | real |
| **no in-tx display lookup**; no read-side fallback | code review: `LoadInstanceVerdicts` reads snapshot inline, no `displayNameReader` call on this path | real |
| empty instance ⇒ `verdicts: []` (present, not omitted) | `TestInstanceVerdicts_None` — instance with no verdicts serializes `[]` | real |
| **D1** — snapshot columns `NOT NULL` + `CHECK (<> '')` after migration; insert binds unconditionally | migration applied on the testdb schema; `TestInsertVerdict_EmptySnapshot_Rejected` (empty name → CHECK violation, fail-closed) | real |
| **D1** — backfill leaves zero null/empty rows | migration includes backfill from `iam_users`; `SET NOT NULL` succeeds on the real schema (migration runs clean in the testdb bootstrap) | real |
| **D2** — no fallback survives on snapshotted-action reads | grep gate: no `coalesce(actor_display_name_snapshot` in the repo; no snapshot-else-`displayNameReader` branch for signed actors in `get_instance_handler.go` | real |
| **D2** — signoff name still projects (snapshot-only) after fallback removal | existing signoff-name real-DB/handler test stays green (snapshot present ⇒ same output) | real |
| regen produces no hand-written DTO consumer; build clean | `oapi-codegen` regen + `go build ./...` + grep gate (`body.data.verdicts`) | real |

> TDD: write the failing real-DB test first, then implement to green.

## ADR needed?

- [x] Durable decision → **ADR 0079 — Approval verdict-history contract + immutable actor-name
  snapshot (no read fallback).** Records: verdict history is projected by-instance onto the view DTO;
  `actor_display_name` is the as-cast snapshot, read as a pure projection that fails closed on `NULL`
  (no `coalesce`/live fallback); the invariant is guaranteed transitively by `iam_users.display_name
  NOT NULL` + write-time snapshot; the two DB/read hardenings (D1/D2) are deferred with triggers.
  Links ADR 0078 (viewer facts), ADR 0035 (generated DTOs), the no-fallback principle.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| **D3** — surface `on_behalf_of` (delegated verdict "X on behalf of Y") in `verdicts[]` | Delegator name is not snapshotted; as-cast display would need a NEW snapshot column (a separate additive migration + a write-path capture change), which is a distinct decision from D1's constraint hardening | Trigger: when the F2d.5 timeline design calls for delegation attribution. Owner: approval module + F2d.5. |

*(D1 and D2 were promoted from defers into F2d.2 scope by the operator HS-2 override, 2026-07-09 —
Option A, global maximum. Only D3 remains deferred.)*
