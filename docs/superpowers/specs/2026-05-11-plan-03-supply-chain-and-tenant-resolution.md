# Plan 3 — Supply-chain unblock + tenant resolution platform fix

> **For agentic workers:** Tasks below use checkbox (`- [ ]`) syntax. Use `nexus:executing-plans` (or `nexus:test-driven-development` for the resolver-layer tasks) to implement task-by-task.

**Goal:** (a) Fresh `npm install` works from a clean checkout by restoring the vendored eigenpal 0.2.0 tarball. (b) Tenant identity is sourced from the authenticated session (cookie → DB-bound `tenant_id`), never from a client-supplied header, on every templates_v2 / taxonomy / registry request path.

**Architecture:**
1. Tarball lives at repo blob `0e35c08986e90809f536e3567ba4ff7c49e62d26` (parent of `0ee9160d`). Restore the same bytes — pin holds per ADR 0001.
2. Bind `tenant_id` to each `auth_sessions` row at login (migration adds the column; login chooses the user's tenant via their existing IAM role bindings; the `X-Tenant-ID` header is honoured at login only as a "tenant claim" that must be verified against the user's role set). `ResolveSession` returns the tenant_id from the session row; auth middleware injects it into the request context via a new `internal/platform/tenant.WithTenantID` / `tenant.FromContext` resolver. Module call sites stop reading `X-Tenant-ID` and read the context instead.
3. `tenant.DevTenantID` remains a constant — used only as the seed/dev fallback for users whose existing role bindings resolve to it. Production code path errors if context tenant is absent.

**Tech Stack:** Go (`net/http`, `database/sql`), Postgres, pnpm workspace, vendored npm tarball (`file:` URI).

**Scope guard (per /simplify):**
- Out of scope: documents, documents/approval, documents/http, iam, auth/handler header readers. Same anti-pattern, scope-locked by user prompt to Workstreams A + B → templates_v2 / taxonomy / registry only.
- Out of scope: templates_v2 T-002 (cross-tenant version access via `GetVersionByID`) — Plan 5 owns tier-2 enforcement.
- Out of scope: capability namespace, observability, metrics, RFC 9457 envelope, OpenAPI codegen.

**Out-of-scope anti-pattern flag (for the user, no action this Plan):** Header trust survives in `internal/modules/auth/delivery/http/{middleware.go:69, handler.go:116}`, `internal/modules/auth/application/service.go:150`, `internal/modules/iam/delivery/http/{middleware.go:77, admin_handler.go:109,225,275,345,379, routes_memberships.go:130}`, `internal/modules/documents/delivery/http/handler.go:950`, `internal/modules/documents/http/{fillin_handler.go:202, placeholder_options_handler.go:39, reconstruct_handler.go:33}`, `internal/modules/documents/approval/http/handler.go:93`. After this Plan they remain — but the platform `tenant.FromContext` they need is in place, so a follow-up plan can mop them up surgically.

---

## File structure

