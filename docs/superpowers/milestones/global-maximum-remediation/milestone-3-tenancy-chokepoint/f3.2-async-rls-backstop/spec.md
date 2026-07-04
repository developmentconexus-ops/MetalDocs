# F3.2 — async RLS backstop (SeedTxTenant + compensating lint + negative proof)

> **Milestone:** M3 · **Binding contract:** `../validation-contract.md` §2 + §4 (worker/jobs rows) + §5.
> Contract governs on any conflict (HS-7). **Approval:** approved for implementation 2026-07-03.

## Consumer contract (who consumes what)

Consumers: (a) the `FORCE ROW LEVEL SECURITY` backstop on 33 tenant tables — for worker/jobs it engages
ONLY when each single-tenant processing tx seeds `metaldocs.tenant_id`; (b) the `ASYNC-TENANT-SEED` CI
lint, which needs every tenant-scoped async write to sit in a seeded tx or on the sanctioned allowlist;
(c) the negative RLS integration proof, which needs `SeedTxTenant` to actually block a cross-tenant touch.

**Required end-state:**
1. New primitive `authz.SeedTxTenant(ctx, tx, tenantID) error` — sets ONLY `metaldocs.tenant_id` via
   `set_config('metaldocs.tenant_id',$1,true)`; requires tenantID non-empty; sets/requires NO actor.
   (RLS reads only tenant; async has no human actor — actor-required `SeedTxIdentity` is the wrong tool.)
   Additive & orthogonal to `authz.BypassSystem` (RLS backstop vs write-tripwire are separate gates).
2. `SeedTxTenant` called at the START of the processing tx (before any lock/write) at the **5 single-tenant
   sites** (contract §2.2): materialize (existing tx), pdf (**wrap in tx**), scheduled-publish (existing
   `runner.Do`), notifications-fanout (**wrap in tx**), staging-outbox **processing** (per claimed row).
3. Blocking `ASYNC-TENANT-SEED` lint (`scripts/api-lint`): a tenant-scoped-table write in worker/jobs
   handler code must be inside a tx that called `SeedTxTenant`/`SeedTxIdentity`, unless on the §2.4
   allowlist. Handler-local AST scan (same technique as M2 lints) — cross-file claim→process is covered by
   allowlist + negative proof (recorded bounded defer, NOT a call-graph lint).
4. Sanctioned allowlist (§2.4): outbox claim steps (ADR 0054 rule 1), cross-tenant scans, system tables
   with no `tenant_id` (`idempotency_keys`, `job_leases`), audit-integrity scan.
5. Negative RLS integration proof (testdb factory, `//go:build integration`): leak pre-seed → blocked
   post-seed, real-DB, captured.

## Non-goals (mandatory)
- **No RLS policy change** — seeding coverage only (contract §5).
- **No actor invention for async** — `SeedTxTenant` is tenant-only; do not seed a fake actor.
- **No call-graph lint** — handler-local AST only; cross-file edges → allowlist + integration proof.
- **No F3.1 re-touch** — the chokepoint/`SeedTxIdentity` census is F3.1's; do not modify it.
- **No tenant-mixing "fix"** — if a handler mixes tenants in one tx, STOP (HS-2), do not workaround.

## Validation gate (positive + negative, captured output)
- **PG-1 (primitive):** unit test — `SeedTxTenant` sets `metaldocs.tenant_id`, requires non-empty, sets NO
  actor GUC.
- **PG-2 (seed sites):** the 5 handlers call `SeedTxTenant` before their first write/lock; pdf +
  notifications-fanout now open a tx. Compile + targeted handler tests green.
- **PG-3 (lint negative):** throwaway unseeded tenant-scoped worker write ⇒ `ASYNC-TENANT-SEED` RED naming
  file:line+table; remove ⇒ GREEN. Clean tree ⇒ 0 violations.
- **PG-4 (negative RLS proof, load-bearing):** real-DB integration — GUC-unset cross-tenant touch of
  tenant B's row SUCCEEDS (leak demo); after `SeedTxTenant(A)` the same touch → 0 rows (SELECT/UPDATE/
  DELETE) AND `42501` on a B-row INSERT. Labeled real-DB (testdb), not sqlmock. Bounded-defer the RUN if
  box can't run `-tags integration`; author regardless.
- **PG-5 (regression):** scheduled-publish + fanout still function (targeted drives); `go build ./...`
  green; H-PRE-1 preserved (seed before locks; no `authz.Require` added to locked paths).

## Named tests / proof commands
- `go test ./internal/modules/iam/authz/...` (SeedTxTenant unit)
- `go test ./internal/platform/worker/... ./internal/modules/documents/approval/jobs/... ./internal/modules/notifications/... ./internal/modules/render/fanout/...` (handlers)
- `go test ./scripts/api-lint/...` (ASYNC-TENANT-SEED unit incl. negative fixture)
- `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` (0 live violations)
- `go test -tags integration -run 'RLS|Tenant|Isolation' ./...` (negative proof; or authored + deferred)
- `go build ./...`

## Interview record

| Q | A (from runtime truth / contract §0.3, §2) |
|---|---|
| Why a new primitive vs reuse SeedTxIdentity? | RLS reads only tenant; async has no human actor; SeedTxIdentity hard-requires actor → wrong tool. SeedTxTenant is the minimum that engages FORCE RLS. |
| Is the async seam sanctioned or a new decision? | Sanctioned — ADR 0054 rule 2 already mandates tenant-scoped per-item processing with tx-local GUCs. F3.2 enforces it (no HS-2). |
| Which async units seed, which stay unseeded? | 5 single-tenant processing txs seed (§2.2); claim steps + cross-tenant scans + no-tenant_id system tables stay GUC-unset, allowlisted (§2.4). |
| Why can't the lint catch cross-file claim→process leaks? | Handler-local AST (M2 technique); call-graph resolution is a larger effort → bounded defer; allowlist + negative integration proof cover the class meanwhile. |
| What if a handler mixes tenants in one tx? | HS-2 surface — stop + report, do not workaround (contract §2.2). |
