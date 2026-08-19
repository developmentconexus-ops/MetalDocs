# R10-T8A — Technical Census & Preliminary Disposition Matrix

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **PRELIMINARY CENSUS; T8-A NOT RATIFIED**  
> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Current product-source baseline:** `main@7f5b8928cc5a13feb8ee3fa7c8ceb1c7d3655a18` — the redesign PR changed documentation/routing/support artifacts, not product implementation paths  
> **Implementation:** BLOCKED

This is the first evidence-backed T8-A disposition matrix. It is not target architecture and does not authorize T8-B→T8-G decisions. Its purpose is to separate **valuable properties** from **legacy physical shapes** before physical target design.

## 1. Binding posture

DevelopmentConexus Method + T8-A bootstrap require:

```text
existing code/schema/API/tests/runtime/history = evidence
existing shape                               != target authority
sunk cost / test count / migration ease      != survival criterion
PRESERVE                                     must be proved
REWRITE / REHOME / DELETE                    are valid Global Maximum outcomes
```

Operator explicitly reconfirmed on 2026-08-19 that T8-A must not protect any implemented local maximum and may refactor/rewrite from zero whenever Method + accepted decisions + evidence justify it.

T7 also removed a major compatibility constraint:

```text
current MetalDocs business data/history = DEV / TEST / THROWAWAY
historical-data compatibility consumer   = NONE
```

Therefore no package/table/route/process survives merely to preserve current DEV data.

## 2. Structural Inversion result

If the current implementation had the opposite package/table/API/frontend/process shape, the following accepted target properties would still be true:

```text
PostgreSQL product-state transactions
Document / Revision / WorkingContent / Submission separation
live T3 Authorization + domain predicates
same-local-commit Audit for required semantic/security changes
exact-content identity independent from storage identity
provider-neutral managed content
River-class durable jobs where T5 has a named consumer
contract-first executable API + generated boundaries
frontend as client, not lifecycle/AuthZ authority
fail-closed authentication/session behavior
restart/readiness/recovery/restore proof
repository verification must be mechanically reproducible locally and in CI
```

Current physical topology is therefore not load-bearing by itself.

## 3. Evidence / disposition matrix — wave 1

