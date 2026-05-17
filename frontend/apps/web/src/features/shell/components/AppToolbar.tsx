import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Icon } from '../../../components/ui/Icon';
import { useAuthStore } from '../../../store/auth.store';
import styles from './AppToolbar.module.css';

export function AppToolbar() {
  const [search, setSearch] = useState('');
  const navigate = useNavigate();
  const user = useAuthStore((s) => s.user);

  function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (search.trim()) {
      navigate(`/documents?q=${encodeURIComponent(search.trim())}`);
    }
  }

  return (
    <header className={styles.toolbar}>
      <form onSubmit={handleSearch} className={styles.searchWrapper}>
        <span className={styles.searchIcon}>
          <Icon name="search" size={13} />
        </span>
        <input
          className={styles.searchInput}
          type="search"
          placeholder="Buscar documentos…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          aria-label="Buscar documentos"
        />
      </form>

      <div className={styles.spacer} />

      {user && (
        <span style={{ fontSize: 12, color: 'var(--text-muted)', fontFamily: 'var(--font-sans)' }}>
          {user.displayName || user.email}
        </span>
      )}

      <button className={styles.bellBtn} aria-label="Notificações" title="Notificações">
        <Icon name="bell" size={16} />
        {/* Notification badge — wired in Library/Dashboard blocks */}
      </button>

      <button
        className={styles.newDocBtn}
        onClick={() => navigate('/documents/new')}
        aria-label="Novo documento"
      >
        <Icon name="plus" size={14} />
        Novo documento
      </button>
    </header>
  );
}
