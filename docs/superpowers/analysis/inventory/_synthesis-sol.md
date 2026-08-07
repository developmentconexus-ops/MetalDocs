# Whole-repo remediation synthesis — ARM=sol

## 1. Axes

| Axis | Root cause | Member findings (exclusive) | Total measured scale | Blocks what |
|---|---|---|---|---|
| A1 — **The verifier is not one trusted product** | Checks accumulated as workflows, scripts, allowlists, and prose without one executable manifest, proven negative fixtures, reproducible toolchain, or authority separate from the author. | CI/CD §§3–8; TEST-01–03; FE-09; GOI-01–03; PERSIST-09; SEC-05 | 20 workflows; 13 check-shaped scripts unwired; 548/553 integration test functions (342 files, 43,691 LOC) absent from the PR gate; 13 shell/PowerShell guards with no common negative-fixture harness; 4 `@latest` CLI families; no locally equivalent full gate | Trusting “green”; fast agent feedback; safe ratcheting of every later axis |
| A2 — **No ratcheted whole-repo code-quality baseline** | Feature-local delivery was allowed to add hand-written variants faster than shared primitives and static rules could make the old variants extinct. | FE-01–03, FE-06–08, FE-10; GOI-04–06, GOI-09; DUP-07; FE-02 | 42.9k TS LOC with zero general ESLint quality rules; 4 components over 400 LOC (max 738); 13 likely-dead `src/` files and 48 unused exports; 80,179 non-generated Go LOC; 110 `panic` sites requiring triage; 53 hand-written clone pairs repo-wide, including two IAM query/scan pairs | Predictable review cost; safe parallel editing; confidence that cleanup stays cleaned |
| A3 — **The API contract stops before runtime behaviour** | OpenAPI/codegen defines routes and shapes, but handlers, clients, validation, error translation, pagination, identity extraction, and idempotency remain parallel hand-authored dialects. | HTTP-01–06, HTTP-08–12; DUP-01, DUP-05, DUP-10; OBS-09–10; FE-07 | 521 error-emission calls through 6 mechanisms and 12+ local definitions; 123 OpenAPI constraints with zero request-validator wiring; 11 anonymous request structs; 18 hand-written frontend DTO files versus 4 generated-client consumers; 17 fail-open actor-helper callers; 4 pagination policies | A stable consumable API; uniform security/error semantics; contract-driven frontend changes; reliable client/server correlation |
| A4 — **Module seams expose implementations instead of capabilities** | Cross-context collaboration is expressed through foreign domain types, sentinels, SQL, and platform-to-module imports rather than consumer-owned ports and composition-root adapters. | LAYERING-01–07, LAYERING-11–13; DUP-02–03; FE-05; GOI-07 | 9/15 domain packages import `database/sql` and/or `platform/db`; 62 cross-module foreign-sentinel checks; 6+ sampled producer-owned seams; 20 platform→module edges already known plus 11 more outside tenant-data registry; 12 tenant-data adapters/1,079 LOC; 66 frontend cross-feature imports; `approval/application` 8,755 LOC/583 exports | Independent module evolution; bounded blast radius; reliable ownership; later extraction if scaling demands it |
| A5 — **Persistence correctness is maintained by visual agreement** | Transaction lifecycle, SQL text, column order, scans, and driver error mapping are repeated by hand instead of being represented once in typed query/transaction machinery. | PERSIST-01–06, PERSIST-08; DUP-04, DUP-06, DUP-08 | 82 direct `BeginTx` sites/25 files; 75 constructors take `*sql.DB` versus 6 `TxRunner`; 20 Tx/non-Tx method twins; 87 SQL-issuing files; 242 scan sites but 16 scan helpers; 15 `23505` checks/10 files; 53 clone pairs include the persistence hotspots | Safe schema evolution; transaction policy changes; retry/telemetry centralization; low-risk repository maintenance |
| A6 — **Security properties are configuration-contingent, not fail-closed facts** | Several controls are strong only when an operator chooses the right role/configuration, while startup does not prove those preconditions. | SEC-01, SEC-03–04, SEC-06, SEC-08 | 37 FORCE-RLS tables and the audit REVOKE depend on a non-superuser/non-BYPASSRLS connection; the shipped compose services use `${POSTGRES_USER}`; one re-auth timing divergence; zero CSP/HSTS/frame/type headers; KEK absence accepted for all tenants | Credible tenant isolation and audit immutability; defensible regulated deployment; safe electronic-signature re-authentication |
| A7 — **Async operations have no closed operational feedback loop** | The API was instrumented as the observable centre, but workers/jobs, propagation, logs, metrics, alerts, and deploy artifacts were not designed as one end-to-end runtime. | OBS-01–08, OBS-11–15 | 2/4 binaries have no health/metrics listener or compose healthcheck; 0 async spans; tracing in 2/15 modules; 170/215 log calls lack context-aware variants; 2 parallel HTTP metric systems; 0 alert rules; 2 mutable `latest` images; 3 near-identical Dockerfiles | Detecting stalled work; incident reconstruction; measurable SLOs; reproducible operations and release confidence |
| A8 — **ADR 0092 grant model remains dual-source** | Tier 1 and tier 2 answer authorization from disjoint grant relations, so route admission and in-transaction enforcement can disagree by construction. | ADR-0092 design program; SEC-02; ME-01/02/03/09/11/12 as already-registered prerequisites, not rediscovered lane findings | 2 grant sources; 4/15 modules have no tier-2 `authz.Require` or tripwire coverage; role vocabulary exists on 7 declaration surfaces; the existing design specifies F1–F9 | Reachable legitimate grants; one auditable authorization answer; removal of role bypasses; dependable document visibility |
| A9 — **Controlled Information is implemented as three peer contexts** | One regulated lineage/revision lifecycle and one draft-integrity question were split by artifact kind, even though template use is a version-bound role of released controlled information. | ADR 0093 implementation; LAYERING-08–10 | Migration surface: 3 modules and about 55k Go LOC; 17+ foreign-table reads among approval/documents/controlleddocuments; target domain has 2 principal aggregates plus `NumberSeries` and `TemplateUsePolicy` | Conceptual integrity of documents/templates; one regulatory lifecycle; elimination of context-internal foreign SQL and duplicated lifecycle rules |

