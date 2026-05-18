import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { MetalDocsEditor, type MetalDocsEditorRef } from '@metaldocs/editor-ui';
import type { Comment } from '@metaldocs/editor-ui';
import { toast } from 'sonner';
import { ApiError, resolveErrorMessage, apiFetch } from '../../../lib/api';
import { useDocumentSession } from '../hooks/editor/useDocumentSession';
import { useDocumentAutosave } from '../hooks/editor/useDocumentAutosave';
import { useDocumentComments } from '../hooks/editor/useDocumentComments';
import { getDocument, finalizeDocument, renameDocument, signedRevisionURL } from '../api/documents';
import { useDocumentPdfStatus } from '../hooks/editor/useDocumentPdfStatus';
import { PDFCell } from '../components/PDFCell';
import type { DocumentResponse } from '../api/documents';
import { EditorMetaSidebar } from '../components/EditorMetaSidebar';
import {
  EditorChrome,
  editorChromeStyles,
  VersionBadge,
  AutosaveStatus,
  type AutosaveState,
} from '../../shared/components/editor-chrome';
import { CodeChip, StatusPill, type DocumentStatus } from '../../../components/ui';
import styles from './styles/DocumentEditorPage.module.css';

export type DocumentEditorPageProps = {
  documentID: string;
  onDone: () => void;
};

