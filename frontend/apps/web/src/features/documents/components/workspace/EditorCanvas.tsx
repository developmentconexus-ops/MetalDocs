import { DocumentShell, type DocumentShellProps } from '../DocumentShell';

export interface EditorCanvasProps extends DocumentShellProps {
  /**
   * Mirrors DocumentEditorPage's session-acquisition guard: false while a
   * writer session is still being acquired (or hasn't been requested yet) on
   * a draft document — the editor must not mount mid-acquisition. Read-only
   * canvases (approving/reviewing/observing) never gate on a session, so this
   * defaults to `true`.
   */
  canMountEditor?: boolean;
}

/**
 * EditorCanvas — F2d.5 S2b.
 *
 * The single canonical writable-canvas mount, extracted from
 * DocumentEditorPage (F3's `<main className={styles.canvas}>` region) so
 * DocumentWorkspacePage's `author-editing` / `author-changes-requested`
 * modes reuse the SAME DocumentShell wiring (editorMode + autosave + comment
 * handlers + chrome) instead of forking it. Pure passthrough to
 * `DocumentShell` behind a `canMountEditor` gate — no session/autosave/
 * comments state lives here; that stays owned by the caller (route/page),
 * matching DocumentShell's own "presentational, caller owns persistence"
 * contract.
 */
export function EditorCanvas({ canMountEditor = true, ...shellProps }: EditorCanvasProps): React.ReactElement | null {
  if (!canMountEditor) {
    return null;
  }
  return <DocumentShell {...shellProps} />;
}
