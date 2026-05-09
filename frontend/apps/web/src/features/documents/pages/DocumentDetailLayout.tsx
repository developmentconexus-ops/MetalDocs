import { NavLink, Outlet } from 'react-router-dom';
import styles from './DocumentDetailLayout.module.css';

export function DocumentDetailLayout() {
  return (
    <div className={styles.root}>
      <nav className={styles.tabStrip}>
        <NavLink
          to="."
          end
          className={({ isActive }) =>
            isActive ? `${styles.tab} ${styles.tabActive}` : styles.tab
          }
        >
          Documento
        </NavLink>
        <NavLink
          to="distribution"
          className={({ isActive }) =>
            isActive ? `${styles.tab} ${styles.tabActive}` : styles.tab
          }
        >
          Distribuição
        </NavLink>
      </nav>
      <Outlet />
    </div>
  );
}