| Surface / structure | Current evidence | Ratified target property | Evidence class | Preliminary disposition | Reason / boundary | Later owner |
|---|---|---|---|---|---|---|
| **One Go module `metaldocs`** | `go.mod`; Go + frontend + SQL + Node renderer repo | no semantic requirement for multiple Go modules | CURRENT-PROVEN | **PRESERVE candidate as repository mechanism only** | One Go module is simple and does not itself violate 4+1 ownership. T8-B may keep it unless package isolation evidence disproves sustainability. | T8-B |
| **Current ~15 semantic module identities under `internal/modules/*`** | current imports/search expose `approval`, `auth`, `audit`, `controlleddocuments`, `distribution`, `documents`, `iam`, `jobs`, `notifications`, `render`, `search`, `security`, `taxonomy`, `templates`, `tokens` | 4 business owners + Audit supporting; mechanisms are not owners | CURRENT-PROVEN shape / old exact counts not relied on | **REWRITE / REHOME** | Legacy module vocabulary conflicts directly with ratified semantic ownership and contains Launch+/Future mechanisms as domains. Package count will be rederived, not mapped 1:1. | T8-B/C |
| **API composition root** | `apps/api/cmd/metaldocs-api/main.go` imports many legacy modules + platform packages and contains cross-domain adapters | explicit composition root; owners communicate through accepted contracts | CURRENT-PROVEN | **REFINE / likely REWRITE wiring** | Property “composition root owns adapters” is healthy; current root is coupled to superseded module vocabulary and legacy capabilities. | T8-B/C/G |
| **Worker host** | `apps/worker/cmd/metaldocs-worker/main.go` composes legacy outbox/render/documents/approval/IAM concerns | only named durable/external effects; smallest runtime topology | CURRENT-PROVEN | **REWRITE / possibly DELETE as separate process** | Useful shutdown/readiness/OTel properties exist, but process existence is not a target requirement. T8-G must derive whether a separate worker is still justified. | T8-G |
| **Jobs host** | `apps/jobs/cmd/metaldocs-jobs/main.go` registers River release work plus notifications, Periodic Review surfacer, approval SLA, tenant lifecycle, retention/watchdog jobs | T5 River for named jobs; Distribution/Periodic Review Launch+; SLA/tenant lifecycle not Launch baseline | CURRENT-PROVEN | **REWRITE** | Host mixes justified durable-job mechanism with capabilities explicitly removed/deferred from Launch. Preserve River property, not current worker registry/process shape. | T8-C/G |
| **Platform → legacy module dependency** | `internal/platform/authn/context.go` imports `internal/modules/iam/domain` | Authentication distinct from Organization; dependency direction must follow accepted ownership | CURRENT-PROVEN | **REHOME / REWRITE** | Fail-closed actor extraction is valuable, but `platform/authn` depending on legacy IAM domain is a concrete inversion against target ownership. Exact replacement belongs later. | T8-B/C |
| **Current module-boundary PowerShell guard** | `scripts/check-module-boundaries.ps1` scans only `internal/modules/**`; allows legacy module `domain/application/api` published surfaces; debt allow-list empty | mechanical enforcement of accepted boundaries | CURRENT-PROVEN | **REFINE / REWRITE policy** | Firing guard is valuable, but its subject is the legacy module topology and it misses platform→module inversion. New boundary guard must protect T8 target, not ADR-0082-era module identities. | T8-B/C + T9 |
| **PostgreSQL as product-state SSOT** | DB overview + current runtime; T2/T5 rely on local ACID + River substrate | local ACID semantic state and durable job coherence | CURRENT-PROVEN + RATIFIED PROPERTY | **PRESERVE** | This is both current evidence and accepted architecture property. Physical schemas/tables are not preserved by this decision. | T8-D/G |
| **Current curated baseline process** | `db/baseline/0001_current_schema.sql`; fresh environments bootstrap prerequisites → schema → reference data → grants; equivalence script historically proves folded baseline | reproducible schema/bootstrap, proof before serving | CURRENT-PROVEN mechanism | **PRESERVE / REFINE property** | Curated deterministic bootstrap is valuable; current 182 KB schema contents are legacy. T8-D/T10 can rebuild baseline from scratch while keeping deterministic bootstrap/proof. | T8-D/T10 |
| **`metaldocs` + legacy `public` schema split** | `wiki/database/schemas.md`; public contains historical unqualified objects | target persistent ownership must reflect accepted owners/mechanisms | CURRENT-PROVEN | **REWRITE** | Existing schema split is historical accident, not semantic ownership. T8-D must derive target schemas/tables without migration entitlement. | T8-D |
| **Tenant/RLS mesh** | baseline enables `tenant_isolation` policies across IAM, sessions, approval, documents, templates, notifications, outboxes, tenant lifecycle and other tables; `metaldocs.tenant_id` GUC | Launch single-company with stable Company identity; T3 live Authorization + scope/domain predicates | CURRENT-PROVEN | **REWRITE / DELETE current mechanism** | Pooled multi-tenant isolation is not a Launch consumer. Do not delete security properties blindly: T8-D must decide the smallest DB constraints/defense-in-depth that enforce current target invariants without legacy tenant GUC/RLS complexity. | T8-D/G |
| **DB capability-assertion GUC/triggers** | baseline `enforce_capability_asserted()` uses `metaldocs.asserted_caps` / bypass GUC and legacy capability names including template/controlled-document/review semantics | T3 product Role/Permission + domain predicates; DB constraints where correctness materially benefits | CURRENT-PROVEN | **REWRITE / adjudicate necessity** | Current mechanism encodes superseded permission vocabulary and legacy subjects. Property “writes fail closed if security wiring is bypassed” may be valuable, but exact trigger/GUC scheme has no survival entitlement. | T8-C/D |
| **Current persistent table families** | baseline includes parallel `documents`, `controlled_documents`, `templates_*`, approval kernel, editor sessions, document comments/exports/placeholders, tenant lifecycle, notifications etc. | T1 core Document/Revision/WorkingContent/Submission/Governance/Release + 4+1 ownership; Launch+Future deferred | CURRENT-PROVEN | **REWRITE / many DELETE candidates** | Persistent meaning conflicts directly with ratified semantic target. T7 says data is disposable, so target schema can be rederived cleanly. | T8-D/T10 |
| **Current OpenAPI 3.0.3 contract** | `api/openapi/v1/openapi.yaml` | T6 OpenAPI 3.0.3 `/api/v1`, generated Go/TS, explicit errors/headers/idempotency | CURRENT-PROVEN | **REWRITE document; PRESERVE contract-first property** | Existing contract contains materially superseded routes/semantics. Rebuild `/api/v1` in place from T6; no compatibility layer. | T8-E |
| **Current auth wire/handler** | OpenAPI + `internal/modules/auth/delivery/http/handler.go`: local identifier/password login, password change, local session service | Keycloak Authorization Code + state/PKCE → ApplicationSession; no local passwords/ROPC/JIT; auth login/callback outside generated JSON API | CURRENT-PROVEN | **REWRITE / DELETE local-credential capability** | Direct contradiction with T6. Preserve only useful fail-closed/cookie/session/error properties that still apply after OIDC realization. | T8-E/G |
| **Generated frontend API types/client direction** | frontend `gen:api` runs `openapi-typescript`; frontend structure documents generated `api-types` + `openapi-fetch` path | generated TypeScript transport boundary; frontend consumes canonical API | CURRENT-PROVEN property; exact current docs stale | **PRESERVE / REFINE** | Good mechanism aligned with T6. Regenerate against rewritten T8-E contract; legacy wrappers should not survive by compatibility. | T8-E/F |
| **Frontend React Router + single TanStack Query provider** | `App.tsx`, `RootProviders.tsx`, `AppRouter.tsx`; React Router + QueryClientProvider | frontend semantic lenses; TanStack Query ownership to be frozen in T8-F | CURRENT-PROVEN | **PRESERVE candidate mechanism** | These mechanisms do not conflict with semantic target and are replaceable implementation choices. Preserve only if T8-F comparison confirms smallest sustainable solution. | T8-F |
| **Frontend feature topology** | route tree composes approval/dashboard/documents/IAM/password-change/taxonomy/templates/tokens; feature tree also has controlled-documents/notifications etc. | Library, My Work, Document Official/Work/History, Governance Case, Audit, Administration | CURRENT-PROVEN | **REWRITE / REHOME** | Current folders/routes mirror legacy domains rather than ratified user lenses. Do not refactor incrementally from names as target assumptions. | T8-F |
| **Current editor dependency stack** | web depends on `@metaldocs/editor-ui`, TipTap, DOCX/ZIP tooling, PDF.js; repo also carries EigenPal wrapper/ACL/reference artifacts | T6: one interactive DOCX provider; EigenPal/browser-buffer first candidate; provider replaceable; fidelity corpus decides | CURRENT-PROVEN inventory; target provider still proof-dependent | **CURRENT-STATE ONLY / REMEASURE fidelity** | Installed/editor code does not prove provider choice. Preserve adapter seam and proof requirement; T8-F/T9 must use current fidelity evidence, not sunk cost. | T8-F/T9 |
| **Compose: db-provision separate from runtime DB role** | compose has one-shot `db-provision`; API/worker/jobs use `metaldocs_runtime` rather than bootstrap superuser | least privilege; serving runtime should not own schema provisioning | CURRENT-PROVEN | **PRESERVE property / REFINE topology** | Strong security/operations property independent from old domain model. Exact service/process can change in T8-G. | T8-G/T10 |
| **Compose: MinIO/local object store** | MinIO + deterministic bucket/bootstrap | T4 provider-neutral store; Local dev/test + S3 reference production profile | CURRENT-PROVEN | **CURRENT-STATE ONLY / preserve store abstraction property** | MinIO is a current provider, not target authority. Provider-neutral contract is ratified; T8-G chooses deployment profiles. | T8-G |
| **Compose: Gotenberg + separate DOCX renderer** | current Gotenberg + Node `docx-renderer` process | T5 renderer replaceable; official rendition exact Submission; fidelity corpus required | CURRENT-PROVEN | **CURRENT-STATE ONLY / REEVALUATE** | Provider/process count is not ratified. Reuse only if fidelity + operational comparison makes it Global Maximum. | T8-F/G/T9 |
| **Redis / multi-replica rate-limit substrate** | compose config uses Redis for shared rate limiting and `METALDOCS_MULTI_REPLICA` | no T1→T7 semantic owner/property requires Redis or multi-replica topology by itself | CURRENT-PROVEN | **CURRENT-STATE ONLY / DELETE candidate** | Keep only if T8-G names a Launch runtime/security consumer that cannot be satisfied more simply. | T8-G |
| **`tools/verify` single verification registry** | `tools/verify/main.go`, `registry.go`, `ci.yml`: local and CI share one registry; PASS/FAIL/SKIP explicit; toolchain preflight; audit; ordering; diff scoping; negative fixtures/closed waivers | proof before implementation; no inert controls; reproducible verification | CURRENT-PROVEN | **PRESERVE** | This is a demonstrated solution to a real repo failure class and is topology-agnostic at its core. | T9/T11/T12 |
| **Current architecture-specific verifier checks** | registry/guards include module-imports, arch-lint analyzers, legacy vocabulary/ownership assumptions | mechanically enforce accepted target architecture, not historical shape | CURRENT-PROVEN | **REFINE / REWRITE with T8** | Existing controls are valuable only insofar as their policy matches new target. New T8 boundaries must replace old identities and fixtures must prove they fire. | T8-B→G / T9 |
| **Current technical docs (`data-model.md`, `backend-blueprint.md`, `frontend-structure.md`, repo/module pages)** | TRRB already marks them current-state/legacy evidence; `data-model.md` explicitly says target not designed | one authority per meaning; no stale target routing | CURRENT-PROVEN classification | **CURRENT-STATE ONLY → REWRITE/DELETE after replacement** | Useful archaeology, but several still contain stale routing/“canonical” language. T8-A must prevent Fresh Actors from treating them as target authority. | T8-A/T10 |
| **Historical target docs (`cohesive-platform-redesign.md`, `backend-target-architecture.md`)** | router/TRRB mark superseded/historical | current R10 authority only | CURRENT-PROVEN classification | **SUPERSEDED** | No target inheritance. Keep only where documentation governance intentionally retains history; never route target design through them. | T8-A/docs governance |