### A1 — The verifier is not one trusted product

An inert check counts as absent. `release-readiness.yml` is manual-only; `govulncheck` is installed but skipped by default; baseline equivalence and backup/restore gates exist but never fire; full integration and race evidence first arrive after merge; `only-new-issues: true` permanently hides the old Go backlog. The immediate product is not another workflow: it is a versioned verification manifest with `fast`, `changed`, `pr`, `full`, and `release` profiles, called identically by agents, local operators, and thin CI wrappers. Every custom guard gets at least one good fixture and one bad fixture proving a non-zero exit. Tool versions and container digests are inputs to that manifest. This axis is first because every later cleanup becomes cheaper and durable once its definition of done fires before handoff.

### A2 — No ratcheted whole-repo code-quality baseline

Do not turn every grep count into a rewrite. Establish a measured baseline, block regressions immediately, and burn down only high-change hotspots. Enable `gofmt`/`goimports`, `unused`, `ineffassign`, `nolintlint`, `noctx`, `unconvert`, carefully scoped `wrapcheck`, TypeScript unused/promise/react-hooks rules, and stronger compiler flags one at a time with fixed baselines and expiry dates. Use `knip`, `jscpd`, and complexity/size reports as annotations until false positives are triaged; then block only new dead files, new clones over an agreed threshold, and new oversized components without an explicit reviewed waiver. `DocumentWorkspacePage`, IAM area queries, and approval application are hotspots, not proof that arbitrary package splitting is correct.

### A3 — The API contract stops before runtime behaviour

