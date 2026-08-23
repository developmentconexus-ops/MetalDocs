# B08 Notifications Full Inbox P8 Realization Plan

> **For agentic workers:** REQUIRED SUB-SKILL: use the repository P8 method; this plan is for disposable frontend-planning evidence only and MUST NOT be treated as Product implementation authorization.

**Goal:** Produce one browser-operable low-fidelity HTML wireframe that lets the operator falsify or confirm the ratified B08 Focused Triage Inbox structure.

**Architecture:** One self-contained HTML file with CSS + vanilla JavaScript and deterministic local fixtures. The fixture model mirrors only accepted Notification engagement/source-read semantics; it does not implement Product persistence, Authorization, API transport, SSE transport or React state architecture.

**Tech Stack:** HTML5, CSS, vanilla JavaScript, deterministic in-memory fixtures, browser `IntersectionObserver` only for the P8 seen-presentation hypothesis.

**Spec:** `docs/work/current/t11-b08-notifications-full-inbox-r1.md` + `docs/decisions/notification-inbox-recognition-read.md` + `docs/decisions/discussion-notifications-launch.md`.

## Global Constraints

- Product implementation remains BLOCKED.
- Canonical P8 medium is one `.html` file under `docs/work/current/`.
- No React, backend calls, OpenAPI client, real SSE, local Authorization evaluator or copied source ACL model.
- Fixed lenses only: `active`, `unread`, `archived`.
- Canonical fixture ordering: `created_at DESC`, then `notification_id DESC`; client does not offer sort.
- Unseen and unread are visibly distinct.
- Fetch/cache does not by itself make a row seen.
- `READ => SEEN`; mark unread preserves seen; archive/unarchive preserves read/seen.
- Source-open uses exact fixture `document_id + message_id` and terminates at an explicit B03 boundary rather than rebuilding B03.
- SSE fixture contains no business payload; it only triggers a simulated canonical refetch.
- No search, arbitrary filters, bulk selection/archive, snooze, priority, preferences, delete, source reply/editor/viewer or B09 work.

---

### Task 1: Build deterministic semantic fixture model

**Files:**
- Create: `docs/work/current/t11-b08-notifications-full-inbox-functional-wireframe.html`

**Interfaces:**
- Consumes: ratified `NotificationInboxItem` semantics and B01N shell/Quick Inbox structure.
- Produces: local `notifications[]`, current lens, cursor-page state, failure toggles and rendered counts used by all later interactions.

- [ ] **Step 1: Define rows covering every engagement combination that matters**

Use deterministic objects equivalent to:

```js
{
  id: 'n-001',
  createdAt: '2026-08-23T20:26:00Z',
  seen: false,
  read: false,
  archived: false,
  source: {
    documentId: 'doc-po-032',
    documentCode: 'PO-032',
    messageId: 'msg-po-032-77',
    author: 'Mariana Alves',
    revisionAtPost: 'REV003',
    preview: '@Leandro consegue verificar o item 4 antes da reunião?'
  }
}
```

Include at least:

```text
unseen + unread active
seen + unread active
seen + read active
seen + read archived
seen + unread archived
second cursor page with unseen rows initially off-screen/unpresented
```

- [ ] **Step 2: Implement derived current counts and fixed lens predicates**

```js
const isActive = n => !n.archived;
const isUnread = n => !n.archived && !n.read;
const isArchived = n => n.archived;
const unseenCount = rows => rows.filter(n => !n.archived && !n.seen).length;
const unreadCount = rows => rows.filter(n => !n.archived && !n.read).length;
```

- [ ] **Step 3: Verify no fixture-only state invents Product semantics**

Run a static scan confirming there are no fixture properties named `priority`, `snoozed`, `deleted`, `permission`, `allowed`, `task`, `acknowledged` or `documentViewed`.

---

### Task 2: Implement Focused Triage Inbox interactions

**Files:**
- Modify: `docs/work/current/t11-b08-notifications-full-inbox-functional-wireframe.html`

**Interfaces:**
- Consumes: Task 1 fixture model.
- Produces: operator-operable fixed lenses and accepted engagement transitions.

- [ ] **Step 1: Implement three fixed lenses**

```js
const lensPredicate = {
  active: isActive,
  unread: isUnread,
  archived: isArchived,
};
```

Switching lenses resets local cursor presentation for that lens and never changes Notification engagement state.

- [ ] **Step 2: Implement per-row read/unread and archive/unarchive**

```js
function applyEngagement(row, patch) {
  if (patch.read === true) { row.read = true; row.seen = true; }
  if (patch.read === false) { row.read = false; }
  if (patch.archived === true) { row.archived = true; row.seen = true; }
  if (patch.archived === false) { row.archived = false; row.seen = true; }
}
```

Mutation failure fixture must leave the pre-action row state unchanged.

- [ ] **Step 3: Implement mark-all-read**

```js
for (const row of notifications) {
  if (!row.archived && !row.read) {
    row.read = true;
    row.seen = true;
  }
}
```

Failure fixture must not mutate any row/count.

- [ ] **Step 4: Implement source-open boundary**

On success:

```text
mark selected row read+seen
→ show explicit B03 boundary state containing exact document code/id + message id
→ do not render Document Official/Discussion itself
```

On armed access-drift `404`:

```text
remove the stale row from the presentable fixture set
→ refetch/render current Inbox
→ show neutral reconciliation message
→ do not navigate
```

---

### Task 3: Prove seen is presentation-driven, not fetch-driven

**Files:**
- Modify: `docs/work/current/t11-b08-notifications-full-inbox-functional-wireframe.html`

**Interfaces:**
- Consumes: unseen rows from Task 1.
- Produces: visible P8 evidence of bounded op84-style batching.

- [ ] **Step 1: Observe materially presented unseen rows**

Use `IntersectionObserver` on rendered row elements with a deterministic threshold and short dwell before enqueueing the row id.

```js
const observer = new IntersectionObserver(entries => {
  for (const entry of entries) {
    if (entry.isIntersecting && entry.intersectionRatio >= 0.6) {
      queueSeen(entry.target.dataset.notificationId);
    }
  }
}, { threshold: [0.6] });
```

- [ ] **Step 2: Flush bounded local seen batch**

The P8 fixture may expose the queued ids in the review console, then apply `seen=true` only to those ids after the simulated op84 succeeds.

- [ ] **Step 3: Prove off-screen second-page rows stay unseen after fetch**

Load page 2 into the list while its rows remain below the visible viewport. The review console must still report those ids as unseen until the operator scrolls them into material view.

---

### Task 4: Implement cursor, failures and realtime invalidation

**Files:**
- Modify: `docs/work/current/t11-b08-notifications-full-inbox-functional-wireframe.html`

**Interfaces:**
- Consumes: current lens/page state.
- Produces: recoverable continuation and invalidation states.

- [ ] **Step 1: Implement `Carregar mais` with a one-shot failure toggle**

Failure leaves page-1 rows rendered and retryable. Success appends the next canonical fixture window without re-sorting.

- [ ] **Step 2: Implement one-shot per-item failure and mark-all-read failure toggles**

Both must preserve authoritative pre-action fixture state and announce recovery intent.

- [ ] **Step 3: Implement SSE invalidation simulation**

The review control emits only the concept `notifications.changed`; handler then performs a fixture `refetchCanonical()` that may reveal a newly created server-side row.

Do not pass row data through the simulated SSE call.

- [ ] **Step 4: Implement initial-load unavailable and all three empty-state fixtures**

The operator must be able to inspect:

```text
Inbox unavailable + retry
active empty
unread empty
archived empty
```

---

### Task 5: Integrate B01/B01N structure and responsive/accessibility behavior

**Files:**
- Modify: `docs/work/current/t11-b08-notifications-full-inbox-functional-wireframe.html`

**Interfaces:**
- Consumes: locked global shell + Quick Inbox semantics.
- Produces: one coherent B08 P8 surface.

- [ ] **Step 1: Reuse low-fidelity global shell geometry**

Keep the locked sidebar model unchanged and keep Notifications absent as a permanent sidebar item.

- [ ] **Step 2: Provide functional bell + Quick Inbox overlay**

It consumes the same fixture engagement state/counts as the Full Inbox. `Ver todas` closes/returns to the already-open Full Inbox rather than creating a second state store.

- [ ] **Step 3: Responsive transformation**

Desktop uses one dominant Inbox column with trailing actions. Narrow layout stacks row recognition and moves secondary actions into a compact row action disclosure without changing semantics.

- [ ] **Step 4: Accessibility structure**

Use semantic buttons/tabs, text for `Nova` and `Não lida`, an `aria-live` reconciliation region, focus-visible controls and explicit accessible action names. Preserve focus after an action removes a row from the current lens where possible.

---

### Task 6: Verify and hand off P8 R1

**Files:**
- Verify: `docs/work/current/t11-b08-notifications-full-inbox-functional-wireframe.html`

**Interfaces:**
- Produces: exact canonical P8 artifact for operator use.

- [ ] **Step 1: Parse HTML and check static duplicate ids**

Run a local parser script and require zero parser errors attributable to malformed structure and zero duplicated static `id` values.

- [ ] **Step 2: Extract inline JavaScript and syntax-check it**

Run `node --check` against the extracted script and require exit code 0.

- [ ] **Step 3: Static forbidden-scope scan**

Require the Product surface to contain no controls for search, snooze, priority, preferences, delete, bulk archive, source reply/editor/viewer or B09/Audit work. Review-only fixture controls must be visibly labeled non-Product.

- [ ] **Step 4: Verify repository blob equals delivered chat artifact**

Compute the Git blob SHA from the local HTML and compare it with the blob returned after repository creation.

- [ ] **Step 5: Fresh exact-HEAD CI**

After all repository writes, require the `required` job and repository-standard envelope to pass on the exact final HEAD before claiming P8 R1 ready.

- [ ] **Step 6: Operator handoff**

Deliver the exact `.html` file in chat and ask the operator to exercise fixed lenses, read/unread, archive/unarchive, mark-all-read, page-2 seen behavior, access drift, failures, SSE refetch, Quick Inbox coherence and mobile layout. Do not open P9/P10 or B09 before explicit B08 LOCK.
