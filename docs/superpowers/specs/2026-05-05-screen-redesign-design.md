# MetalDocs Screen Redesign — Design Spec

**Date:** 2026-05-05
**Branch:** feature/screen-redesign
**Worktree:** `.worktrees/screen-redesign`
**Status:** Approved — ready for implementation planning

---

## 1. Scope

Full visual redesign of all 9 approved screens to match the MetalDocs Wine design system (design source files in `frontend/apps/web/design-source/`). This is not a feature addition — it is a complete UI rebuild with parallel infrastructure cleanup.

Screens in delivery order:
1. Login
2. Library (Document List)
3. Editor
4. Wizard (4-step document creation)
5. Templates
6. Registry (Controlled Documents)
7. Dashboard
8. Approval Inbox
9. Signoff Detail

---

## 2. Decisions

| Topic | Decision | Rationale |
|---|---|---|
| Approach | Foundation-first, screen-by-screen | Each screen verifiable before next starts |
| Shell | Full replace (`WorkspaceRoot` → `AppRoot` + `AppShell`) | Gutting God component; clean separation |
| Legacy topbar | Adapted into `AppToolbar` | Keep existing UX patterns, restyle to design system |
| Token naming | Rename `--vinho*` → `--brand*` everywhere | Single source of truth, matches design system |
| Fonts | Switch to Inter Tight + JetBrains Mono | Matches design exactly |
| Zustand cleanup | Delete server-state stores upfront | Forces clean TanStack Query from day one |
| Zustand kept | `auth.store` (session), `ui.store` (flash messages) | Auth is session state, not server cache |
| Server state | TanStack Query with centralized `QK` query keys | Industry standard, shared cache across screens |
| Editor data | `useDocumentSession` + `useDocumentAutosave` unchanged | Stateful write protocol, not a cacheable query |
| Optimistic updates | Required on all mutations | SaaS-grade feel — UI instant before server confirms |
| CSS primitives | Tier 1 (Icon/Avatar/CodeChip/Logo): global classes. Tier 2+ (StatusPill, pages): CSS Modules | Design system consistency + isolation |
| Login route | Isolated public route outside AppShell | No Rail, no Toolbar — matches design |

---

## 3. Foundation Block

### 3.1 Token Rename

`src/styles/tokens.css` — full rewrite to Wine palette:

```css
:root {
  /* Fonts */
  --font-sans: "Inter Tight", "Inter", system-ui, sans-serif;
  --font-mono: "JetBrains Mono", "IBM Plex Mono", ui-monospace, monospace;

  /* Brand */
  --brand: #6b1f2a;
  --brand-deep: #3e1018;
  --brand-soft: #8b2e3a;
  --brand-pale: #f9f0f0;
  --accent: #c8364a;

  /* Surface */
  --bg: #f4eeee;
  --surface: #ffffff;
  --surface-2: #faf6f6;
  --surface-3: #f0e9e9;

  /* Border */
  --border: #e6dcdc;
  --border-strong: #d4c2c2;

  /* Text */
  --text: #1a0e0e;
  --text-soft: #4a3434;
  --text-muted: #8a7575;
  --text-faint: #b3a0a0;

  /* Rail */
  --rail: #2a1418;
  --rail-text: #e8d6d6;
  --rail-text-muted: #9c7e7e;
  --rail-active: #6b1f2a;
  --rail-divider: #3e2025;

  /* Semantic */
  --success: #1a6b35;
  --success-bg: #e6f5ec;
  --warning: #b07016;
  --warning-bg: #fbf2dc;
  --danger: #c8364a;
  --danger-bg: #fae8eb;
  --info: #1a3a7a;
  --info-bg: #e8eef8;

  /* Shadow */
  --shadow-1: 0 1px 2px rgba(74, 33, 33, 0.06);
  --shadow-2: 0 8px 24px rgba(74, 33, 33, 0.08);

  /* Radii */
  --r-1: 4px; --r-2: 6px; --r-3: 8px; --r-4: 12px; --r-pill: 999px;

  /* Spacing */
  --sp-1: 4px; --sp-2: 8px; --sp-3: 12px; --sp-4: 16px;
  --sp-5: 20px; --sp-6: 24px; --sp-7: 32px; --sp-8: 40px; --sp-9: 56px;
}
```

