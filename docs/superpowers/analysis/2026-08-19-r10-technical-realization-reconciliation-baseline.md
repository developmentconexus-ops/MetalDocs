# R10 — Technical Realization Reconciliation Baseline

> **Status:** ACTIVE STAGING / CURRENT-STATE CENSUS — **OPERATOR REVIEW NEXT**  
> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Program authority:** `wiki/architecture/r10-post-t6-implementation-readiness-program.md`  
> **Implementation:** BLOCKED

This baseline exists to prevent semantic architecture from being translated into code through legacy defaults or undocumented implementation choices.

It is **not target design**. It is a current-state/evidence census plus a decision-gap routing map. T8 will decide physical realization only after this census is operator-reviewed.

---

## 1. Evidence classification

Every claim in this baseline uses one of four classes:

```text
CURRENT-PROVEN
  directly verified against the current PR branch/current durable authority

LAST-REPRODUCED
  mechanically reproduced on a prior pinned runtime baseline;
  direction/risk remains useful but exact count must be rerun before becoming current

STALE / SUPERSEDED
  contradicted by current authority or explicitly retired as target authority

UNKNOWN / REMEASURE
  not currently proven strongly enough for a target decision
```

Important baseline distinction:

```text
old architecture audit runtime baseline = main@418070bf38a9f358f9131bcc36b7a6bcbc069273
current PR base                        = main@7f5b8928cc5a13feb8ee3fa7c8ceb1c7d3655a18
current pre-census PR head             = 817f0ece16e317d12bfa3fa34fc0f3c1314a031b
```

The current PR changes architecture/governance/documentation surfaces, not product runtime source. However, `main` advanced between the Aug-09 audit baseline and the current PR base. Therefore old mechanically calculated counts are intentionally `LAST-REPRODUCED`, not silently promoted to current facts.

---

## 2. A — Repository and runtime topology

### CURRENT-PROVEN

- Repository is one Go module: `module metaldocs`, currently declaring Go `1.26.5`.
- Durable/current-state topology documents and current source paths identify a mixed-language monorepo with Go backend, React/TypeScript frontend, SQL database artifacts and a Node/TypeScript document-rendering service.
- Current executable/runtime evidence includes API, worker, jobs and DOCX-rendering process roots.
- Current infrastructure evidence includes PostgreSQL, object storage and River-backed job mechanics; exact target process count is not inherited.
- `tools/verify/registry.go` explicitly declares itself the repository verification SSOT and defines `fast`, `changed`, `pr`, `full`, `release` profiles.
- Current `.github/workflows/ci.yml` delegates PR correctness to `tools/verify`, including diff-scoped verification and sharded integration tests.

### LAST-REPRODUCED

Aug-09 audit measured:

```text
158 first-party Go packages
15 internal/modules directories
37 internal/platform packages
3 primary Go runtime binaries + E2E tool
one Node DOCX renderer service
```

These counts are T8-A remeasurement inputs, not target commitments.

### TARGET GAP

T8-G must decide the target process/binary/service/deployment topology. T8-B must decide target package/module topology.

---

## 3. B — Backend modules, package graph and internal boundaries

### CURRENT-PROVEN

Current technical documentation/source topology still carries legacy implementation concepts such as:

```text
approval
audit
auth
controlleddocuments
distribution
documents
iam
jobs
notifications
render
search
security
taxonomy
templates
tokens
```