Make the OpenAPI document the upstream for request/response types on both sides, wire request validation at the ingress, and expose one problem writer that sets RFC 9457 fields, trace correlation, content type, status, and structured error logging. Handlers consume generated body/parameter types; the frontend consumes the generated client except for explicitly typed transport exceptions such as presigned object transfers. Pagination and concurrency remain product semantics: encode declared variants in the spec instead of forcing one dialect where offset or `If-Match` is not behaviourally justified. Actor extraction has one fail-closed helper. Idempotency stays per operation but uses one library contract for key parsing, persistence states, replay, TTL, and conflict response.

### A4 — Module seams expose implementations instead of capabilities

The target rule is consumer-owned capability ports, stable value contracts, and adapters in the composition root; a module does not import another module's repository, concrete application type, raw sentinel, or table. A module owns writes to its data and publishes the smallest read capability required by a consumer. Enforce Go edges from the package graph and scan SQL identifiers against a machine-readable ownership catalog; do not use another shrink-only list of file exceptions. Move module-specific composition out of `internal/platform`; keep truly generic transport, DB, telemetry, and middleware there. Do not introduce a reflection DI container: the current explicit `main.go` wiring is legible and is not the defect.

### A5 — Persistence correctness is maintained by visual agreement

Standardize on one unit-of-work/transaction runner so application code cannot manually forget rollback/commit and repositories do not implement Tx/non-Tx bodies twice. Run a short `sqlc` spike on the documents/approval/IAM hotspots: it can make query parameter/result shape compiler-checked and replace most inline scan order, while bespoke dynamic queries can remain behind named scanners/builders. Adopt incrementally by touched repository, not as a flag-day rewrite. Keep SQL visible; the goal is typed derivation, not an ORM. Centralize PostgreSQL error classification on pgx, remove unreachable `lib/pq` fallbacks, and use set-based updates for the confirmed release-coordinator loop.

### A6 — Security properties are configuration-contingent

Before serving, query `pg_roles` for `current_user` and fail unless `rolsuper=false` and `rolbypassrls=false`; separately prove the expected audit-table privileges. Provision and connect as an application role distinct from the bootstrap owner. Require the tenant KEK in regulated/runtime profiles and permit no-op crypto only in an explicit test profile that cannot be selected accidentally in production. Add response headers at the edge with a tested CSP appropriate to the SPA, and give approval re-authentication the same dummy-hash timing path as login. This is an urgent bounded hardening axis, not evidence for replacing the competent login/session core.

### A7 — Async operations have no closed operational feedback loop

Give worker and jobs processes separate liveness/readiness and Prometheus endpoints, with readiness tied to poll/scheduler progress rather than process existence. Use W3C trace context in outbox metadata, start consumer spans with links/parents as semantics require, and use a context-aware `slog.Handler` so trace/tenant fields are attached centrally. Retire the hand-rolled JSON metrics surface or make it a projection of the Prometheus collectors; never emit synthetic zero for “provider unavailable.” Add Prometheus recording/alert rules and Alertmanager routes for outbox age/dead letters, scheduler staleness, job failures, and dependency readiness. Pin images and consolidate the three Docker builds with one parameterized multi-stage Dockerfile.

### A8 — ADR 0092 grant-model unification

Land the written design, not a smaller compatibility layer: one `capability_bindings` relation with referentially valid user/group subjects and tenant/area scopes; one query builder supplies `Granted` and `GrantedAnyScope`, making tier 1 the existential form of tier 2. Generate role surfaces from the checked-in role catalog, keep group membership orthogonal, remove role-identity bypasses, repoint tripwire arms, and preserve grant history. The authorization relation is domain-specific and remains in-house; ME-06 correctly records that a relationship engine becomes justified only when per-object sharing appears. This axis precedes A9, as both prior advisory rounds required, because access semantics and the document-visibility query must not be migrated twice.

### A9 — Controlled Information context

