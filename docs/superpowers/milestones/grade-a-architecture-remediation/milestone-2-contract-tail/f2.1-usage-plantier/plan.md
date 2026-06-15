# Feature F2.1 — Plan (the "how")

> Spec: `spec.md` (Approved 2026-06-14, scope A = contract-only). Judge against the spec, not this file.

## Approach

Contract-first, align spec → live runtime emit. One schema property added to `UsageSnapshot`; regen the
Go server type; no handler, no FE-render, no casing change. FE `gen:api` is the milestone-batched single
regen (run after F2.1–F2.3, gated on HS-2) — **not** in this feature.

## Steps (each with its verify)

1. **TDD red.** Run the OpenAPI assertion (`planTier in UsageSnapshot.properties`) + `api-lint -strict`.
   → verify: assertion **fails**, api-lint may already pass (it lints declared shape, not emit) — record both as the red baseline.
2. **Edit `api/openapi/v1/openapi.yaml`** `UsageSnapshot` (≈line 4799): add
   ```yaml
   planTier:
     type: string
     enum: [free, pro, enterprise]
     nullable: true
   ```
   not added to `required`. → verify: YAML parses; assertion now **passes**.
3. **Regen Go server type:** `go generate ./internal/modules/iam/api/...`.
   → verify: `internal/modules/iam/api/api.gen.go` `UsageSnapshot` gains a nullable `PlanTier *string` (or generated enum type). `git diff` touches only that gen file.
4. **Build + lint:** `go build ./...`; `go vet ./internal/modules/iam/...`; `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .`.
   → verify: build 0, vet 0, api-lint **0 violations**.
5. **Touched-pkg tests:** `go test ./internal/modules/iam/... -p 2`.
   → verify: ok (no behavior change expected; confirms regen didn't break compile/tests).
6. **Runtime emit proof (Docker up):** `.\scripts\start-api.ps1 -Build`, login, `GET /iam/usage` authed.
   → verify: response body carries a `planTier` key (enum value or `null`). Record the actual JSON.
7. **Field-truth table** into `evidence.md`: runtime emit ↔ spec ↔ gen server type all agree on `planTier`.

## Out of this feature (per spec non-goals)

- FE single `gen:api` + `tsc 0` → milestone-batched, HS-2-gated.
- `usageToJSON` raw-map refactor → M3.
- camelCase→snake_case normalization → deferred observation.

## Rollback

Single-property additive schema change + regen. Revert = drop the property + re-`go generate`. No data, no migration, no FE consumer → zero blast radius beyond the regenerated type.
