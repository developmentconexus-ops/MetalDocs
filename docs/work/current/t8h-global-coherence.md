---
id: t8h-global-coherence-review
kind: work
owner: architecture
summary: Branch-only T8-H adversarial review ledger for cross-stage realization coherence.
---

# T8-H — Whole-T8 Global Coherence Review

> **Status:** OPEN / CONVERGED / NOT RATIFIED
> **Opening base:** `main @ 0b4ef6ef891b01f907804cff4bd3c0022aebad80`
> **Implementation:** BLOCKED
> **Scope:** T8-A→T8-G coherence only; T9 is not open.

This is temporary branch-only work. Durable authority lives in the owning T8 documents and mutable program state lives only in `docs/roadmap.md`. This ledger must not enter `main`.

## 1. Review question

> Do T8-A→T8-G form one executable technical realization without duplicate/missing authority, incompatible assumptions, removed required seams or accidental complexity created by the architecture itself?

Bounded Whole-T8 pack:

```text
T8-A  docs/architecture/technical-baseline.md
T8-B  docs/architecture/backend.md
T8-C  docs/architecture/interfaces.md
T8-D  docs/architecture/persistence.md
T8-E  docs/architecture/wire-contract.md
T8-F  docs/architecture/frontend.md
T8-G  docs/architecture/runtime.md
```

## 2. Lead findings

```text
H1  mutable stage/integration snapshots duplicated outside roadmap
H2  executable DocumentOfficialView meaning split across wire + precision record
H3  accepted application/maintenance GC leaf omitted from T8-B projection
```

All three received explicit operator approval for bounded correction.

### H1

Roadmap is the sole mutable stage/status/implementation/next-action authority. Durable Product/T8 records retain immutable ratification and historical handoff facts only.

Initial H1 correction covered T8-F/T8-G and decision records but was later independently falsified as incomplete. Round-1 Fable found older stale progression in Product alignment and T8-A→T8-D. The broader same-class sweep corrected those without reopening Product or T8 semantics.

### H2

`docs/architecture/wire-contract.md` now owns the complete effective `DocumentOfficialView`, including:

```text
open_revision?: OpenRevisionRoutingReference
active_obsolescence_request_id?: Uuid
```

with exact existence/disclosure laws. `docs/decisions/frontend-read-symmetry.md` remains routed provenance only.

### H3

T8-B now includes:

```text
internal/application/maintenance
```

for T5-J managed-content GC choreography only, inside the existing non-semantic application class.

## 3. CI proportionality correction

Operator-approved bounded correction:

```text
valid isolated Draft review/*             PASS
review/* ready or isolation violation     FAIL
forward-obligation counts                 document-owned / self-consistent
Markdown trailing whitespace              WARNING
leftover merge-conflict marker            FAIL
superseded same-PR runs                    CANCEL
material routing/envelope/provenance       BLOCKING
```

No additional CI framework was introduced.

## 4. Proof history

Useful failures were retained as evidence rather than rewritten away:

```text
CI #1081  whitespace-only red -> identified low-signal blocker
CI #1093  unrouted durable provenance -> valid material repository failure
CI #1094  routing correction -> PASS
CI #1096  first proportional CI correction -> PASS
CI #1098  Round-1 technical candidate -> PASS
```

Application census remained exactly 78 and operation 79 remained absent throughout.

## 5. Fable Round 1 — PR #149

```text
reviewed candidate HEAD       96c99fa6edfed5b015c5f93dbec9a40d806b8c95
candidate CI                  #1098 SUCCESS
final review HEAD             4990a8b1c0e3123a135989577d0f13bda2175932
review CI                     #1100 SUCCESS
review delta                  docs/work/current/ai-dialog.md only
verdict                       NOT CONVERGED
MATERIAL                      1
```

Adjudication:

```text
F1 MATERIAL  ACCEPT
  H1 completeness was false; broaden same-class sweep across implicated older authorities.

F2 MINOR     ACCEPT
  whitespace warning-only remains; Git leftover-conflict-marker diagnostic must block.

F3 MINOR     ACKNOWLEDGE / NO BROAD REGEX
  blanket durable-doc text policing cannot distinguish immutable history from live status
  with acceptable signal; bounded Round 2 must attack this decision.
```

PR #149 was closed unmerged.

## 6. Corrected technical candidate

Exact candidate independently reviewed in Round 2:

```text
b940d4e105a8b837ecdac7f71233ff10d735cd5e
```

Required candidate CI:

```text
#1108 SUCCESS
```

Correction delta from Round-1 candidate was limited to:

```text
docs/product/alignment.md
docs/architecture/technical-baseline.md
docs/architecture/backend.md
docs/architecture/interfaces.md
docs/architecture/persistence.md
.github/workflows/ci.yml
docs/roadmap.md
this temporary ledger
```

No wire-contract, Product journey, operation-census, frontend or runtime semantics changed.

## 7. Fable Round 2 — PR #150

```text
reviewed candidate HEAD       b940d4e105a8b837ecdac7f71233ff10d735cd5e
candidate CI                  #1108 SUCCESS
final review HEAD             5564612d07dc0325ac9b81e441f551340872e59d
review CI                     #1110 SUCCESS
review delta                  docs/work/current/ai-dialog.md only
verdict                       CONVERGED
MATERIAL findings             0
Round 3                       NOT JUSTIFIED
```

Independent conclusions:

```text
F1 / H1 closure                         CLOSED
F2 conflict-marker behavior             CLOSED / independently reproduced
F3 no-broad-regex adjudication          UPHELD
78-operation census                     PASS
operation 79                            ABSENT
new Permission                          NONE
new semantic owner                      NONE
new persistence authority               NONE
new runtime capability                  NONE
H2 meaning                              UNCHANGED
H3 meaning                              UNCHANGED
```

Fable recorded one MINOR, explicitly non-blocking residue: two older durable authorities retain safe-direction bare `Implementation remains BLOCKED` echoes. They do not grant implementation, do not contradict roadmap and do not prevent T8-H ratification. No post-review architecture mutation is justified solely for this optional wording cleanup.

PR #150 was closed unmerged.

## 8. Converged Whole-T8 result

```text
Product reopen                     NO
T1→T7 reopen                       NO
T8-A→T8-G semantic reopen          NO
application operations             78
operation 79                       ABSENT
new Permission                     NONE
new semantic owner                 NONE
new persistence authority          NONE
new runtime capability             NONE
Authorization evaluator            SINGLE CANONICAL AUTHORITY
same-local-commit Audit             PRESERVED
River transaction seam             PRESERVED
idempotency/replay                  PRESERVED
exact-content authority             PRESERVED
OfficialRendition                   PRESERVED
Search baseline                     POSTGRESQL / NO SECOND AUTHORITY
T9                                  NOT OPEN
implementation                      BLOCKED
```

## 9. Current decision

```text
T8-H technical result             CONVERGED
Fable Round 1                     NOT CONVERGED / 1 MATERIAL
Round-1 MATERIAL                  ACCEPTED / CORRECTED
Fable Round 2                     CONVERGED / MATERIAL=0
Round 3                           NOT JUSTIFIED
T8-H operator ratification        PENDING
T9                                NOT OPEN
Product implementation            BLOCKED
```

Next authority action is exclusively in `docs/roadmap.md`: operator adjudication/ratification of T8-H. If ratified, the closure path authors durable ratification evidence, removes this temporary ledger, runs the exact required gate and seeks explicit merge authorization. T9 does not open before T8-H is ratified and integrated.
