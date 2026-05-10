# Template Wizard — Architecture Hardening Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to execute task-by-task. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Resolve 3 root-cause architectural smells in the 5-step template wizard (M1+N4+N2 = nav SSOT; M2+M3+N5 = ARIA radiogroup primitive; M4 = dead props). No mock/API work. Out of scope: M5/M6 tests, N3/N6 copy fixes, N7 mock comments.

**Architecture:**
- **Smell 1** — Reducer owns step navigability via `maxReachableStep` derived selector. `GO_TO_STEP` clamps target. Page collapses 8 hand-rolled handlers into one `goToStep` callback.
- **Smell 2** — Headless `useRovingRadioGroup` hook owns ARIA `radiogroup` semantics + roving `tabIndex` + arrow-key nav. `SelectableCard` gains `forwardRef` + optional `tabIndex` prop. Both `StepScope` and `StepStructure` consume the hook.
- **Smell 3** — Drop unused `description` + `scopeType` props from `StepConfirmation` (YAGNI).

**Tech Stack:** React 18, TypeScript, CSS Modules, `useReducer`. No new dependencies.

**Tooling rules:**
- **Codex** (`codex exec --model gpt-5.3-codex` via `codex:codex-rescue` subagent) for all complex code refactors + final code review. Codex prompts in caveman style.
- **Sonnet/Haiku** (main session) for trivial edits (file creation scaffolds, prop strip, inline-style → class), all `git` operations, all artifact writing.
- **Opus** (`frontend-code-reviewer` subagent) for Phase 6 review.
- **Codex never** writes commits or runs `git`. Codex never writes tests (none planned here anyway).
- **`/simplify`** mental pass after each Codex run before commit.
- **Codex parallelism** — Phase 1 boundary tightened to **reducer + page only** (step prop signatures unchanged → no step files touched). Phase 2 owns hook + SelectableCard + StepScope + StepStructure. Truly file-disjoint → parallel.
- **`wiki-curator`** subagent at Phase 8.

---

## File map

**Modify:**
- `frontend/apps/web/src/features/templates/state/templateWizard.reducer.ts` — add `selectMaxReachableStep` selector + clamp logic in `GO_TO_STEP` reducer case
- `frontend/apps/web/src/features/templates/pages/TemplateWizardPage.tsx` — collapse 8 `handle{Advance,Back}FromStepN` → one `goToStep`; remove 5 inline gate `const`s; pass `goToStep` down to step props
- `frontend/apps/web/src/features/templates/components/wizard/steps/StepScope.tsx` — wrap scope + profile grids in `useRovingRadioGroup`; kill inline `style={{ marginBottom: 8 }}` (N1)
- `frontend/apps/web/src/features/templates/components/wizard/steps/StepScope.module.css` — add `.profileSectionKicker { margin-bottom: var(--sp-2); }`
- `frontend/apps/web/src/features/templates/components/wizard/steps/StepStructure.tsx` — adopt `useRovingRadioGroup` for the 2 starting-point cards
- `frontend/apps/web/src/features/templates/components/wizard/steps/StepIdentity.tsx` — adapt prop signature if `onAdvance`/`onBack` collapse
- `frontend/apps/web/src/features/templates/components/wizard/steps/StepPermissions.tsx` — same as above
- `frontend/apps/web/src/features/templates/components/wizard/steps/StepConfirmation.tsx` — drop `description` + `scopeType` props + their `void` lines
- `frontend/apps/web/src/components/ui/SelectableCard.tsx` — convert to `forwardRef`, accept `tabIndex` prop, drop default `role="radio"`-when-not-in-group? (decision: keep `role="radio"` default; group enforces `radiogroup` parent — same contract)

**Create:**
- `frontend/apps/web/src/components/ui/useRovingRadioGroup.ts` — headless hook returning `{ groupProps, getItemProps(index, ref) }`

**No new tests this plan.** Reducer test row (M5) and wizard E2E (M6) deferred per user signal.

---

## Pre-flight: Validate plan with Codex

- [ ] **Step 0.1: Dispatch Codex to validate this plan**

Dispatch `codex:codex-rescue` subagent with caveman prompt:

```
Read plan: docs/superpowers/plans/2026-05-10-template-wizard-architecture.md
Read current code referenced in File map.
Validate:
- Smell 1 reducer change feasible? GO_TO_STEP clamp breaks anything?
- useRovingRadioGroup hook + forwardRef pattern fits existing SelectableCard consumers? Other consumers exist? Grep src for SelectableCard usage.
- Dead-prop strip safe? Anyone read description/scopeType in StepConfirmation today?
- Order: Phases 1+2 truly independent? Any file overlap?
- Missing risk: anything plan misses?
Report findings under 300 lines. Cite file:line. No code edits.
Use: codex exec --model gpt-5.3-codex
```

- [ ] **Step 0.2: Apply plan corrections from Codex feedback**

If Codex flags issue → edit plan inline. If clean → proceed.

- [ ] **Step 0.3: Commit plan**

```bash
git add docs/superpowers/plans/2026-05-10-template-wizard-architecture.md
git commit -m "docs(plans): template wizard architecture hardening plan"
```

---

## Phase 1 — Smell 1: Reducer-owned navigability (Codex)

**Goal:** Single SSOT for step navigability. Stepper click, footer advance, URL deep-link all funnel through one path. Reducer enforces invariant.

### Task 1.1: Codex refactor — reducer + page

- [ ] **Step 1.1.1: Dispatch codex subagent (Phase 1 work)**

Dispatch `codex:codex-rescue` (sequential — wait for completion before Phase 2):

```
Caveman mode prompt:

Refactor template wizard navigability to single SSOT.

Files:
- frontend/apps/web/src/features/templates/state/templateWizard.reducer.ts
- frontend/apps/web/src/features/templates/pages/TemplateWizardPage.tsx
- frontend/apps/web/src/features/templates/components/wizard/steps/StepScope.tsx (prop wiring only)
- frontend/apps/web/src/features/templates/components/wizard/steps/StepIdentity.tsx (prop wiring only)
- frontend/apps/web/src/features/templates/components/wizard/steps/StepStructure.tsx (prop wiring only — DO NOT touch radio cards, Phase 2 owns those)
- frontend/apps/web/src/features/templates/components/wizard/steps/StepPermissions.tsx (prop wiring only)
- frontend/apps/web/src/features/templates/components/wizard/steps/StepConfirmation.tsx (prop wiring only)

Reducer changes:
1. Add pure selector: export function selectMaxReachableStep(state: TemplateWizardState): TemplateWizardStep
   - Returns highest step user has unlocked given current state
   - Step 1 always reachable
   - Step 2 if scopeType !== null AND (scopeType === 'generic' OR profileCode !== null)
   - Step 3 if step 2 reachable AND name.trim().length >= 3
   - Step 4 if step 3 reachable AND startingPoint !== null AND (startingPoint === 'blank' OR selectedDocxName !== null)
   - Step 5 if step 4 reachable AND NOT (permissionsMode === 'roles' AND selectedRoleIds.length === 0)
2. GO_TO_STEP reducer case: clamp action.step to selectMaxReachableStep(state). If requested > max, ignore (no-op return state).
3. **Auto-clamp on every action.** Wrap the existing switch result: `const next = <existing-switch-result>; const max = selectMaxReachableStep(next); return next.step > max ? { ...next, step: max } : next;` This protects against state.step desyncing from reachability when a prerequisite is cleared (e.g. user on step 4 toggles off the only selected role — step 4 must auto-fall back to step 3).
4. Export selectMaxReachableStep from same module.

**Known limitation noted (not fixed this plan):** if the user manually edits `?step=N` in the URL bar to a higher step mid-session, the URL-state sync `useEffect` on line 52 only writes state→URL, not URL→state. The `parseStepParam` initial value handles deep-link page load. This stays as a follow-up. Document in commit message.

Page changes:
1. Delete: handleAdvanceFromStep1..4, handleBackToStep1..4 (8 functions)
2. Add: const goToStep = useCallback((step: TemplateWizardStep) => dispatch({ type: 'GO_TO_STEP', step }), [])
3. Delete: step1Disabled..step5Disabled inline consts
4. Compute: const maxStep = selectMaxReachableStep(state); const advanceDisabled = maxStep <= state.step
5. handleStepClick: parseStepParam(id) instead of raw cast (uses existing helper)
6. Pass to each step: onAdvance={() => goToStep((state.step + 1) as TemplateWizardStep)}, onBack={() => goToStep((state.step - 1) as TemplateWizardStep)}, advanceDisabled
7. Step 1 onCancel preserved; StepConfirmation onSubmit preserved (mocked navigate).

Step file changes:
- Each step prop signature unchanged (still onAdvance, onBack, advanceDisabled). No internal logic touched. Just confirm no breakage.

CRITICAL constraints:
- Do NOT touch reducer action types except GO_TO_STEP behavior.
- Do NOT add new actions.
- Do NOT touch StepStructure radiogroup JSX (Phase 2 territory — explicit boundary).
- Do NOT touch StepConfirmation prop list (Phase 3 owns the dead-prop strip).
- Do NOT run tsc, do NOT commit, do NOT touch tests.

After edit: list every file changed + line ranges. Report any blocker.

Use: codex exec --model gpt-5.3-codex
```

