import type { TrackedChange, EditorComment } from '@metaldocs/editor-ui';
import styles from './RequestedChangesPanel.module.css';

const TRACKED_CHANGE_TYPE_LABELS: Record<string, string> = {
  insertion: 'Inserção',
  deletion: 'Exclusão',
  replacement: 'Substituição',
  paragraphMarkInsertion: 'Inserção de parágrafo',
  paragraphMarkDeletion: 'Exclusão de parágrafo',
  paragraphPropertiesChanged: 'Formatação de parágrafo',
  rowInserted: 'Linha inserida',
  rowDeleted: 'Linha excluída',
  rowPropertiesChanged: 'Formatação de linha',
  cellInserted: 'Célula inserida',
  cellDeleted: 'Célula excluída',
  cellMerged: 'Células mescladas',
  cellPropertiesChanged: 'Formatação de célula',
  tableInserted: 'Tabela inserida',
  tableDeleted: 'Tabela excluída',
  tablePropertiesChanged: 'Formatação de tabela',
};

function trackedChangeTypeLabel(type: string): string {
  return TRACKED_CHANGE_TYPE_LABELS[type] ?? type;
}

export type RequestedChangesPanelProps = {
  trackedChanges: TrackedChange[];
  comments: EditorComment[];
  onAcceptChange: (revisionId: string) => void;
  onRejectChange: (revisionId: string) => void;
  onResolveComment: (comment: EditorComment) => void;
};

/**
 * RequestedChangesPanel — F6/C5. Page-owned (mounted by DocumentEditorPage,
 * NOT a DocumentShell slot). Pure presentational: lists live tracked changes
 * (accept/reject per change) and unresolved comments (resolve per comment).
 * Actions are delegated entirely to the caller via callbacks; this component
 * never touches editorRef or comment mutations directly.
 */
export function RequestedChangesPanel({
  trackedChanges,
  comments,
  onAcceptChange,
  onRejectChange,
  onResolveComment,
}: RequestedChangesPanelProps): React.ReactElement {
  const unresolvedComments = comments.filter((c) => !c.resolved);
  const totalPending = trackedChanges.length + unresolvedComments.length;

  return (
    <aside className={styles.root} aria-label="Mudanças solicitadas">
      <div className={styles.header}>
        <span className={styles.title}>Mudanças solicitadas</span>
        <p className={styles.count}>
          {totalPending === 0
            ? 'Tudo resolvido'
            : `${totalPending} ${totalPending === 1 ? 'pendência' : 'pendências'}`}
        </p>
      </div>

      {totalPending === 0 ? (
        <p className={styles.emptyState}>Nenhuma marcação pendente. Você pode reenviar.</p>
      ) : (
        <>
          {trackedChanges.length > 0 ? (
            <section className={styles.section}>
              <h3 className={styles.sectionLabel}>Marcações de alteração</h3>
              <ul className={styles.list}>
                {trackedChanges.map((change) => (
                  <li key={change.revisionId} className={styles.card}>
                    <div className={styles.cardMeta}>
                      <span className={styles.cardAuthor}>{change.author}</span>
                      <span className={styles.cardType}>{trackedChangeTypeLabel(change.type)}</span>
                    </div>
                    <p className={styles.cardExcerpt}>{change.excerpt}</p>
                    <div className={styles.cardActions}>
                      <button
                        type="button"
                        className={`${styles.actionBtn} ${styles.acceptBtn}`}
                        onClick={() => onAcceptChange(change.revisionId)}
                      >
                        Aceitar
                      </button>
                      <button
                        type="button"
                        className={`${styles.actionBtn} ${styles.rejectBtn}`}
                        onClick={() => onRejectChange(change.revisionId)}
                      >
                        Rejeitar
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            </section>
          ) : null}

          {unresolvedComments.length > 0 ? (
            <section className={styles.section}>
              <h3 className={styles.sectionLabel}>Comentários não resolvidos</h3>
              <ul className={styles.list}>
                {unresolvedComments.map((comment) => (
                  <li key={comment.id} className={styles.card}>
                    <div className={styles.cardMeta}>
                      <span className={styles.cardAuthor}>{comment.author}</span>
                    </div>
                    <div className={styles.cardActions}>
                      <button
                        type="button"
                        className={styles.actionBtn}
                        onClick={() => onResolveComment(comment)}
                      >
                        Resolver
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            </section>
          ) : null}

          <p className={styles.gateNote}>
            Resolva todas as marcações e comentários antes de reenviar.
          </p>
        </>
      )}
    </aside>
  );
}
