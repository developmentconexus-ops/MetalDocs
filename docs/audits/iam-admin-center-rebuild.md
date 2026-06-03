# IAM Admin Center — Full Audit (Rebuild Scope)

> Read-only audit. No code changes in this document. Sole purpose: enumerate every legacy artifact, rule violation, and contract gap on the current `/admin` stack with `file:line` anchors and a classification verdict (`DELETE` / `MIGRATE` / `REDESIGN` / `KEEP`). Phase 2 (industry deep-research → target IA) runs in parallel as `docs/audits/observability-admin-panel-research.md`.

- **Screen:** IAM Admin Center
- **Routes owned:** `/admin`, `/admin/memberships`
- **Page:** [frontend/apps/web/src/features/iam/pages/AdminCenterPage.tsx](../../frontend/apps/web/src/features/iam/pages/AdminCenterPage.tsx)
- **View:** [frontend/apps/web/src/features/iam/AdminCenterView.tsx](../../frontend/apps/web/src/features/iam/AdminCenterView.tsx)
- **Owning module:** [wiki/modules/iam.md](../../wiki/modules/iam.md)
- **QA branch:** `qa/iam-admin-center` (commit `2bf7da17` — critical/medium QA patches landed; **rebuild is a separate branch**)
- **Audit date:** 2026-06-02
- **Auditor:** Claude (Opus 4.7) under user-directed full review

## Verdict summary

| Layer | State | Verdict |
|---|---|---|
| Frontend page composition | KPIs + presence + activity + managed-users CRUD on one screen, hand-rolled, no information density discipline | **REDESIGN** |
| Frontend state | zustand `useUiStore` + `useAdminStore` holding server data, manual `useCallback` + `useEffect` fetch, no cache | **MIGRATE → TanStack Query** |
| Frontend types | `UserRole` union includes phantom `"admin"` / `"reviewer"` not in backend | **DELETE phantoms** |
| Frontend API layer | hand-rolled `request<T>` wrapper bypassing `openapi-fetch` typed client | **MIGRATE → generated client** |
| Frontend auth gate | `AppShell.requiresAdmin` hardcodes `system_admin` role literal, ignores ADR 0016 view-grade caps | **REDESIGN → capability gate** |
| Create-user card | hardcoded department/area dropdowns (decorative), inline form on overview surface, single-role array | **DELETE entirely** (replace with dedicated invite flow on drill-in) |
| Edit-user card | same hardcoded dropdowns, fake `departmentFromRole` mapping, single-role dropdown silently destroys multi-role state | **DELETE entirely** (replace with right-rail drill-in panel) |
| Dead UI affordances | "Ver todos →", "Audit trail →" buttons with no `onClick` | **DELETE** |
| Backend HTTP surface | hand-wired `http.ServeMux` + suffix-dispatcher (`handleUserRoute` parses `strings.Split`), no oapi-codegen | **MIGRATE → oapi-codegen contract-first** |
| Backend overview aggregator | `GET /iam/admin/overview` returns 3 unrelated concerns in one envelope (users / online / audit) | **REDESIGN → split into `/iam/users`, `/iam/presence`, `/iam/activity`** |
| Backend error envelope | RFC 9457 via `internal/platform/problem` — already correct | **KEEP** (wiki T-006 is stale — see §6.2) |
| Backend audit emission on role upsert | `handleUserRoleUpsert` emits at [admin_handler.go:369](../../internal/modules/iam/delivery/http/admin_handler.go:369) | **KEEP** (wiki T-005 is stale — see §6.2) |
| Backend role-replace contract | `parseExactlyOneRole` rejects multi-role payloads | **KEEP single-role contract**, but **DELETE the FE multi-role array intent** (dead code) |

---

## 1. Frontend — composition (`AdminCenterView.tsx`)

### 1.1 Single-file mega-view

[AdminCenterView.tsx](../../frontend/apps/web/src/features/iam/AdminCenterView.tsx) (266 lines) hosts:

| Block | Lines | Concern |
|---|---|---|
| Two coupled `useEffect` (load + form sync) | 22-43 | data load + form bind |
| Three KPI cards inline | 120-163 | Presence / Last-activity / Total |
| Online-users panel | 166-200 | Presence stream |
| Recent-activity panel | 202-241 | Audit stream |
| `<ManagedUsersSection>` | 243-259 | CRUD on users |
| Local label/variant heuristics | 53-93 | string-matching on `action` |

**Violations**