T1→T6 do **not** ratify these as target semantic owners. Ratified semantic ownership is:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit (supporting evidence)
```

Therefore the current module tree is evidence only.

### LAST-REPRODUCED

Aug-09 mechanical dependency audit found:

```text
one module-level SCC of size 9
7 reciprocal module pairs
4 platform packages importing modules / 9 package edges
zero multi-node package-level Go SCCs
```

The size-9 SCC was:

```text
approval
auth
controlleddocuments
documents
iam
render
security
taxonomy
templates
```

It also identified `jobs` as composition-shaped/orchestration-heavy rather than automatically a business context.

Exact graph must be rerun at T8-A.

### TARGET GAP

T8-B/T8-C must freeze:

```text
final package topology
semantic-owner realization boundaries
allowed/forbidden import graph
consumer/producer public seams
composition root
mechanism locations
internal Q/capability/projection/job contracts
```

No Writer may infer this from current module names.

---

## 4. C — Persistence / database ownership

### CURRENT-PROVEN

`wiki/architecture/data-model.md` explicitly states:

```text
CURRENT DATABASE EVIDENCE ONLY — target model not designed yet
```

The current Postgres/baseline/migration/dictionary tree describes runtime storage, not target architecture.

T1→T6 already freeze material semantic persistence requirements such as:

```text
Document != Revision != WorkingContent != Submission
one DRAFT OCC generation
immutable Submission
system Release/effectivity
Organization/AuthZ current truth
Audit evidence
T4 exact-content descriptor/managed handle separation
River as selected/reference durable-job mechanism
idempotency replay semantics
```

But they do not define the complete physical relational model.

### LAST-REPRODUCED

Aug-09 DB ownership audit measured:

```text
55 foreign-table reads
12 foreign-table writes
67 foreign SQL statements total
10 approval→documents raw UPDATE sites
6 deliberate producer-owned projection views as positive examples
```

It also recorded direct transaction/query-maintenance sprawl and ownership-dictionary drift. Exact counts require T8-A remeasurement.

### TARGET GAP

T8-D must decide target schemas/tables, ownership, correctness constraints, transaction/serialization/locking realization and persistent technical state. T10 must decide current→target schema/data transition.

---

## 5. D — API / OpenAPI / code generation / transport

### CURRENT-PROVEN

- Current application contract source exists at `api/openapi/v1/openapi.yaml`.
- Current Go codebase uses generated OpenAPI packages and frontend generated API types.
- Current `backend-api-structure.md` describes the **legacy/current implementation pattern** of one generated package/tag per implementation module. That pattern is not target entitlement because T6 changed semantic topology.
- T6 already ratifies `/api/v1` semantic operation families, RFC 9457 machine-code law, cursor pagination, ETags/If-Match, bounded Idempotency-Key use, upload/admission semantics and generated Go/TypeScript boundaries.
- Exact executable request/response schemas are not yet frozen as the new target OpenAPI.

### LAST-REPRODUCED / STALENESS CAUTION

Aug-09 API audit measured multiple runtime inconsistencies, including generated-package count, pagination/idempotency variants and partial contract enforcement. Some of those findings were remediated later on `main`; for example historical target evidence records later ingress request-validation work.

Therefore individual old defect counts are **not current claims**.

### TARGET GAP

T8-E must author/freeze the executable wire contract and prove runtime/codegen/frontend conformance. No implementation task may invent wire fields, enums, errors or headers.

---

## 6. E — Frontend realization

### CURRENT-PROVEN

Current `wiki/architecture/frontend-structure.md` still describes a feature-sliced SPA organized around legacy feature names including:

```text
documents
templates
taxonomy
iam
approval
controlled-documents
notifications
tokens
```

It also documents useful mechanisms worth re-evaluating rather than inheriting blindly:

```text
React SPA
TanStack Query
OpenAPI-generated TypeScript types
feature-local API/query layers
local UI state separate from server state
shared UI primitives
```

T6 ratifies the target semantic lenses:

```text
Library
My Work
Document Official
Document Work
Governance Case
Document History
Audit
Administration
```

and explicitly removes separate Approval/Template product authority.

### LAST-REPRODUCED

Aug-09 frontend audit measured:

```text
14 feature directories
29 dead production files
multiple cross-feature imports/allowlist edges
no strongly enforced public feature-surface convention
strong generated API-client adoption
strong TanStack Query vs local-state separation
```

Exact counts require T8-A remeasurement.

### TARGET GAP

T8-F must decide target route tree, frontend feature/package topology, public feature surfaces, generated contract consumption, query/cache behavior, editor/viewer boundaries and explicit legacy feature disposition.

---

## 7. F — Async / rendering / external mechanisms

### CURRENT-PROVEN

T5 ratifies:

```text
one PostgreSQL-backed durable-job mechanism
River selected/reference
at-least-once + idempotent + revalidating workers
OfficialRendition render conditional on representation policy
no generic notification/event-bus baseline
no materialized Search baseline
```

T4 forbids provider calls inside semantic transactions and owns exact-content admission/integrity.

T6 keeps interactive DOCX editing separate from the server-side OfficialRendition renderer.

Current worker/jobs/render processes are current-state mechanisms only.

### LAST-REPRODUCED

Aug-09 async/runtime audit contained both positive and negative evidence:

Positive:

```text
single transactional-outbox insert path
fail-loud consumer dispatch
producer-side duplicate protection
bounded retry/dead-letter mechanics
```

Prior gaps included process observability/health and async trace/alert closure. Some may have changed since the audit baseline.

### TARGET GAP

T8-C/T8-G must decide exact internal async contracts and runtime ownership. T9 must define composed failure/restart proofs.

---

## 8. G — Deployment, operations and recovery

### CURRENT-PROVEN

Current repository contains deploy/compose/docker/runtime scripts and operational verification mechanisms. They prove how the existing product is operated, not what the target must preserve structurally.

T4 already freezes restore correctness properties, including:

```text
restored ApplicationSessions invalid
post-snapshot security teardown reconciled before ordinary serving
fail closed when security reconciliation cannot be proven
DB + durable-job semantic facts restore coherently while they share PostgreSQL
```

### UNKNOWN / TARGET NOT DECIDED

Still missing target decisions include:

```text
final process/service count
startup ownership/order
readiness/liveness topology
network/trust segmentation
configuration/secrets layout
timeout/degradation policy realization
observability placement
production deployment topology
backup/restore runtime choreography
```

These belong T8-G and T9. Concrete transition/cutover belongs T10.

---

## 9. H — Tests / CI / verifier

### CURRENT-PROVEN

Current repository has a substantially consolidated verification spine:

```text
tools/verify/registry.go = verification SSOT
profiles = fast / changed / pr / full / release
verify --audit = registry/workflow anti-drift mechanism
ci.yml = PR gate composition
integration suite = sharded in CI
```

`AGENTS.md` currently instructs normal local PR verification through:

```text
go run ./tools/verify --profile=pr
```

Existing controls remain binding safety rails until a later accepted stage deliberately changes them.

### LAST-REPRODUCED / REMEASURE

Aug-09 architecture guard audit identified remaining classes such as:

```text
no module-SCC guard at that baseline
SQL ownership analyzer write blind spot
no foreign-sentinel independence guard
no consumer-owned-port shape guard
some allowlist/baseline bypass surfaces
```

Current exact guard coverage must be re-audited in T8-A/T9 because the verifier evolved materially during Aug-09.

### TARGET GAP

T9 owns the target Validation Baseline/proof matrix. T12 owns implementation-readiness proof completeness. Neither may weaken currently firing safety controls without a ratified replacement property.

---

## 10. I — Technical documentation authority reconciliation

### CURRENT-PROVEN conflict/drift

`AGENTS.md` currently routes target architecture through `wiki/architecture/cohesive-platform-redesign.md`, but that page is explicitly:

```text
SUPERSEDED FOR ACTIVE TARGET ROUTING
```

`wiki/architecture/system-map.md` also still announces the old Cohesive Platform Redesign as active.

Other technical documents describe real current-state mechanisms but are unsafe as target authority after T1→T6:

| Document | Current role for R10 |
|---|---|
| `backend-target-architecture.md` | **SUPERSEDED / HISTORICAL prior target** |
| `data-model.md` | **CURRENT-STATE DATABASE EVIDENCE** |
| `backend-blueprint.md` | **CURRENT-STATE composition/evidence** |
| `backend-api-structure.md` | **CURRENT-STATE/legacy API realization evidence** |
| `frontend-structure.md` | **CURRENT-STATE/legacy frontend realization evidence** |
| `wiki/backend/repo-topology.md` | **CURRENT-STATE runtime/repository evidence** |
| current module pages | **CURRENT-STATE implementation evidence unless explicitly re-promoted** |

### Required routing correction now

Fresh Actors must route:

```text
AGENTS.md
→ Method
→ current handoff/router
→ Product Contract + T1→T6 durable authority
→ r10-post-t6-implementation-readiness-program.md
→ current active stage/staging
→ legacy technical docs/code only as evidence
```

T8-A will later decide which technical documents are rewritten/promoted/deleted after physical realization is accepted.

---

## 11. J — Target decision gap matrix

| Missing decision class | Evidence state | Owning future stage |
|---|---|---|
| truthful historical source semantics/modes/provenance | not yet designed | **T7** |
| legacy technical disposition census | current evidence + remeasure required | **T8-A** |
| backend package/module topology | not target-designed | **T8-B** |
| internal owner communication realization | not target-designed | **T8-C** |
| target relational schema/constraints/locks | not target-designed | **T8-D** |
| exact OpenAPI wire schemas/codegen conformance | semantic surface only frozen | **T8-E** |
| target frontend route/feature/query/cache topology | semantic lenses only frozen | **T8-F** |
| target processes/jobs/deploy/trust/observability | partially constrained, not realized | **T8-G** |
| whole physical realization coherence | not performed | **T8-H** |
| composed Golden Flows + proof matrix | not frozen | **T9** |
| code/schema/API/frontend/runtime transition/cutover | not designed against final physical target | **T10** |
| bounded implementation dependency graph | not authorized | **T11** |
| independent implementation-readiness challenge | not performed | **T12** |

---

## 12. Coverage verdict

```text
semantic target T1→T6                     CLOSED / PRESERVED
physical technical target                  INCOMPLETE BY DESIGN
current technical evidence coverage        SUFFICIENT TO ROUTE NEXT DESIGN
old exact audit metrics                     LAST-REPRODUCED / REMEASURE WHEN MATERIAL
legacy technical target docs                NOT TARGET AUTHORITY
hidden implementation decisions acceptable NO
implementation                              BLOCKED
```

No target package, table, binary, deployment or implementation sequence is selected by this census.

---

## 13. Operator review questions

Operator review of this baseline is only about **coverage and evidence classification**, not choosing T8 solutions yet.

Ratify/correct:

1. Does the census cover every technical surface that must be resolved before implementation?
2. Are any `CURRENT-PROVEN`, `LAST-REPRODUCED`, `STALE` or `UNKNOWN` classifications materially wrong?
3. Is any missing decision routed to the wrong future stage?
4. Is there any material implementation concern that would still be left for a Writer after T7→T12 as defined?

If the answer is clean:

```text
operator accepts TRRB
→ complete technical-document routing reconciliation
→ open redefined T7 Historical Migration Truth & Semantic Mapping
```

No T7 substantive design or product implementation is authorized by this census.
