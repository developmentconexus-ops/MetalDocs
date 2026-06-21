# Feature F3.3 — Spec: approval consumes `v_active_user_areas` (C3, in-tx, H-PRE-1)

> **Milestone:** 3 — Category C  ·  **Folder:** `f3.3-approval-consume-view`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-21 / leandrotca (operator) — via operator-approved mission §7 M3 + ADR-0039 D3(a) + ADR-0037 Model A. The single consumer call site (`ResolveEligibleActors`) is read this session; the change is a one-relation swap inside the existing in-tx SELECT. *No implementation begins until this line is filled — it is.*

> The feature's **contract**, approved **before any code**. The milestone-validator judges F3.3 against this file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | What exactly changes in `ResolveEligibleActors`? | The single `tx.QueryContext` SELECT's relation: `FROM metaldocs.user_process_areas WHERE … AND effective_from <= now() AND (effective_to IS NULL OR effective_to > now())` becomes `FROM metaldocs.v_active_user_areas WHERE tenant_id=$1 AND area_code=$2 AND role=$3` — the entire temporal predicate is **dropped** (the view encodes active-now). The `role` filter stays (the view exposes `role`). Nothing else in the method (signature, scan loop, nil-slice contract) changes. |
| 2 | The old predicate is an **interval** form; the view is `effective_to IS NULL`. Same set? | Yes, under ADR-0037 **Model A** (the system's only write path): active rows have `effective_to = NULL`; revoke stamps a **past** `effective_to`; `effective_from` is set to `now()` at insert (never future). So `effective_from <= now()` is always true and `effective_to > now()` is the **empty set** — the interval form reduces to exactly `effective_to IS NULL`. Repointing both preserves the set **and** retires the Model-B interval leak (`effective_to > now()`, refuted by ADR-0037 D2). The only rows on which the two predicates could differ (future-dated `effective_to`/`effective_from`) are **unreachable** via the write path. |
| 3 | H-PRE-1 — is the in-tx posture preserved? | Yes. It stays a plain, **non-recording** `tx.QueryContext` SELECT on the caller's `db.Tx`. Swapping one relation name adds no authz-recording read, no lock, no extra round-trip, no tx-structure change. The view is `SELECT`-only (no `security_invoker`, no function) — structurally identical read. D5 preserved by construction. |
| 4 | What proves no authz drift? | A parity test that calls `repo.ResolveEligibleActors` **on a real `*sql.Tx`** (view form, post-repoint) and asserts its set == a **verbatim inline copy of the deleted interval SQL** on the same tx, over: two eligible members, a wrong-role member, a **revoked** member, a wrong-area member, a wrong-tenant member. The revoked + wrong-role rows are the discriminators. |
| 5 | Any guard/ledger change? | Drop the `{documents/approval/repository/postgres_approval_repository.go, user_process_areas}` (C3) entry from `hgPendingRemediation`. The cilint `PendingBaseline` fixture was already realigned to a C4 site in F3.2, so **no further fixture edit** is needed; re-green `go test ./tools/cilint/...` and keep `go run ./tools/cilint ./...` exit 0. |

## Consumer contract (FIRST)

- **Consumer:** `postgresApprovalRepository.ResolveEligibleActors` (`postgres_approval_repository.go:1133`) — resolves user_ids holding `required_role` in `area_code` for a tenant **as of now**, inside the caller's approval transaction.
- **Contract:** read eligible actors from **`metaldocs.v_active_user_areas`** filtered by `tenant_id`, `area_code`, `role`. Active-now is the view's responsibility; the consumer states **no** temporal predicate.
- **Source of truth:** the call site (read this session) + F3.1's published view + ADR-0039 D3(a) + ADR-0037 Model A.

## What this feature implements

Repoint `ResolveEligibleActors`'s in-tx SELECT from `metaldocs.user_process_areas` (passthrough view over
the iam base table) to `metaldocs.v_active_user_areas`, dropping the interval temporal predicate. Update the
method doc to record the view-read + active-now semantics. Drain the C3 ledger entry; re-green the cilint suite.

## Non-goals (mandatory)

- No change to the method signature, the `db.Tx` in-tx posture, the scan loop, or the never-nil slice contract.
- No new authz-recording read, lock, or tx-structure change (H-PRE-1 / D5 — the whole point).
- No reinterpretation of active-now as an interval; no exposure of temporal columns to the consumer.
- No touch to C1/C2 (F3.2, done) or C4 (search, M4); no change to the `metaldocs.user_process_areas` passthrough view itself.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| `ResolveEligibleActors` view-form set == raw interval-form set on a real `*sql.Tx`, all scopes | `TestResolveEligibleActors_ViewParityWithRaw` in `internal/modules/documents/approval/repository/eligible_actors_view_parity_integration_test.go` | real (PG :5434) |
| Discriminators: both eligibles present; revoked / wrong-role / wrong-area / wrong-tenant absent | same test's explicit assertions | real (PG :5434) |
| Parity green **before** raw deleted (D6) | `evidence.md` run order: green pre-repoint (raw==raw), repoint, green post-repoint (view==raw) | real |
| H-PRE-1 preserved — still a plain non-recording in-tx SELECT, no tx-structure change | reviewer confirmation in `evidence.md` (diff is one relation token; no lock/record added) | code review |
| Build + full test | `go build ./...`; `go test ./...` | real |
| Guard exit 0, C3 entry drained | `go run ./tools/cilint ./...` (exit 0); `hgPendingRemediation` no longer lists the approval site | real |
| `git grep user_process_areas` under `documents/approval/` → 0 production reads | `git grep -n user_process_areas -- internal/modules/documents/approval/` (excluding the parity test's verbatim baseline) | real |
| Cilint unit suite green (fixture already realigned in F3.2) | `go test ./tools/cilint/...` | real |

> TDD: parity test seeded + asserted; green pre-repoint (raw==raw sanity) then green post-repoint (view==raw).
> Real PG only; HS-3 if down. **HS-PRE-1 (critical):** if the only way to keep the set were to add a
> recording read or restructure the tx, STOP and report — do not symptom-patch.

## ADR needed?

- [x] No — uses F3.1's view under ADR-0039 D3(a); the predicate-equality rests on the existing ADR-0037 Model A. No new durable decision.
