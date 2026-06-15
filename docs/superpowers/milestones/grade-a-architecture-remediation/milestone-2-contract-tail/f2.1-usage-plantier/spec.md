# Feature F2.1 — Spec

> **Milestone:** 2 — Contract Tail (H-D class)  ·  **Folder:** `f2.1-usage-plantier`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-14 / leandrotca.work (operator) — scope decided "Contract-only (kill H-D)".

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | The governing spec says F2.1 "unblock `UsageGauges.tsx` / gauge renders a real value", but the consumer read shows FE does **not** consume `planTier` — `UsageGauges.tsx:274-279` renders a hardcoded `UnavailableCard` placeholder and never reads `data.planTier`. The handler already emits `planTier`. Does F2.1 = (A) contract-only (declare the field, kill H-D, leave FE placeholder) or (B) contract + FE wire-up (replace the UnavailableCard with a real plan-tier card)? | **(A) Contract-only.** Declare `planTier` in OpenAPI + regen the server type; close the H-D drift. FE wiring stays out of scope (a separate UX feature). Operator-selected 2026-06-14. |

**Why no further interview:** with (A) chosen + the program's "align spec to the live runtime/consumer" direction, the producer is fully determined — declare what the handler already emits. No remaining contract ambiguity.

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** the generated FE type `UsageSnapshot` (from `lib/api-types/`, surfaced through `useUsageQuery`); the OpenAPI contract `UsageSnapshot` is itself the contract consumers (FE codegen, any API client) read. There is **no runtime FE field-consumer of `planTier`** today (the screen shows a placeholder) — so the consumer of this contract is the **generated type + codegen pipeline**, which must gain `planTier` so it stops being undeclared drift.
- **Contract — the exact shape the runtime already emits** (read from the producer's live output, `observability_handler.go:81-107`, `usageToJSON`):
  - `planTier`: a JSON string, one of `"free" | "pro" | "enterprise"`, **or JSON `null`** when no plan row exists (`PlanTier == ""` → emitted as `null`).
  - Field name is **`plan_tier`** (snake_case). *Contract correction during execution:* the handler originally emitted camelCase `planTier`; the project's **blocking** `api-lint` `CASING-DRIFT` rule requires snake_case property names (backend snake_case canon). Aligning *both* the contract and the live emit to `plan_tier` (rather than exempting the linter) is the canon-correct closure. Safe because **no consumer existed** — FE has no `.ts/.tsx` data-read of the key, and no Go test asserts the JSON string (tests assert the domain field `got.PlanTier`). Optional / nullable — **not** in `required`.
  - The four existing `UsageSnapshot` properties (`seats`, `storage`, `api_calls`, `active_users`) are unchanged.
- **Source of truth for the contract:** the live handler emit (`internal/modules/iam/delivery/http/observability_handler.go:81-107`) + the domain enum `PlanTier` (`internal/modules/iam/domain/observability.go:5-12`: `free`/`pro`/`enterprise`, empty→null). The OpenAPI `UsageSnapshot` schema (`api/openapi/v1/openapi.yaml:4799`) is the surface being corrected to match that emit.

## What this feature implements

Declare `planTier` on the `UsageSnapshot` response schema in `api/openapi/v1/openapi.yaml` — type `string`, `enum: [free, pro, enterprise]`, `nullable: true`, **not** required — matching exactly what the `/iam/usage` handler already emits. Regenerate the server types (`oapi-codegen`) and the FE types (batched into the milestone's single `gen:api`). After this feature, the `/iam/usage` response carries **zero** emitted-but-undeclared fields: `planTier` is now tri-source-declared (runtime ↔ spec ↔ generated server type ↔ generated FE type).

## Non-goals (mandatory)

- **No FE rendering change.** `UsageGauges.tsx` keeps its `UnavailableCard` placeholder; no plan-tier card is wired. (Operator decision A.)
- **No handler-serialization refactor.** The `usageToJSON` `map[string]any` hand-serialization stays as-is. Converting it to emit from the generated type is the raw-map→generated-type pattern owned by **M3**; doing it here is scope drift. *(Observation, not a fix: because the handler hand-builds the map, the contract and emit can re-drift in future; recorded as a defer below, not closed here.)*
- ~~**No casing normalization.**~~ **Superseded during execution:** the blocking `api-lint` snake_case canon forced `plan_tier` (not camelCase). The handler emit literal was corrected `"planTier"`→`"plan_tier"` (one line, `observability_handler.go:105`) to keep contract↔emit consistent. This stays within contract-only scope (no FE render, no map refactor) and was safe (zero pre-existing consumers, verified). Not scope drift — required by the contract-truth gate.
- **No other route or schema touched.** Only `UsageSnapshot` gains one property. F2.2 (`OnlinePresenceItem.status`) and F2.3 (`/templates` envelope) are separate features; the FE `gen:api` is run **once** for all three (not in this feature).
- **No new endpoint, no plan-tier source change.** The value comes from the existing `PlanTier` domain field; this feature does not change how it is computed.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| `UsageSnapshot` in OpenAPI declares `planTier` (string, enum free/pro/enterprise, nullable, not required); YAML parses | `python -c "import yaml,sys; s=yaml.safe_load(open('api/openapi/v1/openapi.yaml'))['components']['schemas']['UsageSnapshot']; assert 'planTier' in s['properties']; assert 'planTier' not in s.get('required',[])"` → exit 0 | real |
| Generated server type for `UsageSnapshot` has a nullable `PlanTier` field; build clean | `oapi-codegen` regen (canonical command) → `go build ./...` exit 0; grep generated type shows `PlanTier *…` | real |
| `api-lint -strict` reports **0** violations for `/iam/usage` (field no longer undeclared-drift) | `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` → 0 violations | real |
| Field-truth table: runtime emit ↔ spec ↔ generated server type ↔ generated FE type all agree on `planTier` (string-enum-or-null, optional) | route/field-truth table recorded in `evidence.md`; runtime `GET /iam/usage` (Docker up, authed) returns a `planTier` key (value or `null`) | real |
| FE generated `UsageSnapshot` type gains optional `planTier`; FE `tsc` 0 after the **milestone** single regen | `pnpm --filter web gen:api` (milestone-batched) + `pnpm --filter web tsc` → 0 | real (milestone-level) |
| Focused-slice precheck: `/iam/usage` emits no other undeclared field besides `planTier` (now declared) | diff `usageToJSON` keys vs `UsageSnapshot` properties → only `planTier` was the gap, now closed | real |

> TDD note: the OpenAPI assertion + the `api-lint -strict` pass are the failing-first checks (both fail
> before the schema edit, pass after). Runtime emit proof is observed live (Docker up), labeled real.
> The FE `tsc` 0 is a **milestone-level** gate — the single `gen:api` runs once after F2.1–F2.3 land.

## ADR needed?

- [x] No durable decision — skip. This aligns the contract to an existing runtime emit; no architectural choice. (The deferred handler-serialization-refactor and casing observations are tracked below, not ADR-grade.)

## Deferred observations (recorded, not closed here — none block F2.1)

- **Re-drift risk:** `usageToJSON` hand-builds `map[string]any`; declaring the field does not force the handler to serialize from the generated type, so a future field could drift again. Root-cause-hard fix = emit from generated type (M3 raw-map→generated-type pattern, cf. F3.3). Trigger: next touch of `observability_handler.go` serialization, or M3.
- ~~Casing inconsistency~~ — **resolved in this feature** (declared + emitted as snake_case `plan_tier`; api-lint clean).