## 4. High-confidence findings already available for adjudication

### F1 — Current semantic module topology is not a target candidate by default

Current package/module names encode concepts that R10 explicitly removed, rehomed or deferred. Examples from live source include:

```text
approval/http/delegation_handler.go
approval/http/extend_sla_handler.go
approval/http/fast_forward_handler.go
jobs/document_review_surfacer
jobs/approval_sla_surfacer
iam/jobs/tenant_lifecycle_worker.go
distribution/delivery/http/routes.go
notifications/delivery/http/routes.go
templates/* independent lifecycle routes
```

This is not merely naming drift. It is capability/authority drift. T8-B must derive topology from 4+1 + T1→T7, not rename the old 15 modules.

### F2 — Current persistence contains structural accidental complexity from abandoned product assumptions

The baseline still represents:

```text
pooled tenant isolation substrate
parallel Documents / ControlledDocuments / Templates identities
legacy approval kernel variants
comments / placeholders / exports / editor-session state
notifications / distribution-adjacent state
tenant lifecycle / crypto-shred machinery
periodic review / approval SLA machinery
```

Some constraints/transaction techniques may be reusable evidence. The table families themselves have no target entitlement.

### F3 — Security must be decomposed into invariant vs mechanism

Do not conclude either:

```text
"RLS exists, therefore preserve RLS"
OR
"single-company, therefore delete every DB-level security guard"
```

