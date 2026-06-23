import { useState } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';

import { CodeChip, StatusPill } from '../../../components/ui';
import { formatDateTime } from '../../../lib/formatDate';
import { useDocumentCommentsQuery } from '../../documents/queries/useDocumentCommentsQuery';
import { useDocumentDetailQuery } from '../../documents/queries/useDocumentDetailQuery';
import { useDocumentPdfStatus } from '../../documents/hooks/editor/useDocumentPdfStatus';
import { ControlledDocumentDetailPanel } from '../components/ControlledDocumentDetailPanel';
import { commentPlainText } from '../lib/commentPlainText';
import { useActiveDocumentContextQuery } from '../queries/useActiveDocumentContextQuery';
import styles from './SignoffDetailPage.module.css';

type Tab = 'documento' | 'mudancas' | 'comentarios';

const TABS: { id: Tab; label: string }[] = [
  { id: 'documento', label: 'Documento' },
  { id: 'mudancas', label: 'Mudanças vs versão anterior' },
  { id: 'comentarios', label: 'Comentários' },
];

export function SignoffDetailPage() {
  const { documentId = '' } = useParams<{ documentId: string }>();
  const [searchParams] = useSearchParams();
  const decisionParam = searchParams.get('decision');
  const initialSignoffDecision =
    decisionParam === 'approve' || decisionParam === 'reject' ? decisionParam : undefined;

  const [tab, setTab] = useState<Tab>('documento');

  const docQuery = useDocumentDetailQuery(documentId);
  const doc = docQuery.data ?? null;
  const controlledDocumentId = doc?.controlled_document_id ?? '';

  const contextQuery = useActiveDocumentContextQuery(controlledDocumentId);
  const context = contextQuery.data ?? null;

  const pdf = useDocumentPdfStatus(documentId, tab === 'documento');
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
              aria-selected={tab === t.id}
              className={`${styles.tab} ${tab === t.id ? styles.tabActive : ''}`}
              onClick={() => setTab(t.id)}
            >
              {t.label}
            </button>
          ))}
        </nav>

        <div className={styles.body}>
          {tab === 'documento' ? (
            <div className={styles.a4}>
              {pdf.status === 'ready' && pdf.url ? (
                <iframe
                  title="Pré-visualização do documento"
                  src={pdf.url}
                  className={styles.pdfFrame}
                />
              ) : pdf.status === 'failed' ? (
                <div className={styles.a4State} role="alert">
                  <p>Falha ao gerar a pré-visualização.</p>
                  <button type="button" className={styles.retry} onClick={pdf.retry}>
                    Tentar novamente
                  </button>
                </div>
              ) : (
                <div className={styles.a4State}>Gerando visualização do documento…</div>
              )}
            </div>
          ) : null}

          {tab === 'mudancas' ? (
            <div className={styles.deferred}>
              <h2>Comparação de versões</h2>
              <p>
                A comparação visual entre revisões depende de um endpoint de diff de documentos que
                ainda não existe. Esta aba será habilitada quando o backend expuser o diff — veja
                o backlog (<code>wiki/backlog/detalhe-signoff.md</code>). Nenhum diff é exibido aqui
                para não apresentar dados inventados.
              </p>
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
        ) : context && context.content_hash ? (
          <ControlledDocumentDetailPanel
            documentId={documentId}
            approvalState={context.approval_state ?? doc.status}
            contentHash={context.content_hash}
            revisionVersion={context.revision_version ?? doc.revision_version}
            lockedByInstanceId={context.approval_instance_id}
            publishedDocumentId={context.published_document_id}
            autoOpenSignoff={Boolean(initialSignoffDecision)}
            initialSignoffDecision={initialSignoffDecision}
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
