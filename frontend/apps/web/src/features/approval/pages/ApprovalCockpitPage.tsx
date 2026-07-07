import { useRef, useState } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import type { TrackedChange } from '@metaldocs/editor-ui';

import { formatSignedAt } from '../../../lib/format/dates';
import { useAuthStore } from '../../../store/auth.store';
import { useDocumentCommentsQuery } from '../../documents/queries/useDocumentCommentsQuery';
import { useDocumentDetailQuery } from '../../documents/queries/useDocumentDetailQuery';
import { useControlledDocumentActiveDocumentQuery } from '../../documents/queries/useControlledDocumentActiveDocumentQuery';
import {
  useDocumentApprovalArtifact,
  type DocumentApprovalHandlers,
} from '../../documents/adapters/useDocumentApprovalArtifact';
import type { StageInstance } from '../api/approvalTypes';
import { toApprovalState } from '../../documents/lib/approvalWorkflow';
import { ArtifactApprovalScreen } from '../../shared/controlled-artifact/ArtifactApprovalScreen';
import { CancelInstanceDialog } from '../components/CancelInstanceDialog';
import { DocumentApprovalExtras } from '../components/DocumentApprovalExtras';
import { DocumentShell } from '../../documents/components/DocumentShell';
import { SupersedePublishDialog } from '../components/SupersedePublishDialog';
import { useSignoffMutation } from '../hooks/useSignoffMutation';
import { commentPlainText } from '../lib/commentPlainText';
import type { ArtifactViewModel } from '../../shared/controlled-artifact/types';
import styles from './SignoffDetailPage.module.css';

type Tab = 'documento' | 'comentarios';

const TABS: { id: Tab; label: string }[] = [
  { id: 'documento', label: 'Documento' },
  { id: 'comentarios', label: 'Comentários' },
];

/**
 * Resolve the cockpit editor mode from the active approval stage (F3, C1/W2).
 *
 * `'review'` iff the active stage is a review stage AND the current user is an
 * eligible actor on it (present in `actors` with status active/waiting).
 * Everything else — approval stage, non-eligible actor, oversee observer, no
 * active stage — is `'readonly'` (fail-safe default: no writable affordance
 * without positive eligibility, per the spec's mode-resolution table).
 */
export function resolveEditorMode(
  activeStage: StageInstance | undefined,
  currentUserId: string | null | undefined,
): 'review' | 'readonly' {
  if (!activeStage || activeStage.stage_kind !== 'review') {
    return 'readonly';
  }
  if (!currentUserId) {
    return 'readonly';
  }
  const eligible = activeStage.actors.some(
    (actor) => actor.user_id === currentUserId && (actor.status === 'active' || actor.status === 'waiting'),
  );
  return eligible ? 'review' : 'readonly';
}

/**
 * Document-specific route wrapper for the shared ArtifactApprovalScreen. Composes
 * the shared DecisionModel for the inline sign-off (password re-auth + legal-effect
 * confirmation — the document's legal e-signature), owns the remaining interactive
 * state (publish / cancel dialogs) and injects the document review tabs as the
 * `main` slot, the integrity / lock / timeline as the `decisionExtras` slot, and
 * the publish / cancel modals as the `dialogs` slot.
 *
 * F3 (C1/W2): the "Documento" tab mounts `DocumentShell` — the SAME editor-canvas
 * region the author page uses — in a resolved mode (`readonly` for approval
 * stages / non-eligible actors, `review` for eligible review-stage actors). The
 * cockpit mounts NO writer session and NO autosave: DocumentShell is passed no
 * `onAutoSave`, so a reviewer's edits never persist as a document revision (the
 * W2 fix). Suggestions are surfaced via `onTrackedChangesChange` (wired here,
 * rendered by F4); comments still persist through the comments API below.
 *
 * The cockpit is approver-only — submitting a document for review happens
 * exclusively on the document editor, so this route has no submit-picker. The
 * `?decision=` param preselects the sign-off option via `defaultOptionKey`. Gating
 * + data + the action set come from useDocumentApprovalArtifact.
 */
