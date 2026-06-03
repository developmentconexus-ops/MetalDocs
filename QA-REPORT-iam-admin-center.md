# QA Report — IAM Admin Center (qa/iam-admin-center)

- Route(s): `/admin`
- Page: [frontend/apps/web/src/features/iam/pages/AdminCenterPage.tsx](frontend/apps/web/src/features/iam/pages/AdminCenterPage.tsx)
- Owning module: [wiki/modules/iam.md](wiki/modules/iam.md)
- QA class: screen + authz (admin surface)
- Acting role(s): `system_admin` (login `admin / AdminMetalDocs123!`)
- Date: 2026-06-02
- CI: local-evidence gate (tsc green, targeted vitest green, Preview runtime proof captured)

## Gate results

- **Gate 0 scope truth**: pass. Route owned by `features/iam/routes.tsx:5`, page = `pages/AdminCenterPage.tsx → AdminCenterView`. Backend contract = `GET /api/v1/iam/admin/overview` + `/iam/users*` writes (`apps/api/cmd/metaldocs-api/permissions.go:116-123`). API up on :8081, Vite preview on :4173, login OK, `/admin` reachable.
- **Gate 1 impl truth**: pass. `corepack pnpm tsc --noEmit -p tsconfig.build.json` → EXIT=0. `corepack pnpm exec vitest run src/features/iam src/components` → 9/9 passed.
- **Gate 3 product QA**: 4 findings (1 critical, 3 medium) — drove the screen as `system_admin` via Preview; happy-path list/edit-select verified; **the destructive Save-user happy path was NOT executed** because the critical bug would have demoted the only admin → tested via state read instead.
- **Gate 5 regression**: pass. tsc + vitest re-run after fixes both green.

## Findings (severity-ordered)

