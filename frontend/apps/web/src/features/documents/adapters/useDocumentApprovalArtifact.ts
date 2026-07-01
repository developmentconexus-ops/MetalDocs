import { useCallback, useEffect, useState } from 'react';

import { getInstance } from '../../approval/api/approvalApi';
import { etagCache } from '../../approval/api/etagCache';
import type { ApprovalInstance } from '../../approval/api/approvalTypes';
import { formatRevisionCode } from '../../../lib/labels/revisionCode';
import { useAuthStore } from '../../../store/auth.store';
import type {
  ApprovalChainItem,
  ArtifactAction,
  ArtifactViewModel,
  LifecycleStatus,
} from '../../shared/controlled-artifact/types';
import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
import { useActiveDocumentContextQuery } from '../../approval/queries/useActiveDocumentContextQuery';
import { TRANSITION_POLICY, type TransitionPolicy, toApprovalState, mapApprovalChain } from '../lib/approvalWorkflow';
import { resolveOwnerDisplay } from '../../shared/controlled-artifact/resolveOwnerDisplay';

/**
 * Route-supplied dialog/prompt openers bound to each emitted action's `run`.
 * The route owns all interactive state (dialog visibility, canvas ref, prompts);
 * the adapter only decides WHICH actions are allowed in the current state and
 * wires their `run` to these handlers.
 */
export interface DocumentApprovalHandlers {
  /** Open the inline submit route-picker. */
  openSubmit: () => void;
  /** Open the CancelInstanceDialog to collect a reason and cancel the approval instance. */
  cancelInstance: () => void;
  /** Open the SupersedePublishDialog. */
  openPublish: () => void;
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
  /** True when the last instance fetch is older than 30s (staleness banner). */
  isStale: boolean;
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
 * The plain actions appear per state (submit → cancel → publish); signing routes
 * through the shared DecisionPanel instead of a button (gated on
 * `policy.actions.signoff` in the route). The 404 fallback seeds a "v0" ETag so a
 * cold submit still sends a valid If-Match.
 */
export function useDocumentApprovalArtifact(
  documentId: string,
  handlers: DocumentApprovalHandlers,
): DocumentApprovalArtifact {
  const user = useAuthStore((s) => s.user);
  const docQuery = useDocumentDetailQuery(documentId);
  const doc = docQuery.data ?? null;
  const controlledDocumentId = doc?.controlled_document_id ?? '';

  const contextQuery = useActiveDocumentContextQuery(controlledDocumentId);
  const context = contextQuery.data ?? null;

  const [instance, setInstance] = useState<ApprovalInstance | null>(null);
  const [instanceLoading, setInstanceLoading] = useState(false);
  const [instanceError, setInstanceError] = useState<string | null>(null);
  const [lastFetchedAt, setLastFetchedAt] = useState<number | null>(null);
  const [now, setNow] = useState(() => Date.now());

  // Resolve the approval-context fields (with document-detail fallbacks) exactly
  // as the old page did before constructing the panel.
  const hasActiveContext = Boolean(context && context.content_hash != null);
  const approvalState = (context?.approval_state ?? doc?.status ?? 'draft') as string;
  const contentHash = context?.content_hash ?? null;
  const revisionVersion = context?.revision_version ?? doc?.revision_version ?? 0;
  const lockedByInstanceId = context?.approval_instance_id ?? undefined;
  const publishedDocumentId = context?.published_document_id ?? undefined;

  const refetchInstance = useCallback(async () => {
    if (!documentId) {
      return;
    }
    setInstanceLoading(true);
    setInstanceError(null);
    try {
      const next = await getInstance(documentId);
      setInstance(next);
      setLastFetchedAt(Date.now());
    } catch (err) {
      if ((err as { status?: number }).status === 404) {
        setInstance(null);
        setLastFetchedAt(Date.now());
        if (!etagCache.get(documentId)) etagCache.set(documentId, '"v0"');
      } else {
        setInstanceError('Erro ao carregar dados de aprovação.');
        setInstance(null);
      }
    } finally {
      setInstanceLoading(false);
    }
  }, [documentId]);

  // Fetch the instance once the active context is confirmed (parity: the panel
  // only mounted — and thus only fetched — when content_hash was present).
  useEffect(() => {
    if (hasActiveContext) {
      void refetchInstance();
    }
  }, [hasActiveContext, refetchInstance]);

  // 30s staleness clock (1s tick).
  useEffect(() => {
    const timer = window.setInterval(() => {
      setNow(Date.now());
    }, 1000);
    return () => window.clearInterval(timer);
  }, []);

  const isStale = lastFetchedAt != null && now - lastFetchedAt > 30_000;

  const status = toApprovalState(approvalState);
  const policy = TRANSITION_POLICY[status];

  // Approval chain: one ApprovalChainItem per stage (first signoff per stage),
  // matching the document detail adapter's mapping.
  const approvalChain: ApprovalChainItem[] | null = instance ? mapApprovalChain(instance.stages) : null;

  // Ordered, gated actions — emit ONLY the allowed actions, in display order
  // (submit → cancel → publish), matching the old "button appears only when
  // allowed" behavior. Signing is NOT a plain action: it routes exclusively
  // through the shared DecisionPanel, gated on `policy.actions.signoff` in the
  // route (SignoffDetailPage), so no 'signoff' button is emitted here.
  const actions: ArtifactAction[] = [];
  if (policy.actions.submit) {
    actions.push({
      key: 'submit',
      label: 'Submeter para revisão',
      variant: 'primary',
      available: true,
      run: handlers.openSubmit,
    });
  }
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
        status: (doc.status ?? '') as LifecycleStatus,
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
    noActiveContext: Boolean(doc) && !contextQuery.isLoading && !contextQuery.isError && !hasActiveContext,
    refetchInstance,
    isStale,
  };
}
