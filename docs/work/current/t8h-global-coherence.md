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

Additional authority is opened only for a concrete seam.

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

## 3. Lead Whole-T8 findings

| Seam | Lead result |
|---|---|
| T8-A → T8-B legacy-shape inheritance | PASS |
| T8-B → T8-C dependency/package completeness | H3 FOUND / CORRECTED |
| T8-C → T8-D transaction/Audit/AuthZ/idempotency/GC/River persistence | PASS |
| T8-D → T8-E ETag/CSRF/replay/exact-byte executability | PASS |
| T8-E → T8-F fresh-navigation/read symmetry | H2 FOUND / CORRECTED |
| T8-F → T8-G concrete runtime consumers/YAGNI | PASS |
| sole mutable status authority | H1 FOUND / FIRST CORRECTION INCOMPLETE |
| Authorization single ALLOW/default-DENY authority | PASS |
| same-local-commit Audit | PASS |
| River transaction fit | PASS |
| idempotency key/fingerprint/replay/24h semantics | PASS |
| exact-content authority/storage/wire/runtime spool | PASS |
| same-PDF reuse + conditional DOCX→PDF | PASS |
| canonical PostgreSQL Search / no dormant external engine | PASS |
| operation census | PASS — 78; operation 79 absent |

## 4. H1 — mutable status duplication

Root cause: durable Product/architecture/ratification documents captured future program-state snapshots even though roadmap is the sole mutable status authority.

Initial operator-approved correction removed stale progression state from T8-F/T8-G architecture/ratification records and the decision register. That correction was valid but not exhaustive.

Independent Fable Round 1 later found a material same-class contradiction in earlier durable authorities; H1 therefore remained open until the broader bounded sweep below.

## 5. H2 — executable wire authority split

`wire-contract.md` now owns the complete effective `DocumentOfficialView`, including:

```text
open_revision?: OpenRevisionRoutingReference
active_obsolescence_request_id?: Uuid
```

with exact existence/disclosure laws. `frontend-read-symmetry.md` is routed decision provenance only and no longer supersedes executable schema.

No operation, Permission, Product capability or persisted Document pointer was added.

## 6. H3 — missing maintenance application leaf

T8-B now includes:

```text
internal/application/maintenance
```

for T5-J managed-content GC choreography only. No semantic owner, public owner surface, dependency class, runtime shell or application operation was added.

## 7. Initial proof history

The initial candidate mechanically preserved the 78-operation census and passed required CI. CI #1093 also caught a real unrouted-durable-document defect during H2 consolidation.

The initial H1 proof was subsequently falsified for **completeness** by independent review. That falsification is retained rather than rewritten as a pass.

```text
H2 executable-wire SSOT                     PASS
H3 application/maintenance class            PASS
application census = 78                     PASS
operation 79 absent                         PASS
required candidate CI                       PASS
H1 complete durable-status sweep            FALSIFIED BY FABLE F1
```

## 8. CI proportionality correction before Fable

Operator-approved correction already applied:

```text
valid Draft review/* + candidate base + ai-dialog-only delta -> PASS
review/* marked ready or violating isolation                 -> FAIL
forward-obligation counts derive from document declarations
Markdown whitespace                                           -> warning only
superseded same-PR runs                                       -> cancelled
material routing/envelope/provenance guards                   -> blocking
```

Workflow correction commit `bb8164f1f8cc784268f5f1f7515614c4703a37af` passed required CI #1096. Candidate `96c99fa6edfed5b015c5f93dbec9a40d806b8c95` passed #1098. Valid review isolation passed #1099.

## 9. Fable Round 1

Evidence channel:

```text
PR                              #149
branch                          review/t8h-fable
reviewed candidate HEAD         96c99fa6edfed5b015c5f93dbec9a40d806b8c95
final Fable HEAD                4990a8b1c0e3123a135989577d0f13bda2175932
final Evidence CI              #1100 SUCCESS
review delta                    docs/work/current/ai-dialog.md only
verdict                         NOT CONVERGED
MATERIAL                        1
```

### F1 — MATERIAL — ACCEPTED

Claim: H1 was incomplete. Concrete stale live-state examples included:

