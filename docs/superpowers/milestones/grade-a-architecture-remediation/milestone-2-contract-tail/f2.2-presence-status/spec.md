# Feature F2.2 — Spec

> **Milestone:** 2 — Contract Tail (H-D class)  ·  **Folder:** `f2.2-presence-status`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-14 / leandrotca.work (operator) — scope "Contract-only (kill H-D)", same cadence as F2.1.

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | F2.1 precedent set the cadence (contract-only, align contract to live emit, FE wiring out of feature). The admin-overview handler emits `status` on presence items *conditionally* (only the `h.presence != nil` branch, `admin_handler.go:232`); the `else` branch omits it. The FE already **consumes** `status` — but via a hand-shim `OnlinePresenceItem & { status?: PresenceStatus }` (`usePresenceStream.ts:17`) because the generated type lacks it. Does F2.2 = (A) contract-only (declare `status` optional on `OnlinePresenceItem`, kill the H-D drift, leave the FE shim) or (B) contract + drop the now-redundant FE shim? | **(A) Contract-only.** Declare `status` (optional) on `OnlinePresenceItem` + regen the server type; close the H-D drift. The FE shim becomes *redundant* (not broken) and its removal is milestone-batched FE cleanup, done with the single `gen:api` after F2.1–F2.3. Mirrors the F2.1 operator decision. |

