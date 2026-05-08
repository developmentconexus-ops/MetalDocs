import { Avatar } from '../../../components/ui/Avatar';
import type { RichInboxItem } from '../lib/mockInboxData';
import styles from './InboxTimeline.module.css';

interface InboxTimelineProps {
  items: RichInboxItem[];
}

interface Bucket {
  label: string;
  sub: string;
  urgent: boolean;
  items: RichInboxItem[];
}

function groupByDeadlineBucket(items: RichInboxItem[]): Bucket[] {
  const todayEnd = new Date();
  todayEnd.setHours(23, 59, 59, 999);
  const tomorrowEnd = new Date(todayEnd);
  tomorrowEnd.setDate(tomorrowEnd.getDate() + 1);
  const weekEnd = new Date(todayEnd);
  weekEnd.setDate(weekEnd.getDate() + 7);

  const today: RichInboxItem[] = [];
  const tomorrow: RichInboxItem[] = [];
  const thisWeek: RichInboxItem[] = [];
  const nextMonth: RichInboxItem[] = [];

  for (const item of items) {
    const at = new Date(item.deadline_at).getTime();
    if (at <= todayEnd.getTime()) {
      today.push(item);
    } else if (at <= tomorrowEnd.getTime()) {
      tomorrow.push(item);
    } else if (at <= weekEnd.getTime()) {
      thisWeek.push(item);
    } else {
      nextMonth.push(item);
    }
  }

  const fmt = (d: Date) =>
    d.toLocaleDateString('pt-BR', { weekday: 'short', day: '2-digit', month: '2-digit' });

  return [
    { label: 'Hoje', sub: 'até 18:00', urgent: true, items: today },
    { label: 'Amanhã', sub: fmt(tomorrowEnd), urgent: false, items: tomorrow },
    { label: 'Esta semana', sub: `até ${fmt(weekEnd)}`, urgent: false, items: thisWeek },
    { label: 'Próximo mês', sub: 'até 31/05', urgent: false, items: nextMonth },
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

export function InboxTimeline({ items }: InboxTimelineProps) {
  const buckets = groupByDeadlineBucket(items);

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
            <p className="caption">Organizadas por prazo. As de hoje pulsam em laranja.</p>
          </div>

          {/* Heatmap widget — hardcoded; real data deferred */}
          {/* TODO [BACKLOG: caixa-aprovacao.md]: wire real heatmap data */}
          <div className={styles.heatmap}>
            <div className="kicker">SUAS DECISÕES · 14 DIAS</div>
            <div className={styles.heatmapBars}>
              {[3, 5, 2, 7, 4, 6, 8, 5, 3, 9, 4, 7, 5, 2].map((v, i) => (
                <div
                  key={i}
                  className={styles.heatmapBar}
                  style={{
                    height: `${(v / 9) * 100}%`,
                    opacity: 0.3 + (v / 9) * 0.7,
                    animationDelay: `${i * 30}ms`,
                  }}
                  title={`${v} decisões`}
                />
              ))}
            </div>
            <div className={styles.heatmapLabels}>
              <span className={`${styles.heatmapLabel} mono`}>22/abr</span>
              <span className={`${styles.heatmapLabel} mono`}>hoje</span>
            </div>
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
                    bucket.urgent ? styles.railDotUrgent : '',
                    !bucket.urgent && bucket.items.length > 0 ? styles.railDotFilled : '',
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
                  <h2 className={`${styles.bucketLabel}${bucket.urgent ? ` ${styles.bucketLabelUrgent}` : ''}`}>
                    {bucket.label}
                  </h2>
                  <span className={`${styles.bucketSub} mono`}>{bucket.sub}</span>
                  <span className={styles.spacer} />
                  <span
                    className={[
                      styles.bucketCount,
                      bucket.urgent && bucket.items.length > 0 ? styles.bucketCountUrgent : '',
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
                        // TODO [BACKLOG: caixa-aprovacao.md]: open doc from timeline
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
                            <div className={`${styles.itemTimeValue} mono`}>—</div>
                            <div className="caption">vence em {item.deadline}</div>
                          </div>
                          <div className={styles.itemMeta}>
                            <div className={styles.itemMetaTop}>
                              <span className={styles.itemKind}>{item.kind}</span>
                              <span className={`${styles.itemCode} mono`}>{item.code}</span>
                            </div>
                            <div className={styles.itemTitle}>{item.document_title}</div>
                            <div className={styles.itemAuthLine}>
                              <Avatar name={item.submitted_by} size="xs" />
                              <span>{item.submitted_by}</span>
                              <span className={styles.itemDot}>·</span>
                              <span>{item.area_code}</span>
                              <span className={styles.itemDot}>·</span>
                              <span>{item.changes} alterações</span>
                            </div>
                          </div>
                          <div className={styles.itemSpacer} />
                          <div className={styles.stageProgress}>
                            <div className={styles.stageBars}>
                              <span className={`${styles.stageBar} ${styles.stageBarFilled}${bucket.urgent ? ` ${styles.stageBarUrgent}` : ''}`} />
                              <span className={`${styles.stageBar} ${styles.stageBarFilled}${bucket.urgent ? ` ${styles.stageBarUrgent}` : ''}`} />
                              <span className={styles.stageBar} />
                              <span className={styles.stageBar} />
                            </div>
                            <div className={`${styles.stageLabel} caption mono`}>
                              {item.stage_label}
                            </div>
                          </div>
                          <button
                            type="button"
                            className={`${styles.reviewBtn} btn btn-primary btn-sm`}
                            onClick={(e) => {
                              e.stopPropagation();
                              // TODO [BACKLOG: caixa-aprovacao.md]: open doc from timeline
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
