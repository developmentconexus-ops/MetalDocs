# Stage 2 Evaluation — Async Runtime & Audit/Data Integrity

> **Theme:** async-integrity
> **Author:** Stage-2 evaluator agent
> **Date:** 2026-06-11
> **Scope:** F-07 + D-01, F-04, F-19 from `wiki/backend/legacy-register.md`
> **Standards:** Transactional Outbox pattern (microservices.io / Chris Richardson), at-least-once delivery, River job queue, ISO 9001:2015 §8.5.1 (controlled conditions), 21 CFR Part 11 §11.10(e) (audit trail integrity), OWASP ASVS 2.10 (service credential management)
> **Target requirements cited:** REQ-ASYNC-1, REQ-ASYNC-2, REQ-ASYNC-3, REQ-ASYNC-4, RF-7

---

## How to read this document

Each finding section follows the same structure:

1. **Current state** — what the code actually does, confirmed against the file:line anchors in the register.
2. **Standard** — the external reference the finding is judged against, with citation.
3. **Verdict + rationale** — one of KEEP / SIMPLIFY / REFACTOR / DELETE / DEFER with the reasoning.
4. **Smallest correct fix** — the minimum change that reaches the professional bar.
5. **Effort / blast radius** — S/M/L and whether the change is contained, module-scoped, cross-module, or system-wide.
6. **ADR needed?** — yes/no with reason.
7. **Over-engineering check** — explicit note when a heavier fix than stated would be over-engineering for this system's scale.

---

## F-07 + D-01 — Post-commit audit/governance log (atomicity gap)

### Current state

Confirmed by direct code reading.

**Taxonomy (representative sample):**
`internal/modules/taxonomy/application/area_service.go:195-208` — `tx.Commit()` at line 195, `s.govLogger.Log(...)` at line 199. The `committed = true` guard means the log call only happens after a successful commit. Identical structure confirmed at `family_service.go:157-177` and `profile_service.go:190-204`. The `DBGovernanceLogger` (`taxonomy/application/governance_logger.go:21-44`) executes a bare `db.ExecContext` — no transaction, no retry, no outbox. Any crash, OOM, or context cancellation between line 195 (commit) and line 199 (log) drops the governance event permanently.

**Templates:**
`internal/modules/templates/application/lifecycle.go:70-92` — `tx.Commit()` at line 70, `s.repo.AppendAudit(ctx, audit)` at line 92. `AppendAudit` resolves to `r.audit.Record(ctx, event)` at `repository/postgres.go:631-639`, which is a direct `INSERT INTO metaldocs.audit_events` outside any transaction. Same drop window.

**Documents-core:**
`internal/modules/documents/application/service.go:803-808` (ForceReleaseSession) and `:843-848` (Archive) — `repo.ForceReleaseSession` / `repo.MarkArchived` commit internally; `s.audit.Write(...)` is called after the repo call returns. Same pattern.

**Approval:**
`internal/modules/documents/approval/application/decision_service.go:545` — governance event via `s.emitter.Emit(ctx, tx, event)` is called **inside** the transaction, before `tx.Commit()` at line 561. This is the correct pattern. The approval module does **not** have the atomicity gap for its primary governance event.

**Read/write sink split (templates):**
`AppendAudit` writes to `metaldocs.audit_events` (`repository/postgres.go:639`). `ListAudit` reads from `templates_audit_log` (`repository/postgres.go:676-714`). Events written via `AppendAudit` after the migration are invisible to the read endpoint. This is a separate consistency defect layered on top of the atomicity gap.

**Deprecated sink active as nil fallback:**
`taxonomy/application/governance_logger.go:17` — `NewDBGovernanceLogger` is marked `// Deprecated`. `internal/modules/controlleddocuments/module.go:31` imports it as a fallback when `AuditWriter` is nil. This means controlled-documents governance events go to the legacy `governance_events` table when the canonical audit writer is absent.

**D-01 cross-cut summary:**
No module using post-commit `govLogger.Log` or `s.audit.Write` has atomic outbox semantics. The approval module is the exception: its primary event is inside the transaction. Five distinct modules exhibit the gap; the taxonomy module is the highest-exposure instance (11 event types, all post-commit).

