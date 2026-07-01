import { StatusPill, type DocumentStatus } from '../../../components/ui';
import styles from './TemplateApprovalExtras.module.css';

interface TemplateApprovalExtrasProps {
  status: DocumentStatus | null;
  showReason: boolean;
  reason: string;
  onReasonChange: (value: string) => void;
  busy: boolean;
  message: { kind: 'success' | 'error'; text: string } | null;
}

/**
 * Decision-sidebar extras rendered below the shared action buttons for the
 * template approval screen. Surfaces the current version status, an optional
 * reason textarea (for review/approve actions), and the mutation result message.
 */
export function TemplateApprovalExtras({
  status,
  showReason,
  reason,
  onReasonChange,
  busy,
  message,
}: TemplateApprovalExtrasProps) {
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
      ) : showReason ? (
        <div className={styles.reasonField}>
          <label htmlFor="template-approval-reason" className={styles.reasonLabel}>
            Motivo (opcional)
          </label>
          <textarea
            id="template-approval-reason"
            className={styles.reasonTextarea}
            value={reason}
            onChange={(e) => onReasonChange(e.target.value)}
            rows={3}
            disabled={busy}
          />
        </div>
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
