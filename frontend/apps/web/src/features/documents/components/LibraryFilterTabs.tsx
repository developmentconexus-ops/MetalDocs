import styles from './LibraryFilterTabs.module.css';

type FilterTab = {
  value: string;
  label: string;
  statusKey: string | null;
};

const TABS: FilterTab[] = [
  { value: 'todos', label: 'Todos', statusKey: null },
  { value: 'rascunhos', label: 'Rascunhos', statusKey: 'draft' },
  { value: 'em_revisao', label: 'Em Revisão', statusKey: 'under_review' },
  { value: 'aprovados', label: 'Aprovados', statusKey: 'approved' },
  { value: 'publicados', label: 'Publicados', statusKey: 'published' },
  { value: 'rejeitados', label: 'Rejeitados', statusKey: 'rejected' },
  { value: 'obsoletos', label: 'Obsoletos', statusKey: 'obsolete' },
];

type LibraryFilterTabsProps = {
  activeTab: string;
  onTabChange: (tab: string) => void;
  statsByStatus?: Record<string, number>;
  total?: number;
};

export function LibraryFilterTabs({
  activeTab,
  onTabChange,
  statsByStatus = {},
  total = 0,
}: LibraryFilterTabsProps): JSX.Element {
  function getCount(tab: FilterTab): number | null {
    if (tab.statusKey === null) return total;
    const c = statsByStatus[tab.statusKey];
    return typeof c === 'number' ? c : null;
  }

  return (
    <div className={styles.root} role="tablist" aria-label="Filtro de documentos">
      {TABS.map((tab) => {
        const count = getCount(tab);
        const isActive = activeTab === tab.value;
        return (
          <button
            key={tab.value}
            type="button"
            role="tab"
            aria-selected={isActive}
            className={`${styles.tab} ${isActive ? styles.activeTab : ''}`}
            onClick={() => onTabChange(tab.value)}
          >
            <span>{tab.label}</span>
            {count !== null && count > 0 ? (
              <span className={styles.count}>{count.toLocaleString('pt-BR')}</span>
            ) : null}
          </button>
        );
      })}
      <span className={styles.spacer} aria-hidden="true" />
      <button type="button" className={styles.actionBtn}>
        <svg width="13" height="13" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
          <path d="M3 5h14M5 10h10M8 15h4" />
        </svg>
        Filtros
      </button>
      <button type="button" className={styles.actionBtn}>
        <svg width="13" height="13" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
          <path d="M10 3v10m0 0l-4-4m4 4l4-4M3 17h14" />
        </svg>
        Exportar
      </button>
    </div>
  );
}
