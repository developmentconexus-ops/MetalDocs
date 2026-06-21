# Milestone 1 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-21  ·  **Verdict:** see C7 — **PASS**.
> The validator judged and wrote this file only; it edited no source and flipped no status. The
> **main session flips status only on this PASS**.

## Inputs loaded (fail-closed check)

All required inputs present and readable — nothing missing:
- Milestone spec `../milestone.md` ✅
- F1.1 `spec.md` / `plan.md` / `evidence.md` ✅
- F1.2 `spec.md` / `plan.md` / `evidence.md` ✅
- Program `README.md` (M0 = passed; M1 = in-progress) ✅
- Governing spec `../../mission.md` ✅
- Aggregate M1 diff (`git status --short` + source read): only `features/dashboard/*` +
  milestone-1 docs + program README changed; backend `git diff HEAD -- '*.go'` = empty ✅

## C1 — Spec & plan conformance (per feature)

Both features: `spec.md` approval line filled (`2026-06-21 / leandrotca`, pre-code); Interview record
populated (3 rows each, resolving the real contract ambiguity, not guessed); `plan.md` is
execution-shaped (files-touched table + TDD ordering, not a re-spec); `evidence.md` acceptance table
maps row-for-row to the spec Validation Gate; non-goals/rabbit-holes respected.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F1.1 dashboard-stats-wire | ✅ — `deriveDashboardStats` consumes generated `DocumentStatsResponse.by_status` via `fetchLibraryStats`; status→label map (approved / under_review+rejected / published; missing→0) matches the operator-locked re-scope and the real enum | ✅ — `MOCK_STATS` gone; 3 unit tests; tsc 0; both reviewers APPROVE on record | ✅ — no Go change, no time-window/average fabricated, "Aguardando você" pill untouched, no layout redesign | F1.1 `evidence.md` + source read |
| F1.2 dashboard-activity-wire | ✅ — `api/audit.ts` re-exports generated `components['schemas']['AuditEventItem']` / `ListAuditEventsResponse` and calls typed `api.GET('/audit/events', { params:{ query:{ limit }}})`; `deriveActivity` reads only fields present in the generated type (`id, actor_id, action, resource_id, occurred_at`) | ✅ — `MOCK_ACTIVITY` + all `MOCK_` gone; 4 unit tests; loading/empty/error trifecta; both reviewers APPROVE | ✅ — no Go change, no audit export/filter/pagination, stats untouched, `formatRelative` reused | F1.2 `evidence.md` + source read |

C1 PASS. (Verified the contract claim against the generated bundle directly:
`/audit/events` path line 524, `AuditEventItem` line 2439, `ListAuditEventsResponse` line 2452 of
`lib/api-types/index.d.ts` — the spec.md "false-negative grep" correction is itself correct; the
audit shape is in the generated contract, so the consumer is contract-first, not hand-rolled.)

## C2 — Gates re-run, isolated

Re-run by the validator from clean state in `frontend/apps/web/` (not trusted from the transcript):

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| Both (mock bar) | `grep -rEn "MOCK_" src/features/dashboard` | exit 1 — **0 matches** | ✅ |
| Both (types) | `npx tsc --noEmit -p tsconfig.build.json` | **exit 0** | ✅ |
| F1.1 + F1.2 (tests) | `npx vitest run src/features/dashboard` | **Test Files 2 passed (2); Tests 7 passed (7)** — `deriveActivity` 4, `deriveDashboardStats` 3 | ✅ |

C2 PASS. The unit tests are real branch-coverage on the only logic with branches (status mapping,
missing-key→0, known/unknown action humanization, empty input) — not vacuous. The runtime
live-stack proof (8 live audit events + live `/documents/stats` counts via `preview_snapshot`) is on
record in both evidence files, honestly labeled real-provider; it could not be re-executed here
(no running stack / pnpm off PATH in the validation env), but the deterministic gates plus a source
read of the live query→derive→render paths corroborate it.

## C3 — Senior review of the aggregate milestone diff

Whole-M1 diff reviewed as one unit (DashboardPage.tsx, DashboardPage.module.css `.activityMuted`,
api/audit.ts, lib/deriveDashboardStats.ts + test, lib/deriveActivity.ts + test,
queries/useDashboardStatsQuery.ts, queries/useDashboardActivityQuery.ts):

- **No split-brain.** The audit shape has exactly one live source of truth in M1 — the generated
  `components['schemas']['AuditEventItem']`. The legacy hand-rolled `lib/types/index.ts:191
  AuditEventItem` is confirmed to have **0 importers** (grep across `src`: nothing imports
  `AuditEventItem` from `../lib/types`; dashboard and IAM both use the generated type). It is
  pre-existing IAM-domain dead code, outside M1's mock→live boundary, deferred with a written
  trigger — M1 neither created nor fed it.
