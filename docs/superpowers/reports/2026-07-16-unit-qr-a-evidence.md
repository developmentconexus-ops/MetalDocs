# Unit QR-A Evidence — QA-1 Criticals F18 + F22 + F2

- **Date:** 2026-07-16
- **Branch:** `unit/qr-a-kernel-template-wiring` @ base `e7826c30` (main tip 1d148f6d was docs-only ahead; branched from e7826c30 per charter)
- **Charter:** `docs/superpowers/reports/2026-07-16-qa1-browser-qa-verdict.md` (F18, F22, F2)
- **Plan:** `docs/superpowers/specs/2026-07-16-unit-qr-a-plan.md` (P2 batch, Opus planner)
- **Doctrine:** HARNESS-CORE 0.2.0 + docs/HARNESS-PROFILE.md (§10 model matrix) + REVIEW-STANDARD.md
- **Worktree note:** chip worktree arrived UNMATERIALIZED (known gotcha, 3rd occurrence) — materialized via temp `git worktree add` + content move + `git worktree repair` (dir busy blocked direct add). `.env` copied, never read.

## Root-cause statements

- **F18** — Two-layer defect. (1) Contract: spec `CreateRouteRequest.required` includes `profile_code` (openapi.yaml:7253) and `contracts/route.go:143` `Validate()` requires it unconditionally. (2) Service: `route_admin_service.go` createTx binds `in.ProfileCode` verbatim into the INSERT regardless of subject kind. DB check `approval_routes_template_subject_projection_check` (migration 0297: `subject_kind <> 'template' OR profile_code IS NULL`) is DELIBERATE architecture truth (ADR 0082 subject-generic kernel; per-profile policy governs document routes only; 0297 header states template rows leave profile_code NULL). **Ruling: DB side correct; contract + service fixed** — profile_code conditional (required for document/absent kind, must be absent for template, non-empty on template → 400 fail-honest), service binds SQL NULL for template. No migration needed → no escalation.
- **F22** — `main.go:737` rebuilt `approvalServices.Decision` via `NewDecisionService(...)` (brand-new pointer) after template ports (:645/:649) and lifecycle enqueuer (:684) were set on the ORIGINAL pointer — dropping `templateVersionReader`, `templateCompletion`, `lifecycleEnqueuer` from the instance wired into templates (:765) and the approval handler (:770). Additionally `Services.FastForward` (services.go:98) captured the ORIGINAL pointer at construction → post-rebuild divergence: FastForward's Decision lacked `pdfDispatch`/`pinInvoker`/`sigRegistry`. The comment "Decision is finalized above" was false. **Fix: single-instance convergence** — `DecisionService.With*` mutate in place, so the rebuild is replaced with an in-place `With*` chain; boot-time readiness guard + unit tests pin the class.
- **F2** — FE `routeDraft.ts` (:21 defaultStage, :50/:86 `|| 'doc.signoff'` fallbacks) posts capability `doc.signoff`; Go registry canon is `document.signoff` (`iam/domain/model.go:80`), enforced at `contracts/route.go` validateRequiredCapability → every route save 400. No generated FE capability enum exists (api-types type it as plain string). **Fix:** FE single-source constants module (hand-synced mirror, documented; codegen = DEFER-QR-A-1), drop silent-substitute `||` fallbacks (no-fallback principle), fix 6 FE test files + 4 Go fixture files asserting/seeding registry truth. KEEP `audit-event-presenter.ts:61/:67` — those are event-type label keys, not capabilities.

## Dispatch ledger

