import { useEffect, useRef, useState } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';

import { formatDateTime } from '../../../lib/formatDate';
import { useAuthStore } from '../../../store/auth.store';
import { useDocumentCommentsQuery } from '../../documents/queries/useDocumentCommentsQuery';
import { useDocumentDetailQuery } from '../../documents/queries/useDocumentDetailQuery';
import {
  useDocumentApprovalArtifact,
  type DocumentApprovalHandlers,
} from '../../documents/adapters/useDocumentApprovalArtifact';
import { toApprovalState } from '../../documents/lib/approvalWorkflow';
import { ArtifactApprovalScreen } from '../../shared/controlled-artifact/ArtifactApprovalScreen';
import { CancelInstanceDialog } from '../components/CancelInstanceDialog';
import { DocumentApprovalExtras } from '../components/DocumentApprovalExtras';
import { ReviewDocumentCanvas, type ReviewDocumentCanvasRef } from '../components/ReviewDocumentCanvas';
import { SignoffDialog } from '../components/SignoffDialog';
import { SupersedePublishDialog } from '../components/SupersedePublishDialog';
import { commentPlainText } from '../lib/commentPlainText';
import type { ArtifactViewModel } from '../../shared/controlled-artifact/types';
import styles from './SignoffDetailPage.module.css';

type Tab = 'documento' | 'comentarios';

const TABS: { id: Tab; label: string }[] = [
  { id: 'documento', label: 'Documento' },
  { id: 'comentarios', label: 'Comentários' },
];

/**
 * Document-specific route wrapper for the shared ArtifactApprovalScreen. Owns all
 * interactive state (dialog visibility, the review-canvas ref + flushSave, the
 * cancel prompt, the ?decision= auto-open) and injects the document review tabs as
 * the `main` slot, the integrity / lock / timeline / submit-picker as the
 * `decisionExtras` slot, and the sign-off / publish modals as the `dialogs` slot.
 * Gating + data + the action set come from useDocumentApprovalArtifact.
 */
export function SignoffDetailPage() {
  const { documentId = '' } = useParams<{ documentId: string }>();
  const [searchParams] = useSearchParams();
  const decisionParam = searchParams.get('decision');
  const initialSignoffDecision =
    decisionParam === 'approve' || decisionParam === 'reject' ? decisionParam : undefined;

  const [tab, setTab] = useState<Tab>('documento');
  const canvasRef = useRef<ReviewDocumentCanvasRef>(null);

  const [showSignoff, setShowSignoff] = useState(false);
  const [showPublish, setShowPublish] = useState(false);
  const [showSubmit, setShowSubmit] = useState(false);
  const [showCancel, setShowCancel] = useState(false);
  const [decisionError, setDecisionError] = useState<string | null>(null);
  const autoOpenedRef = useRef(false);

  const currentUser = useAuthStore((s) => s.user);

  const docQuery = useDocumentDetailQuery(documentId);
  const doc = docQuery.data ?? null;

  const commentsQuery = useDocumentCommentsQuery(documentId);

  const flushSave = async () => {
    await canvasRef.current?.flushSave();
  };

  const handlers: DocumentApprovalHandlers = {
    openSubmit: () => setShowSubmit(true),
    openSignoff: () => {
      void (async () => {
        setDecisionError(null);
        try {
          await flushSave();
        } catch {
          setDecisionError('Não foi possível salvar as alterações antes de registrar a decisão.');
          return;
        }
        setShowSignoff(true);
      })();
    },
    cancelInstance: () => setShowCancel(true),
    openPublish: () => setShowPublish(true),
  };

  const {
    model,
    instance,
    approvalState,
    contentHash,
    revisionVersion,
    lockedByInstanceId,
    publishedDocumentId,
    policy,
    error,
    contextError,
    noActiveContext,
    refetchInstance,
    isStale,
  } = useDocumentApprovalArtifact(documentId, handlers);

  // Sidebar is "loading" once the document is present but the active context has
  // not yet resolved (no error, no confirmed/absent context).
  const contextLoading = Boolean(doc) && !contextError && !noActiveContext && contentHash == null;

  // ?decision= auto-open: once the instance loads and signoff is allowed, open the
  // SignoffDialog exactly once (queue triage), preselecting the decision.
  useEffect(() => {
    if (
      initialSignoffDecision &&
      !autoOpenedRef.current &&
      instance &&
      policy.actions.signoff &&
      lockedByInstanceId
    ) {
      autoOpenedRef.current = true;
      setShowSignoff(true);
    }
  }, [initialSignoffDecision, instance, policy, lockedByInstanceId]);

  if (docQuery.isLoading) {
    return <div className={styles.state}>Carregando documento…</div>;
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
    decisionExtras = <div className={styles.state}>Carregando dados de aprovação…</div>;
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
      <>
        {decisionError ? (
          <div className={styles.state} role="alert">
            {decisionError}
          </div>
        ) : null}
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
          showSubmit={showSubmit}
          onCloseSubmit={() => setShowSubmit(false)}
        />
      </>
    );
  }

  // Suppress the action buttons until the sidebar is in its ready state so they
  // never render during loading / error / no-context (parity with the old panel,
  // which only rendered action buttons inside the loaded panel).
  const screenModel: ArtifactViewModel = sidebarReady ? model : { ...model, actions: [] };

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
              <ReviewDocumentCanvas
                ref={canvasRef}
                documentId={documentId}
                currentRevisionId={doc.current_revision_id}
                status={doc.status}
                approverDisplay={currentUser?.displayName ?? ''}
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
                      <span className={styles.commentDate}>{formatDateTime(c.created_at)}</span>
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
      {showSignoff && instance && contentHash != null ? (
        <SignoffDialog
          documentId={documentId}
          contentHash={contentHash}
          instanceId={instance.id}
          revisionVersion={revisionVersion}
          initialDecision={initialSignoffDecision}
          onClose={() => setShowSignoff(false)}
          onSuccess={() => void refetchInstance()}
        />
      ) : null}

      {showPublish && contentHash != null ? (
        <SupersedePublishDialog
          documentId={documentId}
          contentHash={contentHash}
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
