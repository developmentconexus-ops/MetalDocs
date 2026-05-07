# Frontend Code Review: novo-documento

**Implementation:** `frontend/apps/web/src/features/documents/` (wizard pages, components/wizard, queries, lib, state) + `frontend/apps/web/src/components/ui/{Stepper,SelectableCard}.{tsx,module.css}`
**Design source:** `frontend/apps/web/design-source/novo-documento/`
**Visual review (prerequisite):** ⚠️ partial — `artifacts/phase4-review.md` + `phase4-review-v2.md` exist but reported REQUEST CHANGES with one Critical (cache-key collision) that is **still present** in code at the time of this code review. Recommend re-running `frontend-screen-reviewer` after the cache-key fix lands.
**Verdict:** **REQUEST CHANGES**

---

## Critical

- [ ] `frontend/apps/web/src/features/documents/components/LibrarySidebar.tsx:44-48` and `frontend/apps/web/src/features/documents/queries/useAreasQuery.ts:7-13` — **TanStack Query cache-key collision with two different return shapes.** LibrarySidebar registers `QK.taxonomy.areas()` with `queryFn: listTaxonomyAreas` which returns `{ items: ProcessAreaItem[] }` (`features/taxonomy/api/catalog.ts:55`). `useAreasQuery` registers the **same key** with `fetchAreas` returning unwrapped `ProcessArea[]` (`features/taxonomy/api/taxonomy.ts:58`). Whichever query runs first poisons the cache for the other consumer. From Library → /documents-v2/new the wizard receives `{ items: [...] }` and `areasQuery.data ?? []` becomes a non-array; `areas.find(...)` in `NewDocumentWizardPage.tsx:77` throws. *Why:* `wiki/architecture/frontend-structure.md § Server state` — one query key, one canonical shape. Also called out in prior `phase4-review-v2.md` as Critical and not addressed. *Fix:* Migrate `LibrarySidebar` to `useAreasQuery` (preferred — `fetchAreas` returns the canonical `ProcessArea[]`), or give the envelope variant a distinct key (`QK.taxonomy.areasEnvelope()`) and stop using `QK.taxonomy.areas()` outside `useAreasQuery`.

- [ ] `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility.module.css:116` — `color: var(--text-on-brand, #fff)` references an **undeclared token**; grep across `frontend/` + `packages/` shows `--text-on-brand` is defined nowhere. The fallback hides the breakage at runtime, but per `wiki/architecture/frontend-structure.md § Styling` and `metaldocs-screen-implementation` SKILL § Phase 3b ("CSS Module uses ONLY tokens"), every `var(--…)` must resolve to a declared token. Fallback values are not a substitute. Already flagged Major in `phase4-review-v2.md`; promoted to Critical here because it is a contract violation that hides drift in CI. *Fix:* Declare `--text-on-brand: #ffffff` (or reuse `--surface`) in `src/styles/tokens.css` brand cluster, then remove the fallback.

- [ ] Phase 3b/4 hard-gate artifacts missing. Worksheet (`design-source/novo-documento/IMPLEMENTATION.md`) marks Phase 2 / 3a / 3b / 3c as `[x]`, but `design-source/novo-documento/artifacts/` contains only `phase4-review.md` + `phase4-review-v2.md`. There is no `phase0-audit.md`, `phase1-map.md`, `phase2-preflight.md`, `phase3a-structure.md`, `phase3b-style.md`, `token-coverage.txt`, `screenshots/{1440,1024,375}-{ref,impl}.png`, `parity-diff.md`, `leakage-probe.md`, or `phase4-behavior.md`. *Why:* `metaldocs-screen-implementation` SKILL § "Iron Law": **declared-done phase = non-empty artifact**. The wizard predates v1.2 of the skill, but the worksheet itself claims Phases 0–3c done. Per the calibration rule supplied with this review (declared-done phase + missing artifact = Critical), this is a Critical Iron-Law violation. *Fix:* Either backfill the artifacts (token-coverage.txt + 3-viewport screenshots are 30-min wins) or revise the worksheet to mark the phases as "delivered without artifact (pre-v1.2)" with explicit user sign-off.

