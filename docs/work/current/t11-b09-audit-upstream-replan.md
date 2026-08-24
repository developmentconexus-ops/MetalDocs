# T11 — B09-F1 Audit Upstream Capability Replan

> **Status:** OPEN / ACTIVE / UPSTREAM FINDING / AUDITOR JOBS + STRUCTURED QUERY + HUMAN RECOGNITION + QUERY ASSIST + OWNER CROSS-LINK POLICY OPERATOR-RATIFIED.  
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

The final exact parameter vocabulary and wire encoding are intentionally not frozen until B09-F1 query-construction semantics are resolved.

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

Structured query satisfies the ratified Launch investigation jobs with higher evidentiary precision and lower semantic ambiguity.

## 7. Human recognition / historical-label semantics — OPERATOR-RATIFIED

### 7.1 Authority split

Immutable Audit evidence remains identity/fact authority:

```text
event_id
occurred_at
actor USER(user_id) | SYSTEM
operation_code
resource_kind
resource_id
historical visibility area_id when AREA
bounded typed facts
```

Human-readable labels are **presentation enrichment**, not retroactive historical evidence, unless the underlying human identifier is itself accepted immutable authority.

The screen must distinguish these semantics rather than silently presenting a mutable current label as if Audit had snapshotted it at event time.

### 7.2 Actor recognition

For `actor=USER(user_id)`:

```text
historical evidence authority
  user_id

optional current recognition
  current admissible UserReference/display_name when available

lawful profile erasure / missing enrichment
  preserve user_id
  render neutral "Nome atual indisponível" or equivalent
  never rewrite/delete the Audit event
```

`display_name` is never filter identity. A human selection resolves to exact `user_id` before the canonical Audit query.

For `actor=SYSTEM`, SYSTEM remains the evidence meaning; Launch does not expose an invented system display-profile model.

### 7.3 Area recognition

Historical visibility authority remains the snapshotted `area_id`.

A current mutable Area name may be shown only as current recognition context. A separately accepted immutable human identifier such as immutable Area code may be shown as stable recognition when current authority proves that immutability.

Current Area rename/retirement never rewrites historical visibility attribution.

### 7.4 Resource recognition

`resource_kind + resource_id` remains the evidence identity.

Where an owner already has a stable immutable human identifier, the Audit projection may compose it for recognition without making Audit its owner. Current mutable labels/titles may be composed only when disclosure-safe and explicitly presented as current recognition context.

Examples of the intended distinction:

```text
Document code
  may be stable recognition when its immutability is already Product authority

Revision title / User display name / mutable entity name
  current enrichment only unless a separate historical snapshot is already owned elsewhere
```

Audit does not create a universal historical-label snapshot merely for presentation.

### 7.5 Filter identity law

Human-facing controls may show names/codes, but canonical query identity remains stable IDs/enums:

```text
"Marina Costa"
  → actor_user_id = UUID

"PO-023"
  → resource_kind=document + resource_id=UUID

friendly action label
  → exact AuditOperationCode
```

No query semantics rely on mutable display-name equality.

### 7.6 Explicit YAGNI

Do not introduce by default:

```text
historical snapshot of display_name
historical snapshot of email
universal snapshot of Group/Area/resource mutable names
copy of current labels into immutable Audit persistence
generic entity resolver / reference-data platform
```

Only recognition data proven necessary for Audit investigation may be composed, and that composition never becomes lifecycle/current-state authority.

## 8. Audit Query Assist — OPERATOR-RATIFIED

### 8.1 Purpose-built query construction

Launch requires a bounded **Audit Query Assist** capability so an Auditor can construct the ratified structured query without knowing opaque UUIDs and without depending on Organization/Access administration directories.

The capability is Audit-query purpose-built. It is not a generic entity/reference-data platform and it does not transfer semantic ownership of User, Area, Document or other resources to Audit.

### 8.2 Action discovery

`AuditOperationCode` is already a closed vocabulary.

The frontend may group and humanize these codes for selection, but canonical filter identity remains the exact enum value. No server-side action directory is required merely to populate the action selector.

### 8.3 Historical Area discovery

The Area selector must be derived from the caller's current `audit.read` authority and historical Audit visibility, not from an unrestricted Organization-admin directory.

Human-facing Area recognition may use the semantics in §7.3, but canonical query identity remains `area_id`.

The default choice remains "all Audit events currently readable by me". Explicit Area narrowing never grants access beyond current `audit.read` authority.

### 8.4 Actor discovery

Actor selection is a bounded purpose-built typeahead over Audit-relevant actor identities.

Candidate law:

```text
candidate USER
  has at least one AuditEvent potentially visible
  under the caller's current Audit authority

candidate SYSTEM
  represented by the closed SYSTEM actor meaning
```

Human recognition may show current admissible `display_name` when available. Lawful erasure or missing profile enrichment leaves the stable `user_id` selectable with neutral recognition text.

Canonical filter identity remains exact `user_id` or SYSTEM. `display_name` equality never becomes query authority.

### 8.5 Resource discovery

Resource selection is closed by `resource_kind` first; MetalDocs does not create a universal global resource search.

For the selected kind, bounded candidate discovery is drawn from resource identities that appear in Audit evidence visible under the caller's Audit authority.

Canonical identity is always:

```text
resource_kind + resource_id
```

