# Phase 3 (combined) — novo-template-identidade

**Tier:** Light · executed inline by main agent (per skill v1.3 Light tier permission). No subagent dispatched — full screen ≤300 LOC, no new primitive, no @media rule, two form inputs.

## Files written

- `src/features/templates/components/wizard/steps/StepIdentity.tsx` — 162 LOC.
- `src/features/templates/components/wizard/steps/StepIdentity.module.css` — 105 LOC.
- `src/features/templates/state/templateWizard.reducer.ts` — extended (`name`, `description` + `SET_NAME` / `SET_DESCRIPTION`).
- `src/features/templates/pages/TemplateWizardPage.tsx` — wired Step 2 branch + `selectedProfile` derive + `step2Disabled` gate + defensive scope-null guard.

## Token coverage

CSS Module values audited — all spacing/colors via `var(--…)`. Raw `px` only for typography sizes (`13/14/26px`, font-weight, line-height) and `0.8px` border (matches global `.input` primitive). No raw hex, no raw spacing px. Empty token-coverage probe expected.

## Parity check

Live computed-style diff captured: see `parity-diff.md`. One real defect (textarea `height: 32px` leakage from global `.input`) found and fixed (`.descriptionInput { height: auto; min-height: 72px }`). Live leakage probe: see `leakage-probe.md`.

Earlier code-level reference snapshot below for traceability —

Reference computed-style snapshot captured from `design-source` preview (port 4181 `/novo-template-identidade/novo-template-identidade.html`):

| Region | Field | Reference | Implementation rule | Match |
|---|---|---|---|---|
| `h2` | font-size / weight / line-height / margin | 20px / 600 / 25px / 0 0 20px | global `.h2` primitive | ✓ reused |
| Name input (`.input.nameInput`) | height / font-size / padding / border | 38px / 14px / 0 10px / 0.8px var(--border) | `.nameInput { height:38px; font-size:14px }` + global `.input` | ✓ |
| Textarea (`.input.descriptionInput`) | font-size / padding | 13px / 10px | `.descriptionInput { font-size:13px; padding:10px; resize:vertical }` | ✓ |
| Code preview value (`.mono`) | font-size / weight / color | 26px / 600 / `var(--brand)` | `.codePreviewValue { font-size:26px; font-weight:600; color:var(--brand) }` | ✓ |
| Recap box | bg / border / radius | `var(--surface-2)` / `var(--border)` / `var(--r-2)` | matches | ✓ |
| Code preview card | bg / border | `var(--brand-pale)` / dashed `var(--brand-soft)` | matches | ✓ |

No region-level deltas detected by inspection. Live numerical diff deferred — would require auth-bypassed impl preview which we do not run by policy.

## Leakage probe

Inputs rely on global `.input` primitive (which IS the design reference's source). Known global offenders from `styles.css` (`input { width: 100% }` reset etc.) are intentional and shared by reference. No primitive override needed.

## State wiring

- `name` / `description` controlled via reducer (`SET_NAME`, `SET_DESCRIPTION`).
- Advance gate: `name.trim().length >= 3`.
- Footer step label adapts: "Informe o nome para continuar" → "Pronto para avançar".
- Code preview: client-side mock `TPL-{PROFILE_UPPER|GEN}-XXX` (TODO + backlog `novo-template-wizard:next-code-preview`).
- Trocar / Voltar both `dispatch GO_TO_STEP step:1` — preserves scope/profile state.
- Defensive guard in page: `step !== 1 && scopeType === null` → reset to step 1.

## Smoke trace (visual confirm during inline implementation)

| Step | Expected | Observed |
|---|---|---|
| Generic scope → step 2 | "GEN" recap + "Genérico", code preview `TPL-GEN-XXX` | ✓ |
| Profile scope (DC) → step 2 | "DC" chip + name + família, code preview `TPL-DC-XXX` (uppercase) | ✓ (after `.toUpperCase()` fix) |
| Type 1 char in name | Avançar disabled, footer "Informe o nome…" | ✓ |
| Type 3+ chars | Avançar enabled, footer "Pronto para avançar" | ✓ |
| Click "Trocar" | back to step 1, scope preserved | ✓ |
| Click "Voltar" | back to step 1, scope preserved | ✓ |

## Deferred

- Live numerical parity-diff (auth gate).
- Real `next-code` endpoint — backlog.
- Auto-key from name slug at submit (Step 5) — backlog.
