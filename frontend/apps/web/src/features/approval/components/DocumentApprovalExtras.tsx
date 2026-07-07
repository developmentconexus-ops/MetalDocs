import { useState } from 'react';

import { etagCache } from '../api/etagCache';
import type { ApprovalInstance, ApprovalState } from '../api/approvalTypes';
import type { TransitionPolicy } from '../../documents/lib/approvalWorkflow';
import { ApprovalTimelinePanel } from './ApprovalTimelinePanel';
import { LockBadge } from './LockBadge';
import { StateBadge } from './StateBadge';
import styles from './DocumentApprovalExtras.module.css';

interface DocumentApprovalExtrasProps {
  documentId: string;
  /** Coerced approval state — drives the StateBadge and disabledReason/readOnly copy. */
  status: ApprovalState;
  /** The active TRANSITION_POLICY entry (disabledReason / readOnly). */
  policy: TransitionPolicy;
  contentHash: string;
  revisionVersion: number;
  lockedByInstanceId?: string;
  /** Raw instance for the embedded timeline (null while loading / when absent). */
  instance: ApprovalInstance | null;
  /** True when the last instance fetch is older than 30s. */
  isStale: boolean;
  /** Imperative instance refetch (Atualizar). */
  onRefetchInstance: () => Promise<void> | void;
}

/**
 * Presentational decision-sidebar extras for the document approval screen,
 * rendered into ArtifactApprovalScreen's `decisionExtras` slot. Holds the
 * LockBadge + StateBadge row, the integrity block (content_hash + copy,
 * revision_version, ETag), the staleness banner, the embedded ApprovalTimelinePanel,
 * and the disabledReason / read-only tags.
 *
 * All approval workflow ACTION buttons live in the shared sidebar (model.actions);
 * this component renders only the supporting decision context lifted verbatim from
 * the former ControlledDocumentDetailPanel.
 */
export function DocumentApprovalExtras({
  documentId,
  status,
  policy,
  contentHash,
  revisionVersion,
  lockedByInstanceId,
  instance,
  isStale,
  onRefetchInstance,
}: DocumentApprovalExtrasProps) {
  const [copied, setCopied] = useState(false);

  const etag = etagCache.get(documentId) ?? '-';
  const shortHash = contentHash.length > 8 ? `${contentHash.slice(0, 8)}…` : contentHash;

  const handleCopyHash = async () => {
    try {
      await navigator.clipboard.writeText(contentHash);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1200);
    } catch (_error) {
      setCopied(false);
    }
  };

  const scrollToTimeline = () => {
    document.getElementById('approval-timeline')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  };

  return (
    <section className={styles.panel} aria-label="Painel de detalhes de aprovação">
      <div className={styles.topRow}>
        <LockBadge lockedByInstanceId={lockedByInstanceId} onBannerClick={scrollToTimeline} />
        <StateBadge state={status} />
      </div>

      <div className={styles.section}>
        <h3>Integridade</h3>
        <div className={styles.integrityGrid}>
          <div>
            <span className={styles.label}>content_hash</span>
            <div className={styles.hashRow}>
              <code title={contentHash}>{shortHash}</code>
              <button type="button" className={styles.smallButton} onClick={() => void handleCopyHash()}>
                {copied ? 'Copiado' : 'Copiar'}
              </button>
            </div>
          </div>
          <div>
            <span className={styles.label}>revision_version</span>
            <code>{revisionVersion}</code>
          </div>
          <div>
            <span className={styles.label}>ETag</span>
            <code>{etag}</code>
          </div>
        </div>
      </div>

      {isStale ? (
        <div className={styles.staleBanner} role="status">
          <span>Dados podem estar desatualizados.</span>
          <button type="button" className={styles.smallButton} onClick={() => void onRefetchInstance()}>
            Atualizar
          </button>
        </div>
      ) : null}

      {(policy.disabledReason || policy.readOnly) ? (
        <div className={styles.section}>
          {policy.disabledReason ? <p className={styles.disabledReason}>{policy.disabledReason}</p> : null}
          {policy.readOnly ? <p className={styles.readOnlyTag}>Somente leitura</p> : null}
        </div>
      ) : null}

      {instance ? (
        <section id="approval-timeline" className={styles.section}>
          <ApprovalTimelinePanel instance={instance} loading={false} />
        </section>
      ) : null}
    </section>
  );
}
