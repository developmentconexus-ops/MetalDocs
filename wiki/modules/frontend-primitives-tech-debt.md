# Tech Debt Register - frontend-primitives

> Companion to `wiki/modules/frontend-primitives.md`. Debt only; no fix prescriptions.

**Last verified:** 2026-07-02 (DOC-07b — T-001 closed: full components/ui inventory table added; FE-07 pass — T-002/T-003 analyzed report-only, no code change)

## Items

### T-001 · Module page scope is narrower than actual `components/ui` surface — CLOSED 2026-07-02 (DOC-07b)
- **Severity:** major (closed)
- **Surface:** `frontend/apps/web/src/components/ui/index.ts`
- **Observation (original):** doc focused on `SelectableCard` and `useRovingRadioGroup` while the package exports many additional primitives.
- **Resolution:** `wiki/modules/frontend-primitives.md` now has a "Full `components/ui/` inventory" table covering all 15 primitive/hook source files (`Icon`, `Avatar`, `CodeChip`, `StatusPill`, `Stepper`, `SelectableCard`, `TabBar`, `WorkspaceHeroHeader`, `Dialog`, `SearchBar`, `SelectMenu`, `DrawerShell`, `useRovingRadioGroup`, `FilterDropdown`, `TextFieldBox`/`DropdownFieldBox`, `TopbarDropdown`, `Logo`), noting barrel-export status and external-consumer counts verified against `frontend/apps/web/src/` at this pass.
- **New finding surfaced by this closure:** 4 files have zero external consumers (`FormFieldBox.tsx`, `FilterDropdown.tsx`, `TopbarDropdown.tsx`, `Logo.tsx`) — reported as observation in the inventory table; not fixed here (code change, different owner).
- **Evidence:** `wiki/modules/frontend-primitives.md` "Full `components/ui/` inventory" section; consumer counts gathered via grep across `frontend/apps/web/src/` on 2026-07-02.
- **Linked backlog row:** `R-001` (closed)
- **Linked ADR:** missing-ADR (doc-completeness item, not a design decision — no ADR needed)

### T-002 · TabBar still uses bespoke roving logic instead of shared hook — analyzed 2026-07-02, still open
- **Severity:** minor
- **Surface:** `frontend/apps/web/src/components/ui/TabBar.tsx` (bespoke), `frontend/apps/web/src/components/ui/useRovingRadioGroup.ts` (shared hook)
- **Observation:** keyboard navigation logic is duplicated instead of unified on shared hook. `TabBar.tsx:18-32` hand-rolls a `useRef<Array<HTMLButtonElement|null>>` + `handleKeyDown` (ArrowLeft/ArrowRight/Home/End, horizontal-only wraparound) that reimplements what `useRovingRadioGroup` already provides.
- **Why not a mechanical swap (root cause of why this stayed open):** `useRovingRadioGroup`'s `groupProps.role` is hardcoded to `'radiogroup'` (line 75) and its item contract has no `aria-selected`/`role` output — it's shaped for a radio-group consumer. `TabBar` renders `role="tablist"` / `role="tab"` / `aria-selected={isActive}` (WAI-ARIA Tabs pattern), which is a distinct ARIA pattern from Radio Group: tabs use `aria-selected` on `role="tab"` children under `role="tablist"`, radio groups use `aria-checked` on `role="radio"` children under `role="radiogroup"`. Pointing `TabBar` at the hook as-is would either (a) silently ship the wrong ARIA role pair (radiogroup semantics under a visually-tabbed UI — a regression, not a refactor), or (b) require generalizing the hook's `groupProps.role`/`getItemProps` return shape to accept a role-pair parameter, which is a shared-primitive API change with its own review surface, not a TabBar-local fix.
- **Recommended refactor (not implemented, per FE-07 report-only instruction):** generalize `useRovingRadioGroup` into a role-parameterized `useRovingTabIndex({ count, selectedIndex, onSelect, orientation, roles: { group: 'radiogroup'|'tablist', item: 'radio'|'tab' } })`, or extract a role-agnostic `useRovingIndex` (index math + focus + keydown only, no ARIA prop emission) that both `SelectableCard`'s radiogroup usage and `TabBar`'s tablist usage compose over, each adding their own `aria-selected`/`aria-checked`. The role-agnostic extraction is the lower-risk path since it doesn't touch the current hook's public contract for existing radiogroup consumers.
- **Evidence:** `TabBar.tsx:1-61` (full file, keydown handler at :20-32), `useRovingRadioGroup.ts:18-24,73-88` (`RovingRadioGroupResult.groupProps.role` fixed to `'radiogroup'`, `getItemProps` returns no ARIA state props).
- **Linked backlog row:** `R-002`
- **Linked ADR:** missing-ADR

### T-003 · Primitive governance boundaries are convention-only — analyzed 2026-07-02, still open
- **Severity:** minor
- **Surface:** `wiki/architecture/frontend-structure.md`, `frontend/apps/web/src/components/ui/**`
- **Observation:** "domain-agnostic only" rule exists in docs but no enforcement rule/check — nothing stops a future PR from importing a `features/**` type or hook into `components/ui/**`, or from adding a domain-specific prop (e.g. a `DocumentStatus`-shaped prop) to a primitive.
- **Recommended rule (not implemented, per FE-07 report-only instruction):** an ESLint `no-restricted-imports` (or a custom rule via `eslint-plugin-boundaries`) scoped to `frontend/apps/web/src/components/ui/**` that forbids importing from `frontend/apps/web/src/features/**`. This is a static, zero-runtime-cost guard consistent with the existing `check-css-token-discipline.sh`/`check-eigenpal-selector-pin.sh` bash-guard pattern used elsewhere in this repo, but as an ESLint rule (not a bash script) since the violation is import-graph shaped, not text-pattern shaped, and ESLint already runs in this repo's toolchain. One pre-existing exception to flag before adding the rule: `StatusPill.tsx` exports a `DocumentStatus` type — worth a one-time audit of whether that's a boundary violation already baked in, or a deliberately shared vocabulary type; the rule addition should not silently break that export without a decision.
- **Evidence:** no lint config or CI check currently references `components/ui` import boundaries (confirmed via absence in `frontend/apps/web/.eslintrc*` / `eslint.config.*` at time of this pass).
- **Linked backlog row:** `R-003`
- **Linked ADR:** missing-ADR

## Coverage stats

- Public symbols undocumented: 0 (all 15 `components/ui` files now listed in the inventory table, T-001 closed)
- Operations missing C4 placement: n/a (frontend component module)
- Cross-deps missing in section map: n/a (partial module doc)
- State transitions missing: n/a
- Decisions without ADR link: 2 (T-002, T-003 — T-001 closed)
