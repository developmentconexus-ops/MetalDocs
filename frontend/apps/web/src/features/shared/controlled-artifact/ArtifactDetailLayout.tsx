import { NavLink, Outlet } from "react-router-dom";
import type { ArtifactTab } from "./types";
import styles from "./ArtifactDetailLayout.module.css";

/**
 * Tabbed shell for the controlled-artifact detail view. Purely presentational —
 * no data fetching, no kind branching, no kind-specific defaults. Tabs are
 * model-driven and REQUIRED: each caller (documents, templates) owns and passes
 * its own tab set. Each tab renders as a router NavLink when `href` is set (active
 * styling preserved), or a static non-link span when `href` is absent (single-tab
 * case). The shared layer holds no document- or template-specific routing.
 */

interface ArtifactDetailLayoutProps {
  tabs: ArtifactTab[];
}

export function ArtifactDetailLayout({ tabs }: ArtifactDetailLayoutProps) {
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
