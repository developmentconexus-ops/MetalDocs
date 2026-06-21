# Feature F0.1 — Plan

> **Milestone:** 0 — Truth reset  ·  **Folder:** `f0.1-tracker-rewrite`
> Engine: inline plan (doc rewrite — `superpowers:writing-plans` not needed for a single-file
> markdown truth-sync). Input: `spec.md`.

## Plan

Single file touched: `wiki/implementation/screen-redesign-tracker.md`. No code, no tests to add.

1. **Build the verified per-screen inventory** (already gathered this session from the router read).
   The routed screens + their verified status:

   | Screen | Route | Component | Status | Milestone |
   |--------|-------|-----------|--------|-----------|
   | Login | `/login` (public) | `auth/pages/LoginPage` | done | done |
   | Dashboard (home) | `/` index | `dashboard/pages/DashboardPage` | partial | M1 (kill MOCK_STATS/MOCK_ACTIVITY) |
   | Operations (dead) | `/` dup index + `/operations` | `operations/pages/OperationsPage` | stub | M0/F0.3 delete |
   | Audit (dead) | `/audit` | `audit/pages/AuditPage` | stub | M0/F0.3 delete |
   | Library | `/documents` | `documents/pages/LibraryPage` | done | done |
   | Documento Publicado | `/documents/:id` index | `documents/pages/DocumentPublishedPage` | partial | M4/F4.1 |
   | Distribuição | `/documents/:id/distribution` | `documents/pages/DocumentDistributionPage` | stub | M2 |
   | Editor | `/documents/:id/edit` | `documents/pages/DocumentEditorRoutePage` | done | done |
   | Novo Documento (wizard) | `/documents/new` | `documents/pages/NewDocumentWizardPage` | done | done |
   | Templates list | `/templates` | `templates/pages/TemplatesListRoutePage` | done | done |
   | Template wizard | `/templates/new` | `templates/pages/TemplateWizardPage` | done | done |
   | Template editor | `/templates/:id/versions/:n` | `templates/pages/TemplateEditorRoutePage` | done | done |
   | Taxonomy Admin | `/admin/taxonomy` | `taxonomy/pages/TaxonomyAdminRoutePage` | partial | M5/F5.2 (inline styles → tokens) |
   | IAM Admin Center | `/admin/*` (+7 tabs) | `iam/pages/AdminCenterPage` | done | done |
   | Approval Inbox | `/approvals` | `approval/pages/InboxPage` | done | done |
   | Route Admin | `/approval-routes` | `approval/pages/route-admin/RouteAdminPage` | done | done |
   | Notifications | `/notifications` | `notifications/pages/NotificationsPage` | stub | M3 |
   | Content Builder | `/content-builder` | `content-builder/pages/ContentBuilderPage` | partial (wrapper) | out-of-scope (HS-6 if real gap) |
   | Change Password | `/change-password` | `password-change/pages/PasswordChangeRoutePage` | done | done |
   | Auth route (unmounted) | — (`authRoutes` not spread in AppRouter) | `auth/pages/AuthRoutePage` | not-started | out-of-scope (record only) |
   | Documento Obsoleto | — (net-new, not routed) | (variant of `DocumentPublishedPage`) | not-started | M4/F4.2 |
   | Detalhe Signoff | — (net-new, not routed) | (none yet) | not-started | M5/F5.1 |
   | alternativas-inicio-caixa | — | (none) | cut | cut (D3) |
   | catalogo-slots | — | (none) | cut | cut (D3) |

2. **Rewrite the tracker:** replace the `## Status` redesign-block table with the per-screen table
   above (+ Notes column for the verified detail). Update header `Last updated → 2026-06-21` and add
   a `Governing program:` pointer to `docs/superpowers/milestones/frontend-screen-completion/mission.md`.
   Replace the redesign-block Legend with the 5-term status vocab legend. Keep Key-Files + Design-System
   reference sections (still accurate).
3. **Self-verify** against the spec Validation Gate (greps + 1:1 reconcile); record in `evidence.md`.

## Files touched
- `wiki/implementation/screen-redesign-tracker.md` (rewrite status section + header + legend)

## Test strategy
- No automated test (markdown). Deterministic cross-check: for each row, `ls`/`grep` the cited
  component file; grep the status vocab; grep the header stamp. Output captured in `evidence.md`.

## Ordering
F0.1 first in M0 — the verified inventory it produces is the reference F0.2/F0.3/F0.4 act on.
