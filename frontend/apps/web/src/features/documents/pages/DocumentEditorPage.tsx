import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { MetalDocsEditor, type MetalDocsEditorRef } from '@metaldocs/editor-ui';
import type { Comment } from '@metaldocs/editor-ui';
import { useQuery } from '@tanstack/react-query';
import { toast } from 'sonner';
import { ApiError, resolveErrorMessage, apiFetch } from '../../../lib/api';
import { QK } from '../../../lib/queryKeys';
import { useDocumentSession } from '../hooks/editor/useDocumentSession';
import { useDocumentAutosave } from '../hooks/editor/useDocumentAutosave';
import { useDocumentComments } from '../hooks/editor/useDocumentComments';
import { finalizeDocument, getApprovalInstance, renameDocument, signedRevisionURL, syncArtifactMetadata } from '../api/documents';
import type { DocumentDetail } from '../api/documents';
import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
import { useDocumentRevisionHistoryQuery } from '../queries/useDocumentRevisionHistoryQuery';
import { EditorMetaSidebar } from '../components/EditorMetaSidebar';
import {
  EditorChrome,
  editorChromeStyles,
  VersionBadge,
  AutosaveStatus,
  type AutosaveState,
} from '../../shared/components/editor-chrome';
import { CodeChip, StatusPill, type DocumentStatus } from '../../../components/ui';
import { useProfilesQuery } from '../../taxonomy/queries/useProfilesQuery';
import { useAreasQuery } from '../queries/useAreasQuery';
import { useControlledDocumentDetailQuery } from '../../registry/queries/useControlledDocumentDetailQuery';
import { buildVisibilityLabel, formatRevisionCode, hasSettledSidebarIdentity, resolveAreaLabel, resolveProfileLabel } from '../lib/documentDetailMeta';
import styles from './styles/DocumentEditorPage.module.css';

export type DocumentEditorPageProps = {
  documentID: string;
  onDone: () => void;
};

type EditorDocumentDetail = DocumentDetail & { RevisionTitle?: string | null };
type ArtifactMetadata = { fileSizeBytes?: number | null; pageCount?: number | null };

const ARTIFACT_SYNC_MAX_ATTEMPTS = 8;
const ARTIFACT_SYNC_RETRY_DELAY_MS = 400;

