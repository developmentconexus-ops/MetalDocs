import type { ReactNode } from 'react';
import styles from './EditorChrome.module.css';

export type EditorChromeProps = {
  /** Slot rendered top-left over eigenpal title bar (back btn, etc.). */
  left?: ReactNode;
  /** Slot rendered centered over eigenpal title bar (title, badges, pill). */
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
 *  - position: relative wrapper with eigenpal styling overrides
 *    (compact title bar, wine formatting bar, gradient scrollbar)
 *  - 3 absolute-positioned overlay slots (left/center/right) atop
 *    eigenpal's 40px title bar
 *  - optional alert banner
 *
 * Used by templates/TemplateAuthorPage and documents/DocumentEditorPage.
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

/* Re-export the module CSS classes so consumers can use the
   button primitives (.iconBtn / .ghostBtn / .primaryBtn) and
   text helpers (.docTitle / .docMeta / .docSep) without
   redefining them per-page. */
export const editorChromeStyles = styles;
