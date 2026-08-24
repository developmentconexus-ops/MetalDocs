# T11 — B09-F1 Audit Upstream Capability Replan

> **Status:** OPEN / ACTIVE / UPSTREAM FINDING / AUDITOR LAUNCH JOBS RATIFIED.  
> **Block:** B09 — Audit.  
> **Method:** Frontend Product Experience Planning Method v2.3.  
> **Trigger:** frontend/reference evidence exposed a potentially material mismatch between the auditor job and the current `listAuditEvents` query surface.  
> **P8:** BLOCKED until this finding is ratified, rejected or deferred.  
> **Implementation:** BLOCKED.

## 1. Why this finding exists

B09 initially recovered the current Audit application wire:

```text
GET /api/v1/audit/events
operationId: listAuditEvents
response: AuditEventPage
order: occurred_at DESC, event_id DESC
query: cursor + limit only
permission: audit.read
```

Reference study then surfaced mature audit-inspection patterns such as search, time-range narrowing, actor/action/resource filters, richer event detail and export.

Under Frontend Method v2.3, the current wire is a falsifiable baseline rather than an immutable UX ceiling. The correct response is therefore not:

```text
op78 lacks capability X
→ omit X from the experience
```

and not:

```text
reference product has capability X
→ invent endpoint X
```

The finding must first prove the auditor job, compare Global-Maximum alternatives, and then reopen only the smallest owning Product/backend/wire authority if evidence requires it.

## 2. Current accepted Audit invariants that remain in force

Until explicitly reopened by evidence, preserve:

```text
AuditEvent = semantic action evidence, not current business state
Audit is append-only evidence, not event sourcing
Audit does not reconstruct current Document lifecycle
Document History remains separate Controlled Documents semantic history
historical visibility is snapshotted at action time
current grants determine read authorization; historical visibility determines event inclusion
historical visibility filtering occurs before pagination
current relocation/rename does not rewrite historical Audit visibility
Audit is PII-minimized
free-form governed reasons/comments/content are not copied into Audit by convenience
```

Current event projection includes stable evidence such as:

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

Current Launch wire has a closed Audit operation-code union and typed fact schemas for specific event families.

## 3. B09-F1 question

Determine the smallest globally coherent Audit inspection/query capability that lets the Auditor / Governance Viewer efficiently answer real evidence questions at expected production scale without:

```text
turning Audit into current-state authority
merging Audit with Document History
copying mutable current profile/resource labels as false historical truth
post-filtering incomplete cursor pages in the browser
creating a generic analytics/search platform without a consumer
inventing export merely because competitors expose it
```

## 4. Candidate capability space to investigate

Each item must be dispositioned independently as `PRESENT-IN-AUTHORITY`, `RATIFY`, `REJECT`, or `DEFER`:

```text
A. chronological inspection baseline
B. exact event detail / typed facts presentation
C. time-range narrowing
D. operation/action filtering
E. actor filtering
F. resource-kind filtering
G. exact resource-id lookup
H. historical Area/Company visibility narrowing where useful
I. free-text search, if a real searchable evidence corpus exists
J. sort alternatives, if canonical reverse chronology is insufficient
K. cross-links to owner lenses without making Audit a resolver
L. export, only if a named auditor/compliance portability job is proven
M. human-recognition enrichment, with historical-truth/PII rules made explicit
```

This list remains an investigation boundary, not a feature checklist.

## 5. Required Global-Maximum analysis

Before any backend/API change or P8 HTML:

```text
1. define the concrete auditor jobs/questions
2. establish expected event volume and inspection frequency
3. study mature audit products by task pattern
4. distinguish query/navigation convenience from real semantic truth
5. test each candidate capability against least privilege and historical visibility
6. decide whether one read operation can remain coherent or whether a separate capability is semantically justified
7. define pagination/query semantics that remain complete under authorization
8. decide whether human labels are historical snapshots, current enrichment, or intentionally absent
9. prove export need separately from on-screen inspection
10. compare YAGNI / complexity / implementation cost against user value
11. choose the Global-Maximum Product contract
12. reopen only the smallest Product/API/wire/frontend authorities affected
13. perform bounded FP0/B09 rebaseline
14. only then resume P7 and P8
```

