import type { TrackedChange } from '@metaldocs/editor-ui';

import type { ArtifactAction, ArtifactDecisionModel } from '../../../shared/controlled-artifact/types';
import type { ApprovalInstance, StageInstance } from '../../api/approvalTypes';
import { ApprovalTimeline } from './ApprovalTimeline';
import { DecisionFooter } from './DecisionFooter';
import { IntegrityDisclosure } from './IntegrityDisclosure';
import { StageContextHeader } from './StageContextHeader';
import { SuggestionList } from './SuggestionList';
import styles from './ApprovalSidebar.module.css';

export interface ApprovalSidebarProps {
  /** 'readonly' here == approval/observer (from cockpit). */
  editorMode: 'review' | 'readonly';
  /** Active stage_kind === 'approval' — decides the footer variant. */
  isApprovalStage: boolean;
  instance: ApprovalInstance;
  activeStage: StageInstance | undefined;
  documentId: string;
  contentHash: string | null;
  frozenContentHash: string | null;
  /** model.decision for approval-mode signoff. */
  decision: ArtifactDecisionModel | null;
  /** Publish/cancel etc. */
  actions: ArtifactAction[];
  /** Review-mode suggestion cards. */
  trackedChanges: TrackedChange[];
  lockedByInstanceId?: string;
  onRefetchInstance: () => Promise<void> | void;
}

/**
 * Self-contained document-approval aside (spec §1) — the composition root for
 * the cockpit sidebar. Flex column: an internally scrolling region
 * (StageContextHeader → ApprovalTimeline → IntegrityDisclosure →
 * SuggestionList in review mode) with the mode-aware DecisionFooter pinned at
 * the bottom via `position: sticky`. Rendered as `ArtifactApprovalScreen`'s
 * `decisionExtras` — the cockpit neutralizes the shared screen's own
 * decision/flow bands via null model fields (see ApprovalCockpitPage).
 */
export function ApprovalSidebar({
  editorMode,
  isApprovalStage,
  instance,
  activeStage,
  documentId,
  contentHash,
  frozenContentHash,
  decision,
  actions,
  trackedChanges,
  lockedByInstanceId,
  onRefetchInstance,
}: ApprovalSidebarProps) {
  // Suggestions render in review mode when signoff isn't currently offered
  // (decision == null) and the active stage isn't an approval stage — belt and
  // suspenders: `decision` is the adapter's actual signoff-offer signal (its
  // gate is document-status-based, not stage_kind-based), `isApprovalStage`
  // guards against the edge case of a decision surviving into an approval-stage
  // render before the instance refetch settles.
  const showSuggestions = editorMode === 'review' && decision == null && !isApprovalStage;

  return (
    <div className={styles.root}>
      <div
        className={styles.scroll}
        data-testid="approval-sidebar-scroll"
        style={{ overflowY: 'auto' }}
      >
        <StageContextHeader
          instance={instance}
          activeStage={activeStage}
          lockedByInstanceId={lockedByInstanceId}
        />

        <ApprovalTimeline instance={instance} loading={false} />

        <IntegrityDisclosure
          documentId={documentId}
          contentHash={contentHash}
          frozenContentHash={frozenContentHash}
        />

        {showSuggestions && <SuggestionList documentId={documentId} trackedChanges={trackedChanges} />}
      </div>

      <DecisionFooter
        decision={decision}
        actions={actions}
        instance={instance}
        activeStage={activeStage}
        onRefetchInstance={onRefetchInstance}
      />
    </div>
  );
}
