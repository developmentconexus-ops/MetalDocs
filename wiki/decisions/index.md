# Decisions

> **Last verified:** 2026-08-19  
> **Status:** HISTORICAL ADR REGISTER / R10 EVIDENCE INDEX — **NOT TARGET AUTHORITY BY ACCEPTED STATUS ALONE**

Historical ADRs remain valuable rationale and implementation evidence. Under R10, an ADR's historical `Accepted` status does **not** grant automatic inheritance into the target physical architecture.

## Current target authority

For target decisions, read in this order:

1. `../architecture/launch-v1-product-contract.md`
2. `../architecture/whole-product-alignment-review.md`
3. `../architecture/launch-v1-ownership-topology.md`
4. `../architecture/r10-t1-semantic-state-invariants.md` through the latest promoted R10 stage authority
5. `../architecture/rebaseline-decision-registry.md` + current amendments
6. `../architecture/r10-technical-architecture.md` — sole current stage/status/next-action router
7. `../references/current-agent-handoff.md`

Current promoted technical-realization disposition:

- `../architecture/r10-t8a-technical-authority-legacy-disposition.md`
- `../architecture/rebaseline-decision-registry-t8a-amendment.md`

## ADR inheritance law

For any pre-R10 ADR:

```text
historical ADR proposes property/mechanism
→ identify a named current R10 consumer
→ map it to a ratified R10 property
→ compare alternatives when material
→ retain only if it remains the smallest sustainable solution
```

Absence of an obvious conflict is **not** proof of Global Maximum.

Current implementation, old tests, migration convenience and sunk cost do not create preservation rights.

## Properties explicitly preserved by current R10 authority

These are preserved because current R10 authorities independently require them, not because their old ADRs exist:

- **Contract-first API / generated Go+TypeScript boundaries** — T6/T8-A.
- **RFC 9457 error envelope** — T6.
- **PostgreSQL product-state substrate** — T2/T5/T8-A.
- **River durable-job mechanism for named T5 jobs** — T5/T8-A.
- **system-owned Release/effectivity** — Product Contract/T1/T2.
- **exact-content SHA/size/fail-closed principles** — T4/T8-A.
- **runtime DB identity separated from schema/DDL ownership** — T8-A.
- **verification registry / local-CI SSOT model** — T8-A.

The current physical implementation of those properties is still subject to later T8 subgates unless explicitly frozen.

## Historical ADRs whose old target meaning is not inherited

Examples include, but are not limited to:

- `0007` / `0022` — old AuthZ grant/capability enforcement model.
- `0021` / `0027` — pooled tenancy / tenant-isolation/RLS assumptions.
- `0069` — Periodic Review implementation; capability is Launch+, not Launch implementation.
- `0077` — approval delegation.
- `0081` — historical governance/signature policy shape.
- `0082` — historical approval-kernel ownership/engine shape.
- `0087` — historical zero-stage route encoding.
- `0092` — historical AuthZ grant-model unification.
- `0093` — earlier Controlled Information redesign wording where it differs from promoted Product Contract/T1→T8-A truth.

These remain evidence. They do not route new implementation.

## T8-A binding disposition

Current legacy shapes such as the 15-module topology, tenant/GUC/RLS mesh, local-password AuthN, current OpenAPI surface, current frontend feature topology, provider-key semantic storage contract and non-Launch job wiring are **REWRITE / REHOME / DELETE / CURRENT-STATE ONLY** according to `r10-t8a-technical-authority-legacy-disposition.md`.

Selective code/mechanism reuse is allowed only when all five T8-A tests pass:

```text
named current R10 consumer
+ public contract free of legacy semantic authority
+ dependency direction fits target
+ proof asserts target property rather than legacy shape
+ reuse remains smaller than rewrite after transition cost
```

## Rule for Fresh Actors

If an ADR conflicts with Product Contract, T1→current promoted R10 authority, Registry amendments, or the active R10 router, **R10 wins**.

Do not edit dozens of historical ADR bodies to rewrite history. This index is the program-level authority boundary; Git history preserves the original rationale.