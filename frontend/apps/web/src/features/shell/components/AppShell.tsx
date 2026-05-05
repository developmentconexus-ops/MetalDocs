import { Outlet, useMatches } from 'react-router-dom';
import { Rail } from './Rail';
import { AppToolbar } from './AppToolbar';
import { SectionPanel } from './SectionPanel';
import styles from './AppShell.module.css';

type RouteHandle = { sectionPanel?: boolean };

export function AppShell() {
  const matches = useMatches();
  const hasSectionPanel = matches.some(
    (m) => (m.handle as RouteHandle | undefined)?.sectionPanel === true,
  );

  return (
    <div className={styles.shell}>
      <Rail />
      <div className={styles.main}>
        <AppToolbar />
        <div className={styles.content}>
          {hasSectionPanel && <SectionPanel />}
          <main className={styles.page}>
            <Outlet />
          </main>
        </div>
      </div>
    </div>
  );
}
