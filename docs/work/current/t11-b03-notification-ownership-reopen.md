# T11 — Notification Ownership Bounded Reopen

> **Status:** OPERATOR-RATIFIED CANDIDATE / FABLE-CONVERGED / PENDING UPSTREAM CONSOLIDATION.  
> **Parent:** `t11-b03-discussion-notification-mini-design.md`.  
> **Reasoning authority:** `developmentconexus-ops/conexus-methodology/METHOD.md` — DevelopmentConexus Engineering Method v1.0.0.  
> **Implementation:** BLOCKED.  
> **Current upstream authority remains effective until the full bounded reopen is consolidated.**

## 1. Material trigger

Current Launch authority classified Notifications as non-semantic/deferred because no concrete Launch consumer or independent lifecycle had been proven.

The operator has now required before Launch V1:

```text
stable-Document Discussion
+ explicit @mention
+ persistent in-app Notification
+ Inbox engagement state
```

The approved Inbox direction includes independent recipient-controlled state such as seen/read/archive. This is new material evidence and satisfies the existing ownership reopen trigger: a previously deferred capability has gained a concrete consumer and an independent lifecycle.

## 2. Target invariant

```text
A Notification is a persistent unit of user attention about an already-valid source fact.

It must never:
- create or replace the source fact;
- grant or preserve access to the source;
- own Document lifecycle, DiscussionMessage content, User identity or Authorization;
- become equivalent to Document read/acknowledgement/governance evidence.
```

Its own semantic lifecycle is limited to Notification identity, recipient, closed source/kind reference, creation time, deduplication and recipient engagement state.

## 3. Alternatives

### A — Controlled Documents owns Notification

Rejected as a **Local Maximum**. It is convenient for the first `DOCUMENT_MENTION` consumer but makes document authority own recipient Inbox engagement (`seen/read/archive`) that remains meaningful even if a future Notification source is unrelated to Controlled Documents.

### B — Organization owns Notification

Rejected. Organization owns who exists and how people are organized. Notification attention state is not organizational identity/lifecycle merely because it references a User.

### C — Notification is a rebuildable projection/mechanism

Rejected once recipient-controlled engagement exists. `seen/read/archive` cannot be reconstructed from the source Mention alone; moving those facts elsewhere would simply create a hidden second Notification authority.

### D — supporting semantic owner `Notifications`

**SELECTED / OPERATOR-RATIFIED.**

Notifications owns its independent attention lifecycle while source meaning remains with the producing business owner.

## 4. Global Maximum decision

Launch topology becomes, after bounded upstream consolidation:

```text
BUSINESS SEMANTIC
1. Authentication
2. Organization
3. Authorization
4. Controlled Documents

SUPPORTING SEMANTIC
5. Audit
6. Notifications
```

`Notifications` is supporting rather than business-semantic because it does not establish the underlying business fact. It preserves user-attention state about facts owned elsewhere.

This promotion does **not** imply:

```text
microservice
separate database
Kafka / RabbitMQ / NATS
Redis
notification SaaS
workflow engine
generic EventBus
generic polymorphic activity platform
```

The accepted physical posture remains one modular monolith + one PostgreSQL product-state database unless later evidence independently reopens it.

## 5. Ownership boundary

Notifications owns:

```text
Notification identity
recipient User reference
closed Notification kind
closed source reference
created_at
deduplication identity
seen/read/archive engagement semantics
Inbox ordering/current engagement state
```

Notifications does not own:

```text
Document identity/title/lifecycle
DiscussionMessage content/validity
Mention target eligibility
User/Profile identity
Authorization/disclosure
governance/audit history
email/push delivery
realtime transport
River jobs
event bus
```

Launch starts with the concrete source family:

```text
DOCUMENT_MENTION
  document_id
  message_id
```

Do not create an `any source_type + source_id + arbitrary JSON` semantic platform. A later source expands a closed union only after a real consumer is approved.

## 6. Cross-owner composition

No owner imports another semantic owner.