Instead T8-D must ask which target invariants need DB-level enforcement and choose the smallest mechanism. Current tenant RLS is not the target security model; T3 live Authorization remains semantic authority.

### F4 — Current API and frontend are replacement evidence, not compatibility constraints

Pre-launch T6 explicitly authorizes rebuilding `/api/v1` in place with no `/api/v2` and no compatibility. Current OpenAPI/frontend therefore serve as:

```text
mechanism evidence
failure/UX evidence
transition/deletion inventory
```

not contract compatibility requirements.

### F5 — Verification control plane is one of the strongest preserve candidates

The current `tools/verify` registry has a named root cause, fail-closed design, CI/local convergence and negative-fixture discipline. It should not be thrown away merely because domain architecture changes. The target-specific checks inside it must evolve as T8 decisions land.

## 5. Remeasure / unknown register

Exact old Aug-09 metrics remain `LAST-REPRODUCED`. Remeasure only when they change a material decision.

| Evidence | Current need | T8-A action |
|---|---|---|
| exact Go package/module/SCC/reciprocal-edge counts | useful to quantify transition blast radius, not to choose target owner model | remeasure before final T8-A disposition if a transition/topology choice relies on exact magnitude |
| exact foreign SQL read/write count | material to prove current ownership leakage and T10 deletion effort | **REMEASURE before T8-A closure** |
| frontend cross-feature import/cache-key counts | useful to quantify current coupling | remeasure if needed for T8-F transition classification; exact count not yet needed for target lens decision |
| current OpenAPI operation/tag count | useful to size deletion/rewrite surface | remeasure before T8-E/T10 planning; not needed to prove current contract is semantically superseded |
| renderer fidelity against current representative DOCX corpus | load-bearing for EigenPal/ONLYOFFICE/Gotenberg/LibreOffice selection | **must be fresh before T8-F/G target provider decision** |
| current process resource/performance envelope | only material if runtime topology comparison depends on it | measure in T8-G when alternatives are concrete |
| current Redis/multi-replica need | no named Launch consumer yet | T8-G must prove or delete |

## 6. Current T8-A direction — not yet ratified

```text
PRESERVE proven properties, not legacy names.
REWRITE semantic/persistent/wire/frontend shapes that contradict T1→T7.
REHOME cross-owner/platform responsibilities to the owners/mechanisms derived later.
DELETE dormant Launch+/Future capability implementation unless a current Launch property consumes it.
Keep current-state evidence only where it helps T10 delete/transition safely.
```

The strongest current hypothesis is **substantial clean-slate physical realization inside the existing repository**, not incremental preservation of the legacy module/data/API/frontend model. This remains a hypothesis until the remaining census, foreign-SQL remeasurement, document-authority reconciliation and adversarial T8-A candidate review close.

## 7. Next evidence wave

```text
1. fresh foreign-SQL / persistent-ownership remeasurement
2. current API/codegen/conformance mechanism census
3. current frontend cross-feature/query/cache/editor boundary census
4. runtime/process/deployment mechanism inventory by named target property
5. technical-document/ADR disposition sweep
6. preliminary matrix challenge: where does PRESERVE still lack proof?
7. T8-A disposition candidate
8. material adjudication + operator ratification
```

T8-B remains closed. Product implementation remains **BLOCKED**.
