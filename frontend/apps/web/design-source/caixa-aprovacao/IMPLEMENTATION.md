# Caixa de Aprovação — Implementation Worksheet

> **Slug:** `caixa-aprovacao`
> **Owning feature:** `features/approval`
> **Target route:** `/approvals`
> **Reference:** `./caixa-aprovacao.html` + `./NOTES.md`
> **Skill version:** 1.2
> **Started:** 2026-05-08
> **Completed:** —

---

## Open Questions Log

| # | Phase | Question | User answer | Resolved |
|---|---|---|---|---|
| 1 | 0 | "Aprovar e assinar" / "Devolver" — open `SignoffDialog` inline or navigate to `/registry-v2/:id`? | Navigate to document (mock behavior) until signoff flow designed. No real signoff in this pass. | ✅ |
| 2 | 0 | Fields absent from API (code, kind, deadline, urgency, summary, changes, version) — cut or mock? | KEEP everything. Use hardcoded mock data flagged with TODO comments. Add all to backlog. No visual cuts. | ✅ |

---

## Phase 0 — Audit (HARD GATE)

### 0.1 Element-by-element audit

Strategy: **implement full visual fidelity**. Missing API fields → hardcoded mock data flagged with `// TODO [BACKLOG: caixa-aprovacao.md]: needs <field> from API`. All mock fields tracked in `wiki/backlog/caixa-aprovacao.md`.

| Element (HTML region) | Maps to (state / role / persona / data) | Keep / Cut / Defer | Reason |
|---|---|---|---|
| Rail section="approvals" | `features/approval`, approver persona | **Keep** | Standard shell wiring |
| InboxToolbar: "APROVAÇÕES › Caixa de entrada" | Section breadcrumb | **Keep** | Real label |
| InboxToolbar: "Filtros" button | Area filter (API supports `area_code`) | **Keep** | Rendered as `aria-disabled` placeholder — filter panel deferred to backlog |
| InboxToolbar: Foco / Linha do tempo view-switcher | Local + persisted (`localStorage['md.inbox.v']`) | **Keep** | Core feature, full implementation |
| InboxStack: "SUA FILA · N decisões" | `total` from API | **Keep** | Real data |
| InboxStack: "N vencem hoje" | Mock — `InboxItem.deadline_at` missing | **Keep (mock)** | `// TODO [BACKLOG]: needs deadline_at field` |
| Queue rail: numbered items | `items[]` from API | **Keep** | Real list |
| Queue item: code chip (`POP-QUA-0148`) | Mock — `InboxItem.code` missing | **Keep (mock)** | `// TODO [BACKLOG]: needs code field` |
| Queue item: kind badge (`POP`/`IT`/`POL`/`DC`) | Mock — `InboxItem.kind` missing | **Keep (mock)** | `// TODO [BACKLOG]: needs kind field` |
| Queue item: urgency blink dot | Mock — `InboxItem.urgent` missing | **Keep (mock)** | `// TODO [BACKLOG]: needs urgent field` |
| Queue item: deadline countdown | Mock — `InboxItem.deadline_at` missing | **Keep (mock)** | Show mock "3h" etc; `// TODO [BACKLOG]` |
| Queue item: title | `item.document_title` | **Keep** | Real |
| Queue item: area | `item.area_code` | **Keep** | Real |
| Ghost cards (z-stack) | Pure CSS decoration | **Keep** | No data needed |
| Card header: urgent gradient vs. neutral | Mock — `InboxItem.urgent` missing | **Keep (mock)** | Derive from mock urgent field |
| Card header: "VENCE EM" countdown | Mock — `InboxItem.deadline_at` missing | **Keep (mock)** | Mock value; `// TODO [BACKLOG]` |
| Card header: kind badge | Mock | **Keep (mock)** | Same as queue item |
| Card header: code + version (`v2.3 → v2.4`) | Mock — `InboxItem.code`, `.version` missing | **Keep (mock)** | `// TODO [BACKLOG]: needs code, version fields` |
| Card header: title (h2) | `item.document_title` | **Keep** | Real |
| Card body: summary paragraph | Mock — `InboxItem.summary` missing | **Keep (mock)** | `// TODO [BACKLOG]: needs summary field` |
| Card body: AUTOR | `item.submitted_by` + `Avatar` | **Keep** | Real (`area_code` for subtitle) |
| Card body: ALTERAÇÕES count | Mock — `InboxItem.changes` missing | **Keep (mock)** | `// TODO [BACKLOG]: needs changes field` |
| Card body: ESTÁGIO | `item.stage_label` + parse `quorum_progress` ("1/3") | **Keep** | Real |
| Card actions: "Abrir documento" | `navigate('/registry-v2/' + item.controlled_document_id)` | **Keep** | Real navigation |
| Card actions: "Devolver" | Mock — navigate to doc (signoff flow not designed) | **Keep (mock)** | `// TODO [BACKLOG]: wire to signoff flow` |
| Card actions: "Aprovar e assinar →" | Mock — navigate to doc | **Keep (mock)** | `// TODO [BACKLOG]: wire to signoff/approval flow` |
| Keyboard shortcuts footer (A/D/←/→) | **Defer** | **Defer** | Non-critical UX; add after core complete |
| Prev/Next nav buttons | Local `selectedIdx` state | **Keep** | |
| InboxTimeline: "X decisões" header | `total` from API | **Keep** | Real |
| InboxTimeline: heatmap sparkline | Mock — no history API | **Keep (mock)** | Static mock array; `// TODO [BACKLOG]: needs GET /approval/my-decisions?days=14` |
| Timeline: deadline buckets (Hoje/Amanhã/Esta semana/Próximo mês) | Group by mock `deadline_at` on mock items | **Keep (mock)** | Mock grouping; `// TODO [BACKLOG]: needs deadline_at` |
| Timeline item: time + "vence em" | Mock | **Keep (mock)** | Mock value |
| Timeline item: code + kind chip | Mock | **Keep (mock)** | Same as queue |
| Timeline item: title + meta (author, area, changes) | Real title/author/area; changes mock | **Keep** | |
| Timeline item: stage progress bars | Parse `quorum_progress` ("1/3") | **Keep** | Real if parseable |
| Timeline item: "Revisar →" | `navigate('/registry-v2/' + item.controlled_document_id)` | **Keep** | Real |

