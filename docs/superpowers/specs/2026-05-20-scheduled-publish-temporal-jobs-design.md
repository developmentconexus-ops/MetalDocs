# Scheduled Publish Temporal Jobs Design

> Date: 2026-05-20
> Status: proposed and user-approved for plan writing
> Scope: migrate scheduled publish ownership from the API process to a dedicated temporal jobs runtime using River, while preserving current domain invariants and removing the legacy runtime path completely at cutover

## 1. Purpose

MetalDocs currently executes scheduled publish cutover (`scheduled -> published`) from an embedded scheduler registered inside the API process. QA validated that the modern `/documents/:id` + approval flow does not trigger publish from frontend refresh; instead, the cutover is owned by API-hosted runtime polling. This is below the desired SaaS hosting boundary because the API process remains the principal owner of temporal execution.

The goal of this design is to move scheduled publish to a dedicated temporal jobs runtime with professional ownership boundaries:

- API owns validation, authz, OCC, and persistence of scheduling intent.
- A dedicated temporal jobs host owns clock/runtime execution.
- Read paths remain side-effect free.
- Domain truth remains DB-driven and concurrency-safe.
- Legacy hosting is fully removed after cutover; no dual runtime remains.

## 2. Problem Statement

Current runtime truth:

- `effective-date-publisher` is registered in `apps/api/cmd/metaldocs-api/main.go`.
- Execution calls `internal/modules/documents/approval/application/SchedulerService.RunDuePublishes`.
- The current `apps/worker/cmd/metaldocs-worker/main.go` owns outbox/PDF work, not scheduled publish.
- Local operational memory is vulnerable to confusion because API and worker startup happen together while ownership of temporal cutover still lives in API.

This creates an architecture hosting gap:

- the API is the owner of the temporal boundary
- outbox worker and temporal ownership are semantically mixed in operations memory
- polling cadence is coupled to API runtime concerns
- the current runtime shape does not match the desired professional SaaS boundary

## 3. Design Principles

### 3.1 Runtime truth first

The design follows current hardened domain behavior. This is not a rewrite of publish rules. It is a hosting-boundary migration.

### 3.2 DB truth remains canonical

`documents.status`, `effective_from`, lineage/supersede fields, and governed approval state remain the source of truth. The queue/runtime does not become the business source of truth.

### 3.3 Job execution must be idempotent

Queue uniqueness and scheduling reduce duplicates operationally, but correctness depends on transactional revalidation against current DB state at execution time.

### 3.4 Read paths remain safe

No read endpoint, page refresh, or query freshness loop may trigger publish.

### 3.5 No legacy fallback at completion

Once the new runtime is cut over, the API-hosted scheduler path for scheduled publish must be removed completely. No hidden duplicate scheduler, no temporary legacy fallback left behind.

## 4. Market-Aligned Architecture Choice

Three categories were evaluated:

1. Dedicated temporal runner over current Postgres/leasing infrastructure
2. River-backed dedicated temporal jobs host
3. Workflow engine / external scheduler platform (Temporal, Cadence, cloud scheduler patterns)

Recommendation:

- adopt `River` as the substrate for the new temporal jobs runtime
- host it in a dedicated process, separate from API and separate from the outbox/PDF worker
- migrate scheduled publish as the first temporal job

Why this is the recommended choice:

- aligns with the existing Go + Postgres stack
- preserves DB-driven domain truth
- avoids introducing Redis only for this workflow
- avoids premature workflow-platform adoption
- uses a market-validated scheduling/runtime model instead of expanding custom scheduler ownership

Workflow engines like Temporal remain compatible as a future direction if MetalDocs later accumulates many long-running human/workflow timers, compensations, or orchestration-heavy flows. They are not the recommended first move for this problem.

## 5. Target Boundary

### 5.1 API ownership

The API will own:

- schedule request validation
- authz and OCC checks
- persistence of `scheduled` state and `effective_from`
- persistence of schedule metadata needed for safe execution
- transaction-safe enqueue of a single temporal job

The API will not own:

- polling for due schedules
- runtime clock ownership for cutover
- background execution of `scheduled -> published`

### 5.2 Temporal jobs host ownership

The new `metaldocs-jobs` process will own:

- River worker runtime
- temporal queue hosting
- execution of scheduled publish jobs
- retries, job observability, and backlog visibility
- future temporal jobs in the same boundary

The jobs host will not become a domain owner. It invokes existing application/domain behavior under a new runtime boundary.

### 5.3 Outbox worker ownership

The existing outbox/PDF worker remains dedicated to:

- outbox consumption
- PDF/docgen fanout
- messaging-related retries

It must not own scheduled publish after the migration.

## 6. Runtime Shape

### 6.1 New process

Create a dedicated process/binary:

- recommended name: `metaldocs-jobs`

The naming intentionally avoids `worker` to prevent repeating the current semantic ambiguity between messaging workers and temporal job execution.

### 6.2 Queue model

Initial runtime organization:

- queue: `temporal`
- first job kind: `scheduled_publish_cutover`

The boundary is generic for future temporal jobs, but only one real business job is migrated in this change.

### 6.3 Execution model

Recommended primary model:

- one scheduled River job per schedule request
- job is enqueued transactionally at schedule time with `scheduled_at = effective_from`

Rejected as primary model:

- periodic DB scanner as the main engine

Reason for rejection:

- scanning reintroduces runtime polling as the principal mechanism
- per-schedule jobs provide cleaner traceability and more precise cutover timing
- River already provides a market-validated execution-at-time boundary

Optional future safety net:

- a low-frequency reconciler may be added later as an operational defense if needed
- it must not become the primary execution engine

## 7. Domain Model Adjustment

The source of truth remains in `documents`, but the design adds a small explicit discriminant for stale-job detection.

### 7.1 Required schema change

Add to `documents`:

- `schedule_generation` bigint/int, non-null, monotonic

Purpose:

- increments on each schedule or reschedule
- allows old jobs to prove they are stale and exit safely

### 7.2 Existing truth retained

Retain current truth in `documents`:

- `status = 'scheduled'`
- `effective_from`
- governed lineage / supersede fields

### 7.3 Optional observability-only metadata

Optional, not required for correctness:

- `scheduled_job_id`
- `scheduled_job_key`

These may aid debugging but must not be the primary correctness mechanism.

## 8. Job Payload Contract

Recommended minimum payload:

- `tenant_id`
- `document_id`
- `expected_revision_version`
- `scheduled_effective_at`
- `schedule_generation`

The payload is intentionally small. It carries enough state to validate freshness and execute the cutover safely, while DB truth remains canonical.

## 9. Execution Semantics

When the job runs, the handler must begin a transaction and revalidate current truth before mutating state.

Required validation order:

1. load current governed document state
2. confirm `status == scheduled`
3. confirm `schedule_generation` matches the payload
4. confirm `effective_from` still matches the scheduled intent
5. confirm revision/OCC assumptions still hold
6. confirm supersede target/head invariants still hold

If validation fails because the schedule is stale, cancelled, already executed, or superseded by a newer scheduling generation:

- exit as successful no-op
- do not retry indefinitely
- do not mutate state

If validation passes:

- transition `scheduled -> published`
- supersede the recorded current published head when applicable
- emit governed events
- commit atomically

## 10. Schedule Lifecycle Rules

### 10.1 Initial schedule

On schedule request:

- increment `schedule_generation`
- persist `status='scheduled'`
- persist `effective_from`
- enqueue a River job in the same transaction

### 10.2 Reschedule

On reschedule:

- increment `schedule_generation`
- update `effective_from`
- enqueue a new River job transactionally

The old job is allowed to exist physically. It becomes logically stale because its `schedule_generation` no longer matches current truth.

