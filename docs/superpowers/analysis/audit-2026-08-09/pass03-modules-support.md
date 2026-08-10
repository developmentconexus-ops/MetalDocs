# MetalDocs Architecture Audit — Pass 3: Support-Module Architecture Maps

**Date:** 2026-08-09
**Baseline:** `main@418070bf38a9f358f9131bcc36b7a6bcbc069273` (branch `docs/architecture-audit-current-state` @ `9e48a6a1`, worktree HEAD `0365ba17`)
**Status:** reproduced-current-state — every claim below was re-verified directly against source in this worktree at the pinned baseline (not carried forward from the prior inventory pass without re-checking), except where explicitly marked "carried from `docs/superpowers/analysis/inventory/*.md`, unverified this pass."
**Scope:** `internal/modules/{audit,distribution,jobs,notifications,render,search}` — the six modules Pass 1/2 classified as "support" rather than the core Documents/Approval/ControlledDocuments/Taxonomy/IAM cluster.

## 0. Seam-classification legend

Per `docs/superpowers/analysis/2026-08-09-metaldocs-architecture-current-state.md` §4 and the engineering rulebook §5/§9, coupling at a module boundary is classified along independent axes — a single seam can carry more than one letter:

| Code | Dimension | Healthy default |
|---|---|---|
| **C** | Module-level cycle/reciprocal relationship (collapse Go package graph to module identity; A↔B both directions) | none — a module should be a DAG node, not part of a 2-cycle |
| **T** | Foreign type / persistence leakage (domain port typed by `*sql.Tx`/`platform/db.Tx`, or by another module's concrete producer struct/free function) | domain ports use only own-module value types |
| **E** | Error identity coupling (`errors.Is(err, <foreign>domain.Err...)`) | only checks own-module sentinels; foreign failures translated at the boundary |
| **S** | SQL/data coupling (raw `FROM`/`JOIN` against another module's owned base table) | only touches own `TenantDataPort.Tables()` (or shared read-model views) |
| **G** | Go-import coupling (import into another module's published `application`/`domain` surface for a capability call) | consumer depends on the minimal capability it needs, not the producer's full surface |
| **P** | Platform inversion (`internal/platform/**` importing `internal/modules/**`) | platform stays domain-free |
| **W** | Composition-only wiring (dependency lives exclusively in an executable root or `internal/composition/**`) | not itself a defect — composition is expected to know implementations |

Verdict vocabulary used per seam below: **consumer-owned** (target shape — R-MOD-3/4 compliant), **producer-owned** (anti-pattern — consumer depends on producer's internal type/interface), **should-own** (evidence-based recommendation for which side ought to declare the port), **adapter** (a composition-root or infrastructure translation exists), **reciprocal** (a C-class 2-cycle), **sync-vs-event** (whether the seam is an in-process call or crosses the transactional-outbox/River boundary).

---

## 1. `internal/modules/audit`

### Responsibility
Immutable, hash-chained regulatory evidence log (`metaldocs.audit_events`, hash-chain per `prev_hash`/`row_hash`) plus governance events and CSV/JSON export jobs. Not a generic "logging" utility — it is the module that answers "what happened, who did it, and can we prove the record was not altered."

### Owned concepts
`Event` (append-only fact with actor/action/resource/trace/hash-chain fields), `IntegrityIssue`/`IntegrityIssueKind` (hash-chain tamper detection result), `ExportJob`/`ExportStatus`/`ExportFormat` (async CSV/JSON export lifecycle), `ListEventsQuery`/`Cursor` (keyset pagination contract).

### Owned tables (writes)
`internal/modules/audit/infrastructure/postgres/tenant_data_port.go:52-57` — `Tables()` returns:
- `metaldocs.audit_events`
- `metaldocs.audit_export_jobs`
- `public.governance_events`

### Foreign tables read
**None.** Every `FROM`/`SELECT` in `internal/modules/audit/infrastructure/postgres/{writer.go,exports.go,tenant_data_port.go}` targets only `metaldocs.audit_events` and `metaldocs.audit_export_jobs` (its own tables). No S-class violation found for this module — a genuine contrast with the Approval/Documents/ControlledDocuments cluster's LAYERING-08/09/10 foreign-table reads.

### Public application/domain surface
- `domain.Writer` — `Record(ctx, event)`, `RecordTx(ctx, tx db.Tx, event)` (`internal/modules/audit/domain/port.go:128-138`)
- `domain.Reader` — `ListEvents(ctx, query) (items, hasMore, err)`
- `domain.Counter`, `domain.IntegrityValidator`, `domain.ExportJobRepository`
- `application.Service` — `ListEvents`, `ExportEvents`, `GetExportStatus`, `LoadExportPayload`, `BuildSignedURL` (`internal/modules/audit/application/service.go`)

### Inbound/outbound deps (module-edge-evidence.txt)
- **Fan-in = 6** module-level producer edges: `module:auth -> module:audit` (2), `module:iam -> module:audit` (2), `module:jobs -> module:audit` (1, via `audit_integrity_validator`), `module:taxonomy -> module:audit` (2), `module:templates -> module:audit` (1), `module:tokens -> module:audit` (2). Plus non-module producers: `app:api`, `platform:bootstrap` (2).
- **Fan-out = 0** module-level edges. `module:audit -> platform:*` only (apibase, authn, crypto, db, httpresponse, httprouter, pagination, problem, sqlescape, tenant, tenantdata).

This confirms the task's stated fan-in-6/fan-out-0 shape exactly.

### Seam classification
| Seam | Class | Verdict |
|---|---|---|
| `domain.Writer.RecordTx(ctx, tx db.Tx, event)` | **T** | `audit/domain` (not `application`) imports `internal/platform/db` for a `db.Tx` parameter — one of LAYERING-02's 9/15 domain packages. Deliberate (see comment at `port.go:120-127`: audit write must share fate with the caller's mutation), but it is *domain*-owned persistence leakage, not the application-layer transitional case the rulebook's R-LAYER-2 carve-out describes. |
| Every other producer module → `audit.Writer`/`audit.Reader` | **G, consumer-owned** | Producers (auth, iam, jobs, taxonomy, templates, tokens) depend on `audit/domain`'s published interfaces, not audit's infrastructure. This is the correct shape: audit is the **producer** of a stable capability (`RecordTx`) that consumers call; audit never reaches back into any producer's tables. |
| `errors.Is(err, domain.ErrExportJobNotFound)` (`delivery/http/handler.go:326,358`) | **E** | Own-module sentinel only — not a foreign-sentinel violation. |
| Module-level cycle (C) | **none** | Fan-out 0 means audit cannot participate in any of the 7 known module-level reciprocal relationships; it is a pure sink in the module graph. |

### Consumer- vs producer-owned ports
Audit is structurally a **pure producer**: every cross-module edge into audit is a downstream module consuming `audit/domain.Writer`/`Reader`, never audit reaching into another module's tables or domain types. The one architectural cost is that the producer's own port (`RecordTx`) is typed by platform transaction plumbing, so every consumer that wants transactional audit writes must import `internal/platform/db` transitively through `audit/domain` — this is the mechanism, not a design flaw in direction.

### Transaction participation
`internal/modules/audit/infrastructure/postgres/writer.go:93` hand-rolls `w.db.BeginTx(ctx, nil)` directly (not via `platform/db.TxRunner.Do`) for the non-transactional `Record` path — adds to PERSIST-02's 82-site count. `RecordTx` accepts the caller's `db.Tx` for the same-transaction path.

### Events/jobs
No River worker/periodic job lives in `audit` itself. It is *read by* `internal/modules/jobs/audit_integrity_validator/job.go` (hourly periodic job, `audit-integrity-validator`), which calls `auditdomain.IntegrityValidator.ValidateIntegrity` — a clean G-class capability call into audit's own published interface, not a foreign-table read.

### HTTP routes
Mounted via generated strict handler (`auditapi.HandlerWithOptions`, `delivery/http/handler.go:93-97`), tag `audit`:
- `GET /audit/events` (`api/openapi/v1/openapi.yaml:888`)
- `GET /audit/events/export` (`:683`)
- `GET /audit/events/export/{export_id}` (`:710`)
- `GET /audit/events/export/{export_id}/download` (`:736`)

### Test shape
`handler_allow_test.go`, `handler_export_test.go`, `handler_test.go`, `handler_typed_test.go`, `route_registration_test.go`, `service_test.go`, `writer_test.go`. Zero `//go:build integration`-tagged files — all HTTP/service-layer coverage is unit-level (fakes/mocks), no DB-integration suite found for this module.

### Findings mapped to the owning issues
- **#93/A4** — `audit/domain.RecordTx`'s `db.Tx` parameter (T-class) is in-scope for the domain-persistence-leakage burn-down alongside the other 8 domain packages, but it is the one instance in this pass where the leak exists *because* the port is a widely-shared cross-module capability (fan-in 6), not an accident — any A4 remediation must preserve "audit write shares fate with the caller's tx" as a first-class requirement, not just erase the `db.Tx` parameter.
- **#92/A5** — `writer.go:93`'s direct `BeginTx` call adds one more site to the 82/25-file PERSIST-02 count; a straightforward `TxRunner.Do` migration candidate since it does not need caller-tx sharing.
- **§12 of the rulebook (audit-as-shared-domain question)** — see §7 below for the explicit verdict this pass reaches.

---

## 2. `internal/modules/distribution`

### Responsibility
Read-only presentation of who a released controlled document was (or is obligated to be) distributed to — recipient list, coverage/area rollup — surfaced on top of Controlled Documents' obligation data. It does not create, mutate, or own any distribution *event*; "distribution" here means "who must read this," not "we emailed/shipped this."

### Owned concepts
`RecipientRow`, `RecipientsPage`, `AreaCoverageRow` (`internal/modules/distribution/domain/types.go`) — all read-model row shapes, no lifecycle, no mutation verbs, no owned aggregate.

### Owned tables (writes)
**None.** No `TenantDataPort` exists for this module (`find internal/modules -iname tenant_data_port.go` has no `distribution` entry) — distribution owns zero independent rows to export/erase.

### Foreign tables read
`internal/modules/distribution/infrastructure/coverage_repository.go` reads exclusively through DB **views**, never base tables:
- `metaldocs.v_cd_obligated_readers` (`:50,154,175,199,308,311`)
- `metaldocs.v_process_area_name` (`:155,176,200,312`)

Both views are defined in `db/baseline/0001_current_schema.sql` (`v_cd_obligated_readers` at `:1813`, sourced from migration `archive/migrations/post-baseline-2026-06-fold/0245_cd_obligated_readers_view.sql`) — a database-level read-model projection over ControlledDocuments/Taxonomy base tables, not a raw cross-module `JOIN`. This is the R-DATA-2-sanctioned "deliberate read model/projection" shape, in contrast to Approval/Documents' raw `FROM documents`/`FROM approval_instances` reads (LAYERING-08/09).

### Public application/domain surface
No `application` package exists. `delivery/http.Handler` declares its own `Repository` interface (`handler.go:22-27`) and calls it directly from the HTTP layer — there is no use-case/service layer between delivery and infrastructure.

### Inbound/outbound deps
- **Fan-in = 0** module-level edges (only `app:api -> module:distribution`, composition-shaped).
- **Fan-out = 1**: `module:distribution -> module:iam` (`infrastructure/coverage_repository.go` → `iam/domain`, 1 edge) for the tenant-scoped grant/actor shape it needs.

### Seam classification
| Seam | Class | Verdict |
|---|---|---|
| `coverage_repository.go` → `metaldocs.v_cd_obligated_readers`/`v_process_area_name` | **S, but adapter-shaped** | Reads a DB view, not another module's base table directly. No module Go package declares/owns these views today; ownership is implicit in the baseline schema. This is a soft S-class gap (no `TenantDataPort`-style ownership catalog entry for views), not a raw foreign-table read. |
| `-> module:iam` | **G, producer-owned** (mild) | distribution imports `iam/domain` types directly rather than declaring its own minimal `ActorRef`-shaped port; low severity given it is a single read-only edge. |
| Module-level cycle (C) | **none** | Fan-out 1, one-directional (iam does not import distribution back). |

### Consumer- vs producer-owned ports
distribution is a pure **consumer** with no owned port surface at all — `Repository` in `delivery/http` is infrastructure-shaped (raw SQL-view queries), not a use-case-shaped application port. There is nothing here for another module to consume.

### Transaction participation
None — every method in `coverage_repository.go` is a single read (`db.QueryContext`/`QueryRowContext`), no `BeginTx`/`TxRunner` usage found.

### Events/jobs
None. No River worker, no periodic job, no outbox producer/consumer role.

### HTTP routes
Mounted via `distributionapi.HandlerWithOptions` (`delivery/http/routes.go:19-24`), tag `distribution`:
- `GET /documents/{id}/distribution` (`api/openapi/v1/openapi.yaml:4385`)
- `GET /documents/{id}/distribution/recipients` (`:4408`)
- `GET /documents/{id}/distribution/coverage` (`:4440`)

### Test shape
Exactly one test file, `coverage_repository_integration_test.go` — `//go:build integration` tagged, DB-backed. No handler-level unit test found for this module (contrast with audit's 6 unit-test files).

### Findings mapped to owning issues
- **#93/A4** — the module has no `application` layer (R-LAYER-3 nominally expects delivery→application→domain); given the module is genuinely thin (three read-only endpoints over one view pair), this is defensible as intentionally minimal rather than a missed layer, but any future write capability added to `distribution` should not be bolted directly onto `delivery/http.Handler`.
- **#95/A7** — no distinct async/observability concern: distribution has no job/worker surface (out of scope for Pass 2's findings).
- Not owned by any existing #87-#95 program as a *new* root cause — it is evidence supporting A4's general "read model/projection ownership should be explicit" property (R-DATA-2), not a novel defect class.

---

## 3. `internal/modules/jobs`

### Responsibility — and the explicit "is this a real module" question
`internal/modules/jobs` is the container for River's leader-elected **periodic maintenance jobs**: `stuck-instance-watchdog`, `idempotency-janitor`, `audit-integrity-validator`, `document-review-surfacer`, `approval-sla-surfacer`, `release-hold-reconciler`, plus `outbox-retention`. It has no owned business vocabulary of its own beyond `metaldocs.idempotency_keys` (an operational/ephemeral platform-adjacent table, explicitly documented as such: `internal/modules/jobs/tenantdata/tenant_data_port.go:1-6` — "owns metaldocs.idempotency_keys ... because it is not tied to one janitor").

**Verdict: `jobs` is composition-shaped orchestration living under `internal/modules/`, not a bounded context.** Evidence:

1. **No shared domain language.** There is no `internal/modules/jobs/domain` package. Each of the seven job subpackages (`approval_sla_surfacer`, `audit_integrity_validator`, `document_review_surfacer`, `idempotency_janitor`, `outbox_retention`, `release_hold_reconciler`, `stuck_instance_watchdog`) is a self-contained River `Args`+`Work` pair that imports whichever *other* module's domain/application it needs to do its check:
   - `approval_sla_surfacer/job.go:34-35` → `approval/domain`, `iam/authz`
   - `release_hold_reconciler/job.go:35-37` → `approval/application`, `approval/domain`, `iam/authz`
   - `stuck_instance_watchdog/job.go:10-12` → `approval/application`, `iam/authz`
   - `document_review_surfacer/job.go` → `documents/domain`
   - `audit_integrity_validator/job.go:8` → `audit/domain`
   - `idempotency_janitor/job.go`, `outbox_retention/job.go` → no module imports (pure platform-table janitors)
2. **Fan-out mirrors a composition root, not a bounded context's capability list.** `module:jobs -> module:approval` (4 edges), `-> module:audit` (1), `-> module:documents` (1), `-> module:iam` (4) — four *different* producer modules' domain/application surfaces reached into by four *different, unrelated* subpackages. A real bounded context has one coherent business language it reaches out from; `jobs` has seven independent janitors that happen to share a directory and a River queue.
3. **Fan-in = 0.** No other module imports anything from `internal/modules/jobs/*` — consistent with "this is a leaf orchestration layer," not a producer of a capability other contexts consume.
4. **`internal/modules/jobs/maintenance/periodic.go`** is explicitly documented (`:1-11`) as holding "ONLY the schedule/args wiring, no worker construction, no DB (ADR 0067)" — i.e., it is itself a composition manifest (which jobs exist, on what cadence), imported by both `apps/api` (to enqueue as elected leader) and `apps/jobs` (to enqueue + execute).

This makes `jobs` structurally the same *shape* as `internal/composition/tenantdata/registry` (LAYERING-12: "composition-root-shaped... imports 12 of 15 modules' infrastructure packages to assemble the registry") — except `jobs` lives under `internal/modules/` and is treated as a peer bounded context in the "15 modules" count, while `tenantdata/registry` correctly lives under `internal/composition/`. Per rulebook R-PLAT-2 ("composition is allowed to know implementations... keep composition separate from platform so the verifier can distinguish W from real module coupling"), the same separation should apply one level up: `jobs`'s cross-module fan-out is **W-shaped** (composition wiring), not **G-shaped** (a bounded context's legitimate capability dependency), and today nothing in the architecture verifier or the 15-module count distinguishes the two.

### Owned concepts
Only `metaldocs.idempotency_keys` row lifecycle (operational/ephemeral, TTL-swept). No domain entity, no aggregate, no invariant beyond "delete rows past `expires_at` + grace".

### Owned tables (writes)
`internal/modules/jobs/tenantdata/tenant_data_port.go:35-37` — `Tables()` returns `["metaldocs.idempotency_keys"]` only.

### Foreign tables read
- `idempotency_janitor/job.go:61,63,90` — own table only (`metaldocs.idempotency_keys`).
- The other six job subpackages read/write through their **target module's own application/domain interfaces** (e.g., `approvaldomain`/`application` calls), not raw SQL against foreign tables — so this dimension is **G**, not **S**, for `jobs`: it is Go-import coupling into producer capability surfaces, correctly mediated by the producer's own published package, not a SQL-string bypass of module boundaries.

### Public application/domain surface
None published outward (fan-in 0 confirms nothing consumes `jobs`). Internally each subpackage exposes only `<Name>Args` (River job args, `Kind() string`) and a `Work(ctx, job)` method satisfying `river.Worker[T]`.

### Inbound/outbound deps
- **Fan-in = 0** module-level edges.
- **Fan-out = 4** distinct producer modules: `approval` (4 package edges across 3 subpackages), `audit` (1), `documents` (1), `iam` (4, `iam/authz` only — capability checks each janitor performs before acting).
- Also: `module:jobs -> platform:db` (2, `release_hold_reconciler`+`stuck_instance_watchdog` direct `BeginTx`), `-> platform:messaging` (1, `outbox_retention` → `messaging/outbox/postgres`), `-> platform:tenantdata` (1).

### Seam classification
| Seam | Class | Verdict |
|---|---|---|
| `jobs/* -> approval\|audit\|documents\|iam` (fan-out 4) | **G, but W-shaped in aggregate** | Each individual edge is a legitimate capability call into the target module's own published surface (no foreign SQL, no foreign sentinel). The *aggregate shape* — one directory reaching into four unrelated producer domains for four unrelated reasons — is composition, not bounded-context collaboration. See verdict above. |
| Module-level cycle (C) | **none** | Fan-in 0 rules out reciprocal edges. |
| `release_hold_reconciler.go:168,194`, `stuck_instance_watchdog.go:119,167` direct `BeginTx` | **T (mechanism), A5-owned** | Hand-rolled transactions, not `TxRunner`; comments at `release_hold_reconciler.go:203` and `stuck_instance_watchdog.go:176` explicitly note "Carrier-less ctx → the TxRunner chokepoint never runs here" — a *documented*, not accidental, deviation, because these jobs run outside any HTTP request's tenant-GUC-seeded context. |

### Consumer- vs producer-owned ports
Every seam here is `jobs`-as-consumer of another module's producer-owned application/domain interface (e.g., `approval/application`'s existing exported types) — there is no jobs-declared minimal port (`R-MOD-3`/`R-MOD-4` shape) anywhere in this module. Given the verdict above (jobs is composition, not a context), the correct target fix per rulebook R-MOD-4 is not "give jobs consumer-owned ports" but "recognize the fan-out as composition and stop counting `jobs`'s producer imports as module-seam debt the same way Approval-vs-Documents imports are counted."

### Transaction participation
Mixed: `idempotency_janitor` and `outbox_retention` use short-lived direct queries; `release_hold_reconciler` and `stuck_instance_watchdog` hand-roll `BeginTx` explicitly because they run in a River-worker context with no request-scoped tenant GUC seed (the TxRunner chokepoint depends on that carrier). `approval_sla_surfacer`, `document_review_surfacer` similarly `BeginTx` directly (`job.go:132,158` and `:122,149`).

### Events/jobs
This *is* the events/jobs layer for maintenance — see `internal/modules/jobs/maintenance/periodic.go` (7 `river.PeriodicJob` definitions, ADR 0067 dual-define pattern; full detail in Pass 2 §5 below). Zero `slog` calls in `periodic.go` itself (confirmed: `grep -c slog internal/modules/jobs/maintenance/periodic.go` → 0); each job's own package logs its own tick.

### HTTP routes
None. No `delivery/http` package exists anywhere under `internal/modules/jobs/`.

### Test shape
`job_test.go` ×4 (unit), `job_integration_test.go` ×6, `retention_integration_test.go` ×1 — DB-integration-heavy (7 of 11 test files are `//go:build integration`-tagged), consistent with each job's logic being mostly "query Postgres, act on rows."

### Findings mapped to owning issues
- **#93/A4** — primary owner. This pass's specific, previously-uncatalogued finding: **`jobs` should be evaluated for demotion from "bounded context" status alongside the already-ruled `documents`/`controlleddocuments`/`templates` (ADR 0093/#94/A9) question** — not because it needs domain consolidation like Controlled Information, but because R-MOD-1 ("a module is a business ownership boundary, not a folder category") applies just as directly: `jobs` owns one operational table and reaches into four other contexts by necessity, which is the textbook shape of `internal/composition/**`, not `internal/modules/**`. This is a **new observation this pass, not yet represented in the 2026-08-09 current-state/rulebook docs**, but it does not need a new root-cause issue per the current-state doc §14 policy — it is squarely inside #93/A4's existing charter ("module seam migration... platform inversion... composition vs module classification").
- **#92/A5** — the four `BeginTx`-hand-rolling job subpackages add to PERSIST-02's count; two of them (`release_hold_reconciler`, `stuck_instance_watchdog`) document *why* `TxRunner` cannot apply as-is (no request-scoped tenant-GUC carrier in a River worker context), which is itself useful evidence for A5's TxRunner-migration design (a River-worker-safe TxRunner variant, or an explicit "background job" carrier, would remove the exception rather than just re-flagging the same 82 sites).
- **#95/A7** — `internal/modules/jobs/maintenance/periodic.go` having zero scheduler-level logging is OBS-11 (Pass 2 §5); it is jobs-owned evidence feeding that Pass-2 finding, not a separate one.

---

## 4. `internal/modules/notifications`

### Responsibility
In-app notification inbox: fan out a "you have something to review/act on" row per recipient when a document lifecycle or approval event fires, and serve the read/unread inbox API. It is the async **consumer** side of two other modules' domain events, not a notification-composition or templating engine.

### Owned concepts
`NotificationRow`, `NotificationsPage` (`internal/modules/notifications/domain/types.go`) — row/pagination shapes only; no owned lifecycle entity beyond "read/unread."

### Owned tables (writes)
`internal/modules/notifications/infrastructure/tenant_data_port.go:28-30` — `Tables()` returns `["metaldocs.notifications"]` only (documented "operational/ephemeral, no long-term PII value").

### Foreign tables read
**None.** `notifications_repository.go:67,80,142` reads only `metaldocs.notifications`. All recipient/context data needed to build a notification row arrives as **River job args**, not via foreign SQL — see next section for why that is still a coupling finding, just a different class (T, not S).

### Public application/domain surface
No `application` package. `delivery/http.Handler` declares its own `Repository` interface directly (`handler.go:27-34`), same thin CRUD shape as `distribution`.

### Inbound/outbound deps
- **Fan-in = 0** module-level edges (only `app:api`/`app:jobs`/`composition -> module:notifications`).
- **Fan-out = 3**: `module:notifications -> module:approval` (1, `infrastructure -> approval/domain`), `-> module:documents` (1, `infrastructure -> documents/domain`), `-> module:iam` (1, `infrastructure -> iam/authz`).

### Seam classification
| Seam | Class | Verdict |
|---|---|---|
| `NotificationsFanoutWorker.Work(ctx, job *river.Job[documentsdomain.LifecycleEventArgs])` (`infrastructure/fanout_worker.go:47`) | **T, producer-owned, sync-vs-event: event** | notifications' River worker is typed directly by `documents/domain.LifecycleEventArgs` — a producer (documents) concrete type crossing an async boundary into a consumer's infrastructure signature. This is LAYERING-03's "producer-owned concrete type" anti-pattern, applied to a River job-args struct rather than an in-process call; it means documents cannot rename/reshape `LifecycleEventArgs` without recompiling/breaking `notifications`, and there is no consumer-declared minimal args shape insulating the seam. |
| `ApprovalNotifyWorker.Work(ctx, job *river.Job[approvaldomain.ApprovalNotificationArgs])` (`infrastructure/approval_notify_worker.go:57`) | **T, producer-owned, sync-vs-event: event** | Same pattern, second producer (approval). Two independent producers each hand notifications their own args shape directly — no shared "notification trigger" vocabulary owned by notifications itself. |
| `-> iam/authz` | **G, producer-owned** (mild) | Capability check inside the worker; standard cross-module authz dependency, not a data-ownership issue. |
| Module-level cycle (C) | **none** | Fan-in 0; approval/documents/iam do not import notifications back. |

**Contrast with `render`** (§5 below): render's own River job args (`PDFDispatchArgs`, `MaterializeDispatchArgs`) are self-contained value types declared in `render/fanout/dispatchjobs`, importing no foreign domain package. `notifications` took the opposite path — it consumes the *producing* module's own domain args type directly rather than declaring its own `NotificationTriggerArgs`-shaped consumer port. Same architectural question (what should a River job's `Args` type look like at a module seam), two different answers in the same codebase.

### Consumer- vs producer-owned ports
notifications has **zero consumer-owned ports**. Both of its cross-module async triggers are typed by the producer's domain package directly, and its one HTTP-facing `Repository` interface is infrastructure-shaped, not a use-case-shaped application port a producer could target.

### Transaction participation
`approval_notify_worker.go:82` and `fanout_worker.go:64` both hand-roll `w.db.BeginTx(ctx, nil)` directly — two more PERSIST-02 sites. `fanoutToReaders`/`fanoutToAuthor`/`insertRow` (`fanout_worker.go:97,119,127`) all take the already-open `*sql.Tx` as a parameter, so the whole fanout for one lifecycle event is one transaction.

### Events/jobs
Two River workers, both consumer-side of async fanout:
- `ApprovalNotifyWorker` — consumes `approval`'s `ApprovalNotificationArgs`.
- `NotificationsFanoutWorker` — consumes `documents`' `LifecycleEventArgs`, fans out to both the document's readers (`fanoutToReaders`) and its author (`fanoutToAuthor`).

Neither worker creates an OTel span (consistent with Pass 2 OBS-04 — zero spans anywhere under `internal/platform/worker`/`internal/platform/messaging`, and these River-queue workers are the same "async processing creates zero spans" gap, just via River rather than the Postgres outbox).

### HTTP routes
Mounted via `notificationsapi.HandlerWithOptions` (`delivery/http/routes.go:19-24`), tag `notifications`:
- `GET /notifications` (`api/openapi/v1/openapi.yaml:4465`)
- `GET /notifications/unread-count` (`:4495`)
- `POST /notifications/{id}/read` (`:4514`)
- `POST /notifications/read-all` (`:4533`)

### Test shape
`approval_notify_worker_integration_test.go`, `fanout_worker_integration_test.go`, `fanout_worker_race_integration_test.go` (concurrency-specific), `fanout_worker_test.go`, `notifications_repository_integration_test.go` — 4 of 5 files integration-tagged; the one race test is notable (suggests a known concurrent-fanout hazard was tested for, not merely assumed safe).

### Findings mapped to owning issues
- **#93/A4** — the two producer-typed River `Args` seams (T-class) are a specific, file:line-cited instance of the same root cause as LAYERING-03/04, extended to the async/River boundary rather than in-process calls. Worth flagging to A4 explicitly because the existing inventory's LAYERING findings sampled only in-process seams; the async job-args shape is architecturally the same defect with a different blast radius (a producer renaming its `Args` struct breaks a *deployed, independently-versioned* consumer binary at job-decode time, not just at compile time).
- **#92/A5** — 2 more `BeginTx` sites for PERSIST-02.
- **#95/A7** — zero spans in either worker; feeds Pass 2 OBS-04 (no new finding, same root cause, extended evidence).

---

## 5. `internal/modules/render`

### Responsibility
Owns the **staging-dispatch outbox** for the two-stage PDF/DOCX materialization pipeline (materialize → PDF), the placeholder "computed resolver" registry used when a template value must be derived rather than looked up (author name, approval date, revision number, etc.), and the HTTP client that fans a render request out to the separate `docx-renderer` Node service. It is the seam between "a document/template needs a rendered artifact" and the actual renderer process.

### Owned concepts
`OutboxRow`/staging-dispatch lifecycle (`enqueued → dispatched/dead-lettered`), `ComputedResolver`/`ResolveInput`/`ResolvedValue` (placeholder-computation contract), `ReconstructionEntry`/`EngineVersions` (render-provenance record), `PDFDispatchArgs`/`MaterializeDispatchArgs` (River job payloads).

### Owned tables (writes)
`internal/modules/render/fanout/tenant_data_port.go:31-36` — `Tables()` returns:
- `metaldocs.pdf_dispatch_outbox`
- `metaldocs.materialize_dispatch_outbox`

(Both explicitly documented as operational/ephemeral, "no FK dependencies in either direction," `:10-15`.)

### Foreign tables read
`internal/modules/render/fanout/pdf_pipeline_state.go:37` reads `metaldocs.outbox_events` — this is the **platform**-owned transactional-outbox table (`internal/platform/messaging/outbox/postgres/publisher.go` is the sole INSERT site; see Pass 2 §8), not a business table owned by any of the 15 modules. Reading it to check pipeline state is a legitimate platform-infrastructure read, not an S-class cross-module violation.

No reads of `documents`, `approval`, `controlleddocuments`, or `templates` base tables found anywhere under `internal/modules/render/` — render's own tables are self-contained, and everything else it needs (document title, revision number, approver list, effective date) arrives through **its own declared reader ports** (next section), not foreign SQL.

### Public application/domain surface
- `resolvers.ComputedResolver` — `Key()`, `Version()`, `Resolve(ctx, ResolveInput)` (`resolvers/resolver.go:52-56`)
- `resolvers.RegistryReader`, `RevisionReader`, `WorkflowReader`, `DocumentReader` — **render-declared, render-scoped** reader ports (`resolvers/resolver.go:79-101`), each exposing only the narrow method(s) a resolver needs (e.g. `DocumentReader.GetDocumentTitle`), typed by render's own `TenantID`/`RevisionID`/`ControlledDocumentID` distinct string types, not by any foreign module's domain struct.
- `resolvers.Registry` — `Register`, `Get`, `Known` (`registry.go`)
- `fanout.StagingOutboxRepository` — `Enqueue`, `EnqueuePDF`, `MarkDispatched`, `MarkFailed`, `CountDeadLettered`, `ReadState`, `PurgeDispatched` (`fanout/staging_outbox.go`)
- `fanout.Client` — `Fanout(ctx, Request) (Response, error)` (HTTP client to the `docx-renderer` service, `fanout/client.go:57-77`)
- `fanout.ReconstructService` — `Reconstruct(ctx, tenantID, revisionID)` (`fanout/reconstruction.go:47-67`), itself declaring its own `InputsReader`/`ClientPort`/`ReconstructionWriter` ports (`:13-24`)

### Inbound/outbound deps
- **Fan-in = 2** producer modules: `module:documents -> module:render` (5 package edges: `application`, `delivery/http`, `infrastructure` ×2 → `render/fanout`, `render/resolvers`), `module:templates -> module:render` (2, → `render/domain`).
- **Fan-out = 1**: `module:render -> module:iam` (2 package edges: `fanout/dispatchjobs`, `fanout/infrastructure` → `iam/authz`, capability checks only).

### Seam classification
| Seam | Class | Verdict |
|---|---|---|
| `resolvers.RegistryReader`/`RevisionReader`/`WorkflowReader`/`DocumentReader` | **G, consumer-owned — positive exemplar** | render is the *consumer* of documents/controlleddocuments/approval data here (it needs to read revision/author/approver/document-title facts to compute a placeholder value), and it declares its own minimal, render-scoped port shapes rather than importing `documentsdomain`/`approvaldomain` types. This is the same discipline as the already-known `DictionaryValueReader` exemplar (LAYERING-05) — a **second, independently-arrived-at instance of the correct pattern**, not previously called out in the prior inventory pass. |
| `PDFDispatchArgs`/`MaterializeDispatchArgs` (River job args, `dispatchjobs/args.go:26-51`) | **G, consumer/producer-neutral — positive exemplar** | Fully self-contained value types (`dispatchFields` embeds only `TenantID`/`RevisionID`/`OutboxID`/`ReleaseGenerationID` strings/bytes) — no foreign domain struct imported into a River `Args` type, in direct contrast to `notifications`' `LifecycleEventArgs`/`ApprovalNotificationArgs` seams (§4 above). |
| `documents -> render/fanout`, `documents -> render/resolvers` (render as producer here) | **G, should-own: consumer (documents) already owns it correctly** | Since render declares the ports (previous row), documents' dependency on `render/fanout`/`render/resolvers` is a normal producer-capability call, not a violation. |
| `-> iam/authz` | **G, producer-owned** (mild) | Standard capability-check dependency; same shape as every other module's authz calls. |
| Module-level cycle (C) | **none** | Fan-in is one-directional from documents/templates; render does not import either back (confirmed no `documents`/`templates` import anywhere under `internal/modules/render/`). |

### Consumer- vs producer-owned ports
render is unusual among the six modules in this pass for having **both** roles cleanly separated: as a *consumer* of documents/approval/controlleddocuments data (via `resolvers.*Reader`), it owns its ports correctly; as a *producer* of the dispatch/resolver capability that documents/templates consume, its published surface (`StagingOutboxRepository`, `ComputedResolver`, `Registry`) is appropriately narrow and does not leak `render`-internal row shapes into its own method signatures.

### Transaction participation
`fanout/infrastructure/seeded_tx.go:29` hand-rolls `db.BeginTx(ctx, nil)` (PERSIST-02 site). `dispatchjobs/enqueuer.go` methods (`EnqueuePDFTx`, `EnqueueMaterializeTx`) accept a `db.Tx` parameter from the caller (documents' own transaction) rather than opening their own — i.e., render's *producer*-facing enqueue API is transaction-participant-shaped (correct for "insert dispatch row in the same tx as the document mutation that triggers it"), while its *internal* staging-outbox purge worker opens its own short-lived tx.

### Events/jobs
- `dispatchjobs.PDFDispatchWorker`/`MaterializeDispatchWorker` — River workers consuming render's own `PDFDispatchArgs`/`MaterializeDispatchArgs`, publishing a `messaging.Event` (`docgen_v2_pdf`/`docx_materialize`) onto the **platform outbox** for the separate `metaldocs-worker` process to pick up and call `docx-renderer` over HTTP (two-stage: River enqueues → outbox event → worker HTTP-fans-out).
- `retention.PeriodicJob()` (`fanout/retention/periodic.go:22-31`) — daily River periodic job purging dispatched staging-outbox rows; explicitly modeled on `jobs/maintenance/periodic.go`'s ADR 0067 dual-define convention (comment `:6-11`), including the same "no worker construction, no DB" split.

### HTTP routes
None directly (render has no `delivery/http` package) — it is consumed in-process by `documents`' HTTP handlers, and its `fanout.Client` is an *outbound* HTTP client to `docx-renderer`, not an inbound route.

### Test shape
23 test files, the largest of the six modules in this pass: unit coverage per resolver (`approval_date_test.go`, `approvers_test.go`, `author_test.go`, `controlled_by_area_test.go`, `doc_code_test.go`, `doc_title_test.go`, `effective_date_test.go`, `revision_number_test.go`), plus `catalog_parity_test.go`/`contract_test.go` (contract-drift guards), plus 3 integration-tagged files (`dispatch_integration_test.go`, `pdf_pipeline_state_integration_test.go`, `retention_integration_test.go`).

### Findings mapped to owning issues
- **#93/A4** — render's `resolvers.*Reader` ports and self-contained job-args types are worth citing as a **second local exemplar** alongside `DictionaryValueReader` when A4 defines the target pattern for other seams to converge on; they were not previously catalogued in the layering inventory.
- **#92/A5** — one `BeginTx` site (`seeded_tx.go:29`) for PERSIST-02.
- **#95/A7** — render's two River workers plus the fanout HTTP client are additional zero-span sites (OBS-04's scope already covers "outbox/worker async processing," and River-dispatched jobs are architecturally the same gap).

---

## 6. `internal/modules/search`

### Responsibility
Cross-entity keyword/filter search over documents and controlled documents for a tenant's visible corpus — one read endpoint, one query shape, backed by pre-computed search-fact views.

### Owned concepts
`Query`, `Document` (search result row) — `internal/modules/search/domain/model.go`. No owned lifecycle, no mutation.

### Owned tables (writes)
**None.** No `TenantDataPort` for `search` — it owns zero independent rows.

### Foreign tables read
`internal/modules/search/infrastructure/v2documents/reader.go:75-109` reads exclusively through views:
- `metaldocs.v_document_search_facts` (`:75`)
- `metaldocs.v_cd_search_facts` (`:76`)
- `metaldocs.v_cd_grantee` (`:109`, visibility-scoping subquery)

All three defined in `db/baseline/0001_current_schema.sql` (`v_document_search_facts:1964`, `v_cd_search_facts:1790`, `v_cd_grantee:1762`), sourced from `archive/migrations/post-baseline-2026-06-fold/0243_cd_search_visibility_contract.sql` and `0244_documents_search_projection.sql`. Same read-model-projection shape as `distribution` (§2) — a genuine, repeatable pattern for this pass's two pure-read modules.

### Public application/domain surface
- `domain.Reader` — `internal/modules/search/domain/port.go:7` (single method surface backing `application.Service`)
- `application.Service` — `SearchDocuments(ctx, domain.Query) ([]domain.Document, error)` (`application/service.go:32`)

### Inbound/outbound deps
- **Fan-in = 0** module-level edges (only `app:api -> module:search`, 3 package edges).
- **Fan-out = 2**: `module:search -> module:iam` (1, `delivery/http -> iam/domain`), `module:search -> module:taxonomy` (1, `infrastructure/v2documents -> taxonomy/domain`).

### Seam classification
| Seam | Class | Verdict |
|---|---|---|
| `v2documents/reader.go` → `v_document_search_facts`/`v_cd_search_facts`/`v_cd_grantee` | **S, adapter-shaped** | Same as distribution: a DB view, not a raw base-table `JOIN`. No explicit ownership catalog entry names these views today (soft gap, not a violation). |
| `-> iam/domain`, `-> taxonomy/domain` | **G, producer-owned** (mild) | Two narrow, read-only capability dependencies (tenant/actor shape from iam, area/family shape from taxonomy) — no application-layer type leakage found in the sampled surface (`domain.Reader` itself takes no foreign type). |
| `errors.Is(err, domain.ErrQueryTenantEmpty)` (`application/service.go:45`) | **E** | Own-module sentinel only. |
| Module-level cycle (C) | **none** | Fan-in 0. |

### Consumer- vs producer-owned ports
search is a pure **consumer**: its own `domain.Reader` is the only port in the module, and nothing else in the codebase implements or consumes it — it exists solely to let `application.Service` be tested against a fake. There is no producer role for search to play (fan-in 0 confirms nothing depends on it).

### Transaction participation
None — `v2documents/reader.go` issues read-only queries; no `BeginTx`/`TxRunner` usage anywhere in the module.

### Events/jobs
None. No worker, no periodic job, no outbox role.

### HTTP routes
Mounted via `searchapi.HandlerWithOptions` (`delivery/http/handler.go:84-96`), tag `search`:
- `GET /search/documents` (`api/openapi/v1/openapi.yaml:950`)

### Test shape
8 test files: `handler_test.go`, `handler_typed_response_test.go`, `reader_test.go`, `reader_like_escape_test.go` (SQL-injection/escaping-specific unit test — worth noting as defensive coverage), `service_test.go` (unit), plus 3 integration-tagged (`reader_contract_parity_integration_test.go`, `reader_family_integration_test.go`, `reader_visibility_integration_test.go` — the visibility test in particular guards the `v_cd_grantee` tenant/grant-scoping subquery).

### Findings mapped to owning issues
- **#93/A4** — same soft S-class gap as distribution (view ownership not catalogued); folds into A4's "single machine-readable table-ownership catalog" acceptance property (rulebook V-ARCH-4) if/when that catalog is extended to cover views, not just base tables.
- No other findings — search is the cleanest of the six modules in this pass on every dimension except the shared soft-view-ownership gap.

---

## 7. Cross-module synthesis for this pass

### 7.1 Is `audit` a legitimate shared domain module despite fan-in 6/fan-out 0? (rulebook §12)

**Yes — audit is correctly a top-level bounded context, not cross-cutting technical infrastructure that should be pushed down into `internal/platform`.** The rulebook's §12 test is explicit: "Audit/evidence that is itself a business/regulatory concept remains a domain module even if many modules use it." Evidence this pass gathered supports the concept side of that test directly:

- The owned aggregate (`Event` with `prev_hash`/`row_hash` chain fields, `audit_sequence`) is a **regulatory evidence artifact** — its correctness properties (tamper-evidence via hash chain, `IntegrityValidator`) are business rules, not generic logging mechanics. Generic logging (what `internal/platform` would own) has no hash-chain-integrity invariant to validate; audit's does, and `jobs/audit_integrity_validator` exists specifically to enforce it hourly.
- High fan-in (6) is the *expected* shape for a shared regulatory-evidence context, exactly analogous to why `iam/domain` legitimately has fan-in 32 (LAYERING-11) — both are "everyone needs to call this one thing," not "this got put in the wrong place." The distinguishing test is not the fan-in number but whether the high-fan-in package has its *own* domain language and invariants (audit: yes — `Event`, `IntegrityIssue`, hash chain) versus being a bag of technical utility functions (which would belong in `internal/platform`).
- Fan-out 0 is additional supporting evidence, not just neutral: audit never needs another module's domain type to do its job (`Record`/`RecordTx` take audit's own `Event` value type), which is what you would expect from a module whose entire responsibility is "receive a fact, store it immutably" — a genuinely leaf-shaped bounded context, unlike `jobs` (§3), whose fan-out-4-into-unrelated-domains shape is the actual tell for "this is not a bounded context."

The one architectural cost identified this pass — `RecordTx`'s `db.Tx` parameter living in `audit/domain` rather than `audit/application` (T-class, §1) — is a layering deviation *within* a legitimately-scoped module, not evidence against the module's existence. It should be tracked as ordinary #93/A4 domain-persistence-leakage debt, not as a reason to reconsider audit's status as a module.

### 7.2 Is `jobs` a legitimate module? (this pass's primary new finding)

**No — see the full analysis in §3.** `jobs` is composition-shaped orchestration of seven independent maintenance routines that each reach into a different producer module's domain/application surface, with no owned business language beyond one operational table. It should be evaluated by #93/A4 for reclassification (not necessarily relocation — the periodic-job hosting responsibility is real and needed somewhere) using the same lens ADR 0093/#94/A9 already applied to `documents`/`controlleddocuments`/`templates`: current directory placement under `internal/modules/` is implementation fact, not proof of bounded-context status (rulebook R-MOD-1).

### 7.3 Pattern found twice independently: consumer-declared read ports

Both `render/resolvers` (§5) and the previously-known `documents/application.DictionaryValueReader` (LAYERING-05) show the same correct shape: a module that needs narrow read access to another module's data declares its own minimal interface, typed by its own value types, rather than importing the producer's domain struct. This pass found it **arrived at independently in a different module** (render was not built by the same author/PR as the tokens-adapter exemplar, per the differing comment style and unrelated ADR references) — which is useful evidence for #93/A4: the pattern is not a one-off, it is discoverable/replicable engineering discipline that the team has now demonstrated twice without being told "copy `dictionary_reader_adapter.go`." Worth citing both exemplars together when A4 writes the target-pattern guidance.

### 7.4 Pattern found twice independently, in the *bad* direction: producer-typed async job args

`notifications`' two River workers (`documentsdomain.LifecycleEventArgs`, `approvaldomain.ApprovalNotificationArgs`) both take a foreign producer's domain struct directly as job-args, while `render`'s own dispatch workers (`PDFDispatchArgs`/`MaterializeDispatchArgs`) are self-contained. Same underlying question — "what type should a River `Args` struct be at a module seam" — answered two different ways in two different modules of this pass. This extends LAYERING-03/04's finding (previously sampled only in-process/HTTP seams) into the async/River boundary explicitly, which the prior inventory pass did not sample.

### 7.5 What this pass adds to the existing hotspot table (§14 of the rulebook)

No new root-cause issue is justified (per current-state doc §14's creation policy — none of the below has a distinct root cause outside #87-#95's existing charters). Additions to existing owners:

| New evidence this pass | Owner | Relation to existing hotspot row |
|---|---|---|
| `jobs` composition-vs-module classification (§3, §7.2) | #93/A4 | Extends "documents/templates/controlleddocuments peer-context split" row's method (R-MOD-1 test) to a module not covered by ADR 0093 |
| `audit/domain.RecordTx`'s `db.Tx` parameter | #93/A4 + #92/A5 | Extends "9/15 domain packages leak SQL/platform DB types" row with a 10th citable file:line, and the specific "shared regulatory port" nuance |
| `notifications`' two producer-typed River `Args` | #93/A4 | Extends "producer-owned types/interfaces at seams" row to the async/River boundary, not just in-process/HTTP |
| `render/resolvers` + self-contained job args | #93/A4 (positive) | Second exemplar for the target pattern, alongside `dictionary_reader_adapter.go` |
| `distribution`/`search` view-based reads, no `TenantDataPort` ownership entry for the views | #93/A4 | Soft addition to "17+ cross-context table reads" row's *mitigation* side — shows the read-model-projection escape hatch (R-DATA-2) is already in live use for 2 of 6 modules, not just a theoretical option |
| 6 new `BeginTx` file:line sites (audit ×1, jobs ×4, notifications ×2, render ×1 = 8 additional sites) | #92/A5 | Extends PERSIST-02's 82-site/25-file count with modules the prior persistence-lane pass did not enumerate individually |
