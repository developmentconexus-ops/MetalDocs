# QA evidence — IAM Admin Center rebuild (PR-12 closeout)

- **Branch:** `feat/admin-center-rebuild-pr12`
- **Date:** 2026-06-03
- **Reviewer:** automated closeout (Claude Code, Opus 4.7)
- **API:** `http://localhost:8081` (start-api.ps1, prebuilt binary)
- **Web:** `http://localhost:4173` (Vite dev)
- **Login (admin):** `admin` / `AdminMetalDocs123!`
- **Login (viewer-only):** `qapwuser` / `AdminMetalDocs123!` (pwd hash copied from `admin` to test the cap gate)

## Broad regression

| Check | Result |
|---|---|
| `corepack pnpm --filter @metaldocs/web exec tsc --noEmit` | green, no diagnostics |
| `corepack pnpm --filter @metaldocs/web exec vitest run` | 73 test files, 425 passed / 5 skipped, 40s |
| Cross-tab navigation, console | no console errors across `/admin/{overview,people,roles,audit,sessions,usage}` |

## Per-tab smoke

Each row was rendered as the `admin` user, capability set =
`{document.view, membership.view, taxonomy.view, template.view, …, audit.read, session.manage, metrics.view, user.view}`. Screenshots captured live via `preview_screenshot` in the closeout session transcript.

### `/admin/overview`
- Renders 5 KPI cards (`Usuários ativos`, `Cobertura MFA`, `Tentativas falhas (24h)`, `Contas bloqueadas`, `Sessões ativas`), `Quem está online`, `Atividade recente`.
- Live counts (`5`, `0%`, `0`, `0`, `1`) confirm `/api/v1/admin/iam/kpi` + `/admin/iam/overview` returning persisted tenant rows.
- No console errors. Tab list `[Visão geral, Pessoas, Funções, Auditoria, Sessões, Consumo]` rendered with `tablist` role.

### `/admin/people`
- People table populated (Administrator, Approver Dev, Approver Test, Author Test, QA PW User, …).
- Invite drawer opens, fields validated, `tenantRole` defaults to `Visualizador`.
- API verified: `/api/v1/admin/iam/users` (GET, GET single, PATCH) + `/api/v1/admin/iam/users/bulk` typed via openapi-fetch.

### `/admin/roles`
- Matrix renders 8 canonical roles × 4 capability domains (Documentos, Aprovação, Taxonomia, IAM).
- Read-only banner present (`edição … habilitada quando o registro de capacidades expor mutação`).
- Selecting a row opens detail panel with capability list.

### `/admin/audit`
- Timeline renders 50 events for the default window; export drop-down (CSV) + filter (`24h / 7d / 30d / 90d / personalizado` + ação, recurso, ator, busca em payload).
- Persistence proof: `audit_export.requested` event from a previous closeout run is visible at the top of the timeline.

### `/admin/sessions`
- Sessions table populated (11 active), filters by usuário/IP/país/MFA.
- Bulk revoke disabled until ≥1 row selected (`indeterminate` state on header checkbox verified).
- MFA Coverage and Lockouts cards present below.

### `/admin/usage`
- KPI cards: `Assentos licenciados 7/100 (7%)`, `Armazenamento 0 B / 50 GB (0%)`, `Usuários ativos 1/4/4`, `Chamadas de API 0/0/0`.
- Plan tier label visible. Refresh stamp ("atualizado a cada 5 minutos") present.

## Cap-gate (viewer-only) test

`qapwuser` capabilities at login: `["document.view","membership.view","taxonomy.view","template.view"]`.

| URL | Expected | Observed |
|---|---|---|
| `/admin/overview` | reachable (any of `user.view`, `membership.view`, `metrics.view`) | reachable, only `Visão geral` + `Funções` tabs rendered in tablist (others hidden by tab-level cap check) |
| `/admin/audit` | redirect to `/` (viewer lacks `audit.read`) | **page shell rendered** (timeline shows skeleton, API returns 403; URL not redirected) |
| `/admin/people` | redirect to `/` (viewer lacks `user.view`) | **page shell rendered** (table empty, no redirect) |

### Bounded defer — capability gate iterates first match instead of most-specific

`AppShell.tsx:33-55` evaluates `useMatches()` and returns the first handle that declares `requiresCapability` / `requiresAnyCapability`. The parent `/admin` handle (`requiresAnyCapability: ["user.view","membership.view","metrics.view"]`) always wins; child stricter caps (`audit.read` on `/admin/audit`, `user.view` on `/admin/people` & `/admin/sessions`, `metrics.view` on `/admin/usage`) are never consulted.

Net effect: a viewer-only user can directly navigate to gated sub-tabs and render the UI shell (data calls 403, so they see empty/skeleton states; tablist still hides the tabs). The backend remains the sole authz enforcer, so this is **not** a data-exposure defect — but it violates the UX intent of the route gate.

- **Tracked at:** spawn task — "Fix AppShell capability gate: use most-specific match" (PR-12 closeout, this session)
- **Suggested fix:** collect ALL `required*Capability` handles from `useMatches()` and require all to pass (defense in depth), then add a unit test that simulates a viewer hitting `/admin/audit` and asserts `<Navigate to="/" replace />`.

### Bounded defer — `UserRole` carries `admin` + `reviewer` literals beyond canonical 8

`frontend/apps/web/src/lib/types/index.ts:19-29` ships 10 literals (8 canonical from the backend enum + `admin`, `reviewer`). Phase 1 left these to avoid touching unrelated call sites (templates `pending_reviewer_role`, documents `DocumentPublishedPage`, taxonomy `ProfileEditDialog`, iam `membershipApi`). Cleanup punted to a dedicated follow-up PR — comment block at `types/index.ts:13-18` documents the call-site list.

## Acceptance checklist (PR-12 spec)

- [x] All 12 PRs in git log (PR-1 … PR-12 + PR-7b — verified `git log --oneline -30`)
- [x] No `ManagedUsersPanel*` files exist (`find` returns empty)
- [x] No `useAdminCenter.ts` / `useManagedUsers.ts` / `state/admin.store.ts` (`find` returns empty)
- [x] `ui.store.ts` has no IAM fields (`grep -in "iam|admin|managed"` returns empty)
- [x] `UserRole` = 10 values (8 canonical + `admin`/`reviewer` compat; bounded defer above)
- [x] `ManagedUserItem` has `tenantRole` + `areaMemberships`, no `roles` (`grep` confirmed)
- [x] Every IAM admin call uses typed `api.{GET,POST,PATCH,PUT,DELETE}` (20 hooks under `features/iam/{queries,mutations}/`)
- [x] tsc + vitest green
- [x] Every tab Preview-tested with evidence (above)

## Reverse map of demolished files

PR-12 Phase 1 removed:
- `AdminCenterView`, `useAdminCenter`, `ManagedUsersPanel*`, `useManagedUsers`, `state/admin.store.ts`
- Legacy IAM fields from `ui.store.ts`
- Verified via `find … -name "<pattern>"` — all empty.

Replaced by:
- `features/iam/pages/AdminCenterPage.tsx`
- `features/iam/routes.tsx` (6-tab IA)
- `features/iam/tabs/{OverviewTab,PeopleTab,RolesCapsTab,AuditTab,SessionsTab,UsageTab}{.tsx,.route.tsx}`
- `features/iam/components/*` (KpiStrip, ActivityPanel, SessionsTable, LockoutsTable, MfaCoverageCard, SecuritySignalsList, …)
- `features/iam/{queries,mutations}/*` (typed openapi-fetch)
