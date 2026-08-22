# T11 — D5 Discussion / Notification Disclosure & Immutability

> **Status:** OPERATOR-RATIFIED CANDIDATE / PENDING UPSTREAM CONSOLIDATION.  
> **Parent:** `t11-b03-discussion-notification-mini-design.md`.  
> **Reasoning authority:** `developmentconexus-ops/conexus-methodology/METHOD.md` — DevelopmentConexus Engineering Method v1.0.0.  
> **Implementation:** BLOCKED.  
> **Current upstream authority remains effective until the full bounded reopen is consolidated.**

## 1. Target invariants

```text
DiscussionMessage accepted truth must remain intelligible after replies, mentions, Notification creation, offboarding and access drift.
Notification persistence must never grant or preserve source access.
Historical User identity may survive while erasable UserProfile enrichment does not.
Current disclosure is enforced server-side before Notification presentation.
```

## 2. Message edit/delete alternatives

### A — free edit/delete

Rejected for Launch. Once a message may produce Mention, Notification and reply references, retroactive content mutation requires additional semantics for mention retraction, newly introduced mentions, notification cancellation/reissue, reply context and edit history.

### B — versioned edit + tombstone delete

Coherent but not currently required. It introduces MessageRevision/edit-history/moderation semantics without a named Launch consumer.

### C — accepted message immutable

**SELECTED / OPERATOR-RATIFIED.**

```text
edit message     not a Launch Product operation
delete message   not a Launch Product operation
correction       new DiscussionMessage
```

No arbitrary edit window is introduced.

## 3. Stable authorship and mention identity

```text
DiscussionMessage.author_user_id -> stable Organization User
Mention.user_id                   -> stable Organization User
```

UserProfile display/contact enrichment is not copied as durable authority merely to preserve historical chat appearance.

If the User remains but Profile enrichment is lawfully erased, the message/mention remains truthful through stable `user_id`; presentation uses the currently admissible bounded User reference or a neutral no-profile rendering.

## 4. Offboarding

Current T3 offboarding disables the User, revokes sessions, memberships and direct grants while preserving stable historical User identity.

D5 law:

```text
offboarding does not rewrite DiscussionMessage
offboarding does not rewrite Mention
offboarding does not delete Notification
offboarding does not mutate seen_at/read_at/archived_at
```

A disabled recipient receives no active Inbox/badge/realtime presentation.

Re-enable alone restores no grants or memberships and therefore does not automatically restore source disclosure.

## 5. Notification persistence vs presentability

Persistent Notification state and current user-visible presentability are distinct.

For the Launch `DOCUMENT_MENTION` source, a Notification is presentable only when at read time:

```text
recipient == current User
+ current User is ENABLED
+ exact Notification source remains currently disclosable to that User
```

When not presentable, the Notification is omitted server-side from:

```text
active Inbox
archived Inbox
unseen badge/count
unread count
bulk engagement target sets
realtime presentation payloads
```

No Document code/title, author/profile or message snippet is leaked to the client for it to hide locally.

Loss of access does not physically delete or rewrite the Notification.

If disclosure later becomes valid again, the same Notification may become presentable with its existing `seen_at/read_at/archived_at` values unchanged.

An item that remained unseen while non-presentable may truthfully reappear unseen after access is later restored.

## 6. Source-data duplication law

Notifications owns attention state, not source content. Persistent Notification source identity remains bounded and closed:

```text
DOCUMENT_MENTION
  document_id
  message_id
```

Do not persist Document title/code, DiscussionMessage text or mutable User display profile as Notification authority merely for rendering convenience.

Current presentation resolves admissible source/user facts under current disclosure.

## 7. Retention / hard deletion

Launch exposes no ordinary Product command to hard-delete DiscussionMessage or Notification.

This is not a claim of indefinite physical retention. A future explicit privacy/retention/records requirement may define physical pruning or redaction and is a material reopen.

`archive` remains recipient Inbox engagement, never physical deletion.

## 8. METHOD result

```text
CURRENT STRUCTURE CONFIRMED within the bounded reopen
```

The selected model preserves stable conversation and engagement truth while avoiding two unsupported Launch subdomains:

```text
message edit/version/retraction machinery
access-drift-driven destructive rewriting of Inbox history
```

## 9. Reopen triggers

Reopen D5 if material evidence proves one of:

```text
real Launch usage requires author correction/edit semantics beyond follow-up messages;
moderation/legal/privacy requirements require delete/redaction workflows;
Notification presentation cannot revalidate source disclosure sustainably at required scale;
a Product retention/records requirement defines bounded physical retention;
a future Notification kind requires source-presentability semantics that cannot fit the closed source contract.
```
