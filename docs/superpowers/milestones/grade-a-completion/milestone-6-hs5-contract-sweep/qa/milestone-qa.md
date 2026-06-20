# Milestone 6 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-19  ·  **Verdict:** see C7 — **PASS** (with 2 recorded non-blocking findings).

## Inputs loaded

- Milestone spec: `../milestone.md` (read).
- Feature artifacts: F6.1, F6.2 (`spec.md`+`plan.md`+`evidence.md`), F6.3 (`f6.3-iam-admin-sessions-observability-typed/` spec+plan+evidence), F6.4, F6.5, F6.6 — all present and read. Note: stray empty folder `f6.3-iam-handlers-typed/` (rename artifact) carries no content; the real F6.3 work lives in `f6.3-iam-admin-sessions-observability-typed/`.
- Program README: `../../README.md` (read). Governing spec: `../../mission.md` (referenced).
- Aggregate diff: `git diff 8f70b8bb..33bf4f02` (M6 commits `1c38f8a0`, `c2c87cd6`, `24e6539b`, `0249d0ec`, `33bf4f02`; F6.5 bundled into `24e6539b`).

No input missing or unreadable.

## C1 — Spec & plan conformance (per feature)

All 6 features carry a filled `Approved before code:` line (2026-06-19 / leandrotca.work via M6 operator scope confirmation), populated interview/contract records, and evidence acceptance that maps row-for-row to each spec's Validation Gate. Consumer contracts honored: all referenced codegen types verified to exist (`iamapi.UpsertUserRoleResponse`, `ReplaceUserRolesResponse`, `ListSessionsResponse`, `SessionItem`, `CursorPage`, `UsageSnapshot`, `IamKpiSnapshot`, `GrantAreaMembershipResponse` at `internal/modules/iam/api/api.gen.go`; templates `*Response`/envelope types in `internal/modules/templates/api/api.gen.go`). Producer emits the generated type — not the reverse.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F6.1 templates-lifecycle-typed | yes — emits generated `*Response`/`TemplateVersionEnvelope`; OpenAPI 200 schemas declared | yes (gate rows 1–6 re-verified) | yes | `routes_lifecycle.go` 0 literals; typed-shape test green |
| F6.2 templates-query-typed | yes — `GetTemplateResponse`/`GetTemplateDocxUrlResponse`/`ListTemplateAuditResponse` declared + emitted | yes (gate rows 1–7) | yes; 1 pre-existing test fixed (non-UUID version id) with rationale | `routes_query.go` 0 literals; typed-shape test green |
| F6.3 iam-admin/sessions/observability-typed | yes — generated iam types emitted | yes (field-name parity + tests green) | yes | see C3 finding F-1 (disclosed wire-value delta on invalid timestamps) |
| F6.4 class-sweep-typed | yes — typed structs on security/taxonomy/templates-catalog/schema | yes | yes; remaining `security/handler.go:54 Evidence map[string]any` is a domain-mirror struct field (`securitydomain.Signal.Evidence`), legitimately `any`-valued, correctly out of scope | 9 sites 0 literals |
| F6.5 iam-memberships-typed | yes — `GrantAreaMembershipResponse` + local `listMembershipsResponse` | yes | yes | `routes_memberships.go` 0 literals |
| F6.6 fe-codegen-regen-final-proof | yes — FE openapi-typescript regen includes all 8 new schemas | yes (grep A = 0, wiki stamp, regression) | yes | re-verified below |

## C2 — Gates re-run, isolated

All re-run by the validator from clean state (not trusted from transcripts):

| Gate | Command re-run | Real output | Pass? |
|------|----------------|-------------|-------|
| H-D Grep A | `grep -rEn 'writeJSON.*map\[string\]any' internal/modules/*/delivery/http/ --include='*.go' \| wc -l` | `0` | yes |
| BE codegen freshness | `GOFLAGS=-mod=mod go generate ./internal/modules/templates/api/...` then `git diff --exit-code .../api.gen.go` | generate exit 0; diff exit 0 (no uncommitted diff) | yes |
| Build | `go build ./...` | exit 0, no output | yes |
| Full suite | `go test -count=1 ./...` | exit 0 — **85 ok, 0 FAIL** | yes |
| Per-feature typed tests | `go test -count=1 -run 'TestLifecycle_TypedResponseShape\|TestQuery_TypedResponseShape' ./internal/modules/templates/delivery/http/...` | `ok ... 2.881s` | yes |
| Per-file site grep (10 cited M6 files) | `grep -cE 'writeJSON\(...map\[string\]any\|WriteJSON\(...map\[string\]any'` per file | `0` for all 10 | yes |
| OpenAPI 200 schemas | `grep -c` of the 8 new schema names in `openapi.yaml` | 18 (defs + refs) | yes |

## C3 — Senior review of the aggregate milestone diff

Diff is tightly scoped to exactly the intended surface: `openapi.yaml`, `templates/api/api.gen.go`, FE `api-types/index.d.ts`, `routes_lifecycle.go`, `routes_query.go(+test)`, `routes_catalog.go`, `routes_schema.go`, iam `admin_handler.go`/`observability_handler.go`/`sessions_handler.go`/`routes_memberships.go`, `security/handler.go`, taxonomy `routes_areas.go`/`routes_families.go`, 2 new typed-shape tests, wiki stamp. No file outside scope touched. No scope creep.

