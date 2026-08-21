---
id: runtime-process-deployment
kind: authority
owner: architecture
summary: T8-G candidate authority for runtime shells, process/deployment topology, trust boundaries, configuration/secrets, observability, content-processing mechanisms and recovery realization.
---

# R10 T8-G — Runtime / Process / Deployment

> **Status:** OPEN / ACTIVE CANDIDATE / NOT RATIFIED
> **Opened:** 2026-08-21
> **Candidate base:** `main @ 8f39184a2b2e2d07a48ff6796dc9efa77c5c3aac`
> **Implementation:** BLOCKED

This document is the T8-G candidate authority for the smallest runtime, process and deployment realization capable of serving the operator-ratified T8-A→T8-F architecture.

It owns runtime shells, long-lived process placement, deployment artifacts, startup/readiness/shutdown, configuration and secrets, trust/network boundaries, runtime observability, renderer/scanner placement, exact-byte buffering realization, operational controls and the runtime side of backup/restore readiness.

It does **not** reopen Product meaning, semantic ownership, lifecycle, Authorization, persistence semantics, the 78-operation application wire, frontend route meaning or transition/cutover sequencing. T8-H, T9, T10, T11, T12 and Product implementation remain outside this gate.

---

## 1. Global Maximum

```text
ONE MODULAR-MONOLITH APPLICATION RUNTIME
+
ONE POSTGRESQL PRODUCT-STATE DATABASE
+
RIVER WORKERS IN THE APPLICATION PROCESS
+
ONE ACTIVE MANAGED-CONTENT STORE PER DEPLOYMENT
+
ONE PRIVATE MALWARE-INSPECTION MECHANISM
+
ONE PRIVATE CONDITIONAL DOCX→PDF RENDERER
+
ONE EXTERNAL OIDC PROVIDER BOUNDARY
+
VERIFIED EPHEMERAL EXACT-BYTE SPOOL
+
FAIL-CLOSED RECOVERY PROFILE
+
OPENTELEMETRY / OTLP OBSERVABILITY BASELINE
+
ONE-SHOT MIGRATION / JOB / RECOVERY OPERATIONS
+
PROVEN THIRD-PARTY MECHANISMS BEFORE LOCAL INFRASTRUCTURE
-
SEMANTIC AUTHORITY IN RUNTIME MECHANISMS
-
SPECULATIVE DISTRIBUTION
-
DORMANT PLATFORM CAPABILITY
```

Binding interpretation:

> Isolate where a current security, correctness, resource or failure-domain property requires isolation. Keep semantic application behavior inside the accepted modular monolith. Prefer a maintained standard-library or proven third-party mechanism for generic infrastructure. Add a thin MetalDocs adapter only where an accepted boundary needs protection. Implement custom infrastructure only when no smaller proven mechanism satisfies a named current consumer.

---

## 2. Consumer census

T8-G is derived from these current consumers only:

```text
one React SPA delivery surface
same-origin /api/v1 application access
OIDC redirect/callback outside the application OpenAPI
HttpOnly ApplicationSession cookie + synchronizer CSRF
exact-byte application-origin DOCX/PDF reads
browser direct-upload capability with bounded provider CORS
PostgreSQL canonical Product state
River durable work for activated named effects
conditional DOCX→OfficialRendition(PDF) transformation
production malware inspection before untrusted bytes cross a governed boundary
periodic managed-content GC reconciliation
canonical PostgreSQL Search baseline
backup/restore exact-content and security/privacy readiness
```

Current non-consumers remain:

```text
SSR
BFF
service worker/offline correctness
WebSocket/realtime channel
EditorSession service
external Search engine
materialized Search job
notification runtime
generic event bus/outbox
Redis
custom scheduler
service mesh
multi-region active-active
automatic failover
```

Application operation census remains exactly 78. Operation 79 is absent.

---

## 3. Long-lived topology

Baseline steady state:

```text
                         Browser
                            │
                         HTTPS
                            │
                    ┌───────▼────────┐
                    │  MetalDocs app │
                    │                │
                    │ SPA + /api/v1  │
                    │ OIDC callback  │
                    │ exact-byte I/O │
                    │ River workers  │
                    │ GC scheduling  │
                    │ health + OTel  │
                    └───────┬────────┘
                            │
             ┌──────────────┼──────────────┬──────────────┬──────────────┐
             ▼              ▼              ▼              ▼              ▼
        PostgreSQL     ManagedContent     OIDC         Renderer       MalwareInspector
         + River           Store        Provider       PRIVATE          PRIVATE
                              ▲
                              │ bounded signed PUT capability
                           Browser
```

Steady-state Launch process/deployment census:

```text
MetalDocs application deployment     1
MetalDocs app replicas                1 baseline
separate River worker deployment      0
custom scheduler deployment           0
private renderer deployment           1 when DOCX→PDF transformation is supported
private malware-inspection mechanism  1 production profile
Redis                                 0
external Search service               0
```

`replicas=1` is an economic/operational baseline, **not correctness authority**. Temporary old/new overlap during replacement must not corrupt semantics. Additional replicas require measured availability/throughput evidence, not preference.

---

## 4. Runtime shells and artifacts

T8-B's `cmd/<runtime-shells>` realization is:

```text
metaldocs serve
metaldocs migrate
metaldocs jobs ...
metaldocs recovery ...
```

These are one executable with explicit modes/subcommands unless implementation evidence proves separate binaries smaller.

### `serve`

Long-lived application composition:

```text
HTTP + embedded SPA
OIDC integration
River workers
GC periodic reconciliation
runtime observability
```

### `migrate`

One-shot deployment operation with schema/DDL identity distinct from the ordinary runtime database identity.

### `jobs`

Private operator inspection/redrive/diagnostics for required durable work. It is not an application API.

### `recovery`

Private non-serving backup/restore verification and reconciliation operations. It is not an application API.

Application build artifact properties:

```text
immutable/reproducible release artifact
compiled Go executable
compiled SPA assets
runtime CA/trust material when needed
no compiler/toolchain requirement at runtime
no source-tree requirement at runtime
no embedded secret
```

An OCI image is the reference deployment artifact. Container vendor/orchestrator/base-image choice is not durable architecture authority.

---

## 5. Public and private trust boundaries

The MetalDocs application is the only public HTTP surface owned by MetalDocs.

Public/runtime surfaces:

```text
/                     SPA/static/history fallback
/api/v1/*             exact T8-E application wire
/auth/login            OIDC browser integration
/auth/callback         OIDC browser integration
/livez                 runtime liveness
/readyz                runtime readiness
```

Not public:

```text
PostgreSQL
River internal state
renderer
MalwareInspector
job redrive/inspection
migration
recovery operations
private diagnostics
```

Health/auth integration paths do not add application operations to the T8-E census.

Browser is always an untrusted caller. It never receives DB credentials, durable storage credentials, renderer/scanner credentials, session-signing/server secret material or a client-side Authorization matrix.

---

## 6. Managed-content network boundary

Browser direct upload is bounded capability use:

```text
Browser
→ short-lived create-only signed PUT capability
→ active ManagedContentStore
```

Provider CORS is limited to the canonical MetalDocs origin, the required upload method and the exact browser-settable headers returned by the allocation contract.

Governed reads remain:

```text
Browser
→ authenticated MetalDocs application operation
→ authorized OpenExact
→ MetalDocs exact-byte proof
→ application-origin response
```

No baseline provider redirect/presigned GET exposes governed reads. Provider/storage identity never becomes Product identity.

One active `ManagedContentStore` exists per deployment. Multiple simultaneously active stores are a T4/T8-G reopen trigger.

---

## 7. PostgreSQL / River boundary

Launch uses one PostgreSQL product-state database. River remains a third-party PostgreSQL-backed durable-job mechanism and keeps its third-party schema semantics separate from MetalDocs Product state.

Runtime database identity:

```text
ordinary Product DML required by accepted owners/application
River runtime needs
no schema ownership
no migration/DDL authority
```

Migration identity:

```text
bounded one-shot use
schema/DDL authority only as required
not present in ordinary `serve` execution context
```

