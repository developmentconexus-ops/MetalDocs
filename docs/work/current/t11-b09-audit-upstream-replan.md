# T11 — B09-F1 Audit Upstream Capability Replan

> **Status:** OPEN / ACTIVE / UPSTREAM FINDING / AUDITOR JOBS + STRUCTURED QUERY DIRECTION OPERATOR-RATIFIED.  
> **Block:** B09 — Audit.  
> **Method:** Frontend Product Experience Planning Method v2.3.  
> **Trigger:** frontend/reference evidence exposed a material mismatch between the Auditor job and the current `listAuditEvents` query surface.  
> **P8:** BLOCKED until B09-F1 is fully adjudicated and any required upstream authority is ratified.  
> **Implementation:** BLOCKED.

## 1. Why this finding exists

Current Audit application wire:

```text
GET /api/v1/audit/events
operationId: listAuditEvents
response: AuditEventPage
order: occurred_at DESC, event_id DESC
query: cursor + limit only
permission: audit.read
```

Frontend/reference investigation showed that a raw reverse-chronological feed is insufficient for the ratified Auditor jobs at realistic production scale.

Frontend Method v2.3 therefore treats the current wire as a falsifiable baseline, not an immutable UX ceiling.

Forbidden shortcuts remain:

```text
op78 lacks capability X
→ omit X from the experience

reference product has capability X
→ invent endpoint X
```

B09-F1 must prove the human job, choose the Global Maximum, and reopen only the smallest owning Product/backend/wire authority.

## 2. Audit invariants preserved

Unless explicitly reopened by evidence:

```text
AuditEvent = semantic action evidence, not current business state
Audit is append-only evidence, not event sourcing
Audit does not reconstruct current Document lifecycle
Document History remains separate Controlled Documents semantic history
historical visibility is snapshotted at action time
current grants determine read authorization
historical visibility determines event inclusion
historical visibility filtering occurs before pagination
current relocation/rename does not rewrite historical Audit visibility
Audit remains PII-minimized
free-form governed reasons/comments/content are not copied into Audit by convenience
```

Current event evidence includes:

```text
event_id
occurred_at
actor: USER(user_id) | SYSTEM
operation_code
resource_kind
resource_id
historical visibility: COMPANY | AREA(area_id)
bounded typed facts where operation/resource identity is insufficient
```

The current Launch wire has a closed Audit operation-code union and typed facts for the event families that require them.

## 3. B09-F1 target

Determine the smallest globally coherent Audit inspection/query capability that lets an Auditor / Governance Viewer answer real evidence questions efficiently without:

```text
turning Audit into current-state authority
merging Audit with Document History
copying mutable current labels as false historical truth
post-filtering incomplete cursor pages in the browser
creating a generic analytics/search platform without a consumer
```

## 4. Auditor Launch jobs — OPERATOR-RATIFIED

### 4.1 Point investigation — LAUNCH CORE

The Auditor must be able to answer:

```text
what happened to this Document / User / access decision / configuration?
who performed the action?
when did it happen?
what exact semantic action was recorded?
what bounded evidence facts belong to that event?
```

Audit remains evidence. Navigation to a current owner lens is a cross-link only and never means Audit reconstructs current state.

### 4.2 Period + scope review — LAUNCH CORE

The Auditor must be able to answer:

```text
what relevant actions occurred during this period?
what occurred within the historically visible Company / Area scope available to me?
which actors/actions/resources account for the evidence set I am reviewing?
```

The result must remain complete under server-side current `audit.read` authorization + historical-visibility filtering + cursor pagination.

### 4.3 External evidence export — DEFERRED

Export is not Launch Core merely because mature products expose it.

A future reopen requires a named consumer such as:

```text
external auditor / certification body / regulator / customer
+ required evidence package or machine-readable handoff
+ explicit scope / format / integrity expectations
```

Until then there is no commitment to CSV/PDF/JSON export, export jobs, governed export artifacts, signatures, hashes or export-specific Audit semantics.

This is Product-scope YAGNI, not a backend limitation.

## 5. Reference evidence direction

Reference products converge on one principle: Audit must be **investigable**, not merely scrollable.

```text
GitHub Audit Log
  structured narrowing by time, actor, action and resource context

Microsoft Purview Audit
  time, activity, user and object narrowing for large evidence sets

Qualio Audit Trail
  date, user, action and document metadata investigation in a controlled-quality context
```

MetalDocs deliberately takes the task pattern, not their full feature breadth.

## 6. Structured Audit Query — OPERATOR-RATIFIED

