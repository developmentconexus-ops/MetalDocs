# Cross-Repository Engineering Alignment — MetalDocs ↔ Marketplace Central

> **TEMPORARY REVIEW EVIDENCE ONLY — NOT PRODUCT / ARCHITECTURE / ROADMAP AUTHORITY**
>
> Delete this file after convergence. Do not merge it into MetalDocs `main`.

## Exact review subjects

```text
METALDOCS
repository        developmentconexus-ops/MetalDocs
current main      82832cce62d11ea90575fb484b97e3c934c03e37
T1→T8-H           CLOSED / OPERATOR-RATIFIED / INTEGRATED
T9                NEXT / NOT STARTED — requires explicit operator authorization
T10→T12           NOT OPEN
implementation    BLOCKED

MARKETPLACE CENTRAL
repository        developmentconexus-ops/marketplace-central
candidate branch  stage/d6-frontend
candidate SHA     cb55238c1908b087989825ff4d2ad9ce6f08527b
candidate PR      #54
D6-B1             OPERATOR-RATIFIED
D6-B2             ACTIVE DECISION
D7–D9             BLOCKED
implementation    BLOCKED UNTIL D9
```

This review was deliberately rebased onto integrated MetalDocs `main` after PR #148 merged during setup. The stale cross-review PR #152 was closed unmerged.

## Operator goal

Align engineering decisions across the two products where the **protected property and failure class are genuinely the same**, while preserving justified differences. The operator explicitly does **not** require shared code or a shared platform today.

Use DevelopmentConexus Root Cause / Global Maximum / YAGNI / falsification discipline.

Required classifications:

```text
ALIGN
DIVERGE_JUSTIFIED
REOPEN_MARKETPLACE
REOPEN_METALDOCS
DEFER
STOP
```

## Interaction rule

1. Start fresh from MetalDocs authority:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ only the bounded task authority pack
```

2. Then inspect Marketplace current authority:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ docs/engineering/rebaseline/D6-FRONTEND.md + ARCHITECTURE.md
```

Switch to Marketplace D2 only for the concrete AuthN/identity question. Use the routed engineering-research guide only for concrete technology research.

3. Read the incoming Marketplace review at:

```text
repository: developmentconexus-ops/marketplace-central
branch:     review/d6b2-metaldocs-alignment
PR:         #57
file:       docs/work/current/ai-dialog.md
```

Do not assume its findings are correct. It is Evidence to attack.

4. Use current official docs/upstream repositories for technology claims that may have changed.

5. Write only under `## MetalDocs reciprocal response` in this file. Modify no other file.

## Review scope

Perform a reciprocal review of Marketplace Central and, where Marketplace evidence exposes a stronger solution, challenge already-ratified MetalDocs technology decisions too.

At minimum adjudicate:

- React SPA realization;
- TanStack Query server-state ownership;
- OpenAPI TypeScript generation;
- native `fetch` thin transport vs `openapi-fetch` / Orval / Hey API / other current alternatives;
- route tree and whether a router dependency is materially justified;
- feature/package topology: MetalDocs lens-first vs Marketplace's unapproved owner-first hypothesis vs bounded hybrid;
- form and UI component dependencies;
- external OIDC boundary;
- Keycloak as first concrete provider while architecture remains provider-neutral;
- browser ApplicationSession / cookie / CSRF model;
- Go modular-monolith topology and package-direction law;
- Go `net/http` vs framework;
- generated Go OpenAPI boundary;
- PostgreSQL + pgx / SQL generation / migration tooling;
- tenant isolation and whether Marketplace justifiably needs a stronger mechanism than MetalDocs;
- River or another durable-job mechanism;
- OpenTelemetry / OTLP / slog;
- configuration, test-environment and security tooling;
- same-origin SPA/API profile;
- dependency-version alignment policy;
- mechanical frontend/backend dependency boundaries;
- deliberate absences: BFF, SSR, Redis, microfrontends, realtime, generic event bus/workflow, ORM.

## Required adversarial questions

1. Is `openapi-typescript + thin native fetch + TanStack Query` still the Global Maximum shared frontend baseline, or does Marketplace expose a real gap?
2. Does Marketplace need a router library where MetalDocs did not? If yes, is that a justified divergence or should MetalDocs reopen too?
3. Is MetalDocs lens-first feature topology stronger than Marketplace owner-first folders, or does Marketplace expose a better hybrid?
4. Does Marketplace's owner/composition model expose any material weakness in MetalDocs T8-F?
5. Is Keycloak the best first concrete IdP for both products while keeping OIDC/provider-neutral architecture?
6. Which MetalDocs Go/runtime decisions should Marketplace D7 begin from rather than rediscovering from zero?
7. Which decisions must *not* be copied because Marketplace's tenant/provider workload differs?
8. Is River a real shared candidate or only familiarity bias?
9. Is MetalDocs' no-RLS Launch posture correctly different from Marketplace tenant-ready isolation?
10. Are any MetalDocs technology choices now materially inferior under current 2026 upstream evidence?
11. Are any Marketplace accepted decisions materially weaker than MetalDocs and worth reopening?
12. What is the smallest shared technology profile that reduces engineering/LLM cognitive drift without creating coupled releases or a generic internal platform?

