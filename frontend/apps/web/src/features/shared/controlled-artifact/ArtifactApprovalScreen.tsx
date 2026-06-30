import type React from "react";
import { Avatar } from "../../../components/ui/Avatar";
import { CodeChip } from "../../../components/ui/CodeChip";
import { formatRevisionCode } from "../../../lib/labels/revisionCode";
import { formatSignedAt } from "../../../lib/format/dates";
import { ArtifactHero } from "./ArtifactHero";
import type { ArtifactAction, ArtifactViewModel } from "./types";
import styles from "./ArtifactApprovalScreen.module.css";
import detailStyles from "./ArtifactDetailView.module.css";

const EM_DASH = "—";

interface ArtifactApprovalScreenProps {
  model: ArtifactViewModel;
  /** Kind-specific main content (document review canvas + tabs, or template preview). */
  main: React.ReactNode;
  /** Kind-specific decision-sidebar extras rendered BELOW the action buttons and
   *  approval chain (e.g. document integrity panel / lock badge / timeline, or a
   *  template reason composer). Optional. */
  decisionExtras?: React.ReactNode;
  /** Route-owned modal dialogs (SignoffDialog, SupersedePublishDialog, …). Optional. */
  dialogs?: React.ReactNode;
}

function variantClass(action: ArtifactAction): string {
  switch (action.variant) {
    case "primary":
      return styles.actionBtnPrimary;
    case "danger":
      return styles.actionBtnDanger;
    default:
      return styles.actionBtnSecondary;
  }
}

/**
 * Purely presentational approval surface for a controlled artifact. Consumes an
 * `ArtifactViewModel` and optional ReactNode slots — no data fetching, no
 * `useParams`, no `kind` branching. Every kind-specific difference is either
 * normalized into the model by the adapter or injected via a slot by the caller.
 */
export function ArtifactApprovalScreen({
  model,
  main,
  decisionExtras,
  dialogs,
}: ArtifactApprovalScreenProps) {
  const code = model.code ?? EM_DASH;
  const versionLabel = model.revisionLabel ?? formatRevisionCode(model.versionNumber);
  const profileLabel = model.meta.profileLabel ?? EM_DASH;
  const areaLabel = model.meta.areaLabel ?? EM_DASH;
  const subtitle = model.hero.subtitle;

  const hasActions = model.actions.length > 0;
  const hasChain = model.approvalChain != null && model.approvalChain.length > 0;

  return (
    <div className={styles.root}>
      {/* Hero — reuses ArtifactHero for breadcrumb + badges + title, consistent
          with ArtifactDetailView. No heroActions slot: approval-screen workflow
          buttons live exclusively in the decision sidebar. */}
      <ArtifactHero
        breadcrumb={model.hero.breadcrumb}
        docCard={
          <div className={detailStyles.docCard}>
            <div className={detailStyles.docCardHeader}>{areaLabel}</div>
            <div className={detailStyles.docCardBody}>
              <div className={detailStyles.docCardCode}>{code}</div>
              <div className={detailStyles.docCardType}>{profileLabel}</div>
              <div className={detailStyles.docCardSpacer} />
              <div className={detailStyles.docCardDivider} />
              <div className={detailStyles.docCardFooter}>
                <span className={detailStyles.docCardVersion}>{versionLabel}</span>
                <span className={detailStyles.docCardDot} />
              </div>
            </div>
          </div>
        }
        badges={
          <>
            {model.hero.badges.map((badge) => {
              if (badge.variant === "code") {
                return (
                  <CodeChip key={badge.key} className={detailStyles.codeChip}>
                    {badge.label}
                  </CodeChip>
                );
              }
              if (badge.variant === "status") {
                return (
                  <span key={badge.key} className={detailStyles.vigenteBadge}>
                    <span className={detailStyles.vigenteDot} />
                    {badge.label}
                  </span>
                );
              }
              return (
                <span key={badge.key} className={detailStyles.typeLabel}>
                  {badge.label}
                </span>
              );
            })}
          </>
        }
        title={model.title}
        subtitle={subtitle ? <span>{subtitle}</span> : null}
      />

      {/* Two-column body */}
      <div className={styles.body}>
        {/* Main content slot */}
        <main className={styles.main}>{main}</main>

        {/* Decision sidebar */}
        <aside aria-label="Decisão de aprovação" className={styles.aside}>

          {/* Ações — only rendered when there are actions */}
          {hasActions && (
            <div className={styles.asideSection}>
              <h2 className={styles.asideSectionTitle}>Ações</h2>
              <div className={styles.actions}>
                {model.actions.map((action) => (
                  <button
                    key={action.key}
                    type="button"
                    data-action={action.key}
                    className={`${styles.actionBtn} ${variantClass(action)}`}
                    disabled={!action.available}
                    title={action.available ? undefined : action.reason}
                    onClick={() => {
                      void action.run();
                    }}
                  >
                    {action.label}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Approval chain — null for templates, skipped when empty */}
          {hasChain && (
            <div className={styles.asideSection}>
              <h2 className={styles.asideSectionTitle}>Cadeia de aprovação</h2>
              <ul className={styles.chainList} aria-label="Cadeia de aprovação">
                {model.approvalChain!.map((item) => (
                  <li key={`${item.stageIndex}-${item.actorUserId ?? "unassigned"}`} className={styles.chainItem}>
                    <div className={styles.chainItemLabel}>{item.label}</div>
                    {item.actorDisplay != null && (
                      <div className={styles.chainItemActor}>
                        <Avatar name={item.actorDisplay} size="sm" />
                        {" "}{item.actorDisplay}
                      </div>
                    )}
                    {item.signedAt != null && (
                      <div className={styles.chainItemMeta}>
                        {formatSignedAt(item.signedAt)}
                      </div>
                    )}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Kind-specific extras (integrity panel, reason composer, etc.) */}
          {decisionExtras}

        </aside>
      </div>

      {/* Route-owned dialogs */}
      {dialogs}
    </div>
  );
}