Mass-replace all `var(--vinho*)` in CSS modules → `var(--brand*)`. Mapping:

| Old | New |
|---|---|
| `--vinho` | `--brand` |
| `--vinho-d` | `--brand-deep` |
| `--vinho-l` | `--brand-soft` |
| `--vinho-pale` | `--brand-pale` |
| `--vinho-soft` | `--surface-2` |
| `--vinho-muted` | `--border-strong` |
| `--border-2` | `--border-strong` |
| `--muted` | `--text-muted` |
| `--muted-soft` | `--text-faint` |

### 3.2 Font Switch

`index.html` — add Google Fonts import:
```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Inter+Tight:wght@400;500;600&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
```

`tokens.css` — `font-family` declaration removed from `:root` (set via `base.css` on `body`).

### 3.3 Zustand Server-State Cleanup

**Delete** the following Zustand slices (and their store setters/getters):
- `documents.store`: `documents`, `selectedDocument`, `versions`, `versionDiff`, `approvals`, `attachments`, `collaborationPresence`, `documentEditLock`, `policies`, `auditEvents`, `loadState`, `contentMode/File/PdfUrl/DocxUrl/Status/Error`
- `registry.store`: `documentProfiles`, `processAreas`, `documentDepartments`, `subjects`, `selectedProfileSchema/Schemas/Governance`
- `notifications.store`: `notifications`

**Keep:**
- `auth.store`: `user`, `authState`, `loginForm`, `passwordForm` + all setters
- `ui.store`: `error`, `message`, `managedUsers` (flash messages + admin user list)

**Cascade:** `WorkspaceRoot`, `WorkspaceShell`, `AuthShell`, `DocumentWorkspaceShell` — delete all. `useAuthSession` — keep but strip all Zustand store calls for deleted slices.

### 3.4 UI Primitives

**Tier 1 — global design system classes (no CSS Module):**

`src/components/ui/Icon.tsx`
- Props: `name: string`, `size?: number` (default 16), `className?: string`
- Renders inline SVG from a path map (30+ icons from `design-source/shell.jsx`)
- 20×20 viewBox, stroke 1.5, `strokeLinecap="round"`, `strokeLinejoin="round"`

`src/components/ui/Avatar.tsx`
- Props: `name: string`, `size?: 'sm' | 'md' | 'lg'`
- Renders `.avatar .avatar-{size}` with 2-letter initials derived from name

`src/components/ui/CodeChip.tsx`
- Props: `children: ReactNode`, `className?: string`
- Renders `<span className="code-chip mono">`

`src/components/ui/Logo.tsx`
- Renders logomark + "MetalDocs" wordmark from design
- Props: `size?: 'sm' | 'md'`

**Tier 2 — CSS Module:**

`src/components/ui/StatusPill.tsx`
- Props: `status: DocumentStatus`
- Maps status → pill modifier class (`pill-draft`, `pill-review`, `pill-approved`, `pill-frozen`, `pill-rejected`, `pill-archived`)
- Own CSS Module for any layout overrides

---

## 4. AppShell Architecture

### 4.1 Router Structure

```tsx
// src/app/AppRouter.tsx
createBrowserRouter([
  {
    path: "/login",
    lazy: () => import("../features/auth/pages/LoginPage"),
    // Full-page, no Rail, no Toolbar
  },
  {
    element: <AppRoot />,
    children: [
      {
        element: <AppShell />,
        children: [
          ...documentsRoutes,
          ...approvalRoutes,
          ...templatesRoutes,
          ...registryRoutes,
          ...taxonomyRoutes,
          ...iamRoutes,
          ...auditRoutes,
          ...notificationsRoutes,
          ...operationsRoutes,
          { path: "*", element: <Navigate to="/" replace /> },
        ],
      },
    ],
  },
])
```

### 4.2 AppRoot