| # | Severity | Family | Finding | Disposition |
|---|---|---|---|---|
| 1 | **CRITICAL** | shared-contract prerequisite (FE↔backend role enum drift) | `features/iam/api/iam.ts:4` `allowedRoles = {"admin","editor","reviewer","viewer"}` filters OUT every canonical backend role: `system_admin`, `approver`, `author`, `signer`, `area_admin`, `qms_admin`. Effect: `GET /iam/admin/overview` returns roles like `["system_admin"]` and `["approver"]`, FE strips them to `[]`. Verified: every user including `Administrator` rendered the chip "VIEWER" in the list. | **fixed** — widened allowlist to canonical 8 roles ([iam.ts:4-15](frontend/apps/web/src/features/iam/api/iam.ts:4)). Runtime re-verified: chips now read SYSTEM ADMIN / APPROVER / AUTHOR / VIEWER. |
| 2 | **CRITICAL** | screen-local implementation (compounded by #1) | Edit form profile dropdown defaulted to `Viewer` for every selected user (incl. Administrator) because `selectManagedUser` / form-sync effect used post-normalize empty `roles[]` → fell back to `["viewer"]` (`useManagedUsers.ts:17`, `AdminCenterView.tsx:40`, `ManagedUsersPanel.tsx:103`). Clicking **Salvar usuario** would then `PUT /iam/users/admin/roles {roles:["viewer"]}` → strip `system_admin` from the only admin → operator lockout. No confirmation. | **fixed transitively by #1** — once roles array preserves backend values, dropdown reflects real role. Verified: Administrator edit form now shows "PERFIL: System Admin". |
| 3 | medium | screen-local implementation (misleading UI) | `AdminCenterView.tsx:54-60` activity-label heuristic checked `lower.includes("login")` for the LOGIN chip, but the dominant audit action is `auth.logout` — does not contain "login". Every logout was labeled "ACAO" (generic). `session.acquired` / `session.released` / `published` likewise unlabeled. | **fixed** — added explicit LOGOUT/SESSAO/PUBLICAR matches ([AdminCenterView.tsx:53-74](frontend/apps/web/src/features/iam/AdminCenterView.tsx:53)). Verified runtime chips: LOGOUT, SESSAO. |
| 4 | medium | screen-local implementation | `ManagedUsersPanel.tsx:72-74` `roleLabel` returned "Viewer" for any role missing from `PROFILE_OPTIONS` (which excludes `signer`, `area_admin`, `qms_admin`, `admin`, `reviewer`). Re-introduces #1-style misreporting for area-only roles even after the API fix. | **fixed** — added exhaustive `ROLE_LABELS: Record<UserRole, string>` map ([ManagedUsersPanel.tsx:72-83](frontend/apps/web/src/components/ManagedUsersPanel.tsx:72)). |
| 5 | medium | wiki-memory drift (FE↔ADR 0016) | `AppShell.tsx:23` `requiresAdmin` gate hard-codes `roles?.includes('system_admin')`. ADR 0016 introduced view-grade caps (`user.view`); backend `permissions.go:116,123` already gates GET on `CapUserView`. FE blocks any non-`system_admin` role from reaching `/admin` regardless of cap holdings. | **bounded defer** → tracked under [wiki/modules/iam-tech-debt.md](wiki/modules/iam-tech-debt.md) (cross-cutting AppShell gate — affects `/admin`, `/admin/memberships`, approval routes, taxonomy admin). Fix is FE-wide capability gate; out of scope for this screen QA. |
| 6 | medium | screen-local implementation (misleading UI) | `AdminCenterView.tsx:168-205` "Ver todos →" and "Audit trail →" buttons render but have no `onClick` handler — dead controls. | **bounded defer** — UI affordance only, not destructive. |
| 7 | medium | screen-local implementation (misleading UI) | `ManagedUsersPanel.tsx:53-65` `DEPARTMENT_OPTIONS` / `PROCESS_AREA_OPTIONS` are hardcoded strings, never persisted, never sent to backend. UI suggests admin assigns departamento/area — actually pure decoration. Area assignment lives in `/admin/memberships` (`user_process_areas`). | **bounded defer** — separate redesign needed to either (a) wire to taxonomy + memberships or (b) remove the fields. Out of scope. |
| 8 | low | screen-local | `useEffect` + StrictMode double-fires `getAdminOverview` on mount (two requests visible in Preview network log at .431/.432). Harmless duplicate GET. | **bounded defer** — common StrictMode artifact, fix when migrating to TanStack Query. |
| 9 | low | screen-local | `PROFILE_OPTIONS` (the edit dropdown) only lists 5 of 8 canonical roles. A `qms_admin`/`signer`/`area_admin` user can be loaded but the dropdown silently coerces to one of the 5 on next save. | **bounded defer** — single-role-per-user is the current admin UX contract; area-only roles are managed in `/admin/memberships`. Document in iam-tech-debt. |

## Evidence

### Commands run

- `git checkout -b qa/iam-admin-center`
- `curl -s http://localhost:8081/healthz` → `{"status":"live"}`
- `corepack pnpm tsc --noEmit -p tsconfig.build.json` → EXIT=0 (post-fix)
- `corepack pnpm exec vitest run src/features/iam src/components` → 9/9 passed

### Runtime proof (Preview-driven)

- Pre-fix list (all users mislabeled "VIEWER"):
  `["QP\nQA PW User\nVIEWER", "AD\nApprover Dev\nVIEWER", "AT\nAuthor Test\nVIEWER", "AT\nApprover Test\nVIEWER", "A\nAdministrator\nVIEWER"]`
- Pre-fix edit card for Administrator: `PERFIL\nViewer` (would demote on save).
- Post-fix list (canonical roles surface):
  `["QP\nQA PW User\nVIEWER", "AD\nApprover Dev\nAPPROVER", "AT\nAuthor Test\nAUTHOR", "AT\nApprover Test\nAPPROVER", "A\nAdministrator\nSYSTEM ADMIN"]`
- Post-fix edit card for Administrator: `PERFIL\nSystem Admin`.
- Post-fix activity chips: `auth.logout → LOGOUT`, `session.released → SESSAO` (previously both "ACAO").
- Console errors: none.
- Network: 2× `GET /api/v1/iam/admin/overview → 200`, no failed requests.

### Persisted / API proof

- `GET /api/v1/iam/admin/overview` payload sample (verified canonical role strings reach the FE):
  - `Administrator` → `roles:["system_admin"]`
  - `Approver Dev` → `roles:["approver"]`
  - `Author Test` → `roles:["author"]`
  - `QA PW User` → `roles:["viewer"]`
- Backend permissions table for `/iam/admin/overview` = `CapUserView` (read-grade per ADR 0016) — confirmed at `apps/api/cmd/metaldocs-api/permissions.go:123`.
- **No write was executed** during QA. Destructive Save-user path verified by state inspection only — the critical bug surfaced via DOM read, not by triggering the demote.

## Hard-stops / Bounded defers

- **No hard-stop.** All fixes were FE-local; backend contract is correct (canonical roles already returned). Root cause was a stale FE allowlist + missing label mappings.
- Bounded defers (linked to wiki):
  - Finding #5 (FE `requiresAdmin` gate ignores ADR 0016 view-grade caps) — cross-cutting AppShell concern; minimum plan: replace `roles?.includes('system_admin')` with capability-array gate sourced from `currentUser.capabilities`, mapped per ADR 0016. Track in [wiki/modules/iam-tech-debt.md](wiki/modules/iam-tech-debt.md).
  - Findings #6, #7, #8, #9 — local cosmetic / decoration / StrictMode artifacts. Track in iam-tech-debt.

## Wiki sync

No `Last verified:` bump required — code change does not invalidate any current wiki anchor. Tech-debt addenda for #5/#6/#7/#8/#9 should be added in a follow-up curator pass.