export function ApprovalCockpitPage() {
  const { documentId = '' } = useParams<{ documentId: string }>();
  const [searchParams] = useSearchParams();
  const decisionParam = searchParams.get('decision');
  const initialSignoffDecision =
    decisionParam === 'approve' || decisionParam === 'reject' ? decisionParam : undefined;

  const [tab, setTab] = useState<Tab>('documento');
  // F4 surface: client-side tracked-change suggestions from review mode. Wired
  // through DocumentShell here, not yet rendered (SuggestionList lands in F4).
  const [trackedChanges, setTrackedChanges] = useState<TrackedChange[]>([]);
  // Bridges the one-hop ordering cycle: decisionSubmit (built before the adapter
  // call) needs refetchInstance (returned BY the adapter call). See below.
  const refetchInstanceRef = useRef<() => Promise<void>>(async () => {});

  const [showPublish, setShowPublish] = useState(false);
  const [showCancel, setShowCancel] = useState(false);

  const currentUser = useAuthStore((s) => s.user);

  const docQuery = useDocumentDetailQuery(documentId);
  const doc = docQuery.data ?? null;

  const commentsQuery = useDocumentCommentsQuery(documentId);

  // Signing routes exclusively through the DecisionPanel now — the adapter emits no
  // 'signoff' action, so there is no handler for it.
  const handlers: DocumentApprovalHandlers = {
    cancelInstance: () => setShowCancel(true),
    openPublish: () => setShowPublish(true),
  };

  // The signoff mutation needs contentHash/revisionVersion BEFORE the adapter call
  // (the adapter now takes the sign submit callback as an input — FE-02), so this
  // route reads the active-document context directly here. Same query key as the
  // one `useDocumentApprovalArtifact` reads internally; react-query deduplicates,
  // so this adds no extra network request.
  const contextQuery = useControlledDocumentActiveDocumentQuery(doc?.controlled_document_id ?? '');
  const contentHash = contextQuery.data?.content_hash ?? null;
  const revisionVersion = contextQuery.data?.revision_version ?? doc?.revision_version ?? 0;

  const { signOff } = useSignoffMutation({
    documentId,
    contentHash: contentHash ?? '',
    revisionVersion,
  });

  // No writable canvas to flush — DocumentShell in review mode carries no
  // autosave (W2 fix). decisionSubmit is now signOff -> refetchInstance.
  const decisionSubmit = async (input: { optionKey: string; reason: string; password: string }) => {
    await signOff({
      decision: input.optionKey === 'approve' ? 'approve' : 'reject',
      reason: input.reason || undefined,
      password: input.password,
    });
    await refetchInstanceRef.current();
  };

  const {
    model,
    instance,
    approvalState,
    lockedByInstanceId,
    publishedDocumentId,
    policy,
    error,
    contextError,
    noActiveContext,
    refetchInstance,
    isStale,
  } = useDocumentApprovalArtifact(documentId, handlers, {
    defaultOptionKey: initialSignoffDecision ?? null,
    submit: decisionSubmit,
  });

  // `decisionSubmit` needs `refetchInstance`, which only exists after the adapter
  // call above — a ref bridges this one-hop ordering cycle without re-deriving
  // any adapter-owned state in the route.
  refetchInstanceRef.current = refetchInstance;

  // Sidebar is "loading" once the document is present but the active context has
  // not yet resolved (no error, no confirmed/absent context).
  const contextLoading = Boolean(doc) && !contextError && !noActiveContext && contentHash == null;

  if (docQuery.isLoading) {
    return <div className={styles.state} role="status" aria-live="polite">Carregando documento…</div>;
  }
  if (docQuery.isError || !doc || !model) {
    return (
      <div className={styles.state} role="alert">
        Não foi possível carregar este documento.
      </div>
    );
  }

  // The sidebar (decisionExtras) renders the real approval panel ONLY when the
  // active context is confirmed; loading / error / no-context / instance-error
  // states render here and suppress the action buttons (empty actions).
  const sidebarReady = !contextLoading && !contextError && !noActiveContext && !error && contentHash != null;

  let decisionExtras: React.ReactNode;
  if (contextLoading) {
    decisionExtras = <div className={styles.state} role="status" aria-live="polite">Carregando dados de aprovação…</div>;
  } else if (contextError) {
    decisionExtras = (
      <div className={styles.state} role="alert">
        Não foi possível carregar os dados de aprovação.
      </div>
    );
  } else if (error) {
    decisionExtras = (
      <div className={styles.state} role="alert">
        <p>{error}</p>
        <button type="button" className={styles.retry} onClick={() => void refetchInstance()}>
          Tentar novamente
        </button>
      </div>
    );
  } else if (noActiveContext || contentHash == null) {
    decisionExtras = (
      <div className={styles.state}>Este documento não está em um fluxo de aprovação ativo.</div>
    );
  } else {
    decisionExtras = (
      <DocumentApprovalExtras
        documentId={documentId}
        status={toApprovalState(approvalState)}
        policy={policy}
        contentHash={contentHash}
        revisionVersion={revisionVersion}
        lockedByInstanceId={lockedByInstanceId}
        instance={instance}
        isStale={isStale}
        onRefetchInstance={refetchInstance}
      />
    );
  }

  // The document sign-off decision model (password re-auth + legal-effect
  // confirmation) is now constructed by the adapter via `buildDocumentSignoffDecision`
  // (FE-02) — this route only supplied the `submit` sequencing above (signOff →
  // refetchInstance) since that depends on route-owned state (the lifted signoff
  // mutation).
  const decision = model.decision;

  // Suppress the action buttons until the sidebar is ready. Signing is not among
  // model.actions (it routes through the DecisionPanel), so no filtering is needed.
  const baseActions = sidebarReady ? model.actions : [];
  const screenModel: ArtifactViewModel = { ...model, actions: baseActions, decision };

  const activeStage = instance?.stages.find((s) => s.status === 'active');
  const editorMode = resolveEditorMode(activeStage, currentUser?.userId ?? null);

  const main = (
    <div className={styles.main}>
      <header className={styles.header}>
        <h1 className={styles.title}>{doc.name}</h1>
      </header>

      <nav className={styles.tabs} role="tablist" aria-label="Seções do documento">
        {TABS.map((t) => (
          <button
            key={t.id}
            type="button"
            role="tab"
            id={`tab-${t.id}`}
            aria-controls={`panel-${t.id}`}
            aria-selected={tab === t.id}
            className={`${styles.tab} ${tab === t.id ? styles.tabActive : ''}`}
            onClick={() => setTab(t.id)}
          >
            {t.label}
          </button>
        ))}
      </nav>

      <div
        className={styles.body}
        role="tabpanel"
        id={`panel-${tab}`}
        aria-labelledby={`tab-${tab}`}
        tabIndex={0}
      >
        {tab === 'documento' ? (
          <div className={styles.a4}>
            {doc.current_revision_id ? (
              <DocumentShell
                documentId={documentId}
                currentRevisionId={doc.current_revision_id}
                editorMode={editorMode}
                author={currentUser?.displayName ?? ''}
                onTrackedChangesChange={setTrackedChanges}
              />
            ) : (
              <div className={styles.a4State}>Este documento ainda não possui conteúdo para revisão.</div>
            )}
          </div>
        ) : null}

        {tab === 'comentarios' ? (
          <div className={styles.comments}>
            {commentsQuery.isLoading ? (
              <p className={styles.commentsState}>Carregando comentários…</p>
            ) : commentsQuery.isError ? (
              <p className={styles.commentsState} role="alert">
                Não foi possível carregar os comentários.
              </p>
            ) : (commentsQuery.data ?? []).length === 0 ? (
              <p className={styles.commentsState}>Nenhum comentário neste documento.</p>
            ) : (
              <ul className={styles.commentList}>
                {(commentsQuery.data ?? []).map((c) => (
                  <li key={c.id} className={styles.comment}>
                    <div className={styles.commentHead}>
                      <strong>{c.author}</strong>
                      <span className={styles.commentDate}>{formatSignedAt(c.created_at)}</span>
                      {c.done ? <span className={styles.resolved}>Resolvido</span> : null}
                    </div>
                    <p className={styles.commentBody}>{commentPlainText(c.content)}</p>
                  </li>
                ))}
              </ul>
            )}
          </div>
        ) : null}
      </div>
    </div>
  );

  const dialogs = (
    <>
      {showPublish && contentHash != null ? (
        <SupersedePublishDialog
          documentId={documentId}
          publishedDocumentId={publishedDocumentId}
          onClose={() => setShowPublish(false)}
          onSuccess={() => void refetchInstance()}
        />
      ) : null}

      {showCancel ? (
        <CancelInstanceDialog
          documentId={documentId}
          onClose={() => setShowCancel(false)}
          onSuccess={() => void refetchInstance()}
        />
      ) : null}
    </>
  );

  return (
    <ArtifactApprovalScreen
      model={screenModel}
      main={main}
      decisionExtras={decisionExtras}
      dialogs={dialogs}
    />
  );
}