- One screen, four unrelated surfaces — fails container/presentational split ([rules/react/patterns.md](../../../../.claude/rules/react/patterns.md)).
- Effect at [AdminCenterView.tsx:26](../../frontend/apps/web/src/features/iam/AdminCenterView.tsx:26) syncs form from list — derived state in `useEffect`, banned by [rules/react/hooks.md §useEffect — When NOT to Use](../../../../.claude/rules/react/hooks.md).
- Activity label/variant/dot/chip heuristics (53-93) are `lower.includes("…")` string sniffing on audit action codes. Belongs in domain/`audit-event-presenter.ts`, keyed off the canonical action enum, not substring guessing.

**Verdict:** REDESIGN into `<AdminOverviewPage>` orchestrator + child feature components:
- `<KpiStrip />`
- `<PresencePanel />`
- `<ActivityPanel />`
- `<UsersDirectory />` (list)
- `<UserDrawer />` (right-rail drill-in — replaces create/edit cards)

### 1.2 Dead controls

| Element | Anchor | Issue |
|---|---|---|
| "Ver todos →" | [AdminCenterView.tsx:174-176](../../frontend/apps/web/src/features/iam/AdminCenterView.tsx:174) | no `onClick`, no `to=`, no handler |
| "Audit trail →" | [AdminCenterView.tsx:209-211](../../frontend/apps/web/src/features/iam/AdminCenterView.tsx:209) | same |

**Verdict:** DELETE both. Replace with router links to `/admin/users` and `/admin/audit` (Phase 2 IA will name the real destinations).

### 1.3 KPI cards

[AdminCenterView.tsx:120-163](../../frontend/apps/web/src/features/iam/AdminCenterView.tsx:120) — three identical card layouts inlined with raw inline SVG icons. Same shape repeated 3×. Violates DRY in [common/coding-style.md](../../../../.claude/rules/common/coding-style.md).

KPIs themselves are weak signals for an observability admin panel:
- "Usuários online agora" — counts a 10-minute window (`activeSince := now.Add(-10 * time.Minute)` at [admin_handler.go:118](../../internal/modules/iam/delivery/http/admin_handler.go:118)) — fine but undocumented.
- "Última atividade" — single timestamp; gives no signal.
- "Total de usuários" — vanity metric, not actionable.

**Verdict:** REDESIGN. Per Phase 2 research; expect: failed-login-attempts trend, locked accounts count, MFA coverage %, dormant users count, role distribution, audit-events-per-minute spark.

---

## 2. Frontend — state management

### 2.1 Server state inside zustand (banned by canonical FE arch)

Canonical rule ([wiki/architecture/frontend-structure.md](../../wiki/architecture/frontend-structure.md)): server state lives in TanStack Query; client/UI state in zustand. Current code violates this in **two** places.

| Store | Holds | Should be |
|---|---|---|
| [`useAdminStore`](../../frontend/apps/web/src/features/iam/state/admin.store.ts) | `users`, `onlineUsers`, `recentActivities`, `loadState` | TanStack Query cache (`['iam','admin','overview']`) |
| [`useUiStore`](../../frontend/apps/web/src/store/ui.store.ts) lines 33,72 | `managedUsers` (DUPLICATE of `useAdminStore.users`) + `userForm` + `managedUserForm` | server state goes to TQ; form state goes to **local component state** or React Hook Form; `userForm`/`managedUserForm` should not live in a global store at all |

[useAdminCenter.ts:18,26](../../frontend/apps/web/src/features/iam/useAdminCenter.ts:18) — writes the same `overview.users` into BOTH stores. Two sources of truth for one list. Any mutation must remember to sync both.

**Verdict:** MIGRATE → TanStack Query for `overview`, `users`, `presence`, `activity`. DELETE `useAdminStore`. DELETE the `userForm` / `managedUserForm` / `managedUsers` fields from `useUiStore` — push form state into the dedicated form components.

### 2.2 Manual fetch lifecycle in `useAdminCenter`

[useAdminCenter.ts:20-32](../../frontend/apps/web/src/features/iam/useAdminCenter.ts:20) — `useCallback` wrapping `setLoadState("loading") → try/catch → setLoadState("ready"|"error")`. Re-implements what TanStack Query does for free, badly:

- no caching
- no dedup (StrictMode double-fires the GET — observed in QA finding #8)
- no background refetch
- no `staleTime`/`gcTime` policy
- no `isFetching` vs `isLoading` distinction
- no retry policy
- error → string message via `asMessage(err)` — drops the RFC 9457 `code` / `errors[]`

**Verdict:** MIGRATE. Per [.agents/skills/metaldocs-tanstack-query/SKILL.md](../../.agents/skills/metaldocs-tanstack-query/SKILL.md).

### 2.3 Mutations in `useManagedUsers` — non-atomic role replace

[useManagedUsers.ts:57-84](../../frontend/apps/web/src/features/iam/useManagedUsers.ts:57) `handleSaveManagedUser`:

```
await api.updateUser(id, {displayName,email,isActive,mustChangePassword});  // call 1
await api.replaceUserRoles(id, {displayName, roles});                       // call 2
await onRefresh();
```

If call 1 succeeds and call 2 fails → user metadata updated, role NOT replaced. No rollback. No idempotency key. No optimistic update with rollback. No invalidation strategy (full refetch of `overview` regardless of which slice changed).

**Verdict:** REDESIGN — backend should expose one transactional endpoint `PATCH /iam/users/{id}` carrying both metadata and roles, OR client uses `mutationFn` with explicit retry/rollback contract. Either way the FE pattern moves to `useMutation` with proper cache-key invalidation.

### 2.4 Selectable user — coerced default

[useManagedUsers.ts:17](../../frontend/apps/web/src/features/iam/useManagedUsers.ts:17): `roles: Array.isArray(item.roles) && item.roles.length > 0 ? item.roles : ["viewer"]`

If a user really has zero roles (legitimate state per backend role-provider's empty-slice return), the form coerces to `["viewer"]` and a Save would assign `viewer`. Hidden side effect, same family as QA finding #2 (destructive demote-by-default).

**Verdict:** REDESIGN — never default roles; if empty, the form shows "No role assigned" and disables Save until the operator picks one.

---

## 3. Frontend — types (`lib/types/index.ts`)

### 3.1 `UserRole` phantom values

[lib/types/index.ts:14-24](../../frontend/apps/web/src/lib/types/index.ts:14) declares:

```ts
export type UserRole =
  | "admin"        // ← PHANTOM. Not in backend iamdomain.Role.
  | "system_admin"
  | "approver"
  | "author"
  | "editor"
  | "reviewer"     // ← PHANTOM. Not in backend iamdomain.Role.
  | "viewer"
  | "signer"
  | "area_admin"
  | "qms_admin";
```

Backend canonical set per [internal/modules/iam/domain/model.go:10-16](../../internal/modules/iam/domain/model.go:10) + [wiki/modules/iam.md §"Last verified"](../../wiki/modules/iam.md): 8 roles (`system_admin, approver, author, editor, viewer, signer, area_admin, qms_admin`). `admin` and `reviewer` are dead intent from an earlier draft.

Knock-on: [ManagedUsersPanel.tsx:72-83](../../frontend/apps/web/src/components/ManagedUsersPanel.tsx:72) `ROLE_LABELS` declares entries for `admin` and `reviewer` to satisfy `Record<UserRole, string>`. Two dead rows.

**Verdict:** DELETE `"admin"` and `"reviewer"` from `UserRole`. Delete the two dead rows from `ROLE_LABELS`. Re-run tsc — any remaining usage proves another phantom call site.

### 3.2 `ManagedUserItem.roles: UserRole[]` vs backend single-role contract

Backend [admin_handler.go:382-407](../../internal/modules/iam/delivery/http/admin_handler.go:382) `handleReplaceUserRoles` calls `parseExactlyOneRole(req.Roles)` — backend **rejects multi-role payloads**. Same for `handleCreateUser` at [admin_handler.go:274](../../internal/modules/iam/delivery/http/admin_handler.go:274). The current admin contract is one tenant role per user.

But FE types model `roles: UserRole[]` and `useManagedUsers.toggleManagedUserRole` (lines 22-31) supports multi-select. Dead intent.

**Verdict:** REDESIGN at the type boundary — either:
- **(A)** keep single-role admin contract and change FE to `role: UserRole` (drop the array everywhere), OR
- **(B)** expand backend to multi-role (`parseRoles` already exists at [admin_handler.go:493-510](../../internal/modules/iam/delivery/http/admin_handler.go:493)) and remove `parseExactlyOneRole`.

Phase 2 research will pick. Industry default for QMS-style products is **multi-role** because area-only roles (`signer`, `area_admin`, `qms_admin`) commonly compose with a tenant role.

---

## 4. Frontend — API client (`features/iam/api/iam.ts`)

### 4.1 Bypasses generated client

[features/iam/api/iam.ts:2](../../frontend/apps/web/src/features/iam/api/iam.ts:2) imports `request<T>` from [lib/api/client.ts:89](../../frontend/apps/web/src/lib/api/client.ts:89) — the hand-rolled wrapper. Project already wires an `openapi-fetch` typed client at [lib/api/client.ts:109](../../frontend/apps/web/src/lib/api/client.ts:109) (`export const api = createClient<paths>({ fetch: apiFetch })`) — but IAM admin endpoints are not in `openapi.yaml` so codegen does not produce them.

Effect: every IAM call is typed by hand with `Record<string, unknown>` bodies (see `createUser`, `updateUser`, `assignRole`, `replaceUserRoles`, `adminResetPassword` at lines 79-113). No compile-time guarantee FE payload matches backend expectation. Drift is invisible until runtime.

**Verdict:** MIGRATE. Backend must publish OpenAPI paths for the full IAM admin surface (§5), then FE re-runs codegen and replaces `iam.ts` with typed `api.POST("/iam/users", {body: ...})` calls. Until then, all FE→backend IAM types are guesswork.

### 4.2 Normalization belongs server-side

`iam.ts:26-63` defines `normalizeManagedUser`, `normalizeOnlineUser`, `normalizeAuditEventItem` — defensive coalescing of `?? ""` and `?? 0`. This exists because backend returns `map[string]any` (see [admin_handler.go:142-193](../../internal/modules/iam/delivery/http/admin_handler.go:142)) instead of typed DTOs. FE compensates by re-typing every field.

**Verdict:** DELETE these normalizers once backend ships typed response structs through oapi-codegen. They are a symptom of the missing contract.

---

## 5. Frontend — `ManagedUsersPanel.tsx` (legacy CRUD cards)

User flagged this whole file as legacy. Below is the artifact-by-artifact enumeration.

### 5.1 Hardcoded department / process-area dropdowns (decorative)

[ManagedUsersPanel.tsx:53-65](../../frontend/apps/web/src/components/ManagedUsersPanel.tsx:53):

```ts
const DEPARTMENT_OPTIONS = [
  { value: "Operacoes",      label: "Operacoes" },
  { value: "Qualidade",      label: "Qualidade" },
  { value: "Engenharia",     label: "Engenharia" },
  { value: "Administrativo", label: "Administrativo" },
];
const PROCESS_AREA_OPTIONS = [
  { value: "Administrativo", label: "Administrativo" },
  { value: "Producao",       label: "Producao" },
  { value: "Logistica",      label: "Logistica" },
  { value: "Suprimentos",    label: "Suprimentos" },
];
```

These render in both Create-user card (lines 215-230) and Edit-user card (lines 316-331). Never sent to backend. Never persisted. **Pure decoration that lies to operators about what the form does.** The real area-membership flow lives at `/admin/memberships` against `user_process_areas` ([wiki/modules/iam.md §5.3](../../wiki/modules/iam.md)).

**Verdict:** DELETE outright. No replacement on the overview surface. Department/area assignment is the membership feature, not the user CRUD.

### 5.2 Fake role→department mapping

[ManagedUsersPanel.tsx:98-103](../../frontend/apps/web/src/components/ManagedUsersPanel.tsx:98):

```ts
function departmentFromRole(role?: UserRole) {
  if (role === "system_admin") return "Operacoes";
  if (role === "editor")       return "Qualidade";
  if (role === "approver")     return "Engenharia";
  return "Administrativo";
}
```

Pure invention. Displayed in the edit hero subtitle ([ManagedUsersPanel.tsx:299](../../frontend/apps/web/src/components/ManagedUsersPanel.tsx:299)) as if it were a real attribute. Tells the operator the user belongs to a department that doesn't exist in any backend column.

**Verdict:** DELETE.

### 5.3 PROFILE_OPTIONS — incomplete role set

[ManagedUsersPanel.tsx:45-51](../../frontend/apps/web/src/components/ManagedUsersPanel.tsx:45) lists only 5 of 8 canonical roles. A user with `signer`/`area_admin`/`qms_admin` cannot have their role surfaced or changed via this dropdown. Same family as QA finding #9.

**Verdict:** Will be replaced by drill-in role/membership UX (Phase 2). DELETE this constant.

### 5.4 Single-role dropdown over a multi-role array

[ManagedUsersPanel.tsx:117](../../frontend/apps/web/src/components/ManagedUsersPanel.tsx:117): `const selectedRole = props.managedUserForm.roles[0] ?? "viewer";`
[ManagedUsersPanel.tsx:133-138](../../frontend/apps/web/src/components/ManagedUsersPanel.tsx:133): `handleManagedRoleChange` writes `[value as UserRole]` — collapses any prior array to one. If a user ever has multi-role state, saving truncates silently.

Currently safe-by-accident because backend rejects multi-role via `parseExactlyOneRole`. Becomes a data-loss vector the day backend goes multi-role (§3.2 option B).

**Verdict:** DELETE entire dropdown. Replace per Phase 2 with role/membership manager that respects the chosen single-vs-multi contract.

### 5.5 Card-on-overview as a UX pattern

User flagged "creating user legacy, editing user legacy". The whole pattern — inline Criar usuário + inline Editar usuário on the same screen as KPIs + presence + activity — is amateur. Industry standard (Okta, Auth0, Linear, Datadog, Vercel admin): drill into a row → right-rail drawer or dedicated `/admin/users/:id` route → focused edit/invite there.

**Verdict:** DELETE both cards entirely. Replace per Phase 2 with:
- `<InviteUserDialog>` triggered from Users directory toolbar
- `<UserDrawer userId>` opened from row click

### 5.6 Imperative card-height syncing

[ManagedUsersPanel.tsx:148-170](../../frontend/apps/web/src/components/ManagedUsersPanel.tsx:148) — `useRef` + `ResizeObserver` to force the "Base de usuarios" card to match the edit card's measured height. CSS-grid-with-`align-items: stretch` solves the same layout in zero JS.

**Verdict:** DELETE.

---

## 6. Frontend — shell auth gate (`AppShell.tsx`)

[AppShell.tsx:19-25](../../frontend/apps/web/src/features/shell/components/AppShell.tsx:19):

```tsx
const requiresAdmin = matches.some(
  (m) => (m.handle as RouteHandle | undefined)?.requiresAdmin === true,
);

if (requiresAdmin && !roles?.includes('system_admin')) {
  return <Navigate to="/" replace />;
}
```

Pre-ADR-0016 ([wiki/decisions/0016-view-grade-capabilities.md](../../wiki/decisions/0016-view-grade-capabilities.md)). Backend already gates `GET /iam/admin/overview` on `CapUserView` ([permissions.go:123](../../apps/api/cmd/metaldocs-api/permissions.go:123)), not on the role literal. The FE gate over-restricts: a `qms_admin` or anyone holding `user.view` cap cannot reach `/admin` even though backend would let them.

`CurrentUser.capabilities` already exists at [lib/types/index.ts:38](../../frontend/apps/web/src/lib/types/index.ts:38) — the gate just doesn't read it.

**Verdict:** REDESIGN — replace `requiresAdmin: true` with `requiresCapability: 'user.view' | 'membership.view' | …` and gate on `capabilities.includes(...)`. Cross-cutting (also affects `/admin/memberships` and any future taxonomy-admin routes), so plan once, ship behind the rebuild PR.

---

## 7. Backend — IAM admin handler (`admin_handler.go`)

### 7.1 No oapi-codegen

[wiki/modules/iam.md §"API Route Truth Table"](../../wiki/modules/iam.md) confirms: every IAM admin route is hand-wired through [admin_handler.go:86-90](../../internal/modules/iam/delivery/http/admin_handler.go:86) (`mux.HandleFunc`). No `ServerInterface` method in `apps/api/internal/api/api.gen.go`. `openapi.yaml` declares request/response schemas for some endpoints but no `operationId` → no codegen stub.

Effect: FE↔BE contract is verbal, not generated. Every drift (see §3, §4) is invisible at compile time.

**Verdict:** MIGRATE. Per [.agents/skills/metaldocs-backend-api/SKILL.md](../../.agents/skills/metaldocs-backend-api/SKILL.md):
1. Author canonical `operationId`s in `openapi.yaml` for: `listUsers`, `createUser`, `patchUser`, `replaceUserRoles`, `resetUserPassword`, `unlockUser`.
2. Run codegen → produces `ServerInterface` stubs.
3. Replace `RegisterRoutes` + `handleUserRoute` suffix dispatcher with the generated router.
4. Implement the stubs against existing services.
5. Regenerate FE types → swap `features/iam/api/iam.ts` to typed `api.POST`/`api.PATCH`/`api.PUT` calls.

### 7.2 Suffix-dispatcher anti-pattern

[admin_handler.go:195-222](../../internal/modules/iam/delivery/http/admin_handler.go:195) `handleUserRoute`:

```go
path := strings.TrimPrefix(r.URL.Path, "/api/v1/iam/users/")
parts := strings.Split(path, "/")
if len(parts) == 2 && ... && parts[1] == "roles" { ... }
if len(parts) == 2 && ... && parts[1] == "reset-password" && r.Method == http.MethodPost { ... }
if len(parts) == 2 && ... && parts[1] == "unlock" && r.Method == http.MethodPost { ... }
if len(parts) == 1 && ... && r.Method == http.MethodPatch { ... }
```

Hand-parsing `parts[N]`, mixing method dispatch + path dispatch + length checks. Bug-prone. A user named `roles` collapses the dispatch. Replaced for free by codegen-driven `chi`/stdlib mux that handles `{userId}` properly.

**Verdict:** DELETE in the codegen migration.

### 7.3 `getAdminOverview` aggregates 3 unrelated concerns

[admin_handler.go:103-194](../../internal/modules/iam/delivery/http/admin_handler.go:103) `handleAdminOverview` fans out to:

| Source | What | Owner |
|---|---|---|
| `authService.ListUsers` | full user list | auth/IAM |
| `authService.ListOnlineUsers` | 10-minute presence window | auth/session |
| `auditReader.ListEvents` | last 25 governance events | audit module |

Returned as one envelope `{users, onlineUsers, recentActivities}`. Sequential, not parallelized. No cache headers. No pagination on the user list (the WHOLE tenant comes back every refresh). No filters on the audit slice.

Cross-module: this endpoint lives in the IAM handler but reads audit. That's a boundary leak. Audit should own its own list endpoint with its own cap gate.

**Verdict:** REDESIGN. Split into three queries on the FE side (TanStack Query handles the parallelism for free), each backed by a focused endpoint:

| Old | New (proposed) | Cap |
|---|---|---|
| `GET /iam/admin/overview` users[] | `GET /iam/users?paginate&filter` | `user.view` |
| `GET /iam/admin/overview` onlineUsers[] | `GET /iam/presence?activeSince=` | `user.view` (or new `presence.view`) |
| `GET /iam/admin/overview` recentActivities[] | `GET /audit/events?limit=&actor=&action=` | `metrics.view` (or new `audit.view`) |

The combined endpoint can stay as a server-side composition for the dashboard if perf demands, but **not** as the primary contract. Phase 2 deep-research will validate the split.

### 7.4 KPI summary is computed on the FE

[AdminCenterView.tsx:45,46](../../frontend/apps/web/src/features/iam/AdminCenterView.tsx:45) computes `onlineCount = adminCenter.onlineUsers.length` and `latestActivity = adminCenter.recentActivities[0]?.occurredAt`. Trivial here, but the moment KPIs get non-trivial (locked accounts, MFA coverage %, failed-logins-24h) the FE will be paginating-and-counting. That pattern should move to a backend KPI endpoint.

**Verdict:** REDESIGN — Phase 2 will name the KPI surface. Expect `GET /iam/kpi?window=24h` returning a typed `IamKpiSnapshot`.

### 7.5 Wiki claims that look stale

Verified against current code; do not propagate unchecked:

| Wiki claim | Reality | Verdict |
|---|---|---|
| [iam.md §"Top 3" T-005](../../wiki/modules/iam.md): "`handleUserRoleUpsert` does not emit `recordAudit`" | [admin_handler.go:369](../../internal/modules/iam/delivery/http/admin_handler.go:369) DOES emit `h.recordAudit(r, userID, "iam.user.role.upserted", ...)` | **wiki drift — T-005 should be marked CLOSED** |
| [iam.md §"Top 3" T-006](../../wiki/modules/iam.md): "IAM emits `{error:{code,message,details,trace_id}}` — does NOT yet match RFC 9457" | [internal/platform/problem/problem.go:11-20](../../internal/platform/problem/problem.go:11) IS RFC 9457; `writeProblem` calls `problem.Write` which sets `application/problem+json` | **wiki drift — T-006 should be marked CLOSED** |
| [iam.md §5.3 table](../../wiki/modules/iam.md): GET `/iam/admin/overview` gated on `user.manage` | [permissions.go:123](../../apps/api/cmd/metaldocs-api/permissions.go:123) gates it on `CapUserView` (post ADR 0016) | **wiki drift — cap column needs refresh** |

Dispatching `wiki-curator` is in scope for the close-out of the rebuild PR, not this audit. Listed here so the rebuild author doesn't act on stale claims.

### 7.6 `recordAudit` swallows errors

[admin_handler.go:465-491](../../internal/modules/iam/delivery/http/admin_handler.go:465) — every failure path inside `recordAudit` is silent:

- `h.audit == nil` → early return (no log)
- `tenant.FromContext` err → early return (no log)
- `json.Marshal(payload)` err → early return (no log)
- only the final `h.audit.Record(...)` failure goes to `log.Printf`

ISO 9001 requires audit trail completeness. A silent skip on tenant context missing is a compliance gap — operator action succeeds, audit row absent.

**Verdict:** MIGRATE — promote these to structured `slog.Warn` with the action + userID + reason. Consider failing the request if the audit write fails for a mutating call (current behavior is fire-and-forget after the response writes — concurrent log can land out-of-order).

### 7.7 `authenticatedActor` fallback to `"system"`

[admin_handler.go:523-532](../../internal/modules/iam/delivery/http/admin_handler.go:523) — if no `userID` in context, the audit `actorId` becomes `"system"`. Comment claims this is a tolerated fallback for bootstrap. But admin routes are gated on `CapUserManage` (a real user) — reaching `handleCreateUser` without an authenticated actor is a middleware bug, not a tolerated path. The fallback masks it.

**Verdict:** REDESIGN — split into:
- bootstrap helper `recordSystemAudit(...)` explicitly used by startup paths
- handler audit calls fail-closed if `authn.UserIDFromContext` returns false

### 7.8 Tenant-from-context: fallback chain

[wiki/modules/iam.md §"Last verified"](../../wiki/modules/iam.md) — IAM middleware tenant resolution: `tenant.FromContext` → `X-Tenant-ID` header → `DevTenantID`. The fallback chain is documented but the prod path should hard-reject if `tenant.FromContext` returns absent. Risk of leaking `DevTenantID` into production was already noted in Failure modes (wiki line 463). Not regressed here — flagged because the rebuild touches `handleAdminOverview` directly.

**Verdict:** KEEP existing behavior but add `bootstrap` config assertion that `AllowDevTenantFallback=false` in production (already noted in iam wiki Failure modes; track in `iam-tech-debt.md` if not already).

---

## 8. Backend — wiki-flagged debt that the rebuild should resolve

From [wiki/modules/iam.md §11 + iam-tech-debt.md](../../wiki/modules/iam.md):

- **T-004** — partially closed; `iam_users` INSERT still tier-1 only (no tier-2 `authz.Require`). User-creation path lacks defense-in-depth.
- **T-007** — `MembershipGovernanceLogger` wired with `nil` in [main.go:217](../../apps/api/cmd/metaldocs-api/main.go); governance log emission for area-membership grants is dead.
- **T-008** — `CachedRoleProvider` has no `InvalidateGroup`; group-membership writes don't invalidate.
- **T-010** — circular dep concern between `auth` and `iam` (auth imports `iamdomain.Role`).
- **T-011** — `tenant_id` per IAM table has no standalone ADR.

These are out of scope for the screen rebuild proper but the rebuild PR can choose to close T-004 (cheap — add `authz.Require(CapUserManage)` inside `iam_users` INSERT) and T-007 (wire the membership governance logger). Phase 2 will rank them.

---

## 9. Removal list (the user's explicit "remove legacy" call)

| Artifact | File | Lines | Why remove |
|---|---|---|---|
| "Criar usuário" inline card | `components/ManagedUsersPanel.tsx` | 188-250 | Wrong UX surface; replace with invite-dialog from Users directory toolbar |
| "Editar usuário" inline card | `components/ManagedUsersPanel.tsx` | 285-370 | Same; replace with right-rail `<UserDrawer>` |
| `DEPARTMENT_OPTIONS` / `PROCESS_AREA_OPTIONS` | `components/ManagedUsersPanel.tsx` | 53-65 | Hardcoded, decorative, not persisted |
| `departmentFromRole` | `components/ManagedUsersPanel.tsx` | 98-103 | Pure invention |
| `PROFILE_OPTIONS` (single-role list) | `components/ManagedUsersPanel.tsx` | 45-51 | Incomplete role set; replaced by Phase 2 role manager |
| `ROLE_LABELS` entries `admin`, `reviewer` | `components/ManagedUsersPanel.tsx` | 81-82 | Phantom role labels |
| `UserRole` `"admin"`, `"reviewer"` members | `lib/types/index.ts` | 15, 19 | Phantom enum members |
| `useUiStore.userForm` / `managedUserForm` / `managedUsers` | `store/ui.store.ts` | 31-33, 55-72, 80-88 | Server data + form state should not live in global store |
| `useAdminStore` (entire file) | `features/iam/state/admin.store.ts` | full | Replaced by TanStack Query |
| `useAdminCenter` (entire file) | `features/iam/useAdminCenter.ts` | full | Replaced by `useIamOverviewQuery()` |
| `useManagedUsers` (entire file) | `features/iam/useManagedUsers.ts` | full | Replaced by per-mutation hooks |
| `ManagedUsersPanel.tsx` / `ManagedUsersSection` | `components/ManagedUsersPanel.tsx` | full | Folded into new `<UsersDirectory>` + `<UserDrawer>` |
| Inline KPI card markup | `features/iam/AdminCenterView.tsx` | 120-163 | Replaced by `<KpiStrip>` + `<KpiCard>` primitives |
| `activityLabel` / `activityVariant` / `activityDotClass` / `activityChipClass` | `features/iam/AdminCenterView.tsx` | 53-93 | Substring sniffing; replace with `audit-event-presenter.ts` keyed off action enum |
| Dead "Ver todos →" button | `features/iam/AdminCenterView.tsx` | 174-176 | No handler |
| Dead "Audit trail →" button | `features/iam/AdminCenterView.tsx` | 209-211 | No handler |
| `handleUserRoute` suffix dispatcher | `internal/modules/iam/delivery/http/admin_handler.go` | 195-222 | Replaced by codegen router |
| `GET /iam/admin/overview` (as primary contract) | `internal/modules/iam/delivery/http/admin_handler.go` | 103-194 | Split into focused endpoints; may remain as a dashboard composition |

---

## 10. Bounded defers (out of scope for the rebuild)

| Item | Reason out-of-scope |
|---|---|
| T-008 group invalidation in `CachedRoleProvider` | Touches auth module + group migration; separate PR |
| T-010 circular `auth`↔`iam` dep | Architectural; needs ADR before code change |
| T-011 missing tenant_id ADR | Documentation-only |
| Migrating `/admin/memberships` page | Same module but different screen; addressed in its own QA pass |
| FE `requiresAdmin` → capability gate retrofit on **all** admin routes (not just `/admin`) | Cross-cutting; should be one focused PR right before this rebuild lands so the rebuild can use the new gate |

---

## 11. Pre-implementation checklist

Before writing any rebuild code, the implementer needs:

- [ ] Phase 2 deliverable `docs/audits/observability-admin-panel-research.md` reviewed and approved by user
- [ ] Concrete IA from Phase 2 (tab tree, surface inventory, drill-in topology)
- [ ] Decision on role contract: single-role vs multi-role (§3.2 option A vs B)
- [ ] Decision on overview endpoint: keep composed `/iam/admin/overview` for the dashboard OR force the FE to compose three queries
- [ ] Backend OpenAPI spec drafted for new IAM admin operations (operationId per route)
- [ ] FE capability gate prototype landed (replaces `requiresAdmin: true`)
- [ ] `wiki-curator` dispatched to refresh stale wiki claims §6.2 before the rebuild branches off, so the new code reads against accurate docs

---

## Appendix A — file inventory touched or deleted by rebuild

```
DELETE
  frontend/apps/web/src/features/iam/AdminCenterView.tsx
  frontend/apps/web/src/features/iam/useAdminCenter.ts
  frontend/apps/web/src/features/iam/useManagedUsers.ts
  frontend/apps/web/src/features/iam/state/admin.store.ts
  frontend/apps/web/src/components/ManagedUsersPanel.tsx
  frontend/apps/web/src/components/ManagedUsersPanel.module.css

MODIFY
  frontend/apps/web/src/features/iam/pages/AdminCenterPage.tsx        (rewire to new orchestrator)
  frontend/apps/web/src/features/iam/routes.tsx                       (cap gate, possibly add /admin/users/:id)
  frontend/apps/web/src/features/iam/api/iam.ts                       (replace with typed client calls or DELETE)
  frontend/apps/web/src/lib/types/index.ts                            (drop UserRole phantoms; revisit ManagedUserItem.roles)
  frontend/apps/web/src/store/ui.store.ts                             (drop iam form/server fields)
  frontend/apps/web/src/features/shell/components/AppShell.tsx        (capability gate)
  apps/api/cmd/metaldocs-api/permissions.go                           (any new cap mappings for split endpoints)
  api/openapi/v1/openapi.yaml                                         (add IAM admin operationIds + schemas)
  internal/modules/iam/delivery/http/admin_handler.go                 (replace with codegen-driven handlers)
  wiki/modules/iam.md                                                 (Last verified + close T-005/T-006)
  wiki/modules/iam-tech-debt.md                                       (close T-005/T-006, update T-004 status)

CREATE
  frontend/apps/web/src/features/iam/pages/AdminOverviewPage.tsx
  frontend/apps/web/src/features/iam/components/KpiStrip.tsx
  frontend/apps/web/src/features/iam/components/PresencePanel.tsx
  frontend/apps/web/src/features/iam/components/ActivityPanel.tsx
  frontend/apps/web/src/features/iam/components/UsersDirectory.tsx
  frontend/apps/web/src/features/iam/components/UserDrawer.tsx
  frontend/apps/web/src/features/iam/components/InviteUserDialog.tsx
  frontend/apps/web/src/features/iam/queries/useIamOverviewQuery.ts
  frontend/apps/web/src/features/iam/queries/useUsersQuery.ts
  frontend/apps/web/src/features/iam/queries/usePresenceQuery.ts
  frontend/apps/web/src/features/iam/queries/useActivityQuery.ts
  frontend/apps/web/src/features/iam/mutations/useInviteUserMutation.ts
  frontend/apps/web/src/features/iam/mutations/useUpdateUserMutation.ts
  frontend/apps/web/src/features/iam/mutations/useReplaceRolesMutation.ts
  frontend/apps/web/src/features/iam/mutations/useResetPasswordMutation.ts
  frontend/apps/web/src/features/iam/mutations/useUnlockUserMutation.ts
  frontend/apps/web/src/features/iam/audit-event-presenter.ts
  docs/audits/observability-admin-panel-research.md                   (Phase 2, in-flight)
```

(The component/query inventory is the audit's **structural prediction** — Phase 2 IA may rename or split further.)
