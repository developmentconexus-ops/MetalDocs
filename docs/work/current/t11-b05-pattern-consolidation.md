# T11 — B05 P10 Bounded Pattern Consolidation

> **Status:** COMPLETE / POST-LOCK PROOF.  
> **Block:** B05 — My Work / Work Queues.  
> **Method:** Frontend Product Experience Planning Method v2.2.  
> **Rule:** shared patterns are derived from LOCKED semantic evidence; cosmetic list/queue similarity is insufficient.

## 1. Goal

Compare locked B05 against already-LOCKED B01/B01N/B02/B03/B04 and graduate only patterns whose purpose, authority, state and recovery semantics actually match another locked consumer.

## 2. Existing locked patterns reused

### Global application shell — REUSE

Origin:

```text
B01 App Shell + Global IA
```

B05 reuses the locked application header/sidebar/navigation posture and the `Minha Caixa` IA. It does not create a Work-specific application shell.

### Notification Quick Inbox — REUSE

Origin:

```text
B01N Notification chrome + Quick Inbox
```

B05 inherits the global bell/Quick Inbox chrome. Notification attention remains transversal utility state and does not merge with assigned-work queues.

## 3. New shared patterns — NONE

No B05 pattern has enough cross-block semantic evidence to graduate.

The fact that B02, B03 or B04 also contain rows, lists, selectors, filters, stale recovery or navigation does not establish a shared Product pattern by itself.

P10 therefore authorizes **no new shared pattern vocabulary** from B05.

## 4. B05-local patterns — DO NOT GRADUATE YET

### Focused My Work intent lane

Status: **LOCAL B05 PATTERN**.

```text
Minha Caixa
→ selected assigned-work intent
→ one full-width focused queue
→ owner-lens handoff
```

The semantic purpose is assigned-work continuation, not discovery, current-document truth or authoring itself.

### Human-recognizable work handoff row

Status: **LOCAL B05 PATTERN**.

Common B05 row intent:

```text
recognize exact work
→ select
→ enter owner lens
```

Authoring and governance rows share that local intent, but their server shapes, ordering and attention semantics differ. Do not create a cross-product generic `WorkItem` abstraction.

### Due-aware governance triage queue

Status: **LOCAL B05 PATTERN**.

Semantics:

```text
active-Step persisted due_at
→ due_at ASC NULLS LAST server order
→ exact + relative deadline presentation
→ no manual priority state
```

This is governance-assigned-work semantics, not a generic sortable task list.

### Bounded deadline preset control

Status: **LOCAL B05 PATTERN**.

```text
Todos
Atrasados
Próximas 24h
Próximos 7 dias
Sem prazo
```

The presets are Product-bounded consequences of the F3/F4 deadline model. They are not a generic date-filter primitive or query-builder vocabulary.

### Cursor-stable relative-time traversal

Status: **LOCAL B05 PATTERN**.

```text
relative deadline filter
→ server first-page anchor
→ opaque cursor authenticates anchor
→ continuation reuses anchor
→ fresh first page receives fresh anchor
```

The underlying cursor mechanism is architecture-wide, but this user-visible temporal traversal has no second locked frontend consumer yet. Do not manufacture a reusable “anchored filter controller” from one use.

### Work-projection stale recovery

Status: **LOCAL B05 PATTERN**.

```text
queue row selected
→ owner destination rejects/disappears
→ return / refresh assigned-work projection
→ stale row not resurrected
```

This differs from B04 DRAFT OCC reconciliation and from B03 management mutation conflicts: B05 has no local human draft to preserve and no mutation to retry.

### Dense keyboard queue traversal

Status: **LOCAL B05 PATTERN**.

```text
ArrowUp / ArrowDown
→ ephemeral selected row
Enter
→ owner-lens handoff
```

Keyboard behavior is useful ergonomics, but no second locked block currently proves the same list-selection + owner-handoff semantic contract.

## 5. Similarity explicitly rejected as insufficient

```text
B02 Library results vs B05 assigned-work queues
  -> discovery/catalog intent vs actor-assigned continuation
  -> different order/filter semantics
  -> no shared queue authority

B01 Home work overview vs B05 detailed queue
  -> current operational situation vs detailed selected intent
  -> duplicating one into the other would erase the accepted IA distinction

B03 revisions/discussion rows vs B05 rows
  -> Document history/discussion context vs assigned-work handoff

B04 Work controls vs B05 authoring row
  -> B05 only routes to Work; B04 owns DRAFT/source/persistence/actions

B04 DRAFT 412 recovery vs B05 stale-row recovery
  -> rich local human input preservation vs read-only projection refresh

B02/B05 filters
  -> visual selector similarity does not imply shared query semantics

Camunda/Flowable/ProcessMaker generic task filters vs B05 deadline presets
  -> external systems are Evidence only and carry broader workflow semantics
```

## 6. Prototype-only constructs — NEVER GRADUATE

```text
Review controls bar
Advance server +6h
Fresh first-page fixture button
force stale / force list error / force empty
fixture serverNow display
fixture cursorAnchor display
local deterministic fake rows
boundary stub modal for unopened B06
fixture toast messages
```

These exist only to make P8 falsifiable and operator-operable.

## 7. Pattern vocabulary effect

Existing shared patterns reused:

```text
Global App Shell
Notification Quick Inbox
```

New shared pattern graduated by B05:

```text
none
```

B05-local patterns retained without cross-block abstraction:

```text
Focused My Work intent lane
Human-recognizable work handoff row
Due-aware governance triage queue
Bounded deadline preset control
Cursor-stable relative-time traversal
Work-projection stale recovery
Dense keyboard queue traversal
```

## 8. Reopen / graduation triggers

A B05-local pattern may graduate only when another LOCKED block proves matching:

```text
purpose / actor goal
state ownership
row identity semantics
server order/filter authority
cursor behavior
stale/failure recovery
navigation handoff meaning
Authorization/disclosure posture
responsive/accessibility behavior
```

Shared code convenience, CSS similarity or generic “task list” terminology is insufficient.

In particular, B06 must not be forced into a B05 queue abstraction merely because governance work navigates there.

## 9. P10 closure

```text
existing locked shared patterns reused          2
new shared semantic patterns graduated          0
B05-local semantic patterns retained            7
false abstractions introduced                   0
unexplained duplicate locked semantic patterns  0
```

P10 is complete for B05.