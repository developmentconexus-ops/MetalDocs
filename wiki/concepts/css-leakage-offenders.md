# CSS Leakage Offenders

> **Last verified:** 2026-05-09
> **Scope:** Global CSS rules in `src/styles.css` that silently clobber element styles inside components. Known offenders, symptoms, and required overrides.
> **Out of scope:** CSS Modules scoping mechanics (general); design token definitions (`tokens.css`).
> **Key files:**
> - `frontend/apps/web/src/styles.css:66` — `.input { height: 32px }` — the primary leakage source for textarea elements

## What is CSS leakage?

Global utility classes in `styles.css` are authored for the default use-case (single-line text inputs). When the same class is applied to semantically different elements (`<textarea>`, `<select>` with unusual sizing), the global rule overrides what the element or browser would otherwise produce — regardless of the element's own attributes.

This is the "leakage" — a rule written for inputs bleeds into non-input elements that share the class.

The Pixel Parity Playbook §2 (leakage probe) catches these during Phase 3 assembly: render the component, inspect computed styles, compare to design spec.

---

## Known offenders

### `.input { height: 32px }` — `styles.css:66`

**What it does:** Sets a fixed 32px height on any element with class `input`.

**Who it hits:** `<textarea class="input">` — the `rows` attribute has no effect once a fixed `height` is set. The textarea collapses to 32px regardless of `rows={3}` or `rows={6}`.

**Symptom:** Textarea appears as a single-line input; content is clipped and scrollable from pixel 1.

**First caught:** Step 2 (Identidade) of Template Creation Wizard — `StepIdentity.tsx:127`.

**Required override pattern** (apply in the component's CSS Module):

```css
.descriptionInput {
  /* Override global .input { height: 32px } — let rows attr drive height. */
  height: auto;
  min-height: <design-spec>px; /* e.g. 72px for rows={3} at 13px/line */
  resize: vertical;            /* allow user resize if design permits */
}
```

Applied at: `StepIdentity.module.css:71–79`.

**Rule:** Every component that renders a `<textarea className="input ...">` MUST include this override. There is no plan to change the global rule (too many dependents). The override is the contract.

---

## Checklist: adding a textarea to a wizard step

Before shipping any step that includes a `<textarea>`:

- [ ] Does the element use class `input`? If yes, apply `height: auto; min-height: Npx` override in the CSS Module.
- [ ] Verify in browser: does the textarea visually show all `rows`? If not, override is missing.
- [ ] If the design spec calls for a non-standard height, set `min-height` to that value and note `/* design-exact */`.

---

## Deferred / not yet catalogued

Additional leakage candidates to audit as new screens ship:

- `.btn` fixed `height: 32px` (`styles.css:39`) — could affect `<a class="btn">` elements that wrap multi-line content.
- Any future use of `.input` on `<select>` with non-standard heights.

Add entries to this doc when new leakage is confirmed during Pixel Parity probes.

---

## See also

- `wiki/modules/templates.md` — Step 2 where this was first documented
- `wiki/backlog/novo-template-wizard.md#font-size-hero` — related design-token gap (26px between tokens)
