# R10-T8A — Technical Authority & Legacy Disposition — Candidate

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **T8-A CANDIDATE / MATERIAL ADJUDICATION NEXT**  
> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** BLOCKED

This candidate closes the T8-A question only:

> **Which current technical properties/shapes are legitimate inputs to target realization, and which must be preserved, refined, rehomed, rewritten, deleted, retained only as current-state evidence, or treated as superseded?**

It does **not** choose the final T8-B→T8-G physical design.

## 1. Evidence basis

Binding target truth:

```text
Product Contract REV001
Whole-Product GCR + 4+1 ownership
T1→T7 durable R10 authorities
Decision Registry amendments
Post-T6 Implementation Readiness Program
TRRB evidence classes
DevelopmentConexus Engineering Method
```

Current implementation evidence inspected across:

```text
repo / Go module / process roots
backend module imports and composition
platform→module dependency examples
PostgreSQL schemas/baseline/RLS/grants
foreign-SQL guard baseline
OpenAPI + codegen/conformance gates
local AuthN/session implementation
frontend routes/features/generated transport/query provider
object storage / verified store / revision freeze pointers
compose/runtime/deploy mechanisms
tools/verify + registry + CI delegation
legacy technical docs / ADR routing
```

Product source on this redesign branch remains the base implementation at `main@7f5b8928cc5a13feb8ee3fa7c8ceb1c7d3655a18`; the redesign PR changed documentation/routing/support artifacts rather than product implementation surfaces.

## 2. Root cause

The current MetalDocs implementation was optimized around a materially different product architecture:

```text
~15 semantic modules
pooled tenant/RLS substrate
parallel Documents / ControlledDocuments / Templates concepts
legacy Approval kernel + delegation/SLA/review variants
local-password AuthN
legacy module-tag API surface
legacy domain-feature frontend topology
provider/storage keys embedded in semantic rows
Distribution/Notifications/Periodic Review/tenant lifecycle and other non-Launch capability code
```

Over time the code acquired useful engineering mechanisms, but those mechanisms are interleaved with superseded domain assumptions.

The root problem is therefore not “technical debt inside the existing architecture.”

It is:

> **the existing physical architecture encodes a semantic model that is no longer the accepted product.**

Incrementally polishing that shape would preserve the wrong decomposition.

## 3. Target invariant for T8-A

> **T8-B→T8-G must be free to derive the smallest sustainable physical realization from T1→T7 without compatibility obligation to disposable DEV data or legacy module/table/API/frontend/process shapes. Proven technical mechanisms may survive only when their value remains independent of the superseded shape.**

Corollaries:

1. `PRESERVE` applies to a proven property/mechanism, not automatically to the current file/package/process arrangement.
2. A semantic mismatch is sufficient reason to `REWRITE / REHOME / DELETE` even when current tests are green.
3. A mechanism is not deleted merely because it is old if it still has a named R10 consumer and falsifiable proof.
4. T8-A may classify shape but may not invent the replacement shape.
5. T10 owns the concrete current→target transition/deletion choreography after T8/T9 freeze the target and proofs.

## 4. Alternatives

### A — Legacy-shaped incremental refactor

Keep current module/table/API/frontend/process topology as the scaffold and progressively rename/remove/refactor until it resembles R10.

**Advantages**
- superficially smaller local diffs;
- maximum immediate reuse of existing tests/code;
- familiar current runtime.

**Why rejected**
- gives existence/sunk cost survival authority;
- encourages 1:1 mapping from 15 legacy modules to 4+1 semantic ownership;
- current foreign SQL proves many existing boundaries are not clean authorities;
- current persistent/API/frontend shapes directly encode removed/deferred concepts;
- current DEV data has no compatibility consumer;
- creates a high risk that T8-B→G merely rename the local maximum.

**Method verdict:** `LOCAL MAXIMUM / REJECT`.

### B — Clean-slate physical realization inside the current project, selective mechanism reuse

T8-B→G derive package, communication, persistence, wire, frontend and runtime topology directly from ratified target properties. Existing code is an implementation quarry: reuse only mechanisms/algorithms/tests whose consumer and proof survive R10; rewrite/rehome/delete the rest.

