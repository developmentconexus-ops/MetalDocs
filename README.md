# MetalDocs

MetalDocs is the company system for creating, governing, approving, publishing, finding and proving the history of official controlled documents.

## Current project state — 2026-08-19

MetalDocs is in the **R10 Post-T6 Implementation Readiness** program.

Product/semantic architecture through T7 is operator-ratified. Physical technical realization, proof architecture, transition planning and implementation decomposition are intentionally being closed **before** product code resumes.

**No product implementation is currently authorized.**

Current stage:

```text
T1→T7   CLOSED / OPERATOR-RATIFIED
T8-A    ACTIVE / TECHNICAL AUTHORITY & LEGACY CENSUS
T8-B→T12 NOT OPEN
CODE     BLOCKED
```

T7 closed with:

```text
NO HISTORICAL BUSINESS MIGRATION REQUIRED FOR LAUNCH
```

Current MetalDocs DB/content/history is DEV/test/throwaway and has no business-history compatibility entitlement.

## Read first

For any non-trivial work:

1. [`AGENTS.md`](AGENTS.md)
2. [`docs/engineering/standards/root-cause-global-maximum-method.md`](docs/engineering/standards/root-cause-global-maximum-method.md)
3. [`wiki/references/current-agent-handoff.md`](wiki/references/current-agent-handoff.md)
4. [`wiki/architecture/r10-technical-architecture.md`](wiki/architecture/r10-technical-architecture.md)
5. the durable authority/current stage named by that router

The canonical wiki landing page is [`wiki/index.md`](wiki/index.md).

Do **not** use `wiki/architecture/cohesive-platform-redesign.md` or `wiki/architecture/backend-target-architecture.md` as current target authority. They are superseded/historical evidence.

## Ratified Launch direction

Launch V1 is single-company and centers on:

- Authentication + local MetalDocs User/session binding;
- Organization: Company, User/Profile, Area, Group, GroupMembership;
- product-owned Authorization with scoped RoleAssignments;
- Controlled Documents: stable Document, business Revision (`REV000` initial issuance), mutable DRAFT WorkingContent, immutable Submission, sequential governance, system-owned Release/effectivity, optional OfficialRendition and governed obsolescence;
- Templates as ordinary governed Documents used to seed independent Documents;
- Audit as supporting evidence, never lifecycle authority;
- current-effective search/read/download;
- backup/restore correctness.

Distribution/Read&Acknowledge and Periodic Review are Launch+. Broader records/evidence/retention, generic interchange/connectors, Training/LMS, pooled tenancy and realtime CRDT remain future capabilities unless explicitly reopened by a concrete consumer.

For exact product and architecture meaning, read the durable R10 authorities through the active router rather than this summary.

## Current T8-A work

T8-A is not yet designing the replacement topology. It is classifying the current technical estate across:

```text
backend packages/modules/import graph
DB/schema/SQL ownership
OpenAPI/codegen/runtime contract mechanisms
frontend routes/features/query/cache/state
async/jobs/rendering
binaries/deploy/config/trust/observability/recovery
verification/tests/CI/architecture guards
technical documentation/ADRs
```

Each material structure receives a disposition:

```text
PRESERVE
REFINE
REHOME
REWRITE
DELETE
CURRENT-STATE ONLY
SUPERSEDED
```

Only after T8-A closes may T8-B choose the target backend/package topology.

## Runtime vs target

The repository still contains the current modular-monolith implementation. Use current code/schema/OpenAPI/frontend/deploy/tests to answer **what runs today**. Use the R10 durable authority to answer **what should exist after the redesign**.

Current implementation existence, sunk cost and DEV/test data do not grant target survival rights.

## Remaining gate before code

```text
T8-A→T8-H Technical Realization
→ T9 Golden Flows & Validation Baseline
→ T10 Transition / Refactor / Cutover
→ T11 Implementation Program & Execution Graph
→ T12 Adversarial Implementation-Readiness
→ Integrated Whole-R10 GCR
→ fresh independent/cold review
→ operator final ratification
→ explicit implementation authorization
```

No code yet.
