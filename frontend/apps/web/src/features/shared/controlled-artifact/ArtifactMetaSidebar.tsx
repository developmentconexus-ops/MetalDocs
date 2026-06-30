import { useState } from "react";
import styles from "./ArtifactMetaSidebar.module.css";
import { Avatar } from "../../../components/ui/Avatar";
import { formatShortDate } from "../../../lib/format/dates";
import { formatFileSize } from "../../../lib/format/fileSize";
import type { ApprovalChainItem, ArtifactMetaModel, LifecycleStatus, VersionHistoryItem } from "./types";

const APPROVAL_BADGE_LABELS: Record<string, string> = {
  approved: "aprovou",
  rejected: "rejeitou",
  active: "proximo",
  waiting: "aguarda",
};

const MAX_COLLAPSED_HISTORY_ITEMS = 3;

function formatPageSizeSummary(pageCount: number | null, fileSizeBytes: number | null): string | null {
  const parts: string[] = [];
  if (typeof pageCount === "number" && Number.isFinite(pageCount) && pageCount > 0) {
    parts.push(`${pageCount} ${pageCount === 1 ? "pagina" : "paginas"}`);
  }
  if (typeof fileSizeBytes === "number" && Number.isFinite(fileSizeBytes) && fileSizeBytes >= 0) {
    parts.push(formatFileSize(fileSizeBytes));
  }
  return parts.length ? parts.join(" · ") : null;
}

interface ArtifactMetaSidebarProps {
  open: boolean;
  onToggle: () => void;
  loading?: boolean;
  code?: string | null;
  status?: LifecycleStatus | null;
  meta: ArtifactMetaModel;
  approvalChain?: ApprovalChainItem[] | null;
  lineage?: VersionHistoryItem[];
}

export function ArtifactMetaSidebar({
  open,
  onToggle,
  loading = false,
  code,
  status,
  meta,
  approvalChain = null,
  lineage = [],
}: ArtifactMetaSidebarProps) {
  const [historyExpanded, setHistoryExpanded] = useState(false);
  const pageSizeSummary = formatPageSizeSummary(meta.pageCount, meta.fileSizeBytes);
  const recentNonCurrentHistory = [...lineage]
    .filter((item) => !item.isCurrent)
    .sort((a, b) => new Date(b.createdAt ?? 0).getTime() - new Date(a.createdAt ?? 0).getTime());
  const visibleHistory = historyExpanded || lineage.length <= MAX_COLLAPSED_HISTORY_ITEMS
    ? lineage
    : [
        ...lineage.filter((item) => item.isCurrent),
        ...recentNonCurrentHistory.slice(0, MAX_COLLAPSED_HISTORY_ITEMS - 1),
      ].slice(0, MAX_COLLAPSED_HISTORY_ITEMS);

  const showApprovalChain = status === "under_review" && approvalChain != null && approvalChain.length > 0;

  return (
    <div className={styles.sidebarOuter}>
      <button
        type="button"
        className={styles.toggleTab}
        onClick={onToggle}
        aria-label={open ? "Fechar painel de metadados" : "Abrir painel de metadados"}
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
              {loading ? (
                <div className={styles.metaRows}>
                  <div className={styles.metaRow}>
                    <span className={styles.metaValue}>Carregando metadados do documento</span>
                  </div>
                </div>
              ) : (
                <div className={styles.metaRows}>
                  {code ? (
                    <div className={styles.metaRow}>
                      <span className={styles.metaLabel}>Codigo</span>
                      <span className={styles.metaCode}>{code}</span>
                    </div>
                  ) : null}
                  <div className={styles.metaRow}>
                    <span className={styles.metaLabel}>Tipo</span>
                    <span className={styles.metaValue}>{meta.profileLabel ?? "---"}</span>
                  </div>
                  <div className={styles.metaRow}>
                    <span className={styles.metaLabel}>Area responsavel</span>
                    <span className={styles.metaValue}>{meta.areaLabel ?? "---"}</span>
                  </div>
                  <div className={styles.metaRow}>
                    <span className={styles.metaLabel}>Visibilidade</span>
                    <span className={styles.metaValue}>{meta.visibilityLabel ?? "---"}</span>
                  </div>
                  {pageSizeSummary ? (
                    <div className={styles.metaRow}>
                      <span className={styles.metaLabel}>Paginas</span>
                      <span className={styles.metaValue}>{pageSizeSummary}</span>
                    </div>
                  ) : null}
                </div>
              )}
            </section>
            <div className={styles.divider} />
            <section className={styles.section}>
              <div className={styles.sectionHeader}>Revisoes</div>
              <div className={styles.revisionList} aria-label="Historico de revisoes">
                {visibleHistory.map((item) => (
                  <div
                    key={`${item.versionNumber}-${item.revisionNumber ?? "r"}-${item.revisionLabel ?? ""}`}
                    className={`${styles.revisionRow} ${item.isCurrent ? styles.revisionRowCurrent : ""}`}
                  >
                    <span className={styles.revisionMarker} aria-hidden="true" />
                    <div className={styles.revisionBody}>
                      <span className={styles.revisionCode}>{item.revisionLabel}</span>
                      <span className={styles.revisionTitle}>{item.title}</span>
                    </div>
                    <time className={styles.revisionDate} dateTime={item.createdAt ?? undefined}>
                      {formatShortDate(item.createdAt)}
                    </time>
                  </div>
                ))}
              </div>
              {lineage.length > MAX_COLLAPSED_HISTORY_ITEMS ? (
                <button
                  type="button"
                  className={styles.historyToggle}
                  onClick={() => setHistoryExpanded((expanded) => !expanded)}
                >
                  {historyExpanded ? "Ver menos revisoes" : "Ver todas as revisoes"}
                </button>
              ) : null}
            </section>
            {showApprovalChain ? (
              <>
                <div className={styles.divider} />
                <section className={styles.section}>
                  <div className={styles.sectionHeader}>Proximos aprovadores</div>
                  <div className={styles.approverList}>
                    {approvalChain?.map((item, idx) => {
                      const badgeClassName =
                        item.status === "approved"
                          ? styles.approverBadgeApproved
                          : item.status === "rejected"
                            ? styles.approverBadgeRejected
                            : item.status === "active"
                              ? styles.approverBadgeNext
                              : styles.approverBadgeWait;
                      return (
                        <div key={`${item.stageIndex}:${item.actorUserId ?? `_${idx}`}:${item.status}`} className={styles.approverRow}>
                          <Avatar name={item.actorDisplay ?? ""} size="sm" />
                          <div className={styles.approverInfo}>
                            <span className={styles.approverName}>{item.actorDisplay}</span>
                            <span className={styles.approverRole}>{item.label}</span>
                          </div>
                          <span className={`${styles.approverBadge} ${badgeClassName}`}>
                            {APPROVAL_BADGE_LABELS[item.status] ?? item.status}
                          </span>
                        </div>
                      );
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
