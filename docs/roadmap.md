---
id: repository-roadmap
kind: authority
owner: architecture
summary: Sole mutable MetalDocs stage, gate, implementation-status, and next-action authority.
---

# MetalDocs roadmap

## Current state

```text
REPOSITORY MODE       CLEAN-SLATE / ARCHITECTURE-FIRST
T1 → T10              CLOSED / OPERATOR-RATIFIED / INTEGRATED
T11                   OPEN / ACTIVE
T12                   NOT OPEN
IMPLEMENTATION         BLOCKED
Draft PR               #162
branch                 arch/t11-implementation-program
opening main           cae6ba48df5d611959c0390e0f2b9b8194d62a9d
```

Current Product/architecture authority lives in `docs/product/**`, `docs/architecture/**`, and `docs/decisions/**`.

Current bounded T11 authorities include:

```text
docs/decisions/discussion-notifications-launch.md
docs/decisions/document-official-actions-read.md
docs/decisions/my-work-governance-identification-read.md
docs/decisions/governance-step-deadline.md
docs/decisions/governance-case-step-deadline-read.md
```

Current system census:

```text
semantic owners                  4 business + 2 supporting
stable SPA routes                11
PermissionCode values            16
application operations           86
Idempotency-Key creations        11
ETag read / mutation domains     13 / 13
exact-byte resources             4
```

`docs/decisions/api-operation-census.md` is the sole numeric census.

## Frontend Product Experience Program

```text
FP0  Frontend Foundation                         CLOSED / R1 86/11 REBASELINED
FP1  Block-by-block Product Experience           ACTIVE
FP2  Integrated Low-Fidelity Product / P11       NOT OPEN
FP3  Whole-Product Adversarial Review / P12      NOT OPEN
FP4  Visual Handoff + Readiness / P13-P14        NOT OPEN
```

Method: `docs/development/functional-html-wireframe-method.md` — Frontend Product Experience Planning Method v2.2.

## FP1 block roadmap

```text
B01   App Shell + Global IA + Home                 LOCKED / OPERATOR-RATIFIED
B01N  Notification global chrome + Quick Inbox     LOCKED / OPERATOR-RATIFIED
B02   Library / Discovery                          LOCKED / OPERATOR-RATIFIED
B03   Document Official / Ficha + Viewer + Discussion
       LOCKED / OPERATOR-RATIFIED
       P8 / P9 / P10                               COMPLETE

B04   Document Work / Authoring
       LOCKED / OPERATOR-RATIFIED
       B04-F1 hybrid persistence                   CLOSED / OPERATOR-RATIFIED
       P8 / P9 / P10                               COMPLETE

B05   My Work / Work Queues
       LOCKED / OPERATOR-RATIFIED
       B05-F1 governance row recognition           CLOSED / OPERATOR-RATIFIED
       B05-F2 neutral governance ordering          SUPERSEDED BY F4
       B05-F3 governance Step deadline             CLOSED / OPERATOR-RATIFIED
       B05-F4 due-aware governance queue           CLOSED / OPERATOR-RATIFIED
       P8 R2                                       APPROVED / COMPLETE
       P9 / P10                                    COMPLETE

B06   Governance Case
       OPEN / ACTIVE
       entry recovery                              CLOSED
       B06-F1 case Step deadline projection        CLOSED / OPERATOR-RATIFIED
       P6                                          COMPLETE
       P7                                          CANDIDATE / AWAITING OPERATOR ADJUDICATION
       P8                                          NOT OPEN
B07   Document History                             NOT OPEN
B08   Notifications Full Inbox                     NOT OPEN
B09   Audit                                        NOT OPEN
B10   Organization Administration                 NOT OPEN
B11   Access Administration                       NOT OPEN
B12   Document Governance Administration           NOT OPEN
```

## Locked global IA preserved

```text
Início       = current operational situation
Minha Caixa  = assigned work
  Para aprovação
  Em edição
Documentos   = official document truth / creation
Gestão       = system configuration
Evidência    = audit/evidence
```

