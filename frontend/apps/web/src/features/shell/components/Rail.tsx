import { useMemo } from 'react';
import { useMatch, useNavigate } from 'react-router-dom';
import { Icon } from '../../../components/ui/Icon';
import { Avatar } from '../../../components/ui/Avatar';
import { useAuthStore } from '../../../store/auth.store';
import { useAuthSession } from '../../auth/useAuthSession';
import styles from './Rail.module.css';

type NavItem = {
  icon: React.ComponentProps<typeof Icon>['name'];
  label: string;
  path: string;
  // Mirrors the route's own `requiresCapability` handle so the rail never
  // offers a destination AppShell would bounce back to "/". UX hint only —
  // backend stays the sole authz enforcer. Capabilities, never roles (ADR 0022).
  requiredCapability?: string;
};

const NAV_ITEMS: NavItem[] = [
  { icon: 'home',     label: 'Início',     path: '/' },
  { icon: 'library',  label: 'Documentos', path: '/documents' },
  { icon: 'template', label: 'Templates',  path: '/templates' },
  { icon: 'inbox',    label: 'Aprovações', path: '/approvals' },
];

const ADMIN_NAV_ITEMS: NavItem[] = [
  { icon: 'taxonomy', label: 'Taxonomia',          path: '/admin/taxonomy',    requiredCapability: 'taxonomy.view' },
  { icon: 'users',    label: 'Membros',            path: '/admin/memberships', requiredCapability: 'membership.view' },
  { icon: 'workflow', label: 'Rotas de aprovação', path: '/approval-routes',   requiredCapability: 'route.manage' },
];

function NavButton({ item }: { item: NavItem }) {
  const navigate = useNavigate();
  const match = useMatch({ path: item.path, end: item.path === '/' });
  const isActive = Boolean(match);

  return (
    <button
      className={`${styles.navItem}${isActive ? ` ${styles.navItemActive}` : ''}`}
      onClick={() => navigate(item.path)}
      title={item.label}
      aria-label={item.label}
      aria-current={isActive ? 'page' : undefined}
    >
      <Icon name={item.icon} size={18} />
    </button>
  );
}

export function Rail() {
  const user = useAuthStore((s) => s.user);
  const { handleLogout } = useAuthSession();
  const avatarName = user ? user.displayName || user.email || user.username : '';

  // Same capability source as the route guard (AppShell reads
  // `user.capabilities` the same way; `useHasCapability` is its single-cap
  // hook form, which cannot be called per item in a list).
  const adminItems = useMemo(() => {
    const granted = user?.capabilities ?? [];
    return ADMIN_NAV_ITEMS.filter(
      (item) => !item.requiredCapability || granted.includes(item.requiredCapability),
    );
  }, [user?.capabilities]);

  return (
    <aside className={styles.rail}>
      <div className={styles.logo}>
        <span
          style={{
            width: 28, height: 28, borderRadius: 6, background: 'var(--brand)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
          }}
        >
          <Icon name="docs" size={14} style={{ color: 'white' }} />
        </span>
      </div>

      <nav className={styles.nav}>
        {NAV_ITEMS.map((item) => (
          <NavButton key={item.path} item={item} />
        ))}
        {adminItems.length > 0 && (
          <>
            <div className={styles.navSpacer} />
            <div className={styles.divider} />
            <div className={styles.navGroup} role="group" aria-label="Administração">
              {adminItems.map((item) => (
                <NavButton key={item.path} item={item} />
              ))}
            </div>
          </>
        )}
      </nav>

      <div className={styles.bottom}>
        <div className={styles.divider} />
        {user && (
          <span title={avatarName}>
            <Avatar name={avatarName} size="sm" />
          </span>
        )}
        <button
          className={styles.logoutBtn}
          onClick={() => void handleLogout()}
          title="Sair"
          aria-label="Sair"
        >
          <Icon name="arrow" size={16} style={{ transform: 'rotate(180deg)' }} />
        </button>
      </div>
    </aside>
  );
}
