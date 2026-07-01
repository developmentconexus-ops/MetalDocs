import { useRef, useState } from 'react';
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
import { SupersedePublishDialog } from '../components/SupersedePublishDialog';
import { useSignoffMutation } from '../hooks/useSignoffMutation';
import { commentPlainText } from '../lib/commentPlainText';
import type { ArtifactDecisionModel, ArtifactViewModel } from '../../shared/controlled-artifact/types';
import styles from './SignoffDetailPage.module.css';

type Tab = 'documento' | 'comentarios';

const TABS: { id: Tab; label: string }[] = [
  { id: 'documento', label: 'Documento' },
  { id: 'comentarios', label: 'Comentários' },
];

/**
 * Document-specific route wrapper for the shared ArtifactApprovalScreen. Composes
 * the shared DecisionModel for the inline sign-off (password re-auth + legal-effect
 * confirmation — the document's legal e-signature), owns the remaining interactive
 * state (publish / cancel / submit-picker dialogs, the review-canvas ref + flushSave)
 * and injects the document review tabs as the `main` slot, the integrity / lock /
 * timeline / submit-picker as the `decisionExtras` slot, and the publish / cancel
 * modals as the `dialogs` slot. The `?decision=` param preselects the sign-off
 * option via `defaultOptionKey`. Gating + data + the action set come from
 * useDocumentApprovalArtifact.
 */
export function SignoffDetailPage() {
  const { documentId = '' } = useParams<{ documentId: string }>();
  const [searchParams] = useSearchParams();
  const decisionParam = searchParams.get('decision');
  const initialSignoffDecision =
    decisionParam === 'approve' || decisionParam === 'reject' ? decisionParam : undefined;

  const [tab, setTab] = useState<Tab>('documento');
  const canvasRef = useRef<ReviewDocumentCanvasRef>(null);

  const [showPublish, setShowPublish] = useState(false);
  const [showSubmit, setShowSubmit] = useState(false);
  const [showCancel, setShowCancel] = useState(false);

  const currentUser = useAuthStore((s) => s.user);

  const docQuery = useDocumentDetailQuery(documentId);
  const doc = docQuery.data ?? null;

  const commentsQuery = useDocumentCommentsQuery(documentId);

  const flushSave = async () => {
    await canvasRef.current?.flushSave();
  };

  // Signing routes exclusively through the DecisionPanel now — the adapter emits no
  // 'signoff' action, so there is no handler for it.
  const handlers: DocumentApprovalHandlers = {
    openSubmit: () => setShowSubmit(true),
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

  const { signOff } = useSignoffMutation({
    documentId,
    contentHash: contentHash ?? '',
    revisionVersion,
  });

  // Sidebar is "loading" once the document is present but the active context has
  // not yet resolved (no error, no confirmed/absent context).
  const contextLoading = Boolean(doc) && !contextError && !noActiveContext && contentHash == null;

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
    );
  }

  // The document sign-off carries legal e-signature weight: a password re-auth + a
  // legal-effect confirmation. Offered only when the active context is confirmed and
  // policy allows signing on a locked instance. The panel owns password/reason state.
  const signoffOffered =
    sidebarReady && policy.actions.signoff && lockedByInstanceId != null && instance != null && contentHash != null;

  const decision: ArtifactDecisionModel | undefined = signoffOffered
    ? {
        kicker: 'Assinatura requerida',
        heading: 'Registrar decisão',
        description:
          'Sua decisão é registrada com efeito de assinatura eletrônica na trilha de auditoria do documento.',
        options: [
          {
            key: 'approve',
            label: 'Assinar e aprovar',
            description: 'Aprova o documento e avança o fluxo de aprovação.',
            tone: 'approve',
            submitLabel: 'Assinar e aprovar',
            requiresReason: false,
          },
          {
            key: 'reject',
            label: 'Assinar e devolver',
            description: 'Devolve o documento ao autor · requer justificativa.',
            tone: 'reject',
            submitLabel: 'Assinar e devolver',
            requiresReason: true,
          },
        ],
        reasonLabel: 'Justificativa',
        reasonPlaceholder: 'Comentário visível na trilha de auditoria…',
        password: { label: 'Senha' },
        legal: {
          text: 'Confirmo que revisei o conteúdo integralmente e que esta decisão tem efeito de assinatura eletrônica conforme a MP 2.200-2/2001.',
        },
        signer: currentUser
          ? {
              name: currentUser.displayName,
              detail: currentUser.email ?? currentUser.username,
              note: 'Assinatura digital gerada no ato da confirmação.',
            }
          : null,
        defaultOptionKey: initialSignoffDecision ?? null,
        submit: async ({ optionKey, reason, password }) => {
          try {
            await flushSave();
          } catch {
            throw new Error('Não foi possível salvar as alterações antes de registrar a decisão.');
          }
          await signOff({
            decision: optionKey === 'approve' ? 'approve' : 'reject',
            reason: reason || undefined,
            password,
          });
          await refetchInstance();
        },
      }
    : undefined;

  // Suppress the action buttons until the sidebar is ready. Signing is not among
  // model.actions (it routes through the DecisionPanel), so no filtering is needed.
  const baseActions = sidebarReady ? model.actions : [];
  const screenModel: ArtifactViewModel = { ...model, actions: baseActions, decision };

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
