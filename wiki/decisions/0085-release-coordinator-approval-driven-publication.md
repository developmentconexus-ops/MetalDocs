# ADR 0085 — Release coordinator: approval-driven publication, manual publish retired

> **Status:** Accepted 2026-07-28 (operator ruling D7: "ADR + coordenador já"; Codex-aligned rev 2)
> **Supersedes:** the manual publish/schedule-publish/supersede endpoints and the
> `document.publish` capability (D3 of 2026-07-28 is void).
> **Amends:** [ADR 0069](0069-document-periodic-review-and-reason-for-change.md) — its
> "`effective_from` is the single effective/target date" ruling is replaced by the
> planned/actual split defined here (`planned_effective_from` = plan,
> `effective_from` = actual). Its review/expiry model is otherwise unchanged.
> **Extends:** [ADR 0082](0082-approval-kernel-extraction.md) (approval kernel),
> [ADR 0067](0067-async-job-infrastructure-consolidated-onto-river.md) (River-only async,
> jobs execute in `metaldocs-jobs`), GMR M4 unified status-transition function.
> **Relates to:** [ADR 0022](0022-authz-capability-coherence.md) (capability model),
> [ADR 0068](0068-stuck-instance-watchdog-alert-only.md) (alert-only operational posture).
> **Scope:** usability-remediation Etapa 2 (tracker 2026-07-28); fixes F-QA4-5,
> F-QA4-13, F-QA4-14 structurally.

## Context

Today a document that clears its approval route rests in `approved` until a human
calls `POST /documents/{id}/publish` (or `/schedule-publish`). The operator ruled
this step is legacy: *"no momento que foi aprovado já não era para estar
publicado?"*. Market check (Qualio `effectiveOnApproval`, MasterControl release
workflow, Veeva lifecycle actions) confirms the industry model is
**approval-driven effectiveness with explicit gates** — not a second manual
decision, and not an ungated instant flip either.

The current pipeline is also structurally defective:

- **Two publish paths, two `effective_from` behaviors.** Immediate publish
  (`publish_service.go:55`) never writes `effective_from`; the scheduled-publish
  worker (`scheduler_service.go:28`) **clears** it on `scheduled→published`. The
  M6 review-scheduling predicate (`review_due_reader.go:40`) requires
  `effective_from IS NOT NULL` — so every published document escapes periodic
  review (**F-QA4-13**).
- **Two approval terminal paths, one broken.** Signoff completion
  (`decision_service.go:582`) calls `FreezeService.Pin` (document freeze +
  materialization outbox); the review-verdict route
  (`review_verdict_service.go:292`) reaches `approved` through `executeFreeze`
  (`freeze.go:58-71`), which pins only `approval_instances.frozen_content_hash`
  — it never invokes `FreezeService.Pin`, so no frozen artifact materializes
  (**F-QA4-14**).
- **No durable artifact-readiness fact.** The materialization pipeline writes
  `final_docx_s3_key` as an unguarded column update
  (`snapshot_repository.go:151-165`); its events and dedupe are revision-keyed
  only (`messaging/events.go:30-33`, `staging_outbox.go:56-77`) and cannot
  distinguish approval generations. Approval lifecycle events carry an instance
  id but no revision/hash/generation and are emitted best-effort
  (`decision_service.go:154-159, 729-736`). `frozen_content_hash` is freeze-time
  identity, not readiness.
- **Publish gate is incoherent** (F-QA4-5): `document.publish` is area-grade
  (`capability_scope.go:43`) yet the UI renders the button for actors who cannot
  execute it and blocks actors who can; the frontend infers publication
  readiness from `content_hash` (`useDocumentArtifact.ts:190-211`).

Patching any of these inside the manual-publish model is a local maximum. The
global-maximum structure — what a workflow engine or eQMS does — is a single
**release coordinator**: an idempotent evaluator fed by durable facts that owns
the only path from `approved` to `published`.

## Decision

Approval is the last human decision about a document's release. Everything after
it is mechanical and automatic, executed by an idempotent **release
coordinator** in the approval module. The manual publish step — endpoint,
button, capability — is **deleted, not repurposed**.

### Release generation: the shared identity

Every fact, timer, coordinator evaluation, and emitted event keys on one durable
**release generation** record:

```
(tenant_id, subject_kind='document', document_id, approval_instance_id,
 revision_id, revision_version, frozen_content_hash)
```

- `subject_kind` is retained from the ADR 0082 kernel: the coordinator is
  **document-only**; template subjects keep their existing skip
  (`decision_service.go:717-724`) and never produce release facts.
- Both fact producers are **mandatory and fail-closed**: emitting the approval
  fact is part of the terminal-approval transaction (no best-effort branch), and
  emitting the artifact fact is part of the artifact-persistence transaction.
  A generation with a missing fact is a defect surfaced by reconciliation
  (below), never silently released.
- Artifact persistence becomes **generation-guarded**: the final-artifact write
  carries the expected generation identity and refuses to overwrite a different
  generation's artifact. Materialization idempotency/dedupe keys widen from
  revision-only to the generation key.

### Definitions