`src/features/shell/pages/AppRoot.tsx`
- Calls `me()` on mount (plain async in `useEffect`, result → `auth.store`)
- `authState === 'idle'` → `<Navigate to="/login" />`
- `authState === 'loading'` → full-screen spinner
- `authState === 'ready'` + `mustChangePassword` → password change overlay
- `authState === 'ready'` → `<Outlet />`
- On `authBus` 401 event → reset `auth.store` → triggers redirect

### 4.3 AppShell

`src/features/shell/components/AppShell.tsx`
- Layout: `display: flex; height: 100vh; overflow: hidden`
- `<Rail />` (56px, fixed width, `var(--rail)` background)
- `<div className={styles.main}>` (flex-1, flex-column)
  - `<AppToolbar />`
  - `<SectionPanel />` (conditional, Library only via route `handle.sectionPanel`)
  - `<main>` (flex-1, overflow-auto) → `<Outlet />`

### 4.4 Rail

`src/features/shell/components/Rail.tsx`
- Width: 56px, background: `var(--rail)`, full height
- Top: `<Logo />` (icon only, 32px)
- Middle: nav items — Home (`/`), Documents (`/documents`), Templates (`/templates-v2`), Registry (`/registry-v2`), Approvals (`/approvals`), Audit (`/audit`)
- Bottom: user `<Avatar />` + logout icon button
- Active: `useMatch` per path → `var(--rail-active)` pill highlight
- Tooltips: each nav icon shows label on hover (title attribute, CSS tooltip)

### 4.5 AppToolbar (adapted legacy topbar)

`src/features/shell/components/AppToolbar.tsx`
- Height: 52px, `var(--surface)`, border-bottom `var(--border)`
- Keeps from legacy: search input, "Novo documento" CTA, notifications bell, user display name
- Drops: old Zustand-driven search handler, hardcoded colors, DM Sans
- Search: local state + `useNavigate('/documents?q=...')` on submit
- "Novo documento": navigates to `/documents-v2/new`
- Notifications bell: `useQuery(QK.notifications.unreadCount())`
- Uses `var(--brand)` for CTA button, Inter Tight font via CSS inheritance
- CSS Module

### 4.6 SectionPanel

`src/features/shell/components/SectionPanel.tsx`
- Width: 224px, `var(--surface-2)`, border-right `var(--border)`
- Renders only when parent route has `handle: { sectionPanel: true }`
- Contains area/profile filter tree (Library-specific content passed via children or context)

---

## 5. Query Keys

`src/lib/queryKeys.ts` — centralized constants:

```ts
export const QK = {
  documents: {
    list: () => ['documents', 'list'] as const,
    detail: (id: string) => ['documents', 'detail', id] as const,
  },
  inbox: (params?: { page?: number; areaFilter?: string; onlyOverdue?: boolean }) =>
    ['approval', 'inbox', params ?? {}] as const,
  audit: {
    // Maps to GET /audit/events — confirmed in src/lib/api-types/index.d.ts
    recent: (limit?: number) => ['audit', 'recent', limit ?? 10] as const,
  },
  controlledDocuments: {
    list: (filter?: object) => ['controlled-documents', 'list', filter ?? {}] as const,
    detail: (id: string) => ['controlled-documents', 'detail', id] as const,
  },
  taxonomy: {
    profiles: () => ['taxonomy', 'profiles'] as const,
    areas: () => ['taxonomy', 'areas'] as const,
  },
  templates: {
    list: () => ['templates', 'list'] as const,
  },
  approval: {
    instance: (documentId: string) => ['approval', 'instance', documentId] as const,
  },
  notifications: {
    unreadCount: () => ['notifications', 'unread-count'] as const,
  },
} as const;
```

---

## 6. Screen Specifications

### 6.1 Login `/login`

**Layout:** full-page, no AppShell
- Left 45%: dark panel `var(--rail)` — Logo + tagline "Documentos controlados para indústrias sérias."
- Right 55%: white — MetalDocs logomark, email/username input, password input, submit button, error message

**Data:** none on load
**Auth:** `useAuthSession()` → `handleLogin()` → on success `navigate('/')`
**Design source:** `screens-2.jsx` → `Login` component