Notifications remains transversal utility chrome, not `Minha Caixa` authority.

## B05 locked references

```text
docs/work/current/t11-b05-my-work-r1.md
docs/work/current/t11-b05-my-work-functional-wireframe.html
docs/work/current/t11-b05-screen-contract.md
docs/work/current/t11-b05-pattern-consolidation.md
```

Locked governance deadline law:

```text
GovernanceRouteStep.due_in_days?
→ immutable attempt snapshot
→ activation freezes activated_at + persisted due_at
→ 1 day = 24 elapsed hours
→ no hidden default
→ overdue has no automatic lifecycle effect
```

Locked governance queue law:

```text
WorkGovernanceItem.due_at?

default order
  due_at ASC NULLS LAST
  document.code ASC
  governance_attempt_id ASC

optional first-page deadline_filter
  overdue | next_24h | next_7d | no_deadline
  omitted = all

relative filters
  one server-trusted first-page anchor
  cursor authenticates/reuses that anchor
  fresh first-page request gets a fresh anchor
```

No manual priority state, generic sort/filter DSL, business-calendar buckets or client-side global re-sort is introduced.

B05 P9 proves 19/19 material regions/controls with operations 54/55 only and zero invented write/API authority. P10 graduates no new shared pattern; B05 queue/deadline/cursor semantics remain local.

## B06 current entry authority

Canonical current work record:

```text
docs/work/current/t11-b06-governance-case-r1.md
```

B06 owns the exact Governance Case lens only:

```text
/work/governance/:attempt_id
→ getGovernanceAttempt
→ exact immutable governed Submission / obsolescence subject
→ ordered Steps + bounded decisions/feedback
→ live server-derived allowed_actions
```

B06-F1 is now operator-ratified:

```text
GovernanceStepView pending
  due_at forbidden

GovernanceStepView active | decided
  timed Step   -> exact persisted due_at present
  untimed Step -> due_at absent
```

Binding authority:

```text
docs/decisions/governance-case-step-deadline-read.md
```

The deadline remains attention/context truth only. B06 owns no deadline mutation, SLA engine, lifecycle breach effect, manual priority or frontend Authorization.

Current P7 leading candidate:

```text
Content-first Governance Workspace
  exact governed content dominant
  + B06-local governance rail
    subject summary
    ordered Steps / active deadline / prior Decisions
    governance feedback
    deliberate ACCEPT / RETURN_FOR_CHANGES decision zone
```

The dossier-first separate-viewer structure is not leading; the three-column workflow cockpit is rejected. P7 remains CANDIDATE until operator adjudication. No P8 HTML has been created.

## Exact next action

```text
1. Operator adjudicates B06 P7 H1 Content-first Governance Workspace.
2. If approved, create B06 functional P8 HTML with deterministic local fixtures/state simulation.
3. Operate / iterate the same B06 block until explicit operator LOCK.
4. Only after LOCK execute B06 P9 Screen Contract then P10 pattern consolidation.
5. Do not design or open B07+ early.
```

## Hard stops

```text
no Product code/schema/OpenAPI implementation/runtime/deploy work
no T12 work
no production frontend framework in P8
no static storyboard accepted as P8 lock evidence
no framework/library redefines Product semantics
no generic EventBus/broker/Redis without a named material trigger
no frontend Authorization matrix
no unopened downstream block design
no legacy restoration by sunk cost
no merge authorization implied
```

## T11 / implementation gate

Implementation remains blocked until:

```text
T11 CLOSED / OPERATOR-RATIFIED
T12 CLOSED / OPERATOR-RATIFIED
Integrated Whole-R10 coherence = PASS
fresh independent challenge = converged
operator implementation authorization = explicit
```

## Reopen law

Accepted Product/R10/frontend LOCK decisions reopen only on material evidence under the DevelopmentConexus Engineering Method. Preference, sunk cost, framework availability or hypothetical scale are not reopen triggers.