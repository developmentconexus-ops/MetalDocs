import { useState } from "react";
import styles from "./ArtifactMetaSidebar.module.css";
import { Avatar } from "../../../components/ui/Avatar";
import { formatShortDate } from "../../../lib/format/dates";
import { formatFileSize } from "../../../lib/format/fileSize";
import type { ApprovalChainItem, ArtifactMetaModel, VersionHistoryItem } from "./types";

const APPROVAL_BADGE_LABELS: Record<string, string> = {
  approved: "aprovou",
  rejected: "rejeitou",
  active: "próximo",
  waiting: "aguarda",
};

const MAX_COLLAPSED_HISTORY_ITEMS = 3;

/**
 * Honest-absence marker. NEVER render a value-shaped placeholder (the old
 * `---`) for a fact the API does carry — F-QA4-4. When the adapter supplies an
 * absence reason it becomes the `title` tooltip so the reader learns WHY the
 * field is empty instead of assuming the screen is broken.
 */
const EM_DASH = "—";

function formatPageSizeSummary(pageCount: number | null, fileSizeBytes: number | null): string | null {
  const parts: string[] = [];
  if (typeof pageCount === "number" && Number.isFinite(pageCount) && pageCount > 0) {
    parts.push(`${pageCount} ${pageCount === 1 ? "página" : "páginas"}`);
  }
  if (typeof fileSizeBytes === "number" && Number.isFinite(fileSizeBytes) && fileSizeBytes >= 0) {
    parts.push(formatFileSize(fileSizeBytes));
  }
  return parts.length ? parts.join(" · ") : null;
}

/**
 * One `label: value` identification row. Renders the real value when present,
 * otherwise an em-dash carrying the adapter-supplied absence reason as its
 * tooltip.
 */
function MetaRow({
  label,
  value,
  absenceReason,
}: {
  label: string;
  value: string | null;
  absenceReason?: string;
}) {
  return (
    <div className={styles.metaRow}>
      <span className={styles.metaLabel}>{label}</span>
      {value ? (
        <span className={styles.metaValue}>{value}</span>
      ) : (
        <span className={styles.metaValue} title={absenceReason} data-absent="true">
          {EM_DASH}
        </span>
      )}
    </div>
  );
}

interface ArtifactMetaSidebarProps {
  open: boolean;
  onToggle: () => void;
  loading?: boolean;
  code?: string | null;
  meta: ArtifactMetaModel;
  approvalChain?: ApprovalChainItem[] | null;
  lineage?: VersionHistoryItem[];
  /** Kind-specific a11y label for the panel; caller supplies "documento"/"template". */
  ariaLabel?: string;
  /** Kind-specific loading copy; caller supplies "documento"/"template". */
  loadingLabel?: string;
  /**
   * Layout variant. Default (`false`) is the standalone editor drawer: a fixed
   * 280px rail with its own border/background and a left pull-tab that collapses
   * the whole rail. `true` is the embedded reuse (WorkspaceSidebar, spec §9.2) —
   * the panel flexes to fill its host column instead of imposing a 300px width
   * (tab + 280px), and drops the redundant rail chrome the host already draws.
   */
  embedded?: boolean;
}

export function ArtifactMetaSidebar({
  open,
  onToggle,
  loading = false,
  code,
  meta,
  approvalChain = null,
  lineage = [],
  ariaLabel = "Identificação do artefato",
  loadingLabel = "Carregando metadados",
  embedded = false,
}: ArtifactMetaSidebarProps) {
  const [historyExpanded, setHistoryExpanded] = useState(false);
  const pageSizeSummary = formatPageSizeSummary(meta.pageCount, meta.fileSizeBytes);
  const absenceReasons = meta.absenceReasons ?? {};
  const recentNonCurrentHistory = [...lineage]
    .filter((item) => !item.isCurrent)
    .sort((a, b) => new Date(b.createdAt ?? 0).getTime() - new Date(a.createdAt ?? 0).getTime());
  const visibleHistory = historyExpanded || lineage.length <= MAX_COLLAPSED_HISTORY_ITEMS
    ? lineage
    : [
        ...lineage.filter((item) => item.isCurrent),
        ...recentNonCurrentHistory.slice(0, MAX_COLLAPSED_HISTORY_ITEMS - 1),
      ].slice(0, MAX_COLLAPSED_HISTORY_ITEMS);

  const showApprovalChain = approvalChain != null && approvalChain.length > 0;

  return (
    <div className={`${styles.sidebarOuter} ${embedded ? styles.sidebarOuterEmbedded : ""}`}>
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
        <aside className={`${styles.sidebar} ${embedded ? styles.sidebarEmbedded : ""}`} aria-label={ariaLabel}>
          <div className={styles.panelFrame}>
            <section className={styles.section}>
              <div className={styles.sectionHeader}>Identificação</div>
              {loading ? (
                <div className={styles.metaRows}>
                  <div className={styles.metaRow}>
                    <span className={styles.metaValue}>{loadingLabel}</span>
                  </div>
                </div>
              ) : (
                <div className={styles.metaRows}>
                  {code ? (
                    <div className={styles.metaRow}>
                      <span className={styles.metaLabel}>Código</span>
                      <span className={styles.metaCode}>{code}</span>
                    </div>
                  ) : null}
                  <MetaRow label="Tipo" value={meta.profileLabel} absenceReason={absenceReasons.profileLabel} />
                  <MetaRow label="Área responsável" value={meta.areaLabel} absenceReason={absenceReasons.areaLabel} />
                  <MetaRow label="Visibilidade" value={meta.visibilityLabel} absenceReason={absenceReasons.visibilityLabel} />
                  {/* The pages row stays hidden for kinds that carry no file
                      metadata at all (templates); documents that legitimately
                      have no page count yet supply an absence reason instead,
                      so the row appears with an explained em-dash. */}
                  {pageSizeSummary || absenceReasons.pageCount ? (
                    <MetaRow label="Páginas" value={pageSizeSummary} absenceReason={absenceReasons.pageCount} />
                  ) : null}
                </div>
              )}
            </section>
            <div className={styles.divider} />
            <section className={styles.section}>
              <div className={styles.sectionHeader}>Revisões</div>
              <div className={styles.revisionList} aria-label="Histórico de revisões">
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
                  {historyExpanded ? "Ver menos revisões" : "Ver todas as revisões"}
                </button>
              ) : null}
            </section>
            {showApprovalChain ? (
              <>
                <div className={styles.divider} />
                <section className={styles.section}>
                  <div className={styles.sectionHeader}>Próximos aprovadores</div>
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
              <span className={styles.panelFillText}>Dossiê governado</span>
            </div>
          </div>
        </aside>
      )}
    </div>
  );
}