### Standard

**Transactional Outbox pattern** (Chris Richardson, microservices.io/patterns/data-management/transactional-outbox.html): "Instead of directly publishing messages to a message broker, the domain logic uses a database transaction to atomically update the database and insert messages into an outbox table." The core invariant: a message is either committed together with the state change or not committed at all. Any write outside the transaction violates this invariant and creates a silent-loss window.

**ISO 9001:2015 §8.5.1** requires that controlled production processes retain documented information demonstrating the process was carried out as planned. For a QMS-regulated system, audit events are the documentary evidence of controlled production operations. A droppable audit event means the QMS record is incomplete — a non-conformity under an ISO 9001 audit.

**21 CFR Part 11 §11.10(e)** (electronic records / audit trails, FDA): "Use of secure, computer-generated, time-stamped audit trails to independently record the date and time of operator entries and actions that create, modify, or delete electronic records." "Independently record" means the record must survive any single point of failure in the creating process. A post-commit fire-and-forget write is not independently durable.

**OWASP ASVS V10 (Malicious Code / Logging)** and ASVS V7 (Error Handling and Logging): log/audit records must be generated in a way that they cannot be trivially lost. Post-commit writes with no retry mechanism fail this bar.

### Verdict: REFACTOR — P1

The post-commit pattern is a correctness gap, not a latency or style issue. For taxonomy, templates, and documents-core mutations, the smallest correct fix is to move the audit/governance write inside the existing transaction — an in-transaction `INSERT INTO governance_events` or `INSERT INTO audit_events` rather than the post-commit call. This is not an outbox migration; the event sinks are already Postgres tables in the same database, so in-transaction insert is both simpler and more correct than adding an outbox stage.

Conditions for this verdict:
- The `govLogger.Log` and `audit.Write` calls write to Postgres tables in the same database as the committing transaction. There is no cross-service call involved.
- The approval module already proves this works: `emitter.Emit(ctx, tx, event)` at decision_service.go:545 is inside the transaction.
- Moving these writes inside the transaction requires: (a) adding a `Tx` variant to the governance logger and audit writer interfaces, and (b) passing the transaction handle into the service call. This is a bounded refactor, not a system redesign.

**The register's severity (high) is correct.** For a QMS product under ISO 9001 or 21 CFR Part 11 audit, a droppable audit trail is a compliance defect, not just a reliability concern.

The read/write sink split in templates (`AppendAudit` → `audit_events`, `ListAudit` → `templates_audit_log`) is a separate bug. The fix is to migrate `ListAudit` to read from `audit_events`. This requires a migration to backfill or accept the seam, but the in-flight code path should be unified first.

The deprecated `DBGovernanceLogger` active as a nil fallback is a dead-code hygiene issue; it should be removed along with the `AuditWriter` nil guard in `controlleddocuments/module.go`.

### Smallest correct fix

1. **Taxonomy / templates / documents-core:** Add `LogTx(ctx, tx, event)` to the `GovernanceLogger` interface and `RecordTx` to the `AuditWriter` interface (both interfaces likely already have `Tx` variants given `AppendAuditTx` exists at `repository/postgres.go:642`). Rewrite each post-commit call to pass the active `tx` handle. No new tables, no new goroutines.
2. **Templates read/write split:** Change `ListAudit` to query `audit_events` filtered by `resource_type='template'` and `resource_id=$templateID`. Accept the historical seam (old rows remain in `templates_audit_log`; a one-time backfill migration is optional and can be deferred).
3. **Deprecated logger cleanup:** Remove `DBGovernanceLogger` import from `controlleddocuments/module.go`. Wire the canonical `AuditWriter` unconditionally.

### Effort / blast radius

**Effort: M** (taxonomy has 11 service methods to update; templates and documents-core have ~5 combined; the interface changes ripple through tests).
**Blast radius: cross-module** (taxonomy, templates, documents-core; approval is already correct and needs no change).

### ADR needed?

No. The fix is not a design decision — it is restoration of the transactional-outbox invariant already established in the approval module and documented in ADR 0009. A PR description citing REQ-ASYNC-1, ISO 9001 §8.5.1, and the approval module as the canonical example is sufficient.

