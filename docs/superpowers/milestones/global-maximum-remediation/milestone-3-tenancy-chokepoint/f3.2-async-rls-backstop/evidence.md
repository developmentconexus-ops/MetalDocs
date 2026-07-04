# F3.2 evidence — async RLS backstop (SeedTxTenant + lint + negative proof)

> Contract: `../validation-contract.md` §2 + §4. Executed via subagent (TDD); main session reviewed the
> aggregate diff, ran an independent carrier-less-write audit, fixed one real gap the lint could not see,
> and commits. Real-DB proof authored; RUN GREEN for real in F3.5 (retargeted to `metaldocs.notifications`).

## What shipped

**Primitive (T1):** `internal/modules/iam/authz/context.go` — `SeedTxTenant(ctx, tx, tenantID)`: trims,
requires non-empty, `SELECT set_config('metaldocs.tenant_id', $1, true)` — **tenant-only, no actor**
(RLS reads only tenant; async has no human actor and keeps `authz.BypassSystem` for the write-tripwire).

**Five seed sites (T2)** — each seeds the claimed row/payload/args tenant at tx start, before any lock
(H-PRE-1; no `authz.Require` added to any locked path):
| Site | Tenant source | Tx |
|---|---|---|
| `internal/platform/worker/materialize_job_runner.go` | `payload.TenantID` | pre-existing `BeginTx` |
| `internal/platform/worker/pdf_job_runner.go` | `payload.TenantID` | **new tx wrapper** (`NewPDFJobRunnerWithDB` + `PDFPersisterInTx`; legacy `NewPDFJobRunner` path untouched for back-compat) |
| `internal/modules/documents/approval/application/scheduler_service.go` `RunScheduledPublishJob` | `input.TenantID` (River args) | pre-existing `runner.Do`; seed **before** the `FOR UPDATE` in `loadScheduledDocumentState` (ordering pinned by test) |
| `internal/modules/notifications/infrastructure/fanout_worker.go` `Work` | `args.TenantID` | **new `BeginTx`** wrapping fanout + `insertRow` |
| `internal/modules/render/fanout/staging_outbox.go` `MarkDispatched`/`MarkFailed` | `OutboxRow.TenantID` (threaded per row) | **new `inSeededTx`** per-row tx; the **claim** step (`ClaimPending` `FOR UPDATE SKIP LOCKED`) stays GUC-unset per ADR 0054 rule 1 |

No handler mixes tenants in one tx (each processes exactly one tenant's row/event) — no HS-2.

**Lint (T3):** `scripts/api-lint/async_tenant_seed_rule.go` — `ASYNC-TENANT-SEED`, blocking, registered in
`RunCodeRules`. AST-scans the async handler roots for `Exec`/`ExecContext` INSERT/UPDATE/DELETE against a
tenant-scoped table (`scripts/api-lint/async-tenant-tables.txt`, 32 FORCE-RLS tables) not inside a
`SeedTxTenant`/`SeedTxIdentity` function and not allowlisted → violation naming file:line+table; plus a
stale-allowlist rule. **Scope:** restricted to the async handler roots (`apps/worker`, `apps/jobs`,
`internal/platform/worker`, `internal/modules/documents/approval/jobs`,
`internal/modules/notifications/infrastructure`, `internal/modules/render/fanout`, + `scheduler_service.go`)
— a whole-tree scan false-positives on every F3.1-chokepoint-covered sync repository write; this scoping
mirrors how F3.1 `SEED-CHOKEPOINT` and M2 `TRIPWIRE-ARM-DRIFT` are exemption-scoped (contract §2.3
"the async binaries' handler packages").

**Allowlist (T4):** `scripts/api-lint/async-tenant-seed-allowlist.txt` — one machine-parsed entry
(`fanout_worker.go` `insertRow`, seeded by its same-file caller `Work` — the documented handler-local AST
limitation, ordering pinned by test). The §2.4 sanctioned categories (outbox/staging claim, watchdog
cross-tenant scan, `idempotency_keys`/`job_leases` no-`tenant_id`, audit scan) are documented as prose and
are **not** live AST matches (they use `QueryContext`, not `Exec`, or the table is absent from the list by
construction) — a real allowlist line for them would immediately register `ASYNC-TENANT-SEED-ALLOWLIST-STALE`.

## Main-session review finding (fixed before commit)

Independent audit: grep every `authz.BypassSystem` in the async roots and require a companion
`SeedTxTenant`. Two hits — `scheduled_publish_job.go` (a **comment** mention; the real seed is in
`scheduler_service.go`) and `stuck_instance_watchdog/job.go` (**3 BypassSystem, 0 seed**). Classifying its
three txs: `listStuckInstances` = cross-tenant **scan** (no write, correctly unseeded, §2.4);
`SystemCancelInstance`→`cancelInstance` = seeded by the F3.1 review fix; **`emitStuckAlert` = a
`governance_events` write** (a tenant-scoped table, in `async-tenant-tables.txt`) on the carrier-less
watchdog tx with `BypassSystem` but **no seed** — an unseeded async tenant-scoped write, which §2.6/§4
forbid outside the §2.4 allowlist, and which the lint **cannot** see (write goes through `emitter.Emit`,
not a literal `Exec`). **Fix:** added `authz.SeedTxTenant(ctx, tx, inst.TenantID)` after `BypassSystem` in
`emitStuckAlert`. The audit confirmed this was the only such gap.

## Negative RLS proof (T5) — authored; retargeted + RUN GREEN for real in F3.5

> **Update 2026-07-03 (F3.5):** the operator required a REAL run (no defer). First real run was RED — the
> proof targeted `documents`, whose M2 capability write-tripwire (`P0001`) fires before the RLS tenant
> policy and masked it. Fix-feature **F3.5** retargeted the proof to `metaldocs.notifications` (a FORCE-RLS
> `tenant_isolation` table, a real F3.2 async seed site, NOT tripwired), then ran it **GREEN** against live
> Postgres — all subtests pass, leak-before reproduced (1 row), no assertion weakened. See
> `../f3.5-rls-proof-real-green/evidence.md`. The bounded run-defer below is **CLOSED**.

`internal/modules/iam/authz/seed_tx_tenant_rls_integration_test.go` (`//go:build integration`, testdb
factory). Creates a **NOBYPASSRLS** role (the dev/test connection role is a BYPASSRLS superuser for which
RLS never applies), two tenants A/B each with a `metaldocs.notifications` row (F3.5; was `documents`), then:
- **leak_before_no_seed:** unseeded tx → B's row visible + a cross-tenant UPDATE affects 1 row (the
  pre-fix "async has zero backstop" evidence).
