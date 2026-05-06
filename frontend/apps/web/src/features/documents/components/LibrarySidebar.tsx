import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { listTaxonomyAreas } from '../../taxonomy/api/catalog';
import { QK } from '../../../lib/queryKeys';
import { LIBRARY_STATUSES, type LibraryFilter } from '../lib/libraryStatus';
import styles from './LibrarySidebar.module.css';

export type { LibraryFilter };

type Props = {
  activeFilter: LibraryFilter;
  onFilterChange: (f: LibraryFilter) => void;
  selectedArea: string | null;
  onAreaChange: (code: string | null) => void;
  statsByStatus: Record<string, number>;
  statsByArea: Record<string, number>;
  totalDocuments: number;
  searchQuery: string;
  onSearchChange: (q: string) => void;
};

// Map backend status → sidebar dot style. Lives next to the only consumer.
const STATUS_DOT: Record<string, string> = {
  draft:        'dotDraft',
  under_review: 'dotReview',
  approved:     'dotApproved',
  published:    'dotPublished',
  rejected:     'dotRejected',
  obsolete:     'dotObsolete',
};

export function LibrarySidebar({
  activeFilter,
  onFilterChange,
  selectedArea,
  onAreaChange,
  statsByStatus,
  statsByArea,
  totalDocuments,
  searchQuery,
  onSearchChange,
}: Props): JSX.Element {
  const navigate = useNavigate();
  const areasQuery = useQuery({
    queryKey: QK.taxonomy.areas(),
    queryFn: listTaxonomyAreas,
  });
  const areas = areasQuery.data?.items ?? [];

  return (
    <nav className={styles.root} aria-label="Navegação da biblioteca">
      {/* Header */}
      <div className={styles.header}>
        <span className={styles.headerTitle}>Documentos</span>
        <button
          type="button"
          className={styles.headerAdd}
          aria-label="Novo documento"
          onClick={() => navigate('/documents/new')}
        >
          +
        </button>
      </div>

      {/* Search */}
      <div className={styles.searchWrap}>
        <input
          type="text"
          className={styles.searchInput}
          placeholder="Buscar..."
          aria-label="Buscar na biblioteca"
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
        />
        <span className={styles.searchKbd}>⌘K</span>
      </div>

      {/* Visões */}
      <section className={styles.section}>
        <span className={styles.sectionLabel}>VISÕES</span>
        <button
          type="button"
          className={`${styles.navItem} ${activeFilter === 'todos' ? styles.navItemActive : ''}`}
          onClick={() => onFilterChange('todos')}
        >
          <span className={styles.navLabel}>Biblioteca</span>
          <span className={styles.navBadge}>{totalDocuments.toLocaleString('pt-BR')}</span>
        </button>
        <button type="button" className={`${styles.navItem} ${styles.navItemDisabled}`} disabled>
          <span className={styles.navLabel}>Meus documentos</span>
        </button>
        <button type="button" className={`${styles.navItem} ${styles.navItemDisabled}`} disabled>
          <span className={styles.navLabel}>Recentes</span>
        </button>
      </section>

      <div className={styles.divider} />

      {/* Estado */}
      <section className={styles.section}>
        <span className={styles.sectionLabel}>ESTADO</span>
        {LIBRARY_STATUSES.map((entry) => {
          const count = statsByStatus[entry.status] ?? 0;
          const isActive = activeFilter === entry.filter;
          return (
            <button
              key={entry.filter}
              type="button"
              className={`${styles.navItem} ${isActive ? styles.navItemActive : ''}`}
              onClick={() => onFilterChange(entry.filter)}
            >
              <span className={`${styles.dot} ${styles[STATUS_DOT[entry.status]]}`} aria-hidden="true" />
              <span className={styles.navLabel}>{entry.label}</span>
              {count > 0 ? <span className={styles.navBadge}>{count}</span> : null}
            </button>
          );
        })}
      </section>

      <div className={styles.divider} />

      {/* Áreas */}
      <section className={styles.section}>
        <span className={styles.sectionLabel}>ÁREAS</span>
        <button
          type="button"
          className={`${styles.navItem} ${selectedArea === null ? styles.navItemActive : ''}`}
          onClick={() => onAreaChange(null)}
        >
          <span className={styles.navLabel}>Todas</span>
        </button>
        {areas.map((area) => {
          const count = statsByArea[area.code] ?? 0;
          return (
            <button
              key={area.code}
              type="button"
              className={`${styles.navItem} ${selectedArea === area.code ? styles.navItemActive : ''}`}
              onClick={() => onAreaChange(area.code)}
            >
              <span className={styles.navLabel}>{area.name}</span>
              {count > 0 ? <span className={styles.navBadge}>{count}</span> : null}
            </button>
          );
        })}
      </section>

      {/* Footer */}
      <div className={styles.footer}>
        <span className={styles.footerDot} aria-hidden="true" />
        <span className={styles.footerText}>Sistema operacional</span>
      </div>
    </nav>
  );
}
