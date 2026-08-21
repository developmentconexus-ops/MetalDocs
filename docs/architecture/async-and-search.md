# R10-T5 — Durable Async, Search & External Effects

> **Status:** ACTIVE / OPERATOR-RATIFIED TECHNICAL AUTHORITY  
> **Ratified:** 2026-08-18  
> **Post-T5 Fable bounded amendment:** 2026-08-18 — canonical Search baseline + conditional materialization + projection serialization + late-rendition no-op  
> **T8-E bounded correction:** 2026-08-21 — OfficialRendition durable work only when source transformation is required
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Decision baseline:** `wiki/architecture/rebaseline-decision-registry.md`  
> **Implementation:** BLOCKED

This page records the operator-ratified T5 architecture plus bounded completeness/correction amendments ratified through the post-T5 independent-review checkpoint. T5 defines which Launch effects require durable asynchronous execution, how those effects converge without becoming semantic authority, and how Search remains discovery mechanism rather than lifecycle/access authority.

The operator accepted the rendition/viewer subgate `RV-1→RV-6`, then accepted corrected `T5-A→T5-P`, explicitly ratified the platform-facing T5 summary, and later ratified the bounded Fable amendments on 2026-08-18 without formally reopening T5.

---

## 1. Preserved authority boundary

```text
Document / Revision / Submission / OfficialRendition / Release = Controlled Documents authority
Authorization = T3 current-grant + scope + domain-predicate authority
exact bytes = T4 semantic descriptor + managed-content mechanism
Search query/view or optional Search projection = discovery mechanism only
jobs / queues / retries / workers = execution mechanism only
provider / renderer = external mechanism only
```

Never infer business truth from queue/job/provider/Search-mechanism state.

---

## 2. Rendition / viewer correction — RV-1→RV-6

```text
PDF source
  → direct PDF viewer by default
  → no duplicate generated PDF without a named need

PDF + RequireOfficialRendition(PDF)
  → establish the OfficialRendition semantic fact over the same admitted PDF handle + descriptor
  → no renderer, provider copy or durable rendition job

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
official_rendition_render exists only when frozen representation policy requires OfficialRendition **and transformation to the required format is necessary**
```

A future rebuildable viewable-PDF cache is mechanism only and must not become Release authority.

Renderer product selection is not frozen by T5. EigenPal is the first SourceOnly DOCX viewer candidate, ONLYOFFICE is a stronger viewer/converter candidate, and Gotenberg/LibreOffice is a server-side conversion candidate. Final mechanism selection requires a representative DOCX fidelity corpus.

---

## 3. One durable-job mechanism — T5-A

Launch uses one PostgreSQL-backed transactional durable-job mechanism for the named durable effects that are actually activated.

River remains the selected/reference mechanism because existing evidence proves the needed transaction-coupled enqueue, retry and scheduling primitives.

This mechanism has a current Launch consumer independent of Search materialization: `RequireOfficialRendition` may require durable renderer execution. The same runtime may schedule periodic managed-content GC reconciliation.

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

Baseline:

```text
conditional durable job:
  official_rendition_render(submission_id, required_format)
  only when the frozen Submission requires OfficialRendition **and its exact source must be transformed to the required format**

already-PDF + required PDF:
  establish OfficialRendition synchronously over the same admitted bytes; no durable job

periodic reconciliation, not per-object durable enqueue:
  managed-content GC over durable GC_PENDING

baseline Search:
  canonical PostgreSQL query/view over current canonical searchable facts
  no durable search_refresh job required merely because Search exists
```

Optional Search materialization seam:

```text
IF T6 proves at least one real derived/expensive searchable fact
OR measured scale/ranking/language behavior that canonical query/view cannot sustainably satisfy
THEN activate:
  PostgreSQL materialized Search projection keyed by Document
  + search_refresh(document_id)
  + rebuild/reconciliation
  + §8 projection-write serialization law
```

Do not invent full-text/content extraction or another derived field merely to justify the projection. T6 must name the concrete consumer first.

No mandatory Launch durable job exists merely for viewing PDF/DOCX.

No mandatory Launch durable job exists for notifications, Distribution acknowledgement, Periodic Review, scheduled Release, Audit hash-chain validation or routine external-IdP disable.

---

## 5. Transaction-coupled durable intent — T5-C

If a local semantic transaction creates a **required activated future effect**, durable intent/job is inserted in that same local transaction.

```text
business fact commits
⇔
required durable work exists
```

Examples:

