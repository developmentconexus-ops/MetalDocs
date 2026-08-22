# T11 — Discussion / Mention / Notifications Global Coherence Review

> **Status:** LEAD GCR — NOT CONVERGED / OPERATOR ADJUDICATION REQUIRED.  
> **Scope:** the operator-ratified D0→D8 bounded reopen candidate only.  
> **Method:** DevelopmentConexus Engineering Method v1.0.0.  
> **Implementation:** BLOCKED.  
> **Independent review:** NOT YET PERFORMED; this Lead GCR is not independent challenge.

## 1. Inputs

Candidate package:

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

Primary current authorities checked:

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

No current GCR evidence invalidates:

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

No operation-count change is proposed by this GCR.

## 3. MATERIAL M1 — Authorization ownership was blurred inside D7

### Finding

D7 currently summarizes the transaction as if Controlled Documents validates Mention target admissibility. That is not authority-coherent.

Current T8-C law is explicit:

```text
Organization authors current User/Group facts
resource owner authors relationship/state predicate facts
Authorization alone computes final ALLOW/default-DENY
application maps/routes/co-ordinates facts
```

A Mention target being able to receive/read the exact Document Discussion is an Authorization/disclosure conclusion, not intrinsic Controlled Documents truth.

### Required correction

For create-message commit:

```text
application/documentofficial
  → gather current author + unique target SecuritySubjects from Organization
  → use protected in-Scope eligibility for author + Mention targets
  → gather exact Controlled Documents access/predicate facts
  → Authorization.DecideIn / DecideManyIn computes final current decisions
  → only after ALLOW for author + every Mention target:
       Controlled Documents validates intrinsic message/reply/content semantics and inserts Message/Mention
       Notifications inserts one Notification per unique Mention target
```

No `document.read_discussion` Permission is added. D1 remains binding:

```text
Discussion read eligibility
= exact current ability to receive the Document Official / Discussion lens
```

The canonical B03 disclosure composition is reused; Notification/Mention callers may not invent a parallel role matrix.

### Offboarding concurrency

Create DiscussionMessage now joins the T3 family whose correctness depends on ENABLED User truth.

```text
author
+ every unique Mention target
```

must serialize eligibility with offboarding through the existing protected Organization subject mechanism while the local Scope remains open.

When multiple Users need protected eligibility, acquire/resolve them in deterministic `user_id ASC` order (after uniqueness) so the new multi-User path does not introduce avoidable deadlock cycles.

RoleAssignment/access drift after a valid current decision remains handled by D5 presentability/current reauthorization; Mention never preserves later access.

### Classification

```text
MATERIAL — authority + security/concurrency boundary
```

The Product decision survives; D7 placement/enforcement requires correction.

## 4. MATERIAL M2 — Notification disclosure must compose before public pagination/counts

### Finding

D5 requires a Notification to disappear completely from items/badge/counts when its source is no longer currently disclosable. Notifications itself does not own source disclosure. A naive implementation:

```text
Notifications pages rows
→ application/frontend filters inaccessible sources
```

would break the current authority model and can produce sparse/incorrect pages, incorrect badge counts and source-existence/cardinality leakage.

Persisting a copied `presentable`/ACL flag in Notifications would instead create duplicate Authorization authority.

### Required correction

`application/notifications` owns the cross-owner read choreography, not semantic truth.

Conceptual current read:

```text
Notifications candidate scan for current recipient + engagement filter, in canonical order
→ batch source identities
→ Controlled Documents batch source/access facts
→ Organization current subject facts
→ Authorization exact Decide/DecideMany
→ retain currently presentable candidates
→ continue candidate scan until requested presentable page (+ lookahead for has_more) is satisfied or source candidates are exhausted
→ compose source presentation only for presentable rows
```

Public cursor semantics are over the canonical Notification ordering and never expose hidden candidate identities. The response cannot rely on frontend post-filtering.

The same current-disclosure composition owns exact first-page unseen/unread summary counts. Launch may evaluate these in bounded internal chunks; no durable copied counter or current-access projection becomes authority. A measured scale failure is an explicit reopen trigger for a rebuildable optimization that proves equivalence.

`mention-candidates` follows the analogous pattern in bounded search form: Organization search is only a candidate source; current Document disclosure is applied server-side before results are returned.

### Classification

```text
MATERIAL — disclosure + pagination/current-authority correctness
```

No new operation or owner is required.

## 5. MATERIAL M3 — SSE/wake-up call graph must preserve the one semantic inbound door

### Finding

T8-B forbids `transport -> platform` and requires:

```text
transport -> application
```

for every semantic application route. A direct SSE handler subscribing to an in-process platform hub would violate that topology even though the hub is non-semantic.

The realtime mechanism must also update multiple tabs after Notification engagement mutations, not only after creation.

### Required correction

```text
GET /api/v1/notifications/events
transport/http
→ application/notifications stream choreography
→ narrow consumer-owned subscription port
→ platform realtime implementation
```

The platform mechanism owns only ephemeral connection/subscription mechanics.