River workers run in the MetalDocs application process at Launch. A worker split is activated only if measured resource isolation, throughput or availability demonstrates that in-process workers materially compromise accepted application serving.

River backlog is durable work mechanism, never business state.

---

## 8. OIDC boundary

Browser redirects to the external OIDC Provider and MetalDocs performs the required callback/exchange/validation through a provider-neutral identity adapter.

OIDC proves external authentication only. It does not establish:

```text
MetalDocs User eligibility
Permission
scope/relationship eligibility
governance participation
command executability
```

Provider roles/groups never become MetalDocs Authorization authority by mapping convenience.

Protocol implementation should use a mature OIDC/OAuth2 implementation; MetalDocs owns only the anti-corruption mapping from verified issuer+subject semantics into the accepted ProviderSubjectBinding/User/ApplicationSession model.

---

## 9. Renderer boundary

Renderer exists only for the activated transformation path:

```text
exact eligible DOCX Submission
→ private renderer
→ candidate PDF bytes
→ T4 admission/verification
→ reload canonical eligibility
→ OfficialRendition semantic finalization
```

Already-PDF + required PDF reuses the exact admitted bytes and invokes no renderer, provider copy or durable rendition job.

Renderer laws:

```text
private only
no PostgreSQL credential
no ManagedContentStore credential
no OIDC/session credential
no Product lifecycle knowledge
no semantic finalization
outbound network denied by baseline
bounded CPU / memory / PIDs / ephemeral disk / execution time
```

Remote/linked-content retrieval during conversion is not baseline. A document that cannot render acceptably without untrusted outbound dependency fails visibly; MetalDocs does not silently grant renderer egress.

### Backpressure

River is the durable backlog mechanism. Renderer-local queueing is bounded mechanical protection only and never a second durable queue.

For a single-instance LibreOffice-class renderer, effective conversion concurrency per renderer replica is one unless the chosen mechanism proves safe parallel capacity. Additional renderer replicas require measured backlog/throughput evidence.

### Fidelity gate

A renderer/provider/version/font/configuration profile is production-eligible only after a representative MetalDocs DOCX corpus proves material fidelity for the real content classes it must support, including typography/fonts, headers/footers, page numbering, tables, images, lists, section/page breaks and orientation where present.

A material renderer, LibreOffice, font-set or conversion-configuration upgrade reruns the corpus before production promotion.

Byte-identical PDF output across renderer versions is not required. The selected admitted OfficialRendition bytes are immutable once semantic finalization succeeds.

Gotenberg + LibreOffice is the reference provider profile, preferably using the smallest LibreOffice-only image/profile that satisfies the proof. The reference is conditional on the fidelity/resource/security gate; Product semantics do not depend on Gotenberg identity.

---

## 10. MalwareInspector boundary

Production invariant inherited from T4:

> Untrusted external bytes cannot become immutable governed MetalDocs content without successful malware inspection of those exact immutable bytes.

Baseline classes remain:

```text
browser/external/imported bytes → UNTRUSTED_EXTERNAL → CLEAN required
authorized managed copy         → TRUSTED_MANAGED_COPY
renderer output                 → TRUSTED_INTERNAL_DERIVATION
```

Fast DRAFT replacement may reach READY without scanning every debounce. The exact candidate bytes must satisfy the malware gate before the governed boundary that requires it.

MalwareInspector laws:

```text
private only
exact bytes in → CLEAN | MALICIOUS | FAILURE/UNAVAILABLE out
no PostgreSQL credential
no ManagedContentStore credential
no Product lifecycle authority
bounded CPU / memory / request length / duration / queue/in-flight work
no production bypass flag such as FORCE_CLEAN or scan-disabled governed admission
```

ClamAV `clamd` is the reference production mechanism. Prefer same deployment locality / Unix-socket or otherwise protected transport because native clamd TCP is not an authenticated/encrypted public protocol.

The content-scanning engine receives no arbitrary document-processing egress entitlement. Signature update is a separate bounded operational concern: the selected updater may have narrowly scoped egress to approved signature sources, or signatures may be delivered by the deployment platform. Signature-update network access never grants scanned document content general outbound access.

