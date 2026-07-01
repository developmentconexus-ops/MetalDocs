import type { ArtifactViewModel } from '../../shared/controlled-artifact/types';
import type { VersionDTO } from '../api/templates';
import type { ActorContext } from '../lib/canActOnVersion';
import { buildTemplateApprovalActions, type TemplateApprovalHandlers } from '../lib/templateApprovalActions';
import { buildTemplateApprovalChain } from '../lib/templateApprovalChain';
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
 * badge/meta mapping) and overrides only `actions` with the ordered, gated
 * approval actions built by `buildTemplateApprovalActions`.
 *
 * The approval-flow chain is built from the version's inline submit/review/approve
 * fields (`buildTemplateApprovalChain`) — the SAME `ApprovalChainItem[]` shape the
 * document mapper produces — so the shared cockpit flow-viz renders identically for
 * both kinds. (The former `approvalChain: null` was an adapter shortcut; templates
 * are not signoff-less.)
 *
 * `useTemplateDetailQuery` is called a second time here; react-query deduplicates
 * by key so no extra network request is made.
 */
export function useTemplateApprovalArtifact(
  templateId: string,
  handlers: TemplateApprovalHandlers,
): TemplateApprovalArtifact {
  const base = useTemplateArtifact(templateId);
  const query = useTemplateDetailQuery(templateId);
  const version = query.data?.latest_version ?? null;
  const user = useAuthStore((s) => s.user);
  const actor: ActorContext = { roles: user?.roles ?? [], capabilities: user?.capabilities ?? [] };
  const actions = version ? buildTemplateApprovalActions(version, actor, handlers) : [];
  const approvalChain = version ? buildTemplateApprovalChain(version) : base.model.approvalChain;
  const model: ArtifactViewModel = { ...base.model, actions, approvalChain };
  return { model, version, isLoading: base.isLoading, isError: base.isError, refetch: base.refetch };
}