Implement one Controlled Information bounded context with two aggregates: `ControlledDocument` owns stable lineage, revision allocation, and the sole effective-revision pointer; `DocumentRevision` owns mutable draft head, immutable submissions, digest/provenance, and approval receipt. The pointer switch accepting the matching approved digest is the atomic invariant; draft editing must not contend on it. A template is a `TemplateUsePolicy` relationship to an exact released revision and hash, not a permanent subtype or floating “latest” flag. Keep Approval & Evidence separate and subject-generic. Migration size may determine phasing, but is not evidence for this target. The standing inversion trigger remains: an independently evolving executable template payload earns a paired `TemplateDefinition` aggregate.

## 2. Noise — drop from the program

- **DUP-09:** domain pre-check plus DB constraint for “actor cannot sign twice” is intended defense in depth; rename only if ambiguity causes a real error.
- **DUP-02, within-context generic sentinel names:** `ErrNotFound` in separate contexts is normal namespacing. Remove only foreign-sentinel contracts and the three Controlled Information duplicates as their owning axes land.
- **FE-04:** manufacturing 11 barrel files to satisfy prose adds indirection without improving boundaries. Replace the prose rule with enforceable public-import rules; do not create barrels for ceremony.
- **FE-06 raw `px`/`rgb` totals:** tokens are not intrinsically correct for every dimension or color. Keep the existing raw-hex regression red; require tokens only for a declared semantic token set.
- **FE-11:** the 21 unused-dependency results are unverified static-analysis leads, not a workstream. Confirm each before removal.
- **GOI-08:** one constructor returning an interface has negligible carrying cost; fix opportunistically.
- **GOI-10:** one generic in 80k LOC is not a defect. Introduce generics only where a measured repeated algorithm disappears.
- **GOI-04 blanket panic count:** wiring-time and exhaustive-state panics can be correct. Triage request/data-reachable sites; do not launch a 110-site conversion.
- **GOI-05 “439 bare `fmt.Errorf`” count:** absence of `%w` does not prove an error was available to wrap. Enforce future wrapping where a causal error exists; no count-driven rewrite.
- **HTTP-07:** universal `If-Match` is not a style rule. Require it only for resources with observable concurrent-edit/stale-write semantics.
- **PERSIST-07:** cross-tenant scope of `auth_failure_counters.actor_id` is unverified. Resolve in a focused security check; do not size remediation from the grep.
- **SEC-07:** a concurrent-session cap is a product/security policy choice, not a universal baseline. Current revocation and secure session design can remain until threat model or customer policy requires a cap.
- **SEC-09:** zip-bomb protection inside the vendored editor is unverified. Audit the boundary first; no slot until exposure and existing caps are known.
- **TEST-05:** an empty backend `tests/e2e` directory is harmless when real Playwright E2E exists; delete the placeholder or ignore it.
- **TEST-06:** 30 `t.Skip` calls are not actionable until divided into environment gates and disabled tests. Report by reason; only unexplained skips become debt.
- **OBS-11 scheduler-level log:** a log line per dispatch is not a substitute for River/Prometheus state and can become noise. Metrics and alerts own this concern.
- **Action SHA pinning:** desirable supply-chain hardening, but not a standalone program axis. Pin reusable actions during A1; do not churn every workflow separately.

## 3. Do-not-touch

- Generated `api.gen.go` duplication and generated OpenAPI surfaces; generation is the anti-drift mechanism.
- Approval's internally coherent registered error catalog; reuse its semantics while replacing only the cross-module writer fragmentation.
- The 11-link middleware chain and `assertSurface` boot checks.
- `pagination.ClampLimit`/cursor primitives, `tenant.FromContext`, centralized query keys, the two intentional Zustand stores, rare TS escape hatches, and sanctioned presigned/ETag transport calls.
- `templates` pagination 50/200 exemption, which is contract-declared rather than silently divergent.
- Auth's in-memory test repository behind the same Go interface.
- Explicit `main.go` composition; no DI framework is needed.
- Zero intra-module Go layer inversions, zero package import cycles, zero frontend import cycles, and domain packages containing no literal SQL.
- Parameter binding, tenant-predicate discipline, centralized outbox insertion, and the rule that external HTTP calls do not share a DB transaction.
- DB-enforced constraints/triggers/RLS design, including app-friendly pre-check plus DB backstop; strengthen connection-role proof rather than remove these controls.
- Migration gap/no-historical-edit CI checks and the CI-only non-owner RLS test role.
- `tools/cilint` and `scripts/api-lint` compiled analyzers with positive/negative fixtures; extend their standard, do not replace them.
- IAM `authz.Require` core tests and live DB tripwire tests.
- Login/session core, secure cookie defaults, CORS validation, trusted-proxy default deny, typed config package, gitleaks, and blocking gosec.
- Audit hash-chain computation inside an immutable DB function.
- `slog` JSON as the single logger, no observed secrets/PII in logs, existing `otelhttp`/`otelsql`, the isolated metrics listener, and the API's real dependency readiness probe.
- CI caching, `api-contract.yml`'s seven-job design, OpenAPI breaking-change check, requirement-trace regeneration, and the explicitly advisory accessibility limitation.

