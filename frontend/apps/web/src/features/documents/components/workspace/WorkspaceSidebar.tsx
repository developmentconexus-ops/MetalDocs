import { useState } from 'react';
import { Link } from 'react-router-dom';

import type { ReactNode } from 'react';
import { ArtifactMetaSidebar } from '../../../shared/controlled-artifact/ArtifactMetaSidebar';
import type { ArtifactAction, ArtifactDecisionModel, ArtifactMetaModel, VersionHistoryItem } from '../../../shared/controlled-artifact/types';
import { ApprovalTimeline } from '../../../approval/components/sidebar/ApprovalTimeline';
import { DecisionFooter } from '../../../approval/components/sidebar/DecisionFooter';
import type { ApprovalInstance, StageInstance } from '../../../approval/api/approvalTypes';
import { mapApprovalChain } from '../../lib/approvalWorkflow';
import { formatRevisionCode } from '../../../../lib/labels/revisionCode';
import { parseDocumentStatus } from '../../lib/parseDocumentStatus';
import type { DocumentDetail } from '../../api/documents';
import styles from './WorkspaceSidebar.module.css';

export interface WorkspaceSidebarProps {
  documentId: string;
  doc: DocumentDetail;
  instance: ApprovalInstance | null;
  instanceLoading: boolean;
  instanceError: string | null;
  onRetryInstance?: () => void;
  activeStage: StageInstance | undefined;
  onRefetchInstance: () => Promise<void> | void;
  /**
   * S2b — the approving-mode signature decision (built by
   * `buildDocumentSignoffDecision`, owner-page-computed). Null in every other
   * mode: DecisionFooter then falls back to its own stage-mode-derived
   * verdict/observing behavior (S2a — unchanged).
   */
  decision?: ArtifactDecisionModel | null;
  /**
   * S2b — a mode-specific contextual panel (currently: RequestedChangesPanel
   * for author-changes-requested) rendered in the scroll stack between the
   * approval timeline and the decision footer.
   */
  contextualPanel?: ReactNode;
}

// S2a: no cancel/publish dialog state lives on this screen yet (those own
// route-level dialog state on the cockpit route — out of scope here). The
// footer therefore renders with an empty lifecycle-actions group; wiring a
// real "Cancelar instância" action is deferred to S2b.
const NO_ACTIONS: ArtifactAction[] = [];

/**
 * F2d.5 S2a — the workspace's right-hand panel stack: embedded
 * ArtifactMetaSidebar (identity + revision lineage + approval chain,
 * rendered open — see spec §9.2: this RETIRES ArtifactMetaSidebar as a
 * standalone composition, reusing it here as one panel among several) →
 * ApprovalTimeline (accountability) → sticky DecisionFooter (mode-aware,
 * `decision` always null in S2a — signature/disclosure surfaces land in
 * S2b) → a discoverable link to the record view.
 */
export function WorkspaceSidebar({
  documentId,
  doc,
  instance,
  instanceLoading,
  instanceError,
  onRetryInstance,
  activeStage,
  onRefetchInstance,
  decision = null,
  contextualPanel = null,
}: WorkspaceSidebarProps) {
  const [metaOpen, setMetaOpen] = useState(true);

  const meta: ArtifactMetaModel = {
    // Reduced surface (mirrors useDocumentApprovalArtifact's cockpit model):
    // profile/area/visibility require the taxonomy queries the editor page
    // composes — out of scope for this thin owner. Honest omission (null),
    // never fabricated.
    profileLabel: null,
    areaLabel: null,
    visibilityLabel: null,
    fileSizeBytes: doc.current_revision_file_size_bytes ?? null,
    pageCount: doc.current_revision_page_count ?? null,
    createdAt: doc.created_at ?? null,
    effectiveFrom: doc.effective_from ?? null,
    nextReviewAt: doc.review_due_at ?? null,
    ownerName: doc.created_by ?? null,
    ownerDescriptor: null,
  };

  // Single-entry lineage from the loaded document itself (no revision-history
  // query in S2a — the embedded panel still gets an honest "current" row
  // instead of an empty list).
  const lineage: VersionHistoryItem[] = [
    {
      versionNumber: doc.revision_number ?? 0,
      revisionNumber: doc.revision_number ?? null,
      revisionLabel: formatRevisionCode(doc.revision_number),
      status: parseDocumentStatus(doc.status),
      title: doc.name ?? null,
      createdAt: doc.created_at ?? null,
      isCurrent: true,
    },
  ];

  const approvalChain = instance ? mapApprovalChain(instance.stages) : null;

  return (
    <aside className={styles.sidebar} aria-label="Painel do documento" data-testid="workspace-sidebar">
      <div className={styles.scroll}>
        <ArtifactMetaSidebar
          open={metaOpen}
          onToggle={() => setMetaOpen((open) => !open)}
          code={doc.code ?? null}
          meta={meta}
          approvalChain={approvalChain}
          lineage={lineage}
          ariaLabel="Identificação do documento"
          loadingLabel="Carregando metadados do documento"
        />
        <ApprovalTimeline
          instance={instance}
          loading={instanceLoading}
          error={instanceError}
          onRetry={onRetryInstance}
        />
        {contextualPanel}
      </div>

      {instance ? (
        <DecisionFooter
          decision={decision}
          actions={NO_ACTIONS}
          instance={instance}
          activeStage={activeStage}
          onRefetchInstance={onRefetchInstance}
        />
      ) : null}

      <Link to={`/documents/${documentId}/details`} className={styles.detailsLink}>
        Ver ficha completa do documento
      </Link>
    </aside>
  );
}
