# T8-G Fable independent review

> **Evidence only — non-authoritative.**
> Candidate authority remains `arch/t8g-runtime-deployment`; this review branch must never merge.

## Lead handoff

Repository: `developmentconexus-ops/MetalDocs`

Candidate branch: `arch/t8g-runtime-deployment`

Exact candidate HEAD under review: `4d93066070a08fa49271dbd58cd43a830d921509`

Required candidate CI: **#1062 SUCCESS** on that exact HEAD.

Review branch: `review/t8g-fable`

Gate: **T8-G — Runtime / Process / Deployment**

Canonical Method: `developmentconexus-ops/conexus-methodology/METHOD.md` v1.0.0

Repository Standard: `developmentconexus-ops/conexus-methodology/REPOSITORY-STANDARD.md` v1.0.0

### Fresh-actor route

Reconstruct authority independently:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ docs/architecture/runtime.md
→ only the smallest accepted upstream authority needed for a concrete finding
```

Expected upstream owners that may be consulted when a specific runtime claim requires them include:

```text
docs/architecture/technical-baseline.md   # T8-A clean-slate/selective reuse
docs/architecture/backend.md              # T8-B modular monolith/composition/runtime shells seam
docs/architecture/interfaces.md           # T8-C owner/mechanism boundaries
docs/architecture/persistence.md          # T8-D PostgreSQL/River persistence laws
docs/architecture/wire-contract.md        # T8-E application wire/exact-byte laws
docs/architecture/frontend.md             # T8-F concrete browser/runtime consumers
docs/architecture/content-integrity.md    # T4 managed-content/malware/recovery authority
docs/architecture/async-and-search.md     # T5 durable work/Search authority
```

Do not broad-read these merely because they are listed. Use the smallest owning authority for each challenge. Repository current authority wins over this handoff.

Do not use removed implementation, old runtime/deployment code, legacy branches, closed PR chronology or historical mechanism shape as target authority unless a material falsifier requires exact provenance.

### Candidate target

Adversarially challenge whether T8-G is the **smallest sustainable runtime/process/deployment contract** capable of realizing the accepted T8-A→T8-F consumers without importing Product semantics into infrastructure or prebuilding scale/platform capability.

Candidate Global Maximum under attack:

```text
one modular-monolith application runtime
+ one PostgreSQL product-state database
+ River workers in the application process
+ one active ManagedContentStore per deployment
+ one private MalwareInspector
+ one private conditional DOCX→PDF renderer
+ one external OIDC provider boundary
+ verified ephemeral exact-byte spool
+ fail-closed recovery profile
+ OpenTelemetry / OTLP observability baseline
+ one-shot migration / job / recovery operations
+ proven third-party mechanisms before local infrastructure
```

The candidate claims the inverse subtraction:

```text
no Redis
no separate worker deployment
no BFF / SSR / WebSocket runtime
no external Search service
no custom scheduler/event bus
no service mesh
no Kubernetes requirement
no custom telemetry framework
no custom queue/migration/OIDC/S3/antivirus/DOCX-conversion infrastructure
no operation 79
```

A finding must attack a concrete required property, hidden authority transfer, unsafe failure mode, unjustified mechanism, missing runtime consumer, or contradiction with accepted upstream authority. Framework taste, hypothetical hyperscale and legacy-shape preference are not findings.

### Lead evidence already claimed by the candidate

Do **not** trust these claims; re-execute or falsify them where material:

```text
T8-G adds zero application operations; 78 census remains closed; operation 79 absent
T8-H remains unopened; Product implementation remains blocked
only MetalDocs app is a public MetalDocs HTTP service
River durable work remains PostgreSQL-backed mechanism state and runs in-process
renderer exists only for the conditional required DOCX→PDF transformation path
already-PDF required PDF path performs no renderer/copy/job
MalwareInspector exists because T4 requires CLEAN for exact UNTRUSTED_EXTERNAL bytes before governed admission
renderer/scanner have no DB/object-store/Product authority
renderer ordinary outbound network is denied
scanner signature update egress is bounded separately from scanned-content processing
browser direct upload uses bounded create-only capability; governed read bytes return through application origin
exact-byte serving uses complete proof before a successful response can complete; T8-G chooses verified ephemeral spool rather than whole-file heap buffering
PostgreSQL/schema/recovery barrier determine readiness; renderer/scanner/storage/OIDC outages are scoped degradation, not global readiness failure
shutdown makes readiness false first and relies on accepted DB/idempotency/River retry semantics rather than new Product states
runtime config cannot encode Product policy; secrets are process-scoped and never browser/telemetry material
ordinary runtime DB identity is separate from migration/DDL identity
backup recovery point = coherent DB + required exact content + descriptor manifest + post-snapshot privacy/security reconciliation evidence
restore remains non-serving until exact-content verification, session invalidation and privacy/security non-resurrection are complete
no invented RPO/RTO number or automatic failover requirement
OpenTelemetry metrics/traces + OTLP are the vendor-neutral observability baseline; slog JSON is the logging baseline while OTel Go logs are not treated as equivalent-stability authority
OTel is used directly rather than wrapped in a proprietary generic telemetry framework
reuse-first law prefers stdlib/proven third party/thin adapter/custom infra last
selected/reference families do not transfer Product authority and exact versions/digests remain implementation/release pins
component-consumer census gives every surviving runtime component a named current consumer
```

CI success proves only Repository Standard envelope conformance; it is not architecture convergence evidence.

### Review focus

Try to **falsify**, not confirm, the candidate. Attack at least these classes.

1. **Consumer completeness / coverage**
   - Trace every concrete T8-F runtime consumer into T8-G.
   - Look for an accepted browser, content, job, auth, recovery or operational property with no runtime home.
   - Look for a retained runtime component that exists only because it is familiar infrastructure rather than a named consumer.
   - Verify T8-G does not create operation 79 or an unadmitted Product capability.

2. **Process topology / Structural Inversion**
   - Attack the one-application-process baseline.
   - Determine whether in-process River + HTTP + GC can preserve accepted failure/isolation semantics without hidden singleton correctness.
   - Look for a real current reason a separate worker/scheduler process is required; conversely reject separation with no property.
   - Verify `metaldocs serve|migrate|jobs|recovery` as one executable does not merge privilege/trust contexts merely because the binary is shared.

3. **Semantic-owner leakage**
   - Trace transport/jobs/runtime mechanisms back through T8-B/T8-C.
   - Flag any point where renderer, scanner, queue, storage, OIDC, config, health, telemetry, recovery metadata or deployment state becomes lifecycle/Authorization/Audit/Product truth.
   - Verify operational controls cannot bypass semantic eligibility.

4. **PostgreSQL / River correctness**
   - Verify one DB and River-in-process preserve T5 transaction-coupled intent and T8-D River-schema isolation.
   - Attack shutdown/crash/redrive semantics for lost work, duplicate semantic effect or stale authority.
   - Check that backlog metrics/readiness do not accidentally become Product or serving authority.

5. **Trust / network boundaries**
   - Verify only the app is public.
   - Attack browser direct-upload CORS/capability boundaries.
   - Attack renderer/scanner privacy: no DB/storage/session credentials, no Product identity, no accidental public endpoint.
   - Attack renderer outbound denial and scanner signature-update egress separation for SSRF/data-exfiltration holes or an impossible operational profile.
   - Verify OIDC provider roles/groups never become MetalDocs Authorization by deployment configuration.

6. **Configuration / secrets / deployment identity**
   - Search for configuration that can silently change Product behavior rather than runtime mechanics.
   - Verify migration credentials cannot leak into `serve` merely because one binary implements both modes.
   - Attack workload identity/secret-loading assumptions against the platform-neutral contract.
   - Flag generic config/secret machinery with no consumer or insufficiently defined fail-closed startup behavior.

7. **Startup / liveness / readiness / degradation**
   - Attack whether PostgreSQL is exactly the dependency that must gate global readiness.
   - Find any accepted operation that would make storage/renderer/scanner/OIDC a global readiness dependency, or prove why scoped degradation is sufficient.
   - Verify liveness cannot cause restart storms on dependency outage.
   - Check deterministic startup faults fail rather than remain permanently unready.

8. **Exact-byte delivery / ephemeral spool**
   - Reconcile the verified-spool decision against T4/T8-E complete-response integrity.
   - Look for a path that could send unverified/corrupted bytes while still completing 200/Content-Length.
   - Attack heap/disk/resource behavior for accepted content profiles.
   - Verify spool is disposable mechanism state and cannot acquire recovery/retention/business meaning.
   - Compare against smaller alternatives only if they preserve the same exact-byte proof with lower total cost.

9. **Renderer / OfficialRendition**
   - Verify the renderer is invoked only when a required format transformation is actually necessary.
   - Attack late/duplicate renderer completion against T5 no-op/idempotency laws.
   - Attack resource bounding, queue/backpressure placement, concurrency assumptions and egress denial.
   - Challenge whether Gotenberg+LibreOffice is an acceptable **reference** rather than accidental semantic/vendor lock-in.
   - Verify the representative DOCX fidelity corpus gate is strong enough to reject material rendering errors without demanding byte-deterministic PDFs.

10. **Malware governed-boundary gate**
    - Reconcile scanner placement exactly with T4 trust classes and admission timing.
    - Verify the scan proof binds the same exact immutable bytes admitted to governed truth.
    - Attack false-positive handling, scanner outage and signature policy for any bypass.
    - Challenge ClamAV/clamd only on current security/fit evidence, not product taste.

11. **Backup / restore / recovery**
    - Re-execute the T4 recovery set: DB + required exact bytes + manifest + backup pin/GC safety.
    - Attack restoration of River intents, sessions, UserProfile erasures and post-snapshot offboarding/revocation.
    - Look for a path from RECOVERY to NORMAL that can be forced by a config boolean or incomplete evidence.
    - Verify deferring exact RPO/RTO numbers is legitimate and does not leave a current Product promise unimplementable.
    - Challenge the absence of automatic failover only if a current accepted availability requirement requires it.

12. **Observability / operational evidence**
    - Attack whether OTel metrics/traces + OTLP + slog JSON answer T5 operational-visibility requirements without a custom telemetry authority.
    - Verify AuditEvent remains distinct from logs/traces/jobs.
    - Attack metric cardinality, secret/PII leakage and correlation requirements.
    - Challenge direct OTLP/no-Collector baseline if a named current pipeline need makes it insufficient; otherwise reject speculative Collector/mesh additions.

13. **Reuse-first third-party law**
    - Challenge each selected/reference family on current consumer, maintenance/security posture, API stability, license, proofability and replacement boundary where relevant.
    - Current families under explicit attack include `pgx/v5`, `otelpgx`, OpenTelemetry/OTLP, `coreos/go-oidc/v3` + `x/oauth2`, AWS SDK Go v2, `sethvargo/go-envconfig`, River, ClamAV/clamd and Gotenberg+LibreOffice.
    - Separate architecture family selection from exact version/digest pinning.
    - Flag a local generic implementation where a proven mechanism should be used.
    - Also flag a third-party abstraction that is larger/riskier than a small stdlib/thin implementation.
    - Adjacent T9/T11 candidates (`sqlc`, `tern/v2`, Testcontainers-Go, Schemathesis, `govulncheck`, OSV-Scanner) must not be promoted into T8-G semantic authority merely because they are useful.

14. **Frontend/wire preservation**
    - Verify same-origin SPA, exact T8-E headers/problems/bytes, native-fetch thin transport and TanStack Query requirements remain realizable.
    - Verify T8-G does not reopen stable SPA route meaning or create a BFF/runtime API.
    - Verify runtime/health/auth integration paths remain outside the 78 application-operation census for legitimate reasons.

15. **YAGNI / subtraction / activation law**
    - Attack every rejected component: Redis, separate workers, external Search, CDN, Collector, service mesh, automatic failover, multipart upload, multiple renderer/scanner replicas.
    - If one is already required, give the concrete accepted consumer and smallest correction.
    - If not, ensure the activation trigger preserves a viable evolution seam without dormant implementation.

16. **Falsification/reopen discipline**
    - Evaluate T8-G F1→F15 for missing/overbroad reopen triggers.
    - A genuine contradiction may reopen T4/T5/T8-B/D/E/F; identify the smallest owner.
    - Do not reopen accepted authority by framework preference or legacy implementation shape.

### Output contract

Write your independent review **only in this file** below `## Fable response`.

For every finding provide:

```text
ID
severity: MATERIAL | MINOR
claim
owning authority implicated
candidate section(s)
concrete counterexample / failure mode
smallest correction
whether it reopens accepted upstream authority
```

For third-party findings additionally state, when applicable:

```text
mechanism/library
current evidence challenged
replacement consequence if the concern is valid
```

Separate non-material observations from findings.

Re-execute the strongest closure claims you can falsify mechanically or structurally, especially:

```text
candidate PR delta = runtime authority + router + roadmap only
review branch delta = exactly this ai-dialog file
application operation census stays 78; operation 79 absent
runtime component census has a named consumer for every KEEP
no Product semantic owner added
no T8-H opening / no implementation permission
no runtime/provider identifier becomes Product identity
no generic third-party mechanism bypasses current owner/application boundaries
```

If no MATERIAL finding survives, say so explicitly and list the strongest attacks attempted. Use `CONVERGED` only if you judge the exact reviewed candidate has no surviving material contradiction after those attacks.

Do not edit the candidate branch or any other file. Reviewer output is Evidence; the Lead adjudicates every finding. Do not ratify T8-G, do not merge this Evidence PR and do not open T8-H.

## Fable response