### 0.2 Cut list confirmed

- [x] User confirmed: no cuts — full visual, mock data for missing API fields
- [x] User confirmed: action buttons navigate to document (mock signoff) in this pass

**Mock fields → backlog (all tracked in `wiki/backlog/caixa-aprovacao.md`):**
- `InboxItem.code` — document code (e.g. `POP-QUA-0148`)
- `InboxItem.kind` — document type short (POP/IT/POL/DC)
- `InboxItem.deadline_at` — absolute deadline timestamp
- `InboxItem.urgent` — boolean urgency flag
- `InboxItem.summary` — revision summary text
- `InboxItem.changes` — edit count integer
- `InboxItem.version` — version transition string (e.g. `v2.3 → v2.4`)
- `GET /api/v1/approval/my-decisions?days=14` — history for heatmap sparkline
- Signoff flow — "Aprovar e assinar" / "Devolver" → design + implement approval path

---

## Phase 1 — Map (HARD GATE)

### 1.1 Reusability scan — backward

| Design element | Existing primitive | Path | Action |
|---|---|---|---|
| Author avatar (sm/xs) | `Avatar` | `components/ui/Avatar.tsx` | use `sm`; add `xs` size in Phase 2 (primitive gap) |
| Filter icon, eye icon | `Icon` | `components/ui/Icon.tsx` | use |
| Doc code chip (`POP-QUA-0148`) | `CodeChip` | `components/ui/CodeChip.tsx` | use (mock data) |
| Segmented view-switcher | `TabBar` | `components/ui/TabBar.tsx` | NOT suitable — design uses icon+label buttons in a pill; build inline in `InboxToolbar` |
| Timeline grouped buckets | `TimelineRail` | `components/ui/TimelineRail.tsx` | NOT suitable — too simple; `InboxTimeline` needs custom grouped structure |

### 1.2 Reusability scan — forward

| Name | Generic? | 2+ screens? | Placement | Rationale |
|---|---|---|---|---|
| `InboxToolbar` | No | No | `features/approval/components/InboxToolbar.tsx` | Approval section header + view-switcher |
| `InboxStack` | No | No | `features/approval/components/InboxStack.tsx` | Foco view — queue rail + card stack layout |
| `InboxApprovalCard` | No | No | `features/approval/components/InboxApprovalCard.tsx` | Active approval card (header + body + actions) |
| `InboxTimeline` | No | No | `features/approval/components/InboxTimeline.tsx` | Linha do tempo view — grouped deadline buckets |

Queue rail items and ghost cards are inline sub-components within `InboxStack` (not reused elsewhere, <50 lines each).

### 1.3 Component decomposition

