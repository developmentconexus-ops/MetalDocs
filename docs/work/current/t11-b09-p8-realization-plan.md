# B09 Audit P8 Functional Evidence Realization Plan

> **Review status:** FABLE-ADJUDICATED / I-1 + M-1..M-7 INCORPORATED / P8 NOT STARTED.  
> **For agentic workers:** REQUIRED SUB-SKILL: use the repository P8 method. This plan realizes disposable frontend-planning evidence only; it MUST NOT be treated as Product implementation authorization.

**Goal:** Produce one browser-operable low-fidelity HTML wireframe that lets the operator falsify or confirm the operator-ratified B09 Audit Investigation Ledger structure.

**Architecture:** One self-contained HTML file with semantic HTML, low-fidelity CSS, vanilla JavaScript and deterministic in-memory fixtures. The prototype simulates only the already-ratified op78/op87/op88/op89 frontend truth boundary; it does not implement APIs, Authorization, persistence, owner services, production routing, React state, generic search infrastructure or Product writes.

**Tech Stack:** HTML5, CSS, vanilla JavaScript, deterministic local fixtures/state, browser History API only for prototype query-state simulation.

**Spec:** `docs/work/current/t11-b09-audit-r1.md` + `docs/work/current/t11-b09-p7-exit.md` + `docs/work/current/t11-b09-fable-review-adjudication.md` + `docs/decisions/audit-investigation-read.md`.

## Global Constraints

- Product implementation remains BLOCKED.
- Canonical P8 medium is one `.html` file under `docs/work/current/`.
- Canonical artifact path: `docs/work/current/t11-b09-audit-functional-wireframe.html`.
- No React, backend/API calls, OpenAPI client, persistence, real Authorization evaluator, admin directory, generic entity resolver or production router.
- Default `/audit` posture is recent-first with no hidden time cutoff.
- Canonical order remains `occurred_at DESC,event_id DESC`; no alternate sort.
- Main query dimensions are exactly Period, Historical Scope, Actor, Action and Resource.
- Draft filter state is distinct from the applied query; URL/chips/ledger represent applied truth only.
- Applied URL state must round-trip: initial load, refresh, browser back/forward and pasted canonical query strings reconstruct the same applied investigation from stable IDs/enums/UTC instants.
- Applied chips are semantic-dimension atoms: Period clears `from+before`, USER Actor clears `actor_kind+actor_user_id`, and exact Resource clears `resource_kind+resource_id` together.
- Relative period preset names exist only in the draft editor. A refreshed/copied URL reconstructs an exact interval label, never a newly evaluated `Hoje`/`Últimos 7 dias`/`Últimos 30 dias` label.
- Query Assist is simulated as server-authored bounded options; loaded Audit rows never become selector completeness authority.
- op88/op89 max-20 results may show a non-claiming refine-search hint but must not invent `has_more`.
- `actor_kind=user` without exact `actor_user_id` is not a Launch Product filter; the ratified P7 disposition rejects an all-human actor category filter.
- Evidence and current recognition remain structurally distinct.
- Detail uses the already-loaded inspection item; no detail endpoint or route exists.
- Cursor continuation appends older events; no infinite scroll, page numbers or total count.
- B09 is read-only in the business domain.
- Owner handoffs terminate at explicit existing-route boundaries; the prototype MUST NOT rebuild B03/B06/B07 or unopened admin blocks.
- Loading, known-empty and failure are separate states.
- Low fidelity only: no final palette, typography, iconography, tokens, production component styling or animation authority.
- P8 must be delivered as the exact `.html` bytes in chat for operator operation before any B09 LOCK.

---

### Task 1: Build the deterministic Audit evidence fixture model

**Files:**
- Create: `docs/work/current/t11-b09-audit-functional-wireframe.html`

**Interfaces:**
- Consumes: ratified `AuditInspectionItem`, op78 predicates/order/cursor law and optional recognition semantics.
- Produces: local immutable `allEvents`, applied/draft query state, loaded-page state and review-only failure toggles used by later tasks.

- [ ] **Step 1: Define a closed local event shape that keeps evidence and recognition separate**

Use objects equivalent to:

```js
{
  evidence: {
    eventId: '11111111-1111-4111-8111-111111111101',
    occurredAt: '2026-08-24T12:42:17Z',
    actor: { kind: 'user', userId: '21111111-1111-4111-8111-111111111101' },
    operationCode: 'governance.accepted',
    resource: { kind: 'document', resourceId: '31111111-1111-4111-8111-111111111101' },
    visibility: { kind: 'area', areaId: '41111111-1111-4111-8111-111111111101' },
    facts: {
      governanceAttemptId: '51111111-1111-4111-8111-111111111101',
      documentId: '31111111-1111-4111-8111-111111111101'
    }
  },
  recognition: {
    actor: { currentLabel: 'Marina Costa' },
    resource: { stableLabel: 'PO-023' },
    visibilityArea: { stableLabel: 'COM', currentLabel: 'Comercial' }
  }
}
```

- [ ] **Step 2: Cover the P7 evidence variants with deterministic rows**

The fixture set must include at least:

```text
USER actor with current display name
USER actor with recognition unavailable
SYSTEM actor
Document with immutable stable code
Area resource with stable code + current name
resource with current recognition only
resource with no safe recognition -> kind + compact UUID fallback
AREA historical visibility
COMPANY historical visibility
Governance typed facts with governance_attempt_id
release/cancellation/obsolescence facts with document_id
an event with no admitted owner handoff
multiple events on the same local day
at least one continuation page crossing into an older local day
```

Create enough rows for two deterministic op78 windows of 20 events each plus a final short page so `Load older events` and end-of-results behavior are both operable.

Every Area offered later by the simulated op87 selector MUST have at least one corresponding historically visible fixture event. If the selector offers `COM`, `FIN` and `RH`, the fixture set must contain admitted Area evidence for all three; the prototype must not demonstrate an op87 option impossible under the ratified candidate law.

- [ ] **Step 3: Keep fixture-only review controls out of Product truth**

Review flags may exist only under one object such as:

```js
const reviewMode = {
  failNextMainQuery: false,
  failNextContinuation: false,
  failAreaAssist: false,
  failActorAssist: false,
  failResourceAssist: false,
  forceKnownEmpty: false,
};
```

Do not add Product fields such as `reviewed`, `resolved`, `comment`, `case`, `exported`, `priority`, `permission`, `allowed`, `currentState` or `historicalNameSnapshot`.

---

### Task 2: Implement applied-query truth and draft-filter editing

**Files:**
- Modify: `docs/work/current/t11-b09-audit-functional-wireframe.html`

**Interfaces:**
- Consumes: Task 1 evidence fixtures.
- Produces: `draftQuery`, `appliedQuery`, canonical prototype search params, chips and first-page reload behavior.

- [ ] **Step 1: Define separate draft and applied query objects**

```js
const emptyQuery = () => ({
  occurredAtFrom: null,
  occurredAtBefore: null,
  visibilityAreaId: null,
  actorKind: null,
  actorUserId: null,
  operationCodes: [],
  resourceKind: null,
  resourceId: null,
});

let appliedQuery = emptyQuery();
let draftQuery = emptyQuery();
```

- [ ] **Step 2: Implement canonical query serialization and visible canonical evidence**

Serialize only applied stable IDs/enums/UTC instants. Multiple operation codes must serialize in the closed local enum order, never in user click order.

```js
function serializeAppliedQuery(q) {
  const p = new URLSearchParams();
  if (q.occurredAtFrom) p.set('occurred_at_from', q.occurredAtFrom);
  if (q.occurredAtBefore) p.set('occurred_at_before', q.occurredAtBefore);
  if (q.actorKind) p.set('actor_kind', q.actorKind);
  if (q.actorUserId) p.set('actor_user_id', q.actorUserId);
  if (q.operationCodes.length) p.set('operation_codes', canonicalizeCodes(q.operationCodes).join(','));
  if (q.resourceKind) p.set('resource_kind', q.resourceKind);
  if (q.resourceId) p.set('resource_id', q.resourceId);
  if (q.visibilityAreaId) p.set('visibility_area_id', q.visibilityAreaId);
  return p.toString();
}
```