- [ ] `frontend/apps/web/src/features/documents/components/wizard/steps/StepConfirm.module.css:198-208` and `StepConfirm.module.css:186-196` — page-level CSS Module **resets a global rule from `src/styles.css`** (`label span { text-transform: uppercase ... }` and `input { width: 100%; ... }`). The reset is correct local behavior, but the comment block at `:198-201` admits the cause is a global selector that targets ALL `label > span`. That global is the leak. Per `metaldocs-screen-implementation` SKILL § Phase 3b ("Global CSS leakage probe — known offenders → reset in page CSS Module **or scope the global narrower in `styles.css`, separate commit**") the right long-term fix is to scope the global, otherwise every future screen with a `label span` repeats the reset. *Why:* leakage propagates indefinitely without a leakage map (which is also missing — see prior Critical). *Fix:* File a follow-up commit narrowing `label span` and the bare `input { width: 100% }` selector in `src/styles.css`; add a `leakage-probe.md` row recording the offender. Until then keep the local reset, but log the debt.

---

## Major

### Architecture

- [ ] `frontend/apps/web/src/features/documents/queries/useTemplatesByProfileQuery.ts:6` — Inline query key `['templates', { docType: profileCode }] as const` bypasses the `QK` factory used elsewhere (e.g. `useProfilesQuery.ts:9`). *Why:* `wiki/architecture/frontend-structure.md § Server state` — query keys live in the `QK` registry so cache invalidation is centralized. *Fix:* Add `QK.templates.byProfile(profileCode)` and consume it here.

- [ ] `frontend/apps/web/src/features/documents/queries/useTemplatesByProfileQuery.ts:1-10` — No `staleTime` set; defaults to 0 (always refetched on mount). The other two query hooks in the same folder set `FIVE_MINUTES`. Templates change rarely; mismatch is unintentional. *Fix:* Set `staleTime: 5 * 60 * 1000` (extract a shared constant in `queries/`).

### Component design

- [ ] `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility.tsx` (290 LOC). At the cap. Two inline sub-components (`PeopleSubcontrols`, `ExternalSubcontrols`) plus the main step in one file. Per `metaldocs-frontend` SKILL § "Anti-patterns" / phase-9 review §"Local component split. If a page has 3+ inline sub-components..." this is borderline; it has 2 sub-components, but each is non-trivial, the file is at the LOC threshold, and they have their own lifecycle (rendered conditionally on visibility key). *Fix:* Extract `PeopleSubcontrols.tsx` and `ExternalSubcontrols.tsx` into the same `steps/` folder.

- [ ] `frontend/apps/web/src/features/documents/components/wizard/steps/StepConfirm.tsx:74-80` and `StepTemplate.tsx:95-101` — Inline-styled `<div style={{ width: ... }}>` used for paper-thumbnail line-bars, computed via `${55 + (idx * N) % M}%`. Same pattern duplicated in two files. *Why:* `metaldocs-frontend` § "Inline styles for theming" / DRY rule. This is layout, not theming, so it's borderline — but the same magic formula in two files is a copy-paste. *Fix:* Extract a tiny `<DocPaperPreview lineCount={N} />` primitive in `features/documents/components/wizard/` (or simply hoist the formula to a helper in `lib/paperPreview.ts`) and reuse from both Step 3 + Step 4.

- [ ] `frontend/apps/web/src/components/ui/SelectableCard.tsx:1-46` — Public API takes a single `onSelect: () => void`; the radio role implies a group, but the primitive has no concept of `name` / group context. Consumers stitch the radio group manually (Step 1 / Step 2 / Step 3 each map+pass). *Why:* `metaldocs-frontend` § "Discriminated unions for variants over many optional booleans" — currently boolean `disabled` + `selected` + optional `title` + optional `ariaLabel`. Acceptable for now (only 1 caller pattern), but if a 2nd wizard reuses the primitive, hoist a `<SelectableCardGroup name value onChange>` wrapper. *Fix:* Track in backlog as a follow-up; do not block this PR.

