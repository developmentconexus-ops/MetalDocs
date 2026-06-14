# Feature F1.3 — approval signoff display-name reach (contain off-tx) — Spec

> **Milestone:** 1 (Reach-A Blockers)  ·  **Folder:** `f1.3-approval-displayname-reach`
> **Closes (partial):** Grade-A blocker — H-G class (cross-module reach + hardcoded domain
> state) **and** H-PRE-1 (advisory-lock deadlock constraint). This is the **contained** step
> (Approach-3 step 1). It does **not** generalize the cross-module reach into a shared port —
> that is M4/F4.1 (HS-6 scope guard).

## Interview record (fail-closed gate — read from source, not guessed)

| Q | A | Source |
|---|---|--------|
| Where is the offending read? | A raw `SELECT display_name FROM metaldocs.iam_users WHERE user_id=$1 AND tenant_id=$2::uuid`, run via `tx.QueryRowContext` **inside** the signoff transaction closure | `internal/modules/documents/approval/application/decision_service.go:266-274` |
| What tx is that closure? | The lock-holding signoff atomic tx: `runner.Do(ctx, func(tx *sql.Tx)…)` opens it (`:158`); inside, `LoadInstance` locks child stage rows `FOR UPDATE` (`:165-166`) and `authz.SeedTxIdentity` seeds the tenant GUC (`:161`) | `decision_service.go:158,161,165-166` |
| What does the read feed? | `ActorDisplayNameSnapshot: actorDisplayName.String` on the domain Signoff built by `domain.NewSignoff(…)` | `decision_service.go:289`; getter `domain/signoff.go:ActorDisplayNameSnapshot()`; param `domain.SignoffParams.ActorDisplayNameSnapshot` |
| What does the read depend on? | Only `req.ActorUserID` + `req.TenantID` — both **server-derived** request inputs available **before** the tx opens (already trusted at `:161` `SeedTxIdentity` and `:190` `authz.Require`) | `decision_service.go:150,161,190,271` |
| Why is in-tx the problem? | H-PRE-1 (advisory-lock-deadlock constraint): a cross-module read on a connection inside a lock-holding atomic tx can contend/deadlock; CD-create/signoff must stay fast | governing spec H-PRE-1; `wiki` advisory-lock-deadlock-constraint |
| Is an off-tx pool read on `iam_users` safe under RLS? | **Yes.** `metaldocs.iam_users` has `ENABLE + FORCE ROW LEVEL SECURITY` with a **NULL-permissive** `tenant_isolation` policy: `GUC unset/empty → rows visible`. The pool connection sets no `metaldocs.tenant_id` GUC, so the `IS NULL` branch applies; the explicit `tenant_id = $2::uuid` predicate keeps it tenant-correct | `db/migrations/0237_rls_all_tenant_tables.sql:11-14,109-117` |
| Is there an existing off-tx repo-method precedent? | **Yes** — `ListRoutes(ctx, tenantID)` reads `approval_routes` off the pool via `r.db.QueryContext` (no `tx`), same NULL-permissive RLS regime | `repository/postgres_approval_repository.go:421-428` |
| Where should the read live? | A **contained** method on the approval module's **own** `ApprovalRepository`, using the repo's `*sql.DB` pool. **Not** a shared cross-module `UserDisplayNameReader` port (that is M4/F4.1) | `repository/approval_repository.go:61`; milestone.md F1.3 row + HS-6 |

**No design fork** — the approach was pre-decided in `milestone.md` (Approach-3 step 1). No operator
hard-stop is required before implementing; the next hard-stop is HS-1 at the M1 close gate.

## Invariant under change (the "contract")

`RecordSignoff` MUST produce a `domain.Signoff` whose `ActorDisplayNameSnapshot` equals
`metaldocs.iam_users.display_name` for `(req.TenantID, req.ActorUserID)` — **byte-for-byte the same
value as today** (empty string when the user row is absent: today's read tolerates `sql.ErrNoRows`
and yields `""`). The **only** change is *where* and *how* that value is read:

- **Before:** `tx.QueryRowContext(...)` inside the lock-holding signoff tx (`decision_service.go:266-274`).
- **After:** `s.repo.LoadActorDisplayName(ctx, req.TenantID, req.ActorUserID)` called **pre-flight**,
  before `runner.Do(...)`, on the repository's pool (`*sql.DB`) — off the lock tx. The resolved
  string is captured by the tx closure (like `result`/`eligibilityEvent` already are) and passed to
  `domain.NewSignoff` unchanged.

**Deliberate, documented trade-off:** the off-tx read drops the GUC mismatch *tripwire* (it relies on
the explicit `tenant_id` predicate, not the seeded GUC). This is acceptable because `req.TenantID` is
server-derived (already authoritative for `SeedTxIdentity`/`authz.Require`), and the explicit predicate
is the authoritative tenant scope. Trading the tripwire for deadlock-safety is the whole point of F1.3.

