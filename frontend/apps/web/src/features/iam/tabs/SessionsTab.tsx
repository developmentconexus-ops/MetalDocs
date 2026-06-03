import SessionsTable from "../components/SessionsTable";
import MfaCoverageCard from "../components/MfaCoverageCard";
import LockoutsTable from "../components/LockoutsTable";
import SecuritySignalsList from "../components/SecuritySignalsList";
import styles from "./SessionsTab.module.css";

export default function SessionsTab() {
  return (
    <section
      className={styles.tab}
      data-testid="admin-sessions-tab"
      aria-labelledby="admin-sessions-heading"
    >
      <div className={styles.headerBlock}>
        <h2 id="admin-sessions-heading" className={styles.heading}>
          Sessões & Segurança
        </h2>
        <p className={styles.lede}>
          Logins ativos, cobertura de MFA, bloqueios e sinais de risco do
          tenant.
        </p>
      </div>

      <SessionsTable />

      <div className={styles.bento}>
        <SecuritySignalsList />
        <MfaCoverageCard />
      </div>

      <LockoutsTable />
    </section>
  );
}
