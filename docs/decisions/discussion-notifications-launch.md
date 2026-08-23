---
id: discussion-notifications-launch
kind: authority
owner: architecture
summary: Owns the bounded Launch V1 Discussion, @Mention, Notifications, Inbox and realtime amendment across Product/T1/T3/T5/T6/T8-B→G/T9.
---

# Launch V1 Discussion / Mention / Notifications — bounded current authority

> **Status:** OPERATOR-RATIFIED / LEAD-GCR-CONVERGED / FABLE-CONVERGED.  
> **Scope:** bounded Product/architecture reopen triggered during T11 B03 planning.  
> **Method:** DevelopmentConexus Engineering Method v1.0.0.  
> **Implementation:** BLOCKED by the repository roadmap.

## 1. Authority and supersession law

This page is the **single bounded current authority** for the Launch V1 Document Discussion + `@mention` + Notifications capability discovered during T11.

It does not replace the Product/T1→T9 authorities wholesale. It supersedes **only** their previously closed statements that conflict with the decisions below. Every unchanged statement in those authorities remains current.

Bounded supersession map:

```text
docs/product/contract.md
  Launch Core did not include stable-Document Discussion/@Mention/in-app Inbox

docs/architecture/ownership.md
  Launch topology = 4 business + Audit only
  Notifications classified as non-semantic

docs/architecture/domain-model.md
  owner/family census omitted DiscussionMessage/Mention/Notification

docs/architecture/authorization-and-audit.md
  Permission vocabulary = 15 values
  role bundles omitted document.discuss
  Discussion/Mention protected eligibility absent

docs/architecture/async-and-search.md
  Notifications deferred absent a consumer
  no Notification realtime consumer

docs/product/journeys.md
  stable SPA routes = 10
  no Notifications Product route/workspace
  application census = 78

docs/architecture/backend.md
  exactly five semantic homes/public owner packages
  Notifications explicitly absent as semantic owner

docs/architecture/interfaces.md
  five-owner service list / no Notifications owner contract
  no Discussion→Notification same-Scope choreography

docs/architecture/persistence.md
  owner namespace/state census omitted controlled-document Discussion + notifications.*

docs/architecture/wire-contract.md
  exact application wire/census = 78
  PermissionCode = 15 values
  no Discussion/Notification schemas or SSE operation

docs/architecture/frontend.md
  stable routes = 10
  no Notifications route/feature/runtime consumer
  78/78 operation coverage

docs/architecture/runtime.md
  notification/realtime channel listed as non-consumer

docs/architecture/validation-baseline.md
  Golden Flows/properties omitted Discussion/Mention/Notification/realtime proof
```

When this page and an older sentence above disagree on the bounded subject, **this page wins**. `docs/decisions/index.md`, `forward-obligations.md` and `api-operation-census.md` must route the current disposition accordingly.

This bounded-authority form deliberately avoids rewriting unrelated material in large ratified documents merely to change one coherent cross-layer capability. A later substantive rewrite may absorb these clauses, but may not change their meaning without a normal reopen.

## 2. Material trigger and Product boundary

Launch V1 now requires:

```text
stable-Document human Discussion
+ explicit @mention of eligible MetalDocs Users
+ persistent in-app Notification for accepted Mention
+ global Notification Inbox
```

Discussion belongs to the **stable Controlled Document**, not to WorkingContent, a Revision, Submission, GovernanceAttempt or a generic chat/thread platform.

Semantic separation:

```text
Document Discussion != DRAFT EditorialComment
Document Discussion != SubmissionFeedback
Document Discussion != GovernanceCase feedback
Notification != access grant
Notification != document-read acknowledgement
Notification != governance evidence
```

One chronological Discussion exists per stable Document. A message may optionally reference one earlier message, but no separate Thread aggregate or arbitrary nesting exists.

## 3. Semantic ownership — current Launch topology = 4 + 2

Current Launch semantic owners are:

```text
BUSINESS
1. Authentication
2. Organization
3. Authorization
4. Controlled Documents

SUPPORTING SEMANTIC
5. Audit
6. Notifications
```

### Controlled Documents owns

```text
DocumentDiscussionMessage
Mention
stable-Document Discussion membership of messages
reply-to relationship
contextual official Revision-at-post reference
```

### Notifications owns

```text
Notification identity
recipient User reference
closed Notification kind/source reference
created_at
deduplication identity
seen/read/archive engagement state
Inbox ordering/current engagement state
```

Current Notification kind is a closed union with one member:

```text
DOCUMENT_MENTION
  document_id
  message_id
```

Notifications does **not** own source content, Document lifecycle, User identity/profile, Authorization, access grants, delivery transport, River jobs or realtime connection state.

