import { useState } from 'react';
import styles from './EditorMetaSidebar.module.css';
import { CodeChip } from '../../../components/ui/CodeChip';
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

type EditorMetaSidebarProps = {
  open: boolean;
  onToggle: () => void;
  code?: string;
  profileLabel?: string;
  areaLabel?: string;
  visibilityLabel?: string;
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
  history = [],
  approvalChain = null,
  documentStatus = '',
}: EditorMetaSidebarProps) {
  const [historyExpanded, setHistoryExpanded] = useState(false);
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
        <aside className={styles.sidebar} aria-label="Metadados do documento">
          <section className={styles.section}>
            <div className={styles.sectionHeader}>Metadados</div>
            <div className={styles.metaRows}>
              {code ? (
                <div className={styles.metaRow}>
                  <span className={styles.metaLabel}>Codigo</span>
                  <CodeChip>{code}</CodeChip>
                </div>
              ) : null}
              <div className={styles.metaRow}>
                <span className={styles.metaLabel}>Perfil</span>
                <span className={styles.metaValue}>{profileLabel ?? '---'}</span>
              </div>
              <div className={styles.metaRow}>
                <span className={styles.metaLabel}>Area</span>
                <span className={styles.metaValue}>{areaLabel ?? '---'}</span>
              </div>
              <div className={styles.metaRow}>
                <span className={styles.metaLabel}>Visibilidade</span>
                <span className={styles.metaValue}>{visibilityLabel ?? '---'}</span>
              </div>
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
        </aside>
      )}
    </div>
  );
}
