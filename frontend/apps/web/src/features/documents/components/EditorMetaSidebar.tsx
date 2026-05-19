import { useState } from 'react';
import styles from './EditorMetaSidebar.module.css';
import { Avatar } from '../../../components/ui/Avatar';
import { displayRevisionTitle, formatShortDate } from '../lib/documentDetailMeta';

export type EditorSidebarRevisionItem = {
  documentId: string;
  revisionCode: string;
  revisionTitle: string;
  status: string;
  createdAt: string;
  isCurrent: boolean;
};

export type EditorSidebarApprovalStage = {
  id: string;
  label: string;
  status: string;
  actors: Array<{
    user_id: string;
    display_name: string;
    status: string;
    decision?: string | null;
  }>;
  signoffs: Array<{
    id: string;
    actor_user_id: string;
    decision: string;
  }>;
};

const APPROVAL_BADGE_LABELS: Record<string, string> = {
  approved: 'aprovou',
  rejected: 'rejeitou',
  active: 'proximo',
  waiting: 'aguarda',
};

const MAX_COLLAPSED_HISTORY_ITEMS = 3;

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  const units = ['KB', 'MB', 'GB'];
  let value = bytes / 1024;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }
  return `${new Intl.NumberFormat('pt-BR', { maximumFractionDigits: 1 }).format(value)} ${units[unitIndex]}`;
}

function formatPageSizeSummary(pageCount?: number | null, fileSizeBytes?: number | null): string | null {
  const parts: string[] = [];
  if (typeof pageCount === 'number' && Number.isFinite(pageCount) && pageCount > 0) {
    parts.push(`${pageCount} ${pageCount === 1 ? 'pagina' : 'paginas'}`);
  }
  if (typeof fileSizeBytes === 'number' && Number.isFinite(fileSizeBytes) && fileSizeBytes >= 0) {
    parts.push(formatFileSize(fileSizeBytes));
  }
  return parts.length ? parts.join(' \u00b7 ') : null;
}

type EditorMetaSidebarProps = {
  open: boolean;
  onToggle: () => void;
  code?: string;
  profileLabel?: string;
  areaLabel?: string;
  visibilityLabel?: string;
  fileSizeBytes?: number | null;
  pageCount?: number | null;
  history?: EditorSidebarRevisionItem[];
  approvalChain?: { stages: EditorSidebarApprovalStage[] } | null;
  documentStatus?: string;
};