The promotion does not imply a service boundary. Current physical posture remains one Go modular-monolith application runtime + one PostgreSQL product-state database.

## 4. DiscussionMessage and Mention semantics

Accepted message shape:

```text
DocumentDiscussionMessage
  message_id
  document_id
  author_user_id
  created_at
  content: MessageContentSegment[]
  reply_to_message_id?
  official_revision_at_post?
```

`official_revision_at_post` is the exact current official Revision identity when one exists at acceptance time; absent before first Release. It is context/provenance only and never moves ownership from stable Document to Revision.

Content union:

```text
Text(text)
Mention(user_id)
```

Mention authority is stable `user_id`; mutable display text is presentation only. No accepted message is stored as Lexical JSON, HTML, ProseMirror or reparsed `@Name` text.

Launch accepted messages are immutable:

```text
edit message    absent
hard delete     absent
correction      new DiscussionMessage
```

A future moderation/privacy/retention requirement may reopen redaction/deletion semantics; Launch does not claim indefinite physical retention.

Reply law:

```text
reply_to_message_id, when present,
  references one message in the same Document Discussion

reply to reply
  remains one ordinary new message with one reference

no semantic nesting depth
```

## 5. Canonical Discussion disclosure predicate

The bounded reopen names one reusable predicate:

```text
DocumentDiscussionDisclosure(actor, document)
```

Meaning:

> the current enabled actor is currently permitted to receive the exact stable-Document Official/Discussion context for that Document under canonical Organization + Authorization grants/scopes + Controlled Documents relationship/state disclosure facts.

This is **not** a new Permission. It is the named composition of the same canonical current disclosure that makes the stable Document Official lens receivable.

It is the single citable source for:

```text
Discussion read/list
Mention autocomplete admission
Mention target commit-time validation
DOCUMENT_MENTION Notification presentability
Notification engagement target admission
```

Authority partition remains:

```text
Organization        -> User existence / Company / ENABLED / Group facts
Controlled Documents -> exact Document relationship/state/disclosure predicate facts
Authorization       -> final ALLOW / default DENY
application         -> choreography only
```

No frontend, Controlled Documents package or Notifications package owns a parallel permission matrix.

## 6. `document.discuss` Permission and role bundles

Current Permission vocabulary adds exactly:

```text
document.discuss
```

Current PermissionCode count becomes **16**.

Role-bundle delta:

```text
viewer
  + document.discuss

author
  + document.discuss

approver
  + document.discuss

area_manager
  + document.discuss

governance_viewer
  no document.discuss by default

governance_admin
  no document.discuss by default
```

Reading Discussion does not require `document.discuss`; it follows `DocumentDiscussionDisclosure`. Writing/replying/mentioning requires both:

```text
document.discuss in matching scope
+ DocumentDiscussionDisclosure(actor, document)
```

Commands always recheck current truth.

## 7. Mention candidate + acceptance law

Mention candidate eligibility:

```text
existing MetalDocs User
+ same Company
+ ENABLED
+ DocumentDiscussionDisclosure(target, document)
+ target != message author
```

A target does not need `document.discuss`; a read-only actor may be mentioned without gaining reply authority.

Autocomplete is purpose-built for the exact Document. It returns only bounded current `UserReference`-like selection data and never exposes email, role/permission sets, admin profile fields or exclusion reasons.

Autocomplete result is guidance only. Message acceptance atomically revalidates current author + targets.

Protected eligibility/offboarding law:

```text
author + all unique Mention targets
→ deduplicate User ids
→ acquire protected Organization eligibility in user_id ASC order
→ hold eligibility protection through the caller-owned Scope
```

If offboarding linearizes first, the command observes DISABLED and fails closed. If the message transaction linearizes first, it may commit under valid current truth; later access drift is handled by current Notification presentability.

If any explicit Mention is invalid at command acceptance:

```text
reject whole message
zero DiscussionMessage
zero Mention
zero Notification
```

No silent drop of an invalid Mention is allowed.

Trigger law:

```text
explicit accepted Mention -> exactly one DOCUMENT_MENTION Notification per unique target/message
same User repeated in one message -> one Notification
author self-mention -> not admitted
reply without explicit Mention -> no Notification merely because it is a reply
reply + explicit Mention -> normal Mention law
```

## 8. Cross-owner atomicity

The protected business invariant is:

```text
accepted explicit Mention
<=>
required persistent DOCUMENT_MENTION Notification exists
```

Application enforces this inside one caller-owned local PostgreSQL Scope:

```text
BEGIN Scope
→ protected subject facts
→ Controlled Documents predicate facts
→ Authorization current decisions
→ Controlled Documents inserts Message + Mention
→ Notifications inserts one Notification per unique target/message
COMMIT
```

Notification persistence failure aborts the message transaction.

No River job, generic outbox, generic EventBus or external broker mediates this local invariant. Same-database transaction composition is the smaller and stronger structure for the current one-producer/one-semantic-consumer fact.

Completed `createDocumentDiscussionMessage` idempotency replay rechecks current caller authorization/disclosure before replay recognition but does **not** rerun historical Mention-target eligibility. Replay returns the stored `message_id` and creates no new Message/Mention/Notification.

## 9. Notification engagement lifecycle

Persistent semantic state:

```text
created_at
seen_at?
read_at?
archived_at?
```

### Seen

```text
seen_at absent  -> unseen
seen_at present -> seen
```

`seen_at` is monotonic and represents actual presentation/recipient interaction, not source acknowledgement.

### Read / unread

```text
read_at absent  -> unread
read_at present -> read
READ => SEEN
```

Mark read sets `seen_at` if absent and sets `read_at`. Mark unread clears only `read_at`; it never makes the item unseen/new.

### Archive

```text
archived_at absent  -> active Inbox
archived_at present -> archived
```

Archive/unarchive preserves seen/read. Any deliberate per-item engagement implies `seen_at` if absent.

Binding distinctions:

```text
Notification SEEN     != Document viewed
Notification READ     != Document read
Notification READ     != Read & Acknowledge
Notification READ     != governance evidence
Notification ARCHIVED != delete/source resolution
Notification UNREAD   != unseen/new
```

No durable User-level unseen/unread counter becomes authority; counts are derived from current Notification + disclosure truth.

## 10. Presentability and access drift

Notification persistence and current presentation are distinct.

A `DOCUMENT_MENTION` Notification is presentable only when:

```text
recipient == current User
+ current User ENABLED
+ DocumentDiscussionDisclosure(current User, source Document)
```

When not presentable, the Notification is omitted server-side from:

```text
active Inbox
archived Inbox
unseen badge/count
unread count
batch/bulk engagement target sets
source presentation
```

No source code/title/author/message preview is leaked for the client to hide.

Loss of access does not delete the Notification or mutate seen/read/archive. If access later becomes valid again, the same Notification may reappear with unchanged engagement state.

Notification persistence stores only source identities; source/user presentation is resolved under current disclosure at read time.

## 11. Product IA / stable SPA routes

Current stable Product SPA routes become **11**:

```text
/documents
/documents/:document_id
/documents/:document_id/work
/documents/:document_id/history
/work
/work/governance/:attempt_id
/notifications
/audit
/admin/organization
/admin/access
/admin/document-governance
```

`/notifications` is a global attention utility route and does **not** become a permanent sidebar item.

B01 mental model remains:

```text
Início       = current operational situation
Minha Caixa  = assigned work
Documentos   = official truth / creation
Gestão       = configuration
Evidência    = audit/evidence
```

Notifications is not placed under `Minha Caixa`; attention and assigned work remain distinct.

Global shell delta:

```text
utility header
  Notification bell + unseen badge
  desktop Quick Inbox
  narrow/mobile accessible sheet/full-screen quick surface

sidebar
  unchanged
```

Quick Inbox and `/notifications` are two presentations of the same Notifications owner/read family, never separate stores.

A smallest-scope rendered B01 P8 delta still requires operator re-LOCK before the structural frontend block is closed again.

## 12. API operation census — current bounded delta

Eight application operations are added:

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
   success media type: text/event-stream
```

Current application census after promotion:

```text
operations                  86
Idempotency-Key creations   11
ETag read/mutation domains  13 / 13
exact-byte resources        4
```

Only operation 80 is a new non-idempotent semantic creation requiring `Idempotency-Key`.

## 13. Discussion wire semantics

### Pagination / ordering

Potentially unbounded Discussion uses the existing stateless integrity-protected cursor law.

Canonical selection order:

```text
created_at DESC
message_id DESC
```

Initial ordinary page selects the newest `limit` messages. The response emits the selected window in chronological reading order (oldest → newest). `next_cursor` advances to the next **older** window.

Default/maximum limits reuse the global list baseline unless a stricter generated schema states otherwise:

```text
default 20
max     100
```

### `anchor_message_id`

`anchor_message_id` is an optional first-page navigation filter for `listDocumentDiscussionMessages` only.

```text
cursor absent + anchor_message_id
  select a window of up to limit messages whose newest item is the exact anchor
  emit oldest -> anchor
  next_cursor continues to older windows