## 4. Target methodology

### Gate topology

| Level / firing point | Blocking | Annotating | Why |
|---|---|---|---|
| **Unrepresentable in code/data** | Generated API/role types; FKs/checks/unique effective pointer; one authz evaluator; typed query outputs; one transaction runner API | — | Prefer this whenever the wrong state can be removed rather than detected. |
| **Boot-fatal** | DB role is non-superuser/NOBYPASSRLS; required secrets/KEK by runtime profile; API surface parity; role/capability registry parity; required dependencies configured | Optional dependency degradation only where product behaviour explicitly permits it | A process that cannot uphold tenant/audit/security invariants must not become ready. |
| **Red local/PR build** | Formatting; compile/typecheck; generated drift; custom boundary/sole-reader guards; migration invariants; OpenAPI lint/breaking checks; unit tests; affected integration shards; full integration+races before merge for DB/platform/security changes; gitleaks; gosec; govulncheck; High/Critical Grype; negative-fixture suites; critical accessibility failures | Duplication, complexity, coverage trend, knip, medium/low CVEs, noncritical accessibility, performance until stable budgets exist | A blocker must be deterministic, locally reproducible, and have a demonstrated bad fixture. Otherwise it teaches bypassing. |
| **Runtime assertion/alert** | Request-schema validation, authorization denial, optimistic concurrency where declared, idempotency state machine, DB tripwire | SLO warning thresholds before an error budget is ratified | Input and distributed-runtime properties cannot all be compile-time facts; fail closed at the owning boundary. |
| **Discipline** | Only domain judgments that cannot be mechanized: role-bundle correctness, acceptance of an ADR boundary, one-time migration review | Architecture trigger review | Discipline is never the only guard for an enumerable or syntactic property. |

Per axis, the highest attainable firing level is:

| Axis | Highest attainable level | Why not higher |
|---|---|---|
| A1 | **Red build plus external merge authority** | Trust in the judge is a governance property, not representable in product code. |
| A2 | **Red build** | “Good decomposition” is judgment; formatting, dead code, forbidden APIs, new clones, and size regressions are mechanical. |
| A3 | **Unrepresentable + boot/runtime** | Generated types remove shape drift; client input remains runtime data and requires validation. |
| A4 | **Compiler/red build** | Go types can enforce ports; ownership hidden in SQL strings needs an AST/parser gate unless DB roles are split per context, which is disproportionate for this monolith. |
| A5 | **Unrepresentable** | Typed generated queries and a unit-of-work API can make scan-order and manual transaction drift unwritable; truly dynamic SQL retains tested runtime paths. |
| A6 | **Boot-fatal** | Deployment identity/secrets are runtime facts; headers and timing equivalence remain tests/runtime behaviour. |
| A7 | **Runtime assertion/alert** | Liveness, lag, and dependency failure exist only in a running distributed system. |
| A8 | **Unrepresentable** | One binding relation, FKs, one builder, and generated catalogs eliminate dual-source states; bundle content remains human policy. |
| A9 | **Unrepresentable** | Aggregate APIs plus a DB uniqueness/pointer invariant prevent two effective revisions; whether a future payload deserves a new aggregate remains judgment. |