## Required response shape

For each material item:

```text
ID
PROPERTY / FAILURE CLASS
METALDOCS CURRENT DECISION
MARKETPLACE CURRENT DECISION
CURRENT PRIMARY EVIDENCE
CLASSIFICATION = ALIGN | DIVERGE_JUSTIFIED | REOPEN_MARKETPLACE | REOPEN_METALDOCS | DEFER | STOP
RATIONALE
SMALLEST NEXT ACTION
REOPEN TRIGGER
```

End with exactly these sections:

```text
SHARED PROFILE — safe to align now
REPO-SPECIFIC DIFFERENCES — must remain different
DEFERRED PROFILE — align later only after owning stage proves need
METALDOCS REOPEN DECISION
MARKETPLACE REOPEN DECISION
CONTINUATION RECOMMENDATION
```

Do not open T9. Do not begin Marketplace D7. Do not implement code. Reviewer output is Evidence, not authority.

---

## MetalDocs reciprocal response

### Reciprocal verdict

The Marketplace Lead review is directionally strong, but two preliminary conclusions do not survive current evidence unchanged:

1. **Browser bearer OAuth is not merely a D7 mechanism question for Marketplace.** The current accepted human-browser security shape is materially weaker than the same-origin session pattern already ratified by MetalDocs and than current IETF browser-application guidance. A targeted Marketplace D5 authentication/wire reopen is justified before D6-B2 freezes browser transport.
2. **MetalDocs' hand-authored native-fetch wrapper is no longer the strongest common transport candidate.** Current `openapi-typescript` upstream now provides `openapi-fetch`, a tiny typed wrapper over native Fetch that preserves raw `Response`, headers, streams and middleware while removing hand-authored path/parameter/body serialization. This is a real anti-drift property, not symmetry. A bounded MetalDocs T8-F mechanism reopen is justified for adjudication; no Product/wire semantic reopen follows from it.

No other MetalDocs T1→T8 semantic or runtime decision is materially falsified by Marketplace evidence.

### XR-01 — React SPA + server-state ownership

```text
ID
XR-01

PROPERTY / FAILURE CLASS
Prevent browser UI state from becoming a second Product/server authority while retaining reusable asynchronous server-state caching.

METALDOCS CURRENT DECISION
React SPA; TanStack Query owns server state; URL/navigation, form draft and ephemeral UI are separate classes; no Redux/Zustand/global entity mirror.

MARKETPLACE CURRENT DECISION
React accepted; D6-B1 ratifies the same four state classes and TanStack Query server-state ownership.

CURRENT PRIMARY EVIDENCE
MetalDocs T8-F §§7–8; Marketplace D6 §§1–3; current TanStack Query/React guidance remains compatible with this separation.

CLASSIFICATION = ALIGN

RATIONALE
The protected property and failure class are identical. A second normalized/global server-state store would duplicate authority in both products. Marketplace's larger owner count does not change the browser-state law.

SMALLEST NEXT ACTION
Freeze the same state-class law in Marketplace D6-B2. Keep query identity based on operation + canonical semantic inputs and keep mutation invalidation bounded to affected lenses/reads.

REOPEN TRIGGER
A concrete frontend consumer proves server state cannot be represented safely/sustainably through TanStack Query plus the three local/navigation classes.
```

### XR-02 — OpenAPI TypeScript generation and browser transport

```text
ID
XR-02

PROPERTY / FAILURE CLASS
Prevent hand-authored wire drift in paths, params, serialization, request/response types and headers without generating a second business/query authority.

METALDOCS CURRENT DECISION
`openapi-typescript` paths/components + one thin native-`fetch` transport + feature query/command functions.

MARKETPLACE CURRENT DECISION
`openapi-typescript` projection is already proved; exact D6 runtime client remains open.

CURRENT PRIMARY EVIDENCE
Current `openapi-typescript` official docs recommend `openapi-fetch`; current `openapi-fetch` is a ~6 kB thin wrapper over native Fetch, infers method/path/path+query params/body/success/error from generated `paths`, exposes the original `Response`, supports custom serializers, middleware and `blob|stream|arrayBuffer` parsing. Orval generates TanStack Query hooks/query keys and SDK layers; Hey API exposes an opt-in SDK/plugin platform. Those broader surfaces are not required by either repository's current authority.

CLASSIFICATION = REOPEN_METALDOCS

RATIONALE
`openapi-fetch` now protects a material failure class that MetalDocs' manually-authored transport would otherwise have to reimplement: exact path/parameter/body serialization tied to the OAD. It does so without owning cache, Authorization, lifecycle, retries or business meaning. MetalDocs' required CSRF, Idempotency-Key, If-Match/If-None-Match, Problem decoding, ETag preservation and exact-byte flows remain explicit thin transport policy because `openapi-fetch` exposes middleware/options and raw `Response`.

For Marketplace, `openapi-fetch` is a stronger D6-B2 baseline than Orval/Hey API/openapi-react-query because Marketplace already chose TanStack Query and needs operation-/lens-specific query identity, retries and invalidation rather than generated generic hooks becoming a second behavior layer.

SMALLEST NEXT ACTION
Marketplace D6-B2: select `openapi-typescript + openapi-fetch + TanStack Query`, subject to XR-05 browser-auth repair. MetalDocs: bounded T8-F mechanism adjudication only, replacing "native-fetch transport" with "thin Product transport backed by openapi-fetch/native Fetch" after a disposable proof across representative JSON, conditional, idempotent and exact-byte operations.

REOPEN TRIGGER
Reject `openapi-fetch` if the exact MetalDocs/Marketplace OAD proves a required serialization/header/body/stream behavior cannot be represented without unsafe casts, hidden semantic behavior or a larger custom bypass than the native wrapper it replaces.
```