| term | meaning |
|---|---|
| **approved(G)** | The approval route reached terminal ACCEPT on generation G. Both routes (signoff and review-verdict) end here via the same unified terminal path: `FreezeService.Pin` + approval fact, in-tx (closes F-QA4-14). |
| **artifact-ready(G)** | The complete final artifact set for G — **final DOCX and final PDF** — exists in blob storage, recorded by the durable artifact fact emitted in the same tx as the last artifact write. Column presence is never readiness. |
| **`planned_effective_from`** | Immutable plan data declared in the publication plan at submission (optional; absent = "effective on release"). New column; never mutated by release. |
| **`effective_from` (actual)** | The timestamp the document actually became effective — written by the winning release transaction as coordinator-evaluation time, **never cleared** (closes F-QA4-13). `scheduled` rows have planned non-null and actual **null**; `published` rows have actual non-null. |
| **released / published** | The coordinator's predicate held and its CAS transition won: document `published`, effective, artifact immutable, predecessors superseded. |
| **readiness hold** | `approved` (or `scheduled`) with the predicate not yet satisfied. A visible, queryable state with an explicit reason — no human action required. |

### Release predicate

The coordinator releases generation G of document D iff **all** hold, evaluated
in one transaction:

1. **Approval fact** for G exists and G is still D's freeze head (no newer
   generation superseded it).
2. **Artifact fact** for G exists (full DOCX+PDF set).
3. **Effective-date gate:** `planned_effective_from` absent or ≤ now. If in the
   future, the coordinator transitions `approved→scheduled` and arms a timer.
4. **Supersession-head check:** D's currently-published revision and every
   cross-document supersede target named in the plan are still supersedable.

### The release transaction (single CAS winner)

- **Source CAS:** one guarded UPDATE through the unified transition function:
  `WHERE status IN ('approved','scheduled') AND <generation match on G>`. Zero
  rows → a concurrent evaluation won or the generation moved; no-op.
- **Targets:** all supersession targets (same-document predecessor + plan-named
  cross-document targets) are locked in **deterministic order** (sorted by
  document id) and each `published→superseded` update is guarded on its expected
  head revision. **Any target conflict rolls back the whole transaction** — the
  document stays in readiness hold with reason `supersede_conflict` and the
  conflict is surfaced (alert-only posture, ADR 0068); no partial release.
- **Writes:** status, actual `effective_from :=` evaluation time,
  `effective_to`/`review_due_at` per plan (validation below), supersessions.
- **Events:** the winning tx enqueues exactly one governance event and one
  lifecycle event per affected transition (`document.published` for D,
  `document.superseded` per predecessor/target) through the existing outbox +
  `EnqueueLifecycleEventTx` fanout, each idempotency-keyed on
  `(G, transition, affected document)`. Consumers see no semantic change.

### Timers and topology (ADR 0067)

The future-effective-date timer is a **River job hosted and executed by
`metaldocs-jobs`** (dual-define per ADR 0067), replacing today's
scheduled-publish worker. The timer payload carries G; firing just invokes the
same idempotent evaluation. Statuses `approved`/`scheduled`/`published` keep
their meaning and UI presentation; the `approved→scheduled→published` edges
become coordinator-only (no HTTP path reaches them). The status model stays the
current **eight-value / eleven-arc** transition set
(`0272_documents_remove_rejected.sql`, `state.go`, `state_parity_test.go`) — no
new statuses, no removed arcs.

### Publication plan at submission

`SubmitDocumentRequest` (spec `openapi.yaml:6737`) gains an optional publication
plan: `planned_effective_from`, `effective_to`, `review_due_at`, and
cross-document supersede targets — the same parameters the retiring
schedule-publish contract accepts today (`contracts/publish.go:16-50`).
Same-document supersession (new revision replaces the currently-published one)
is implicit and automatic.

**Delayed-release validation:** DB invariants `effective_to > effective_from`
and `review_due_at >= effective_from` (migration 0274) are evaluated against
**actual** release time. At release, a plan whose `effective_to` is already ≤
actual time fails the predicate (readiness hold, reason `plan_invalid`, alert)
— never a constraint violation mid-tx. A planned `review_due_at` before actual
release is recomputed from actual release time using the profile's review
interval (ADR 0069 semantics); an explicitly planned future `review_due_at` is
kept.