**Advantages**
- Structural Inversion compliant;
- preserves freedom to simplify ownership and transactions;
- no fake backward compatibility;
- retains expensive proven infrastructure only where it still solves the right problem;
- keeps current repo/history/tooling available for archaeology, transition and verification.

**Costs**
- larger physical rewrite surface;
- T10 must explicitly plan deletion/transition instead of relying on gradual semantic compatibility;
- some current tests become legacy evidence and must be replaced by T9 proof architecture.

**Method verdict:** `GLOBAL MAXIMUM CANDIDATE / PREFERRED`.

### C — Full greenfield repository/toolchain/runtime reset

Create a separate clean project and discard current repo/tooling/code wholesale.

**Advantages**
- maximum psychological separation from legacy;
- no accidental import graph/table/route inheritance.

**Why rejected now**
- no evidence that the repository container, Go/Postgres stack or verification substrate blocks the R10 target;
- would discard proven mechanisms with named current consumers;
- increases transition/provenance/verification risk without solving a demonstrated root cause;
- “rewrite everything” is as unjustified as “preserve everything.”

**Method verdict:** `ACCIDENTAL RESET COMPLEXITY / REJECT unless T8 proves substrate obstruction`.

## 5. Candidate strategy

```text
SELECT B

CLEAN-SLATE TARGET REALIZATION
+ SELECTIVE REUSE OF PROVEN R10-CONSUMED MECHANISMS
+ NO LEGACY SHAPE COMPATIBILITY ENTITLEMENT
```

“Clean-slate” here is a **design freedom law**, not a statement that every source file must be deleted.

T8-B→G must be capable of choosing a target that shares zero legacy package/table/API/frontend/process boundaries if that is the Global Maximum.

Conversely, when a current mechanism independently proves the accepted property, reuse is preferred over ritual rewrite.

## 6. High-confidence disposition by technical surface

### 6.1 Semantic module/package topology

Current legacy semantic modules include `approval`, `auth`, `iam`, `documents`, `controlleddocuments`, `templates`, `taxonomy`, `distribution`, `notifications`, `tokens`, `render`, `search`, `jobs`, etc.

Current source also contains concrete capability baggage such as delegation, approval SLA, fast-forward/review variants, Periodic Review surfacing and tenant lifecycle.

```text
DISPOSITION = REWRITE / REHOME
```

T8-B must derive package topology from the accepted owners/properties, not rename 15 modules. No 1 owner = 1 package law is implied.

### 6.2 Internal communication / foreign SQL

Current `hgcrossmodule` baseline records concrete `approval` reads/writes against `documents`, `controlled_documents`, `document_revisions`, `document_comments` and governance/audit tables. A current platform package also imports a legacy IAM domain package.

```text
DISPOSITION = REWRITE / REHOME communication boundaries
```

Exact historical `55 reads + 12 writes` is not load-bearing to this decision and need not be ritualistically reproduced in T8-A. T10 may remeasure exact current blast radius when transition sizing requires it.

### 6.3 PostgreSQL substrate

T2/T5 already require local PostgreSQL transaction semantics and selected River durability/recovery coherence.

```text
PostgreSQL product-state substrate = PRESERVE
current physical schema/table model = REWRITE
```

Preserving PostgreSQL does not preserve current schemas, tables, RLS or grants.

### 6.4 Current multi-tenant / RLS substrate

Current baseline carries broad `tenant_id`, GUC and RLS policy machinery plus tenant lifecycle/security data.

Launch is single-company and T3 live product Authorization is semantic authority.

```text
current tenant/RLS mechanism = REWRITE / DELETE candidate
DB-level invariant enforcement as a property = RE-DERIVE in T8-D
```

Do not infer either “keep RLS because security” or “remove all DB security because single-company.” T8-D must choose the smallest enforcement that protects the actual target invariants.

### 6.5 DB capability assertion / tripwire mechanisms

Current DB triggers/GUCs assert legacy capability names and subjects.

```text
current scheme = REWRITE
property “bypassing HTTP must not silently bypass material invariants” = PRESERVE AS DESIGN QUESTION
```

T8-C/D decides whether DB grants/constraints/triggers or another smaller enforcement achieves the target property.

