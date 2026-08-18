# R10-T5 — Corrected Operator Adjudication Packet

> **Status:** ACTIVE STAGING — T5-A→T5-P OPERATOR-ADJUDICATED / ACCEPTED; PLATFORM SUMMARY NEXT  
> **Date:** 2026-08-18  
> **Parent candidate:** `2026-08-18-r10-t5-durable-async-search-external-effects-candidate.md`  
> **Accepted subgate:** `2026-08-18-t5-rendition-viewer-strategy-evaluation.md` — RV-1→RV-6 ACCEPTED  
> **Adjudication record:** `2026-08-18-r10-t5-operator-adjudication.md`  
> **Implementation:** BLOCKED

This packet is the final corrected T5 decision surface after the operator-ratified rendition/viewer subgate. The operator accepted T5-A→T5-P. T5 is not closed yet: the mandatory platform-facing T5 summary must still be explicitly ratified before promotion, Decision Registry update, staging removal and T6 opening.

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

Binding distinction:

```text
viewer/preview mechanism != OfficialRendition
SourceOnly viewing != durable rendering requirement
official_rendition_render exists only when frozen representation policy requires it
```

Renderer product selection remains empirical: EigenPal is first SourceOnly DOCX viewer candidate; ONLYOFFICE is the stronger viewer fallback; Gotenberg/LibreOffice versus ONLYOFFICE conversion remains a fidelity-corpus decision for required PDF rendering.

## 2. Corrected Launch async census

```text
always-required durable job:
  search_refresh(document_id)

conditional durable job:
  official_rendition_render(submission_id, required_format)
  only when the frozen Submission representation policy requires OfficialRendition

periodic reconciliation, not per-object durable enqueue:
  managed-content GC over durable GC_PENDING
```

No mandatory Launch durable job exists merely for viewing PDF/DOCX.

## 3. T5-A→T5-P — accepted decisions

```text
T5-A ACCEPT — use one Postgres-backed transactional durable-job mechanism; River remains selected/reference mechanism. Do not run a parallel custom scheduler, lease framework or second outbox-dispatch runtime.

T5-B ACCEPT / REFINED — always-required durable job = search_refresh; official_rendition_render is conditional only when the frozen representation policy requires an OfficialRendition; managed-content GC uses periodic reconciliation over durable GC_PENDING rather than one durable job per handle.

T5-C ACCEPT — if a semantic transaction creates a required future effect, its durable job/intent is inserted in that same local transaction. Business fact commit and required durable work existence are atomic; provider/network execution is not.

T5-D ACCEPT / REFINED BY RV — ordinary PDF/DOCX viewing is a T6 viewer concern and creates no durable-render requirement. When OfficialRendition is required, rendering occurs outside the semantic transaction; final T4 admission plus semantic OfficialRendition creation and any eligible Release revalidation occur inside the final local transaction.

T5-E ACCEPT / REFINED BY RV — required OfficialRendition rendering is at-least-once, idempotent and revalidating; duplicate physical outputs cannot create duplicate semantic Renditions. Renderer failure cannot silently downgrade RequireOfficialRendition to SourceOnly. SourceOnly viewing itself has no rendition job.

T5-F ACCEPT — Search is one PostgreSQL-backed rebuildable projection keyed by stable Document identity. No external Search engine is a Launch baseline without measured need.

T5-G ACCEPT — search_refresh(document_id) reloads the latest canonical state at execution time; duplicate and out-of-order jobs converge to current truth, so per-Document FIFO is not required for correctness.

T5-H ACCEPT — Search may temporarily lag by omission, but a Search hit never grants access or establishes effectivity. Current canonical lifecycle state and T3 Authorization are re-resolved before serving actionable/readable content.

T5-I ACCEPT — a full Search rebuild/reconciliation path is mandatory proof that Search is derivative; an always-on global reconciliation crawler is not Launch baseline.

T5-J ACCEPT — managed-content cleanup is periodic reconciliation over T4 GC_PENDING. Immediately before physical delete, canonical eligibility and absence of live/governed/backup-protected references are re-proven. No per-handle durable outbox is required.

T5-K ACCEPT — no mandatory Launch notification inbox/email/push fanout or generic domain-event bus. Add a named delivery job/projection only when a concrete Launch/Future consumer is promoted.

T5-L ACCEPT — no mandatory durable external IdP-disable job. T3 offboarding already removes MetalDocs access atomically; provider disable is defense-in-depth unless a future assurance requirement mandates eventual convergence.

T5-M ACCEPT — durable jobs use at-least-once delivery, idempotent/revalidating workers, bounded retry/backoff, fail-loud terminal visibility, manual redrive after cause correction, and bounded payloads containing stable IDs/minimum immutable routing facts rather than business-content/AuthZ snapshots.

T5-N ACCEPT — no generic ExternalEffectReceipt semantic family. Current effects already have natural outcomes: OfficialRendition, Search projection state, or finalized managed-content cleanup. Add a semantic receipt only for a future business requirement to prove third-party acceptance.

T5-O ACCEPT — required async mechanisms must expose minimum operational visibility: worker processing health, available/retry/terminal-failure counts by required job kind, oldest outstanding age, recent success/failure and subject/correlation needed to investigate. Exact metrics/endpoints/process topology remain implementation design.

T5-P ACCEPT — future capabilities add only named jobs/effects/receipts for proven consumers; T5 must never grow into a generic integration/event platform by anticipation.
```

## 4. Explicit non-decisions preserved

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

## 5. Current gate

```text
RV-1→RV-6                         = ACCEPTED
corrected T5-A→T5-P               = OPERATOR-ADJUDICATED / ACCEPTED
T5 platform-facing summary        = NEXT
T5 promotion/closure              = PENDING SUMMARY RATIFICATION
Decision Registry update          = PENDING T5 CLOSURE
T6                                = NOT OPEN
implementation                    = BLOCKED
```

After this technical adjudication, **do not open T6**. Present the mandatory platform-facing T5 summary and obtain explicit operator ratification first.