### 6.2 Library `/documents`

**Layout:** AppShell + SectionPanel (224px) + collapsible activity sidebar (320px, right)
**Design source:** `selected-library.jsx` → `SelectedLibrary`

**Components:**
- Header: title "Acervo controlado" + doc count + fanout timestamp + toggle activity button
- Stat strip: 4 cards (Em revisão / Aprovação pendente / Frozen este mês / Próx. revisão) — derived from document list
- Filter tabs: Todos / Meus / Em revisão / Aprovação pendente / Frozen / Rascunhos — local `useState` filter
- Table: 9-column grid — Código / Título / Área / Perfil / Rev. / Estado / Autor / Atualizado / actions
- Activity sidebar: pending approvals (top 3) + audit trail stream

**Queries:**
- `useQuery(QK.documents.list())` → `listDocuments()`
- `useQuery(QK.inbox({ limit: 3 }))` → pending approvals in sidebar
- `useQuery(QK.audit.recent())` → audit trail in sidebar

### 6.3 Editor `/documents-v2/:documentId`

**Layout:** AppShell, no SectionPanel
**Design source:** `selected-editor-v2.jsx` → `SelectedEditorV2`

**Components:**
- Ultra-slim doc bar (full width): back chevron + code chip + title (editable) + version + save indicator + "Revisões" button + "Salvar" + "Submeter para revisão"
- Floating mini-toolbar: paragraph style select + B/I/U + list/link + word count
- Paper canvas: 760px centered white card, Georgia serif, `contentEditable` sections, document header with code + version + "EM EDIÇÃO"
- Right metadata sidebar (300px): metadata table + revision timeline + next approvers list

**Data:**
- `useQuery(QK.documents.detail(id))` → `getDocument(id)` for metadata
- `useDocumentSession(id)` + `useDocumentAutosave` → content (unchanged from current)
- `useQuery(QK.approval.instance(id))` → next approvers
- `useMutation(submit)` + `invalidateQueries(QK.documents.detail(id))`

### 6.4 Wizard `/documents-v2/new`

**Layout:** AppShell, no SectionPanel
**Design source:** `selected-wizard-v2.jsx` + `screens-2.jsx` → `WizardStep2V2`, `CreateWizard`

**State:** local `useState` form object `{ profileCode, areaCode, title, visibility, invitees, templateId }` + `activeStep: 0-3`

**Steps:**
- Step 1 (Perfil): profile card grid — `useQuery(QK.taxonomy.profiles())`
- Step 2 (Área & Código): area select + title + code preview + visibility 2×2 selector + conditional sub-controls — `useQuery(QK.taxonomy.areas())`
- Step 3 (Template): 3-col template card grid — `useQuery(QK.templates.list())`
- Step 4 (Confirmação): summary of all selections

**Submit:** `useMutation(createDocument)` → on success `navigate('/documents-v2/:newId')` + `invalidateQueries(QK.documents.list())`

### 6.5 Templates `/templates-v2`

**Layout:** AppShell, no SectionPanel
**Design source:** `screens-2.jsx` → `Templates`

- 3-column card grid, mini doc preview per card
- `useQuery(QK.templates.list())` → `listTemplates()` from `features/templates/api/templatesV2.ts`
- "Novo template" CTA → existing `TemplateCreateDialog` at `features/templates/TemplateCreateDialog.tsx` (kept as-is)

### 6.6 Registry `/registry-v2`

**Layout:** AppShell, no SectionPanel
**Design source:** `screens-2.jsx` → `Registry`

- 6-column sequence counter grid: profile code + next seq + dotline revision dots
- Full controlled documents table below
- `useQuery(QK.controlledDocuments.list())` → `fetchControlledDocuments()`
- `useMutation(createControlledDocument)` + `invalidateQueries` — optimistic update
- `useMutation(obsoleteControlledDocument)` — optimistic update

### 6.7 Dashboard `/`

**Layout:** AppShell, no SectionPanel
**Design source:** `priority-screens.jsx` → `Dashboard`

