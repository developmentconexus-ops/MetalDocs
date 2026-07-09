import { useCallback } from 'react';

import type { ApprovalInstance } from '../../approval/api/approvalTypes';
import { deriveWorkspaceMode } from '../../approval/lib/workspaceMode';
import { useApprovalInstanceQuery } from '../../approval/queries/useApprovalInstanceQuery';
import { formatRevisionCode } from '../../../lib/labels/revisionCode';
import { useAuthStore } from '../../../store/auth.store';
import type {
  ApprovalChainItem,
  ArtifactAction,
  ArtifactDecisionModel,
  ArtifactViewModel,
} from '../../shared/controlled-artifact/types';
import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
import { useControlledDocumentActiveDocumentQuery } from '../queries/useControlledDocumentActiveDocumentQuery';
import { TRANSITION_POLICY, type TransitionPolicy, toApprovalState, mapApprovalChain } from '../lib/approvalWorkflow';
import { resolveOwnerDisplay } from '../../shared/controlled-artifact/resolveOwnerDisplay';
import { buildDocumentSignoffDecision } from '../lib/documentSignoffDecision';
import { parseDocumentStatus } from '../lib/parseDocumentStatus';

/**
 * Route-supplied dialog/prompt openers bound to each emitted action's `run`.
 * The route owns all interactive state (dialog visibility, canvas ref, prompts);
 * the adapter only decides WHICH actions are allowed in the current state and
 * wires their `run` to these handlers.
 */
export interface DocumentApprovalHandlers {
  /** Open the CancelInstanceDialog to collect a reason and cancel the approval instance. */
  cancelInstance: () => void;
  /** Open the SupersedePublishDialog. */
  openPublish: () => void;
}

/**
 * Route-owned inputs to the sign-off decision (FE-02). The route keeps ownership of
 * the review-canvas ref (`flushSave`) and the signoff mutation (the actual submit
 * sequence) since those are interactive/imperative concerns outside the adapter's
 * query+gating responsibility; the adapter only decides WHETHER to offer the
 * decision and constructs its shape via `buildDocumentSignoffDecision`.
 */
export interface DocumentSignoffDecisionInputs {
  /** Preselects a decision option from the `?decision=` query param. */
  defaultOptionKey: 'approve' | 'reject' | null;
  /** Flush pending canvas edits, call the signoff mutation, then refetch the instance. */
  submit: (input: { optionKey: string; reason: string; password: string }) => Promise<void>;
}

export interface DocumentApprovalArtifact {
  /** Kind-agnostic view-model for the shared approval screen. Null until the
   *  document detail has loaded (the screen is not rendered before then). */
  model: ArtifactViewModel | null;
  /** Raw approval instance (timeline + signoff dialog inputs). Null = no active instance. */
  instance: ApprovalInstance | null;
  approvalState: string;
  contentHash: string | null;
  revisionVersion: number;
  lockedByInstanceId?: string;
  publishedDocumentId?: string;
  /** The active TRANSITION_POLICY entry — drives disabledReason / readOnly copy in extras. */
  policy: TransitionPolicy;
  /** True while the document detail OR the approval instance is loading. */
  loading: boolean;
  /** Document-level load error message (the page-level error state). Null when healthy. */
  error: string | null;
  /** True when the active-document context query has errored (sidebar error state). */
  contextError: boolean;
  /** True once the document loaded but has no active approval context. */
  noActiveContext: boolean;
  /** Imperative instance refetch — seeds etagCache via getInstance. */
  refetchInstance: () => Promise<void>;
}