**Cross-document supersede authorization:** the `document.supersede` capability
is **retained** (area-grade, `capability_scope.go:45`) and re-homed to
submission time: naming cross-document supersede targets in the plan requires
`document.supersede` on each target's area, checked in the submit tx. Without
this, any `document.submit` holder would gain supersession power over arbitrary
documents. Same-document supersession needs no extra capability (it is the
document's own lifecycle).

### Retirement inventory (hard break, no compat shims)

| item | anchors | action |
|---|---|---|
| `POST /documents/{id}/publish`, `/schedule-publish`, `/supersede` | `publish_handler.go:33,77`, `supersede_handler.go:23`; spec paths `openapi.yaml:3659,3697,3791`; `permissions.go:179-180`; generated `routes_generated.go:18-27`, `router.go:28-36`, `api.gen.go` registrations | delete + full spec regen |
| `PublishService`, `SchedulerService`, `SupersedeService` + composition | `publish_service.go`, `scheduler_service.go`, `supersede_service.go`, `services.go:35-48,84-98` | replaced by coordinator |
| Scheduled-publish job type, enqueuer, worker registration | `scheduled_publish_job.go:26-104`, `apps/jobs/.../main.go:104-112`, `apps/api/.../main.go:681-687` | replaced by coordinator timer job |
| Event producers for published/scheduled/superseded | `publish_service.go:133-173`, `scheduler_service.go:152-176`, `supersede_service.go:137-176` | move into the winning release tx |
| `document.publish` capability | `iam/domain/model.go:81`, `capability_scope.go:43`, `catalog.go:113`, seed `0001_product_reference_data.sql:181-182`, cardinality/scope tests `model_test.go:93-101`, `capability_scope_test.go:28-50` | delete; **no tripwire arm removal needed** — no arm carries `document.publish` (`arms.go:156-176`); regen + parity-verify only |
| Frontend publish/schedule/supersede surface | `approvalApi.ts:171,181`, `DocumentDetailRoute.tsx:230-248` publish button, `SupersedePublishDialog.tsx`, E2E `scheduled_publish.spec.ts`, integration `e2e_happy_test.go:265-277` | delete; UI consumes the readiness-hold projection instead |
| FE readiness inference from `content_hash` | `useDocumentArtifact.ts:190-211` | replaced by coordinator projection |
| D3 ruling (publish = capability) | tracker 2026-07-28 | void — the gate it governed no longer exists |

`Obsolete` retirement flow is untouched — obsolescence remains a human
governance decision.

### Observability, attribution, reconciliation

- **Readiness-hold projection:** a queryable coordinator state per generation —
  `awaiting_approval_fact` / `materializing` / `awaiting_effective_date` /
  `supersede_conflict` / `plan_invalid` / `failed` — exposed on the document
  read model. This is the UI's only source of release status.
- **Actor attribution:** governance events require a non-null actor
  (`events.go:45-54`). The release governance event is attributed to a stable
  **system principal** (the coordinator), with causal identities — approval
  instance, final approver, submitter — carried in the payload. Human agency is
  recorded where it happened (approval events); the release is mechanical and
  says so.
- **Reconciliation:** a River periodic maintenance job (ADR 0067 pattern,
  `metaldocs-jobs`) sweeps generations stuck in readiness hold beyond a
  threshold (lost fact, dead-lettered consumer, dead timer) and **alerts** —
  alert-only posture per ADR 0068; duplicate-safe because recovery is always
  "re-run the idempotent evaluation".

### DB invariants and dictionary sync

- `published` ⇒ `effective_from IS NOT NULL`; `scheduled` ⇒
  `planned_effective_from IS NOT NULL AND effective_from IS NULL` — trigger/CHECK
  enforced, not app-only.
- Unified transition function keeps the eight-value/eleven-arc edge set.
- Status dictionary (wiki + enum docs) updated in the same change; the
  hand-synced enumeration meta-defect list gains no new member.

### In-flight disposition (executable backfill)

Dev-stage dataset only (no production tenants; consistent with
`migration-policy.md` dev-seed separation). The backfill is an idempotent
migration + job pass that creates the prerequisites **before** any evaluation:

1. **Synthesize approval facts** for every existing `approved`/`scheduled`
   document from its terminal approval instance, generation-scoped per the key
   above.
2. **Migrate planned dates:** existing `scheduled` rows carry their planned date
   in `effective_from` (`publish_service.go:305-320`) — move it to
   `planned_effective_from`, null the actual.
3. **Repair materialization** for generations without artifacts (review-verdict
   victims): perform the Pin-equivalent repair (`FreezeService.Pin` writes freeze
   state + enqueues materialization; unpinned documents are rejected at
   `freeze_service.go:239-245`, so the repair must pin first). The staging-outbox
   dedupe (`staging_outbox.go:56-77`) is generation-aware after this ADR, so
   re-enqueue is not silently swallowed.
4. **Enqueue evaluation** for every generation only after 1–3 hold. Eligible
   documents release with actual `effective_from` = evaluation time; `scheduled`
   keep their planned date; artifact-pending generations stay in readiness hold
   with an honest reason.

Pre-fix quarantined artifacts (`ba24c4f2…`, `45c9e784…`, `d18fbfdf…`) are
handled by the Etapa 5 release gate, not by this migration.

## Consequences

- One human decision (approval), one mechanical release path, one
  `effective_from` semantics — review scheduling (M6) works for every published
  document.
- F-QA4-5 dissolves rather than being fixed: there is no publish button to gate.
- The approval module gains a small durable state — release generations, facts,
  a projection — which is the standard price of exactly-once release semantics
  over an outbox.
- Submission grows into the single point where a human parameterizes release
  (dates, supersessions) — one mental model instead of three endpoints.
- Any future "hold before effective" feature (e.g. training-complete gates, as
  MasterControl models) is a new predicate conjunct + hold reason, not a new
  endpoint.