### In-loop agent gate

The authoring loop must fire before PR review:

1. `verify preflight --changed-from <base>` derives affected modules, contracts, DB objects, and mandatory checks from the diff graph; the agent does not choose the lane.
2. Before implementation, the agent records observable invariants and at least one falsifying counterexample in the task evidence, using product/standard nouns rather than current type names.
3. After each coherent patch, `verify changed` runs formatting, generators, type/compile, semantic guards, focused unit tests, and the smallest affected integration shard. It must stay below roughly five minutes on existing hardware.
4. A separate context-isolated reviewer receives the brief, diff, canonical contract/standards, and counterexamples—but not the author's rationale first—and attempts to falsify noun choice, ownership, authorization, tenant isolation, and migration assumptions.
5. `verify pr` runs all blocking PR gates and emits a content-addressed JSON evidence bundle: source SHA, guardian-policy SHA, tool/container versions, commands, exit codes, and artifacts. No prose “green” substitutes for it.

The wrong-noun failure is addressed by deriving guards from independent upstreams and testing how they fail. Every semantic guard declares: the noun it checks, fixtures that must fail, and explicit non-claims. String greps are acceptable only for literal extinction; architectural rules use Go AST/package graphs, parsed OpenAPI, migration/schema queries, or property tests. An agent may add a product test in its change, but may not make that self-authored test the sole evidence for a security, boundary, or migration invariant.

### Separation of powers with a solo operator

GitHub branch protection is unavailable/unverifiable on the present private-repo tier, so do not pretend workflow exit codes enforce merge. Enforcement is credential custody plus an external guardian:

- Agents work without the `main` push credential. Only the operator can merge/push.
- Keep the verification manifest, gate runner, protected-path policy, and merge command in a small separate `metaldocs-guardian` repository to which agent credentials are read-only. Product CI calls a pinned guardian commit.
- An operator-owned `mergectl` outside this writable repository fetches the candidate SHA into a clean worktree, rejects mixed product+policy changes, runs the guardian's `verify pr`, checks the evidence hash, and only then fast-forwards/merges with the operator credential.
- Product changes may not modify `.github/workflows/**`, lint configs, guard sources, allowlists/baselines, or the guardian pin. Policy changes use a policy-only PR/change, must add or update negative fixtures, run against the old and new policy, and receive explicit operator approval before the guardian pin advances.
- CI remains fast feedback; `mergectl` is the actual enforcement point. The operator can override only with a signed, reasoned break-glass record committed to the evidence log. Agents cannot create that signature.

This separates author from judge without paid SaaS or a second human. It does not claim the solo operator is unable to subvert their own process; it prevents the machine author from silently weakening its judge in the same change.

### Free tooling and run cost

| Tool | Replaces / supplies | Run cost |
|---|---|---|
| Pinned `golangci-lint`, `staticcheck`, `go vet`, `gofmt`, `goimports`, existing `cilint` | Grep review and formatting/idiom discipline | Seconds to a few minutes; cacheable |
| `tsc`, ESLint (`@typescript-eslint`, react-hooks, promises), `knip`, `jscpd` | Hand review of TS correctness, dead surfaces, clones | Seconds/minutes; knip/jscpd annotate initially |
| `oapi-codegen`, `openapi-typescript`/`openapi-fetch`, Redocly, `oasdiff`, `kin-openapi` validation | Hand-synced route/DTO/constraint dialects | Generation/lint minutes; validation per request |
| `sqlc` plus pgx | Most handwritten result scanning and dual driver handling | Build-time generation; no service/license; migration labor is the cost |
| `go/packages`/AST custom analyzers and a SQL parser such as `pg_query_go` | Regex boundary and sole-reader checks/allowlists | Seconds to low minutes |
| Pester + Bats, existing Go fixture tests | Untested PowerShell/shell gates | Seconds per guard suite |
| PostgreSQL 16 service, Go test sharding, race detector, Playwright/Vitest | Post-merge-only correctness evidence | 10–25 min PR wall time when parallel; nightly stress longer; use the operator's self-hosted runner to avoid paid minutes |
| gitleaks, gosec, govulncheck, Syft, Grype, `actionlint`, `zizmor` | Secret, code, dependency, SBOM, workflow checks | 1–5 min each with pinned DB/tool caches |
| OpenTelemetry Collector, Prometheus, Alertmanager, Grafana or Jaeger OSS | Ad-hoc log grep, duplicate JSON metrics, no alerts/trace view | Existing host CPU/disk; no license/SaaS spend |
| Git + SSH signing + external `metaldocs-guardian`/`mergectl` | Unverifiable branch protection/CODEOWNERS | Seconds plus the same verification workload |