- **blocked_after_seed_tenant_a:** `SeedTxTenant(A)` → B invisible (SELECT), UPDATE/DELETE of B = 0 rows,
  A's own row still visible (scoped, not a blanket lockout); a re-tenant UPDATE producing a B-row → error
  **SQLSTATE 42501**.
Labeled **real-DB (testdb), not sqlmock**. Compiles + `go vet -tags integration` clean. **RUN GREEN for
real** in F3.5 (`ok metaldocs/internal/modules/iam/authz`, all subtests PASS) via a `.env`-loading
PowerShell wrapper that redacts the DB password from all output (no `.env` secret exposed).

## Gates (captured)

| Gate | Result |
|---|---|
| `go build ./...` / `go build -tags integration ./...` | exit 0 / exit 0 |
| `go test` iam/authz, platform/worker, approval/*, notifications/infrastructure, render/fanout, jobs/stuck_instance_watchdog, scripts/api-lint | ok |
| `go run ./scripts/api-lint -strict …` (live) | **0 violations** |
| PG-3 negative | stray unseeded `documents` UPDATE in a worker root → `ASYNC-TENANT-SEED` **RED** (names `zz_pg3_throwaway.go:12`, op UPDATE, table documents); removed → **GREEN** 0 |
| PG-4 negative RLS proof | authored + compiling; **RUN GREEN for real** (F3.5, on `metaldocs.notifications`) — all subtests PASS, leak-before=1 row, blocked-after=0/0, retenant=42501 |

## Behavior-change risks reviewed
- **pdf** + **notifications-fanout** moved from untransacted writes to tx-wrapped writes (required by
  §2.2). Verified via targeted sqlmock ordering tests: commit/rollback + error/retry preserved; fanout's
  idempotent `ON CONFLICT DO NOTHING` insert still runs inside the new tx; single event = single tenant
  (no ADR 0054 rule 2 risk). pdf legacy constructor path left intact for back-compat.
- **staging-outbox** `MarkDispatched`/`MarkFailed` gained a leading `tenantID` param; all production
  callers + the `fakeOutboxRepo` test double updated; no other callers.

## Bounded defers
- ~~PG-4 live run~~ — **CLOSED in F3.5**: run GREEN for real against live Postgres (proof retargeted to
  `metaldocs.notifications`; the `documents` capability write-tripwire had masked RLS). See
  `../f3.5-rls-proof-real-green/evidence.md`.
- Cross-file claim→process call-graph coverage — both async/sync lints are handler-local (contract §6);
  covered meanwhile by the allowlist + the negative proof + the main-session BypassSystem audit.

## Contract conformance (§2.6 exit criteria)
`SeedTxTenant` tenant-only ✓ · 5 processing txs seeded (pdf + fanout wrapped) ✓ · no tenant-mixing (no
HS-2) ✓ · `ASYNC-TENANT-SEED` registered/blocking, GREEN clean / RED synthetic ✓ · sanctioned allowlist
enumerated ✓ · negative RLS proof (leak→blocked, 42501) RUN GREEN for real (F3.5, notifications) ✓ · scheduled-publish +
fanout still function (targeted drives green) ✓ · `go build` green ✓ · H-PRE-1 preserved ✓ · **plus** the
`emitStuckAlert` gap closed so no tenant-scoped async write runs unseeded outside §2.4 (§4 posture met).
