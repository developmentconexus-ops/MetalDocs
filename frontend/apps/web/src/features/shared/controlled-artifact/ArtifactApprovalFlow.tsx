import { Icon } from "../../../components/ui/Icon";
import { formatSignedAt } from "../../../lib/format/dates";
import type { ApprovalChainItem, ApprovalFlowState } from "./types";
import styles from "./ArtifactApprovalFlow.module.css";

interface ArtifactApprovalFlowProps {
  /** Ordered approval-chain steps (submit → review → approve for both kinds). */
  items: ApprovalChainItem[];
}

/** Dot glyph for a step, driven purely by its normalized flow state. */
function StepDot({ flowState }: { flowState: ApprovalFlowState }) {
  const className = `${styles.dot} ${styles[`dot_${flowState}`]}`;
  return (
    <span className={className} aria-hidden="true">
      {flowState === "approved" && <Icon name="check" size={11} />}
      {flowState === "rejected" && <Icon name="x" size={11} />}
      {flowState === "sent" && <span className={styles.dotInner} />}
    </span>
  );
}

/**
 * Presentational vertical approval-flow timeline for the cockpit sidebar.
 *
 * Renders the SAME `ApprovalChainItem[]` produced for both documents (from the
 * approval instance) and templates (from the version's inline submit/review/approve
 * fields) — so the flow-viz looks identical regardless of kind. No data fetching,
 * no kind branching: the adapter has already normalized every step's `flowState`.
 *
 * Honest-omit: when a step has no signer/timestamp yet (a pending future step),
 * the actor line falls back to the role and the time renders an em-dash — never a
 * fabricated name or date.
 */
export function ArtifactApprovalFlow({ items }: ArtifactApprovalFlowProps) {
  return (
    <ol className={styles.flow} aria-label="Fluxo de aprovação">
      {items.map((item, index) => {
        const isLast = index === items.length - 1;
        const isCurrent = item.flowState === "current";
        const who = item.actorDisplay ?? item.roleLabel ?? "Pendente";
        const time = item.signedAt != null ? formatSignedAt(item.signedAt) : "—";
        return (
          <li
            key={`${item.stageIndex}-${item.actorUserId ?? "unassigned"}`}
            className={styles.step}
            data-flow-state={item.flowState}
          >
            {!isLast && <span className={styles.connector} aria-hidden="true" />}
            <StepDot flowState={item.flowState} />
            <div className={styles.stepBody}>
              <div className={`${styles.who} ${isCurrent ? styles.whoCurrent : ""}`}>{who}</div>
              <div className={styles.meta}>
                {item.roleLabel != null && <span>{item.roleLabel} · </span>}
                <span className={styles.time}>{time}</span>
              </div>
            </div>
          </li>
        );
      })}
    </ol>
  );
}
