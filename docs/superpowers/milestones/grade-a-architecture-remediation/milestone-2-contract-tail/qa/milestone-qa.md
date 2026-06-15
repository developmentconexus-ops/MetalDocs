# Milestone 2 — Validation Verdict (C1–C7) — Contract Tail (H-D class)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-14  ·  **Verdict:** see C7 (PASS).
> Judge-only: this file is the only artifact written; no source/spec/status was edited.

## Inputs loaded (all present, all readable)

- Milestone spec: `milestone-2-contract-tail/milestone.md` ✅
- Features (spec.md + evidence.md): `f2.1-usage-plantier`, `f2.2-presence-status`, `f2.3-templates-envelope` ✅ (plan.md not required for the C1–C7 judgment; spec+evidence present and complete for all three)
- Program README: `grade-a-architecture-remediation/README.md` ✅
- Governing spec §6 M2: `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md:137-146` ✅
- Aggregate diff: `git log 4c6a1e83..HEAD` = 5 M2 commits (69ad234d, 0f0fb1ee, 4ab670b1, 728783f6, 33fbb48c) + the pre-M2 `a242e5a4` (M1 test-infra-rebaseline, not M2 scope) ✅

## C1 — Spec & plan conformance (per feature)

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F2.1 plan_tier | ✅ read from live emit (`observability_handler.go:81-107`) + domain enum; contract declares the snake_case `plan_tier` the handler emits. No FE data-read exists (scope-A placeholder) → consumer-of-contract is the codegen pipeline, correctly served. | ✅ OpenAPI declares `plan_tier` string enum[free,pro,enterprise] nullable, not required; api-lint 0; gen struct `PlanTier *UsageSnapshotPlanTier omitempty`. | ✅ no FE render change, no map-serialization refactor; camelCase→snake_case rename recorded as canon-forced, not drift. | `f2.1/evidence.md` + my C2 re-runs |
| F2.2 status | ✅ read from live emit (`admin_handler.go:224-244`, conditional branch) + `presence.Status` enum; declared optional matching the FE shim `status?: PresenceStatus` (`usePresenceStream.ts:17,52`). | ✅ OpenAPI declares `status` enum[online,idle], NOT in `required`; gen `Status *OnlinePresenceItemStatus omitempty`; named test `TestHandleAdminOverview_PresenceCarriesStatus` PASS. | ✅ no shim removal, no else-branch change, no required-list change, no `PresenceStreamItem` touch. | `f2.2/evidence.md` + my C2 re-runs |
| F2.3 templates envelope | ✅ direction (envelope) decided contract-first from the real FE consumer `templates.ts:131,141-142` (`body.data.templates`, `body.meta.{limit,offset}`); matches live emit `routes_query.go:67-75`. | ✅ 200 = `$ref ListTemplatesResponse`; requires `data.templates:[TemplateDTO]` + `meta.{limit,offset}`; `x-pagination-exempt:true` retained; gen `ListTemplates200JSONResponse ListTemplatesResponse`. | ✅ no FE re-wire, no handler refactor, no request-param fix, no pagination-model change, `TemplateDTO` reused unchanged. | `f2.3/evidence.md` + my C2 re-runs |

All three consumer contracts were **read from consumer + runtime, not guessed** — the H-D root-cause requirement. No missing spec/evidence rows. **C1 PASS.**

## C2 — Gates re-run, isolated (clean tree `git status` empty; reproduced by validator, not trusted)

