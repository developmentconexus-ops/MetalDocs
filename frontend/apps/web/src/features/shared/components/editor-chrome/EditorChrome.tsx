import type { ReactNode } from 'react';
import styles from './EditorChrome.module.css';

export type EditorChromeProps = {
  /** Slot rendered top-left over eigenpal title bar (back btn, etc.). */
  left?: ReactNode;
  /**
   * Slot rendered centered over eigenpal title bar (title, badges, pill).
   *
   * FE-06 T-007: the container this slot renders into sets
   * `pointer-events: none` (see `.overlayCenter` in EditorChrome.module.css)
   * so clicks pass through to eigenpal's title bar underneath. This is a
   * CSS-only contract — `ReactNode` accepts any JSX, including interactive
   * elements that would silently lose mouse activation (keyboard activation
   * still works). If you pass an interactive element here (button, link,
   * dropdown trigger), it MUST set its own `pointer-events: auto`.
   */
  center?: ReactNode;
  /** Slot rendered top-right over eigenpal title bar (autosave, actions). */
  right?: ReactNode;
  /** Optional alert banner rendered just below the 40px title bar. */
  alert?: ReactNode;
  /** The eigenpal editor instance (DocxEditor / MetalDocsEditor). */
  children: ReactNode;
  /** Extra class on the wrapper for page-specific tweaks. */
  className?: string;
};

/**
 * EditorChrome — shared wrapper for eigenpal-based document editors.
 *
 * Provides:
 *  - `position: relative` wrapper with eigenpal styling overrides
 *    (compact title bar, wine formatting bar, gradient scrollbar)
 *    targeting selectors from @eigenpal/docx-editor-core@1.9.0.
 *    See the coupling-pin comment in EditorChrome.module.css for the
 *    upgrade checklist.
 *  - 3 absolute-positioned overlay slots (left/center/right) atop
 *    eigenpal's 40px title bar
 *  - optional alert banner rendered below the title bar
 *
 * Slot semantics:
 *  - `left`   — rendered top-left; pointer-events enabled (back button, etc.).
 *  - `center` — rendered centered; `pointer-events: none` on the container.
 *               Children that need interactivity must set `pointer-events: auto`
 *               themselves (see overlayCenter comment in the CSS).
 *  - `right`  — rendered top-right; pointer-events enabled (actions, autosave).
 *  - `alert`  — optional banner below title bar; pass `null`/`undefined` to hide.
 *  - `children` — the eigenpal editor instance (always rendered).
 *
 * @example
 * ```tsx
 * <EditorChrome
 *   center={<><CodeChip>{code}</CodeChip><span>{title}</span></>}
 *   right={<><AutosaveStatus status={autosaveState} /><button>Submit</button></>}
 * >
 *   <MetalDocsEditor ... />
 * </EditorChrome>
 * ```
 *
 * `editorChromeStyles` re-export gives consumers access to the shared button
 * primitives (.primaryBtn, .ghostBtn, .iconBtn) and text helpers (.docTitle,
 * .docMeta, .docSep) without duplicating class definitions per page.
 *
 * Used by templates/TemplateEditorPage and documents/DocumentEditorPage.
 */
export function EditorChrome({ left, center, right, alert, children, className }: EditorChromeProps) {
  return (
    <div className={`${styles.wrapper}${className ? ` ${className}` : ''}`}>
      {left && <div className={styles.overlayLeft}>{left}</div>}
      {center && <div className={styles.overlayCenter}>{center}</div>}
      {right && <div className={styles.overlayRight}>{right}</div>}
      {alert && <div className={styles.overlayAlert}>{alert}</div>}
      {children}
    </div>
  );
}

/* FE-06 T-006: named union of the class keys EditorChrome.module.css
   actually defines. `styles` itself is typed via vite/client's
   `{ readonly [key: string]: string }` CSS-module ambient declaration —
   an index signature that lets ANY property access (including a typo
   like `.primarybtn`) type-check as `string` instead of erroring. Consumers
   should prefer `editorChromeStyles[EditorChromeClass]` (or destructure
   through this union) so a typo is caught by TS instead of silently
   resolving to `undefined` at runtime. */
export type EditorChromeClass =
  | 'wrapper'
  | 'overlayLeft'
  | 'overlayCenter'
  | 'overlayRight'
  | 'overlayAlert'
  | 'docTitle'
  | 'docSep'
  | 'docMeta'
  | 'iconBtn'
  | 'ghostBtn'
  | 'primaryBtn';

/* Re-export the module CSS classes so consumers can use the
   button primitives (.iconBtn / .ghostBtn / .primaryBtn) and
   text helpers (.docTitle / .docMeta / .docSep) without
   redefining them per-page. Cast (not a fresh object) — `styles` from
   vite/client's ambient `*.module.css` typing is a readonly index
   signature; the cast narrows the *consumer-facing* type to the known
   key union above without copying the runtime object. */
export const editorChromeStyles = styles as Record<EditorChromeClass, string>;
