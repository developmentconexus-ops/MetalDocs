# Biblioteca (Library)

**Owning feature:** `features/documents`
**Target route:** `/documents` (protected — inside AppShell)
**Page file:** `features/documents/pages/LibraryPage.tsx + LibraryPage.module.css`
**Design source:** `selected-library.jsx` → `SelectedLibrary` (canonical reference)
**Companion refs:** `library.jsx` (alt explorations), `screens/01 Biblioteca.html`, `styles.css`

---

## Audit findings (run before implementing — see `wiki/concepts/design-workflow-audit.md`)

The mockup was AI-generated without domain knowledge of MetalDocs document states, RBAC roles, or operator personas. Findings below.

### Real model (verified 2026-05-06)

- **Document states (Spec 2, `internal/modules/documents/approval/domain/state.go:11`):** 8 — `draft`, `under_review`, `approved`, `rejected`, `scheduled`, `published`, `superseded`, `obsolete`.
- **Legacy strings rejected:** `finalized`, `archived` → `ErrLegacyStateRejected`.
- **List endpoint:** `GET /api/v2/documents` (`internal/modules/documents/delivery/http/handler.go:185`). RBAC-gated to `system_admin` or `document_filler`. Admin sees all; filler sees own only. **No pagination params, no filter params.** Returns flat array.
- **Approval instance states (`approval/domain/instance.go:20`):** `pending`, `approved`, `rejected`. Distinct from doc state.
- **No "frozen" state.** "Frozen" is freeze-and-hashing concept (immutability snapshot, see `wiki/concepts/freeze-and-hashing.md`), not a document status. Closest mapping: `published`.

### Mockup states vs real states

| Mockup status | Maps to real | Action |
|---|---|---|
| `draft` | `draft` | ✅ Keep |
| `review` | `under_review` | ✅ Keep, rename label |
| `approved` | `approved` | ✅ Keep |
| `rejected` | `rejected` | ✅ Keep |
| `frozen` | ❌ no state | ❌ **Cut** — replace with `published` if "current effective" was the intent |
| `archived` | ❌ legacy | ❌ **Cut** — use `obsolete` if "decommissioned" was the intent |
| (missing) | `scheduled` | ➕ Add |
| (missing) | `superseded` | ➕ Add |
| (missing) | `obsolete` | ➕ Add |

### Stat strip (4 cards) — Keep / Cut / Defer

| Card | Decision | Reason |
|---|---|---|
| `Em revisão` | ✅ Keep | Maps to `count(status = under_review)` |
| `Aprovação pendente` | ⚠ Defer | Distinct from `under_review` only via approval instance state. Needs join `documents` × `approval_instances`. Cut from first commit; add when stats endpoint lands |
| `Frozen este mês` | ❌ Cut | "Frozen" is not a state; "este mês" is invented metric. If we want "published this month" rephrase + verify column exists |
| `Próx. revisão` | ❌ Cut | We don't track scheduled review dates — would need new field `next_review_at`. No backend support |

**Recommend first commit ships: 1 card (`Em revisão`) only. Or: 0 cards (cut entire strip; revisit when stats endpoint exists).**

### Filter tabs — Keep / Cut / Defer

| Tab | Decision | Reason |
|---|---|---|
| `Todos` | ✅ Keep | Default view |
| `Meus` | ✅ Keep | Backend already filters by author for `document_filler` role; admin needs client filter `createdBy === currentUser.userId` |
| `Em revisão` | ✅ Keep | Maps to `status === 'under_review'` |
| `Aprovação pendente` | ⚠ Defer | Same as stat card — needs approval-instance join |
| `Frozen` | ❌ Cut | Replace with `Publicados` (`status === 'published'`) |
| `Rascunhos` | ✅ Keep | Maps to `status === 'draft'` |
| (missing) | ➕ Add `Rejeitados` | Real state, useful for authors to find rework |
| (missing) | ➕ Add `Obsoletos` | Real state, useful for audit trail |

**Recommend tab set:** `Todos / Meus / Rascunhos / Em revisão / Publicados / Rejeitados / Obsoletos`. Cut `Aprovação pendente` until stats endpoint joins approval instances. Counts come from stats endpoint (deferred — see Backend gaps).

### Activity sidebar — persona check

The mockup ships pending-approvals + audit-trail sidebar default-open. Audit:

- **Pending approvals widget** — useful only for `approver` persona (someone with active approval-instance assignments). For viewer / author personas it's noise.
- **Audit trail stream** — useful for compliance / admin personas. For everyday viewers it's noise.

**Decision:** Keep collapsible sidebar, **default-collapsed** (not default-open as mockup shows). Open state persists per-user in `localStorage`. Inside, gate sub-widgets by role (pending approvals visible only when user has assignments).

### Table columns — Keep / Cut

| Column | Decision | Reason |
|---|---|---|
| Código | ✅ Keep | Domain identity (`document.Code`) |
| Título | ✅ Keep | `document.Name` |
| Área | ✅ Keep | `document.ProcessAreaCodeSnapshot` |
| Perfil | ✅ Keep | `document.ProfileCodeSnapshot` |
| Rev. | ✅ Keep | `document.RevisionVersion` (or version label) |
| Estado | ✅ Keep | `document.Status` (real 8-state) |
| Autor | ⚠ Soften | `document.CreatedBy` is a UUID. Avatar+initials needs user lookup join — defer or hide for v1 |
| Atualizado | ✅ Keep | `document.UpdatedAt` |
| (more menu) | ⚠ Defer | Actions need explicit list — RBAC-gated. Skip in v1 |

