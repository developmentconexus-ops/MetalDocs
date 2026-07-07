# Milestone 2 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's evidence + the governing
> spec `docs/superpowers/specs/2026-07-06-lifecycle-ux-coherence-design.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-07-07 · **Diff judged:** `b7784284..HEAD` (99ad8cab, 2aed6c8b, 3e61f80d, fbffebce) · **Verdict:** see C7.

## Inputs loaded

- `milestone.md` (objective, R1/R5 constraints, runtime baseline w/ file:line, F2.1–F2.5 table, DoD) — present, approval-dated 2026-07-06.
- `f21-23-editor-submit/evidence.md` (T1), `f24-cockpit-approver-only/evidence.md` (T2), `f25-template-single-trigger/evidence.md` (T3) — all present, closed 2026-07-07, acceptance tables + live QA.
- Program `README.md` (status table, M1 PASSED precondition, deferred register) — present.
- Governing spec — present; R1 (§line 24) and R5 (§line 28) match milestone.md restatement.
- Aggregate diff — 21 FE files (+517/−436) + docs; ZERO backend files.

**Structural note (C1):** M2 features carry a grouped `evidence.md` per implementer task (T1/T2/T3),
not per-feature `spec.md`/`plan.md`. The up-front contract lives in `milestone.md` — per-feature
outcomes, binding constraints, runtime baseline verified in code with file:line, and DoD, approval-dated
2026-07-06 before code. Per the C1 rule ("skill absent + equivalent inline output present = PASS; binds
on artifacts, not on which skill produced them"), the milestone.md + task-grouped evidence together
satisfy the spec/plan/evidence intent: contract stated up front, acceptance mapped back row-for-row.
Not a C1 fail.

## C1 — Spec & plan conformance (per feature)

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F2.1 editor-submit-unify | ✅ sole client `approvalApi.submit` → `SubmitRequest = components['schemas']['SubmitDocumentRequest']`, no route_id/content_hash | ✅ `finalizeDocument`/`submitDocumentForReview` grep NONE | ✅ no new tracking screen | `documents.ts` diff (both fn+types deleted); `DocumentEditorPage.tsx:18,294` |
| F2.2 editor-reason-for-change | ✅ gate keys `doc.revision_number` (`:365`); body from `SubmitDocumentRequest` (`:392`); schema index.d.ts:3302 has revision_title/reason_for_change/reason_category enum | ✅ REV0 `submitForReview({})` no dialog; REV≥1 dialog required title+reason, optional category omitted when blank; 422 codes mapped (`:346-351`) | ✅ contract-first, no hand DTO | `DocumentEditorPage.tsx:363-398` |
| F2.3 editor-polish | ✅ n/a (UX) | ✅ `isSubmitting` guard (`:273,364,379`); "submeter" strings; success toast `:342` | ✅ | `DocumentEditorPage.tsx` |
| F2.4 cockpit-approver-only | ✅ `TransitionPolicy.actions` = {signoff,cancelInstance,publishOrSchedule} — no submit; 404 branch seeds no `"v0"` (`useDocumentApprovalArtifact.ts:129-135`) | ✅ policy/adapter/SignoffDetailPage/Extras all stripped; `openSubmit`/`showSubmit`/`onCloseSubmit` grep NONE | ✅ | `approvalWorkflow.ts:32-80`; SignoffDetailPage diff |
| F2.5 template-single-trigger | ✅ `buildTemplateApprovalActions` no draft case → `[]`; `submitForReview` retained in `templates/api` for editor | ✅ `runSubmit`/`canSubmit`/`submitForReview` grep NONE in `templates/lib`+route cockpit | ✅ | `templateApprovalActions.ts:29-84`; `TemplateApprovalRoute.tsx` diff |

Consumer contract (M1 backend `SubmitDocumentRequest`) was **read, not guessed** — the FE alias points
at the codegen schema and the REV≥1 gate mirrors the backend's "reason_for_change required for rev≥1"
rule stated in milestone.md's runtime baseline. All acceptance-vs-spec tables map back to milestone.md
outcomes row-for-row.

## C2 — Gates re-run, isolated (validator, clean state)

| Gate | Command re-run | Real output | Pass? |
|------|----------------|-------------|-------|
| R5 doc submit single path | `grep -rn "submitDocumentForReview\|finalizeDocument" src` | **No matches found** | ✅ |
| R5 template cockpit no submit | `grep runSubmit\|canSubmit\|submitForReview` in `templates/lib/templateApprovalActions.ts` + `pages/TemplateApprovalRoute.tsx` | **No matches** (editor `canActOnVersion.ts`/`TemplateEditorPage.tsx` retain `canSubmit` — the sole trigger; correct) | ✅ |
| R1 no submit in policy | grep `submit` in `approvalWorkflow.ts` | only `TRANSITION_POLICY` type-decl line; `actions` = signoff/cancelInstance/publishOrSchedule, no `submit` field | ✅ |
| R1 no cold `"v0"` seed | read `useDocumentApprovalArtifact.ts:129-135` | 404 branch sets `instance=null`; no `etagCache` seed | ✅ |
| Error-code parity | `vitest run errorMessages.coverage` | `2 passed` (bidirectional) | ✅ |
| FE full suite | `vitest run` (node_modules/.bin) | **122 files / 751 passed** (matches evidence exactly) | ✅ |
| Typecheck | `tsc --noEmit -p tsconfig.build.json` | `EXIT=0` | ✅ |

FE toolchain was **not** wedged — the local `vitest`/`tsc` ran clean; the known node_modules junction
drift did not manifest this run. Suite re-run from clean state reproduced 751/122 independently.

## C3 — Senior review of the aggregate milestone diff

- **Deletions clean, both sides updated.** `DocumentApprovalHandlers` dropped `openSubmit`; both call
  sites (SignoffDetailPage, adapter) and the Extras props (`showSubmit`/`onCloseSubmit`) removed. Grep
  confirms zero dangling references. `finalizeDocument` + its two type exports deleted together.
- **No split-brain.** One document-submit client (`approvalApi.submit`), one template-submit client
  (`templates/api.submitForReview`), one transition policy (`approvalWorkflow.ts` — comment forbids a
  second fork). Submit-request shape sourced solely from the codegen `SubmitDocumentRequest`.
- **No dead code.** Superseded route-picker (`route.manage`-gated, authors could never use — dead
  affordance) removed with its CSS. `templateApprovalActions` draft case removed; switch simplified.
- **No feature broke another.** M2 touches zero backend files; the M1 `/submit` producer + `/finalize`
  removal (openapi.yaml:3394 description confirms removal intact) are untouched.
- **Generated artifacts** (`lib/api-types/index.d.ts`, `error-codes.generated.json`) regenerated, not
  hand-edited — the parity test guards them. Staff-engineer bar met.

- Findings: none. Staff-engineer bar met? ✅

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (screen / FE) | pass | full vitest 751/122 green; tsc clean; contract-first types; feature-sliced; API via `lib/api`/`mutationClient` |
| Regression vs M1 (backend canonical submit) | all still pass | M2 diff has **zero** backend files; `/finalize` removal + `/submit` producer intact (openapi.yaml:3394); no Go touched |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| One submit impl per artifact kind (2 impls → 4 defects) | 2 doc submit paths + cockpit route-picker cold-submit + template cockpit dup trigger | 1 doc client, 1 template client, cockpits approver-only | Paths **deleted** (grep NONE), not masked — the second implementations no longer exist; policy has no submit field |
| Author submits from authoring context (R1) | submit affordances on cockpit surfaces | cockpits render zero submit; editor is sole trigger | `TRANSITION_POLICY` no submit; live draft-cockpit screenshot (zero submit); template draft case → `[]` |

Root cause fixed by deletion, not symptom-patch. Could it be built better? The REV≥1 dialog state is
colocated in `DocumentEditorPage.tsx` (large component); a future extraction into a dedicated
`SubmitReasonDialog` would improve cohesion — a non-blocking retrospective note, not a defect.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clear; each F2.x acceptance mapped in C1, live 201 proofs cited*
- [ ] Fixture/mock passed off as real-provider proof — *clear; REV0 & REV≥1 submit = real POST 201 on preview :4173; unit-only clearly labeled for F2.5*
- [ ] Consumer contract guessed rather than read from the consumer — *clear; `SubmitDocumentRequest` read from codegen schema*
- [ ] Split-brain (one fact, two sources of truth) — *clear; one client per kind, one policy*
- [ ] Self-judged close / validator edited or fixed code — *clear; validator wrote only this file*
- [ ] Scope drift (work beyond the spec, no rationale) — *clear; diff maps to F2.1–F2.5*
- [ ] Symptom-patch (bar moved by masking) — *clear; second paths deleted*

(All unchecked = clean.)

### Bounded defer judged
F2.5 draft-cockpit "no submit" was **unit-tested only** (no seeded template versions this session).
Judged **acceptable**: the change is a pure deletion of the `draft` case so `buildTemplateApprovalActions`
returns `[]` for draft/default — the "draft still shows submit" risk is structurally impossible, and the
unit suite asserts it. F2.2 REV≥1 was driven live (fresh-instance 201 proves the server accepted
`reason_for_change` at the rev≥1 gate) via an API-created fixture; documented defer to a seeded
rev-of-published fixture. Both defers have triggers/owners and do not gate close.

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass: **code-wise** (single client per kind, contract-first, clean deletions, no
  split-brain / dead code / dangling refs; tsc + 751 tests green) and **function-wise** (live REV0 +
  REV≥1 submit 201 from the editor; cockpit renders zero submit in draft — live for documents, unit +
  structural for templates).
- Handed back to the main session to flip M2 status and present the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — main session, only on this PASS
> - Commits remain local; do not push (program rule); HS-1 before M3.
