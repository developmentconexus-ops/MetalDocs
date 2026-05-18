import styles from './EditorMetaSidebar.module.css';
import { CodeChip } from '../../../components/ui/CodeChip';
import { Avatar } from '../../../components/ui/Avatar';
import { TimelineRail } from '../../../components/ui/TimelineRail';
import { formatShortDate } from '../lib/documentDetailMeta';

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
  signoffs: Array<{
    id: string;
    actor_user_id: string;
    decision: string;
  }>;
};

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
  const historyItems = history.map((item) => ({
    id: item.documentId,
    title: item.revisionCode,
    subtitle: item.revisionTitle,
    aside: formatShortDate(item.createdAt),
    active: item.isCurrent,
  }));

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
                <span className={styles.metaValue}>{profileLabel ?? '—'}</span>
              </div>
              <div className={styles.metaRow}>
                <span className={styles.metaLabel}>Area</span>
                <span className={styles.metaValue}>{areaLabel ?? '—'}</span>
              </div>
              <div className={styles.metaRow}>
                <span className={styles.metaLabel}>Visibilidade</span>
                <span className={styles.metaValue}>{visibilityLabel ?? '—'}</span>
              </div>
            </div>
          </section>
          <div className={styles.divider} />
          <section className={styles.section}>
            <div className={styles.sectionHeader}>Revisoes</div>
            <TimelineRail items={historyItems} ariaLabel="Historico de revisoes" variant="flat" />
          </section>
          {documentStatus === 'under_review' && approvalChain ? (
            <>
              <div className={styles.divider} />
              <section className={styles.section}>
                <div className={styles.sectionHeader}>Proximos aprovadores</div>
                <div className={styles.approverList}>
                  {approvalChain.stages.map((stage) => {
                    const primarySignoff = stage.signoffs[0];
                    return (
                      <div key={stage.id} className={styles.approverRow}>
                        <Avatar name={primarySignoff?.actor_user_id ?? stage.label} size="sm" />
                        <div className={styles.approverInfo}>
                          <span className={styles.approverName}>{primarySignoff?.actor_user_id ?? stage.label}</span>
                          <span className={styles.approverRole}>{stage.label}</span>
                        </div>
                        <span className={`${styles.approverBadge} ${stage.status === 'active' ? styles.approverBadgeNext : styles.approverBadgeWait}`}>
                          {stage.status === 'active' ? 'proximo' : stage.status}
                        </span>
                      </div>
                    );
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