- [ ] **Step 1.1.2: Read Codex output, verify scope**

Read what Codex changed. Confirm: only listed files touched, no test/commit/tsc invocations, no scope creep into Phase 2/3 files.

- [ ] **Step 1.1.3: /simplify pass on reducer + page**

Mental review of changed files. Look for: unnecessary `useMemo`, redundant casts, dead branches, names that lie. Inline-edit any that fail the test.

- [ ] **Step 1.1.4: Run tsc (templates feature only check)**

```bash
cd frontend/apps/web && pnpm.cmd tsc --noEmit -p tsconfig.build.json 2>&1 | grep "src/features/templates"
```

Expected: zero errors in `src/features/templates/`. Pre-existing errors in other features ignored.

- [ ] **Step 1.1.5: Commit Phase 1**

```bash
git add frontend/apps/web/src/features/templates/state/templateWizard.reducer.ts frontend/apps/web/src/features/templates/pages/TemplateWizardPage.tsx frontend/apps/web/src/features/templates/components/wizard/steps/Step*.tsx
git commit -m "$(cat <<'EOF'
refactor(template-wizard): reducer owns step navigability (single SSOT)

- Add selectMaxReachableStep selector to reducer; gate predicates derived from state
- GO_TO_STEP reducer case clamps target to maxReachableStep (impossible to skip gates)
- Page collapses 8 handle{Advance,Back}FromStepN handlers into one goToStep callback
- Stepper click + Footer advance + URL deep-link all funnel through reducer
- Removes M1 (gate bypass via stepper) by construction
- Removes N4 (handler proliferation) and N2 (unchecked step cast)

No behavior change for valid navigation paths; invalid paths now no-op instead of dispatching invalid state.

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Phase 2 — Smell 2: Roving radiogroup primitive (Codex, parallel with Phase 1)

**Goal:** ARIA radiogroup pattern (parent role + roving tabIndex + arrow-key nav) lives in one place. Consumers cannot regress.

### Task 2.1: Create headless hook + adapt SelectableCard + consumers

- [ ] **Step 2.1.1: Dispatch codex subagent (Phase 2 work)**

Dispatch `codex:codex-rescue` (sequential — only after Phase 1 commit landed; Phase 2 rebases from Phase 1's StepScope/StepStructure prop wiring):

```
Caveman mode prompt:

Goal: extract ARIA radiogroup pattern (parent role + roving tabindex + arrow keys) into headless hook. Apply to StepScope (2 grids) + StepStructure (1 grid).

Files:
- CREATE: frontend/apps/web/src/components/ui/useRovingRadioGroup.ts
- MODIFY: frontend/apps/web/src/components/ui/SelectableCard.tsx
- MODIFY: frontend/apps/web/src/features/templates/components/wizard/steps/StepScope.tsx (radiogroup wiring + N1 inline-style fix)
- MODIFY: frontend/apps/web/src/features/templates/components/wizard/steps/StepScope.module.css (add .profileSectionKicker)
- MODIFY: frontend/apps/web/src/features/templates/components/wizard/steps/StepStructure.tsx (radiogroup wiring on the 2 starting-point cards)

Hook contract (useRovingRadioGroup.ts):

```typescript
import { useRef, useCallback, type KeyboardEvent, type RefCallback } from 'react';

export type RovingRadioGroupConfig = {
  count: number;
  selectedIndex: number; // -1 if none
  onSelect: (index: number) => void;
  ariaLabel: string;
  orientation?: 'horizontal' | 'vertical' | 'both'; // default 'both' (all 4 arrows)
};

export type RovingRadioGroupResult = {
  groupProps: {
    role: 'radiogroup';
    'aria-label': string;
  };
  getItemProps: (index: number) => {
    ref: RefCallback<HTMLElement | null>;
    tabIndex: number;
    onKeyDown: (e: KeyboardEvent<HTMLElement>) => void;
  };
};

export function useRovingRadioGroup(config: RovingRadioGroupConfig): RovingRadioGroupResult;
```

