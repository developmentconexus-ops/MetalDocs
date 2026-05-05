import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { MetalDocsEditor, type MetalDocsEditorRef } from '@metaldocs/editor-ui';
import type { Comment } from '@metaldocs/editor-ui';
import { toast } from 'sonner';
import { ApiError, resolveErrorMessage, apiFetch } from '../../../lib/api';
import { useDocumentSession } from './hooks/useDocumentSession';
import { useDocumentAutosave } from './hooks/useDocumentAutosave';
import { useDocumentComments } from './hooks/useDocumentComments';
import { getDocument, finalizeDocument, renameDocument, signedRevisionURL } from './api/documentsV2';
import { useDocumentPdfStatus } from './hooks/useDocumentPdfStatus';
import { PDFCell } from './PDFCell';
import type { DocumentResponse } from './api/documentsV2';
import { CheckpointsDialog } from './CheckpointsDialog';
import { ExportMenuButton } from './ExportMenuButton';
import styles from './styles/DocumentEditorPage.module.css';

export type DocumentEditorPageProps = {
  documentID: string;
  onDone: () => void;
};

export function DocumentEditorPage({ documentID, onDone }: DocumentEditorPageProps): React.ReactElement {
  const session = useDocumentSession(documentID);
  const [doc, setDoc] = useState<DocumentResponse | null>(null);
  const [documentName, setDocumentName] = useState('');
  const [buffer, setBuffer] = useState<ArrayBuffer | null | undefined>(undefined);
  const [checkpointsOpen, setCheckpointsOpen] = useState(false);
  const editorRef = useRef<MetalDocsEditorRef>(null);

  const fetchRevisionBuffer = useCallback(async (revisionID: string) => {
    if (!revisionID) {
      setBuffer(null);
      return;
    }
    const signedPayload = await apiFetch<{ url?: string }>(signedRevisionURL(documentID, revisionID));
    if (!signedPayload.url) {
      throw new Error('missing_signed_url');
    }
    const fileRes = await fetch(signedPayload.url);
    if (!fileRes.ok) throw Object.assign(new Error(`http_${fileRes.status}`), { status: fileRes.status });
    setBuffer(await fileRes.arrayBuffer());
  }, [documentID]);

  useEffect(() => {
    void (async () => {
      try {
        setBuffer(undefined);
        const loadedDoc = await getDocument(documentID);
        const name = loadedDoc.Name ?? loadedDoc.name ?? 'Document';
        const revisionID = loadedDoc.CurrentRevisionID ?? loadedDoc.current_revision_id ?? '';
        setDoc(loadedDoc);
        setDocumentName(name);
        await fetchRevisionBuffer(revisionID);
      } catch {
        toast.error('Failed to load document.');
      }
    })();
  }, [documentID, fetchRevisionBuffer]);

  // Refetch doc status when the user returns to the tab so the editor mode
  // updates without requiring a manual reload (E1).
  useEffect(() => {
    const onFocus = () => {
      void getDocument(documentID).then((d) => setDoc(d)).catch(() => {});
    };
    window.addEventListener('focus', onFocus);
    return () => window.removeEventListener('focus', onFocus);
  }, [documentID]);

  const sessionPhase = session.state.phase;
  const sessionID = sessionPhase === 'writer' ? session.state.sessionID : '';
  const lastAckRevisionID = sessionPhase === 'writer' ? session.state.lastAckRevisionID : '';
  const { setLastAck } = session;

  const autosaveArgs = useMemo(() => {
    if (sessionPhase === 'writer') {
      return {
        documentID,
        sessionID,
        baseRevisionID: lastAckRevisionID,
        onAdvanceBase: (newRevisionID: string) => {
          setLastAck(newRevisionID);
        },
        onSessionLost: () => {
          toast.error('Writer session lost.');
        },
      };
    }
    return {
      documentID,
      sessionID: '',
      baseRevisionID: '',
      onAdvanceBase: () => {},
      onSessionLost: () => {},
    };
  }, [documentID, sessionPhase, sessionID, lastAckRevisionID, setLastAck]);

  const autosave = useDocumentAutosave(autosaveArgs);

  const prevAutosaveStatus = useRef(autosave.status);
  useEffect(() => {
    if (autosave.status === prevAutosaveStatus.current) {
      return;
    }
    prevAutosaveStatus.current = autosave.status;
    if (autosave.status === 'error' || autosave.status === 'session_lost' || autosave.status === 'stale') {
      toast.error(`Autosave ${autosave.status.replace('_', ' ')}.`);
    }
  }, [autosave.status]);

  const prevSessionPhase = useRef(sessionPhase);
  useEffect(() => {
    if (sessionPhase !== prevSessionPhase.current && (sessionPhase === 'readonly' || sessionPhase === 'lost')) {
      toast.warning(
        sessionPhase === 'readonly'
          ? 'Readonly session. Another user is editing this document.'
          : 'Session lost. Reload to reacquire writer access.',
      );
    }
    prevSessionPhase.current = sessionPhase;
  }, [sessionPhase]);

  const handleRename = useCallback((name: string) => {
    const prev = documentName;
    setDocumentName(name);
    void renameDocument(documentID, name).catch((err: unknown) => {
      setDocumentName(prev);
      const code = (err && typeof err === 'object' && 'code' in err)
        ? (err as { code?: string }).code
        : undefined;
      toast.error(resolveErrorMessage(code, 'Falha ao renomear documento.'));
    });
  }, [documentID, documentName]);

  async function handleSave() {
    if (!editorRef.current) return;
    if (!doc) return;
    const buf = await editorRef.current.getDocumentBuffer();
    if (!buf) return;
    await autosave.queue(buf, doc.FormDataJSON ?? doc.form_data ?? null);
  }

  async function handleFinalize() {
    if (session.state.phase !== 'writer' || !doc) return;
    try {
      await autosave.flush();
      await finalizeDocument(documentID);
      await session.release();
      onDone();
    } catch (err) {
      if (err instanceof ApiError) {
        toast.error(resolveErrorMessage(err.code, err.message));
      } else {
        toast.error('Erro ao finalizar documento.');
      }
    }
  }

  async function handleRestored(newRevisionID: string) {
    try {
      await fetchRevisionBuffer(newRevisionID);
      const refreshedDoc = await getDocument(documentID);
      setDoc(refreshedDoc);
      setDocumentName(refreshedDoc.Name ?? refreshedDoc.name ?? 'Document');
      session.setLastAck(newRevisionID);
    } catch {
      toast.error('Failed to refresh document after restore.');
    }
  }

  const docStatus = doc?.Status ?? doc?.status ?? '';
  const isEditable = session.state.phase === 'writer' && docStatus === 'draft';
  // Poll view endpoint for PDF status when doc is not a draft (E11).
  const pdf = useDocumentPdfStatus(documentID, docStatus !== '' && docStatus !== 'draft');
  const docCode = doc?.Code ?? doc?.code ?? '';
  const revNum = doc?.RevisionVersion ?? doc?.revision_version ?? 0;
  const displayName = documentName.replace(/\.docx$/i, '');
  const statusPillClass = {
    draft: styles.draft,
    under_review: styles.inReview,
    approved: styles.approved,
    published: styles.published,
  }[docStatus] ?? '';
  const userID = doc?.CreatedBy ?? doc?.created_by ?? '';
  const authorDisplay = String(userID);
  const commentsHook = useDocumentComments(documentID, authorDisplay);
  const canMountEditor = !!doc
    && session.state.phase !== 'idle'
    && session.state.phase !== 'acquiring'
    && buffer !== undefined;

  return (
    <div className={styles.page} data-editor-root>
      <div className={styles.body}>

        <main className={styles.canvas}>
          <div className={styles.editorWrapper}>

            <div className={styles.overlayBack}>
              <button className={styles.overlayBackBtn} onClick={onDone} aria-label="Voltar">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#fff" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="15 18 9 12 15 6" />
                </svg>
              </button>
            </div>

            <div className={styles.overlayTitle}>
              <span className={styles.docMeta}>Documento</span>
              <span className={styles.docSep}>·</span>
              <span className={styles.docTitle}>{docCode ? `${docCode}-${displayName}` : (displayName || 'Documento')}</span>
              {doc && <span className={styles.versionBadge}>REV{String(revNum).padStart(2, '0')}</span>}
              {docStatus && (
                <span className={`${styles.statusPill} ${statusPillClass}`}>
                  {docStatus.replace('_', ' ')}
                </span>
              )}
            </div>

            <div className={styles.overlayRight}>
              {autosave.status === 'saving' ? (
                <span className={styles.autosaveStatus}>
                  <span className={styles.autosaveDot} aria-hidden="true" />
                  Saving…
                </span>
              ) : autosave.status === 'error' ? (
                <span className={styles.autosaveStatus} style={{ color: '#dc2626' }}>Save failed</span>
              ) : autosave.status === 'saved' ? (
                <span className={styles.autosaveStatus}>✓ Saved</span>
              ) : (
                <span className={styles.autosaveStatus} />
              )}
              <button
                type="button"
                className={styles.editorSubmitBtn}
                onClick={() => setCheckpointsOpen(true)}
              >
                Checkpoints
              </button>
              <ExportMenuButton
                documentID={documentID}
                canExport={sessionPhase === 'writer' || sessionPhase === 'readonly'}
              />
              {docStatus !== 'draft' && docStatus !== '' && (
                <PDFCell status={pdf.status} url={pdf.url} onRetry={pdf.retry} />
              )}
              <button
                type="button"
                className={styles.editorSubmitBtn}
                onClick={() => void handleFinalize()}
                disabled={!isEditable}
              >
                Finalizar
              </button>
            </div>

            {canMountEditor ? (
              <MetalDocsEditor
                ref={editorRef}
                mode={isEditable ? 'document-edit' : 'readonly'}
                documentBuffer={buffer ?? undefined}
                author={authorDisplay}
                comments={commentsHook.comments}
                onCommentsChange={commentsHook.setComments}
                onCommentAdd={(c: Comment) => { if (isEditable) void commentsHook.add(c); }}
                onCommentResolve={(c: Comment) => { if (isEditable) void (c.done ? commentsHook.resolve(c) : commentsHook.reopen(c)); }}
                onCommentDelete={(c: Comment) => { if (isEditable) void commentsHook.remove(c); }}
                onCommentReply={(reply: Comment, parent: Comment) => { if (isEditable) void commentsHook.reply(reply, parent); }}
                onAutoSave={handleSave}
                onDocumentNameChange={handleRename}
                showRuler={false}
              />
            ) : null}

          </div>
        </main>

      </div>

      <CheckpointsDialog
        open={checkpointsOpen}
        onClose={() => setCheckpointsOpen(false)}
        documentID={documentID}
        disabled={!isEditable}
        onRestored={(rev) => {
          setCheckpointsOpen(false);
          void handleRestored(rev);
        }}
      />
    </div>
  );
}
