# Milestone 8 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md` + mission §8 (incl. the F8.6 scope amendment).
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-20  ·  **Verdict:** see C7 — **PASS**.
> Run after every feature is closed (each has a complete `evidence.md`). The validator judges and
> writes this file; the **main session flips status only on a PASS**. The validator never edits code,
> fixes findings, or flips status.

**Inputs loaded (none missing/unreadable):** `milestone.md`; all six features' `spec.md`/`plan.md`/`evidence.md`
(F8.1–F8.6); program `README.md`; governing `mission.md` (incl. §8 amendment); ADR 0038; `wiki/architecture/api-contract.md`
§5b; the aggregate M8 diff `a00dd78a..HEAD` (the six `f8.*` feature commits on `main`).

**Code baseline:** `a00dd78a` (M8 scaffold). **HEAD:** `ca8cc817` (F8.6). Aggregate diff = 47 files, +2802/−81,
scoped to the six features + docs/ADR + FE codegen regen. No source change in the working tree (only untracked
M2 smoke logs; the validator made no edits).

## C1 — Spec & plan conformance (per feature)

All six features have an approved `spec.md` (each `Approved before code: ✅ 2026-06-20`, with a populated
Interview record), an execution-shaped `plan.md` (files-touched + test strategy + task order, not a re-spec), and an
`evidence.md` whose acceptance table maps row-for-row to the spec Validation Gate. Consumer contracts were read from
the consumer site, not guessed (each spec carries a "Source of truth" line). Non-goals respected. ADR required by
F8.3 exists (ADR 0038); F8.6's durable scope change recorded as the mission §8 amendment.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F8.1 presence-typed-body | ✅ emits `iamapi.PresenceSnapshotResponse` per OpenAPI `PresenceSnapshotResponse` | ✅ | ✅ (WS `/stream`, OpenAPI schema, StrictServerInterface untouched) | handler.go:83 `WriteJSON(...toPresenceSnapshotResponse...)`; no response `map[string]any` |
| F8.2 metrics-typed-envelope | ✅ typed `MetricsResponse`; OpenAPI declares every emitted top-level key; FE regen matches | ✅ | ✅ (provider interfaces / sub-object typing / scrape format untouched) | http.go:183–207; OpenAPI `:5024/:5028`; FE `index.d.ts` +8 |
| F8.3 search-taxonomy-port | ✅ search depends on taxonomy `FamilyCodeResolver` interface; taxonomy owns impl (ADR 0038) | ✅ | ✅ (no schema/migration; result shape/ordering/visibility unchanged) | reader.go subqueries removed; ADR 0038; global-PK discovery pinned by test |
| F8.4 deactivation-session-enforcement | ✅ revoke-on-deactivate + fail-closed-at-resolve → 401 (CWE-613) | ✅ | ✅ (no token rotation / TTL change / new endpoint; force-logout deferred PR-7 with trigger) | service.go:627/887; middleware.go:75 |
| F8.5 problem-json-405 | ✅ stdlib `text/plain` 404/405 → `application/problem+json`, `Allow` preserved (D-03) | ✅ | ✅ (no route migration / new routes / hand-coded-405 change) | method_not_allowed.go; chain `method_not_allowed` innermost; REQ-MW-7 updated |
| F8.6 gate-scope-widening | ✅ §5b/§8 cover the full public surface; AST guard enforces it | ✅ | ✅ (gate/doc/CI only — no source-behavior change) | noresponsemap.go (registered in RunAll); §5b + §8 amended; health exemption recorded |

## C2 — Gates re-run, isolated

