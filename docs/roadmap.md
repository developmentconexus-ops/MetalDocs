---
id: repository-roadmap
kind: authority
owner: architecture
summary: Sole mutable MetalDocs stage, gate, implementation-status, and next-action authority.
---

# MetalDocs roadmap

## Current state

```text
REPOSITORY MODE                       CLEAN-SLATE / ARCHITECTURE-FIRST
REPOSITORY RESET                      MERGED / OPERATOR-RATIFIED
REPOSITORY STANDARD V1 ALIGNMENT      MERGED
PRODUCT / OWNERSHIP                   OPERATOR-APPROVED
T1 → T10                              CLOSED / OPERATOR-RATIFIED / INTEGRATED
T11                                   OPEN / ACTIVE CANDIDATE
T12                                   NOT OPEN
IMPLEMENTATION                        BLOCKED
LEGACY IMPLEMENTATION                 ABSENT FROM LIVE TREE
```

## Current gate

T11 — Implementation Program & Execution Graph is **OPEN / ACTIVE** on Draft PR #162.

Current branch-only T11 pack:

```text
docs/work/current/t11-implementation-program.md
docs/work/current/t11-frontend-blueprint.md
docs/work/current/t11-wireframes.md
docs/work/current/t11-interaction-ledger.md
```

Reusable method:

```text
docs/development/functional-html-wireframe-method.md
```

Approved bounded precision:

```text
decisions/responsible-owner-selection-read.md
```

The work pack is candidate Evidence, not durable T11 authority. T11 does not authorize Product implementation or T12.

## T11 candidate proof

```text
opening integrated main               cae6ba48df5d611959c0390e0f2b9b8194d62a9d
branch                                 arch/t11-implementation-program
Draft PR                               #162
operator T11 authorization             EXPLICIT / 2026-08-22
operator node/frontend precision       EXPLICIT / 2026-08-22
operator T8-E-RO correction            APPROVED / 2026-08-22
application operations                 78
orphaned / invented                    0 / 0
operation 79                           ABSENT
Idempotency-Key creations             10
ETag read / mutation domains          13 / 13
exact-byte resources                  4
accepted human goals                  16 / 16
material frontend surfaces            36 / 36
Screen Contracts                      36 / 36 READY
Navigation/Data Graph                 COMPLETE
Material Interaction Ledger           COMPLETE CANDIDATE
frontend↔backend trace                78 / 78
Functional HTML method                 ROUTED / CI #1212 SUCCESS
HTML prototype                         REVIEW CANDIDATE / external Evidence
HTML routes / surfaces / operations   10 / 36 / 78
HTML SHA-256                           37378abfb7671767823f07552d9ef2feabad107d33451466d79159d7d7728a12
Product implementation                BLOCKED
```

The HTML prototype remains outside the tracked tree while Repository Standard v1 admits architecture-first Markdown only. T11 does not weaken the allowlist merely to host prototype Evidence.

## Final implementation DAG candidate

```text
P0 authority / implementation-admission pin
 ↓
P1 structural + executable-contract spine
 ├────────────────────┐
P2 persistence        P3 runtime/dependency/non-serving bootstrap
 └──────────┬─────────┘
            ↓
S1 Identity + Organization + Access                         33
 ↓
S2 Document Governance configuration                       10
 ↓
S3 Document core + creation + authoring + Submission       22
   + Library + My Work authoring + History
 ↓
S4 Governance work + Governance Case + Release/rendition    9
 ↓
S5 Obsolescence + Audit                                     4
 ↓
P4 runtime / durable-work / recovery closure
 ↓
T10 B1 private target
 ↓
P5 whole implementation proof closure
 ↓
T10 B2 → B3 → B4
```

`33 + 10 + 22 + 9 + 4 = 78`.

S3 closes only when the real journey is complete:

```text
Library → create/open Document → Document Work → DRAFT edit/upload → Submission
```

No material user-facing node closes as “backend complete; frontend later”.

## T8-E-RO — approved responsible-owner precision

Existing operation 47 remains `getDocument`. `DocumentOfficialView` gains:

```text
responsible_owner_candidates?: UserReference[]
```

Binding law:

```text
present iff current document.owner.manage = ALLOW for the exact Document
contents = complete existing + same-Company + ENABLED Users
order = user_id ASC
absence discloses neither candidate existence nor reason
```

The list grants no authority and is outside the ResponsibleOwner ETag domain. Replacement still rechecks current AuthZ, D4 eligibility/offboarding serialization and `If-Match`.

