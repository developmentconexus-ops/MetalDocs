# T11 — B09 Audit R1 — P7 Ratified Design

> **Status:** OPERATOR-RATIFIED / FABLE-ADJUDICATED CLARIFICATIONS.  
> **Block:** B09 — Audit.  
> **Method:** Frontend Product Experience Planning Method v2.3 + DevelopmentConexus Engineering Method.  
> **Predecessors:** B01–B08 LOCKED / OPERATOR-RATIFIED; B09-F1 CLOSED / OPERATOR-RATIFIED.  
> **Durable authority:** `../../decisions/audit-investigation-read.md`.  
> **Finding ledger:** `t11-b09-audit-upstream-replan.md`.  
> **P7 clean exit:** `t11-b09-p7-exit.md`.  
> **Fable adjudication:** `t11-b09-fable-review-adjudication.md`.  
> **P8:** ELIGIBLE / NOT STARTED; explicit operator execution authorization still required.  
> **Implementation:** BLOCKED.

## 1. P7 purpose and exit posture

This document records the operator-ratified B09 P7 leading hypothesis after the bounded Audit capability reopen and 89-operation FP0 rebaseline.

Binding Method v2.3 law:

```text
material user need
+ current authority insufficient
= blocking UPSTREAM FINDING
```

P7 exits only when the leading hypothesis has no unresolved material authority gap. This ratified design records the chosen structure, the credible alternatives rejected, the exact authority dependencies, responsive/accessibility behavior, scale assumptions, failure behavior and the required P7 disposition matrix before any P8 HTML exists.

The independent Fable review found no P7 architecture failure or upstream reopen need. The operator accepted its bounded clarifications M-1, M-2, M-5 and M-6 here; they remove ambiguity without changing Product/API authority.

No Product code, schema, OpenAPI, runtime, deployment or T12 work is authorized by this design.

## 2. Human jobs and Product boundary

Launch Audit supports two ratified jobs:

```text
A. point investigation / exact evidence question
B. period + authorized historical-scope review
```

The frontend job is therefore not merely to display a reverse-chronological event feed. It must let an authorized auditor quickly narrow, inspect and continue evidence without confusing current recognition with immutable historical proof.

Product boundary:

```text
Audit
  = observe / prove / investigate immutable evidence

owner lens
  = inspect or act on current resource state under that owner's rules
```

B09 remains read-only in the business domain. No administrative or corrective mutation is owned by Audit.

## 3. Reference evidence retained from P6/B09-F1

The reference study used mature audit products as task-pattern evidence rather than feature checklists:

```text
GitHub Audit Log
  structured narrowing by time, actor, action and resource context

Microsoft Purview Audit
  time, activity, user and object narrowing for large evidence sets

Veeva Vault
  structured audit-trail inspection

Qualio Audit Trail
  date, user, action and document-oriented investigation
```

Bounded conclusion:

> MetalDocs Audit must be efficiently investigable by real evidence questions; it does not need a generic search, analytics, export or query-builder platform at Launch.

## 4. Credible P7 hypotheses compared

### H1 — Evidence Ledger + horizontal investigation bar + detail drawer — LEADING

Strengths:

```text
high scan density
strong recognition/comparison
preserves width for evidence
supports recent-first exploration and exact investigation
fits fixed server order + seek cursor
keeps immutable evidence distinct from current recognition
responsive reflow is credible
owner handoffs remain secondary
```

### H2 — Permanent filter sidebar + evidence ledger — REJECTED

Rejected because the sidebar consumes horizontal space continuously for controls that are episodic while the evidence ledger benefits materially from width. It also weakens narrow-screen viability without improving backend truth fit.

### H3 — Query-first page or card/timeline-first results — REJECTED

A query-first empty landing weakens recent evidence exploration. Card-first loses density. Timeline-first suggests a narrative/history model closer to Document History and is a poorer fit for cross-resource Audit investigation.

Also rejected as unnecessary Launch complexity:

```text
spreadsheet-style column configuration
column resize/reorder/pin
grouping controls
alternate views
density selector
bulk row selection
infinite scroll
numbered pages
```

## 5. Leading hypothesis — Audit Investigation Ledger

Stable route:

```text
/audit
```

Default entry:

```text
recent-first
no hidden temporal cutoff
no dashboard cards/totals/charts
```

An unfiltered `/audit` loads the first page of all events currently auditable by the caller under canonical order:

```text
occurred_at DESC,event_id DESC
```

“Recent” is presentation posture only. It never means an implicit last-7/30-day predicate.

High-level composition:

```text
Audit
  investigation bar
  applied-filter chips
  evidence ledger
  explicit cursor continuation
  contextual detail drawer
```

## 6. Investigation bar

Desktop uses a horizontal progressive-disclosure bar:

```text
[Período] [Escopo] [Ator] [Ação] [Recurso]   [Limpar edição] [Aplicar]
```

The bar preserves ledger width while keeping the five ratified query dimensions directly discoverable. Each control has Audit-specific semantics; B09 does not introduce a generic filter-engine Product pattern.

Narrow/mobile collapses this into a compact control such as:

```text
[Filtros · N] [Aplicar*]
```

and opens a sheet/full-surface editor using the same semantic dimensions.

### 6.1 Draft vs applied query

Two states are binding:

```text
draft filters
  = editor controls currently being changed

applied query
  = URL + chips + ledger/result truth
```

When draft differs from applied:

```text
controls show draft
ledger remains last applied result
URL remains last applied query
chips remain last applied query
explicit “alterações não aplicadas” state is visible
Apply is enabled only when draft is valid and different
```

`Aplicar` validates and promotes draft to applied query, canonicalizes the URL, closes an open detail drawer, replaces the ledger with first-page loading, then renders the new first page.

`Limpar edição` only resets the draft editor. It does not silently change the applied evidence question until Apply.

Applied chips represent result truth. Removing an applied chip is an explicit immediate query action:

```text
remove chip
→ update canonical URL
→ close drawer if open
→ first-page query
→ replace results
```

Chip granularity follows semantic query dimensions, not individual wire members. Dependent predicates are one compound chip and removal clears the whole dimension:

```text
Período
  occurred_at_from + occurred_at_before
  = one chip

Ator USER
  actor_kind=user + actor_user_id
  = one chip

Recurso exato
  resource_kind + resource_id
  = one chip
```

This prevents chip removal from constructing wire-invalid combinations such as `resource_id` without `resource_kind` or `actor_user_id` without `actor_kind=user`.

The known-empty recovery action `Limpar filtros` likewise immediately applies the unfiltered query.

No undo stack, autosaved draft, query-history subsystem or revert button is introduced.

## 7. Filter construction

### 7.1 Period

One popover provides explicit convenience presets plus a custom interval:

```text
Hoje
Últimos 7 dias
Últimos 30 dias
Personalizado: from date/time + to date/time
```

No preset is selected by default.

Preset semantics are explicit and converted at Apply time into canonical UTC instants:

```text
Hoje
  local start of current calendar day -> apply-time instant

Últimos 7 dias
  local start of the calendar day 6 days before today -> apply-time instant

Últimos 30 dias
  local start of the calendar day 29 days before today -> apply-time instant
```

Preset names are draft-editor conveniences only. Once applied, URL authority is the exact canonical `from/before` interval. A chip reconstructed from a refreshed or copied URL renders that exact interval in local presentation time; it must not re-label the interval as `Hoje`/`Últimos 7 dias`/`Últimos 30 dias` based on a later clock date.

Custom whole-day selection uses local calendar boundaries and an exclusive next-day upper bound. Custom time selection converts the chosen local instants directly to UTC.

The wire remains:

```text
occurred_at_from   inclusive
occurred_at_before exclusive
```

Invalid `from >= before` remains draft-only and shows inline validation; it never replaces the currently applied query.

### 7.2 Historical scope

