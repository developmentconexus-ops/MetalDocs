# T11 — B08 P10 Bounded Pattern Consolidation

> **Status:** COMPLETE / POST-LOCK PROOF.  
> **Block:** B08 — Notifications Full Inbox.  
> **Method:** Frontend Product Experience Planning Method v2.2.  
> **Rule:** shared patterns graduate only from repeated LOCKED semantic/protected behavior; visual similarity is insufficient.

## 1. Goal

Compare B08 against already-LOCKED B01/B01N/B02/B03/B04/B05/B06/B07 and reuse or graduate only patterns whose semantics truly match.

## 2. Existing shared patterns reused

### Global App Shell — REUSE

Origin: B01.

B08 reuses the locked global navigation/application shell. Notifications remains a transversal utility reached from the utility header and stable `/notifications` route; B08 creates no Notification-specific primary IA or sidebar section.

### Notification Quick Inbox — REUSE

Origin: B01N.

B08 reuses the locked global bell + unseen badge + Quick Inbox behavior as the glance/entry surface. Quick Inbox and Full Inbox consume the same Notifications owner/read/engagement authority.

Shared law preserved:

```text
one server-state authority
presentability before client receipt
unseen badge != unread count
opening a surface != blindly seen-all
source activation never grants source access
```

B08 does not broaden Quick Inbox into a second page store, owner or task queue.

## 3. Existing patterns deliberately not imported

### Exact Read-Only Content Viewer Shell — NOT A B08 CONSUMER

B08 does not render source Document/Discussion content beyond the bounded server-composed Notification preview. Source activation hands off to B03.

Adding the viewer here would duplicate source workspace semantics and is explicitly rejected.

### Work queue / Governance Case patterns — NOT APPLICABLE

Notifications are personal attention, not assigned work or governance responsibility. B05 queue ordering/deadline patterns and B06 Decision/feedback patterns are not reused merely because all surfaces contain rows and actions.

### B07 chronological History pattern — NOT APPLICABLE

B08 follows Notification recency order and engagement lenses. It does not become a lifecycle timeline or event-history abstraction.

## 4. B08-local semantic patterns — DO NOT GRADUATE

### Focused Triage Inbox

Status: **LOCAL B08 PATTERN**.

```text
unseen/unread summaries
+ three fixed engagement lenses
+ one recency-ordered Notification list
+ per-item engagement
+ source handoff
```

This is specifically the current Notifications Product surface and must not become a generic work/activity/inbox framework.

### Fixed active / unread / archived lens set

Status: **LOCAL B08 PATTERN**.

These views are direct projections of the current Notification engagement model:

```text
active    = presentable + non-archived
unread    = presentable + non-archived + read_at absent
archived  = presentable + archived
```

Do not generalize them into a generic filter/tab engine. Other collections have different server-owned filters and ordering laws.

### Human-recognizable Notification row

Status: **LOCAL B08 PATTERN**.

The row consumes the closed B08-F1 projection:

```text
Notification identity + engagement
current-disclosable DocumentReference
exact message_id
author UserReference
exact Revision-at-post when present
bounded message preview
```

It is not a generic Activity/Event row. Source optionality and presentability semantics belong to the current closed `DOCUMENT_MENTION` kind.

### Novelty separate from read state

Status: **LOCAL B08 PATTERN**.

```text
seen_at absent -> new/unseen
read_at absent -> unread
mark unread never restores unseen/new
```

The visual treatment must preserve this semantic distinction. It is not a generic badge-state abstraction for unrelated Product concepts.

### Presentation-driven seen batching

Status: **LOCAL B08 PATTERN**.

```text
actually presented unseen rows
-> transient bounded id batch
-> op84
-> server intersects presentability/recipient truth
```

This is interaction bookkeeping around the Notification `seen_at` meaning, not a generic viewport analytics framework. Exact threshold/debounce implementation remains disposable.

### Notification engagement reconciliation

Status: **LOCAL B08 PATTERN**.

Per-item read/archive actions and mark-all-read preserve authoritative server state on failure and reconcile through op82. This does not justify a generic optimistic mutation engine; current frontend law already rejects fabricated lifecycle/business truth.

### Source engagement + owner-lens handoff

Status: **LOCAL B08 COMPOSITION PATTERN**.

```text
op82 exact source ids
-> op83 read=true
-> B03 route boundary
-> B03 op79 anchor disclosure/read
```

The pattern preserves source-owner separation. Do not graduate a generic cross-feature deep-link resolver or parse preview text into navigation identity.

### SSE invalidation-only reconciliation

Status: **LOCAL B08 REALTIME PATTERN**.

```text
notifications.changed {}
-> invalidate/refetch op82
```

No business payload or state patch travels through the stream. This does not graduate a generic frontend EventBus or realtime entity store.

## 5. Candidate shared Notification row abstraction — DO NOT GRADUATE YET

B01N Quick Inbox and B08 Full Inbox both display Notifications, but their protected jobs differ:

```text
B01N
  bounded glance/entry
  limited actions
  presentation-driven seen of actually surfaced items

B08
  full triage
  fixed engagement lenses
  read/unread + archive/unarchive
  cursor continuation
```

A future production implementation may share low-level rendering primitives, but frontend planning does **not** freeze a durable `NotificationRow` Product abstraction now. The common owner data contract is already sufficient authority; component factoring remains an implementation detail until another locked semantic consumer proves a stable shared behavior set.

## 6. Quick Inbox / Full Inbox boundary locked

```text
B01N Quick Inbox
  global glance / entry
  recent presentable Notifications
  unseen badge
  mark-all-read
  Ver todas -> /notifications

B08 Full Inbox
  persistent triage
  active/unread/archived
  per-item read/unread
  archive/unarchive
  cursor continuation
```

Both use one Notifications server authority. Neither may cache a second durable engagement model.

## 7. Cross-block source boundary preserved

B08 does not absorb B03:

```text
B08
  recognize source + exact ids
  record Notification engagement
  hand off

B03
  authoritative Document Official / Discussion lens
  exact anchor read
  current source disclosure
```

Notification preview is recognition only. It is not a mini Discussion, mini Document viewer or source content authority.

## 8. Anti-abstraction decisions

Rejected:

```text
Generic Inbox framework
  B05 My Work and B08 Notifications have different ownership and action semantics

Generic Activity/Event feed
  B07 History, B03 Discussion and B08 Notifications have distinct truth models

Generic Tab/Filter engine as Product pattern
  B08 lenses are one closed operation filter set only

Generic Realtime entity sync
  op86 is invalidation only

Generic DeepLink resolver
  source identity and destination owner are explicit already

Generic Seen/Read state machine
  Notification engagement is a specific Product semantic lifecycle
```

## 9. Shared-pattern result

```text
existing locked shared patterns reused          2
new shared semantic patterns graduated           0
B08-local semantic/composition patterns retained 8
false abstractions introduced                    0
Notifications/Minha Caixa semantic merges        0
source-workspace duplications                     0
realtime business-state abstractions              0
```

## 10. P10 closure

B08 closes without changing the established shared-pattern vocabulary. Production implementation may factor code internally, but no reusable Product/component contract beyond the already-locked shared patterns is frozen by P10.

P10 is complete for the operator-locked B08 scope.