- Hero: "Bom dia, {user.displayName}" + date
- Stat row (4 cards) — derived from documents query (no extra query)
- Left col: pending approvals preview — `useQuery(QK.inbox({ limit: 3 }))` (shared cache with Library sidebar)
- Right col: shortcuts grid + my drafts — `useQuery(QK.documents.list())` filtered client-side (shared cache with Library)

### 6.8 Approval Inbox `/approvals`

**Layout:** AppShell, no SectionPanel
**Design source:** `priority-screens.jsx` → `ApprovalInbox`

- Filter strip: area dropdown + "Apenas vencidos" toggle — local state
- Multi-select + bulk action bar (appears on selection count > 0)
- Dense table: code + title + submitter + area + submitted at + urgency pill + action button
- `useQuery(QK.inbox({ page, areaFilter, onlyOverdue }))` → `listInbox()`
- Pagination: prev/next, page in query key
- `useMutation(signoff)` → optimistic update on instance + `invalidateQueries(QK.inbox(...))`

### 6.9 Signoff Detail `/approvals/:documentId`

**Layout:** AppShell, no SectionPanel
**Design source:** `priority-screens.jsx` → `SignoffDetail`

- Left: A4 paper with diff highlights on changed sections
- Right panel (340px): approval flow visualization + decision radio (Aprovar / Rejeitar / Solicitar revisão) + comment textarea + signature block + submit button
- `useQuery(QK.approval.instance(documentId))` → `getInstance()`
- `useMutation(signoff)` → optimistic cache update on instance + `invalidateQueries(QK.inbox(...))`

---

## 7. File Structure Changes

### New files
```
src/
  lib/
    queryKeys.ts                          ← centralized QK constants
  components/ui/
    Icon.tsx                              ← SVG icon sprite
    Avatar.tsx                            ← initials avatar
    CodeChip.tsx                          ← code chip wrapper
    Logo.tsx                              ← logomark + wordmark
    StatusPill.tsx + StatusPill.module.css
  features/
    shell/
      components/
        AppShell.tsx + AppShell.module.css
        Rail.tsx + Rail.module.css
        AppToolbar.tsx + AppToolbar.module.css
        SectionPanel.tsx + SectionPanel.module.css
      pages/
        AppRoot.tsx
    auth/
      pages/
        LoginPage.tsx + LoginPage.module.css
    documents/
      pages/
        LibraryPage.tsx + LibraryPage.module.css
        EditorPage.tsx + EditorPage.module.css (replaces DocumentEditorPage)
        WizardPage.tsx + WizardPage.module.css (replaces DocumentCreatePage)
    templates/
      pages/
        TemplatesPage.tsx + TemplatesPage.module.css
    registry/
      pages/
        RegistryPage.tsx + RegistryPage.module.css
    operations/
      pages/
        DashboardPage.tsx + DashboardPage.module.css
    approval/
      pages/
        InboxPage.tsx (rebuild, CSS Module)
        SignoffDetailPage.tsx + SignoffDetailPage.module.css
```

### Deleted
```
src/features/shell/pages/WorkspaceRoot.tsx
src/features/shell/WorkspaceShell.tsx
src/components/AuthShell.tsx
src/components/DocumentWorkspaceShell.tsx
src/components/WorkspaceViewFrame.tsx
src/components/WorkspaceDataState.tsx
src/components/WorkspacePlaceholder.tsx
src/components/AppShellHeader.tsx
src/features/documents/state/documents.store.ts (server-state slices)
src/features/registry/state/registry.store.ts
src/features/notifications/state/notifications.store.ts
```

---

## 8. Constraints & Rules

- All new screen CSS: CSS Modules only. Zero inline styles for theming. Zero hardcoded color hex values.
- All new screens: TanStack Query for server state. No `useState` + `useEffect` for data fetching.
- Editor content session: `useDocumentSession` + `useDocumentAutosave` unchanged.
- `design-source/<slug>/NOTES.md` written per screen at implementation time.
- `auth.store` and `ui.store` are the only Zustand stores that remain post-cleanup.
- Every mutation with user-visible side effects implements optimistic update + rollback.
- Query keys must come from `src/lib/queryKeys.ts` — no inline string arrays.