| # | Role | Model | Vehicle | Input | Output | Status |
|---|---|---|---|---|---|---|
| D1 | Investigator (breadth) | sonnet | Agent sync | inline prompt (root-cause breadth: service/repo/handler/FE/registry/main.go/tests) | compressed report in-session (agentId a3014329dd69f94b9) | DONE |
| D2 | P2 batch planner | opus | Agent sync | inline prompt-pack (verified anchors + required outputs) | `docs/superpowers/specs/2026-07-16-unit-qr-a-plan.md` | DONE |
| D3 | Implementer S2 (F22) | sonnet | Agent sync | plan §S2 + doctrine pins | diff on worktree (committed 9adbbab6) | DONE |
| D4 | Implementer S3 (F2 FE) | sonnet | Agent sync | plan §S3 + doctrine pins | diff on worktree | DONE (after correction) |
| D5 | Reviewer S2 | sonnet | Agent sync | S2 diff + plan §S2 | APPROVE | DONE |
| D6 | Reviewer S3 (round 1) | sonnet | Agent sync | S3 diff + plan §S3 | REJECT — 1 blocking (no literal-assertion pin on POST body; reverting constant kept 235 green) | DONE |
| D7 | Implementer S3 correction | sonnet | SendMessage (a9f594189020e80eb) | blocking finding | literal assertion added, RED/GREEN proof (flip constant → 1 fail; restore → 235/235) | DONE |
| D8 | Mechanical S4 sweep | haiku | Agent sync | plan §S4 (4 Go fixture files) | 7 literal swaps; build+vet green; flagged 4 extra literals in iam/authz/authz_test.go | DONE |
| D9 | Reviewer S3 delta (§9) + S4 | sonnet | Agent sync | delta since D6 + S4 diff | JOB1 APPROVE (finding resolved, path traced defaultStage→toStageRequests→POST body) · JOB2 APPROVE (pure swaps, no semantic drift) | DONE |
| D10 | Implementer S1 (F18) | sonnet | Agent sync | plan §S1 + hub lock conditions | diff on worktree; RED→GREEN proof (check-violation FAIL → pass); regen churned ONLY approval api.gen.go (per-module embedded spec) | DONE |
| D11 | Reviewer S1 | sonnet | Agent sync | S1 diff + plan §S1 + hub conditions | APPROVE — authz.Require unconditional both paths, only friendly G1 policy check skipped for template (DB constraint = last line); no *string nil-deref (handler decodes via contracts, generated type unconsumed) | DONE |
| D12 | Dual gate arm 1 | opus (cold) | Agent background (retry after 1 API-error abort) | gate prompt @ fixed SHA 5ee67417, range e7826c30..5ee67417 | ACCEPT — no blocking; 2 minor + 1 info | DONE |
| D13 | Dual gate arm 2 | GPT-5.6 Sol medium | codex exec read-only, stdin closed, prompt from file | same fixed SHA/range | REJECT — 4 findings (2 CRITICAL NULL-scan read paths, 1 HIGH FE regen missing, 1 HIGH RouteSummary schema) | DONE |
| D14 | Implementer S5 (gate remediation) | sonnet | Agent sync | findings 1–3 verified + fix spec | diff on worktree; RED→GREEN stash proof (`converting NULL to string is unsupported`) | DONE |
| D15 | Reviewer S5 | sonnet | Agent sync | S5 diff | APPROVE — doc-row safety proven via 0297 companion check `approval_routes_document_subject_projection_check` + contract minLength=1; repo-wide scan audit (LoadRoute already safe); FE 409 drift confirmed pre-existing | DONE |
| D16 | Implementer S6 (RouteSummary nullable) | sonnet | Agent sync | hub lock-extension ruling + fix spec | diff on worktree; raw-JSON null serialization test; FE consumers audited | DONE |
| D17 | Reviewer S6 | sonnet | Agent sync | S6 diff | APPROVE — spec matches route_id nullable convention; 0297 checks plain ADD CONSTRAINT (no NOT VALID → no legacy-row gap); edit dialog never sends profile_code (UpdateRouteRequest has no such field) → `?? ''` is controlled-input normalization, not substitute; api.gen.go delta = field type + gofmt realignment only | DONE |
| D18 | Dual gate round 2 arm 1 | opus (cold) | Agent background (1 API-error abort, resumed in-context) | gate-2 prompt @ ffcfccf5, range e7826c30..ffcfccf5 | ACCEPT — round-1 findings resolved at root; live real-DB test execution; 1 minor (template accepts present-but-empty "") + 2 info | DONE |
| D19 | Dual gate round 2 arm 2 | GPT-5.6 Sol medium | codex exec read-only | same | round-1 findings all RESOLVED-at-root; REJECT with 4 NEW findings (idemp hash subject gap CRITICAL, presence-tracking HIGH, event/response "" sentinel HIGH, sweep straggler MEDIUM) | DONE |
| D20 | Implementer S7 | sonnet | Agent sync | 4 verified round-2 findings | diff on worktree; conflict test RED→GREEN | DONE |
| D21 | Reviewer S7 | sonnet | Agent sync | S7 diff | APPROVE — H-PRE-1 intact (pure subject resolution pre-store, authz in-tx); all presence directions pinned; 2 non-blocking notes (hash deploy-window 409 ≤24h TTL fail-safe → comment corrected in-commit; update/deactivate ProfileCode never populated, pre-existing, now null instead of "") | DONE |
| D22 | Dual gate round 3 arm 1 | opus (cold) | Agent background | gate-3 prompt @ FINAL SHA e03580db, range e7826c30..e03580db | ACCEPT — all 4 round-2 findings resolved at root; fresh-eyes delta clean | DONE |
| D23 | Dual gate round 3 arm 2 | GPT-5.6 Sol medium | codex exec read-only | same | round-2 findings 2–4 RESOLVED; REJECT — 1 residual CRITICAL (idemp hash trimmed subject.Key vs exact validation/persistence) | DONE |
| D24 | Implementer S8 (hash exact key) | sonnet | Agent sync | verified round-3 finding + ratified fix spec | diff on worktree; RED→GREEN (replay nil → ErrConflict) | DONE |
| D25 | Reviewer S8 | sonnet | Agent sync | S8 working-tree delta | ACCEPT — no findings; fake-store replay path traced (old trim ⇒ cached replay, nil error ⇒ test genuinely RED) | DONE |
| D26 | Dual gate round 4 arm 1 | opus (cold) | SendMessage resume (same gate agent, delta re-review) | round-3 fix summary + delta e03580db..36ccc0cf | ACCEPT — RED proof re-executed live (one-line revert → FAIL → restore, tree verified clean); 2 defer notes | DONE |
| D27 | Dual gate round 4 arm 2 | GPT-5.6 Sol medium | codex exec read-only, fresh full-range prompt | gate-4 prompt @ FINAL SHA 36ccc0cf, range e7826c30..36ccc0cf | ACCEPT — round-3 CRITICAL resolved at root; no delta-introduced findings | DONE |

