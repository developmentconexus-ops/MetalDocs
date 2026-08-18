# R10-T5 — Durable Async, Search & External Effects

> **Status:** ACTIVE / OPERATOR-RATIFIED TECHNICAL AUTHORITY  
> **Ratified:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Decision baseline:** `wiki/architecture/rebaseline-decision-registry.md`  
> **Implementation:** BLOCKED

This page records the operator-ratified T5 architecture. T5 defines which Launch effects require durable asynchronous execution, how those effects converge without becoming semantic authority, and how Search remains a rebuildable discovery projection.

The operator accepted the rendition/viewer subgate `RV-1→RV-6`, then accepted corrected `T5-A→T5-P`, and explicitly ratified the platform-facing T5 summary on 2026-08-18.

---

## 1. Preserved authority boundary

```text
Document / Revision / Submission / OfficialRendition / Release = Controlled Documents authority
Authorization = T3 current-grant + scope + domain-predicate authority
exact bytes = T4 semantic descriptor + managed-content mechanism
Search = projection only
jobs / queues / retries / workers = execution mechanism only
provider / renderer = external mechanism only
```

Never infer business truth from queue/job/provider state.

---

## 2. Rendition / viewer correction — RV-1→RV-6

```text
PDF source
  → direct PDF viewer by default
  → no duplicate generated PDF without a named need

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

A future rebuildable viewable-PDF cache is mechanism only and must not become Release authority.

Renderer product selection is not frozen by T5. EigenPal is the first SourceOnly DOCX viewer candidate, ONLYOFFICE is a stronger viewer/converter candidate, and Gotenberg/LibreOffice is a server-side conversion candidate. Final mechanism selection requires a representative DOCX fidelity corpus.

---

## 3. One durable-job mechanism — T5-A

Launch uses one PostgreSQL-backed transactional durable-job mechanism.

River remains the selected/reference mechanism because existing evidence proves the needed transaction-coupled enqueue, retry and scheduling primitives.

Laws:

```text
River job row != business state
queue status != Document lifecycle
worker success != Release
worker failure != new Document/Submission lifecycle state
```

Do not run a parallel custom scheduler, lease framework or second outbox-dispatch runtime beside the selected mechanism.

---

## 4. Launch durable-effect census — T5-B

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

No mandatory Launch durable job exists for notifications, Distribution acknowledgement, Periodic Review, scheduled Release, Audit hash-chain validation or routine external-IdP disable.

---

## 5. Transaction-coupled durable intent — T5-C

If a local semantic transaction creates a required future effect, durable intent/job is inserted in that same local transaction.

```text
business fact commits
⇔
required durable work exists
```

Examples:

```text
Release / obsolescence / current-effective change
  → search_refresh(document_id)

Submission requiring OfficialRendition(PDF)
  → official_rendition_render(submission_id, PDF)