### 6.6 Database bootstrap / execution identities

Current repo has a deterministic curated bootstrap and a separation between schema owner/provisioning identity and non-owner serving runtime identity.

```text
deterministic bootstrap/proof property = PRESERVE / REFINE
runtime DB identity != DDL/schema owner = PRESERVE
current db-provision process/file choreography = T8-G/T10 RE-DERIVE
```

### 6.7 OpenAPI / code generation

T6 independently requires OpenAPI 3.0.3, generated Go/TypeScript boundaries and runtime/conformance proof. Current verifier checks contract sync and codegen drift.

```text
contract-first + generated boundaries + drift/conformance enforcement = PRESERVE
current openapi.yaml routes/schemas/tag/module mapping = REWRITE
```

Pre-launch `/api/v1` is rebuilt in place. No `/api/v2` and no compatibility shim.

### 6.8 Authentication/session implementation

Current API implements local identifier/password login and password changes; current DB/runtime carry legacy local-auth/tenant semantics.

T6 requires Keycloak Authorization Code + state/PKCE → ApplicationSession, no local passwords/ROPC/JIT.

```text
local credential capability = DELETE
current AuthN delivery/application shape = REWRITE
fail-closed session/cookie/error/security properties that still apply = reusable evidence
```

### 6.9 Frontend realization

Current route/feature tree mirrors legacy domains rather than T6 semantic lenses.

```text
current feature/route topology = REWRITE / REHOME
current generated transport practice = PRESERVE / REFINE property
React Router / TanStack Query / Zustand / TipTap stack = CURRENT-STATE ONLY until T8-F comparison
```

No installed frontend library gets survival status merely because it exists.

### 6.10 Exact-content / object-storage realization

Current code has useful verified-storage behavior:

```text
SHA-256 verification
size bounds
fail-closed missing/hash mismatch
revision pinning
stored content hashes
```

But current semantic rows/repos persist `storage_key`, `body_docx_snapshot_s3_key`, `final_docx_s3_key`, `final_pdf_s3_key` and current `VerifiedStore` returns `VerifiedPointer{StorageKey, ContentHash, SizeBytes}`. It is MinIO-specific and embeds tenant-key-prefix semantics and generic key deletion/enumeration.

T4 requires provider-neutral:

```text
managed_content_id mechanism handle
ExactContentDescriptor {sha256,size_bytes,content_format}
server-derived admission truth
OPEN→READY
create-once/no-overwrite
admission binding
semantic-reference-aware reclaimable GC
```

Therefore:

```text
current storage/objectstore public contract = REWRITE / REFINE
provider key as semantic reference = DELETE
hashing/size/fail-closed algorithms/tests = reuse candidate
MinIO provider = CURRENT-STATE ONLY
```

### 6.11 Async / River

T5 selected PostgreSQL/River for named durable jobs.

```text
River durable-job mechanism = PRESERVE
current jobs registry / old job vocabulary = REWRITE / DELETE non-Launch jobs
current API/worker/jobs process split = CURRENT-STATE ONLY until T8-G
```

### 6.12 Rendering/editor providers

Current repo contains EigenPal/TipTap history, Gotenberg and a separate Node DOCX renderer.

T5/T6 keep renderer/editor provider replaceable and require a representative fidelity corpus.

```text
adapter/provider-replaceability + fidelity proof requirement = PRESERVE
current provider/process/library choices = CURRENT-STATE ONLY / REMEASURE
```

No provider is ratified by installed-code sunk cost.

### 6.13 Redis / multi-replica rate limiting

Current compose uses Redis for shared rate limiting and carries multi-replica assumptions.

No ratified Launch semantic/runtime decision independently requires Redis.

```text
Redis = CURRENT-STATE ONLY / DELETE candidate
```

T8-G must name a current Launch consumer and prove simpler alternatives inadequate before Redis survives.

### 6.14 Verification control plane

`tools/verify` is one current mechanism with independent proof of value:

```text
one registry for local + CI verification
explicit PASS / FAIL / SKIP
toolchain preflight
fail-closed dependency/selection handling
registry↔CI audit
negative fixture or closed classified waiver for blocking controls
CI delegation instead of parallel YAML gate definitions
```

