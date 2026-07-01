import type React from "react";
import { Fragment } from "react";
import { Link } from "react-router-dom";
import { Avatar } from "../../../components/ui/Avatar";
import { Icon } from "../../../components/ui/Icon";
import { StatusPill } from "../../../components/ui/StatusPill";
import { ArtifactApprovalFlow } from "./ArtifactApprovalFlow";
import { ArtifactDecisionPanel } from "./ArtifactDecisionPanel";
import type { ArtifactAction, ArtifactAuthorMeta, ArtifactViewModel } from "./types";
import styles from "./ArtifactApprovalScreen.module.css";

interface ArtifactApprovalScreenProps {
  model: ArtifactViewModel;
  /** Kind-specific main body: the A4 document / review canvas. Scrolls under the header. */
  main: React.ReactNode;
  /** Optional route-owned tab strip rendered between the doc header and the body. */
  tabs?: React.ReactNode;
  /** Kind-specific sidebar extras rendered BELOW the decision panel / actions (e.g.
   *  document integrity panel / lock badge / timeline). Optional. */
  decisionExtras?: React.ReactNode;
  /** Route-owned modal dialogs (SupersedePublishDialog, CancelInstanceDialog, …). Optional. */
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

/** Author / submission descriptor row — renders only the segments that are present. */
function AuthorMetaRow({ authorMeta }: { authorMeta: ArtifactAuthorMeta }) {
  const segments: { key: string; label: string }[] = [];
  if (authorMeta.authorRole != null) segments.push({ key: 'role', label: authorMeta.authorRole });
  if (authorMeta.areaLabel != null) segments.push({ key: 'area', label: authorMeta.areaLabel });
  if (authorMeta.submittedLabel != null) segments.push({ key: 'submitted', label: authorMeta.submittedLabel });
  if (authorMeta.priorDecisionLabel != null)
    segments.push({ key: 'prior', label: authorMeta.priorDecisionLabel });

  const hasAny = authorMeta.authorName != null || segments.length > 0;
  if (!hasAny) return null;

  return (
    <div className={styles.authorRow}>
      {authorMeta.authorName != null && (
        <span className={styles.authorLead}>
          <Avatar name={authorMeta.authorName} size="sm" />
          {authorMeta.authorName}
        </span>
      )}
      {segments.map((seg, i) => (
        <Fragment key={seg.key}>
          {(i > 0 || authorMeta.authorName != null) && <span className={styles.authorSep}>·</span>}
          <span>{seg.label}</span>
        </Fragment>
      ))}
    </div>
  );
}

/**
 * Approval cockpit for a controlled artifact. Purely presentational: consumes an
 * `ArtifactViewModel` (+ optional ReactNode slots) with no data fetching, no
 * `useParams`, and no `kind` branching. Mirrors the detalhe-signoff reference —
 * a focused review header + body on the left, and a decision sidebar (approval
 * flow-viz + decision panel) on the right.
 *
 * The sidebar is lifecycle-state-driven: when the adapter supplies a `decision`
 * model (document under_review sign, template review/approve) the cockpit renders
 * the §DECISÃO header + flow + `ArtifactDecisionPanel`. Otherwise it falls back to
 * the plain `actions[]` band + the route's `decisionExtras` slot (draft submit,
 * approved publish, read-only states).
 */
export function ArtifactApprovalScreen({
  model,
  main,
  tabs,
  decisionExtras,
  dialogs,
}: ArtifactApprovalScreenProps) {
  const hasActions = model.actions.length > 0;
  const chain = model.approvalChain;
  const hasChain = chain != null && chain.length > 0;
  const decision = model.decision;

  const actionsBand = hasActions ? (
    <div className={styles.asideSection}>
      <h2 className={styles.bandKicker}>{decision != null ? "Outras ações" : "Ações"}</h2>
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
  ) : null;

  return (
    <div className={styles.root}>
      <div className={styles.body}>
        {/* ─── Main: focused doc header + body ─────────────────────────────── */}
        <main className={styles.main}>
          <header className={styles.docHeader}>
            {model.hero.breadcrumb.length > 0 && (
              <nav aria-label="Breadcrumb" className={styles.breadcrumb}>
                {model.hero.breadcrumb.map((item, index) => {
                  const isLast = index === model.hero.breadcrumb.length - 1;
                  return (
                    <Fragment key={`${item.label}-${index}`}>
                      {item.href && !isLast ? (
                        <Link to={item.href} className={styles.breadcrumbLink}>
                          {item.label}
                        </Link>
                      ) : (
                        <span>{item.label}</span>
                      )}
                      {!isLast && <Icon name="chevron" size={10} className={styles.breadcrumbSep} />}
                    </Fragment>
                  );
                })}
              </nav>
            )}

            <div className={styles.titleRow}>
              {model.code != null && <span className={styles.codeChip}>{model.code}</span>}
              <h1 className={styles.title}>{model.title}</h1>
            </div>

            <div className={styles.statusRow}>
              <StatusPill status={model.status} />
              <span className={styles.versionTag}>v{model.versionNumber}</span>
              {model.revisionLabel != null && (
                <>
                  <span className={styles.statusSep}>·</span>
                  <span className={styles.versionTag}>{model.revisionLabel}</span>
                </>
              )}
            </div>

            {model.authorMeta != null && <AuthorMetaRow authorMeta={model.authorMeta} />}
          </header>

          {tabs != null && <div className={styles.tabStrip}>{tabs}</div>}

          <div className={styles.bodyContent}>{main}</div>
        </main>

        {/* ─── Sidebar: decision cockpit ───────────────────────────────────── */}
        <aside aria-label="Decisão de aprovação" className={styles.aside}>
          {decision != null ? (
            <>
              <div className={styles.decisionHead}>
                <div className={styles.decisionKicker}>{decision.kicker}</div>
                <h2 className={styles.decisionHeading}>{decision.heading}</h2>
                <p className={styles.decisionDesc}>{decision.description}</p>
              </div>

              {hasChain && (
                <div className={styles.flowBand}>
                  <div className={styles.bandKicker}>Fluxo de aprovação</div>
                  <ArtifactApprovalFlow items={chain} />
                </div>
              )}

              <ArtifactDecisionPanel model={decision} />

              {(actionsBand != null || decisionExtras != null) && (
                <div className={styles.fallback}>
                  {actionsBand}
                  {decisionExtras}
                </div>
              )}
            </>
          ) : (
            <div className={styles.fallback}>
              {actionsBand}

              {hasChain && (
                <div className={styles.asideSection}>
                  <h2 className={styles.bandKicker}>Fluxo de aprovação</h2>
                  <ArtifactApprovalFlow items={chain} />
                </div>
              )}

              {decisionExtras}
            </div>
          )}
        </aside>
      </div>

      {dialogs}
    </div>
  );
}
