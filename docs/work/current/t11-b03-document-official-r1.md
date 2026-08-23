# T11 — B03 Document Official R1 — Method v2.2 candidate

> **Status:** CURRENT FP1 BLOCK / CANDIDATE / NOT LOCKED.  
> **Block:** B03 — Document Official / Ficha + Viewer + stable-Document Discussion.  
> **Method:** Frontend Product Experience Planning Method v2.2.  
> **Global frontend map:** `t11-frontend-foundation-r1.md`.  
> **Current Product/architecture amendment:** `../../decisions/discussion-notifications-launch.md`.  
> **Predecessors:** B01 / B01N / B02 LOCKED.  
> **Implementation:** BLOCKED.

## 1. Previous evidence disposition

Preserved design learning:

```text
viewer-first whole-page B03              REJECTED — too collapsed
record/ficha-first direction             OPERATOR-APPROVED LEADING DIRECTION
sectioned vertical hierarchy A           OPERATOR-APPROVED LEADING DIRECTION
stable-Document Discussion on ficha      CURRENT Product authority
viewer deliberately separate             preserved
Notification -> exact message deep-link  current authority
```

The previous file:

```text
t11-b03-document-official-wireframe.html
```

is now:

```text
REJECTED — wrong representation medium
```

It remains evidence only. It was an HTML storyboard/static review board rather than the canonical functional P8 required by Method v2.2. Rejecting that artifact does **not** reject the approved hierarchy A.

## 2. User needs

B03 must let a person answer and act on:

```text
What controlled Document is this?
What is official now?
Is newer work underway without replacing official truth?
Who owns/responsibly maintains it?
How do I deliberately read exact official content?
How do I reach Work or History without confusing their truth with this ficha?
What are people discussing about this stable Document?
Was I mentioned, and can I return to the exact message?
What management actions are currently offered by the server?
```

## 3. Leading structure A — preserved

Canonical reading order:

```text
Shell inherited from B01/B01N
↓
Breadcrumb / return to Library
↓
Document hero
  code
  governed official title
  lifecycle status
  official Revision
  primary: Visualizar documento
  secondary: Download / History where admitted
↓
Current-context banner
  disclosed open Revision or active obsolescence context
↓
Ficha do documento
  Type
  Area
  Responsible owner
  official Revision
  released-at
  representation/source information
↓
Revisões — current context only
  official Revision
  disclosed open Revision
  link to B07 full History
↓
Gestão
  server-hinted current actions only
↓
Discussão do Documento
  stable-Document timeline
  reply reference
  @mention composer
```

Discussion remains intentionally below official/current-management truth. A Notification deep-link reveals/scrolls directly to the target message without creating a second B03 layout.

## 4. Distinct viewer surface

`Visualizar documento` changes the material viewing mode and therefore opens a distinct B03-owned read-only surface.

```text
Ficha
-> Visualizar documento
-> B03 read-only exact official-content viewer
-> Voltar para ficha
```

Viewer laws:

```text
no edit
no submit
no governance controls
explicit Document code + Revision outside rendered bytes
exact source vs official rendition remains labeled truth
no official Release -> no fake viewer / no DRAFT substitution
```

This is not B04 Document Work.

## 5. Discussion interaction laws represented in P8

P8 must behaviorally demonstrate:

```text
load stable-Document Discussion fixture
reply to one prior message
composer draft
@mention autocomplete
select Mention candidate by stable User identity
send simulated accepted message
Notification entry -> anchor_message_id -> exact message reveal/highlight
```

P8 simulation must preserve Product meaning:

```text
Discussion != WorkingContent comments
Discussion != SubmissionFeedback
Discussion != Governance feedback
Mention display text != identity authority
Notification != access grant
```

No production Lexical integration is required in P8. The low-fi composer may use plain browser controls while representing the accepted Text | Mention semantics truthfully.

## 6. Locked Notification chrome reuse

B03 reuses B01N; it does not redesign it.

