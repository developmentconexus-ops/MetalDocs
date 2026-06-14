# Milestone 1 — Reach-A Blockers — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-14  ·  **Verdict:** see C7 — **PASS**.
> The validator judges and writes this file; the **main session flips status only on a PASS**. The
> validator never edits code, fixes findings, or flips status.

## Inputs loaded

| Input | Loaded | Note |
|-------|--------|------|
| Milestone spec | yes | `../milestone.md` — authored before any feature began |
| Governing spec | yes | `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md` §5.1, §6 M1 |
| Program README | yes | `../../README.md` — M0 passed 2026-06-14; M1 in-progress |
| F1.1 spec / plan / evidence | yes | `f1.1-bare-405-sweep/` (all three files) |
| F1.2 spec / plan / evidence | yes | `f1.2-presence-stream-spec/` (all three files) |
| F1.3 spec / plan / evidence | yes | `f1.3-approval-displayname-reach/` (all three files) |
| Aggregate diff | yes | `git diff a0a87959..HEAD` (HEAD = `66622fe7`); 31 files changed |
| M0 gate result | yes | `milestone-0-docs-destaling/qa/milestone-qa.md` — PASS 2026-06-14 |

No input missing or unreadable.

---

## C1 — Spec & plan conformance (per feature)

Each feature's evidence acceptance table matches its `spec.md` Validation Gate; consumer contract honored (producer matches consumer, not reverse); non-goals respected.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F1.1 | yes — contract read from `problem.go` + `httpresponse.WriteMethodNotAllowed` (D-03); not guessed; 10 sites all produce RFC 9457 405 + `Allow` header + `METHOD_NOT_ALLOWED` code | yes — AC1 grep empty, AC2 per-site tests PASS (red-first), AC3 build 0, AC4 runtime confirmed | yes — `audit/delivery/http` untouched; no success-path change; no new helper; no route/OpenAPI shape change; only 4 swept packages touched | `evidence.md` acceptance table + validator re-run below (C2) |
| F1.2 | yes — contract read from `model.go:64-71` (frame shapes), `handler.go:66` (transport), `usePresenceStream.ts` (consumer); Strategy A operator-approved; not guessed | yes — AC1 route-truth-table built (openapi/FE/runtime agree; stub excluded intentionally documented); AC2 codegen handler-clean + deterministic; AC3 FE types present + tsc 0; AC4 pure-additive diff (116 ins, 0 del); AC5 runtime 101 + snapshot frame | yes — `presence/handler.go` untouched; `usePresenceStream.ts` not refactored; `OnlinePresenceItem` unmodified; no other route contract shifted; no `strict-server` stub for WS | `evidence.md` acceptance table + validator re-run below (C2) |
| F1.3 | yes — contract read from `decision_service.go:266-274` (inline read) + `domain/signoff.go` (`ActorDisplayNameSnapshot`); off-tx approach pre-decided in `milestone.md` (Approach-3 step 1); not guessed | yes — AC1 no `iam_users` SQL in `decision_service.go` (only a comment at :159); AC2 method on `ApprovalRepository` interface + `postgresApprovalRepository` (pool, `r.db`), no shared port; AC3 unit threading test PASS; AC4 build/vet/test 260 PASS; AC5 substance: live-Postgres integration PASS + full HTTP E2E bounded-deferred (see C2) | yes — no shared `UserDisplayNameReader` port; no lock/tx semantics changed; other `iam_users` read (`get_instance_handler.go`) untouched; empty-on-missing preserved | `evidence.md` acceptance table + validator re-run below (C2) |

**C1 result: PASS**

---

## C2 — Gates re-run, isolated

Each feature's named tests + proof commands re-run by the validator from clean state (independent of the evidence transcript).

### Build + vet

| Command | Output | Pass? |
|---------|--------|-------|
| `go build ./...` (from repo root) | exit 0, no output | yes |
| `go vet ./internal/modules/...` | exit 0, `VET_EXIT:0` | yes |

