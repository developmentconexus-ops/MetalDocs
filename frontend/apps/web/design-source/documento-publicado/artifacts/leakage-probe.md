# Leakage Probe — documento-publicado Phase 3b

> Produced by Phase 3b style-port subagent (2026-05-08).
> Tests performed against impl at http://localhost:4174/documents/c0be0f69-2d06-4770-be36-c9328208630d

---

## Interactive elements audited

### 1. "Visualizar documento" button (`btn btn-primary btn-lg`)

- Uses global `.btn .btn-primary .btn-lg` classes — no CSS Module wrapping.
- `button:disabled { opacity: 0.5 }` from base.css: button is enabled so not triggered.
- `.btn:disabled { opacity: 0.45; cursor: not-allowed; pointer-events: none }` from styles.css: more specific, would override base.css if disabled. No conflict.
- `button[aria-label="Font size"]`: button has no aria-label — no leak.
- Result: CLEAN

### 2. "Iniciar revisão" button (`btn`)

- Same global class analysis as above.
- No `disabled` attribute present — base.css disabled rule not triggered.
- Result: CLEAN

### 3. "Copiar link" button (`btn btn-ghost`)

- Same global class analysis.
- `.btn-ghost { border-color: transparent; background: transparent }` applied correctly.
- Result: CLEAN

### 4. Breadcrumb `<a>` elements

- No `href` set (placeholder data). No unexpected underline or pointer from global bare selectors.
- `.breadcrumbLink` sets `color: inherit; text-decoration: none` — overrides browser default underline.
- Result: CLEAN

### 5. `<nav>` element

- No global bare `nav` selector in base.css or styles.css. No leak.
- Result: CLEAN

### 6. `<header>` element

- Global `.hero h1` selector in styles.css is scoped under `.hero` class — our hero is a CSS Module class `_hero_17lwa_39`, not the `.hero` global class. No leak.
- Result: CLEAN

### 7. `<section>` elements (x2)

- No global bare `section` selector. No leak.
- Result: CLEAN

### 8. `<h1>` element

- `.display-1` global class not applied (correct — heroTitle uses CSS Module class).
- No global bare `h1` selector in base.css or styles.css.
- Result: CLEAN

### 9. `<h2>` elements

- No global bare `h2` selector. `.h2` global class not applied. No leak.
- Result: CLEAN

---

## Global selectors that touch this page

| Global selector | Effect on this page | Risk |
|---|---|---|
| `button, input, select, textarea { font: inherit }` | All 3 buttons inherit font | DESIRED — expected |
| `button { cursor: pointer }` | All 3 buttons get pointer cursor | DESIRED |
| `button:disabled { opacity: 0.5 }` | No disabled buttons currently | NONE |
| `.btn:disabled { opacity: 0.45; … }` | More specific — would win over above if disabled | DESIRED |
| `button[aria-label="Font size"] { … }` | No button with that aria-label | NONE |

---

## Conclusion

No unexpected global CSS leakage detected. All interactive elements render with correct styles. The `.btn` global system is used correctly (no CSS Module overrides needed for button styles).
