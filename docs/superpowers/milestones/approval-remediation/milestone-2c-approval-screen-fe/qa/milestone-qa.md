# Milestone 2c — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (up-front spec) + each feature's `spec.md` + governing spec §8.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-07-07  ·  **Verdict:** see C7 — **PASS**.
> The validator judges and writes this file; the **main session flips status only on a PASS**. The
> validator never edits code, fixes findings, or flips status.

## Inputs loaded (all present, all readable)

- Milestone spec `../milestone.md` (features F0–F8, validation definition §1–6, hard-stops HS-1..HS-6). ✅
- Every feature `spec.md`/`plan.md`/`evidence.md` for F0–F7; F8 `evidence.md` + `qa/live-qa-log.md`
  (F8 is the close feature — no spec/plan by design, its contract is the `milestone.md` F8 row). ✅
- Program `README.md` (status table + HS-2 F0 resolution + bounded-defer registry). ✅
- Governing spec `docs/superpowers/specs/2026-07-07-approval-remediation-design.md` §8 (C1–C5). ✅
- Aggregate diff: 127 files, +7335/−6309 since M2b close (`git diff --stat 6628f7c1`), F1–F7 committed
  (`148fab1c..44dbe1dd`, F0 at `d5b4bf96`+`66187625`); F8 backend bug-fixes uncommitted in working
  tree (expected — close feature; NOT pushed per F8 contract, main session commits post-PASS). ✅

No missing/unreadable input. Did not fail-fast. Both dimensions (code-wise + function-wise) judged.

## C1 — Spec & plan conformance (per feature)

