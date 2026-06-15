# Feature F2.1 — Evidence

> **Milestone:** 2  ·  **Feature:** `f2.1-usage-plantier`  ·  **Closed:** 2026-06-14
> **Contract:** `spec.md` (scope A = contract-only; align contract to live emit, snake_case canon).

## What was implemented

Closed the H-D drift on `/iam/usage`: the handler emitted a `plan_tier` value the OpenAPI contract never
declared. Producer now matches the (corrected) consumer contract:

- **`api/openapi/v1/openapi.yaml`** — `UsageSnapshot` gains `plan_tier` (`type: string`, `enum: [free, pro, enterprise]`, `nullable: true`, **not** in `required`). Matches the live emit exactly.
- **`internal/modules/iam/delivery/http/observability_handler.go:105`** — emit key `"planTier"` → `"plan_tier"` (one line) to satisfy the project's blocking snake_case `api-lint` canon and keep contract↔emit consistent. Safe: zero pre-existing consumers (verified, see below).
- **`internal/modules/iam/api/api.gen.go`** — regenerated (`go generate`): new `UsageSnapshotPlanTier` enum type + `Valid()`, struct field `PlanTier *UsageSnapshotPlanTier json:"plan_tier,omitempty"` (nullable, optional). (Base64 churn = embedded-spec blob re-gzip, expected.)

**Not committed yet** — held for operator (no-merge / commit-on-ask). Diff staged in working tree.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — red then green | `python -c "...UsageSnapshot...'plan_tier' in properties..."` | RED: `planTier present: False` (props seats/storage/api_calls/active_users) → GREEN: `plan_tier declared, optional` | real |
| Contract gate (was the failing-first canon check) | `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` | RED on first attempt: `4819 CASING-DRIFT: "planTier" is not snake_case` (1 violation) → after snake_case fix: **0 violation(s)** | real |
| Static build | `go build ./...` | exit 0 | — |
| Static vet | `go vet ./internal/modules/iam/...` | exit 0 | — |
| Regen correctness | `go generate ./internal/modules/iam/api/...` + grep | `PlanTier *UsageSnapshotPlanTier json:"plan_tier,omitempty"` (nullable, snake_case) | real |
| Targeted test | `go test ./internal/modules/iam/... -p 2` | all `ok` (8 pkgs; delivery/http 2.74s) — incl. `observability_service_test` asserting domain `PlanTier` | real |
| Consumer-safety | grep FE `.ts/.tsx` for `planTier`/`plan_tier`; grep `_test.go` for JSON key | FE data-read: **0**; tests assert `got.PlanTier` (domain field), not JSON string → rename safe | real |
| Runtime emit proof (Docker/API up `:8081`) | `start-api.ps1 -Build`, cookie login, `GET /api/v1/iam/usage` | `{"active_users":…,"api_calls":…,"plan_tier":"pro","seats":…,"storage":…}` — `plan_tier` key present, enum value `"pro"` | real |
| Focused-slice precheck (H-D) | `usageToJSON` emitted keys vs `UsageSnapshot` properties | both = `{seats, storage, api_calls, active_users, plan_tier}` → **0 emitted-but-undeclared fields** on `/iam/usage` | real |

**Field-truth table (tri-source) — all agree on `plan_tier`:** runtime `"pro"` (string-enum) ↔ spec `string enum[free,pro,enterprise] nullable` ↔ gen Go `*UsageSnapshotPlanTier omitempty` ↔ FE gen type *(pending milestone-batched `gen:api`, HS-2-gated — deferred per spec, milestone-level gate)*.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| `UsageSnapshot` declares `plan_tier` (string enum, nullable, not required); YAML parses | yes | TDD green row |
| Generated server type has nullable `PlanTier`; build clean | yes | regen + static build rows |
| `api-lint -strict` 0 violations for `/iam/usage` | yes | contract-gate row (1→0) |
| Field-truth table agrees across runtime ↔ spec ↔ codegen ↔ FE | yes (FE deferred to milestone regen) | field-truth table |
| FE gen `UsageSnapshot` gains optional `plan_tier`; `tsc` 0 | **deferred** — milestone-level, single `gen:api` after F2.1–F2.3, HS-2-gated | per spec (milestone gate, not feature) |
| `/iam/usage` emits no other undeclared field | yes | focused-slice precheck row |

## Review disposition

- **Spec-compliance + code-quality review** (independent subagent, read-only, over the 3-file source diff): **no findings.** Confirmed contract↔emit exact match (string enum free/pro/enterprise, nullable, optional), zero consumer breakage (no FE data-read, no test asserts JSON key), no enum-const collision in `iam/api`, scope strictly contract-only.
- One execution-time contract correction (camelCase→snake_case) recorded in `spec.md` non-goals as superseded-by-the-blocking-canon — not scope drift.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| FE `gen:api` + `tsc 0` for `plan_tier` | Milestone-level single regen (F2.1–F2.3 batched), and HS-2 (FE eigenpal `file:` path) must clear first | Milestone M2 close gate; trigger = HS-2 resolved → run single `gen:api` |
| `usageToJSON` raw-map → emit-from-generated-type | H-D is *declared-closed* here; hard-prevent re-drift = serialize from gen type (M3 raw-map pattern, cf. F3.3) | Next touch of `observability_handler.go` serialization, or M3 |
