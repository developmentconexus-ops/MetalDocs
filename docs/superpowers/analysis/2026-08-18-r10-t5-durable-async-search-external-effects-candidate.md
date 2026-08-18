# R10-T5 — Durable Async, Search & External Effects — Reconciled Candidate

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **RENDITION/VIEWER SUBGATE ACTIVE; WHOLE T5 ADJUDICATION PAUSED**  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Decision baseline:** `wiki/architecture/rebaseline-decision-registry.md`  
> **T1 authority:** `wiki/architecture/r10-t1-semantic-state-invariants.md`  
> **T2 authority:** `wiki/architecture/r10-t2-governance-effectivity-transactions.md`  
> **T3 authority:** `wiki/architecture/r10-t3-authorization-audit-enforcement.md`  
> **T4 authority:** `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`  
> **Active subgate:** `docs/superpowers/analysis/2026-08-18-t5-rendition-viewer-strategy-evaluation.md`  
> **Implementation:** BLOCKED

T5 derives only the durable-async/Search/external-effect decisions still open after T1→T4. Existing River/outbox/notifications/search code is current-state/prior-design evidence only. T5 does not preserve runtime machinery merely because it already exists.

**Important:** the operator challenged the original implication that DOCX viewing requires a persisted `OfficialRendition` PDF. The active rendition/viewer subgate must close before T5-A→T5-P can be adjudicated as a whole. The corrected direction is that `official_rendition_render` is conditional on frozen `RequireOfficialRendition(PDF)` policy, while SourceOnly viewing remains a viewer concern rather than a durable-rendition requirement.

T5 does not define final SQL/table/index syntax, Go package/binary layout, public API/frontend routes or historical-migration execution.

---

# 1. T5 decision question

> **Which Launch effects truly require durable asynchronous execution, how must they retry/reconcile without becoming semantic authority, and what is the smallest Search projection that remains rebuildable, eventually consistent and incapable of granting access?**

---

# 2. Registry baseline — not open for aesthetic redesign

T5 MUST consume:

```text
Search = rebuildable/eventually-consistent discovery projection
Search never grants access or establishes effectivity
canonical current state/AuthZ is re-resolved before serving
notifications are delivery projection, never lifecycle authority
real required async/external work may need durable intent in same local transaction
current River/job implementation is evidence only, not target authority
no global SERIALIZABLE/global worker-lock framework
provider/external calls never join local semantic transactions
Document/Submission/Release/Audit remain semantic authority
OfficialRendition is an optional Launch Release gate only when frozen representation policy requires it
SourceOnly remains a valid representation policy
preview/viewer mechanism is not semantic authority
T4 managed content uses create-once handles + exact descriptors
T4 GC_PENDING is technical eligibility, not Records disposition
T4 provider/scanner work remains outside semantic transactions
```

T5 MUST NOT revive:

```text
generic event-sourcing/domain-event log
Audit as async bus
custom parallel schedulers/lease frameworks beside the chosen job runtime
Distribution/PeriodicReview notifications as Launch state
scheduled Release
old lifecycle notification fanout merely because it exists today
external Search engine without a demonstrated Launch consumer
provider effect receipts as a generic semantic family
universal DOCX→PDF persistence merely for viewing convenience
```

---

# 3. Evidence from prior/current implementation

Prior implementation evidence proves a real failure class: MetalDocs previously operated multiple parallel async infrastructures — River, a custom Postgres lease scheduler and a custom staging-outbox poller — with duplicated scheduling/retry/retention responsibilities. The earlier consolidation work correctly identified that as a local maximum.

Current-state evidence also shows:

```text
River already supports transaction-coupled job enqueue in the repository
old notification fanout was built on lifecycle/domain-event jobs
current Search has used document-derived projections
custom outbox/worker state and periodic jobs have accumulated operational complexity
```

These facts prove the value of **one job mechanism**, not the validity of every old job, event type, notification or outbox table.

---

# 4. Credible async alternatives

## A — generic durable event bus/outbox for every business event

Reject. Launch has no named need for replayable generic domain-event infrastructure. Audit remains action evidence, not an event bus.