### 6.1 Launch query dimensions

The Global-Maximum/YAGNI direction is a **structured server-side Audit query**, preserving one canonical reverse-chronological result stream.

Launch must support evidence-backed narrowing by:

```text
time
  occurred_at bounded interval

actor
  exact stable USER identity
  OR SYSTEM

action
  one or more closed AuditOperationCode values

resource
  resource_kind
  + exact resource_id when known

historical scope
  default = every event currently readable by the caller
  optional explicit narrowing to a historically attributed Area that is within the caller's current audit.read authority

order
  occurred_at DESC
  event_id DESC
```

Filters are server-side and combine conjunctively across dimensions. Multi-valued action selection is an OR within the action dimension.

The query result must continue to apply:

```text
current audit.read scope authority
→ historical Audit visibility admission
→ structured query predicates
→ canonical ordering
→ cursor pagination
```

No browser-side filtering over incomplete pages may masquerade as complete Audit truth.

### 6.2 Cursor law direction

The existing stateless cursor model remains the preferred architecture:

```text
first traversal
  admitted structured filters + optional limit

continuation
  cursor + optional limit only

cursor authenticates
  operation identity
  + normalized query predicates
  + canonical ordering
  + seek position
```

The final exact parameter vocabulary and wire encoding are intentionally not frozen until B09-F1 human-recognition semantics are resolved.

### 6.3 Full-text and generic query platform dispositions

```text
free-text search over Audit      DEFERRED
query DSL                        REJECTED for Launch
saved searches                   DEFERRED
custom sort                      REJECTED for Launch
analytics/dashboard              REJECTED as B09 responsibility
export                           DEFERRED by §4.3
```

The reason to defer free-text is Product semantics, not backend absence. MetalDocs does not yet have one unambiguous searchable Audit corpus whose meaning justifies full-text infrastructure.

Questions such as these are deliberately unresolved rather than guessed:

```text
should q search current actor name or historical actor name?
current Document title or title-at-event?
translated action label or operation_code?
resource identifiers?
typed facts?
text intentionally excluded from Audit evidence?
```

Structured query satisfies the ratified Launch investigation jobs with higher evidentiary precision and lower semantic ambiguity.

## 7. Remaining capability dispositions

Already ratified:

```text
chronological inspection baseline        PRESENT-IN-AUTHORITY
exact event detail / typed facts         PRESENT-IN-AUTHORITY conceptually; UX projection still to close
time-range narrowing                     RATIFY
operation/action filtering               RATIFY
actor filtering                          RATIFY
resource-kind filtering                  RATIFY
exact resource-id lookup                 RATIFY
historical Area narrowing                RATIFY within current audit.read authority
free-text search                         DEFER
sort alternatives                        REJECT
export                                   DEFER
```

Still blocking B09-F1:

```text
human-recognition enrichment / historical-label semantics
owner-lens cross-link policy
exact op78 query/read projection refinement after those semantics close
bounded Product/API/wire ratification
bounded FP0/B09 rebaseline
```

## 8. Known non-solutions

```text
browser-side filtering over already-loaded Audit pages
client reconstruction of current state from event sequences
client AuthZ/historical-visibility matrix
arbitrary generic query DSL
full-text engine before a searchable corpus and job are proven
copying current User/Profile/Document names into immutable Audit semantics without historical-truth analysis
using Audit as Document History
```

## 9. Current gate

```text
Frontend Method v2.3                         OPERATOR-RATIFIED
B01-B08                                      LOCKED / OPERATOR-RATIFIED
B09                                          OPEN / ACTIVE
B09-F1 upstream Audit capability replan      OPEN / BLOCKING FINDING
Auditor point investigation                  LAUNCH CORE / OPERATOR-RATIFIED
Auditor period + scope review                LAUNCH CORE / OPERATOR-RATIFIED
Audit export                                 DEFERRED / OPERATOR-RATIFIED
Structured Audit Query                       OPERATOR-RATIFIED
Full-text generic Audit search               DEFERRED
Human recognition / historical labels       NEXT DECISION
B09 P7                                       PAUSED pending B09-F1
B09 P8                                       BLOCKED pending B09-F1
B10-B12                                      NOT OPEN
Product implementation                       BLOCKED
```

Next step:

> Adjudicate how actors, Areas and resources become human-recognizable without rewriting historical evidence, leaking current-state data, or forcing auditors to work from opaque UUIDs. Only after B09-F1 closes may the exact op78 refinement be ratified and B09 P7/P8 resume.