export function DocumentEditorPage({ documentID, onDone }: DocumentEditorPageProps): React.ReactElement {
  const session = useDocumentSession(documentID);
  const [doc, setDoc] = useState<DocumentResponse | null>(null);
  const [pageLoadError, setPageLoadError] = useState<string | null>(null);
  const [editorLoadError, setEditorLoadError] = useState<string | null>(null);
  const [documentName, setDocumentName] = useState('');
  const [buffer, setBuffer] = useState<ArrayBuffer | null | undefined>(undefined);
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
        setPageLoadError(null);
        setEditorLoadError(null);
        setBuffer(undefined);
        const loadedDoc = await getDocument(documentID);
        const name = loadedDoc.Name ?? loadedDoc.name ?? 'Document';
        const revisionID = loadedDoc.CurrentRevisionID ?? loadedDoc.current_revision_id ?? '';
        setDoc(loadedDoc);
        setDocumentName(name);
        try {
          await fetchRevisionBuffer(revisionID);
        } catch {
          setBuffer(null);
          setEditorLoadError('Falha ao carregar o arquivo do documento. Tente novamente.');
        }
      } catch (err) {
        setDoc(null);
        setBuffer(null);
        if (err instanceof ApiError) {
          setPageLoadError(err.code === 'not_found' ? 'Documento não encontrado.' : resolveErrorMessage(err.code, err.message));
        } else if (err && typeof err === 'object' && 'code' in err) {
          const code = (err as { code?: string }).code;
          const message = 'message' in err ? (err as { message?: string }).message : undefined;
          setPageLoadError(code === 'not_found' ? 'Documento não encontrado.' : resolveErrorMessage(code, message));
        } else if (err instanceof Error && err.message.trim()) {
          setPageLoadError(err.message);
        } else {
          setPageLoadError('Falha ao carregar documento.');
        }
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

  async function handleSave(buf: ArrayBuffer) {
    if (!doc) return;
    await autosave.queue(buf, doc.FormDataJSON ?? doc.form_data ?? null);
  }

  async function handleFinalize() {
    if (session.state.phase !== 'writer' || !doc) return;
    try {
      const latestBuf = await editorRef.current?.saveNow();
      if (latestBuf) {
        await autosave.queue(latestBuf, doc.FormDataJSON ?? doc.form_data ?? null);
      }
      const flushOk = await autosave.flush();
      if (!flushOk) {
        toast.error('Erro ao salvar documento antes da submissão.');
        return;
      }
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

  const docStatus = doc?.Status ?? doc?.status ?? '';
  const canEditContent = session.state.phase === 'writer' && docStatus === 'draft';
  const canComment = Boolean(doc) && (docStatus === 'draft' || docStatus === 'under_review' || docStatus === 'rejected');
  // Poll view endpoint for PDF status when doc is not a draft (E11).
  const pdf = useDocumentPdfStatus(documentID, docStatus !== '' && docStatus !== 'draft');
  const docCode = doc?.Code ?? doc?.code ?? '';
  const revNum = doc?.RevisionVersion ?? doc?.revision_version ?? 0;
  const displayName = documentName.replace(/\.docx$/i, '');
  const userID = doc?.CreatedBy ?? doc?.created_by ?? '';
  const authorDisplay = String(userID);
  const commentsHook = useDocumentComments(documentID, authorDisplay);
  const canUseComments = canComment && !commentsHook.loadError;
  const canMountEditor = !!doc
    && session.state.phase !== 'idle'
    && session.state.phase !== 'acquiring'
    && buffer !== undefined;

  const [sidebarOpen, setSidebarOpen] = useState<boolean>(
    () => localStorage.getItem('editor-sidebar-open') !== 'false'
  );

  // AutosaveState (7-state) is structurally identical to useDocumentAutosave.AutosaveStatus.
  const autosaveState: AutosaveState = autosave.status;

  const statusForPill: DocumentStatus | null = (() => {
    if (!docStatus) return null;
    const allowed: DocumentStatus[] = [
      'draft', 'review', 'under_review', 'approved', 'frozen',
      'rejected', 'archived', 'finalized', 'scheduled', 'published',
      'superseded', 'obsolete',
    ];
    return (allowed as string[]).includes(docStatus) ? (docStatus as DocumentStatus) : null;
  })();

  return (
    <div className={styles.page} data-editor-root>
      <div className={styles.body}>
        <aside className={styles.rail}>
          <button
            type="button"
            className={styles.railBackBtn}
            onClick={onDone}
            aria-label="Voltar"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="15 18 9 12 15 6" />
            </svg>
            <span className={styles.railTip}>Voltar</span>
          </button>
        </aside>
        <main className={styles.canvas}>
          {pageLoadError ? (
            <div role="alert" className={styles.error}>
              {pageLoadError}
            </div>
          ) : (
            <EditorChrome
            center={
              <>
                {docCode && <CodeChip>{docCode}</CodeChip>}
                {displayName && <span className={editorChromeStyles.docTitle}>{displayName}</span>}
                {revNum > 0 && <VersionBadge>{`v${revNum}`}</VersionBadge>}
                {statusForPill && <StatusPill status={statusForPill} />}
              </>
            }
            right={
              <>
                <AutosaveStatus status={autosaveState} />
                <button
                  type="button"
                  className={editorChromeStyles.primaryBtn}
                  onClick={() => void handleFinalize()}
                  disabled={!canEditContent}
                >
                  Submeter para revisão
                </button>
              </>
            }
          >
            {commentsHook.loadError ? (
              <div className={styles.errorBanner} role="alert">
                <span>{commentsHook.loadError}</span>
                <button
                  type="button"
                  className={styles.retryButton}
                  onClick={() => void commentsHook.retry()}
                >
                  Tentar novamente
                </button>
              </div>
            ) : null}
            {editorLoadError ? (
              <div role="alert" className={styles.error}>
                {editorLoadError}
              </div>
            ) : canMountEditor ? (
              <MetalDocsEditor
                ref={editorRef}
                mode={canEditContent ? 'document-edit' : 'readonly'}
                documentBuffer={buffer ?? undefined}
                author={authorDisplay}
                comments={commentsHook.comments}
                onCommentsChange={commentsHook.setComments}
                onCommentAdd={(c: Comment) => { if (canUseComments) void commentsHook.add(c); }}
                onCommentResolve={(c: Comment) => { if (canUseComments) void (c.done ? commentsHook.resolve(c) : commentsHook.reopen(c)); }}
                onCommentDelete={(c: Comment) => { if (canUseComments) void commentsHook.remove(c); }}
                onCommentReply={(reply: Comment, parent: Comment) => { if (canUseComments) void commentsHook.reply(reply, parent); }}
                onAutoSave={handleSave}
                onDocumentNameChange={handleRename}
                showRuler={false}
              />
            ) : null}
          </EditorChrome>
          )}
        </main>
        {!pageLoadError ? (
          <EditorMetaSidebar
          open={sidebarOpen}
          onToggle={() => {
            setSidebarOpen((prev) => {
              const next = !prev;
              localStorage.setItem('editor-sidebar-open', String(next));
              return next;
            });
          }}
          code={docCode || undefined}
          />
        ) : null}
      </div>
    </div>
  );
}

