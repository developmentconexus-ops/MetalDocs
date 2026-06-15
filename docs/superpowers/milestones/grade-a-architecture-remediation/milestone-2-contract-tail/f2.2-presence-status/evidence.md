# Feature F2.2 — Evidence

> **Milestone:** 2  ·  **Feature:** `f2.2-presence-status`  ·  **Closed:** 2026-06-14
> **Contract:** `spec.md` (scope A = contract-only; align contract to live emit; `status` optional, snake-safe).

## What was implemented

Closed the H-D drift on the admin-overview `presence[]` response: the handler emitted a `status` value the
OpenAPI contract never declared on `OnlinePresenceItem`. Producer now matches the (already-correct) consumer
contract:

- **`api/openapi/v1/openapi.yaml`** — `OnlinePresenceItem` gains `status` (`type: string`, `enum: [online, idle]`, **not** in `required`). Mirrors `PresenceStreamItem.status` (line 4101) except cardinality (optional here, because the handler emits it only in the `presence != nil` branch). Shared schema → both `PresenceSnapshotResponse.items` and `AdminOverviewResponse.presence` gain the optional field.
- **`internal/modules/iam/api/api.gen.go`** — regenerated (`go generate`): new `OnlinePresenceItemStatus` enum type + consts `Online`/`Idle` + `Valid()`, struct field `Status *OnlinePresenceItemStatus json:"status,omitempty"` (optional). (Base64 churn = embedded-spec blob re-gzip, expected.)
- **`internal/modules/iam/delivery/http/admin_handler_test.go`** — added `stubPresenceReader` + `TestHandleAdminOverview_PresenceCarriesStatus`: drives the `presence != nil` branch via `WithPresenceReader` and asserts the emitted JSON `presence[0].status == "online"`. Characterization/guard test for the contract↔emit agreement the live runtime could not show (no online users → empty `presence[]`). **No handler change** — the emit is pre-existing.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — red then green | `python -c "...OnlinePresenceItem...'status' in properties, enum==[online,idle], not in required"` | RED: `status present: False` → AssertionError → GREEN: `PASS: status optional, enum ['online','idle']` | real |
| Contract gate (green-stays-green guard) | `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` | before **0**, after **0** violations (`status` is snake_case; lint checks declared shape) | real |
| Static build | `go build ./...` | exit 0 | — |
| Static vet | `go vet ./internal/modules/iam/...` | exit 0 | — |
| Regen correctness | `go generate ./internal/modules/iam/api/...` + grep | `Status *OnlinePresenceItemStatus json:"status,omitempty"` (optional); consts `Online="online"`, `Idle="idle"`; `Valid()`; only `openapi.yaml`+`api.gen.go` changed by regen | real |
| Positive-emit test (presence-wired branch) | `go test ./internal/modules/iam/delivery/http/ -run TestHandleAdminOverview_PresenceCarriesStatus -v` | `--- PASS` — `presence[0].status == "online"` | real |
| Targeted regression | `go test ./internal/modules/iam/... -p 2` | all `ok` (8 pkgs; delivery/http 2.661s) — existing `DropsUsersField`/`TenantIsolation`/`RunsInParallel` still green | real |
| Runtime emit proof (Docker/API up `:8081`) | `start-api.ps1 -Build`, cookie login, `GET /api/v1/iam/admin/overview` | `200`; `"presence":[]` (no online users in dev → else/legacy branch, `status` omitted). Confirms the **optional** cardinality live; positive emit proven by the handler test above | real |
| Focused-slice precheck (H-D) | `presenceOut` map keys vs `OnlinePresenceItem` properties | presence-branch keys `{user_id, username, display_name, last_seen_at, status}` all declared; else-branch keys (4) all declared (status omitted = optional) → **0 emitted-but-undeclared fields** on `presence[]` | real |

**Field-truth table (tri-source) — all agree on `status`:** runtime emit `string(item.Status)` → `"online"/"idle"` (handler test) ↔ spec `string enum[online,idle]`, optional ↔ gen Go `*OnlinePresenceItemStatus omitempty` ↔ FE consumer `status?: PresenceStatus` (`"online"|"idle"`, `usePresenceStream.ts:16-17`) — *FE generated type gains `status?` at the milestone-batched `gen:api` (HS-2-gated); the FE shim already matches the shape.*

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| `OnlinePresenceItem` declares `status` (string enum online/idle, not required); YAML parses | yes | TDD green row |
| Generated server type has optional `Status` (pointer/omitempty); build clean | yes | regen + static build rows |
| `api-lint -strict` 0 violations | yes | contract-gate row (0→0; status snake-safe) |
| Field-truth table agrees across runtime ↔ spec ↔ codegen ↔ FE | yes (FE gen deferred to milestone regen; shim already matches) | field-truth table |
| FE generated `OnlinePresenceItem` gains optional `status`; `tsc` 0 | **deferred** — milestone-level, single `gen:api` after F2.1–F2.3, HS-2-gated | per spec (milestone gate, not feature) |
| `presence[]` emits no other undeclared field | yes | focused-slice precheck row |

## Review disposition

- **Spec-compliance + code-quality review** (independent `caveman:cavecrew-reviewer` subagent, read-only, over the 3-file diff): **NO FINDINGS.** Verified all 7 contract facts — `status` not in `required` (optional, matches conditional emit), enum `[online,idle]` exact, generated struct `*OnlinePresenceItemStatus omitempty`, generated consts `Online`/`Idle` exact, handler emits status only in `presence != nil` branch, new test exercises the presence-wired branch via `WithPresenceReader`, test asserts JSON `status == "online"`. No scope creep beyond contract-only (no handler/FE/behavior change).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| FE `gen:api` + `tsc 0` for `status`, and FE shim removal (`usePresenceStream.ts:17` `& { status?: PresenceStatus }`) | Milestone-level single regen (F2.1–F2.3 batched); HS-2 (FE eigenpal `file:` path) must clear first; after regen the generated `OnlinePresenceItem` carries `status?` → shim redundant | Milestone M2 close gate; trigger = HS-2 resolved → run single `gen:api`, then drop the intersection, `tsc 0` |
| `presenceOut` raw-map → emit-from-generated-type | H-D is *declared-closed* here; hard-prevent re-drift = serialize from gen type (M3 raw-map pattern) | Next touch of `admin_handler.go` presence serialization, or M3 |
| `else`/legacy `onlineUsers` branch never emits `status` | Behavior question (retire legacy path) outside the H-D contract scope; the optional cardinality correctly describes today's conditional emit | Presence-reader rollout completion, or M3 |
