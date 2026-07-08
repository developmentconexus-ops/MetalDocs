import { useEffect, useState } from 'react';
import { MetalDocsEditor, type MetalDocsEditorRef, type EditorComment, type TrackedChange } from '@metaldocs/editor-ui';
import { apiFetch } from '../../../lib/api';
import { signedRevisionURL } from '../api/documents';
import { EditorChrome } from '../../shared/components/editor-chrome';
import styles from './DocumentShell.module.css';

export interface DocumentShellProps {
  documentId: string;
  /** Shell fetches the buffer from the signed URL for this revision. */
  currentRevisionId: string | null;
  editorMode: 'document-edit' | 'readonly' | 'review';
  editorRef?: React.Ref<MetalDocsEditorRef>;
  author: string;
  comments?: EditorComment[];
  onCommentsChange?: (c: EditorComment[]) => void;
  onCommentAdd?: (c: EditorComment) => void;
  onCommentResolve?: (c: EditorComment) => void;
  onCommentDelete?: (c: EditorComment) => void;
  onCommentReply?: (reply: EditorComment, parent: EditorComment) => void;
  /** Absent => the editor never persists a document revision. */
  onAutoSave?: (buf: ArrayBuffer) => void | Promise<void>;
  onChange?: () => void;
  /**
   * Additive beyond the F3 spec's literal prop list: DocumentEditorPage relies
   * on eigenpal's inline document-name edit affordance to drive its rename
   * flow (`handleRename`). Not consumed by the cockpit (F3/F4 have no rename
   * affordance in-editor); optional and inert when omitted.
   */
  onDocumentNameChange?: (name: string) => void;
  /** F1 surface; F4 consumes. */
  onTrackedChangesChange?: (c: TrackedChange[]) => void;
  /** Optional EditorChrome header slots. Absent => bare editor (caller owns the header). */
  chrome?: { center?: React.ReactNode; right?: React.ReactNode };
  /** Optional banner rendered above the editor. */
  notice?: React.ReactNode;
}

/**
 * DocumentShell — the shared editor-canvas region (F3, C1).
 *
 * Presentational + buffer-owning; owns the buffer state machine (`undefined`
 * loading / `null` error / `ArrayBuffer` ready) fetched from the signed
 * revision URL, the loading/error render, and the `MetalDocsEditor` mount.
 *
 * Never imports `useDocumentSession` or `useDocumentAutosave` — persistence
 * is the caller's concern, injected via `onAutoSave`. Absent `onAutoSave`
 * means the editor mounts but never persists (the W2 fix for the approval
 * review canvas: no writable session, no autosave).
 */
export function DocumentShell({
  documentId,
  currentRevisionId,
  editorMode,
  editorRef,
  author,
  comments,
  onCommentsChange,
  onCommentAdd,
  onCommentResolve,
  onCommentDelete,
  onCommentReply,
  onAutoSave,
  onChange,
  onDocumentNameChange,
  onTrackedChangesChange,
  chrome,
  notice,
}: DocumentShellProps): React.ReactElement {
  const [buffer, setBuffer] = useState<ArrayBuffer | null | undefined>(undefined);
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    if (!currentRevisionId) {
      setBuffer(null);
      setLoadError(null);
      return;
    }

    let cancelled = false;
    setBuffer(undefined);
    setLoadError(null);

    void (async () => {
      try {
        const signedPayload = await apiFetch<{ url?: string }>(
          signedRevisionURL(documentId, currentRevisionId),
        );
        if (!signedPayload.url) throw new Error('missing_signed_url');
        const fileRes = await fetch(signedPayload.url);
        if (!fileRes.ok) {
          throw Object.assign(new Error(`http_${fileRes.status}`), { status: fileRes.status });
        }
        const ab = await fileRes.arrayBuffer();
        if (!cancelled) setBuffer(ab);
      } catch {
        if (!cancelled) {
          setBuffer(null);
          setLoadError('Falha ao carregar o arquivo do documento. Tente novamente.');
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [documentId, currentRevisionId]);

  let body: React.ReactElement | null;
  if (loadError || buffer === null) {
    body = (
      <div role="alert" className={styles.error}>
        {loadError ?? 'Documento indisponível.'}
      </div>
    );
  } else if (buffer === undefined) {
    body = (
      <div role="status" aria-live="polite" className={styles.loading}>
        Carregando documento…
      </div>
    );
  } else {
    body = (
      <MetalDocsEditor
        ref={editorRef}
        mode={editorMode}
        documentBuffer={buffer}
        author={author}
        comments={comments}
        onCommentsChange={onCommentsChange}
        onCommentAdd={onCommentAdd}
        onCommentResolve={onCommentResolve}
        onCommentDelete={onCommentDelete}
        onCommentReply={onCommentReply}
        onAutoSave={onAutoSave ? async (buf: ArrayBuffer) => { await onAutoSave(buf); } : undefined}
        onChange={onChange}
        onDocumentNameChange={onDocumentNameChange}
        onTrackedChangesChange={onTrackedChangesChange}
        showRuler={false}
      />
    );
  }

  if (chrome) {
    return (
      <EditorChrome center={chrome.center} right={chrome.right} alert={notice}>
        {body}
      </EditorChrome>
    );
  }

  return (
    <>
      {notice}
      {body}
    </>
  );
}