```text
Submission requiring DOCX→OfficialRendition(PDF) transformation
  → official_rendition_render(submission_id, PDF)

Submission already PDF + required OfficialRendition(PDF)
  → no future external effect; no job

IF materialized Search has been activated by T6:
  Release / obsolescence / relevant searchable-current change
  → search_refresh(document_id)
```

Provider/network execution never joins the local semantic transaction.

A canonical query/view Search baseline creates no asynchronous work to enqueue.

---

## 6. Conditional OfficialRendition execution — T5-D / T5-E

When a frozen Submission policy requires an OfficialRendition, execution is conditional on whether transformation is actually needed.

Already-PDF path:

```text
load exact eligible Submission whose source format is PDF
→ open local transaction
→ reload current eligibility
→ prove required PDF rendition still absent/required
→ revalidate the exact Submission handle + descriptor
→ create OfficialRendition over that same handle + descriptor
→ required T3 Audit
→ if human gate also satisfied, execute T2 system Release
→ COMMIT
```

No renderer, provider copy or durable job participates.

Transformation path (current Launch: DOCX→PDF):

```text
load Submission + requirement
→ load exact T4 source bytes
→ execute renderer outside semantic transaction
→ admit/verify READY output through T4
→ open local transaction
→ reload current eligibility
→ prove exact Submission remains the current eligible pre-Release candidate
→ prove required rendition still absent/required
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

### Late renderer result

If the exact Submission is no longer eligible because the attempt was returned/withdrawn or its Revision was cancelled before finalization:

```text
semantic OfficialRendition creation = NO-OP
Release = NO-OP
produced READY output = reclaimable mechanism state after T4 claim/binding release/expiry
```

Do not freeze a permanent OfficialRendition for a dead attempt merely because rendering physically completed.

No silent downgrade is allowed:

```text
RequireOfficialRendition(PDF)
+ renderer failure
≠ SourceOnly
```

A terminal renderer failure leaves the still-eligible Submission truthfully `SUBMITTED` with the Release gate unsatisfied until corrected through an authorized existing path.

---

## 7. Search baseline and optional projection — T5-F

Search is a required Launch journey, but **materialized Search infrastructure is not required merely because Search exists**.

### Baseline

The smallest current mechanism is a canonical PostgreSQL query/view over the current facts required by the Product Contract, such as:

```text
stable code
current EFFECTIVE Revision title
DocumentType
Area
responsible owner
status/current-effectivity facts
```

These facts already belong to canonical product state. T6 owns the exact query/filter/ranking UX and may prove that this baseline is sufficient.

### Conditional materialization

A PostgreSQL-backed rebuildable projection keyed by stable Document identity may be activated only when T6 or measured operating evidence names a real consumer that canonical query/view cannot sustainably satisfy — for example a derived/expensive searchable fact. No external Search engine is baseline.

Whether canonical query/view or optional projection is used, Search never owns:

```text
current EFFECTIVE truth
Authorization
Document lifecycle
exact-content identity
```

Do not introduce an external Search engine without measured need for scale/ranking/language features the product database cannot sustainably provide.

---

## 8. Latest-state materialized refresh + concurrency law — T5-G

This section applies **only if T6 activates a materialized Search projection**.

Job payload:

```text
search_refresh(document_id)
```

For one Document, each refresh execution must acquire the selected per-Document **projection-write serialization** before reading canonical searchable state and hold it through the projection rewrite/removal.

```text
acquire per-Document projection-write serialization
→ reload latest canonical state
→ rewrite/remove projection
→ release serialization
```

A full/rebuild write for that same Document obeys the same serialization law.

This closes overlapping-execution races where an older worker could otherwise read first and write last.

Consequences:

```text
duplicate delivery remains acceptable
out-of-order job start remains acceptable
concurrent overlap cannot leave an older observation as final projection state
retry after later lifecycle transition converges to latest state
```

Per-Document FIFO/broker ordering infrastructure is still not required for correctness. Exact lock/upsert/SQL primitive remains implementation design.

---

## 9. Search freshness / authorization — T5-H

### Canonical query/view baseline

Because the baseline reads current canonical state, there is no asynchronous projection freshness lag to reconcile. T6 still applies current T3 Authorization and domain predicates to the discovery/serve journey.

### If materialized projection is activated

The projection may lag temporarily by omission:

```text
new/current EFFECTIVE Document not yet discoverable
```

But a projection hit never grants access or establishes effectivity.

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

A full rebuild/reconciliation path is mandatory **only if a materialized Search projection is activated**:

```text
recreate projection
→ enumerate canonical Documents
→ for each Document obey §8 serialization
→ compute current searchable facts
→ upsert/remove projection
→ reconciliation check
```

This proves the optional projection is derivative.

A canonical query/view baseline has nothing materialized to rebuild.

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
→ re-prove no live admission claim/binding
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

Historical restore is separate: T4 invalidates all restored ApplicationSessions and gates ordinary serving on required post-snapshot security-teardown reconciliation.

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

Current/conditional effects already have natural outcomes:

```text
renderer                    → admitted content + OfficialRendition
optional materialized Search→ projection state
GC                          → confirmed provider absence + finalized mechanism state
```

A future capability may define a semantic receipt only when proving third-party acceptance is itself a business requirement.

---

## 16. Operational visibility — T5-O

Required activated async work cannot be an invisible background mechanism.

Operations must be able to answer at least:

```text
is the worker processing required work?
available/retry/terminal-failure counts by active required job kind
age of oldest available/retrying required job
recent successful execution / terminal failures
subject id/correlation needed to investigate and redrive
```

If materialized Search is not activated, no Search-job metrics/backlog exist merely for architectural symmetry.

T5 does not freeze endpoint names, metric names, dashboard product, queue names or process topology.

---

## 17. Future evolution law — T5-P

Future capabilities add only named jobs/effects/receipts for proven consumers.

Do not grow T5 into a generic integration/event platform by anticipation.

Examples of future seams may include Distribution delivery, repository publish, governed export packaging, Records disposition enforcement or Training/LMS sync — only when those capabilities are promoted.

Search follows the same law: materialization/extraction/indexing activates only on a named consumer or measured requirement.

---

## 18. Explicit non-decisions

T5 does not decide:

```text
final River version/configuration
final queue names/process placement
SQL/table/index/lock syntax
HTTP health endpoints/metric names
Search API/ranking/filter UX
whether T6 will prove a derived searchable fact requiring materialization
viewer/editor final implementation
renderer product selection
Historical Migration jobs
future notification UX
```

---

## 19. Proof obligations before implementation

Later implementation/tests must prove at least:

```text
conditional rendition enqueue rolls back with Submission transaction rollback
already-PDF + required PDF creates OfficialRendition over the same admitted handle/descriptor with zero renderer/copy/job
DOCX→required PDF still commits the rendition intent atomically when renderer work is activated
SourceOnly viewing creates no official-rendition requirement
process crash after semantic commit cannot permanently lose required activated durable work
duplicate rendition execution cannot create duplicate semantic OfficialRendition
late rendition execution cannot create semantic Rendition/Release for a returned/withdrawn/cancelled candidate
renderer output cannot bypass T4 exact-content admission
renderer failure cannot silently produce Release
canonical Search query/view returns current canonical facts under current AuthZ/domain predicates
IF materialized Search is activated:
  required search-refresh enqueue rolls back with semantic transaction rollback
  overlapping refresh executions serialize before canonical read through projection write
  duplicate/out-of-order refresh ends at latest canonical state
  Search hit cannot bypass current lifecycle/AuthZ revalidation
  full Search rebuild reproduces projection from canonical state under same per-Document serialization
GC recheck prevents current/governed/claim-protected/backup-protected content deletion
GC crash leaves GC_PENDING discoverable
terminal required jobs are operator-visible and redrivable
job payload carries no durable Authorization snapshot/business-content copy
provider request IDs never become business truth
```

---

## 20. Reopen / activation triggers

No formal T5 reopen is required merely to choose the canonical Search baseline in T6.

The optional materialized Search seam is activated when T6/operations proves:

```text
a real derived/expensive searchable fact is required
OR canonical query/view cannot sustainably meet measured scale/ranking/language behavior
```

Reopen only the implicated T5 decision if material evidence instead proves:

```text
multiple durable-job mechanisms are required by a real isolation/trust boundary
external Search engine features/scale are necessary beyond the optional PostgreSQL projection seam
Search omission lag violates a concrete synchronous/SLA requirement after materialization
a Launch notification consumer becomes explicit
provider account disable must converge durably for assurance
GC volume requires per-object durable scheduling
future effect requires durable third-party acceptance receipt
job payload cannot remain bounded IDs without unacceptable cost/coupling
SourceOnly native viewer fidelity/performance proves inadequate and requires a rebuildable viewable-rendition cache
```

Implementation remains **BLOCKED** until all remaining R10 stages and final reviews close.
