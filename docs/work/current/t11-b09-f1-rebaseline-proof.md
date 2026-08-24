# T11 — B09-F1 Bounded Rebaseline Proof

> **Status:** PASS / OPERATOR-RATIFIED AUTHORITY PROMOTED.  
> **Block:** B09 — Audit.  
> **Finding:** B09-F1.  
> **Durable authority:** `../../decisions/audit-investigation-read.md`.  
> **Numeric authority:** `../../decisions/api-operation-census.md`.  
> **Implementation:** BLOCKED.

## 1. Purpose

Prove that the operator-ratified Audit investigation reopen changed only the smallest owning Product/API/frontend authority needed to satisfy the real Auditor jobs, while preserving unrelated Product/T1→T10 and B01-B08 decisions.

The current supersession mechanism is the bounded durable decision `audit-investigation-read.md`; large previously ratified Product/T8 documents are not rewritten wholesale merely to absorb one coherent T11 delta.

## 2. Affected-surface matrix

```text
Product/T6
  Audit job = point investigation + period/authorized-scope review
  /audit route unchanged
  export/full-text/saved-search remain deferred
  custom sort/analytics remain rejected for Launch/B09

T8-C
  Audit structured predicates + historical visibility stay Audit-owned
  application composes optional current recognition and Query Assist owner facts
  no generic resolver/query service

T8-E
  op78 refined, not replaced
  op87-op89 added as SAFE_READ
  evidence/recognition structurally separated
  cursor/filter laws preserved

T8-F / FP0
  operation coverage 86 -> 89
  stable routes 11 unchanged
  /audit consumes 78 + 87 + 88 + 89
  no frontend AuthZ evaluator
  no client post-filter over incomplete pages

Censuses
  operations 89
  routes 11
  permissions 16
  idempotency 11
  ETag 13/13
  exact-byte 4
  owners 4+2

Locked blocks
  B01-B08 preserved
  no contradiction found
```

## 3. Product / journey rebaseline

Ratified Launch Audit jobs are now:

```text
A. investigate one exact action/evidence question
B. review a bounded period inside current audit.read + historical visibility
```

The stable Product route remains exactly:

```text
/audit
```

No new top-level workspace or SPA route is introduced.

Explicit non-Launch/disposition posture remains:

```text
free-text Audit search       DEFERRED
external export              DEFERRED
saved searches               DEFERRED
custom sort                  REJECTED
analytics/dashboard          REJECTED as B09 responsibility
query DSL                    REJECTED
```

Audit remains action evidence and never becomes current lifecycle-state authority.

## 4. T8-C internal-contract rebaseline

Existing read boundary remains:

```text
Organization current subject
→ Authorization.AuthorizedScopes(audit.read)
→ application maps to Audit.ReadVisibility
→ Audit historical visibility before pagination
```

Bounded addition:

```text
Audit
  owns structured evidence predicates
  owns admitted historical identities
  owns seek/page truth

semantic owners
  own current recognition facts for their entities

application
  composes/intersects bounded recognition and Query Assist facts
```

No new semantic owner, generic service locator, generic query service, shared semantic contract package or cross-owner repository is created.

## 5. T8-E wire rebaseline

Current Audit operation family:

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

Operation 78 remains the sole Audit evidence traversal authority.

Its first-page structured query now covers:

```text
occurred_at_from / occurred_at_before
actor_kind + actor_user_id when USER
operation_codes set
resource_kind + optional exact resource_id
optional visibility_area_id
limit
```

Canonical order remains:

```text
occurred_at DESC,event_id DESC
```

Cursor remains stateless and authenticates operation + normalized filters + order + seek position. Current `audit.read` is rechecked every page.

The inspection response structurally separates:

```text
AuditEventView     immutable evidence
recognition        optional current/non-historical presentation context
```

No `GET /audit/events/{event_id}` operation, generic search endpoint, entity directory or deep-link resolver is added.

## 6. T8-F / FP0 rebaseline

Current application coverage is:

```text
89 application operations
11 stable SPA routes
```

B09 `/audit` consumes exactly:

```text
78 listAuditEvents
87 listAuditQueryAreas
88 searchAuditQueryActors
89 searchAuditQueryResources
```

Frontend laws:

```text
server filters complete evidence before pagination
frontend never evaluates audit.read scope itself
frontend never post-filters incomplete pages as complete truth
mutable human labels never become query identity
same-actor/resource/action investigation stays inside Audit
owner links are secondary and destination rechecks current AuthZ/disclosure
```

No B01-B08 locked screen/route/pattern is invalidated by this delta.

## 7. Numeric proof

```text
original journeys census                    76
bounded read-symmetry precision             +2
T11 Discussion/Mention/Notifications reopen +8
T11 Audit investigation bounded reopen      +3
                                            ---
current application census                  89
```

Supporting counts:

```text
stable SPA routes                11
PermissionCode values            16
Idempotency-Key creations        11
ETag read / mutation domains     13 / 13
exact-byte resources             4
semantic owners                  4 business + 2 supporting
```

The three new operations are all safe reads; no supporting count changes.

## 8. Lock-preservation proof

```text
B01 App Shell + IA + Home                 PRESERVED
B01N Notification chrome                  PRESERVED
B02 Library                               PRESERVED
B03 Document Official                     PRESERVED
B04 Document Work                         PRESERVED
B05 My Work                               PRESERVED
B06 Governance Case                       PRESERVED
B07 Document History                      PRESERVED
B08 Notifications Full Inbox              PRESERVED
```

No new evidence contradicts their locked Product decisions.

Audit/History separation is strengthened rather than weakened:

```text
Audit   = who did what and when across the Product
History = Controlled Documents semantic lifecycle story
```

## 9. Finding closure verdict

```text
B09-F1 finding trigger                     VALID
current backend-as-ceiling assumption      REJECTED by Method v2.3
Auditor jobs A+B                           RATIFIED
Structured Audit Query                     RATIFIED
Human recognition                         RATIFIED
Audit Query Assist                         RATIFIED
Owner-lens cross-links                     RATIFIED
Exact op78 + op87-op89 package             RATIFIED / DURABLE
bounded authority promoted                 YES
numeric census promoted                    89
material contradiction after promotion     0
B01-B08 reopen required                    0
```

Verdict:

```text
B09-F1 CLOSED / OPERATOR-RATIFIED
B09 P7 RESUMED / NEXT
B09 P8 BLOCKED pending clean P7 exit
B10-B12 NOT OPEN
Product implementation BLOCKED
T12 NOT OPEN
```
