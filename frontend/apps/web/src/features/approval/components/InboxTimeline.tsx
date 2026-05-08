import { Avatar } from '../../../components/ui/Avatar';
import styles from './InboxTimeline.module.css';

export function InboxTimeline() {
  return (
    <div className={styles.timelineContainer}>
      <div className={styles.inner}>
        {/* Header */}
        <div className={styles.timelineHeader}>
          <div className={styles.timelineHeadingBlock}>
            <div className="kicker">APROVAÇÕES · LINHA DO TEMPO</div>
            <h1 className={styles.timelineTitle}>
              Suas <span className={styles.titleAccent}>4 decisões</span> na ordem que importam
            </h1>
            <p className="caption">Organizadas por prazo. As de hoje pulsam em laranja.</p>
          </div>

          {/* Heatmap widget */}
          <div className={styles.heatmap}>
            <div className="kicker">SUAS DECISÕES · 14 DIAS</div>
            <div className={styles.heatmapBars}>
              {[3, 5, 2, 7, 4, 6, 8, 5, 3, 9, 4, 7, 5, 2].map((v, i) => (
                <div
                  key={i}
                  className={styles.heatmapBar}
                  style={{ height: `${(v / 9) * 100}%` }}
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

        {/* Timeline */}
        <div className={styles.timeline}>
          <div className={styles.timelineRail} aria-hidden="true" />

          {/* Bucket: Hoje */}
          <section className={styles.bucketSection}>
            <div className={`${styles.bucketDot} ${styles.bucketDotUrgent}`} aria-hidden="true" />
            <div className={styles.bucketHeader}>
              <h2 className={`${styles.bucketLabel} ${styles.bucketLabelUrgent}`}>Hoje</h2>
              <span className={`${styles.bucketSub} mono`}>até 18:00</span>
              <span className={styles.spacer} />
              <span className={`${styles.bucketCount} ${styles.bucketCountUrgent}`}>2 docs</span>
            </div>

            <div className={styles.bucketItems}>
              <div role="button" tabIndex={0} className={styles.itemRow}>
                <div className={styles.itemTime}>
                  <div className={`${styles.itemTimeValue} mono`}>18:00</div>
                  <div className="caption">vence em 3h 28min</div>
                </div>
                <div className={styles.itemMeta}>
                  <div className={styles.itemMetaTop}>
                    <span className={styles.itemKind}>POP</span>
                    <span className={`${styles.itemCode} mono`}>POP-QUA-0148</span>
                  </div>
                  <div className={styles.itemTitle}>Inspeção de recebimento — solda MIG/MAG</div>
                  <div className={styles.itemAuthLine}>
                    <Avatar name="Carla Pinheiro" size="xs" />
                    <span>Carla Pinheiro</span>
                    <span className={styles.itemDot}>·</span>
                    <span>Qualidade</span>
                    <span className={styles.itemDot}>·</span>
                    <span>12 alterações</span>
                  </div>
                </div>
                <div className={styles.itemSpacer} />
                <div className={styles.stageProgress}>
                  <div className={styles.stageBars}>
                    <span className={`${styles.stageBar} ${styles.stageBarFilled} ${styles.stageBarUrgent}`} />
                    <span className={`${styles.stageBar} ${styles.stageBarFilled} ${styles.stageBarUrgent}`} />
                    <span className={styles.stageBar} />
                    <span className={styles.stageBar} />
                  </div>
                  <div className={`${styles.stageLabel} caption mono`}>ETAPA 2/4 · L2 — você</div>
                </div>
                <button type="button" className={`${styles.reviewBtn} btn btn-primary btn-sm`}>
                  Revisar →
                </button>
              </div>

              <div role="button" tabIndex={0} className={styles.itemRow}>
                <div className={styles.itemTime}>
                  <div className={`${styles.itemTimeValue} mono`}>23:30</div>
                  <div className="caption">vence em 9h</div>
                </div>
                <div className={styles.itemMeta}>
                  <div className={styles.itemMetaTop}>
                    <span className={styles.itemKind}>IT</span>
                    <span className={`${styles.itemCode} mono`}>IT-PROD-0072</span>
                  </div>
                  <div className={styles.itemTitle}>Setup da prensa hidráulica P-440</div>
                  <div className={styles.itemAuthLine}>
                    <Avatar name="João Aguiar" size="xs" />
                    <span>João Aguiar</span>
                    <span className={styles.itemDot}>·</span>
                    <span>Produção</span>
                    <span className={styles.itemDot}>·</span>
                    <span>8 alterações</span>
                  </div>
                </div>
                <div className={styles.itemSpacer} />
                <div className={styles.stageProgress}>
                  <div className={styles.stageBars}>
                    <span className={`${styles.stageBar} ${styles.stageBarFilled} ${styles.stageBarUrgent}`} />
                    <span className={`${styles.stageBar} ${styles.stageBarFilled} ${styles.stageBarUrgent}`} />
                    <span className={styles.stageBar} />
                    <span className={styles.stageBar} />
                  </div>
                  <div className={`${styles.stageLabel} caption mono`}>ETAPA 2/4 · L2 — você</div>
                </div>
                <button type="button" className={`${styles.reviewBtn} btn btn-primary btn-sm`}>
                  Revisar →
                </button>
              </div>
            </div>
          </section>

          {/* Bucket: Amanhã */}
          <section className={styles.bucketSection}>
            <div className={styles.bucketDot} aria-hidden="true" />
            <div className={styles.bucketHeader}>
              <h2 className={styles.bucketLabel}>Amanhã</h2>
              <span className={`${styles.bucketSub} mono`}>sáb 05/05</span>
              <span className={styles.spacer} />
              <span className={styles.bucketCount}>1 docs</span>
            </div>

            <div className={styles.bucketItems}>
              <div role="button" tabIndex={0} className={styles.itemRow}>
                <div className={styles.itemTime}>
                  <div className={`${styles.itemTimeValue} mono`}>09:00</div>
                  <div className="caption">vence em 18h</div>
                </div>
                <div className={styles.itemMeta}>
                  <div className={styles.itemMetaTop}>
                    <span className={styles.itemKind}>POL</span>
                    <span className={`${styles.itemCode} mono`}>POL-RH-0011</span>
                  </div>
                  <div className={styles.itemTitle}>Política de afastamento parcial</div>
                  <div className={styles.itemAuthLine}>
                    <Avatar name="Fábio Ribas" size="xs" />
                    <span>Fábio Ribas</span>
                    <span className={styles.itemDot}>·</span>
                    <span>RH</span>
                    <span className={styles.itemDot}>·</span>
                    <span>4 alterações</span>
                  </div>
                </div>
                <div className={styles.itemSpacer} />
                <div className={styles.stageProgress}>
                  <div className={styles.stageBars}>
                    <span className={`${styles.stageBar} ${styles.stageBarFilled}`} />
                    <span className={styles.stageBar} />
                    <span className={styles.stageBar} />
                  </div>
                  <div className={`${styles.stageLabel} caption mono`}>ETAPA 1/3 · L2 — você</div>
                </div>
                <button type="button" className={`${styles.reviewBtn} btn btn-primary btn-sm`}>
                  Revisar →
                </button>
              </div>
            </div>
          </section>

          {/* Bucket: Esta semana */}
          <section className={styles.bucketSection}>
            <div className={styles.bucketDot} aria-hidden="true" />
            <div className={styles.bucketHeader}>
              <h2 className={styles.bucketLabel}>Esta semana</h2>
              <span className={`${styles.bucketSub} mono`}>até 11/05</span>
              <span className={styles.spacer} />
              <span className={styles.bucketCount}>1 docs</span>
            </div>

            <div className={styles.bucketItems}>
              <div role="button" tabIndex={0} className={styles.itemRow}>
                <div className={styles.itemTime}>
                  <div className={`${styles.itemTimeValue} mono`}>seg 14:00</div>
                  <div className="caption">vence em 2 dias</div>
                </div>
                <div className={styles.itemMeta}>
                  <div className={styles.itemMetaTop}>
                    <span className={styles.itemKind}>DC</span>
                    <span className={`${styles.itemCode} mono`}>DC-TI-0203</span>
                  </div>
                  <div className={styles.itemTitle}>Diagrama de rede — DMZ planta 2</div>
                  <div className={styles.itemAuthLine}>
                    <Avatar name="Ana Iuri" size="xs" />
                    <span>Ana Iuri</span>
                    <span className={styles.itemDot}>·</span>
                    <span>TI</span>
                    <span className={styles.itemDot}>·</span>
                    <span>6 alterações</span>
                  </div>
                </div>
                <div className={styles.itemSpacer} />
                <div className={styles.stageProgress}>
                  <div className={styles.stageBars}>
                    <span className={`${styles.stageBar} ${styles.stageBarFilled}`} />
                    <span className={`${styles.stageBar} ${styles.stageBarFilled}`} />
                    <span className={styles.stageBar} />
                    <span className={styles.stageBar} />
                  </div>
                  <div className={`${styles.stageLabel} caption mono`}>ETAPA 2/4 · L2 — você</div>
                </div>
                <button type="button" className={`${styles.reviewBtn} btn btn-primary btn-sm`}>
                  Revisar →
                </button>
              </div>
            </div>
          </section>

          {/* Bucket: Próximo mês — empty */}
          <section className={styles.bucketSection}>
            <div className={styles.bucketDot} aria-hidden="true" />
            <div className={styles.bucketHeader}>
              <h2 className={styles.bucketLabel}>Próximo mês</h2>
              <span className={`${styles.bucketSub} mono`}>até 31/05</span>
              <span className={styles.spacer} />
              <span className={`${styles.bucketCount} ${styles.bucketCountEmpty}`}>0 docs</span>
            </div>
            <div className={styles.emptyBucket}>Nada ainda. Continue assim.</div>
          </section>
        </div>
      </div>
    </div>
  );
}