- [ ] `frontend/apps/web/src/features/documents/components/wizard/WizardShell.tsx:43-90` — `WizardFooter` is exported from the same file as `WizardShell` but is rendered by every step component (`StepProfile`, `StepAreaCodeVisibility`, `StepTemplate`, `StepConfirm`), not by the shell. The colocation is misleading: the shell renders the chrome, but the footer is per-step. *Why:* `metaldocs-frontend` § Naming/co-location — file boundaries should reflect responsibility. *Fix:* Split `WizardFooter.tsx` into its own file under `components/wizard/`. Low-effort; clarifies import sites.

### State / data flow

- [ ] `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx:88-93` — `useEffect` dependency array is `[profilesQuery.data, state.profileCode]` but the body also dispatches based on `state.profileCode`. Reading the captured value is fine, but the effect's purpose ("clear orphaned profile selection") is **derived state masquerading as effect**. After `profilesQuery.data` resolves, the truth of "is selected profile valid" is computable synchronously. *Why:* `metaldocs-frontend` § "useEffect only for side effects, never for derived state". *Fix:* Compute `selectedProfile` from `profiles.find(...)`; if profileCode is set but selectedProfile is null and `profilesQuery.isSuccess`, render an inline notice + Cancel button rather than silently dispatching `clearProfile` from an effect. Alternatively wrap the dispatch in `useEffect` with a clear comment that this is reconciliation, not derivation. Keep the dispatch in the effect if you must, but document the rationale.

- [ ] `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx:187-192` — Effect with dep `[state]` (the entire reducer state object) dispatches `goToStep` if `step > maxReachableStep(state)`. Two issues: (a) depending on `[state]` re-runs on **every** field change (typing in title, toggling visibility, etc.), unnecessary work; (b) the clamp logic already lives in the reducer's `selectProfile` / `clearProfile` cases — duplicating it as an effect is defensive coding without a documented threat model. *Why:* `metaldocs-frontend` § Effect hygiene + DRY. *Fix:* Move the clamp inside the affected reducer actions (e.g. in `clearProfile`, return `step: 1`; that's already done). Drop the effect, or scope its dep to `[state.profileCode, state.areaCode, state.title, state.templateVersionID]`.

- [ ] `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx:54-60` — URL `?step=N` sync effect depends on `[state.step, searchParams, setSearchParams]`. `setSearchParams` is referentially stable from `react-router`, but `searchParams` is a fresh `URLSearchParams` instance every render → effect re-runs every render → calls `setSearchParams(next, { replace: true })` whenever the same step is in the URL → guarded by the `next.get('step') !== String(state.step)` check, so no infinite loop, but the effect runs more than necessary. *Why:* dep hygiene — should depend on the value (`state.step`) not the carrier. *Fix:* Drop `searchParams` from deps and read it via `setSearchParams((prev) => { ... })` updater form, or split into "write step → URL" (deps `[state.step]`) and "read URL on mount" (one-shot via lazy reducer init, already done).

### Type safety

