# T11 — B01 Notifications Smallest-Scope Reopen

> **Status:** P8 RENDERED / OPERATOR ADJUDICATION REQUIRED / NOT RE-LOCKED.  
> **Parent:** B01 App Shell + Global Information Architecture.  
> **Current Product/architecture authority:** `../../decisions/discussion-notifications-launch.md`.  
> **Reasoning authority:** `developmentconexus-ops/conexus-methodology/METHOD.md` v1.0.0.  
> **Implementation:** BLOCKED.  
> **Preservation law:** the prior B01 LOCK remains binding outside the exact Notification chrome delta below.

## 1. Material trigger

Launch V1 now has a current Notifications supporting semantic owner and persistent in-app Inbox with `seen`, `read/unread` and `archive/unarchive` engagement.

That current Product requirement reopens only the smallest B01 scope needed to make global attention discoverable.

## 2. Preserved B01 mental model

Unchanged:

```text
Início       = current operational situation
Minha Caixa  = assigned work
Documentos   = official document truth / creation
Gestão       = system configuration
Evidência    = audit/evidence
```

Notifications is deliberately **not** placed under `Minha Caixa`: attention items and assigned work have different semantics.

## 3. Ratified structure before P8

Alternatives already adjudicated:

```text
A  Notifications inside Minha Caixa        REJECTED — conflates attention with assigned work
B  permanent sidebar Notifications item    REJECTED — overweights a transversal utility in primary IA
C  bell/popover only                        REJECTED — insufficient for persistent Inbox triage/archive/history
D  global bell + Quick Inbox + full Inbox   SELECTED / OPERATOR-RATIFIED
```

Current stable route:

```text
/notifications
```

The route exists without a permanent sidebar entry.

## 4. P8 rendered artifact

Canonical candidate for the bounded reopen:

```text
docs/work/current/t11-b01-notifications-wireframe.html
```

The original locked baseline remains preserved separately:

```text
docs/work/current/t11-b01-app-shell-wireframe.html
```

The P8 candidate intentionally renders only three affected states:

```text
1. utility header with bell + unseen badge, closed
2. desktop Quick Inbox open as an anchored overlay
3. narrow/mobile Quick Inbox transformed to a full-width sheet/material surface
```

It does **not** redesign Home, sidebar, primary navigation or existing work cards.

## 5. Desktop structure under review

```text
utility header
  existing brand/tagline
  spacer
  Notification bell + unseen badge
  existing session control

bell open
  anchored Quick Inbox overlay
  content layout does not shift
```

Quick Inbox hierarchy:

```text
Notificações
+ Mark all read
+ active / unread quick lenses
+ bounded recent presentable items
+ source context sufficient for recognition
+ Ver todas -> /notifications
```

Quick Inbox is glance/action only. Archive/unarchive and full triage remain the full Inbox's responsibility.

## 6. Narrow/mobile structure under review

Do not squeeze the desktop popover into a narrow viewport.

```text
utility header remains globally reachable
bell opens full-width sheet/material surface below header
items become one-column touch targets
Ver todas -> /notifications
sidebar/drawer behavior remains independently unchanged
```

## 7. Engagement / authority constraints represented by P8

```text
badge = presentable + non-archived + unseen
unread != unseen
opening bell does not blindly mark every loaded/paginated item seen
only actually presented items may become seen
Quick Inbox and full Inbox share one server-state authority
click source -> Notification read/seen + authoritative source navigation
source always rechecks current disclosure
SSE/toast never becomes Notification truth
```

No client-side permission/disclosure matrix is introduced.

## 8. Accessibility / responsive checks for operator review

P8 structurally preserves:

```text
bell has an accessible name including new-item quantity
badge novelty is not communicated by color alone
no material action depends on hover
focus order can proceed header -> Quick Inbox controls -> items -> full Inbox transition
mobile sheet exposes an explicit close/back control
bounded touch targets replace compressed desktop popover
```

Exact focus-trap/popover library, final icons, colors, typography and motion remain visual/implementation choices and are not locked by P8.

## 9. Operator visual adjudication questions

Review only the reopened delta:

1. Is the bell in the correct global location relative to session/user controls?
2. Is the Quick Inbox visually large enough for recognition without becoming a second workspace?
3. Is `Marcar todas como lidas` appropriately visible without competing with opening a Notification?
4. Is `Ver todas` discoverable enough to communicate that full Inbox exists?
5. Does the overlay preserve the already-approved Home hierarchy rather than displacing it?
6. On mobile, is sheet/full-width transformation preferable to a compressed popover?
7. Does anything make Notifications feel like assigned work under `Minha Caixa`? If yes, the structure fails.
8. Is any material interaction dependent on hover, color alone, or inaccessible reading order? If yes, the structure fails.

## 10. Lock law

Current status remains:

```text
B01 baseline              LOCKED
Notification architecture CURRENT
Notification P8 delta     CANDIDATE / NOT LOCKED
```

Only the operator may re-LOCK this delta after viewing the rendered artifact. A re-LOCK updates only the implicated header/Quick-Inbox/responsive structure; unrelated B01 remains untouched.

After re-LOCK, B03 P8 resumes with current Discussion/Notification semantics.