---

## Layout (post-audit)

AppShell + SectionPanel (224px, area tree) + main content + collapsible activity sidebar (320px, right, default-collapsed).

Main content top-down:
1. **Header row** — kicker `Documentos · Biblioteca`, h1 `Acervo controlado`, doc count + activity-toggle button.
2. **Lede paragraph** — short description, max-width 580px.
3. **Stat strip** — 0 or 1 card in v1 (see audit). Skip until stats endpoint lands.
4. **Filter tabs** — post-audit set: `Todos / Meus / Rascunhos / Em revisão / Publicados / Rejeitados / Obsoletos`. Counts deferred.
5. **Table card** — 8-column grid (drop Autor for v1): Código (160) / Título (1fr) / Área (100) / Perfil (90) / Rev. (70) / Estado (110) / Atualizado (110) / actions (36).
6. **Pagination footer** — page-size selector `[10, 20, 50]` default 20 + page controls. **Client-side pagination over flat list in v1** (see Backend gaps).

SectionPanel (left, 224px):
- Real area tree from `useQuery(QK.taxonomy.areas())` — RH, Qualidade, Produção, TI & Segurança, Financeiro, etc.
- Selecting an area filters the table (client-side for v1).

Activity sidebar (right, 320px, **default-collapsed**):
- Pending approvals (top 3) — gated by role (visible only if user has approval-instance assignments).
- Audit trail stream — gated by role (admin / compliance only).

## Reused primitives

- `components/ui/Icon` — chevron, history, filter, more, x.
- `components/ui/Avatar` — sm size, only if Autor column kept.
- `components/ui/StatusPill` — extend to 8 states (add `.pill-published`, `.pill-scheduled`, `.pill-superseded`, `.pill-obsolete`).
- `components/ui/CodeChip` — mono code wrapper.
- `features/shell/components/SectionPanel` — host for area tree (route `handle: { sectionPanel: true }`).

## New primitives needed (page-local)

- `LibraryStatStrip` (page-local, only if any stat card survives audit).
- `LibraryFilterTabs` (page-local, single use).
- `LibraryAreaTree` (page-local for v1; promote to `features/taxonomy/` if reused).
- `Pagination` + `PageSizeSelector` (page-local first; promote to `components/ui/` when Approval Inbox lands).

CSS: `LibraryPage.module.css`. No inline styles for theming.

## Data sources

- `useQuery(QK.documents.list())` → `listDocuments()` → `GET /api/v2/documents` (flat array, no pagination, no filter).
- `useQuery(QK.taxonomy.areas())` → area tree for SectionPanel.
- Stats endpoint: **does not exist** — see Backend gaps.

State (page-local):
- `useState<Filter>('todos')` — tab.
- `useState<string|null>(null)` — selected area code (from SectionPanel).
- `useState<boolean>(false)` — `activityOpen` (default-collapsed; persist to `localStorage` key `metaldocs.library.activityOpen`).
- `useState<number>(1)` — `page`.
- `useState<10|20|50>(20)` — `pageSize`, persisted to `localStorage` key `metaldocs.library.pageSize`.

## Pagination strategy (v1)

Backend `GET /documents` returns full list, no paging. Decision:

- **v1 — client-side pagination**: fetch full list, slice in memory. Page-size selector `[10, 20, 50]` works locally. Acceptable up to ~10k docs.
- **v2 — server-side pagination**: when scale demands, add `?page=&pageSize=` query params + `X-Total-Count` header (or paginated envelope). Server caps `pageSize` at 50.

NOTES: `QK.documents.list()` keeps stable cache key in v1. When server pagination lands, change to `QK.documents.list({ page, pageSize, area, filter })` and split cache.

## Backend gaps (track in follow-ups)

1. **Stats endpoint missing.** Need `GET /documents/stats` returning per-state counts + per-area counts. Without it, all stat cards + filter-tab counts are decorative or client-derived from the page (misleading).
2. **No pagination on `GET /documents`.** Add `?page=&pageSize=` query params + total-count metadata. Server-side `pageSize` cap = 50.
3. **No filter params.** Add `?status=&area=&author=&q=` for server-side filtering (mirrors filter tabs + SectionPanel selection).
4. **No "Aprovação pendente" view.** Needs join `documents` × `approval_instances WHERE instance_status = 'pending'`. New endpoint or filter param.
5. **Viewer / read-only role missing.** Endpoint requires `system_admin` or `document_filler` to even list documents. If we want a true reader role (e.g. `document_viewer`), add it to capability table + handler check.
6. **`document.CreatedBy` is UUID.** Author column needs user lookup. Defer column or join in repo layer.

Open these as separate plan items before merging real-data wires for Library.

## Open questions for product

1. Should `Em revisão` in stat strip count documents or count approval-instance items? They diverge when one document has multiple staged approvals.
2. "Frozen" intent: is the design surfacing `published` (immutable, current effective) or `superseded` (replaced by newer rev)? Stat card label needs renaming.
3. Should activity sidebar default-open for `approver` role personas, default-closed for everyone else?
4. Is the Library `/documents` route the canonical landing page for **all** roles, or a separate "approver inbox" handles the approver-first persona?