```text
application
  ├── Authorization
  ├── Controlled Documents
  └── Notifications
```

For accepted explicit Mentions, the leading invariant is:

```text
accepted Mention
<=>
required in-app Notification exists
```

Application may therefore compose Controlled Documents + Notifications in the same caller-owned local PostgreSQL transaction when the final reopen confirms that this is the smallest enforcement of the invariant.

A broker is not required merely to express that a source fact occurred. Typed domain-event semantics and transport/event-bus mechanism remain separate decisions.

## 7. Structural inversion / adversarial result

If the first Notification source had been Organization, Access, or another future owner rather than Controlled Documents, the following would still be true:

```text
Notification recipient identity
seen/read/archive lifecycle
Inbox ordering
source reference
no access grant
no ownership of the source fact
```

That survival under inversion is evidence that Notification lifecycle does not belong to Controlled Documents.

Strongest objection:

> only one current Notification kind exists, so a new owner may be premature.

Adjudication:

The owner is justified by the **current independent lifecycle**, not hypothetical future source count. A source-derived projection with no mutable engagement would not justify promotion; the approved persistent Inbox does.

## 8. METHOD outcome

```text
OUTCOME: RESTRUCTURE NOW — bounded ownership reopen
```

This means restructure the target authority now during planning, not implement Product code. The old 4+1 authority remains effective until the full reopen delta is coherently consolidated and ratified across owning documents.

## 9. Required downstream coherence work

The bounded reopen implicates the smallest coherent current-authority set:

```text
Product Contract                      Notification/Inbox = Launch V1
T1 / domain model                     Notification semantic state
Ownership                             4+1 -> 4+2
T3                                   document.discuss + Notification self-access rules
T5                                   async/realtime/event-delivery classification
T6                                   Discussion/Inbox journeys + 11 stable SPA routes
T8-B                                 Notifications owner realization
T8-C                                 Notifications owner contract + cross-owner transaction composition
T8-D                                 Notification persistence/concurrency
T8-E                                 exact API/schema/census delta
T8-F                                 Inbox/shell/frontend realization
T8-G                                 realtime runtime
T9                                   new golden-flow/validation obligations
T11                                  B01 smallest-scope reopen + B03 Discussion

docs/decisions/forward-obligations.md
  ASY-02 DEFERRED must be refined/superseded by the ratified Launch requirement

docs/decisions/api-operation-census.md
  78-operation authority must be promoted to the exact 86-operation bounded-reopen census
```

The two decision-register targets above are mandatory parts of the same consolidation so the repository cannot retain a stale `DEFERRED Notifications` disposition or stale `78` census beside promoted Product/Architecture authority.

Only the actually implicated decisions may be reopened during consolidation.

## 10. Independent challenge result

Fresh Fable review on Evidence PR #165 independently re-falsified the whole D0→D8 package and Lead GCR corrections.

```text
VERDICT: CONVERGED
MATERIAL: 0
IMPORTANT: 1
OPTIONAL: 3
```

The sole IMPORTANT finding was the missing explicit decision-register targets corrected in §9 above. The operator accepted it. It changes no Product capability, owner count, route count, Permission count, operation count or mechanism choice.

The operator also accepted three bounded wire/consolidation precisions: exact `anchor_message_id` continuation semantics, explicit message/fan-out bounds, and one named canonical Discussion-disclosure predicate reused by Discussion read/Mention/presentability.

No Fable Round 2 is justified because no material contradiction survived and the IMPORTANT finding is a consolidation-completeness correction only.

## 11. Reopen triggers

Reopen this ownership decision if evidence proves one of:

```text
Notification has no independent persistent lifecycle after final UX closure;
a simpler existing owner can own the entire lifecycle without unrelated authority;
implementation evidence shows the owner boundary causes more cross-owner accidental complexity than it prevents;
a materially different Product requirement changes Notification into a pure rebuildable projection;
a new trust/deployment boundary independently justifies a different topology.
```

Preference for fewer packages/owners, sunk cost, framework shape, or broker availability are not reopen evidence.
