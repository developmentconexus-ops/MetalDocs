# R10-T8A — Technical Document / ADR Authority Reconciliation

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **T8-A DOCUMENT ROUTING CANDIDATE**  
> **Date:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** BLOCKED

## 1. Problem

Several maintained legacy technical pages still contain words such as `canonical`, `MUST`, `target layer` or `current target authority`, while the durable R10 router/TRRB explicitly demoted their physical topology to evidence.

A Fresh Actor who lands on one of those pages directly can therefore inherit a local maximum by prose even when the repository router forbids it.

T8-A must reconcile **document authority**, not rewrite target architecture inside old documents.

## 2. Binding routing law

Current target authority is:

```text
AGENTS.md
→ DevelopmentConexus Engineering Method
→ current-agent-handoff.md
→ r10-technical-architecture.md
→ Product Contract / GCR / ownership
→ T1→T7 durable authorities
→ Registry amendments
→ post-T6 program / TRRB
→ active T8-A staging
```

Current code/schema/API/frontend/runtime and historical ADRs are evidence only unless R10 explicitly preserves/promotes the exact property.

## 3. Preliminary document disposition

| Document family | Evidence role | Preliminary T8-A disposition | Required action |
|---|---|---|---|
| `wiki/architecture/r10-*`, Product Contract, Whole-Product GCR, ownership, current Registry amendments | current durable R10 authority | **PRESERVE** | remain authoritative; router continues to own current stage/status |
| `AGENTS.md`, `current-agent-handoff.md`, `r10-technical-architecture.md` | current routing/bootstrap | **PRESERVE / REFINE** | keep synchronized; never duplicate architecture meaning |
| `wiki/architecture/backend-blueprint.md` | detailed current backend composition / Aug-09 audit evidence | **CURRENT-STATE ONLY** | remove/override any apparent target/canonical authority during T8-A closure; later rewrite/delete after T8 target replaces it |
| `wiki/architecture/backend-api-structure.md` | current HTTP/codegen module realization evidence | **CURRENT-STATE ONLY** | its contract-first enforcement property may survive, but tag-per-module / generated-package-per-legacy-module rules do not bind T8-E |
| `wiki/architecture/frontend-structure.md` | current frontend realization evidence | **CURRENT-STATE ONLY** | feature-sliced/TanStack/API mechanisms are evidence; legacy domain feature tree and `ArtifactViewModel` cannot route T8-F |
| `wiki/architecture/data-model.md` | current DB pointer/evidence page | **CURRENT-STATE ONLY / REFINE ROUTING** | already says target not designed, but stale Cohesive Redesign links must be replaced by current R10 router |
| `wiki/backend/repo-topology.md` | current repo/process inventory | **CURRENT-STATE ONLY** | preserve as archaeology inventory until T10; remove target links/claims that imply current process count is target |
| `wiki/modules/*` | per-current-module implementation documentation | **CURRENT-STATE ONLY** | useful for T8-A/T10 archaeology; never target owner authority |
| `wiki/architecture/backend-target-architecture.md` | prior physical target | **SUPERSEDED** | no re-promotion by inertia; Git/wiki history only |
| `wiki/architecture/cohesive-platform-redesign.md` | prior redesign routing | **SUPERSEDED** | no active target routing |
| `wiki/decisions/index.md` | historical ADR register | **REWRITE ROUTING** | currently points to Cohesive Redesign and calls outdated tenancy/RLS/Periodic Review assumptions retained; must route through R10 and classify ADRs as historical evidence unless an exact R10 property preserved them |
| pre-R10 ADRs under `wiki/decisions/*` | rationale/history + possible reusable mechanism evidence | **CURRENT-STATE/HISTORICAL EVIDENCE BY DEFAULT** | `Accepted` is not target inheritance; preserve a property only where Product Contract/T1→T7 or later T8 independently proves it |
| pure engineering/test mechanism docs (e.g. `tools/verify` operating rationale) | mechanism evidence | **PRESERVE/REFINE candidate** | only while still consistent with current mechanism; target-specific guard policies change with T8 |

## 4. ADR inheritance law

The old rule “pure infrastructure ADRs may continue when they do not conflict” is too weak for T8-A because absence of an obvious contradiction is not proof of Global Maximum.

Replace inheritance posture with:

```text
historical ADR proposes property/mechanism
→ identify current named consumer
→ map to ratified R10 property
→ compare against alternatives when material
→ PRESERVE only if still smallest sustainable solution
```

Examples:

```text
ADR contract-first API
→ named T6 property exists
→ mechanism already has proof/guards
→ strong PRESERVE/REFINE candidate

ADR tenant RLS
→ pooled tenant isolation consumer removed from Launch
→ current mechanism is accidental complexity for target unless another invariant proves a need
→ REWRITE/DELETE candidate

ADR approval delegation / SLA / fast-forward
→ no Launch consumer and T2 explicitly rejects baseline engine
→ SUPERSEDED / DELETE capability implementation

ADR periodic review
→ named product capability still exists but moved to Launch+
→ no Launch implementation survives merely because ADR was Accepted
```

## 5. Closure action

Before T8-A can close:

1. update the durable documentation landing/routing pages that still misroute Fresh Actors;
2. do **not** rewrite legacy detail pages into target architecture before T8-B→G decides the target;
3. instead mark them unambiguously `CURRENT-STATE EVIDENCE ONLY` or `SUPERSEDED`;
4. preserve Git history as archive;
5. let T10 own final deletion/replacement of legacy technical documentation when current→target transition is fully known.

This reconciliation is a candidate until T8-A adjudication/ratification. It does not open T8-B.
