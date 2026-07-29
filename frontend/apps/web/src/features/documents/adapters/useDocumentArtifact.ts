import { useAuthStore } from '../../../store/auth.store';
import { useHasCapability } from '../../iam/hooks/useHasCapability';
import { formatRevisionCode } from '../../../lib/labels/revisionCode';
import { formatPageCount, formatPublishedAt, resolveAreaLabel, resolveProfileLabel } from '../lib/documentDetailMeta';
import { resolveOwnerDisplay } from '../../shared/controlled-artifact/resolveOwnerDisplay';
import {
  ACTIVE_SIBLING_STATES,
  getActiveSiblingCtaLabel,
  getActiveSiblingDestination,
  type ActiveSiblingState,
} from '../lib/documentWorkflow';
import { mapApprovalChain } from '../lib/approvalWorkflow';
import type {
  ApprovalChainItem,
  ArtifactBadge,
  ArtifactKpiCell,
  ArtifactViewModel,
  BreadcrumbItem,
  LifecycleStatus,
  VersionHistoryItem,
} from '../../shared/controlled-artifact/types';
import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
import { useApprovalInstanceQuery } from '../queries/useApprovalInstanceQuery';
import { useControlledDocumentActiveDocumentQuery } from '../queries/useControlledDocumentActiveDocumentQuery';
import { useDocumentRevisionHistoryQuery } from '../queries/useDocumentRevisionHistoryQuery';
import { useDistributionSummaryQuery } from '../queries/useDistributionSummaryQuery';
import { useAreasQuery } from '../queries/useAreasQuery';
import { useProfilesQuery } from '../../taxonomy/queries/useProfilesQuery';
import { formatShortDate } from '../../../lib/format/dates';
import { getDocumentStatusPresentation } from '../lib/documentStatusPresentation';
import {
  LIFECYCLE_POLL_INTERVAL_MS,
  getDocumentReleasePresentation,
  isDocumentLifecycleSettling,
  type ReleasePresentation,
} from '../lib/documentReleasePresentation';
import { parseDocumentStatus } from '../lib/parseDocumentStatus';
import type { DocumentDetail } from '../api/documents';
import type { fetchActiveDocumentInstance } from '../../controlled-documents/api/controlledDocuments';

type ControlledDocumentActiveDocument = Awaited<ReturnType<typeof fetchActiveDocumentInstance>>;

const EM_DASH = '—';

/**
 * Lifecycle/capability gating derived by the adapter for the detail hero's action
 * buttons. The route owns all interactive/dialog state (revision composer, PDF
 * export/viewer, copy link) and renders the hero JSX itself — `ArtifactDetailView`
 * takes `heroActions` as a raw ReactNode slot, not a generic action-button list — but
 * the adapter decides WHICH gates are open in the current state so the route consumes
 * these booleans/labels instead of re-deriving them from a second set of queries.
 *
 * ADR 0085 (Stage B): there is no manual publish gate any more. Release is an
 * approval-driven coordinator outcome, so the detail screen only reports the
 * document's status. ADR 0085 (Stage C): release STATUS comes from the
 * readiness-hold projection (`DocumentArtifact.release`), never from gating.
 */
export interface DocumentDetailGating {
  isObsolete: boolean;
  isApproved: boolean;
  isPublished: boolean;
  /** True when the user holds the `document.edit` capability (ADR 0022). */
  canInitiateRevision: boolean;
  /** Gated on canInitiateRevision + isPublished + controlled_document_id + no active sibling. */
  canCreateRevision: boolean;
  activeSiblingDocumentId: string | null;
  activeSiblingState: ActiveSiblingState | null;
  activeSiblingCtaLabel: string;
  activeSiblingDestination: string | null;
}