Behavior:
- tabIndex: selected item = 0; if no selection, first item = 0; others = -1.
- onKeyDown: ArrowRight/ArrowDown → next index (wrap); ArrowLeft/ArrowUp → prev index (wrap); Home → 0; End → count-1. On move, call onSelect(newIndex) AND focus refs[newIndex].
- orientation 'horizontal' → only ArrowLeft/Right; 'vertical' → only ArrowUp/Down; 'both' (default) → all 4.

SelectableCard changes:
- Wrap component export in React.forwardRef<HTMLButtonElement, SelectableCardProps>.
- Add optional tabIndex?: number to props; default behavior unchanged when omitted (default tabIndex=0 from native button).
- When tabIndex prop passed, spread onto <button>.
- Spread an optional onKeyDown prop onto <button> too (additive — allow callers to pass keyboard handlers).
- Keep role="radio" + aria-checked default (no opt-out).

StepScope changes:
1. Two radiogroups: scope-type (2 cards) + profile (N cards, only when scopeType==='profile').
2. Compute scopeIndex: scopeType === 'generic' ? 0 : scopeType === 'profile' ? 1 : -1.
3. const scopeGroup = useRovingRadioGroup({ count: 2, selectedIndex: scopeIndex, onSelect: (i) => onSelectScopeType(i === 0 ? 'generic' : 'profile'), ariaLabel: 'Tipo de escopo do template' }).
4. Wrap scope grid div with {...scopeGroup.groupProps}. Each SelectableCard gets {...scopeGroup.getItemProps(0)} / (1).
5. Profile grid (only renders when profileGroup applicable): const profileIndex = profiles.findIndex(p => p.code === selectedCode); const profileGroup = useRovingRadioGroup({ count: profiles.length, selectedIndex: profileIndex, onSelect: (i) => onSelect(profiles[i].code), ariaLabel: 'Perfil do template' }).
6. Wrap profile grid div with {...profileGroup.groupProps}. Each SelectableCard maps to {...profileGroup.getItemProps(idx)}.
7. Disabled profiles: still call onSelect but the SelectableCard's disabled prop blocks the click. Hook does not need to skip (acceptable — user can still arrow over but Enter/Space won't act). Document this behavior in a one-line comment.
8. N1 fix: replace inline style={{marginBottom:8}} on .kicker with className={styles.profileSectionKicker}. Add the class to StepScope.module.css with margin-bottom: var(--sp-2).

