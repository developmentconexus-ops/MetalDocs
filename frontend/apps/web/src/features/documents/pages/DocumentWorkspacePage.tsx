import { useMemo, useRef, useState } from 'react';
import { Link, useParams, useSearchParams } from 'react-router-dom';
import type { MetalDocsEditorRef, EditorComment, TrackedChange } from '@metaldocs/editor-ui';

import { CodeChip, InlineAlert, StatusPill } from '../../../components/ui';
import { useAuthStore } from '../../../store/auth.store';
import { formatRevisionCode } from '../../../lib/labels/revisionCode';
import type { ArtifactAction } from '../../shared/controlled-artifact/types';
import { useHasCapability } from '../../iam/hooks/useHasCapability';
import { CancelInstanceDialog } from '../../approval/components/CancelInstanceDialog';
import { deriveWorkspaceMode } from '../../approval/lib/workspaceMode';
import { useSignoffMutation } from '../../approval/hooks/useSignoffMutation';
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
 *    delegation badge) + a signature decision seeded from `?decision=`.
 *  - reviewing / observing / author-waiting / lifecycle: unchanged S2a read
 *    canvas.
 */

// F2d.5b D1 — statuses whose official PDF the backend serves via
// GET /documents/:id/view (view_service.go viewableStatuses). Keyed on
// status, NOT on mode: 'lifecycle' also covers superseded/obsolete, which
// /view does not serve — those keep the docx read canvas.
const OFFICIAL_PDF_STATUSES = new Set(['approved', 'scheduled', 'published']);

export function DocumentWorkspacePage() {
  const { documentId = '' } = useParams<{ documentId: string }>();
  const currentUser = useAuthStore((s) => s.user);
  const [searchParams] = useSearchParams();

  const docQuery = useDocumentDetailQuery(documentId);
  const instanceQuery = useApprovalInstanceQuery(documentId);

  const doc = docQuery.data ?? null;
  const instance = instanceQuery.data ?? null;
  const viewer = instance?.viewer ?? null;

  // Mode is computed up-front (deriveWorkspaceMode accepts a possibly-null
  // doc) so every hook below can gate on it without violating the rules of
  // hooks — the loading/doc-not-found early returns further down must stay
  // the LAST thing before render, no hooks after them.
  const mode = deriveWorkspaceMode(doc, instance, viewer);
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

  const decisionParam = searchParams.get('decision');
  const defaultOptionKey: 'approve' | 'reject' | null =
    decisionParam === 'approve' || decisionParam === 'reject' ? decisionParam : null;

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
        defaultOptionKey,
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
            <ModeChip mode={mode} viewer={viewer} />
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
