# Fable Review — T11 Discussion / Mention / Notifications bounded reopen

## Review target

Repository: `developmentconexus-ops/MetalDocs`  
Candidate branch: `arch/t11-implementation-program`  
Exact candidate HEAD: `a9047924aa2e31aaa1418a15c8786b7e9ad2967f`  
Required candidate CI: `#1263 SUCCESS`  
Review branch: `review/t11-notifications-fable`

Follow the canonical DevelopmentConexus Fable workflow and METHOD v1.0.0.

Read repository authority in the normal bounded order:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ the smallest owning authority/work package needed for this review
```

Primary candidate package:

```text
docs/work/current/t11-b03-discussion-notification-mini-design.md
docs/work/current/t11-b03-notification-ownership-reopen.md
docs/work/current/t11-b03-notification-engagement.md
docs/work/current/t11-b03-discussion-notification-d5.md
docs/work/current/t11-b01-notifications-reopen.md
docs/work/current/t11-b03-discussion-notification-d7-contract.md
docs/work/current/t11-b03-notification-technology-spike.md
docs/work/current/t11-notifications-global-coherence-review.md
```

Current upstream authorities should be read only as required to challenge the candidate, especially:

```text
docs/product/contract.md
docs/product/journeys.md
docs/architecture/ownership.md
docs/architecture/domain-model.md
docs/architecture/authorization-and-audit.md
docs/architecture/async-and-search.md
docs/architecture/backend.md
docs/architecture/interfaces.md
docs/architecture/persistence.md
docs/architecture/wire-contract.md
docs/architecture/frontend.md
docs/architecture/runtime.md
```

The current integrated authority remains 4 business + Audit, 10 SPA routes, 78 application operations, 10 Idempotency-Key creations. The package under review is a bounded material reopen candidate; do not treat its new counts as already promoted authority.

## Candidate result to attack

```text
Product
  stable-Document Discussion
  immutable DiscussionMessage
  chronological timeline + optional one-message reply reference
  semantic Mention(user_id)
  in-app Notification for explicit Mention
  no Launch email/push

Authorization
  new document.discuss Permission for writing
  reading Discussion follows current Document disclosure
  Mention targets need current ability to receive/read exact Discussion context

Ownership
  Notifications becomes second supporting semantic owner
  4 business + 2 supporting owners

Inbox
  seen_at monotonic
  read_at reversible
  archived_at reversible
  unseen badge
  unread filter/count
  archive/unarchive
  mark all read

Frontend IA
  global bell + Quick Inbox
  stable /notifications route
  sidebar unchanged

Wire candidate
  +3 Discussion/Mention operations
  +4 Notification state operations
  +1 SSE invalidation operation
  78 -> 86 application operations
  10 -> 11 Idempotency-Key creations
  13/13 ETag domains unchanged
  4 exact-byte resources unchanged

Atomicity
  accepted explicit Mention <=> required DOCUMENT_MENTION Notification exists
  one local PostgreSQL Scope coordinated by application
  Authorization alone decides final ALLOW/DENY
  protected author + target eligibility serializes with offboarding

Disclosure
  Notification source disclosure is recomputed server-side
  presentability is applied before public pagination/counts
  no copied ACL/presentable authority

Realtime
  SSE is best-effort invalidation only
  transport -> application -> narrow realtime mechanism
  in-process coalescing hub for one-replica Launch
  wake only after committed Notification state changes
  River remains the only durable future-work mechanism
  no generic EventBus / external broker / Redis baseline

Mechanisms
  Lexical core + @lexical/react for PlainText + custom Mention node/typeahead
  Lexical state never persists as Product truth
  native MetalDocs Inbox over OpenAPI + TanStack Query
  native browser EventSource
  narrow Go stdlib SSE realization
  LISTEN/NOTIFY only as a future multi-replica candidate
  Watermill only as a future EventStore candidate if real multi-consumer pressure appears