- [ ] `frontend/apps/web/src/features/documents/components/wizard/steps/StepProfile.tsx:14`, `StepTemplate.tsx:13`, `StepAreaCodeVisibility.tsx:18` — `error: unknown` props plus `error instanceof ApiError` discrimination at every callsite. The pattern is repeated 3× within the wizard. *Why:* `metaldocs-frontend` § "Discriminated unions for finite states" + DRY. *Fix:* Extract `function resolveQueryError(err: unknown): string` into `lib/api/` (or co-locate in `wiki/concepts/error-ux.md`'s helper) and use everywhere. The `wizard.reducer.ts:152-155` mutation already does the same dance — three call sites is the threshold.

- [ ] `frontend/apps/web/src/features/documents/components/wizard/steps/StepConfirm.tsx:53` — `summaryFields: ReadonlyArray<readonly [string, string]>` works but loses the relationship between label and field. *Fix:* Define `type SummaryField = { label: string; value: string }` and an array of those. Tuple-index access is not a meaningful win here.

### Error UX / a11y

- [ ] `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility.tsx:97-104` — Area-fetch error renders `<div role="alert" aria-live="assertive">` **without `className="card"`**. Other steps (StepProfile:51, StepTemplate:56, StepConfirm:135) wrap inline alerts in `card`. *Why:* `wiki/concepts/error-ux.md` — visual scaffolding consistency. Also flagged in `phase4-review-v2.md`. *Fix:* Add `className="card"`.

- [ ] `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility.tsx:203,253` — Sub-component prop names use `_onAddInvitee`, `_onRemoveInvitee`, `_onSetExternal` (underscore prefix to silence "unused" lint). The handlers are real props plumbed from the page; they're just unused while the feature is deferred. Underscore-prefixing in the **type signature** is misleading — it suggests the caller can rely on them being no-ops. *Fix:* Drop the underscore in the type; keep destructuring with rename if you must silence the lint locally, or just delete the props from the type until the feature is wired (the page still maintains them via reducer; the sub-components don't need them today).

- [ ] `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility.tsx:265,269,273` — `<label>` wrapping checkbox + text without `htmlFor`/`id`, mixed with `disabled aria-disabled readOnly` simultaneously on the inputs. `disabled` removes the input from the tab order and keyboard activation; `readOnly` does not apply to checkboxes (silently ignored). *Why:* `wiki/concepts/error-ux.md` a11y rule + MDN: `readonly` is not valid on `type=checkbox`. *Fix:* Drop `readOnly` from checkbox inputs. Keep `disabled aria-disabled="true"`. Optionally split the label so the text isn't part of the click target on disabled rows.

### Reusability / scalability

- [ ] `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility.module.css:9-10` — Cross-component CSS reach via `:global(label.kicker)` and `:global([data-selected="true"]) .visibilityIconTile`. The `data-selected` selector is OK because `SelectableCard` documents `data-selected` as a public hook (`SelectableCard.tsx:40`), but `:global(label.kicker)` reaches outside the module to style any descendant `label.kicker`. *Why:* CSS Modules § encapsulation; if `label.kicker` styling is universal it belongs in `styles.css`, not a step CSS Module. *Fix:* Either move the `display: block; margin-bottom: 6px` rule to a `.kicker--block` utility in `styles.css`, or wrap each label in a local class that is the actual styling target.

- [ ] No tests anywhere for the wizard. `__tests__/` contains nothing for `wizard.reducer.ts`, `useTemplatesByProfileQuery`, `NewDocumentWizardPage`, or `WizardShell`. *Why:* Phase 9 — `wizard.reducer.ts` has 8 cases with non-trivial guards (empty-code rejection, profile-clear cascade, invitee dedup, external patch merge), `maxReachableStep` + `clampStep` + `canAdvance` are pure functions with branchy logic — exactly the unit-test footprint described in the agent definition. *Fix:* Add `wizard.reducer.test.ts` covering: each action's state transition; empty-code guard; profile-change template invalidation; `maxReachableStep` step gating; `canAdvance` per step. ~80 LOC of vitest, no React Testing Library needed.

- [ ] `frontend/apps/web/design-source/novo-documento/IMPLEMENTATION.md:54-55` — Phase 0.2 user-confirmation checkboxes are still `[ ]`. Per the worksheet's own template these gate Phase 0 done. Documentation hygiene only, but the worksheet is supposed to be a record. *Fix:* Either tick them with a date or document the deviation.

### Maintainability

- [ ] `frontend/apps/web/src/features/documents/components/wizard/steps/StepAreaCodeVisibility.module.css:75-92` and `StepConfirm.module.css:130-167` — Both files have ≥ 30-line block comments quoting design coordinates by line number from `selected-wizard.jsx` / `selected-wizard-v2.jsx`. The line numbers will rot the moment those JSX mockups are touched (and they should be deleted post-merge per the design-workflow audit). *Why:* WHAT-comments anchored to volatile sources. *Fix:* Replace line citations with structural references ("Step 4 doc thumbnail dimensions"). Delete the JSX mockups from `design-source/` if they have no further role.