Wake-up law:

```text
successful Notification creation or engagement mutation commits
→ after commit, relevant application leaf invokes narrow Notification-changed wake-up mechanism for that recipient
```

This includes:

```text
DOCUMENT_MENTION Notification creation
mark seen
mark read
mark unread
archive
unarchive
mark all read
```

Wake-up remains best-effort and non-durable. Semantic owners never invoke the realtime mechanism directly. A wake-up failure never rolls back committed Product state.

Different application leaves may own narrow technical ports to the same platform implementation; do not create a generic application EventBus merely to share the mechanism.

### Classification

```text
MATERIAL — accepted dependency-direction / runtime boundary
```

SSE and the in-process Launch mechanism remain selected.

## 6. IMPORTANT I1 — batch seen must be race-safe without becoming a disclosure oracle

`markNotificationsSeen(ids[])` is generated from Notifications that were visible to the client, but access may drift before the request arrives.

Required law:

```text
unique bounded request ids
→ intersect with current recipient's currently presentable Notifications
→ set seen_at only on that intersection
→ absent / foreign-recipient / now-non-presentable ids produce no mutation and no per-id status
→ response discloses no existence/cardinality detail for skipped ids
```

A direct single-Notification engagement operation on an item that is no longer presentable returns the normal non-disclosing not-found behavior and the client refetches.

## 7. IMPORTANT I2 — completed create-message replay must not re-run historical Mention eligibility

The current global idempotency law rechecks current caller AuthZ/disclosure before completed replay, then returns the stored historical success without re-running historical lifecycle/preconditions.

For `createDocumentDiscussionMessage`:

```text
current caller session / CSRF / current document.discuss + source disclosure
→ completed replay recognition
→ same stored message_id
→ zero new Message
→ zero new Notification
```

Do not re-evaluate the old Mention targets' current eligibility on completed replay. Their later access/offboarding state is governed by D5 presentability; replay is not a second Mention command.

The semantic fingerprint includes the exact Document, optional reply target and ordered normalized Text/Mention content. ReplaySnapshot remains free of message text and stores only stable success identity.

## 8. IMPORTANT I3 — Discussion and Notifications do not become duplicate Audit/History streams

Launch baseline:

```text
DiscussionMessage
  immutable trusted author/time/content authority
  -> no duplicate semantic AuditEvent merely to copy the message
  -> not inserted into Document lifecycle History timeline

Notification creation/engagement/realtime
  -> no mandatory semantic AuditEvent in Launch
```

This mirrors the existing SubmissionFeedback principle: an immutable owner record with its own actor/time/content does not need a second Audit copy solely because it exists.

A future regulatory requirement for messaging/notification audit is a T3 reopen trigger.

## 9. IMPORTANT I4 — persistence enforcement required by the new invariants

Upstream T8-D consolidation must at minimum preserve:

```text
new notifications.* owner namespace
identity-only cross-owner references only
unique DOCUMENT_MENTION Notification per recipient + message
read_at present -> seen_at present
archived_at present -> seen_at present
immutable DiscussionMessage/Mention accepted state by application privileges/structure
reply reference cannot cross Document Discussion
```

Exact table decomposition remains T8-D/T11 realization, not this GCR.

## 10. IMPORTANT I5 — runtime/OpenAPI proof remains a closure gate, not Product authority

Operation 86 remains part of the candidate `/api/v1` application census only if the selected Go OpenAPI boundary can prove server-side `text/event-stream` realization without a manual parallel route/DTO registry.

Until that proof exists:

```text
SSE = accepted Product/UX mechanism candidate
manual contract escape hatch = forbidden
```

Failure of the chosen generator is a bounded mechanism/toolchain reopen; it does not justify silently dropping realtime or moving the route outside authority.

T8-G consolidation must additionally own heartbeat/flush/proxy-timeout/shutdown/resource-limit behavior appropriate to a long-lived SSE response while preserving one application runtime baseline.

## 11. IMPORTANT I6 — visual and upstream sequencing remains binding

Even after semantic corrections converge:

```text
1. consolidate Product/T1/T3/T5/T6/T8-B→G/T9 candidate deltas coherently
2. render the smallest B01 P8 notification-header/Quick-Inbox reopen
3. operator re-LOCKs only the implicated B01 delta
4. update B03 P8 with real Discussion semantics
5. operator adjudicates/LOCKs B03
```

No B04+ baseline opens while this material dependency remains unresolved.

## 12. Lead verdict

Current candidate before adjudication:

```text
MATERIAL   3
IMPORTANT  6
```

No finding currently changes the candidate Product capability set or census:

```text
owners                4+2 candidate
routes                 11 candidate
operations             86 candidate
Idempotency creations  11 candidate
ETag domains           13/13
exact-byte resources   4
```

The three MATERIAL findings are bounded authority/enforcement corrections. If they are accepted and applied, rerun this Lead GCR over the corrected package. Only a converged corrected package may proceed to the METHOD-required fresh independent challenge and then upstream consolidation/visual reopens.