The control label is `Escopo`, but applied evidence and the ledger use the explicit term `Escopo histórico`.

Default:

```text
Todos os eventos que posso auditar
```

Area options come only from op87 and display immutable code first with optional current name:

```text
COM · Comercial
FIN · Financeiro
```

Applied identity is `visibility_area_id`, never the mutable current name.

Launch does not invent a Company-only historical-scope filter. Company-attributed events remain visible in the default all-admitted query when authorized.

### 7.3 Actor

Actor is a specific selector, not a free-text evidence predicate.

```text
SYSTEM
  closed frontend option

USER
  op88 Query Assist
  selected user_id is query identity
```

Selecting USER applies:

```text
actor_kind=user
actor_user_id=<uuid>
```

Selecting SYSTEM applies:

```text
actor_kind=system
```

Typed text that is not selected never becomes an applied actor predicate. Query Assist loading, known-empty and failure states are distinct.

When op88 returns exactly its max-20 result set, the UI may state only a bounded refinement hint such as `Mostrando até 20 opções. Refine a busca para localizar outro resultado.` It must not claim that additional results definitely exist because the response carries no `has_more` authority.

A category filter for “all human actors” (`actor_kind=user` without `actor_user_id`) is not part of Launch. It is explicitly REJECTED because neither ratified Auditor job requires human-vs-system category analysis: point investigation uses an exact USER identity, while SYSTEM is itself the closed system actor. A future proven comparative automation-vs-human job may reopen this decision.

### 7.4 Action

Action is a local multi-select over the closed 37-value `AuditOperationCode` vocabulary.

The UI may group human labels by semantic family for recognition, for example Documents, Governance, Access and Organization/Configuration. Group membership is presentation only; groups never enter backend query semantics.

A local text field may filter the 37 human labels inside the selector. This is not Audit full-text search.

Applied selection remains the exact unique canonical enum set. Multiple selected actions are OR within the action dimension and are canonicalized into closed enum order for stable query representation.

### 7.5 Resource

Resource construction is kind-first:

```text
1. select resource_kind
2. optionally narrow to an exact resource through op89
```

`resource_kind` alone is a valid applied predicate. Exact resource selection adds `resource_id` and stores stable identity, not the displayed label.

Changing resource kind clears any draft exact resource identity that belonged to the previous kind.

Query Assist uses only op89 candidates admitted by Audit-visible evidence. No admin directory, generic entity platform or loaded-page-derived completeness is used.

When op89 returns exactly its max-20 result set, the same non-claiming refinement hint applies: the UI may say it is showing up to 20 options and invite a more specific search, but it may not invent `has_more`.

## 8. Applied URL state

Only the applied evidence question is durable browser state.

Canonical search params may include:

```text
occurred_at_from
occurred_at_before
actor_kind
actor_user_id
operation_codes
resource_kind
resource_id
visibility_area_id
```

Human labels never become URL authority. IDs/enums are canonical.

Properties:

```text
refresh preserves the applied question
browser back/forward navigates applied questions
a copied URL reproduces the question, not authorization
the server rechecks audit.read + historical visibility
draft edits are local only
cursor / number of loaded continuation pages are ephemeral
```

After reloading a query for which 80 events had previously been appended, the browser starts again from the first page of that same applied question.

## 9. Evidence Ledger

Desktop primary representation is a dense ledger/table:

```text
Quando | Ator | Ação | Recurso | Escopo histórico
```

The main row deliberately omits:

```text
event_id
raw operation_code
full UUIDs
typed facts
other technical evidence details
```

Rows prioritize human recognition while keeping stable identity visible where useful.

Examples:

```text
Marina Costa
Usuário · …9c31

PO-023
Documento · …a710
```

The action column uses a human label. The canonical `operation_code` belongs in detail.

A canonical semantic `Ver detalhes` action is available per row. Whole-row activation may be a convenience but cannot be the only accessible action.

### 9.1 Local-day separators

The ledger inserts lightweight non-interactive separators by local presentation date:

```text
Hoje · 24 ago 2026
Ontem · 23 ago 2026
```

They do not collapse, count, query, group backend data or alter canonical order. Continuation crossing a day inserts the next separator; continuation inside the same local day does not repeat it unnecessarily.

### 9.2 Time presentation

Ledger time is local/browser-user presentation for scanability.

Detail shows both:

```text
Local
24/08/2026 09:42:17
America/Sao_Paulo · UTC−03:00

Recorded evidence
2026-08-24T12:42:17Z
```

`occurred_at` UTC remains canonical evidence. Relative-time text is not evidence authority.

## 10. Recognition fallback

Recognition never suppresses or rewrites evidence.

Priority:

```text
accepted stable human label
→ primary

otherwise safe current recognition
→ primary + stable identity secondary

otherwise
→ neutral type label + compact stable UUID suffix
```

Examples:

```text
Nome atual indisponível
Usuário · …9c31

Documento · …a710
```

SYSTEM is displayed as `Sistema`.

The ledger never invents historical names, leaves a material identity blank or uses a full UUID as primary scan text. Full exact identity remains available in detail.

## 11. Historical scope presentation

Ledger heading is explicitly:

```text
Escopo histórico
```

For Area evidence:

```text
COM
Comercial · nome atual
```

Immutable Area code is stable recognition. Current Area name is secondary current recognition only.

If current recognition is unavailable:

```text
COM
Área · …62ac
```

Company historical visibility displays humanly as:

```text
Empresa inteira
```

The detail section exposes exact historical visibility kind/id. The UI must not imply that historical scope establishes current resource ownership or current Area membership.

## 12. Continuation and scale

Launch uses the op78 default first-page size:

```text
20 events
```

Continuation is an explicit control:

```text
[Carregar eventos anteriores]
```

Behavior:

```text
cursor continuation
→ append next page beneath current rows
→ preserve existing rows and scroll context
```

No infinite scroll, offset pagination, page numbers or total count is shown.

`has_more=false` removes the continuation action and presents an end-of-available-events message for the applied query.

A new applied query discards accumulated continuation pages and starts again from the first page.

## 13. Detail Drawer

Desktop opens one contextual side drawer at a time while preserving ledger scroll position and selected-row indication. The drawer has independent scroll and does not trap the rest of the desktop page; filters/continuation remain operable.

Narrow/mobile detail occupies the full working surface. `Voltar` restores the exact ledger position.

Accessibility behavior:

```text
canonical semantic Ver detalhes control
Close control
Escape closes on desktop
focus moves into detail heading/content on open
focus returns to invoking control/row on close
selected event state is programmatically perceivable
```

Query change while detail is open closes detail and replaces the ledger. Loading older events while detail is open keeps the current detail open and appends rows beneath the ledger.

No `/audit/events/:event_id` route and no Audit detail endpoint are introduced. Detail uses the already-loaded `AuditInspectionItem`.

### 13.1 Drawer hierarchy

```text
RESUMO
  human action
  actor
  resource
  local time

EVIDÊNCIA CANÔNICA
  occurred_at UTC
  event_id
  operation_code
  actor identity
  resource_kind + resource_id
  historical visibility

FATOS DO EVENTO
  typed facts with human labels and exact values

RECONHECIMENTO ATUAL
  optional, safe, explicitly non-historical context

INVESTIGAR NO AUDIT
  same actor
  same resource
  same action

CONTEXTO ATUAL
  admitted owner handoffs only
```

Typed facts use closed semantic renderers/label maps for the closed Audit event families. B09 does not expose raw JSON, schema inspection or developer mode.

## 14. Investigation shortcuts and owner handoff

Audit-native shortcuts are primary detail actions:

```text
Mesmo ator
Mesmo recurso
Mesma ação
```

Each is an explicit immediate new applied query. It updates the canonical URL, closes detail and loads the first page. It does not first enter draft and wait for a second Apply.

Secondary current-context handoffs appear only when already-admitted identities/routes exist:

```text
Document resource
  → /documents/:document_id

release / revision cancellation / obsolescence + admitted document_id
  → /documents/:document_id/history

governance Decision + admitted governance_attempt_id
  → /work/governance/:governance_attempt_id
```

The section label is `Contexto atual` to prevent historical/current-state conflation. Every destination independently rechecks current authorization/disclosure.

No generic `Abrir recurso`, URL resolver, links array, admin deep-link platform or row kebab menu is introduced.

## 15. Loading, known-empty and failure states

Binding law:

```text
loading != known-empty != failure
```

### New first-page query

After the applied URL/chips change, old result rows leave the evidence surface and the ledger shows explicit loading until the new first page resolves. Old rows are not shown dimmed beneath a new applied question.

### Known-empty

A valid completed query with zero rows states only that no event matched that investigation, never that history does not exist.

Recovery offers removal/adjustment of applied filters and an immediate `Limpar filtros` action.

### First-page failure

The applied URL/chips remain so the same question is recoverable. The evidence region shows query failure + `Tentar novamente`, not an empty ledger.

### Continuation failure

Already loaded evidence remains intact. An inline continuation error and retry control replace only the failed continuation action.

### Query Assist states

Each op87/op88/op89-dependent selector distinguishes:

```text
loading
known-empty
actionable failure + retry
```

A failed Query Assist must never appear as `Nenhuma opção encontrada`.

### Authorization

Absence of `audit.read` is route/shell authorization behavior, not a B09 empty state. Audit must not imply `no evidence` when the actual state is `not authorized to read evidence`.

## 16. Sticky/context-preservation behavior

Desktop uses only the sticky context that materially improves long investigations:

```text
sticky
  investigation bar / applied-query context
  ledger column header

normal flow
  event rows
  local-day separators
  continuation action
```

The title, day separators and individual-row actions are not separately frozen.

Mobile uses a compact sticky filter summary/control rather than the full desktop bar.

The open desktop drawer has independent scroll; the ledger preserves its own scroll position.

## 17. Responsive behavior

Desktop:

```text
dense semantic table/ledger
horizontal investigation bar
contextual side drawer
```

Narrow/mobile:

```text
compact event list preserving same information priority
filter sheet/full-surface editor
detail full-surface with Back
same applied-query semantics
same cursor continuation
same evidence/current-recognition distinction
```

Responsive adaptation may change arrangement and visible secondary text density; it may not change query semantics, authority, action ownership or evidence meaning.

## 18. Material writes — P7 disposition

B09 owns no business-domain write.

Allowed frontend state changes are local navigation/query state only:

```text
draft filters
applied URL/query
selected event
loaded cursor pages
open/closed detail
```

Explicit Product dispositions:

```text
operational mutation from Audit       REJECTED — belongs to semantic owner
mark event read/reviewed               REJECTED — no Launch job/state model
comments/annotations                   DEFERRED — collaboration/case subsystem unproven
saved searches                         DEFERRED — current Launch scope unproven
export                                 DEFERRED — external evidence handoff unproven
investigation case management          DEFERRED — new subsystem unproven
```

## 19. P7 authority-disposition matrix