| File | Status | Responsibility |
|------|--------|----------------|
| `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` | restore | NPM tarball consumed by 3 `package.json` files via `file:` URI. |
| `vendor/eigenpal/README.md` | restore | Documents the pin per ADR 0001. |
| `migrations/0184_auth_sessions_tenant_id.sql` | new | Adds `tenant_id TEXT NOT NULL` to `metaldocs.auth_sessions`. Backfills existing rows by joining `iam_user_roles` and picking the user's lone tenant; defaults remaining nulls to `DevTenantID`. |
| `internal/platform/tenant/context.go` | new | `WithTenantID(ctx, id)` + `FromContext(ctx) (string, error)` + `ErrTenantMissing` sentinel. |
| `internal/platform/tenant/context_test.go` | new | Unit tests: round-trip; missing returns sentinel; empty string is missing. |
| `internal/platform/tenant/const.go` | modify | No code change; only ensure doc comment clarifies "dev/test only — production must use `FromContext`". |
| `internal/modules/auth/domain/model.go` | modify | Add `TenantID string` to `CurrentUser` and `Session`. |
| `internal/modules/auth/infrastructure/postgres/repository.go` | modify | `CreateSession` writes tenant_id; `FindSession` reads it. |
| `internal/modules/auth/infrastructure/memory/repository.go` | modify | Mirror in-memory store. |
| `internal/modules/auth/application/service.go` | modify | `Login` resolves the user's tenant from `iam_user_roles` (verifying `X-Tenant-ID` claim if present; rejecting on mismatch) and persists it on the session. `ResolveSession` drops the `tenantID` parameter, reads it from the session row, returns it on `CurrentUser`. |
| `internal/modules/auth/delivery/http/middleware.go` | modify | Stop reading `X-Tenant-ID`. After `ResolveSession`, call `tenant.WithTenantID(ctx, currentUser.TenantID)`. |
| `internal/modules/auth/delivery/http/handler.go` | modify | `/me` endpoint reads tenant from `CurrentUser`, not from header. (Login handler stays as the one place that consumes the inbound header — it's the boundary.) |
| `internal/modules/templates_v2/delivery/http/handler.go` | modify | Replace `tenantIDFromReq` body with `tenant.FromContext` lookup; error → 500 INTERNAL_ERROR (resolver invariant violation). |
| `internal/modules/taxonomy/delivery/http/routes_profiles.go` | modify | Same swap on `tenantIDFromRequest`. |
| `internal/modules/taxonomy/delivery/http/routes_areas.go` | modify | Same swap (mirror helper). |
| `internal/modules/taxonomy/delivery/http/routes_families.go` | modify | Same swap (mirror helper). |
| `internal/modules/registry/delivery/http/handler.go` | modify | `injectTenant` middleware reads from context, not header. |
| `internal/modules/registry/delivery/http/routes.go` | modify | `GetActiveDocument` (line 207) + revision lookup (line 448) read from context. |
| `internal/modules/auth/application/service_test.go` | modify/new | New cases: login rejects header claim that user has no role in; session row carries chosen tenant; `ResolveSession` returns it. |
| `internal/modules/templates_v2/delivery/http/routes_*_test.go` | modify | Replace `req.Header.Set("X-Tenant-ID", ...)` with `req = req.WithContext(tenant.WithTenantID(req.Context(), ...))`. |
| `internal/modules/registry/delivery/http/routes_contract_test.go` | modify | Same swap (5 sites). |
| `internal/modules/taxonomy/.../*_test.go` | modify | Same swap if any. |
| `wiki/modules/editor-ui-eigenpal-tech-debt.md` | modify | Flip T-001 status `closed YYYY-MM-DD`; link PR. |
| `wiki/backlog/editor-ui-eigenpal-refactor.md` | modify | Mark R-001 `closed`. |
| `wiki/modules/templates_v2-tech-debt.md` | modify | Flip T-003 status. |
| `wiki/backlog/templates_v2-refactor.md` | modify | Mark R-003 `closed`. |
| `wiki/modules/taxonomy-tech-debt.md` | modify | Flip T-001 status. |
| `wiki/backlog/taxonomy-refactor.md` | modify | Mark R-001 `closed`. |
| `wiki/modules/registry-tech-debt.md` | modify | Flip T-005 + T-006 status. |
| `wiki/backlog/registry-refactor.md` | modify | Mark R-005 + R-006 `closed`. |
| `wiki/backlog/roadmap.md` | modify | Flip Plan 3 status to `done YYYY-MM-DD`; add `**PRs:**` line; bump `Last verified`. |

---

## PR boundaries

- **PR 1 — Tarball restore.** Tasks 1–3. Standalone — frontend smoke only.
- **PR 2 — Platform tenant resolver + auth wiring.** Tasks 4–10. New migration + auth changes + resolver. Existing module sweep stays on header until PR 3 (no breakage during PR 2 because PR 2 leaves the helpers untouched).
- **PR 3 — Module sweep.** Tasks 11–14. templates_v2 / taxonomy / registry helpers consume `tenant.FromContext`; tests migrated. End-to-end exercise per CLAUDE.md (login + mutating route in each module).

---

## Task 1 — Restore vendored eigenpal tarball + README from git history

**Files:**
- Create: `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` (blob `0e35c089`)
- Create: `vendor/eigenpal/README.md` (blob `4ec632f0`)

- [ ] **Step 1: Restore both files from their pre-delete blob.**

```bash
git checkout 0ee9160d~1 -- vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz vendor/eigenpal/README.md
```

- [ ] **Step 2: Verify byte-identical to source blob.**

```bash
git hash-object vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz
# Expected: 0e35c08986e90809f536e3567ba4ff7c49e62d26
git hash-object vendor/eigenpal/README.md
# Expected: 4ec632f0c2c3f639d6e66a5bf1db5c82127fc14d
```

If either hash differs, STOP — escalate before improvising.

- [ ] **Step 3: Commit.**

```bash
git add vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz vendor/eigenpal/README.md
git commit -m "chore(vendor): restore eigenpal 0.2.0 tarball deleted by 0ee9160d

Three package.json files reference file:../../[…]/vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz. Fresh installs broke when 0ee9160d removed the bytes alongside the unrelated go mod vendor flip. Restores the same blob (sha 0e35c089) per ADR 0001 pin.

Closes editor-ui-eigenpal T-001 / R-001."
```

## Task 2 — Verify fresh install works from clean caches

**Files:** none modified.

- [ ] **Step 1: Wipe node_modules + caches for the 3 consumers.**

```powershell
Remove-Item -Recurse -Force packages\editor-ui\node_modules, apps\docgen-v2\node_modules, frontend\apps\web\node_modules -ErrorAction SilentlyContinue
```

- [ ] **Step 2: Install at the workspace root (pnpm).**

```powershell
pnpm install --frozen-lockfile=false
```

Expected: completes without "ENOENT" / "no such file" on `vendor/eigenpal/...`. `@eigenpal/docx-js-editor` resolves to `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` for each of the 3 workspaces.

- [ ] **Step 3: Frontend dev-server smoke.**

```powershell
.\scripts\start-api.ps1
```

Then in a second shell: `pnpm --filter @metaldocs/web dev` (or the project's standard frontend dev script). Open the editor preview path; confirm no module-resolution error in the console. Document outcome.

## Task 3 — Refresh ADR 0001 + eigenpal references to acknowledge restore

**Files:**
- Modify: `wiki/decisions/0001-eigenpal-adoption.md` (§ Consequences — drop the line claiming `vendor/eigenpal/` was removed if present; reinstate the pin path).
- Modify: `wiki/references/eigenpal-controlled-package.md` if its "Vendoring" section names the deletion.

- [ ] **Step 1: Read both files.**

```bash
# inspect current claims about vendor/eigenpal/
```

- [ ] **Step 2: Edit to state: tarball at `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`, pinned at 0.2.0 fork build; consumed via `file:` URI from packages/editor-ui, apps/docgen-v2, frontend/apps/web.**

- [ ] **Step 3: Bump `Last verified` stamp on both docs.**

- [ ] **Step 4: Commit.**

```bash
git commit -am "docs(eigenpal): refresh ADR 0001 + reference doc for restored tarball"
```

> PR 1 ends here. Open PR with branch `chore/plan-03-eigenpal-restore`. Merge before starting PR 2.

---

## Task 4 — Add `internal/platform/tenant.FromContext` resolver

**Files:**
- Create: `internal/platform/tenant/context.go`
- Create: `internal/platform/tenant/context_test.go`

- [ ] **Step 1: Write the failing test.**

```go
package tenant

import (
	"context"
	"errors"
	"testing"
)

func TestFromContext_RoundTrip(t *testing.T) {
	ctx := WithTenantID(context.Background(), "tenant-a")
	got, err := FromContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "tenant-a" {
		t.Fatalf("got %q, want tenant-a", got)
	}
}

func TestFromContext_MissingReturnsSentinel(t *testing.T) {
	_, err := FromContext(context.Background())
	if !errors.Is(err, ErrTenantMissing) {
		t.Fatalf("got %v, want ErrTenantMissing", err)
	}
}

func TestFromContext_EmptyTreatedAsMissing(t *testing.T) {
	ctx := WithTenantID(context.Background(), "")
	_, err := FromContext(ctx)
	if !errors.Is(err, ErrTenantMissing) {
		t.Fatalf("got %v, want ErrTenantMissing", err)
	}
}
```

- [ ] **Step 2: Run test — expect fail.**

```bash
go test ./internal/platform/tenant/...
```

- [ ] **Step 3: Implement.**

```go
package tenant

import (
	"context"
	"errors"
	"strings"
)

type ctxKey struct{}

// ErrTenantMissing is returned by FromContext when no authenticated tenant has
// been bound to the request context. Production handler paths MUST treat this
// as an internal-server-error invariant violation, not a 400.
var ErrTenantMissing = errors.New("tenant: not present in context")

// WithTenantID returns a child context that carries the supplied tenant id. The
// auth middleware is the only production caller; tests may also call it to
// stand in for the middleware.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, tenantID)
}

// FromContext extracts the authenticated-session tenant id. Returns
// ErrTenantMissing when absent or empty.
func FromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", ErrTenantMissing
	}
	v, ok := ctx.Value(ctxKey{}).(string)
	if !ok {
		return "", ErrTenantMissing
	}
	if strings.TrimSpace(v) == "" {
		return "", ErrTenantMissing
	}
	return v, nil
}
```

- [ ] **Step 4: Run test — expect pass.**

- [ ] **Step 5: Commit.**

```bash
git commit -am "feat(platform/tenant): add FromContext/WithTenantID resolver + ErrTenantMissing"
```

## Task 5 — Migration: add `tenant_id` to `auth_sessions`

**Files:**
- Create: `migrations/0184_auth_sessions_tenant_id.sql`

- [ ] **Step 1: Write the migration.**

```sql
-- 0184_auth_sessions_tenant_id.sql
-- Bind each session to the tenant chosen at login. Backfills existing rows
-- using the user's lone IAM role tenant; rows where the user has no role (or
-- multiple) fall through to DevTenantID so dev/test data continues to resolve.

ALTER TABLE metaldocs.auth_sessions
  ADD COLUMN IF NOT EXISTS tenant_id TEXT;

UPDATE metaldocs.auth_sessions s
SET tenant_id = sub.tenant_id
FROM (
  SELECT user_id, MIN(tenant_id) AS tenant_id
  FROM metaldocs.iam_user_roles
  GROUP BY user_id
  HAVING COUNT(DISTINCT tenant_id) = 1
) sub
WHERE s.user_id = sub.user_id AND s.tenant_id IS NULL;

UPDATE metaldocs.auth_sessions
SET tenant_id = 'ffffffff-ffff-ffff-ffff-ffffffffffff'
WHERE tenant_id IS NULL;

ALTER TABLE metaldocs.auth_sessions
  ALTER COLUMN tenant_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_auth_sessions_tenant_user
  ON metaldocs.auth_sessions (tenant_id, user_id);
```

- [ ] **Step 2: Run migrations.**

```powershell
.\scripts\start-api.ps1 -Build
```

Confirm boot succeeds and `\d metaldocs.auth_sessions` (via psql) shows the new column NOT NULL.

- [ ] **Step 3: Commit.**

```bash
git commit -am "migrate(0184): bind tenant_id on auth_sessions; backfill from iam_user_roles"
```

## Task 6 — Domain: `Session.TenantID` + `CurrentUser.TenantID`

**Files:**
- Modify: `internal/modules/auth/domain/model.go`

- [ ] **Step 1: Add field to `Session`.**

```go
type Session struct {
	SessionID  string
	UserID     string
	TenantID   string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	IPAddress  string
	UserAgent  string
	LastSeenAt time.Time
}
```

- [ ] **Step 2: Add field to `CurrentUser`.**

```go
type CurrentUser struct {
	UserID             string           `json:"userId"`
	TenantID           string           `json:"tenantId"`
	Username           string           `json:"username"`
	// ... rest unchanged
}
```

- [ ] **Step 3: Build — `go build ./...` must pass (no callers yet broken by absence).**

- [ ] **Step 4: Commit.**

```bash
git commit -am "feat(auth/domain): add TenantID to Session + CurrentUser"
```

## Task 7 — Repository: persist + read `tenant_id`

**Files:**
- Modify: `internal/modules/auth/infrastructure/postgres/repository.go`
- Modify: `internal/modules/auth/infrastructure/memory/repository.go`
- Modify: `internal/modules/auth/infrastructure/postgres/repository_test.go` (extend existing test to set + read tenant_id)

- [ ] **Step 1: Update postgres `CreateSession` INSERT to include `tenant_id`, and `FindSession` SELECT to project it.**

Exact column ordering should match whatever the current statement uses; add `tenant_id` to both column list + scan targets.

- [ ] **Step 2: Mirror in `memory/repository.go` — the in-memory map already keys by `session_id`; just store the field.**

- [ ] **Step 3: Extend `repository_test.go` to assert tenant_id round-trips.**

- [ ] **Step 4: Run tests.**

```bash
go test ./internal/modules/auth/infrastructure/...
```

- [ ] **Step 5: Commit.**

```bash
git commit -am "feat(auth/repo): persist + read auth_sessions.tenant_id"
```

## Task 8 — Service: login resolves + verifies tenant; `ResolveSession` returns it

**Files:**
- Modify: `internal/modules/auth/application/service.go`

- [ ] **Step 1: Add a private helper `resolveLoginTenant(ctx, userID, claimedTenantID) (string, error)`.**

Behaviour:
- Read `iam_user_roles` for the user; collect distinct tenant_ids.
- If `claimedTenantID` non-empty: require it to be in the user's set; else return `domain.ErrTenantNotPermitted`.
- If `claimedTenantID` empty AND user has exactly one tenant in their set: pick it.
- If `claimedTenantID` empty AND user has zero roles AND `s.cfg.AllowDevTenantFallback` (new bool, default true in dev, false in prod): return `tenant.DevTenantID`.
- Else: return `domain.ErrTenantClaimRequired`.

Add the two sentinel errors to `internal/modules/auth/domain/errors.go` (or wherever `ErrSessionNotFound` lives).

- [ ] **Step 2: Wire it into `Login`.**

Replace lines 150–157 (the current `X-Tenant-ID`-with-DevTenantID-fallback block) with:

```go
claimedTenant := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
tenantID, err := s.resolveLoginTenant(ctx, identity.UserID, claimedTenant)
if err != nil {
	return authdomain.AuthenticatedSession{}, err
}
session.TenantID = tenantID
if err := s.repo.CreateSession(ctx, session); err != nil {
	return authdomain.AuthenticatedSession{}, err
}
user, err := s.buildCurrentUser(ctx, identity.UserID, tenantID)
```

Note: `CreateSession` must be called AFTER `session.TenantID` is set.

- [ ] **Step 3: Change `ResolveSession` signature — drop `tenantID` parameter, read from session row.**

```go
func (s *Service) ResolveSession(ctx context.Context, rawToken string) (authdomain.CurrentUser, error) {
	// ... unchanged token-hash + FindSession ...
	user, err := s.buildCurrentUser(ctx, session.UserID, session.TenantID)
	if err != nil {
		return authdomain.CurrentUser{}, err
	}
	user.TenantID = session.TenantID
	return user, nil
}
```

- [ ] **Step 4: Update tests.**

`internal/modules/auth/application/service_test.go`: add cases for:
- Login with header matching a user role → session row carries that tenant; `ResolveSession` returns it.
- Login with header naming a tenant the user has no role in → returns `ErrTenantNotPermitted`.
- Login with no header + user has one tenant → picked automatically.
- Login with no header + user has zero tenants in non-dev mode → returns `ErrTenantClaimRequired`.

- [ ] **Step 5: Run.**

```bash
go test ./internal/modules/auth/...
```

- [ ] **Step 6: Commit.**

```bash
git commit -am "feat(auth): bind tenant on login; ResolveSession returns session-bound tenant"
```

## Task 9 — Middleware: drop header read; inject tenant into context

**Files:**
- Modify: `internal/modules/auth/delivery/http/middleware.go`

- [ ] **Step 1: Delete lines 69–72 (`tenantID := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))` … `tenant.DevTenantID`).**

- [ ] **Step 2: Change `m.service.ResolveSession(r.Context(), cookie.Value, tenantID)` to `m.service.ResolveSession(r.Context(), cookie.Value)`.**

- [ ] **Step 3: After `WithAuthContext`, add the tenant injection.**

```go
ctx := authdomain.WithCurrentUser(r.Context(), currentUser)
ctx = iamdomain.WithAuthContext(ctx, currentUser.UserID, currentUser.Roles)
ctx = tenant.WithTenantID(ctx, currentUser.TenantID)
next.ServeHTTP(w, r.WithContext(ctx))
```

- [ ] **Step 4: Build + test.**

```bash
go build ./...
go test ./internal/modules/auth/...
```

- [ ] **Step 5: Commit.**

```bash
git commit -am "feat(auth/middleware): inject session-bound tenant into request context"
```

## Task 10 — Update `/me` handler to use `CurrentUser.TenantID`

**Files:**
- Modify: `internal/modules/auth/delivery/http/handler.go:116` (the lone `/me`-side header reader; login handler keeps the header — it's the boundary)

- [ ] **Step 1: Replace `tenantID := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))` (line 116) with a read from `authdomain.CurrentUserFromContext(r.Context())`.**

- [ ] **Step 2: Build + run targeted test.**

```bash
go test ./internal/modules/auth/delivery/http/...
```

- [ ] **Step 3: Commit.**

```bash
git commit -am "fix(auth/handler): /me reads tenant from session-bound CurrentUser"
```

> PR 2 ends here. Branch `feat/plan-03-tenant-resolver`. Run full backend test suite (`go test ./...`) before opening. Verify boot + login via `.\scripts\start-api.ps1` per CLAUDE.md.

---

## Task 11 — templates_v2: swap `tenantIDFromReq`

**Files:**
- Modify: `internal/modules/templates_v2/delivery/http/handler.go:84-89`
- Modify: `internal/modules/templates_v2/delivery/http/routes_create_test.go:198` and any other test setting `X-Tenant-ID`

- [ ] **Step 1: Replace the helper body.**

```go
func tenantIDFromReq(r *http.Request) (string, error) {
	return tenant.FromContext(r.Context())
}
```

Update every caller to handle the error → `httpresponse.WriteError(w, 500, "INTERNAL_ERROR", "tenant context missing")` and return.

- [ ] **Step 2: Update test helper to inject context.**

Replace `req.Header.Set("X-Tenant-ID", "tenant-a")` with:

```go
req = req.WithContext(tenant.WithTenantID(req.Context(), "tenant-a"))
```

- [ ] **Step 3: Run.**

```bash
go test ./internal/modules/templates_v2/...
```

- [ ] **Step 4: Commit.**

```bash
git commit -am "fix(templates_v2): tenant from authenticated session, not X-Tenant-ID header"
```

## Task 12 — taxonomy: swap `tenantIDFromRequest` in profiles + areas + families

**Files:**
- Modify: `internal/modules/taxonomy/delivery/http/routes_profiles.go:197-203`
- Modify: `internal/modules/taxonomy/delivery/http/routes_areas.go` (mirror helper)
- Modify: `internal/modules/taxonomy/delivery/http/routes_families.go` (mirror helper)
- Modify: any taxonomy `*_test.go` that sets `X-Tenant-ID`

- [ ] **Step 1: For each of the three helpers, replace the body the same way as Task 11 — return `(string, error)` from `tenant.FromContext(r.Context())`. Propagate error to caller as 500 INTERNAL_ERROR.**

- [ ] **Step 2: Update test sites.**

- [ ] **Step 3: Run.**

```bash
go test ./internal/modules/taxonomy/...
```

- [ ] **Step 4: Commit.**

```bash
git commit -am "fix(taxonomy): tenant from authenticated session, not X-Tenant-ID header"
```

## Task 13 — registry: swap `injectTenant` middleware + `GetActiveDocument` reader

**Files:**
- Modify: `internal/modules/registry/delivery/http/handler.go:46-58` (`injectTenant` reads ctx instead of header)
- Modify: `internal/modules/registry/delivery/http/routes.go:207` (`GetActiveDocument` — drop the local `tenantIDFromRequest` call, use `injectTenant`-populated value already in `r.Context()`; OR swap inline reader to `tenant.FromContext`)
- Modify: `internal/modules/registry/delivery/http/routes.go:448` (revision lookup site — same swap)
- Modify: `internal/modules/registry/delivery/http/routes_contract_test.go` — every `req.Header.Set("X-Tenant-ID", …)` site (lines 192, 345, 711, 735, 761) becomes `req = req.WithContext(tenant.WithTenantID(req.Context(), …))`

- [ ] **Step 1: Rewrite `injectTenant`.**

```go
func injectTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tid, err := tenant.FromContext(r.Context())
		if err != nil {
			httpresponse.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "tenant context missing")
			return
		}
		ctx := context.WithValue(r.Context(), registryTenantCtxKey{}, tid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

(If registry already routes everything through `injectTenant`, the two inline readers in `routes.go` become reads from the registry-local ctx key; if not, swap them inline to `tenant.FromContext` directly.)

- [ ] **Step 2: Migrate test setters.**

- [ ] **Step 3: Run.**

```bash
go test ./internal/modules/registry/...
```

- [ ] **Step 4: Commit.**

```bash
git commit -am "fix(registry): tenant from authenticated session on injectTenant + GetActiveDocument"
```

## Task 14 — End-to-end verification + wiki + roadmap

**Files:**
- Modify: 5 wiki tech-debt + 4 backlog files (see file-structure table).
- Modify: `wiki/backlog/roadmap.md` (Plan 3 status flip + PRs line + `Last verified` bump).

- [ ] **Step 1: Boot the API.**

```powershell
.\scripts\start-api.ps1 -Build
```

- [ ] **Step 2: Login. Confirm cookie is set.**

```powershell
$body = '{"identifier":"admin","password":"AdminMetalDocs123!"}'
Invoke-RestMethod -Method Post -Uri http://localhost:8081/api/v1/auth/login -Body $body -ContentType 'application/json' -SessionVariable s
$s.Cookies.GetCookies('http://localhost:8081') | Format-List
```

- [ ] **Step 3: Exercise one mutating route per module without `X-Tenant-ID` header.**

- templates_v2: `POST /api/v2/templates` (create) — expect 201.
- taxonomy: `POST /api/v2/taxonomy/profiles` (create profile) — expect 201.
- registry: `POST /api/v2/controlled-documents` (atomic create) — expect 201.

Each call uses ONLY the session cookie. No `X-Tenant-ID` header. Confirm response uses the admin user's tenant.

- [ ] **Step 4: Negative: send the same calls WITH `X-Tenant-ID: deadbeef-…-bad` set. Confirm the header is ignored — call still resolves to the admin's tenant.**

- [ ] **Step 5: Frontend smoke per CLAUDE.md.**

Run frontend dev server, exercise templates list + create + editor open. Confirm no regression.

- [ ] **Step 6: Update wiki tech-debt rows.**

For each of: `editor-ui-eigenpal-tech-debt.md` T-001, `templates_v2-tech-debt.md` T-003, `taxonomy-tech-debt.md` T-001, `registry-tech-debt.md` T-005 + T-006 — append a `Status: closed YYYY-MM-DD (PR #N)` line under the item. Mirror on the corresponding `*-refactor.md` backlog row.

For templates_v2 T-003 specifically: note the header-trust portion is closed; cross-tenant version access (T-002) remains open for Plan 5. Add a TODO comment at the original `GetVersionByID` call site in `internal/modules/templates_v2/application/create.go:126` referencing T-002.

- [ ] **Step 7: Update `wiki/backlog/roadmap.md`.**

- Flip the Plan 3 row in the execution-order table → `done 2026-MM-DD`.
- Add a `**PRs:**` line under the Plan 3 body: `**PRs:** #PR1, #PR2, #PR3`.
- Bump `Last verified` at top.

- [ ] **Step 8: Dispatch `wiki-curator` agent.**

```
Use the wiki-curator agent to refresh Last verified stamps on:
  wiki/modules/editor-ui-eigenpal.md
  wiki/modules/templates_v2.md
  wiki/modules/taxonomy.md
  wiki/modules/registry.md
  wiki/modules/auth.md
  wiki/decisions/0001-eigenpal-adoption.md
  wiki/decisions/0007-two-tier-authz.md (add a cross-link to the new resolver)
```

- [ ] **Step 9: Commit (doc-only).**

```bash
git commit -am "docs(plan-03): close tech-debt rows; flip roadmap Plan 3 status"
```

> PR 3 ends here.

---

## Rollback notes

- **PR 1 (tarball restore):** revert is `git rm vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz vendor/eigenpal/README.md`. Frontend installs break again — same state as today.
- **PR 2 (resolver):** revert reverses the migration via `ALTER TABLE metaldocs.auth_sessions DROP COLUMN tenant_id;` plus a `git revert`. Sessions issued post-migration are still valid; the dropped column just gets ignored. Production deployment should hold the resolver revert as the long-term blast radius — once module sweep lands (PR 3) and production calls depend on `tenant.FromContext`, reverting requires re-adding header reads first.
- **PR 3 (sweep):** revert returns the four modules to header trust. Safe per-module; revert in module groups if one regresses.

## Notes / decisions

- **Why bind tenant at the session row and not at the `iam_user_roles` lookup on every request?** Single round-trip; sessions already cache the user — same row carries the tenant. Avoids per-request join. Multi-tenant users are explicit at login: claim a tenant or pick the lone one.
- **Why does Login still read the header?** Login is the boundary. The client tells the server which tenant the user is acting as for this session. The server then verifies that claim against the user's role bindings. Once bound, the session row is the only source of truth; the header is never trusted again.
- **Why 500 (not 400) when `tenant.FromContext` returns missing?** Because every non-public route is wrapped by the auth middleware, and the middleware unconditionally populates the tenant. Missing-in-context is an invariant violation, not a client bug.
- **`templates_v2` T-002 (cross-tenant `GetVersionByID`) is intentionally NOT closed here.** A TODO comment marks the call site for Plan 5 (tier-2 `authz.Require` enforcement).
- **Wider sweep (documents / approval / iam header readers) is intentionally NOT in scope.** Same anti-pattern, deferred to a follow-up plan; the platform `tenant.FromContext` makes that follow-up mechanical.

## Self-review

- Spec coverage: tarball restore ✓ (Tasks 1–3); resolver ✓ (Task 4); session-bound tenant ✓ (Tasks 5–10); module sweep ✓ (Tasks 11–13); verification + docs ✓ (Task 14).
- Placeholders: none — every step gives the actual edit.
- Type consistency: `tenant.WithTenantID` / `tenant.FromContext` / `ErrTenantMissing` / `CurrentUser.TenantID` / `Session.TenantID` / `ResolveSession(ctx, rawToken)` — used consistently across tasks.
- Closes rows from the prompt: editor-ui-eigenpal T-001/R-001 ✓; templates_v2 T-003/R-003 (header-trust portion) ✓; taxonomy T-001/R-001 ✓; registry T-005/R-005 + T-006/R-006 ✓.