Always render the canonical query string in a clearly review-only evidence region so the operator can compare the applied question with `location.search`.

History API use is defensive P8 evidence, not routing authority:

```js
function writePrototypeUrl(q, mode = 'push') {
  const search = serializeAppliedQuery(q);
  renderCanonicalQueryEvidence(search);
  const url = `${location.pathname}${search ? `?${search}` : ''}${location.hash}`;
  try {
    const fn = mode === 'replace' ? 'replaceState' : 'pushState';
    history[fn]({ auditQuery: search }, '', url);
  } catch (error) {
    announceReviewStatus('Não foi possível atualizar a URL local; a consulta canônica continua visível nesta evidência P8.');
  }
}
```

- [ ] **Step 3: Implement URL deserialization on initial load and `popstate`**

The P8 artifact MUST prove the reverse direction, not merely write a decorative URL.

```js
function parseCanonicalQuery(search) {
  const p = new URLSearchParams(search);
  const q = emptyQuery();
  q.occurredAtFrom = p.get('occurred_at_from');
  q.occurredAtBefore = p.get('occurred_at_before');
  q.actorKind = p.get('actor_kind');
  q.actorUserId = p.get('actor_user_id');
  q.operationCodes = canonicalizeCodes((p.get('operation_codes') || '').split(',').filter(Boolean));
  q.resourceKind = p.get('resource_kind');
  q.resourceId = p.get('resource_id');
  q.visibilityAreaId = p.get('visibility_area_id');
  return q;
}

function restoreAppliedQueryFromLocation() {
  const next = parseCanonicalQuery(location.search);
  appliedQuery = cloneQuery(next);
  draftQuery = cloneQuery(next);
  closeDetailIfOpen();
  resetLoadedPages();
  renderCanonicalQueryEvidence(serializeAppliedQuery(next));
  runFirstPageQuery({ writeHistory: false });
}

window.addEventListener('popstate', restoreAppliedQueryFromLocation);
```

On initial boot, call `restoreAppliedQueryFromLocation()` before rendering canonical results. This is the proof path for refresh and pasted/copied query strings. Parsing must preserve the ratified dependency laws; the prototype must not manufacture an applied `resource_id` without `resource_kind` or `actor_user_id` without `actor_kind=user`.

- [ ] **Step 4: Implement Apply and explicit unapplied-changes state**

`Aplicar` must:

```text
validate draft
→ copy draft into applied
→ canonicalize prototype URL
→ close detail if open
→ clear loaded continuation pages
→ remove old ledger rows
→ show first-page loading
→ resolve deterministic op78 simulation
→ update chips + ledger
```

Invalid period (`from >= before`) remains draft-only with inline error and leaves applied URL/chips/ledger unchanged.

- [ ] **Step 5: Implement immediate applied-chip removal as semantic-dimension clearing**

Removing a chip must alter the applied query immediately, update the canonical prototype URL and trigger a new first-page query without requiring a second Apply click.

Dependent predicates are indivisible chip dimensions:

```js
function clearAppliedDimension(dimension) {
  if (dimension === 'period') {
    appliedQuery.occurredAtFrom = null;
    appliedQuery.occurredAtBefore = null;
  }
  if (dimension === 'actor') {
    appliedQuery.actorKind = null;
    appliedQuery.actorUserId = null;
  }
  if (dimension === 'resource') {
    appliedQuery.resourceKind = null;
    appliedQuery.resourceId = null;
  }
}
```

`Limpar edição` resets only draft controls to the currently applied query. `Limpar filtros` from a known-empty result immediately applies the unfiltered query.

---

### Task 3: Implement the five Audit-specific investigation controls

**Files:**
- Modify: `docs/work/current/t11-b09-audit-functional-wireframe.html`

**Interfaces:**
- Consumes: Task 2 draft-query state.
- Produces: operable Period, Scope, Actor, Action and Resource editors without a generic filter engine.

- [ ] **Step 1: Period popover**

Provide exactly:

```text
Hoje
Últimos 7 dias
Últimos 30 dias
Personalizado: from date/time + to date/time
```

No preset is selected by default. Whole-day custom input converts to local start boundaries with an exclusive next-day upper bound.

