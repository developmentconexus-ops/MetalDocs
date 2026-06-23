# Milestone 4 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-23  ·  **Verdict:** see C7 — **PASS**.
> The validator judged and wrote this file; it edited no source and flipped no status. The **main
> session flips status only on this PASS**.

**Milestone:** M4 — Documento Publicado completion + Documento Obsoleto (frontend-only, 2 features).
**Aggregate diff judged:** `ae7c8a1a..4f8e5c3b` (M4 commits `d8f03953` F4.1, `bf730e43` F4.1 close-out, `4f8e5c3b` F4.2). 15 files, +1158/−66; **0 Go source files changed** (`git diff --name-only … -- '*.go'` empty). Source changes confined to the declared component surface: `DocumentPublishedPage.tsx` / `.module.css` / `.test.tsx`, `lib/documentDetailMeta.ts(.test.ts)`, and the backlog doc.

**Inputs loaded:** `milestone.md`; F4.1 + F4.2 `spec.md`/`plan.md`/`evidence.md` (all present); program `README.md`; governing `mission.md` reference; aggregate diff. No input missing — judged with full sight.

## C1 — Spec & plan conformance (per feature)

Both features: `spec.md` `Approved before code:` line filled **2026-06-23 / leandrotca**; interview record populated (F4.1 = 3 Q&A rows incl. the runtime-truth correction that page-count/file-size already ship; F4.2 = 4 Q&A rows resolving parity depth + the capability choice); `plan.md` is execution-shaped (TDD task list, files touched, ordering); `evidence.md` acceptance table maps **row-for-row** to the `spec.md` Validation Gate; non-goals respected.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F4.1 | ✅ — consumes `DistributionSummaryResponse.total_targets` (denominator only, numerator parked ADR-0042), `exportPDF` POST client, and `DocumentResponse.current_revision_page_count` / `current_revision_file_size_bytes` (nullable → honest em-dash) — exactly the shipped, frozen shapes; no hand mapper | ✅ — 8/8 gate rows met; coverage count renders (no fabricated `%`/progressbar), PDF button enabled + wired with 429 + non-429 error handling, páginas/tamanho from real fields, every surviving "em breve" eliminated (grep = NONE) | ✅ — no Go change, no numerator, no `ExportMenu` swap, no restyle, no obsolete work | `f4.1-publicado-stubs/evidence.md`; grep "em breve"=NONE |
| F4.2 | ✅ — obsolete variant driven by **real** `DocumentResponse.status === 'obsolete'` (not a prop hack/fork); capability gate via existing `useHasCapability('document.obsolete')` reading `user.capabilities`; `document.obsolete` is a canonical backend capability (`internal/modules/iam/domain/model.go:97 CapDocumentObsolete`) | ✅ — 9/9 gate rows met; OBSOLETO watermark + dim + hidden vigente pill + capability-gated Visualizar + disabled Baixar/Copiar, all with named tests incl. negative controls | ✅ — no new capability, no second page file (reuse asserted), no restyle, no F4.1-wiring change, **no backend-enforcement claim** (UX-hint-only, comment cites `wiki/concepts/authz-tiers.md`) | `f4.2-obsoleto-variant/evidence.md` |

Minor doc imprecision (not a fail): F4.2 `spec.md` cites the capability's home as `catalog.go`; it actually lives in `iam/domain/model.go:97`. The capability genuinely exists and the consumer contract is honored — the reference is slightly off, the contract was not guessed.

## C2 — Gates re-run, isolated

Re-run by the validator from clean state (`frontend/apps/web`), not trusted from the transcripts.

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| F4.1 + F4.2 | `npx vitest run DocumentPublishedPage documentDetailMeta` | `Test Files 2 passed (2) · Tests 43 passed (43)` (35 page + 8 formatter), 3.03s. The lone `stderr` line (`[api] unmapped error code pdf.render_failed`) is the expected log from the 502 PDF-failure test case, not a failure. | ✅ |
| F4.1 + F4.2 | `npx tsc --noEmit` | **exit 0** | ✅ |
| F4.2 reuse | `ls src/features/documents/pages` + grep | One `DocumentPublishedPage.tsx`; **no obsolete page file**; obsolete branch lives in-component (`isObsolete`/`canViewObsolete`/`rootObsolete`) | ✅ |

## C3 — Senior review of the aggregate milestone diff

Whole-M4 diff reviewed as one unit.