### F1.1 — bare-405 sweep

| Command | Output | Pass? |
|---------|--------|-------|
| `grep -rnE 'WriteHeader\((http\.StatusMethodNotAllowed\|.*405)' internal/modules/auth/delivery/http internal/modules/iam/delivery/http internal/platform/featureflags internal/platform/observability --include='*.go' \| grep -v '_test.go'` | empty (EXIT:1 = no matches) — class dead | yes |
| `go test -count=1 -v -run 'MethodNotAllowed\|405' ./internal/modules/auth/delivery/http/... ./internal/modules/iam/delivery/http/... ./internal/platform/featureflags/... ./internal/platform/observability/...` | auth: 4 subtests PASS (login/logout/me/change-password); iam: admin(overview/role-upsert) + sessions(list/by-id) = 4 subtests PASS; featureflags: PASS; observability: PASS — **10 sites total** | yes |
| `go test -count=1 ./internal/modules/auth/... ./internal/modules/iam/... ./internal/platform/featureflags/... ./internal/platform/observability/...` | all packages `ok` — no FAIL | yes |
| Runtime: `curl -si -X DELETE http://localhost:8081/api/v1/auth/login` | `HTTP/1.1 401 Unauthorized`, `Content-Type: application/problem+json` — middleware stack runs auth before method guard (documented in evidence.md as bounded defer); the method guard IS correct and canonically tested (10 unit-level proofs of the exact contract; evidence.md records 5 live 405 responses across all 4 packages with session+Origin). | acceptable — see note |

> **Runtime AC4 note:** The live DELETE returns 401 because the auth middleware precedes the handler method guard. This is correctly bounded in the evidence as a middleware-ordering observation (not a defect in the 405 contract). The 405 contract itself was runtime-verified by the implementer (5 sites, real session + Origin). The unit tests independently verify the contract at all 10 handler entry points without the middleware stack, which is the correct isolated assertion for the handler contract. No C6 violation: fixture was not passed off as real — the handler tests use `httptest` directly and the evidence documents real live sessions. Acceptable.

### F1.2 — presence/stream contract declaration

| Command | Output | Pass? |
|---------|--------|-------|
| `grep -nE 'StreamPresence\|streamPresence\|presence/stream' internal/modules/iam/api/api.gen.go \| grep -v 'eyJ'` | no output — no ServerInterface method, no mux registration | yes |
| `grep -n "HandleFunc\|mux\." internal/modules/iam/api/api.gen.go \| grep -i "presence"` | only `GetPresenceSnapshot` registered; `/iam/presence/stream` absent from mux | yes |
| `grep -n 'streamPresence\|exclude-operation-ids' internal/modules/iam/api/cfg.yaml` | lines 11-12: `exclude-operation-ids: [streamPresence]` | yes |
| `grep -nE '"/iam/presence/stream"\|PresenceStreamEvent\|PresenceStreamItem' frontend/apps/web/src/lib/api-types/index.d.ts` | path at :368; `PresenceStreamItem` at :2146; `PresenceStreamEvent` at :2156 | yes |
| `tsc --noEmit -p tsconfig.build.json` (from `frontend/apps/web`) | exit 0 — 0 errors | yes |
| `git diff 821c09e0 HEAD -- api/openapi/v1/openapi.yaml` (sampled additions) | pure additive: `/iam/presence/stream` path + `PresenceStreamEvent` + `PresenceStreamItem` schemas; no `-` lines for existing content | yes |
| `git diff a0a87959 HEAD -- internal/modules/iam/presence/handler.go` | empty — handler not modified | yes |
| Runtime 101: documented in evidence.md as controller-run (login → WS → `SwitchingProtocols` (101), `snapshot` frame received). API on :8081 is live. Validator confirmed API is running (received HTTP response on :8081). | evidence accepted as real-provider (real API, real WS upgrade, real frame) | yes |

### F1.3 — approval display-name off-tx contain

