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

T11 remains **OPEN / ACTIVE** on Draft PR #162.

Reusable frontend planning method:

```text
docs/development/functional-html-wireframe-method.md
Frontend Product Experience Planning Method v2.1
Fable Round 2: CONVERGED / MATERIAL=0
```

The prior all-at-once MetalDocs HTML prototype is **SUPERSEDED / NOT ACCEPTED** and is not an implementation baseline.

The current frontend gate is:

```text
B01 — App Shell + Global Information Architecture   LOCKED / OPERATOR-RATIFIED
B02 — Library / discovery                           LOCKED / OPERATOR-RATIFIED
B03 — Document Official                             OPEN / CANDIDATE / NOT LOCKED
B04 — Document Work / authoring                     NOT OPEN
```

B01 was operator-locked on 2026-08-22 after the revised **Home A — Central operacional** visual walkthrough. Its bounded durable record and canonical rendered structural artifact are:

```text
docs/work/current/t11-b01-app-shell.md
docs/work/current/t11-b01-app-shell-wireframe.html
```

B01 P9 Screen Contract / bidirectional trace and bounded P10 consolidation are complete in the B01 record. The lock freezes the current structural baseline but remains subject to the normal smallest-scope evidence-backed reopen law. The operator-required in-application Notification capability discovered during B03 may justify exactly such a bounded B01 reopen for notification chrome; no B01 structure is changed until that mini-design is ratified.

B02 was operator-locked on 2026-08-22 after iterative visual adjudication from table-first hypotheses to **C3 — Discovery-first Library with Por tipo + Por área**. Its bounded record and canonical rendered structural artifact are:

```text
docs/work/current/t11-b02-library.md
docs/work/current/t11-b02-library-wireframe.html
```

B02-F1 is resolved by the bounded `B02-LD` Library read-composition precision: the first page of the existing `listDocuments` operation supplies complete disclosure-safe discovery options for Document Type, Area and responsible owner from the actor's currently disclosable status universe. It adds no operation, owner, Permission or route; operation 79 remains absent. B02 P9 Screen Contract / bidirectional trace and bounded P10 consolidation are complete in the B02 record. The precision must be consolidated into effective T6/T8-E/T8-F owners before T11 integration.

B03 is now OPEN / CANDIDATE. Its bounded current record is:

```text
docs/work/current/t11-b03-document-official.md
```

Direct operator feedback plus bounded legacy evidence corrected the initial viewer-first hypothesis. Current B03 direction is **record/ficha first + deliberate distinct read-only official-content viewer**. The stable Document record must communicate identity, official truth, responsibility, current work/obsolescence context, history entry, management affordances and the operator-required Document Discussion surface.

Two material B03 findings are now open:

```text
B03-F1
server-derived command-affordance hints for
create Revision / replace responsible owner / create obsolescence / withdraw obsolescence
without rebuilding Authorization in React

B03-F2
operator-required stable-Document Discussion
+ @mention of current MetalDocs Users
+ required in-application Notification
```

B03-F2 is a genuine bounded Product/architecture reopen trigger. Current Product/T3/T5/T6/T8-E/T8-F authority has no Document Discussion semantic model, mention contract or Notification surface; Notifications were explicitly deferred absent a named consumer. The operator has now supplied that consumer and requires the capability before Launch V1.

B03 is temporarily paused at the B03-F2 mini-design gate. The mini-design must close the user-visible semantics that can change B01/B03 and backend contracts before visual adjudication resumes. Framework/repository/library selection is explicitly later evidence work and may not define Product semantics.

Only the operator may set `LOCKED`. B04 remains NOT OPEN.

## T11 fixed system invariants

```text
opening integrated main               cae6ba48df5d611959c0390e0f2b9b8194d62a9d
branch                                 arch/t11-implementation-program
Draft PR                               #162
application operations                 78
orphaned / invented                    0 / 0
operation 79                           ABSENT
Idempotency-Key creations             10
ETag read / mutation domains          13 / 13
exact-byte resources                  4
Product implementation                BLOCKED
```

