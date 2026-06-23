# Milestone 3 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-22  ·  **Verdict:** see C7 — **PASS**.
> Inputs loaded: `milestone.md`; `f3.1/f3.2/f3.3/f3.4` `spec.md`+`plan.md`+`evidence.md`; program
> `README.md`; governing `mission.md`; aggregate code diff `4be6479c..0a99fa73` (M2-close → HEAD).
> All inputs present and readable — no fail-fast on missing input.

## C1 — Spec & plan conformance (per feature)

Every feature has `spec.md` (Approval line filled: 2026-06-22 / leandrotca), a populated Interview
record (fail-closed), an execution-shaped `plan.md`, and an `evidence.md` whose acceptance table maps
row-for-row to the spec Validation Gate. Non-goals respected (verified against the aggregate diff — no
emit-site semantic change, no channels/prefs/SSE/approver-routing). Note: F3.1/F3.2 specs were authored
under ADR-0043 (projector + two compensating views); F3.3 was operator-re-scoped to ADR-0044 (typed
River-job domain events, both views eliminated). The re-scope is recorded in `milestone.md`, the program
README HS table, and F3.3 `spec.md`/`research-and-design.md` — a documented, approved deviation, not drift.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F3.1 contract | YES — `Notification` schema is the snake_case mirror of FE `NotificationItem`; `event_type` open string, `status` closed 3-enum (openapi.yaml:4205-4227) | YES — api-lint 0; generated Go+FE types present; cap registered + deferred across 4 files | YES — no handler/migration/wire/emitter | api.gen.go + ADR-0043 |
| F3.2 backend | YES — `Handler` implements F3.1 `StrictServerInterface`; self-scope in SQL predicate (`recipient_user_id=$caller`) + tier-1 cap | YES — 7/7 live-PG subtests (re-run, see C2) | YES — `git diff publish_service.go` empty for F3.2; no emitter/FE | evidence.md |
| F3.3 emitter | YES — worker imports `documents/domain` (sanctioned `<module>/domain` cross-module layer); domain file has no `river` import; port takes `db.Tx`, `*sql.Tx` assertion isolated to adapter | YES — 7/7 live-PG fanout subtests (re-run, see C2); 5 additive in-tx enqueues, terminal-only author events | YES — additive enqueue only; no audit-payload edit, no new event table | evidence.md + fanout_worker.go |
| F3.4 wire | YES — components consume generated types directly; no snake→camel mapper (tsc clean) | YES — read-all 7/7 live PG; 33/33 vitest; noop deleted; route removed; both reviewers APPROVE-with-nits | YES — only backend touch is additive `read-all`; legacy retired; SSE left as-is | evidence.md |

## C2 — Gates re-run, isolated

Re-run by the validator from clean state (not trusted from transcripts). Note F3.3 `evidence.md`
**honestly labeled** its integration tests "compile-verified; runtime deferred to CI" — no fixture was
passed off as real. The deferred live run is now independently executed here and PASSES.

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| build | `go build ./...` | exit 0 | YES |
| F3.1/F3.4 contract | `go run ./scripts/api-lint -strict ./api/openapi/v1/openapi.yaml .` | `0 violation(s)` | YES |
| F3.2 + F3.4 (read surface + read-all) | `go test -tags integration ./internal/modules/notifications/... -run TestNotifications` (live PG via C:\tmp runner) | `--- PASS: TestNotifications (27.75s)` — 7/7 subtests incl. self_scope_isolation, mark_all_read | YES |
| F3.3 (fanout) | same runner (prefix-match also runs fanout) | `--- PASS: TestNotificationsFanoutWorker (119.01s)` — 7/7: published/superseded/obsoleted→readers, approved/rejected→submitter, redelivery_is_noop, no_cd_link_skips | YES |
| integration exit | runner | `INTEGRATION_EXIT=0` | YES |
| all-Go unit suite | `go test ./...` | `GOTEST_EXIT=0`, FAIL_LINES=0 (incl. unchanged approval/publish service tests) | YES |
| 6 CI guards | `go test ./tools/cilint/...` | `ok metaldocs/tools/cilint/internal/analyzers` | YES |
| F3.4 typecheck | `npx tsc --noEmit -p tsconfig.build.json` | `TSC_EXIT=0` (generated types, no mapper) | YES |
| F3.4 FE tests | `npx vitest run src/features/notifications` | `7 passed (7) … 33 passed (33)` | YES |
| F3.4 FE build | `npx vite build` | `✓ built in 1m 5s`, exit 0 | YES |

## C3 — Senior review of the aggregate milestone diff

Aggregate code diff `4be6479c..0a99fa73`: 64 files, +4169/−262 (excluding docs/wiki). Reviewed as one unit.

- **Emit sites (HS-2 critical):** `publish_service.go` / `supersede_service.go` / `obsolete_service.go` /
  `decision_service.go` each gain ONLY an additive in-tx enqueue after the existing `events.Emit`, gated
  on `lifecycleEnqueuer != nil`; author events fire only at `result.InstanceApproved`/`InstanceRejected`
  (terminal). No reordering, no behavior change — confirmed by the unchanged approval service tests
  passing in `go test ./...`.
