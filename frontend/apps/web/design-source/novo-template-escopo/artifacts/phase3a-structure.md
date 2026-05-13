# Phase 3a — Structure Mirror: novo-template-escopo

## Files Created / Modified

| Action   | Path |
|----------|------|
| Created  | `src/features/templates/pages/TemplateWizardPage.tsx` |
| Created  | `src/features/templates/pages/TemplateWizardPage.module.css` |
| Created  | `src/features/templates/components/wizard/steps/StepScope.tsx` |
| Created  | `src/features/templates/components/wizard/steps/StepScope.module.css` |
| Created  | `src/features/shared/components/wizard/WizardFooter.tsx` |
| Created  | `src/features/shared/components/wizard/WizardFooter.module.css` |
| Modified | `src/features/templates/routes.tsx` — added `templates/novo` route stub |

## Class-Name Mapping

| Design element (inline style / class) | TSX CSS Module class | File |
|---------------------------------------|----------------------|------|
| profile grid container (`display: grid; repeat(2, 1fr)`) | `styles.profileGrid` | StepScope.module.css |
| profile card button | `styles.profileCard` | StepScope.module.css |
| profile card header row (code + name + badges) | `styles.profileHeader` | StepScope.module.css |
| profile code mono span | `styles.profileCode` | StepScope.module.css |
| profile name span | `styles.profileName` | StepScope.module.css |
| profile meta row (family + count) | `styles.profileMeta` | StepScope.module.css |
| page scroll wrapper | `styles.scrollWrapper` | TemplateWizardPage.module.css |
| page container (max-width centered) | `styles.container` | TemplateWizardPage.module.css |
| page header block | `styles.header` | TemplateWizardPage.module.css |
| page description paragraph | `styles.description` | TemplateWizardPage.module.css |
| footer row | `styles.footerRow` | WizardFooter.module.css |
| footer divider | `styles.footerDivider` | WizardFooter.module.css |

## Primitive Substitution Table

| Design element | Primitive used | Notes |
|----------------|----------------|-------|
| Profile card `<button>` with selected state | `SelectableCard` from `components/ui/SelectableCard.tsx` | Has native `disabled` prop — CHK card passes `disabled={true}` |
| Stepper (`TplStepper`) | `Stepper` from `components/ui/Stepper.tsx` | Used directly in page; accepts `steps[]` + `current` string |
| Wizard footer (`TplFooter`) | `WizardFooter` from `features/shared/components/wizard/WizardFooter.tsx` | See structural decision below |
| Check icon `<Icon name="check"/>` | `Icon` from `components/ui/Icon.tsx` | `check` is a valid `IconName` |
| Page shell (`TplShell`) | Inline layout in `TemplateWizardPage.tsx` | See structural decision below |
| Global classes `.kicker`, `.h2`, `.caption`, `.mono`, `.pill` | Used as-is via global CSS | Not CSS Modules — per codebase convention |

## Structural Decisions

### 1. No shared `WizardShell` component (Phase 2 gap)
Phase 2 was supposed to create a parameterizable `WizardShell` in `features/shared/components/wizard/WizardShell.tsx`. That component was not created. The documents `WizardShell` is not usable by templates because it hardcodes 4 steps and Documentos-specific title/kicker.

**Decision:** Inline the shell layout (scrollWrapper → container → header → Stepper → children) directly in `TemplateWizardPage.tsx`. This is the same pattern as the documents wizard page. Phase 3c can extract a shared shell if desired, but inlining avoids a cross-feature dependency and is simpler.

### 2. Shared `WizardFooter` extracted to `features/shared`
The documents `WizardFooter` imports from `WizardShell.module.css` (same-directory) for footer CSS. A re-export from shared would require importing across feature boundaries. Created a standalone `features/shared/components/wizard/WizardFooter.tsx` + `WizardFooter.module.css` with the same props API and a minimal CSS stub (Phase 3b ports styles).

### 3. TabBar entirely cut (phase0 decision confirmed)
Design had a two-tab bar: "Para um perfil" (active) + "A partir de um documento" (disabled). Phase 0 audit decision: cut the TabBar entirely — only profile-based scope is supported. The "document-based scope" concept is invalid for the template wizard at this stage. A comment in `StepScope.tsx` documents this decision.

### 4. `disabled` prop on `SelectableCard`
`SelectableCard` has a native `disabled?: boolean` prop (confirmed in source). CHK profile card passes `disabled={profile.disabled}`. No workaround needed.

### 5. Template count deferred
Profile cards show `— templates publicados` with a TODO comment referencing `wiki/backlog/novo-template-wizard.md`. No summary endpoint exists today. Mirrors the same pattern used in `StepProfile.tsx` for document counts.

### 6. Route stub added
`routes.tsx` did not have the wizard route (Phase 2 gap). Added `templates/novo` pointing to `TemplateWizardPage` with lazy import.

## TSC Result

```
npx tsc --noEmit -p tsconfig.build.json 2>&1 | grep -E "TemplateWizard|StepScope|WizardFooter.*shared"
(no output — zero errors in new files)
```

Pre-existing errors in the repo (missing node_modules: react-router-dom, @tanstack/react-query, @tiptap/* not installed in worktree) are unrelated to Phase 3a changes.

## Open Questions

None. All design elements mapped cleanly.
