import { useAuthStore } from '../../../store/auth.store';
import { formatRevisionCode } from '../../../lib/labels/revisionCode';
import { formatPageCount, formatPublishedAt, resolveAreaLabel, resolveProfileLabel } from '../lib/documentDetailMeta';
import { resolveOwnerDisplay } from '../../shared/controlled-artifact/resolveOwnerDisplay';
import {
  ACTIVE_SIBLING_STATES,
  canInitiateRevision as canInitiateRevisionGate,
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

const EM_DASH = '—';

type StatusPresentation = {
  badgeLabel: string;
  subtitle: string | null;
  ownerMeta: string;
};

// Lifted verbatim from DocumentPublishedPage so governed status copy is preserved byte-for-byte.
// ownerMeta is intentionally DISTINCT from subtitle for scheduled/superseded (governed requirement).
function getDocumentStatusPresentation(status: string, publishedAt: string): StatusPresentation {
  const hasPublishedAt = publishedAt !== EM_DASH;

  switch (status) {
    case 'approved':
      return {
        badgeLabel: 'aprovado',
        subtitle: hasPublishedAt ? `aprovado em ${publishedAt}` : 'Aprovado',
        ownerMeta: hasPublishedAt ? `aprovado em ${publishedAt}` : 'aprovado',
      };
    case 'scheduled':
      return {
        badgeLabel: 'agendado',
        subtitle: hasPublishedAt ? `aprovação concluída em ${publishedAt}` : 'Publicação agendada',
        ownerMeta: hasPublishedAt ? `publicação agendada · aprovado em ${publishedAt}` : 'publicação agendada',
      };
    case 'published':
      return {
        badgeLabel: 'publicado',
        subtitle: hasPublishedAt ? `publicado em ${publishedAt}` : 'Publicado',
        ownerMeta: hasPublishedAt ? `publicado em ${publishedAt}` : 'publicado',
      };
    case 'superseded':
      return {
        badgeLabel: 'substituído',
        subtitle: hasPublishedAt ? `publicado em ${publishedAt}` : 'Substituído por revisão posterior',
        ownerMeta: hasPublishedAt ? `substituído · publicado em ${publishedAt}` : 'substituído por revisão posterior',
      };
    case 'obsolete':
      return {
        badgeLabel: 'obsoleto',
        subtitle: hasPublishedAt ? `obsoleto após publicação em ${publishedAt}` : 'Documento obsoleto',
        ownerMeta: hasPublishedAt ? `obsoleto · publicado em ${publishedAt}` : 'obsoleto',
      };
    case 'draft':
      return { badgeLabel: 'rascunho', subtitle: 'Rascunho — ainda não publicado', ownerMeta: 'rascunho' };
    case 'under_review':
      return { badgeLabel: 'em revisão', subtitle: 'Aguardando decisão de aprovação', ownerMeta: 'em revisão' };
    case 'rejected':
      return { badgeLabel: 'rejeitado', subtitle: 'Revisão rejeitada', ownerMeta: 'rejeitado' };
    default:
      return {
        badgeLabel: status || 'sem status',
        subtitle: hasPublishedAt ? `publicado em ${publishedAt}` : null,
        ownerMeta: hasPublishedAt ? `publicado em ${publishedAt}` : (status || EM_DASH),
      };
  }
}

export interface DocumentArtifact {
  model: ArtifactViewModel;
  isLoading: boolean;
  isError: boolean;
  refetch: () => void;
}

/**
 * Document-kind adapter for the shared controlled-artifact view layer. Composes the
 * existing document queries into a kind-agnostic `ArtifactViewModel`.
 *
 * Lifecycle/capability flags drive the action-set `available` booleans; the full
 * interactive run() handlers are dialog-driven and owned by the route wrapper
 * (DocumentDetailRoute) — completed in T11.
 */
export function useDocumentArtifact(documentId: string): DocumentArtifact {
  const user = useAuthStore((s) => s.user);

  const docQuery = useDocumentDetailQuery(documentId, { pollScheduledLifecycle: true });
  const shouldPollScheduledLifecycle = docQuery.data?.status === 'scheduled';
  const scheduledLifecycleRefetchInterval = 5_000;
  const approvalQuery = useApprovalInstanceQuery(documentId);
  const controlledDocumentId = docQuery.data?.controlled_document_id ?? null;
  const activeDocumentQuery = useControlledDocumentActiveDocumentQuery(controlledDocumentId, {
    refetchInterval: shouldPollScheduledLifecycle ? scheduledLifecycleRefetchInterval : false,
  });
  const revisionHistoryQuery = useDocumentRevisionHistoryQuery(documentId, {
    refetchInterval: shouldPollScheduledLifecycle ? scheduledLifecycleRefetchInterval : false,
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
  const status = (doc?.status ?? '') as LifecycleStatus;
  const versionNumber = doc?.revision_number ?? 0;
  const revisionLabel = doc ? formatRevisionCode(doc.revision_number) : null;
  const areaCode = doc?.process_area_code_snapshot ?? '';
  const profileCode = doc?.profile_code_snapshot ?? '';
  const areaLabel = areaCode ? resolveAreaLabel(areaCode, areasQuery.data ?? []) : null;
  const profileLabel = profileCode ? resolveProfileLabel(profileCode, profilesQuery.data ?? []) : null;

  const publishedAt = formatPublishedAt(approval?.completed_at ?? undefined);
  const sinceDateHint = approval?.completed_at ? formatShortDate(approval.completed_at) : EM_DASH;
  const statusPresentation = getDocumentStatusPresentation(status, publishedAt);

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

  // Approval chain: one ApprovalChainItem per signoff slot, grouped downstream by stageIndex.
  // Mirrors the page's "first signoff per stage" timeline (one actor per stage in current workflow).
  const approvalChain: ApprovalChainItem[] | null = approval ? mapApprovalChain(approval.stages) : null;

  const lineage: VersionHistoryItem[] = revisionHistoryItems.map((item) => ({
    versionNumber: item.revision_number ?? 0,
    revisionNumber: item.revision_number ?? null,
    revisionLabel: formatRevisionCode(item.revision_number),
    status: (item.status ?? '') as LifecycleStatus,
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
  const canInitiateRevision = canInitiateRevisionGate(user);
  const canCreateRevision =
    canInitiateRevision && isPublished && Boolean(doc?.controlled_document_id) && activeSiblingDocumentId == null;
  const canPublish = canInitiateRevision && isApproved && Boolean(activeDocument?.content_hash);

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
      href: '/distribution',
    },
    {
      key: 'nextReview',
      label: 'Próxima revisão',
      value: EM_DASH,
      hint: 'sem data de revisão definida',
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
      effectiveFrom: approval?.completed_at ?? null,
      nextReviewAt: null,
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

  return {
    model,
    isLoading: docQuery.isLoading,
    isError: docQuery.isError || !doc,
    refetch: () => {
      void docQuery.refetch();
    },
  };
}
