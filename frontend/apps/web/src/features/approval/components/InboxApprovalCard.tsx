import { Avatar } from '../../../components/ui/Avatar';
import { Icon } from '../../../components/ui/Icon';
import type { InboxItem } from '../api/approvalTypes';
import styles from './InboxApprovalCard.module.css';

interface InboxApprovalCardProps {
  item: InboxItem;
  onOpenDocument?: () => void;
  onApprove?: () => void;
  onReject?: () => void;
}

export function InboxApprovalCard({
  item,
  onOpenDocument,
  onApprove,
  onReject,
}: InboxApprovalCardProps) {
  return (
    <article className={styles.card}>
      <div className={styles.cardHeader}>
        <div className={styles.cardHeaderInner}>
          <div className={styles.cardHeaderMeta}>
            <div className={`${styles.codeVersion} mono`}>
              {new Date(item.submitted_at).toLocaleString('pt-BR')}
            </div>
            <h2 className={styles.cardTitle}>{item.document_title}</h2>
          </div>
        </div>
      </div>
      <div className={styles.cardBody}>
        <div className={styles.statsGrid}>
          <div className={styles.statCell}>
            <div className="kicker">AUTOR</div>
            <div className={styles.authorRow}>
              <Avatar name={item.submitted_by} size="sm" />
              <div>
                <div className={styles.authorName}>{item.submitted_by}</div>
                <div className="caption">{item.area_code}</div>
              </div>
            </div>
          </div>
          <div className={styles.statCell}>
            <div className="kicker">ÁREA</div>
            <div className={styles.stageName}>{item.area_code}</div>
            <div className="caption mono">{item.controlled_document_id}</div>
          </div>
          <div className={styles.statCell}>
            <div className="kicker">ESTÁGIO</div>
            <div className={styles.stageName}>{item.stage_label}</div>
            <div className="caption mono">{item.quorum_progress}</div>
          </div>
        </div>
        <div className={styles.cardActions}>
          <button
            type="button"
            className={`${styles.btnOpen} btn`}
            onClick={onOpenDocument}
          >
            <Icon name="eye" size={14} /> Abrir documento
          </button>
          <button
            type="button"
            className={`${styles.btnReturn} btn`}
            onClick={onReject}
          >
            Devolver
          </button>
          <button
            type="button"
            className={`${styles.btnApprove} btn btn-primary`}
            onClick={onApprove}
          >
            Aprovar e assinar →
          </button>
        </div>
      </div>
    </article>
  );
}