- **Finding F-1 (wire-value delta, disclosed, unpinned)** — `sessions_handler.go`: prior code rendered an invalid `sql.NullTime` as `""` via `nullTimeRFC3339`; the typed `iamapi.SessionItem.{CreatedAt,LastSeenAt,ExpiresAt time.Time}` now serializes a never-set timestamp as `"0001-01-01T00:00:00Z"`. F6.3 evidence (line 25) discloses and rationalizes this as MVP-acceptable (active sessions always have valid CreatedAt/ExpiresAt; LastSeenAt may be zero on brand-new sessions). The change is honest and bounded to never-set timestamps, but **no test pins it** — `sessions_handler_test.go` asserts status codes / display-name enrichment / tenant isolation only, never timestamp values. Not a forbidden-list hit; recorded as a quality nit + next-milestone test-pin candidate.
- **Finding F-2 (dead code)** — `nullTimeRFC3339` (`sessions_handler.go:305`) is now defined with zero call-sites after the F6.3 swap. Build stays green (unexported, vet-tolerated) but it is dead. Minor; record for drive-by cleanup.
- Pre-existing duplication `kpiToJSON` vs `kpiToOverviewTyped` (both now `iamapi.IamKpiSnapshot`) correctly left untouched and disclosed in F6.3 evidence — not introduced by M6.

Staff-engineer bar met? **yes** — with F-1/F-2 noted. Construction is sound; no split-brain, no guessed contract, no feature breaking another.

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend-api / contract sweep) | pass | typed responses on all 24 cited sites; OpenAPI 200 schemas declared + BE/FE codegen contract-first regen; freshness verified |
| Regression vs prior milestones (M0–M5) | all still pass | full suite 85 ok / 0 FAIL; H-G = 0; F5.1 (`templates/infrastructure/template_version_reader.go`) and F5.2 (`auth/infrastructure/postgres/repository.go`) present and untouched by M6 diff |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| H-D Grep A (`writeJSON.*map[string]any`) | 24 | **0** | validator re-ran grep → 0; per-file counts 0 on all 10 cited files; root cause (untyped envelope literals) replaced by generated/typed structs, not masked |
| Contract/API ≥ A− | B− | indicatively A− on the swept surface | 9 OpenAPI 200 schemas declared; producer = generated type; contract-first regen order respected (OpenAPI → BE codegen → FE codegen) |

- **Could it be built better? Material retrospective note (input to terminal re-audit / HS-5 watch):** the milestone's binding metric (H-D Grep A, pattern `writeJSON.*map[string]any`) is **0**, but the milestone's own forbidden-list phrasing "no `map[string]any` literal surviving on any public delivery route" is **not** true in absolute terms. Pre-existing response-body literals via `httpresponse.WriteJSON(...)` / `writeFillInJSON(...)` survive outside the 24-site scope and were untouched by M6 (verified pre-existing, not regressions): `audit/delivery/http/handler.go:127,216,268`; `auth/delivery/http/handler.go:90,161`; `search/delivery/http/handler.go:134`; `documents/delivery/http/{fillin_handler.go:58,116, pdf_webhook_handler.go:113, placeholder_options_handler.go:67,74, view_handler.go:46}`. They fall outside the grep pattern (capital-`W` `WriteJSON` / wrapper `writeFillInJSON`), so the binding gate passes by design, but Contract/API is **not yet absolutely literal-free**. The terminal re-audit should weigh whether these affect the contract-api grade; Contract/API has missed twice — if the post-M6 re-audit treats these as a third §8 miss, that is the HS-5 → HS-2 boundary (codegen-first adoption), not a bounded M7. This note does not FAIL M6: M6's construction is sound and its declared, operator-approved scope (24 sites + 9 schemas + 2 regens) is fully delivered.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — clean (per-feature acceptance mapped in C1/C2)
- [ ] Fixture/mock passed off as real-provider proof — clean (handler-via-mux proof explicitly labeled; codegen/grep proofs are real commands)
- [ ] Consumer contract guessed rather than read from the consumer — clean (codegen types verified present; consumer sites cited in specs)
- [ ] Split-brain (one fact, two sources of truth) — clean (OpenAPI/codegen single source; pre-existing kpi duplication out of scope, disclosed)
- [ ] Self-judged close / validator edited or fixed code — clean (validator only re-ran read-only gates + regen freshness check that produced no committed change; wrote only this verdict)
- [ ] Scope drift — clean (diff strictly within the 24 sites + 9 schemas + 2 regens; the 1 pre-existing-test fix in F6.2 carries a rationale)
- [ ] Symptom-patch — clean (H-D = 0 via real typing, not grep-evasion; no `interface{}(map…)` cloak)

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- M6's binding gates all pass on isolated re-run: H-D Grep A = 0, BE codegen fresh (no uncommitted diff), `go build ./...` clean, `go test -count=1 ./...` 85 ok / 0 FAIL, F6.1/F6.2 typed-shape tests green, all 24 cited sites typed, 9 OpenAPI 200 schemas + FE codegen contract-first. The declared, operator-approved scope is fully delivered; no forbidden-list hit; no scope drift; no prior-milestone regression.
- **Non-blocking findings carried forward (do not gate this milestone):**
  - F-1 — F6.3 disclosed wire-value delta on never-set session timestamps (`""` → `"0001-01-01T00:00:00Z"`) is unpinned by test. Recommend a value-asserting sessions wire test in the next milestone or terminal close-out.
  - F-2 — dead `nullTimeRFC3339` at `sessions_handler.go:305`. Recommend drive-by removal.
  - C5 retrospective — pre-existing `WriteJSON`/`writeFillInJSON` `map[string]any` response literals survive outside the 24-site scope (audit/auth/search/documents). Not an M6 regression, but material to whether the terminal re-audit grades Contract/API ≥ A−. **HS-5 watch:** a third Contract/API miss is the HS-2 (codegen-first) boundary, not a bounded M7.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — only on PASS + HS-1, by the main session