export function EditorMetaSidebar({
  open,
  onToggle,
  code,
  profileLabel,
  areaLabel,
  visibilityLabel,
  fileSizeBytes,
  pageCount,
  history = [],
  approvalChain = null,
  documentStatus = '',
}: EditorMetaSidebarProps) {
  const [historyExpanded, setHistoryExpanded] = useState(false);
  const pageSizeSummary = formatPageSizeSummary(pageCount, fileSizeBytes);
  const recentNonCurrentHistory = [...history]
    .filter((item) => !item.isCurrent)
    .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
  const visibleHistory = historyExpanded || history.length <= MAX_COLLAPSED_HISTORY_ITEMS
    ? history
    : [
        ...history.filter((item) => item.isCurrent),
        ...recentNonCurrentHistory.slice(0, MAX_COLLAPSED_HISTORY_ITEMS - 1),
      ].slice(0, MAX_COLLAPSED_HISTORY_ITEMS);

  return (
    <div className={styles.sidebarOuter}>
      <button
        type="button"
        className={styles.toggleTab}
        onClick={onToggle}
        aria-label={open ? 'Fechar painel de metadados' : 'Abrir painel de metadados'}
        aria-expanded={open}
      >
        <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          {open
            ? <polyline points="9 18 15 12 9 6" />
            : <polyline points="15 18 9 12 15 6" />}
        </svg>
      </button>
      {open && (
        <aside className={styles.sidebar} aria-label="Identificacao do documento">
          <div className={styles.panelFrame}>
            <section className={styles.section}>
              <div className={styles.sectionHeader}>Identificacao</div>
              <div className={styles.metaRows}>
                {code ? (
                  <div className={styles.metaRow}>
                    <span className={styles.metaLabel}>Codigo</span>
                    <span className={styles.metaCode}>{code}</span>
                  </div>
                ) : null}
                <div className={styles.metaRow}>
                  <span className={styles.metaLabel}>Tipo</span>
                  <span className={styles.metaValue}>{profileLabel ?? '---'}</span>
                </div>
                <div className={styles.metaRow}>
                  <span className={styles.metaLabel}>Area responsavel</span>
                  <span className={styles.metaValue}>{areaLabel ?? '---'}</span>
                </div>
                <div className={styles.metaRow}>
                  <span className={styles.metaLabel}>Visibilidade</span>
                  <span className={styles.metaValue}>{visibilityLabel ?? '---'}</span>
                </div>
                {pageSizeSummary ? (
                  <div className={styles.metaRow}>
                    <span className={styles.metaLabel}>Paginas</span>
                    <span className={styles.metaValue}>{pageSizeSummary}</span>
                  </div>
                ) : null}
              </div>
            </section>
            <div className={styles.divider} />
            <section className={styles.section}>
              <div className={styles.sectionHeader}>Revisoes</div>
              <div className={styles.revisionList} aria-label="Historico de revisoes">
                {visibleHistory.map((item) => (
                  <div
                    key={item.documentId}
                    className={`${styles.revisionRow} ${item.isCurrent ? styles.revisionRowCurrent : ''}`}
                  >
                    <span className={styles.revisionMarker} aria-hidden="true" />
                    <div className={styles.revisionBody}>
                      <span className={styles.revisionCode}>{item.revisionCode}</span>
                      <span className={styles.revisionTitle}>
                        {displayRevisionTitle(item.revisionTitle, item.revisionCode)}
                      </span>
                    </div>
                    <time className={styles.revisionDate} dateTime={item.createdAt}>
                      {formatShortDate(item.createdAt)}
                    </time>
                  </div>
                ))}
              </div>
              {history.length > MAX_COLLAPSED_HISTORY_ITEMS ? (
                <button
                  type="button"
                  className={styles.historyToggle}
                  onClick={() => setHistoryExpanded((expanded) => !expanded)}
                >
                  {historyExpanded ? 'Ver menos revisoes' : 'Ver todas as revisoes'}
                </button>
              ) : null}
            </section>
            {documentStatus === 'under_review' && approvalChain ? (
              <>
                <div className={styles.divider} />
                <section className={styles.section}>
                  <div className={styles.sectionHeader}>Proximos aprovadores</div>
                  <div className={styles.approverList}>
                    {approvalChain.stages.map((stage) => {
                      return stage.actors.map((actor) => {
                        const badgeClassName =
                          actor.status === 'approved'
                            ? styles.approverBadgeApproved
                            : actor.status === 'rejected'
                              ? styles.approverBadgeRejected
                              : actor.status === 'active'
                                ? styles.approverBadgeNext
                                : styles.approverBadgeWait;
                        return (
                          <div key={`${stage.id}:${actor.user_id}:${actor.status}`} className={styles.approverRow}>
                            <Avatar name={actor.display_name} size="sm" />
                            <div className={styles.approverInfo}>
                              <span className={styles.approverName}>{actor.display_name}</span>
                              <span className={styles.approverRole}>{stage.label}</span>
                            </div>
                            <span className={`${styles.approverBadge} ${badgeClassName}`}>
                              {APPROVAL_BADGE_LABELS[actor.status] ?? actor.status}
                            </span>
                          </div>
                        );
                      });
                    })}
                  </div>
                </section>
              </>
            ) : null}
            <div className={styles.panelFill} aria-hidden="true">
              <span className={styles.panelFillLine} />
              <span className={styles.panelFillText}>Dossie governado</span>
            </div>
          </div>
        </aside>
      )}
    </div>
  );
}
