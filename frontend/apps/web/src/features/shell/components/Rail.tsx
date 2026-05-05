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
};

const NAV_ITEMS: NavItem[] = [
  { icon: 'home',     label: 'Início',     path: '/' },
  { icon: 'library',  label: 'Documentos', path: '/documents' },
  { icon: 'template', label: 'Templates',  path: '/templates-v2' },
  { icon: 'registry', label: 'Registro',   path: '/registry-v2' },
  { icon: 'inbox',    label: 'Aprovações', path: '/approvals' },
  { icon: 'audit',    label: 'Auditoria',  path: '/audit' },
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
      </nav>

      <div className={styles.bottom}>
        <div className={styles.divider} />
        {user && (
          <span title={user.displayName ?? user.email}>
            <Avatar name={user.displayName || user.email} size="sm" />
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