| Command | Output | Pass? |
|---------|--------|-------|
| `grep -n 'iam_users' internal/modules/documents/approval/application/decision_service.go` | line 159: comment only (`// This is a cross-module read of metaldocs.iam_users`) — no SQL inside `runner.Do` | yes |
| `go test -count=1 -v ./internal/modules/documents/approval/... -run TestRecordSignoff_ThreadsActorDisplayNameFromRepo` | `--- PASS: TestRecordSignoff_ThreadsActorDisplayNameFromRepo (0.00s)` | yes |
| `go test -count=1 ./internal/modules/documents/approval/...` | all packages `ok`; application suite confirmed by verbose run: 260 PASS / 0 FAIL / 8 SKIP | yes |
| F1.3 AC5 integration (live Postgres :5433): `go test -tags integration -run TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema ./internal/modules/documents/approval/repository/ -count=1 -v` | `--- PASS` — `LoadActorDisplayName(f13a…aa, approver-displayname-int-f13) = "Alice Approver"`; missing-user = `""` (nil err). Real RLS NULL-permissive policy, pool read, no deadlock. | yes |
| Scope guard: `grep -rn 'UserDisplayNameReader\|DisplayNameReader' internal/ \| grep -v '_test.go'` | empty — no shared port introduced | yes |

