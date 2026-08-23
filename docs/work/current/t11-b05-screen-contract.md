# T11 — B05 P9 Screen Contract / Bidirectional Trace

> **Status:** COMPLETE / POST-LOCK PROOF.  
> **Block:** B05 — My Work / Work Queues.  
> **Depends on:** B05 LOCK, B05-F1/F3/F4, current T6/T8-E Work projections and T8-F frontend realization.  
> **Implementation:** BLOCKED.

## 1. Goal

Prove that every material region/control in the locked B05 functional wireframe is realizable by current authority without inventing write behavior, a generic Work authority, a second priority model, a frontend global ordering authority or a screen-shaped API.

## 2. Screen contract

| Surface / region | User goal | Current read truth | Material control / request | Identity source | Material failure / safe UX | Forbidden frontend authority | Status |
|---|---|---|---|---|---|---|---|
| B05-01 `/work` lens | enter assigned-work surface | `listAuthoringWork` / `listGovernanceWork` according to selected presentation | lane selection only | stable `/work` route + browser presentation state | lane read failure isolated to selected queue | no new route or generic work aggregate | READY |
| B05-02 Minha Caixa header | understand surface purpose | locked B01 IA + current lane | none | none new | no total KPI invented | no cross-lane priority summary | READY |
| B05-03 intent switch | choose `Para aprovação` vs `Em edição` | separate current Work projections | switch lane / initiate selected list read | browser presentation state | switching clears stale row selection; server truth refetched as needed | no merged cross-lane feed | READY |
| B05-04 governance recognition row | identify governed subject before opening | op55 `WorkGovernanceItem.document + revision + subject_kind + created_at` | row selection/open only | returned `governance_attempt_id`, Document/Revision refs | missing/stale row is not case authority | no per-row B06 enrichment | READY |
| B05-05 governance deadline presentation | know exact deadline / whether attention is overdue | op55 `WorkGovernanceItem.due_at?` | none | returned persisted active-Step `due_at` | absent deadline -> truthful `Sem prazo` | no manual urgency/priority state | READY |
| B05-06 relative deadline label | scan time remaining quickly | returned `due_at` + presentation clock | presentation formatting only | no new Product identity | clock drift may alter label only; exact `due_at` remains visible | relative label is not lifecycle/state authority | READY |
| B05-07 deadline filter | narrow governance work by deadline horizon | op55 first-page `deadline_filter?` | `Todos / Atrasados / Próximas 24h / Próximos 7 dias / Sem prazo` | filter enum + server-trusted anchor when relative | empty filtered result is not global no-work truth | no arbitrary date range / generic filter DSL | READY |
| B05-08 governance order | see time-critical work first | op55 canonical server order `due_at ASC NULLS LAST, document.code, governance_attempt_id` | none | server cursor/order | frontend preserves returned order | no client global re-sort | READY |
| B05-09 relative-filter cursor anchor | keep one filter meaning across pages | op55 opaque cursor binds filter + first-page anchor + order + seek position | load continuation with cursor + optional limit only | opaque cursor | invalid/tampered cursor -> safe reload/fresh first page | no client-generated time anchor authority | READY |
| B05-10 load more | continue queue without false totals | op54/op55 `Page.next_cursor` / `has_more` law | continuation request | returned opaque cursor | failed continuation preserves already shown rows but does not claim completeness | no total count / offset model | READY |
| B05-11 governance handoff | continue exact approval work | selected op55 row | navigate `/work/governance/:attempt_id` | returned `governance_attempt_id` | stale/non-disclosable destination -> refresh My Work | no approve/return mutation in B05 | READY |
| B05-12 authoring recognition row | identify current authoring work | op54 `WorkAuthoringItem` | row selection/open only | returned Document/Revision refs | stale row resolved by destination/current refetch | no DRAFT/content preview in queue | READY |
| B05-13 authoring handoff | continue current Work | selected op54 row | navigate B04 `/documents/:document_id/work` | returned `document_id` | B04 resolves current truth again | no B05 authoring state machine | READY |
| B05-14 keyboard traversal | move quickly through dense queue | already-loaded row order | ArrowUp/ArrowDown + Enter | selected row index is ephemeral UI | selection clamps/resets when collection changes | no persisted selected-work authority | READY |
| B05-15 stale-row recovery | recover when projection changed before entry | fresh op54/op55 read | refresh current lane | current list operation | removed row disappears; changed row re-renders | no resurrection of stale work | READY |
| B05-16 list failure / retry | distinguish unknown from empty | failed op54/op55 read | retry same first-page intent | none new | no cached stale list presented as current truth | no offline/saved queue authority | READY |
| B05-17 empty state | know current selected queue/filter has no rows | successful empty op54/op55 page | clear/change filter where applicable | current lane + current deadline filter | filtered empty is explicitly scoped to filter | no global “all work complete” inference | READY |
| B05-18 B01N global chrome reuse | preserve application-wide attention affordance | locked Notification chrome | source navigation remains B01N/B08-owned | Notification source identity | B05 does not interpret Notification business truth | no B05 Notification store | READY |
| B05-19 responsive reflow | preserve queue semantics on narrow screens | same op54/op55 truths | presentation-only stack/reflow | none new | material deadline/identity/open affordance remains accessible | no mobile-specific work semantics | READY |

