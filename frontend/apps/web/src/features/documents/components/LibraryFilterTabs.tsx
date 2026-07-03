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
      {/* FE-19 N8: Filtros/Exportar removed rather than shipped permanently
          disabled — no working control until server-side advanced filters and
          an export job endpoint exist. */}
    </div>
  );
}
