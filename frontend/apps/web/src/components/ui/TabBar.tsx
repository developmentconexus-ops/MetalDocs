import { useRef } from "react";
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
  const buttonsRef = useRef<Array<HTMLButtonElement | null>>([]);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLButtonElement>, idx: number) => {
    if (e.key !== "ArrowLeft" && e.key !== "ArrowRight" && e.key !== "Home" && e.key !== "End") return;
    e.preventDefault();
    let next = idx;
    if (e.key === "ArrowLeft") next = (idx - 1 + tabs.length) % tabs.length;
    else if (e.key === "ArrowRight") next = (idx + 1) % tabs.length;
    else if (e.key === "Home") next = 0;
    else if (e.key === "End") next = tabs.length - 1;
    const nextTab = tabs[next];
    if (!nextTab) return;
    onTabChange(nextTab.key);
    buttonsRef.current[next]?.focus();
  };

  return (
    <div role="tablist" aria-label={ariaLabel} className={styles.bar}>
      {tabs.map((tab, idx) => {
        const isActive = tab.key === activeKey;
        return (
          <button
            key={tab.key}
            ref={(el) => {
              buttonsRef.current[idx] = el;
            }}
            type="button"
            role="tab"
            aria-selected={isActive}
            tabIndex={isActive ? 0 : -1}
            className={`${styles.tab} ${isActive ? styles.tabActive : ""}`.trim()}
            onClick={() => onTabChange(tab.key)}
            onKeyDown={(e) => handleKeyDown(e, idx)}
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