## 5. Sequence

| Order | Axis | Rough size | Dependency and parallelism |
|---|---|---:|---|
| 1 | **A1 trusted verification spine** | 6–9 days | First and highest leverage. Build the manifest, negative-fixture standard, changed-slice selector, pinned tools, evidence bundle, guardian, and merge command. Do not wait to wire every slow release check before proceeding. |
| 2 | **A8 ADR 0092 grant unification** | 10–15 days | Starts immediately after the minimal A1 spine. It must precede A9 and must not be shelved behind it. Includes migration, evaluator, OpenAPI/frontend re-point, parity corpus, tripwire, and F1–F9. |
| 2a | **A6 fail-closed security posture** | 4–7 days | Parallel with A8 after A1. DB-role/privilege boot proof, explicit app role, KEK profiles, headers, and re-auth timing are bounded and urgent. Coordinate only where A8 touches authz tripwires. |
| 3 | **A3 contract-to-runtime completion** | 10–15 days | After A1; can overlap late A8 except on IAM routes. Land one writer/validator/generated client path, then migrate modules in slices. This makes later module moves cheaper. |
| 3a | **A2 quality ratchet and hotspot cleanup** | 8–15 days, then ongoing | Parallel with A3. First land regression-only rules (3–5 days), then clean high-change hotspots. Do not block A8 on historic lint debt. |
| 4 | **A5 persistence mechanics** | 15–25 days | Needs A1; benefits from A3's settled error/DTO conventions. Spike `sqlc` first (2 days); adopt transaction runner and migrate approval/documents/IAM hotspots before the long tail. |
| 5 | **A4 executable module seams** | 12–20 days | After A8 fixes IAM vocabulary/evaluator and after A3 establishes published contract patterns. Can overlap A5 on modules not being edited. Land graph/SQL ownership gates before moving seams. |
| 6 | **A9 Controlled Information implementation** | 20–30 days | After A8; preferably after A3 and A4 guards. Phase by schema/domain, application adapters, HTTP/frontend, then deletion. Migration cost affects this order only, not the target. |
| 7 | **A7 async operational loop** | 8–12 days | Health/metrics can start after A1 in parallel; trace propagation and outbox schema work should follow A3/A5 to avoid rework. Alerts land only after trustworthy metrics exist. |
| 8 | **Long-tail ratchet closure** | 10–20 days spread across normal work | Remove expired baselines/waivers, finish typed-query migration, and promote stable annotations to blockers. It is not a separate refactor freeze. |

A1 is the axis that makes every later axis cheaper. A8 and A6 are the first product changes; A3 and A2 can parallelise; A5 and A4 can parallelise by module; A7's process health is independent while trace propagation is not.

## 6. Mandatory inversion tests