## B — ad-hoc goroutines / post-commit best-effort calls

Reject for required effects. If process death can permanently lose work required by an accepted semantic transition, best-effort execution is not truthful.

## C — one Postgres-backed durable-job mechanism for named required effects

Recommended. Use one job runtime, transactionally enqueue only work whose eventual execution is required, and keep each job payload bounded to stable IDs needed to re-load canonical state.

---

# 5. Selected durable-job mechanism

T5 recommends one Postgres-backed transactional durable-job runtime for Launch. River remains the selected/reference implementation because current evidence proves it already supports the needed transaction-coupled enqueue, retry and scheduling primitives.

This is a mechanism selection, not semantic authority:

```text
River job row != business state
queue/attempt/status != Document lifecycle
worker receipt != Release/Rendition/Search truth
```

Do not add a second hand-rolled lease scheduler, staging poller or parallel outbox-dispatch framework beside the selected runtime.

Exact binary/process topology, queue names and operational configuration remain implementation design.

---

# 6. Corrected Launch durable-effect census

The rendition/viewer subgate refines the original census.

```text
always-required durable job:
  search_refresh(document_id)

conditional durable job only when the exact frozen Submission representation policy requires it:
  official_rendition_render(submission_id, required_format)

periodic reconciliation mechanism, not per-handle enqueue:
  managed-content GC over durable GC_PENDING
```

Preview/read-only viewing of PDF or DOCX is **not** itself a durable-job requirement. A SourceOnly DOCX may be viewed directly through the T6 viewer mechanism without generating/persisting a governed PDF.

No mandatory Launch job exists for:

```text
notifications
Distribution acknowledgement
Periodic Review
scheduled Release
Audit integrity/hash chain
routine offboarding IdP disable
Search crawler reconciliation on every document
```

A named future capability may add its own job later.

---

# 7. Transaction-coupled enqueue law

If a semantic transaction creates a required future effect, durable intent/job is inserted in that same local transaction.

Examples:

```text
Release / obsolescence / relevant current-effective change
  → search_refresh(document_id)

Submission with frozen RequireOfficialRendition(PDF)
  → official_rendition_render(submission_id, PDF)
```

Required invariant:

```text
business fact commits
⇔
required durable work exists
```

The enqueue mechanism joins the local PostgreSQL transaction; provider/network execution does not.

---

# 8. Conditional OfficialRendition render execution

This section applies only after the rendition/viewer subgate accepts that the frozen representation policy for this Submission requires an OfficialRendition.

Worker flow:

```text
load Submission + frozen representation requirement
→ load exact T4 source content
→ call renderer outside semantic transaction
→ create/verify READY managed-content output through T4 mechanism
→ open local transaction
→ reload Submission/current lifecycle eligibility
→ prove required rendition still absent and output bound to exact Submission
→ revalidate READY descriptor/admission proof
→ create OfficialRendition semantic record + required T3 Audit
→ if human gate also satisfied, execute system Release under T2
→ COMMIT
```

No renderer/provider call is made inside semantic transaction.

If rendering succeeds physically but the final transaction fails, output remains reclaimable managed-content mechanism state.

If the Submission no longer needs/accepts the result, the worker does not force semantic admission.

---

# 9. Rendition idempotency / at-least-once

Workers must tolerate duplicate delivery/retry.

Semantic uniqueness target:

```text
one required OfficialRendition
per exact Submission + required format
```

Two physical render attempts may transiently produce two managed handles, but only one exact eligible output may become semantic OfficialRendition; loser output is reclaimable mechanism state.

No silent fallback:

```text
RequireOfficialRendition(PDF)
+ renderer terminal failure
≠ silently treat as SourceOnly
```

The Submission remains SUBMITTED and Release remains blocked until the requirement is truthfully resolved or an authorized business action changes the lifecycle via existing rules.

---

# 10. Search projection — smallest sustainable form

Search is one rebuildable PostgreSQL-backed projection over current searchable Document facts.

Do not introduce an external Search engine in Launch without measured requirement for scale/ranking/language features the product database cannot sustainably provide.

