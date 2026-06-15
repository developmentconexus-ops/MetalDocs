# Feature F4.1 — Spec

> **Milestone:** 4 — Systemic Ports (H-G class)  ·  **Folder:** `f4.1-user-display-name-reader`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-15 / backend agent — *internal Go port; consumer contract derived from the 3 existing consumers (read, not invented); no public/OpenAPI contract change, behavior-preserving. Engineering-grade decisions recorded below; no operator contract decision required. The one operator-grade question (security scope) was resolved at the milestone Phase-2 gate → defer.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Are `security`'s `iam_users` reads an F4.1 target? | **Resolved at milestone Phase-2 (operator, 2026-06-15): defer.** They are tenant-scope JOINs on auth tables, not display-name reaches; need a different port. Bounded defer recorded in `milestone.md`. |
| 2 | The 3 consumers have **non-identical** semantics (tenant-scoped vs not; in-tx snapshot vs off-tx pool; raw "" vs COALESCE→userID). Does the port unify them, or preserve each verbatim? | **Engineering call (recorded below):** port exposes a uniform tenant-scoped `(tenantID, userID)` contract returning the **raw** display_name ("" / omitted when absent); each consumer keeps its **own presentation fallback** (handler's `→userID`). The two divergences from current code (documents read becomes tenant-scoped + moves off-tx) are deliberate, behavior-preserving for the normal single-tenant path, and **improve** boundary/H-PRE-1 posture. No operator decision needed — internal, no public contract change. |

## Consumer contract (FIRST — before any producer)

Three existing consumers reach directly into `metaldocs.iam_users` for `display_name`. The port is read
**from** these consumers; the producer (iam-owned impl) is built to match.

- **Consumer(s):**
  1. `documents/approval/application/decision_service.go:163` — single actor display-name for the
     signoff snapshot, via `s.repo.LoadActorDisplayName(ctx, tenantID, userID)` (impl
     `postgres_approval_repository.go:446`). Currently **tenant-scoped**, **off-tx** (pool), `"" on
     missing`, raw value (no fallback). This is **F1.3's contained fix** — the one to generalize.
  2. `documents/repository/repository.go:134` — `created_by` display-name snapshot written into
     `documents.created_by_display_name_snapshot`. Currently **NOT tenant-scoped** (`WHERE user_id=$1`),
     read **inside** the create tx, `sql.NullString` ("" when null/absent), raw value.
  3. `documents/approval/http/get_instance_handler.go:127` (`resolveEligibleActorNames`) — **batch**
     `map[userID]displayName`, **tenant-scoped**, `COALESCE(NULLIF(display_name,''), user_id)` plus a
     post-loop fallback filling any still-missing actorID with its own id.

- **Contract (the iam-owned port):**
  ```go
  // package iam/domain (owned by iam — the module that owns metaldocs.iam_users)
  type UserDisplayNameReader interface {
      // DisplayName returns iam_users.display_name for (tenantID, userID); "" when the user
      // is absent in that tenant (best-effort snapshot — matches all current single-read sites).
      DisplayName(ctx context.Context, tenantID, userID string) (string, error)

      // DisplayNames returns userID -> display_name for the users present in the tenant.
      // Users that are absent OR whose display_name is null/empty are OMITTED; the caller
      // applies its own fallback (the handler maps any missing id -> the id itself), which
      // reproduces the current COALESCE(NULLIF(display_name,''), user_id) behavior consumer-side.
      DisplayNames(ctx context.Context, tenantID string, userIDs []string) (map[string]string, error)
  }
  ```
  Implementation lives in `iam/infrastructure` against `metaldocs.iam_users`, on the connection
  **pool** (`*sql.DB`) — never on a caller's lock-holding tx connection.

- **Source of truth for the contract:** the three consumers above (their argument lists, null/empty
  handling, and the columns written/rendered). No OpenAPI/route involved — these are internal module
  ports; no generated type, no FE regen.

## What this feature implements

1. Add `UserDisplayNameReader` to **iam/domain** and its postgres impl to **iam/infrastructure**
   (pool-backed; single + batch; tenant-scoped `WHERE tenant_id = $1::uuid AND user_id …`).
2. Wire the iam impl into the composition root and inject it into the three consumers (adapter where a
   consumer's existing interface — e.g. approval's `ApprovalRepository.LoadActorDisplayName` — should
   delegate to the port rather than issue SQL).
3. Migrate each consumer to the port, preserving observable behavior:
   - **approval signoff** — `LoadActorDisplayName` delegates to `port.DisplayName`; stays **off-tx**
     (H-PRE-1 unchanged); "" on missing preserved.
   - **documents created_by snapshot** — read via `port.DisplayName` **before** the create tx
     (off-tx), pass the value into `Create`; the snapshot column gets the same value for the normal
     single-tenant path. Read becomes **tenant-scoped** (uses `d.TenantID`, already in scope) — a
     deliberate correctness tightening (see non-goals).
   - **eligible-actor batch** — `resolveEligibleActorNames` calls `port.DisplayNames`; the handler
     keeps its post-loop `missing → userID` fallback so the rendered names are byte-identical.
4. Result: **zero** raw `SELECT … display_name … FROM metaldocs.iam_users` outside `iam/` in these
   files (grep-provable).

## Non-goals (mandatory)

- **No** snapshot/denormalization semantics or "freeze actor name" product change (D4/Approach-3 — reads
  stay live).
- **No** migration of `security/infrastructure/postgres/repository.go` (tenant-scope JOINs — deferred at
  Phase-2), `iam/presence/*` (intra-module, not a cross-module reach), or `password_reauth.go` (already
  behind `IamUserReader`, reads `password_hash`).
- **No** OpenAPI / route / FE-type change; no change to any HTTP response shape (the rendered
  `StageActor` names and the `created_by_display_name_snapshot` value are preserved).
- **No** change to the signoff tx/lock model; the off-tx placement of the approval read is preserved,
  not re-touched.
- **Behavior-change scope is bounded to exactly two deliberate tightenings**, both recorded: documents
  created_by read becomes (a) tenant-scoped and (b) read off-tx. Both are no-ops on the normal
  single-tenant, user-present path; (a) only changes the pathological "same user_id in another tenant"
  case (now correctly scoped), (b) only moves a non-authz read out of the lock-holding tx (strictly
  better H-PRE-1 posture). Anything beyond these is scope drift.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Port exists in iam/domain; impl in iam/infrastructure; pool-backed | `go build ./...`; grep port decl in `iam/domain` | real |
| `DisplayName` returns value for present user, "" for absent, tenant-scoped | new iam port unit/integration test (single) | real (live PG integration) + fixture unit |
| `DisplayNames` batch returns present users, omits absent/empty | new iam port test (batch) | real + fixture |
| approval signoff snapshot unchanged; read stays off-tx (H-PRE-1) | `postgres_approval_repository_displayname_integration_test.go` (migrated to port) green; `decision_service_test.go` snapshot assertion green; signoff path `pg_locks`=0 cross-module-read-in-lock | real (integration + runtime pg_locks) |
| documents `created_by_display_name_snapshot` value preserved (single-tenant path) | documents create test asserts snapshot value; live create → row read-back | real |
| eligible-actor rendered names byte-identical | approval get-instance handler test asserts `StageActor` names incl. `missing→userID` fallback | real/fixture |
| **Class root cause:** 0 raw `iam_users` display-name SQL outside `iam/` in the 3 files | `grep -n "FROM metaldocs.iam_users" <3 files>` → 0 display-name matches | real |
| build + vet clean; backend-api-qa-checklist green | `go build ./... && go vet ./...`; checklist | real |

> TDD: failing test first (port contract + each migrated consumer), then implement to green. The live
> PG integration proofs are real-provider; fakes used for the consumer unit tests are labeled fixture.

## ADR needed?

- [x] Durable decision made → **ADR authored in F4.3** (`UserDisplayNameReader` boundary: iam owns the
  `iam_users` read; cross-module consumers depend on the port, reads stay live, no snapshot). Linked
  here when F4.3 lands.