```
InboxPage
├── InboxToolbar (view state + setter)
└── view === 'stack'
    └── InboxStack (items[], selectedIdx)
        ├── aside.queueRail
        │   ├── queue header (count, "vencem hoje" mock)
        │   └── button[] queue items (code, title, area, deadline)
        └── main
            ├── counter + prev/next buttons
            ├── div.ghostCard × 2 (CSS only)
            └── InboxApprovalCard (item, onNavigate)
                ├── div.cardHeader (kind, code, version, title, deadline)
                ├── div.cardBody
                │   ├── summary paragraph
                │   ├── stats grid (AUTOR | ALTERAÇÕES | ESTÁGIO)
                │   └── actions row (Abrir / Devolver / Aprovar)
                └── (keyboard hints deferred)
    view === 'timeline'
    └── InboxTimeline (items[], total)
        ├── header (title + heatmap widget)
        └── section[] × buckets (Hoje/Amanhã/Esta semana/Próximo mês)
            ├── bucket dot + label
            └── div.itemRow[] (time, kind, code, title, author, stage bars, Revisar →)
```

### 1.4 Status / enum meta SSOT

No new workflow status enum. Existing `ApprovalState` in `approvalTypes.ts` covers states.

New: `features/approval/lib/mockInboxData.ts` — holds the mock-enrichment function.
```ts
// TODO [BACKLOG: caixa-aprovacao.md]: All fields in RichInboxItem beyond InboxItem base are MOCK
export type RichInboxItem = InboxItem & {
  code: string;       // TODO: needs InboxItem.code
  kind: string;       // TODO: needs InboxItem.kind
  deadline: string;   // TODO: needs InboxItem.deadline_at
  urgent: boolean;    // TODO: needs InboxItem.urgent
  summary: string;    // TODO: needs InboxItem.summary
  changes: number;    // TODO: needs InboxItem.changes
  version: string;    // TODO: needs InboxItem.version
};
export function enrichInboxItem(item: InboxItem, idx: number): RichInboxItem
```

### 1.5 State design

| Type | Item | Notes |
|---|---|---|
| Server state | `useInboxQuery()` | `features/approval/queries/useInboxQuery.ts` — wraps `listInbox()`, uses `QK.inbox()` |
| Local state | `selectedIdx: number` | `InboxStack` — which item is active |
| Persisted | `view: 'stack' \| 'timeline'` | lazy `useState(() => (localStorage.getItem('md.inbox.v') as ViewType) ?? 'stack')` in `InboxPage` |
| Debounced | none in this pass | Filter panel deferred |

### 1.6 Backend contract

| Endpoint | Path | Status | Missing for full design | Backlog |
|---|---|---|---|---|
| List inbox | `GET /api/v1/approval/inbox` | existing | `code`, `kind`, `deadline_at`, `urgent`, `summary`, `changes`, `version` | `wiki/backlog/caixa-aprovacao.md` |
| Decision history | `GET /api/v1/approval/my-decisions?days=14` | **needed** | `{ date: string; count: number }[]` | `wiki/backlog/caixa-aprovacao.md` |
| Signoff flow | design + endpoint TBD | **needed** | approve/reject path + password challenge | `wiki/backlog/caixa-aprovacao.md` |

Mock fallback strategy:
- `enrichInboxItem()` in `features/approval/lib/mockInboxData.ts` adds mock values for all missing fields
- Heatmap sparkline: static mock array `[3,5,2,7,4,6,8,5,3,9,4,7,5,2]` flagged with `// TODO [BACKLOG]`
- Action buttons: `navigate('/registry-v2/' + item.controlled_document_id)` — flagged `// TODO [BACKLOG]`
- Every mock value has co-located `// TODO [BACKLOG: caixa-aprovacao.md]` comment

**Pre-existing tech debt** (not in scope — separate backlog):
- `approvalApi.ts` uses raw `fetch` instead of `lib/api/client.ts` — tracked separately

### 1.7 User review checkpoint

- [x] Phase 0 confirmed by user 2026-05-08
- [x] Reusability classifications reviewed (Avatar xs = Phase 2 fix; TimelineRail not suitable)
- [x] Backend contract reviewed (mock strategy agreed)
- [x] No open Phase-1 questions

---

## Phase 2 — Pre-flight (advisory)

_To be filled after Phase 1._

---

## Phase 3a — Structure mirror (HARD GATE)

_To be filled after Phase 2._

---

## Phase 3b — Style port (HARD GATE)

_To be filled after Phase 3a._

---

## Phase 3c — State wiring (advisory)

_To be filled after Phase 3b._

---

## Phase 4 — Verify (HARD GATE)

_To be filled after Phase 3c._

---

## Phase 5 — Document (advisory)

_To be filled after Phase 4._