## 3. Exact operation homes used by B05

Current retained operation numbering:

```text
54  listAuthoringWork    GET /api/v1/work/authoring
55  listGovernanceWork   GET /api/v1/work/governance
```

B05 performs no Product write.

Owner-lens continuation is navigation only from this block:

```text
Authoring row
→ B04 route
→ B04 current truth reads/writes remain B04-owned

Governance row
→ B06 route boundary
→ B06 will load exact Governance Case truth when that block opens
```

B05 adds no operation 87+.

## 4. Bidirectional trace

### Product/backend -> frontend

```text
WorkAuthoringPage
→ Em edição focused queue
→ Document/Revision/title/state/owner/updated-at recognition

WorkGovernancePage
→ Para aprovação focused queue
→ governed Revision/title/subject recognition
→ exact active-Step due_at when present

listGovernanceWork F4 order
→ rendered governance row order

first-page deadline_filter + server anchor
→ bounded deadline subset
→ cursor-bound continuation semantics

Page.next_cursor / has_more
→ Carregar mais
```

### Frontend -> Product/backend

```text
choose Em edição
→ op54 first page

choose Para aprovação
→ op55 first page

choose deadline preset
→ fresh op55 first page with exact deadline_filter

Carregar mais
→ op54/op55 cursor + optional limit only

fresh refresh
→ new first-page request
→ relative governance filter receives a fresh server anchor

open authoring row
→ navigate to B04 stable Work lens

open governance row
→ navigate to B06 stable Governance Case boundary
```

Unbound material controls: **0**.  
Invented application operations: **0**.  
Screen-shaped APIs: **0**.

## 5. Client state classes

B05 uses only accepted state classes:

```text
SERVER STATE
  WorkAuthoringPage
  WorkGovernancePage
  returned Page/cursor truth

NAVIGATION / URL
  stable /work route
  selected Minha Caixa intent may be encoded as presentation state/query
  selected deadline preset may be encoded as presentation state/query

FORM DRAFT
  none

EPHEMERAL UI
  selected row
  loading/retry state
  relative-time formatted label
  open/closed Quick Inbox chrome state
  responsive presentation state
```

No localStorage saved-filter baseline, offline queue, global Work entity store or client priority model is admitted.

## 6. Material deadline / cursor laws

```text
exact due_at
  = server-projected persisted active-Step deadline

OVERDUE label
  = presentation predicate over exact due_at and current display clock
  = not GovernanceStepState

op55 default order
  = due_at ASC NULLS LAST
  + document.code ASC
  + governance_attempt_id ASC

relative deadline_filter first page
  = server captures trusted anchor A

continuation
  = opaque cursor reuses authenticated A
  = frontend supplies no replacement deadline_filter/time anchor

fresh first page
  = new request
  = new trusted anchor for relative filter
```

No client-side page reordering or dynamic “today/week” calendar interpretation exists.

## 7. Material failure intent

```text
401 / 403
  current access truth wins; no stale cached queue becomes authority

400 invalid/tampered cursor
  discard failed continuation token
  issue a fresh first-page read under current lane/filter intent

list read dependency/transport failure
  show unknown/load-failure state + retry
  never translate into empty queue

successful empty filtered governance page
  state means empty for that exact filter traversal only

stale destination after row open
  owner lens rejects/does not disclose
  return/refresh current My Work projection

clock advances during relative-filter cursor traversal
  visual relative labels may change
  cursor filter anchor does not change
```

Final copywriting remains outside P9; these are semantic message intents.

## 8. Access / authority proof

```text
row presence != permanent work entitlement
row identity != Governance Case authority
row due_at != mutation authority
OVERDUE label != lifecycle state
deadline filter != Authorization filter
Carregar mais != complete frozen snapshot
selected row != reservation/claim
Em edição label != DRAFT-only invariant
B05 navigation != owner-lens truth
Notification chrome != My Work authority
```

Every owner destination rechecks current truth.

## 9. P9 closure

```text
material B05 regions/controls traced        19 / 19
unbound material controls                   0
invented operations                         0
operation 87+                               absent
screen-shaped APIs                          0
frontend Authorization evaluator            0
frontend global queue sorter                 0
manual priority state                        0
per-row B06 enrichment                       0
navigation identities unsourced              0
material B05 Screen Contract findings        0
```

P9 is complete for the locked B05 scope.