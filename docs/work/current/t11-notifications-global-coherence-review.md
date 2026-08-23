# T11 — Discussion / Mention / Notifications Global Coherence Review

> **Status:** LEAD GCR + FRESH FABLE — CONVERGED / READY FOR UPSTREAM CONSOLIDATION.  
> **Scope:** operator-ratified D0→D8 bounded reopen candidate only.  
> **Method:** DevelopmentConexus Engineering Method v1.0.0.  
> **Implementation:** BLOCKED.

## 1. Candidate package

```text
t11-b03-discussion-notification-mini-design.md
t11-b03-notification-ownership-reopen.md
t11-b03-notification-engagement.md
t11-b03-discussion-notification-d5.md
t11-b01-notifications-reopen.md
t11-b03-discussion-notification-d7-contract.md
t11-b03-notification-technology-spike.md
t11-notification-architecture-research.md
```

Current authorities checked:

```text
Product / journeys
T1 semantic state
Ownership topology
T3 Authorization + Audit
T5 async/search/external effects
T8-B backend owner topology
T8-C internal communication contracts
T8-D persistence
T8-E executable wire
T8-F frontend realization
T8-G runtime/process/deployment
B01/B03 T11 planning records
```

The review attacks duplicate authority, owner placement, transaction direction, cross-owner access, pagination/disclosure, offboarding races, durable-vs-ephemeral effects, frontend authority, runtime ingress direction and proofability.

## 2. Preserved candidate conclusions

The corrected package preserves:

```text
Document-level Discussion as Controlled Documents state
Message timeline + optional one-message reply reference
semantic Mention(user_id)
new document.discuss write Permission
Notifications as the second supporting semantic owner
4 business + 2 supporting semantic owners
seen / read-unread / archive-unarchive Inbox lifecycle
immutable Launch DiscussionMessage
current-source disclosure for Notification presentation
global bell + Quick Inbox + /notifications
stable SPA routes 10 -> 11
application operations 78 -> 86
Idempotency-Key creations 10 -> 11
ETag domains 13/13 unchanged
exact-byte resources 4 unchanged
same-local-transaction Mention -> Notification invariant
River remains the only durable future-work mechanism
SSE as non-authoritative invalidation transport
Lexical as replaceable Discussion-composer mechanism
generic EventBus / broker / Redis absent at Launch
```

## 3. Lead GCR Round 1 — NOT CONVERGED

Round 1 found:

```text
MATERIAL   3
IMPORTANT  6
```

Operator adjudication approved all nine bounded corrections.

### M1 — Authorization ownership / offboarding serialization

Finding: D7 blurred final Mention-target admission into Controlled Documents.

Correction applied:

```text
Organization authors current subject/eligibility facts
Controlled Documents authors exact Document predicate facts
Authorization alone computes ALLOW/default-DENY
application coordinates
```

Author + all unique Mention targets now use protected Organization eligibility in one Scope, resolved in deterministic `user_id ASC` order. Controlled Documents validates only intrinsic message/reply/content structure after current decisions are ALLOW.

**CLOSED.**

### M2 — current disclosure before pagination/counts

Finding: paging Notifications then filtering inaccessible sources would break cursor/count semantics; copied presentability would duplicate AuthZ authority.

Correction applied:

```text
Notifications candidate scan
→ batch source facts
→ current Organization/AuthZ composition
→ retain presentable candidates
→ bounded continued scan until requested presentable page + lookahead is filled or exhausted
→ compose presentation only for retained items
```

The same current-disclosure composition owns unseen/unread summary counts. Frontend post-filtering is forbidden.

**CLOSED.**

### M3 — SSE dependency direction / post-commit invalidation

Finding: a direct `transport -> platform/realtime` SSE handler would violate the one semantic inbound door; wake-up also needs to cover engagement changes.

Correction applied:

```text
transport/http
→ application/notifications
→ narrow consumer-owned subscription/wake-up port
→ platform realtime mechanism
```

Recipient wake-up happens only after committed Notification creation or engagement change: creation, seen, read, unread, archive, unarchive, mark-all-read. Wake-up is best-effort, non-durable and never called by semantic owners.

**CLOSED.**

## 4. Lead GCR Round-1 IMPORTANT findings — closure

### I1 — batch seen disclosure oracle

Corrected law:

```text
bounded unique ids
→ intersect current recipient + currently presentable
→ mutate only intersection
→ absent/foreign/non-presentable ids expose no per-id outcome/cardinality
```

**CLOSED.**

### I2 — completed create-message replay

Corrected replay:

```text
current caller AuthZ/disclosure recheck
→ completed idempotency recognition
→ stored message_id
→ zero new Message/Mention/Notification
```

Historical Mention-target eligibility is not rerun.

**CLOSED.**

### I3 — duplicate Audit/History streams

Corrected disposition:

```text
DiscussionMessage own immutable author/time/content truth
→ no duplicate Audit solely for creation
→ not Document lifecycle History

Notification creation/engagement/realtime
→ no mandatory semantic Audit in Launch
```

**CLOSED.**

### I4 — persistence enforcement

D7 requires upstream T8-D consolidation to preserve at minimum:

```text
notifications.* owner namespace
identity-only cross-owner FKs
unique DOCUMENT_MENTION per recipient + message
read_at -> seen_at
archived_at -> seen_at
immutable accepted DiscussionMessage/Mention
reply cannot cross Document Discussion
```

**CLOSED at architecture level.**

### I5 — OpenAPI/SSE proof

Operation 86 remains admitted only under proof that the selected Go OpenAPI boundary represents server-side `text/event-stream` without manual parallel route/DTO authority.

**CLOSED as explicit proof gate.**

### I6 — visual/upstream sequencing

Sequence remains:

```text
converged candidate
→ fresh independent challenge
→ reviewer adjudication
→ upstream consolidation
→ smallest B01 P8 Notification reopen
→ operator re-LOCKs B01 delta
→ B03 P8 with real Discussion semantics
```

**CLOSED.**

## 5. Lead GCR Round 2

Round 2 re-falsified authority duplication, transaction cycles, disclosure, async/realtime, API/census and YAGNI/framework posture.

```text
VERDICT: CONVERGED
MATERIAL: 0
IMPORTANT: 0
OPTIONAL: 0
```

No Lead finding changed the candidate result.

## 6. Fresh independent Fable challenge

Evidence PR #165 reviewed the exact corrected candidate independently under METHOD and Repository Standard isolation.

Verified Evidence posture:

```text
base candidate   arch/t11-implementation-program @ a9047924aa2e31aaa1418a15c8786b7e9ad2967f
review branch    review/t11-notifications-fable
review delta     docs/work/current/ai-dialog.md only
review CI        #1265 SUCCESS on final Fable head
review PR        DRAFT / NEVER MERGE
```

Fable independently re-derived and re-attacked all former Lead findings before reading the Lead closure and confirmed them closed.

Final Fable verdict:

```text
VERDICT: CONVERGED
MATERIAL: 0
IMPORTANT: 1
OPTIONAL: 3
UNSUPPORTED_PREFERENCE: 0
```

### F-1 — IMPORTANT — ACCEPTED

Finding: the consolidation-target map did not explicitly name:

```text
docs/decisions/forward-obligations.md
docs/decisions/api-operation-census.md
```

Without those targets, promoted Product/Architecture could say Notifications Launch + 86 operations while the decision register still said ASY-02 DEFERRED and census still said 78.

Adjudication:

```text
ACCEPT
```

The consolidation map now explicitly requires both decision-register changes in the same coherent promotion. This is consolidation precision only; it changes no Product capability, owner, route, Permission, operation count or mechanism choice.

### O-1 — OPTIONAL — ACCEPTED

`anchor_message_id` is one first-page navigation filter of `listDocumentDiscussionMessages`; continuation cursor authenticates the same operation/filter/order semantics. T8-E must provide a named fixture. No second pagination authority.

### O-2 — OPTIONAL — ACCEPTED

T6/T8-E must set explicit closed bounds for message segment count and unique Mention targets per message. The global 65,536-byte JSON ceiling alone is insufficient to cap protected-user lock/AuthZ/Notification fan-out.

### O-3 — OPTIONAL — ACCEPTED

T3/T6 must promote one named canonical Discussion-disclosure predicate reused by Discussion read, Mention-candidate/commit validation and Notification presentability. T8-G must record that transient deploy-overlap wake loss is tolerated because SSE is non-authoritative and canonical GET reconciles.

No Fable Round 2 is justified because no material contradiction survived and F-1 was a one-step consolidation-completeness correction accepted by the Lead/operator.

## 7. Independent survival statement

Fable explicitly concluded all of the following survive the independent challenge:

```text
4+2 owners                                      SURVIVES
11 stable SPA routes                            SURVIVES
16 PermissionCode values                        SURVIVES
86 application operations                       SURVIVES
11 Idempotency-Key creations                    SURVIVES
same-Scope Mention -> Notification              SURVIVES
presentability before paging/counts             SURVIVES
Lexical                                          SURVIVES
SSE + in-process wake-up                        SURVIVES
River as sole durable async                     SURVIVES
no generic EventBus / broker / Redis            SURVIVES
```

## 8. Converged review outcome

The coherent review chain is now:

```text
D0→D8 operator-ratified candidate
→ Lead GCR R1: NOT CONVERGED, 3M/6I
→ operator adjudication
→ corrected candidate
→ Lead GCR R2: CONVERGED, 0M/0I
→ fresh Fable: CONVERGED, 0M/1I/3O
→ Lead/operator adjudication: F-1 + O1→O3 ACCEPTED
→ Fable Round 2 NOT JUSTIFIED
→ upstream consolidation NEXT
```

Candidate counts remain candidate until that consolidation atomically updates every current authority that owns their meaning.

## 9. Required consolidation targets

At minimum:

```text
Product contract
T1 semantic state
Ownership topology
T3 Authorization/Audit
T5 async/realtime posture
T6 journeys/routes/API meaning
T8-B backend topology
T8-C internal contracts
T8-D persistence
T8-E executable wire
T8-F frontend realization
T8-G runtime
T9 golden-flow/validation obligations
T11 implementation-readiness planning

docs/decisions/forward-obligations.md
docs/decisions/api-operation-census.md
```

Only the actually implicated sections may change. No unrelated T1→T10 decision is reopened.