Re-run by the validator from clean state — not trusted from the evidence transcript. Integration tests run against
the real docker Postgres (`metaldocs-postgres`, healthy, port 5433) with `METALDOCS_DATABASE_URL` set, `-tags=integration`.

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| build | `GOFLAGS=-mod=mod go build ./...` | exit 0 | ✅ |
| F8.1 | `go test -count=1 -v -run TestHandler_Snapshot_TypedBody ./internal/modules/iam/presence/` | `--- PASS: TestHandler_Snapshot_TypedBody`; pkg `ok` | ✅ |
| F8.2 | `go test ... -run TestMetricsHandler_TypedEnvelope... ./internal/platform/observability/` | `--- PASS` KeySet + ItemsAlwaysPresent; pkg `ok` | ✅ |
| F8.3 (real DB) | `go test -tags=integration -run TestFamilyCodeResolverRepository ./internal/modules/taxonomy/infrastructure/` | **6/6 PASS** incl. `_CodeIsGlobalPrimaryKey`, `_SentinelFallback` (79.9s) | ✅ real |
| F8.3 (real DB) | `go test -tags=integration -run 'TestListDocuments_Family\|...Visibility' ./internal/modules/search/infrastructure/v2documents/` | **4/4 PASS** (FamilyProjection, FamilyFilter, FamilyFilterPagination, Visibility) (94.6s) | ✅ real |
| F8.4 | `go test ... -run 'TestUpdateUser_Deactivate.../TestResolveSession_FailsClosed.../TestMiddleware_DeactivatedIdentity_Returns401' ./internal/modules/auth/...` | all 4 named `--- PASS`; auth suite `ok` | ✅ real |
| F8.5 | `go test ... -run 'TestMethodNotAllowedJSON_RewritesStdlib40[45]\|TestAPIChainOrder_REQMW7' ./internal/platform/middleware/ ./apps/api/cmd/metaldocs-api/` | named `--- PASS`; both pkgs `ok` | ✅ real |
| F8.6 | `go run ./tools/cilint ./...`; analyzer unit tests; **validator-planted** violation | cilint exit 0; `go test ./tools/cilint/...` `ok`; planted `page:=map[...]; WriteJSON(...page)` in `audit/delivery/http/` → guard exit 1 with `[noresponsemap] … H-D`, exit 0 after revert (file restored) | ✅ |

The F8.3 real-DB tests, F8.4 service+middleware tests, and the F8.6 planted-violation are genuine end-to-end /
real-provider proof — not fixtures. The presence/metrics "parity-lock" tests are honestly labelled characterization
tests (green before and after a pure typed-parity refactor); the feature's real red→green is the H-D grep/guard.

## C3 — Senior review of the aggregate milestone diff

Reviewed `a00dd78a..HEAD` as one unit.

- **F8.4** (`service.go`): atomic-tx revoke (`UpdateUserTx`+`RevokeSessionsByUserIDTx`+`Commit`) with an in-memory
  fallback; `deactivating` guard prevents over-broad revoke; H-PRE-1 honored (no authz-recording read inside the
  tx — same shape as `AdminResetPassword`); fail-closed placed *before* tenant/role resolution; sentinel mapped to
  401 (not 500). Defense-in-depth, not a symptom patch.
- **F8.3** (`reader.go`): both cross-schema `metaldocs.document_profiles` subqueries removed; family **filter** stays
  in SQL (`= ANY($14)`) so `LIMIT/OFFSET` pagination is unchanged; projection moved to a Go batch resolve over the
  identical join key; `NoopFamilyCodeResolver` null-object guards nil callers; `pq.Array` for `= ANY`; parameter
  numbering correct. HS-6 global-PK discovery surfaced and resolved in-boundary (production SQL retained verbatim for
  parity/forward-safety; precedence documented as defensive/unreachable; invariant pinned by test). Provider-owned
  port (ADR 0031 mirror) — dependency direction correct, no split-brain.
- **F8.2** (`http.go` + OpenAPI + FE): one source of truth — typed envelope, OpenAPI schema, and FE codegen all
  declare the same top-level keys; the three sub-objects are declared-dynamic by design (operator decision), commented
  as such. No module import from platform (REQ-TOP-2).
- **F8.6** (`noresponsemap.go`): AST, two-pass, laundering-resistant; covers writer aliases; scoped to the widened
  surface; named exemptions + inline directive; registered in `RunAll` (rides `invariants.yml`). The `txownership`
  drive-by (testdb seed harness) is correctly classified as pre-existing repair (CLAUDE.md §4).
