# T11 — B03 P10 Bounded Pattern Consolidation

> **Status:** COMPLETE / POST-LOCK PROOF.  
> **Block:** B03 — Document Official.  
> **Method:** Frontend Product Experience Planning Method v2.2.  
> **Rule:** shared patterns are derived from LOCKED semantic evidence; cosmetic similarity is insufficient.

## 1. Goal

Compare B03 against already-LOCKED B01/B01N/B02 and graduate only patterns with real repeated semantics.

## 2. Existing locked patterns reused

### Global application shell — REUSE

Origin:

```text
B01 App Shell + Global IA
```

B03 reuses:

```text
utility header
sidebar mental model
global navigation placement
narrow navigation transformation
```

No B03-specific shell variant is created.

### Notification Quick Inbox — REUSE

Origin:

```text
B01N Notification chrome + Quick Inbox
```

B03 consumes the already-locked interaction:

```text
bell
-> Quick Inbox
-> Notification source click
```

B03 owns only the destination behavior after the source resolves to the Document/Discussion context.

No second Notification store or B03-specific inbox pattern is created.

## 3. B03-local patterns — DO NOT GRADUATE YET

### Two-column Document dossier

```text
left  = current context + ficha + management
right = official-content preview
lower = revisions + Discussion
```

Status: **LOCAL B03 PATTERN**.

Reason: no second LOCKED screen yet shares the same stable-record + official-content-preview semantics. Cosmetic two-column similarity elsewhere would not justify abstraction.

### Official-content preview card

Status: **LOCAL B03 PATTERN**.

Semantics:

```text
recognition/context only
bound to current official Revision
click -> exact read-only viewer
never semantic/exact-content authority
```

No shared `PreviewCard` production abstraction is justified until another locked surface needs the same truth/interaction contract.

### Exact official viewer shell

Status: **LOCAL CANDIDATE PATTERN**.

Semantics:

```text
explicit back to owning ficha
code + Revision outside rendered bytes
exact source/rendition labeling
read-only
```

Potential future consumers may appear, but none is LOCKED yet. Keep local.

### Server-derived management affordance group

Status: **LOCAL CANDIDATE PATTERN**.

B03 uses `DocumentOfficialView.allowed_actions`, while future Governance screens may use their own `allowed_actions`. Shared implementation is deferred until at least two LOCKED screens prove equivalent interaction/accessibility/failure semantics.

Do not create a generic `AllowedActionsMenu` merely because both read models contain a similarly named member.

### Responsible-owner selection drawer

Status: **LOCAL B03 PATTERN**.

Semantics include:

```text
returned candidate set
separate ResponsibleOwner ETag
mutation-time target revalidation
```

No generic User picker is justified. Other User selectors may have different eligibility/disclosure laws.

### Stable-Document Discussion composer

Status: **LOCAL B03 PATTERN**.

Semantics:

```text
Text | Mention(user_id)
optional one-message reply reference
immutable accepted message
Mention autocomplete purpose-bound to exact Document
```

Do not generalize into a chat/comment platform.

### Notification-to-anchor reveal

Status: **LOCAL CROSS-FLOW PATTERN**.

Semantics:

```text
source identity
-> owning route
-> bounded anchor query
-> exact target reveal/focus/highlight
```

Potentially reusable later, but only B03 has a LOCKED anchored source target today.

## 4. Prototype-only constructs — NEVER GRADUATE

```text
transition stub dialogs for unopened B04/B07/B08
local fixture toasts
fixture-only owner mutation
fixture-only message insertion
review header/status chrome outside the Product shell
```

These exist only to make P8 operable and are not production design patterns.

## 5. Pattern vocabulary effect

New shared cross-product patterns graduated by B03:

```text
0
```

Existing shared patterns reused:

```text
Global App Shell
Notification Quick Inbox
```

New B03-local patterns retained without abstraction:

```text
Two-column Document dossier
Official-content preview
Exact official viewer shell
Document management affordance group
Responsible-owner selection drawer
Stable-Document Discussion composer
Notification-to-Discussion anchor reveal
```

This is intentional YAGNI, not an incomplete design system.

## 6. Reopen / graduation triggers

A B03-local pattern may graduate only when another LOCKED block proves matching:

```text
purpose
state ownership
protected semantics
authorization/disclosure posture
failure/recovery class
accessibility behavior
responsive transformation
```

Two components looking similar is insufficient.

## 7. P10 closure

```text
existing locked shared patterns reused          2
new shared abstractions created                 0
B03-local semantic patterns retained            7
false abstractions introduced                   0
unexplained duplicate locked semantic patterns  0
```

P10 is complete for B03.