```

Provider/network execution never joins the local semantic transaction.

---

## 6. Conditional OfficialRendition execution — T5-D / T5-E

When a frozen Submission policy requires an OfficialRendition:

```text
load Submission + requirement
→ load exact T4 source bytes
→ execute renderer outside semantic transaction
→ admit/verify READY output through T4
→ open local transaction
→ reload current eligibility
→ prove required rendition still absent/eligible
→ revalidate exact READY descriptor + binding
→ create OfficialRendition + required T3 Audit
→ if human gate also satisfied, execute T2 system Release
→ COMMIT
```

At-least-once execution is allowed. Semantic uniqueness is:

```text
at most one required OfficialRendition
per exact Submission + required format
```

Duplicate physical outputs may exist transiently; only one can become semantic truth and loser outputs remain reclaimable mechanism state.

No silent downgrade is allowed:

```text
RequireOfficialRendition(PDF)
+ renderer failure
≠ SourceOnly
```

A terminal renderer failure leaves the Submission truthfully `SUBMITTED` with the Release gate unsatisfied until corrected through an authorized existing path.

---

## 7. Search projection — T5-F

Launch Search is one PostgreSQL-backed rebuildable discovery projection keyed by stable Document identity.

Do not introduce an external Search engine without measured need for scale/ranking/language features the product database cannot sustainably provide.

Search never owns:

```text
current EFFECTIVE truth
Authorization
Document lifecycle
exact-content identity
```

The exact searchable field set belongs to T6/implementation design after user journeys are fixed.

---

## 8. Latest-state refresh — T5-G

Job payload:

```text
search_refresh(document_id)
```

The worker reloads the latest canonical state at execution time and rewrites/removes the projection to match that state.

Consequences:

```text
duplicate jobs    → harmless
out-of-order jobs → converge to latest state
retry after later lifecycle transition → converges to latest state
```

Per-Document FIFO/ordering infrastructure is not required for correctness.

---

## 9. Search freshness / authorization — T5-H

Search may lag temporarily by omission:

```text
new/current EFFECTIVE Document not yet discoverable
```

But a Search hit never grants access or establishes effectivity.

Before actionable/readable serve:

```text
Search hit stable Document id
→ reload current canonical state
→ apply current T3 Authorization
→ serve only current allowed truth
```

A stale projection must never become stale authority.

---

## 10. Search rebuild — T5-I

A full Search rebuild/reconciliation path is mandatory:

```text
recreate projection
→ enumerate canonical Documents
→ compute current searchable facts
→ upsert projection
→ reconciliation check
```

This is the proof that Search is derivative.

An always-on global reconciliation crawler is not Launch baseline. Add one only if operating evidence proves recurring drift that transaction-coupled refresh cannot sustainably prevent.

---

## 11. Managed-content GC — T5-J

T4 already owns durable technical eligibility:

```text
GC_PENDING
```

Therefore Launch uses periodic reconciliation rather than one durable job per abandoned handle:

```text
claim bounded GC_PENDING set
→ immediately re-prove no current WorkingContent reference
→ re-prove no immutable governed/imported reference
→ re-prove no backup exclusion/pin
→ provider DeleteReclaimable outside semantic tx
→ finalize/remove mechanism state after confirmed absence
```

Crash leaves `GC_PENDING` discoverable. Safe failure is leaked storage, not lost business truth.

---

## 12. Notifications / generic event bus — T5-K

Launch has no mandatory notification inbox/email/push consumer.

Do not preserve by sunk cost:

```text
lifecycle notification fanout
per-recipient notification inbox/read state
generic domain-event bus
generic replayable business-event log
```

A concrete future/Launch+ consumer may add the smallest named delivery job/projection then.

Audit remains action evidence and is never the async bus.

---

## 13. External IdP disable — T5-L

T3 offboarding already atomically removes MetalDocs access by disabling User eligibility, revoking ApplicationSessions and removing current memberships/direct grants.

External provider account disable is defense-in-depth unless a future assurance requirement mandates eventual provider convergence.

No mandatory `provider_subject_disable` durable job or generic identity-sync engine is Launch baseline.

---

## 14. Durable-job semantics — T5-M

```text
at-least-once delivery
idempotent/revalidating workers
bounded retry/backoff
fail-loud terminal visibility
manual redrive after cause correction
bounded payloads containing stable IDs + minimum immutable routing facts
worker reloads canonical state before effect/finalization
```

Do not place full business content, mutable Authorization snapshots or large domain-state copies in job payloads.

Long-lived jobs never grant authority.

---

## 15. No generic ExternalEffectReceipt — T5-N

Launch has no generic `ExternalEffectReceipt` semantic family.

Current effects already have natural outcomes:

```text
renderer → admitted content + OfficialRendition
Search   → Search projection state
GC       → confirmed provider absence + finalized mechanism state
```

A future capability may define a semantic receipt only when proving third-party acceptance is itself a business requirement.

---

## 16. Operational visibility — T5-O

Required async work cannot be an invisible background mechanism.

Operations must be able to answer at least:

```text
is the worker processing required work?
available/retry/terminal-failure counts by required job kind
age of oldest available/retrying required job
recent successful execution / terminal failures
subject id/correlation needed to investigate and redrive
```

T5 does not freeze endpoint names, metric names, dashboard product, queue names or process topology.

---

## 17. Future evolution law — T5-P

Future capabilities add only named jobs/effects/receipts for proven consumers.

Do not grow T5 into a generic integration/event platform by anticipation.

Examples of future seams may include Distribution delivery, repository publish, governed export packaging, Records disposition enforcement or Training/LMS sync — only when those capabilities are promoted.

---

## 18. Explicit non-decisions

T5 does not decide:

```text
final River version/configuration
final queue names/process placement
SQL/table/index syntax
HTTP health endpoints/metric names
Search API/ranking/filter UX
viewer/editor final implementation
renderer product selection
Historical Migration jobs
future notification UX
```

---

## 19. Proof obligations before implementation

Later implementation/tests must prove at least:

```text
required search-refresh enqueue rolls back with semantic transaction rollback
conditional rendition enqueue rolls back with Submission transaction rollback
SourceOnly viewing creates no official-rendition requirement
process crash after semantic commit cannot permanently lose required durable work
duplicate rendition execution cannot create duplicate semantic OfficialRendition
renderer output cannot bypass T4 exact-content admission
renderer failure cannot silently produce Release
out-of-order Search refresh converges to latest canonical state
Search hit cannot bypass current lifecycle/AuthZ revalidation
full Search rebuild reproduces projection from canonical state
GC recheck prevents current/governed/backup-protected content deletion
GC crash leaves GC_PENDING discoverable
terminal required jobs are operator-visible and redrivable
job payload carries no durable Authorization snapshot/business-content copy
provider request IDs never become business truth
```

---

## 20. Reopen triggers

Reopen only the implicated decision if material evidence proves:

```text
multiple durable-job mechanisms are required by a real isolation/trust boundary
external Search features/scale are necessary
Search omission lag violates a concrete synchronous/SLA requirement
a Launch notification consumer becomes explicit
provider account disable must converge durably for assurance
GC volume requires per-object durable scheduling
per-Document Search ordering is necessary despite latest-state refresh
future effect requires durable third-party acceptance receipt
job payload cannot remain bounded IDs without unacceptable cost/coupling
SourceOnly native viewer fidelity/performance proves inadequate and requires a rebuildable viewable-rendition cache
```

Implementation remains **BLOCKED** until all remaining R10 stages and final reviews close.
