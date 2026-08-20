---
id: stage-program
kind: authority
owner: architecture
summary: Defines the remaining architecture, validation, transition, planning, and implementation-readiness stages after T8-E.
---

# Remaining stage program

This page is the durable successor to the pre-reset post-T6 implementation-readiness program. It preserves only the stage definitions still needed after the clean-slate repository reset.

## Program invariant

> No implementation task may contain a material architecture decision that should have been decided before execution.

The current implementation is absent from the live tree and provides no default target shape.

## Remaining stages

| Stage | Owns | Opens when |
|---|---|---|
| T8-E — Executable Wire Contract | Exact OpenAPI application wire, schemas, headers, problems, ETags, idempotency, pagination, upload/exact-byte contract, generated Go/TypeScript boundaries | Current architecture gate after repository reset |
| T8-F — Frontend Realization | Route tree, feature/package topology, generated transport consumption, TanStack Query/query-key behavior, local-vs-server state, read-model consumption, editor/viewer boundaries | T8-E ratified |
| T8-G — Runtime / Process / Deployment | Binaries/processes, River worker ownership, renderer/provider boundary, startup/readiness/shutdown, configuration/secrets, trust/network boundaries, observability, backup/restore runtime roles, environment profiles | T8-F sufficiently closed for runtime consumers |
| T8-H — Whole-T8 Global Coherence Review | Cross-check backend, persistence, wire, frontend, and runtime realization as one system | T8-A→T8-G closed |
| T9 — Golden Flows & Validation Baseline | Falsifiable composed-system flows and proof classes: contract, integration, concurrency, security, recovery, E2E, restore | Whole T8 coherent |
| T10 — Transition / Cutover | Any current→target transition that actually remains, data/schema transition, switch-over, rollback barriers, legacy deletion/cutover readiness | T9 baseline known |
| T11 — Implementation Program & Execution Graph | Bounded work graph, dependencies, scope, touched contracts/persistence/runtime surfaces, proof obligations, rollback/replan triggers | T1→T10 architecture accepted |
| T12 — Adversarial Implementation-Readiness | Independent attack on ownership, dependency graph, persistence, API/frontend/runtime parity, proof completeness, security/recovery, transition, and hidden Writer decisions | T11 complete |

## Final implementation gate

Implementation remains blocked until all are true:

```text
T8  CLOSED / OPERATOR-RATIFIED
T9  CLOSED / OPERATOR-RATIFIED
T10 CLOSED / OPERATOR-RATIFIED
T11 CLOSED / OPERATOR-RATIFIED
T12 CLOSED / OPERATOR-RATIFIED
Integrated Whole-R10 coherence = PASS
fresh independent challenge = converged
operator implementation authorization = explicit
```

T11 may translate architecture into executable work; it may not choose module topology, persistent authority, wire meaning, trust boundary, runtime topology, or migration semantics.

## Provenance

Source authority before reset: PR #131 `wiki/architecture/r10-post-t6-implementation-readiness-program.md`, commit `d8b1c6d31e704e9552a14faa7764c634a29b081d`.

Legacy references to that old path are non-navigational provenance. This page owns the surviving stage definitions.