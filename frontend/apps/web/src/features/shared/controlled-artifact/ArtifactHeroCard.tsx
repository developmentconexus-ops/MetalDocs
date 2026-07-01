import { CodeChip } from "../../../components/ui/CodeChip";
import { formatRevisionCode } from "../../../lib/labels/revisionCode";
import type { ArtifactViewModel } from "./types";
import styles from "./ArtifactHeroCard.module.css";

const EM_DASH = "—";

/**
 * Presentational hero doc-card for a controlled artifact — the small "physical
 * document" card in the hero's left column. Shared verbatim by ArtifactDetailView
 * and ArtifactApprovalScreen so the two surfaces stay in pixel-parity and neither
 * reaches across into the other's CSS module.
 */
export function ArtifactHeroDocCard({ model }: { model: ArtifactViewModel }) {
  const code = model.code ?? EM_DASH;
  const versionLabel = model.revisionLabel ?? formatRevisionCode(model.versionNumber);
  const profileLabel = model.meta.profileLabel ?? EM_DASH;
  const areaLabel = model.meta.areaLabel ?? EM_DASH;

  return (
    <div className={styles.docCard}>
      <div className={styles.docCardHeader}>{areaLabel}</div>
      <div className={styles.docCardBody}>
        <div className={styles.docCardCode}>{code}</div>
        <div className={styles.docCardType}>{profileLabel}</div>
        <div className={styles.docCardSpacer} />
        <div className={styles.docCardDivider} />
        <div className={styles.docCardFooter}>
          <span className={styles.docCardVersion}>{versionLabel}</span>
          <span className={styles.docCardDot} />
        </div>
      </div>
    </div>
  );
}

/**
 * Presentational hero badge row for a controlled artifact — model-driven badges
 * (code chip / vigente status pill / plain type label). Shared by
 * ArtifactDetailView and ArtifactApprovalScreen.
 */
export function ArtifactHeroBadges({ badges }: { badges: ArtifactViewModel["hero"]["badges"] }) {
  return (
    <>
      {badges.map((badge) => {
        if (badge.variant === "code") {
          return (
            <CodeChip key={badge.key} className={styles.codeChip}>
              {badge.label}
            </CodeChip>
          );
        }
        if (badge.variant === "status") {
          return (
            <span key={badge.key} className={styles.vigenteBadge}>
              <span className={styles.vigenteDot} />
              {badge.label}
            </span>
          );
        }
        return (
          <span key={badge.key} className={styles.typeLabel}>
            {badge.label}
          </span>
        );
      })}
    </>
  );
}
