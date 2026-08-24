# T11 — B09-F1 Audit Investigation Capability — Written Decision Candidate

> **Status:** CANDIDATE / IN-CHAT DESIGN OPERATOR-APPROVED / WRITTEN RATIFICATION PENDING.  
> **Block:** B09 — Audit.  
> **Finding:** B09-F1 — Audit query/evidence capability.  
> **Method:** Frontend Product Experience Planning Method v2.3 + DevelopmentConexus Engineering Method.  
> **Upstream working evidence:** `t11-b09-audit-upstream-replan.md`.  
> **Implementation:** BLOCKED.  
> **P7/P8:** PAUSED/BLOCKED until this written candidate is operator-ratified and the bounded authority rebaseline is complete.

## 1. Decision objective

Close the pre-implementation B09-F1 finding with the smallest globally coherent Product/API/wire capability that makes Audit genuinely investigable without turning it into current-state authority, a generic search platform, a generic entity directory, or a screen-shaped API.

The current baseline remains:

```text
GET /api/v1/audit/events
operationId listAuditEvents
AuditEventPage
cursor + limit only
occurred_at DESC,event_id DESC
audit.read
```

That baseline is insufficient for the already operator-approved Launch jobs:

```text
A. point investigation / exact evidence question
B. period + authorized historical-scope review
```

External evidence export remains `DEFERRED` until a named auditor/certification/regulatory/customer handoff proves a concrete package/format/integrity requirement.

## 2. Global-Maximum selection

### Selected

```text
preserve and refine op78 listAuditEvents
+ add exactly three purpose-built safe reads for Audit query construction
+ keep current Audit evidence model immutable and PII-minimized
+ compose optional current human recognition outside evidence authority
```

### Rejected / deferred alternatives

```text
raw chronological feed only
  REJECTED — fails the ratified Auditor jobs at production scale

all query-assist data embedded into every op78 page
  REJECTED — couples evidence traversal to selector discovery and repeats unrelated data

one generic /audit/query-candidates?kind=... endpoint
  REJECTED — creates a conditional mini reference-data platform with heterogeneous semantics

admin User/Area/resource directories as selector infrastructure
  REJECTED — wrong disclosure/authority boundary and fails historical-evidence cases

generic entity/reference-data resolver
  REJECTED — no cross-product consumer justifies the abstraction

manual UUID entry as normal workflow
  REJECTED — evidentially pure but operationally poor

free-text Audit search
  DEFERRED — no single unambiguous searchable Audit corpus is currently proven

saved searches
  DEFERRED

custom sort
  REJECTED for Launch

analytics/dashboard
  REJECTED as B09 responsibility

export
  DEFERRED
```

## 3. Binding Audit invariants preserved

This decision does **not** reopen:

```text
AuditEvent = semantic action evidence, not current business state
Audit is append-only evidence, not event sourcing
Audit != Document History
historical visibility is snapshotted at action time
current grants determine read authorization
historical visibility determines event inclusion
current relocation/rename does not rewrite historical Audit visibility
Audit remains PII-minimized
free-form governed reasons/comments/content are not copied into Audit by convenience
current owner labels never become retroactive event-time facts
```

Current immutable evidence remains:

```text
event_id
occurred_at
actor USER(user_id) | SYSTEM
operation_code
resource_kind
resource_id
historical visibility COMPANY | AREA(area_id)
bounded typed facts where required
```

## 4. Exact application-operation delta

The candidate operation surface is:

```text
78  GET /api/v1/audit/events
    listAuditEvents
    REFINED — same semantic operation, structured query + inspection projection

87  GET /api/v1/audit/query-areas
    listAuditQueryAreas
    NEW SAFE READ

88  GET /api/v1/audit/query-actors
    searchAuditQueryActors
    NEW SAFE READ

89  GET /api/v1/audit/query-resources
    searchAuditQueryResources
    NEW SAFE READ
```

Resulting census if ratified:

```text
application operations           86 → 89
stable SPA routes                11 unchanged
PermissionCode values            16 unchanged
Idempotency-Key creations        11 unchanged
ETag read / mutation domains     13 / 13 unchanged
exact-byte resources             4 unchanged
semantic owners                  4 business + 2 supporting unchanged
new writes                       0
```

