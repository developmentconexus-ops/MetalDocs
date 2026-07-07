import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { MetalDocsEditorRef } from '@metaldocs/editor-ui';
import type { EditorComment, TrackedChange } from '@metaldocs/editor-ui';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { ApiError, resolveErrorMessage } from '../../../lib/api';
import { QK } from '../../../lib/queryKeys';
import { useDocumentSession } from '../hooks/editor/useDocumentSession';
import { useDocumentAutosave } from '../hooks/editor/useDocumentAutosave';
import { useDocumentComments } from '../hooks/editor/useDocumentComments';
import {
  getApprovalInstance,
  getDocumentRevisionHistory,
  renameDocument,
} from '../api/documents';
import type { DocumentDetail } from '../api/documents';
import { submit as submitForReviewRequest } from '../../approval/api/approvalApi';
import type { components } from '../../../lib/api-types';
import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
import { useDocumentRevisionHistoryQuery } from '../queries/useDocumentRevisionHistoryQuery';
import { ArtifactMetaSidebar } from '../../shared/controlled-artifact/ArtifactMetaSidebar';
import type { VersionHistoryItem } from '../../shared/controlled-artifact/types';
import { actorStatusToFlowState } from '../lib/approvalWorkflow';
import {
  editorChromeStyles,
  VersionBadge,
  AutosaveStatus,
  type AutosaveState,
} from '../../shared/components/editor-chrome';
import { DocumentShell } from '../components/DocumentShell';
import { RequestedChangesPanel } from '../components/RequestedChangesPanel';
import { CodeChip, StatusPill } from '../../../components/ui';
import { useProfilesQuery } from '../../taxonomy/queries/useProfilesQuery';
import { useAreasQuery } from '../queries/useAreasQuery';
import { useControlledDocumentDetailQuery } from '../../controlled-documents/queries/useControlledDocumentDetailQuery';
import { buildVisibilityLabel, displayRevisionTitle, formatRevisionCode, hasSettledSidebarIdentity, resolveAreaLabel, resolveProfileLabel } from '../lib/documentDetailMeta';
import { parseDocumentStatus } from '../lib/parseDocumentStatus';
import styles from './styles/DocumentEditorPage.module.css';

export type DocumentEditorPageProps = {
  documentID: string;
  onDone: () => void;
};

type EditorDocumentDetail = DocumentDetail;
type ArtifactMetadata = { fileSizeBytes?: number | null; pageCount?: number | null };
type SubmitDocumentRequest = components['schemas']['SubmitDocumentRequest'];
type ReasonCategory = NonNullable<SubmitDocumentRequest['reason_category']>;

const REASON_CATEGORY_OPTIONS: Array<{ value: ReasonCategory; label: string }> = [
  { value: 'content', label: 'Conteúdo' },
  { value: 'corrective', label: 'Corretiva' },
  { value: 'regulatory', label: 'Regulatória' },
  { value: 'periodic_review', label: 'Revisão periódica' },
  { value: 'administrative', label: 'Administrativa' },
];

