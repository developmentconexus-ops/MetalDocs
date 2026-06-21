# Screen Redesign — Implementation Tracker

**Governing program:** `docs/superpowers/milestones/frontend-screen-completion/mission.md` (frontend-screen-completion mission)
**Per-screen Definition of Done (D2 gate):** [`wiki/quality/screen-definition-of-done.md`](../quality/screen-definition-of-done.md) — a screen's `done` status here means it passed that gate.
**Original redesign spec:** `docs/superpowers/specs/2026-05-05-screen-redesign-design.md`
**Design source:** `frontend/apps/web/design-source/`
**Last updated:** 2026-06-21 — full per-screen truth-reset (M0/F0.1) from a verified router + page read. Supersedes the 2026-05-08 redesign-block table.

---

## Status (verified per-screen — 2026-06-21)

Status vocab: `done` · `partial` · `stub` · `not-started` · `cut`. Every **routed** screen in
`src/app/AppRouter.tsx` + `features/**/routes.tsx` has exactly one row; net-new and CUT screens are
listed for completeness. "Milestone" = the `frontend-screen-completion` milestone that completes it
(or `done` / `out-of-scope` / `cut`).

| Screen | Route | Component | Status | Milestone | Notes |
|--------|-------|-----------|--------|-----------|-------|
| Login | `/login` (public) | `auth/pages/LoginPage.tsx` | done | done | Full-page split layout, no Rail. |
| Dashboard (home) | `/` (index) | `dashboard/pages/DashboardPage.tsx` | partial | M1 | Redesign tokens applied; ships `MOCK_STATS` + `MOCK_ACTIVITY` (finding 6) — only approval-inbox query is real. |
| Operations | — (route removed M0/F0.3) | `operations/pages/OperationsPage.tsx` *(deleted)* | cut | cut (M0/F0.3 — deleted) | **Deleted** in M0/F0.3 (D7): was an empty `OperationsCenter` shell, no API, with a duplicate root `index:true` shadowing Dashboard (findings 2–3). F0.2 removed the dup index; F0.3 deleted the page + route. IAM Admin Center owns metrics/audit/sessions. |
| Audit | — (route removed M0/F0.3) | `audit/pages/AuditPage.tsx` *(deleted)* | cut | cut (M0/F0.3 — deleted) | **Deleted** in M0/F0.3 (D7): was an identical empty `OperationsCenter` copy (finding 4). IAM Admin Center owns audit. |
| Library | `/documents` | `documents/pages/LibraryPage.tsx` | done | done | Server-paginated document library; real `/api/v2/documents` + stats. |
| Documento Publicado | `/documents/:id` (index) | `documents/pages/DocumentPublishedPage.tsx` | partial | M4 / F4.1 | Core lifecycle real; 3 TODOs + 6 "em breve" placeholders (PDF download, coverage, related docs, comments…) (finding 12). |
| Distribuição | `/documents/:id/distribution` | `documents/pages/DocumentDistributionPage.tsx` | stub | M2 | Canvas ready; all KPI cards "Dados ilustrativos · Em breve", CTAs disabled; **no fanout backend** (finding 8–9). |
| Editor | `/documents/:id/edit` | `documents/pages/DocumentEditorRoutePage.tsx` | done | done | Ships (corrects 2026-05-08 "Not started"). |
| Novo Documento (wizard) | `/documents/new` | `documents/pages/NewDocumentWizardPage.tsx` | done | done | 4-step wizard; deferred items in `wiki/backlog/novo-documento.md`. |
| Templates list | `/templates` | `templates/pages/TemplatesListRoutePage.tsx` | done | done | Real API; deferred items in `wiki/backlog/templates.md`. |
| Template wizard | `/templates/new` | `templates/pages/TemplateWizardPage.tsx` | done | done | |
| Template editor | `/templates/:id/versions/:n` | `templates/pages/TemplateEditorRoutePage.tsx` | done | done | |
| Taxonomy Admin | `/admin/taxonomy` | `taxonomy/pages/TaxonomyAdminRoutePage.tsx` | partial | M5 / F5.2 | Functional + API-wired (`/api/v1/taxonomy/*`) but inline styles, off redesign tokens (finding 15). |
| IAM Admin Center | `/admin/*` (+ overview/people/roles/memberships/audit/sessions/usage tabs) | `iam/pages/AdminCenterPage.tsx` | done | done | Owns metrics/audit/sessions (D7 — why Operations/Audit stubs are deleted). |
| Approval Inbox | `/approvals` | `approval/pages/InboxPage.tsx` | done | done | Foco/Linha-do-tempo views; deferred items in `wiki/backlog/caixa-aprovacao.md`. |
| Route Admin | `/approval-routes` | `approval/pages/route-admin/RouteAdminPage.tsx` | done | done | Capability-gated (`route.manage`). |
| Notifications | `/notifications` | `notifications/pages/NotificationsPage.tsx` | stub | M3 | Empty stub; `notifications.ts` returns empty arrays, stream noop; **no notification backend** (finding 10–11). |
| Content Builder | `/content-builder` | `content-builder/pages/ContentBuilderPage.tsx` | partial | out-of-scope | Thin wrapper delegating to a shared view; not a target screen. HS-6 if it proves a real gap. |
| Change Password | `/change-password` | `password-change/pages/PasswordChangeRoutePage.tsx` | done | done | |
| Auth route | — (`authRoutes` **not** mounted in `AppRouter.tsx`) | `auth/pages/AuthRoutePage.tsx` | not-started | out-of-scope | Exported but unmounted dead route; `/login` is the live public auth entry. Recorded as truth; not in M0 delete scope (D7 = Operations/Audit only). |
| Documento Obsoleto | — (net-new, not routed) | (planned `obsolete` variant of `DocumentPublishedPage`) | not-started | M4 / F4.2 | `documento-obsoleto` design source exists; no variant logic yet (finding 13). |
| Detalhe Signoff | — (net-new, not routed) | (none yet) | not-started | M5 / F5.1 | `detalhe-signoff` design source exists; approval panels exist but no standalone screen (finding 14). |
| `alternativas-inicio-caixa` | — | (none) | cut | cut (D3) | Design-source slug, no NOTES/route/product intent. CUT. |
| `catalogo-slots` | — | (none) | cut | cut (D3) | Design-source slug, no NOTES/route/product intent. CUT. |

