import { useRef, useState } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';

import { CodeChip, StatusPill } from '../../../components/ui';
import { formatDateTime } from '../../../lib/formatDate';
import { useDocumentCommentsQuery } from '../../documents/queries/useDocumentCommentsQuery';
import { useDocumentDetailQuery } from '../../documents/queries/useDocumentDetailQuery';
import { ControlledDocumentDetailPanel } from '../components/ControlledDocumentDetailPanel';
import { ReviewDocumentCanvas, type ReviewDocumentCanvasRef } from '../components/ReviewDocumentCanvas';
import { commentPlainText } from '../lib/commentPlainText';
import { useActiveDocumentContextQuery } from '../queries/useActiveDocumentContextQuery';
import styles from './SignoffDetailPage.module.css';

type Tab = 'documento' | 'comentarios';

const TABS: { id: Tab; label: string }[] = [
  { id: 'documento', label: 'Documento' },
  { id: 'comentarios', label: 'Comentários' },
];

export function SignoffDetailPage() {
  const { documentId = '' } = useParams<{ documentId: string }>();
  const [searchParams] = useSearchParams();
  const decisionParam = searchParams.get('decision');
  const initialSignoffDecision =
    decisionParam === 'approve' || decisionParam === 'reject' ? decisionParam : undefined;

  const [tab, setTab] = useState<Tab>('documento');
  const canvasRef = useRef<ReviewDocumentCanvasRef>(null);

  const docQuery = useDocumentDetailQuery(documentId);
  const doc = docQuery.data ?? null;
  const controlledDocumentId = doc?.controlled_document_id ?? '';

  const contextQuery = useActiveDocumentContextQuery(controlledDocumentId);
  const context = contextQuery.data ?? null;

  const commentsQuery = useDocumentCommentsQuery(documentId);

  if (docQuery.isLoading) {
    return <div className={styles.state}>Carregando documento…</div>;
  }
  if (docQuery.isError || !doc) {
    return (
      <div className={styles.state} role="alert">
        Não foi possível carregar este documento.
      </div>
    );
  }

  return (
    <div className={styles.page}>
      <div className={styles.main}>
        <header className={styles.header}>
          <div className={styles.headerTop}>
            <CodeChip>{doc.code}</CodeChip>
            <StatusPill status={doc.status} />
            <span className={styles.version}>v{doc.revision_version}</span>
          </div>
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
                  approverDisplay={String(doc.created_by ?? '')}
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

      <aside className={styles.sidebar} aria-label="Decisão de aprovação">
        {contextQuery.isLoading ? (
          <div className={styles.state}>Carregando dados de aprovação…</div>
        ) : contextQuery.isError ? (
          <div className={styles.state} role="alert">
            Não foi possível carregar os dados de aprovação.
          </div>
        ) : context && context.content_hash != null ? (
          <ControlledDocumentDetailPanel
            documentId={documentId}
            approvalState={context.approval_state ?? doc.status}
            contentHash={context.content_hash}
            revisionVersion={context.revision_version ?? doc.revision_version}
            lockedByInstanceId={context.approval_instance_id}
            publishedDocumentId={context.published_document_id}
            autoOpenSignoff={Boolean(initialSignoffDecision)}
            initialSignoffDecision={initialSignoffDecision}
            beforeDecision={async () => { await canvasRef.current?.flushSave(); }}
          />
        ) : (
          <div className={styles.state}>
            Este documento não está em um fluxo de aprovação ativo.
          </div>
        )}
      </aside>
    </div>
  );
}