/**
 * Document-kind adapter for the shared controlled-artifact APPROVAL screen.
 *
 * Owns the imperative `getInstance` fetch (which seeds `etagCache` from the
 * response ETag — required before any If-Match signoff/publish write), the
 * instance state + refetch, the 30s staleness clock, and the composition of the
 * `ArtifactViewModel` (approvalChain + the ordered, gated `actions`). The route
 * wrapper owns dialog state and passes the openers in via `handlers`.
 *
 * The plain actions appear per state (cancel → publish); signing routes through
 * the shared DecisionPanel instead of a button (offered when the workspace mode
 * is `approving` — server-derived viewer truth, not document status). The cockpit
 * is approver-only: submitting a document for review
 * happens exclusively on the document editor, so there is no cold-submit path or
 * ETag seed here — when there is no active instance the document is in draft and
 * no cockpit write is possible.
 */
export function useDocumentApprovalArtifact(
  documentId: string,
  handlers: DocumentApprovalHandlers,
  decisionInputs: DocumentSignoffDecisionInputs,
): DocumentApprovalArtifact {
  const user = useAuthStore((s) => s.user);
  const docQuery = useDocumentDetailQuery(documentId);
  const doc = docQuery.data ?? null;
  const controlledDocumentId = doc?.controlled_document_id ?? '';

  const contextQuery = useControlledDocumentActiveDocumentQuery(controlledDocumentId);
  const context = contextQuery.data ?? null;

  // Resolve the approval-context fields (with document-detail fallbacks) exactly
  // as the old page did before constructing the panel.
  const hasActiveContext = Boolean(context && context.content_hash != null);
  const approvalState = (context?.approval_state ?? doc?.status ?? 'draft') as string;
  const contentHash = context?.content_hash ?? null;
  const revisionVersion = context?.revision_version ?? doc?.revision_version ?? 0;
  const lockedByInstanceId = context?.approval_instance_id ?? undefined;
  const publishedDocumentId = context?.published_document_id ?? undefined;

  const instanceQuery = useApprovalInstanceQuery(documentId, hasActiveContext);
  const instance = instanceQuery.data ?? null;
  // isFetching (not isLoading): the old imperative adapter set loading on EVERY
  // fetch incl. manual refetch; v5 isLoading is first-fetch-only, so a post-404
  // manual refetch would not surface the spinner. isFetching preserves the old
  // semantics (still masked by `instance == null` once a fetch has resolved data).
  const instanceLoading = instanceQuery.isFetching;
  const instanceError = instanceQuery.isError ? 'Erro ao carregar dados de aprovação.' : null;
  const refetchInstance = useCallback(async () => {
    await instanceQuery.refetch();
  }, [instanceQuery]);

  const status = toApprovalState(approvalState);
  const policy = TRANSITION_POLICY[status];

  // Approval chain: one ApprovalChainItem per stage (first signoff per stage),
  // matching the document detail adapter's mapping.
  const approvalChain: ApprovalChainItem[] | null = instance ? mapApprovalChain(instance.stages) : null;

  // Ordered, gated actions — emit ONLY the allowed actions, in display order
  // (cancel → publish), matching the old "button appears only when allowed"
  // behavior. Signing is NOT a plain action: it routes exclusively through the
  // shared DecisionPanel, offered via the workspace-mode selector (mode
  // 'approving'), so no 'signoff' button is emitted here. The cockpit is
  // approver-only, so there is no 'submit' action either — submitting for review
  // lives exclusively on the document editor.
  const actions: ArtifactAction[] = [];
  if (policy.actions.cancelInstance) {
    actions.push({
      key: 'cancel',
      label: 'Cancelar instância',
      variant: 'secondary',
      available: true,
      run: handlers.cancelInstance,
    });
  }
  if (policy.actions.publishOrSchedule) {
    actions.push({
      key: 'publish',
      label: 'Publicar / Agendar',
      variant: 'primary',
      available: true,
      run: handlers.openPublish,
    });
  }

  // Same gates the route (SignoffDetailPage) used to compute inline: the sidebar
  // decision panel is only ready once the document loaded, the active-context query
  // settled with a confirmed context (not loading/error/absent), and the instance
  // fetch itself didn't error.
  const noActiveContext = Boolean(doc) && !contextQuery.isLoading && !contextQuery.isError && !hasActiveContext;
  const contextLoading = Boolean(doc) && !contextQuery.isError && !noActiveContext && contentHash == null;
  const sidebarReady = !contextLoading && !contextQuery.isError && !noActiveContext && !instanceError && contentHash != null;

  // The document sign-off carries legal e-signature weight: a password re-auth + a
  // legal-effect confirmation. Offered only when the active context is confirmed and
  // policy allows signing on a locked instance (FE-02: single decision construction
  // path, owned by `buildDocumentSignoffDecision` — was inline in SignoffDetailPage).
  // Stage eligibility is server-derived (viewer truth), NOT document status: the
  // signature panel is offered only when the workspace mode is `approving` (an
  // approval-kind active stage the caller is eligible for). This is the F2d.3 fix
  // for the M2c defect where the old status-based `policy.actions.signoff` gate
  // offered signoff on a review stage → 412 content_hash_mismatch.
  const signoffOffered =
    sidebarReady &&
    deriveWorkspaceMode(doc, instance, instance?.viewer ?? null) === 'approving' &&
    lockedByInstanceId != null &&
    instance != null &&
    contentHash != null;

  const decision: ArtifactDecisionModel | undefined = buildDocumentSignoffDecision({
    offered: signoffOffered,
    signer: user
      ? { displayName: user.displayName, email: user.email ?? null, username: user.username }
      : null,
    defaultOptionKey: decisionInputs.defaultOptionKey,
    submit: decisionInputs.submit,
  });

  // The approval cockpit deliberately builds a REDUCED ArtifactViewModel here
  // rather than composing `useDocumentArtifact` (the way `useTemplateApprovalArtifact`
  // composes `useTemplateArtifact`). The cockpit hero uses an "Aprovações" breadcrumb,
  // a single code chip, and empty profile/area/visibility meta — the sign-off surface
  // intentionally omits the detail screen's richer identity block. Both this hook and
  // `useDocumentArtifact` read the SAME `useDocumentDetailQuery` key, so there is no
  // extra network cost; the divergence is presentational scope, not a second source of
  // truth. If the cockpit ever needs the full identity block, compose the detail model
  // and override hero/meta instead of widening this literal. (Reviewer M-6, T11 follow-up.)
  const model: ArtifactViewModel | null = doc
    ? {
        kind: 'document',
        id: documentId,
        code: doc.code ?? null,
        title: doc.name ?? doc.code ?? '',
        status: parseDocumentStatus(doc.status),
        versionNumber: doc.revision_number ?? 0,
        revisionLabel: formatRevisionCode(doc.revision_number),
        hero: {
          breadcrumb: [
            { label: 'Aprovações', href: '/approvals' },
            ...(doc.code ? [{ label: doc.code }] : []),
          ],
          badges: [
            ...(doc.code ? [{ key: 'code', label: doc.code, variant: 'code' as const }] : []),
          ],
          subtitle: null,
        },
        meta: {
          profileLabel: null,
          areaLabel: null,
          visibilityLabel: null,
          fileSizeBytes: doc.current_revision_file_size_bytes ?? null,
          pageCount: doc.current_revision_page_count ?? null,
          createdAt: doc.created_at ?? null,
          effectiveFrom: null,
          nextReviewAt: null,
          ownerName: resolveOwnerDisplay(doc.created_by, user),
          ownerDescriptor: null,
        },
        kpis: [],
        approvalChain,
        lineage: [],
        tabs: [],
        actions,
        decision,
      }
    : null;

  return {
    model,
    instance,
    approvalState,
    contentHash,
    revisionVersion,
    lockedByInstanceId,
    publishedDocumentId,
    policy,
    loading: docQuery.isLoading || (hasActiveContext && instanceLoading && instance == null),
    error: instanceError,
    contextError: contextQuery.isError,
    noActiveContext,
    refetchInstance,
  };
}