## Events sent to hub

- REQUEST (contract-lock api/openapi, single CreateRouteRequest edit) — sent 2026-07-16.
- Contract-lock GRANTED by hub (verified in hub transcript): adjudication = DB check 0297 + ADR 0082 are truth, spec over-requires profile_code. Conditions: full canonical regen (option-A parity) ✔ (`go generate ./internal/modules/...` run in full; churn = approval api.gen.go only — embedded spec is per-module, so single-file churn IS the complete canonical result, not a partial regen); both-direction integration/unit pinning ✔ (template+profile_code → 400, document−profile_code → 400, template create → SQL NULL asserted in DB, document create → populated); STOP-on-unexpected-churn ✔ (none); lock releases at CLOSED.

## Slice reviews

- **S2 (F22)** — APPROVE (D5). In-place `With*` chain restores single-instance convergence (heals template ports + lifecycleEnqueuer + FastForward divergence); `Ready()` boot guard fail-fast; 7 wiring tests incl. rebuild-divergence regression. Committed 9adbbab6.
- **S3 (F2 FE)** — round 1 REJECT (D6): no test pinned `required_capability` on create-route POST body via default path. Correction (D7): `RouteAdminPage.test.tsx` asserts literal `'document.signoff'` on `routeAdminApi.createRoute` mock call payload; RED/GREEN proven. Delta re-review (D9): APPROVE — assertion on real POST body, exercises defaultStage→toStageRequests chain.
- **S4 (Go fixture sweep)** — APPROVE (D9): exactly 7 `doc.signoff`→`document.signoff` swaps across 4 files, no other changes, no stale comparisons; build + vet -tags=integration clean.
- **S1 (F18)** — APPROVE (D11). Validate() conditional exact (template never hits pattern/required; unknown kind explicit error; wrapValidation on every branch); createTx binds SQL nil only for template, `authz.Require(CapRouteManage)` unconditional in-tx both paths — only the friendly G1 per-profile policy check is skipped for template (no profile exists; DB deferrable constraint remains authoritative). Spec diff minimal (required list + 2 descriptions). Generated ProfileCode `*string` has zero consumers (handler decodes via contracts type) — no nil-deref. Mock handler test repaired (template fixture drops profile_code) + new test asserts 400 without reaching service. Non-blocking note: subject resolution moved before tx/authz — same precedent as HTTP-layer validation. Committed 5ee67417.
- **Straggler adjudication:** `internal/modules/iam/authz/authz_test.go` retains 4 `doc.signoff` literals — NOT drift. File uses a fake SQL driver (query-matcher stub on `role_capabilities`); capability string is self-consistent fixture data testing authz mechanics, registry-independent. Left untouched.