export function DocumentEditorPage({ documentID, onDone }: DocumentEditorPageProps): React.ReactElement {
  const queryClient = useQueryClient();
  const docQuery = useDocumentDetailQuery(documentID);
  const [documentName, setDocumentName] = useState('');
  const [revisionTitleDialogOpen, setRevisionTitleDialogOpen] = useState(false);
  const [revisionTitleInput, setRevisionTitleInput] = useState('');
  const [revisionTitleError, setRevisionTitleError] = useState<string | null>(null);
  const [reasonForChangeInput, setReasonForChangeInput] = useState('');
  const [reasonForChangeError, setReasonForChangeError] = useState<string | null>(null);
  const [reasonCategoryInput, setReasonCategoryInput] = useState('');
  const [reasonCategoryError, setReasonCategoryError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [editorDirty, setEditorDirty] = useState(false);
  const [submitStatusOverride, setSubmitStatusOverride] = useState<'under_review' | null>(null);
  const [artifactMetadata, setArtifactMetadata] = useState<ArtifactMetadata>({});
  const [trackedChanges, setTrackedChanges] = useState<TrackedChange[]>([]);
  const editorRef = useRef<MetalDocsEditorRef>(null);
  const skipInitialEditorChangeRef = useRef(false);
  const baseDoc = (docQuery.data as EditorDocumentDetail | undefined) ?? null;
  const doc: EditorDocumentDetail | null = useMemo(() => {
    if (!baseDoc) {
      return null;
    }
    if (!submitStatusOverride) {
      return baseDoc;
    }
    return {
      ...baseDoc,
      Status: submitStatusOverride,
      status: submitStatusOverride,
    };
  }, [baseDoc, submitStatusOverride]);
  const docStatus = doc?.status ?? '';
  const session = useDocumentSession(documentID, { enabled: docStatus === 'draft' });
  const controlledDocumentId = doc?.controlled_document_id ?? '';
  const profilesQuery = useProfilesQuery();
  const areasQuery = useAreasQuery();
  const controlledDocumentQuery = useControlledDocumentDetailQuery(controlledDocumentId);
  const revisionHistoryQuery = useDocumentRevisionHistoryQuery(documentID);

  useEffect(() => {
    setSubmitStatusOverride(null);
  }, [documentID]);

  useEffect(() => {
    if ((baseDoc?.status ?? '') !== 'draft') {
      setSubmitStatusOverride(null);
    }
  }, [baseDoc]);

  useEffect(() => {
    if (!doc) {
      return;
    }
    setDocumentName(doc.name ?? 'Document');
    setArtifactMetadata({
      fileSizeBytes: doc.current_revision_file_size_bytes ?? null,
      pageCount: doc.current_revision_page_count ?? null,
    });
  }, [doc]);

  // Buffer ownership moved to DocumentShell; the dirty-guard resets whenever the
  // current revision changes (a fresh buffer load is imminent), matching the old
  // per-buffer reset without the page tracking the buffer itself.
  useEffect(() => {
    skipInitialEditorChangeRef.current = true;
    setEditorDirty(false);
  }, [doc?.current_revision_id]);

  const pageLoadError = (() => {
    const err = docQuery.error;
    if (!docQuery.isError || !err) return null;
    if (err instanceof ApiError) {
      return err.code === 'NOT_FOUND' ? 'Documento não encontrado.' : resolveErrorMessage(err.code, err.message);
    }
    if (err && typeof err === 'object' && 'code' in err) {
      const code = (err as { code?: string }).code;
      const message = 'message' in err ? (err as { message?: string }).message : undefined;
      return code === 'NOT_FOUND' ? 'Documento não encontrado.' : resolveErrorMessage(code, message);
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
    await autosave.queue(buf, doc.form_data_json ?? null, pageCount);
    setEditorDirty(false);
  }

  function handleEditorChange() {
    if (skipInitialEditorChangeRef.current) {
      skipInitialEditorChangeRef.current = false;
      return;
    }
    setEditorDirty(true);
  }

  async function submitForReview(body: SubmitDocumentRequest = {}) {
    if (session.state.phase !== 'writer' || !doc) return;
    if (isSubmitting) return;
    setIsSubmitting(true);
    try {
      // Clean-buffer re-submit sequence (F6/C5): strip the visual marks of
      // already-resolved comments BEFORE the flush below persists the buffer,
      // so the re-submitted revision doesn't carry stale comment marks. Order
      // matters: removeCommentMark (per resolved comment) -> saveNow/queue/
      // flush -> submitForReviewRequest. removeCommentMark mutates the buffer
      // without flipping editorDirty, so force the flush path whenever
      // changesRequested even if editorDirty is otherwise false.
      let resolvedAnyMark = false;
      if (changesRequested) {
        for (const comment of commentsHook.comments) {
          if (comment.resolved) {
            editorRef.current?.removeCommentMark(String(comment.id));
            resolvedAnyMark = true;
          }
        }
      }
      if (editorDirty || resolvedAnyMark) {
        const latestBuf = await editorRef.current?.saveNow();
        if (latestBuf) {
          const pageCount = editorRef.current?.getPageCount() ?? null;
          await autosave.queue(latestBuf, doc.form_data_json ?? null, pageCount);
        }
        const flushOk = await autosave.flush();
        if (!flushOk) {
          toast.error('Erro ao salvar documento antes da submissão.');
          return;
        }
        setEditorDirty(false);
      }
      // If-Match is mandatory server-side (CON-01 OCC parity with canonical
      // /submit, DEC-01). Autosave commits advance the revision *id*, not
      // documents.revision_version (that only bumps on the draft->under_review
      // transition itself), so the cached doc.revision_version read here is
      // still the version the server expects to match.
      await submitForReviewRequest(documentID, body, {
        ifMatch: `"v${doc.revision_version}"`,
        idempotencyKey: crypto.randomUUID(),
      });
      setSubmitStatusOverride('under_review');
      queryClient.setQueryData(QK.documents.detail(documentID), (current: DocumentDetail | undefined) => {
        if (!current) return current;
        return {
          ...current,
          Status: 'under_review',
          status: 'under_review',
        } as DocumentDetail;
      });
      setEditorDirty(false);
      let rehydrationFailed = false;
      try {
        const detailResult = await docQuery.refetch();
        if (detailResult.error) {
          throw detailResult.error;
        }
        await Promise.allSettled([
          queryClient.fetchQuery({
            queryKey: QK.documents.revisionHistory(documentID),
            queryFn: () => getDocumentRevisionHistory(documentID),
          }),
          queryClient.fetchQuery({
            queryKey: QK.approval.instance(documentID),
            queryFn: () => getApprovalInstance(documentID),
          }),
        ]);
      } catch {
        rehydrationFailed = true;
        void queryClient.invalidateQueries({ queryKey: QK.documents.detail(documentID) });
        void queryClient.invalidateQueries({ queryKey: QK.documents.revisionHistory(documentID) });
        void queryClient.invalidateQueries({ queryKey: QK.approval.instance(documentID) });
      }
      await session.release();
      setRevisionTitleDialogOpen(false);
      setRevisionTitleInput('');
      setRevisionTitleError(null);
      setReasonForChangeInput('');
      setReasonForChangeError(null);
      setReasonCategoryInput('');
      setReasonCategoryError(null);
      onDone();
      if (rehydrationFailed) {
        toast.warning('Documento submetido. Recarregando dados atualizados do editor.');
      } else {
        toast.success('Documento submetido para revisão.');
      }
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === 'validation.revision_title_required') {
          setRevisionTitleError(resolveErrorMessage(err.code, err.message));
        } else if (err.code === 'validation.reason_for_change_required') {
          setReasonForChangeError(resolveErrorMessage(err.code, err.message));
        } else if (err.code === 'validation.reason_category_invalid') {
          setReasonCategoryError(resolveErrorMessage(err.code, err.message));
        } else {
          toast.error(resolveErrorMessage(err.code, err.message));
        }
      } else {
        toast.error('Erro ao submeter documento.');
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  function handleSubmitForReview() {
    if (isSubmitting) return;
    if ((doc?.revision_number ?? 0) === 0) {
      void submitForReview({});
      return;
    }
    setRevisionTitleInput(doc?.revision_title ?? '');
    setRevisionTitleError(null);
    setReasonForChangeInput('');
    setReasonForChangeError(null);
    setReasonCategoryInput('');
    setReasonCategoryError(null);
    setRevisionTitleDialogOpen(true);
  }

  function handleConfirmSubmitForReview() {
    if (isSubmitting) return;
    const trimmedTitle = revisionTitleInput.trim();
    const trimmedReason = reasonForChangeInput.trim();
    let hasError = false;
    if (!trimmedTitle) {
      setRevisionTitleError('Informe o título da revisão para submeter.');
      hasError = true;
    }
    if (!trimmedReason) {
      setReasonForChangeError('Informe o motivo da alteração para submeter.');
      hasError = true;
    }
    if (hasError) return;
    const body: SubmitDocumentRequest = {
      revision_title: trimmedTitle,
      reason_for_change: trimmedReason,
      ...(reasonCategoryInput ? { reason_category: reasonCategoryInput as ReasonCategory } : {}),
    };
    void submitForReview(body);
  }

  // Broadened detection gate (F6/C5): a `request_changes` verdict reverts the
  // document to `draft` in-tx (review_verdict_service.go:340-347), so the
  // instance query must also fire for `draft` — otherwise the author never
  // sees the `changes_requested` instance that prompted the reopen. A plain
  // never-submitted draft has no instance; the query resolves to undefined
  // gracefully (no panel, no error) rather than throwing.
  const approvalInstanceQuery = useQuery({
    queryKey: QK.approval.instance(documentID),
    queryFn: () => getApprovalInstance(documentID),
    enabled: documentID.length > 0 && (docStatus === 'under_review' || docStatus === 'draft'),
    retry: false,
  });
  const changesRequested = approvalInstanceQuery.data?.status === 'changes_requested';
  const canEditContent = session.state.phase === 'writer' && docStatus === 'draft';
  const canComment = Boolean(doc) && (docStatus === 'draft' || docStatus === 'under_review');
  const docCode = doc?.code ?? '';
  const revisionBadge = doc?.revision_number != null ? formatRevisionCode(doc.revision_number) : null;
  const displayName = documentName.replace(/\.docx$/i, '');
  const userID = doc?.created_by ?? '';
  const authorDisplay = String(userID);
  const commentsHook = useDocumentComments(documentID, authorDisplay);
  const canUseComments = canComment && !commentsHook.loadError;
  const unresolvedComments = commentsHook.comments.filter((c) => !c.resolved);
  const hasUnresolvedTrackedChanges = trackedChanges.length > 0;
  const hasUnresolvedComments = unresolvedComments.length > 0;
  const requestedChangesBlocksSubmit = changesRequested
    && (hasUnresolvedTrackedChanges || hasUnresolvedComments);
  const canMountEditor = !!doc
    && (docStatus === 'draft'
      ? session.state.phase !== 'idle' && session.state.phase !== 'acquiring'
      : true);

  const [sidebarOpen, setSidebarOpen] = useState<boolean>(
    () => localStorage.getItem('editor-sidebar-open') !== 'false'
  );

  // AutosaveState (7-state) is structurally identical to useDocumentAutosave.AutosaveStatus.
  const autosaveState: AutosaveState = autosave.status;

  const statusForPill = parseDocumentStatus(docStatus);

  const profileLabel = doc?.profile_code_snapshot && profilesQuery.data
    ? resolveProfileLabel(doc.profile_code_snapshot, profilesQuery.data)
    : null;
  const areaLabel = doc?.process_area_code_snapshot && areasQuery.data
    ? resolveAreaLabel(doc.process_area_code_snapshot, areasQuery.data)
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
  const lineage: VersionHistoryItem[] = (revisionHistoryQuery.data?.items ?? []).map((item) => {
    const revisionLabel = formatRevisionCode(item.revision_number);
    return {
      // API exposes only revision_number; versionNumber shadows it until the schema exposes a separate version counter.
      versionNumber: item.revision_number,
      revisionNumber: item.revision_number,
      revisionLabel,
      status: item.status as VersionHistoryItem['status'],
      title: displayRevisionTitle(item.revision_title, revisionLabel),
      createdAt: item.created_at,
      isCurrent: item.is_current,
    };
  });

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
          ) : canMountEditor ? (
            <DocumentShell
              documentId={documentID}
              currentRevisionId={doc?.current_revision_id ?? null}
              editorMode={canEditContent ? 'document-edit' : 'readonly'}
              editorRef={editorRef}
              author={authorDisplay}
              comments={commentsHook.comments}
              onCommentsChange={commentsHook.setComments}
              onCommentAdd={(c: EditorComment) => { if (canUseComments) void commentsHook.add(c); }}
              onCommentResolve={(c: EditorComment) => { if (canUseComments) void (c.resolved ? commentsHook.resolve(c) : commentsHook.reopen(c)); }}
              onCommentDelete={(c: EditorComment) => { if (canUseComments) void commentsHook.remove(c); }}
              onCommentReply={(reply: EditorComment, parent: EditorComment) => { if (canUseComments) void commentsHook.reply(reply, parent); }}
              onAutoSave={handleSave}
              onChange={handleEditorChange}
              onDocumentNameChange={handleRename}
              onTrackedChangesChange={setTrackedChanges}
              chrome={{
                center: (
                  <>
                    {docCode && <CodeChip>{docCode}</CodeChip>}
                    {displayName && <span className={editorChromeStyles.docTitle}>{displayName}</span>}
                    {revisionBadge && <VersionBadge>{revisionBadge}</VersionBadge>}
                    {statusForPill && <StatusPill status={statusForPill} />}
                  </>
                ),
                right: (
                  <>
                    <AutosaveStatus status={autosaveState} />
                    <button
                      type="button"
                      className={editorChromeStyles.primaryBtn}
                      onClick={handleSubmitForReview}
                      disabled={!canEditContent || isSubmitting || requestedChangesBlocksSubmit}
                    >
                      Submeter para revisão
                    </button>
                  </>
                ),
              }}
              notice={
                commentsHook.loadError ? (
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
                ) : null
              }
            />
          ) : null}
        </main>
        {!pageLoadError ? (
          <ArtifactMetaSidebar
            open={sidebarOpen}
            onToggle={() => {
              setSidebarOpen((prev) => {
                const next = !prev;
                localStorage.setItem('editor-sidebar-open', String(next));
                return next;
              });
            }}
            loading={!sidebarIdentityReady}
            code={docCode || null}
            meta={{
              profileLabel,
              areaLabel,
              visibilityLabel,
              fileSizeBytes: artifactMetadata.fileSizeBytes ?? null,
              pageCount: artifactMetadata.pageCount ?? null,
              createdAt: null,
              effectiveFrom: null,
              nextReviewAt: null,
              ownerName: null,
              ownerDescriptor: null,
            }}
            // Editor sidebar surfaces the signoff chain only while the document is
            // actively under review; once approved/published the dedicated approval
            // screen owns the chain. ArtifactMetaSidebar is status-agnostic by design
            // (m-3) — this gate is deliberate call-site data-scoping, not lifecycle
            // coupling leaking back into the presentational layer.
            approvalChain={
              docStatus === 'under_review' && approvalInstanceQuery.data
                ? approvalInstanceQuery.data.stages.flatMap((stage, stageIndex) =>
                    stage.actors.map((actor) => ({
                      stageIndex,
                      label: stage.label,
                      status: actor.status,
                      // The stage label doubles as the role descriptor in the flow-viz
                      // (parity with the document mapApprovalChain — no separate role field).
                      roleLabel: stage.label,
                      // This instance DTO carries the resolved outcome in `actor.status`
                      // directly (approved/rejected/active/waiting); the shared helper
                      // maps it to ApprovalFlowState (single source of truth).
                      flowState: actorStatusToFlowState(actor.status),
                      actorUserId: actor.user_id,
                      actorDisplay: actor.display_name,
                      decision: actor.decision ?? null,
                      signedAt: null,
                    })),
                  )
                : null
            }
            lineage={lineage}
            ariaLabel="Identificação do documento"
            loadingLabel="Carregando metadados do documento"
          />
        ) : null}
        {!pageLoadError && changesRequested ? (
          <RequestedChangesPanel
            trackedChanges={trackedChanges}
            comments={commentsHook.comments}
            onAcceptChange={(revisionId) => editorRef.current?.acceptChange(revisionId)}
            onRejectChange={(revisionId) => editorRef.current?.rejectChange(revisionId)}
            onResolveComment={(comment) => void commentsHook.resolve(comment)}
          />
        ) : null}
        {revisionTitleDialogOpen ? (
          <div className={styles.dialogBackdrop} role="presentation">
            <div className={styles.dialogPanel} role="dialog" aria-modal="true" aria-labelledby="revision-title-dialog-title">
              <h2 id="revision-title-dialog-title" className={styles.dialogTitle}>Título da revisão</h2>
              <p className={styles.dialogText}>Informe o motivo governado desta revisão antes de submeter o documento para aprovação.</p>
              <label className={styles.dialogField}>
                <span>Título</span>
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
              <label className={styles.dialogField}>
                <span>Motivo da alteração</span>
                <textarea
                  value={reasonForChangeInput}
                  onChange={(event) => {
                    setReasonForChangeInput(event.target.value);
                    if (reasonForChangeError) setReasonForChangeError(null);
                  }}
                  placeholder="Descreva o motivo desta alteração governada"
                />
              </label>
              {reasonForChangeError ? <p role="alert" className={styles.dialogError}>{reasonForChangeError}</p> : null}
              <label className={styles.dialogField}>
                <span>Categoria</span>
                <select
                  value={reasonCategoryInput}
                  onChange={(event) => {
                    setReasonCategoryInput(event.target.value);
                    if (reasonCategoryError) setReasonCategoryError(null);
                  }}
                >
                  <option value="">— Selecionar —</option>
                  {REASON_CATEGORY_OPTIONS.map((option) => (
                    <option key={option.value} value={option.value}>{option.label}</option>
                  ))}
                </select>
              </label>
              {reasonCategoryError ? <p role="alert" className={styles.dialogError}>{reasonCategoryError}</p> : null}
              <div className={styles.dialogActions}>
                <button type="button" className={styles.dialogSecondaryBtn} onClick={() => setRevisionTitleDialogOpen(false)}>
                  Cancelar
                </button>
                <button
                  type="button"
                  className={styles.dialogPrimaryBtn}
                  onClick={handleConfirmSubmitForReview}
                  disabled={isSubmitting}
                >
                  Confirmar submissão
                </button>
              </div>
            </div>
          </div>
        ) : null}
      </div>
    </div>
  );
}


