# T11 — B03 Document Official R2 — Method v2.2

> **Status:** LOCKED / OPERATOR-RATIFIED.  
> **Block:** B03 — Document Official / Ficha + Viewer + stable-Document Discussion.  
> **Method:** Frontend Product Experience Planning Method v2.2.  
> **Global frontend map:** `t11-frontend-foundation-r1.md`.  
> **Current Product/architecture amendment:** `../../decisions/discussion-notifications-launch.md`.  
> **Document Official action precision:** `../../decisions/document-official-actions-read.md`.  
> **Predecessors:** B01 / B01N / B02 LOCKED.  
> **Implementation:** BLOCKED.

## 1. Operator adjudication

The operator explicitly approved:

```text
record/ficha-first B03 semantics
stable-Document Discussion on ficha
distinct read-only official viewer
Notification -> exact Discussion message deep-link
historical C — Two-column dossier visual direction
functional P8 R2 layout + interaction
B03-F1 DocumentOfficialView.allowed_actions precision
```

Therefore B03 is now `LOCKED / OPERATOR-RATIFIED`.

No downstream block is implicitly approved by this LOCK.

## 2. Locked composition

```text
Document hero
↓
TWO-COLUMN DOSSIER
  left
    current-work context
    ficha / classification / responsibility
    server-derived management affordances

  right
    official-content preview
    exact current official Revision / representation label
    deliberate Visualizar completo action
↓
Revisions context — full width
↓
Stable-Document Discussion — full width
```

Canonical P8 evidence:

```text
t11-b03-document-official-functional-wireframe.html
```

The preview is contextual recognition only. It never becomes exact-content authority and never substitutes DRAFT content for official truth.

## 3. Locked interaction laws

```text
hero / preview / current official Revision
-> Visualizar documento
-> distinct B03 read-only official viewer
-> Voltar para ficha

bell
-> Quick Inbox inherited from B01N
-> DOCUMENT_MENTION
-> same Document ficha
-> Discussion
-> anchor_message_id target reveal/highlight

Discussion
-> stable-Document chronological timeline
-> one-message reply reference
-> @mention autocomplete
-> Mention selected by stable User identity
-> local prototype send demonstrates accepted message semantics
```

B04/B07/B08 destinations remain explicit transition boundaries only; they are not designed by B03.

## 4. Locked truth hierarchy

```text
stable Document identity
-> current official truth
-> current work context without replacing official truth
-> classification / responsibility
-> bounded management affordances
-> current Revision context
-> stable-Document Discussion
```

Laws:

```text
DRAFT/SUBMITTED never becomes official truth
History never becomes current-state authority
preview/viewer never becomes semantic content authority
Discussion != WorkingContent comments
Discussion != SubmissionFeedback
Discussion != Governance feedback
Notification != access grant
```

## 5. B03-F1 — CLOSED / OPERATOR-RATIFIED

Current durable precision:

```text
../../decisions/document-official-actions-read.md
```

Accepted shape:

```text
DocumentOfficialAction =
  create_revision
  replace_responsible_owner
  create_obsolescence_request
  withdraw_obsolescence_request

DocumentOfficialView.allowed_actions: unique DocumentOfficialAction[]
```

Meaning:

```text
server-derived UX guidance only
same canonical predicates as corresponding commands
always present on disclosed DocumentOfficialView; [] valid
commands recheck current truth
no frontend Authorization evaluator
no denial-reason channel
```

Current system census remains unchanged:

```text
application operations             86
Idempotency-Key creations          11
ETag read / mutation domains       13 / 13
exact-byte resources               4
PermissionCode values              16
stable SPA routes                  11
semantic owners                    4 business + 2 supporting
```

## 6. Responsible-owner management

The existing precision remains current:

```text
../../decisions/responsible-owner-selection-read.md
```

`responsible_owner_candidates` provides bounded selection guidance when admitted; `getDocumentResponsibleOwner` + strong ETag remains the concurrency truth; `replaceDocumentResponsibleOwner` rechecks current target eligibility and Authorization.

## 7. Responsive/accessibility structure

Locked structural requirements:

```text
desktop
  two-column dossier

narrow
  single-column semantic reflow
  ficha/current context remains before Discussion
  preview remains recognizably official and opens same viewer

viewer
  explicit back control
  Document code + official Revision exposed outside rendered bytes

Discussion
  keyboard-reachable composer/reply/Mention path
  anchor highlight not color-only as the sole context signal
```

Exact production styling, component library, focus implementation and final visual design are not locked by B03.

## 8. Post-LOCK obligations

Per Method v2.2:

```text
P9  exact Screen Contract + bidirectional backend trace
P10 bounded pattern consolidation
```

Only after those close may B04 become the normal next block.

A later material finding reopens only the smallest affected B03 decision under the Method; preference or visual polish alone does not reopen this LOCK.