## Ladder receipts

All run from the worktree root at SHA 5ee67417 (branch tip, all 4 slices committed):

- **L0** `go build ./...` — clean.
- **L0** `go vet -tags=integration ./internal/modules/approval/... ./internal/modules/jobs/stuck_instance_watchdog/... ./apps/api/cmd/...` — clean.
- **L0** gofmt — no NEW drift: repo has pre-existing gofmt drift (~30 files incl. one S4-touched test file); verified identical drift exists at base e7826c30 (`git show base:file | gofmt -d` byte-identical diff region). None of the unit's other touched files listed.
- **L0** `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` — `0 violation(s)`.
- **L0** `.\scripts\check-governance.ps1` — `[governance-check] OK`.
- **L1** `.\scripts\test-integration.ps1 -Package "./internal/modules/approval/...","./internal/modules/jobs/stuck_instance_watchdog/..." -GoFlags "-count=1"` — PASS, all 9 test packages ok (application 25.0s, http 5.1s, contracts 1.0s, infrastructure 16.0s, idempotency 1.7s, signature 2.5s, jobs 87.6s, domain 1.0s, watchdog 13.5s). DATABASE_URL derived by runner from .env (never hand-set).
- **L1** FE `pnpm exec tsc --noEmit` — exit 0.
- **L1** FE `pnpm exec vitest run src/features/approval` — 30 files, 235/235 passed.

At FINAL SHA **36ccc0cf** (post-S8, post-convergence):