### XR-03 — Frontend router

```text
ID
XR-03

PROPERTY / FAILURE CLASS
Keep route params and shareable URL state typed/validated without letting routing acquire server/business authority or duplicating server-state loading.

METALDOCS CURRENT DECISION
Stable route meanings are ratified; router library intentionally unfrozen. MetalDocs has route params plus list/search/filter/cursor URL state.

MARKETPLACE CURRENT DECISION
Router dependency open in D6-B2; Organization/Installation context, periods/comparison, search/filter and route params are materially richer URL state.

CURRENT PRIMARY EVIDENCE
Current TanStack Router documentation provides end-to-end typed routes and runtime-validated typed search params, including parent/child search inheritance. Current React Router v7 provides strong route-module/URL-param types, but its normal search-param API remains `URLSearchParams`; Framework/Data modes additionally introduce loaders/actions/revalidation that overlap with the repositories' explicit TanStack Query server-state model.

CLASSIFICATION = ALIGN

RATIONALE
TanStack Router directly protects the same "URL state is a separate typed state class" property in both products and is especially valuable for Marketplace. It does not require SSR or router-owned server state. MetalDocs did not ratify an alternative library, so adopting it later does not reopen route semantics.

SMALLEST NEXT ACTION
Marketplace D6-B2 may select `@tanstack/react-router` as the router dependency. Shared profile should prefer the same router for MetalDocs implementation. Keep search validation transport/presentation-only; do not let Zod/Valibot/router schemas become Product-domain authority.

REOPEN TRIGGER
Measured TypeScript/build complexity, route-scale performance or an exact browser behavior requirement proves TanStack Router creates more complexity than it removes; or either product needs router-owned data/action semantics that materially conflict with TanStack Query ownership.
```

### XR-04 — Frontend package topology

