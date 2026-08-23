# T11 — B03 Document Official R2 — Method v2.2 candidate

> **Status:** CURRENT FP1 BLOCK / CANDIDATE / NOT LOCKED.  
> **Block:** B03 — Document Official / Ficha + Viewer + stable-Document Discussion.  
> **Method:** Frontend Product Experience Planning Method v2.2.  
> **Global frontend map:** `t11-frontend-foundation-r1.md`.  
> **Current Product/architecture amendment:** `../../decisions/discussion-notifications-launch.md`.  
> **Predecessors:** B01 / B01N / B02 LOCKED.  
> **Implementation:** BLOCKED.

## 1. Evidence disposition

Preserved Product/UX decisions:

```text
viewer-first whole-page B03              REJECTED — too collapsed
record/ficha-first B03                   OPERATOR-APPROVED
stable-Document Discussion on ficha      CURRENT Product authority
viewer deliberately separate             OPERATOR-APPROVED
Notification -> exact message deep-link  CURRENT Product authority
```

Previous P8 storyboard:

```text
t11-b03-document-official-wireframe.html
REJECTED — wrong representation medium
```

Method v2.2 replaced static/storyboard P8 evidence with a browser-operable functional low-fi prototype.

## 2. P7 visual-hypothesis adjudication

The operator asked to recover the pre-Discussion B03 visual hypothesis that had an official-content preview on the same ficha.

Historical evidence in the prior B03 ledger named:

```text
C — Two-column dossier
left: identity / current Revision / responsibility / work context
right: official-content summary card + actions
lower: revisions / management / discussion
```

The operator reviewed a reconstructed C hypothesis updated with current Discussion/Notification semantics and **APPROVED** it as the visual/layout direction for P8.

Selected composition:

```text
Document hero
↓
TWO-COLUMN DOSSIER
  left
    current-work context
    ficha / metadata
    management actions

  right
    official-content preview card
    exact Revision / representation label
    deliberate Visualizar completo action
↓
Revisions context — full width
↓
Stable-Document Discussion — full width
```

The preview is contextual recognition only. It never becomes exact-content authority and never replaces the separate viewer.

## 3. User needs

B03 must let a person answer and act on:

```text
What controlled Document is this?
What is official now?
Is newer work underway without replacing official truth?
Who responsibly maintains it and where does it belong?
Can I recognize the official content without leaving the ficha?
How do I deliberately open the exact official viewer?
How do I reach Work or History without confusing their truth with this ficha?
What are people discussing about this stable Document?
Was I mentioned, and can I return to the exact message?
What management actions are currently offered by the server?
```

## 4. Canonical P8 R2

Current functional evidence:

```text
t11-b03-document-official-functional-wireframe.html
```

Medium:

```text
HTML
CSS
vanilla JavaScript
local deterministic fixtures
```

P8 R2 combines the selected historical C layout with the already-reviewed interactions:

```text
bell -> Quick Inbox
DOCUMENT_MENTION -> exact Discussion anchor
preview -> exact-content viewer -> back to ficha
hero/revision Visualizar -> same viewer
reply select/cancel
@mention autocomplete/select
send local Discussion fixture
responsible-owner drawer
server-hinted management disclosure/menu
responsive navigation/dossier reflow
explicit B04/B07/B08 transition stubs
```

The prototype is Evidence only; no local fixture becomes Product/server truth.

## 5. Official preview vs exact viewer

Preview laws:

```text
contextual recognition only
bound to current official Revision presentation
no DRAFT substitution
no edit/governance controls
click opens the same B03 exact-content viewer
responsive layout may move preview below ficha without changing meaning
```

Viewer laws:

```text
Ficha -> Visualizar documento / preview -> viewer -> Voltar para ficha
read-only
explicit Document code + official Revision outside rendered bytes
exact source vs official rendition remains labeled truth
no edit
no submit
no governance controls
no official Release -> no fake viewer
```

This remains distinct from B04 Document Work.

## 6. Discussion interaction laws represented in P8

```text
load stable-Document Discussion fixture
reply to one prior message
composer draft
@mention autocomplete
select Mention candidate by stable User identity
send simulated accepted message
Notification entry -> anchor_message_id -> exact message reveal/highlight
```

Semantic separations remain:

```text
Discussion != WorkingContent comments
Discussion != SubmissionFeedback
Discussion != Governance feedback
Mention display text != identity authority
Notification != access grant
```

No production Lexical integration is required in P8.

## 7. Locked Notification chrome reuse

B03 reuses B01N; it does not redesign it.

```text
bell
-> Quick Inbox
-> DOCUMENT_MENTION item
-> local seen/read simulation
-> exact B03 Document
-> Discussion anchor reveal
```

Quick Inbox is not full B08 triage.

## 8. Downstream transition law

```text
Abrir trabalho
  -> /documents/:document_id/work stub
  -> B04 NOT OPEN

Histórico
  -> /documents/:document_id/history stub
  -> B07 NOT OPEN

Ver todas as notificações
  -> /notifications stub
  -> B08 NOT OPEN
```

No unopened block is visually designed inside B03.

## 9. P7/P8 data feasibility

| Requirement | Source | Status |
|---|---|---|
| stable Document code/status/type/area/owner | `getDocument` / current read authority | PRESENT-IN-AUTHORITY |
| official Revision/title/released-at/content summary | `getDocument` + Release reads | PRESENT-IN-AUTHORITY |
| official preview input | exact current official content summary/representation | PRESENT-IN-AUTHORITY |
| disclosed open Revision context | `getDocument.open_revision` | PRESENT-IN-AUTHORITY |
| active obsolescence context | current Document/request read | PRESENT-IN-AUTHORITY |
| Discussion timeline + anchor navigation | op79 | PRESENT-IN-AUTHORITY |
| create Discussion message/reply | op80 | PRESENT-IN-AUTHORITY |
| mention autocomplete | op81 | PRESENT-IN-AUTHORITY |
| Notification source navigation | ops82–86 + current authority | PRESENT-IN-AUTHORITY |
| management action visibility | `DocumentOfficialView.allowed_actions` | **FINDING B03-F1** |

No screen-shaped API is justified by the selected layout.

## 10. B03-F1 — required before final LOCK

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

P8 may simulate `allowed_actions`; B03 cannot receive final operator LOCK until this precision is durably reconciled or proven unnecessary.

## 11. Current operator gate

P7 visual direction is approved. P8 R2 is now the only current visual/interaction candidate.

Operator should operate it and judge:

```text
Does the two-column dossier feel more modern without becoming a dashboard?
Does the official preview materially improve recognition/context?
Is the preview large enough without overpowering the ficha?
Does opening the full viewer feel like the natural next step?
Does current work remain clearly separate from official truth?
Are management controls appropriately secondary?
Do Revisions and Discussion work well at full width below the dossier?
Does Notification -> exact message remain context-preserving?
Does narrow/mobile reflow preserve the same reading order?
```

## 12. Exit

```text
P8 R2 operated by operator
-> findings iterated only inside B03
-> B03-F1 closed
-> operator-only B03 LOCK
-> P9 exact Screen Contract + bidirectional trace
-> P10 bounded pattern consolidation
-> only then B04 may open normally
```
