---
id: audit-investigation-read
kind: authority
owner: architecture
summary: Bounded T11 authority for investigable Audit reads, structured query, human recognition, Query Assist and owner-lens handoff.
---

# Audit Investigation Read — bounded T11 authority

> **Status:** OPERATOR-RATIFIED / BOUNDED T11 REOPEN.  
> **Ratified:** 2026-08-24.  
> **Method:** Frontend Product Experience Planning Method v2.3 + DevelopmentConexus Engineering Method.  
> **Implementation:** BLOCKED by `../roadmap.md`.

## 1. Authority and supersession

This page is the single bounded current authority for the B09 Audit investigation capability discovered during T11 frontend planning.

It does not replace Product/T1→T10 wholesale. It supersedes only conflicting current-tense clauses concerning Audit read/query realization in:

```text
../product/journeys.md
../architecture/interfaces.md
../architecture/wire-contract.md
../architecture/frontend.md
api-operation-census.md
```

All unchanged Audit write/evidence, historical-visibility, Authorization, lifecycle, persistence, runtime and Product laws remain current.

When this page conflicts with an older current-tense Audit clause on the bounded subject, this page wins. Historical stage snapshots remain truthful for the stage at which they were ratified.

## 2. Product jobs and Global-Maximum boundary

Launch Audit must support:

```text
A. point investigation / exact evidence question
B. period + authorized historical-scope review
```

Audit must be investigable, not merely scrollable.

Explicit Launch dispositions:

```text
structured Audit query            RATIFIED
human recognition                 RATIFIED
Audit Query Assist                RATIFIED
owner-lens cross-links            RATIFIED
full-text Audit search            DEFERRED
external evidence export          DEFERRED
saved searches                    DEFERRED
custom sort                       REJECTED for Launch
analytics/dashboard               REJECTED as B09 responsibility
query DSL                         REJECTED for Launch
```

Export may reopen only on a named auditor/certification/regulatory/customer handoff with explicit package/format/integrity requirements.

## 3. Audit invariants preserved

This decision does not reopen:

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
current mutable labels never become retroactive event-time facts
```

Immutable evidence remains:

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

```text
78  GET /api/v1/audit/events
    listAuditEvents
    REFINED — same semantic operation

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

Current census after this bounded reopen:

```text
application operations           89
stable SPA routes                11
PermissionCode values            16
Idempotency-Key creations        11
ETag read / mutation domains     13 / 13
exact-byte resources             4
semantic owners                  4 business + 2 supporting
new writes                       0
```

No Audit detail endpoint is added. Operation 90+ requires unchanged semantic normalization already permitted by current authority or a new lawful bounded Product/T6 reopen.

## 5. op78 — structured Audit traversal

### 5.1 Route, permission and order

```text
GET /api/v1/audit/events
operationId listAuditEvents
SAFE_READ
permission audit.read
order occurred_at DESC,event_id DESC
```

Operation 78 remains the sole Audit evidence traversal authority.

### 5.2 First-page query

When `cursor` is absent:

```text
occurred_at_from?     UtcInstant
occurred_at_before?   UtcInstant
actor_kind?           user | system
actor_user_id?        Uuid
operation_codes?      AuditOperationCode[]
resource_kind?        AuditResourceKind
resource_id?          Uuid
visibility_area_id?   Uuid
limit?                integer 1..100; default 20
```

Time law:

```text
occurred_at_from   inclusive
occurred_at_before exclusive

both present
  occurred_at_from < occurred_at_before
  otherwise 400 request.invalid
```

No arbitrary Launch maximum time interval is imposed.

Actor law:

```text
actor_kind=user   -> actor_user_id REQUIRED
actor_kind=system -> actor_user_id FORBIDDEN
actor_kind absent -> actor_user_id FORBIDDEN
```

Resource law:

```text
resource_kind alone is valid
resource_id requires resource_kind
resource_id without resource_kind -> 400 request.invalid
```

Historical Area law:

```text
visibility_area_id=X
  -> exact historical visibility.kind=area + area_id=X
```

It never includes Company-attributed events merely because the caller has Company-wide `audit.read`. Launch does not add a separate Company-only historical-visibility filter absent a proven job.

Filters combine by AND across dimensions. Multiple action codes are OR inside the action dimension.