```text
ID
XR-04

PROPERTY / FAILURE CLASS
Prevent frontend folders from either mirroring backend domain topology mechanically or duplicating reusable Product-operation consumption across multiple user flows.

METALDOCS CURRENT DECISION
Top-level `features/*` are stable user/lens flows (Library, Document Official, Document Work, Governance Work, History, Audit, Admin); frontend topology is explicitly not a one-for-one projection of Go semantic owners.

MARKETPLACE CURRENT DECISION
D6-B1 freezes task-oriented UX and cross-owner composition; D6-B2 topology is open. An owner-first feature-folder hypothesis has not been ratified.

CURRENT PRIMARY EVIDENCE
Marketplace `Venda`/sale experiences compose Sales + Economics + Materialization/Fulfillment; Overview and Settings also compose multiple authorities. Conversely Performance, Market and Economics have owner-native user spaces. MetalDocs similarly has lens routes that consume multiple accepted read/authorization/configuration authorities.

CLASSIFICATION = ALIGN

RATIONALE
Strict owner-first UI folders fail on real composed screens and encourage UI-to-domain mirroring. Strict lens-only code can duplicate Product-operation adapters used by several lenses. The Global Maximum is a bounded hybrid:

UI composition/route ownership follows stable human lenses/flows.
Reusable Product operation adapters may be grouped by semantic owner/operation family, but are stateless API-consumption code and acquire no client business authority.

A representative non-normative shape is:

src/app/                 shell + router/composition
src/features/<lens>/     route/lens UI + local form/UI state
src/api/<owner-family>/  reusable Product query/command adapters only
src/lib/api/             generated types + thin transport
src/ui/                  bounded reusable presentation primitives

Exact folder spelling is not the decision.

SMALLEST NEXT ACTION
Marketplace D6-B2 should reject strict D1-owner mirroring and freeze this hybrid law. MetalDocs T8-F remains coherent; no topology reopen is required.

REOPEN TRIGGER
A concrete feature proves the lens/API-adapter split creates circular imports, duplicated semantic logic or an unavoidable fifth state/authority class.
```

### XR-05 — Human browser authentication/session

```text
ID
XR-05

PROPERTY / FAILURE CLASS
Prevent theft/exfiltration of OAuth access/refresh tokens from browser JavaScript in a first-party business application while preserving external OIDC authentication and current Product authorization.

METALDOCS CURRENT DECISION
External OIDC authentication; server-side ApplicationSession; Secure HttpOnly SameSite=Lax host-only cookie; synchronizer CSRF token on unsafe requests; browser never receives OIDC access/refresh tokens as Product API bearer credentials.

MARKETPLACE CURRENT DECISION
D5-B2-A selected human Authorization Code + PKCE as a browser public OAuth client, then audience-bound bearer access token to the MPC Product API; the canonical OAD uses `MpcBearerAuth` on every Product operation. Machine Client Credentials uses the same bearer resource-server boundary.

CURRENT PRIMARY EVIDENCE
Current IETF `draft-ietf-oauth-browser-based-apps-27` (July 2026, intended Best Current Practice) states that a proxying BFF is strongly recommended for business applications, sensitive applications and applications handling personal data; it states browser-only OAuth substantially increases attack surface and is not recommended for those applications; and its same-domain discussion explicitly notes that a SPA backed by its own server often does not need OAuth between frontend and backend and can use protected cookie-based session state while retaining OIDC for login. Marketplace is exactly a first-party business application with PII and consequential writes. MetalDocs' already-ratified session pattern matches that threat model.

CLASSIFICATION = REOPEN_MARKETPLACE

RATIONALE
This is not library preference. Marketplace's browser bearer decision exposes transferable access tokens to the JavaScript execution context, while its current Product does not require the browser to call an independently hosted third-party resource server. The current IETF guidance materially changes the Global Maximum for this failure class.

Human and machine authentication should split without splitting Product authority:

H browser:
  browser -> same-origin MPC application session cookie
  server -> OIDC Authorization Code exchange as confidential/server-side client
  OIDC tokens remain server-side
  unsafe browser requests -> CSRF protection

A/S machine:
  OAuth Client Credentials (or later stronger confidential-client auth)
  audience-bound bearer token -> Product API

No shared MetalDocs/MPC session, user table or Permission model follows from this alignment.

SMALLEST NEXT ACTION
Before final D6-B2 transport freeze, perform a bounded D5-B2-A/W4/canonical-OAD security repair for dual human-session vs machine-bearer admission. Preserve D2 Principal/Membership/Permission semantics and all operation/client-class mappings; change only the authentication carrier/profile needed by each allowed caller class. Do not begin D7 implementation.

REOPEN TRIGGER
Already satisfied: current IETF browser-app security guidance + Marketplace first-party same-domain business/PII profile materially falsifies "browser bearer is the smallest secure human-client baseline". A future independent-resource/third-party delegated-client requirement may reopen the human profile again.
```

### XR-06 — OIDC provider and Keycloak

```text
ID
XR-06

PROPERTY / FAILURE CLASS
Use mature standards-based AuthN without making provider roles/claims Product authority or creating vendor lock-in.

METALDOCS CURRENT DECISION
Provider-neutral external OIDC boundary; `coreos/go-oidc/v3 + golang.org/x/oauth2` reference mechanism; provider roles/groups never become Product Authorization.

MARKETPLACE CURRENT DECISION
Same provider-neutral OIDC authority; Keycloak is the preferred first self-hosted candidate; provider/deployment/realm topology deferred.

CURRENT PRIMARY EVIDENCE
Current Keycloak documentation continues to expose standards-compliant OIDC and explicitly recommends using normal application-ecosystem OIDC support before tightly coupled Keycloak adapters. Keycloak 26.7 remains actively maintained as of July 2026.

CLASSIFICATION = ALIGN

RATIONALE
Keycloak is a credible common first provider because both repositories need the same standards boundary and self-hosted identity capability. The shared technology should be Keycloak-as-provider, not Keycloak-specific business APIs/adapters.

SMALLEST NEXT ACTION
Use Keycloak as the first concrete provider candidate for both implementations, with provider-neutral Go OIDC libraries. Keep per-application clients and Product bindings distinct. Whether one physical Keycloak deployment hosts separate realms/clients is an operations decision, not part of this cross-repo architecture result.

REOPEN TRIGGER
Keycloak cannot satisfy a proven availability, federation, MFA, lifecycle, audit, upgrade or operational property at lower total complexity than an alternative conformant OIDC provider.
```

### XR-07 — Go modular-monolith class and dependency law

```text
ID
XR-07

PROPERTY / FAILURE CLASS
Keep semantic authority isolated while allowing one-process/local-transaction composition without service-distribution tax or cross-owner private coupling.

METALDOCS CURRENT DECISION
One Go module; owner-first modular monolith; one public surface per semantic owner; stateless application orchestration; transport/platform/composition classes; default-deny first-party dependency graph.

MARKETPLACE CURRENT DECISION
Go backend canonical; stable target reasoning already separates contexts/business authorities, adapters, tiny kernel, platform, composition and views; exact D7 realization is intentionally unfrozen.

CURRENT PRIMARY EVIDENCE
Both authority models reject cross-owner SQL/private imports, generic service/repository/workflow platforms and mechanism-as-authority. Marketplace's 13 semantic boundaries are more numerous, but that changes the catalog, not the dependency failure class.

CLASSIFICATION = ALIGN

RATIONALE
Marketplace should not rediscover the structural class from zero. The reusable decision is the class law, not MetalDocs owner names or application leaves:

semantic owner/context public boundary
+ stateless application/use-case orchestration
+ external adapters owned by consumer semantics
+ non-semantic platform mechanisms
+ wiring-only composition root
+ mechanically default-deny first-party edges

SMALLEST NEXT ACTION
When D7 opens, begin from this class vocabulary as the default candidate and falsify it against Marketplace provider-ingress, fan-out and tenant requirements. Do not freeze it in D6 and do not copy MetalDocs package names.

REOPEN TRIGGER
A concrete Marketplace interaction requires a legal dependency that cannot be represented without transferring authority or creating repeated adapter/orchestration contortions.
```

### XR-08 — Go HTTP routing

```text
ID
XR-08

PROPERTY / FAILURE CLASS
Dispatch the exact canonical Product paths without path rewrites, shadow routing semantics or a broad framework dependency.

METALDOCS CURRENT DECISION
Go `net/http` is the HTTP reference; MetalDocs' accepted path grammar is compatible with its server profile.

MARKETPLACE CURRENT DECISION
D7 router is open, but D5 already proved canonical partial-segment paths such as `{id}:verb` and explicitly warned that standard `ServeMux` is not a neutral proof vehicle.

CURRENT PRIMARY EVIDENCE
Current Go 1.26 `net/http.ServeMux` documentation still requires wildcards to occupy a complete path segment; `/b_{id}`-style partial-segment wildcard patterns are invalid. Go's own routing guidance says third-party routers remain appropriate for advanced routing needs.

CLASSIFICATION = DIVERGE_JUSTIFIED

RATIONALE
Marketplace cannot copy MetalDocs' pure ServeMux choice while retaining its ratified wire. This is a real path-grammar difference. It does not justify Gin/Echo/Fiber or a broad web framework; the HTTP server/middleware can still be `net/http` with the smallest compatible mux/router or generated dispatch layer.

SMALLEST NEXT ACTION
Marketplace D7, when authorized, should compare a narrow router (for example Chi-class routing) against the already-proved bounded custom/generated mux interface specifically for exact colon-suffix dispatch. Choose the smaller proved option. MetalDocs remains on `net/http`.

REOPEN TRIGGER
Marketplace Product paths are intentionally changed to a ServeMux-compatible grammar by the owning API stage, or a selected narrow router creates an unacceptable correctness/maintenance burden.
```

### XR-09 — PostgreSQL, transaction substrate, SQL generation and tenant isolation

```text
ID
XR-09

PROPERTY / FAILURE CLASS
Preserve local ACID business transitions and owner-private persistence while preventing tenant leaks and avoiding ORM/shared-repository authority.

METALDOCS CURRENT DECISION
One PostgreSQL Product-state DB; `database/sql` transaction-family internal contract with pgx-compatible runtime; pgx/v5 + pgxpool runtime reference; owner-private SQL; no ORM; RLS deferred because Launch has one Company root.

MARKETPLACE CURRENT DECISION
PostgreSQL canonical MPC state; no Direct Oracle fallback; exact D7 transaction/driver/RLS mechanism open; Organization is a real tenant/isolation root and target isolation may not rely only on remembered predicates.

CURRENT PRIMARY EVIDENCE
Current pgx/v5 documentation recommends the native pgx interface when an application targets PostgreSQL only and no dependency requires `database/sql`. Marketplace currently has no accepted D7 consumer requiring `database/sql`. Its tenant invariant is materially stronger than MetalDocs' singleton-Company Launch posture.

CLASSIFICATION = DIVERGE_JUSTIFIED

RATIONALE
Align the database and persistence philosophy, not necessarily the transaction interface. Marketplace should begin with pgx/v5 + pgxpool as the default driver candidate and should not inherit MetalDocs' `database/sql` contract unless a real D7 dependency proves the need. `sqlc` is a strong common SQL-generation candidate because it preserves explicit owner SQL rather than introducing ORM authority.

RLS/isolation is deliberately different: Marketplace D7 owes a fail-closed Organization isolation substrate (PostgreSQL RLS is the leading candidate to prove); MetalDocs correctly deferred pooled RLS under its one-Company Launch interlock.

SMALLEST NEXT ACTION
No D7 work now. Record for D7: compare pgx-native transaction scope first; prove RLS or an equally strong structural tenant mechanism with cross-Organization negative tests; evaluate `sqlc` and `tern/v2` as common tooling candidates; keep ORM absent.

REOPEN TRIGGER
A library/current transaction property requires `database/sql`; RLS cannot express a required owner/system operation safely; or measured persistence/query complexity proves another explicit approach smaller without weakening isolation.
```

### XR-10 — Durable jobs / River

```text
ID
XR-10

PROPERTY / FAILURE CLASS
Persist asynchronous/recoverable work and transactionally couple required follow-up work to committed PostgreSQL state without adding a second infrastructure service or turning queue state into business truth.

METALDOCS CURRENT DECISION
River is ratified for named durable consumers and runs in-process at Launch; River state is mechanism only.

MARKETPLACE CURRENT DECISION
D3 requires recoverable consequential propagation and Marketplace clearly has provider acquisition/effect/reconciliation workloads, but D7 has not selected the durable mechanism.

CURRENT PRIMARY EVIDENCE
Current River documentation continues to provide PostgreSQL-backed typed jobs, transaction-safe enqueueing, retries, recurring jobs, uniqueness and operational tooling without adding Redis/broker infrastructure.

CLASSIFICATION = DEFER

RATIONALE
River is not merely familiarity bias: its core property maps unusually well to Marketplace's Go + PostgreSQL + recoverable-propagation constraints. But D3 has more event/fan-out/provider reconciliation shapes than MetalDocs, so D7 must prove that one job-per-required reaction, scheduling, retry and transaction semantics cover those exact edges without creating a generic event/workflow authority.

SMALLEST NEXT ACTION
When D7 opens, test River first rather than doing an unconstrained queue survey. Reject Redis/Kafka/NATS/custom outbox until a concrete River failure class exists.

REOPEN TRIGGER
D7 proves River cannot satisfy a required atomic-enqueue, fan-out, ordering, isolation, throughput, scheduling, recovery or deployment property at lower total complexity.
```

### XR-11 — Observability, configuration, test and security tooling

```text
ID
XR-11

PROPERTY / FAILURE CLASS
Use standard replaceable technical mechanisms for telemetry/config/proof/security without product authority or duplicated operational stacks.

METALDOCS CURRENT DECISION
OpenTelemetry Go + OTLP, `slog` JSON, `otelhttp`, `otelpgx`; `go-envconfig` reference; sqlc/tern/Testcontainers-Go/govulncheck/OSV-Scanner are preferred adjacent candidates.

MARKETPLACE CURRENT DECISION
Owning D7/later choices remain open; its engineering guide independently prefers standard/native mechanisms, real dependency tests and vulnerability scanning.

CURRENT PRIMARY EVIDENCE
These mechanisms remain current maintained ecosystem choices and match generic rather than domain-specific properties.

CLASSIFICATION = ALIGN

RATIONALE
There is no Marketplace-specific property that benefits from a parallel telemetry/logging/config/test-security ecosystem. Aligning these choices reduces human/LLM operating and upgrade drift without coupling code or releases.

SMALLEST NEXT ACTION
Adopt the MetalDocs mechanism set as Marketplace's default candidate list when the owning stages open. When both repos select the same dependency, prefer the same supported version line and upgrade/security-review cadence unless one repo documents a concrete compatibility constraint.

REOPEN TRIGGER
A repository-specific runtime/provider/deployment requirement is unsupported, or an upstream project becomes unmaintained/insecure/incompatible.
```

### XR-12 — Same-origin profile and deliberate infrastructure absences

```text
ID
XR-12

PROPERTY / FAILURE CLASS
Minimize cross-origin/auth/deployment complexity while avoiding speculative infrastructure that has no current consumer.

METALDOCS CURRENT DECISION
One same-origin app serves SPA + API + OIDC callback; no SSR, service worker/offline correctness, realtime, Redis, external Search, service mesh, generic event bus/workflow or separate BFF service.

MARKETPLACE CURRENT DECISION
React client and stable Product origin exist; D7 serving topology is open; D6 already excludes BFF/screen API, SSR-by-fashion, offline-first, realtime and microfrontends.

CURRENT PRIMARY EVIDENCE
Current IETF browser guidance makes same-origin/server-mediated human auth more attractive, not less. Neither product has an SEO/server-rendering, disconnected/offline, realtime or independently-scaled cache/search consumer.

CLASSIFICATION = ALIGN

RATIONALE
Use one same-origin modular-monolith serving profile as the first Marketplace D7 candidate. XR-05 does not require a new "BFF service": the same application can perform the security BFF/session role for its browser while remaining the resource/application server. Continue rejecting a screen-shaped business BFF.

SMALLEST NEXT ACTION
Keep these absences explicit. Do not add SSR/Redis/realtime/microfrontends/service mesh/external Search/generic event bus/ORM merely to create a common platform.

REOPEN TRIGGER
A measured SEO/rendering, cross-origin independent-client, availability/throughput, offline, low-latency realtime, search, cache or process-isolation requirement creates a real consumer.
```

### XR-13 — OpenAPI dialect, generators and exact version alignment

```text
ID
XR-13

PROPERTY / FAILURE CLASS
Keep one executable wire authority and compatible generated projections without version churn that protects no behavior.

METALDOCS CURRENT DECISION
OpenAPI 3.0.3; `openapi-typescript 7.13.0` and `oapi-codegen v2.8.0` proved as generation mechanisms.

MARKETPLACE CURRENT DECISION
OpenAPI 3.1.2; the same `openapi-typescript 7.13.0` / `oapi-codegen v2.8.0` baseline is already accepted/proved in D5.

CURRENT PRIMARY EVIDENCE
Both generator families remain current and both accepted OADs are executable with them. Marketplace uses 3.1.2 because its accepted schema/tooling package selected that expressiveness. MetalDocs 3.0.3 currently expresses its complete 78-operation wire and has no missing semantic feature.

CLASSIFICATION = DIVERGE_JUSTIFIED

RATIONALE
The valuable alignment already exists at the authority/generator level. Rewriting MetalDocs OAD solely from 3.0.3 to 3.1.x would be compatibility churn with no protected property. Exact dependency versions likewise belong to manifests/upgrade policy rather than durable architecture unless a version-specific behavior is required.

SMALLEST NEXT ACTION
Keep current OAD dialects. Prefer common generator version lines at implementation time and upgrade them together when compatible, but prove each repository's generated/conformance contract independently.

REOPEN TRIGGER
MetalDocs requires a 3.1-only schema property or 3.0.3 tooling becomes unsupported/insecure; Marketplace discovers a generator incompatibility requiring dialect/tool change.
```

### XR-14 — Form/UI component dependencies

```text
ID
XR-14

PROPERTY / FAILURE CLASS
Avoid hand-building complex accessibility/form mechanics while also avoiding a universal design-system/form/schema platform before real component needs exist.

METALDOCS CURRENT DECISION
No broad form/UI/design-system framework is frozen; editor/viewer is a bounded special seam.

MARKETPLACE CURRENT DECISION
D6 explicitly rejects router/form/state/UI libraries by preference and universal design-system platform work.

CURRENT PRIMARY EVIDENCE
Current interaction authority proves forms, filters, dialogs/tables and accessible controls exist, but does not prove one broad form/schema/component framework is required across both products.

CLASSIFICATION = DEFER

RATIONALE
React Hook Form, Zod/Valibot, Radix/shadcn/MUI and Storybook solve different properties. Selecting a bundle now would be technology shopping. Router search validation must remain URL-bound and must not turn a validation library into Product schema authority.

SMALLEST NEXT ACTION
Enumerate concrete repeated form/accessibility primitives during implementation planning and choose the smallest maintained dependency for each proven repeated failure class. Prefer common choices across repos only when the component/form property matches.

REOPEN TRIGGER
Repeated complex form lifecycle, accessibility/focus primitives or component consistency cost is demonstrated in both products.
```

## SHARED PROFILE — safe to align now

```text
Architecture
  independent repositories/applications
  one Product wire SSOT per repo
  React SPA client; no client business authority
  Go modular monolith default architecture class
  PostgreSQL canonical Product state
  external OIDC; provider claims != Product Permissions
  owner/context-private persistence and default-deny dependency direction
  same-origin first-party serving profile as default candidate

Frontend
  React + TypeScript strict
  TanStack Query = server state
  TanStack Router = preferred common router
  openapi-typescript = generated wire shapes
  openapi-fetch = preferred common low-level Product HTTP client, subject to bounded MetalDocs reopen and Marketplace XR-05 repair
  lens/flow-owned UI features + stateless owner/operation-family API adapters
  URL / form draft / ephemeral UI remain separate from server state
  no Redux/Zustand server mirror
  no generated generic workflow/action/query behavior layer

Authentication
  OIDC/provider-neutral architecture
  Keycloak = first concrete self-hosted provider candidate
  ecosystem-standard OIDC libraries rather than Keycloak-specific business adapters
  Product Permission/Principal/Membership remain application-owned

Backend/tooling default candidates
  oapi-codegen generated Go wire boundary
  pgx/v5 + pgxpool PostgreSQL runtime candidate
  sqlc explicit SQL generation candidate
  tern/v2 migration candidate
  OpenTelemetry + OTLP
  log/slog structured logs
  Testcontainers-Go for real dependency proof
  govulncheck + OSV-Scanner
  small typed env config (`go-envconfig` class)

Shared subtraction
  no ORM
  no broad Go web framework by default
  no separate screen-shaped BFF service
  no Redis by default
  no generic event bus/workflow platform
  no microfrontends
  no SSR by fashion
  no offline-first/service worker correctness
  no realtime channel without consumer
  no service mesh
  no shared platform/code repository merely for symmetry

Version policy
  when both repos independently select the same dependency, prefer one supported version line and coordinated review cadence
  no shared lockfile/release/deployment coupling
```

## REPO-SPECIFIC DIFFERENCES — must remain different

```text
Tenant isolation
  MetalDocs: one-Company Launch interlock; pooled RLS legitimately deferred.
  Marketplace: real Organization tenant root; D7 owes fail-closed isolation and should prove RLS or equivalent.

HTTP routing
  MetalDocs: accepted path grammar fits net/http reference profile.
  Marketplace: canonical `{id}:verb` partial-segment paths do not fit Go ServeMux wildcard grammar; needs a narrow compatible mux/router in D7.

Transaction API
  MetalDocs: database/sql transaction-family contract is already ratified.
  Marketplace: should start pgx-native unless a real D7 dependency requires database/sql.

Domain/package census
  MetalDocs: five Launch semantic homes and document-specific application lenses.
  Marketplace: many marketplace/business authorities and provider adapters. Copy the class law, never the owner/package names.

Wire dialect
  MetalDocs: OAS 3.0.3 remains sufficient.
  Marketplace: OAS 3.1.2 already proved/accepted.

Content-specific runtime
  MetalDocs renderer/scanner/managed-content/exact-byte mechanisms are Product-specific and must not be copied into Marketplace absent a consumer.
```

## DEFERRED PROFILE — align later only after owning stage proves need

```text
Marketplace D7 only after authorization
  exact HTTP mux/router (narrow candidate, not broad framework)
  exact pgx transaction/RLS realization
  River selection after D3 edge-by-edge durable-work proof
  same-process worker/scheduler topology
  Keycloak realm/deployment/HA/backups
  same-origin deployment/reverse proxy details
  exact secrets/config mechanics

Both implementations
  form library
  schema validation library for UI-only/form concerns
  UI primitive/component library
  Storybook/design-system tooling
  exact Vite/build/package-manager choices
  exact dependency pins
  shared physical Keycloak deployment

Do not defer the Marketplace human browser-auth question: it changes D6 transport/security assumptions and must be adjudicated before D6-B2 closes.
```

## METALDOCS REOPEN DECISION

```text
REOPEN RECOMMENDED — BOUNDED T8-F MECHANISM ONLY

Subject:
  browser Product transport implementation

Current:
  openapi-typescript + hand-authored thin native fetch

Candidate:
  openapi-typescript + openapi-fetch-backed thin transport

Reason:
  current upstream now supplies OAD-bound path/param/body/response serialization/type safety while preserving native Response/headers/streams and without adding query/business authority.

Not reopened:
  Product/T1→T7 semantics
  78-operation census
  T8-E wire
  state classes
  route meanings
  Authorization/session/CSRF semantics
  backend/persistence/runtime topology
  T9

Required proof before any authority mutation:
  representative safe JSON read
  conditional ETag mutation
  idempotent POST + CSRF
  exact-byte/stream response + Content headers
  Problem Details decoding
  no unsafe casts / no hidden retry / no generated query authority

If that proof fails, retain native fetch and classify transport divergence as justified.
```

## MARKETPLACE REOPEN DECISION

```text
REOPEN REQUIRED — TARGETED D5 HUMAN AUTHENTICATION/WIRE PROFILE

Owning scope:
  D5-B2-A client/authentication admission
  W4/OAD security carrier only as required for consistency

Material finding:
  direct browser bearer OAuth is no longer the strongest baseline for this first-party business/PII SPA under current IETF browser-app security guidance.

Target to adjudicate:
  H browser -> same-origin HttpOnly ApplicationSession + CSRF; server-side OIDC code exchange
  A/S machine -> audience-bound OAuth bearer / Client Credentials

Preserve:
  Principal kinds H/A/S
  D2 Principal binding and eligibility
  Organization Membership
  ordinary Permissions
  business disposition/Governance
  operation inventory 99/30 unless the auth repair itself proves otherwise
  Technical Ingress separation

D6-B2 strict owner-first frontend folders are rejected as a candidate, but this is NOT an accepted-authority reopen: D6-B2 has not ratified a topology yet.

D7 remains BLOCKED.
```

## CONTINUATION RECOMMENDATION

```text
1. Marketplace Lead independently adjudicates XR-01→XR-14 against current repository authority and the current upstream sources named above.

2. Resolve XR-05 first.
   If accepted, perform only the bounded Marketplace D5 authentication/wire repair and re-prove the canonical OAD/client-class surface.
   Do not begin D7.

3. Then close Marketplace D6-B2 around the smallest common frontend profile:
   React + TypeScript
   TanStack Query
   TanStack Router
   openapi-typescript
   openapi-fetch thin transport
   lens/flow features + stateless owner/operation API adapters
   no second server store / no generic query-workflow layer
   form/UI framework still unfrozen unless concrete evidence requires it.

4. Separately route XR-02 to MetalDocs operator adjudication.
   If the bounded openapi-fetch proof is approved and passes, reopen T8-F mechanism wording only and perform the smallest Whole-T8 coherence revalidation required by repository governance.
   Do NOT open T9 as part of that work.

5. Preserve the explicit divergences:
   Marketplace tenant isolation/RLS obligation
   Marketplace partial-segment HTTP routing requirement
   Marketplace pgx-native transaction candidate
   Product-specific MetalDocs content-processing runtime
   current OAS dialect difference.

6. Treat River, OTel/config/SQL/migration/test/security mechanisms as pre-vetted Marketplace D7 candidates, not D6 decisions.

7. After reciprocal adjudication reaches no surviving material contradiction, write the bounded operator-facing alignment result into each owning repo only through its normal stage/reopen process; then delete this Evidence file and close PR #153 and Marketplace PR #57 unmerged.

T9 remains NOT OPEN.
Marketplace D7 remains BLOCKED.
Product implementation remains BLOCKED in both repositories.
```