The low-fi P8 must allow the operator to exercise the already-locked flow:

```text
bell
-> Quick Inbox
-> DOCUMENT_MENTION item
-> mark item seen/read in local fixture
-> close Quick Inbox
-> remain/navigate to exact B03 Document
-> reveal Discussion
-> highlight anchor message
```

Quick Inbox is not full B08 triage.

## 7. Downstream transition law

B03 may demonstrate destination intent without designing unopened blocks:

```text
Abrir trabalho
  -> transition stub: /documents/:document_id/work
  -> B04 NOT OPEN; no B04 baseline generated

Histórico
  -> transition stub: /documents/:document_id/history
  -> B07 NOT OPEN; no B07 baseline generated
```

This preserves the block progression law.

## 8. P7 lightweight data/action feasibility

Leading hypothesis A requires:

| Requirement | Source | Status |
|---|---|---|
| stable Document code/status/type/area/owner | `getDocument` / current read authority | PRESENT-IN-AUTHORITY |
| official Revision/title/released-at/content summary | `getDocument` + Release reads | PRESENT-IN-AUTHORITY |
| disclosed open Revision context | `getDocument.open_revision` | PRESENT-IN-AUTHORITY |
| active obsolescence context | current Document/request read | PRESENT-IN-AUTHORITY |
| Discussion timeline + anchor navigation | op79 | PRESENT-IN-AUTHORITY |
| create Discussion message/reply | op80 | PRESENT-IN-AUTHORITY |
| mention autocomplete | op81 | PRESENT-IN-AUTHORITY |
| Notification source navigation | ops82–86 + current authority | PRESENT-IN-AUTHORITY |
| management action visibility | `DocumentOfficialView.allowed_actions` | **FINDING B03-F1** |

No screen-shaped API is justified by the candidate.

## 9. B03-F1 — required before final LOCK

Candidate precision remains:

```text
DocumentOfficialAction =
  create_revision
  replace_responsible_owner
  create_obsolescence_request
  withdraw_obsolescence_request

DocumentOfficialView.allowed_actions: unique DocumentOfficialAction[]
```

Law:

```text
server-derived UX hints only
same canonical AuthZ/domain predicates or shared equivalent
every command rechecks current truth
no frontend permission matrix
no new operation / route / Permission / semantic owner
```

P8 may simulate a server-provided `allowed_actions` fixture to evaluate placement, but B03 cannot receive final operator LOCK until this precision is durably reconciled or proven unnecessary.

## 10. Canonical P8 contract under Method v2.2

Required medium:

```text
HTML
CSS
vanilla JavaScript
local deterministic fixtures
```

The artifact must be one browser-operable B03 prototype, not multiple static screens stacked as a review board.

Material local interactions must work:

```text
Notification Quick Inbox open/close
Notification -> Discussion anchor
Visualizar documento -> viewer -> Voltar
Discussion reply select/cancel
@mention candidate open/select
send local Discussion fixture
management disclosure/menu/drawer used by current candidate
responsive shell/drawer plausibility
```

Controls that merely cross into B04/B07 may resolve to explicit transition stubs rather than inventing those screens.

## 11. P8 review questions

Operator should operate the prototype and judge:

```text
Does the ficha feel like the Document record rather than a dashboard?
Is official truth visible before work/management/discussion?
Does current open work remain clearly separate from official truth?
Is Visualizar documento discoverable and does the separate viewer feel natural?
Does Notification -> exact message feel obvious and context-preserving?
Is Discussion too low, too high or appropriately secondary to official truth?
Does @mention interaction feel natural without turning B03 into chat?
Are management controls visible enough without dominating reader intent?
Does mobile/narrow behavior preserve the same semantic order?
```

## 12. Exit

```text
functional P8 operated by operator
-> findings iterated only inside B03
-> B03-F1 closed
-> operator-only LOCK
-> P9 exact Screen Contract + bidirectional trace
-> P10 bounded pattern consolidation
-> only then B04 may open normally
```
