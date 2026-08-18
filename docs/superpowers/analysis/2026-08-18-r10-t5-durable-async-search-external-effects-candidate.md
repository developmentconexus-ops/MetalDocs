# R10-T5 — Durable Async, Search & External Effects — Reconciled Candidate

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — OPERATOR ADJUDICATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Decision baseline:** `wiki/architecture/rebaseline-decision-registry.md`  
> **T1 authority:** `wiki/architecture/r10-t1-semantic-state-invariants.md`  
> **T2 authority:** `wiki/architecture/r10-t2-governance-effectivity-transactions.md`  
> **T3 authority:** `wiki/architecture/r10-t3-authorization-audit-enforcement.md`  
> **T4 authority:** `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`  
> **Implementation:** BLOCKED

T5 derives only the durable-async/Search/external-effect decisions still open after T1→T4. Existing River/outbox/notifications/search code is current-state/prior-design evidence only. T5 does not preserve runtime machinery merely because it already exists.

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
OfficialRendition is a real optional Launch Release gate
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

```text
all mutations
→ domain event log
→ projector/fanout framework
```

**Reject — accidental platform.** Launch has only a few named effects. Audit already owns action evidence and must not become the bus; a second durable domain-event history would create replay/versioning/retention complexity without a second current consumer.

## B — ad-hoc goroutines/callbacks/best-effort provider calls

**Reject — incorrect for required effects.** A process crash after semantic commit could permanently lose required OfficialRendition/Search propagation work.

## C — one transactional durable-job mechanism for named effects only

```text
business transaction
  semantic truth
  required Audit
  required durable job enqueue
COMMIT

worker
  load canonical current truth
  perform effect idempotently
  finalize through proper owner use case
```

**Recommended Global Maximum.**

No generic event platform is introduced. Only a named effect that must survive process restart gets a durable job.

---

# 5. One durable job runtime — T5-A

T5 recommends one Postgres-backed transactional durable-job mechanism for Launch.

The existing River framework is retained as the **selected/reference implementation** because current repository evidence already proves the needed primitive set and because keeping it avoids rebuilding queue, retry, scheduling and lease machinery.

Architectural law:

```text
River/job runtime = mechanism
job row            = durable work intent, not domain state
job completion     != semantic business completion
```

Do not operate a second custom lease scheduler, generic outbox dispatcher or custom retry framework beside it.

If a later implementation replaces River, the replacement must preserve the same T5 contract; River identity never enters domain semantics.

---

# 6. Named Launch durable job classes — T5-B

T5 finds only two mandatory durable async classes in Launch Core:

```text
1. official_rendition_render
2. search_refresh
```

Plus one bounded technical periodic reconciliation:

```text
3. reclaim_gc_pending_managed_content
```

The GC item is **not** one durable job per handle; durable `GC_PENDING` mechanism state already survives crashes, so one periodic/reconciling worker can repeatedly discover eligible rows.

Not mandatory durable async Launch work:

```text
malware scan before SUBMIT
  → synchronous/retriable preflight; no semantic state waits on a background scan

browser upload/admission
  → request-driven T4 mechanism

provider account disable on User offboarding
  → local User-disabled/session/grant teardown already fail-closes MetalDocs access;
     no mandatory external-IdP convergence job absent a separate product/ops requirement

notifications
  → no accepted Launch consumer; deferred below
```

---

# 7. Transactional enqueue law — T5-C

When a semantic transition creates a required future effect, durable enqueue occurs in that same local PostgreSQL transaction.

```text
BEGIN
semantic mutation
required T3 Audit
required durable job enqueue
COMMIT
```

If enqueue fails, the semantic transition that depends on guaranteed future work does not report success.

Job payloads carry only stable routing identities and bounded mechanism configuration, for example:

```text
Submission id
Document id
required ContentFormat
```

They do not freeze a second copy of mutable business truth. Workers re-read canonical current state before acting.

---

# 8. Official Rendition execution — T5-D

A Submission whose snapshotted representation policy requires an OfficialRendition transactionally enqueues:

```text
official_rendition_render(submission_id, required_format)
```

Enqueue occurs when the immutable Submission is created. Rendering may therefore run while human governance proceeds; Release remains governed by T2's independent human + representation gates.

Worker path:

```text
load exact immutable Submission + T4 source descriptor/handle
→ verify the Submission still requires this rendition
→ call renderer outside semantic transaction
→ renderer output enters T4 managed-content admission as TRUSTED_INTERNAL_DERIVATION
→ server derives exact output descriptor / READY
→ BEGIN semantic transaction
→ serialize Document under T2
→ re-prove Submission/representation requirement is still eligible
→ create exactly one OfficialRendition for required format
→ required T3 Audit
→ re-evaluate Release gates
→ if all gates satisfied, system Release may complete in same transaction
→ COMMIT
```

If the attempt was returned/withdrawn or otherwise no longer eligible before final admission, provider output never creates a Rendition; it remains reclaimable mechanism content.

No provider call joins the final semantic transaction.

---

# 9. Rendition idempotency / retry semantics — T5-E

The runtime is at-least-once. Therefore rendering must be idempotent at the semantic boundary.

Conceptual identity:

```text
one successful semantic OfficialRendition
per (Submission, required ContentFormat)
```

Duplicate/retried jobs may render more than once physically, but only one eligible semantic success can commit. Extra READY outputs are reclaimable mechanism content.

Failures:

```text
transient renderer/provider failure → retry with bounded backoff
permanent unsupported/invalid render → terminal/discarded operational failure
```

A terminal render failure leaves the truthful Revision `SUBMITTED`; it never fabricates Release or falls back to SourceOnly when the snapshotted policy requires a Rendition.

Operational failure must be visible; product/admin remediation may retry after the cause is corrected.

---

# 10. Search projection boundary — T5-F

Search remains a **rebuildable mechanism projection**, not a semantic owner.

Launch target:

```text
one PostgreSQL-backed search projection
keyed by stable Document identity
containing only the current search fields required by T6
```

T5 deliberately does not freeze the final searchable-field list; T6 owns the actual search/read journey.

Projection rules:

```text
no User/Group ACL materialization
no permission cache
no effectivity authority
no lifecycle mutation
no history ownership
no external Elasticsearch/OpenSearch dependency without proven scale/feature need
```

PostgreSQL is the Launch default search substrate because no current requirement proves a separate search service.

---

# 11. Search refresh propagation — T5-G

Normal propagation uses one transactionally enqueued job:

```text
search_refresh(document_id)
```

At minimum enqueue on:

```text
Release that establishes/replaces EFFECTIVE content
successful governed obsolescence
```

T6 may add enqueue call sites for other current fields it actually exposes in Search (for example a searchable display field), without changing the T5 mechanism.

The job payload carries only `document_id`.

Worker always loads **latest canonical truth at execution time** and then:

```text
current EFFECTIVE exists + searchable under canonical product rules
  → upsert projection from latest canonical state

no current EFFECTIVE / OBSOLETE
  → remove projection row
```

Because every job reads latest state, out-of-order duplicate refresh jobs are commutative/idempotent; an old job cannot rewrite the projection to an old Revision merely because it ran late.

---

# 12. Search freshness / serving law — T5-H

Normal Search is eventually consistent.

Allowed transient condition:

```text
newly EFFECTIVE document may be temporarily missing from Search
```

Never allowed as product truth:

```text
stale Search hit grants access
stale hit proves current effectivity
stale indexed content is served without canonical revalidation
```

Before a hit is returned/used as an actionable document result, the request path must re-resolve current canonical Document state and current T3 Authorization. T6 owns the exact query/response choreography.

A stale projection row for a now-obsolete/superseded document is therefore filtered/dropped when canonical revalidation fails.

Search may degrade by omission during lag; it may not degrade by granting stale authority.

---

# 13. Search rebuild / reconciliation — T5-I

Search must be derivable entirely from canonical product state.

Required operational capability:

```text
full rebuild:
  clear/rebuild projection from canonical current-effective documents
```

Rebuild is used after:

```text
projection corruption
projection-schema/algorithm change
restore/cutover
known propagation defect
```

A permanent always-on reconciliation crawler is **not** mandatory Launch baseline because transactional refresh jobs already provide normal propagation. Add periodic reconciliation only if operational evidence proves it is needed.

Rebuild tooling may run through the same durable-job runtime but its state remains rebuildable mechanism only.

---

# 14. Managed-content GC execution — T5-J

T4 already owns the correctness fence:

```text
GC_PENDING
+ no current WorkingContent reference
+ no immutable governed reference
```

T5 therefore rejects a second per-object durable-outbox family.

One periodic/reconciling worker is sufficient:

```text
scan bounded GC_PENDING batch
→ re-read/re-prove T4 eligibility immediately before provider delete
→ DeleteReclaimable(handle)
→ on success remove/finalize mechanism row
→ on transient failure leave GC_PENDING for future retry
```

Provider age/listing never authorizes deletion.