StepStructure changes:
1. Replace the 2 hand-rolled <button role="radio"> elements with native buttons consuming useRovingRadioGroup.
2. const startIndex = startingPoint === 'docx' ? 0 : startingPoint === 'blank' ? 1 : -1.
3. const startGroup = useRovingRadioGroup({ count: 2, selectedIndex: startIndex, onSelect: (i) => i === 0 ? pickDocx() : pickBlank(), ariaLabel: 'Ponto de partida do template', orientation: 'horizontal' }).
4. Replace existing role="radiogroup" div with {...startGroup.groupProps} (drop hand-rolled aria-label which is now in groupProps).
5. Each card button: spread {...startGroup.getItemProps(0)} / (1). Keep existing className/onClick wiring (onClick still triggers pickDocx/pickBlank for mouse users).
6. Remove the hand-rolled aria-checked attribute on the buttons — buttons keep role=radio + aria-checked themselves (set explicitly, hook doesn't manage that).

CRITICAL constraints:
- Do NOT touch reducer or page (Phase 1 territory).
- Do NOT touch StepConfirmation (Phase 3).
- Do NOT touch StepPermissions (its mode tabs already work and use a different visual primitive).
- Do NOT add new dependencies.
- Hook must NOT use useEffect for ref management — RefCallback only.
- Do NOT run tsc, do NOT commit, do NOT touch tests.

After edit: list every file changed + line ranges. Report any blocker.

Use: codex exec --model gpt-5.3-codex
```

- [ ] **Step 2.1.2: Read Codex output, verify scope**

Confirm only listed files touched. No test/commit/tsc.

- [ ] **Step 2.1.3: /simplify pass on hook + consumers**

Look for: unnecessary refs, redundant memoization, off-by-one, dead branches.

- [ ] **Step 2.1.4: Run tsc**

```bash
cd frontend/apps/web && pnpm.cmd tsc --noEmit -p tsconfig.build.json 2>&1 | grep -E "(useRovingRadioGroup|SelectableCard|StepScope|StepStructure)"
```

Expected: zero errors in scope.

- [ ] **Step 2.1.5: Commit Phase 2**

```bash
git add frontend/apps/web/src/components/ui/useRovingRadioGroup.ts frontend/apps/web/src/components/ui/SelectableCard.tsx frontend/apps/web/src/features/templates/components/wizard/steps/StepScope.tsx frontend/apps/web/src/features/templates/components/wizard/steps/StepScope.module.css frontend/apps/web/src/features/templates/components/wizard/steps/StepStructure.tsx
git commit -m "$(cat <<'EOF'
refactor(ui): extract roving radiogroup pattern to headless hook

- New: useRovingRadioGroup hook owns role=radiogroup parent, roving tabIndex, arrow/Home/End nav
- SelectableCard: forwardRef + optional tabIndex/onKeyDown props (additive, default unchanged)
- StepScope: scope grid + profile grid both consume hook (closes M2 — orphaned role=radio)
- StepStructure: starting-point cards consume hook (closes M3 — missing arrow nav, N5 — missing aria-label)
- StepScope: kill inline style={{marginBottom:8}} → .profileSectionKicker class (closes N1)

ARIA radiogroup invariant now enforced by primitive shape; new step radiogroups cannot regress by forgetting parent role or keyboard handler.

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Phase 3 — Smell 3: Strip dead props (Sonnet inline, no Codex)

### Task 3.1: Drop description + scopeType from StepConfirmation

- [ ] **Step 3.1.1: Edit StepConfirmation.tsx**

Remove from `StepConfirmationProps`:

```typescript
description: string;
scopeType: 'generic' | 'profile' | null;
```

Remove from destructure:

```typescript
description,
scopeType,
```

Remove the two `void` lines:

```typescript
void description;
void scopeType;
```

- [ ] **Step 3.1.2: Edit TemplateWizardPage.tsx — remove props from call site**

Delete from the StepConfirmation JSX (lines around 235–255):

```tsx
description={state.description}
scopeType={state.scopeType}
```

- [ ] **Step 3.1.3: Run tsc**

```bash
cd frontend/apps/web && pnpm.cmd tsc --noEmit -p tsconfig.build.json 2>&1 | grep "StepConfirmation"
```

Expected: zero errors.

- [ ] **Step 3.1.4: Commit Phase 3**

```bash
git add frontend/apps/web/src/features/templates/components/wizard/steps/StepConfirmation.tsx frontend/apps/web/src/features/templates/pages/TemplateWizardPage.tsx
git commit -m "$(cat <<'EOF'
refactor(template-wizard): strip unused description + scopeType props from StepConfirmation

YAGNI: props were accepted then immediately suppressed with void. Removed prop surface and caller plumbing. When backend wiring lands (confirmacao-backend-submit backlog), reintroduce with real consumer in same change.

Closes M4.

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Phase 4 — N1 already handled in Phase 2

Inline style fix bundled with Phase 2 commit (single visit to `StepScope.tsx` + `.module.css`). No separate phase needed.

---

## Phase 5 — /simplify final pass (Sonnet inline)

### Task 5.1: Whole-feature simplify scan

- [ ] **Step 5.1.1: Read all modified files in one pass**

Read in single message: reducer, page, useRovingRadioGroup, SelectableCard, StepScope, StepStructure, StepConfirmation.

- [ ] **Step 5.1.2: Apply simplify pass**

Look for: dead imports, unused vars introduced by refactor, redundant `useCallback`/`useMemo`, names that lie, comment debt that no longer matches code. Inline-edit anything found. If nothing found, no commit.

- [ ] **Step 5.1.3: If edits made, commit**

```bash
git add <files>
git commit -m "refactor(template-wizard): simplify pass after architecture hardening"
```

If no edits: skip commit.

---

## Phase 6 — Independent senior review (Opus → frontend-code-reviewer)

### Task 6.1: Dispatch frontend-code-reviewer

- [ ] **Step 6.1.1: Dispatch reviewer**

Dispatch `frontend-code-reviewer` subagent (Opus). Self-contained prompt:

```
Senior code review of template wizard architecture refactor (3 commits).

Branch: main, last 3 commits cover:
1. Reducer owns step navigability (selectMaxReachableStep + GO_TO_STEP clamp)
2. useRovingRadioGroup hook + SelectableCard forwardRef + StepScope/StepStructure adoption
3. StepConfirmation dead-prop strip (description, scopeType)

Files in scope:
- frontend/apps/web/src/features/templates/state/templateWizard.reducer.ts
- frontend/apps/web/src/features/templates/pages/TemplateWizardPage.tsx
- frontend/apps/web/src/features/templates/components/wizard/steps/StepScope.tsx
- frontend/apps/web/src/features/templates/components/wizard/steps/StepStructure.tsx
- frontend/apps/web/src/features/templates/components/wizard/steps/StepConfirmation.tsx
- frontend/apps/web/src/components/ui/SelectableCard.tsx
- frontend/apps/web/src/components/ui/useRovingRadioGroup.ts (new)

Verify:
- M1 (gate bypass) impossible by construction — Stepper click cannot land on locked step
- M2/M3/N5 (radiogroup) — both StepScope grids + StepStructure cards have role=radiogroup parent + roving tabIndex + arrow nav
- M4 (dead props) — props gone from prop type, destructure, call site
- N1 (inline style) — replaced with CSS Module class
- N2/N4 (handler proliferation, unchecked cast) — handlers collapsed, parseStepParam used

Verify also:
- No regression of previously-good patterns (StepPermissions ARIA pattern intact, error UX intact, mock-data discipline intact)
- Hook is correctly memoized (no infinite re-render risk on parent re-render)
- forwardRef on SelectableCard does not break any existing consumer

Out of scope: M5/M6/N3/N6/N7 (deferred separately).

Report Critical/Major/Minor with APPROVE / REQUEST CHANGES verdict. Read-only.
```

- [ ] **Step 6.1.2: Address findings**

If Critical or Major: dispatch Codex (or fix inline if trivial) to resolve. Loop back to Step 5.1 if changes made. If only Minor or APPROVE: proceed to Phase 7.

- [ ] **Step 6.1.3: Write phase6-review.md artifact**

Save reviewer output verbatim to `docs/superpowers/plans/2026-05-10-template-wizard-architecture-review.md`.

```bash
git add docs/superpowers/plans/2026-05-10-template-wizard-architecture-review.md
git commit -m "docs(plans): senior review of template wizard architecture refactor"
```

---

## Phase 7 — Push

### Task 7.1: Push to origin

- [ ] **Step 7.1.1: Push**

```bash
git push origin main
```

Verify push succeeded.

---

## Phase 8 — Wiki update (wiki-curator subagent)

### Task 8.1: Dispatch wiki-curator

- [ ] **Step 8.1.1: Dispatch**

Dispatch `wiki-curator` subagent. Caveman prompt:

```
Architecture refactor commits landed on main covering:
1. Reducer-owned step navigability in features/templates/state/templateWizard.reducer.ts (selectMaxReachableStep selector added; GO_TO_STEP clamps)
2. New primitive: components/ui/useRovingRadioGroup.ts (headless ARIA radiogroup hook)
3. SelectableCard gained forwardRef + optional tabIndex/onKeyDown props
4. StepScope + StepStructure adopt the hook
5. StepConfirmation dropped unused description + scopeType props

Tasks:
- Bump Last verified stamps in any wiki doc referencing these files
- Update wiki/modules/templates-v2.md Step 5 prop list (description/scopeType removed)
- Add wiki/modules/frontend-primitives.md row for useRovingRadioGroup if file exists; otherwise note in templates-v2.md Key files
- Update wiki/README.md index entries that need it
- Do NOT touch backlog (separate concern)
- Do NOT create new docs unless an existing module clearly requires one
```

- [ ] **Step 8.1.2: Verify wiki-curator output, commit + push if changes**

If wiki-curator made commits, push:

```bash
git push origin main
```

---

## Self-review checklist (run inline before saving)

- [x] Spec coverage: M1 (Phase 1), M2 (Phase 2), M3 (Phase 2), M4 (Phase 3), N1 (Phase 2), N2 (Phase 1), N4 (Phase 1), N5 (Phase 2). Deferred per user signal: M5, M6, N3, N6, N7. ✅
- [x] No placeholders, no TBD, no "implement later". ✅
- [x] Type consistency: `selectMaxReachableStep`, `useRovingRadioGroup`, `RovingRadioGroupConfig`, `goToStep` named consistently across phases. ✅
- [x] Every code-changing step shows the code or a Codex prompt with the contract. ✅
- [x] Exact file paths everywhere. ✅
- [x] Codex parallelism: Phases 1+2 dispatched in single message. ✅
- [x] Codex never commits, never runs tests. ✅
