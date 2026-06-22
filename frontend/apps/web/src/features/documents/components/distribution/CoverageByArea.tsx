import type { DistributionAreaCoverage } from '../../api/distribution';
import styles from './CoverageByArea.module.css';

export interface CoverageByAreaProps {
  rows: DistributionAreaCoverage[];
  loading?: boolean;
  error?: boolean;
}

export function CoverageByArea({ rows, loading = false, error = false }: CoverageByAreaProps) {
  if (error) {
    return (
      <div className={styles.card}>
        <div className={styles.header} role="alert">
          <span className={styles.headerNote}>Não foi possível carregar a cobertura por área.</span>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className={styles.card}>
        <div className={styles.header} role="status" aria-live="polite">
          <span className={styles.headerNote}>Carregando cobertura por área…</span>
        </div>
      </div>
    );
  }

  if (rows.length === 0) {
    return (
      <div className={styles.card}>
        <div className={styles.header}>
          <span className={styles.headerNote}>Nenhuma área registrada.</span>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <span className={styles.headerNote}>
          Áreas ordenadas por nome · total de destinatários obrigatórios
        </span>
      </div>
      <div className={styles.rows}>
        {rows.map((a) => (
          <div key={a.area_code} className={styles.areaRow}>
            <div className={styles.areaLabel}>
              <div className={styles.areaName}>{a.area_name}</div>
            </div>
            <div className={styles.pctWrap}>
              <span className={styles.pctValue}>{a.total}</span>
              <span className={styles.pctLabel}>destinatários</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
