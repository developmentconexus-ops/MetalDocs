import { useEffect, useMemo, useState } from 'react';
import { CodeChip } from '../../../components/ui/CodeChip';
import { StatusPill } from '../../../components/ui/StatusPill';
import LibraryAreaTree from '../components/LibraryAreaTree';
import { LibraryFilterTabs } from '../components/LibraryFilterTabs';
import { PageSizeSelector } from '../components/PageSizeSelector';
import { Pagination } from '../components/Pagination';
import { useLibraryQuery } from '../queries/useLibraryQuery';
import { useLibraryStatsQuery } from '../queries/useLibraryStatsQuery';
import styles from './LibraryPage.module.css';

type LibraryFilter = 'todos' | 'rascunhos' | 'em_revisao' | 'publicados' | 'rejeitados' | 'obsoletos';

const PAGE_SIZE_KEY = 'metaldocs.library.pageSize';
const ACTIVITY_OPEN_KEY = 'metaldocs.library.activityOpen';
const PAGE_SIZE_OPTIONS = [10, 20, 50] as const;

function readStoredPageSize(): number {
  const raw = localStorage.getItem(PAGE_SIZE_KEY);
  const parsed = raw ? Number(raw) : NaN;
  return PAGE_SIZE_OPTIONS.includes(parsed as (typeof PAGE_SIZE_OPTIONS)[number]) ? parsed : 20;
}

function readStoredActivityOpen(): boolean {
  return localStorage.getItem(ACTIVITY_OPEN_KEY) === 'true';
}

function filterToStatus(filter: LibraryFilter): string | undefined {
  switch (filter) {
    case 'rascunhos':
      return 'draft';
    case 'em_revisao':
      return 'under_review';
    case 'publicados':
      return 'published';
    case 'rejeitados':
      return 'rejected';
    case 'obsoletos':
      return 'obsolete';
    default:
      return undefined;
  }
}

export default function LibraryPage(): JSX.Element {
  const [selectedArea, setSelectedArea] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<LibraryFilter>('todos');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [activityOpen, setActivityOpen] = useState(false);

  useEffect(() => {
    setPageSize(readStoredPageSize());
    setActivityOpen(readStoredActivityOpen());
  }, []);

  useEffect(() => {
    localStorage.setItem(PAGE_SIZE_KEY, String(pageSize));
  }, [pageSize]);

  useEffect(() => {
    localStorage.setItem(ACTIVITY_OPEN_KEY, String(activityOpen));
  }, [activityOpen]);

  const status = useMemo(() => filterToStatus(activeTab), [activeTab]);

  const libraryQuery = useLibraryQuery({
    page,
    pageSize,
    status,
    areaCode: selectedArea ?? undefined,
  });
  const statsQuery = useLibraryStatsQuery();

  const items = libraryQuery.data?.items ?? [];
  const total = libraryQuery.data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  return (
    <div className={styles.root}>
      <aside className={styles.leftAside}>
        <LibraryAreaTree
          selected={selectedArea}
          onSelect={(areaCode) => {
            setSelectedArea(areaCode);
            setPage(1);
          }}
          counts={statsQuery.data?.byArea ?? {}}
        />
      </aside>

      <main className={styles.main}>
        <header className={styles.header}>
          <div>
            <p className={styles.kicker}>Documentos · Biblioteca</p>
            <h1 className={styles.title}>Acervo controlado</h1>
          </div>
          <div className={styles.headerMeta}>
            <span className={styles.totalBadge}>{total} documentos</span>
            <button
              type="button"
              className={styles.activityButton}
              onClick={() => setActivityOpen((current) => !current)}
            >
              {activityOpen ? 'Ocultar atividade' : 'Mostrar atividade'}
            </button>
          </div>
        </header>

        <LibraryFilterTabs
          activeTab={activeTab}
          onTabChange={(tab) => {
            setActiveTab(tab as LibraryFilter);
            setPage(1);
          }}
        />

        <section className={styles.tableCard}>
          <div className={styles.tableHeader}>
            <span>Código</span>
            <span>Nome</span>
            <span>Área</span>
            <span>Perfil</span>
            <span>Revisão</span>
            <span>Status</span>
            <span>Atualizado</span>
          </div>

          {libraryQuery.isPending ? <div className={styles.empty}>Carregando...</div> : null}
          {libraryQuery.isError ? <div className={styles.empty}>Erro ao carregar documentos.</div> : null}

          {!libraryQuery.isPending && !libraryQuery.isError && items.length === 0 ? (
            <div className={styles.empty}>Nenhum documento encontrado.</div>
          ) : null}

          {items.map((d) => (
            <div key={d.ID} className={styles.row}>
              <CodeChip>{d.Code}</CodeChip>
              <span className={styles.nameCell}>{d.Name}</span>
              <span>{d.ProcessAreaCodeSnapshot ?? '-'}</span>
              <span>{d.ProfileCodeSnapshot ?? '-'}</span>
              <span>v{d.RevisionVersion}</span>
              <StatusPill status={d.Status} />
              <span>{new Date(d.UpdatedAt).toLocaleString('pt-BR')}</span>
            </div>
          ))}

          <footer className={styles.tableFooter}>
            <PageSizeSelector
              pageSize={pageSize}
              options={[...PAGE_SIZE_OPTIONS]}
              onPageSizeChange={(size) => {
                setPageSize(size);
                setPage(1);
              }}
            />
            <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
          </footer>
        </section>
      </main>

      {activityOpen ? (
        <aside className={styles.rightAside}>
          <p className={styles.activityPlaceholder}>Atividade em breve.</p>
        </aside>
      ) : null}
    </div>
  );
}
