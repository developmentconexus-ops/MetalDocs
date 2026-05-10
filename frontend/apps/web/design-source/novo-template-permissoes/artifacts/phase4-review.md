# Screen Review: novo-template-permissoes

**Implementation:** `frontend/apps/web/src/features/templates/components/wizard/steps/StepPermissions.tsx` + `StepPermissions.module.css`
**Design source:** `frontend/apps/web/design-source/novo-template-permissoes/` (renders via `design-source/template-wizard.jsx` TplStep4)
**Visual comparison:** Design server (4181) live, measured via `preview_eval`. Impl server (4174) requires auth session (API on 8081 not running); impl values derived from CSS source + token resolution against `src/styles/tokens.css`.
**Verdict:** REQUEST CHANGES

---

## Critical

### Falsified Phase 3b artifacts

- [ ] `parity-diff.md` -- Nine or more rows echo design raw pixel values on both ref and impl columns without resolving impl CSS tokens. Examples: .areaCard padding ref=14px impl=14px -- impl CSS is `var(--sp-3)` = 12px (delta -2px). .areaGrid gap ref=10px impl=10px -- impl CSS is `var(--sp-2)` = 8px (delta -2px). .allBanner padding ref=18px impl=18px -- impl CSS is `var(--sp-4)` = 16px (delta -2px). The artifact was built by copying design values to both columns, not by measuring the implementation. *Why:* `metaldocs-screen-implementation/SKILL.md` Phase 3b Pixel Parity Playbook requires resolved token values to be recorded. *Fix:* Re-run parity-diff computing actual browser computed-style values from the impl server; replace all falsified rows.

- [ ] `token-coverage.txt` -- Lists design-spec raw pixel values as if they appear verbatim in the impl CSS (marked /* design-exact */), but the CSS uses tokens resolving to different values. Specific false claims: padding: 3px for .modeSegmented -- CSS has `var(--sp-1)` = 4px; padding: 14px for .areaCard/.roleCard/.coverageSummary -- CSS has `var(--sp-3)` = 12px; gap: 10px for both grids -- CSS has `var(--sp-2)` = 8px; padding: 18px for .allBanner -- CSS has `var(--sp-4)` = 16px. *Why:* same rule. *Fix:* Regenerate from actual CSS source, or mark as design target not yet met.

---

## Major

### Visual / tokens -- Spacing gaps
All deltas verified against `tokens.css:74` (sp-1=4px, sp-2=8px, sp-3=12px, sp-4=16px, sp-5=20px) and design computed-style measurements via `preview_eval` on the design server (4181).

- [ ] `StepPermissions.module.css:159` -- .areaGrid { gap: var(--sp-2) } resolves to 8px. Design: 10px (measured). Delta: -2px. *Why:* no 10px token in `tokens.css`. *Fix:* gap: 10px; /* design-exact */

- [ ] `StepPermissions.module.css:204` -- .roleGrid { gap: var(--sp-2) } resolves to 8px. Design: 10px (measured). Delta: -2px. *Fix:* same.

- [ ] `StepPermissions.module.css:165` -- .areaCard { padding: var(--sp-3) } resolves to 12px. Design: 14px (measured, all four sides). Delta: -2px. Cards are 2px tighter than spec. *Why:* `tokens.css:74` -- --sp-3: 12px. *Fix:* padding: 14px; /* design-exact */

- [ ] `StepPermissions.module.css:210` -- .roleCard { padding: var(--sp-3) } resolves to 12px. Design: 14px. Delta: -2px. *Fix:* same.

- [ ] `StepPermissions.module.css:250` -- .coverageSummary { padding: var(--sp-3) } resolves to 12px. Design: 14px (measured). Delta: -2px. *Fix:* padding: 14px; /* design-exact */

- [ ] `StepPermissions.module.css:62` -- .allBanner { padding: var(--sp-4) } resolves to 16px. Design: 18px (measured). Delta: -2px. *Fix:* padding: 18px; /* design-exact */

- [ ] `StepPermissions.module.css:61` -- .allBanner { gap: var(--sp-4) } resolves to 16px. Design: 14px (measured). Delta: +2px. Icon and body text sit 2px further apart than spec. *Fix:* gap: 14px; /* design-exact */

- [ ] `StepPermissions.module.css:66` -- .allBanner { margin-bottom: var(--sp-5) } resolves to 20px. Design: 18px (measured). Delta: +2px. *Fix:* margin-bottom: 18px; /* design-exact */

- [ ] `StepPermissions.module.css:15` -- .modeSegmented { margin-bottom: var(--sp-5) } resolves to 20px. Design: 22px (measured). Delta: -2px. *Fix:* margin-bottom: 22px; /* design-exact */