Signature database freshness is an explicit operations/security policy. Inability to satisfy that policy makes required governed admission unavailable rather than implicitly CLEAN.

False-positive recovery is provider/signature correction followed by rescan of the exact bytes, never semantic bypass.

---

## 11. Exact-byte verified spool

T8-E leaves exact-byte buffering strategy to T8-G. Launch baseline is a **verified ephemeral spool**:

```text
OpenExact from ManagedContentStore
→ stream into private ephemeral spool while proving size + SHA-256 + format coherence
→ only after complete proof, commit successful exact-byte response headers
→ stream the verified spool to the browser
```

Properties:

```text
whole governed file in Go heap     forbidden baseline
memory usage                        bounded transfer/hash buffers
spool disk usage                    O(content size)
spool durability                    none
spool Product authority             none
```

Temporary spool/workspace may be discarded wholesale after process/container restart. Anything requiring crash survival belongs in ManagedContentStore, not spool storage.

Renderer/scanner I/O follows the same bounded-spool principle where required. The renderer and scanner do not receive object-store credentials merely to save a copy step.

If accepted content cannot be handled sustainably by this spool profile, the buffering mechanism is falsified; the exact-content guarantee is not weakened.

---

## 12. Configuration and secrets

Runtime configuration is external, typed, startup-validated and immutable for the process lifetime.

Non-secret deployment configuration may include:

```text
canonical public origin
listen address/port
OIDC issuer/client id
PostgreSQL endpoint/database reference
active object-store bucket/location
renderer endpoint/profile
MalwareInspector endpoint/profile
operation-class technical timeouts
log/telemetry configuration
deployment/build metadata
```

Runtime configuration does not define Product Permissions, lifecycle, governance or accepted operations.

Secret laws:

```text
never committed to Git
never baked into the immutable image
never logged or emitted in metrics/health
never returned to browser
least privilege by process/mode
```

Prefer provider/workload identity over exported long-lived credentials. When explicit secrets are necessary, prefer the deployment platform's managed secret capability. Custom secret synchronization/rotation daemons are not Launch baseline.

Secret version change may use controlled process restart unless provider-native identity/refresh already provides safer transparent rotation. No custom hot-reload framework is required.

Config parsing should use a small maintained typed configuration mechanism rather than a broad generic configuration framework. `sethvargo/go-envconfig` is the current reference Go mechanism; cross-field MetalDocs invariants remain explicit local validation.

---

## 13. Startup and readiness

`metaldocs serve` startup order:

```text
load deployment config
→ resolve required secret capabilities
→ validate config/cross-field invariants
→ initialize observability
→ establish PostgreSQL pool
→ prove DB reachable
→ prove schema compatible with executable
→ construct composition graph
→ start required in-process River/GC machinery
→ prove recovery/restore serving barrier clear
→ start HTTP
→ READY
```

Renderer, ManagedContentStore, MalwareInspector and OIDC availability are not global startup probes. They are dependency-scoped mechanisms and are exercised on paths that need them.

### `/livez`

Process liveness only. It does not query PostgreSQL, storage, OIDC, renderer, scanner or River backlog.

### `/readyz`

Ordinary serving readiness requires:

```text
startup complete
not draining
PostgreSQL reachable
schema compatible
required in-process components running
recovery serving barrier cleared
```

PostgreSQL outage makes the process unready but does not require liveness failure/restart. Renderer/scanner/storage/OIDC outage does not globally make the app unready; affected operations fail/degrade explicitly.

Health responses are tiny fixed mechanism responses and expose no topology, credentials, schema version or detailed dependency diagnostics.

Unexpected terminal exit of a required in-process subsystem is process-fatal; the root process performs bounded shutdown and exits non-zero. No nested supervisor framework is required.

---

## 14. Dependency degradation