No new route, PermissionCode, semantic owner, mutation, ETag domain, idempotency command or byte resource is introduced.

## 5. op78 — `listAuditEvents` refinement

### 5.1 Route / order / permission

```text
GET /api/v1/audit/events
SAFE_READ
audit.read
order = occurred_at DESC,event_id DESC
```

The operation remains the sole Audit evidence traversal authority.

### 5.2 First-page structured query

When `cursor` is absent, admitted query members are:

```text
occurred_at_from?     UtcInstant   inclusive lower bound
occurred_at_before?   UtcInstant   exclusive upper bound

actor_kind?           user | system
actor_user_id?        Uuid

operation_codes?      unique AuditOperationCode set

resource_kind?        AuditResourceKind
resource_id?          Uuid

visibility_area_id?   Uuid

limit?                1..100, default 20
```

Semantic validation:

```text
both time bounds present
  occurred_at_from < occurred_at_before
  otherwise 400 request.invalid

actor_kind=user
  actor_user_id required

actor_kind=system
  actor_user_id forbidden

actor_kind absent
  actor_user_id forbidden

resource_id present
  resource_kind required

operation_codes
  closed enum only
  duplicates rejected
```

No arbitrary Launch maximum is placed on the requested time interval absent evidence that one is needed.

`operation_codes` is one logical set-valued filter; final OAS encoding must preserve uniqueness and canonical normalization for cursor binding. No generic query DSL is introduced.

### 5.3 Filter composition

```text
time + actor + action + resource + historical Area
  combine by AND across dimensions

multiple operation_codes
  combine by OR inside the action dimension
```

`visibility_area_id=X` means historically Area-attributed events for `X`; it never converts Company-wide events into Area events and never grants access.

Unknown/unseen actor/resource/Area identities do not create a disclosure oracle: after normal `audit.read` admission they simply cannot produce unauthorized evidence.

### 5.4 Authorization / visibility / pagination order

Binding evaluation order:

```text
Organization current subject
→ Authorization.AuthorizedScopes(audit.read)
→ Audit historical visibility admission
→ structured evidence predicates
→ occurred_at DESC,event_id DESC
→ seek position
→ page selection
→ optional recognition composition
```

Recognition never participates in event admission, equality, ordering, cursor position or historical evidence.

### 5.5 Continuation law

Current stateless seek-cursor law is preserved:

```text
first page
  structured filters + optional limit

continuation
  cursor + optional limit only

cursor + repeated first-page filter
  400 request.invalid
```

Cursor authenticates:

```text
operationId
+ normalized structured predicates
+ canonical ordering
+ seek position
```

Current `audit.read` authority is re-evaluated on every page. No frozen multi-page snapshot, offset or total count is added.

## 6. Inspection projection — evidence separated from recognition

`AuditEventView` remains the closed immutable evidence union.

The list response becomes an inspection projection rather than contaminating `AuditEventView` with mutable labels:

```text
AuditRecognitionLabel {
  stable_label?: ShortText
  current_label?: ShortText
}

AuditEventRecognition {
  actor?: AuditRecognitionLabel
  visibility_area?: AuditRecognitionLabel
  resource?: AuditRecognitionLabel
}

AuditInspectionItem {
  evidence: AuditEventView
  recognition?: AuditEventRecognition
}

AuditInspectionPage {
  items: AuditInspectionItem[]
  page: Page
}
```

Semantic law:

```text
stable_label
  only an already-authoritative immutable human identifier
  examples: immutable Document code, immutable Area code

current_label
  optional current presentation enrichment
  never event-time evidence
  absent when erased, unavailable or not disclosure-safe
```

Examples:

```text
actor USER
  evidence.user_id always preserved
  current_label may be current display_name
  erased/missing profile → current_label absent

Document resource
  stable_label may be immutable Document code

Area visibility/resource
  stable_label may be immutable Area code
  current_label may be current Area name
```

Email, provider identity, roles/permissions, free-form governed content and admin profile data are not recognition payload.

No `GET /audit/events/{event_id}` detail operation is added. The loaded `AuditInspectionItem.evidence` already contains the exact event/facts needed by the B09 detail surface.

## 7. op87 — `listAuditQueryAreas`

