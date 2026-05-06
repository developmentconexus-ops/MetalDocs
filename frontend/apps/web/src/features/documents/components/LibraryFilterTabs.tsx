import { LIBRARY_STATUSES, type LibraryFilter } from '../lib/libraryStatus';
import styles from './LibraryFilterTabs.module.css';

type LibraryFilterTabsProps = {
  activeTab: LibraryFilter;
  onTabChange: (tab: LibraryFilter) => void;
  statsByStatus?: Record<string, number>;
  total?: number;
};

export function LibraryFilterTabs({
  activeTab,
  onTabChange,
  statsByStatus = {},
  total = 0,
}: LibraryFilterTabsProps): JSX.Element {
  return (
    <div className={styles.root} role="tablist" aria-label="Filtro de documentos">
      <button
        type="button"
        role="tab"
        aria-selected={activeTab === 'todos'}
        className={`${styles.tab} ${activeTab === 'todos' ? styles.activeTab : ''}`}
        onClick={() => onTabChange('todos')}
      >
        <span>Todos</span>
        {total > 0 ? <span className={styles.count}>{total.toLocaleString('pt-BR')}</span> : null}
      </button>
      {LIBRARY_STATUSES.map((entry) => {
        const count = statsByStatus[entry.status] ?? 0;
        const isActive = activeTab === entry.filter;
        return (
          <button
            key={entry.filter}
            type="button"
            role="tab"
            aria-selected={isActive}
            className={`${styles.tab} ${isActive ? styles.activeTab : ''}`}
            onClick={() => onTabChange(entry.filter)}
          >
            <span>{entry.label}</span>
            {count > 0 ? (
              <span className={styles.count}>{count.toLocaleString('pt-BR')}</span>
            ) : null}
          </button>
        );
      })}
      <span className={styles.spacer} aria-hidden="true" />
      {/* TODO: wire Filtros panel + Exportar action. Disabled until backend
          endpoints exist (server-side advanced filters + export job). */}
      <button
        type="button"
        className={`${styles.actionBtn} ${styles.actionBtnDisabled}`}
        disabled
        aria-disabled="true"
        title="Em breve"
      >
        <svg width="13" height="13" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
          <path d="M3 5h14M5 10h10M8 15h4" />
        </svg>
        Filtros
      </button>
      <button
        type="button"
        className={`${styles.actionBtn} ${styles.actionBtnDisabled}`}
        disabled
        aria-disabled="true"
        title="Em breve"
      >
        <svg width="13" height="13" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
          <path d="M10 3v10m0 0l-4-4m4 4l4-4M3 17h14" />
        </svg>
        Exportar
      </button>
    </div>
  );
}
