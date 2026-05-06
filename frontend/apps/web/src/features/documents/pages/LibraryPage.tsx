import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ActivityPanel } from '../components/ActivityPanel';
import { LibraryFilterTabs } from '../components/LibraryFilterTabs';
import { LibraryStatCards } from '../components/LibraryStatCards';
import { LibrarySidebar } from '../components/LibrarySidebar';
import type { LibraryFilter } from '../components/LibrarySidebar';
import { PageSizeSelector } from '../components/PageSizeSelector';
import { Pagination } from '../components/Pagination';
import { useLibraryQuery } from '../queries/useLibraryQuery';
import { useLibraryStatsQuery } from '../queries/useLibraryStatsQuery';
import styles from './LibraryPage.module.css';

const PAGE_SIZE_KEY = 'metaldocs.library.pageSize';
const ACTIVITY_OPEN_KEY = 'metaldocs.library.activityOpen';
const PAGE_SIZE_OPTIONS = [10, 20, 50] as const;

function readStoredPageSize(): number {
  const raw = localStorage.getItem(PAGE_SIZE_KEY);
  const parsed = raw ? Number(raw) : NaN;
  return PAGE_SIZE_OPTIONS.includes(parsed as (typeof PAGE_SIZE_OPTIONS)[number]) ? parsed : 20;
}

function readStoredActivityOpen(): boolean {
  const raw = localStorage.getItem(ACTIVITY_OPEN_KEY);
  // Default open when never set
  return raw === null ? true : raw === 'true';
}

function filterToStatus(filter: LibraryFilter): string | undefined {
  switch (filter) {
    case 'rascunhos':  return 'draft';
    case 'em_revisao': return 'under_review';
    case 'aprovados':  return 'approved';
    case 'publicados': return 'published';
    case 'rejeitados': return 'rejected';
    case 'obsoletos':  return 'obsolete';
    default:           return undefined;
  }
}

export default function LibraryPage(): JSX.Element {
  const navigate = useNavigate();
  const [selectedArea, setSelectedArea] = useState<string | null>(null);
  const [activeFilter, setActiveFilter] = useState<LibraryFilter>('todos');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [activityOpen, setActivityOpen] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    setPageSize(readStoredPageSize());
    setActivityOpen(readStoredActivityOpen());
  }, []);

  useEffect(() => {
    localStorage.setItem(PAGE_SIZE_KEY, String(pageSize));
  }, [pageSize]);

  const status = useMemo(() => filterToStatus(activeFilter), [activeFilter]);

  function handleFilterChange(f: LibraryFilter) {
    setActiveFilter(f);
    setPage(1);
  }

  function handleAreaChange(code: string | null) {
    setSelectedArea(code);
    setPage(1);
  }

  function toggleActivity() {
    const next = !activityOpen;
    setActivityOpen(next);
    localStorage.setItem(ACTIVITY_OPEN_KEY, String(next));
  }

  const libraryQuery = useLibraryQuery({ page, pageSize, status, areaCode: selectedArea ?? undefined });
  const statsQuery = useLibraryStatsQuery();

  const items = libraryQuery.data?.items ?? [];
  const total = libraryQuery.data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const statsByStatus = statsQuery.data?.byStatus ?? {};
  const statsByArea = statsQuery.data?.byArea ?? {};

  return (
    <div className={`${styles.root} ${activityOpen ? styles.withActivity : ''}`}>

      {/* Dark left sidebar */}
      <aside className={styles.sidebar}>
        <LibrarySidebar
          activeFilter={activeFilter}
          onFilterChange={handleFilterChange}
          selectedArea={selectedArea}
          onAreaChange={handleAreaChange}
          statsByStatus={statsByStatus}
          statsByArea={statsByArea}
          totalDocuments={total}
          searchQuery={searchQuery}
          onSearchChange={(q) => { setSearchQuery(q); setPage(1); }}
        />
      </aside>

      {/* Main content */}
      <main className={styles.main}>
        <header className={styles.header}>
          <p className={styles.kicker}>Documentos · Biblioteca</p>
          <div className={styles.titleRow}>
            <h1 className={styles.title}>Acervo<br />controlado</h1>
            <div className={styles.titleMeta}>
              <span className={styles.metaPair}>
                <span className={styles.metaValue}>{total.toLocaleString('pt-BR')}</span>
                <span className={styles.metaLabel}>documentos</span>
              </span>
              <span className={styles.metaDivider} aria-hidden="true" />
              <span className={styles.metaPair}>
                <span className={styles.metaLabel}>Última fanout</span>
                <span className={styles.metaValue}>{new Date().toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })}</span>
              </span>
              <button type="button" className={styles.activityButton} onClick={toggleActivity}>
                {activityOpen ? (
                  <>
                    <svg width="13" height="13" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                      <path d="M13 5l-5 5 5 5" />
                    </svg>
                    Recolher atividade
                  </>
                ) : (
                  <>
                    <svg width="13" height="13" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                      <circle cx="10" cy="10" r="7" />
                      <path d="M10 6v4l3 2" />
                      <path d="M3 10a7 7 0 0 1 12-5" />
                    </svg>
                    Mostrar atividade
                  </>
                )}
              </button>
            </div>
          </div>
          <p className={styles.description}>
            Documentos controlados por código, perfil e área. Filtre por estado para revisões pendentes ou trilha de auditoria.
          </p>
        </header>

        <LibraryStatCards total={total} statsByStatus={statsByStatus} />

        <LibraryFilterTabs
          activeTab={activeFilter}
          onTabChange={(tab) => handleFilterChange(tab as LibraryFilter)}
          statsByStatus={statsByStatus}
          total={total}
        />

        <section className={styles.tableCard}>
          <div className={styles.tableHeader}>
            <span>Código</span>
            <span>Título</span>
            <span>Área</span>
            <span>Perfil</span>
            <span>Rev.</span>
          </div>

          {libraryQuery.isPending ? (
            <div className={styles.empty}>Carregando...</div>
          ) : null}

          {libraryQuery.isError ? (
            <div className={styles.empty}>Erro ao carregar documentos.</div>
          ) : null}

          {!libraryQuery.isPending && !libraryQuery.isError && items.length === 0 ? (
            <div className={styles.empty}>Nenhum documento encontrado.</div>
          ) : null}

          {items.map((d) => (
            <button
              key={d.ID}
              type="button"
              className={styles.row}
              onClick={() => navigate(`/documents/${d.ID}`)}
            >
              <span className={styles.codeLink}>{d.Code}</span>
              <span className={styles.nameCell}>{d.Name}</span>
              <span className={styles.metaCell}>{d.ProcessAreaCodeSnapshot ?? '–'}</span>
              <span className={styles.metaCell}>{d.ProfileCodeSnapshot ?? '–'}</span>
              <span className={styles.metaCell}>v{d.RevisionVersion}</span>
            </button>
          ))}

          <footer className={styles.tableFooter}>
            <PageSizeSelector
              pageSize={pageSize}
              options={[...PAGE_SIZE_OPTIONS]}
              onPageSizeChange={(size) => { setPageSize(size); setPage(1); }}
            />
            <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
          </footer>
        </section>
      </main>

      {/* Right activity panel */}
      {activityOpen ? <ActivityPanel onClose={toggleActivity} /> : null}
    </div>
  );
}
