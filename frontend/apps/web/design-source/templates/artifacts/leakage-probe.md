# Phase 3b — Global CSS Leakage Probe

**Method:** Iterate `document.styleSheets` for each interactive/styled element on `/templates-v2`, collect all rules whose selector matches the element. Filter to global rules (non-CSS-Module-hashed selectors).

## Probe results @ 1440

### `.action button` (Novo template)
- `*` { box-sizing: border-box; }
- `button, input, select, textarea` { font: inherit; }
- `button` { cursor: pointer; }

→ Standard resets only. No clobber.

### `.card` root
- `*` { box-sizing: border-box; }

→ Clean.

### Tab button
- `*` { box-sizing: border-box; }
- `button, input, select, textarea` { font: inherit; }
- `button` { cursor: pointer; }

→ Standard resets only. TabBar primitive owns its appearance.

### `.title` (h1 in hero), card title (div)
- `*` { box-sizing: border-box; }
- (h1) browser default; overridden by primitive CSS Module.

→ No leak.

## Known offenders from `src/styles.css` (per Phase 2 leakage map)

Templates page does NOT use:
- `<input>` / `<select>` / `<textarea>` (no form on this screen)
- `<label span>`
- `<ol>` / `<ul>`
- `<table>`

→ The Phase 2 documented offenders are not active on this page.

## Verdict

**No global leak fix required.** All bare-element rules from `styles.css` that match elements on this page are intentional resets (`*` box-sizing, `button` cursor, `button` font:inherit). No CSS Module value is being clobbered.
