# Feature F4.4 — Plan

> TDD. Failing handler test first → implement to green; live-PG integration proves the JOIN removal.
> Five touched files, one feature commit. No adjacent refactor (CLAUDE.md §5.3).

## Slices (ordered)

1. **RED — handler test.** In `iam/delivery/http/sessions_handler_test.go`:
   - Add a `fakeDisplayNameReader` implementing `iamdomain.UserDisplayNameReader` (`DisplayNames`
     returns a fixed map; `DisplayName` unused → "").
   - Migrate `fakeSessionAdmin.ListActiveSessions` to stop setting the (about-to-be-removed)
     `DisplayName` field; have it return two rows (user-1, user-2) for the caller tenant.
   - New `TestSessionsHandler_ListEnrichesDisplayNameViaPort`: reader maps `user-1→"Alice"`, omits
     `user-2`; GET `/api/v1/auth/sessions`; assert body has `"display_name":"Alice"` for user-1 and
     `"display_name":"user-2"` (fallback) for user-2. Wire reader via `WithDisplayNameReader`.
   - Compiles-red first (no `WithDisplayNameReader`, no enrichment) → fails.

2. **GREEN — auth domain.** `auth/domain/session_admin.go`: remove `DisplayName` from
   `SessionListItem`; rewrite the doc comment (rows are auth-owned session columns; tenant scope via
   `s.tenant_id`; display names resolved consumer-side by iam).

3. **GREEN — auth repo.** `auth/infrastructure/postgres/sessions_admin.go`: drop the
   `JOIN metaldocs.iam_users u …` and the `COALESCE(NULLIF(u.display_name,''), s.user_id)` SELECT
   column; drop `&item.DisplayName` from the scan. Update the function doc to note the tenant scope is
   `auth_sessions.tenant_id` only.

4. **GREEN — iam handler.** `iam/delivery/http/sessions_handler.go`:
   - Add field `displayNameReader iamdomain.UserDisplayNameReader`; default to
     `iamdomain.NoopUserDisplayNameReader{}` in `NewSessionsHandler` (keeps 6 existing ctor sites
     compiling).
   - Add `WithDisplayNameReader(r iamdomain.UserDisplayNameReader) *SessionsHandler` setter.
   - In `handleSessions`, after `ListActiveSessions`: collect unique `user_id`s, call
     `h.displayNameReader.DisplayNames(r.Context(), tenantID, ids)` (log+continue on error, names map
     empty → all fall back), render `display_name = names[item.UserID]` else `item.UserID`.
   - Import `iamdomain "metaldocs/internal/modules/iam/domain"`.

5. **GREEN — wiring.** `apps/api/cmd/metaldocs-api/main.go` sessions block (≈244-247): construct a
   pool-backed reader and chain `.WithDisplayNameReader(iampg.NewUserDisplayNameRepository(sqlDB))`.

6. **REAL — live-PG integration.** New `sessions_admin_integration_test.go`
   (`//go:build integration`) in `auth/infrastructure/postgres/`: seed one `auth_sessions` row in
   tenant T whose `user_id` has **no** `iam_users` row; assert `ListActiveSessions` returns it
   (proves the INNER JOIN is gone) and that a row in another tenant is not returned (tenant scope).
   Mirror the existing `repository_test.go` live harness.

7. **PROVE + CLOSE.** Run the Validation-Gate greps (`iam_users` in auth = 0 real; `DisplayName` in
   `session_admin.go` = 0); `go build ./...`; `go vet` (plain + `-tags integration`) on auth/iam/api;
   run the handler unit suite + the new live integration test; migrate
   `TestSessionsHandler_ListOnlyOwnTenant`; write `evidence.md`; one-line ADR-0029 cross-ref; commit.

## Files touched

- `internal/modules/auth/domain/session_admin.go` (remove field, doc)
- `internal/modules/auth/infrastructure/postgres/sessions_admin.go` (drop JOIN/column/scan, doc)
- `internal/modules/iam/delivery/http/sessions_handler.go` (port field + setter + enrich)
- `internal/modules/iam/delivery/http/sessions_handler_test.go` (fake reader + migrate fake + new case)
- `internal/modules/auth/infrastructure/postgres/sessions_admin_integration_test.go` (new live test)
- `apps/api/cmd/metaldocs-api/main.go` (WithDisplayNameReader wiring)
- `wiki/decisions/0029-user-display-name-reader-port.md` (one-line F4.4 consumer cross-ref)
- this feature's `spec.md` / `plan.md` / `evidence.md`
