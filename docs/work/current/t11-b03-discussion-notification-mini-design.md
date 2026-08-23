# T11 — B03 Discussion / Mention / Notification Mini-Design

> **Status:** D0→D8 OPERATOR-RATIFIED / GCR-CONVERGED / FABLE-CONVERGED / PENDING UPSTREAM CONSOLIDATION.  
> **Parent:** B03 — Document Official / Ficha do Documento.  
> **Purpose:** close the smallest Product/UX semantics required by the operator-mandated Launch V1 Document Discussion + `@mention` + in-app Notification capability before upstream consolidation.  
> **Implementation:** BLOCKED.  
> **Current census authority:** 78 operations / 10 Idempotency-Key creations until the bounded reopen is fully consolidated and promoted.  
> **Converged candidate census:** 86 operations / 11 Idempotency-Key creations.  
> **Reasoning authority:** `developmentconexus-ops/conexus-methodology/METHOD.md` — DevelopmentConexus Engineering Method v1.0.0.

## 1. Analysis law

Every decision in this mini-design follows the canonical METHOD decision core proportionally:

```text
Evidence
→ Known / Inferred / Unknown / Deferred
→ Root Cause
→ Target Invariant
→ Constraints
→ Credible Alternatives
→ Local Maximum vs Global Maximum
→ Essential vs Accidental Complexity
→ YAGNI / Overengineering / Future Cost
→ Authority / Boundary when relevant
→ Enforcement
→ Proof Strategy
→ Adversarial Challenge
→ Decision
→ Reopen Triggers
```

Repository-local application:

```text
current MetalDocs authority first
→ real user need / named consumer
→ alternatives independent of current implementation shape
→ legacy/reference evidence only as falsification/input, never authority
→ structural inversion/adversarial test
→ Global Maximum comparison
→ smallest sustainable semantic change that closes the need
→ bidirectional impact across Product / T1–T10 / frontend / API
→ operator adjudication
→ upstream consolidation only after the mini-design closes
```

METHOD laws particularly binding here:

```text
mechanism != authority
YAGNI must remove unsupported capability, not required invariants or justified seams
prepare the seam, not the entire future capability
new real consumer / changed requirement is a valid bounded-reopen trigger
Global Maximum is not maximum abstraction or infrastructure
```

Technology/framework/library choice was intentionally deferred until Product semantics were frozen. A framework may implement the accepted model; it may not define the model.

## 2. Fixed Product requirement

Operator-required for Launch V1:

```text
stable-Document human discussion
+ explicit @mention of MetalDocs Users
+ in-application Notification for mention
```

Semantic separation:

```text
Document Discussion != DRAFT/editor comments
Document Discussion != SubmissionFeedback
Document Discussion != GovernanceCase feedback
Notification != access grant
Notification != lifecycle/history authority
```

Legacy evidence supports the user need: the published-document backlog identified a separate display-side `CommentsCard` / `discussion` model because editor comments were not appropriate for the published Document page. Legacy also contained a Notifications feature. Neither legacy model is current authority.

## 3. D0 — delivery channel baseline — OPERATOR-RATIFIED

**Decision:** Launch V1 `@mention` requires **in-app Notification only**.

```text
in-app notification   REQUIRED
email                 NOT Launch V1 baseline
push                  NOT Launch V1 baseline
```

Later channels may attach to the same accepted Notification intent without redefining Document Discussion.

## 4. D1 — Discussion read/write authorization — OPERATOR-RATIFIED

### Read

Reading Discussion follows the actor's current ability to receive/open the Document Official / Discussion lens. Discussion is not an access-grant mechanism.

```text
current Document disclosure/access required
notification cannot preserve or expand access
loss of access -> discussion becomes non-disclosable even if an old notification exists
```

T3/T6 consolidation must promote one named canonical Discussion-disclosure predicate and reuse it for Discussion read, Mention discovery/commit validation and Notification presentability.

### Write / reply / mention

A distinct Product Permission is required:

```text
document.discuss
```

Writing requires:

```text
enabled User
+ current document.discuss grant in matching scope
+ current Document access/disclosure
+ accepted Discussion-specific state predicate
```

Approved role-bundle delta:

```text
viewer              + document.discuss
author              + document.discuss
approver            + document.discuss
area_manager         + document.discuss

governance_viewer   no document.discuss by default
governance_admin    no document.discuss by default
```

Commands always recheck current authorization; frontend affordance presence is UX guidance only.

## 5. D2 — message / reply semantics — OPERATOR-RATIFIED

**Decision:** Discussion is one chronological linear timeline over the stable Document. A message may optionally reference one prior message, but replies do not create a separate semantic Thread aggregate or an arbitrarily nested tree.

```text
DocumentDiscussionMessage
  message_id
  document_id
  author_user_id
  created_at
  body
  reply_to_message_id?
  official_revision_at_post?
```

Binding laws:

```text
message belongs to stable Document identity
reply remains an ordinary message in the same chronological timeline
reply_to_message_id, when present, must reference a message in the same Document Discussion
no separate Thread owner/lifecycle is introduced
no semantic nesting depth exists
```

The server records the official Revision that existed when the message was accepted, when one exists; otherwise the contextual revision reference is absent. This snapshot is provenance/context only and never moves Discussion ownership from Document to Revision or binds it to WorkingContent/DRAFT.

## 6. D3 — @mention discovery + accepted-message validation — OPERATOR-RATIFIED

`@mention` is server-derived, Document-scoped and disclosure-safe.

Candidate eligibility:

```text
existing MetalDocs User
+ same Company
+ ENABLED
+ currently eligible to receive/read this exact Document Discussion
+ candidate != message author
```

`document.discuss` is deliberately **not** required of the mentioned User. A read-only actor may be mentioned without gaining reply authority.

### Purpose-built discovery

```text
@bea
→ server-side search within currently mention-eligible Users for this Document
→ bounded UserReference-like results
```

No administrative User directory, email, Role/Permission set or exclusion reason is exposed merely to populate the composer.

### Stable mention identity

A Mention is not authoritative merely because text contains `@Name`. The accepted message carries stable `user_id` semantics:

```text
MessageContent
  = Text
  | Mention(user_id)
```

The selected Lexical mechanism never becomes persisted Product format.

### Accepted-message revalidation

Autocomplete results are UX guidance only. Message acceptance rechecks current truth atomically. Organization authors User/ENABLED facts, Controlled Documents authors exact Document predicate facts, Authorization alone computes ALLOW/default-DENY, and application coordinates.

Author + every unique Mention target use protected eligibility serialized with offboarding in deterministic `user_id ASC` order.

If any Mention becomes invalid:

```text
reject whole message command
zero DiscussionMessage
zero Notification
preserve local composer input for explicit reconciliation
```

### Notification trigger law

```text
explicit accepted Mention -> one in-app Notification per target/message
same User mentioned multiple times -> at most one Notification
author self-mention -> not admitted
reply without explicit Mention -> no Notification solely because it is a reply
reply + explicit Mention -> normal Mention notification
```

Mention never grants access, creates a governance participant, changes lifecycle state or creates a persistent Discussion member.

## 7. D4 — Notification ownership + engagement — OPERATOR-RATIFIED

Notifications becomes the second supporting semantic owner after consolidation.

```text
BUSINESS
Authentication
Organization
Authorization
Controlled Documents

SUPPORTING
Audit
Notifications
```

Notification engagement has independent dimensions:

```text
seen_at?       monotonic first presentation
read_at?       reversible read/unread state
archived_at?   reversible Inbox placement
```

Laws:

```text
READ => SEEN
mark unread preserves seen
archive/unarchive preserves seen/read
archive != delete
bell badge = presentable + non-archived + unseen
unread filter/count = presentable + non-archived + unread
```

Launch includes per-item read/unread/archive/unarchive and mark-all-read, but not mark-unseen, mark-all-unread, bulk archive, snooze, priority or preferences.

## 8. D5 — disclosure / offboarding / immutability — OPERATOR-RATIFIED

Launch DiscussionMessage is immutable after acceptance:

```text
edit     absent
remove   absent
correction = new DiscussionMessage
```

Stable User identity remains historical; erasable UserProfile enrichment is not copied as authority.

Notification persists independently of current presentability. It is presentable only while recipient is current ENABLED User and the exact source remains currently disclosable. Non-presentable items are omitted server-side from Inbox, archive, counts, bulk engagement targets and realtime presentation. Loss/restoration of source disclosure does not rewrite engagement history.

