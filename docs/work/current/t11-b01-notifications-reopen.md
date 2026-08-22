# T11 — B01 Notifications Smallest-Scope Reopen

> **Status:** OPERATOR-RATIFIED CANDIDATE / PENDING UPSTREAM CONSOLIDATION.  
> **Parent:** B01 App Shell + Global Information Architecture.  
> **Reasoning authority:** `developmentconexus-ops/conexus-methodology/METHOD.md` — DevelopmentConexus Engineering Method v1.0.0.  
> **Implementation:** BLOCKED.  
> **Current locked B01 baseline remains authority except for this bounded reopen candidate until the full reopen is consolidated.**

## 1. Material trigger

Launch V1 now requires a persistent Notifications supporting semantic owner and an in-app Notification Inbox whose engagement lifecycle includes `seen`, `read/unread` and `archive/unarchive`.

That new Product requirement creates a real global-attention consumer and triggers the normal smallest-scope evidence-backed reopen law for B01.

## 2. Preserved B01 mental model

The sidebar meaning remains unchanged:

```text
Início       = current operational situation
Minha Caixa  = assigned work
Documentos   = official document truth / creation
Gestão       = system configuration
Evidência    = audit/evidence
```

Notifications is deliberately **not** placed under `Minha Caixa`, because attention items and assigned work have different semantics.

## 3. Alternatives

```text
A  Notifications inside Minha Caixa        REJECTED — conflates attention with assigned work
B  permanent sidebar Notifications item    REJECTED — overweights a transversal utility in primary IA
C  bell/popover only                        REJECTED — insufficient for persistent Inbox triage/archive/history
D  global bell + Quick Inbox + full Inbox   SELECTED / OPERATOR-RATIFIED
```

## 4. Global shell delta

Desktop:

```text
utility header
  + global Notification bell
  + unseen badge
  + Quick Inbox popover

sidebar
  unchanged

content region
  unchanged
```

Narrow/mobile:

```text
bell remains globally reachable
quick surface transforms to accessible sheet/full-screen material surface
full Inbox remains available through its stable route
```

The badge represents only currently presentable + non-archived + unseen Notifications.

## 5. Quick Inbox

The Quick Inbox is a bounded glance/action surface over the same Notifications authority as the full Inbox.

It may provide:

```text
recent presentable Notifications
open Notification source
mark all applicable items as read
entry to full Inbox
```

It is not a second store, second lifecycle authority or complete triage workspace.

Opening the bell does not automatically mark all loaded data as `seen`. Only Notifications actually presented under the accepted visibility semantics may become seen.

## 6. Full Inbox route

A real persistent Inbox requires a dedicated stable Product route:

```text
/notifications
```

Therefore, after bounded upstream consolidation:

```text
stable Product SPA routes  10 -> 11
```

This is a legitimate requirement-driven reopen; preserving the old route count is not an architectural goal.

The route supports the accepted engagement model, conceptually including:

```text
active Inbox
unread filter/lens
archived filter/lens
mark read
mark unread
archive
unarchive
mark all applicable as read
```

Exact query-state encoding, pagination and wire operations belong the later D7/API precision.

## 7. Navigation law

`/notifications` does not gain a permanent sidebar entry in the Launch baseline.

```text
primary IA sidebar  unchanged
utility header      Notification entry
```

A Notification source remains owned by its real Product lens. `DOCUMENT_MENTION` navigates back to the existing Document Official route and exact Discussion message context; the Inbox does not become a Document viewer.

## 8. Frontend authority law

Quick Inbox and full Inbox are two presentations of the same Notifications owner/read family.

```text
no header-only notification truth
no page-only notification truth
no frontend permission/disclosure matrix
no toast/realtime signal as notification authority
```

A realtime signal may later trigger refetch/reconciliation only.

## 9. P8 reopen requirement

Because B01 is already LOCKED, this textual approval does not silently rewrite the locked structural artifact.

Before the bounded B01 delta is re-LOCKED:

```text
render smallest P8 visual delta
→ utility-header bell/badge
→ desktop Quick Inbox
→ narrow/mobile transformation
→ operator visual adjudication
→ re-LOCK only the implicated B01 scope
```

Sidebar and unrelated B01 structure remain locked and are not reopened.

## 10. METHOD outcome

```text
CURRENT STRUCTURE CONFIRMED
+ smallest-scope B01 structural reopen
```

The existing B01 mental model remains globally sound. The new requirement needs a transversal utility surface and one new stable route, not a new primary-navigation category or a redesign of Home/My Work.

## 11. Reopen triggers

Revisit this decision only if evidence proves one of:

```text
Inbox becomes assigned-work authority rather than attention;
Notification volume/consumer diversity makes header + Inbox structure unusable;
a stronger global IA emerges from later assembled Product evidence;
the accepted Notification lifecycle is materially reduced and no longer needs a full Inbox;
accessibility/responsive proof invalidates the selected shell transformation.
```
