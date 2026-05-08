import type { ReactNode } from "react";
import { SearchBar } from "./SearchBar";
import styles from "./WorkspaceHeroHeader.module.css";

type WorkspaceHeroHeaderProps = {
  title: string;
  subtitle?: string;
  kicker?: string;
  searchQuery?: string;
  onSearchQueryChange?: (value: string) => void;
  action?: ReactNode;
  variant?: "default" | "compact";
  tone?: "banner" | "flat";
};

export function WorkspaceHeroHeader(props: WorkspaceHeroHeaderProps) {
  const isCompact = props.variant === "compact";
  const isFlat = props.tone === "flat";
  const headerClassName = `${styles.header} ${isCompact ? styles.headerCompact : ""} ${isFlat ? styles.headerFlat : ""}`.trim();
  const heroClassName = `${styles.hero} ${isCompact ? styles.heroCompact : ""}`.trim();
  const titleClassName = `${styles.title} ${isCompact ? styles.titleCompact : ""}`.trim();
  const hasSearch = props.searchQuery !== undefined && props.onSearchQueryChange !== undefined;

  return (
    <header className={headerClassName}>
      <div className={heroClassName}>
        <div className={styles.copy}>
          {props.kicker && <div className={styles.kicker}>{props.kicker}</div>}
          <h1 className={titleClassName}>{props.title}</h1>
          {!isCompact && props.subtitle && <p className={styles.subtitle}>{props.subtitle}</p>}
        </div>
        {hasSearch && (
          <div className={styles.search}>
            <SearchBar value={props.searchQuery as string} onChange={props.onSearchQueryChange as (v: string) => void} />
          </div>
        )}
        {props.action && <div className={styles.action}>{props.action}</div>}
      </div>
    </header>
  );
}