These are the **currently accepted** system invariants. B03-F2 is an authorized material reopen candidate, not permission to silently change them. Exact operation, Permission, idempotency and async deltas are updated only after the Discussion/@Mention/Notification mini-design is operator-ratified and the smallest implicated authorities are explicitly reopened.

The operator-approved T8-E-RO precision remains binding and adds no operation, Permission, owner or route.

## Frontend planning correction

The first T11 frontend pass proved backend/screen coverage but moved directly from semantic Screen Contracts to a whole-product HTML prototype. Operator review rejected that workflow because it skipped deliberate product/UX planning for:

```text
user needs / jobs
information architecture
navigation mental model
reference study
layout alternatives
screen hierarchy
relative size / density
cards vs tables vs lists vs master-detail
progressive disclosure
screen-by-screen operator visual walkthrough
pattern derivation after reviewed repetition
```

No Product/T1→T10 authority was reopened by this methodological correction.

## Frontend Product Experience Planning Method v2.1

The converged process is:

```text
GLOBAL FOUNDATION
P0  accepted authority
P1  actors / jobs / user needs
P2  end-to-end user flows
P3  frontend coverage
P4  candidate Information Architecture
P5  screen / material-surface inventory

PER-BLOCK LOOP
P6  targeted reference study when triggered
P7  competing layout hypotheses when ambiguity is real
P8  rendered structural wireframe + operator visual adjudication
    → only operator may LOCK
P9  Screen Contract + bidirectional backend trace
P10 bounded pattern consolidation after LOCK
P11 interaction realization / low-fi prototype after structural lock

ASSEMBLED CLOSURE
P10 terminal pattern reconciliation
P11 assembled interaction prototype when required
P12 adversarial UX + architecture walkthrough
P13 visual-design handoff + structural-conformance review
P14 frontend implementation-readiness closure
```

Hard laws include:

```text
no whole-product wireframe generation in one pass
no assistant/reviewer/tool may set LOCKED
P8 must be a rendered/viewable visual artifact; prose alone is insufficient
no layout chosen merely from backend shape
no shared pattern frozen before reviewed repetition exists
no screen-shaped API for frontend convenience
accessibility + responsive behavior are structural
material assumptions remain registered until probed/resolved
visual design cannot silently change locked structure
```

### Methodology review proof

```text
v2 candidate HEAD                     a9e6f3b3ae2b8e56c65d8114e1551e40ec1d7161
candidate CI                          #1218 SUCCESS
Round 1 Evidence PR                   #163 CLOSED / UNMERGED
Round 1 verdict                       NOT CONVERGED
Round 1 findings                      MATERIAL=2 / IMPORTANT=10
corrected v2.1 HEAD                   4b0bf734e70e59e06008d401a0d3d12d9540310e
corrected CI                          #1221 SUCCESS
Round 2 Evidence PR                   #164 CLOSED / UNMERGED
Round 2 valid Evidence HEAD           e2336bab468e04a56ed850cfa93a1fb6f53ca530
Round 2 CI                            #1223 SUCCESS
Round 2 verdict                       CONVERGED
Round 2 findings                      MATERIAL=0 / IMPORTANT=0 / OPTIONAL=3
Round 3                               NOT JUSTIFIED
```

Fable's original Round-2 response commit `f950ccabe1293926e481951ebda3cfebbccf91a1` was accidentally authored on the old Round-1 lineage. Its exact `ai-dialog.md` blob was re-anchored byte-identically onto the correct Round-2 branch rooted at `4b0bf734...`; CI #1223 verified Repository Standard isolation. The review text was not altered.

The three Round-2 notes are non-blocking realization precision and do not justify a methodology change or Round 3.

## MetalDocs UX block program

Current block sequence:

```text
B01  App Shell + Global Information Architecture   LOCKED / OPERATOR-RATIFIED
B02  Library / discovery                            LOCKED / OPERATOR-RATIFIED
B03  Document Official                              CURRENT / CANDIDATE / NOT LOCKED; B03-F2 MINI-DESIGN GATE
B04  Document Work / authoring                      NOT OPEN
B05  My Work                                        NOT OPEN
B06  Governance                                     NOT OPEN
B07  History / Audit                                NOT OPEN
B08  Administration                                 NOT OPEN
```

