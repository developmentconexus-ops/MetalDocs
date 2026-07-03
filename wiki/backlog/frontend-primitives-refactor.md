# Refactor Backlog - frontend-primitives

**Last verified:** 2026-07-02 (DOC-07b — R-001 closed; FE-07 — R-002/R-003 analyzed report-only, still open)

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-001 | Expand frontend-primitives module coverage to full exported UI surface | T-001 | M | major | - | - | closed 2026-07-02 | - |
| R-002 | Normalize roving keyboard behavior by migrating TabBar to shared hook | T-002 | S | minor | - | - | open — see analysis | - |
| R-003 | Add governance rule for `components/ui` domain-agnostic boundary | T-003 | S | minor | - | - | open — see analysis | - |

## R-001 closure evidence (DOC-07b)

- `wiki/modules/frontend-primitives.md` now has a full inventory table for all 15 `components/ui` primitive/hook files, with barrel-export status and consumer counts verified against `frontend/apps/web/src/`.
- Surfaced (not fixed — code change, different owner): `FormFieldBox.tsx` (`TextFieldBox`/`DropdownFieldBox`), `FilterDropdown.tsx`, `TopbarDropdown.tsx`, and `Logo.tsx` have zero consumers in the current tree.
- See `wiki/modules/frontend-primitives-tech-debt.md` T-001 (closed) for full citation.

## R-002 / R-003 analysis (FE-07, 2026-07-02, report-only — no code change)

Both items were scoped as report-only for this pass (design-system hygiene wave); the following is the recommended shape for a future PR, not an implementation.

- **R-002 (TabBar → shared hook):** effort re-estimated **S → M**. `useRovingRadioGroup` hardcodes `groupProps.role: 'radiogroup'` and emits no `aria-selected`/`role="tab"` item props — it is shaped for the Radio Group ARIA pattern, while `TabBar` correctly implements the distinct Tabs ARIA pattern (`role="tablist"`/`role="tab"`/`aria-selected`). A mechanical migration would ship wrong ARIA roles. The real fix is extracting a role-agnostic `useRovingIndex` (index math + focus + keydown, no ARIA emission) that both `SelectableCard` (radiogroup) and `TabBar` (tablist) compose over, each layering their own ARIA props. Full detail in `wiki/modules/frontend-primitives-tech-debt.md` T-002.
- **R-003 (governance rule):** recommend an ESLint `no-restricted-imports`/`eslint-plugin-boundaries` rule forbidding `components/ui/**` → `features/**` imports, following this repo's existing bash-guard pattern (`scripts/check-css-token-discipline.sh`, `scripts/check-eigenpal-selector-pin.sh`) but as a lint rule since the violation is import-graph shaped. Flag before implementing: `StatusPill.tsx` already exports a `DocumentStatus` type from `components/ui` — needs a one-time audit (deliberate shared vocabulary vs. pre-existing boundary violation) before the rule lands, so it doesn't silently break that export. Full detail in `wiki/modules/frontend-primitives-tech-debt.md` T-003.
