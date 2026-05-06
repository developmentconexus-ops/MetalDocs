import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ActivityPanel } from '../components/ActivityPanel';
import { AuthorCell } from '../components/AuthorCell';
import { LibraryFilterTabs } from '../components/LibraryFilterTabs';
import { LibraryStatCards } from '../components/LibraryStatCards';
import { LibrarySidebar } from '../components/LibrarySidebar';
import { PageSizeSelector } from '../components/PageSizeSelector';
import { Pagination } from '../components/Pagination';
import { useLibraryQuery } from '../queries/useLibraryQuery';
import { useLibraryStatsQuery } from '../queries/useLibraryStatsQuery';
import { filterToStatus, type LibraryFilter } from '../lib/libraryStatus';
import { StatusPill, type DocumentStatus } from '../../../components/ui/StatusPill';
import { ApiError, resolveErrorMessage } from '../../../lib/api/errors';
import { useDebouncedValue } from '../../../lib/hooks/useDebouncedValue';
import styles from './LibraryPage.module.css';

const PAGE_SIZE_KEY = 'metaldocs.library.pageSize';
const ACTIVITY_OPEN_KEY = 'metaldocs.library.activityOpen';
const PAGE_SIZE_OPTIONS = [10, 20, 50] as const;
type PageSize = (typeof PAGE_SIZE_OPTIONS)[number];

function readStoredPageSize(): PageSize {
  if (typeof window === 'undefined') return 20;
  const raw = window.localStorage.getItem(PAGE_SIZE_KEY);
  const parsed = raw ? Number(raw) : NaN;
  return PAGE_SIZE_OPTIONS.includes(parsed as PageSize) ? (parsed as PageSize) : 20;
}

function readStoredActivityOpen(): boolean {
  if (typeof window === 'undefined') return true;
  const raw = window.localStorage.getItem(ACTIVITY_OPEN_KEY);
  return raw === null ? true : raw === 'true';
}

export default function LibraryPage(): JSX.Element {
  const navigate = useNavigate();
  const [selectedArea, setSelectedArea] = useState<string | null>(null);
  const [activeFilter, setActiveFilter] = useState<LibraryFilter>('todos');
  const [page, setPage] = useState(1);
  // Lazy init — reads localStorage once on mount, prevents hydration flash.
  const [pageSize, setPageSize] = useState<PageSize>(readStoredPageSize);
  const [activityOpen, setActivityOpen] = useState<boolean>(readStoredActivityOpen);
  const [searchQuery, setSearchQuery] = useState('');
  const debouncedQuery = useDebouncedValue(searchQuery, 300);

  useEffect(() => {
    window.localStorage.setItem(PAGE_SIZE_KEY, String(pageSize));
  }, [pageSize]);

  // Reset to page 1 whenever the debounced search query changes — avoids
  // landing on page N of a smaller filtered result set.
  useEffect(() => {
    setPage(1);
  }, [debouncedQuery]);

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
    window.localStorage.setItem(ACTIVITY_OPEN_KEY, String(next));
  }

  function openDocument(id: string) {
    navigate(`/documents/${id}`);
  }

  const libraryQuery = useLibraryQuery({
    page,
    pageSize,
    status,
    areaCode: selectedArea ?? undefined,
    q: debouncedQuery.trim() || undefined,
  });
  const statsQuery = useLibraryStatsQuery();

  const items = libraryQuery.data?.items ?? [];
  const total = libraryQuery.data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const statsByStatus = statsQuery.data?.byStatus ?? {};
  const statsByArea = statsQuery.data?.byArea ?? {};

  const errorMessage =
    libraryQuery.isError
      ? resolveErrorMessage(
          libraryQuery.error instanceof ApiError ? libraryQuery.error.code : undefined,
          libraryQuery.error instanceof Error ? libraryQuery.error.message : undefined,
        )
      : null;

  return (
    <div className={`${styles.root} ${activityOpen ? styles.withActivity : ''}`}>
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
          onSearchChange={setSearchQuery}
        />
      </aside>

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
          onTabChange={handleFilterChange}
          statsByStatus={statsByStatus}
          total={total}
        />

        <section className={styles.tableCard}>
          <div className={styles.tableHeader}>
            <span>Código</span>
            <span>Título</span>
            <span>Área</span>
            <span>Estado</span>
            <span>Autor</span>
            <span>Rev.</span>
            <span />
          </div>

          {libraryQuery.isPending ? (
            <div className={styles.empty}>Carregando...</div>
          ) : null}

          {errorMessage ? (
            <div className={styles.empty} role="alert">{errorMessage}</div>
          ) : null}

          {!libraryQuery.isPending && !libraryQuery.isError && items.length === 0 ? (
            <div className={styles.empty}>Nenhum documento encontrado.</div>
          ) : null}

          {items.map((d) => (
            <div
              key={d.ID}
              role="button"
              tabIndex={0}
              className={styles.row}
              onClick={() => openDocument(d.ID)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  openDocument(d.ID);
                }
              }}
            >
              <span className={styles.codeLink}>{d.Code}</span>
              <span className={styles.nameCell}>{d.Name}</span>
              <span className={styles.metaCell}>{d.ProcessAreaCodeSnapshot ?? '–'}</span>
              <StatusPill status={d.Status as DocumentStatus} />
              <AuthorCell name={d.CreatedBy} />
              <span className={styles.monoCell}>v{d.RevisionVersion}</span>
              <button
                type="button"
                className={styles.moreBtn}
                aria-label="Mais opções"
                onClick={(e) => e.stopPropagation()}
              >
                ···
              </button>
            </div>
          ))}

          <footer className={styles.tableFooter}>
            <PageSizeSelector
              pageSize={pageSize}
              options={[...PAGE_SIZE_OPTIONS]}
              onPageSizeChange={(size) => { setPageSize(size as PageSize); setPage(1); }}
            />
            <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
          </footer>
        </section>
      </main>

      {activityOpen ? <ActivityPanel onClose={toggleActivity} /> : null}
    </div>
  );
}
