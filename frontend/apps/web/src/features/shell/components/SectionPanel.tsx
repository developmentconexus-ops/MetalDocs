import styles from './SectionPanel.module.css';

// Slot component — Library screen will fill this panel via route handle.
export function SectionPanel() {
  return <aside className={styles.panel} aria-label="Filtros" />;
}