Stable immutable human identifiers may be shown when already authoritative; current mutable labels may be shown only under the §7 recognition/disclosure law.

When no safe human label exists, a compact kind + stable-id representation remains truthful.

### 8.6 Security / completeness law

Query Assist is server-authored. It must not leak actors, Areas or resources merely because they exist in current administrative directories.

```text
current audit.read authority
+ historical Audit visibility
→ admissible query-assist candidates
```

The browser never constructs completeness by collecting candidates from already-loaded Audit pages.

A selected human label resolves to stable query identity before `listAuditEvents` execution; changing or erasing the label later does not change the historical identity being filtered.

### 8.7 Explicit non-solutions

```text
/admin users/areas/documents as required Audit filter infrastructure
manual UUID entry as the normal workflow
global /reference-data or /entities search platform
client-side candidate derivation from loaded Audit pages
mutable display name as filter identity
Audit persistence copy of current labels merely to support selectors
```

The exact count and wire shape of any auxiliary read operation(s) remain deliberately unchosen until the remaining B09-F1 semantics close.

## 9. Owner-lens cross-link policy — OPERATOR-RATIFIED

### 9.1 Audit-native investigation remains universal

Every evidence item may continue investigation inside Audit using the already-ratified structured query:

```text
same actor
same resource
same action
```

This is the universal next-step capability. An owner-lens link is secondary and optional.

### 9.2 Owner handoff admission

A current owner-lens cross-link may be offered only when all are true:

```text
an accepted stable owner route already exists
the exact routing identity is already present in admitted Audit evidence/facts
the destination meaning is unambiguous
navigation does not imply current access or current-state truth
```

The destination always performs its own current Authorization/disclosure check. Historical Audit visibility never grants access to the current owner resource.

Audit does not persist URLs, invent a generic resource resolver or derive current domain state to construct navigation.

### 9.3 Current admitted cross-links

```text
Document resource event
  resource_kind=document + resource_id=document_id
  → /documents/:document_id
  → Document Official current-owner lens

Release / revision cancellation / obsolescence events
  typed facts already expose document_id
  → /documents/:document_id/history
  → Document History semantic lens

Governance Decision event
  GovernanceDecisionAuditFacts already expose governance_attempt_id
  → /work/governance/:governance_attempt_id
  → exact Governance Case lens
```

The link meaning is a handoff from historical action evidence to the appropriate owner lens; it never means Audit and that lens share authority.

### 9.4 Deferred admin-owned destinations

Exact deep-links for current User / Area / Group / RoleAssignment / DocumentType administration are **DEFERRED** until B10/B11/B12 establish and lock their own route/detail semantics.

B09 records the future seam but does not design unopened administration blocks or guess routes now.

### 9.5 Insufficient routing identity

When an Audit event does not already expose the exact identity required by an accepted owner route, Launch does not add new Audit facts merely for navigation by convenience.

Example:

```text
revision.created
  resource_id = revision_id
  no currently admitted document_id routing fact
  → no fabricated Document/History link
  → same-resource Audit investigation remains available
```

A later functional/proven user need may reopen the smallest owning read composition, but navigation convenience alone is insufficient.

### 9.6 Access drift / unavailable destination

```text
historical Audit event remains legitimately visible
+ current owner resource is absent or non-disclosable
→ Audit evidence remains visible and unchanged
→ destination owner returns its ordinary current non-disclosing result
→ Audit does not explain hidden current access/state
```

No failed owner handoff mutates or suppresses historical Audit evidence.

### 9.7 Explicit non-solutions

```text
every resource_kind automatically links somewhere
generic resource/deep-link resolver
URLs persisted inside AuditEvent
Audit-generated current-state projection
historical visibility treated as current owner access
new Audit facts added solely to fabricate navigation
B10-B12 route/detail design inside B09
```

## 10. Remaining capability dispositions

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
human-recognition semantics              RATIFY
Audit Query Assist                       RATIFY
owner-lens cross-link policy             RATIFY
```

Still blocking B09-F1:

```text
exact op78 structured-query/read projection refinement
smallest Audit Query Assist wire realization
bounded Product/API/wire ratification
bounded FP0/B09 rebaseline
```

## 11. Known non-solutions

```text
browser-side filtering over already-loaded Audit pages
client reconstruction of current state from event sequences
client AuthZ/historical-visibility matrix
arbitrary generic query DSL
full-text engine before a searchable corpus and job are proven
copying current labels into immutable Audit semantics as convenience
using admin-only directories as required Audit filter infrastructure
generic entity/reference-data platform for Audit selectors
generic resource/deep-link resolver
using Audit as Document History
```

## 12. Current gate

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
Human recognition / historical labels       OPERATOR-RATIFIED
Audit Query Assist                           OPERATOR-RATIFIED
Owner-lens cross-link policy                 OPERATOR-RATIFIED
Exact Product/API/wire refinement            NEXT DECISION
B09 P7                                       PAUSED pending B09-F1
B09 P8                                       BLOCKED pending B09-F1
B10-B12                                      NOT OPEN
Product implementation                       BLOCKED
```

Next step:

> Ratify the smallest exact `listAuditEvents` + Audit Query Assist Product/API/wire refinement that realizes the already-ratified B09-F1 semantics without generic search/reference infrastructure. Then perform the bounded B09 rebaseline and only afterward resume B09 P7/P8.
