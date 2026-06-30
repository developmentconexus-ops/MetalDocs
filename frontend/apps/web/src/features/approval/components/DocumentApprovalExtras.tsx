import { useEffect, useState } from 'react';

import { submit } from '../api/approvalApi';
import { etagCache } from '../api/etagCache';
import { listRoutes, type RouteSummary } from '../api/routeAdminApi';
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
  /** Imperative instance refetch (Atualizar + post-submit). */
  onRefetchInstance: () => Promise<void> | void;
  /** Whether the inline submit route-picker is open (driven by the Submeter action). */
  showSubmit: boolean;
  /** Close the inline submit route-picker. */
  onCloseSubmit: () => void;
}

/**
 * Presentational decision-sidebar extras for the document approval screen,
 * rendered into ArtifactApprovalScreen's `decisionExtras` slot. Holds the
 * LockBadge + StateBadge row, the integrity block (content_hash + copy,
 * revision_version, ETag), the staleness banner, the embedded ApprovalTimelinePanel,
 * the disabledReason / read-only tags, and the inline submit route-picker.
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
  showSubmit,
  onCloseSubmit,
}: DocumentApprovalExtrasProps) {
  const [copied, setCopied] = useState(false);

  const [routes, setRoutes] = useState<RouteSummary[]>([]);
  const [routesLoading, setRoutesLoading] = useState(false);
  const [routesError, setRoutesError] = useState<string | null>(null);
  const [selectedRouteId, setSelectedRouteId] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const etag = etagCache.get(documentId) ?? '-';
  const shortHash = contentHash.length > 8 ? `${contentHash.slice(0, 8)}…` : contentHash;

  // Load active routes when the submit picker opens (parity with openSubmitSection).
  useEffect(() => {
    if (!showSubmit) {
      return;
    }
    let cancelled = false;
    setRoutesLoading(true);
    setRoutesError(null);
    setSubmitError(null);
    void (async () => {
      try {
        const response = await listRoutes();
        const activeRoutes = response.routes.filter((route) => route.active);
        if (cancelled) {
          return;
        }
        setRoutes(activeRoutes);
        setSelectedRouteId((prev) => prev || activeRoutes[0]?.id || '');
      } catch (_error) {
        if (!cancelled) {
          setRoutesError('Erro ao carregar rotas.');
        }
      } finally {
        if (!cancelled) {
          setRoutesLoading(false);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [showSubmit]);

  const handleCopyHash = async () => {
    try {
      await navigator.clipboard.writeText(contentHash);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1200);
    } catch (_error) {
      setCopied(false);
    }
  };

  const handleSubmitForReview = async () => {
    if (!selectedRouteId || submitting) {
      return;
    }
    setSubmitting(true);
    setSubmitError(null);
    try {
      await submit(documentId, {
        route_id: selectedRouteId,
        content_hash: contentHash,
      });
      onCloseSubmit();
      await onRefetchInstance();
    } catch (_error) {
      setSubmitError('Erro ao submeter para revisão.');
    } finally {
      setSubmitting(false);
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

      {showSubmit ? (
        <div className={styles.submitSection}>
          <h4>Submeter para revisão</h4>
          {routesLoading ? <p>Carregando rotas...</p> : null}
          {routesError ? <p role="alert">{routesError}</p> : null}
          {!routesLoading && !routesError ? (
            <>
              <label htmlFor="route-select" className={styles.label}>
                Rota
              </label>
              <select
                id="route-select"
                className={styles.select}
                value={selectedRouteId}
                onChange={(event) => setSelectedRouteId(event.target.value)}
              >
                {routes.map((route) => (
                  <option key={route.id} value={route.id}>
                    {route.name}
                  </option>
                ))}
              </select>
              {routes.length === 0 ? <p>Nenhuma rota configurada.</p> : null}
              {submitError ? <p role="alert">{submitError}</p> : null}
              <div className={styles.inlineActions}>
                <button
                  type="button"
                  className={styles.actionButton}
                  onClick={() => void handleSubmitForReview()}
                  disabled={!selectedRouteId || submitting}
                >
                  {submitting ? 'Enviando...' : 'Submeter'}
                </button>
                <button type="button" className={styles.actionButtonSecondary} onClick={onCloseSubmit}>
                  Cancelar
                </button>
              </div>
            </>
          ) : null}
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
