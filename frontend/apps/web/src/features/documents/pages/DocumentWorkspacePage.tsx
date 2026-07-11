import { useEffect, useMemo, useRef, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import type { MetalDocsEditorRef, EditorComment, TrackedChange } from '@metaldocs/editor-ui';

import { CodeChip, InlineAlert, StatusPill } from '../../../components/ui';
import { useAuthStore } from '../../../store/auth.store';
import { formatRevisionCode } from '../../../lib/labels/revisionCode';
import { QK } from '../../../lib/queryKeys';
import { ApiError, resolveErrorMessage } from '../../../lib/api';
import type { components } from '../../../lib/api-types';
import type { ArtifactAction } from '../../shared/controlled-artifact/types';
import { useHasCapability } from '../../iam/hooks/useHasCapability';
import { CancelInstanceDialog } from '../../approval/components/CancelInstanceDialog';
import { deriveWorkspaceMode } from '../../approval/lib/workspaceMode';
import { useSignoffMutation } from '../../approval/hooks/useSignoffMutation';
import { submit as submitForReviewRequest } from '../../approval/api/approvalApi';
import type { DocumentDetail } from '../api/documents';
import { AutosaveStatus, editorChromeStyles, type AutosaveState } from '../../shared/components/editor-chrome';
import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
import { useApprovalInstanceQuery } from '../queries/useApprovalInstanceQuery';
import { useControlledDocumentActiveDocumentQuery } from '../queries/useControlledDocumentActiveDocumentQuery';
import { useDocumentSession } from '../hooks/editor/useDocumentSession';
import { useDocumentAutosave } from '../hooks/editor/useDocumentAutosave';
import { useDocumentComments } from '../hooks/editor/useDocumentComments';
import { buildDocumentSignoffDecision } from '../lib/documentSignoffDecision';
import { DocumentShell } from '../components/DocumentShell';
import { EditorCanvas } from '../components/workspace/EditorCanvas';
import { PdfCanvas } from '../components/workspace/PdfCanvas';
import { ApprovingDisclosure } from '../components/workspace/ApprovingDisclosure';
import { RequestedChangesPanel } from '../components/RequestedChangesPanel';
import { AuthorCommentsPanel } from '../components/workspace/AuthorCommentsPanel';
import { ModeChip } from '../components/workspace/ModeChip';
import { WorkspaceSidebar } from '../components/workspace/WorkspaceSidebar';
import { parseDocumentStatus } from '../lib/parseDocumentStatus';
import styles from './DocumentWorkspacePage.module.css';

/**
 * F2d.5 S2b — the mode-adaptive single working screen owner.
 *
 * Constant shell (header + ModeChip + canvas + WorkspaceSidebar) across every
 * `WorkspaceMode` (F2d.3, `deriveWorkspaceMode`); the canvas content and the
 * sidebar's decision footer vary by mode:
 *  - author-editing / author-changes-requested: the writable EditorCanvas
 *    (autosave-backed), with a teaching-copy banner for changes-requested
 *    plus the RequestedChangesPanel (F6) as a sidebar contextual panel.
 *  - approving: the frozen read canvas + ApprovingDisclosure (hash/ETag +
 *    delegation badge) + a signature decision (unselected on mount — no
 *    preselection, spec §5).
 *  - reviewing / observing / author-waiting / lifecycle: unchanged S2a read
 *    canvas.
 */

// F2d.5b D1 — statuses whose official PDF the backend serves via
// GET /documents/:id/view (view_service.go viewableStatuses). Keyed on
// status, NOT on mode: 'lifecycle' also covers superseded/obsolete, which
// /view does not serve — those keep the docx read canvas.
const OFFICIAL_PDF_STATUSES = new Set(['approved', 'scheduled', 'published']);

// F2d.8 RED-1 — the draft submit-for-approval affordance, re-homed from the
// retired (unrouted) DocumentEditorPage into the single-screen workspace. The
// submit contract (If-Match "v{revision_version}", UUID Idempotency-Key, Origin
// via the shared transport) is unchanged; only its host screen moved.
type SubmitDocumentRequest = components['schemas']['SubmitDocumentRequest'];
type ReasonCategory = NonNullable<SubmitDocumentRequest['reason_category']>;

const REASON_CATEGORY_OPTIONS: Array<{ value: ReasonCategory; label: string }> = [
  { value: 'content', label: 'Conteúdo' },
  { value: 'corrective', label: 'Corretiva' },
  { value: 'regulatory', label: 'Regulatória' },
  { value: 'periodic_review', label: 'Revisão periódica' },
  { value: 'administrative', label: 'Administrativa' },
];

export function DocumentWorkspacePage() {
  const { documentId = '' } = useParams<{ documentId: string }>();
  const currentUser = useAuthStore((s) => s.user);

  const docQuery = useDocumentDetailQuery(documentId);
  const instanceQuery = useApprovalInstanceQuery(documentId);

  const doc = docQuery.data ?? null;
  const instance = instanceQuery.data ?? null;
  const viewer = instance?.viewer ?? null;

  // Design brief §"No instance + draft = author-editing": a never-submitted
  // draft has no approval instance (GET .../approval-instance 404s), so there is
  // no server `viewer` to carry is_author — and deriveWorkspaceMode reaches the
  // author lens only through viewer.is_author. Authorship for that pre-submission
  // case is document OWNERSHIP (created_by), not stage eligibility, so deriving
  // it here is consistent with ADR 0078 (which governs eligibility, still
  // server-only). The server stays the authority on submit (writer lease +
  // authz.Require). Once an instance exists its real viewer takes precedence.
  const effectiveViewer = viewer
    ?? (currentUser && doc?.created_by === currentUser.userId
      ? { is_author: true, eligible_for_active_stage: false, has_signed_active_stage: false, via_delegation_from: null }
      : null);

  // Mode is computed up-front (deriveWorkspaceMode accepts a possibly-null
  // doc) so every hook below can gate on it without violating the rules of
  // hooks — the loading/doc-not-found early returns further down must stay
  // the LAST thing before render, no hooks after them.
  const mode = deriveWorkspaceMode(doc, instance, effectiveViewer);
  const docStatus = doc?.status ?? '';

  const session = useDocumentSession(documentId, { enabled: docStatus === 'draft' });
  const editorRef = useRef<MetalDocsEditorRef>(null);
  const [trackedChanges, setTrackedChanges] = useState<TrackedChange[]>([]);

  const sessionPhase = session.state.phase;
  const sessionID = sessionPhase === 'writer' ? session.state.sessionID : '';
  const lastAckRevisionID = sessionPhase === 'writer' ? session.state.lastAckRevisionID : '';
  const { setLastAck } = session;

  const autosaveArgs = useMemo(() => {
    if (sessionPhase === 'writer') {
      return {
        documentID: documentId,
        sessionID,
        baseRevisionID: lastAckRevisionID,
        onAdvanceBase: (newRevisionID: string) => setLastAck(newRevisionID),
        onSessionLost: () => {},
      };
    }
    return {
      documentID: documentId,
      sessionID: '',
      baseRevisionID: '',
      onAdvanceBase: () => {},
      onSessionLost: () => {},
    };
  }, [documentId, sessionPhase, sessionID, lastAckRevisionID, setLastAck]);

  const autosave = useDocumentAutosave(autosaveArgs);

  // Mirrors DocumentEditorPage's authorDisplay: the document's creator id,
  // not the current viewer — comment attribution stays keyed to the
  // document's own author, not whoever happens to be looking at it.
  const authorDisplay = String(doc?.created_by ?? '');
  const commentsHook = useDocumentComments(documentId, authorDisplay);

  // Only fetch the active-document context (content_hash/revision_version)
  // while approving — every other mode passes undefined, which the query
  // hook's own `enabled: Boolean(controlledDocumentId)` gate turns into a
  // no-op.
  const controlledDocumentId = mode === 'approving' ? doc?.controlled_document_id ?? undefined : undefined;
  const contextQuery = useControlledDocumentActiveDocumentQuery(controlledDocumentId);
  const contentHash = contextQuery.data?.content_hash ?? null;
  const revisionVersion = contextQuery.data?.revision_version ?? doc?.revision_version ?? 0;

  const { signOff } = useSignoffMutation({
    documentId,
    contentHash: contentHash ?? '',
    revisionVersion,
  });

  // ADR 0022 — the cancel affordance is gated on the SAME capability the
  // backend enforces at BOTH tiers (tier-1 permissions.go route→capability +
  // tier-2 authz.Require in cancel_service.go): document.edit, area-scoped.
  // Never role-reasoned. This is only the affordance gate; the server remains
  // the authority (area scope it can't see from here, plus a fail-closed
  // tripwire), so a false-positive here degrades to a 403, never a bypass.
  const canCancelInstance = useHasCapability('document.edit');
  const [showCancelDialog, setShowCancelDialog] = useState(false);

  // F2d.8 RED-1 — submit-for-approval affordance state (ported from the retired
  // DocumentEditorPage). editorDirty gates the pre-submit buffer flush; the
  // dialog fields collect the governed revision reason for revision_number > 0.
  const queryClient = useQueryClient();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [editorDirty, setEditorDirty] = useState(false);
  const [revisionDialogOpen, setRevisionDialogOpen] = useState(false);
  const [revisionTitleInput, setRevisionTitleInput] = useState('');
  const [revisionTitleError, setRevisionTitleError] = useState<string | null>(null);
  const [reasonForChangeInput, setReasonForChangeInput] = useState('');
  const [reasonForChangeError, setReasonForChangeError] = useState<string | null>(null);
  const [reasonCategoryInput, setReasonCategoryInput] = useState('');
  const [reasonCategoryError, setReasonCategoryError] = useState<string | null>(null);
  const skipInitialEditorChangeRef = useRef(false);

  // Dirty-guard resets whenever the current revision changes (a fresh buffer
  // load is imminent) — mirrors the retired editor's per-buffer reset.
  useEffect(() => {
    skipInitialEditorChangeRef.current = true;
    setEditorDirty(false);
  }, [doc?.current_revision_id]);

  // §6 loading — shell skeleton, no central spinner.
  if (docQuery.isLoading) {
    return (
      <div className={styles.page} data-testid="workspace-page">
        <div className={styles.body}>
          <div className={styles.rail} aria-hidden="true" />
          <div className={styles.main}>
            <div className={styles.header} data-testid="workspace-skeleton">
              <div className={styles.skeletonBlock} />
              <div className={styles.skeletonChip} />
            </div>
            <div className={styles.canvas}>
              <div className={styles.skeletonCanvas} />
            </div>
          </div>
          <div className={styles.sidebarSkeleton} aria-hidden="true" />
        </div>
      </div>
    );
  }

  // Doc-not-found — teaching-copy empty state (doc query error, or settled
  // with no data).
  if (docQuery.isError || !doc) {
    return (
      <div className={styles.page} data-testid="workspace-page">
        <div className={styles.emptyState} role="alert">
          <p className={styles.emptyStateTitle}>Não foi possível localizar este documento.</p>
          <p className={styles.emptyStateHint}>
            Verifique se o link está correto ou se você ainda tem acesso a este documento.
          </p>
        </div>
      </div>
    );
  }

  const activeStage = instance?.stages?.find((s) => s.status === 'active');
  // isFetching (not isLoading): a manual retry after an error should still
  // surface as "loading", which v5 isLoading (first-fetch-only) would miss.
  const instanceLoading = instanceQuery.isFetching && instance == null;
  // §6 instance error — the canvas stays readable; only the sidebar
  // surfaces the error (instance is optional context, not a blocking read).
  const instanceError = instanceQuery.isError ? 'Não foi possível carregar os dados de aprovação.' : null;
  const statusForPill = parseDocumentStatus(doc.status);

  const canEditContent = session.state.phase === 'writer' && docStatus === 'draft';
  const canMountEditor = docStatus === 'draft'
    ? session.state.phase !== 'idle' && session.state.phase !== 'acquiring'
    : true;

  const formDataJson = doc.form_data_json ?? null;
  async function handleSave(buf: ArrayBuffer) {
    const pageCount = editorRef.current?.getPageCount() ?? null;
    await autosave.queue(buf, formDataJson, pageCount);
  }

  // --- F2d.8 RED-1: submit-for-approval (draft → under_review) ---
  const isAuthorEditing = mode === 'author-editing' || mode === 'author-changes-requested';
  const changesRequested = mode === 'author-changes-requested';
  const autosaveState: AutosaveState = autosave.status;
  const unresolvedComments = commentsHook.comments.filter((c) => !c.resolved);
  // A changes-requested resubmit is blocked while unresolved tracked changes or
  // comments remain — the author must respond before reopening the flow.
  const requestedChangesBlocksSubmit = changesRequested
    && (trackedChanges.length > 0 || unresolvedComments.length > 0);

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
      // Clean-buffer re-submit (F6/C5): strip visual marks of already-resolved
      // comments BEFORE the flush persists the buffer, so the re-submitted
      // revision carries no stale marks. Order: removeCommentMark -> saveNow/
      // queue/flush -> submit. removeCommentMark mutates without flipping
      // editorDirty, so force the flush path when any mark was stripped.
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
          await autosave.queue(latestBuf, formDataJson, pageCount);
        }
        const flushOk = await autosave.flush();
        if (!flushOk) {
          toast.error('Erro ao salvar documento antes da submissão.');
          return;
        }
        setEditorDirty(false);
      }
      // If-Match is mandatory server-side (OCC parity with canonical /submit).
      // Autosave commits advance the revision *id*, not documents.revision_version
      // (that only bumps on the draft->under_review transition itself), so the
      // cached doc.revision_version read here is still the version to match.
      await submitForReviewRequest(documentId, body, {
        ifMatch: `"v${doc.revision_version}"`,
        idempotencyKey: crypto.randomUUID(),
      });
      // Flip the cached status immediately so the mode recomputes off
      // under_review before the refetch lands (no editable-canvas flash).
      queryClient.setQueryData(QK.documents.detail(documentId), (current: DocumentDetail | undefined) => {
        if (!current) return current;
        return { ...current, Status: 'under_review', status: 'under_review' } as DocumentDetail;
      });
      setEditorDirty(false);
      await docQuery.refetch();
      await instanceQuery.refetch();
      await session.release();
      setRevisionDialogOpen(false);
      setRevisionTitleInput('');
      setRevisionTitleError(null);
      setReasonForChangeInput('');
      setReasonForChangeError(null);
      setReasonCategoryInput('');
      setReasonCategoryError(null);
      toast.success('Documento submetido para revisão.');
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
    // First submission of a brand-new draft carries no governed revision reason;
    // a resubmit of an existing revision requires title + reason via the dialog.
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
    setRevisionDialogOpen(true);
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

  // signOff (the shared useSignoffMutation, If-Match/content_hash unchanged)
  // then refetch the instance so the mode/decision recompute post-signature.
  const decisionSubmit = async (input: { optionKey: string; reason: string; password: string }) => {
    await signOff({
      decision: input.optionKey === 'approve' ? 'approve' : 'reject',
      reason: input.reason || undefined,
      password: input.password,
    });
    await instanceQuery.refetch();
  };

  const signer = currentUser
    ? { displayName: currentUser.displayName, email: currentUser.email ?? null, username: currentUser.username }
    : null;

  const decision = mode === 'approving'
    ? buildDocumentSignoffDecision({
        offered: instance != null && contentHash != null,
        signer,
        submit: decisionSubmit,
      }) ?? null
    : null;

  const delegatedFrom = viewer?.via_delegation_from?.display_name ?? null;
  const frozenContentHash = instance?.frozen_content_hash ?? null;

  // F2d.6 — build a reply EditorComment from composer text (fresh monotonic id;
  // single text paragraph body) and delegate to the existing reply mutation.
  const handleAuthorReply = (text: string, parent: EditorComment) => {
    const nextId = commentsHook.comments.reduce((max, c) => Math.max(max, Number(c.id) || 0), 0) + 1;
    const reply: EditorComment = {
      id: nextId,
      parentId: Number(parent.id),
      author: currentUser?.displayName ?? '',
      body: [{ type: 'paragraph', content: [{ type: 'text', text }] }],
      resolved: false,
    };
    void commentsHook.reply(reply, parent);
  };

  const contextualPanel = mode === 'author-changes-requested'
    ? (
        <RequestedChangesPanel
          trackedChanges={trackedChanges}
          comments={commentsHook.comments}
          onAcceptChange={(revisionId) => editorRef.current?.acceptChange(revisionId)}
          onRejectChange={(revisionId) => editorRef.current?.rejectChange(revisionId)}
          onResolveComment={(comment) => void commentsHook.resolve(comment)}
        />
      )
    : mode === 'author-waiting'
    ? (
        <AuthorCommentsPanel
          comments={commentsHook.comments}
          onReply={handleAuthorReply}
          onResolve={(c) => void commentsHook.resolve(c)}
          onReopen={(c) => void commentsHook.reopen(c)}
        />
      )
    : null;

  const canUseComments = docStatus === 'draft' || docStatus === 'under_review';

  // Lifecycle actions surfaced in the sidebar footer's "Outras ações" group.
  // Cancel requires an actively-in-progress instance (instance.status
  // 'in_progress' ↔ document under_review); a terminal / changes_requested
  // instance is not cancelable. This mirrors the retired cockpit's
  // TRANSITION_POLICY.cancelInstance gate (under_review only), now keyed on
  // instance truth rather than document status. Offered across every
  // in_progress mode (reviewing / approving / observing / author-waiting) since
  // cancel authority is capability-scoped, not stage-eligibility-scoped.
  const lifecycleActions: ArtifactAction[] = [];
  if (instance?.status === 'in_progress' && canCancelInstance) {
    lifecycleActions.push({
      key: 'cancel',
      label: 'Cancelar instância',
      variant: 'secondary',
      available: true,
      run: () => setShowCancelDialog(true),
    });
  }

  return (
    <div className={styles.page} data-testid="workspace-page">
      <div className={styles.body}>
        <Link to="/documents" className={styles.rail} aria-label="Voltar para a biblioteca">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="15 18 9 12 15 6" />
          </svg>
        </Link>

        <div className={styles.main}>
          <header className={styles.header}>
            <div className={styles.headerLeft}>
              {doc.code ? <CodeChip>{doc.code}</CodeChip> : null}
              <h1 className={styles.title}>{doc.name}</h1>
              {statusForPill ? <StatusPill status={statusForPill} /> : null}
              {doc.revision_number != null ? (
                <span className={styles.revision}>{formatRevisionCode(doc.revision_number)}</span>
              ) : null}
            </div>
            <div className={styles.headerRight}>
              {isAuthorEditing ? (
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
              ) : null}
              <ModeChip mode={mode} viewer={effectiveViewer} />
            </div>
          </header>

          <main className={styles.canvas} data-testid="workspace-canvas">
            {mode === 'approving' ? (
              <>
                <ApprovingDisclosure
                  documentId={documentId}
                  contentHash={contentHash}
                  frozenContentHash={frozenContentHash}
                  delegatedFrom={delegatedFrom}
                />
                {doc.current_revision_id ? (
                  <DocumentShell
                    documentId={documentId}
                    currentRevisionId={doc.current_revision_id}
                    editorMode="readonly"
                    author={currentUser?.displayName ?? ''}
                  />
                ) : (
                  <div className={styles.canvasEmpty}>Este documento ainda não possui conteúdo para exibir.</div>
                )}
              </>
            ) : mode === 'author-editing' || mode === 'author-changes-requested' ? (
              <>
                {mode === 'author-changes-requested' ? (
                  <InlineAlert
                    tone="warning"
                    className={styles.changesBanner}
                    message="Mudanças foram solicitadas nesta etapa. Revise os comentários e marcações ao lado e responda antes de reenviar para revisão."
                  />
                ) : null}
                {doc.current_revision_id ? (
                  <EditorCanvas
                    canMountEditor={canMountEditor}
                    documentId={documentId}
                    currentRevisionId={doc.current_revision_id}
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
                    onTrackedChangesChange={setTrackedChanges}
                  />
                ) : (
                  <div className={styles.canvasEmpty}>Este documento ainda não possui conteúdo para exibir.</div>
                )}
              </>
            ) : OFFICIAL_PDF_STATUSES.has(docStatus) ? (
              <PdfCanvas documentId={documentId} />
            ) : doc.current_revision_id ? (
              <DocumentShell
                documentId={documentId}
                currentRevisionId={doc.current_revision_id}
                editorMode="readonly"
                author={currentUser?.displayName ?? ''}
              />
            ) : (
              <div className={styles.canvasEmpty}>Este documento ainda não possui conteúdo para exibir.</div>
            )}
          </main>
        </div>

        <WorkspaceSidebar
          documentId={documentId}
          doc={doc}
          instance={instance}
          instanceLoading={instanceLoading}
          instanceError={instanceError}
          onRetryInstance={() => void instanceQuery.refetch()}
          activeStage={activeStage}
          onRefetchInstance={async () => {
            await instanceQuery.refetch();
          }}
          decision={decision}
          contextualPanel={contextualPanel}
          lifecycleActions={lifecycleActions}
        />
      </div>

      {revisionDialogOpen ? (
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
              <button type="button" className={styles.dialogSecondaryBtn} onClick={() => setRevisionDialogOpen(false)}>
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

      {showCancelDialog ? (
        <CancelInstanceDialog
          documentId={documentId}
          onClose={() => setShowCancelDialog(false)}
          onSuccess={() => {
            // Cancel flips the instance → cancelled and the document → cancelled;
            // refetch both so the mode/footer/canvas recompute off the new truth.
            void instanceQuery.refetch();
            void docQuery.refetch();
          }}
        />
      ) : null}
    </div>
  );
}
