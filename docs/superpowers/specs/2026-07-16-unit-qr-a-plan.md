# Unit QR-A — Kernel + Template Wiring Remediation (Batch Plan)

- **Date:** 2026-07-16
- **Planner:** P2 Batch Planner (Opus), PLAN-ONLY
- **Branch:** `unit/qr-a-kernel-template-wiring` @ base `e7826c30`
- **Binding:** HARNESS-CORE §4 (slice cards + WRITE-SET + contract-satisfiability + additive locks), REVIEW-STANDARD (G2 alternatives-considered, §15 disjointness)
- **Invariants in force:** capabilities-never-roles (ADR 0022), contract-first (spec = route truth), RFC 9457, no-fallback, legacy-fallback extermination, testdb factory hard gate, actor ids TEXT.
- **Collision ownership:** QR-A owns `internal/modules/approval/**`, `apps/api/cmd/metaldocs-api/main.go`, `frontend/apps/web/src/features/approval/**`, capability registry files, `api/openapi` (locked). QR-B owns `apps/worker/**`, `apps/jobs/**`, render materialize consumer, documents publish path + its FE surface. **No overlap** on approval, main.go, or openapi — verified against QR-B scope. **No migrations in QR-A.**

---

## Root-cause verification log (re-verified against code this session)

| Claim | Verified anchor | Status |
|---|---|---|
| F18 spec requires profile_code | `api/openapi/v1/openapi.yaml:7253` `required: [profile_code, name, stages]` | CONFIRMED |
| F18 contract unconditionally requires profile_code | `internal/modules/approval/http/contracts/route.go:143-148`; enum switch for subject_kind already present :152-156 | CONFIRMED |
| F18 service never nulls profile_code for template | `route_admin_service.go:218-231` resolveCreateRouteSubject (no ProfileCode nulling); `:267-275` INSERT binds `in.ProfileCode` as `$2` | CONFIRMED |
| F18 DB truth | migration `db/migrations/0297_*.sql:141-145` `approval_routes_template_subject_projection_check CHECK (subject_kind <> 'template' OR profile_code IS NULL)`; `:135-139` document check `... OR profile_code IS NOT NULL` | CONFIRMED |
| F18 false-green mock test | `route_admin_handler_test.go:289` posts `profile_code:"ops"` + `subject_kind:"template"` through mock `svc` — never hits DB | CONFIRMED |
| F22 in-place mutation lines | main.go `:645` WithTemplateVersionReader, `:649` WithTemplateCompletionWriter, `:684` WithLifecycleEnqueuer — Services.With* mutate `s` and return same pointer (services.go:114-134,158-182) | CONFIRMED |
| F22 rebuild drops ports | main.go `:737-741` `approvalServices.Decision = NewDecisionService(...).WithPDFOutbox(...).WithPinInvoker(...).WithSignatureRegistry(...).WithCDFieldReader(...)` — BRAND-NEW pointer; missing templateVersionReader/templateCompletion/lifecycleEnqueuer | CONFIRMED |
| F22 DecisionService With* mutate-in-place | decision_service.go:118-169 all `s.field = x; return s` | CONFIRMED |
| F22 FastForward captures original pointer | services.go:98 `FastForward: newFastForwardService(reviewVerdict, decision)` — captures the ORIGINAL `decision`, diverges after :737 rebuild | CONFIRMED |
| F22 stripped instance flows to templates + handler | main.go `:765` WithApprovalKernel(..., approvalServices.Decision, ...), `:770` NewHandler(approvalServices, ...) | CONFIRMED |
| F2 FE drift | routeDraft.ts:21 (defaultStage), :50 (toDraft `||`), :86 (toStageRequests `||`) use `'doc.signoff'` | CONFIRMED |
| F2 canonical cap | iam/domain/model.go:80 `CapDocumentSignoff = "document.signoff"`; no `doc.signoff` in registry | CONFIRMED |
| F2 KEEP audit presenter | audit-event-presenter.ts:61 `doc.signoff` + :67 `document.signoff` are **event-type** label-map keys (historical audit event names), NOT capabilities — distinct namespace, legitimately retained | CONFIRMED |
| Collision: QR-B untouched on openapi/approval | QR-B scope = worker/jobs/render/documents-publish; no approval/openapi/main.go paths | CONFIRMED |
| Tooling present | `scripts/api-lint`, `scripts/test-integration.ps1`, `internal/modules/approval/api/gen.go:3` go:generate directive | CONFIRMED |

---

## Slice ordering (dependency rationale)