- **No split-brain on facts:** notifications owns its table; the worker reads cross-module data only via
  the published `metaldocs.v_cd_obligated_readers` view (`fanout_worker.go:58`) + the `submitted_by`
  carried in the event payload — no base-table reads (`hgcrossmodule` green). The domain-event contract
  lives once, in `documents/domain/notification_events.go`.
- **No dead code from the re-scope:** ADR-0043's two compensating views were never built (eliminated by
  ADR-0044 before code); no superseded-approach residue in the diff.
- **One non-fatal finding (doc nit, not split-brain):** the OpenAPI `Notification.event_type` description
  (openapi.yaml:4216) still names the superseded ADR-0043 strings (`document_published`, `signoff_recorded`,
  `signoff.rejected`), whereas the runtime emits the ADR-0044 constants (`document.published/.superseded/
  .obsoleted/.approved/.rejected`, `notification_events.go:30-34`). Because `event_type` is an **open
  string** (not an enum), the generated type is `string` regardless — this is a stale descriptive comment,
  not a second source of truth for the contract. Recommend a one-line docstring refresh (carry to a future
  touch); it does not gate this milestone. F3.1 `evidence.md` likewise mentions the old-style strings,
  same cosmetic lag.
- Staff-engineer bar met? **YES** (the one finding is a comment-text refresh, not a structural defect).

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| BE canonical (`backend-api-qa-checklist` + 6 CI guards + api-lint + integration) | pass | api-lint 0; cilint ok; 14/14 live-PG subtests; build/vet/test green |
| FE canonical (`screen-definition-of-done` D2 + `screen-qa-checklist`) | pass-with-defers | both reviewers APPROVE-with-nits (nits resolved at root, on record in F3.4 evidence); runtime functional pass recorded; bounded defers (Playwright E2E, popover focus-trap) have written triggers/owners |
| Regression vs M0 | all still pass | `/notifications` dead stub removed (18→17 routes); tracker updated — improves M0's "0 dead-stub routes" bar, no regression |
| Regression vs M1 | all still pass | dashboard untouched; `go test ./...` green |
| Regression vs M2 | all still pass | sacred views `v_cd_grantee`/`v_cd_obligated_readers` migrations diff-empty (`4be6479c..0a99fa73`); worker only reads the published view; distribution untouched |
| Publish/approval semantics | unchanged | additive enqueue only; existing approval/publish service tests pass unmodified in `go test ./...` |
| FE suite baseline | held | notifications subset 33/33; full `make test` pre-existing unrelated failures (InboxPage/DocumentEditorPage/templates.create) confirmed node_modules junction drift, not introduced by M3 |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Notifications noop stub (D2 / mission §8) | page renders hard-coded `[]` behind a noop `notifications.ts`; no backend | met | `grep -nE "items: \[\]\|never\[\]\|Stubbed pending" notifications/api/notifications.ts` = **0** (noop deleted at root, file rewritten to typed `api.GET/POST`); `NotificationsPage`+route removed; real producers (F3.1–F3.3) wired; mark-read + mark-all flip state — proven live (14/14 integration subtests + F3.4 runtime row). Root cause fixed, not flag-hidden. |
| Live-data production | no producer existed | met | 5 typed domain events emitted in-tx + 1 idempotent fan-out worker; `published/superseded/obsoleted → v_cd_obligated_readers`, `approved/rejected → submitted_by`; redelivery is a no-op (re-run confirmed). |

- Could it be built better? Two small carry-forwards (neither unsound, neither gates M3): (1) refresh the
  OpenAPI `event_type` docstring to the ADR-0044 strings (C3 nit). (2) The fanout integration tests call
  `worker.Work` directly rather than driving emit-site → River runtime → worker end-to-end; per-site
  enqueue is covered by unit tests + the wired binaries build, but a future full-loop integration test
  would tighten the seam. Both are next-touch / parked-mission inputs, not defects.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — clean (per-criterion mapping + per-subtest re-run)
- [ ] Fixture/mock passed off as real-provider proof — clean (F3.3 honestly labeled the defer; live run independently executed here)
- [ ] Consumer contract guessed rather than read from the consumer — clean (consumer-contract-first from `NotificationItem`; `<module>/domain` sanctioned import)
- [ ] Split-brain (one fact, two sources of truth) — clean (event constants single-sourced; the stale docstring is descriptive prose, not a contract)
- [ ] Self-judged close / validator edited or fixed code — clean (fresh validator; verdict file only)
- [ ] Scope drift (work beyond the spec, no rationale) — clean (diff = exactly F3.1–F3.4; ADR-0044 re-scope is documented + operator-approved)
- [ ] Symptom-patch (bar moved by masking, root cause intact) — clean (noop deleted at root; live producers wired)

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass: **code-wise** (additive-only emit-site edits, clean module boundaries, no
  split-brain, no dead code, generated-type consumption) and **function-wise** (14/14 live-PG integration
  subtests re-run from clean state prove real per-recipient fan-out + self-scope + idempotent mark-read/
  mark-all; FE renders real rows end-to-end). One non-fatal doc-comment nit recorded (C3/C5) for a future
  touch — it does not gate the milestone.
- Handed back to the main session to flip status and present the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — only on PASS + operator HS-1