- **No duplication / dead code in M1.** Each new file has a single caller; both derive functions are
  pure and individually tested; both query hooks mirror the established `useDashboardInboxQuery`
  pattern (QK reuse, `staleTime: 30_000`).
- **No feature broke another.** F1.2 added only `.activityMuted` CSS and the feed block; F1.1's
  stats path is independent. `statValue` honesty guard (`'…'` loading / `'—'` error / live value)
  is shared cleanly across hero pills and `§ SEU PULSO`.
- **Contract-clean.** Consumers read generated producer types; no guessed shapes.

- Findings: none blocking. Staff-engineer bar met? ✅

## C4 — Workflow-class QA + regression

Workflow class = **screen** (`wiki/quality/screen-definition-of-done.md` D2 gate + runtime functional
pass per `screen-qa-checklist.md`).

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (screen D2) | pass | Both `frontend-screen-reviewer` and `frontend-code-reviewer` APPROVE on record (0 Critical, 0 Major in M1 scope) for F1.1 and F1.2; honest loading/empty/error states; redesign tokens respected; generated types consumed |
| Runtime functional pass | pass | Live stack (`/` on web :4173 + api :8081): `§ MURMÚRIOS` renders 8 live audit events, hero pills + `§ SEU PULSO` render live `/documents/stats` counts (honest `0` for current seed, not fabricated) — on record in evidence |
| Regression vs M0 | all still pass | Single app-root index route intact (`dashboard/routes.tsx` `index: true`; other `index:true` are distinct nested layouts, M0's resolved design); no `OperationsPage`/`AuditPage`/`OperationsCenter` reintroduced (grep = 0); M1 reintroduced no mock in the dashboard boundary |
| Backend untouched | confirmed | `git diff HEAD -- '*.go'` empty; no Go change → no HS-6 scope drift |

Note: `MOCK_` still appears in `features/documents/components/distribution/*` — that is the
**Distribuição** screen, **M2 scope**, pre-existing and outside M1's `features/dashboard` boundary.
Not an M1 regression.

C4 PASS.

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| "no `MOCK_`/illustrative data on the Dashboard" (D2 crit. 1) | violated (`MOCK_STATS` + `MOCK_ACTIVITY` in DashboardPage) | met | `grep -rEn "MOCK_" src/features/dashboard` = 0 (validator re-run, exit 1); cards/feed render **live** values at runtime — mock deleted at the root, not hidden behind a flag or a permanent `'—'` placeholder (the `'—'`/`'…'` paths are honest loading/error states only) |

Root cause (mock data source) is removed, not symptom-patched: the constants are deleted and replaced
by live queries against generated-contract producers.

- Could it be built better? No material rebuild needed. Minor future polish already captured as
  bounded defers with triggers (stale `"5 dias úteis"` / `"5 dias úteis"`-style window-free label,
  `activityMuted` naming, `who`=raw `actor_id` until a display-name lookup exists, forced empty/error
  E2E harness, pre-existing `PendingRow` a11y/inline-style nits, legacy dead `AuditEventItem`). All
  are genuinely outside the mock→live wiring boundary and each carries a written trigger — none is a
  silent cap.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean; each feature's acceptance row maps to a named gate.*
- [ ] Fixture/mock passed off as real-provider proof — *clean; fixture unit tests and the live-API runtime proof are explicitly distinguished in both evidence files.*
- [ ] Consumer contract guessed rather than read from the consumer — *clean; both consumers use generated `components['schemas']` types, verified against the real bundle.*
- [ ] Split-brain (one fact, two sources of truth) — *clean; legacy `AuditEventItem` has 0 importers; M1 uses only the generated type.*
- [ ] Self-judged close / validator edited or fixed code — *clean; validator wrote only this verdict file.*
- [ ] Scope drift (work beyond the spec, no rationale) — *clean; only F1.1 + F1.2 dashboard files + docs changed; backend untouched.*
- [ ] Symptom-patch (bar moved by masking) — *clean; mock deleted at root, live data rendered.*

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass: **code-wise** (contract-clean, generated types, no split-brain, no dead code
  introduced, staff-engineer bar met) and **function/QA-wise** (the Dashboard renders 100% live data
  end-to-end; `MOCK_` = 0; reviewers APPROVE; runtime live-API proof on record). M0 regression check
  holds; backend untouched.
- Handed back to the main session to flip M1 status (validator PASS) and present the HS-1 operator
  gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — main session, only after this PASS + HS-1
