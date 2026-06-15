# Feature F4.1 — Plan (the "how")

> Spec: `spec.md` (Approved pre-code 2026-06-15). Executed subagent-driven (implementer → spec-review →
> quality-review), TDD. One PR for the port (governing spec §M4 "one PR per port").

## Structural anchors (verified 2026-06-15)

- **Port pattern to mirror:** `iam/domain/login_context_port.go` (interface) + `iam/infrastructure/postgres/login_context_repository.go` (pool-backed impl, `db *sql.DB`, tenant-scoped `WHERE user_id=$1 AND tenant_id=$2::uuid`).
- **Wiring sites:** `apps/api/cmd/metaldocs-api/main.go` (≈179, 411) and `apps/api/cmd/metaldocs-e2e-seed/main.go:60` both construct services/repos with `iampg.New…`.
- **Consumers:**
  - approval impl: `internal/modules/documents/approval/repository/postgres_approval_repository.go:446` `LoadActorDisplayName`; interface `approval/repository/approval_repository.go:98`.
  - documents: `internal/modules/documents/repository/repository.go` — `Repository{db *sql.DB}`, `New(db)`; inline read at `:134` inside `CreateDocumentTx`.
  - handler: `internal/modules/documents/approval/http/get_instance_handler.go:127` `resolveEligibleActorNames` (`h.db` direct).

## Steps

1. **Port (iam/domain).** New `iam/domain/user_display_name_port.go`:
   `UserDisplayNameReader` with `DisplayName(ctx, tenantID, userID) (string, error)` and
   `DisplayNames(ctx, tenantID string, userIDs []string) (map[string]string, error)`. Doc-comment the
   bounded-context rationale (iam owns `iam_users`; cross-module consumers meet it here, not at SQL).

2. **Impl (iam/infrastructure/postgres).** `user_display_name_repository.go`, pool-backed
   (`db *sql.DB`), mirror `LoginContextRepository`:
   - `DisplayName`: `SELECT display_name FROM metaldocs.iam_users WHERE user_id=$1 AND tenant_id=$2::uuid`; `sql.ErrNoRows → ("", nil)`; `MapPgError` on other errors.
   - `DisplayNames`: `SELECT user_id, display_name FROM metaldocs.iam_users WHERE tenant_id=$1::uuid AND user_id = ANY($2)`; **omit** rows whose `display_name` is null/empty; return `map[string]string`. Use `pq.Array`.
   - Constructor `NewUserDisplayNameRepository(db *sql.DB)`.

3. **Migrate approval signoff.** Inject the port into the approval repo (new field + constructor param,
   or an adapter). `LoadActorDisplayName` **delegates** to `port.DisplayName` — interface unchanged
   (`ApprovalRepository.LoadActorDisplayName` stays; consumer contract preserved). Read stays off-tx
   (port uses its own pool). Remove the raw `iam_users` SQL from `postgres_approval_repository.go`.

4. **Migrate documents created_by snapshot.** Add `displayName iamdomain.UserDisplayNameReader` field to
   documents `Repository`; thread through `New` (update all call sites + tests). In `CreateDocumentTx`
   replace the inline `tx.QueryRowContext(... iam_users ...)` (`:134`) with
   `r.displayName.DisplayName(ctx, d.TenantID, d.CreatedBy)` → assign into the existing
   `createdByDisplayName` (now a plain string / NullString-equiv) used by the INSERT. **Off-tx by
   construction** (port reads iam pool, not `tx`); **tenant-scoped** (uses `d.TenantID`).

5. **Migrate eligible-actor batch.** Inject the port into the approval HTTP `Handler`; in
   `resolveEligibleActorNames` replace the `h.db` query with `port.DisplayNames(ctx, tenantID, actorIDs)`;
   **keep** the existing post-loop `missing → userID` fallback (now also covers port-omitted
   empty/null names → identical rendered output). Drop `pq`/`h.db` use if now unreferenced there.

6. **Wire.** Construct `iampg.NewUserDisplayNameRepository(deps.SQLDB)` once in `main.go` and
   `metaldocs-e2e-seed/main.go`; inject into documents.New, the approval repo, and the approval handler.

## TDD order

1. iam port unit test (fake) + live PG integration test (`-tags integration`): present→value, absent→"",
   tenant-scoped; batch present/omit-empty. (Fails: port doesn't exist.)
2. Migrate `postgres_approval_repository_displayname_integration_test.go` to assert delegation/off-tx;
   `decision_service_test.go` snapshot assertion stays green.
3. documents create test: `created_by_display_name_snapshot` value preserved on a live create.
4. handler test: `StageActor` names byte-identical incl. `missing→userID` fallback.
5. Implement to green. Then grep gate: 0 `FROM metaldocs.iam_users` display-name SQL outside `iam/` in
   the 3 files.

## Verify (gate)

- `go build ./... && go vet ./...` clean.
- Touched-tree tests green (incl. `-tags integration` iam port + approval displayname live probes).
- `grep -rn "FROM metaldocs.iam_users" internal/modules/documents` → only non-display-name rows remain
  (area-name lookup etc.); zero display-name reads.
- Signoff runtime `pg_locks` proof: no cross-module read held in the signoff lock tx (H-PRE-1 intact).
- backend-api-qa-checklist green.