Preset names are only draft-editor conveniences. After Apply, the authoritative value is the exact canonical UTC interval. If state is reconstructed from `location.search`, render the applied chip as that exact interval in local presentation time; never recalculate and label an old canonical interval as `Hoje` on a later date.

- [ ] **Step 2: Historical Scope selector using simulated op87 options**

Use a bounded deterministic list such as:

```js
const areaOptions = [
  { areaId: '41111111-1111-4111-8111-111111111101', code: 'COM', currentName: 'Comercial' },
  { areaId: '41111111-1111-4111-8111-111111111102', code: 'FIN', currentName: 'Financeiro' },
  { areaId: '41111111-1111-4111-8111-111111111103', code: 'RH', currentName: 'Recursos Humanos' },
];
```

Default means all currently auditable evidence. Do not add a Company-only filter. Each option must be backed by at least one historically visible fixture event from Task 1.

The selector must expose review-operable `loading`, `known-empty` and `failure + retry` states using `reviewMode.failAreaAssist` without turning failure into an empty option set.

- [ ] **Step 3: Actor Query Assist**

Expose `Sistema` as a closed local option. USER search uses a deterministic simulated op88 result list with separate `loading`, `known-empty` and `failure + retry` review paths. Typed text that is never selected must not become an actor predicate.

Do not offer `Todos os usuários` or any equivalent `actor_kind=user` category filter; it is explicitly REJECTED for Launch because no ratified Auditor job requires human-vs-system category analysis.

If the simulated op88 response returns exactly 20 items, display a non-claiming hint equivalent to:

```text
Mostrando até 20 opções. Refine a busca para localizar outro resultado.
```

Do not say that more results definitely exist and do not invent pagination/`has_more`.

- [ ] **Step 4: Action multi-select**

Represent the closed 37-value action vocabulary in human groups. The local search box filters only these 37 labels. Group headers are presentation only and cannot be selected as backend query categories. Multi-select remains OR inside the action dimension.

- [ ] **Step 5: Resource kind-first Query Assist**

Require resource kind before exact resource selection. Simulated op89 results must vary by selected kind. Changing kind clears any previous exact resource id. Show an exact-UUID search result path and recognition-unavailable fallback. Do not offer a universal resource search.

If the simulated op89 response returns exactly 20 items, use the same non-claiming refinement hint as Actor. Do not invent `has_more` or Query Assist pagination.

---

### Task 4: Build the Evidence Ledger and continuation behavior

**Files:**
- Modify: `docs/work/current/t11-b09-audit-functional-wireframe.html`

**Interfaces:**
- Consumes: applied query + filtered deterministic page windows.
- Produces: dense desktop ledger, compact narrow list, local-day separators and cursor continuation.

- [ ] **Step 1: Render the desktop ledger hierarchy**

Use the visible headings:

```text
Quando | Ator | Ação | Recurso | Escopo histórico
```

Do not render `event_id`, raw `operation_code`, full UUIDs or typed facts as main-row text.

- [ ] **Step 2: Implement recognition priority**

```js
function recognitionPrimary(kind, evidenceId, label) {
  if (label?.stableLabel) return label.stableLabel;
  if (label?.currentLabel) return label.currentLabel;
  return `${humanKind(kind)} · …${evidenceId.slice(-4)}`;
}
```

Actor recognition absence uses neutral copy such as `Nome atual indisponível` plus `Usuário · …xxxx`. SYSTEM renders as `Sistema`.

- [ ] **Step 3: Render historical scope truthfully**

AREA rows show immutable code first and optional current Area name second. COMPANY shows `Empresa inteira`. Never label the ledger column merely `Área`.

- [ ] **Step 4: Insert non-interactive local-day separators**

Separators follow the local presentation date without changing canonical sort, cursor or server grouping semantics. Appending a page within the same day must not duplicate the separator; crossing a day must insert the new separator.

- [ ] **Step 5: Implement cursor continuation**

Start with 20 items. `Carregar eventos anteriores` appends the next deterministic window. A one-shot continuation failure leaves every already-rendered row intact and exposes retry. Final page removes the button and displays an end-of-available-events message. No infinite scroll, page number or total count.