```text
docs/architecture/persistence.md
  T8-E ACTIVE
  T8-F→T8-H NOT OPEN
  T9→T12 NOT OPEN
  implementation BLOCKED

docs/product/alignment.md
  T1 ACTIVE
  T2→T7 NOT OPEN
  absent active staging packet
  stale NEXT action

docs/architecture/interfaces.md
  next active stage = T8-D

docs/architecture/technical-baseline.md
  next allowed stage = T8-B
```

Lead verified the same class also remained in T8-B's `next open stage = T8-C` closure block.

Adjudication:

```text
ACCEPT
reopen scope = T8-H H1 only
upstream semantic reopen = NO
```

### F2 — MINOR — ACCEPTED

`git diff --check` detects both whitespace and leftover conflict markers. The first CI proportionality correction downgraded the entire command result to warning, accidentally downgrading conflict-marker corruption too.

Adjudication:

```text
whitespace findings             WARNING
leftover conflict marker        BLOCKING ERROR
```

Use Git's own diagnostic; do not create a duplicate marker grammar.

### F3 — MINOR — ACKNOWLEDGED / NO BROAD REGEX

Recurrence risk is real, but a blanket grep over all durable authority text would conflate:

```text
immutable ratification state
historical handoff snapshots
pre-reset provenance strings
actual mutable current-program state
```

That would reintroduce low-signal/brittle CI. Existing documentation governance already makes roadmap the sole mutable status authority and explicitly classifies imported provenance. Round 2 must challenge whether the bounded sweep plus this law is sufficient.

PR #149 was closed unmerged after adjudication.

## 10. Round-1 correction package

### F1 / H1 completion

`docs/product/alignment.md`
- stale T1 current gate/staging packet/NEXT removed;
- A1-A10, 4+1 and T1→T7 ratification facts retained;
- current program state routes only to roadmap.

`docs/architecture/technical-baseline.md`
- T8-A closure remains immutable;
- T8-B is framed as historical downstream handoff, not current next stage;
- current stage/implementation/next action route to roadmap.

`docs/architecture/backend.md`
- T8-B closure remains immutable;
- T8-C is framed as historical downstream handoff;
- live `next open stage` / implementation snapshot removed.

`docs/architecture/interfaces.md`
- T8-C closure remains immutable;
- T8-D/T8-E sections describe authority partition, not live progression;
- T8-D handoff is historical; mutable state routes to roadmap.

`docs/architecture/persistence.md`
- direct stale T8-E ACTIVE / T8-F→T8-H NOT OPEN block removed;
- T8-D closure remains immutable;
- T8-E handoff is historical; mutable state routes to roadmap.

Imported pre-reset title/provenance blocks elsewhere remain governed by `docs/development/documentation.md`; no recursive cosmetic normalization is justified.

### F2

`.github/workflows/ci.yml` now separates `git diff --check` findings:

```text
leftover conflict marker -> error + exit 1
other diff-check output  -> warning
```

Material repository-envelope/routing/provenance guards remain unchanged.

## 11. Preserved invariants

```text
Product reopen           NO
T1→T7 reopen             NO
T8-A semantic reopen     NO
H2 wire semantics         UNCHANGED
H3 topology semantics     UNCHANGED
application operations   78
operation 79             ABSENT
new Permission           NONE
new semantic owner       NONE
new persistence authority NONE
new runtime capability   NONE
T9                       NOT OPEN
implementation           BLOCKED
```

## 12. Current decision

```text
T8-H outcome              NOT YET RATIFIED
Fable Round 1             NOT CONVERGED / 1 MATERIAL
F1                         ACCEPTED / CORRECTION APPLIED
F2                         ACCEPTED / CORRECTION APPLIED
F3                         ACKNOWLEDGED / BROAD REGEX REJECTED
Round 1 Evidence PR        CLOSED / UNMERGED
Round 2                    REQUIRED AFTER EXACT CANDIDATE CI
```

Next: prove the exact corrected candidate with required CI. If green, create a **fresh** isolated bounded Round-2 review from that exact candidate. Round 2 must attack F1 closure, F2 conflict-marker behavior, the F3 adjudication and any semantic/census regression; it must not restart T8-H from preference.