export interface DocumentArtifact {
  model: ArtifactViewModel;
  isLoading: boolean;
  isError: boolean;
  refetch: () => void;
  /** Raw document detail (route needs doc.name/controlled_document_id/template_version_id/current_revision_id
   *  for the revision composer + PDF export + "Visualizar documento" link — not worth normalizing into the
   *  kind-agnostic model). Null until loaded. */
  doc: DocumentDetail | null;
  /** Distribution coverage count, pre-formatted (EM_DASH on error/missing) — same value as the coverage KPI cell. */
  obligatedCount: string;
  /** Gating booleans + sibling-CTA copy the route's hero-actions JSX renders directly. */
  gating: DocumentDetailGating;
  /**
   * ADR 0085 Stage C — release status derived from the readiness-hold projection
   * that rides the document detail response (no extra query). `null` when the
   * document has no release generation at all (draft / in-review / legacy rows
   * never backfilled): the route then renders no release block whatsoever.
   */
  release: ReleasePresentation | null;
}

/**
 * Document-kind adapter for the shared controlled-artifact view layer. Composes the
 * existing document queries into a kind-agnostic `ArtifactViewModel` AND owns the
 * lifecycle/capability gating derivation the detail route consumes for its hero
 * actions — the route no longer re-runs these queries or re-derives gating itself
 * (FE-02).
 */
