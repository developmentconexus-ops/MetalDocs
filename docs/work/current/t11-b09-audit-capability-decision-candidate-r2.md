# T11 — B09-F1 Audit Investigation Capability — Written Decision Candidate R2

> **Status:** CANDIDATE / IN-CHAT DESIGN OPERATOR-APPROVED / WRITTEN RATIFICATION PENDING.  
> **Supersedes as review candidate:** `t11-b09-audit-capability-decision-candidate.md` (R1 remains work evidence only).  
> **Block:** B09 — Audit.  
> **Finding:** B09-F1 — Audit query/evidence capability.  
> **Method:** Frontend Product Experience Planning Method v2.3 + DevelopmentConexus Engineering Method.  
> **Implementation:** BLOCKED.  
> **P7/P8:** PAUSED/BLOCKED until written ratification + bounded authority rebaseline.

## 1. Ratification package

R2 freezes the exact pre-implementation decision already approved section-by-section in chat:

```text
Launch jobs
  A. point investigation / exact evidence question
  B. period + authorized historical-scope review

Audit traversal
  preserve op78 listAuditEvents
  add structured server-side narrowing
  preserve occurred_at DESC,event_id DESC
  preserve stateless authenticated seek cursor

Human recognition
  immutable evidence stays authoritative
  mutable labels are optional current/non-historical enrichment

Audit Query Assist
  purpose-built selector reads
  no admin-directory dependency
  no generic entity/reference-data platform

Owner handoff
  Audit-native investigation is universal
  bounded secondary links only from already-admitted identities

YAGNI
  full-text Audit search DEFERRED
  export DEFERRED
  saved searches DEFERRED
  custom sort REJECTED for Launch
  analytics/dashboard REJECTED as B09 responsibility
```

## 2. Exact operation delta

```text
78  GET /api/v1/audit/events
    listAuditEvents
    REFINED

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

If ratified:

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

No Audit detail endpoint is added.

## 3. op78 exact first-page query

```text
GET /api/v1/audit/events
SAFE_READ
audit.read
```

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

### 3.1 Time law

```text
occurred_at_from  = inclusive
occurred_at_before = exclusive

both present
  occurred_at_from < occurred_at_before
  else 400 request.invalid
```

No arbitrary Launch maximum interval is imposed.

### 3.2 Actor law

```text
actor_kind=user
  actor_user_id REQUIRED

actor_kind=system
  actor_user_id FORBIDDEN

actor_kind absent
  actor_user_id FORBIDDEN
```

### 3.3 Action-set wire law

`operation_codes` is one optional OAS array query member:

```text
style=form
explode=false
uniqueItems=true
minItems=1
maxItems=37
items=AuditOperationCode
```

Wire example:

```text
?operation_codes=governance.accepted,governance.returned_for_changes
```

Duplicates or unknown enum values → `400 request.invalid`.

For cursor authority, the server canonicalizes the accepted set into the closed `AuditOperationCode` enum order before cursor binding. User selection order is never semantic.

### 3.4 Resource law

```text
resource_kind alone
  valid; narrows by kind

resource_id present
  resource_kind REQUIRED

resource_id without resource_kind
  400 request.invalid
```

### 3.5 Historical-Area law

```text
visibility_area_id=X
  exact historical visibility.kind=area + area_id=X
```

It never includes Company-attributed events merely because the caller has Company-wide `audit.read`.

Launch does **not** add a separate “Company-only historical visibility” filter: current proven jobs require all-readable review plus optional Area narrowing; a Company-only slice remains unproven YAGNI.

An arbitrary actor/resource/Area UUID does not create an existence/access oracle. Normal current `audit.read` + historical visibility admission remains authoritative; identities outside it produce no unauthorized evidence.

### 3.6 Predicate composition

```text
time
AND actor
AND operation-code set
AND resource
AND historical Area
```

Multiple `operation_codes` are OR inside that one dimension.

## 4. op78 authorization, paging and response law

Binding evaluation:

```text
Organization current subject
→ Authorization.AuthorizedScopes(audit.read)
→ Audit historical visibility admission
→ structured predicates
→ occurred_at DESC,event_id DESC
→ cursor seek
→ page selection
→ optional human-recognition composition
```

Continuation preserves the global list law:

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

Current `audit.read` is rechecked every page. No offset, total count, generic sort or frozen multi-page snapshot.

## 5. Evidence / recognition response split

`AuditEventView` stays the immutable closed evidence union.

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

Binding semantics:

```text
stable_label
  accepted immutable human identifier only
  e.g. Document code, Area code

