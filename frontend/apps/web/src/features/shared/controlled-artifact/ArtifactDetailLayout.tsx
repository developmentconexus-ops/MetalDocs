import { NavLink, Outlet } from "react-router-dom";
import type { ArtifactTab } from "./types";
import styles from "./ArtifactDetailLayout.module.css";

/**
 * Tabbed shell for the controlled-artifact detail view. Purely presentational —
 * no data fetching, no kind branching. Tabs are model-driven: each tab renders as
 * a router NavLink when `href` is set (active styling preserved), or a static
 * non-link span when `href` is absent (template single-tab case).
 *
 * The default tab set reproduces the documents shell ("Documento" / "Distribuição")
 * so the document route can mount this without passing any props. The template route
 * passes a reduced single-tab set.
 */

const DEFAULT_TABS: ArtifactTab[] = [
  { key: "documento", label: "Documento", href: "." },
  { key: "distribuicao", label: "Distribuição", href: "distribution" },
];

interface ArtifactDetailLayoutProps {
  tabs?: ArtifactTab[];
}

export function ArtifactDetailLayout({ tabs = DEFAULT_TABS }: ArtifactDetailLayoutProps) {
  return (
    <div className={styles.root}>
      <nav className={styles.tabStrip}>
        {tabs.map((tab) =>
          tab.href ? (
            <NavLink
              key={tab.key}
              to={tab.href}
              end={tab.href === "."}
              className={({ isActive }) =>
                isActive ? `${styles.tab} ${styles.tabActive}` : styles.tab
              }
            >
              {tab.label}
            </NavLink>
          ) : (
            <span key={tab.key} className={`${styles.tab} ${styles.tabActive}`}>
              {tab.label}
            </span>
          ),
        )}
      </nav>
      <Outlet />
    </div>
  );
}