```text
GET /api/v1/audit/query-areas
SAFE_READ
audit.read
no query parameters
complete bounded projection
order: stable Area code ASC,area_id ASC
```

Purpose:

> Populate historical-Area narrowing without requiring Organization administration access or opaque UUID knowledge.

Candidate law:

```text
current audit.read authority
+ historically visible Audit evidence
→ Area ids that may construct a useful Audit query
```

Response concept:

```text
AuditQueryAreaOption {
  area_id: Uuid
  code: CodeToken          // immutable stable recognition
  current_name?: ShortText // current non-historical recognition
}

AuditQueryAreaOptionsView {
  items: AuditQueryAreaOption[]
}
```

No event counts, analytics, admin lifecycle fields or permission details are exposed.

The default Audit scope remains “all events currently readable by me”; choosing an Area narrows op78 by stable `area_id`.

## 8. op88 — `searchAuditQueryActors`

```text
GET /api/v1/audit/query-actors?query=<SearchQuery>
SAFE_READ
audit.read
query required
max 20 items
no pagination
```

Purpose:

> Resolve a human actor selection to stable `user_id` without exposing a general User directory.

Candidate law:

```text
USER candidate
  must have at least one AuditEvent potentially visible under current caller Audit authority

SYSTEM
  remains a closed UI option; no synthetic system profile/directory row is needed
```

Response concept:

```text
AuditQueryActorOption {
  user_id: Uuid
  current_display_name?: ShortText
}

AuditQueryActorOptionsView {
  items: AuditQueryActorOption[] // max 20
}
```

Matching is bounded selection search, not Audit full-text search:

```text
exact user_id when query parses as UUID
OR current admissible display_name match when present

no email search
no role/permission search
no fuzzy/stem/vector semantics
```

Suggested deterministic ranking for the executable contract:

```text
exact user_id
→ display-name prefix
→ display-name contains
→ user_id
```

A profile-erased actor remains filterable through same-actor investigation from an event and through exact `user_id` when known; erasure never rewrites Audit evidence.

## 9. op89 — `searchAuditQueryResources`

```text
GET /api/v1/audit/query-resources
  ?resource_kind=<AuditResourceKind>
  &query=<SearchQuery>
SAFE_READ
audit.read
resource_kind required
query required
max 20 items
no pagination
```

Purpose:

> Resolve a human resource selection to exact `resource_kind + resource_id` without creating a universal entity search.

Candidate law:

```text
resource identity
  must appear in Audit evidence potentially visible under current caller Audit authority

searchable recognition
  exact UUID always admissible
  stable human identifier where already owned/immutable
  current label only where a bounded owner fact/search can safely provide it
```

Response concept:

```text
AuditQueryResourceOption {
  resource_kind: AuditResourceKind
  resource_id: Uuid
  stable_label?: ShortText
  current_label?: ShortText
}

AuditQueryResourceOptionsView {
  items: AuditQueryResourceOption[] // max 20
}
```

No promise exists that every resource kind has a searchable mutable name.

Examples:

```text
document
  exact UUID or immutable Document code
  current short recognition only when disclosure-safe

area
  exact UUID or immutable Area code
  current Area name when admissible

resource kinds with no accepted human identifier
  exact UUID remains the truthful search identity
```

Suggested deterministic ranking for the executable contract:

```text
exact resource_id
→ exact stable_label
→ stable_label prefix
→ current_label prefix
→ stable/current label contains
→ resource_id
```

No generic `/entities`, `/reference-data`, global resolver or cross-owner search platform is introduced.

## 10. Query Assist is guidance, never authorization

The three selector reads are server-authored purpose-built projections.

```text
Query Assist candidate appears
!=
caller is authorized to receive any future event/resource truth without recheck
```

Actual `listAuditEvents` independently re-evaluates current `audit.read`, historical visibility and exact predicates.

The browser must not derive selector completeness from already-loaded Audit pages.

Application may compose bounded owner facts/recognition after Audit establishes admissible identities; semantic owners retain their own User/Area/Document/resource truth.

## 11. Owner-lens cross-links — no wire link authority

Cross-link policy remains Product/frontend composition, not Audit evidence schema.

Universal next action:

```text
same actor
same resource
same action
→ structured Audit query
```

Admitted secondary handoffs:

```text
resource_kind=document + resource_id=document_id
  → /documents/:document_id

release / revision cancellation / obsolescence facts with document_id
  → /documents/:document_id/history

governance Decision facts with governance_attempt_id
  → /work/governance/:governance_attempt_id
```

Every destination rechecks current owner AuthZ/disclosure.

Exact User/Area/Group/RoleAssignment/DocumentType admin deep-links remain deferred until B10/B11/B12 lock their own detail-route semantics.

No URLs, `links[]`, owner-route objects or generic resource resolver are added to Audit wire data.

## 12. Internal read-composition impact

Existing T8-C law remains the base:

```text
Organization current subject
→ Authorization.AuthorizedScopes(audit.read)
→ application maps current scopes to Audit.ReadVisibility
→ Audit applies historical visibility before pagination
```

Bounded refinement required if this candidate is ratified:

```text
Audit.ListEvents
  accepts the structured evidence predicates above
  preserves historical visibility + cursor authority

Audit Query Assist
  Audit identifies admissible historical identities
  application may obtain bounded current recognition facts from the semantic owner
  application intersects/composes; never re-owns owner truth

op78 recognition
  composed only after event page selection
  recognition failure/absence must not rewrite, reorder or suppress admitted evidence
```

No generic ServiceLocator, cross-owner repository, shared semantic model or generic query service is created.

## 13. Product/API supersession boundary after written ratification

If operator-ratified as written, the durable bounded decision must supersede only conflicting current-tense clauses concerning:

```text
docs/product/journeys.md
  Audit inspection = raw cursor traversal only
  no structured Audit narrowing / Query Assist

docs/architecture/interfaces.md
  Audit read contract lacks the ratified structured predicates / Query Assist composition

docs/architecture/wire-contract.md
  op78 cursor+limit-only clause
  AuditEventPage as the complete B09 inspection projection
  operation 79+ historical closure statements where superseded by current census law

docs/architecture/frontend.md
  Launch Audit inspection/paging-only clause
  no filters absent from old op78

docs/decisions/api-operation-census.md
  86-operation count
```

Unrelated Product/T1→T10 authority remains untouched.

After durable promotion, the numeric census authority becomes:

```text
operations                  89
Idempotency-Key creations   11
ETag read/mutation domains  13 / 13
exact-byte resources        4
```

## 14. Bounded frontend rebaseline impact

After durable Product/API/wire ratification:

```text
FP0
  update only Audit flow/coverage/surface mappings affected by B09-F1

B01-B08
  preserve LOCK unless contradicted by new material evidence

B09 P7
  resume with structured query + Query Assist + recognition + cross-link authority

B09 P8
  remains blocked until P7 has no unresolved upstream finding
```

No B10+ work is opened by this decision.

## 15. Self-review checklist for written ratification

The written candidate must satisfy all before promotion:

```text
[ ] every new capability maps to ratified Auditor Launch job A or B
[ ] op78 remains one Audit evidence traversal authority
[ ] no browser post-filter over incomplete pages
[ ] recognition is structurally separated from immutable evidence
[ ] mutable labels never become filter identity or cursor authority
[ ] profile erasure preserves historical evidence
[ ] Query Assist exposes only Audit-relevant bounded candidates
[ ] admin directories are not required selector infrastructure
[ ] no generic entity/reference-data/search platform
[ ] no new Audit detail endpoint without a separate need
[ ] owner cross-links use existing admitted identities only
[ ] destination AuthZ/disclosure remains independent
[ ] B10-B12 route details remain unopened
[ ] operation delta is exactly +3 safe reads
[ ] routes/permissions/writes/idempotency/ETag/byte/owner censuses unchanged except operations 86→89
[ ] export/full-text/saved-search/custom-sort/analytics dispositions remain explicit
[ ] B01-B08 remain preserved by bounded-rebaseline law
[ ] implementation remains blocked
```

## 16. Exact gate

```text
NOW
  operator reviews this written candidate

IF RATIFIED AS WRITTEN
  promote one bounded durable B09 Audit investigation decision
  update sole API numeric census 86→89
  apply bounded Product/T8-C/T8-E/T8-F supersession/rebaseline
  close B09-F1
  invoke planning workflow for B09 P7/P8 realization

NOT YET
  no production implementation
  no P8 HTML
  no B10+
  no merge
```