current_label
  optional current presentation context
  never event-time evidence
  absent when erased, unavailable or not disclosure-safe
```

Recognition:

```text
never filters events
never changes ordering
never enters cursor authority
never establishes current lifecycle state
never grants current owner access
```

After page selection, failure/absence of optional recognition does not rewrite, reorder or suppress admitted Audit evidence.

Closed disclosure posture:

```text
actor current recognition
  User display_name only; never email/provider/roles/permissions

Area current recognition
  current Area name; immutable Area code may be stable_label

Document resource recognition
  immutable Document code may be stable_label

other resources
  current/stable label only when an already-owned bounded owner fact safely provides it
  otherwise label absent and stable UUID remains truthful
```

No historical display-name/email/name snapshot is created by this decision.

## 6. op87 — Area Query Assist

```text
GET /api/v1/audit/query-areas
operationId listAuditQueryAreas
SAFE_READ
audit.read
no query parameters
complete bounded response
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

Order:

```text
code ASC,area_id ASC
```

No counts, analytics, lifecycle/admin fields or permission details.

## 7. op88 — Actor Query Assist

```text
GET /api/v1/audit/query-actors?query=<SearchQuery>
operationId searchAuditQueryActors
SAFE_READ
audit.read
query REQUIRED
maxItems=20
no pagination
```

Public candidate law:

```text
USER id has at least one historically visible AuditEvent
under current caller Audit authority
```

SYSTEM is a closed frontend option and is not returned as a synthetic User row.

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
  exact user_id match only

otherwise
  Unicode case-fold current display_name
  diacritics preserved
  no accent folding
  no stemming/fuzzy/vector matching
```

Deterministic ranking:

```text
exact display_name
→ display_name prefix
→ display_name contains
→ user_id
```

The application may page an internal Organization-owned bounded recognition search as needed and intersect candidate User ids against Audit historical visibility; it must not stop at an arbitrary first owner page if that would make the public max-20 result incorrect.

No email, role, permission or admin-profile search.

Profile erasure leaves the historical `user_id` intact. Such an actor remains directly filterable through “same actor” from visible evidence and through exact UUID when known.

## 8. op89 — Resource Query Assist

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

Public candidate law:

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
  exact resource_id match only

otherwise
  match only owner-provided recognition admitted for this Audit purpose:
    stable_label where available
    current_label where available

  Unicode case-fold
  diacritics preserved
  no accent folding
  no stemming/fuzzy/vector matching
```

Deterministic ranking:

```text
exact stable_label
→ exact current_label
→ stable_label prefix
→ current_label prefix
→ stable_label/current_label contains
→ resource_id
```

The application may page a bounded owner-owned recognition search as required and intersects owner candidates with Audit-visible resource identities before disclosure. It may not expose an owner candidate that has no admissible Audit evidence.

Closed examples:

```text
document
  exact UUID or immutable Document code

area
  exact UUID or immutable Area code; current Area name may assist recognition

user/user_profile
  exact UUID; current display_name may assist recognition

resource kind with no accepted bounded human label
  exact UUID only
```

This does not create a universal resource-name promise.

## 9. Query Assist authorization law

```text
Query Assist result
  = query-construction guidance only
```

Actual op78 always independently rechecks:

```text
current session
current audit.read
historical visibility
exact structured predicates
```

The browser never derives selector completeness from loaded Audit pages.

Admin directories are not required public infrastructure for these controls.

## 10. Owner-lens handoff policy

Universal action stays inside Audit:

```text
same actor
same resource
same action
```

