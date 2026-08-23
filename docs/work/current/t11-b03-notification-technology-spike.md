# T11 — Notification Technology / Reuse Spike

> **Status:** OPERATOR-RATIFIED CANDIDATE / FABLE-CONVERGED / PENDING UPSTREAM CONSOLIDATION.  
> **Parent:** `t11-b03-discussion-notification-mini-design.md`.  
> **Reasoning authority:** `developmentconexus-ops/conexus-methodology/METHOD.md` — DevelopmentConexus Engineering Method v1.0.0.  
> **Implementation:** BLOCKED.

## 1. Purpose

Select the smallest sustainable implementation mechanisms for the already-ratified Discussion / Mention / Notifications semantics without allowing a framework, SaaS or existing repository to redefine Product authority.

Frozen requirements entering this spike include:

```text
Document-level persistent Discussion
Text + semantic Mention(user_id)
server-backed disclosure-safe mention autocomplete
persistent Notifications supporting semantic owner
seen / read-unread / archive-unarchive Inbox lifecycle
same-local-transaction Mention -> Notification invariant
native MetalDocs Notification API and stable /notifications route
SSE is invalidation/wake-up only, never Notification authority
River remains the single durable future-work mechanism
86 application operations / 11 Idempotency-Key creations candidate census
```

## 2. Evaluation law

Each candidate is judged by:

```text
semantic fit
ownership preservation
maintenance/activity
license
React / TypeScript or Go fit
accessibility
integration surface
runtime/deployment dependencies
vendor / schema lock-in
ability to replace later without migrating Product truth
proof burden
```

Reference implementations and mature products are evidence, not target authority.

## 3. Discussion composer — Lexical selected

### Alternatives considered

```text
Tiptap / ProseMirror
react-mentions family
Lexical + third-party mention plugins
Lexical core + @lexical/react + MetalDocs Mention node/typeahead
```

### Decision

```text
SELECT:
Lexical core + @lexical/react
PlainText-only Discussion composer
MetalDocs-owned custom MentionNode / typeahead integration
server-backed mention candidate lookup
```

Rationale:

- Lexical provides a mature minimal editor substrate and official typeahead/mention reference implementation without forcing a ProseMirror document model.
- Tiptap is strong but introduces materially richer editor/document semantics than current Discussion requires.
- the legacy/deprecated `react-mentions` lineage and string-markup identity models are a poor fit for D3 stable `user_id` Mention authority.
- an additional third-party Mention plugin is unnecessary while the required behavior can be implemented against Lexical's public primitives.

Binding boundary:

```text
Lexical EditorState = frontend mechanism only
Lexical JSON / HTML = NOT persistence or API authority

accepted wire/domain content =
  Text(text)
  | Mention(user_id)
```

Display name is presentation enrichment only. Replacing Lexical later must require zero migration of accepted Discussion content.

## 4. Notification frontend — native MetalDocs feature selected

Do not adopt Novu, Knock, MagicBell or another Notification-domain SDK/runtime for Launch.

Those systems remain useful behavioral/UI evidence, but importing their runtime model would require mapping or duplicating concepts such as Subscriber/User, Notification, Feed/Inbox engagement or backend URLs beside MetalDocs authority.

Target frontend realization:

```text
OpenAPI-generated wire types
+ thin existing lib/api transport
+ TanStack Query server state
+ native features/notifications presentation
```

Quick Inbox and `/notifications` consume the same Notifications owner/API family. No browser global Notification truth store is introduced.

## 5. Realtime client — native EventSource selected

The accepted realtime contract is server -> browser invalidation only.

```text
GET /api/v1/notifications/events
Content-Type: text/event-stream

signal semantics:
  notifications.changed
```

The signal carries no Document title, message body, User profile or Notification business payload. On signal:

```text
browser invalidates/refetches canonical Notification queries
```

Use the browser's native `EventSource` mechanism. No SSE client library is required for the current same-origin cookie-authenticated GET stream.

Lost realtime signals never lose Notification truth; normal refetch/reconnect/focus recovery remains authoritative.

## 6. Realtime server — narrow Go SSE adapter selected

Use Go `net/http` streaming/flush primitives behind a narrow application-mediated transport adapter rather than introducing a standalone SSE framework for the current one-event invalidation protocol.

Accepted dependency direction:

```text
transport/http
→ application/notifications
→ narrow application-owned subscription port
→ platform realtime mechanism
```

`transport -> platform` remains forbidden. The platform implementation owns only connection/subscription mechanics and carries no Product state, source data or access decision.

Required properties:

```text
text/event-stream response
bounded heartbeat where operationally required
flush after event emission
request-context cancellation closes subscriber
no Product truth in connection/session state
no replay/event-history authority
```

The OpenAPI/code-generation toolchain must prove server-side representation of the SSE operation. Operation 86 may not become a manually invented route outside the contract SSOT.

## 7. Launch wake-up mechanism — in-process coalescing hub selected

Current T8-G baseline has one MetalDocs application replica. The smallest current wake-up mechanism is therefore an in-process recipient-scoped signal hub behind the narrow mechanism seam.

Conceptually:

```text
Subscribe(user_id)
Wake(user_id)
```

Properties:

```text
one or more active browser connections per User
wake-up occurs only after successful semantic commit
buffer/coalescing semantics reduce repeated pending invalidations to one signal
slow/disconnected subscribers never block or fail Product mutation
wake-up failure never rolls back already-committed Notification truth
no durable event queue
```

