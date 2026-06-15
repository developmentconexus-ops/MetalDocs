# Feature F4.4 — Evidence

> **Milestone:** 4 (Systemic Ports)  ·  **Feature:** `f4.4-auth-session-display-name-port`  ·  **Closed:** 2026-06-15
> **Contract:** [`spec.md`](spec.md). HS-4 fix feature (validator FAIL on `auth/sessions_admin.go:32`),
> executed under operator Option-2 full close. A feature is closed only when every row below is filled
> with real, honestly-labeled output.

## What was implemented

- **auth domain** `internal/modules/auth/domain/session_admin.go` — removed `DisplayName` from
  `SessionListItem` (auth does not own display names); doc comment rewritten (auth-owned columns only,
  tenant scope `s.tenant_id`, names resolved by the iam consumer via the port).
- **auth repo** `internal/modules/auth/infrastructure/postgres/sessions_admin.go` — `ListActiveSessions`
  query dropped the `JOIN metaldocs.iam_users u …` and the `COALESCE(NULLIF(u.display_name,''),
  s.user_id)` SELECT column; scan no longer reads `DisplayName`. Auth now issues **0** `iam_users` SQL.
- **iam consumer** `internal/modules/iam/delivery/http/sessions_handler.go` — added
  `displayNameReader iamdomain.UserDisplayNameReader` (defaults to `NoopUserDisplayNameReader{}` in the
  ctor, so all 6 existing `NewSessionsHandler` call sites compile unchanged), a `WithDisplayNameReader`
  setter, and a `resolveDisplayNames` helper. `handleSessions` collects unique `user_id`s, batch-reads
  via `DisplayNames`, and renders `display_name = name else user_id` — byte-identical to the old
  `COALESCE(NULLIF(display_name,''), user_id)`. Best-effort: on port error it logs and renders fallbacks
  rather than failing the list.
- **wiring** `apps/api/cmd/metaldocs-api/main.go` sessions block — chained
  `.WithDisplayNameReader(iampg.NewUserDisplayNameRepository(sqlDB))` (pool-backed, off-tx).
- **ADR 0029** — added the sessions_handler consumer to Key files; **corrected the falsified census claim**
  (security JOINs were mislabeled "tenant-scope only, no display-name"); narrowed the bounded defer to
  security's no-display-name aggregate JOINs; pointed forward to F4.5/F4.6.

## Verification

Integration run used the live dev Postgres
(`postgres://…@127.0.0.1:5433/metaldocs?sslmode=disable&search_path=metaldocs,public`),
`-tags integration -count=1`.

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first | add `WithDisplayNameReader`/enrich case before impl | `vet.exe: …WithDisplayNameReader undefined` (RED) → green after impl | real + fixture |
| Handler enriches via port + missing→user_id fallback | `go test -count=1 ./internal/modules/iam/delivery/http/` (`TestSessionsHandler_ListEnrichesDisplayNameViaPort`: user-1→"Alice", user-2 omitted→`"user-2"`) | `ok …/iam/delivery/http 3.009s` | fixture (fake port) |
| Tenant isolation preserved (migrated fake) | same suite — `TestSessionsHandler_ListOnlyOwnTenant` | `ok` (body has `sess-a`, not `sess-b`) | fixture |
| `ListActiveSessions` returns a session whose user has **no** `iam_users` row (proves INNER JOIN gone) + tenant scope holds | `go test -tags integration -count=1 -run TestListActiveSessions_NoIamUsersJoin ./internal/modules/auth/infrastructure/postgres/` | `ok …/auth/infrastructure/postgres 2.273s` | **real (live PG)** |
| **Class root cause** — `auth/` issues 0 real `iam_users` SQL | `grep -rn iam_users internal/modules/auth/ --include=*.go` (excl. tests) | only doc comments remain (service.go login comment; session_admin doc; sessions_admin doc) — 0 SQL | real |
| `SessionListItem` has no `DisplayName` field | `grep -n DisplayName internal/modules/auth/domain/session_admin.go` | 1 hit, a doc-comment word only — field gone | real |
| Static — build | `go build ./...` | `BUILD DONE` (exit 0) | — |
| Static — vet (plain) | `go vet ./internal/modules/auth/... ./internal/modules/iam/... ./apps/api/...` | `BUILD+VET OK` | — |
| Static — vet (integration tag) | `go vet -tags integration ./internal/modules/auth/infrastructure/postgres/ ./internal/modules/iam/... ./apps/api/...` | `VET-INTEGRATION DONE` | — |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| `auth/` issues 0 `iam_users` SQL | yes | grep clean (doc comments only) |
| `SessionListItem` has no `DisplayName` field | yes | grep — field removed |
| Handler renders via port + missing→user_id (byte-identical) | yes | `…_ListEnrichesDisplayNameViaPort` PASS (fixture) |
| `ListActiveSessions` tenant-scoped without reaching `iam_users` (incl. no-iam_users-row session) | yes | `TestListActiveSessions_NoIamUsersJoin` PASS (**real live PG**) |
| Tenant isolation preserved | yes | `…_ListOnlyOwnTenant` PASS |
| build + vet (incl. integration) clean | yes | BUILD+VET OK, VET-INTEGRATION DONE |
| backend-api-qa green | yes | no route/OpenAPI/shape change; `display_name` key+value preserved; only internal port migration |

## Review disposition

- Spec-compliance: consumer-contract-first honored — auth narrowed to auth-owned rows; the iam consumer
  (owner of `iam_users`) enriches via its own port. Zero cross-module reach (not relocated). The one
  deliberate behavior note (INNER-JOIN drop → orphan sessions no longer filtered) is documented in
  spec interview Q3 and is a no-op on the real path (a session in tenant T implies `iam_users`
  membership in T); the membership *filter* is F4.5's concern, not absorbed here.
- Code-quality: optional-collaborator pattern mirrors `WithSessionService`; Noop default keeps the 6
  existing ctor sites + unit tests compiling; port read is off-tx on the pool (H-PRE-1 not in play —
  list endpoint, no lock-holding tx); best-effort error handling degrades to fallback names, never
  500s the list on a name-lookup hiccup.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Orphan-session membership *filter* (sessions whose user lacks an `iam_users` row in the tenant are no longer dropped) | Not reachable on the real path (sessions exist only after a login that requires membership); the filter is a tenant-scope concern, not a display-name concern | F4.5 (`iam-tenant-membership-port`) provides the membership reader; revisit only if a real orphan-session path appears. Owner: backend |
