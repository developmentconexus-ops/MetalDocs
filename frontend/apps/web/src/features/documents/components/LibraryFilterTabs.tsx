import styles from './LibraryFilterTabs.module.css';

type FilterTab = {
  value: string;
  label: string;
};

const TABS: FilterTab[] = [
  { value: 'todos', label: 'Todos' },
  { value: 'rascunhos', label: 'Rascunhos' },
  { value: 'em_revisao', label: 'Em Revisão' },
  { value: 'publicados', label: 'Publicados' },
  { value: 'rejeitados', label: 'Rejeitados' },
  { value: 'obsoletos', label: 'Obsoletos' },
];

type LibraryFilterTabsProps = {
  activeTab: string;
  onTabChange: (tab: string) => void;
};

export function LibraryFilterTabs({ activeTab, onTabChange }: LibraryFilterTabsProps) {
  return (
    <div className={styles.root}>
      {TABS.map((tab) => (
        <button
          key={tab.value}
          type="button"
          className={`${styles.tab} ${activeTab === tab.value ? styles.activeTab : ''}`}
          onClick={() => onTabChange(tab.value)}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}
