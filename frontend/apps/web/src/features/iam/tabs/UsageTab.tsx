import UsageGauges from "../components/UsageGauges";
import styles from "./UsageTab.module.css";

export default function UsageTab() {
  return (
    <section
      className={styles.tab}
      data-testid="admin-usage-tab"
      aria-labelledby="admin-usage-heading"
    >
      <div className={styles.headerBlock}>
        <h2 id="admin-usage-heading" className={styles.heading}>
          Consumo
        </h2>
        <p className={styles.lede}>
          Capacidade contratada, atividade real e volume de chamadas — atualizado a cada 5 minutos.
        </p>
      </div>

      <UsageGauges />
    </section>
  );
}