- **Trusted judge outside the product change:** survives an opposite implementation because author/judge conflict exists for any codebase whose author can edit its own acceptance mechanism.
- **Generated contract as the API upstream:** survives an opposite route/handler layout because one externally observable protocol still requires one machine-readable source and derived consumers.
- **Consumer-owned ports and owner-mediated data access:** survives an opposite import graph because independent concepts need change insulation; producer types and foreign SQL make consumers depend on implementation decisions in any layout.
- **One unit-of-work and typed query authority:** survives an opposite repository layout because begin/commit/rollback and row-shape agreement are cross-cutting correctness laws, not module-placement facts.
- **Boot-proven non-bypass DB identity:** survives an opposite schema/compose design because FORCE RLS and privilege revocation are ineffective for a bypassing connection in every topology.
- **One Prometheus/OTel operational truth:** survives an opposite binary split because contradictory metrics and lost causal context remain failure modes regardless of process placement.
- **One authorization binding relation and evaluator:** survives an opposite table/module arrangement because route admission and in-transaction authorization answer the same product question and must be monotonic views of one relation.
- **One Controlled Information context:** survives an opposite current implementation because regulated control follows stable lineage, immutable released revision, effectivity, and exact template-version provenance—not existing packages, routes, or tables.
- **Two aggregates (`ControlledDocument`, `DocumentRevision`):** survives an opposite current transaction model because only the effective-pointer switch requires atomic coordination, while keystroke drafts have a different contention/lifecycle boundary.
- **Template as exact-version use policy:** survives an opposite current type hierarchy because a blank/master form must be approved and traceable to an exact revision, while designation can move to a later released revision; neither a permanent subtype nor floating latest reference preserves both behaviours.
- **Approval & Evidence remains separate and subject-generic:** survives an opposite current coupling because signature/case evidence has its own immutability and route-snapshot lifecycle and can govern future non-document subjects.
- **Template inversion trigger:** also survives: if executable template payload/schema migration gains an independent lifecycle, a paired `TemplateDefinition` aggregate becomes justified even if the current merge is complete.

## 7. Where I disagree with a lane

- **SEC-03/04 is over-asserted.** The compose file proves its own services use `${POSTGRES_USER}` and therefore commonly use the bootstrap superuser; it also explicitly labels key PostgreSQL settings dev/test-only. This proves the shipped compose shape cannot validate RLS/audit privilege posture and that no boot assertion protects other deployments. It does **not** prove an unseen production deployment uses a superuser. A6 is sized around eliminating the uncertainty.
- **PERSIST-10 is too broad when it says the RLS trap is closed.** It is closed for the integration isolation suite by `metaldocs_ci`; it is not closed for the application connection in compose. The two lane claims concern different roles.
- **TEST-02's table and prose conflict.** The finding row says 5 of 13 guards lack tests; its detail says all 13 shell/PowerShell guards lack a sibling/common harness. The actionable claim is that the shell guard layer lacks systematic bad-fixture proof; A1 must inventory exact per-script coverage before setting a completion count.
- **GOI-05 is mis-sized as error-chain loss.** `fmt.Errorf` without `%w` is often formatting a new error, so 439 is an upper bound on candidates, not broken chains. Use AST-aware `wrapcheck`/review and count only sites wrapping a causal error.
- **GOI-07 does not prove a package split.** 8,755 LOC/583 exports proves a hotspot and high change surface; cohesion and product use cases must determine subpackages. It belongs to A4/A2 discovery, not an automatic restructuring mandate.
- **HTTP-05's “0 of 13 modules validate” is the wrong deployment unit.** Validation should be wired once at the composed ingress, not once per module; the measured absence is real, the denominator should be one server surface.
- **HTTP-07 overgeneralizes optimistic concurrency.** Approval's 17-file use does not imply every mutation in 12 other modules needs ETags. Product-observable stale-write risk decides it.
- **FE-04 mistakes a documentation convention for engineering value.** Missing barrels do not cause the 66 cross-feature imports; enforce public boundaries directly and amend the rule.
- **FE-06's 2,005 raw-pixel count is not a defect size.** Only literals that duplicate a declared semantic token are drift. Raw dimensions can be local and correct.
- **CICD's 359 integration-tagged files and TEST-01's 342 files are not necessarily contradictory:** the former counts all tagged Go files repo-wide, the latter integration test files in its scoped census. Use TEST-01's 553-function/342-file figure for the pre-merge test gap and recompute from the future manifest once.
