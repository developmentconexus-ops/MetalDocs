# Feature F2.2 — Plan (the "how")

> Spec: `spec.md` (Approved 2026-06-14, scope A = contract-only). Judge against the spec, not this file.

## Approach

Contract-first, align spec → live runtime emit. One schema property added to `OnlinePresenceItem`
(optional, mirroring the conditional handler emit); regen the Go server type; no handler change, no
FE-shim removal, no `else`-branch change. `status` is snake-safe → no casing fix needed (unlike F2.1).
FE `gen:api` is the milestone-batched single regen (after F2.1–F2.3, gated on HS-2) — **not** in this feature.

## Steps (each with its verify)

1. **TDD red.** Run the OpenAPI assertion (`status in OnlinePresenceItem.properties`, enum `[online,idle]`,
   not required) + `api-lint -strict`.
   → verify: assertion **fails** (status absent); api-lint already passes (lints declared shape, not the
   undeclared-emit gap) — record both as the red baseline (assertion red, api-lint green-stays-green guard).
2. **Edit `api/openapi/v1/openapi.yaml`** `OnlinePresenceItem` (≈line 4083, in `properties`): add
   ```yaml
   status: { type: string, enum: [online, idle] }
   ```
   **not** added to `required`. Mirror `PresenceStreamItem.status` (line 4100) exactly except cardinality.
   → verify: YAML parses; assertion now **passes**.
3. **Regen Go server type:** `go generate ./internal/modules/iam/api/...`.
   → verify: `internal/modules/iam/api/api.gen.go` `OnlinePresenceItem` gains an optional `Status`
   (`*OnlinePresenceItemStatus` enum type + consts + `omitempty`). `git diff` touches only that gen file
   (plus the embedded-spec blob churn, expected).
4. **Build + lint:** `go build ./...`; `go vet ./internal/modules/iam/...`; `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .`.
   → verify: build 0, vet 0, api-lint **0 violations**.
5. **Touched-pkg tests:** `go test ./internal/modules/iam/... -p 2`.
   → verify: ok (no behavior change expected; confirms regen didn't break compile/tests; `admin_handler_test`
   still green).
6. **Runtime emit proof (Docker up):** `.\scripts\start-api.ps1 -Build`, cookie login, `GET /api/v1/iam/admin/overview` authed.
   → verify: response `presence[]` items carry a `status` key (`"online"`/`"idle"`) when the presence reader
   is wired. Record the actual JSON. (If presence reader unwired in dev → `status` absent, which is the
   optional-cardinality proof; note which branch was live.)
7. **Field-truth table** into `evidence.md`: runtime emit ↔ spec ↔ gen server type ↔ FE consumer all agree
   on `status` (string-enum online/idle, optional).

## Out of this feature (per spec non-goals)

- FE single `gen:api` + `tsc 0`, and FE shim removal (`usePresenceStream.ts:17`) → milestone-batched, HS-2-gated.
- `presenceOut` raw-map refactor → M3.
- `else`/legacy-branch retirement → out of H-D contract scope.

## Rollback

Single-property additive schema change + regen. Revert = drop the property + re-`go generate`. No data, no
migration; the FE shim already tolerates `status?` so no consumer break either way → zero blast radius
beyond the regenerated type.