- [ ] `StepPermissions.module.css:11` -- .modeSegmented { padding: var(--sp-1) } resolves to 4px. Design: 3px (measured). Delta: +1px. Active tab pill sits 1px more inset than spec. *Fix:* padding: 3px; /* design-exact */
### Visual / tokens -- Sub-element margin-bottom gaps

- [ ] `StepPermissions.module.css:131` -- .cardCode { margin-bottom: var(--sp-1) } = 4px. Design: 2px (`template-wizard.jsx:421` and :453). Delta: +2px. No 2px token exists. *Fix:* margin-bottom: 2px; /* design-exact */

- [ ] `StepPermissions.module.css:142` -- .cardName { margin-bottom: var(--sp-1) } = 4px. Design: 2px. Delta: +2px. *Fix:* same.

- [ ] `StepPermissions.module.css:88` -- .allTitle { margin-bottom: var(--sp-1) } = 4px. Design: 2px (measured). Delta: +2px. *Fix:* same.

### Visual / tokens -- h2 font-size (cross-step systemic gap)

- [ ] `StepPermissions.tsx:81` -- h2 with className=h2 inherits `styles.css:22` global .h2 { font-size: 22px } but design specifies 20px for all template wizard step headings (`template-wizard.jsx:361`: style={{ fontSize: 20 }}; same on Steps 1-5). Delta: +2px. Same gap exists in `StepStructure.tsx:60` and all other template wizard steps -- none override .h2 font-size. *Why:* `styles.css:22` global. *Fix:* Add .container > :global(.card) > :global(.h2) { font-size: 20px; } to `WizardShell.module.css` so all wizard steps inherit the correction.

### A11y

- [ ] `StepPermissions.tsx:87-101` -- role=radiogroup with role=radio buttons does not implement arrow-key navigation. ARIA APG Radio Group pattern requires arrow keys (left/right) to move focus and update selection between options; Tab currently cycles through all three buttons individually. Screen reader users will hear radio group announced but keyboard users cannot navigate with arrows. *Why:* ARIA APG Radio Group Pattern. *Fix:* Add onKeyDown handler on the radiogroup container (ArrowRight/Down advances, ArrowLeft/Up retreats); use roving tabindex (tabIndex={active ? 0 : -1}) so only the active radio is in the tab stop.

---

## Minor

- [ ] `StepPermissions.module.css:238` -- .roleIdRow { margin-bottom: var(--sp-1) } = 4px. Design: 3px (`template-wizard.jsx:453` marginBottom: 3). Delta: +1px. *Fix:* 3px /* design-exact */ or document as accepted near-miss.

- [ ] `StepPermissions.tsx:97` -- .modeTabLabel class body is empty (comment: inherits from .modeTab). No declarations, purely a placeholder for future scoping. Not a bug.
---

## Artifact cross-check

| Artifact | Status |
|---|---|
| `parity-diff.md` | Critical -- 9+ rows falsified (impl token values not resolved) |
| `token-coverage.txt` | Critical -- claims design-exact for token-mismatched values |
| `phase3a-structure.md` | Pass -- DOM structure confirmed correct via accessibility tree snapshot |
| `phase3b-style.md` | Fail -- claims no regions with delta but 13 spacing/size deltas exist |

---

## What is good

- DOM structure accurately mirrors the design HTML across all three mode states (roles / areas / all), coverage summary, and WizardFooter wiring. Phase 3a was done correctly.
- A11y semantic choices (role=radiogroup/radio for mutex mode control; role=group/checkbox for multi-select card grids) are correct and improve on the design, which uses role=tablist/tab -- semantically wrong for a content-panel replacement pattern.
- Responsive breakpoints match NOTES.md requirements exactly: `@media (max-width: 640px)` (area grid 3-to-2-col, role grid 2-to-1-col) and `@media (max-width: 480px)` (area grid 2-to-1-col).
- State logic is correct: advance gate (mode=roles && selectedRoleIds.length===0), TOGGLE_ROLE_ID/TOGGLE_AREA_ID reducers, mode-switch state persistence, empty areas selection allowed per design spec.
- No raw hex colors or unscoped global CSS leakage -- all color values use `var(--token-*)` tokens correctly.
- Card selected state (`cardSelected` with 2px brand border + brand-pale background + `calc(var(--sp-3) - 1px)` padding compensation) follows the Step 3 pattern from `StepStructure.module.css` exactly.
- All mock data constants carry correct TODO backlog references matching NOTES.md.