import type { TrackedChange, TrackedChangeType } from '@metaldocs/editor-ui';

import { formatSignedAt } from '../../../../lib/format/dates';
import { useDocumentCommentsQuery } from '../../../documents/queries/useDocumentCommentsQuery';
import { commentPlainText } from '../../lib/commentPlainText';
import styles from './SuggestionList.module.css';

interface SuggestionListProps {
  documentId: string;
  trackedChanges: TrackedChange[];
}

const TYPE_LABEL: Record<TrackedChangeType, string> = {
  insertion: 'Inserção',
  deletion: 'Remoção',
  replacement: 'Substituição',
  paragraphMarkInsertion: 'Novo parágrafo',
  paragraphMarkDeletion: 'Remoção de parágrafo',
  paragraphPropertiesChanged: 'Formatação de parágrafo',
  rowInserted: 'Linha inserida',
  rowDeleted: 'Linha removida',
  rowPropertiesChanged: 'Formatação de linha',
  cellInserted: 'Célula inserida',
  cellDeleted: 'Célula removida',
} as unknown as Record<TrackedChangeType, string>;

/** Review-mode-only sidebar section (spec §1.4): client-side tracked-change
 *  suggestion cards, followed by the document's persisted comments. */
export function SuggestionList({ documentId, trackedChanges }: SuggestionListProps) {
  const commentsQuery = useDocumentCommentsQuery(documentId);
  const comments = commentsQuery.data ?? [];

  return (
    <section className={styles.section} aria-label="Sugestões e comentários">
      <div className={styles.group}>
        <h3 className={styles.heading}>Sugestões</h3>
        {trackedChanges.length === 0 ? (
          <p className={styles.empty}>Nenhuma sugestão registrada nesta revisão.</p>
        ) : (
          <ul className={styles.list}>
            {trackedChanges.map((change) => (
              <li className={styles.card} key={change.revisionId}>
                <div className={styles.cardHead}>
                  <strong>{change.author}</strong>
                  <span className={styles.tag}>{TYPE_LABEL[change.type] ?? change.type}</span>
                </div>
                <p className={styles.excerpt}>{change.excerpt}</p>
              </li>
            ))}
          </ul>
        )}
      </div>

      <div className={styles.group}>
        <h3 className={styles.heading}>Comentários</h3>
        {commentsQuery.isLoading ? (
          <p className={styles.empty}>Carregando comentários…</p>
        ) : commentsQuery.isError ? (
          <p className={styles.empty} role="alert">
            Não foi possível carregar os comentários.
          </p>
        ) : comments.length === 0 ? (
          <p className={styles.empty}>Nenhum comentário neste documento.</p>
        ) : (
          <ul className={styles.list}>
            {comments.map((c) => (
              <li className={styles.card} key={c.id}>
                <div className={styles.cardHead}>
                  <strong>{c.author}</strong>
                  <span className={styles.date}>{formatSignedAt(c.created_at)}</span>
                  {c.done ? <span className={styles.tag}>Resolvido</span> : null}
                </div>
                <p className={styles.excerpt}>{commentPlainText(c.content)}</p>
              </li>
            ))}
          </ul>
        )}
      </div>
    </section>
  );
}