cursor present
  anchor_message_id must not be repeated as ordinary query input;
  the opaque cursor authenticates the original anchor/filter/order semantics
```

Absent/non-disclosable anchor uses normal non-disclosing not-found behavior. T8-E contract proof requires a named fixture for anchor + continuation.

### Message bounds

Current executable bounds:

```text
MessageContentSegment count           1..64
unique Mention targets per message    0..20
aggregate accepted Text code points   <= 4096
```

These limits are deliberately human-scale and cap one-command protected User-lock/AuthZ/Notification fan-out. They are Product/wire bounds, not editor defaults.

A message must contain at least one meaningful segment; empty/blank-only text without a Mention is invalid. Duplicate target mentions remain valid presentation but collapse to one Notification target for trigger semantics.

### Message response

Discussion response resolves author/Mention display through bounded current `UserReference` enrichment. Stable User identity remains even if human profile enrichment was erased.

## 14. Notification wire semantics

### List

`GET /api/v1/notifications` admits:

```text
view = active | unread | archived
cursor
limit
```

Canonical Notification order:

```text
created_at DESC
notification_id DESC
```

First page includes derived presentable non-archived:

```text
unseen_count
unread_count
```

These are engagement summaries, not generic pagination totals.

Public pagination/counts are formed **after** server-side current disclosure composition:

```text
Notification candidate scan
→ batch source predicate facts
→ current Organization/AuthZ decision
→ retain presentable
→ continue bounded scan until requested page + one presentable lookahead or exhaustion
→ compose presentation only for retained items
```

No React post-filter or copied ACL/presentability state is permitted.

### Per-item engagement

`PATCH /notifications/{id}/engagement` accepts at least one of:

```text
read: boolean
archived: boolean
```

Direct engagement requires recipient ownership + current presentability. Non-presentable/foreign/absent uses normal non-disclosing not-found behavior.

### Batch seen

`PUT /notifications/seen` accepts a unique list of at most **100** Notification ids.

Server intersects the request with current-recipient + currently-presentable items and sets seen only on that intersection. Absent/foreign/non-presentable ids expose no per-id result and no count/cardinality. Success is bodyless `204`.

### Mark all read

`PUT /notifications/read` marks all current-recipient Notifications that are presentable + non-archived + unread at the operation point. It sets seen if absent and read to the accepted action time. Success is bodyless `204`.

## 15. SSE / realtime

`GET /api/v1/notifications/events` is an authenticated application operation whose successful representation is `text/event-stream`.

Call graph:

```text
transport/http
→ application/notifications
→ narrow application-owned subscription port
→ platform realtime mechanism
```

Accepted event is an invalidation only:

```text
event: notifications.changed
data: {}
```

It contains no Document/message/User/Notification business payload.

Browser uses native `EventSource`; signal invalidates/refetches canonical Notification queries.

Every successful Notification state change that affects another tab attempts recipient wake-up **after commit**:

```text
DOCUMENT_MENTION creation
mark seen
mark read
mark unread
archive
unarchive
mark all read
```

Wake-up failure never rolls back Product truth and is not a River job.

Launch one-replica mechanism is an in-process coalescing recipient hub. A slow subscriber cannot block Product mutation. Multiple pending invalidations may collapse into one because the signal means only “canonical Notifications changed”.

Transient wake loss during process/deploy overlap is tolerated because canonical `GET /notifications`, reconnect and focus/refetch restore truth. If multiple live replicas become a real steady-state requirement, PostgreSQL LISTEN/NOTIFY is the first qualified wake-mechanism candidate; it never becomes Notification authority.

Operation 86 remains subject to implementation-readiness proof that the selected Go OpenAPI boundary realizes server-side `text/event-stream` without a manual parallel route/DTO registry.

## 16. Technology realization boundaries

Current selected mechanism evidence:

```text
Discussion composer
  Lexical core + @lexical/react
  PlainText + MetalDocs MentionNode/typeahead only
  Lexical state never persisted

Notification frontend
  native MetalDocs feature
  OpenAPI-generated types
  existing thin transport
  TanStack Query server state

realtime client
  native browser EventSource

realtime server
  narrow Go net/http SSE adapter

Launch wake-up
  in-process coalescing hub

durable future work
  River remains sole PostgreSQL-backed durable-job mechanism