Every F0–F7 folder has `spec.md` (with a filled `Approved before code: 2026-07-07` line + populated
Interview record), an execution-shaped `plan.md`, and an `evidence.md` whose acceptance table maps
row-for-row to the spec Validation Gate. Consumer contracts read from the consumer site, not guessed.
F8 = close feature (contract = `milestone.md` F8 row); evidence + live-QA log present.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F0 backend-contract-closure | ✅ producer (openapi + Go) regenerated to the enum/`stage_kind` the FE consumes; api-types has both (generated, ADR 0035) | ✅ `TestInstanceStatusWireEnumComplete` + `TestStageRequestStageKindValidation` PASS (re-run below); regen verified | ✅ markup gate removed per HS-2 path A, not built | HS-2 recorded in README; grep `ScanForUnresolvedMarkup`→0 |
| F1 editor-ui track-changes | ✅ neutral `TrackedChange` surface consumed by F4/F6 with real names `acceptChange`/`rejectChange`; ACL wall intact | ✅ `trackChanges.test.tsx` + typecheck green (docx-v2) | ✅ body-only, no server-state | editor-ui suite Done |
| F2 approval-api-layer | ✅ `reviewVerdict`/inbox filters/delegations consume `components['schemas']` only (no hand DTO); consumed by F4/F5 | ✅ `approvalApi.test.ts` green; F0-regen fallout (5 tsc errors, task#11) resolved — web tsc clean now | ✅ no dashboard/admin UI, no openapi edit | web tsc exit 0 |
| F3 cockpit-document-shell | ✅ `DocumentShell` mounts same canvas for author+cockpit; W2 vector removed at root | ✅ `ApprovalCockpitPage.test.tsx`/`DocumentShell.test.tsx` green; session/autosave call-count 0 | ✅ shell = canvas region only; `ReviewDocumentCanvas` deleted | grep session/autosave→only test-mocks |
| F4 approval-sidebar-ia | ✅ single `ApprovalSidebar` into cockpit; shared `ArtifactApprovalScreen` neutralized not edited (empty diff) | ✅ single-timeline `toHaveLength(1)`, integrity behind disclosure, decision footer mode-aware — all green | ✅ shared component untouched; §6 legal jurisdiction flagged to HS-1 (not silently swapped) | grep `DocumentApprovalExtras`→0 |
| F5 worklist-single-destination | ✅ item open → `/approvals/:documentId` cockpit, not author editor (closes a W2-class hole) | ✅ `InboxFilters.test.tsx` + extended `InboxPage.test.tsx` green | ✅ doc-type filter deferred (D1, contract gap, flagged); reactive oversee (D2) | grep `/edit` in InboxPage→0 |
| F6 author-request-changes-panel | ✅ panel gated on instance `status==='changes_requested'`; F1 ref API reused; clean-buffer order asserted | ✅ `RequestedChangesPanel.test.tsx` + graceful-404 test green | ✅ page-owned panel; shell untouched; D1 no backend gate flagged | shell diff empty |
| F7 visual-a11y-polish | ✅ token purity/focus/reduced-motion/contrast on in-scope files; a keyboard/AT+reduced-motion consumer | ✅ grep proofs 0; 2 behavior tests RED→GREEN (stash-verified); contrast table with real ratios | ✅ shared files untouched; StateBadge dead-code documented (D1); token-level contrast deferred (D4/D5/D6) | 13-file grep→0 |
| F8 close | ✅ full suites + real-stack live QA + validator dispatch; 2 backend defects = producer→OpenAPI-contract fixes | ✅ all gates green (below); live walkthrough recorded | ✅ NOT pushed; no product scope added beyond the 2 HS-3 repairs | live-qa-log.md |

C1 **PASS**. Every feature's approval line is filled; every interview record populated; no missing
artifact; all disclosed deviations carry written rationale in the owning evidence.md.

## C2 — Gates re-run, isolated (validator's own clean-state runs, not trusted from transcript)

| Feature / gate | Command re-run | Real output | Pass? |
|----------------|----------------|-------------|-------|
| Static | `go build ./...` | `BUILD_EXIT=0` (no output) | ✅ |
| F0 | `go test -count=1 .../http/contracts/ -run 'TestInstanceStatusWireEnumComplete\|TestStageRequestStageKindValidation' -v` | both `--- PASS`; 4 stage_kind subcases PASS; `ok 1.218s` | ✅ |
| F0 regen | `grep -c "stage_kind\|changes_requested" api-types/index.d.ts` | 7 (generated DTOs; `status` enum incl. `changes_requested`, `stage_kind?: "review"\|"approval"`) | ✅ |
| Approval module (F0 + F8 fixes) | `go test ./internal/modules/documents/approval/...` | all `ok`, `APPROVAL_TEST_EXIT=0` | ✅ |
| Full Go suite | `go test ./...` | `FULL_GO_TEST_EXIT=0`, zero non-ok lines (no FAIL) | ✅ |
| docx-v2 types (F1) | `npm run typecheck:docx-v2` | 8 projects Done, `DOCX_TYPECHECK_EXIT=0` | ✅ |
| Web types (F2–F7) | `cd frontend/apps/web && npx tsc --noEmit -p tsconfig.build.json` | `WEB_TSC_EXIT=0`, clean (F2 regen-fallout errors resolved) | ✅ |
| FE tests (M2c surfaces) | `pnpm exec vitest run src/features/approval src/features/documents src/lib/inbox src/lib/format` | **54 files / 374 tests passed**, `FE_SCOPED_EXIT=0` — incl. `ApprovalSidebar` single-timeline, `DecisionFooter`, `formatDueRelative`, `InboxFilters`, `RequestedChangesPanel`, `ApprovalCockpitPage` (session/autosave call-count 0) | ✅ |

DB integration tests (`//go:build integration`, incl. `TestInsertStageInstances_...NilDueInDays...`
and `TestLoadInstanceByDocumentForView_SeesChangesRequested`) SKIP without DATABASE_URL — expected and
honest; their RED→GREEN was proven live on :8081 (F8 log), not treated as SKIP-is-pass. C2 **PASS**.

## C3 — Senior review of the aggregate milestone diff

Reviewed the 127-file / +7335/−6309 diff as one unit.

- **No split-brain:** contract is single-sourced — openapi → oapi-codegen → FE `api-types` (F0);
  FE consumes generated `components['schemas']` only (ADR 0035), verified zero hand-written body DTOs.
  The canonical hash chain is untouched (no second source of integrity truth introduced).
- **No dead code left:** `ReviewDocumentCanvas`(+test), `DocumentApprovalExtras`(+css), the W2 writable
  session on the approval surface, the duplicate "Fluxo de aprovação" band, and the "dados
  desatualizados" polling banner are **deleted** (grep-zero in `src/features/approval`, non-test). The
  misdirected `ScanForUnresolvedMarkup` deleted (grep-zero in `internal`). `TemplateReviewCanvas` is a
  separate templates-feature component, correctly out of scope.
- **No feature broke another:** shared `ArtifactApprovalScreen` neutralized via null model fields, not
  edited (empty diff) — template approval screen untouched; `DocumentShell` shared with cockpit is
  byte-identical to F6/F7 non-owners; `DocumentEditorPage` author path stays visually inert (22/22).
- **Two backend defects (BUG#1 42P08 nil-SLA submit; BUG#2 `changes_requested` view-404):** judged
  **legitimately in-boundary HS-3 prerequisite repairs**, not unplanned scope. Both were exposed by F8
  live QA as compile≠work / fixture≠real gaps that hard-blocked the mandated walkthrough (BUG#2 made
  the C5 author panel — a core M2c deliverable — non-functional live despite green fixture units). Both
  are root-caused, minimal, contract-conforming (producer made to match the OpenAPI enum the FE already
  consumes), keep publish/mutation on the narrow method (no semantic drift), keep identical ADR-0022
  authz gates (no capability weakening), and carry non-tautological integration tests
  (`LoadInstanceByDocumentForView` byte-identical to the active read except the status set; BUG#2 test
  also asserts the narrow method still returns `ErrNoActiveInstance`). Independently code-reviewed
  APPROVE, 0 findings. This is exactly what HS-3 sanctions.

Staff-engineer bar met? ✅ Findings: none blocking.

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Frontend screen QA (`qa-operating-system.md` + `frontend-structure.md`) | pass | generated-DTO discipline, react-query invalidation replaces polling, per-surface loading/empty/error states, PT-BR/wine tokens, visible focus + reduced-motion (F7) |
| F0 backend slice (`backend-api-structure.md` + `api-contract.md`) | pass | contract-first: openapi+oapi-codegen regen before handlers; enum/`stage_kind` on the wire |
| **Regression vs M2b** (`go test ./internal/modules/documents/approval/...`) | **all pass** | F0 + both F8 fixes edit this module; module suite `ok` exit 0 from clean state |
| Regression vs whole backend (`go test ./...`) | all pass | `FULL_GO_TEST_EXIT=0`, no FAIL |
| Root-cause re-measure (§4) | pass | see C5 |

C4 **PASS** — no prior-milestone regression.

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence (not symptom-patch) |
|-------------|--------|-------|-----------------------------------------------|
| W2 writable-session vector on approval surface | reviewer opened a writable `useDocumentSession`+autosave during `under_review` | **gone at the root** | `grep -r "useDocumentSession\|useDocumentAutosave" src/features/approval` → only the two `vi.mock` declarations in `ApprovalCockpitPage.test.tsx` that assert the hooks are **never called** (call-count 0). Hooks removed from the feature, not hidden behind a readonly flag. |
| Duplicate approval timeline | two "Fluxo de aprovação" bands rendered | **single timeline** | `ApprovalSidebar.test.tsx:154` `expect(timelines).toHaveLength(1)` PASS — duplicate band **deleted** (shared component neutralized via null model fields, `DocumentApprovalExtras` removed), not CSS-hidden. |
| Contract truth (ADR 0035) | hand-written body consumers risk | generated-only | api-types regenerated; zero hand-written body DTOs across F2/F5/F6 (grep + reviewer attribution). |

- **Could it be built better?** The one open structural note is the **server-authoritative
  suggestion-resolution freeze gate** — today tracked-change cleanliness before re-submit is
  client-authoritative (eigenpal) + caught by the frozen-content-hash chain; the backend gate is
  comments-only. This is correctly registered as a bounded defer (HS-2, program README) with an
  explicit trigger, not a masked gap. It does not make the current construction unsound (the hash
  chain fails closed). No FAIL on this basis. Retrospective input for a post-program backend feature.

C5 **PASS** — bars moved at the root, not symptom-patched.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — **clean** (C1 maps every feature to its named test/grep).
- [ ] Fixture/mock passed off as real-provider proof — **clean** (every evidence.md labels unit=fixture/mock and defers real end-to-end to the F8 live-stack log; F8 log honestly notes DOM-introspection where the screenshot tool times out on the docx iframe).
- [ ] Consumer contract guessed rather than read from the consumer — **clean** (each spec cites the consumer site file:line; F6 corrected the stale `…ById` naming to the real ref API by reading `types.ts`).
- [ ] Split-brain — **clean** (single contract source openapi→codegen→api-types; hash chain single-sourced; publish/mutation kept on the narrow read while GET uses the dedicated view read — two explicit status-scoped queries, not one polymorphic fallback).
- [ ] Self-judged close / validator edited code — **clean** (validator wrote only this file; F8 backend fixes were built + independently reviewed by separate subagents, not by this validator).
- [ ] Scope drift — **clean** (rabbit-holes — delegation admin UI, oversee dashboard, W12 DAG, author re-theming — all respected; the only beyond-plan work is the two HS-3 backend repairs, each recorded with rationale as HS-1 deviations; F5 doc-type filter deferred as a flagged contract gap, not silently dropped).
- [ ] Symptom-patch — **clean** (W2 removed at root; duplicate timeline deleted).

All unchecked = clean.

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass: **code-wise** (contract-clean, generated DTOs, no split-brain, dead code
  deleted, staff-engineer bar met) and **function-wise** (full lifecycle proven live RED→GREEN on the
  real stack :8081 — route review→approval, submit, suggesting, request_changes, author panel, clean
  buffer, freeze + frozen_content_hash pin, sign-with-meaning, publish, visibility 404, oversee
  200/403, cancel-with-reason). The two backend defects live QA exposed are in-boundary HS-3 repairs,
  root-caused and honestly tested — they strengthen the PASS (compile≠work was caught before close),
  not weaken it.
- Handed back to the main session to flip status and present the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending — surface the accumulated deviations for ratification:
>   F4 §6 (21 CFR 11.50 vs MP 2.200-2/2001 signature-legal jurisdiction — conservative default shipped,
>   operator chooses); F5 D1 (doc-type filter deferred, contract gap) / D2 (reactive oversee vs probe);
>   F6 D1 (no server-authoritative tracked-changes freeze gate — client-authoritative + hash chain) /
>   D2 (ref naming); F7 D1–D6 (StateBadge dead-code, dialog scrims, token-level contrast D4/D5/D6);
>   and the two HS-3 backend repairs (BUG#1/#2). Program terminal acceptance sits behind this gate.
> - Commit the uncommitted F8 backend bug-fixes (working tree) after PASS. NOT pushed.
> - Status flipped in program `README.md`: only on PASS + operator HS-1.
