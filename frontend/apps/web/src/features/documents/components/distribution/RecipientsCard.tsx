import { Avatar } from '../../../../components/ui/Avatar';
import type { DistributionRecipient } from '../../api/distribution';
import styles from './RecipientsCard.module.css';

// Origin chip label derived from the `source` field + optional area_name
function originLabel(recipient: DistributionRecipient): string {
  if (recipient.source === 'area_grant') return recipient.area_name ?? recipient.area_code ?? 'Área';
  if (recipient.source === 'user_grant') return 'Concessão direta';
  return 'Toda a empresa';
}

export interface RecipientsCardProps {
  items: DistributionRecipient[];
  hasMore: boolean;
  onLoadMore: () => void;
  loading?: boolean;
  error?: boolean;
}

export function RecipientsCard({
  items,
  hasMore,
  onLoadMore,
  loading = false,
  error = false,
}: RecipientsCardProps) {
  if (error) {
    return (
      <div className={styles.card}>
        <div className={styles.filterBar} role="alert">
          <span>Não foi possível carregar os destinatários.</span>
        </div>
      </div>
    );
  }

  if (loading && items.length === 0) {
    return (
      <div className={styles.card}>
        <div className={styles.filterBar} role="status" aria-live="polite">
          <span>Carregando destinatários…</span>
        </div>
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div className={styles.card}>
        <div className={styles.filterBar}>
          <span>Nenhum destinatário obrigatório.</span>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.card}>
      {/* Column header row */}
      <div className={styles.headerRow}>
        <span className={styles.colRecipient}>Destinatário</span>
        <span className={styles.colArea}>Origem</span>
      </div>

      {/* Data rows */}
      {items.map((p, i) => (
        <div
          key={p.user_id}
          className={`${styles.recipientRow} ${i < items.length - 1 ? styles.recipientRowBorder : ''}`}
        >
          <div className={styles.colRecipient}>
            <Avatar name={p.name} size="sm" />
            <div className={styles.recipientInfo}>
              <div className={styles.recipientName}>{p.name}</div>
            </div>
          </div>
          <span className={styles.colArea}>
            <span className={styles.originChip}>{originLabel(p)}</span>
          </span>
        </div>
      ))}

      {/* Load-more footer */}
      {hasMore && (
        <div className={styles.paginationFooter}>
          <button
            type="button"
            className={styles.loadMoreBtn}
            onClick={onLoadMore}
            disabled={loading}
          >
            {loading ? 'Carregando…' : 'Carregar mais'}
          </button>
        </div>
      )}
    </div>
  );
}