### Over-engineering check

The obvious over-engineered alternative is to route every governance event through the platform outbox worker. That would add a relay stage (domain write → outbox row → relay worker → governance_events) and increase latency and complexity. It is not warranted here because: (a) the event sinks are in the same Postgres instance as the business data, so in-transaction INSERT is available and is the simpler path; (b) the outbox pattern adds value when the event consumer is external or unreliable — governance_events is neither. In-transaction INSERT is the correct and simplest fix.

---

## F-04 — Duplicate outbox worker/repository clones in render pipeline

### Current state

Confirmed by direct code reading.

**Worker clones:**
`internal/modules/render/fanout/pdf_outbox_worker.go` and `internal/modules/render/fanout/materialize_outbox_worker.go` are structurally identical: same field set (`repo`, `pub`, `pollEvery`, `batchSize`, `maxAttempt`, `staleAfter`, `log`), same `Run` loop (ticker, `ResetStaleClaims`, `ClaimPending`, `dispatchOne`), same backoff formula (`min(30min, 30s * 2^cappedAttempts)`). The only differences are the concrete repo type, the `EventType` constant, and the `Payload` struct populated in `dispatchOne`. Combined ~200 lines, ~95% identical.

**Repository clones:**
`pdf_outbox_repository.go` (~162 lines) and `materialize_outbox_repository.go` (~159 lines) differ only in table name (`metaldocs.pdf_dispatch_outbox` vs `metaldocs.materialize_dispatch_outbox`) and row type (`OutboxRow` vs `MaterializeOutboxRow`). All six methods (`Enqueue`, `ClaimPending`, `MarkDispatched`, `MarkFailed`, `ReadState`, `ResetStaleClaims`) are verbatim copies with the table name substituted.

**Two-stage outbox relay chain:**
The render fanout staging tables (`pdf_dispatch_outbox`, `materialize_dispatch_outbox`) feed into `outbox_events`, which `apps/worker` consumes. This is intentional by design: staging tables allow the domain transaction to durably commit PDF/DOCX dispatch intent, while `outbox_events` is the generic relay consumed by the worker binary. The relay workers (`PDFOutboxWorker`, `MaterializeOutboxWorker`) exist solely to bridge between the two levels.

**Dead restart loop:**
`apps/api/cmd/metaldocs-api/main.go:462-486` — `startOutboxWorker` has a retry loop that re-enters `run(ctx)` if it returns a non-nil error. `PDFOutboxWorker.Run` and `MaterializeOutboxWorker.Run` both return `nil` unconditionally on context cancellation (`return nil` at worker line 47 / 46). The retry branch is never taken. The loop is dead code in production.

