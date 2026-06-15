# Feature F0.1 — Spec

> **Milestone:** 0 — Auth / authz / session correctness  ·  **Folder:** `f0.1-authz-effective-from`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-15 / leandrotca.work (effective_from-only; effective_to deliberately out of scope) — *implementation may begin.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Is the contract ambiguous? | **None needed.** Mission §5 B1 + §7 F0.1 lock the change ("`Require` honors `effective_from <= now()`, matching `ResolveEligibleActors`"). The canonical predicate is verified at `internal/modules/documents/approval/repository/postgres_approval_repository.go:1140`. Column `user_process_areas.effective_from` is `TIMESTAMPTZ NOT NULL` (migration 0125, part of PK) so `<= now()` needs no NULL handling. |
| 2 | Also align the `effective_to` half (`effective_to IS NULL` → `(effective_to IS NULL OR effective_to > now())`)? | **No — out of scope (Non-goal).** That divergence makes `Require` *stricter* today (denies a membership scheduled to end in the future); "fixing" it would *widen* access and is **not** an F5.1-confirmed finding. Changing it is a security-direction change → HS-6 scope drift. Recorded as a flagged observation, not touched. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** every tier-2 authz caller — i.e. every code path that calls
  `authz.Require(ctx, tx, capability, areaCode)` inside its transaction (controlled-documents create,
  approval submit/decision, document publish, etc.). They rely on `Require` to encode the *single*
  definition of "does this actor hold this capability in this area right now".
- **Contract:** `Require` returns `nil` (grant) **iff** the actor is `system_admin` (existing bypass,
  unchanged) **or** holds an **active** `user_process_areas` membership whose role carries the
  capability in the area, where **active** means `effective_from <= now()` **and** the membership is
  not ended (current behavior: `effective_to IS NULL`). Otherwise it returns `ErrCapDenied`. The new
  clause is `effective_from <= now()`: a **future-dated** membership must be **denied**.
- **Source of truth for the contract:** the canonical eligibility predicate in
  `ResolveEligibleActors` (`postgres_approval_repository.go:1140-1141`) and ADR 0022 (two-tier authz).
  The fix makes `Require`'s tier-2 query agree with that canonical predicate on the `effective_from`
  dimension.

## What this feature implements

Add `AND upa.effective_from <= now()` to the capability-grant query in `authz.Require`
(`internal/modules/iam/authz/authz.go:115-126`), at the **shared predicate** — not per-caller. No other
behavior changes: the system_admin bypass, the cap-cache, the asserted-caps write, and the
`effective_to IS NULL` clause are untouched.

## Non-goals (mandatory)

- **Not** altering the `effective_to` clause (see interview Q2) — no `effective_to > now()` addition.
- **Not** changing the system_admin bypass, cap-cache, asserted-caps, or any GUC handling.
- **Not** touching any caller of `Require` (the whole point is a single shared-predicate fix).
- **No** schema/migration change — `effective_from` already exists, NOT NULL.
- **No** change to `ResolveEligibleActors` (it is the reference, already correct).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| A **future-dated** membership (effective_from > now) granting the cap → `Require` returns `ErrCapDenied` | `TestRequire_FutureDatedMembershipDenied` (new, `//go:build integration`, real Postgres via `tests/integration/testdb`) | **real** (real DB; the sqlmock unit driver cannot exercise the SQL predicate) |
| A **current** membership (effective_from <= now) granting the cap → `Require` returns `nil` (grant) | `TestRequire_CurrentMembershipGranted` (new, same integration file) | **real** |
| No regression in existing authz behavior | `go test ./internal/modules/iam/authz/...` (unit) green | fixture (sqlmock) |
| Whole-repo green | `go test ./...` green; `go test -tags integration ./internal/modules/iam/authz/...` green | mixed |

> TDD: write `TestRequire_FutureDatedMembershipDenied` first (must **fail** against current code —
> current `Require` grants the future-dated membership), then add the predicate to make it green.
> The existing sqlmock unit tests stay green (they return canned `granted` regardless of SQL) — that
> is exactly why a **real-DB** integration test is required to prove this fix; fixture proof here is
> not sufficient on its own.

## ADR needed?

- [x] No durable decision — skip. This conforms `Require` to the *existing* ADR 0022 canonical
  predicate; it does not introduce a new decision.