export function DocumentEditorPage({ documentID, onDone }: DocumentEditorPageProps): React.ReactElement {
  const docQuery = useDocumentDetailQuery(documentID);
  const [editorLoadError, setEditorLoadError] = useState<string | null>(null);
  const [documentName, setDocumentName] = useState('');
  const [buffer, setBuffer] = useState<ArrayBuffer | null | undefined>(undefined);
  const [revisionTitleDialogOpen, setRevisionTitleDialogOpen] = useState(false);
  const [revisionTitleInput, setRevisionTitleInput] = useState('');
  const [revisionTitleError, setRevisionTitleError] = useState<string | null>(null);
  const [editorDirty, setEditorDirty] = useState(false);
  const [artifactMetadata, setArtifactMetadata] = useState<ArtifactMetadata>({});
  const editorRef = useRef<MetalDocsEditorRef>(null);
  const skipInitialEditorChangeRef = useRef(false);
  const syncedArtifactRevisionRef = useRef<string | null>(null);
  const artifactSyncRetryTimeoutRef = useRef<number | null>(null);
  const doc: EditorDocumentDetail | null = (docQuery.data as EditorDocumentDetail | undefined) ?? null;
  const docStatus = doc?.Status ?? '';
  const session = useDocumentSession(documentID, { enabled: docStatus === 'draft' });
  const controlledDocumentId = doc?.ControlledDocumentID ?? '';
  const profilesQuery = useProfilesQuery();
  const areasQuery = useAreasQuery();
  const controlledDocumentQuery = useControlledDocumentDetailQuery(controlledDocumentId);
  const revisionHistoryQuery = useDocumentRevisionHistoryQuery(documentID);

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
    if (!doc) {
      setBuffer(docQuery.isLoading ? undefined : null);
      setEditorLoadError(null);
      return;
    }

    const name = doc.Name ?? 'Document';
    const revisionID = doc.CurrentRevisionID ?? '';
    let cancelled = false;

    setDocumentName(name);
    setArtifactMetadata({
      fileSizeBytes: doc.currentRevisionFileSizeBytes ?? null,
      pageCount: doc.currentRevisionPageCount ?? null,
    });
    setEditorLoadError(null);
    setBuffer(undefined);

    void (async () => {
      try {
        await fetchRevisionBuffer(revisionID);
      } catch {
        if (cancelled) return;
        setBuffer(null);
        setEditorLoadError('Falha ao carregar o arquivo do documento. Tente novamente.');
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [doc, docQuery.isLoading, fetchRevisionBuffer]);

  useEffect(() => {
    if (buffer === undefined) return;
    skipInitialEditorChangeRef.current = true;
    setEditorDirty(false);
  }, [buffer]);

  useEffect(() => {
    if (!doc) {
      syncedArtifactRevisionRef.current = null;
      return;
    }
    if (doc.CurrentRevisionID !== syncedArtifactRevisionRef.current) {
      syncedArtifactRevisionRef.current = null;
    }
  }, [doc]);

  const pageLoadError = (() => {
    const err = docQuery.error;
    if (!docQuery.isError || !err) return null;
    if (err instanceof ApiError) {
      return err.code === 'not_found' ? 'Documento não encontrado.' : resolveErrorMessage(err.code, err.message);
    }
    if (err && typeof err === 'object' && 'code' in err) {
      const code = (err as { code?: string }).code;
      const message = 'message' in err ? (err as { message?: string }).message : undefined;
      return code === 'not_found' ? 'Documento não encontrado.' : resolveErrorMessage(code, message);
    }
    if (err instanceof Error && err.message.trim()) {
      return err.message;
    }
    return 'Falha ao carregar documento.';
  })();

  // Refetch doc status when the user returns to the tab so the editor mode
  // updates without requiring a manual reload (E1).
  useEffect(() => {
    const onFocus = () => {
      void docQuery.refetch().catch(() => {});
    };
    window.addEventListener('focus', onFocus);
    return () => window.removeEventListener('focus', onFocus);
  }, [docQuery]);

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
        onArtifactMetadata: setArtifactMetadata,
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
      onArtifactMetadata: setArtifactMetadata,
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
    const pageCount = editorRef.current?.getPageCount() ?? null;
    await autosave.queue(buf, doc.FormDataJSON ?? null, pageCount);
    setEditorDirty(false);
  }

  function handleEditorChange() {
    if (skipInitialEditorChangeRef.current) {
      skipInitialEditorChangeRef.current = false;
      return;
    }
    setEditorDirty(true);
  }

  async function submitForReview(revisionTitle?: string) {
    if (session.state.phase !== 'writer' || !doc) return;
    try {
      if (editorDirty) {
        const latestBuf = await editorRef.current?.saveNow();
        if (latestBuf) {
          const pageCount = editorRef.current?.getPageCount() ?? null;
          await autosave.queue(latestBuf, doc.FormDataJSON ?? null, pageCount);
        }
        const flushOk = await autosave.flush();
        if (!flushOk) {
          toast.error('Erro ao salvar documento antes da submissão.');
          return;
        }
        setEditorDirty(false);
      }
      const trimmedRevisionTitle = revisionTitle?.trim();
      await finalizeDocument(documentID, trimmedRevisionTitle ? { revisionTitle: trimmedRevisionTitle } : {});
      setEditorDirty(false);
      await session.release();
      setRevisionTitleDialogOpen(false);
      setRevisionTitleInput('');
      setRevisionTitleError(null);
      onDone();
    } catch (err) {
      if (err instanceof ApiError) {
        toast.error(resolveErrorMessage(err.code, err.message));
      } else {
        toast.error('Erro ao finalizar documento.');
      }
    }
  }

  function handleFinalize() {
    if ((doc?.RevisionNumber ?? 0) === 0) {
      void submitForReview();
      return;
    }
    setRevisionTitleInput(doc?.RevisionTitle ?? '');
    setRevisionTitleError(null);
    setRevisionTitleDialogOpen(true);
  }

  function handleConfirmFinalize() {
    const trimmed = revisionTitleInput.trim();
    if (!trimmed) {
      setRevisionTitleError('Informe o tÃ­tulo da revisÃ£o para submeter.');
      return;
    }
    void submitForReview(trimmed);
  }

  const approvalInstanceQuery = useQuery({
    queryKey: QK.approval.instance(documentID),
    queryFn: () => getApprovalInstance(documentID),
    enabled: documentID.length > 0 && docStatus === 'under_review',
  });
  const canEditContent = session.state.phase === 'writer' && docStatus === 'draft';
  const canComment = Boolean(doc) && (docStatus === 'draft' || docStatus === 'under_review' || docStatus === 'rejected');
  const docCode = doc?.Code ?? '';
  const revisionBadge = doc?.RevisionNumber != null ? formatRevisionCode(doc.RevisionNumber) : null;
  const displayName = documentName.replace(/\.docx$/i, '');
  const userID = doc?.CreatedBy ?? '';
  const authorDisplay = String(userID);
  const commentsHook = useDocumentComments(documentID, authorDisplay);
  const canUseComments = canComment && !commentsHook.loadError;
  const canMountEditor = !!doc
    && (docStatus === 'draft'
      ? session.state.phase !== 'idle' && session.state.phase !== 'acquiring'
      : true)
    && buffer !== undefined;

  useEffect(() => {
    if (!doc || docStatus !== 'draft' || sessionPhase !== 'writer') {
      return;
    }
    if (artifactMetadata.fileSizeBytes != null && artifactMetadata.pageCount != null) {
      return;
    }
    if (!doc.CurrentRevisionID || syncedArtifactRevisionRef.current === doc.CurrentRevisionID) {
      return;
    }

    let cancelled = false;

    const clearRetry = () => {
      if (artifactSyncRetryTimeoutRef.current != null) {
        window.clearTimeout(artifactSyncRetryTimeoutRef.current);
        artifactSyncRetryTimeoutRef.current = null;
      }
    };

    const attemptSync = (attempt: number) => {
      const pageCount = editorRef.current?.getPageCount() ?? null;
      syncedArtifactRevisionRef.current = doc.CurrentRevisionID;

      void syncArtifactMetadata(documentID, {
        session_id: sessionID,
        page_count: pageCount && pageCount > 0 ? pageCount : null,
      }).then((metadata) => {
        if (cancelled) {
          return;
        }

        setArtifactMetadata({
          fileSizeBytes: metadata.file_size_bytes ?? null,
          pageCount: metadata.page_count ?? null,
        });

        if ((metadata.page_count ?? null) == null && attempt < ARTIFACT_SYNC_MAX_ATTEMPTS) {
          clearRetry();
          artifactSyncRetryTimeoutRef.current = window.setTimeout(() => {
            attemptSync(attempt + 1);
          }, ARTIFACT_SYNC_RETRY_DELAY_MS);
        }
      }).catch(() => {
        if (cancelled) {
          return;
        }
        syncedArtifactRevisionRef.current = null;
        if (attempt < ARTIFACT_SYNC_MAX_ATTEMPTS) {
          clearRetry();
          artifactSyncRetryTimeoutRef.current = window.setTimeout(() => {
            attemptSync(attempt + 1);
          }, ARTIFACT_SYNC_RETRY_DELAY_MS);
        }
      });
    };

    attemptSync(0);

    return () => {
      cancelled = true;
      clearRetry();
    };
  }, [artifactMetadata.fileSizeBytes, artifactMetadata.pageCount, doc, docStatus, documentID, sessionID, sessionPhase]);

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

  const profileLabel = doc?.ProfileCodeSnapshot && profilesQuery.data
    ? resolveProfileLabel(doc.ProfileCodeSnapshot, profilesQuery.data)
    : null;
  const areaLabel = doc?.ProcessAreaCodeSnapshot && areasQuery.data
    ? resolveAreaLabel(doc.ProcessAreaCodeSnapshot, areasQuery.data)
    : null;
  const visibilityLabel = areasQuery.data && controlledDocumentQuery.data?.visibility
    ? buildVisibilityLabel(controlledDocumentQuery.data.visibility, areasQuery.data)
    : null;
  const sidebarIdentityReady = hasSettledSidebarIdentity({
    code: docCode || null,
    profileLabel,
    areaLabel,
    visibilityLabel,
  });
  const sidebarHistory = (revisionHistoryQuery.data?.items ?? []).map((item) => ({
    documentId: item.documentId,
    revisionCode: formatRevisionCode(item.revisionNumber),
    revisionTitle: item.revisionTitle,
    status: item.status,
    createdAt: item.createdAt,
    isCurrent: item.isCurrent,
  }));

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
                {revisionBadge && <VersionBadge>{revisionBadge}</VersionBadge>}
                {statusForPill && <StatusPill status={statusForPill} />}
              </>
            }
            right={
              <>
                <AutosaveStatus status={autosaveState} />
                  <button
                    type="button"
                    className={editorChromeStyles.primaryBtn}
                    onClick={handleFinalize}
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
                onChange={handleEditorChange}
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
          loading={!sidebarIdentityReady}
          profileLabel={profileLabel ?? undefined}
          areaLabel={areaLabel ?? undefined}
          visibilityLabel={visibilityLabel ?? undefined}
          fileSizeBytes={artifactMetadata.fileSizeBytes ?? null}
          pageCount={artifactMetadata.pageCount ?? null}
          history={sidebarHistory}
          approvalChain={docStatus === 'under_review' ? approvalInstanceQuery.data ?? null : null}
          documentStatus={docStatus}
          />
        ) : null}
        {revisionTitleDialogOpen ? (
          <div className={styles.dialogBackdrop} role="presentation">
            <div className={styles.dialogPanel} role="dialog" aria-modal="true" aria-labelledby="revision-title-dialog-title">
              <h2 id="revision-title-dialog-title" className={styles.dialogTitle}>TÃ­tulo da revisÃ£o</h2>
              <p className={styles.dialogText}>Informe o motivo governado desta revisÃ£o antes de submeter o documento para aprovaÃ§Ã£o.</p>
              <label className={styles.dialogField}>
                <span>TÃ­tulo</span>
                <input
                  type="text"
                  value={revisionTitleInput}
                  onChange={(event) => {
                    setRevisionTitleInput(event.target.value);
                    if (revisionTitleError) setRevisionTitleError(null);
                  }}
                  placeholder="Ex.: Ajuste de procedimento operacional"
                />
              </label>
              {revisionTitleError ? <p role="alert" className={styles.dialogError}>{revisionTitleError}</p> : null}
              <div className={styles.dialogActions}>
                <button type="button" className={styles.dialogSecondaryBtn} onClick={() => setRevisionTitleDialogOpen(false)}>
                  Cancelar
                </button>
                <button type="button" className={styles.dialogPrimaryBtn} onClick={handleConfirmFinalize}>
                  Confirmar submissÃ£o
                </button>
              </div>
            </div>
          </div>
        ) : null}
      </div>
    </div>
  );
}