- No duplicated logic across features, no dead code introduced, no contract defined two ways, no feature breaking
  another (whole suite green, below).

- Findings: none material.
- Staff-engineer bar met? ✅

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend-api: typed-body / error-envelope / module-boundary) | pass | typed bodies (F8.1/F8.2), problem+json error envelope (F8.5), module-boundary port (F8.3), session lifecycle (F8.4) |
| Regression vs prior milestones (M0–M7) | all still pass | full-repo `go test -count=1 ./...` = **85 packages ok, 0 FAIL**; `go build ./...` exit 0 |

## C5 — Quality-bar re-measure + retrospective

Re-measured the declared quality bar (the §8 H-D/H-G classes) **at the widened scope** — the root-cause fix, not a symptom patch.

| Bar / class | Before (post-M7) | After (M8, validator-re-run) | Root-cause-fixed evidence |
|-------------|------------------|------------------------------|---------------------------|
| Honest H-D (response-literal `map[string]any` on any public route) | 2 (presence + metrics) | **0** | §5b Part A returns only the 2 recorded-exempt `health.go` lines (0 elsewhere); Part B survivors all allowlist-legitimate (audit `recordAudit` payloads, command `FormData`/`SignaturePayload`, declared-dynamic metrics fields, provider-interface return types, 2 comments) — verified line-by-line, no response literal; `go run ./tools/cilint ./...` exit 0 |
| Honest H-G (cross-module raw read of another module's owned table) | 1 (search→`document_profiles`) | **0** | `grep document_profiles` outside taxonomy returns only 1 comment + test-seed SQL; search module fully clean (0 matches); read now behind ADR-0038 port |
| Gate non-regressability | path-scoped grep (blind to aliases/laundered locals) | mechanical AST guard, full surface, in CI | validator-planted laundered-local violation caught (exit 1); guard registered in `RunAll` |

The root cause (gate scoped narrower than §8 intent) is fixed at the gate itself plus the mechanical guard — the
class is now closed and non-regressable, not merely the two instances patched.

- Could it be built better? The bounded defer in F8.6 (the analyzer does not follow a map literal laundered through
  a helper-function *return*) is the only residual; it is honestly recorded with a written trigger (a future Part-B
  survivor that is a returned response literal). No current instance exists and Part B's whole-package grep would
  still catch it at audit time. Not a soundness gap → does not affect this verdict; it is correct input to the
  terminal re-audit's vigilance. Otherwise the construction is sound.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean (each feature's named tests re-run and mapped above)*
- [ ] Fixture/mock passed off as real-provider proof — *clean (F8.3 real DB; parity-lock tests honestly labelled characterization)*
- [ ] Consumer contract guessed rather than read from the consumer — *clean (each spec cites a Source-of-truth)*
- [ ] Split-brain (one fact, two sources of truth) — *clean (F8.2 envelope/OpenAPI/FE aligned; F8.3 port = single owner)*
- [ ] Self-judged close / validator edited or fixed code — *clean (validator only judged; planted violation reverted, tree restored; no status flipped)*
- [ ] Scope drift (work beyond the spec, no rationale) — *clean (only recorded drive-bys: F8.3 sqlmock-signature update, F8.6 `txownership` testdb allowlist — both CLAUDE.md §4 pre-existing repair with rationale)*
- [ ] Symptom-patch (bar "moved" by masking, root cause intact) — *clean (gate widened + mechanical guard added = root-cause closure)*

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass independently: **code-wise** (senior-level, contract-clean, no split-brain, no dead code, no
  guessed contracts — C1/C3) and **function-wise/QA** (end-to-end real-provider proof for the heavy features; honest
  H-D=0 / H-G=0 at the widened scope; whole-repo build+test green — C2/C4/C5). No forbidden-list hit (C6).
- Handed back to the main session to flip status and present the HS-1 operator gate. The program **terminal** gate
  (post-M8 10-dimension re-audit + `mission-validator` against §8 4/4) is separate and is the main session's next
  action after HS-1 — this verdict closes only the M8 per-milestone gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — only on PASS, by the main session
