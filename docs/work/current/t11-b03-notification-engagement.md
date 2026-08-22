# T11 — Notification Engagement Lifecycle

> **Status:** OPERATOR-RATIFIED CANDIDATE / PENDING UPSTREAM CONSOLIDATION.  
> **Parent:** `t11-b03-discussion-notification-mini-design.md`.  
> **Ownership prerequisite:** `t11-b03-notification-ownership-reopen.md` — OPERATOR-RATIFIED CANDIDATE.  
> **Reasoning authority:** `developmentconexus-ops/conexus-methodology/METHOD.md` — DevelopmentConexus Engineering Method v1.0.0.  
> **Implementation:** BLOCKED.

## 1. Decision

The Notification Inbox uses three independent engagement dimensions rather than one flat status enum:

```text
Notification
  created_at
  seen_at?
  read_at?
  archived_at?
```

These facts are owned by the supporting semantic owner `Notifications` after bounded upstream consolidation.

## 2. Seen

`seen_at` means the Notification has been presented to the recipient in an active Inbox surface.

```text
seen_at absent  -> unseen
seen_at present -> seen
```

`seen_at` is monotonic. Once set, it never returns to absent.

Seen does not mean:

```text
Notification read
Document viewed/read
Read & Acknowledge
Governance decision
source fact resolved
```

A later technical realization must define a falsifiable presentation/visibility rule; the semantic contract does not freeze an IntersectionObserver threshold or realtime mechanism.

## 3. Read / unread

`read_at` is reversible recipient engagement state:

```text
read_at absent  -> unread
read_at present -> read
```

Laws:

```text
READ => SEEN
mark read:
  seen_at = now if absent
  read_at = now

mark unread:
  read_at = absent
  seen_at unchanged
```

Marking unread never makes a Notification unseen/new again.

## 4. Archive / unarchive

`archived_at` controls active Inbox placement and is orthogonal to read state:

```text
archived_at absent  -> active Inbox
archived_at present -> archived
```

Laws:

```text
archive != delete
archive preserves seen/read state
unarchive preserves seen/read state
direct archive/unarchive interaction implies SEEN
```

Launch V1 admits per-item archive/unarchive. It does not imply bulk archive/unarchive.

## 5. Direct engagement law

Any deliberate recipient action on one Notification proves that the Notification has been seen:

```text
mark read
mark unread
archive
unarchive
click/open source
→ seen_at must be present
```

This prevents contradictory states such as recipient-archived-but-unseen.

## 6. Badge and Inbox semantics

The global bell badge represents **new/unseen** attention, not the set the recipient deliberately keeps unread:

```text
bell badge
= currently presentable + non-archived + unseen Notifications
```

The active Inbox may independently expose unread emphasis/filtering:

```text
unread count/filter
= currently presentable + non-archived + unread Notifications
```

Marking a seen Notification unread therefore does not recreate bell badge novelty.

Exact presentability under current disclosure/access is owned by D5.

## 7. Mark all as read

Launch V1 includes `mark all as read` for the applicable current Inbox set.

Semantics:

```text
for each applicable Notification:
  seen_at = now if absent
  read_at = now
```

D5 closes which Notifications remain applicable after access/disclosure drift.

Launch V1 deliberately does not add:

```text
mark all unread
mark unseen / mark all unseen
archive all / unarchive all
snooze
priority system
notification preferences
```

No current consumer requires them.

## 8. Source navigation

Opening a Notification engages the Notification and then navigates to the authoritative source lens. For the current `DOCUMENT_MENTION` kind:

```text
Notification(document_id, message_id)
→ mark read (therefore seen)
→ existing /documents/:document_id B03 lens
→ reveal Discussion
→ resolve/highlight exact message_id as browser presentation/navigation context
```

No new stable Product route is implied merely to address a Notification or DiscussionMessage.

Notification engagement and source access are independent truths. The source always rechecks current disclosure/Authorization; Notification never acts as an access token.

## 9. Counter authority prevention

No `User.unseen_notification_count`, `User.unread_notification_count`, or equivalent stored counter becomes semantic authority. Counts are derived from Notification truth; later caches/projections, if measured need requires them, remain rebuildable mechanisms.

Binding distinctions:

```text
Notification SEEN       != Document viewed
Notification READ       != Document read
Notification READ       != Read & Acknowledge
Notification READ       != governance evidence
Notification ARCHIVED   != delete
Notification ARCHIVED   != source resolved
Notification UNREAD     != new/unseen
```

## 10. METHOD result

Within the operator-ratified Notifications owner, the selected engagement model is the smallest sustainable Inbox lifecycle that preserves the current user requirement without forcing a later semantic migration from `read:boolean` to separate novelty/read/placement concepts.

Reference evidence from mature Inbox systems supported the distinction but did not define MetalDocs authority.

```text
OUTCOME: CURRENT STRUCTURE CONFIRMED inside the new Notifications owner
```

## 11. Reopen triggers

Reopen only if material evidence proves one of:

```text
seen/read separation has no current user-observable value;
archive is not required for a persistent Inbox;
a regulatory/business meaning requires Notification read to bind another owner explicitly;
scale evidence requires a derived counter/projection while preserving Notification as authority;
a materially different Inbox requirement introduces another independent engagement dimension.
```
