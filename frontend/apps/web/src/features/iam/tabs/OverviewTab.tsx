import KpiStrip from "../components/KpiStrip";
import PresencePanel from "../components/PresencePanel";
import ActivityPanel from "../components/ActivityPanel";
import styles from "./OverviewTab.module.css";

export default function OverviewTab() {
  return (
    <section
      className={styles.tab}
      data-testid="admin-overview-tab"
      aria-labelledby="admin-overview-heading"
    >
      <div className={styles.headerBlock}>
        <h2 id="admin-overview-heading" className={styles.heading}>
          Visão geral
        </h2>
        <p className={styles.lede}>
          Saúde do tenant em tempo real — indicadores, presença e últimas ações auditadas.
        </p>
      </div>

      <KpiStrip />

      <div className={styles.bento}>
        <PresencePanel />
        <ActivityPanel />
      </div>
    </section>
  );
}
