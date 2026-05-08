import { Avatar } from '../../../components/ui/Avatar';
import { Icon } from '../../../components/ui/Icon';
import type { RichInboxItem } from '../lib/mockInboxData';
import styles from './InboxApprovalCard.module.css';

interface InboxApprovalCardProps {
  item: RichInboxItem;
}

export function InboxApprovalCard({ item }: InboxApprovalCardProps) {
  return (
    <article className={styles.card}>
      <div className={`${styles.cardHeader}${item.urgent ? ` ${styles.cardHeaderUrgent}` : ''}`}>
        <div className={styles.cardHeaderInner}>
          <div className={styles.kindBadge}>{item.kind}</div>
          <div className={styles.cardHeaderMeta}>
            <div className={`${styles.codeVersion} mono`}>{item.code} · {item.version}</div>
            <h2 className={styles.cardTitle}>{item.document_title}</h2>
          </div>
          <div className={styles.deadlineBlock}>
            <div className={styles.deadlineLabel}>VENCE EM</div>
            <div className={styles.deadlineValue}>{item.deadline}</div>
          </div>
        </div>
      </div>
      <div className={styles.cardBody}>
        <p className={styles.summary}>{item.summary}</p>
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
            <div className="kicker">ALTERAÇÕES</div>
            <div className={styles.changesRow}>
              <span className={styles.changesNum}>{item.changes}</span>
              <span className="caption">edições</span>
            </div>
          </div>
          <div className={styles.statCell}>
            <div className="kicker">ESTÁGIO</div>
            <div className={styles.stageName}>{item.stage_label}</div>
            <div className="caption">de 4 etapas</div>
          </div>
        </div>
        <div className={styles.cardActions}>
          <button
            type="button"
            className={`${styles.btnOpen} btn`}
            onClick={() => { /* TODO [BACKLOG: caixa-aprovacao.md]: navigate to doc */ }}
          >
            <Icon name="docs" size={14} /> Abrir documento
          </button>
          <button
            type="button"
            className={`${styles.btnReturn} btn`}
            onClick={() => { /* TODO [BACKLOG: caixa-aprovacao.md]: return flow */ }}
          >
            Devolver
          </button>
          <button
            type="button"
            className={`${styles.btnApprove} btn btn-primary`}
            onClick={() => { /* TODO [BACKLOG: caixa-aprovacao.md]: signoff flow */ }}
          >
            Aprovar e assinar →
          </button>
        </div>
      </div>
    </article>
  );
}