export function useDocumentArtifact(documentId: string): DocumentArtifact {
  const user = useAuthStore((s) => s.user);
  // FE-11: revision-initiate gate is capability-based (ADR 0022), not role-based.
  // 'document.edit' is the same capability the backend enforces on
  // POST /controlled-documents/{id}/revisions (see documentWorkflow.ts header comment).
  const canInitiateRevision = useHasCapability('document.edit');

  const docQuery = useDocumentDetailQuery(documentId, { pollLifecycleUntilSettled: true });
  // Same gate the detail query polls on, applied to the sibling queries so the whole
  // screen advances together. Post-ADR-0085 this covers the coordinator's transient
  // release holds (document status stays 'approved' while it works), not just the
  // legacy scheduled wait — see `isDocumentLifecycleSettling` for the transient vs
  // terminal-anomaly split.
  const shouldPollLifecycle = isDocumentLifecycleSettling(docQuery.data);
  const approvalQuery = useApprovalInstanceQuery(documentId);
  const controlledDocumentId = docQuery.data?.controlled_document_id ?? null;
  const activeDocumentQuery = useControlledDocumentActiveDocumentQuery(controlledDocumentId, {
    refetchInterval: shouldPollLifecycle ? LIFECYCLE_POLL_INTERVAL_MS : false,
  });
  const revisionHistoryQuery = useDocumentRevisionHistoryQuery(documentId, {
    refetchInterval: shouldPollLifecycle ? LIFECYCLE_POLL_INTERVAL_MS : false,
  });
  const areasQuery = useAreasQuery();
  const profilesQuery = useProfilesQuery();
  // Coverage denominator — wired into the kpis array (Cobertura cell). EM_DASH on error/missing
  // so we never render a fabricated 0 (parity with DocumentPublishedPage L232-234).
  const distributionSummaryQuery = useDistributionSummaryQuery(documentId);

  const doc = docQuery.data;
  const approval = approvalQuery.data ?? null;
  const activeDocument = activeDocumentQuery.data ?? null;

  const code = doc?.code ?? null;
  // Parse-boundary narrowing (never cast): null = absent/unrecognized wire value,
  // which downstream renders as the honest "sem status" presentation fallback.
  const status: LifecycleStatus | null = parseDocumentStatus(doc?.status);
  const versionNumber = doc?.revision_number ?? 0;
  const revisionLabel = doc ? formatRevisionCode(doc.revision_number) : null;
  const areaCode = doc?.process_area_code_snapshot ?? '';
  const profileCode = doc?.profile_code_snapshot ?? '';
  const areaLabel = areaCode ? resolveAreaLabel(areaCode, areasQuery.data ?? []) : null;
  const profileLabel = profileCode ? resolveProfileLabel(profileCode, profilesQuery.data ?? []) : null;

  // ADR 0085 Stage C: the release projection is the sole source of the publication
  // FACT — `released_at` is when the release actually happened. `approval.completed_at`
  // is approval-COMPLETION time and stands in only for legacy documents that have no
  // release generation at all (never backfilled). A document that HAS a generation but
  // is still on hold has no publication date yet, so we render none rather than
  // relabelling its approval timestamp as a publication date.
  const publicationTimestamp = doc?.release ? doc.release.released_at : (approval?.completed_at ?? null);
  const publishedAt = formatPublishedAt(publicationTimestamp);
  const sinceDateHint = publicationTimestamp ? formatShortDate(publicationTimestamp) : EM_DASH;
  const statusPresentation = getDocumentStatusPresentation(status ?? '', publishedAt);

  // Owner name: created_by; if creator is the current user, use displayName (L217-222 parity).
  const ownerName = resolveOwnerDisplay(doc?.created_by, user) ?? EM_DASH;

  // Coverage denominator (obligated audience) — EM_DASH on error/missing (L232-234 parity).
  const obligatedCount = distributionSummaryQuery.isError
    ? EM_DASH
    : String(distributionSummaryQuery.data?.total_targets ?? EM_DASH);

  // Page count label — canonical formatter (honest em-dash when absent).
  const pageCountLabel = formatPageCount(doc?.current_revision_page_count);

  // scheduled→published-head current version override (L300-309 parity).
  const revisionHistoryItems = revisionHistoryQuery.data?.items ?? [];
  const currentPublishedHistoryItem =
    status === 'scheduled' && activeDocument?.published_document_id
      ? revisionHistoryItems.find((item) => item.document_id === activeDocument.published_document_id) ?? null
      : null;
  const currentVersionLabel = currentPublishedHistoryItem
    ? formatRevisionCode(currentPublishedHistoryItem.revision_number)
    : (doc ? formatRevisionCode(doc.revision_number) : EM_DASH);
  const currentVersionHint = currentPublishedHistoryItem ? EM_DASH : sinceDateHint;

  // Hero badges (data only): code chip, version·status pill, profile/type label.
  const badges: ArtifactBadge[] = [];
  if (code) {
    badges.push({ key: 'code', label: code, variant: 'code' });
  }
  const isObsolete = status === 'obsolete';
  if (!isObsolete && revisionLabel) {
    badges.push({ key: 'status', label: `${revisionLabel} · ${statusPresentation.badgeLabel}`, variant: 'status' });
  }
  if (profileLabel) {
    badges.push({ key: 'type', label: profileLabel, variant: 'type' });
  }

  const breadcrumb: BreadcrumbItem[] = [
    { label: 'Biblioteca', href: '/documents' },
    ...(areaLabel ? [{ label: areaLabel }] : []),
    ...(code ? [{ label: code }] : []),
  ];

  // Approval chain: one ApprovalChainItem per stage ACTOR (F-QA4-8) — decided
  // actors plus the still-eligible ones, each with its backend-resolved display
  // name. Consumers group by stageIndex, so a stage may contribute several rows.
  const approvalChain: ApprovalChainItem[] | null = approval ? mapApprovalChain(approval.stages) : null;

  const lineage: VersionHistoryItem[] = revisionHistoryItems.map((item) => ({
    versionNumber: item.revision_number ?? 0,
    revisionNumber: item.revision_number ?? null,
    revisionLabel: formatRevisionCode(item.revision_number),
    status: parseDocumentStatus(item.status),
    title: item.revision_title ?? null,
    createdAt: item.created_at ?? null,
    isCurrent: Boolean(item.is_current),
  }));

  // Lifecycle / capability flags (parity with DocumentPublishedPage RBAC logic).
  const isPublished = status === 'published';
  const isApproved = status === 'approved';
  const activeSiblingDocumentId =
    activeDocument?.document_id &&
    activeDocument.document_id !== documentId &&
    ACTIVE_SIBLING_STATES.includes(activeDocument.approval_state as ActiveSiblingState)
      ? activeDocument.document_id
      : null;
  const activeSiblingStateCandidate = activeDocument?.approval_state as ActiveSiblingState | undefined;
  const activeSiblingState =
    activeSiblingDocumentId && activeSiblingStateCandidate && ACTIVE_SIBLING_STATES.includes(activeSiblingStateCandidate)
      ? activeSiblingStateCandidate
      : null;
  const activeSiblingCtaLabel = activeSiblingState ? getActiveSiblingCtaLabel(activeSiblingState) : 'Iniciar revisão';
  const activeSiblingDestination =
    activeSiblingDocumentId && activeSiblingState
      ? getActiveSiblingDestination(activeSiblingDocumentId)
      : null;
  const canCreateRevision =
    canInitiateRevision && isPublished && Boolean(doc?.controlled_document_id) && activeSiblingDocumentId == null;

  // KPI strip cells — fully composed here so the shared view has zero kind awareness.
  // Cell order matches DocumentPublishedPage L579-608: currentVersion, coverage, nextReview, pages.
  const kpis: ArtifactKpiCell[] = [
    {
      key: 'currentVersion',
      label: 'Versão atual',
      value: currentVersionLabel,
      hint: currentVersionHint !== EM_DASH ? `desde ${currentVersionHint}` : statusPresentation.ownerMeta,
    },
    {
      key: 'coverage',
      label: 'Cobertura',
      value: obligatedCount,
      hint: 'destinatários obrigados · abrir fanout →',
      href: 'distribution',
    },
    {
      key: 'nextReview',
      label: 'Próxima revisão',
      value: doc?.review_due_at ? formatShortDate(doc.review_due_at) : EM_DASH,
      hint: doc?.review_due_at ? 'ciclo de revisão periódica' : 'sem data de revisão definida',
    },
    {
      key: 'pages',
      label: 'Páginas',
      value: pageCountLabel,
      hint: 'metadado do arquivo',
    },
  ];

  // T11b: real run() handlers (submit/review/approve/reject/publish/createVersion) are
  // dialog-driven and owned by DocumentDetailRoute — wired in T11b. The detail adapter
  // emits an empty array; the approval-screen adapter will emit the ordered action list.
  const actions = [] as ArtifactViewModel['actions'];

  const model: ArtifactViewModel = {
    kind: 'document',
    id: documentId,
    code,
    title: doc?.name ?? code ?? '',
    status,
    versionNumber,
    revisionLabel,
    hero: {
      breadcrumb,
      badges,
      subtitle: statusPresentation.subtitle,
    },
    meta: {
      profileLabel,
      areaLabel,
      visibilityLabel: null,
      fileSizeBytes: doc?.current_revision_file_size_bytes ?? null,
      pageCount: doc?.current_revision_page_count ?? null,
      createdAt: doc?.created_at ?? null,
      // M6 F6.2: effective_from/review_due_at now flow from the document
      // response (contract-first, T6). Fall back to the approval-completion
      // timestamp for effectiveFrom when the document has no explicit
      // effective_from set (legacy rows / pre-review-model documents).
      effectiveFrom: doc?.effective_from ?? approval?.completed_at ?? null,
      nextReviewAt: doc?.review_due_at ?? null,
      ownerName,
      ownerDescriptor: statusPresentation.ownerMeta,
    },
    kpis,
    approvalChain,
    lineage,
    tabs: [
      { key: 'documento', label: 'Documento', href: '.' },
      { key: 'distribuicao', label: 'Distribuição', href: 'distribution' },
    ],
    actions,
  };

  // ADR 0085 Stage C: the coordinator's readiness-hold projection is the ONLY
  // source of release status. Nothing here infers readiness from content hashes,
  // capabilities, or the lifecycle status — those answer different questions.
  const release = getDocumentReleasePresentation(doc?.release);

  const gating: DocumentDetailGating = {
    isObsolete,
    isApproved,
    isPublished,
    canInitiateRevision,
    canCreateRevision,
    activeSiblingDocumentId,
    activeSiblingState,
    activeSiblingCtaLabel,
    activeSiblingDestination,
  };

  return {
    model,
    isLoading: docQuery.isLoading,
    isError: docQuery.isError || !doc,
    refetch: () => {
      void docQuery.refetch();
    },
    doc: doc ?? null,
    obligatedCount,
    gating,
    release,
  };
}
