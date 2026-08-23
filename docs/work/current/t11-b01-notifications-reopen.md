# T11 — B01 Notifications Smallest-Scope Reopen

> **Status:** LOCKED / OPERATOR-RATIFIED on 2026-08-22 — bounded Notification chrome delta only.  
> **Parent:** B01 App Shell + Global Information Architecture.  
> **Current Product/architecture authority:** `../../decisions/discussion-notifications-launch.md`.  
> **Reasoning authority:** `developmentconexus-ops/conexus-methodology/METHOD.md` v1.0.0.  
> **Implementation:** BLOCKED.  
> **Preservation law:** the prior B01 LOCK remains binding outside the exact Notification chrome delta below.

## 1. Material trigger

Launch V1 now has a current Notifications supporting semantic owner and persistent in-app Inbox with `seen`, `read/unread` and `archive/unarchive` engagement.

That current Product requirement reopened only the smallest B01 scope needed to make global attention discoverable.

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

## 3. Ratified structure

Alternatives adjudicated:

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

## 4. P8 rendered artifact — LOCKED delta

Canonical rendered evidence for the bounded reopen:

```text
docs/work/current/t11-b01-notifications-wireframe.html
```

The original locked baseline remains preserved separately:

```text
docs/work/current/t11-b01-app-shell-wireframe.html
```

The operator visually approved the Notification delta on 2026-08-22. The lock covers exactly three affected states:

```text
1. utility header with bell + unseen badge, closed
2. desktop Quick Inbox open as an anchored overlay
3. narrow/mobile Quick Inbox transformed to a full-width sheet/material surface
```

It does **not** redesign Home, sidebar, primary navigation or existing work cards.

## 5. Locked desktop structure

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

## 6. Locked narrow/mobile structure

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

## 8. Accessibility / responsive structure locked

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

## 9. Operator adjudication

The operator approved the rendered delta without requested structural changes.

```text
bell global location                 ACCEPTED
Quick Inbox relative size            ACCEPTED
Mark-all-read placement               ACCEPTED
Ver todas discoverability             ACCEPTED
overlay preserving Home hierarchy     ACCEPTED
mobile sheet transformation            ACCEPTED
Minha Caixa separation                 ACCEPTED
accessibility/responsive structure      ACCEPTED
```

## 10. Re-LOCK result

```text
B01 original baseline       LOCKED / unchanged
Notification architecture  CURRENT
Notification P8 delta       LOCKED / OPERATOR-RATIFIED
B01 resulting structure     LOCKED
```

This re-LOCK changes only the implicated header / Quick-Inbox / responsive Notification structure. Unrelated B01 decisions remain exactly as previously locked.

After this re-LOCK, B03 P8 resumes with current Discussion/Notification semantics.

## 11. Reopen triggers

Reopen this bounded delta only if material evidence proves one of:

```text
Inbox becomes assigned-work authority rather than attention;
Notification volume/consumer diversity makes header + Inbox structure unusable;
a stronger global IA emerges from later assembled Product evidence;
the accepted Notification lifecycle is materially reduced and no longer needs a full Inbox;
accessibility/responsive proof invalidates the selected shell transformation.
```