---

### Task 5: Implement the contextual Detail Drawer

**Files:**
- Modify: `docs/work/current/t11-b09-audit-functional-wireframe.html`

**Interfaces:**
- Consumes: one already-loaded `AuditInspectionItem`.
- Produces: exact evidence inspection, typed-fact presentation, Audit-native follow-up queries and bounded owner handoff boundaries.

- [ ] **Step 1: Open detail from a semantic per-row action**

Every row must contain a real `Ver detalhes` button. Whole-row activation may be a convenience, never the sole control.

Desktop opens one side drawer while preserving ledger position and selected-row state. Narrow/mobile opens a full working surface with `Voltar` restoring the ledger position.

- [ ] **Step 2: Render the ratified detail hierarchy**

```text
RESUMO
EVIDÊNCIA CANÔNICA
FATOS DO EVENTO
RECONHECIMENTO ATUAL
INVESTIGAR NO AUDIT
CONTEXTO ATUAL
```

Canonical evidence must include exact UTC `occurred_at`, full `event_id`, raw `operation_code`, exact actor identity, `resource_kind + resource_id` and exact historical visibility. Local time is presentation context only.

- [ ] **Step 3: Render typed facts semantically**

Use closed label maps by fixture event family, for example:

```js
const factLabels = {
  governanceAttemptId: 'Tentativa de governança',
  documentId: 'Documento',
  revisionId: 'Revisão',
  priorRevisionId: 'Revisão anterior',
};
```

Do not expose raw JSON, schema viewer, developer mode or arbitrary key dumping as the Product presentation.

- [ ] **Step 4: Implement Audit-native shortcuts as immediate applied queries**

`Mesmo ator`, `Mesmo recurso` and `Mesma ação` must directly update applied query + URL, close detail and load the first page. They do not enter draft and wait for Apply.

- [ ] **Step 5: Implement only admitted owner-lens boundaries**

Show explicit non-rendering boundary states for:

```text
Document -> /documents/:document_id
Document history -> /documents/:document_id/history
Governance case -> /work/governance/:governance_attempt_id
```

Label the section `Contexto atual`. State that destination authorization/disclosure is independently rechecked. Do not render B03/B06/B07 content and do not create admin deep links.

---

### Task 6: Prove loading, known-empty, failure and authorization distinctions

**Files:**
- Modify: `docs/work/current/t11-b09-audit-functional-wireframe.html`

**Interfaces:**
- Consumes: review-only failure toggles.
- Produces: operator-inspectable recovery states without false evidence claims.

- [ ] **Step 1: First-page loading**

When a new query is applied, remove old ledger rows while the new query is loading. Do not leave old evidence dimmed beneath new chips/URL.

- [ ] **Step 2: Known-empty**

Render copy equivalent to:

```text
Nenhum evento corresponde a esta investigação.
Tente remover ou alterar algum filtro.
[Limpar filtros]
```

Do not claim that no history exists.

- [ ] **Step 3: First-page failure**

Keep applied URL/chips and render `Não foi possível consultar as evidências` + `Tentar novamente`. Do not display an empty ledger as if the query succeeded.

- [ ] **Step 4: Query Assist loading/empty/failure for all three server-authored assists**

Area/Actor/Resource assist must separately expose:

```text
Buscando…
Nenhuma opção encontrada
Não foi possível buscar… / Tentar novamente
```

Use `failAreaAssist`, `failActorAssist` and `failResourceAssist` independently. A failed assist must never be rendered as a known-empty result.

- [ ] **Step 5: Route authorization review state**

A review-only toggle may demonstrate that absence of `audit.read` terminates at a shell/route boundary. It must not render a B09 empty result. Label this clearly as a review state, not a frontend Authorization implementation.

---

### Task 7: Integrate shell geometry, sticky context, responsive and accessibility behavior

**Files:**
- Modify: `docs/work/current/t11-b09-audit-functional-wireframe.html`

**Interfaces:**
- Consumes: locked B01 global shell/IA and all B09 interactions.
- Produces: one coherent low-fidelity B09 surface.

