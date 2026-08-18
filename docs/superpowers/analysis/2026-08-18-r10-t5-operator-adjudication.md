# R10-T5 — Operator Adjudication / Summary Ratification Gate

> **Status:** ACTIVE STAGING — T5 DECISIONS ADJUDICATED / PLATFORM SUMMARY RATIFICATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Corrected packet:** `docs/superpowers/analysis/2026-08-18-r10-t5-corrected-adjudication-packet.md`  
> **Accepted rendition/viewer subgate:** `docs/superpowers/analysis/2026-08-18-t5-rendition-viewer-strategy-evaluation.md`  
> **Decision baseline:** `wiki/architecture/rebaseline-decision-registry.md`  
> **Implementation:** BLOCKED

This record captures operator adjudication of the corrected T5 architecture after the rendition/viewer subgate RV-1→RV-6 was accepted. It does **not** close T5 or open T6. T5 closes only after explicit operator ratification of the required platform-facing T5 summary.

## 1. Accepted rendition/viewer correction consumed by T5

```text
PDF source
  → direct PDF viewer
  → no duplicate generated PDF by default

DOCX + SourceOnly
  → direct read-only DOCX viewer
  → no persistent governed PDF merely for viewing

DOCX + RequireOfficialRendition(PDF)
  → conditional durable server-side render from exact Submission
  → T4 admission
  → immutable OfficialRendition in ManagedContentStore
  → Release gate
```

Binding distinctions:

```text
preview/viewer mechanism != OfficialRendition
SourceOnly viewing != durable rendering requirement
official_rendition_render exists only when frozen representation policy requires it
```

Renderer product remains evidence-driven and is not frozen by T5.

## 2. T5 adjudication

The operator accepted corrected T5-A→T5-P:

```text
T5-A ACCEPT — one Postgres-backed transactional durable-job mechanism; River remains selected/reference mechanism; no parallel custom scheduler, lease framework or second outbox-dispatch runtime.

T5-B ACCEPT / REFINED — always-required durable job = search_refresh; official_rendition_render is conditional only when the frozen representation policy requires an OfficialRendition; managed-content GC uses periodic reconciliation over durable GC_PENDING rather than one durable job per handle.

T5-C ACCEPT — if a semantic transaction creates a required future effect, its durable job/intent is inserted in that same local transaction; business fact and required durable work commit atomically while provider/network execution stays outside.

T5-D ACCEPT / REFINED BY RV — ordinary PDF/DOCX viewing is a T6 viewer concern and creates no durable-render requirement; when OfficialRendition is required, rendering occurs outside semantic transaction and final T4 admission + OfficialRendition creation + eligible Release revalidation occur in the final local transaction.

T5-E ACCEPT / REFINED BY RV — required OfficialRendition rendering is at-least-once, idempotent and revalidating; duplicate physical outputs cannot create duplicate semantic Renditions; renderer failure cannot silently downgrade RequireOfficialRendition to SourceOnly; SourceOnly viewing has no rendition job.

T5-F ACCEPT — Search is one PostgreSQL-backed rebuildable projection keyed by stable Document identity; no external Search engine Launch baseline without measured need.

T5-G ACCEPT — search_refresh(document_id) reloads latest canonical state at execution time; duplicate/out-of-order jobs converge to current truth, so per-Document FIFO is not required for correctness.

T5-H ACCEPT — Search may lag by omission, but a Search hit never grants access or establishes effectivity; current canonical lifecycle state and T3 Authorization are re-resolved before actionable/readable serve.

T5-I ACCEPT — full Search rebuild/reconciliation is mandatory proof that Search is derivative; always-on global crawler is not Launch baseline.

T5-J ACCEPT — managed-content cleanup is periodic reconciliation over T4 GC_PENDING; immediately before physical delete, canonical eligibility and absence of live/governed/backup-protected references are re-proven; no per-handle durable outbox is required.

T5-K ACCEPT — no mandatory Launch notification inbox/email/push fanout or generic domain-event bus; add named delivery job/projection only when a concrete consumer is promoted.

T5-L ACCEPT — no mandatory durable external IdP-disable job; T3 offboarding already removes MetalDocs access atomically; provider disable is defense-in-depth unless future assurance requires eventual convergence.

T5-M ACCEPT — durable jobs use at-least-once delivery, idempotent/revalidating workers, bounded retry/backoff, fail-loud terminal visibility, manual redrive after cause correction, and bounded stable-ID/minimum immutable-routing payloads rather than business-content/AuthZ snapshots.

T5-N ACCEPT — no generic ExternalEffectReceipt semantic family; current effects use their natural outcomes and future semantic receipts require a named business consumer.

T5-O ACCEPT — required async mechanisms must expose minimum operational visibility: worker processing health, available/retry/terminal-failure counts by required job kind, oldest outstanding age, recent success/failure and subject/correlation for investigation; exact metrics/endpoints/process topology remain implementation design.

T5-P ACCEPT — future capabilities add only named jobs/effects/receipts for proven consumers; T5 must never become a generic integration/event platform by anticipation.
```

## 3. Explicit non-decisions preserved

```text
final River version/configuration
final queue names / process placement
SQL/table/index syntax
HTTP health endpoints / metric names
Search API/ranking/filter UX
viewer/editor final implementation
renderer product selection
Historical Migration jobs
future notification UX
```

## 4. Current gate

```text
RV-1→RV-6                         = ACCEPTED
T5 material decisions A→P         = ADJUDICATED / ACCEPTED
T5 platform summary               = NEXT
T5 final closure/promotion         = PENDING SUMMARY RATIFICATION
Decision Registry update          = PENDING T5 CLOSURE
T6                                = NOT OPEN
implementation                    = BLOCKED
```

Per the operator-approved stage protocol:

```text
T5 corrected design
→ T5 adjudication ✅
→ platform-facing T5 summary NEXT
→ explicit operator summary ratification
→ promote/close T5
→ update Decision Registry
→ remove completed T5 staging
→ only then open T6
```