Conceptual projection key:

```text
one row / search document keyed by Document id
```

It may contain only the facts actually needed by T6 search UX, such as current searchable title/code/type/Area/current-effective searchable text and stable IDs. Exact field set belongs T6/implementation spec after journeys are settled.

Search never owns:

```text
current EFFECTIVE truth
Authorization
Document lifecycle
exact-content identity
```

---

# 11. Search refresh — latest-state projector

Job payload:

```text
search_refresh(document_id)
```

The worker never trusts stale event payload such as “revision X became effective”. It loads the **latest canonical state** for the Document at execution time and rewrites/removes the projection to match that state.

Consequences:

```text
duplicate jobs    → harmless
out-of-order jobs → converge to latest state
retry after later lifecycle transition → converges to latest state
```

Example:

```text
Release REV001 enqueues refresh A
replacement Release REV002 enqueues refresh B
B runs first
A runs later

A still reloads current canonical state = REV002
→ projection remains REV002
```

No per-Document FIFO/ordering infrastructure is required for correctness.

---

# 12. Search freshness law

After a canonical write commits, Search may lag temporarily by **omission**:

```text
new EFFECTIVE Document not yet discoverable
```

But a stale projection hit must never be served as authority.

Before returning actionable/readable content:

```text
Search hit stable Document id
→ reload current canonical Document/Revision state
→ apply current T3 Authorization
→ serve only current allowed truth
```

Therefore:

```text
Search false negative during lag = acceptable bounded projection behavior
Search false positive granting stale access/effectivity = forbidden
```

T6 defines UX for indexing delay where material.

---

# 13. Search rebuild / reconciliation

A full Search rebuild is mandatory capability:

```text
truncate/recreate projection
→ enumerate canonical Documents
→ compute current searchable facts
→ upsert projection
→ reconciliation check/counts
```

This is the proof that Search is derivative.

Launch does not need an always-on global crawler merely to compensate for unreliable job design. Durable transaction-coupled refresh handles normal convergence; explicit rebuild/reconciliation is an operational recovery path.

A periodic low-frequency integrity check may be added later if operation evidence proves recurring drift.

---

# 14. Managed-content GC execution

T4 already persists the durable eligibility fence:

```text
GC_PENDING
```

Therefore T5 does **not** require one transactional durable job per abandoned handle.

Smallest mechanism:

```text
periodic reconciler
→ claim bounded GC_PENDING set
→ immediately re-prove no current WorkingContent / immutable governed reference / backup exclusion
→ provider DeleteReclaimable outside semantic tx
→ finalize/remove mechanism state after confirmed absence
```

If the process crashes, `GC_PENDING` itself preserves discoverability for a later run. The safe failure is leaked storage, not lost business truth.

The selected job runtime may schedule this periodic reconciliation, but no per-handle outbox is required.

---

# 15. Notifications — no Launch consumer

Current Product Contract does not require notification inbox/email/push as a Launch-Core capability.

Therefore do not preserve old:

```text
lifecycle notification events
fan-out worker
per-recipient notification inbox
notification read/unread state
```

merely because they exist in current code.

T6 worklists/current-state screens provide the required in-product discovery journeys. If a named Launch journey later proves that email/inbox delivery is required, add the smallest named job/projection then.

Launch+/Future Distribution may later require notification delivery, but that is not current backward pressure.

---

# 16. Authentication provider disable — no mandatory durable job baseline

T3 offboarding already atomically:

```text
disables User
revokes ApplicationSessions
removes GroupMemberships
drops direct RoleAssignments
```

So MetalDocs access stops even if external IdP disable is delayed or unavailable.

Provider account disable may be a post-commit defense-in-depth call. T5 does not make it a required durable job without an explicit assurance requirement that provider state must eventually converge.

If such a requirement appears, add a named `provider_subject_disable` job; do not create a generic identity-sync engine now.

---

# 17. Durable-job delivery semantics

Launch job semantics:

```text
at-least-once delivery
idempotent/revalidating workers
bounded retry/backoff
terminal/discarded failure visible to operators
no hidden infinite retry
job payload contains stable IDs + minimum immutable routing facts only
worker reloads canonical state before effect
```