| Leading-hypothesis requirement | P7 disposition | Current authority / Product reason |
|---|---|---|
| Recent-first admitted evidence | PRESENT-IN-AUTHORITY | op78 |
| Exact time interval | PRESENT-IN-AUTHORITY | `occurred_at_from` / `occurred_at_before` |
| Historical Area narrowing | PRESENT-IN-AUTHORITY | `visibility_area_id` |
| Area Query Assist | PRESENT-IN-AUTHORITY | op87 |
| USER actor identity | PRESENT-IN-AUTHORITY | `actor_kind=user` + `actor_user_id` + op88 |
| SYSTEM actor | PRESENT-IN-AUTHORITY | closed actor enum + frontend option |
| All-human actor category filter | REJECTED — Product reason | no distinct Launch Auditor job requires human-vs-system category analysis; exact USER identity + SYSTEM cover ratified jobs |
| Multi-action narrowing | PRESENT-IN-AUTHORITY | `operation_codes[]` |
| Human action labels/grouping | PRESENT-IN-AUTHORITY | closed frontend mapping over enum |
| Resource-kind narrowing | PRESENT-IN-AUTHORITY | `resource_kind` |
| Exact resource narrowing | PRESENT-IN-AUTHORITY | `resource_id` + op89 |
| Optional safe recognition | PRESENT-IN-AUTHORITY | `AuditEventRecognition` |
| Recognition fallback | PRESENT-IN-AUTHORITY | evidence kind + stable identity |
| Query Assist max-20 refinement affordance | PRESENT-IN-AUTHORITY | op88/op89 bounded max-20 response; UI may invite refinement without asserting `has_more` |
| Typed facts | PRESENT-IN-AUTHORITY | closed `AuditEventView` evidence union |
| Detail inspection | PRESENT-IN-AUTHORITY | loaded `AuditInspectionItem`; no new endpoint |
| Same actor/resource/action | PRESENT-IN-AUTHORITY | new op78 query |
| Owner handoffs | PRESENT-IN-AUTHORITY | bounded accepted routes/identities |
| Applied query in URL | PRESENT-IN-AUTHORITY | frontend representation of canonical predicates |
| Cursor continuation | PRESENT-IN-AUTHORITY | op78 seek cursor |
| Total count | REJECTED — Product reason | no proven job; absent from cursor authority |
| Numbered pages | REJECTED — Product reason | false model over seek cursor/no total |
| Custom sort | REJECTED — Product reason | canonical evidence order is sufficient |
| Column chooser/grouping | REJECTED — Product reason | YAGNI; no Auditor job proven |
| Free-text Audit search | DEFERRED — scope reason | structured investigation covers Launch jobs |
| Export | DEFERRED — scope reason | requires named external handoff/package proof |
| Saved searches | DEFERRED — scope reason | persistence/library job not proven |
| Annotations/case management | DEFERRED — scope reason | new subsystem not proven |
| Operational mutation | REJECTED — Product reason | belongs to current-state semantic owner |
| Audit detail endpoint | REJECTED — Product reason | loaded item is sufficient |
| Generic deep-link resolver | REJECTED — Product reason | bounded owner handoffs are sufficient |

## 20. P7 clean-exit proof

Required Method v2.3 declarations:

```text
fields / summaries
  PRESENT — ledger columns + drawer evidence hierarchy

identity sources
  PRESENT — immutable event IDs/resource IDs/user IDs/codes + optional recognition

pagination / scale
  PRESENT — default 20 + explicit seek-cursor continuation

sort / filter
  PRESENT — fixed canonical order + exact ratified structured predicates

preview / content truth
  PRESENT — bounded typed facts + recognition split; no raw-content preview invented

material writes
  NONE in Audit business domain; operational writes REJECTED, future collaboration DEFERRED
```

Result:

```text
material frontend need without authority     0
unresolved B09 upstream finding              0
screen-shaped API invention                  0
backend-shaped UX suppression                0
browser filtering as evidence truth          0
Audit/current-state reconstruction            0
```

The independent Fable review confirmed that P7 itself stands: no architecture-level blocker or upstream reopen is required. The operator adjudicated all Fable findings before P8 execution.

## 21. Gate after Fable adjudication

```text
B09 P7   CLOSED / OPERATOR-RATIFIED
Fable    ADJUDICATED
B09 P8   ELIGIBLE / NOT STARTED; explicit execution authorization required
B09 P9-P10 NOT OPEN
B10-B12  NOT OPEN
T12      NOT OPEN
Product implementation BLOCKED
merge    NOT AUTHORIZED
```

The next methodological step, only after explicit operator authorization, is P8 functional low-fidelity HTML using deterministic local fixtures and material interactions. P8 remains Evidence, not production implementation, and must be operated by the operator before any B09 LOCK.