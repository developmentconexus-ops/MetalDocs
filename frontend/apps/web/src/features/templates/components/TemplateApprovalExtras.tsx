import { StatusPill, type DocumentStatus } from '../../../components/ui';
import styles from './TemplateApprovalExtras.module.css';

interface TemplateApprovalExtrasProps {
  status: DocumentStatus | null;
  message: { kind: 'success' | 'error'; text: string } | null;
}

/**
 * Decision-sidebar extras rendered below the shared action buttons for the
 * template approval screen. Surfaces the current version status and the
 * fire-and-forget mutation result message. Only mounted when NO decision is
 * offered (draft submit / published read-only); when a decision is active the
 * shared DecisionPanel owns the motivo, so no reason field lives here.
 */
export function TemplateApprovalExtras({ status, message }: TemplateApprovalExtrasProps) {
  return (
    <section className={styles.section}>
      {status != null ? (
        <div className={styles.statusRow}>
          <span className={styles.statusLabel}>Status</span>
          <StatusPill status={status} />
        </div>
      ) : null}

      {status === 'published' ? (
        <p className={styles.publishedNote}>Esta versão está publicada.</p>
      ) : null}

      {message != null ? (
        message.kind === 'error' ? (
          <p role="alert" className={styles.messageError}>{message.text}</p>
        ) : (
          <p role="status" className={styles.messageSuccess}>{message.text}</p>
        )
      ) : null}
    </section>
  );
}
