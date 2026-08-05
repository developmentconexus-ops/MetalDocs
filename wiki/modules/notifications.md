# Module: notifications

> **Last verified:** 2026-08-05 (approval accountability loop, Tasks 4-7 — this doc's first pass. `notifications` was named in `CLAUDE.md`'s 15-module list and in [ADR 0043](../decisions/0043-notifications-module-and-lifecycle-bundle.md) but had no `wiki/modules/` page until now; the doc below is written and verified directly against the current tree, not against ADR 0043's single-worker design (which this module has since outgrown — see §Drift below).)
> **Status:** active
> **Maturity:** L1 — first documented pass, not yet a full Arc42/C4 promotion
> **Scope:** `internal/modules/notifications/` — the per-recipient notification inbox (3 HTTP read/write endpoints, self-scoped) and its TWO River consumer workers.
> **Key files:**
> - `internal/modules/notifications/infrastructure/fanout_worker.go` — `NotificationsFanoutWorker`, consumes `documentsdomain.LifecycleEventArgs` (the document-lifecycle bundle: published/superseded/obsoleted/approved/rejected)
> - `internal/modules/notifications/infrastructure/approval_notify_worker.go` — `ApprovalNotifyWorker`, consumes `approvaldomain.ApprovalNotificationArgs` (approval accountability loop, Task 5)
> - `internal/modules/notifications/infrastructure/notifications_repository.go` — `NotificationsRepository`, the sole reader/writer of `metaldocs.notifications`
> - `internal/modules/notifications/domain/types.go` — `NotificationRow`, `NotificationsPage`
> - `internal/modules/notifications/delivery/http/handler.go`, `routes.go` — `GET /notifications`, `GET /notifications/unread-count`, `POST /notifications/{id}/read`, `POST /notifications/read-all`
> - `internal/modules/iam/domain/model.go` — `CapNotificationRead = "notification.read"` (self-scope cap, deferred — not seeded to any role by default)
> - `apps/jobs/cmd/metaldocs-jobs/main.go:131-132` — both workers are registered ONLY in the `metaldocs-jobs` binary

---

## 1. What this module owns, and what it deliberately does not

`notifications` owns exactly one table, `metaldocs.notifications`, and the HTTP surface for a caller to read and mark-read their own rows. It does not own recipient resolution, does not know why an event happened, and does not read any other module's tables — [ADR 0039](../decisions/0039-cross-module-base-table-read-boundary.md)/[ADR 0043](../decisions/0043-notifications-module-and-lifecycle-bundle.md) drew that boundary and it still holds: `notifications_repository.go`'s own package doc states it "reads and writes ONLY `metaldocs.notifications`... no cross-module reads."

Two producer modules (`documents` and `approval`) resolve their own recipients and hand `notifications` an already-resolved envelope. `notifications` is delivery only.

## 2. Two workers, not one

ADR 0043 designed one worker (`NotificationsFanoutWorker`) for the document-lifecycle bundle. The approval accountability loop (Task 4/5) added a second, `ApprovalNotifyWorker`, rather than widening the first — the two envelopes carry different producer-resolved recipient shapes and the module boundary they cross is different (documents vs. approval). Both are registered only in `apps/jobs/cmd/metaldocs-jobs/main.go` — `metaldocs-api` and `metaldocs-worker` never subscribe either worker.

| Worker | Consumes | River kind | Recipient source | Notes |
|---|---|---|---|---|
| `NotificationsFanoutWorker` (`fanout_worker.go:28`) | `documentsdomain.LifecycleEventArgs` | `"notification_fanout"` (via the `documents` module's `LifecycleEventArgs.Kind()`) | Reader-targeted events (`document_published`/`superseded`/`obsoleted`) resolve via the published view `metaldocs.v_cd_obligated_readers`, queried in-worker with an explicit `tenant_id`/`controlled_document_id` predicate; author-targeted events (`document_approved`/`document_rejected`) use `args.SubmittedBy` carried in the envelope | This is the one cross-module-adjacent read in the module — a published `v_*` view, per ADR 0039, never a raw base table |
| `ApprovalNotifyWorker` (`approval_notify_worker.go:41`) | `approvaldomain.ApprovalNotificationArgs` | `"approval_notification"` | The envelope's own `RecipientUserIDs` — a list the `approval` module already resolved and snapshotted (the stage's `eligible_actor_ids`) before enqueueing | Contains **no approval SQL** and knows nothing about stages, routes, or eligibility — the doc comment on the type is explicit about this being the load-bearing boundary decision (invariant 6, [ADR 0082](../decisions/0082-approval-kernel-extraction.md)) |

Both workers run inside a single tenant-seeded transaction (`authz.SeedTxTenant`, a `SET LOCAL` config write, not an authz-recording read — H-PRE-1 safe, no lock held) and insert via the same `ON CONFLICT (recipient_user_id, source_event_id) WHERE source_event_id IS NOT NULL DO NOTHING` idempotency shape, so a redelivered job is a no-op rather than a duplicate row.

## 3. Unknown event type: error and dead-letter, never a silent drop

Both workers reject any event type they do not recognize by returning a `fmt.Errorf`, not `nil`:

- `fanout_worker.go:56-61` — the `default` branch of the `switch args.EventType` in `Work` returns `fmt.Errorf("fanout_worker: unknown event type %q", args.EventType)`. The comment on that branch is explicit about why: "returning nil dropped the event with no error and no dead-letter, so the divergence was invisible."
- `approval_notify_worker.go:60-66` — the same shape: an event type absent from `approvalMessages` returns an error before any DB work happens.

River's retry policy then applies and the job eventually dead-letters, which is the intended failure mode: a producer/consumer type divergence is a bug that must be visible, not a quietly dropped notification. `ApprovalNotifyWorker` treats an EMPTY recipient list differently from an unknown type, though — a livre (zero-stage) route has no eligible actors and that is legitimate, not a divergence, so `len(args.RecipientUserIDs) == 0` returns `nil` (checked before the subject-kind lookup, `approval_notify_worker.go:73-75`).

## 4. HTTP surface — self-scope, not module-boundary

`GET /notifications`, `GET /notifications/unread-count`, `POST /notifications/{id}/read`, `POST /notifications/read-all` are all gated by `CapNotificationRead` at tier-1 and then self-scoped by an explicit `recipient_user_id = <caller>` SQL predicate in every repository call (`notifications_repository.go`) — the capability's `ScopeTenant` classification exists only because the area/tenant grade has no "self" concept to select; the real guard is the predicate, not the scope enum (ADR 0043 §3 documents this explicitly and it remains true).

## 5. Drift from ADR 0043

ADR 0043 scoped this module to one worker and five document-lifecycle event types, with the notification-approver-routing/SLA/reminder surface explicitly "parked." The approval accountability loop built exactly that parked surface, as a second, independent worker rather than an extension of the first. ADR 0043 itself has not been amended to record this — flagged here as a documentation gap in that ADR, not fixed in this pass (out of Task 11's file scope, which is `wiki/modules/*` only).

---

## Cross-links

- [`wiki/modules/approval.md`](approval.md) — producer of `ApprovalNotificationArgs`; the accountability-loop changelog entry there is the authoritative source for why the envelope carries a resolved recipient list.
- [`wiki/modules/documents.md`](documents.md) — producer of `LifecycleEventArgs`.
- [`wiki/modules/jobs.md`](jobs.md) — the `metaldocs-jobs` binary that hosts both workers; `apps/jobs/cmd/metaldocs-jobs/main.go:131-132`.
- [ADR 0039](../decisions/0039-cross-module-base-table-read-boundary.md) — published-view read boundary (`v_cd_obligated_readers`).
- [ADR 0043](../decisions/0043-notifications-module-and-lifecycle-bundle.md) — original module + document-lifecycle bundle decision; superseded in scope (not in force) by the approval accountability loop's second worker, see §5 above.
- [ADR 0082](../decisions/0082-approval-kernel-extraction.md) — invariant 6 (module-boundary) that `ApprovalNotifyWorker`'s delivery-only design exists to preserve.

## Changelog (this doc)

- 2026-08-05 - First pass. Written and verified against the current tree (both workers, the HTTP surface, the idempotency shape, the unknown-event-type error/dead-letter behavior) rather than transcribed from ADR 0043, which predates the second worker. Created as part of the approval accountability loop's Task 11 doc pass.