### 5.3 Action-set wire/cursor law

`operation_codes` is one optional OAS array query member:

```text
style=form
explode=false
uniqueItems=true
minItems=1
maxItems=37
items=AuditOperationCode
```

Example:

```text
?operation_codes=governance.accepted,governance.returned_for_changes
```

Unknown/duplicate values -> `400 request.invalid`.

For cursor authority, the accepted set is canonicalized into the closed `AuditOperationCode` enum order. User selection order is not semantic.

### 5.4 Authorization, visibility and pagination

Binding evaluation order:

```text
Organization current subject
→ Authorization.AuthorizedScopes(audit.read)
→ Audit historical visibility admission
→ structured evidence predicates
→ occurred_at DESC,event_id DESC
→ cursor seek
→ page selection
→ optional human-recognition composition
```

Continuation preserves the global stateless seek-cursor law:

```text
first page
  admitted filters + optional limit

next page
  cursor + optional limit only

cursor + any repeated initial filter
  400 request.invalid
```

Cursor authenticates:

```text
operationId
+ canonical normalized predicates
+ occurred_at DESC,event_id DESC
+ seek position
```

Current `audit.read` is rechecked on every page. No offset, total count, generic sort or frozen multi-page snapshot is added.

Arbitrary actor/resource/Area UUID input never creates an existence/access oracle; current `audit.read` + historical visibility remains authoritative.

## 6. Evidence and human recognition are structurally distinct

`AuditEventView` remains the immutable closed evidence union.

The inspection projection is:

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

current_label
  optional current presentation context
  never event-time evidence
  absent when erased, unavailable or not disclosure-safe
```

Recognition never filters events, changes ordering, enters cursor authority, establishes current lifecycle state or grants current owner access.

Closed recognition posture:

```text
actor
  current User display_name only; never email/provider/roles/permissions

Area
  immutable Area code may be stable_label
  current Area name may be current_label

Document
  immutable Document code may be stable_label

other resources
  stable/current label only when an already-owned bounded fact safely supplies it
  otherwise the stable UUID remains truthful
```

No historical display-name/email/name snapshot is created by this decision. Recognition failure after page selection never rewrites, reorders or suppresses admitted Audit evidence.

## 7. op87 — Area Query Assist

```text
GET /api/v1/audit/query-areas
operationId listAuditQueryAreas
SAFE_READ
audit.read
no query parameters
complete bounded response
order code ASC,area_id ASC
```

Candidate set:

```text
Area ids occurring in historically visible Audit evidence
∩ current caller audit.read authority
```

Response:

```text
AuditQueryAreaOption {
  area_id: Uuid
  code: CodeToken
  current_name?: ShortText
}

AuditQueryAreaOptionsView {
  items: AuditQueryAreaOption[]
}
```

No counts, analytics, lifecycle/admin fields or permission details are exposed. Default Audit scope remains all events currently readable by the caller.

## 8. op88 — Actor Query Assist

```text
GET /api/v1/audit/query-actors?query=<SearchQuery>
operationId searchAuditQueryActors
SAFE_READ
audit.read
query REQUIRED
maxItems=20
no pagination
```

Candidate law:

```text
USER id has at least one historically visible AuditEvent
under current caller Audit authority
```

SYSTEM is a closed frontend option, not a synthetic User row.

Response:

```text
AuditQueryActorOption {
  user_id: Uuid
  current_display_name?: ShortText
}

AuditQueryActorOptionsView {
  items: AuditQueryActorOption[] // max 20
}
```

Search semantics:

```text
if normalized query parses as UUID
  exact user_id only
otherwise
  Unicode case-fold current display_name
  diacritics preserved
  no accent folding
  no stemming/fuzzy/vector matching
```

Ranking:

```text
exact display_name
→ display_name prefix
→ display_name contains
→ user_id
```

Application may page an internal Organization-owned bounded recognition search and intersect with Audit-visible actor ids; it may not stop at an arbitrary first owner page if that would make the public max-20 result incorrect.

No email, role, permission or admin-profile search. Profile erasure preserves historical `user_id` and same-actor investigation.

## 9. op89 — Resource Query Assist

```text
GET /api/v1/audit/query-resources
  ?resource_kind=<AuditResourceKind>
  &query=<SearchQuery>
