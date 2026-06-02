import { useMemo } from 'react';
import { Navigate, Outlet, useMatches } from 'react-router-dom';
import { Rail } from './Rail';
import { AppToolbar } from './AppToolbar';
import { SectionPanel } from './SectionPanel';
import { useAuthStore } from '../../../store/auth.store';
import styles from './AppShell.module.css';

// Capability-gated route metadata. Backend remains the sole authz enforcer
// (wiki/concepts/authz-tiers.md, ADR 0016); this gate is a UX hint that
// matches the backend's view-grade cap on the corresponding endpoint so
// users don't land on a screen whose data load will 403.
type RouteHandle = {
  sectionPanel?: boolean;
  workspaceView?: string;
  requiresCapability?: string;
  requiresAnyCapability?: string[];
};

type RequiredCapability =
  | { kind: 'single'; cap: string }
  | { kind: 'any'; caps: string[] };

export function AppShell() {
  const matches = useMatches();
  const capabilities = useAuthStore((state) => state.user?.capabilities);
  const hasSectionPanel = matches.some(
    (m) => (m.handle as RouteHandle | undefined)?.sectionPanel === true,
  );
  const hideToolbar = matches.some(
    (m) => (m.handle as RouteHandle | undefined)?.workspaceView === 'document-editor',
  );
  const required = useMemo<RequiredCapability | null>(() => {
    for (const m of matches) {
      const handle = m.handle as RouteHandle | undefined;
      if (handle?.requiresCapability) {
        return { kind: 'single', cap: handle.requiresCapability };
      }
      if (handle?.requiresAnyCapability && handle.requiresAnyCapability.length > 0) {
        return { kind: 'any', caps: handle.requiresAnyCapability };
      }
    }
    return null;
  }, [matches]);

  if (required) {
    const granted = capabilities ?? [];
    const allowed =
      required.kind === 'single'
        ? granted.includes(required.cap)
        : required.caps.some((cap) => granted.includes(cap));
    if (!allowed) {
      return <Navigate to="/" replace />;
    }
  }

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