1. **S1 (F18 contract+service+spec)** — self-contained backend + spec; gated on api/openapi contract-lock. Independent of S2/S3.
2. **S2 (F22 composition-root wiring + guard)** — backend main.go + approval application guard/test; independent of S1/S3.
3. **S3 (F2 FE capability single-source + FE tests)** — frontend-only; independent.
4. **S4 (Go test-fixture capability sweep)** — mechanical Go fixture rename; touches approval + jobs test files only. Disjoint from S1-S3 sources but shares the approval test tree with S1 (sequence S1→S4 to avoid file churn overlap; see §Disjointness).

All four are logically independent (no shared production symbol). Ordering is by contract-lock readiness (S1 first once lock granted) and to keep the approval test tree edited by one slice at a time (S1 then S4).

---

## Slice cards

### S1 — F18: template route creation (contract conditional + service NULL bind + spec)

**Goal:** Make `subject_kind=template` route creation succeed with `profile_code` persisted NULL; reject a template create that carries a non-empty `profile_code` (fail honest, no silent drop); keep document create requiring profile_code. Align spec to DB truth.

**Failing-test-first step:** Add a **real service + DB** integration test (extend `route_admin_service_subject_integration_test.go`, testdb factory) — *Chosen level: real service+DB, not full HTTP handler.* Rationale below (alt-considered). Cases:
  - (a) template create, empty profile_code → success + assert `SELECT profile_code FROM approval_routes` IS NULL + `subject_kind='template'`.
  - (b) template create, non-empty profile_code → `contracts.CreateRouteRequest.Validate()` returns validation error (400 class). (Unit test in `contracts` package — no DB.)
  - (c) document create (absent/`document` subject_kind), empty profile_code → validation error (unit, contracts).
  - (d) document create, valid profile_code → success + profile_code populated (regression guard; real service+DB).
  Run → RED (template create currently binds 'ops' → DB check violation 500 / or contract rejects empty).

**Change set:**
  - `contracts/route.go` `Validate()`: make profile_code conditional — REQUIRED when `subject_kind ∈ {"", "document"}`; when `subject_kind == "template"`, `profile_code` MUST be empty (non-empty → `wrapValidation(fmt.Errorf("profile_code must be absent for template routes"))`). Keep route-code pattern check only when non-empty. Preserve RFC 9457 mapping (existing wrapValidation → problem+json).
  - `route_admin_service.go` `resolveCreateRouteSubject`/`createTx`: for template subject, bind `profile_code` as SQL NULL (not `in.ProfileCode`). Concretely: derive a `sql.NullString`/`*string` profileArg = NULL when `subject.Kind == template`, else `in.ProfileCode`; pass to `$2` in INSERT. `resolvePolicy`/route-shape check: template path must not call the profile-policy reader with a profile_code it will null — decide: skip policy resolution for template subject (there is no profile). State: for template, `policy` resolution short-circuits (no per-profile signature policy applies to a template route in this slice).
  - `api/openapi/v1/openapi.yaml:7250-7277`: drop `profile_code` from `required:` (→ `required: [name, stages]`); add conditional-rule prose to `profile_code`/`subject_kind` descriptions ("REQUIRED when subject_kind is document or absent; MUST be omitted when subject_kind is template"). **GATED on api/openapi contract-lock (see §Locks).**
  - Regen: `go generate ./internal/modules/...` (full canonical regen — embedded swaggerSpec churns ALL modules; commit spec + all regens together).
  - Delete/repair false-green mock test `route_admin_handler_test.go:289` block: it asserts template+profile_code success through a mock. Under the new contract that body is now a 400 (profile_code present on template). Repair it to assert the 400, OR delete if it becomes redundant with the new negative unit test — state choice in evidence.

**Done criteria:** New RED tests GREEN; `go build ./...` + `go vet` + gofmt clean; `api-lint -strict` on the changed spec passes; full regen committed with spec; template create persists NULL profile_code (asserted in DB); template-with-profile_code and document-without-profile_code both 400.

**Est. LOC:** ~180-240 (contract ~30, service ~40, spec ~15, regen churn is generated/not counted toward the 300 human-line budget, tests ~90, mock-test repair ~20).

**Disjointness:** Edits `contracts/route.go`, `route_admin_service.go`, `route_admin_service_subject_integration_test.go`, `route_admin_handler_test.go`, `api/openapi/*`, all `*/api.gen.go`. Shares approval test tree with S4 — sequence S1 before S4.

---

### S2 — F22: composition-root single-Decision convergence + boot guard

**Goal:** One `DecisionService` instance carries ALL ports (templateVersionReader, templateCompletion, lifecycleEnqueuer, pdfDispatch, pinInvoker, sigRegistry, cdRead); FastForward and templates handler observe the same fully-wired instance. Add a fail-fast readiness guard for the compile-green/runtime-dead class.