operationId searchAuditQueryResources
SAFE_READ
audit.read
resource_kind REQUIRED
query REQUIRED
maxItems=20
no pagination
```

Candidate law:

```text
resource_kind + resource_id occurs in historically visible Audit evidence
under current caller Audit authority
```

Response:

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

Search semantics:

```text
if normalized query parses as UUID
  exact resource_id only
otherwise
  match only owner-provided stable_label/current_label admitted for this purpose
  Unicode case-fold
  diacritics preserved
  no accent folding
  no stemming/fuzzy/vector matching
```

Ranking:

```text
exact stable_label
→ exact current_label
→ stable_label prefix
→ current_label prefix
→ stable_label/current_label contains
→ resource_id
```

Application may page bounded owner-owned recognition search and intersect with Audit-visible identities before disclosure. It may never expose an owner candidate with no admissible Audit evidence.

Closed examples:

```text
document      exact UUID or immutable Document code
area          exact UUID or immutable Area code; current Area name may assist
user/profile  exact UUID; current display_name may assist
no accepted label -> exact UUID only
```

No universal resource-name promise or generic entity search is introduced.

## 10. Query Assist is guidance, never Authorization

```text
Query Assist result != authorization
```

Actual op78 independently rechecks current session, current `audit.read`, historical visibility and exact stable predicates.

Query Assist is server-authored. The browser never derives candidate completeness by collecting already-loaded Audit pages. Administrative directories are not required public infrastructure for these selectors.

Filter identity is always stable IDs/enums; mutable labels never become canonical query identity.

## 11. Owner-lens handoff

Universal follow-up remains Audit-native:

```text
same actor
same resource
same action
→ structured Audit query
```

Secondary handoffs are permitted only from already-admitted identities to already-accepted stable routes:

```text
Document resource
  → /documents/:document_id

release / revision cancellation / obsolescence with admitted document_id
  → /documents/:document_id/history

governance Decision with admitted governance_attempt_id
  → /work/governance/:governance_attempt_id
```

Every destination rechecks its own current Authorization/disclosure. Historical Audit visibility never grants destination access.

User/Area/Group/RoleAssignment/DocumentType admin deep-links remain deferred until B10/B11/B12 lock their own detail-route semantics.

No URL, `links[]`, owner-route object or generic deep-link/resource resolver enters Audit wire data. No new Audit facts are added solely to fabricate navigation.

## 12. Internal composition boundary

Existing law remains:

```text
Organization current subject
→ Authorization.AuthorizedScopes(audit.read)
→ application maps to Audit.ReadVisibility
→ Audit historical visibility BEFORE pagination
```

Bounded refinement:

```text
Audit
  owns evidence predicates
  owns historical-visibility filtering
  owns seek position/page
  can answer whether stable identities occur in admitted Audit evidence

Organization / Controlled Documents / other semantic owners
  own current recognition/search facts for their entities

application
  composes/intersects bounded facts
  never re-owns User/Area/Document/current-state truth
```

No shared semantic model, generic query service, ServiceLocator, cross-owner repository or generic search engine is introduced.

## 13. Frontend realization boundary

Stable route remains:

```text
/audit
```

B09 frontend consumes:

```text
78 listAuditEvents
87 listAuditQueryAreas
88 searchAuditQueryActors
89 searchAuditQueryResources
```

Audit presentation must preserve:

```text
server-owned complete filtering before pagination
immutable evidence vs optional current recognition distinction
Audit-native same-actor/resource/action investigation
bounded owner handoffs only under §11
no frontend Authorization matrix
no client-side post-filter over incomplete pages
```

P8 functional HTML remains blocked until B09 P7 exits with no unresolved upstream finding.

## 14. Bounded reopen / preservation proof

Affected current authority only:

```text
Product/T6 Audit investigation job/read surface
T8-C Audit read/query composition
T8-E op78 + op87-op89 wire semantics
T8-F Audit operation coverage/realization
T6-API numeric census
```

Preserved unchanged:

```text
Audit persistence/write semantics
same-commit Audit evidence law
historical-visibility snapshots
PermissionCode vocabulary
semantic-owner topology
stable SPA route census
Document History semantics
B01-B08 locked Product Experience decisions
Product implementation gate
T12 gate
```