- [ ] `frontend/apps/web/src/features/documents/components/wizard/steps/StepProfile.module.css:50-56` and `StepConfirm.module.css:79-85` and `StepTemplate.module.css:79-83` — Repeated raw `font-size: 13px / 14px / 16px / 11px / 12px / 12.5px / 26px` with TODO comments referencing `wiki/backlog/novo-documento.md#typography-tokens`. The TODOs are correct but the type-scale debt lives in 5 different files. *Fix:* Add a typography token cluster (`--fs-1` ... `--fs-7`) in `tokens.css` in a follow-up PR; reference `wiki/backlog/novo-documento.md#typography-tokens` for the full list. Track in backlog (already done).

- [ ] `frontend/apps/web/src/features/documents/state/wizard.reducer.ts:73-78` — `selectProfile` case clears `templateID`+`templateVersionID` but does **not** reset `step` if user is on Step ≥ 3. The page-level effect at `NewDocumentWizardPage.tsx:187-192` clamps step *afterwards*. Defense-in-depth, but the reducer should be self-contained — the truth ("changing profile invalidates template selection AND any step that depends on template") belongs in the reducer. *Fix:* In `selectProfile`, after clearing template, set `step: Math.min(state.step, 2)`.

- [ ] `frontend/apps/web/src/features/documents/components/wizard/steps/StepConfirm.module.css:144-149` — `.nextStepsCallout :global(.kicker) { line-height: 1 }` is a known surgical fix from the visual-parity loop but lives only in this file. If another step uses `.kicker` inside a callout, the same fix is repeated. Tactical, not strategic. *Fix:* Long-term — define a `.kicker--callout` modifier in global `styles.css`. Short-term — leave as is and log to the leakage map (which is the missing artifact, see Critical).

---

## Minor

- `frontend/apps/web/src/features/documents/components/wizard/CodePreviewBanner.tsx:14-15` — `kicker` text changes whether `ready` or not, but the `???` placeholder is rendered either way. The two states differ only in the kicker; consider one source of truth (e.g. tooltip-only).
- `frontend/apps/web/src/features/documents/components/wizard/steps/StepConfirm.tsx:74,99` — `Array.from({ length: 11 })` and `Array.from({ length: 7 })` for paper-thumbnail line counts: hard-coded magic numbers. If extracted into the helper suggested above, the count becomes a prop.
- `frontend/apps/web/src/features/documents/components/wizard/steps/StepConfirm.tsx:159-167` — `formatDateTime` swallows errors (`catch { return date.toISOString() }`) — fine, but `Intl.DateTimeFormat('pt-BR', { dateStyle: 'short', timeStyle: 'short' })` is more idiomatic and a single-pass.
- `frontend/apps/web/src/features/documents/pages/NewDocumentWizardPage.tsx:27` — `export type NewDocumentWizardPageProps = Record<string, never>;` is unused (page takes no props). Drop.
- `frontend/apps/web/src/features/documents/components/wizard/WizardShell.tsx:35` — `String(currentStep)` round-trip via Stepper's stringly-typed `id` is awkward. Stepper could accept numeric ids; or this wrapper could be a typed stepper. Not blocking.
- `frontend/apps/web/src/features/documents/queries/useTemplatesByProfileQuery.ts:8` — `queryFn: () => listTemplates({ doc_type: profileCode ?? undefined })` — when `profileCode` is null, `enabled: false` already gates the call; the `?? undefined` is dead defense.
- `frontend/apps/web/src/features/documents/components/wizard/steps/StepConfirm.tsx:73` — `<div className={styles.thumbnailCode}>{codePreview} v1</div>` hard-codes `v1` — first version is correct for new docs but the literal sits next to a token-derived `codePreview`. Use a named local: `const versionLabel = 'v1'`.
- Phase 4 worksheet checkboxes (`IMPLEMENTATION.md:378-381`) still `[ ]`. Tick or annotate.

---

## What's good