A failure only leaks reclaimable storage; it does not change semantic truth. This lower criticality is why per-handle guaranteed enqueue is unnecessary.

---

# 15. Notifications — T5-K

T5 finds **no accepted Launch-Core notification consumer** in the current Product Contract/T1→T4 authority.

Therefore Launch target contains no mandatory:

```text
Notification semantic owner
notification inbox state
lifecycle-notification fanout
notification domain-event bus
email/push dispatcher
```

Users can discover current work through canonical T6 worklists/journeys over governance state.

Distribution/Read&Acknowledge and richer notification behavior remain Launch+/Future. Prior notification designs stay evidence only.

If a concrete Launch notification requirement is later added, it must define its trigger/recipient/delivery truth and may reuse the T5 durable-job mechanism without retroactively turning existing lifecycle events into a generic bus.

---

# 16. Provider identity/offboarding effects — T5-L

T3 local offboarding is sufficient to stop MetalDocs access:

```text
User disabled
Sessions revoked
memberships/direct grants removed
```

Therefore T5 does not require a durable external IdP-disable job in Launch correctness.

A post-commit provider-disable call may be operational defense-in-depth. If product/ops later requires eventual provider-account convergence, add a bounded named job then; do not create a generic provider-sync engine now.

---

# 17. Job execution semantics — T5-M

All durable jobs obey:

```text
at-least-once execution
idempotent semantic finalization
bounded retry/backoff
terminal failure/discard visibility
no silent success on unknown job kind
stable IDs in payload; mutable business state re-read canonically
no secrets/large governed content copied into job payload
no job becomes semantic history authority
```

Concurrency is narrow and effect-specific. Do not introduce a global worker lock or global SERIALIZABLE posture.

The selected job runtime owns its own claim/lease/retry housekeeping; application code must not build a second scheduler/lease/reaper framework around it.

---

# 18. No generic external-effect receipt family — T5-N

T5 does not introduce a universal `ExternalEffectReceipt` semantic table.

Success proof is already owned where meaning exists:

```text
renderer effect → OfficialRendition semantic fact
Search effect   → rebuildable projection row
GC delete       → mechanism state/object absence
```

Job runtime history is operational evidence only.

A future external effect that requires a durable business-facing receipt may add one in its owning capability. Do not prebuild a polymorphic receipt registry.

---

# 19. Async operational visibility — T5-O

Because a stalled renderer can indefinitely block a required Release, the durable-job runtime must be operationally observable.

Implementation proof must expose enough health/metrics/log correlation to answer at least:

```text
is the durable-job executor alive/ready?
how many jobs are available/retrying/terminal-failed by kind?
what is the age of the oldest required rendition/search job?
when did each required worker kind last succeed?
which correlation/semantic IDs identify a failed job?
```

Exact Prometheus/HTTP endpoint/logging topology is implementation-spec work, but a required async effect may not rely only on an unobserved background process.

No generic business dashboard is implied.

---

# 20. Future capability attack — T5-P

| Future capability | T5 seam preserved |
|---|---|
| Distribution | may add named fanout/delivery jobs without making jobs the Distribution authority |
| Periodic Review | due surfacing may use scheduled jobs/projection while review truth remains domain-owned |
| Dossier/Evidence | may add named projection/effect jobs while semantic owners remain separate |
| Records/Hold/Disposition | physical disposition may use durable jobs only after Records authority creates eligibility/fence |
| Governed Export | package assembly/delivery may use named durable jobs; export semantic request/receipt stays owned by future capability |
| Repository connector | import/publish effects may add bounded jobs/receipts; no generic integration bus required now |
| Training/LMS | may add named delivery/sync work without broadening document roles |
| Change Control | orchestration may enqueue named document operations but cannot bypass each Document's domain transactions |
| pooled tenancy | job isolation/routing may reopen; job mechanism remains non-semantic |
| CRDT | collaboration transport is separate from durable business-effect queue |

Future job kinds are introduced only with a named consumer; adding a capability does not justify a generic event platform retroactively.

---

# 21. Proof obligations before implementation

Later implementation design/tests must falsifiably prove at least:

```text
required rendition job is transactionally enqueued with Submission creation
semantic Submission can never be inferred from job existence
renderer duplicate/retry yields at most one semantic OfficialRendition per required format
returned/withdrawn ineligible Submission cannot gain late Rendition/Release from stale job
renderer terminal failure leaves truthful SUBMITTED state
no provider call occurs inside Rendition/Release semantic transaction
search refresh job always re-reads latest canonical Document state
out-of-order search refresh cannot restore an old Revision into projection
stale Search hit never bypasses canonical effectivity/AuthZ checks
full Search rebuild yields the same projection as canonical current-effective truth
GC worker cannot delete current WorkingContent or governed content
GC transient failure leaves durable GC_PENDING eligibility for retry
no Launch notification fanout/event bus exists without a named requirement
User offboarding blocks MetalDocs access even if external IdP disable never completes
job runtime retries are bounded and terminal failures are operationally visible
unknown job kind fails loud
no custom second lease/scheduler/retry framework exists beside selected runtime
```

---

# 22. Explicit non-decisions

T5 does not decide:

```text
final River version/config values
final job table/index names
exact queue names or binary/process topology
exact retry counts/backoff durations/retention periods
exact Prometheus/health endpoint schema
final Search columns/query grammar/ranking
final PostgreSQL FTS/trigram index definition
public worklist/Search/API routes
frontend notification UX
Historical Migration execution jobs
future Distribution/Records/Repository job catalogs
```

---

# 23. Reopen triggers

Reopen only the implicated T5 seam on material evidence that:

```text
River cannot satisfy the required transactional/retry/scheduling contract economically
Launch gains a concrete notification requirement
Search scale/linguistic/ranking needs exceed PostgreSQL projection capability
a second real consumer requires durable replayable domain-event history
GC volume makes periodic reconciliation insufficient
provider-account disable becomes a contractual/security convergence requirement
a future external effect needs a business-facing durable receipt
required async operational SLA requires stronger queue partitioning/priority semantics
```

---

# 24. Operator adjudication packet

Recommended dispositions:

```text
T5-A ACCEPT — one Postgres-backed transactional durable-job mechanism; retain River as selected/reference implementation, never semantic authority; no parallel custom scheduler/outbox/retry framework.
T5-B ACCEPT — mandatory Launch async classes are official_rendition_render + search_refresh; managed-content GC uses periodic reconciliation over durable GC_PENDING rather than one durable job per handle.
T5-C ACCEPT — a required future effect is transactionally enqueued in the same local semantic transaction; job payload carries stable IDs/bounded routing only and workers re-read canonical state.
T5-D ACCEPT — required OfficialRendition render job is enqueued with immutable Submission creation; render outside tx, T4-admit exact output, then local tx creates Rendition and may complete Release after revalidating T2 gates.
T5-E ACCEPT — rendition runtime is at-least-once/idempotent with one semantic success per (Submission, required format); transient retry, permanent visible terminal failure; never fall back to SourceOnly.
T5-F ACCEPT — Search remains one PostgreSQL-backed rebuildable projection keyed by Document; exact searchable fields wait T6; no embedded user ACL/permission authority and no external search service baseline.
T5-G ACCEPT — transactionally enqueue search_refresh(document_id) at least on Release/obsolescence; worker always reads latest canonical truth, making duplicates/out-of-order execution idempotent/commutative.
T5-H ACCEPT — Search may lag by omission, but stale hits never grant access/effectivity; canonical current state + T3 AuthZ are revalidated before actionable result/serve.
T5-I ACCEPT — Search supports full rebuild from canonical current-effective truth; no mandatory always-on reconciliation crawler unless evidence proves need.
T5-J ACCEPT — GC worker periodically reconciles bounded GC_PENDING rows, re-proves T4 eligibility immediately before provider delete, and retries by leaving GC_PENDING; no per-handle durable outbox family.
T5-K ACCEPT — no mandatory Launch notifications/inbox/fanout/domain-event bus; defer until a named consumer appears.
T5-L ACCEPT — no mandatory durable external IdP-disable job; T3 local offboarding already fail-closes access; provider convergence remains optional defense-in-depth/reopen trigger.
T5-M ACCEPT — all jobs are at-least-once, idempotent, bounded-retry, fail-loud/terminal-visible, ID-only/bounded payload, canonical-reload workers; selected runtime owns leases/retry housekeeping.
T5-N ACCEPT — no generic ExternalEffectReceipt family; semantic result/mechanism state owns success evidence; add bounded receipt only with a future named consumer.
T5-O ACCEPT — required async runtime must expose minimal readiness/backlog/retry/terminal-failure/age/correlation observability; exact telemetry surface is implementation work.
T5-P ACCEPT — future capabilities add only named jobs/effects without turning T5 into a generic integration/event platform.
```

T5 remains non-authoritative until operator adjudication. After technical adjudication, **T6 still must not open**: the mandatory platform-facing T5 summary must be presented and explicitly ratified first.