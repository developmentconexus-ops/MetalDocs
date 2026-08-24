# T11 — B09-F1 Audit Upstream Capability Replan

> **Status:** CLOSED / OPERATOR-RATIFIED / DURABLE AUTHORITY PROMOTED.  
> **Block:** B09 — Audit.  
> **Method:** Frontend Product Experience Planning Method v2.3.  
> **Durable authority:** `../../decisions/audit-investigation-read.md`.  
> **Numeric census:** 89 operations.  
> **Rebaseline proof:** `t11-b09-f1-rebaseline-proof.md`.  
> **P7:** RESUMED / NEXT.  
> **P8:** BLOCKED pending clean P7 exit.  
> **Implementation:** BLOCKED.

## 1. Why this finding existed

B09 initially recovered this accepted Audit baseline:

```text
GET /api/v1/audit/events
operationId: listAuditEvents
response: AuditEventPage
order: occurred_at DESC,event_id DESC
query: cursor + limit only
permission: audit.read
```

External/reference and frontend investigation showed that a raw reverse-chronological feed was insufficient for realistic Auditor work at production scale.

The initial wrong local-maximum inference was:

```text
op78 lacks search/filter capability
→ omit it from the experience
```

Frontend Method v2.3 corrected the process law:

```text
current Product/backend plan during pre-implementation
  = falsifiable baseline, not immutable UX ceiling

material user need + insufficient authority
  = blocking upstream FINDING
```

The opposite shortcut was also rejected:

```text
reference product has capability X
→ invent endpoint X
```

B09-F1 therefore reopened the smallest owning Product/read authority and adjudicated each capability by real human need, Global Maximum and YAGNI.

## 2. Audit invariants preserved throughout the reopen

```text
AuditEvent = semantic action evidence, not current business state
Audit is append-only evidence, not event sourcing
Audit != Document History
historical visibility is snapshotted at action time
current grants determine read authorization
historical visibility determines event inclusion
historical visibility filtering precedes pagination
current relocation/rename never rewrites historical visibility
Audit remains PII-minimized
free-form governed reasons/comments/content are not copied by convenience
```

Immutable evidence identity remained:

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

## 3. Reference-study evidence retained

The P6 study used mature Audit products as task-pattern evidence rather than feature checklists:

```text
GitHub Audit Log
  structured narrowing by time, actor, action and resource context

Microsoft Purview Audit
  time, activity, user and object narrowing for large evidence sets

Veeva Vault
  structured audit trail/event inspection patterns

Qualio Audit Trail
  date, user, action and document-oriented investigation in a controlled-quality context
```

The bounded conclusion was:

> MetalDocs Audit must be efficiently investigable by real evidence questions; it does not need a generic audit analytics/query platform merely because mature products expose one.

## 4. Ratified decision sequence

### 4.1 Auditor Launch jobs

Operator-ratified Launch Core:

```text
A. point investigation / exact evidence question
B. period + authorized historical-scope review
```

External evidence export was `DEFERRED` until a named auditor/certification/regulatory/customer handoff proves a concrete evidence package/format/integrity requirement.

### 4.2 Structured Audit Query

Operator-ratified query dimensions:

```text
occurred_at interval
exact USER actor identity or SYSTEM
one-or-more AuditOperationCode values
resource_kind
exact resource_id when known
optional historical Area narrowing within current audit.read authority
fixed occurred_at DESC,event_id DESC
```

Server-side predicates combine by AND across dimensions; multiple operation codes are OR inside the action dimension. Browser filtering over incomplete pages is never evidence truth.

Explicit YAGNI:

```text
free-text Audit search   DEFERRED
query DSL                REJECTED
saved searches           DEFERRED
custom sort              REJECTED
analytics/dashboard      REJECTED as B09 responsibility
export                   DEFERRED
```

### 4.3 Human recognition / historical labels

Operator-ratified authority split:

```text
immutable IDs/facts
  = historical Audit evidence authority

current mutable labels
  = optional non-historical presentation enrichment

accepted immutable human identifiers
  = stable recognition when already owned as immutable truth
```

Examples:

```text
User display_name
  current enrichment only

Area code
  stable recognition because Product authority makes code immutable

Document code
  stable recognition because Product authority makes code immutable
```

Filter identity remains IDs/enums, never mutable display-name equality. Lawful profile erasure preserves the historical user_id and event.

### 4.4 Audit Query Assist

Operator-ratified purpose-built query construction:

```text
Action
  closed AuditOperationCode vocabulary; frontend may humanize/group

Area
  candidates bounded by current audit.read + historically visible Audit evidence

Actor
  bounded typeahead over Audit-visible USER identities + optional current display name
  SYSTEM remains a closed actor option

Resource
  resource_kind first
  candidates must occur in Audit-visible evidence
  stable/current recognition only where safely owned
```

Rejected:

```text
admin directories as required Audit selector infrastructure
manual UUID entry as normal workflow
generic /reference-data or /entities platform
client-derived candidate completeness
mutable label as canonical filter identity
```

### 4.5 Owner-lens cross-links

Operator-ratified universal follow-up remains Audit-native:

```text
same actor
same resource
same action
```

Secondary handoffs are allowed only when an already-accepted route and exact admitted routing identity exist:

```text
Document resource
  → /documents/:document_id

release / revision cancellation / obsolescence with admitted document_id
  → /documents/:document_id/history

governance Decision with admitted governance_attempt_id
  → /work/governance/:governance_attempt_id
```

Every destination rechecks current AuthZ/disclosure. B10/B11/B12 admin deep-links remain deferred until those blocks define their own detail-route semantics.

## 5. Final ratified Product/API/wire package

The written R2 package was operator-ratified and promoted into `../../decisions/audit-investigation-read.md`.

Exact operation surface:

```text
78  GET /api/v1/audit/events
    listAuditEvents
    REFINED

87  GET /api/v1/audit/query-areas
    listAuditQueryAreas
    SAFE_READ

88  GET /api/v1/audit/query-actors
    searchAuditQueryActors
    SAFE_READ

89  GET /api/v1/audit/query-resources
    searchAuditQueryResources
    SAFE_READ
```

Current census:

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

No Audit detail endpoint, full-text engine, generic entity resolver, generic search platform, URL/links authority or new mutation was added.

## 6. op78 closure laws

First-page structured query:

```text
occurred_at_from?     inclusive UtcInstant
occurred_at_before?   exclusive UtcInstant
actor_kind?           user | system
actor_user_id?        exact Uuid only with actor_kind=user
operation_codes?      closed unique AuditOperationCode set
resource_kind?
resource_id?          only with resource_kind
visibility_area_id?
limit?                1..100; default 20
```

`operation_codes` wire form is comma-separated `form/explode=false`, unique, max 37, canonicalized into closed enum order for cursor authority.

Binding evaluation remains:

```text
Organization current subject
→ Authorization.AuthorizedScopes(audit.read)
→ Audit historical visibility admission
→ structured predicates
→ occurred_at DESC,event_id DESC
→ cursor seek
→ page selection
→ optional recognition composition
```

Cursor remains stateless and authenticates operation + normalized predicates + order + seek position. Current `audit.read` is rechecked on every page.

`AuditEventView` remains immutable evidence; `AuditInspectionItem` wraps evidence with optional recognition so mutable labels never participate in filtering, ordering, cursor position or historical truth.

## 7. Bounded supersession / rebaseline result

Current bounded authority supersedes only conflicting current-tense Audit read clauses in:

```text
docs/product/journeys.md
docs/architecture/interfaces.md
docs/architecture/wire-contract.md
docs/architecture/frontend.md
docs/decisions/api-operation-census.md
```

Large ratified authorities were not rewritten wholesale. Current routing is through `../../decisions/audit-investigation-read.md` and the decision register.

Rebaseline result:

```text
Product/T6                    PASS
T8-C internal read contract   PASS
T8-E wire/census              PASS
T8-F / FP0 coverage           PASS
B01-B08 lock preservation     PASS
material contradictions       0
```

## 8. Closure gate

```text
Frontend Method v2.3                         OPERATOR-RATIFIED
B01-B08                                      LOCKED / OPERATOR-RATIFIED
B09                                          OPEN / ACTIVE
B09-F1 upstream Audit capability replan      CLOSED / OPERATOR-RATIFIED
Durable Audit investigation authority        PROMOTED
Current application census                   89
B09 P7                                       RESUMED / NEXT
B09 P8                                       BLOCKED pending P7
B10-B12                                      NOT OPEN
Product implementation                       BLOCKED
T12                                          NOT OPEN
Merge                                        NOT AUTHORIZED
```

Next step:

> Resume B09 P7 under the ratified structured-query, Query Assist, recognition and cross-link authority. Do not generate P8 functional HTML until P7 exits with no unresolved upstream finding.