| Gate | Command re-run | Real output | Pass? |
|------|----------------|-------------|-------|
| Build | `go build ./...` | exit 0 | ✅ |
| Vet | `go vet ./internal/modules/iam/... ./internal/modules/templates/...` | exit 0 | ✅ |
| API contract lint | `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)` | ✅ |
| Targeted tests (uncached) | `go test -count=1 ./internal/modules/iam/delivery/http/ ./internal/modules/templates/delivery/http/ -p 2` | both `ok` (2.79s / 2.94s) | ✅ |
| Full module tests | `go test ./internal/modules/iam/... ./internal/modules/templates/... -p 2` | all `ok` (14 pkgs) | ✅ |
| F2.2 named test | `go test -count=1 ...delivery/http/ -run TestHandleAdminOverview_PresenceCarriesStatus -v` | `--- PASS` | ✅ |
| OpenAPI F2.1 | python yaml(utf-8) assert | `plan_tier` string enum[free,pro,enterprise] nullable, not required → PASS | ✅ |
| OpenAPI F2.2 | python yaml(utf-8) assert | `status` enum[online,idle]; `required`=[user_id,username,display_name,last_seen_at] (status absent → optional) → PASS | ✅ |
| OpenAPI F2.3 | python yaml(utf-8) assert | 200=`$ref ListTemplatesResponse`; data.templates→TemplateDTO; meta requires [limit,offset]; x-pagination-exempt=True → PASS | ✅ |
| Gen struct (iam) | grep `api.gen.go` | `PlanTier *UsageSnapshotPlanTier omitempty`; `Status *OnlinePresenceItemStatus omitempty`; enum consts Online/Idle, Free/Pro/Enterprise | ✅ |
| Gen struct (templates) | read `api.gen.go:92-101,1462` | `ListTemplatesResponse{Data.Templates []TemplateDTO; Meta.{Limit,Offset int}}`; `ListTemplates200JSONResponse ListTemplatesResponse` | ✅ |
| FE typecheck (single-regen gate) | `./node_modules/.bin/tsc --noEmit -p tsconfig.build.json` (frontend/apps/web) | exit 0 | ✅ |
| FE gen types | grep `lib/api-types/index.d.ts` | `plan_tier?: "free"\|"pro"\|"enterprise"\|null`; `OnlinePresenceItem.status?: "online"\|"idle"`; `ListTemplatesResponse{data.templates:TemplateDTO[]; meta.{limit,offset}}` | ✅ |

Every milestone-deferred FE gate (the single `gen:api` + `tsc 0`) is now reproduced live and green. **C2 PASS.**