Secondary links may be composed only from already-admitted identities + already-accepted stable routes:

```text
Document resource
  → /documents/:document_id

release / revision cancellation / obsolescence with admitted document_id
  → /documents/:document_id/history

governance Decision with admitted governance_attempt_id
  → /work/governance/:governance_attempt_id
```

Destination rechecks current AuthZ/disclosure. Historical Audit visibility never grants destination access.

B10/B11/B12 admin resource deep-links remain deferred until those blocks define their own route/detail semantics.

No URL, `links[]`, owner-route object or generic deep-link resolver enters Audit wire data.

## 11. Internal composition boundary

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
  can answer whether candidate stable identities occur in admitted Audit evidence

Organization / Controlled Documents / other semantic owner
  owns current recognition/search facts for its entities

application
  composes/intersects bounded facts
  never becomes owner of User/Area/Document/current-state truth
```

No shared semantic model, generic query service, ServiceLocator, cross-owner repository or generic search engine.

## 12. Exact supersession boundary after written ratification

Promote one bounded durable decision that supersedes only conflicting current-tense clauses in:

```text
docs/product/journeys.md
  raw Audit inspection-only assumption

docs/architecture/interfaces.md
  Audit read contract without structured predicates / Query Assist composition

docs/architecture/wire-contract.md
  op78 cursor+limit-only query
  AuditEventPage as complete B09 inspection projection

docs/architecture/frontend.md
  Audit inspection/paging-only + no-filter clause

docs/decisions/api-operation-census.md
  86-operation numeric count
```

Unrelated Product/T1→T10 authority remains unchanged.

Durable post-ratification census:

```text
operations                  89
Idempotency-Key creations   11
ETag read/mutation domains  13 / 13
exact-byte resources        4
```

## 13. Bounded frontend rebaseline

After durable promotion:

```text
FP0
  rebaseline Audit-only flows/coverage/surface mappings

B01-B08
  preserve LOCK unless new evidence specifically falsifies one

B09-F1
  close after authority/census rebaseline proves consistency

B09 P7
  resume

B09 P8
  remains blocked until P7 exits with no unresolved upstream finding

B10+
  remain NOT OPEN
```

## 14. R2 self-review — PASS

R2 was checked against the ratified Product jobs, T3 Audit laws, T8-C read-composition law, T8-E cursor law, current B09 findings and the sole numeric census posture.

```text
PASS  every new capability maps to Launch job A or B
PASS  op78 remains sole Audit evidence traversal
PASS  historical visibility remains pre-pagination
PASS  browser post-filter cannot impersonate complete truth
PASS  evidence and mutable recognition are structurally separate
PASS  mutable labels never become filter/cursor identity
PASS  profile erasure preserves event evidence
PASS  action-set query encoding/canonicalization is explicit
PASS  Query Assist search semantics/ranking are deterministic
PASS  Query Assist public results are Audit-visible-only
PASS  no admin directory is required public selector infrastructure
PASS  no generic entity/reference-data/search platform
PASS  no new Audit detail endpoint
PASS  cross-links use admitted identities only
PASS  destination current AuthZ/disclosure is independent
PASS  B10-B12 detail routes remain unopened
PASS  exact delta = +3 SAFE_READ operations
PASS  operation census 86→89; all other stated censuses unchanged
PASS  full-text/export/saved-search/custom-sort/analytics dispositions explicit
PASS  B01-B08 preserved by bounded-rebaseline law
PASS  implementation and P8 remain blocked
```

No unresolved placeholder, `suggested` wire behavior or screen-shaped convenience operation remains in R2.

## 15. Exact gate

```text
NOW
  operator reviews R2 as written

IF OPERATOR-RATIFIED
  create/promote one durable bounded Audit investigation decision
  update sole API census 86→89
  apply bounded Product/T8-C/T8-E/T8-F supersession/rebaseline
  close B09-F1 after consistency proof
  invoke planning workflow for B09 P7/P8

NOT YET
  no Product implementation
  no P8 HTML
  no B10+
  no merge
```
