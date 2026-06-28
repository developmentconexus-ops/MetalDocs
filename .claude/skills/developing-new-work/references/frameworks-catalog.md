# Frameworks catalog — reuse, don't reinvent

**Last verified:** 2026-06-28

MetalDocs has platform frameworks for the recurring concerns. Reuse them. A hand-rolled equivalent of
any row below is a defect (the "we create frameworks" standard) — it bypasses the invariant the
framework exists to enforce. For each primitive the work touches, confirm reuse in §6 of the artifact.

| Primitive | Import / anchor | Use when |
|-----------|-----------------|----------|
| `TxRunner` (`Do` / `DoReadOnly`) | `internal/platform/db/runner.go:21` | Any DB work. Service depends on the tx port, not `*sql.DB`. Owns the tx boundary; nil tx rejected. |
| `tenant.FromContext` | `internal/platform/tenant/context.go:27` | Reading the current tenant. Never thread tenant id by hand. |
| `authz.SeedTxIdentity` | `internal/modules/iam/authz/context.go:58` | Start of every business tx — sets the tx-local GUCs the DB tripwire reads. |
| `authz.Require(ctx, tx, cap, area)` | pattern `internal/modules/templates/application/create.go:63` | Tier-2 in-tx capability check. The friendly first line before the DB tripwire. |
| `problem.New` / `problem.Write` | `internal/platform/problem/problem.go:76`; codes `codes.go:9` | Every error response. RFC 9457 `problem+json`. Never bare `http.Error`. |
| `httpresponse.WriteError` / success helpers | `internal/platform/httpresponse/` | Writing handler responses consistently. |
| `audit.NewEvent` / `Record` / `RecordTx` | `internal/modules/audit/` | Any state change that must be auditable. `RecordTx` to record inside the business tx. |
| Outbox repo | `internal/modules/render/fanout/staging_outbox.go:29` | Any external side effect. Enqueue in the business tx; a consumer does the network call idempotently. |
| `contracts.Decode` | `internal/platform/contracts/` | Strict JSON request decoding (rejects unknown fields). Don't `json.Decode` by hand. |
| `testdb.Open` + factory builders | `tests/integration/testdb/` | Integration tests. `Open(t)`, builders, `SeedWithCaps`, `Qualified`. See `test-qa-gates.md`. |

**Rule of thumb:** if you're about to write code that opens a tx, reads the tenant, checks a
capability, formats an error, calls out to the network, or seeds a test DB — there is already a
framework for it. Find it above first. If a genuinely new cross-cutting concern appears (no row fits),
that's a signal to design a *new* platform framework, not to inline a one-off — surface it as a
global-maximum question in §2 of the artifact.
