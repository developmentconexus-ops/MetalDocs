import { forwardRef, useCallback, useEffect, useImperativeHandle, useRef, useState } from 'react';
import { MetalDocsEditor, type MetalDocsEditorRef, type Comment } from '@metaldocs/editor-ui';
import { apiFetch } from '../../../lib/api';
import { signedRevisionURL } from '../../documents/api/documents';
import { useDocumentSession } from '../../documents/hooks/editor/useDocumentSession';
import { useDocumentAutosave } from '../../documents/hooks/editor/useDocumentAutosave';
import { useDocumentComments } from '../../documents/hooks/editor/useDocumentComments';

export interface ReviewDocumentCanvasRef {
  flushSave(): Promise<void>;
}

export interface ReviewDocumentCanvasProps {
  documentId: string;
  currentRevisionId: string;
  status: string;
  approverDisplay: string;
}

export const ReviewDocumentCanvas = forwardRef<ReviewDocumentCanvasRef, ReviewDocumentCanvasProps>(
  function ReviewDocumentCanvas(
    { documentId, currentRevisionId, status, approverDisplay },
    ref,
  ) {
    const editorRef = useRef<MetalDocsEditorRef>(null);

    // Buffer state: undefined = loading, null = error/not-found, ArrayBuffer = ready
    const [buffer, setBuffer] = useState<ArrayBuffer | null | undefined>(undefined);

    // ---------------------------------------------------------------------------
    // Session — enabled only while document is under_review so this component
    // can hold a writer session (reviewers may annotate but not edit content).
    // ---------------------------------------------------------------------------
    const session = useDocumentSession(documentId, { enabled: status === 'under_review' });

    const sessionPhase = session.state.phase;
    const sessionID = sessionPhase === 'writer' ? session.state.sessionID : '';
    const baseRevisionID = sessionPhase === 'writer' ? session.state.lastAckRevisionID : '';
    const { setLastAck } = session;

    // ---------------------------------------------------------------------------
    // Autosave
    // ---------------------------------------------------------------------------
    const autosave = useDocumentAutosave({
      documentID: documentId,
      sessionID,
      baseRevisionID,
      onAdvanceBase: setLastAck,
      onSessionLost: () => {
        // Silently absorb session-lost in review mode; the cockpit layer handles
        // any necessary UX escalation.
      },
    });

    // ---------------------------------------------------------------------------
    // Buffer fetch — load the working revision from its signed URL
    // ---------------------------------------------------------------------------
    useEffect(() => {
      if (!currentRevisionId) {
        setBuffer(null);
        return;
      }

      let cancelled = false;
      setBuffer(undefined);

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
          if (!cancelled) setBuffer(null);
        }
      })();

      return () => {
        cancelled = true;
      };
    }, [documentId, currentRevisionId]);

    // ---------------------------------------------------------------------------
    // Comments
    // ---------------------------------------------------------------------------
    const commentsHook = useDocumentComments(documentId, approverDisplay);

    // ---------------------------------------------------------------------------
    // Imperative handle — flushSave() for the cockpit save-on-decision path
    // ---------------------------------------------------------------------------
    useImperativeHandle(ref, () => ({
      async flushSave(): Promise<void> {
        const latestBuf = await editorRef.current?.saveNow();
        if (latestBuf) {
          const pageCount = editorRef.current?.getPageCount() ?? null;
          await autosave.queue(latestBuf, null, pageCount);
        }
        await autosave.flush();
      },
    }), [autosave]);

    // ---------------------------------------------------------------------------
    // Autosave handler wired to MetalDocsEditor.onAutoSave
    // ---------------------------------------------------------------------------
    const handleAutoSave = useCallback(async (buf: ArrayBuffer) => {
      const pageCount = editorRef.current?.getPageCount() ?? null;
      await autosave.queue(buf, null, pageCount);
      await autosave.flush();
    }, [autosave]);

    // ---------------------------------------------------------------------------
    // Render
    // ---------------------------------------------------------------------------
    const canMount = buffer !== undefined && buffer !== null;

    if (!canMount) {
      // While loading or on error, render nothing (the cockpit provides shell UI)
      return null;
    }

    return (
      <MetalDocsEditor
        ref={editorRef}
        mode="review"
        documentBuffer={buffer}
        author={approverDisplay}
        comments={commentsHook.comments}
        onCommentsChange={commentsHook.setComments}
        onCommentAdd={(c: Comment) => void commentsHook.add(c)}
        onCommentResolve={(c: Comment) => void (c.done ? commentsHook.resolve(c) : commentsHook.reopen(c))}
        onCommentDelete={(c: Comment) => void commentsHook.remove(c)}
        onCommentReply={(reply: Comment, parent: Comment) => void commentsHook.reply(reply, parent)}
        onAutoSave={handleAutoSave}
        showRuler={false}
      />
    );
  },
);
