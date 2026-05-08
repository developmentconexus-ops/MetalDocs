# Global CSS Leakage Map — documento-publicado

> Preliminary — produced by Phase 2 Pre-flight subagent (2026-05-08).
> Phase 3 assembly subagent must reset or isolate any HIGH-risk items.

Sources audited:
- `src/styles/base.css`
- `src/styles.css` (via `@import "./styles/base.css"`)

---

## Bare-element selectors (unscoped)

### From `src/styles/base.css`

| Selector | Properties | Risk for documento-publicado |
|---|---|---|
| `button, input, select, textarea` | `font: inherit` | MEDIUM — buttons and any form elements inherit doc font. Neutral but expected. |
| `button` | `cursor: pointer` | LOW — all buttons get pointer cursor. Desired behavior. |
| `button:disabled` | `opacity: 0.5; cursor: not-allowed` | MEDIUM — disabled "Iniciar revisão" button will get opacity 0.5 automatically. Check design wants 0.45 (btn system) vs 0.5. |
| `input:not([type="checkbox"]):not([type="radio"]):not([type="file"]):not([type="range"]):not([type="color"]):not([type="image"]), select, textarea` | `width: 100%; border-radius: var(--r-2); border: 1px solid var(--border); background: var(--surface); color: var(--text-soft); padding: 0.5rem 0.75rem` | LOW — documento-publicado is read-only display; no bare input/select/textarea expected in this screen. If any future input is added (e.g. comment field), this will force width:100%. |
| `textarea` | `resize: vertical` | LOW — no textarea in documento-publicado (comments deferred). |

### From `src/styles.css`

| Selector | Properties | Risk |
|---|---|---|
| `table` | `width: 100%; border-collapse: collapse` | LOW — no bare `<table>` expected in this screen. SignoffPipeline is built with divs. |
| `th, td` | `padding: 0.95rem 1rem; border-bottom: 1px solid var(--border); text-align: left; vertical-align: top` | LOW — no table cells expected. |
| `tr` | `cursor: pointer` | LOW — no table rows expected. |
| `tr:hover, .row-active` | `background: var(--brand-pale)` | LOW — no table rows expected. |

---

## Scoped selectors targeting bare elements (safe — NOT leaking)

These target bare elements but are safely gated behind a component-level ancestor class.
Listed here so Phase 3 implementer knows they won't leak into the module.

| Selector | Scoped under | Risk |
|---|---|---|
| `.hero h1` | `.hero` | None — our hero is a new CSS Module, not `.hero` |
| `.create-doc-docx-dropzone input` | `.create-doc-docx-dropzone` | None |
| `.content-builder-field input, select, textarea` | `.content-builder-field` | None |
| `.content-builder-table-row select, input` | `.content-builder-table-row` | None |
| `.content-builder-checklist-row input[type="checkbox"]` | `.content-builder-checklist-row` | None |
| `.create-doc-content input, select, textarea` | `.create-doc-content` | None |
| `.metadata-row input` | `.metadata-row` | None |
| `.panel-heading h2, .card h3` | `.panel-heading`, `.card` | None — our cards are CSS Module classes, not `.card` |
| `.create-doc-section-head h3` | `.create-doc-section-head` | None |
| `.create-doc-page-title p` | `.create-doc-page-title` | None |
| `.toast-body strong` | `.toast-body` | None |
| `.workspace-breadcrumb span` | `.workspace-breadcrumb` | None — our breadcrumb uses CSS Module class |
| `.catalog-stat span, .catalog-alert span, .catalog-info-grid span, .diff-grid span` | respective ancestors | None |
| `.create-doc-step-item strong, small` | `.create-doc-step-item` | None |
| `.content-builder-section-head strong, small` | `.content-builder-section-head` | None |
| `.content-builder-field span` | `.content-builder-field` | None |
| `.preview-list li` | `.preview-list` | None |
| `.preview-checklist li` | `.preview-checklist` | None |
| `.preview-table th, td` | `.preview-table` | None |
| `.approvals-list li p` | `.approvals-list` | None |
| `.card h3` | `.card` | None — our cards are scoped CSS Module classes |
| `body > .ep-root button.text-left` | `body > .ep-root` | None — eigenpal-specific |
| `button[aria-label="Font size"]` | aria-label attribute | MEDIUM — this is an unscoped attribute selector. If any button in the documento-publicado page has `aria-label="Font size"`, it would receive forced sizing. Unlikely but note it. |

---

## Phase 3 action items

**HIGH priority resets needed in DocumentPublishedPage.module.css or component modules:**

None. The documento-publicado screen is a read-only display screen with no forms.

**MEDIUM — monitor:**
- `button:disabled { opacity: 0.5 }`: The "Iniciar revisão" button uses `disabled` state. The `.btn:disabled` rule in styles.css overrides to `opacity: 0.45`. Confirm the more-specific `.btn:disabled` rule wins (it will, since it's more specific). No action needed.
- `button[aria-label="Font size"]`: Confirm no button on this page uses that aria-label.

**LOW — no action needed:**
- Input/table leakage rules are irrelevant since the page has no bare inputs or tables.