```text
PostgreSQL unavailable
  → readiness false; Product core cannot serve

ManagedContentStore unavailable
  → byte/upload/admission/GC provider operations fail
  → unrelated Product truth is not rewritten

Renderer unavailable
  → required transformation retries/fails visibly
  → Submission remains truthful; no SourceOnly downgrade

MalwareInspector unavailable or signature policy unsatisfied
  → governed admissions requiring CLEAN are blocked/retriable
  → ordinary reads/other Product operations may continue

OIDC unavailable
  → new login/provider-dependent operations may fail
  → already-valid MetalDocs ApplicationSessions do not disappear merely because IdP network is unavailable
```

No continuous health-polling mesh is created for partial dependencies merely to report green/red state.

---

## 15. Graceful shutdown

Termination flow:

```text
READY=false
→ stop accepting new ordinary work
→ stop new GC scheduling
→ stop River from fetching new jobs
→ bounded HTTP drain
→ bounded completion of in-flight jobs
→ cancel remaining work at deadline
→ flush/shutdown telemetry within remaining budget
→ close DB/provider clients
→ exit
```

Use Go's standard HTTP graceful-shutdown behavior rather than a custom HTTP drain framework.

All external/network operations have bounded operation-class-specific budgets. T8-G does not impose one global request timeout across tiny JSON requests, large exact-byte responses, OIDC calls and renderer jobs.

Forced termination never invents Product `INTERRUPTED` states. PostgreSQL rollback, T8-E idempotency and T5 at-least-once/revalidation provide retry safety where already ratified.

---

## 16. Deployment profile

T8-G freezes required deployment properties, not a cloud vendor.

A valid deployment substrate must provide:

```text
public HTTPS/TLS ingress for the MetalDocs app
private network reachability for required backing mechanisms
immutable application artifact execution
one-shot process execution
health probes
resource limits
bounded graceful termination
stdout/stderr log capture
managed secret/workload identity capability
PostgreSQL backing service
managed-content backing service
backup/restore capability
```

Not durable requirements:

```text
Kubernetes
Helm
ECS
Cloud Run
Fly.io
Nomad
Docker Compose
systemd
service mesh
custom ingress controller
Vault cluster
Redis
```

Build/release/run remain distinct: a source build creates an immutable artifact; a release combines that artifact with deployment config/secret references; runtime executes the release. Staging/production use the same artifact shape with separate backing resources/credentials/origins.

---

## 17. Migration and release safety

Deployment migration flow:

```text
release candidate
→ one-shot migration using migration identity
→ success permits application release start
→ failure blocks release
```

Ordinary `serve` does not receive DDL authority.

T8-G does not invent permanent N/N-1 schema compatibility, dual writes or zero-downtime migration ceremony without an availability requirement. Temporary old/new application overlap is permitted only when the specific schema transition is compatible with both. T10 owns detailed cutover/rollback sequencing.

Application artifact rollback is not database recovery and must never pretend to undo a persistent incompatible schema/data change.

Use a mature SQL migration mechanism rather than implementing a migration framework. `tern/v2` is the current reference candidate for the one-shot Go/PostgreSQL profile; exact version/pinning belongs to implementation manifests.

---

## 18. Backup and restore profile

A restorable recovery point is:

```text
one consistent PostgreSQL recovery point
+ all ManagedContent required by that DB snapshot
+ exact bytes for those handles
+ manifest(handle + expected ExactContentDescriptor)
+ independently retained evidence required for post-snapshot privacy/security reconciliation
```

The backup manifest is operations metadata, not Product authority.

In-progress backup must prevent selected reclaimable DRAFT content from disappearing before capture through the T4 backup pin/GC-exclusion property.

Because River durable intents share PostgreSQL with Product state, the DB recovery point restores pre-snapshot semantic facts and their transaction-coupled intents coherently. Launch needs no separate queue-backup choreography.

T8-G does not freeze `pg_dump`, PITR, provider snapshot or object-replication product. The chosen backup profile must prove the complete recovery set and descriptor integrity.

### Recovery mode

Restore enters non-serving RECOVERY/MAINTENANCE mode. Ordinary authenticated serving remains off until all required proofs pass:

```text
DB restored and schema-compatible
all required semantic content handles exist/read
actual size == semantic size_bytes
SHA-256(actual bytes) == semantic sha256
format coherent
all restored ApplicationSessions invalidated
required post-snapshot UserProfile erasures reconciled
required post-snapshot offboarding/revocation teardown reconciled or otherwise proven safe
runtime config/secrets valid
```

