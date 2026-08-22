# T11 — Notification Architecture Research

> **Status:** OPERATOR-APPROVED ARCHITECTURE DIRECTION / EVIDENCE FOR B03 MINI-REOPEN.  
> **Parent:** `t11-b03-discussion-notification-mini-design.md`.  
> **Reasoning authority:** `developmentconexus-ops/conexus-methodology/METHOD.md` — DevelopmentConexus Engineering Method v1.0.0.  
> **Implementation:** BLOCKED.  
> **Upstream consolidation:** pending closure of the full bounded reopen.

## 1. Question

Determine how a mature in-application notification capability should fit MetalDocs after Document Discussion + explicit `@mention` became a Launch V1 requirement, without importing a generic notification/event platform by convenience.

## 2. Evidence studied

Reference implementations and architecture evidence included mature/open systems such as:

```text
Discourse
GitLab
Mattermost
Zulip
Novu
Google open-source architecture examples
```

Cross-product conclusion:

```text
domain/business fact
!=
persistent Notification state
!=
durable asynchronous work
!=
realtime client wake-up
```

Mature products frequently use events, but they do not all require an external distributed event broker. Brokers appear where fan-out, throughput, replay, independent services or realtime-stream scale justify them.

## 3. MetalDocs authority baseline

Current accepted architecture already provides:

```text
owner-first modular monolith
one PostgreSQL product-state database
River as the single PostgreSQL-backed durable-job mechanism
same-local-transaction composition when one invariant requires it
mechanism != authority
no generic EventBus/outbox without a named consumer
```

Discussion / Mention / Notification is now the named consumer that legitimately reopens the previous `Notifications deferred` assumption. It does not by itself prove Kafka, NATS, RabbitMQ, Redis, a notification microservice, or a full external notification platform.

## 4. Operator-approved Global Maximum direction

Preferred structure:

```text
BUSINESS FACT
Document DiscussionMessage + explicit Mention
        │
        │ same local business transaction where required
        ▼
ATTENTION FACT
persistent Notification for recipient User
        │
        ├── Inbox engagement state
        └── B01 notification surface
        │
        ▼
REALTIME SIGNAL
best-effort “Inbox changed” signal
        │
        ▼
Browser reloads/refetches canonical Notification truth
```

Binding architectural direction:

1. **Notification is persistent Product state, not a transient queue message.**
2. With recipient, kind/source, Inbox lifecycle and engagement state, Notification is a candidate **supporting semantic owner** rather than `platform/*` mechanism state.
3. `DiscussionMessage + Mention -> required in-app Notification` should be capable of committing atomically in the same local PostgreSQL transaction so accepted mention cannot exist while its required Notification is lost.
4. Notification never grants access and never owns Document/Discussion lifecycle truth.
5. Realtime delivery is only an acceleration mechanism. Losing a realtime signal must not lose the Notification.
6. River remains the one durable asynchronous-work mechanism for future external effects such as email/push when those channels are actually admitted.
7. **Typed domain/event semantics are useful; a generic external Event Bus is not yet justified.**
8. Preserve a clean event seam so an internal EventStore/pub-sub layer can be introduced later if multiple independent consumers create real producer-coupling pressure.
9. Kafka/NATS/RabbitMQ/PubSub-class infrastructure requires evidence such as independent services, sustained high event throughput, heavy distributed fan-out, replay/stream processing, or a failure-domain need PostgreSQL + River cannot satisfy.

## 5. Realtime candidate

A promising smallest mechanism to probe later is:

```text
persistent Notification in PostgreSQL
+ lightweight commit-safe wake-up
+ SSE server -> browser
+ normal HTTP for Inbox commands
```

PostgreSQL `LISTEN/NOTIFY` is a candidate wake-up mechanism because canonical data remains in tables and a notification signal can simply tell connected clients that their Inbox changed. SSE is a candidate because current Notification realtime is server-to-browser only.

Neither mechanism is ratified yet. D7/D8 must prove driver behavior, reconnect/reconciliation, authorization, deployment implications, and future replica behavior before mechanism selection.

## 6. Rejected-by-current-evidence directions

```text
Notification buried as Controlled Documents private mechanism
generic EventBus added merely because events exist
Kafka / RabbitMQ / NATS by default
Novu or another full notification platform by default
separate notification microservice
parallel durable queue beside River
Redis introduced solely for notifications
WebSocket mandated solely for the notification bell
```

These may reopen only if a concrete requirement proves them smaller or safer than the accepted architecture.

## 7. Required upstream reopen if final design confirms this direction

Likely bounded impact:

```text
Product      Notification becomes Launch capability
T1           semantic family / supporting-owner topology
T3           document.discuss + Notification access semantics
T5           notification/realtime/external-channel effect census
T8-B         supporting Notifications owner/package if confirmed
T8-C         exact cross-owner same-transaction contract; event seam
T8-D         Notification persistence + invariants
T8-E/T6      exact Inbox/Discussion operations and schemas
T8-F         B01/B03 frontend realization
T8-G         only the smallest proven realtime/runtime mechanism
T11          implementation graph / proof obligations
```

No upstream file is silently rewritten until the B03 mini-design closes and the exact delta is operator-ratified.

## 8. Reopen triggers for an actual Event Bus/broker

Reconsider an internal EventStore/pub-sub abstraction when:

```text
one producer must know multiple independent consumers
repeated cross-owner producer edits are required whenever a subscriber is added
multiple event consumers need independent retry/delivery semantics
```

Reconsider an external broker when:

```text
services become independently deployed/trusted
measured event volume/fan-out exceeds the sustainable PostgreSQL/River posture
replay/stream processing becomes a named product/operational requirement
failure-domain isolation requires durable transport outside the product database
```

Until one of these is evidenced, preserve the seam rather than the infrastructure.