Names/order after B01 remain candidate because later evidence may refine grouping without changing Product authority. Any locked-block reopen follows the normal smallest-scope finding law and does not invalidate later work automatically.

Each material block follows:

```text
bounded authority + user goals
→ targeted reference study when triggered
→ 2–3 structural hypotheses when real ambiguity exists
→ lightweight backend/data feasibility check
→ rendered visual structural candidate
→ operator walkthrough: hierarchy / position / size / density / discoverability
→ findings / revision
→ OPERATOR LOCK
→ Screen Contract + backend trace
→ bounded pattern consolidation
→ interaction realization only after lock
```

The assistant and operator converse block-by-block. Remaining screens are never generated automatically as baseline.

## Final implementation DAG candidate

The backend/implementation partition remains:

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

`33 + 10 + 22 + 9 + 4 = 78` is the current accepted implementation partition. B03-F2 may require a bounded DAG/census amendment after ratification; no speculative count is assigned beforehand.

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

Before T11 integration, effective T6/T8-E/T8-F owners must consolidate this approved precision together with the B02-LD Library discovery precision and any later operator-approved B03 read/collaboration precision.

## T10 preserved authority

Binding barriers remain:

```text
B0 source truth
→ B1 private target
→ B2 exact candidate proof + verified clean seal
→ B3 first authoritative Product mutation / point of no return
→ B4 authoritative recovery point + serving fence + canonical activation
```

No historical business migration, dual Product authority, legacy fallback, compatibility bridge or Product activation marker is introduced. Any future operation-census delta from B03-F2 must be explicitly ratified before it becomes current authority.

## Exact next action

```text
B01 remains LOCKED / operator-ratified, subject only to evidence-backed smallest-scope reopen
→ B02 remains LOCKED / operator-ratified
→ B03 is OPEN / CANDIDATE / NOT LOCKED
→ B03-F2 mini-design gate is CURRENT
→ close, one decision at a time:
     Discussion read/write eligibility
     message/reply semantics
     @mention candidate + validation semantics
     in-app Notification create/read/unread/navigation semantics
     information-disclosure/offboarding behavior
     smallest B01 notification-chrome impact
     smallest Product/T3/T5/T6/T8-E/T8-F reopen set
→ do NOT select chat/notification framework before semantics are frozen
→ operator ratifies mini-design
→ consolidate only the smallest approved upstream authority delta
→ update exact operation / Permission / idempotency / async census from approved delta
→ apply bounded B01 reopen only if notification chrome requires it
→ resume B03 record-first wireframe with real Discussion semantics
→ resolve B03-F1 before any B03 LOCK
→ only operator may LOCK B03
→ only after B03 LOCK complete P9 Screen Contract / bidirectional trace and P10 bounded pattern consolidation
→ do not open B04 as baseline before B03 LOCK
→ do not generate B04+ in advance
→ do not begin T12
→ do not implement Product code
```

## Remaining architecture program

| Stage | Owns | State |
|---|---|---|
| T8-E | Executable application wire | CLOSED / INTEGRATED; T8-E-RO + B02-LD consolidation pending T11 close; B03-F2 may cause bounded reopen |
| T8-F | Frontend realization | CLOSED / INTEGRATED; T8-E-RO + B02-LD consolidation pending T11 close; B03-F2 may cause bounded reopen |
| T8-G | Runtime / process / deployment | CLOSED / INTEGRATED |
| T8-H | Whole-T8 coherence | CLOSED / INTEGRATED |
| T9 | Golden Flows / validation baseline | CLOSED / INTEGRATED |
| T10 | Transition / cutover | CLOSED / INTEGRATED |
| T11 | Implementation graph + implementation-readiness | OPEN / B01+B02 LOCKED; B03 OPEN / B03-F2 MINI-DESIGN GATE |
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

Completed Product/R10 decisions reopen only on material evidence defined by the owning authority or the DevelopmentConexus Engineering Method. Preference, sunk cost, old implementation shape, hypothetical future capability or infrastructure fashion are not reopen triggers. The operator-required B03-F2 capability is material current-product evidence and is therefore a valid bounded reopen trigger.
