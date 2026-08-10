# Invariant checklist — the 6 non-negotiables

**Last verified:** 2026-07-06

These are MetalDocs' non-negotiable invariants (CLAUDE.md). Violating one is a defect, not a design
choice. For each: is the work *touched* by it? *how* is it satisfied? which *helper* do you reuse?
Anchors are for **targeted verify only** — read an anchor only when an item is genuinely uncertain.

> When an anchor has moved, the code wins: fix the anchor here and bump `Last verified`.

## 1. AuthZ = capabilities, never roles
Two-tier PDP + DB tripwire. Never reason "admin/author can X" — reason in capabilities.
- tier-1 route→capability (middleware): `apps/api/cmd/metaldocs-api/permissions.go`
- tier-2 capability×area in-tx: `authz.Require(ctx, tx, cap, areaCode)` after `authz.SeedTxIdentity`;
  pattern in `internal/modules/templates/application/create.go:67` (sig `iam/authz/authz.go:76`; ScopeTenant passes areaCode `"tenant"`)
- DB tripwire is the last line. Governed by ADR 0022.
- Adding a capability? → full 10-touchpoint walk in `capability-wiring.md`.

## 2. Contract-first (OpenAPI + oapi-codegen)
Routes change ONLY by editing the spec, then regenerating. The spec is route truth.
- spec: `api/openapi/v1/openapi.yaml`
- per-module generation: `api/cfg.yaml` + `gen.go`; run `go generate ./internal/modules/<m>/api/...`
- Never hand-add a route in Go that isn't in the spec.

## 3. Multi-tenant pooled
Every tenant table has `tenant_id`; tx-local GUCs only; tenant-namespaced blob keys; cross-tenant URL
→ 404 (not 403).
- read tenant from context: `tenant.FromContext` — `internal/platform/tenant/context.go:27`
- seed the in-tx identity (sets GUCs): `authz.SeedTxIdentity` — `internal/modules/iam/authz/context.go:48`
- new tenant table ⇒ `tenant_id` column + the tenant predicate on every query.

## 4. Async = transactional outbox
A state-write + an external side effect never share a tx with a network call. Enqueue in the business
tx; consumers are idempotent.
- outbox pattern: `internal/modules/render/fanout/staging_outbox.go:29`
- if the work calls out to anything (email, blob, webhook, render) → it goes through the outbox, not
  an inline network call inside the handler tx.

## 5. DB enforces invariants
Triggers/constraints are the enforcement; app checks are the friendly first line, not the guarantee.
- add the DB constraint/trigger for any new invariant; don't rely on app code alone.
- capability-format tripwires: `db/baseline/0001_current_schema.sql` (`ck_cap_format`, `ck_cap_not_legacy`).

## 6. Cross-module access via published interface only
Never touch another module's repository, SQL, or domain internals. Cross-module access goes through a
module's application service or published Go interface.
- published ports live in the owner's `domain/port.go` (provider) / `application/ports.go` (consumer).
- if you need data from module B, depend on B's interface — adding a method to B's port if needed,
  never reaching into B's tables.

## Supporting platform invariants (not in the 6, but enforced)
- **TxRunner**: services depend on the tx port, not `*sql.DB`; nil tx is rejected.
  `internal/platform/db/runner.go:21` (`Do` / `DoReadOnly`).
- **Errors are RFC 9457 `problem+json`**: never bare `http.Error`.
  `internal/platform/problem/problem.go:77` (`Write`); codes `internal/platform/problem/codes.go:9`.
- **Fixed request lifecycle**: `panic_recovery → otel → http_obs → cors → origin_protection →
  pre_auth_login_rate_limit → authn → iam_authz → presence_bump → rate_limit → contract_validation → method_not_allowed`
  is inherited; new routes don't re-wire it. `apps/api/cmd/metaldocs-api/chain.go:25`. Idempotency is
  **not** a chain link — it is enforced per-handler/per-service where needed (e.g.
  `internal/modules/documents/approval/application/signoff_idemp.go`).
- **H-PRE-1** (LIVE — never retired): never call an authz-recording read inside a lock-holding atomic tx;
  hoist it off-tx. Motivating lock = the audit hash-chain writer's `pg_advisory_xact_lock`
  (`internal/modules/audit/infrastructure/postgres/writer.go:59`) + `authz.Require` recording a
  system_admin bypass audit in the caller's tx (`internal/modules/iam/authz/authz.go:119`). Also governs the
  auth-repo / documents-repo / migrate advisory locks. **M5 note:** M5 removed the stuck-instance-watchdog's
  `pg_try_advisory_lock` — that was single-runner mutual exclusion (River's elector+queue now subsume it),
  **unrelated** to H-PRE-1's authz-read-in-locked-tx hazard. Removing it neither triggers nor retires H-PRE-1.
  ([[advisory-lock-deadlock-constraint]].)
