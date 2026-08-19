# R10-T8B — Backend Module & Package Topology Bootstrap

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **T8-B OPEN / TARGET TOPOLOGY DERIVATION NEXT**  
> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** BLOCKED

T8-A is CLOSED / OPERATOR-RATIFIED / PROMOTED. T8-B is now the only open T8 subgate.

## 1. Authority chain

Read in order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/launch-v1-product-contract.md`
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md` + D4 amendment
10. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
11. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
12. `wiki/architecture/r10-t6-canonical-api-frontend-journeys.md`
13. `wiki/architecture/r10-t7-historical-migration-truth-semantic-mapping.md`
14. `wiki/architecture/r10-t8a-technical-authority-legacy-disposition.md`
15. Decision Registry + D4/T6/post-T6/T7/T8-A amendments
16. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
17. `wiki/architecture/r10-technical-realization-reconciliation-baseline.md`
18. `wiki/architecture/r10-technical-architecture.md`
19. this staging bootstrap
20. current code/import/package evidence only when a concrete T8-B claim needs it

Current legacy module/package existence is evidence only.

## 2. Binding T8-A law

```text
CLEAN-SLATE PHYSICAL TARGET FREEDOM
+ SELECTIVE PROOF-BACKED MECHANISM REUSE
- LEGACY SHAPE INHERITANCE
- FULL-GREENFIELD PURITY RESET WITHOUT EVIDENCE
```

T8-B must therefore derive the backend topology from target semantics rather than map old modules one-for-one.

Selective reuse still requires all five T8-A proofs:

```text
named current R10 consumer
+ public contract free of legacy semantic authority
+ dependency direction fits target
+ proof asserts target property rather than legacy shape
+ reuse remains smaller than rewrite after transition cost
```

## 3. T8-B question

T8-B answers:

> **What is the smallest backend package/module topology that gives each ratified semantic authority one clear home, keeps supporting mechanisms non-semantic, exposes only intentional public surfaces, and makes forbidden dependencies mechanically understandable?**

## 4. T8-B owns

Per the post-T6 program, T8-B freezes:

```text
target repository/package layout
semantic-owner realization boundaries
layering within owners
public/internal Go package surfaces
allowed dependency graph
forbidden dependency graph
composition root / dependency injection
location of shared mechanisms
```

A semantic owner may realize through multiple cohesive packages. Package count follows isolation and clarity, not owner count or legacy module count.

## 5. T8-B does not own

Do not decide by stealth:

```text
exact inter-owner query/capability/read contracts  → T8-C
exact schemas/tables/constraints/locks             → T8-D
exact OpenAPI paths/schemas                         → T8-E
frontend route/package/query realization            → T8-F
process/container/deployment topology                → T8-G
current→target migration/deletion sequence           → T10
implementation task graph                            → T11
```

T8-B may identify that a seam is required to justify dependency direction, but it must not invent the detailed contract that belongs to T8-C.

## 6. Ratified semantic owners to realize

The durable semantic ownership topology is:

```text
1. Authentication
2. Organization
3. Authorization
4. Controlled Documents
5. Audit — supporting evidence authority
```

Supporting mechanisms are not semantic owners merely because they need code/packages:

```text
storage/exact-content
search
rendering/viewer adapters
async/River
HTTP transport/codegen
observability
configuration/bootstrap
database/migrations
idempotency
security middleware
```

Templates are ordinary governed Documents with a role/designation, not a separate semantic owner.

Release/effectivity is system-owned inside Controlled Documents semantics, not a separate business context.

## 7. Core invariants T8-B topology must make obvious

At minimum the package layout/dependency graph must make it difficult to violate:

```text
Authentication != Organization != Authorization
Authorization does not own Organization identity/profile data
Controlled Documents owns document/revision/working/submission/governance/release semantics
Audit records evidence but does not own business state
supporting mechanisms do not become semantic authorities
no package reaches another owner's private persistence/internals
no foreign SQL as owner communication
no shared write authority hidden in platform/common packages
no legacy Artifact/Template/Approval semantic owner resurrection
no dormant Launch+/Future capability modules in Launch
composition owns adaptation; owners do not mutually import private implementation
```

## 8. Evidence posture

Current evidence useful to T8-B includes:

```text
legacy 15-module map                      CURRENT-STATE ONLY
old module SCC / reciprocal-edge counts  LAST-REPRODUCED until load-bearing
current cross-owner SQL leakage           CURRENT-PROVEN qualitatively
go.mod single-module fact                CURRENT-PROVEN / no target entitlement
composition roots                        CURRENT-PROVEN / shape not preserved
module-boundary guards                    mechanism evidence; legacy policy not target
platform package list                     CURRENT-STATE ONLY
```

Remeasure an old metric only if the exact value changes a material topology decision.

## 9. Method loop

For each material topology choice:

```text
Evidence
→ Known / Inferred / Unknown / Deferred
→ Root Cause
→ Target Invariant
→ Constraints
→ 2–3 materially distinct approaches
→ Local vs Global Maximum
→ Essential vs Accidental Complexity
→ YAGNI / future-cost check
→ Ownership / dependency boundary
→ mechanical enforcement concept
→ proof strategy
→ adversarial challenge
→ candidate decision
→ reopen trigger
```

Do not choose a package layout because it resembles Clean Architecture, DDD, hexagonal architecture, the legacy tree, or a fashionable convention. Choose it because it is the smallest sustainable realization of the ratified owners and invariants.

## 10. Required T8-B outputs

Before T8-B can close, staging must contain enough detail to ratify:

1. target backend repository/package tree at meaningful package granularity;
2. semantic owner → package responsibility map;
3. public vs internal package surface law;
4. allowed dependency graph;
5. forbidden dependency graph;
6. composition-root/adaptation law;
7. shared-mechanism placement rules;
8. explicit legacy package disposition classes where needed for T10, without designing the transition sequence;
9. mechanical-enforcement/proof strategy for package boundaries;
10. adversarial comparison of at least 2–3 materially distinct topology approaches;
11. platform-facing summary for operator ratification.

## 11. Exact process

```text
re-anchor current authority
→ derive responsibilities from 4+1 owners + T1→T8-A
→ inspect only load-bearing current package/import evidence
→ identify topology invariants
→ compare 2–3 backend/package approaches
→ choose Global Maximum candidate
→ map allowed/forbidden dependencies
→ adversarial challenge
→ T8-B candidate/design
→ platform-facing summary
→ explicit operator ratification
→ durable promotion + Registry reconciliation + staging cleanup
→ only then T8-C may open
```

## 12. Exact next action

```text
derive target backend responsibilities from ratified semantic owners
→ identify which responsibilities must be cohesive vs isolated
→ compare materially distinct package topology approaches
→ do not start by mapping the 15 legacy modules
```

No product code, migration, schema, exact OpenAPI or frontend implementation is authorized.