- [ ] **Step 1: Reuse locked global shell geometry without redesigning B01**

Keep Audit under the locked `Evidência` navigation home. Do not add a second Audit route or expose backend nouns as navigation.

- [ ] **Step 2: Implement minimal sticky behavior**

Desktop keeps only investigation/applied-query context and ledger column headings sticky. Title, day separators and row actions remain normal flow. Drawer scrolls independently.

- [ ] **Step 3: Responsive transformation**

At narrow width:

```text
ledger table -> compact semantic event list
full investigation bar -> compact Filtros · N control + sheet/full-surface editor
detail drawer -> full-surface detail
cursor semantics -> unchanged
```

Do not remove historical-scope meaning, stable identity fallback or applied-query truth on mobile.

- [ ] **Step 4: Accessibility structure**

Use semantic buttons, labels/fieldsets where appropriate, keyboard-operable popovers/sheets, visible focus, `aria-live` for query/retry status, Escape close for desktop detail, focus move into detail heading and focus restoration to invoking control on close. Programmatically expose selected row state.

---

### Task 8: Verify and hand off B09 P8 R1

**Files:**
- Verify: `docs/work/current/t11-b09-audit-functional-wireframe.html`

**Interfaces:**
- Produces: exact canonical B09 P8 artifact for operator operation.

- [ ] **Step 1: Parse HTML and check static duplicate ids**

Run a local HTML parser check and require zero malformed-structure errors attributable to the artifact and zero duplicate static `id` values.

- [ ] **Step 2: Extract inline JavaScript and syntax-check it**

Run `node --check` against the extracted script and require exit code 0.

- [ ] **Step 3: Static forbidden-scope scan**

Require zero Product controls for:

```text
export
saved searches
free-text global Audit search
custom sort
column chooser/reorder/pin
total count/page numbers
bulk selection
comments/annotations
case management
mark reviewed/read
operational mutation
raw JSON/developer mode
generic resource resolver
admin-directory browsing
all-human actor category filter
```

Review-only scenario toggles must be visibly separated from Product controls.

- [ ] **Step 4: Exercise the P7 falsification matrix**

Manually verify in-browser:

```text
recent-first unfiltered entry with no hidden period
draft filters do not alter ledger/chips/URL
Apply changes query and replaces rows through loading
invalid period remains draft-only
compound Period/Actor/Resource chip removal clears the whole semantic dimension
Period presets + custom interval
reloaded/copied period renders exact interval, not a newly evaluated relative preset label
Scope op87-style selection
all op87 options correspond to historically visible fixture evidence
Area Query Assist loading + known-empty + failure/retry
Actor SYSTEM + exact USER typeahead + empty/failure
no all-human actor category filter
Actor exactly-20 response shows non-claiming refine-search hint
Action multi-select + local 37-label search
Resource kind-first + op89-style search + kind change clears exact id
Resource exactly-20 response shows non-claiming refine-search hint
human recognition + unavailable fallbacks
AREA + COMPANY historical-scope presentation
local-day separators
20-row first page + append continuation + failure/retry + end state
detail exact evidence + typed facts + current recognition separation
same actor/resource/action immediate investigation
bounded owner handoff boundaries
known-empty vs first-page failure
refresh preserves applied query
browser back/forward navigates applied queries
pasted canonical query string reproduces the investigation
visible canonical query evidence matches applied URL/query state
History API failure guard preserves operable in-page canonical evidence
mobile filter sheet + full-surface detail
keyboard/focus/escape restoration
```

- [ ] **Step 5: Verify repository blob equals delivered chat artifact**

Compute the Git blob SHA of the final HTML bytes. Materialize those exact bytes as a local chat attachment and require the repository blob SHA to match before handoff.

- [ ] **Step 6: Fresh exact-HEAD CI**

After repository creation, require the exact final HEAD `required` job and Repository Standard v1 envelope to succeed before claiming P8 R1 ready for operator review.

- [ ] **Step 7: Operator handoff**

Deliver the exact `.html` attachment in chat and ask the operator to operate the P7 falsification matrix. Do not mark B09 LOCKED and do not open P9/P10/B10 until explicit operator approval of the operated P8 artifact.