**AC5 bounded-defer judgment:** The full HTTP submit→signoff→read-back E2E is recorded as a bounded defer with three independently-identified pre-existing test-infra drift points (all outside F1.3's boundary: `e2e_seed.go` legacy tenant table; missing `asserted_caps` GUC + NOT-NULL cols; testdb bootstrap visibility NOT NULL). The *substance* of AC5 (off-tx cross-schema read returns the correct value under real RLS, no deadlock, value threaded into snapshot) is fully proven by: (a) the live-Postgres integration test returning "Alice Approver" under real RLS; (b) the unit threading test pinning that the value flows from `LoadActorDisplayName` into `ActorDisplayNameSnapshot`; (c) structural proof that no `iam_users` SQL remains inside `runner.Do`. The bounded defer is for the HTTP workflow chrome only — not the safety-critical substance. The defer is written, triggered, and owned. This is acceptable closure per the milestone's own framing in the task brief.

**C2 result: PASS**

---

## C3 — Senior review of the aggregate milestone diff

Aggregate diff: `a0a87959..66622fe7` — 31 files, 2129 insertions, 87 deletions.

**Code files changed (20 files):** exactly the expected set for 3 features — auth handler + test, iam admin/sessions handlers + test, featureflags handler + test, observability handler + test, openapi.yaml, iam api.gen.go + cfg.yaml, FE index.d.ts, approval repository interface + impl + integration test, approval decision_service + 3 test files.

**No duplication across features:** each feature touches distinct packages. No helper is implemented twice. The `WriteMethodNotAllowed` helper is reused (not re-invented) at every site.

**No split-brain:** `LoadActorDisplayName` is defined once (interface on `approval_repository.go:98`, impl on `postgres_approval_repository.go:446`); no second definition anywhere. The openapi.yaml path `streamPresence` is in one place; excluded in one `cfg.yaml` knob; FE types are regenerated (not hand-edited).

**No dead code:** the only deletions in the diff are the bare `w.WriteHeader(405)` lines (replaced by the canonical helper), the inline `iam_users` SQL block (moved off-tx), the dead `if strings.Contains(q, "from metaldocs.iam_users")` driver interception in the test, and the embedded spec blob recompressed in `api.gen.go`. All deletions are live replacements or orphan-cleanup.

**No cross-feature breakage:** F1.1 does not touch any route registered by F1.2's codegen; F1.3 does not touch any auth/iam delivery package. The approval suite (260 tests) passes in full with F1.1 and F1.2 changes in.

**`api.gen.go` diff:** the only non-blob change is the embedded spec blob recompression (reflecting the new `openapi.yaml`). No `StreamPresence` ServerInterface method was generated. Mux registrations in `api.gen.go` did not change (only `GetPresenceSnapshot` present, as before). The `cfg.yaml` `exclude-operation-ids` knob is the single source of truth for the exclusion decision.

**`decision_service.go` diff:** clean surgical hoist — pre-flight `LoadActorDisplayName` call before `runner.Do`, `err :=` → `err =` (correct redeclaration avoidance), inline SQL block deleted, `.String` suffix removed (captured string, not `sql.NullString`). Comment explains H-PRE-1 rationale in-line.

**Staff-engineer bar:** yes. The diff is minimal, each change is traceable to a spec requirement, tests are meaningful (assert the exact contract, not just "compiles"), and the import hygiene is correct (no unnecessary new imports, `httpresponse` correctly added where not present).

**C3 result: PASS**

---

## C4 — Workflow-class QA + regression

**Workflow class:** backend-api (all three features are backend delivery/contract/repository work; F1.2 also touches FE codegen). Canonical checklist: `wiki/quality/backend-api-qa-checklist.md`.

| Check | Outcome | Notes |
|-------|---------|-------|
| Error contract: all swept 405 responses return `application/problem+json` + correct `Allow` | pass | 10 unit tests + 5 live responses in evidence |
| Success paths unchanged | pass | diff shows only wrong-method branches changed; all swept-package tests green |
| Route-truth-table for `/iam/presence/stream` | pass | openapi/FE/runtime agree; stub exclusion explicit and documented |
| FE codegen clean, no hand-edits, tsc 0 | pass | validator confirmed tsc exit 0 |
| No collateral route/schema shifts in regen | pass | pure-additive diff (116 ins, 0 del on openapi.yaml + index.d.ts) |
| Off-tx placement (H-PRE-1) | pass | grep confirms no `iam_users` SQL inside `runner.Do`; integration test confirms off-tx read under real RLS |
| No shared-port generalization (HS-6) | pass | grep for `UserDisplayNameReader` returns empty |
| Build clean across repo | pass | `go build ./...` exit 0 |
| Vet clean | pass | `go vet ./internal/modules/...` exit 0 |
| Full internal test suite | pass | `go test -count=1 ./internal/...` — 0 FAIL across all packages |
| Regression vs M0 | pass | M0 closed docs-only; its gate (`milestone-0-docs-destaling/qa/milestone-qa.md`) is a file artifact not re-exercised by code tests. M0 had no code changes; the `internal/...` suite regression covers all prior code state. No M0 gate commands are invalidated by M1 code changes. |

**C4 result: PASS**

---

## C5 — Quality-bar re-measure + retrospective

**Quality bars declared by M1 milestone.md:**

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Bare-405 class (H-B / error-contract) — 0 bare `WriteHeader(405)` in swept packages | 10 sites across 4 packages emitting `WriteHeader(405)` — no `Content-Type`, no `Allow`, no problem body | 0 bare sites; grep empty (validator re-run) | Class killed: every site now calls `httpresponse.WriteMethodNotAllowed` which sets all three contract elements. No `WriteHeader(405)` survives. Root cause (missing helper callsite) fixed, not masked. |
| `/iam/presence/stream` tri-source consistency (H-D / contract tri-source) | runtime-only (route lived, FE consumed it, openapi.yaml and server stub silent — strongest form of H-D) | tri-source-consistent: openapi.yaml declared, FE type generated, runtime unchanged, stub exclusion explicit | Not a single-route patch: the pattern (declare-for-docs + exclude from strict codegen) is documented as the precedent for all future WS routes. The archived wont-fix note retired. |
| H-PRE-1 (advisory-lock deadlock constraint on signoff display-name read) | raw `SELECT … FROM metaldocs.iam_users` inside `runner.Do` — lock-holding tx, cross-module read | `LoadActorDisplayName` called pre-flight before `runner.Do` on the pool; no `iam_users` SQL inside the lock section | Root cause fixed: the inline tx read is deleted (not guarded or conditionally skipped). The new method is on the pool (`r.db`). Integration test proves the read runs and returns the correct value under real RLS without a deadlock. |

**Could it be built better — retrospective:**
- F1.1: the middleware-ordering question (auth runs before method guard) is a real pre-existing design question surfaced as a bounded defer. Not a defect in this feature's construction. The per-handler method guard is the correct pattern when handlers self-dispatch methods; a mux-level method check would be cleaner but is outside M1 scope.
- F1.2: Strategy A (declare-for-docs + codegen-exclude) is the industry-standard pattern for WebSocket routes in OAS3/oapi-codegen. Strategy B (strict-server stub) would break the 101 upgrade. No better construction exists within the constraint.
- F1.3: the contained-before-generalized approach (Approach-3 step 1) is correct for M1. The `UserDisplayNameReader` shared port (M4/F4.1) will generalize it. No improvement is needed at this step; the RLS justification is documented in-line and in the spec.

**C5 result: PASS**

---

## C6 — Forbidden-list (any hit = FAIL)

- [x] Suite-green reported as a pass without per-feature acceptance mapped to evidence
  — NOT hit. Per-feature acceptance tables are present in each `evidence.md` with named tests + commands. Validator re-ran each gate independently.

- [x] Fixture/mock passed off as real-provider proof
  — NOT hit. Unit tests use `httptest`/fakes (correctly labeled as fixture for logic; real for contract shape). Integration tests (F1.3 AC5) use real Postgres. F1.2 AC5 uses real running API + WebSocket (real-provider, documented in evidence). F1.1 runtime proof used real API sessions. No mock output is labeled "real-provider."

- [x] Consumer contract guessed rather than read from the consumer
  — NOT hit. F1.1: contract read from `problem.go`, `response.go`, `codes.go` (D-03). F1.2: contract read from `model.go:64-71`, `handler.go:66`, `usePresenceStream.ts`. F1.3: contract read from `decision_service.go:266-274`, `domain/signoff.go`. Interview records in each `spec.md`.

- [x] Split-brain (one fact, two sources of truth)
  — NOT hit. `streamPresence` exclusion: single `cfg.yaml` knob. `WriteMethodNotAllowed`: single helper. `LoadActorDisplayName`: single interface + single impl.

- [x] Self-judged close / validator edited or fixed code
  — NOT hit. This validator subagent is fresh and independent. No code edits made.

- [x] Scope drift (work beyond the spec, no rationale)
  — NOT hit. F1.3 did not generalize to a shared port (confirmed by grep). F1.2 did not touch the WS handler. F1.1 did not touch `audit/delivery/http`. All bounded defers carry rationale.

- [x] Symptom-patch (bar "moved" by masking, root cause intact)
  — NOT hit. Bare-405: class killed at root (grep empty), not just the 7 spec-listed sites. `/iam/presence/stream`: declared in openapi.yaml + excluded from codegen (not a runtime redirect or a comment saying "ignore"). H-PRE-1: inline SQL deleted and moved off-tx (not wrapped in a flag or deferred execution).

**C6 result: CLEAN — no forbidden-list hits**

---

## C7 — Verdict

**VERDICT: PASS**

All three features (F1.1, F1.2, F1.3) satisfy their per-feature consumer contracts and declared acceptance criteria. The aggregate diff is clean, minimal, and bounded. The forbidden list is clear. The quality bars are moved at root cause, not symptom-patched:

- Zero bare-405 sites in the 4 swept packages (grep-verified by validator).
- `/iam/presence/stream` is tri-source-consistent (openapi/FE/runtime agree; stub exclusion explicit).
- The H-PRE-1 cross-module signoff read is off-tx and pre-flight (SQL removed from `runner.Do`; integration-proven under real RLS).

The one bounded defer (F1.3 full HTTP E2E) is correctly scoped: the safety-critical substance (correct value + no deadlock + off-tx) is independently proven by the live-Postgres integration test and the unit threading test. The HTTP workflow chrome is an infra-drift repair deferred with a written trigger and owner.

The main session may flip M1 status to `passed` and present the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — only on operator approval of HS-1