Do not place full Document/Submission content or mutable authorization snapshots in job payloads.

Long-lived jobs never grant authority: current state and required technical eligibility are checked at execution/finalization.

---

# 18. Failure / terminal-state handling

Required-effect failure cannot silently disappear.

For a job with exhausted bounded retries:

```text
job remains terminally inspectable
cause/classification is observable
subject stable id is visible to operators
manual retry/re-drive is possible after cause correction
```

T5 does not create a business `FAILED` Document/Submission state merely because a mechanism failed.

Examples:

```text
rendition job terminal failure
→ Submission stays SUBMITTED / Release gate unsatisfied

search refresh terminal failure
→ canonical truth remains correct; projection may omit/stale but serve path revalidates
```

T6/ops design later decides exact operator surfaces.

---

# 19. Provider effect receipts

Reject a generic `ExternalEffectReceipt` semantic family for Launch.

Each current effect already has a natural outcome:

```text
renderer result → admitted managed content + OfficialRendition
Search effect   → Search projection row
GC effect       → provider absence + finalized mechanism state
```

Provider-specific request IDs/error codes may live in bounded operational/job telemetry where useful; they do not become domain authority.

Add a durable semantic receipt only if a future capability has a real business requirement to prove third-party acceptance, for example governed external publication.

---

# 20. Operational observability obligation

Because OfficialRendition may block Release and Search drives discovery, async runtime cannot be an invisible background mechanism.

Implementation/ops must expose enough to answer at least:

```text
is the job worker processing work?
current available/retry/terminal-failure counts by required job kind
age of oldest available/retrying required job
last successful execution / recent terminal failures
subject id/correlation needed to investigate a failed required effect
```

This is an implementation proof obligation, not a new business domain.

T5 does not freeze HTTP health endpoints, Prometheus metric names, Grafana dashboards or process layout; implementation spec chooses the smallest operational surface.

---

# 21. Future capability attack

| Future capability | T5 seam preserved |
|---|---|
| Distribution | may add named recipient-resolution/delivery jobs only when promoted; no generic bus needed today |
| Periodic Review | may add timer/surfacing jobs only when promoted |
| Dossier/Evidence | may enqueue named projections/effects without changing semantic ownership |
| Records/Hold/Disposition | may add disposition-enforcement jobs; policy remains Records authority, job is mechanism |
| Governed Export | may need durable export packaging/provider publish receipt; add explicit job/receipt then |
| External Repository | may need IMPORT/PUBLISH jobs and provider acceptance proof; no generic connector bus now |
| Training/LMS | may add delivery/sync jobs without changing current job semantics |
| Change Control | orchestration may coordinate named jobs but cannot make queue state lifecycle authority |
| pooled tenancy | job isolation/routing may reopen deployment substrate; payload remains stable IDs and canonical re-load |
| CRDT | collaboration transport is separate from durable business-effect jobs |

---

# 22. Proof obligations before implementation

Later implementation/tests must falsifiably prove at least:

```text
required search-refresh enqueue rolls back if semantic transaction rolls back
conditional required-rendition enqueue rolls back if Submission transaction rolls back
SourceOnly DOCX creates no official-rendition durable requirement merely for viewing
process crash after semantic commit cannot permanently lose a required job
at-least-once duplicate rendition job cannot create duplicate semantic OfficialRendition
renderer output is never accepted without T4 exact-content/admission proof
renderer failure cannot silently produce Release
out-of-order Search refresh jobs converge to latest canonical Document state
Search hit cannot bypass current state/AuthZ revalidation
full Search rebuild reproduces projection from canonical state
GC reconciler cannot delete current/immutable/backup-protected content
GC crash leaves durable GC_PENDING work discoverable
terminal required jobs are operator-visible and redrivable
job payload contains no durable Authorization snapshot/business-content copy
provider request IDs never become business truth
```

---

# 23. Explicit non-decisions

T5 does not decide:

```text
final River version/configuration
table names/indexes/SQL
exact queue names
exact process/binary placement
HTTP operational endpoints/metrics names
public Search API/filters/ranking UI
frontend worklists
viewer/editor implementation
renderer product selection for SourceOnly viewing or OfficialRendition generation
Historical Migration jobs
future notification UX
```

---

# 24. Reopen triggers

Reopen the implicated T5 decision only on material evidence that:

```text
multiple durable-job mechanisms are required by a proven isolation/trust boundary
external Search engine features/scale are necessary
Search omission lag violates a concrete business SLA requiring synchronous projection
a Launch notification consumer is explicitly required
provider account disable must be guaranteed eventually for assurance
managed-content GC volume requires per-object durable scheduling instead of reconciliation
per-Document Search ordering is proven necessary despite latest-state projector
future effect requires durable third-party acceptance receipt
job payload cannot remain bounded IDs without unacceptable load/coupling
```

Rendition/viewer strategy has its own active subgate and reopen criteria in `2026-08-18-t5-rendition-viewer-strategy-evaluation.md`.

---

# 25. Operator adjudication packet — PAUSED ON RV-1→RV-6

Do **not** adjudicate this packet as a whole before the rendition/viewer subgate closes.

Current corrected recommendations:

```text
T5-A ACCEPT CANDIDATE — one Postgres-backed transactional durable-job mechanism; River remains selected/reference mechanism; no parallel custom scheduler/outbox runtime.
T5-B REFINED CANDIDATE — always-required durable job = search_refresh; official_rendition_render is conditional only when frozen RequireOfficialRendition policy requires it; managed-content GC uses periodic reconciliation over durable GC_PENDING.
T5-C ACCEPT CANDIDATE — required future effect/job is enqueued in the same local semantic transaction that creates the requirement.
T5-D PAUSED ON RV SUBGATE — exact viewer vs OfficialRendition behavior must follow RV-1→RV-6; when OfficialRendition is required, render occurs outside tx and final T4 admission + semantic Rendition/Release revalidation occurs inside local tx.
T5-E PAUSED ON RV SUBGATE — required OfficialRendition rendering is at-least-once/idempotent and cannot silently fall back to SourceOnly; SourceOnly viewing itself needs no durable rendition job.
T5-F ACCEPT CANDIDATE — Search = one PostgreSQL-backed rebuildable projection keyed by Document; no external Search engine Launch baseline.
T5-G ACCEPT CANDIDATE — search_refresh(document_id) reloads latest canonical state; duplicates/out-of-order execution converge without per-document FIFO.
T5-H ACCEPT CANDIDATE — Search may lag by omission, but every hit is revalidated against current canonical state/AuthZ before serve; stale projection never grants authority/effectivity.
T5-I ACCEPT CANDIDATE — full Search rebuild/reconciliation is mandatory; always-on crawler is not baseline.
T5-J ACCEPT CANDIDATE — managed-content GC is periodic reconciliation over GC_PENDING with immediate canonical recheck before physical delete; no per-handle durable outbox required.
T5-K ACCEPT CANDIDATE — no mandatory Launch notifications/inbox/fanout/domain-event bus; add named delivery job only on concrete consumer.
T5-L ACCEPT CANDIDATE — no mandatory durable external IdP-disable job; current MetalDocs offboarding is already access-correct.
T5-M ACCEPT CANDIDATE — jobs are at-least-once, idempotent/revalidating, bounded-retry, fail-loud/terminal-visible and carry bounded stable-ID payloads.
T5-N ACCEPT CANDIDATE — no generic ExternalEffectReceipt family; effect-specific semantic outcome is authority.
T5-O ACCEPT CANDIDATE — required async mechanism must expose minimum worker/readiness/backlog/retry/terminal-failure observability.
T5-P ACCEPT CANDIDATE — future capabilities add only named jobs/effects/receipts on proven consumers; T5 never becomes a generic integration/event platform.
```

Current exact next gate:

```text
RV-1→RV-6 operator adjudication
→ incorporate accepted rendition/viewer decision
→ adjudicate corrected T5-A→T5-P
```

T5 remains non-authoritative. T6 remains NOT OPEN.