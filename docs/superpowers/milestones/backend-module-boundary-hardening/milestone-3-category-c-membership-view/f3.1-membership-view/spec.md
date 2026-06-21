# Feature F3.1 — Spec: iam publishes `metaldocs.v_active_user_areas`

> **Milestone:** 3 — Category C: published active-membership view + consumption  ·  **Folder:** `f3.1-membership-view`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-21 / leandrotca (operator) — via the operator-approved mission §7 M3 + ADR-0039 D3(a)/D4 contract; consumer-contract derivation below confirms the exact view shape. *No implementation begins until this line is filled — it is.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the feature
> must do and how it will be proven — not how it will be built (that is `plan.md`). The milestone-validator
> judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

Engine: `superpowers:brainstorming` unavailable as a subagent here → ran the consumer-contract discovery
dialog inline against the three M3 consumer call sites (read in source this session). The view shape is
**derived from the consumers**, not invented; the mission's `v_active_user_areas` is a working title only.

| # | Question | Answer |
|---|----------|--------|
| 1 | Which columns must the view expose? | Read from consumers: C1/C2 (`controlleddocuments/infrastructure/repository.go:158,500`) correlate the membership leg on `(tenant_id, user_id, area_code)`. C3 (`documents/approval/.../postgres_approval_repository.go:1135-1142`) projects `user_id` and **filters by `role`** (`role = $3`). Union ⇒ **`(tenant_id, user_id, area_code, role)`**. `role` is required by C3. |
| 2 | What "active" predicate does the view encode? | C1/C2 already gate on `effective_to IS NULL` (canonical Model-A, ADR 0037 D1). View encodes **exactly `effective_to IS NULL`** — ADR-0039 D4 mandates this and forbids interval reinterpretation. |
| 3 | C3 currently uses `effective_from <= now() AND (effective_to IS NULL OR effective_to > now())` — does moving it to an `effective_to IS NULL` view change its result set? | **No, under Model A.** ADR 0037: write path always sets `effective_from = now()` (so `effective_from <= now()` is always true for any existing row) and `effective_to = NULL`; revoke stamps a *past* `effective_to`; the set `effective_to > now()` is empty; the partial unique indexes *define* active = `effective_to IS NULL`. So C3's interval form and `effective_to IS NULL` select the identical set. Proven empirically in F3.3's parity test (incl. a seeded revoked row both exclude). This is a seam change, not a behavior change. |
| 4 | Should the view expose `effective_from` / `effective_to` so consumers keep applying their own predicate? | **No.** Encoding active-now *is* the view's contract (the reason the existing `metaldocs.user_process_areas` 1:1 passthrough is **not** a published contract and is still a guard violation). Exposing the temporal columns would invite re-deriving the predicate. Consumers drop all temporal logic and read membership = "row present in the view". |
| 5 | Which schema owns the view, and does it change RLS/tenancy posture? | `metaldocs` schema (consistent with the existing `metaldocs.user_process_areas` exposure view). The new view is defined with the **same security posture** as that existing view (no `security_invoker` clause) so RLS behavior over `public.user_process_areas` is identical. Consumers already filter `tenant_id` explicitly; parity tests assert wrong-tenant exclusion. |
| 6 | Does F3.1 repoint any consumer? | **No.** F3.1 publishes the producer only (the view migration + its self-parity proof). F3.2 (CD) and F3.3 (approval) are the consumer features. This keeps the producer landing isolated and parity-provable on its own. |

## Consumer contract (FIRST — before any producer)

What depends on this feature, and the exact shape that dependency requires.

- **Consumer(s):** three M3 membership reads, repointed in F3.2/F3.3 —
  - C1 `controlleddocuments/infrastructure/repository.go:150` — `ListControlledDocuments` restricted-visibility EXISTS leg (off-tx), correlates `upa.tenant_id = cd.tenant_id AND upa.user_id = $actor AND upa.area_code = cdag.area_code`.
  - C2 `controlleddocuments/infrastructure/repository.go:492` — `CanRead` restricted-visibility EXISTS leg (off-tx), same correlation.
  - C3 `documents/approval/repository/postgres_approval_repository.go:1136` — `ResolveEligibleActors(ctx, tx, …)` in-tx, `SELECT user_id … WHERE tenant_id=$1 AND area_code=$2 AND role=$3`.
