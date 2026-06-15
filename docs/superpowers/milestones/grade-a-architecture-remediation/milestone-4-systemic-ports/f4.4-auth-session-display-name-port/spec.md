# Feature F4.4 — Spec

> **Milestone:** 4 — Systemic Ports (H-G class)  ·  **Folder:** `f4.4-auth-session-display-name-port`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-15 / operator (HS-4 fix feature named by the milestone-validator;
> scope confirmed under operator **Option 2 / full close**). Internal Go port migration; no public
> contract change. Engineering-grade decisions recorded in the interview record; the cross-module
> reach this closes is the validator's finding.

> This is the feature's **contract**, written and approved **before any code**. The milestone-validator
> judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Close the `sessions_admin` reach via (a) inject the iam port into auth's `Repository`, or (b) auth returns auth-owned rows and the **iam** consumer enriches names? | **(b).** The caller is iam (`iam/delivery/http/sessions_handler.go`), which *owns* `iam_users`. Having iam enrich via its own `UserDisplayNameReader` means **auth stops touching `iam_users` entirely** — zero cross-module reach anywhere, not merely relocated. Matches F4.1's `get_instance_handler` batch+fallback pattern. (a) would leave auth depending on the iam port — weaker boundary. |
| 2 | Auth's `SessionListItem` has a `DisplayName` field. Keep it (empty) or remove it? | **Remove it.** Auth does not own display names; the field is a leak of iam's concern into auth's domain type. The iam handler builds `display_name` from its port. Honest ISP — auth's row carries only auth-owned columns. |
| 3 | The old query is `auth_sessions s INNER JOIN iam_users u`. Dropping the JOIN changes the row set (orphan sessions no longer dropped). Behavior-preserving? | **Yes on the real path; documented.** `auth_sessions.tenant_id` already scopes the query. The INNER JOIN additionally dropped sessions whose user has no `iam_users` membership row in that tenant — but a session in tenant T exists only after a login to T, which requires `iam_users` membership, so no such orphan rows occur in practice. The membership *filter* belongs to the deferred iam tenant-scope port (F4.5), **not** to this display-name fix; absorbing it here would be scope creep. Recorded as a bounded, deliberate note (mirrors F4.1's two documented tightenings). |
| 4 | `display_name` value semantics? | Old: `COALESCE(NULLIF(u.display_name,''), s.user_id)`. New: `UserDisplayNameReader.DisplayNames` returns only present+non-empty names; the handler maps any missing `user_id → user_id`. **Byte-identical** rendered value. Tenant scoping preserved — the port is called with the same `tenantID` and scopes `iam_users.tenant_id`. |

## Consumer contract (FIRST — before any producer)

- **Consumer:** `iam/delivery/http/sessions_handler.go` `handleSessions` — renders `display_name` per
  session row in the `/api/v1/auth/sessions` list response (`entry["display_name"]`).
- **Producer (already exists, F4.1):** `iamdomain.UserDisplayNameReader.DisplayNames(ctx, tenantID, userIDs) (map[string]string, error)` — present+non-empty names; absent/empty omitted.
- **Auth's role (narrowed):** `SessionAdmin.ListActiveSessions` returns auth-owned session rows only
  (`session_id, user_id, ip_address, user_agent, created_at, last_seen_at, expires_at`) — **no**
  `display_name`. Tenant-scoped via `auth_sessions.tenant_id` (unchanged).
- **Source of truth:** the handler's existing render loop (`sessions_handler.go:111-129`) and the
  existing `UserDisplayNameReader` contract (`iam/domain/user_display_name_port.go`).

## What this feature implements

1. **auth `sessions_admin.go`** — `ListActiveSessions` query drops `JOIN metaldocs.iam_users u` and the
   `COALESCE(NULLIF(u.display_name,''), s.user_id)` column; selects only `auth_sessions` columns; scan
   drops `DisplayName`. Auth issues **0** `iam_users` SQL.
2. **auth `domain/session_admin.go`** — remove `DisplayName` from `SessionListItem`; update the doc
   comment (no longer "joined from iam_users").
3. **iam `sessions_handler.go`** — add a `displayNameReader iamdomain.UserDisplayNameReader` field,
   injected via a `WithDisplayNameReader(...)` setter (optional-collaborator pattern, mirrors
   `WithSessionService`; defaults to `iamdomain.NoopUserDisplayNameReader{}` so the 6 existing ctor
   call sites compile unchanged). In `handleSessions`, after `ListActiveSessions`: collect `user_id`s,
   call `DisplayNames(ctx, tenantID, ids)`, render `display_name = names[user_id]` falling back to
   `user_id` when absent.
4. **main.go wiring** — pass the existing pool-backed `UserDisplayNameRepository` into the sessions
   handler via `WithDisplayNameReader`.
5. Reads stay **live**; the port read is on the pool (off-tx) — H-PRE-1 not in play (list endpoint,
   no lock-holding tx).

## Non-goals (mandatory)

- **No** membership/tenant-scope filtering reintroduced (that is F4.5 / out of this feature). The
  documented orphan-session note is accepted, not "fixed".
- **No** OpenAPI / route / response-shape change — `display_name` key and value preserved.
- **No** change to the revoke path, `FindSession`, or any other `SessionAdmin` method.
- **No** snapshot/denormalization (D4/Approach-3 — reads live).
- **No** adjacent refactor beyond the named files (CLAUDE.md §5.3).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| `auth/` issues 0 `iam_users` SQL | `grep -rn "iam_users" internal/modules/auth/ --include=*.go` → 0 (excluding doc comments / tests) | real |
| `SessionListItem` has no `DisplayName` field | `grep -n "DisplayName" internal/modules/auth/domain/session_admin.go` → 0 | real |
| Handler renders display_name via port + `missing→user_id` fallback (byte-identical) | new `sessions_handler_test.go` case: fake `UserDisplayNameReader` returns a name for one user, omits another → response shows the name for the first, `user_id` for the second | fixture |
| `ListActiveSessions` returns tenant-scoped sessions without reaching `iam_users` (incl. a session whose user has no `iam_users` row — proves JOIN gone) | new live-PG integration test (`-tags integration`) on `*Repository.ListActiveSessions` | **real (live PG)** |
| Tenant isolation preserved (only caller's tenant) | existing `TestSessionsHandler_ListOnlyOwnTenant` green (migrated for new fake shape) | fixture |
| `go build ./...` + `go vet` (incl. `-tags integration`) clean | `go build ./...`; `go vet ./internal/modules/auth/... ./internal/modules/iam/... ./apps/api/...` | — |
| backend-api-qa-checklist green | checklist at close | — |

> TDD: failing handler test (port-rendered name + fallback) first, then implement to green; live-PG
> integration proves the JOIN removal.

## ADR needed?

- [ ] No new durable decision — F4.4 consumes the **F4.1** `UserDisplayNameReader` boundary already
  recorded in ADR 0029. The orphan-session membership note points forward to the **F4.5** tenant-scope
  port (its own decision/ADR). This feature adds a one-line cross-reference to ADR 0029.
