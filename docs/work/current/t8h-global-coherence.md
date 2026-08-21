---
id: t8h-global-coherence-review
kind: work
owner: architecture
summary: Branch-only T8-H adversarial review ledger for cross-stage realization coherence.
---

# T8-H — Whole-T8 Global Coherence Review

> **Status:** OPEN / ACTIVE / NOT RATIFIED
> **Opening base:** `main @ 0b4ef6ef891b01f907804cff4bd3c0022aebad80`
> **Implementation:** BLOCKED
> **Scope:** T8-A→T8-G coherence only; T9 is not open.

This is temporary branch-only work. Durable authority lives in the owning T8 documents and mutable program state lives only in `docs/roadmap.md`. This ledger must not enter `main`.

## 1. Review question

> Do T8-A→T8-G form one executable technical realization without duplicate/missing authority, incompatible assumptions, removed required seams or accidental complexity created by the architecture itself?

Bounded review pack:

```text
T8-A  docs/architecture/technical-baseline.md
T8-B  docs/architecture/backend.md
T8-C  docs/architecture/interfaces.md
T8-D  docs/architecture/persistence.md
T8-E  docs/architecture/wire-contract.md
T8-F  docs/architecture/frontend.md
T8-G  docs/architecture/runtime.md
```

Additional authority was opened only for concrete seams: frontend read-symmetry provenance, T8-F/T8-G ratification records, decision register and documentation/CI governance.

## 2. Opening revalidation

```text
main                              0b4ef6ef891b01f907804cff4bd3c0022aebad80
expected main SHA                  MATCH
open PRs                           0
candidate branch                   arch/t8h-global-coherence
Draft PR                           #148
T8-G closeout PR #147              MERGED
opening required CI                #1079 / #1080 SUCCESS
```

Remote historical/candidate/review branches are housekeeping only and do not control stage state.

## 3. Whole-T8 attack result

| Seam | Result |
|---|---|
| T8-A → T8-B legacy-shape inheritance | PASS |
| T8-B → T8-C dependency/package completeness | H3 FOUND / CORRECTED |
| T8-C → T8-D transaction/Audit/AuthZ/idempotency/GC/River persistence | PASS |
| T8-D → T8-E ETag/CSRF/replay/exact-byte executability | PASS |
| T8-E → T8-F fresh-navigation/read symmetry | H2 FOUND / CORRECTED |
| T8-F → T8-G concrete runtime consumers/YAGNI | PASS |
| sole mutable status authority | H1 FOUND / CORRECTED |
| Authorization single ALLOW/default-DENY authority | PASS |
| same-local-commit Audit | PASS |
| River transaction fit | PASS |
| idempotency key/fingerprint/replay/24h semantics | PASS |
| exact-content authority/storage/wire/runtime spool | PASS |
| same-PDF reuse + conditional DOCX→PDF | PASS |
| canonical PostgreSQL Search / no dormant external engine | PASS |
| operation census | PASS — 78; operation 79 absent |

Current River `database/sql` support was checked as external mechanism evidence and remains compatible with the accepted T8-C/T8-D transaction seam. No new substrate decision was required.

## 4. H1 — mutable status duplication

**Root cause:** durable architecture/ratification documents captured future program-state snapshots even though roadmap is the sole mutable status authority.

**Approved correction:** preserve immutable ratification evidence; remove stale integration/downstream-stage assertions; route current progression only to roadmap.

**Result:** T8-F/T8-G architecture records, ratification records and decision register no longer act as mutable roadmap copies.

## 5. H2 — executable wire authority split

**Root cause:** a valid T8-E precision discovered during T8-F remained as a live schema override in `frontend-read-symmetry.md` instead of being folded back into the exact wire SSOT.

**Approved correction:** `wire-contract.md` now owns the complete effective `DocumentOfficialView`, including:

```text
open_revision?: OpenRevisionRoutingReference
active_obsolescence_request_id?: Uuid
```

with their exact existence/disclosure laws. `frontend-read-symmetry.md` remains routed decision provenance only and no longer supersedes an executable schema.

No operation, Permission, Product capability or persisted Document pointer was added.

## 6. H3 — missing maintenance application leaf

**Root cause:** T8-C later admitted the concrete T5-J GC consumer inside the already-existing `application` class, but T8-B's frozen package projection was not reflected.

**Approved correction:** T8-B now includes:

```text
internal/application/maintenance
```

for T5-J managed-content GC choreography only. No semantic owner, public owner surface, dependency class, runtime shell or application operation was added.

## 7. H1–H3 proof

```text
P1  roadmap sole mutable stage/status/next-action authority               PASS
P2  stale T8-F/T8-G progression snapshots removed                         PASS
P3  effective DocumentOfficialView lives in wire-contract.md               PASS
P4  frontend-read-symmetry.md is provenance, not executable schema SSOT   PASS
P5  application/maintenance present with no new dependency class           PASS
P6  canonical application census = 78; operation 79 absent                 PASS
P7  required CI #1095                                                      PASS
```

CI #1093 caught a real defect during this proof: after H2 consolidation, the provenance document was initially unrouted. It is now reachable through the decision register without returning to the executable-wire route.

## 8. CI proportionality review

Evidence showed both useful protection and avoidable red noise:

```text
CI #1081
  FAIL only on three Markdown trailing-space findings
  no architecture/property failure

review PR #145 / CI #1064
review PR #146 / CI #1068
  structurally valid Evidence PRs
  workflow deliberately forced red after successful isolation validation

CI #1093
  FAIL on unrouted durable document
  material repository-structure defect; correct blocker
```

Operator-approved correction applied to `.github/workflows/ci.yml`:

```text
KEEP BLOCKING
  architecture-only path envelope
  bootstrap + budget
  sole-roadmap leakage guard
  durable routing/reachability
  docs/work merge-candidate exclusion
  required authority presence
  archive/provenance reachability

CHANGE
  valid Draft review/* + correct candidate base + ai-dialog-only delta -> PASS
  review/* marked ready or isolation violation -> FAIL
  forward-obligation counts derive from the document's own declarations
  Markdown whitespace -> warning only
  superseded same-PR runs -> cancelled
```

The workflow correction commit `bb8164f1f8cc784268f5f1f7515614c4703a37af` passed required CI **#1096**. The protected aggregate context remains named `required`.

## 9. Current T8-H state

```text
H1                       APPROVED / CORRECTED / PROVEN
H2                       APPROVED / CORRECTED / PROVEN
H3                       APPROVED / CORRECTED / PROVEN
CI proportionality       APPROVED / CORRECTED / PROVEN ON #1096
Product reopen           NO
T1→T7 reopen             NO
T8-A reopen              NO
application operations   78
operation 79             ABSENT
T9                       NOT OPEN
implementation           BLOCKED
```

Next: required CI on the exact current candidate, then a fresh isolated Fable challenge of the corrected Whole-T8 package and CI proportionality. Reviewer findings remain Evidence and may reopen only the smallest authority they materially implicate.
