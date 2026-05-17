import { Outlet, useMatches } from 'react-router-dom';
import { Rail } from './Rail';
import { AppToolbar } from './AppToolbar';
import { SectionPanel } from './SectionPanel';
import styles from './AppShell.module.css';

type RouteHandle = { sectionPanel?: boolean; workspaceView?: string };

export function AppShell() {
  const matches = useMatches();
  const hasSectionPanel = matches.some(
    (m) => (m.handle as RouteHandle | undefined)?.sectionPanel === true,
  );
  const hideToolbar = matches.some(
    (m) => (m.handle as RouteHandle | undefined)?.workspaceView === 'document-editor',
  );

  return (
    <div className={styles.shell}>
      <Rail />
      <div className={styles.main}>
        {!hideToolbar && <AppToolbar />}
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