This solves a documented real failure class and is independent from the superseded domain topology.

```text
tools/verify registry/control-plane model = PRESERVE
current target-specific architecture checks = REFINE / REWRITE as T8 freezes new target
```

New T8 invariants must receive firing guards/proofs rather than keeping old policies green.

### 6.15 Technical documentation / ADRs

Current maintained pages still use `canonical/MUST/target` language around the old 15-module, tenancy and module-tag realization. The ADR index still routes through Cohesive Platform Redesign and treats some superseded assumptions as retained.

```text
R10 durable authority = PRESERVE
current implementation detail pages = CURRENT-STATE ONLY
historical target pages = SUPERSEDED
ADR Accepted status = historical evidence, not target inheritance
legacy routing language = REWRITE during T8-A closure
```

T10 later owns final deletion/replacement map when target realization is known.

## 7. What is explicitly NOT decided by this candidate

T8-A does not decide:

```text
number/names of target Go packages
whether one Go module remains optimal
exact allowed import graph
exact owner query/capability interfaces
exact target schemas/tables/columns/constraints
whether any target RLS remains
exact OpenAPI operations/schemas
React Router/TanStack Query survival
exact frontend directory tree
which interactive DOCX provider wins
which rendering converter wins
number of runtime processes/containers
whether Redis exists in target
exact deployment topology
current→target migration/deletion sequence
```

Those remain T8-B→T8-G/T10 decisions.

## 8. Proof strategy

T8-A closes on classification, not implementation.

Required proof:

1. each `PRESERVE` item maps to a named T1→T7/Method consumer or demonstrated repository correctness failure it solves;
2. each `REWRITE/DELETE` material shape has concrete source/schema/API/frontend/runtime evidence of semantic mismatch or unjustified complexity;
3. current technical docs no longer misroute Fresh Actors as target authority;
4. stale exact metrics remain labeled rather than promoted to current truth;
5. no material target package/schema/API/frontend/runtime design is smuggled into T8-A;
6. T8-B→G receive a clean decision surface rather than legacy defaults.

T9 later proves the composed target behavior. T10 proves transition/deletion. T12 attacks implementation readiness.

## 9. Reopen triggers

Reopen only the affected T8-A disposition if concrete evidence proves:

- a current shape embodies a named R10 property that cannot be preserved sustainably after the proposed rewrite/rehome;
- a deleted/deferred mechanism has a real Launch consumer previously missed;
- the current repository/Go/Postgres/tooling substrate materially prevents the Global Maximum target;
- a transition constraint becomes a true correctness constraint rather than convenience;
- a regulatory/security requirement requires preserving or strengthening a currently rejected mechanism.

Do not reopen for familiarity, file count, test volume or migration discomfort.

## 10. Candidate verdict

```text
LEGACY PHYSICAL SHAPE                           NO SURVIVAL ENTITLEMENT
CLEAN-SLATE TARGET REALIZATION FREEDOM          YES
SELECTIVE REUSE OF PROVEN MECHANISMS            YES
CURRENT DEV DATA COMPATIBILITY                  NO
CURRENT MODULE/TABLE/API/FRONTEND TOPOLOGY      REWRITE / REHOME
NON-LAUNCH CAPABILITY IMPLEMENTATION             DELETE / DEFER
POSTGRES                                         PRESERVE
RIVER                                            PRESERVE
CONTRACT-FIRST + GENERATED BOUNDARIES            PRESERVE
VERIFICATION REGISTRY / CI DELEGATION            PRESERVE
LEAST-PRIVILEGE RUNTIME-vs-DDL PROPERTY          PRESERVE
CURRENT PROVIDERS/LIBRARIES/PROCESS COUNTS       CURRENT-STATE ONLY unless later proved
T8-B→T8-G                                        STILL NOT OPEN
IMPLEMENTATION                                   BLOCKED
```

Candidate Global Maximum:

> **Derive the R10 physical architecture cleanly from ratified product/semantic truth, using the current implementation only as evidence and selectively reusing mechanisms whose value survives Structural Inversion.**

This candidate requires material adjudication and operator ratification before T8-A may be promoted/closed.
