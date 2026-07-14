# ADR 0063 — `EditorChrome`: promote-on-second-caller extraction + slot-based composition

- **Status:** Accepted
- **Last verified:** 2026-07-02
- **Date:** 2026-07-02
- **Scope:** Records two existing frontend decisions for `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.tsx`: (a) why it lives under `features/shared/` rather than feature-local or a generic `components/ui/` primitive, and (b) why its composition shape is named slots (`left/center/right/alert/children`) rather than compound components or render props. Closes tech-debt T-008 (`wiki/modules/editor-chrome-tech-debt.md`).
- **Depends on:** ADR 0046 (Eigenpal Anti-Corruption Layer) — `EditorChrome` wraps eigenpal-based editors but does not itself cross the ACL wall (it takes `children: ReactNode`, never touches `@eigenpal/*` types).

---

## Context

Two load-bearing decisions about `EditorChrome` exist only in code and in a general architecture doc (`wiki/architecture/frontend-structure.md`), not as standalone ADRs: the placement rule that put it in `features/shared/` at all, and the slot-based API shape chosen over alternatives.

### Verified runtime facts

- **Two, and only two, consumers today.** `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx` and `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx` both import `EditorChrome`. The component's own doc comment states this explicitly: `EditorChrome.tsx:55` — "Used by templates/TemplateEditorPage and documents/DocumentEditorPage."
- **Placement rule is stated generally in the architecture doc, and `editor-chrome/` is its cited example.** `wiki/architecture/frontend-structure.md:150-153` — "When in doubt: feature-local first. Promote when a **second** caller appears: ... Feature-coupled shared component (has domain context, used by 2+ features) → `features/shared/components/` (e.g. `editor-chrome/`)." Line 65-68 places `editor-chrome/` inside `features/shared/components/` alongside `controlled-artifact/` (ADR 0053's shared shell) — both are "used by 2+ features" promotions, not generic UI primitives.
- **Slot API: 4 named `ReactNode` slots + mandatory `children`, truthy-gated rendering.** `EditorChrome.tsx:4-17` (`EditorChromeProps`) — `left?`, `center?`, `right?`, `alert?` (all optional `ReactNode`), `children: ReactNode` (mandatory — "the eigenpal editor instance (DocxEditor / MetalDocsEditor)"). Implementation (`:57-67`) renders each optional slot only if truthy: `{left && <div className={styles.overlayLeft}>{left}</div>}`, same pattern for `center`/`right`/`alert`. `children` always renders unconditionally.
- **Slots are positioned overlays on top of eigenpal's own chrome, not a generic layout component.** Doc comment `:19-39` — `left`/`center`/`right` are absolute-positioned over eigenpal's 40px title bar (`center` sets `pointer-events: none` on its container so it doesn't block eigenpal's own title-bar interactions underneath, per the CSS-module comment referenced at line 36); `alert` is a banner below the title bar. This is fundamentally a "decorate a third-party component's chrome without modifying it" shape, not a general-purpose layout primitive.
- **Style re-export avoids duplicated primitives.** `EditorChrome.tsx:69-73` — `editorChromeStyles` re-exports the CSS module so both consumers share `.primaryBtn`/`.ghostBtn`/`.iconBtn`/`.docTitle`/`.docMeta`/`.docSep` instead of redefining them per page.

## Decision

**(a) Extraction placement: promote to `features/shared/components/` on the second caller, not before, and only when the component carries domain/feature context (not a generic UI primitive).** `EditorChrome` was extracted when `TemplateEditorPage` needed the same eigenpal-title-bar-overlay chrome `DocumentEditorPage` already had. This follows the general rule already stated at `wiki/architecture/frontend-structure.md:150-153` (feature-local first, promote on 2nd caller for feature-coupled shared components); this ADR pins `EditorChrome` as the canonical example that rule was written to cover, and makes the promotion threshold ("second caller," not "anticipated future caller") binding for this component specifically — do not further-generalize `EditorChrome` (e.g. into a fully generic overlay-chrome primitive under `components/ui/`) until a **third** consumer with a *different* underlying editor or chrome shape appears; two consumers wrapping the same eigenpal-based editor is not evidence that a generic (non-eigenpal-specific) version is needed.

**(b) Composition shape: named slots (`left/center/right/alert` + `children`), not compound components or render props.** Slots were chosen over the alternatives because:
- **Compound components** (`<EditorChrome.Left>`, `<EditorChrome.Right>`, etc.) would add API surface and context-provider machinery for a component with a fixed, small, non-recursive set of regions — overkill for 4 fixed positions.
- **Render props** (`children={(slots) => ...}`) would complicate the common case (most usages pass static JSX per slot, not a function needing chrome-internal state) for no benefit here, since `EditorChrome` doesn't expose any internal state slots need to react to.
- **Named-slot props directly matching the 4 fixed visual regions** (three overlay positions + one banner) is the simplest shape that matches the fixed, non-extensible layout eigenpal's title bar imposes. Truthy-gated rendering (`{slot && <div>...}`) keeps unused slots out of the DOM entirely rather than rendering empty wrapper divs.
- `children` is kept separate from the slots (not itself a slot) because it is mandatory and single-purpose (always the editor instance), whereas the four named props are all optional and page-specific.

## Consequences

- T-008 (`wiki/modules/editor-chrome-tech-debt.md`) is closed by this ADR.
- Future editor-chrome-style extractions elsewhere in the frontend can cite this ADR as the precedent for "promote on second caller, slots over compound-components for fixed small region counts" rather than re-deriving the reasoning each time.
- If a third eigenpal-adjacent (or non-eigenpal) chrome consumer appears with materially different region needs, re-open this decision rather than bolting ad-hoc new optional props onto `EditorChromeProps` indefinitely.
- No migration, schema change, or code change is required by this ADR — it documents and binds existing, verified runtime behavior.

## References

- `frontend/apps/web/src/features/shared/components/editor-chrome/EditorChrome.tsx` — full implementation, doc comments cited above.
- `wiki/architecture/frontend-structure.md:65-68,150-153` — general placement rule, `editor-chrome/` cited as its example.
- `wiki/modules/editor-chrome-tech-debt.md` T-008 — tech-debt row closed by this ADR.
- `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx`, `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx` — the two verified consumers.
- ADR [`0046-eigenpal-anti-corruption-layer.md`](0046-eigenpal-anti-corruption-layer.md) — the ACL wall `EditorChrome` sits outside of (it never imports `@eigenpal/*`).
- ADR [`0053-shared-controlled-artifact-view-layer.md`](0053-shared-controlled-artifact-view-layer.md) — sibling `features/shared/components/` promotion (`controlled-artifact/`), same placement rule applied to a different component.