**Approval idempotency store duplication:**
`internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go` and `postgres_route_admin_idemp_store.go` are ~95% identical (noted in register; not re-confirmed line-by-line as it is out of scope for this theme's primary focus).

### Standard

**DRY principle** (Hunt & Thomas, "The Pragmatic Programmer", principle 8): "Every piece of knowledge must have a single, authoritative, unambiguous representation within a system." The six repository methods replicated verbatim across two files violate this: a bug fix or schema change must be applied twice, and the next developer adding a third staging outbox type will copy again.

**Go generics** (Go 1.18+, spec §Type parameters): Go 1.18 introduced type parameters, which are the idiomatic Go solution for this exact pattern — a generic outbox repository parameterised on the row type and table name. The standard library's `database/sql` package passes typed scan destinations as interface values, making generic repository wrappers straightforward.

The two-stage relay chain is **not a duplication problem** — it is an intentional architectural choice (domain transactions write to staging tables; the relay promotes rows into the generic platform outbox). This design follows the Transactional Outbox pattern correctly and does not need to be changed.

### Verdict: SIMPLIFY — P2

The register's severity (medium) is correct. The duplication is a maintenance burden and a bug-propagation risk, but it is not a correctness or compliance defect. The fix is a bounded simplification: extract a generic `StagingOutboxWorker[R]` and a generic `StagingOutboxRepository[R]` parameterised on the row type, with the table name passed at construction. The two concrete types become thin wrappers or are eliminated entirely.

**The two-stage relay chain is KEEP.** The register asks "collapse to one table?" — the answer is no. The staging tables serve a real purpose: they allow domain code to enqueue within the committing transaction without depending on the platform `outbox_events` table or the `messaging.Publisher` interface. Collapsing to one table would require domain code to import the platform outbox publisher, violating the module boundary. The current design is correct; the duplication is in the implementation, not the architecture.

**The dead restart loop is DELETE (trivial).** `startOutboxWorker` can be replaced by a plain `go func() { workerWG.Add(1); defer workerWG.Done(); run(ctx) }()`. No restart logic is needed because both workers return nil on context cancellation, which is the only expected exit path.

### Smallest correct fix

1. **Generic worker:** Define `type StagingOutboxWorker[R any] struct { ... }` with a `dispatchFn func(R) messaging.Event` passed at construction. Both `PDFOutboxWorker` and `MaterializeOutboxWorker` collapse to `NewStagingOutboxWorker(repo, pub, pdfDispatchFn, log)` calls in `main.go`.
2. **Generic repository:** Define `type StagingOutboxRepository[R any] struct { db *sql.DB; table string; scanRow func(*sql.Rows) (R, error) }`. Both concrete repositories are eliminated. The SQL strings differ only in table name — parameterise at construction.
3. **Dead loop:** Remove the retry loop in `startOutboxWorker`; inline a direct goroutine launch.

### Effort / blast radius

**Effort: M** (generic extraction needs a careful interface refactor and test pass; the change is purely in `render/fanout` and `main.go`).
**Blast radius: module** (render/fanout package + main.go wiring; no cross-module impact).

### ADR needed?

No. The two-stage relay architecture is unchanged. The simplification is an implementation detail within a single package.

### Over-engineering check

Do not introduce a database-agnostic generic outbox framework. The staging tables are Postgres-specific, the SQL is already correct, and the only problem is literal duplication. A Go generic struct with a table-name parameter and a scan function solves this in ~80 lines, replacing ~320 lines of duplicate code. That is the right scope. Do not add a message-broker abstraction layer, do not extract a separate platform package, do not add configuration for retry parameters beyond what already exists.

---

## F-19 — jobs binary absent from Docker Compose / no Dockerfile

### Current state

Confirmed by direct observation:

- `deploy/docker/` contains only `api.Dockerfile` and `worker.Dockerfile`. No `jobs.Dockerfile` exists.
- `deploy/compose/docker-compose.yml` defines services: postgres, redis, minio, minio-init, gotenberg, docx-renderer, api, worker, web, gateway. No `jobs` service exists.
- `apps/jobs/cmd/metaldocs-jobs/main.go` exists — the binary has Go source code.
- `scripts/start-jobs.ps1` exists — local dev only.
- `apps/api/cmd/metaldocs-api/main.go:439` calls `bootstrap.MigrateRiverSchema`. `internal/platform/bootstrap/jobs.go:36` also calls `MigrateRiverSchema`. Both binaries independently run River schema migrations at startup — no single declared owner.
- `internal/modules/jobs/scheduler/lease_reaper.go:38` — the `lease_reaper` subquery `SELECT doc.tenant_id FROM public.documents doc WHERE doc.id::text = d.job_name LIMIT 1` joins `public.documents` matching on `doc.id::text = d.job_name`. For the four maintenance jobs (`stuck-instance-watchdog`, `idempotency-janitor`, `audit-integrity-validator`, `lease-reaper`), `job_name` is a plain string, not a UUID. The subquery always returns NULL. `tenant_id` is always invalid. All four reaped scheduler leases silently fail the governance write (`rowErrs` appended, `reclaimed` stays 0). This is a confirmed code defect.

**Production consequence without jobs binary running:** `scheduled_publish_cutover` River rows accumulate in River's Postgres tables indefinitely. Documents scheduled for future publication never become `published`. This is a silent, user-visible failure — not a performance degradation.

### Standard

**REQ-ASYNC-4** (MetalDocs backend target architecture §6): "Every async pipeline has a watchdog for stuck work and a metric for queue depth + oldest-item age." An async pipeline with no deployed consumer binary has infinite queue depth growth by definition.

**Twelve-Factor App (Factor VI — Processes):** "Execute the app as one or more stateless processes." Every process that participates in the application must be runnable and declared in the deployment manifest. An undeclared process is operationally invisible — it cannot be started, scaled, monitored, or restarted by the orchestrator.

**River job queue documentation** (riverqueue.com/docs): River requires that its worker process (the `river.Client` with registered workers) be running and consuming jobs. River accumulates jobs in the `river_jobs` table with `state='available'` or `state='scheduled'`; they are not processed until a worker claims them. No worker process = perpetual queue growth.

**ISO 9001:2015 §7.1.3** (Infrastructure): "The organization shall determine, provide and maintain the infrastructure necessary for the operation of its processes." A scheduled-publish flow whose executor binary is absent from the deployment manifest does not meet the infrastructure provision requirement for that process.

**Dual River migration ownership:**
The River documentation (riverqueue.com/docs/migrations) explicitly recommends running migrations from a single, controlled location — typically a separate migration step or the primary API binary. Running `MigrateRiverSchema` from two independent binaries that start concurrently creates a TOCTOU risk on the migration state table. Under normal circumstances River's migrations are idempotent (they check the current version), but concurrent startup races on the schema version table can produce errors. This is low-probability but non-zero and leaves schema lifecycle ownership undeclared.

**lease_reaper JOIN bug:**
The query at `lease_reaper.go:38` is a logic error: `doc.id::text = d.job_name` will never match a maintenance job name string. This means governance events for reaped maintenance leases are never written — a silent audit trail gap for the management plane, which is also a QMS concern (ISO 9001 §8.5.1: controlled process documentation).

### Verdict: REFACTOR — P0-prerequisite (jobs deployment) + P1 (dual migration ownership + lease_reaper bug)

**Jobs deployment (P0-prerequisite):** The absence of `jobs.Dockerfile` and the compose service is a deployment gap that makes the scheduled-publish feature non-functional in any containerised environment — including the staging and production environments the company runs. This must be resolved before the async pipeline can be considered production-ready. It is not blocked by any other finding in this register.

**Dual River migration ownership (P1):** Assign `MigrateRiverSchema` to the API binary only (which already calls it). Remove the call from `bootstrap/jobs.go`. The jobs binary should start-fail if River tables are absent, which is the correct fail-fast behavior if the API has not run yet. This is a 2-line change.

**lease_reaper JOIN bug (P1):** The governance write for reaped scheduler leases is a no-op due to the wrong JOIN. Fix: remove the `public.documents` subquery. Maintenance job lease reaps are system-level events with no document-scoped tenant attribution; the governance row should use a system tenant or be written to a system event log without a tenant FK. The simplest fix: omit the tenant attribution for job names that are not UUIDs, and write the governance event to a system audit log or simply log it as a structured slog event. A QMS-grade system should not silently fail governance writes.

### Smallest correct fix

**Jobs deployment:**
1. Create `deploy/docker/jobs.Dockerfile` — copy the pattern from `worker.Dockerfile`, substituting `metaldocs-jobs` binary path.
2. Add a `jobs` service to `docker-compose.yml` with `depends_on: postgres (healthy), api (healthy)` and the environment variables consumed by `jobs/cmd/metaldocs-jobs/main.go` (PGHOST/PORT/DATABASE/USER/PASSWORD, METALDOCS_JOBS_RIVER_SCHEMA, METALDOCS_JOBS_RIVER_MAX_WORKERS).
3. Add a healthcheck (River exposes no HTTP healthcheck by default; use a simple process liveness probe or omit and rely on compose restart policy).

**Dual migration ownership:**
Remove `MigrateRiverSchema` call from `internal/platform/bootstrap/jobs.go:36`. Document in the jobs binary startup that it requires River schema already provisioned by the API binary.

**lease_reaper bug:**
Replace the `public.documents` JOIN with a NULL-safe path: if `job_name` does not parse as a UUID, write a system-level governance event (tenant_id = system sentinel or skip the governance row and emit a structured log line instead). The governance event for scheduler lease reaps is management-plane telemetry, not user-scoped audit; the semantics of requiring a tenant_id are wrong for this event type.

### Effort / blast radius

**Effort: S** (jobs Dockerfile + compose service: ~30 lines; migration ownership: 2-line deletion; lease_reaper: ~10-line change).
**Blast radius: contained** (Dockerfile and compose are deployment artifacts; bootstrap/jobs.go change is isolated; lease_reaper.go change is self-contained).

### ADR needed?

No ADR needed for adding the Dockerfile/compose service — it is restoration of expected deployment completeness. The dual migration ownership fix is a trivial cleanup, not a design decision. The lease_reaper tenant attribution change is a bug fix. None of these rise to the level of an architectural decision requiring an ADR.

### Over-engineering check

Do not introduce a separate River migration binary or a migration-version health gate between the API and jobs startup. The simplest path — API owns River migrations, jobs requires them to exist — is correct and matches how Postgres migrations already work in this system. Do not add a distributed lock around River migrations; the idempotent check built into River's migration client is sufficient for the startup race if ownership is assigned to one binary.

For the lease_reaper: do not introduce a tenant-resolution service or a cross-module lookup. The job names for maintenance jobs are not document IDs; the attempt to derive tenant attribution from them is the root cause. Remove the invalid lookup; emit a system-scoped event or a log line. Keep it simple.

---

## Summary table

| Finding | Verdict | Priority | Effort | Blast radius | ADR? |
|---------|---------|----------|--------|-------------|------|
| F-07 + D-01: post-commit audit/governance gap (taxonomy, templates, documents-core) | REFACTOR | P1 | M | cross-module | No |
| F-07 sub: templates read/write audit sink split | REFACTOR | P1 | S | module | No |
| F-07 sub: deprecated govLogger nil fallback (controlled-documents) | DELETE | P2 | S | module | No |
| F-04: duplicate staging outbox worker/repo clones | SIMPLIFY | P2 | M | module | No |
| F-04: dead restart loop in startOutboxWorker | DELETE | P3 | S | contained | No |
| F-19: jobs binary absent from Docker Compose + no Dockerfile | REFACTOR | P0-prerequisite | S | contained | No |
| F-19: dual River migration ownership | REFACTOR | P1 | S | contained | No |
| F-19: lease_reaper JOIN bug (governance writes always no-op) | REFACTOR | P1 | S | contained | No |

---

## Register severity calibration

| Finding | Register severity | This evaluation | Calibration note |
|---------|------------------|-----------------|-----------------|
| F-07 | high | Confirmed high | ISO 9001 / 21 CFR 11 compliance defect; not over-stated |
| D-01 | cross-area observation | High (same root cause as F-07) | Register correctly identifies that no module has atomic outbox semantics except approval |
| F-04 (duplication) | medium | Confirmed medium — P2 | Maintenance burden, not a correctness defect |
| F-04 (two-stage chain) | "collapse to one table?" | KEEP — not a defect | Stage-2 answer: do not collapse; the staging tables are correct design |
| F-19 (jobs binary) | high | Elevated to P0-prerequisite | Silent feature failure in all containerised deployments; the register's "high" under-states the operational impact |
| F-19 (lease_reaper) | high (tertiary defect) | P1 | Governance no-op on every maintenance lease reap is a QMS gap |

---

## Sources and citations

- microservices.io/patterns/data-management/transactional-outbox.html — Chris Richardson, Transactional Outbox pattern
- ISO 9001:2015 §8.5.1 — Production and service provision (controlled conditions, documented information)
- 21 CFR Part 11 §11.10(e) — Electronic records: time-stamped audit trails
- OWASP ASVS v4.0 V7 (Error Handling and Logging), V10 (Malicious Code) — audit log integrity requirements
- The Twelve-Factor App (12factor.net) Factor VI — Processes: declare and run all processes
- riverqueue.com/docs — River job queue: worker process requirements and migration management
- Go specification §Type parameters (go.dev/ref/spec#Type_parameter_declarations) — generics for DRY repository patterns
- wiki/architecture/backend-target-architecture.md — REQ-ASYNC-1, REQ-ASYNC-2, REQ-ASYNC-3, REQ-ASYNC-4
- wiki/backend/legacy-register.md — F-07, F-04, F-19, D-01