- **L1** `.\scripts\test-integration.ps1 -Package './internal/modules/approval/...' -GoFlags '-count=1'` — PASS, all 8 approval test packages ok (application 21.8s, domain 2.0s, http 5.6s, contracts 2.0s, infrastructure 14.5s, idempotency 2.8s, signature 3.5s, jobs 61.6s). DATABASE_URL derived by runner from .env. (Note: runner requires pwsh 7 — Windows PowerShell 5.1 fails parsing the script's UTF-8 literals.)
- **L0** S8: `go build ./internal/modules/approval/...`, `go vet`, gofmt on touched files — clean (D24 receipts).

## Dual gate

Round 1 at fixed SHA **5ee67417** (range e7826c30..5ee67417):

- **Arm 1 — cold Opus (D12): ACCEPT.** Root-cause confirmation on all three fixes; test honesty confirmed (revert→RED reasoning per fix); regen confirmed canonical (per-module `include-tags:[approval]` + `embedded-spec:true` in cfg.yaml — approval-only churn is the complete result, not partial drift); no invariant violations; no blocking findings. Minors: (1) toDraft empty-capability flows to backend reject (fail-honest, aligned no-fallback, no action); (2) cdReader wired twice in main.go (cosmetic).
- **Arm 2 — GPT-5.6 Sol medium via codex read-only (D13): REJECT**, 4 findings, all downstream of the F18 NULL projection:
  1. CRITICAL `postgres_approval_repository.go` scanRouteListRows — profile_code scanned into plain string; one template route breaks route-admin listing. **VERIFIED REAL → fixed in S5.**
  2. CRITICAL `route_admin_service.go` loadActiveRouteProfileCode — same NULL-scan on the update path; template route update fails pre-write. **VERIFIED REAL → fixed in S5.**
  3. HIGH FE `api-types/index.d.ts` stale — canonical FE regen (`pnpm run gen:api`) never run for the amended spec. **VERIFIED REAL → fixed in S5.**
  4. HIGH RouteSummary (openapi.yaml:7319) still requires non-nullable profile_code — cannot represent template rows' NULL projection (serializes as "" sentinel). **Outside granted lock scope → hub REQUEST sent (extend lock vs defer), holding final convergence on reply.**

Interim commits after round 1: 0bb8e716 (handler-layer document-requires-profile_code test, per hub grant layering condition), e0d0a1a6 (S5 remediation of findings 1–3).

Hub ruling on finding 4: option (a) GRANTED — lock extended to RouteSummary; sentinel "" ruled a no-fallback violation; conditions: serialization test pinning null-not-"" at projection, FE null-safety verified, dual gate re-converges on FINAL SHA (both arms). Findings 1–3 remediation accepted as reported; approval-only regen churn accepted as canonical; getTemplateDocxUrl 409 types-drift absorption accepted (spec had it at base — that IS the contract; types were stale).

**Slice S6** (commit ffcfccf5, reviewed APPROVE D17): RouteSummary.profile_code `nullable: true` (present-but-null, route_id convention) + conditional prose; canonical regens (approval api.gen.go: ProfileCode *string; FE index.d.ts: string | null); wire fix at hand-written contracts.ListRouteItem (*string, nil iff SubjectKind==template — row-local proxy exact per 0297 paired projection checks, plain ADD CONSTRAINT so legacy rows validated at apply); raw-JSON serialization test (`map[string]json.RawMessage`) pins template row `profile_code == null` and sibling document row `== "ops"` in the same response; FE: toDraft null→'' (controlled input; edit dialog never sends profile_code — UpdateRouteRequest has no such field; create still requires non-empty), RouteListTable renders '—' for null; 2 new routeDraft tests (237 total green).

Round 2 at **ffcfccf5** (both arms re-run per hub condition):

- **Arm 1 — Opus (D18): ACCEPT.** All 4 round-1 findings resolved at root (explicit dispositions, live real-DB re-execution of the F18 integration tests). Fresh-eyes sweep clean. Minor: template branch accepted present-but-empty `""` profile_code (spec says omitted). Info: capabilities.ts hand-sync = DEFER-QR-A-1 (known); audit-event-presenter `doc.signoff` = historical display label, not sweep debt.
- **Arm 2 — GPT-5.6 Sol (D19):** round-1 findings 1–4 all **resolved at root** (explicit dispositions), but **REJECT** with 4 NEW findings on the newly-enabled template-create path:
  1. CRITICAL `route_admin_idemp.go` — create hash omitted subject identity: template creates sharing Idempotency-Key+name+stages but differing subject_key silently replayed. **VERIFIED REAL → fixed S7.**
  2. HIGH contracts ProfileCode plain string — present-but-empty/explicit-null indistinguishable from omitted (overlaps Opus minor). **VERIFIED → fixed S7 (*string; null≡omitted documented tolerance).**
  3. HIGH `""` sentinel in route.config.created event + RouteResponse; event lacked subject identity. **VERIFIED (its RouteResponse schema-divergence sub-claim REFUTED: spec RouteResponse is `additionalProperties: true, required: [route_id]` — loose by design, no spec edit needed) → fixed S7 Go-side (event null + subject fields; RouteResponse *string).**
  4. MEDIUM sweep straggler `tests/integration/scenarios/concurrency_test.go:369`. **VERIFIED → fixed S7.**

**Slice S7** (commit e03580db, reviewed APPROVE D21): all four fixes; full ladder re-run green (build, vet, unit, integration approval+scenarios, FE tsc+vitest 237). Non-blocking notes recorded: idemp hash not byte-stable across THIS deploy (pre-deploy in-flight envelopes 409 fail-safe ≤24h TTL — code comment corrected to state this); update/deactivate handlers never populated RouteResponse.ProfileCode (pre-existing gap, formerly `""` now `null`, FE doesn't consume).

Round 3 (convergence, both arms at **e03580db**, range e7826c30..e03580db):

- **Arm 1 — Opus (D22): ACCEPT.** All 4 round-2 findings verified resolved at root (explicit dispositions: hash folds resolved subject, single resolution point feeds hash+createTx identically, H-PRE-1 safe; *string nil-derefs all guarded; event/RouteResponse null with raw-JSON wire tests; straggler swept, both keeps verified correct). Fresh-eyes on ffcfccf5..e03580db delta: no introduced defects. Build + touched unit packages green.
- **Arm 2 — GPT-5.6 Sol (D23):** round-2 findings 2–4 RESOLVED; **REJECT** — finding 1 INCOMPLETE: 1 residual CRITICAL (`route_admin_idemp.go:66`): hash used `strings.TrimSpace(subject.Key)` while validation (domain.Subject.Validate rejects only empty), persistence (resolveCreateRouteSubject assigns raw `in.SubjectKey`), and events use the EXACT key — `"tmpl-a"` vs `" tmpl-a "` are distinct persisted subjects sharing one hash → reused Idempotency-Key wrongly REPLAYS instead of conflicting. **VERIFIED REAL by direct reads (no trim/format check on SubjectKey anywhere in contract Validate, resolution, or domain) → fixed S8.**

**Slice S8** (commit **36ccc0cf**, reviewed ACCEPT D25): hash the EXACT resolved key (trim dropped) — hash identity now matches validation/persistence/event identity byte-for-byte; doc comment amended (padded-key pre-deploy envelopes covered by the same documented ≤24h 409 fail-safe); `TestRouteAdminCreate_IdempotencyKeyConflict_TemplateWhitespaceSubjectKey` added (RED under old trim: fake-store replay returned nil; GREEN: `idempotency.ErrConflict`); build/vet/gofmt/full application package green. Silent trim-at-resolution deliberately rejected (no-fallback: never mutate client input); subject_key FORMAT validation = DEFER-QR-A-4 (pre-existing looseness).

Round 4 (convergence, both arms at FINAL SHA **36ccc0cf**, range e7826c30..36ccc0cf): **CONVERGED — double ACCEPT.**

- **Arm 1 — Opus (D26): ACCEPT.** Round-3 CRITICAL resolved at root: hash line now bare `subject.Key`, identity matches validation/resolution/persistence/event byte-for-byte; doc comment forbids reintroducing the trim. RED proof re-executed live (temporarily reverted the one line → whitespace test FAILs with nil error → restored; tree verified clean, `git status` confirmed post-run). Fresh-eyes on 2-file delta: no introduced defects; remaining hash trims safe (template profileCode always ""; document profileCode pattern-validated, no whitespace admissible). Defer notes: `name` trim-vs-raw same class (pre-existing at base, folded into DEFER-QR-A-4); subject_key format validation concurs out of scope.
- **Arm 2 — GPT-5.6 Sol (D27): ACCEPT.** Round-3 CRITICAL RESOLVED at root (`route_admin_idemp.go:76` hashes exact resolved key; validation, SQL persistence, event emission same byte-exact value). No delta-introduced findings. HEAD verified 36ccc0cf290e; delta 2 files +59/−1; `git diff --check` clean; subject-key format validation deferred as pre-existing. (Read-only sandbox denied Go temp build dir → static-evidence verdict, consistent with all prior codex rounds; test execution covered by Opus arm + D24/D25 receipts.)

## Defers

- DEFER-QR-A-1 — capability registry→FE codegen (hand-synced enumeration meta-defect). Owner: hub.
- DEFER-QR-A-2 — live template-signoff + document-signoff lifecycle E2E (J3/J5 re-drive is hub-owned post-merge).
- DEFER-QR-A-3 — 0296 compat trigger `default_approval_subject()` retirement (migration DEBT, not QR-A).
- DEFER-QR-A-4 — subject_key FORMAT validation at contract layer (whitespace-padded/garbage keys currently accepted verbatim; pre-existing looseness since M3 kernel — hash/persistence/event identity is now byte-consistent, but a padded key persists as a distinct, likely-unfindable subject). Companion (Opus round-4 defer note): `name` is trimmed in create+update hashes but persisted raw and not whitespace-validated — same divergence class, predates branch at e7826c30, all subject kinds. Owner: hub.