```

## Lead GCR history

Lead GCR Round 1 found `MATERIAL=3 / IMPORTANT=6`; operator approved all corrections. Corrected D7/technology records now encode them. Lead GCR Round 2 says `CONVERGED / MATERIAL=0 / IMPORTANT=0`.

Do **not** accept that self-review on trust. Reconstruct and attack it independently.

The three former material classes were:

```text
M1 Authorization ownership + offboarding serialization
M2 current disclosure before Notification pagination/counts
M3 SSE call graph + post-commit invalidation
```

The former important classes covered batch-seen disclosure, completed idempotency replay, Audit/History duplication, persistence constraints, OpenAPI SSE proof and visual/upstream sequencing.

## Mandatory adversarial questions

### 1. Root cause / Product boundary

Is stable-Document Discussion actually the smallest coherent product concept, or is it accidentally recreating DRAFT editor comments, SubmissionFeedback, chat, activity feed or generic collaboration?

Does Discussion ownership at stable Document survive Revision changes and first-release/no-release states correctly?

### 2. Notifications ownership

Does `seen/read/archive` truly justify a second supporting semantic owner, or is this an unnecessary owner boundary? Test Controlled Documents, Organization and rebuildable-projection alternatives seriously.

If Notifications is a real owner, is the proposed boundary complete enough without becoming a generic notification platform?

### 3. Authorization/disclosure

Attack the exact split:

```text
Organization -> subject facts
Controlled Documents -> resource predicate facts
Authorization -> final ALLOW/DENY
application -> choreography
```

Look for hidden second permission matrices in Mention autocomplete, Discussion reads, Notification presentability, batch seen, direct engagement, Quick Inbox or SSE.

Test offboarding/access races and whether deterministic protected-user acquisition is actually sufficient under current T3/T8-D laws.

### 4. Cross-owner transaction

Try to falsify the same-Scope Mention -> Notification invariant.

Is synchronously creating Notifications in the same PostgreSQL transaction the Global Maximum, or would an outbox/EventStore/River subscriber produce lower total complexity or better recovery?

Conversely, would introducing an event bus here be accidental complexity?

### 5. Pagination/count correctness

Attack the candidate-scan -> batch disclosure -> continue-until-presentable-page algorithm.

Can it preserve deterministic cursor semantics, `has_more`, unseen/unread counts, non-disclosure and bounded work without copied access authority?

Look for denial-of-service/scale traps or observable cardinality leaks. If optimization is required now, prove the current candidate structurally insufficient rather than assuming future scale.

### 6. Idempotency

Attack the 11th Idempotency-Key operation.

Does completed replay correctly recheck only current caller authorization/disclosure without rerunning historical Mention-target eligibility? Is the fingerprint sufficient and free of replay ambiguity? Can retries ever duplicate Notification or expose now-forbidden source truth?

### 7. Message immutability / retention

Is no-edit/no-delete the correct Launch Global Maximum after mentions/replies/notifications, or does it create unreasonable user harm? If edit/delete is required, specify the smallest semantics that preserve truthful notifications/replies.

Check offboarding/profile erasure and future retention/privacy compatibility.

### 8. Inbox engagement

Attack `seen_at`, `read_at`, `archived_at` as independent dimensions, badge=`unseen`, unread reversible, archive reversible, no mark-unseen.

Look for contradictory states, concurrency needs, duplicate counters or conflict with future Read & Acknowledge.

### 9. API surface / census

Try to reduce the candidate 8 new operations without:

```text
screen-shaped APIs
generic /actions
frontend AuthZ
hidden semantics
per-item network explosion
loss of deep-link/realtime behavior
```

Also test whether any missing operation/read model makes the 86 count incomplete.

### 10. SSE / runtime

Attack whether SSE belongs in `/api/v1` application census, whether the OpenAPI/codegen proof is realistic, and whether the call graph preserves the one semantic inbound door.

Challenge native EventSource, Go stdlib SSE, heartbeat/reconnect/resource limits, one-replica in-process hub, multi-tab behavior and post-commit wake-up.

Could polling be materially smaller? Could WebSocket be required? Could LISTEN/NOTIFY be needed now? Require evidence.

### 11. Technology choices

Challenge Lexical against Tiptap/ProseMirror, a smaller contenteditable/textarea solution, react-mentions-style libraries and custom editor code.

Challenge native Inbox against Novu/Knock/MagicBell or another mature reusable library/service. Mechanism reuse must not import competing Product authority.

Challenge the absence of Watermill/EventBus/broker/Redis under both current and foreseeable declared requirements.

### 12. Persistence / proofability

Check whether current T8-D architecture can admit a `notifications.*` owner namespace, identity-only cross-owner references, message immutability, same-Document reply constraint, Notification uniqueness and engagement constraints without violating foreign-SQL laws.

Call out any invariant that cannot actually be enforced or falsified.

### 13. B01/B03 UX coherence

Does bell + Quick Inbox + full `/notifications` preserve `Minha Caixa = assigned work`, or is the global IA still locally optimal rather than globally coherent?

Does Discussion belong on B03 ficha rather than the exact-content viewer/work/governance/history lenses?

### 14. YAGNI / Global Maximum

Attack both directions:

```text
under-design: retrofit traps hidden by “YAGNI”
over-design: owner/framework/event/realtime machinery not required by current consumer
```

Do not reward complexity merely because it resembles large-scale platforms.

## Required verdict

Use exactly:

```text
VERDICT: CONVERGED | NOT CONVERGED
MATERIAL: <count>
IMPORTANT: <count>
OPTIONAL: <count>
UNSUPPORTED_PREFERENCE: <count>
```

For each MATERIAL/IMPORTANT finding provide:

```text
finding id
exact candidate/authority location
concrete counterexample/failure mode
protected property
why current candidate is insufficient
smallest sustainable correction
whether it changes Product capability, owner count, route count, operation count, Permission count, mechanism choice, or only enforcement precision
```

Explicitly state whether the following survive your challenge:

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
no generic EventBus/broker/Redis
```

Do not implement code. Do not modify any file except this review dialogue. Reviewer output is Evidence, never authority.

---

## Fable response

<!-- Fable: append the independent adversarial review below this line. -->