No ordinary hard-delete Product operation exists for DiscussionMessage or Notification. Future retention/privacy/redaction is a material reopen.

## 9. D6 — B01 shell impact — OPERATOR-RATIFIED

Smallest B01 reopen:

```text
utility header
  + Notification bell
  + unseen badge
  + Quick Inbox

sidebar unchanged

stable full Inbox route
  /notifications
```

Notifications does not move into `Minha Caixa`; assigned work and attention remain distinct mental models.

Desktop uses a Quick Inbox popover; narrow/mobile uses an accessible sheet/full-screen material surface. A rendered P8 delta and operator re-LOCK are still required before the B01 structural change becomes locked visual authority.

## 10. D7 — exact contract/census — OPERATOR-RATIFIED / GCR+FABLE-CORRECTED

Candidate API delta:

```text
79 GET  /api/v1/documents/{document_id}/discussion/messages
80 POST /api/v1/documents/{document_id}/discussion/messages
81 GET  /api/v1/documents/{document_id}/discussion/mention-candidates
82 GET  /api/v1/notifications
83 PATCH /api/v1/notifications/{notification_id}/engagement
84 PUT   /api/v1/notifications/seen
85 PUT   /api/v1/notifications/read
86 GET   /api/v1/notifications/events   text/event-stream
```

```text
operations             78 -> 86
Idempotency-Key create 10 -> 11
ETag domains            13/13 unchanged
exact-byte resources    4 unchanged
```

Message + Mention + required Notification commit in one caller-owned PostgreSQL Scope. River/outbox/EventBus do not mediate this local biconditional.

Notification list/count presentability is composed server-side **before** public pagination/counts. `anchor_message_id` is one first-page navigation filter of the Discussion list operation with continuation under the same cursor authority. T8-E must set explicit segment-count and unique-Mention-target bounds in addition to the global JSON ceiling.

SSE is invalidation only:

```text
transport -> application/notifications -> narrow realtime port -> mechanism
```

Wake-up occurs only after committed Notification creation/engagement changes and may be lost without losing Product truth.

## 11. D8 — technology/reuse — OPERATOR-RATIFIED / FABLE-CONFIRMED

```text
Discussion composer      Lexical core + @lexical/react; custom Mention node/typeahead
Inbox frontend           MetalDocs + generated OpenAPI types + TanStack Query
browser realtime         native EventSource
server realtime          narrow Go stdlib SSE adapter
Launch wake-up           in-process coalescing recipient hub
multi-replica candidate  PostgreSQL LISTEN/NOTIFY
required durable work    River
current generic EventBus absent
external broker          absent
Redis baseline           absent
future event candidate   Watermill only on real multi-consumer trigger
```

No vendor/runtime/framework owns MetalDocs Notification semantics.

## 12. GCR + independent challenge

Lead GCR:

```text
Round 1  NOT CONVERGED  MATERIAL=3 / IMPORTANT=6
Round 2  CONVERGED      MATERIAL=0 / IMPORTANT=0
```

Fresh Fable Evidence PR #165:

```text
VERDICT: CONVERGED
MATERIAL: 0
IMPORTANT: 1
OPTIONAL: 3
UNSUPPORTED_PREFERENCE: 0
```

F-1 (missing decision-register consolidation targets) and O1→O3 were operator-accepted. No Fable Round 2 is justified.

Fable explicitly confirmed survival of:

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

## 13. Consolidation gate

Current integrated authority remains `4+1 / 10 routes / 78 ops / 10 Idempotency-Key creates` until all implicated current owners are changed coherently.

Mandatory consolidation includes Product/T1/Ownership/T3/T5/T6/T8-B→G/T9/T11 plus:

```text
docs/decisions/forward-obligations.md
  ASY-02 DEFERRED -> refined/superseded by Launch requirement

docs/decisions/api-operation-census.md
  78 -> 86 with bounded-reopen provenance
```

Only after whole-current-authority coherence may `4+2 / 11 / 86 / 11` be promoted from candidate to current authority.

No B04+ baseline opens while the B01 notification P8 re-lock and B03 lock remain unresolved.
