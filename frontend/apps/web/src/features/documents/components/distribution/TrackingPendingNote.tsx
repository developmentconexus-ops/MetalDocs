import styles from './TrackingPendingNote.module.css';

/**
 * Honest disclosure for numerator surfaces (read %, ack %, timeline, per-recipient
 * read/ack status). Renders NO fabricated numbers — the backend producer for these
 * surfaces is not yet available.
 *
 * Planned: wiki/backlog/document-distribution-mission.md
 */
export function TrackingPendingNote() {
  return (
    <div className={styles.note} role="note" aria-label="Acompanhamento pendente">
      <p className={styles.heading}>
        Acompanhamento de leitura/ciência ainda não disponível
      </p>
      <p className={styles.body}>
        O registro de leitura e reconhecimento por destinatário ainda não está disponível
        no backend. Esta funcionalidade está planejada na missão de distribuição de
        documentos.
      </p>
    </div>
  );
}
