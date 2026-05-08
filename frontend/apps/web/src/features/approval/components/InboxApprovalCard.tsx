import { Avatar } from '../../../components/ui/Avatar';
import { Icon } from '../../../components/ui/Icon';
import styles from './InboxApprovalCard.module.css';

export function InboxApprovalCard() {
  return (
    <article className={styles.card}>
      <div className={styles.cardHeader}>
        <div className={styles.cardHeaderInner}>
          <div className={styles.kindBadge}>POP</div>
          <div className={styles.cardHeaderMeta}>
            <div className={`${styles.codeVersion} mono`}>POP-QUA-0148 · v2.3 → v2.4</div>
            <h2 className={styles.cardTitle}>Inspeção de recebimento — solda MIG/MAG</h2>
          </div>
          <div className={styles.deadlineBlock}>
            <div className={styles.deadlineLabel}>VENCE EM</div>
            <div className={styles.deadlineValue}>3h 28min</div>
          </div>
        </div>
      </div>
      <div className={styles.cardBody}>
        <p className={styles.summary}>
          Revisão da inspeção de soldas em juntas críticas. Adiciona §3.2 sobre inspeção visual ampliada conforme NBR ISO 5817 (qualidade B).
        </p>
        <div className={styles.statsGrid}>
          <div className={styles.statCell}>
            <div className="kicker">AUTOR</div>
            <div className={styles.authorRow}>
              <Avatar name="Carla Pinheiro" size="sm" />
              <div>
                <div className={styles.authorName}>Carla Pinheiro</div>
                <div className="caption">Qualidade</div>
              </div>
            </div>
          </div>
          <div className={styles.statCell}>
            <div className="kicker">ALTERAÇÕES</div>
            <div className={styles.changesRow}>
              <span className={styles.changesNum}>12</span>
              <span className="caption">edições</span>
            </div>
          </div>
          <div className={styles.statCell}>
            <div className="kicker">ESTÁGIO</div>
            <div className={styles.stageName}>L2 · Você</div>
            <div className="caption">de 4 etapas</div>
          </div>
        </div>
        <div className={styles.cardActions}>
          <button type="button" className={`${styles.btnOpen} btn`}>
            <Icon name="docs" size={14} /> Abrir documento
          </button>
          <button type="button" className={`${styles.btnReturn} btn`}>
            Devolver
          </button>
          <button type="button" className={`${styles.btnApprove} btn btn-primary`}>
            Aprovar e assinar →
          </button>
        </div>
      </div>
    </article>
  );
}