```text
new capability / Permission / owner / route / operation  none
operation 79                                             absent
```

Before T11 integration, effective T6/T8-E/T8-F owners must consolidate this approved precision.

## T10 preserved authority

Durable authority: `architecture/transition.md`. Ratification evidence: `decisions/t10-ratification.md`.

Integrated proof summary:

```text
candidate PR                #158
Round 1                     MATERIAL=3 / corrected
Round 2                     CONVERGED / MATERIAL=0
operator ratification       EXPLICIT / 2026-08-22
squash-integrated main      e8f415ec16df9cc2d4623981412e1ac21c3c6647
closeout PR                  #161
T10 closeout / opening main cae6ba48df5d611959c0390e0f2b9b8194d62a9d
```

Binding barriers remain:

```text
B0 source truth
→ B1 private target
→ B2 exact candidate proof + verified clean seal
→ B3 first authoritative Product mutation / point of no return
→ B4 authoritative recovery point + serving fence + canonical activation
```

Explicitly absent: historical business migration, dual Product authority, legacy fallback, compatibility bridge, Product activation marker and operation 79.

## Preserved integrated baseline

```text
application operations                  78
operation 79                            absent
Idempotency-Key creations               10
ETag read / mutation domains            13 / 13
exact-byte resources                    4
stable SPA route meanings               accepted T6 set
frontend semantic owner                 none added
frontend Authorization engine           absent
parallel global server-truth store      absent
runtime                                 one modular monolith
Product-state database                  one PostgreSQL database
River                                   in-process
ManagedContentStore                     one active
renderer / MalwareInspector             private conditional mechanisms
Redis / BFF / realtime / external Search / generic event bus absent
Product implementation                  BLOCKED
```

T7 remains binding: Launch has no historical business corpus to migrate. Contrary concrete evidence triggers the smallest bounded reopen.

## Exact next action

```text
operator walkthrough of the current Functional HTML prototype
→ inspect shell, routes, pattern reuse, screens, controls, navigation and material failure/recovery states
→ classify findings against Screen Contract / trace / owning Product authority; visual preference alone does not reopen architecture
→ iterate until functional prototype behavior is accepted
→ reconcile HTML ↔ blueprint ↔ Interaction Ledger; require 36/36 surfaces, 78/78 operations and zero unbound material control
→ record final accepted HTML SHA-256
→ self-review repository candidate + run required CI on exact HEAD
→ present exact HEAD + CI + HTML hash for separate operator approval before independent review
→ only then create isolated review/t11-fable from the approved candidate and provide the exact HTML artifact as reviewer Evidence
→ re-review until MATERIAL=0 or route the smallest justified reopen
→ after convergence consolidate T8-E-RO into T6/T8-E/T8-F, promote durable T11 authority, record ratification and remove temporary T11 work
→ do not begin T12
→ do not implement Product code
```

## Remaining architecture program

| Stage | Owns | State |
|---|---|---|
| T8-E | Executable application wire | CLOSED / INTEGRATED; T8-E-RO consolidation pending T11 close |
| T8-F | Frontend realization | CLOSED / INTEGRATED; T8-E-RO consolidation pending T11 close |
| T8-G | Runtime / process / deployment | CLOSED / INTEGRATED |
| T8-H | Whole-T8 coherence | CLOSED / INTEGRATED |
| T9 | Golden Flows / validation baseline | CLOSED / INTEGRATED |
| T10 | Transition / cutover | CLOSED / INTEGRATED |
| T11 | Implementation graph + implementation-readiness | OPEN / HTML walkthrough gate |
| T12 | Adversarial implementation-readiness | NOT OPEN |

## Final implementation gate

Implementation remains blocked until all are true:

```text
T8  CLOSED / OPERATOR-RATIFIED / INTEGRATED
T9  CLOSED / OPERATOR-RATIFIED / INTEGRATED
T10 CLOSED / OPERATOR-RATIFIED / INTEGRATED
T11 CLOSED / OPERATOR-RATIFIED
T12 CLOSED / OPERATOR-RATIFIED
Integrated Whole-R10 coherence = PASS
fresh independent challenge = converged
operator implementation authorization = explicit
```

## Reopen law

Completed Product/R10 decisions reopen only on material evidence defined by the owning authority or the DevelopmentConexus Engineering Method. Preference, sunk cost, old implementation shape, hypothetical future capability or infrastructure fashion are not reopen triggers.
