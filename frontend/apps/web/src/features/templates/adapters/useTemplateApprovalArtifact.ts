import type { ArtifactViewModel } from '../../shared/controlled-artifact/types';
import type { VersionDTO } from '../api/templates';
import type { ActorContext } from '../lib/canActOnVersion';
import { buildTemplateApprovalActions, type TemplateApprovalHandlers } from '../lib/templateApprovalActions';
import { buildTemplateApprovalChain } from '../lib/templateApprovalChain';
import {
  buildTemplateApprovalDecision,
  type TemplateApprovalDecisionSubmit,
} from '../lib/templateApprovalDecision';
import { useTemplateDetailQuery } from '../queries/useTemplateDetailQuery';
import { useAuthStore } from '../../../store/auth.store';
import { useTemplateArtifact } from './useTemplateArtifact';

export type { TemplateApprovalHandlers };

export interface TemplateApprovalArtifact {
  model: ArtifactViewModel;
  version: VersionDTO | null;
  isLoading: boolean;
  isError: boolean;
  refetch: () => void;
}

/**
 * Template-kind adapter for the shared controlled-artifact APPROVAL screen.
 *
 * Reuses `useTemplateArtifact` for the base view-model (DRY — no duplicated
 * badge/meta mapping) and overrides `actions` with the ordered, gated approval
 * actions built by `buildTemplateApprovalActions`, plus `decision` (FE-02) — the
 * single construction path for this kind's `ArtifactDecisionModel`, built by
 * `buildTemplateApprovalDecision`.
 *
 * `ArtifactApprovalScreen` does not itself suppress `actions[]` when a
 * `decision` is offered — it renders the decision panel AND an "Outras ações"
 * band for whatever is left in `actions[]`. For under_review that band would
 * otherwise duplicate the decision panel's accept/reject (same gate, same
 * capability), so this adapter zeroes `actions` whenever `decision` is
 * defined. When the signoff gate is denied, `decision` is undefined and the
 * (disabled, reason-tooltipped) accept/reject pair surfaces through the plain
 * `actions[]` band instead — that is the only path `handlers.runSignoff` can
 * ever actually be invoked from, and even then the buttons are disabled.
 *
 * The route still owns the raw `submit` callback (it knows `versionNum` and
 * calls `signoffVersion`) and passes it in via `decisionSubmit`.
 *
 * The approval-flow chain is built from the version's kernel-truth status
 * fields (`buildTemplateApprovalChain`) — the SAME `ApprovalChainItem[]` shape
 * the document mapper produces — so the shared cockpit flow-viz renders
 * identically for both kinds.
 *
 * `useTemplateDetailQuery` is called a second time here; react-query dedupes
 * by key so no extra network request is made.
 */
export function useTemplateApprovalArtifact(
  templateId: string,
  handlers: TemplateApprovalHandlers,
  decisionSubmit: TemplateApprovalDecisionSubmit,
): TemplateApprovalArtifact {
  const base = useTemplateArtifact(templateId);
  const query = useTemplateDetailQuery(templateId);
  const version = query.data?.latest_version ?? null;
  const user = useAuthStore((s) => s.user);
  const actor: ActorContext = { capabilities: user?.capabilities ?? [] };
  const rawActions = version ? buildTemplateApprovalActions(version, actor, handlers) : [];
  const signoffAvailable = rawActions.find((a) => a.key === 'accept')?.available ?? false;
  const signer = user
    ? { displayName: user.displayName, email: user.email ?? null, username: user.username }
    : null;
  const decision = buildTemplateApprovalDecision({
    status: version?.status ?? null,
    signoffAvailable,
    signer,
    submit: decisionSubmit,
  });
  const actions = decision != null ? [] : rawActions;
  const approvalChain = version ? buildTemplateApprovalChain(version) : base.model.approvalChain;
  const model: ArtifactViewModel = { ...base.model, actions, approvalChain, decision };
  return { model, version, isLoading: base.isLoading, isError: base.isError, refetch: base.refetch };
}