```

No Novu/Knock/MagicBell runtime, Watermill, EventBus, broker or Redis is a Launch dependency. Watermill is only a future candidate if multiple independent consumers create proven event-coupling pressure.

## 17. Persistence obligations

Current owner namespace catalog expands with:

```text
notifications.*   Notifications
```

Controlled Documents persistence must represent immutable DiscussionMessage + Mention state. Notifications persistence represents persistent Notification + engagement state.

Structural properties required:

```text
identity-only cross-owner references only
unique DOCUMENT_MENTION per recipient_user_id + message_id
read_at IS NOT NULL     -> seen_at IS NOT NULL
archived_at IS NOT NULL -> seen_at IS NOT NULL
accepted DiscussionMessage/Mention immutable through serving privileges/structure
reply reference cannot cross Document Discussion
```

Cross-owner FKs do not authorize foreign-owner SQL and use normal no-cascade authority laws.

Exact table names/decomposition remain implementation realization detail as long as these properties are mechanically proven.

## 18. Audit / History

DiscussionMessage owns its own immutable trusted author/time/content record. It does not create a duplicate AuditEvent merely to copy message creation and is not inserted into Document lifecycle History.

Notification creation/engagement/realtime has no mandatory semantic AuditEvent in Launch. Operational telemetry remains allowed.

A later regulatory/customer requirement for messaging/notification Audit is a bounded T3 reopen.

## 19. Validation / proof obligations

T9/implementation-readiness proof must falsify at least:

```text
Discussion read/Mention/presentability use one canonical disclosure predicate
frontend contains no parallel AuthZ matrix
protected author/target eligibility serializes with offboarding in deterministic user order
invalid Mention -> zero Message + zero Notification
Notification persistence failure -> zero accepted Message
same Idempotency-Key replay -> same message_id + zero duplicate Notification
completed replay does not rerun historical target eligibility
one target repeated -> one Notification per target/message
reply cannot cross Document
non-presentable Notification -> absent from items/counts + zero source metadata leak
presentable pagination has no sparse-page/frontend-postfilter authority
batch seen leaks no requested-id existence/cardinality
mark unread preserves seen
archive/unarchive preserves read/seen
mark-all-read excludes currently non-presentable items
SSE payload carries zero source business truth
lost SSE wake -> canonical GET still restores Notification truth
transport -> application -> mechanism dependency direction
post-commit-only wake-up
Discussion/Notifications do not duplicate Audit/History truth
persistence constraints match the protected invariants
OpenAPI server boundary realizes `text/event-stream` without a parallel manual contract
application operation census = 86
Idempotency-Key creation census = 11
ETag domains = 13/13
exact-byte resources = 4
zero generic EventBus/broker/Redis runtime dependency
Lexical serialization never becomes Product persistence
```

## 20. Deferred/future seams

Still not Launch:

```text
email Notification delivery
push Notification delivery
Notification preferences/subscriptions/digests
snooze/priority platform
message editing/deletion/moderation workflow
generic event-processing platform
multi-replica wake mechanism without a real runtime requirement
Read & Acknowledge equivalence to Notification read
```

Future email/push/webhook delivery must attach to persistent Notification intent under existing T5 named durable-effect/River laws and may not redefine Discussion or engagement semantics.

## 21. Review proof

Review chain:

```text
D0→D8 operator-ratified candidate
→ Lead GCR Round 1: NOT CONVERGED / MATERIAL=3 / IMPORTANT=6
→ operator accepted bounded corrections
→ Lead GCR Round 2: CONVERGED / MATERIAL=0 / IMPORTANT=0
→ fresh Fable PR #165: CONVERGED / MATERIAL=0 / IMPORTANT=1 / OPTIONAL=3
→ operator accepted F-1 + O1→O3
→ Fable Round 2 NOT JUSTIFIED
```

Fresh Fable explicitly confirmed survival of:

```text
4+2 owners
11 stable SPA routes
16 PermissionCode values
86 application operations
11 Idempotency-Key creations
same-Scope Mention -> Notification
server-side presentability before paging/counts
Lexical
SSE + in-process wake-up
River as sole durable async
no generic EventBus / broker / Redis
```

Reviewer output remains Evidence; the operator-ratified decisions above are authority.

## 22. Reopen triggers

Reopen only on material evidence such as:

```text
Discussion usage proves immutable follow-up correction materially insufficient
privacy/moderation/retention creates a concrete delete/redaction requirement
Notification engagement no longer has independent persistent lifecycle
presentability-before-pagination becomes unsustainable at measured scale
multiple real independent event consumers create temporal-coupling pressure
multiple steady-state application replicas require cross-process wake
SSE cannot meet the required same-origin freshness experience under qualification
Lexical cannot satisfy accessibility/IME/composition without disproportionate custom work
external email/push delivery becomes a current Product requirement
new assurance requires messaging/Notification semantic Audit
```

Framework popularity, speculative scale or desire for fewer/more packages are not reopen evidence.