- Mutation flow in `NewDocumentWizardPage.tsx:116-155` is textbook: `useMutation` with `onMutate`/`onSuccess`/`onError`, error funneled through `ApiError` → `resolveErrorMessage` → reducer + sonner toast. Matches `wiki/concepts/error-ux.md`.
- `wizard.reducer.ts` is pure, exhaustive over a discriminated `WizardAction`, with an empty-code guard on `selectProfile` and a separate `clearProfile` for the unset path. Cleanest piece of the implementation.
- `Stepper` (components/ui) is a real generic primitive — keyboard navigation, `aria-current`, `aria-label` on the `<ol>`, completed/active/upcoming states, optional `onStepClick`. Promotion to `components/ui/` was the correct call per Phase 1.2.
- Selected state for visibility icon-tile uses `SelectableCard`'s public `data-selected` data attribute (`StepAreaCodeVisibility.module.css:114-117`) rather than passing a render prop or duplicating selection logic — clean primitive composition.

---

## Worksheet + audit cross-check (Phase 11)

- Cut items absent: ✅ — "soon" disabled badge cut, no model-field UI surfaces.
- Keep items present: ✅ — every Keep row in §0.1 has a code path (or a documented mock fallback for deferred ones).
- §1.1/§1.2 placement decisions respected: ✅ — `Stepper` + `SelectableCard` in `components/ui/`; `WizardShell` + steps + `CodePreviewBanner` in `features/documents/components/wizard/`; `visibilityMeta.ts` in `features/documents/lib/`.
- §1.4 status-meta SSOT single-file: ✅ — `visibilityMeta.ts` is the only place declaring visibility labels/icons.
- §1.5 state design honored: ⚠️ — server state via TanStack Query (✓), local via `useReducer` (✓), persisted not used (✓). But `useTemplatesByProfileQuery` lacks the documented 5-min `staleTime` (Major above).
- §1.6 backend contract — endpoints real or mocked-with-trail: ✅ — visibility / share / external all rendered with TODO + `wiki/backlog/novo-documento.md` rows. Slot rollback TODO present.
- Phase 2 primitive audit fixes verifiable in git: ❌ — no `phase2-preflight.md` artifact; cannot verify which primitive fixes (if any) the subagent committed.
- Phase 2 Global CSS Leakage Map applied in page CSS: ⚠️ — leakage IS reset (StepConfirm.module.css:186-208) but no `leakage-probe.md` records it (Critical above).
- Open Questions Log fully resolved: ✅ — 6 questions logged, 6 answered with dates. Worksheet hygiene for Phase 0/1 is exemplary.
- All declared-done phases have non-empty artifacts: ❌ — only `phase4-review.md` + `phase4-review-v2.md` present. Iron Law violation (Critical above).

## Backlog hygiene

- TODOs in code matching `wiki/backlog/novo-documento.md`: ✅ — backlog file exists with sections matching `TODO(novo-documento:visibility|sharing|template-versions|blank-template|slot-rollback|profile-counts|typography-tokens)` and key files line anchors.
- Mock data with TODO trail: ✅ — every deferred control (people invite, external share, em-branco, profile counts) has the comment block + backlog row.

## Stats

- Files in scope (wizard + new primitives + queries + lib + state + page): 18
- Total LOC: ~2,157 (TSX+CSS+ts).
- Largest TSX: `StepAreaCodeVisibility.tsx` 290 LOC (under the 400-LOC cap; near the 3-inline-subcomponent threshold).
- Largest CSS: `StepConfirm.module.css` 208 LOC.
- Page TSX: `NewDocumentWizardPage.tsx` 280 LOC — within budget.
- Reducer: `wizard.reducer.ts` 148 LOC — clean.
- Tests added: 0.
- New `components/ui/` primitives: 2 (Stepper, SelectableCard). Both reusable.
- New `features/documents/` files: 13 (page, 4 step TSX + 4 step CSS, WizardShell TSX+CSS, CodePreviewBanner TSX+CSS, visibilityMeta, wizard.reducer, 3 query hooks).