**Why no further interview:** with (A) chosen + the program's "align contract to the live runtime/consumer" direction, the producer is fully determined — declare what the handler already emits, with the cardinality the handler already uses (conditional → optional). No remaining contract ambiguity. `status` is snake-safe, so no casing-canon conflict (unlike F2.1).

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** (1) the generated FE type `OnlinePresenceItem` (from `lib/api-types/`, surfaced through `usePresenceStream`/`useAdminOverview`); the OpenAPI `OnlinePresenceItem` schema is itself the contract that FE codegen and any API client read. (2) A **real runtime FE field-consumer exists** here (unlike F2.1): `usePresenceStream.ts:17` extends the generated type with `status?: PresenceStatus` and reads `ev.status` (line 52). The shim exists *because* the generated type is missing `status` — i.e. the FE is already compensating for this exact H-D drift. Declaring `status` makes the contract honest; the shim then becomes redundant (milestone-batched removal).
- **Contract — the exact shape the runtime already emits** (read from the producer's live output, `admin_handler.go:224-244`, `presenceOut`):
  - `status`: a JSON string, one of `"online" | "idle"` (`presence.Status` enum, `internal/modules/iam/presence/model.go:49-50`).
  - **Cardinality: OPTIONAL.** The handler emits `status` only in the `h.presence != nil` branch (`admin_handler.go:227-233`); the legacy `else` branch (`onlineUsers`, lines 236-243) omits it. So a presence item may or may not carry `status` → **not** in `required`. This matches the FE shim's own `status?` (optional).
  - The four existing `OnlinePresenceItem` properties (`user_id`, `username`, `display_name`, `last_seen_at`) and the `required: [user_id, username, display_name, last_seen_at]` list are **unchanged** (status is added to `properties` only, not to `required`).
- **Mirror reference:** `PresenceStreamItem` (`openapi.yaml:4090`) already declares `status: { type: string, enum: [online, idle] }` (there it is `required`, because the WS stream always carries it). F2.2 reuses the **same property shape** on `OnlinePresenceItem` — only the cardinality differs (optional here vs required on the stream), matching each producer's actual emit.
- **Source of truth for the contract:** the live handler emit (`internal/modules/iam/delivery/http/admin_handler.go:224-244`) + the domain enum `presence.Status` (`internal/modules/iam/presence/model.go:45-50`: `online`/`idle`). The OpenAPI `OnlinePresenceItem` schema (`api/openapi/v1/openapi.yaml:4080`) is the surface being corrected to match that emit.

## What this feature implements

Declare `status` on the `OnlinePresenceItem` response schema in `api/openapi/v1/openapi.yaml` — type `string`, `enum: [online, idle]`, **not** in `required` — matching exactly what the admin-overview handler conditionally emits. `OnlinePresenceItem` is shared by `PresenceSnapshotResponse.items` and `AdminOverviewResponse.presence`; both gain the optional field (correct — the snapshot endpoint, if/when it emits, uses the same item shape). Regenerate the server types (`oapi-codegen`); FE types are batched into the milestone's single `gen:api`. After this feature, the admin-overview `presence[]` response carries **zero** emitted-but-undeclared fields: `status` is now tri-source-declared (runtime ↔ spec ↔ generated server type ↔ generated FE type).

## Non-goals (mandatory)

- **No FE shim removal in this feature.** `usePresenceStream.ts:17`'s `OnlinePresenceItem & { status?: PresenceStatus }` stays as-is for now. After the milestone's single `gen:api`, the generated `OnlinePresenceItem` will carry `status?` itself, making the intersection redundant — its removal is **milestone-batched FE cleanup** (with the regen + `tsc`), not feature scope. (Operator decision A.) Recorded as a defer below.
- **No handler-serialization refactor.** The `presenceOut` `[]map[string]any` hand-serialization in `admin_handler.go` stays as-is. Converting it to emit from the generated type is the raw-map→generated-type pattern owned by **M3**; doing it here is scope drift. *(Observation, not a fix: because the handler hand-builds the map, contract and emit can re-drift; recorded as a defer, not closed here.)*
- **No change to the `else`/legacy branch.** The `onlineUsers` path (`admin_handler.go:236-243`) keeps omitting `status` — that is precisely why the field is **optional**. Not "fixing" it to always emit status is correct: the contract must describe the real (conditional) emit, not an idealized one.
- **No required-list change.** `status` is added to `properties` only. `OnlinePresenceItem.required` stays `[user_id, username, display_name, last_seen_at]`.
- **No other route or schema touched.** Only `OnlinePresenceItem` gains one property. F2.1 (`UsageSnapshot.plan_tier`, done) and F2.3 (`/templates` envelope) are separate features; the FE `gen:api` runs **once** for all three (not in this feature).
- **No `PresenceStreamItem` change.** It already declares `status` (required); untouched.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| `OnlinePresenceItem` in OpenAPI declares `status` (string, enum online/idle, **not** in `required`); YAML parses | `python -c "import yaml; s=yaml.safe_load(open('api/openapi/v1/openapi.yaml', encoding='utf-8'))['components']['schemas']['OnlinePresenceItem']; assert 'status' in s['properties']; assert s['properties']['status']['enum']==['online','idle']; assert 'status' not in s.get('required',[])"` → exit 0 | real |
| Generated server type for `OnlinePresenceItem` gains an optional `Status` field (pointer / `omitempty`); build clean | `oapi-codegen` regen (canonical command) → `go build ./...` exit 0; grep generated type shows `Status *…OnlinePresenceItemStatus … omitempty` | real |
| `api-lint -strict` reports **0** violations (status is snake-safe; field no longer undeclared-drift) | `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` → 0 violations | real |
| Field-truth table: runtime emit ↔ spec ↔ generated server type ↔ FE consumer all agree on `status` (string-enum online/idle, optional) | route/field-truth table in `evidence.md`; runtime `GET /iam/admin/overview` (Docker up, authed, presence reader wired) returns `presence[].status` | real |
| FE generated `OnlinePresenceItem` type gains optional `status`; FE `tsc` 0 after the **milestone** single regen | `pnpm --filter web gen:api` (milestone-batched) + `pnpm --filter web tsc` → 0 | real (milestone-level) |
| Focused-slice precheck: admin-overview `presence[]` emits no other undeclared field besides `status` (now declared) | diff `presenceOut` map keys vs `OnlinePresenceItem` properties → only `status` was the gap, now closed | real |

> TDD note: the OpenAPI assertion is the failing-first check (fails before the schema edit, passes after).
> `api-lint -strict` is expected to pass both before and after (status is snake_case; it lints declared
> shape, not the undeclared-emit gap) — recorded as the green-stays-green guard. Runtime emit proof is
> observed live (Docker up, presence reader wired), labeled real. FE `tsc` 0 is a **milestone-level** gate.

## ADR needed?

- [x] No durable decision — skip. This aligns the contract to an existing runtime emit; no architectural choice. (The deferred handler-serialization-refactor and FE-shim-removal are tracked below, not ADR-grade.)

## Deferred observations (recorded, not closed here — none block F2.2)

- **FE shim redundancy:** after the milestone single `gen:api`, `usePresenceStream.ts:17`'s `OnlinePresenceItem & { status?: PresenceStatus }` becomes redundant (the generated type will carry `status?`). Remove the intersection then. Trigger: milestone M2 FE regen step (with `tsc 0`).
- **Re-drift risk:** `presenceOut` hand-builds `[]map[string]any`; declaring the field does not force the handler to serialize from the generated type, so a future field could drift again. Root-cause-hard fix = emit from generated type (M3 raw-map→generated-type pattern). Trigger: next touch of `admin_handler.go` presence serialization, or M3.
- **Legacy `else` branch:** the `onlineUsers` path never emits `status`. Whether that path should be retired (always use the presence reader) is a behavior question outside the H-D contract scope. Trigger: presence-reader rollout completion, or M3.
