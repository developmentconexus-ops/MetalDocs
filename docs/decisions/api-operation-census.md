---
id: api-operation-census
kind: authority
owner: architecture
summary: Owns the current 89-operation /api/v1 application census after operator-ratified T11 bounded reopens, including B11 Access Assignment read precision.
---

# API operation census

The current Launch `/api/v1` application census contains **89 operations**.

Current semantic authority is the combination of:

```text
docs/product/journeys.md
+ docs/decisions/discussion-notifications-launch.md
+ docs/decisions/audit-investigation-read.md
+ docs/decisions/access-assignment-read.md
```

The bounded T8-E read-symmetry precision previously added two reads to the original 76-operation census:

```text
GET /api/v1/users/{user_id}/profile
operationId: getUserProfile

GET /api/v1/areas/{area_id}/lifecycle
operationId: getAreaLifecycle
```

The T11 Discussion / `@mention` / Notifications bounded reopen added exactly eight application operations:

```text
79 GET   /api/v1/documents/{document_id}/discussion/messages
   operationId: listDocumentDiscussionMessages

80 POST  /api/v1/documents/{document_id}/discussion/messages
   operationId: createDocumentDiscussionMessage

81 GET   /api/v1/documents/{document_id}/discussion/mention-candidates
   operationId: searchDocumentDiscussionMentionCandidates

82 GET   /api/v1/notifications
   operationId: listNotifications

83 PATCH /api/v1/notifications/{notification_id}/engagement
   operationId: updateNotificationEngagement

84 PUT   /api/v1/notifications/seen
   operationId: markNotificationsSeen

85 PUT   /api/v1/notifications/read
   operationId: markAllNotificationsRead

86 GET   /api/v1/notifications/events
   operationId: streamNotificationEvents
   representation: text/event-stream
```

The T11 B09 Audit investigation bounded reopen preserves operation 78 `listAuditEvents` and adds exactly three purpose-built safe reads:

```text
87 GET   /api/v1/audit/query-areas
   operationId: listAuditQueryAreas

88 GET   /api/v1/audit/query-actors
   operationId: searchAuditQueryActors

89 GET   /api/v1/audit/query-resources
   operationId: searchAuditQueryResources
```

Operation 78 remains the sole Audit evidence traversal authority and is refined by `audit-investigation-read.md`; it is not counted twice.

The T11 B11 Access Administration bounded reopen refines existing operation 31 only:

```text
31 GET /api/v1/role-assignments
   operationId: listRoleAssignments
   REFINED — server-side User / Group / Scope / Role filters
             + human-recognizable subject/scope read projection
```

No B11 operation is added or removed. Mutation operations 28, 29, 32 and 33 remain unchanged.

The T11 B12 Document Governance Administration bounded reopen refines existing operation 43 only:

```text
43 GET /api/v1/document-governance/templates
   operationId: listTemplateConfigurations
   REFINED — server-side q search + eligible-type / template-role /
             effective-revision filters before pagination
```

No B12 operation is added or removed; the reopen is query-only.

## Count proof

```text
original journeys census                    76
bounded read-symmetry precision             +2
T11 Discussion/Mention/Notifications reopen +8
T11 Audit investigation bounded reopen      +3
T11 B11 op31 read precision                  +0
T11 B12 op43 read precision                  +0
                                            ---
current application census                  89
```

## Idempotency / ETag / exact-byte census

Exactly one of the eight Discussion/Notifications operations is a new non-idempotent semantic POST creation:

```text
createDocumentDiscussionMessage
```

The three Audit investigation additions are all `SAFE_READ` and add no Idempotency-Key creation, ETag domain or exact-byte resource.

The B11 op31 refinement is a `SAFE_READ` precision and likewise adds no Idempotency-Key creation, ETag domain or exact-byte resource.

Current cross-contract counts are:

```text
application operations        89
Idempotency-Key creations     11
ETag read / mutation domains  13 / 13
exact-byte resources          4
```

The T11 Discussion creation reuses the current global durable Idempotency-Key law; its ReplaySnapshot stores stable `message_id` only and does not copy message/free text.

## Authority / historical-snapshot law

`docs/decisions/discussion-notifications-launch.md` is the bounded current authority for operations 79–86.

`docs/decisions/audit-investigation-read.md` is the bounded current authority for the op78 structured-query/inspection refinement and operations 87–89.

`docs/decisions/access-assignment-read.md` is the bounded current authority for the op31 Access Assignment read/query refinement. It does not create operation 90, a new semantic owner or an effective-access engine.

`docs/decisions/content-format-vocabulary.md` is the bounded current authority for the ContentFormat vocabulary, per-format structural admission and converter-bound official rendition. It adds no operation: the census stays at 89 and the 11/13/4 supporting censuses are unchanged.

`docs/decisions/template-configuration-read.md` is the bounded current authority for the op43 template-configuration read/query refinement. It does not create operation 90, a "general template" concept or Area-scoped templates.

Together these bounded decisions supersede only the conflicting current-tense Product/T6/T8-E/T8-F/T8-H/T9 statements on their exact subjects.

This page is the **sole current numeric census authority**. Therefore older `78`, `78/78`, `86`, `operation 79 absent`, `operation 87 absent`, `no operation 79`, `operation 87+ requires reopen`, or equivalent closed-count statements found in:

```text
Product/T6 authority
T4/T8/T9/T10 architecture pages
T8-F/T8-H/T9/T10 ratification records
frontend-read/responsible-owner precision provenance
transition/content-integrity closure prose
current T11 bounded-decision census/proof blocks that predate a later T11 bounded reopen
other pre-current-T11 closure snapshots
```

are interpreted as one of:

```text
historical stage snapshot, truthful for that stage when ratified
OR
bounded current-tense clause superseded by the current T11 decisions + this census authority
```

They never override the current:

```text
89 operations
11 Idempotency-Key creations
13 / 13 ETag read/mutation domains
4 exact-byte resources
```

All unchanged journey and wire laws remain owned by `docs/product/journeys.md` and `docs/architecture/wire-contract.md` plus the bounded decisions that explicitly supersede them.

Any operation **90 or later** requires unchanged semantic normalization already permitted by current authority or a new explicit bounded Product/T6 reopen. Framework convenience, screen shape or implementation preference is not authority to add an operation.

T8-E executable realization must prove exactly **89** application operations and the current 11/13/4 supporting censuses, with operation 31 realized according to `access-assignment-read.md` and operation 43 realized according to `template-configuration-read.md`.
