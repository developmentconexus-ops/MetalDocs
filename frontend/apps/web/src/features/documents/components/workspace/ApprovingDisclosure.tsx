import { IntegrityDisclosure } from '../../../approval/components/sidebar/IntegrityDisclosure';
import styles from './ApprovingDisclosure.module.css';

export interface ApprovingDisclosureProps {
  documentId: string;
  contentHash: string | null;
  frozenContentHash: string | null;
  /**
   * Server-derived display name of the actor the current signer is standing
   * in for (`viewer.via_delegation_from.display_name`, ADR 0078 — never
   * client-derived). Null when signing under the caller's own eligibility.
   */
  delegatedFrom: string | null;
}

/**
 * ApprovingDisclosure — F2d.5 S2b.
 *
 * The approving-canvas disclosure band, rendered atop the frozen read canvas:
 * the reused auditor `IntegrityDisclosure` (hash/ETag, collapsed by default,
 * spec §1.3) plus a delegation badge when the workspace mode resolved
 * `approving` via a delegated eligibility (`viewer.via_delegation_from`) —
 * the same server-derived fact `ModeChip`'s tooltip already surfaces, made
 * visible here as a persistent badge instead of hover-only copy.
 */
export function ApprovingDisclosure({ documentId, contentHash, frozenContentHash, delegatedFrom }: ApprovingDisclosureProps) {
  return (
    <div className={styles.root}>
      {delegatedFrom ? (
        <span className={styles.delegationBadge} data-testid="delegation-badge">
          Assinando por delegação de {delegatedFrom}
        </span>
      ) : null}
      <IntegrityDisclosure
        documentId={documentId}
        contentHash={contentHash}
        frozenContentHash={frozenContentHash}
      />
    </div>
  );
}
