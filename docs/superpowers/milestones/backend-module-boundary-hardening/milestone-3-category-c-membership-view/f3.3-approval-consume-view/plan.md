# Feature F3.3 — Plan: approval consumes `v_active_user_areas` (C3, H-PRE-1)

> Input: `spec.md` (approved 2026-06-21). Engine: `superpowers:writing-plans` shape, inline.

## Plan

### Files touched
- **NEW** `internal/modules/documents/approval/repository/eligible_actors_view_parity_integration_test.go`
  — the in-tx parity gate (`//go:build integration`, package `repository`).
- **EDIT** `internal/modules/documents/approval/repository/postgres_approval_repository.go`
  — repoint `ResolveEligibleActors` (`:1133`) SELECT from `metaldocs.user_process_areas` (interval predicate)
  to `metaldocs.v_active_user_areas` (predicate dropped); refresh method doc.
- **EDIT** `tools/cilint/internal/analyzers/hgcrossmodule.go`
  — drain the `{documents/approval/repository/postgres_approval_repository.go, user_process_areas}` (C3) entry.

### Test strategy (TDD, D6 — parity green BEFORE raw deleted; H-PRE-1 preserved)
1. Write `eligible_actors_view_parity_integration_test.go`:
   - `TestResolveEligibleActors_ViewParityWithRaw` — seed tenant + two fixed areas (`quality`, `safety`);
     members: eligibleA (active, qms_admin, quality), eligibleB (active, qms_admin, quality), wrongRole
     (active, approver, quality), revoked (qms_admin, quality, past `effective_to`), wrongArea (active,
     qms_admin, safety), wrongTenant (active, qms_admin, quality, other tenant). Open a real `*sql.Tx`;
     assert `repo.ResolveEligibleActors(ctx, tx, tenant, "quality", "qms_admin")` set == an inline
     `rawEligible(tx, …)` running the **verbatim deleted interval SQL** on the same tx.
   - Explicit assertions: {eligibleA, eligibleB} present; revoked / wrongRole / wrongArea / wrongTenant absent.
2. Run **pre-repoint**: green (raw repo == raw baseline — the revoked/interval rows prove the baseline itself).
3. Repoint `ResolveEligibleActors` to the view (drop the interval predicate).
4. Run **post-repoint**: green (view repo == raw baseline — parity; the revoked + wrong-role rows are the
   discriminators). Confirm the diff is a single relation token (H-PRE-1: no lock/record/tx-structure change).
5. Drain C3 ledger entry; re-green `go test ./tools/cilint/...` (fixture already realigned in F3.2);
   `go run ./tools/cilint ./...` exit 0.

### Ordering
spec✓ → plan✓ → parity test → pre-repoint green → repoint → post-repoint green → `go build ./...` →
drain ledger → cilint suite green + guard exit 0 → `git grep` clean → H-PRE-1 reviewer confirm → evidence.md → commit.

### Risks / mitigations
- **HS-PRE-1 (critical):** if set-equality required adding a recording read or restructuring the tx, STOP.
  Mitigation: it does not — a `SELECT`-only view swapped into the same `tx.QueryContext` is structurally
  identical; verified by the one-token diff.
- **Predicate-equality rests on Model A.** Mitigation: the parity test seeds only write-path-reachable rows
  (active / revoked-past); evidence records that future-dated rows (the only divergence) are unreachable.
- **Inline-raw baseline must be the verbatim deleted interval SQL** — copy it before editing the method.
- **HS-3** if PG :5434 down → parity not-run, do NOT delete raw, stop.