> **Evidence base:** verified router read of `src/app/AppRouter.tsx` + every `features/**/routes.tsx`,
> `ls` of every live-presented page file (the 18 routed pages after M0/F0.3 deleted Operations + Audit —
> the two `cut` rows cite their now-deleted files for lineage only, marked *(deleted)*), mission §5
> inventory, and `discovery-brief.md` findings 1–16, reconciled to post-F0.3 reality on 2026-06-21.

---

## Legend

| Term | Meaning |
|---|---|
| `done` | Production-complete: real data, redesign tokens, shipped. |
| `partial` | Routed + partly real, but has mock data / placeholders / off-token styling to finish. |
| `stub` | Routed but empty / no-API shell (or noop data layer). |
| `not-started` | No implemented page (net-new owed work, or an unmounted dead export). |
| `cut` | Out of scope by operator decision (D3) — never built. |

---

## Key Files Reference

| File | Purpose |
|---|---|
| `src/styles/tokens.css` | Wine palette CSS vars (`--brand`, `--bg`, `--rail`, etc.) |
| `src/lib/queryKeys.ts` | Centralized TanStack Query key constants |
| `src/components/ui/Icon.tsx` | SVG icon component |
| `src/components/ui/Avatar.tsx` | Initials avatar |
| `src/components/ui/StatusPill.tsx` | Document status pill |
| `src/components/ui/CodeChip.tsx` | Code chip wrapper |
| `src/features/shell/pages/AppRoot.tsx` | Auth guard + bootstrap |
| `src/features/shell/components/AppShell.tsx` | Rail + Toolbar + Outlet layout |
| `src/features/shell/components/Rail.tsx` | 56px dark nav sidebar |
| `src/features/shell/components/AppToolbar.tsx` | 52px top bar |
| `src/app/AppRouter.tsx` | Router (public `/login` + protected routes) |
| `src/features/auth/pages/LoginPage.tsx` | Full-page login, no Rail |
| `src/features/documents/pages/LibraryPage.tsx` | Server-side paginated document library at `/documents` |
| `src/features/documents/routes.tsx` | Document routes: Library, detail (Publicado/Distribuição), editor, wizard |

---

## Design System Reference

**Palette:** Wine — `--brand: #6b1f2a`, `--rail: #2a1418`, `--bg: #f4eeee`, `--accent: #c8364a`

**Fonts:** Inter Tight (sans) + JetBrains Mono (mono)

**Key classes:** `.card`, `.btn`, `.btn-primary`, `.btn-ghost`, `.btn-sm`, `.pill`, `.pill-draft/review/approved/frozen/rejected/archived`, `.code-chip`, `.mono`, `.kicker`, `.avatar`, `.avatar-sm`, `.row`, `.divider`, `.spacer`

**Shell dimensions:** Rail 56px · Toolbar 52px · SectionPanel 224px (Library only) · Activity sidebar 320px (Library only)
