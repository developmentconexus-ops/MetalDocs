# Subagent prompt — Phase 3 Combined (Light tier)

Fresh git worktree subagent. Phase 2 produced primitives + atoms + status-meta + route stub. Your job: structure + style + state in one pass for a Light-tier screen.

## Inputs

- Worksheet: `frontend/apps/web/design-source/<SLUG>/IMPLEMENTATION.md`
- Reference HTML: `frontend/apps/web/design-source/<SLUG>/<SLUG>.html` (read it yourself)
- Page path: `frontend/apps/web/src/features/<DOMAIN>/pages/<PAGENAME>.tsx`
- CSS Module path: `<PAGENAME>.module.css`
- Tokens: `frontend/apps/web/src/styles/tokens.css`

## Required reading

- Worksheet (Phases 0+1+2).
- Reference HTML.
- Pixel Parity Playbook in `templates/subagent-phase3b.md` §1 §3 §4.

## Steps

1. **Mirror DOM.** Write TSX with same tags, nesting, DOM order as reference HTML. Class names mirror reference. Domain primitives (`SelectableCard`, `Button`, etc.) replace bare `<div>` where Phase 1 mapped them.

2. **Port styles** to CSS Module — tokens only. Verify `token-coverage.txt` empty:
   ```bash
   cd frontend/apps/web
   grep -REn '#[0-9a-fA-F]{3,8}|rgb\(|[0-9]+px' src/features/<DOMAIN>/pages/<PAGENAME>.module.css | grep -v 'var(--' | grep -vE '\b(0|1)px\b' > design-source/<SLUG>/artifacts/token-coverage.txt
   ```
   Non-empty → fix raw values.

3. **Wire state.** Query hooks (TanStack), error UX (`ApiError` + `resolveErrorMessage` + `role="alert"`), four states (loading/empty/error/success), disabled CTAs (`aria-disabled` + `title="Em breve"`), persisted state lazy `useState(() => readStored())`, `useDebouncedValue` for inputs. Semantic HTML: no `<button>` in `<button>`; non-button click rows use `<div role="button" tabIndex={0} onClick onKeyDown>` with `:focus-visible`.

4. **Parity-diff (1440 only).** Run dev server. Run Pixel Parity Playbook §1 snapshot on impl AND on design HTML preview. Per region, write to `phase3-combined.md` parity table:
   ```
   region | field | ref | impl | delta
   ```
   Any non-zero delta in spacing/typography/layout → fix in same phase. Then re-snapshot.

5. **Conditional leakage probe.** If any `<input>`/`<select>`/`<textarea>`/`<label>` rendered → run Playbook §2 → `leakage-probe.md`. Otherwise SKIP.

6. **Conditional multi-viewport.** If CSS Module has any `@media` rule → also screenshot 1024 + 375. Otherwise single 1440 ref+impl pair.

7. **Screenshots** to `artifacts/screenshots/1440-{ref,impl}.png` (and 1024/375 if conditional).

## Output `phase3-combined.md` (≤30 lines)

- Tokens added (commit hash if any)
- Page commit hash
- Parity-diff table (zero deltas required)
- Leakage probe: ran/skipped + N findings if ran
- Viewports captured (1440 always; 1024/375 if media query)
- Smoke trace: route loads, all four states reachable
- "User approved" — leave `[ ]`. Main agent / user marks.

## Hard rules

- Tokens-only. Raw hex / spacing px → fail.
- No tsc (Phase 4 main).
- Single subagent boot — do all three substeps in this run.
- Restructuring design HTML in TSX = fail.
- Self-grading parity = fail. User approves.
- Stop on missing token (ask main agent).
