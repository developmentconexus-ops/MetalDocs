import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthorCell } from '../components/AuthorCell';
import { LibraryFilterTabs } from '../components/LibraryFilterTabs';
import { LibraryStatCards } from '../components/LibraryStatCards';
import { LibrarySidebar } from '../components/LibrarySidebar';
import { PageSizeSelector } from '../components/PageSizeSelector';
import { useLibraryQuery } from '../queries/useLibraryQuery';
import { useLibraryStatsQuery } from '../queries/useLibraryStatsQuery';
import { filterToStatus, type LibraryFilter } from '../lib/libraryStatus';
import { StatusPill } from '../../../components/ui/StatusPill';
import { ApiError, resolveErrorMessage } from '../../../lib/api/errors';
import { useDebouncedValue } from '../../../lib/hooks/useDebouncedValue';
import styles from './LibraryPage.module.css';

const PAGE_SIZE_KEY = 'metaldocs.library.pageSize';
const PAGE_SIZE_OPTIONS = [10, 20, 50] as const;
type PageSize = (typeof PAGE_SIZE_OPTIONS)[number];

function readStoredPageSize(): PageSize {
  if (typeof window === 'undefined') return 20;
  const raw = window.localStorage.getItem(PAGE_SIZE_KEY);
  const parsed = raw ? Number(raw) : NaN;
  return PAGE_SIZE_OPTIONS.includes(parsed as PageSize) ? (parsed as PageSize) : 20;
}

export default function LibraryPage(): JSX.Element {
  const navigate = useNavigate();
  const [selectedArea, setSelectedArea] = useState<string | null>(null);
  const [activeFilter, setActiveFilter] = useState<LibraryFilter>('todos');
  // FD-2 keyset pagination: a client-side stack of opaque cursors. Entry 0 is
  // the first page (empty cursor); the active cursor is the last entry. Next
  // pushes the response's page.next_cursor; Prev pops. Page number = stack length.
  const [cursorStack, setCursorStack] = useState<string[]>(['']);
  // Lazy init — reads localStorage once on mount, prevents hydration flash.
  const [pageSize, setPageSize] = useState<PageSize>(readStoredPageSize);
  const [searchQuery, setSearchQuery] = useState('');
  const debouncedQuery = useDebouncedValue(searchQuery, 300);

  useEffect(() => {
    window.localStorage.setItem(PAGE_SIZE_KEY, String(pageSize));
  }, [pageSize]);

  // Reset to the first page whenever the debounced search query changes — avoids
  // landing on a stale cursor against a smaller filtered result set.
  useEffect(() => {
    setCursorStack(['']);
  }, [debouncedQuery]);

  const status = useMemo(() => filterToStatus(activeFilter), [activeFilter]);

  function resetToFirstPage() {
    setCursorStack(['']);
  }

  function handleFilterChange(f: LibraryFilter) {
    setActiveFilter(f);
    resetToFirstPage();
  }

  function handleAreaChange(code: string | null) {
    setSelectedArea(code);
    resetToFirstPage();
  }

  function openDocument(id: string) {
    navigate(`/documents/${id}`);
  }

  const activeCursor = cursorStack[cursorStack.length - 1];
  const libraryQuery = useLibraryQuery({
    cursor: activeCursor || undefined,
    limit: pageSize,
    status,
    area_code: selectedArea ?? undefined,
    q: debouncedQuery.trim() || undefined,
  });
  const statsQuery = useLibraryStatsQuery();

  const items = libraryQuery.data?.items ?? [];
  const total = libraryQuery.data?.total ?? 0;
  const pageNumber = cursorStack.length;
  const nextCursor = libraryQuery.data?.page?.next_cursor ?? null;
  const hasMore = libraryQuery.data?.page?.has_more ?? false;

  function goNextPage() {
    if (hasMore && nextCursor) setCursorStack((s) => [...s, nextCursor]);
  }
  function goPrevPage() {
    setCursorStack((s) => (s.length > 1 ? s.slice(0, -1) : s));
  }
  const statsByStatus = statsQuery.data?.by_status ?? {};
  const statsByArea = statsQuery.data?.by_area ?? {};

  const errorMessage =
    libraryQuery.isError
      ? resolveErrorMessage(
          libraryQuery.error instanceof ApiError ? libraryQuery.error.code : undefined,
          libraryQuery.error instanceof Error ? libraryQuery.error.message : undefined,
        )
      : null;

  return (
    <div className={styles.root}>
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
            </div>
          </div>
          <p className={styles.description}>
            Documentos controlados por código, perfil e área. Filtre por estado para revisões pendentes ou trilha de auditoria.
          </p>
        </header>

        <LibraryStatCards statsByStatus={statsByStatus} />

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
              key={d.id}
              role="button"
              tabIndex={0}
              className={styles.row}
              onClick={() => openDocument(d.id)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  openDocument(d.id);
                }
              }}
            >
              <span className={styles.codeLink}>{d.code}</span>
              <span className={styles.nameCell}>{d.name}</span>
              <span className={styles.metaCell}>{d.process_area_code_snapshot ?? '–'}</span>
              <StatusPill status={d.status} />
              <AuthorCell name={d.created_by} />
              <span className={styles.monoCell}>REV{String(d.revision_number).padStart(2, '0')}</span>
            </div>
          ))}

          <footer className={styles.tableFooter}>
            <PageSizeSelector
              pageSize={pageSize}
              options={[...PAGE_SIZE_OPTIONS]}
              onPageSizeChange={(size) => { setPageSize(size as PageSize); resetToFirstPage(); }}
            />
            <div className={styles.cursorPager}>
              <button
                type="button"
                className={styles.activityButton}
                onClick={goPrevPage}
                disabled={pageNumber <= 1 || libraryQuery.isFetching}
              >
                Anterior
              </button>
              <span className={styles.pagerLabel}>Página {pageNumber}</span>
              <button
                type="button"
                className={styles.activityButton}
                onClick={goNextPage}
                disabled={!hasMore || libraryQuery.isFetching}
              >
                Próxima
              </button>
            </div>
          </footer>
        </section>
      </main>
    </div>
  );
}