**Failing-test-first step:** Add a unit test in `internal/modules/approval/application` (no DB) pinning:
  - (a) `Services.WithTemplateVersionReader`/`WithTemplateCompletionWriter`/`WithLifecycleEnqueuer` set fields on the **same** `Decision` pointer (pointer-identity: `svc.Decision` before == after).
  - (b) `Services.FastForward` and `Services.Decision` reference the **same** `*DecisionService` after all With* calls (single-instance invariant).
  - (c) A new exported readiness check `DecisionService.Ready()` (or `MissingPorts() []string`) returns non-nil/error when a required port is nil.
  Run → RED (Ready() does not exist yet; single-instance test passes today at the application layer — the divergence is created only in main.go, so add a construction-order regression that mirrors main.go's mistake: build Services, rebuild `.Decision` via NewDecisionService, assert FastForward now diverges → documents the bug, then flip to assert convergence after the guard/helper lands).

**Change set:**
  - main.go `:737-741`: **replace the rebuild** with in-place mutation of the existing pointer:
    `approvalServices.Decision = approvalServices.Decision.WithPDFOutbox(pdfDispatchEnqueuer).WithPinInvoker(fanoutCfg.freezeService).WithSignatureRegistry(newSignoffReauthRegistry(...)).WithCDFieldReader(cdReader)` — since With* return `s`, this mutates the ORIGINAL instance already held by FastForward and (post-`:684`) carrying template + lifecycle ports. Net: one instance, all ports.
  - main.go `:761-764`: delete the false comment "Must happen after approvalServices.Decision is finalized above (line ~737-741)" / "Decision is finalized above" — reword to reflect in-place wiring.
  - `decision_service.go`: add exported `Ready() error` (or `MissingRequiredPorts() []string`) enumerating the ports that MUST be non-nil for full runtime function (templateVersionReader, templateCompletion, pdfDispatch, pinInvoker, sigRegistry — decide the required set: those whose absence causes a 500 or silent drop; lifecycleEnqueuer is best-effort/nil-tolerant per existing comments → document whether it is required or optional).
  - main.go composition root: call the guard **before** `templatesModule.WithApprovalKernel` (`:765`) and `NewHandler` (`:770`) — on non-nil error `slog.Error` + `deps.Cleanup()` + `os.Exit(1)`, matching the existing boot fail-fast idiom (e.g. :651-654).

**Alt-considered (guard mechanism):** (1) exported `Ready()` + boot fail-fast [CHOSEN — testable without DB, catches the class at boot, minimal surface]; (2) constructor that takes all ports positionally (compile-time completeness) — rejected: large blast radius across all With* call sites + breaks the established builder idiom; (3) no guard, rely on integration test — rejected: does not prevent a future re-introduction at a different call site (the actual failure mode here).

**Done criteria:** Unit tests GREEN; `go build ./...`/vet/gofmt clean; boot guard present + covered; template signoff no longer hits `decision_service.go:352` "template version reader not configured" path (asserted indirectly via the port-presence unit test — full live template-signoff is QR-B/hub live-QA territory, note as such).

**Est. LOC:** ~120-170 (main.go ~15 net, Ready() ~30, unit test ~90).

**Disjointness:** Edits `apps/api/cmd/metaldocs-api/main.go`, `internal/modules/approval/application/decision_service.go`, new `*_test.go` in approval/application. No overlap with S1 production files; shares approval/application test dir with S4 — but distinct files. Safe to run parallel with S1 if desired; recommend sequential for clean review diffs.

---

### S3 — F2: FE capability single-source + FE test truth

**Goal:** Eliminate `doc.signoff` from approval FE production paths; source the canonical `document.signoff` from one constants module mirroring the Go registry; remove silent-substitute fallbacks where they mask server truth.

**Failing-test-first step:** Update the FE tests that currently hardcode `doc.signoff` to assert canonical `document.signoff` FIRST (they turn RED against current source): `api/routeAdminApi.test.ts:34,46`; `pages/route-admin/RouteAdminPage.test.tsx:59,68`; `pages/route-admin/StageCard.test.tsx:52`; `queries/useRouteAdminMutations.test.tsx:41,108`. Run vitest → RED.

**Change set:**
  - New `frontend/apps/web/src/features/approval/pages/route-admin/capabilities.ts` (or feature-level) exporting canonical strings (e.g. `SIGNOFF_CAPABILITY = 'document.signoff'`), with a doc comment: "Mirrors the Go capability registry `internal/modules/iam/domain/model.go` (CapDocumentSignoff). No generated enum exists yet — hand-synced; see hub defer DEFER-QR-A-1."
  - `routeDraft.ts:21` defaultStage → use `SIGNOFF_CAPABILITY` (KEEP the default here — legitimate new-stage UX seed).
  - `routeDraft.ts:50` toDraft `stage.required_capability || 'doc.signoff'` → **drop the `|| fallback`**, preserve server value: `stage.required_capability` (a route stage always has a required_capability from the server; a silent substitute violates no-fallback and could mask a real backend value).
  - `routeDraft.ts:86` toStageRequests `stage.requiredCapability.trim() || 'doc.signoff'` → **drop the `|| fallback`**; the draft always carries a capability seeded by defaultStage. If a defensive non-empty guard is wanted, fail honest via existing `validateDraft` rather than substituting.
  - KEEP `audit-event-presenter.ts:61,67` untouched (event-type label map, not capabilities — verified).

**Alt-considered (fallbacks):** (1) keep default only in defaultStage, drop `||` in toDraft/toStageRequests [CHOSEN — no-fallback: server/draft value is authoritative, only the genuinely-empty new-stage case needs a seed]; (2) keep all three with canonical constant — rejected: the `||` in toDraft/toStageRequests is a silent substitute that can mask a divergent server value, the exact anti-pattern no-fallback forbids; (3) route-level codegen enum — rejected, out of unit scope (see DEFER-QR-A-1).

**Done criteria:** `tsc` clean; vitest for the six files GREEN asserting `document.signoff`; no `doc.signoff` remains under `features/approval/**` except none (grep clean); constants module documented as hand-synced mirror.

**Est. LOC:** ~90-130 (constants ~15, routeDraft ~10, six test files ~80).

**Disjointness:** Frontend-only; zero overlap with S1/S2/S4.

---

### S4 — Go test-fixture capability sweep (mechanical)

**Goal:** Replace `doc.signoff` seed literals with canonical `document.signoff` in Go integration test fixtures so they assert registry truth (currently seed a non-registry capability).

**Failing-test-first step:** N/A (these ARE tests). Verification is: the touched suites still pass AND now seed the canonical capability. Optionally add a one-line assertion that the seeded capability is registry-valid where a natural hook exists (do not over-engineer).

**Change set (mechanical literal swap `doc.signoff` → `document.signoff`):**
  - `internal/modules/approval/application/read_service_area_parity_integration_test.go:114`
  - `internal/modules/approval/application/read_service_instance_verdicts_integration_test.go:34`
  - `internal/modules/approval/application/read_service_viewer_facts_integration_test.go:77`
  - `internal/modules/jobs/stuck_instance_watchdog/job_integration_test.go:161`

**Alt-considered:** delegate to a mechanical builder subagent (caveman-builder) — acceptable given the 4-file bounded scope; or inline. Either way it is a pure literal swap; keep as its own slice so the review diff isolates fixture-truth from behavioral change.

**Done criteria:** `go build ./...` clean; `test-integration.ps1` for the two touched packages (approval/application, jobs/stuck_instance_watchdog) GREEN; grep shows no `doc.signoff` in these files.

**Est. LOC:** ~8 (4 one-line swaps).

**Disjointness:** Touches approval/application + jobs test files. Shares approval/application test dir with S1 (different files) and S2 (different files). Sequence **after** S1 to avoid concurrent edits to the approval test tree. No production code.

---

## Per-slice WRITE-SET (exact paths)

- **S1:** `internal/modules/approval/http/contracts/route.go`, `internal/modules/approval/application/route_admin_service.go`, `internal/modules/approval/application/route_admin_service_subject_integration_test.go`, `internal/modules/approval/http/route_admin_handler_test.go`, `internal/modules/approval/http/contracts/route_test.go` (or existing contract test file — negative unit cases), `api/openapi/v1/openapi.yaml`, regen outputs under `internal/modules/**/api/*.gen.go` + `internal/modules/**/http/*_generated.go` (full `go generate ./internal/modules/...`).
- **S2:** `apps/api/cmd/metaldocs-api/main.go`, `internal/modules/approval/application/decision_service.go`, new `internal/modules/approval/application/composition_wiring_test.go` (or `decision_service_wiring_test.go`).
- **S3:** new `frontend/apps/web/src/features/approval/pages/route-admin/capabilities.ts`, `frontend/apps/web/src/features/approval/pages/route-admin/routeDraft.ts`, `frontend/apps/web/src/features/approval/api/routeAdminApi.test.ts`, `frontend/apps/web/src/features/approval/pages/route-admin/RouteAdminPage.test.tsx`, `frontend/apps/web/src/features/approval/pages/route-admin/StageCard.test.tsx`, `frontend/apps/web/src/features/approval/queries/useRouteAdminMutations.test.tsx`.
- **S4:** the four Go integration test files listed in S4.

---

## Contract-satisfiability check

- **Spec section availability:** `CreateRouteRequest` (`openapi.yaml:7250-7277`) already carries `subject_kind` enum `[document, template]` and `subject_key`; only the `required:` array (`:7253`) + two prose descriptions change. `StageRequest.required_capability` remains a plain string (no enum) — S3 does NOT touch the spec; capability validity stays server-enforced (registry) as today, so no spec change for F2. **No new schema, no new route.**
- **Contract path diff vs current state:** verified `contracts/route.go:142-158` already validates the subject_kind enum; S1 adds the conditional profile_code rule inside the same `Validate()`. The Go contract and spec must move together (contract-first): spec `required:` drop + contract conditional land in the SAME slice/commit.
- **Regen blast radius:** the OpenAPI spec is embedded per-module (`swaggerSpec`) — any edit churns **every** module's `api.gen.go`. Full `go generate ./internal/modules/...` is canonical; partial regen is forbidden drift (known meta-gotcha). Commit spec + all regenerated files together. Expected large generated diff, zero human-reviewed logic in the churn.
- **QR-B non-collision:** QR-B does not touch `api/openapi` — confirmed against its scope (worker/jobs/render/documents-publish). No spec-lock contention.

---

## Additive contract-locks needed

- **LOCK-QR-A-1 (api/openapi):** additive edit to `CreateRouteRequest` (drop `profile_code` from `required`, add conditional prose). **Contract-lock REQUESTED from hub** (task states: assume granted by implement time; S1's spec sub-step is gated on it). This is the ONLY contract-lock QR-A needs. No other spec surface changes.

No further locks: S2 is composition-root Go only; S3/S4 are FE/test-fixture only.

---

## Verification mapping (ladder lanes per slice)

| Slice | go build/vet/gofmt | api-lint -strict | test-integration.ps1 (pkgs) | FE tsc + vitest |
|---|---|---|---|---|
| S1 | YES (after regen) | YES (spec changed) | `internal/modules/approval/application`, `internal/modules/approval/http/contracts` (+ handler pkg) | — |
| S2 | YES | — | `internal/modules/approval/application` (unit; no DB needed, runs under go test) | — |
| S3 | — | — | — | YES (`tsc` + vitest on the 6 files) |
| S4 | YES (build) | — | `internal/modules/approval/application`, `internal/modules/jobs/stuck_instance_watchdog` | — |

- Integration lane policy: touched packages + guard suites only (selective ladder policy); full `./...` NOT required (no db/platform-layer touch). testdb factory hard gate applies to S1's new DB integration test.
- Dual gate (unit close): cold Opus review + GPT-5.6 Sol medium at a fixed SHA over the full QR-A diff (per profile).

---

## Defers for the hub queue (named, dated)

- **DEFER-QR-A-1 (2026-07-16) — Capability registry→FE codegen.** Global-max fix for the hand-synced enumeration meta-defect: generate a TS capability enum from `internal/modules/iam/domain/model.go` (or a spec component enum) consumed by FE, replacing the hand-synced `capabilities.ts` constants module S3 introduces. New codegen infra = out of QR-A scope. Owner: hub. Ties to the standing "hand-synced enumerations" meta-defect (final-architecture-review 2026-07-03).
- **DEFER-QR-A-2 (2026-07-16) — Live template signoff + document-signoff lifecycle E2E.** S2 pins port-presence at boot/unit level; end-to-end confirmation that template signoff succeeds (no 500) and document-signoff lifecycle events actually enqueue belongs to hub live-QA / QR-B's runtime surface (login-password prohibition prevents planner-side browser QA). Owner: hub.
- **DEFER-QR-A-3 (2026-07-16) — 0296 compat trigger `default_approval_subject()` retirement.** Noted in migration 0297 header as tracked DEBT; not QR-A (no migrations). Owner: hub.

---

## Disjointness summary (REVIEW-STANDARD §15)

- S3 fully disjoint (frontend). S1/S2/S4 all backend but touch **distinct production files** (S1: contracts+route_admin_service; S2: main.go+decision_service; S4: test fixtures only).
- Overlap risk: S1 and S4 both edit files under `internal/modules/approval/application/*_test.go` (different files). **Mitigation: sequence S1 → S4.** S2's new test file is distinct from both.
- Recommended execution order: **S1 → S2 → S3 → S4** (S3 may run in parallel any time; S4 strictly after S1).