No `force_ready`, `skip_corrupt_content` or privacy/security bypass exists.

Recovery evidence is not an unaudited boolean flag. Promotion RECOVERY→NORMAL occurs only after the selected recovery choreography establishes the readiness proof.

A repeatable isolated restore-drill path is required before production readiness is considered proven. Backup success alone is not restore proof.

T8-G does not invent RPO/RTO numbers. Every production recovery profile must state and measure its achieved/target RPO and restore duration; stronger infrastructure activates only when a real business/availability requirement requires it.

Automatic region failover, active-active and cross-region consensus are not Launch baseline.

---

## 19. Observability baseline

OpenTelemetry is the Launch instrumentation baseline for metrics and traces. OTLP is the vendor-neutral export boundary.

```text
METRICS / TRACES
  OpenTelemetry Go SDK

HTTP
  OTel net/http instrumentation (`otelhttp` reference)

POSTGRESQL
  pgx/pgxpool instrumentation (`otelpgx` reference)

AWS/S3 reference provider
  OTel AWS SDK v2 instrumentation (`otelaws` reference)

RIVER
  River native hooks/plugins/metric emission bridged into the OTel meter/tracer model where useful

RENDERER
  use native OTel support when supplied by the selected Gotenberg profile

EXPORT
  OTLP
```

OpenTelemetry Logs SDK is not Launch baseline while its Go implementation remains less stable than metrics/traces. Structured application logs use Go `log/slog` JSON to stdout with trace/span/request/job correlation where available.

Do not wrap OpenTelemetry in a broad proprietary MetalDocs telemetry framework. `internal/platform/observability` owns initialization, common resource attributes, redaction/cardinality policy and shutdown; application/platform code uses the real accepted OTel contracts where appropriate.

A Collector/Alloy/agent is not mandatory for the single-application baseline. Activate one only for a concrete pipeline need such as central buffering/redaction, tail sampling, multiple destinations/producers or environment enrichment.

Observability backend is replaceable behind OTLP. Grafana Cloud is a reference production candidate, not Product or architecture authority. A local all-in-one OTel/Grafana test profile may be used for development/proof.

### Logs and metrics laws

Never emit secrets, session/CSRF tokens, OIDC tokens, signed upload URLs, governed document bytes or full sensitive business payloads into telemetry.

Metrics use bounded dimensions such as operationId, outcome/status class, job kind, dependency kind and content format. High-cardinality IDs such as user/document/submission/job/managed-content ids belong in controlled logs when investigation requires them, not baseline metric labels.

Minimum operational questions:

```text
HTTP latency / traffic / errors / saturation
PostgreSQL reachability/latency/pool saturation
River available/retry/terminal counts by active required job kind
age of oldest required available/retrying work
recent job success/failure
renderer conversion success/failure/duration/timeout/saturation
MalwareInspector CLEAN/MALICIOUS/FAILURE, duration, health and signature-policy status
ManagedContentStore dependency failures/latency
spool active bytes/capacity/cleanup/verification failure
backup/recovery-point and restore-drill failure
```

Alert thresholds derive from SLO/capacity/required-work age/security policy, not arbitrary architecture constants. An alert must have a useful operator action.

AuditEvent remains Product/action evidence and is never replaced by logs, traces or job records.

---

## 20. Operational controls

Migration, job redrive and recovery controls remain private operations surfaces, not `/api/v1` operations.

Use existing third-party operational tooling before building dashboards. River APIs/tooling and a private River UI may satisfy job inspection/redrive if they preserve trust boundaries. No custom MetalDocs job dashboard is baseline.

Redrive always reloads/revalidates canonical state. It never bypasses eligibility; a dead/returned/cancelled rendition candidate remains semantic no-op even if old work is manually retried.

Scanner/renderer/recovery operator controls never include semantic bypasses such as force-CLEAN, force-Release or force-ready.

---

## 21. Reuse-first dependency law

For generic technical mechanisms the decision order is:

```text
Go/browser standard library
→ maintained official ecosystem component
→ established third-party mechanism/library
→ thin MetalDocs anti-corruption adapter
→ custom infrastructure implementation LAST
```

A dependency enters only when all are satisfied:

```text
1. named current consumer exists
2. generic/mechanism concern, not Product semantic authority
3. maintained/current security evidence is acceptable
4. API/behavior is sufficiently stable and bounded
5. license is acceptable
6. dependency surface is smaller than local reimplementation
7. actual integration/conformance proof is possible
8. replacement boundary is clear where provider coupling exists
```

Exact versions and image digests are pinned in implementation/release manifests, not frozen forever in architecture text. Upgrades require security/advisory review and the proof relevant to that boundary.

### Runtime/reference implementation families

T8-G selects or recommends these mechanism families without transferring semantic authority:

```text
HTTP                    Go net/http
observability           OpenTelemetry Go + OTLP; log/slog JSON
PostgreSQL runtime      pgx/v5 + pgxpool
PostgreSQL telemetry    otelpgx
OIDC protocol           coreos/go-oidc/v3 + golang.org/x/oauth2
S3 reference adapter    AWS SDK for Go v2 + OTel AWS instrumentation
config parsing          sethvargo/go-envconfig reference
jobs                    River (already ratified upstream)
malware                 ClamAV/clamd reference
DOCX→PDF                Gotenberg + LibreOffice reference, proof-gated
```

Adjacent implementation/tooling should follow the same law but remains owned by its proper stage. Current preferred candidates include SQL code generation (`sqlc`), SQL migration execution (`tern/v2`), real dependency test environments (Testcontainers-Go), OpenAPI property/conformance attack (Schemathesis for T9), Go vulnerability analysis (`govulncheck`) and broader dependency/image vulnerability analysis (OSV-Scanner). T8-G does not turn those T9/T11 tools into Product authority.

Do not adopt a third-party abstraction merely because one exists. Examples currently rejected without a named need include ORM, broad DI framework, Gin/Echo/Fiber, Viper-class configuration framework, custom queue/scheduler, custom OIDC implementation, custom S3 protocol/signing, custom telemetry framework, custom antivirus or custom DOCX converter.

Frontend remains the ratified T8-F shape: `openapi-typescript` generated contract shapes + one thin native-`fetch` transport + TanStack Query. `openapi-fetch` is not a T8-G requirement and must not be introduced merely for symmetry.

Strict T8-E wire behavior that no stable general mechanism fully provides remains small explicit MetalDocs transport enforcement; cryptographic primitives come from the Go standard library rather than custom cryptography.

---

## 22. Component-consumer subtraction

Every surviving runtime component has a named current consumer:

| Runtime component | Current consumer | Disposition |
|---|---|---|
| MetalDocs app | SPA + auth integration + 78 operations | KEEP |
| PostgreSQL | canonical Product state + River | KEEP |
| ManagedContentStore | exact managed content | KEEP |
| River | rendition intent + GC mechanism | KEEP |
| Renderer | conditional DOCX→PDF | KEEP / proof-gated provider |
| MalwareInspector | untrusted governed admission | KEEP |
| OIDC Provider | external Authentication | KEEP |
| HTTPS/TLS ingress | browser public origin | KEEP |
| ephemeral spool | exact-byte proof + bounded scanner/renderer I/O | KEEP |
| one-shot migration | schema evolution | KEEP |
| recovery operations | T4 restore safety | KEEP |
| Redis | none | REMOVE / NOT BASELINE |
| separate worker service | none demonstrated | REMOVE / NOT BASELINE |
| BFF | none | REMOVE |
| SSR service | none | REMOVE |
| WebSocket/realtime | none | REMOVE |
| external Search service | none | REMOVE |
| custom scheduler | none | REMOVE |
| generic event bus/outbox | none | REMOVE |
| service mesh | none | REMOVE |
| Kubernetes requirement | no architecture property | NOT REQUIRED |
| CDN | no measured consumer | NOT BASELINE |
| telemetry collector | no current pipeline need | NOT BASELINE |
| custom job dashboard | existing mechanisms sufficient | REMOVE |
| automatic failover | no RTO/availability consumer | NOT BASELINE |