### 10.3 Cancel schedule

On cancel:

- clear or transition out of the `scheduled` state according to current domain rules
- increment or otherwise invalidate `schedule_generation` as needed by the final implementation

Any older scheduled job that still fires must exit as a no-op.

### 10.4 Duplicate delivery / retry

If the same job runs more than once:

- current state revalidation prevents double publish
- post-publish state makes later replays no-op

Correctness does not depend solely on queue uniqueness.

## 11. Legacy Removal Rule

The migration is not complete until all legacy scheduled publish runtime ownership is removed.

Must be removed:

- API registration of `effective-date-publisher`
- API-hosted scheduled publish polling/ownership path
- any local startup or operational documentation that implies the API or outbox worker owns scheduled publish

Must not remain:

- hidden fallback scheduler in the API
- duplicated scheduler in two processes
- “temporary” dual-running ownership after final cutover

This is a hard-cutover requirement, not a polish item.

## 12. Implementation Outline

### Phase A: Introduce jobs runtime

- add River dependency and bootstrap
- create `metaldocs-jobs` process
- add local startup support for the new process

### Phase B: Introduce schedule generation model

- add DB migration for `documents.schedule_generation`
- update schedule write path to persist/increment generation

### Phase C: Enqueue per-schedule jobs transactionally

- update `schedule-publish` write path to enqueue River job in the same DB transaction

### Phase D: Implement scheduled publish job handler

- reuse hardened application/domain services where possible
- perform transactional revalidation and cutover

### Phase E: Remove legacy ownership

- remove API registration of scheduled publish job
- remove API-hosted temporal ownership for this flow

### Phase F: Sync operational memory

- update module docs and startup docs
- document `metaldocs-jobs` as the owner of scheduled publish
- explicitly document that the outbox worker is not the temporal owner

## 13. Verification Requirements

The migration is not complete without fresh verification evidence.

### 13.1 Domain/runtime tests

- schedule request persists `scheduled`, `effective_from`, and increments `schedule_generation`
- schedule request enqueues River job transactionally
- due job publishes successfully
- supersede of prior published head still occurs consistently
- replay/duplicate execution is a no-op
- stale job after reschedule is a no-op
- cancelled schedule job is a no-op

### 13.2 Hosting/runtime tests

- `metaldocs-jobs` boots and processes the temporal queue
- API no longer registers `effective-date-publisher`
- local startup makes ownership explicit
- outbox worker does not claim scheduled publish ownership

### 13.3 Contract and regression checks

- no read path triggers publish
- authz/OCC/lineage invariants remain intact
- no fallback legacy code path survives after cutover

## 14. Wiki and Operational Memory Updates

Required sync after code truth changes:

- `wiki/modules/approval.md`
- `wiki/modules/documents.md`
- `wiki/references/local-dev-startup.md`
- any relevant worker/jobs operational docs
- module sync logs for affected documented modules

Required facts to update:

- API is no longer the temporal runtime owner
- `metaldocs-jobs` owns scheduled publish cutover
- outbox worker owns outbox/PDF only
- scheduled publish uses per-schedule jobs, not API polling

## 15. Risks and Non-Goals

### Risks

- partial migration that leaves legacy ownership alive
- stale-job handling implemented inconsistently
- operational memory drift that causes local/dev confusion

### Non-goals

- adopting Temporal/Cadence in this change
- redesigning publish business rules
- moving outbox/PDF worker into the new temporal boundary
- introducing a separate schedule aggregate/table unless a later need proves it necessary

## 16. Success Criteria

This design is successful when all of the following are true:

- scheduled publish ownership is hosted in `metaldocs-jobs`, not API
- API only persists intent and enqueues work transactionally
- read paths remain side-effect free
- DB truth remains canonical
- reschedule/cancel/replay are safe and idempotent
- outbox worker and temporal runner are no longer semantically mixed
- all legacy runtime ownership for scheduled publish is removed after cutover