## What this implements

1. **Add** `LoadActorDisplayName(ctx context.Context, tenantID, userID string) (string, error)` to
   `ApprovalRepository` (`repository/approval_repository.go`). It is an **off-tx** method (no `db.Tx`
   param), mirroring the existing `ListRoutes` precedent.
2. **Implement** it on `postgresApprovalRepository` using `r.db.QueryRowContext` (the pool) with the
   exact same SQL + tenant predicate as the inline read; tolerate `sql.ErrNoRows` → `""`; map other
   errors via `MapPgError`.
3. **Hoist** the call in `decision_service.go`: invoke `s.repo.LoadActorDisplayName(...)` **before**
   `runner.Do(...)`, capture the result, and **delete** the inline `tx.QueryRowContext` read. Feed the
   captured string into `domain.NewSignoff(… ActorDisplayNameSnapshot: actorDisplayName …)`.
4. **Align tests:** add the explicit override to the three fakes that drive `RecordSignoff`
   (`fakeDecisionRepo`, `fakeDecisionRepoWithCounter`, `phase5Repo` — all use the nil no-op embed, so
   the override is required only to avoid a nil-method panic, not for compilation); remove the now-dead
   in-memory driver interception of the `iam_users`/`display_name` query
   (`decision_service_test.go:357-360`) as an orphan of this change; add a positive proof test that the
   repo's returned display name is threaded into `ActorDisplayNameSnapshot`.

## Non-goals (mandatory — HS-6 scope guard)

- **No shared cross-module port.** Do **not** introduce a `UserDisplayNameReader`/IAM port or move the
  read to a shared package. The reach into `metaldocs.iam_users` **stays** (only its *location* moves
  off-tx, onto the approval module's own repo). Generalizing it is **M4/F4.1**.
- **No change to the signoff semantics, lock strategy, or tx boundary** beyond removing the one inline
  read. `SeedTxIdentity`, `FOR UPDATE`, `authz.Require`, content-pin, SoD, eligibility — all untouched.
- **No touch to the other `iam_users` read** at `http/get_instance_handler.go:129` (a different
  instance-read path, not the signoff snapshot) — out of F1.3 scope.
- **No value change.** Empty-on-missing behavior is preserved exactly.

## Validation Gate (acceptance — objectively checkable)

| # | Criterion | Named proof | Real vs fixture |
|---|-----------|-------------|-----------------|
| AC1 | The raw `iam_users` SELECT **no longer executes inside** the advisory-lock signoff tx (H-PRE-1 honored — read is off-tx / pre-flight) | `grep -n 'iam_users' decision_service.go` → **no** match inside the `runner.Do` closure; the read is a `s.repo.LoadActorDisplayName(...)` call **before** `runner.Do` (`:158`) | real |
| AC2 | The read lives on `ApprovalRepository`, **contained** — not a shared port | new method on `repository/approval_repository.go` + `postgres_approval_repository.go` using `r.db` (pool); **no** new cross-module/shared package introduced (HS-6) | real |
| AC3 | The persisted snapshot value is **unchanged** — `ActorDisplayNameSnapshot` still equals `iam_users.display_name` for `(tenant, actor)` | unit test: fake repo returns a sentinel display name → `insertedSignoff.ActorDisplayNameSnapshot()` equals it; empty-on-missing preserved | real |
| AC4 | Full approval test suite green; build + vet clean | `go build ./...` exit 0; `go vet ./internal/modules/documents/approval/...`; `go test ./internal/modules/documents/approval/...` all pass | real |
| AC5 | **Runtime proof:** a live signoff still returns the correct approver display name, with **no deadlock** | API on `:8081` (`.\scripts\start-api.ps1 -Build`); real login → submit → signoff; read back `approval_signoffs.actor_display_name_snapshot` = the approver's `iam_users.display_name`; signoff completes promptly (no lock wait) — by us, not deferred | real |

> AC5 is the H-PRE-1 close: prove the value is right **and** the lock section no longer holds the
> cross-module read. A successful, prompt signoff with the correct persisted snapshot is the
> acceptance — not "looks wired".

## ADR / decision record needed?

- [ ] No heavy ADR. The off-tx-contained pattern + the RLS-NULL-permissive justification are recorded
  here and will be folded into the wiki at M1 close (`wiki-curator`). The shared-port generalization
  decision belongs to M4/F4.1, not here.

## Approval

Approach pre-decided in `milestone.md` (F1.3 row, Approach-3 step 1) under the operator-approved
governing spec. Consumer/source contract read from source (table above). No implementation began
before this spec.