- **No split-brain** — coverage denominator has one source (`useDistributionSummaryQuery.total_targets`); obsolete state has one source (`DocumentResponse.status`); capability has one source (`user.capabilities` via `useHasCapability`). No fact stored twice.
- **No dead code** — superseded `.kpiValuePlaceholder` and `.coverageCardBar` CSS classes removed with the markup that used them; the fabricated `—%` value + `role="progressbar"` bar deleted.
- **No feature breaking another** — F4.2 adds an obsolete branch on top of F4.1's wirings without altering them; `disabled={isObsolete || …}` extends, not replaces.
- **No false enforcement** — the capability gate carries an explicit "UX hint; backend document.view is the real boundary" comment; Visualizar's `onClick={handleView}` is retained (disabled is presentational), consistent with the `canActOnVersion` precedent.
- New helpers `formatFileSize`/`formatPageCount` are pure, unit-tested (8 cases), em-dash on null/NaN/negative.
- Findings: none blocking. `.pdfAlert` uses `var(--danger, #b42318)` — a tokened value with a hardcoded fallback, consistent with the page's pre-existing (deferred) color-literal pattern.
- Staff-engineer bar met? ✅

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (screen DoD / screen-qa) | pass | Real-status-driven render, generated types consumed directly, tsc clean, a11y `role="alert"` on PDF + obsolete states, honest empty/absent states, both reviewer agents APPROVE on record. |
| Regression vs prior milestones | all still pass | **0 Go change** (M2/M3 backend untouched). `npx vitest run DocumentDistributionPage NotificationBell` ⇒ `2 passed / 12 tests` — M2 distribution + M3 notifications intact. M4 is component-local to `DocumentPublishedPage`. |
| Known pre-existing failure (not M4) | not attributed to M4 | `DocumentEditorPage` suite fails with duplicate-React `useState` from `node_modules/.pnpm` (memory `fe-node-modules-junction-drift`) — failure is entirely in node_modules; M4 changed no editor file and no node_modules. Pre-existing infrastructure, not an M4 regression. |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| mission §8 "no silent stub / no fabricated number for in-scope items" (Publicado + Obsoleto) | 7 "em breve" placeholders, a fabricated `—%` + fake progressbar | `grep -n "em breve" DocumentPublishedPage.tsx` ⇒ **NONE**; fabricated `%`/bar removed | Root cause was "endpoint/field existed but unwired" — fixed by **wiring** (coverage denominator, PDF, páginas, tamanho to real producers), not by a nicer placeholder. The 4 genuinely-unbacked fields are explicit defer-with-trigger rows in `wiki/backlog/documento-publicado.md` rendering honest `—`/`não disponível`. |
| Obsolete = status-driven branch of one component (R2 anti-fork) | half-built obsolete branch | full design-source parity (dim + hidden pill + gated actions + watermark), one component | No new page file (C2); F4.2 delta +113 LOC, 0 new files/exports. |

- Could it be built better? The 824-LOC god-component remains (>400 LOC Major gate) — an **honest, pre-existing F4.1 inheritance** with a written trigger ("next substantive feature touching this page → extract `ObsoleteBanner`/`HeroActions`/…"), not a new M4 defect; F4.2 worsens it by no new file. Recorded as next-feature input, does not FAIL this milestone (the current construction is sound).

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean: per-feature gate rows each map to a named test/grep.*
- [ ] Fixture/mock passed off as real-provider proof — *clean: vitest cases are explicitly labeled fixture-level (mocked query/capability); grep + tsc + reuse + 0-Go-change are the real checks; live producers proven by M2 + existing export flow.*
- [ ] Consumer contract guessed rather than read from the consumer — *clean: contracts read from generated types + the real `status`/`capabilities` fields; capability verified to exist in backend (`model.go:97`).*
- [ ] Split-brain — *clean (C3).*
- [ ] Self-judged close / validator edited or fixed code — *clean: validator wrote only this file; status not flipped.*
- [ ] Scope drift — *clean: no deferred-field backend, no numerator, no forked page; all deltas within F4.1+F4.2 spec.*
- [ ] Symptom-patch — *clean: bar moved by wiring real data + honest defers, root cause fixed.*

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass independently: **code-wise** (contract-clean, no split-brain, no dead code, senior-level, no false enforcement claim) and **function-wise** (the Publicado screen renders real coverage denominator / PDF / páginas / tamanho and the Obsoleto variant from real `status`, all proven by re-run gates + grep + reuse assertion). All in-scope "em breve" stubs eliminated; remaining absent states are backlog-backed defers with triggers.
- **PASS** — handed back to the main session to flip M4 status in `README.md` and present the HS-1 operator gate. No M5 / no merge without explicit approval.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: pending — only after this PASS + HS-1