---

## 23. Falsification / reopen triggers

T8-G candidate is materially falsified if evidence proves any of:

```text
F1  an accepted T8-F consumer cannot run through this topology
F2  a runtime mechanism acquires semantic Product authority
F3  in-process River workers materially compromise required HTTP availability/resource isolation
F4  one steady-state app replica cannot meet a ratified availability/throughput requirement
F5  representative DOCX corpus fails material renderer fidelity
F6  renderer cannot operate inside bounded CPU/RAM/PID/disk/time envelope
F7  MalwareInspector cannot scan accepted content safely or satisfy the T4 governed-boundary guarantee
F8  verified-spool strategy cannot support the accepted content profile sustainably
F9  chosen substrate cannot enforce required trust/network/secret boundaries
F10 backup/recovery cannot satisfy exact-content, privacy or security non-resurrection
F11 a concrete RPO/RTO requirement cannot be met by the simple recovery profile
F12 measured Search behavior activates T5 materialization or requires stronger Search mechanism
F13 multiple simultaneously active content stores become a real requirement
F14 a new application operation is required
F15 a selected third-party mechanism is unmaintained/insecure/incompatible such that replacement materially changes the topology
```

F14 requires the smallest Product/T6/T8-E reopen; T8-G must never silently invent operation 79.

Preference, generic cloud-fashion architecture, hypothetical scale and sunk cost are not reopen triggers.

---

## 24. Activation law for deferred complexity

```text
additional app replicas
  only on measured availability/throughput requirement

separate worker process
  only on measured isolation/scaling need

Redis
  only when a named cache/coordination property cannot sustainably use PostgreSQL/process-local state

materialized/external Search
  only on T5 activation conditions / measured Search failure

CDN
  only on measured SPA/static-delivery performance/cost need

service mesh
  only on a real multi-service identity/network problem

telemetry collector / full tracing pipeline
  only when direct OTLP/correlation is insufficient for a named operational need

automatic failover
  only when concrete RTO/availability requirement falsifies operator recovery

multipart upload
  only when supported-file-size evidence requires it

multiple renderer/scanner replicas
  only on measured throughput/availability need
```

Prepare the seam; do not ship dormant infrastructure.

---

## 25. T9 proof handoff

T8-G hands T9 falsifiable properties, not Product implementation permission. T9 should be able to prove at least:

```text
bad config / missing required secret → serve fails closed
incompatible schema → serve fails
PostgreSQL down → unready while process remains live
renderer down → ordinary Product serving survives; rendition work fails/retries truthfully
scanner down → required governed admission blocked while unrelated serving survives
malicious sample → cannot become immutable governed content
clean sample → scan evidence binds the exact admitted bytes
renderer outbound network → denied
representative DOCX corpus → fidelity gate passes for selected profile
provider byte corruption → complete successful exact-byte response cannot be emitted
accepted content profile → bounded heap/spool/resource behavior
SIGTERM → unready first + bounded drain + retry-safe interruption
River terminal required work → visible, inspectable and redrivable
backup recovery point → DB + required exact-content manifest complete
restore drill → session invalidation + privacy/security reconciliation before serving
secrets → absent from logs/metrics/browser
component census → every surviving component has a named consumer
selected third-party adapters → exercised against real mechanisms rather than mocks-only
```

Exact executable proof shape belongs to T9.

---

## 26. Candidate gate

This candidate is design-approved by the operator but **not ratified**.

Required closure sequence:

```text
candidate authority complete
→ Repository Standard required CI on exact candidate HEAD
→ full independent Fable review on an isolated Evidence PR
→ Lead adjudication of every finding
→ bounded corrections only when evidence supports them
→ additional independent round when material corrections change the candidate
→ CONVERGED / no surviving MATERIAL contradiction
→ explicit operator ratification
→ integration authorization and squash merge
```

Evidence PRs are never authority and never merge.

T8-H remains NOT OPEN until T8-G is operator-ratified and integrated. Product implementation remains BLOCKED.