- **Contract:** a published view **`metaldocs.v_active_user_areas`** with columns **`tenant_id uuid, user_id text, area_code text, role text`**, containing **exactly one row per active-now membership** — i.e. the rows of `public.user_process_areas` where `effective_to IS NULL`. No temporal columns. Stable, versioned name (a published read contract per ADR-0039 D3(a)). Set-equality with the base-table active-now projection is the binding invariant.
- **Source of truth for the contract:** the three consumer call sites (read this session) + **ADR-0037 D1** (active ⟺ `effective_to IS NULL`) + **ADR-0039 D3(a)/D4** (published-view mechanism; this exact view named). Base table DDL: `db/baseline/0001_current_schema.sql:1634`.

## What this feature implements

A forward-only migration `db/migrations/0242_iam_v_active_user_areas_view.sql` that creates
`metaldocs.v_active_user_areas` projecting `(tenant_id, user_id, area_code, role)` from
`public.user_process_areas WHERE effective_to IS NULL`, with a `COMMENT ON VIEW` recording it as the
ADR-0039 D3(a) published active-membership contract, and the mandatory `public.schema_migrations` row.
ADR-0039 D4 already names this view; F3.1 annotates ADR-0039's "Related code" with the concrete migration
(no new ADR — the durable decision was made in M0/F0.1; this is its implementation).

## Non-goals (mandatory)

- **No consumer repoint.** C1/C2 (F3.2) and C3 (F3.3) are not touched here.
- **No change to `public.user_process_areas`** (the base table) or the existing `metaldocs.user_process_areas`
  passthrough view (other RLS/exposure code depends on it).
- **No interval / Model-B semantics.** The view is `effective_to IS NULL` only; `effective_to > now()` stays
  refuted (ADR 0037 D2). No `effective_from`/`effective_to` columns exposed.
- **No `hgcrossmodule` guard or ledger edit.** The view name is unknown to the guard's owned-table map → reads
  of it are compliant automatically; the ledger drains happen in F3.2/F3.3 when the *base-table* reads are
  deleted.
- **No speculative columns** (e.g. `granted_by`, `effective_from`) — only the four the consumers consume.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Migration applies cleanly in the full bootstrap (baseline + all `db/migrations`) on the test PG | `testdb.Open` bootstrap inside the parity test (harness runs `ApplyCuratedBootstrap`); plus `go test -tags integration ./tests/integration/testdb/...` green | real (PG :5434) |
| View exists and returns **exactly** the base-table active-now set | `TestActiveUserAreasView_ParityWithBaseActiveNow` in `internal/modules/iam/infrastructure/postgres/active_user_areas_view_parity_integration_test.go`: `SELECT tenant_id,user_id,area_code,role FROM metaldocs.v_active_user_areas ORDER BY …` == `… FROM public.user_process_areas WHERE effective_to IS NULL ORDER BY …`, over a seeded set incl. **active row, revoked (past `effective_to`) row, two roles same user/area-set, wrong-tenant row** | real (PG :5434) |
| View encodes active-now (revoked row excluded; only `effective_to IS NULL` present) | same test asserts the revoked seeded row is **absent** from the view and the active rows are **present** | real |
| Build unaffected | `go build ./...` exit 0 | real |
| Guard unchanged (no ledger edit in F3.1) | `go run ./tools/cilint ./...` exit 0; `git diff` shows no change to `hgcrossmodule.go` | real |

> TDD: the parity test is written **first** and is RED (view does not exist → query errors / undefined
> relation) before the migration is added; GREEN after. Real PG only — if :5434 is down, mark **not-run
> (HS-3)**, never false-green.

## ADR needed?

- [x] No *new* durable decision — the decision (published-view mechanism + this view's contract) was made in
  **M0/F0.1, ADR-0039 D3(a)/D4** and ADR-0037 D1. F3.1 **annotates** ADR-0039's "Related code" line with the
  concrete migration path. Link: [`wiki/decisions/0039-cross-module-base-table-read-boundary.md`](../../../../../wiki/decisions/0039-cross-module-base-table-read-boundary.md).
