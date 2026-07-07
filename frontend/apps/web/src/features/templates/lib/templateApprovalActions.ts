import type { VersionDTO } from '../api/templates';
import type { ActorContext } from './canActOnVersion';
import { canApprove, canPublish, canReview } from './canActOnVersion';
import type { ArtifactAction } from '../../shared/controlled-artifact/types';

/**
 * Route-supplied runners bound to each emitted action's `run`. The route owns the
 * reason textarea state + the API calls (reviewVersion / approveVersion) +
 * idempotency keys + refetch; this module only decides WHICH actions are allowed
 * in the current version state and wires their `run` to these.
 *
 * `runReview(accept)` → reviewer flow (reviewVersion). `runApprove(accept)` →
 * approver/publish flow (approveVersion).
 */
export interface TemplateApprovalHandlers {
  runReview: (accept: boolean) => void;
  runApprove: (accept: boolean) => void;
}

/**
 * Ordered, gated approval actions for a template version. Covers:
 *   - draft            → no cockpit actions (submit lives solely in the template editor — R1)
 *   - under_review + reviewer → Aprovar revisão / Rejeitar (canReview → runReview)
 *   - under_review, no reviewer → Publicar / Rejeitar     (canPublish → runApprove)
 *   - approved         → Publicar / Rejeitar              (canApprove → runApprove)
 *   - published / obsolete → no actions
 * The gate's `reason` becomes the disabled-button tooltip (available=false ⇒ rendered disabled).
 */
export function buildTemplateApprovalActions(
  version: VersionDTO,
  actor: ActorContext,
  handlers: TemplateApprovalHandlers,
): ArtifactAction[] {
  const actions: ArtifactAction[] = [];
  switch (version.status) {
    case 'under_review': {
      const hasReviewer = version.pending_reviewer_role != null;
      const gate = hasReviewer ? canReview(version, actor) : canPublish(version, actor);
      const acceptLabel = hasReviewer ? 'Aprovar revisão' : 'Publicar';
      const run = hasReviewer ? handlers.runReview : handlers.runApprove;
      actions.push({
        key: 'accept',
        label: acceptLabel,
        variant: 'primary',
        available: gate.allowed,
        reason: gate.allowed ? undefined : gate.reason,
        run: () => run(true),
      });
      actions.push({
        key: 'reject',
        label: 'Rejeitar',
        variant: 'danger',
        available: gate.allowed,
        reason: gate.allowed ? undefined : gate.reason,
        run: () => run(false),
      });
      break;
    }
    case 'approved': {
      const gate = canApprove(version, actor);
      actions.push({
        key: 'accept',
        label: 'Publicar',
        variant: 'primary',
        available: gate.allowed,
        reason: gate.allowed ? undefined : gate.reason,
        run: () => handlers.runApprove(true),
      });
      actions.push({
        key: 'reject',
        label: 'Rejeitar',
        variant: 'danger',
        available: gate.allowed,
        reason: gate.allowed ? undefined : gate.reason,
        run: () => handlers.runApprove(false),
      });
      break;
    }
    default:
      // published / obsolete: read-only, no actions.
      break;
  }
  return actions;
}
