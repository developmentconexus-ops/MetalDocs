import type { VersionDTO } from '../api/templates';
import type { ActorContext } from './canActOnVersion';
import { canPublish, canSignoff } from './canActOnVersion';
import type { ArtifactAction } from '../../shared/controlled-artifact/types';

/**
 * Route-supplied runners bound to each emitted action's `run`.
 *
 * `runSignoff(accept)` is wired to the under_review accept/reject pair. In
 * practice it is never reached by a click: whenever canSignoff allows the
 * pair, the adapter also offers `model.decision` for the same version+actor,
 * and the cockpit strips `actions[]` to `[]` while a decision is offered — the
 * real signoff (password + motivo) is owned by the DecisionPanel, not this
 * fallback. It still has to exist so the pair has somewhere to point while
 * disabled (gate denied, no decision offered either).
 *
 * `runPublish()` is wired to the approved-state Publicar action, which has no
 * signature ceremony and is always reached through this plain-action path.
 */
export interface TemplateApprovalHandlers {
  runSignoff: (accept: boolean) => void;
  runPublish: () => void;
}

/**
 * Ordered, gated approval actions for a template version. Covers:
 *   - draft                 → no cockpit actions (submit lives solely in the template editor — R1)
 *   - under_review          → Assinar aprovação / Rejeitar (canSignoff → runSignoff)
 *   - approved              → Publicar only (canPublish → runPublish) — no kernel
 *                             equivalent for "reject at approved" (no approved→draft endpoint)
 *   - published / obsolete  → no actions
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
      const gate = canSignoff(version, actor);
      actions.push({
        key: 'accept',
        label: 'Assinar aprovação',
        variant: 'primary',
        available: gate.allowed,
        reason: gate.allowed ? undefined : gate.reason,
        run: () => handlers.runSignoff(true),
      });
      actions.push({
        key: 'reject',
        label: 'Rejeitar',
        variant: 'danger',
        available: gate.allowed,
        reason: gate.allowed ? undefined : gate.reason,
        run: () => handlers.runSignoff(false),
      });
      break;
    }
    case 'approved': {
      const gate = canPublish(version, actor);
      actions.push({
        key: 'publish',
        label: 'Publicar',
        variant: 'primary',
        available: gate.allowed,
        reason: gate.allowed ? undefined : gate.reason,
        run: () => handlers.runPublish(),
      });
      break;
    }
    default:
      // draft / published / obsolete: read-only, no actions.
      break;
  }
  return actions;
}
