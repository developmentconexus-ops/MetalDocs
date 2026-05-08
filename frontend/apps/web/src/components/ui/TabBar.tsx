import styles from "./TabBar.module.css";

export type TabBarItem = {
  key: string;
  label: string;
  count?: number;
};

type TabBarProps = {
  tabs: TabBarItem[];
  activeKey: string;
  onTabChange: (key: string) => void;
  ariaLabel?: string;
};

export function TabBar({ tabs, activeKey, onTabChange, ariaLabel }: TabBarProps) {
  return (
    <div role="tablist" aria-label={ariaLabel} className={styles.bar}>
      {tabs.map((tab) => {
        const isActive = tab.key === activeKey;
        return (
          <button
            key={tab.key}
            type="button"
            role="tab"
            aria-selected={isActive}
            className={`${styles.tab} ${isActive ? styles.tabActive : ""}`.trim()}
            onClick={() => onTabChange(tab.key)}
          >
            <span className={styles.label}>{tab.label}</span>
            {tab.count !== undefined && (
              <span className={styles.count}>· {tab.count}</span>
            )}
          </button>
        );
      })}
    </div>
  );
}