**Runtime honesty label:** the API was **not** started in this validation run. Build/vet/lint/test/typecheck/struct gates were all **reproduced-live** from a clean tree. The runtime payload rows (`GET /iam/usage` → `plan_tier:"pro"`; `/iam/admin/overview` → `presence:[]` confirming optional cardinality; `GET /templates?limit=5` → envelope) are **asserted-from-evidence** (each feature's `evidence.md`), not reproduced-live by me. The positive `status` emit is independently proven by the re-run handler test, not only runtime.

## C3 — Senior review of the aggregate milestone diff

Reviewed `4c6a1e83..HEAD` as one unit (struct-level; api.gen.go base64 blob churn ignored as benign per brief).

- **OpenAPI yaml diff** = exactly 3 surfaces: `plan_tier` added to UsageSnapshot, `status` added to OnlinePresenceItem, `/templates` 200 `array→$ref ListTemplatesResponse` + new `ListTemplatesResponse` schema. Nothing else.
- **FE index.d.ts diff** = exactly those 3 types; one route ref flipped (`TemplateDTO[]`→`ListTemplatesResponse`). **No unrelated route's contract moved.**
- **Non-generated Go diff (M2-only, 69ad234d~1..HEAD)** = one line `"planTier"→"plan_tier"` (`observability_handler.go:105`) + the F2.2 characterization test (`admin_handler_test.go`, +52). No handler success-path logic changed. (`e2e_seed.go` churn traced to `a242e5a4` = M1 rebaseline, **not** M2 — correctly excluded.)
- No duplication, no dead code, no split-brain: each fact (plan_tier/status/envelope) has a single source (the contract) the gen types and FE follow. The known re-drift risk (raw `map[string]any` hand-serialization) is **explicitly recorded as an M3 defer in all three specs**, not a silent split-brain.
- Staff-engineer bar **met** ✅ — surgical, contract-first, one batched regen as the milestone mandated.

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (`backend-api-qa-checklist`) | pass | Contract-first order honored; route/field-truth tri-source (runtime↔spec↔gen↔FE) agrees for all three; api-lint -strict 0; build/vet/test clean. |
| FE screen-impact (UsageGauges F2.1 consumer + templates-list F2.3 consumer) | pass | Single regen did not break consumers; `tsc --noEmit` exit 0. UsageGauges keeps placeholder (scope-A); templates.ts hand-parse still matches the now-declared envelope. |
| Focused audit slice (contract dimension, HS-4) | pass — 0 undeclared | Swept the H-D-relevant delivery handlers. `usageToJSON` keys = {seats,storage,api_calls,active_users,plan_tier} all declared. `listTemplates` top-level = {data,meta} declared. `admin_handler` overview (most complex multi-block emit): `presence` item {…,status} declared; `kpi`→IamKpiSnapshot, `recent_activities`→AuditEventItem (incl. trace_id, payload) all declared. api-lint -strict 0 corroborates project-wide declared-shape consistency. **No 4th live emitted-but-undeclared instance found → no HS-4 trip.** |
| Regression vs M0 | all still pass | M0/M1 `qa/milestone-qa.md` artifacts intact; docs progression surface unbroken. |
| Regression vs M1 | all still pass | `/iam/presence/stream` still declared in openapi.yaml (F1.2); bare-405/error-contract tail holds (api-lint -strict 0); presence/admin regression tests (DropsUsersField, TenantIsolation, RunsInParallel) green in the uncached suite run. The single FE regen moved no unrelated route. |

**C4 PASS.**

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Contract dimension — H-D class (handler emits undeclared field) | 3 known instances (plan_tier, OnlinePresenceItem.status, /templates envelope) live; tri-source drift | 0 of the 3 remain; audit slice finds 0 others | Each field/shape declared in openapi.yaml, regenerated, consumed through a generated FE type (no hand-typed shim is *load-bearing* — the F2.2 shim and F2.3 hand-parse are now redundant, removal deferred). Tri-source agreement reproduced for all three. Not a symptom-patch: the contract is corrected to the real emit, not the symptom hidden. |

- **Could it be built better?** Yes, and it is already recorded as an M3 defer in all three specs: the handlers hand-build `map[string]any`, so declaring the field does not *force* serialization from the generated type — a future field can re-drift. The hard-prevent fix is emit-from-generated-type (M3 raw-map→generated-type pattern). This is a known, owned, triggered defer — it does **not** make the current construction unsound (the contract is honest today), so it does not FAIL M2. Likewise the F2.2 shim / F2.3 hand-parse removals are correctly batched as bounded FE cleanups. **C5 PASS.**

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean: per-feature acceptance mapped in C1/C2.*
- [ ] Fixture/mock passed off as real-provider proof — *clean: runtime rows explicitly labeled asserted-from-evidence; positive status emit proven by re-run test.*
- [ ] Consumer contract guessed rather than read from the consumer — *clean: all three read from FE consumer + live emit (C1).*
- [ ] Split-brain (one fact, two sources of truth) — *clean: single contract source; re-drift risk is a recorded M3 defer, not a live dual-source.*
- [ ] Self-judged close / validator edited or fixed code — *clean: validator wrote only this file; status not flipped.*
- [ ] Scope drift — *clean: only the 3 contract surfaces + the one canon-forced rename (recorded); H-G (M4) and mechanical-quality (M3) untouched.*
- [ ] Symptom-patch — *clean: root cause (undeclared contract) fixed, not masked.*

All unchecked = clean. **C6 PASS.**

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass: **code-wise** (surgical, contract-first, no split-brain/dead code, no guessed contract) and **function-wise/QA** (each feature does end-to-end what its spec promised; tri-source agreement reproduced; the single FE regen leaves `tsc` at 0 and moved no unrelated route; H-D class audit slice = 0 remaining).
- Handed back to the main session to flip M2 status and present the **HS-1** operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending operator review.
> - Status flipped in `README.md`: no — only the main session, only after operator approval.

VERDICT: PASS