## 6. Known non-solutions

```text
browser-side filtering over already-loaded Audit pages
client reconstruction of current state from event sequences
client AuthZ/historical-visibility matrix
adding arbitrary generic query DSL
adding a full-text engine before proving searchable text and scale
copying current User/Profile/Document names into immutable Audit semantics without historical-truth analysis
using Audit as Document History
```

## 7. Auditor Launch jobs — OPERATOR-RATIFIED

The first B09-F1 scope decision is ratified:

```text
A. Point investigation / exact evidence question     LAUNCH CORE
B. Period + scope review                              LAUNCH CORE
C. External evidence export                           DEFERRED
```

### 7.1 Launch Core job A — point investigation

The Auditor must be able to answer questions such as:

```text
what happened to this Document / User / access decision / configuration?
who performed the action?
when did it happen?
what exact semantic action was recorded?
what bounded evidence facts belong to that event?
```

Audit remains evidence. Any navigation to a current owner lens is a cross-link and never means Audit reconstructs current business state.

### 7.2 Launch Core job B — period / scope review

The Auditor must be able to answer bounded review questions such as:

```text
what relevant actions occurred during this period?
what occurred within the historically visible Company / Area scope available to me?
which actors/actions/resources account for the evidence set I am reviewing?
```

The capability must remain complete under server-side historical-visibility filtering and cursor pagination. Browser filtering over an already-loaded subset is not acceptable evidence truth.

### 7.3 Export — DEFERRED by YAGNI

Export is not Launch Core merely because mature products expose it.

A future export reopen requires a named consumer such as:

```text
external auditor / certification body / regulator / customer
+ required evidence package or machine-readable handoff
+ explicit scope/format/integrity expectations
```

Until then, MetalDocs does not pre-commit to CSV/PDF/JSON export, export jobs, governed export artifacts, signatures, hashes or export-specific Audit semantics.

This is a Product-scope deferral, not a backend limitation.

## 8. Reference evidence direction after job ratification

Current reference evidence supports the Launch Core emphasis on investigation rather than raw chronological traversal:

```text
GitHub Audit Log
  structured qualifiers for time, actor, action/operation and resource context
  demonstrates that investigation is a first-class Audit job

Microsoft Purview Audit
  date/time, activities, users and object scope
  demonstrates explicit narrowing for large evidence sets

Qualio Audit Trail
  simple/complex search across date, user, action and document metadata
  demonstrates the same investigation need in a controlled-quality-document context
```

The reference conclusion is intentionally narrower than their feature sets:

> MetalDocs Audit must be efficiently investigable by real evidence questions. It does not need a generic audit analytics/query language merely because mature platforms have one.

## 9. Current gate

```text
Frontend Method v2.3                         OPERATOR-RATIFIED
B01-B08                                      LOCKED / OPERATOR-RATIFIED
B09                                          OPEN / ACTIVE
B09-F1 upstream Audit capability replan      OPEN / BLOCKING FINDING
Auditor point investigation                  LAUNCH CORE / OPERATOR-RATIFIED
Auditor period + scope review                LAUNCH CORE / OPERATOR-RATIFIED
Audit export                                 DEFERRED / OPERATOR-RATIFIED
B09 query dimensions                         NEXT DECISION
B09 P7                                       PAUSED pending B09-F1
B09 P8                                       BLOCKED pending B09-F1
B10-B12                                      NOT OPEN
Product implementation                       BLOCKED
```

Next step:

> Adjudicate the smallest structured query and human-recognition surface required to satisfy the ratified Launch Core jobs. Do not choose a final wire/API shape or build the B09 functional HTML until B09-F1 is fully adjudicated and any required upstream authority is ratified.
