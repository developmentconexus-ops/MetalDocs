# Feature F3.3 — Evidence

> **Milestone:** 3 · **Feature:** `f3.3-cd-raw-map-to-type` · **Closed:** 2026-06-14
> **Contract:** `spec.md` (consumer = generated `AtomicCreateResponse`, already consumed by the FE).

## What was implemented

Replaced the raw `map[string]any` 201 response at
`internal/modules/controlleddocuments/delivery/http/routes.go:123` with the generated
`controlleddocumentsapi.AtomicCreateResponse`, built from the **existing** domain→api mappers
(`controlledDocumentResponse`, `documentRefResponse`) already used by Get/List. Mapper errors mirror
the canonical Get pattern (`WriteError(500, "INTERNAL_ERROR", …)`). One test-only fixture change: the
spy `Create` now returns UUID-valid data (and gained a `createResult` override) so the typed path's
`uuid.Parse` succeeds.

**Net wire effect (the drift-fix):** absent optionals `department_code`,
`override_template_version_id`, `sequence_num` are now **omitted** instead of serialized as `null` —
aligning the producer with the already-declared `,omitempty` contract the FE already types. Required
fields, present optionals, and the `document` object are byte-unchanged.

## Verification

| Gate | Command | Result (real output) |
|------|---------|----------------------|
| G1 (TDD red→green) | `go test ./internal/modules/controlleddocuments/delivery/http/ -run TestAtomicCreate_UsesGeneratedResponse -count=1` | **RED before impl** — body showed `"department_code":null,"sequence_num":null,"override_template_version_id":null` → `FAIL` at `routes_contract_test.go:603`. **GREEN after impl.** |
| G2 (no raw map at site) | `grep -n 'map\[string\]any' routes.go` | only `:89` + `:224` (request-side `formData` decode targets); the `:123` 201-response map is gone. |
| G3 (module tests) | `go test ./internal/modules/controlleddocuments/... -count=1` | `ok` application / delivery/http / domain / infrastructure. `TestAtomicCreate_ForwardsGeneratedOnlyFields` (201) stays green through the now-typed path via the UUID-valid fixture. |
| G4 (build + codegen) | `go build ./...`; `go generate ./internal/modules/controlleddocuments/api/...` | build clean. Regen touched **only** the embedded gzip-base64 `swaggerSpec` blob (api.gen.go ~`:2043`) — **zero** Go type/signature changes (`AtomicCreateResponse` at `:135` and `ControlledDocument`/`DocumentRef` types unchanged). That blob churn is pre-existing codegen-freshness drift unrelated to F3.3 (no OpenAPI/api-type edit in this feature) → **reverted** `git checkout -- api.gen.go`; recorded as a bounded defer. |
| G5 (spec lint) | `npx @redocly/cli lint api/openapi/v1/openapi.yaml` | `Your API description is valid. 🎉` (no OpenAPI edit). |
| G6 (FE consumer) | read `frontend/apps/web/src/features/controlled-documents/api/controlledDocuments.ts:31-44` | already `apiFetch<AtomicCreateResponse>` where `AtomicCreateResponse = components["schemas"]["AtomicCreateResponse"]`; optionals typed `?` (present-or-absent). `null`→omitted is **more** type-accurate; no regen, no FE code change. |

Real vs fixture: tests use the in-process `spyControlledDocumentService` (handler-level contract tests,
no DB) — labeled fixture. The wire-shape assertions (G1) are real serialization of the production
handler + generated types via `httptest`.

## Acceptance vs spec Validation Gate

All six gates (G1–G6) met — see table. The one contract nuance (`null`→omitted on absent optionals)
is a drift-*fix* toward the already-declared optional contract, not a contract change: **HS-2 did not
trip** (no OpenAPI edit, no FE regen, FE optional-typed fields already expect omission).

## Review disposition

- Spec-compliance: producer now emits the exact generated `AtomicCreateResponse` the FE consumer
  types; consumer-contract-first honored (contract read from FE + OpenAPI before the swap).
- Code-quality: reused existing mappers (no new mapper, no duplication); error handling mirrors the
  sibling Get handler; surgical — only the one response site + a test-only fixture changed.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `go generate` rewrites the embedded `swaggerSpec` gzip-blob in CD `api.gen.go` (no type change) | Pure codegen-freshness drift in the embedded spec bytes — unrelated to F3.3 (no OpenAPI/api-type edit here); reverting keeps the feature surgical. No runtime type/contract impact. | Codegen-freshness micro-task (regenerate all module `api.gen.go` from a single pinned oapi-codegen + spec serialization), or fold into M4/wiki-curator; owner: backend agent. Likely a generator-version/serialization difference vs the committed blob. |

## Memory / cross-refs

H-D class context: this closes a producer-side optional-nullability drift (the inverse of M2's
undeclared-field drift) on the CD AtomicCreate response. Related: `[[backend-standardization-parameters]]`.
