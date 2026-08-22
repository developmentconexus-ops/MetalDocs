# T11 — Discussion / Mention / Notifications Global Coherence Review

> **Status:** LEAD GCR ROUND 2 — CONVERGED / READY FOR FRESH INDEPENDENT CHALLENGE.  
> **Scope:** operator-ratified D0→D8 bounded reopen candidate only.  
> **Method:** DevelopmentConexus Engineering Method v1.0.0.  
> **Implementation:** BLOCKED.  
> **Independent review:** NOT YET PERFORMED; this Lead GCR is not independent evidence.

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

## 3. Round 1 — NOT CONVERGED

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

**Round-2 result: CLOSED.** No parallel permission/disclosure authority remains.

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

**Round-2 result: CLOSED.** Pagination is now compatible with current disclosure without copied ACL truth.

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

**Round-2 result: CLOSED.** T8-B/T8-C dependency laws remain intact.

## 4. Round-1 IMPORTANT findings — closure

### I1 — batch seen disclosure oracle

Corrected law:

```text
bounded unique ids
→ intersect current recipient + currently presentable
→ mutate only intersection
→ absent/foreign/non-presentable ids expose no per-id outcome/cardinality
```

Direct non-presentable engagement uses normal non-disclosing not-found behavior.

**CLOSED.**

### I2 — completed create-message replay

Corrected replay:

```text
current caller AuthZ/disclosure recheck
→ completed idempotency recognition
→ stored message_id
→ zero new Message/Mention/Notification
```

Historical Mention-target eligibility is not rerun. Fingerprint includes Document, reply reference and normalized ordered Text/Mention content; ReplaySnapshot excludes free text.

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

Future regulatory messaging/notification audit remains a T3 reopen trigger.

**CLOSED.**

### I4 — persistence enforcement

D7 now requires upstream T8-D consolidation to preserve at minimum:

```text
notifications.* owner namespace
identity-only cross-owner FKs
unique DOCUMENT_MENTION per recipient + message
read_at -> seen_at
archived_at -> seen_at
immutable accepted DiscussionMessage/Mention
reply cannot cross Document Discussion
```

**CLOSED at architecture level; exact DDL remains downstream consolidation/implementation proof.**

### I5 — OpenAPI/SSE proof

Operation 86 remains candidate only under a closure proof that the selected Go OpenAPI boundary represents server-side `text/event-stream` without a manual route/DTO authority. Toolchain failure reopens mechanism/tooling, not Product meaning.

**CLOSED as an explicit proof gate.**

### I6 — visual/upstream sequencing

Sequence remains binding:

```text
converged candidate
→ fresh independent challenge
→ reviewer adjudication
→ smallest upstream Product/T1→T9/T11 consolidation
→ smallest B01 P8 Notification reopen
→ operator re-LOCKs implicated B01 delta
→ B03 P8 with real Discussion semantics
→ operator B03 adjudication/LOCK
```

B04+ remains unopened.

**CLOSED.**

## 5. Round-2 adversarial sweep

### Authority duplication

No duplicate owner survives:

```text
Controlled Documents -> Discussion/Mention source facts
Authorization        -> final current ALLOW/DENY
Notifications        -> persistent attention/engagement
Audit                -> separate evidence authority only
application          -> choreography only
realtime             -> mechanism only
```

**PASS.**

### Transaction/cycle challenge

Same-Scope Message + Notification creation has no owner→owner import and no event-bus cycle. Application is the sole orchestration class. Protected multi-User eligibility has deterministic acquisition order.

**PASS.**

### Disclosure challenge

Mention autocomplete, Notification list/counts, batch seen and direct engagement all reapply current server-side disclosure. No client-side hide, copied ACL or notification-as-access-token remains.

**PASS.**

### Async/realtime challenge

Persistent Notification truth is independent of wake-up. River remains reserved for required durable future work. SSE loss is recoverable by canonical GET. In-process hub is an admitted one-replica mechanism, replaceable by PostgreSQL LISTEN/NOTIFY if a multi-replica consumer later appears.

**PASS.**

### API/census challenge

No material evidence found a smaller operation family without collapsing independent behaviors or creating generic `/actions` semantics. The candidate remains exactly:

```text
Discussion / Mention     +3
Notification state       +4
SSE invalidation         +1
                         ---
                         +8
78 -> 86 operations
10 -> 11 Idempotency-Key creations
```

**PASS pending independent challenge and later executable-wire consolidation.**

### YAGNI / framework challenge

No notification SaaS runtime, generic EventBus, broker, Redis or persistent Lexical format is introduced. Lexical is a replaceable UI mechanism only; Watermill/LISTEN-NOTIFY remain trigger-bound future candidates.

**PASS.**

## 6. Lead GCR Round-2 verdict

```text
VERDICT: CONVERGED
MATERIAL: 0
IMPORTANT: 0
OPTIONAL: 0
```

No Lead finding changes the candidate result:

```text
owners                4+2 candidate
routes                 11 candidate
PermissionCode         16 candidate
operations             86 candidate
Idempotency creations  11 candidate
ETag domains           13/13
exact-byte resources   4
```

This is still **candidate evidence**, not promoted Product/T1→T9 authority.

## 7. Required next gate

Because this package creates/moves a semantic authority and binds multiple architecture layers, METHOD requires a fresh independent challenge.

```text
exact corrected candidate HEAD + green CI
→ isolated review/<gate>-fable branch
→ delta = docs/work/current/ai-dialog.md only
→ Fable reconstructs authority and attacks the whole coherent package
→ reviewer output = Evidence only
→ Lead adjudicates every material/important finding
→ Round 2 only if a real contradiction survives
→ only then upstream consolidation
```
