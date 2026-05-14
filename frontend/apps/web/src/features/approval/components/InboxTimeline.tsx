import { Avatar } from '../../../components/ui/Avatar';
import type { InboxItem } from '../api/approvalTypes';
import styles from './InboxTimeline.module.css';

interface InboxTimelineProps {
  items: InboxItem[];
  onOpenDocument?: (item: InboxItem) => void;
}

interface Bucket {
  label: string;
  sub: string;
  items: InboxItem[];
}

function groupBySubmittedBucket(items: InboxItem[]): Bucket[] {
  const now = new Date();
  const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
  const yesterdayStart = todayStart - 24 * 60 * 60 * 1000;
  const weekStart = todayStart - 7 * 24 * 60 * 60 * 1000;

  const today: InboxItem[] = [];
  const yesterday: InboxItem[] = [];
  const thisWeek: InboxItem[] = [];
  const older: InboxItem[] = [];

  for (const item of items) {
    const at = new Date(item.submitted_at).getTime();
    if (at >= todayStart) {
      today.push(item);
    } else if (at >= yesterdayStart) {
      yesterday.push(item);
    } else if (at >= weekStart) {
      thisWeek.push(item);
    } else {
      older.push(item);
    }
  }

  return [
    { label: 'Hoje', sub: 'enviados hoje', items: today },
    { label: 'Ontem', sub: 'enviados ontem', items: yesterday },
    { label: 'Últimos 7 dias', sub: 'enviados nesta semana', items: thisWeek },
    { label: 'Mais antigos', sub: 'enviados antes desta semana', items: older },
  ];
}

function handleRowKeyDown(
  e: React.KeyboardEvent<HTMLDivElement>,
  onClick: () => void,
): void {
  if (e.key === 'Enter' || e.key === ' ') {
    e.preventDefault();
    onClick();
  }
}

export function InboxTimeline({ items, onOpenDocument }: InboxTimelineProps) {
  const buckets = groupBySubmittedBucket(items);

  return (
    <div className={styles.timelineContainer}>
      <div className={styles.inner}>
        {/* Header */}
        <div className={styles.timelineHeader}>
          <div className={styles.timelineHeadingBlock}>
            <div className="kicker">APROVAÇÕES · LINHA DO TEMPO</div>
            <h1 className={styles.timelineTitle}>
              Suas <span className={styles.titleAccent}>{items.length} decisões</span> na ordem que importam
            </h1>
            <p className="caption">Organizadas por data de envio para revisão.</p>
          </div>

          <div className={styles.heatmap}>
            <div className="kicker">SUAS DECISÕES</div>
            <p className="caption">Histórico de decisões ainda não disponível</p>
          </div>
        </div>

        {/* Timeline — each bucket is a 2-col grid row: [rail col] [content col] */}
        <div className={styles.timeline}>
          {buckets.map((bucket, bIdx) => (
            <div key={bucket.label} className={styles.bucketRow}>
              {/* Rail column: dot + connecting line — both centered in the same column */}
              <div className={styles.railCol} aria-hidden="true">
                <div
                  className={[
                    styles.railDot,
                    bucket.items.length > 0 ? styles.railDotFilled : '',
                  ].filter(Boolean).join(' ')}
                />
                {bIdx < buckets.length - 1 && (
                  <div
                    className={[
                      styles.railLine,
                      bIdx === 0 ? styles.railLineFirst : '',
                    ].filter(Boolean).join(' ')}
                  />
                )}
              </div>

              {/* Content column */}
              <section className={styles.bucketSection}>
                <div className={styles.bucketHeader}>
                  <h2 className={styles.bucketLabel}>{bucket.label}</h2>
                  <span className={`${styles.bucketSub} mono`}>{bucket.sub}</span>
                  <span className={styles.spacer} />
                  <span
                    className={[
                      styles.bucketCount,
                      bucket.items.length === 0 ? styles.bucketCountEmpty : '',
                    ].filter(Boolean).join(' ')}
                  >
                    {bucket.items.length} docs
                  </span>
                </div>

                {bucket.items.length === 0 ? (
                  <div className={styles.emptyBucket}>Nada ainda. Continue assim.</div>
                ) : (
                  <div className={styles.bucketItems}>
                    {bucket.items.map((item) => {
                      const handleClick = () => {
                        onOpenDocument?.(item);
                      };
                      return (
                        <div
                          key={item.instance_id}
                          role="button"
                          tabIndex={0}
                          className={styles.itemRow}
                          onClick={handleClick}
                          onKeyDown={(e) => handleRowKeyDown(e, handleClick)}
                        >
                          <div className={styles.itemTime}>
                            <div className={`${styles.itemTimeValue} mono`}>
                              {new Date(item.submitted_at).toLocaleTimeString('pt-BR', {
                                hour: '2-digit',
                                minute: '2-digit',
                              })}
                            </div>
                            <div className="caption">enviado para revisão</div>
                          </div>
                          <div className={styles.itemMeta}>
                            <div className={styles.itemMetaTop}>
                              <span className={`${styles.itemCode} mono`}>{item.controlled_document_id}</span>
                            </div>
                            <div className={styles.itemTitle}>{item.document_title}</div>
                            <div className={styles.itemAuthLine}>
                              <Avatar name={item.submitted_by} size="xs" />
                              <span>{item.submitted_by}</span>
                              <span className={styles.itemDot}>·</span>
                              <span>{item.area_code}</span>
                              <span className={styles.itemDot}>·</span>
                              <span>{item.quorum_progress}</span>
                            </div>
                          </div>
                          <div className={styles.itemSpacer} />
                          <div className={styles.stageProgress}>
                            <div className={`${styles.stageLabel} caption mono`}>
                              {item.stage_label}
                            </div>
                          </div>
                          <button
                            type="button"
                            className={`${styles.reviewBtn} btn btn-primary btn-sm`}
                            onClick={(e) => {
                              e.stopPropagation();
                              onOpenDocument?.(item);
                            }}
                          >
                            Revisar →
                          </button>
                        </div>
                      );
                    })}
                  </div>
                )}
              </section>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