All Notification changes that can affect another tab invoke the wake-up after commit through the application layer:

```text
DOCUMENT_MENTION creation
mark seen
mark read
mark unread
archive
unarchive
mark all read
```

Semantic owners never call realtime directly.

Transient deploy/restart overlap may briefly leave clients connected to different process instances and therefore miss an in-process wake. This is explicitly tolerated because the wake is not Product truth: reconnect/focus/ordinary canonical `GET /notifications` reconciles persisted Notification state. T8-G consolidation must preserve this tolerance rather than invent cross-process delivery guarantees in the one-replica baseline.

This mechanism is not a generic EventBus and owns no Notification state.

## 8. Multi-replica seam — PostgreSQL LISTEN/NOTIFY first candidate

PostgreSQL LISTEN/NOTIFY is not activated in the one-replica Launch baseline solely for speculative scale. It is the first qualified replacement candidate for the wake-up mechanism if T8-G later proves multiple application replicas or another cross-process wake-up consumer.

The intended seam permits:

```text
InProcessWakeup
   ↓ replace mechanism only
PostgresWakeup (LISTEN/NOTIFY)
```

without changing Notification semantics, API, SSE, frontend or Product authority.

Do not turn LISTEN/NOTIFY payloads into business truth; canonical Notification rows remain the recovery source.

## 9. Durable async — River unchanged

River remains the sole PostgreSQL-backed durable future-work mechanism for activated effects that must survive process failure/retry.

```text
SSE / wake-up      = ephemeral freshness mechanism
River durable job  = required future work mechanism
```

Do not route in-app Notification creation through River. Do not enqueue every SSE invalidation. Future email/push/webhook delivery, if later admitted, must be evaluated under the existing T5 transaction-coupled durable-effect laws.

## 10. Event bus/framework disposition

### Current Launch

```text
generic internal EventBus     ABSENT
external broker               ABSENT
Redis                         ABSENT
Kafka / RabbitMQ / NATS       ABSENT
```

Typed event semantics remain conceptually admissible, but the current producer/consumer graph does not justify a generic event-processing platform.

### Future trigger

Reopen an internal EventStore/EventBus only when evidence shows repeated producer knowledge of multiple independent consumers, e.g. one accepted fact driving several independent owners/mechanisms such that direct application choreography materially increases coupling.

Watermill is a qualified future candidate because it is a mature Go event/pub-sub library supporting multiple backends and stable event-routing patterns. It is **not** a Launch dependency merely because it exists.

An external broker requires additional evidence such as independent services/trust boundaries, sustained fan-out/throughput, replay/streaming as a Product requirement, or failure-domain isolation that PostgreSQL + River + the modular monolith cannot sustainably satisfy.

## 11. Implementation proof gate

Before implementation-readiness may claim these mechanisms are proven, targeted spikes/tests must be able to falsify at least:

```text
Lexical composer emits/rehydrates exact Text | Mention(user_id) semantics
Lexical/HTML/JSON never becomes persistent authority
mention autocomplete handles async stale-result/cancellation/keyboard/IME behavior safely
same User mentioned repeatedly produces one Notification
multiple tabs for one User receive invalidation
multiple pending invalidations coalesce without losing canonical truth
slow/disconnected SSE client cannot block a Product mutation
request cancellation removes subscriber state
lost SSE signal is recovered through canonical GET /notifications
transient deploy-overlap wake loss does not affect Product correctness
SSE payload contains zero source business truth
selected OpenAPI Go generator supports the declared server-side text/event-stream operation
no manual parallel route/DTO authority is introduced
no generic EventBus/broker/Redis dependency appears without a reopen
```

Accessibility proof must cover keyboard operation, focus management and screen-reader recognition for mention typeahead and Notification surfaces.

## 12. Independent challenge result

Fresh Fable Evidence PR #165 independently challenged Lexical, native Inbox, EventSource/SSE, same-Scope Notification creation, in-process wake-up, River and the absence of EventBus/broker/Redis.

```text
VERDICT: CONVERGED
MATERIAL: 0
IMPORTANT: 1
OPTIONAL: 3
```

No technology/mechanism finding survived. Fable explicitly marked Lexical, SSE + in-process wake-up, River as sole durable async and the absence of generic EventBus/broker/Redis as surviving the challenge.

The operator accepted the one IMPORTANT consolidation-completeness finding and all three OPTIONAL precisions. No second Fable round is justified.

## 13. METHOD outcome

```text
CURRENT STRUCTURE CONFIRMED
+ bounded mechanism additions
```

The new Product requirements justify Lexical mention composition and an ephemeral Notification realtime seam, but do not justify a second backend platform, notification SaaS runtime, external broker, generic EventBus or Redis.

## 14. Reopen triggers

Reopen this mechanism decision only when material evidence shows one of:

```text
Lexical cannot satisfy the exact accessibility/IME/composition contract without disproportionate custom behavior;
a richer Discussion content requirement is operator-approved;
multiple application replicas become a current runtime requirement;
realtime fan-out/connection scale makes the in-process mechanism unsustainable;
multiple independent consumers create material event-coupling pressure;
external delivery channels become Launch requirements;
OpenAPI/codegen constraints make the selected SSE realization materially inferior;
a proven maintained third-party mechanism becomes smaller than the selected local adapter without owning Product semantics.
```

Framework popularity, available infrastructure, or hypothetical future features are not sufficient reopen evidence.
