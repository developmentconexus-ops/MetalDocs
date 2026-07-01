import { useAuthStore } from '../../../store/auth.store';
import { formatRevisionCode } from '../../../lib/labels/revisionCode';
import { resolveOwnerDisplay } from '../../shared/controlled-artifact/resolveOwnerDisplay';
import type {
  ArtifactBadge,
  ArtifactKpiCell,
  ArtifactViewModel,
  BreadcrumbItem,
  LifecycleStatus,
} from '../../shared/controlled-artifact/types';
import { useTemplateDetailQuery } from '../queries/useTemplateDetailQuery';
import { templateStatusLabel } from '../lib/templateStatusPresentation';

const EM_DASH = '—';

export interface TemplateArtifact {
  model: ArtifactViewModel;
  isLoading: boolean;
  isError: boolean;
  refetch: () => void;
}

/**
 * Template-kind adapter for the shared controlled-artifact view layer. Composes the
 * template detail query into a kind-agnostic `ArtifactViewModel`.
 *
 * approvalChain is null — templates have no approval-instance/signoff model; only
 * pending reviewer/approver roles exist on the version DTO.
 *
 * actions is [] — the detail screen renders no actions; template approval actions
 * are a later task.
 */
export function useTemplateArtifact(templateId: string): TemplateArtifact {
  const user = useAuthStore((s) => s.user);

  const query = useTemplateDetailQuery(templateId);

  const template = query.data?.template;
  const version = query.data?.latest_version;

  const status = (version?.status ?? '') as LifecycleStatus;
  const revisionLabel = version ? formatRevisionCode(version.revision_number) : null;
  const versionNumber = version?.revision_number ?? version?.version_number ?? 0;
  const badgeLabel = templateStatusLabel(version?.status ?? '');

  // Owner name: created_by; if creator is the current user, use displayName.
  const ownerName = resolveOwnerDisplay(template?.created_by, user) ?? EM_DASH;

  // Hero badges (push in order; skip when source missing).
  const badges: ArtifactBadge[] = [];
  if (template?.key) {
    badges.push({ key: 'key', label: template.key, variant: 'code' });
  }
  if (revisionLabel) {
    badges.push({ key: 'status', label: `${revisionLabel} · ${badgeLabel}`, variant: 'status' });
  }
  if (template?.doc_type_code) {
    badges.push({ key: 'type', label: template.doc_type_code, variant: 'type' });
  }

  const breadcrumb: BreadcrumbItem[] = [
    { label: 'Modelos', href: '/templates' },
    ...(template?.name ? [{ label: template.name }] : []),
  ];

  // KPI strip — honest template-applicable subset only.
  // coverage/pages/nextReview are document-only KPIs and are deliberately OMITTED here,
  // not faked, to avoid presenting fabricated data for a concept that doesn't apply to templates.
  const kpis: ArtifactKpiCell[] = [
    {
      key: 'currentVersion',
      label: 'Versão atual',
      value: revisionLabel ?? EM_DASH,
      hint: badgeLabel,
    },
    {
      key: 'publishedVersion',
      label: 'Versão publicada',
      value: template?.published_version_id != null
        ? formatRevisionCode(template.current_revision_number ?? undefined)
        : EM_DASH,
      hint: template?.published_version_id != null ? 'vigente' : 'não publicada',
    },
  ];

  // No version-list endpoint exists yet; lineage stays empty until one is added.
  const lineage: ArtifactViewModel['lineage'] = [];

  const model: ArtifactViewModel = {
    kind: 'template',
    id: templateId,
    code: null, // templates have no regulated code, only a slug key
    title: template?.name ?? '',
    status,
    versionNumber,
    revisionLabel,
    hero: {
      breadcrumb,
      badges,
      subtitle: null, // templates have no regulated subtitle
    },
    meta: {
      profileLabel: template?.doc_type_code ?? null,
      areaLabel: null,
      visibilityLabel: null,
      fileSizeBytes: null,
      pageCount: null,
      createdAt: template?.created_at ?? null,
      effectiveFrom: null,
      nextReviewAt: null,
      ownerName,
      ownerDescriptor: null,
    },
    kpis,
    approvalChain: null, // templates have no approval-instance/signoff model
    lineage,
    tabs: [{ key: 'documento', label: 'Documento' }], // single tab; templates have no Distribuição
    actions: [],
  };

  return {
    model,
    isLoading: query.isLoading,
    isError: query.isError || !query.data,
    refetch: () => { void query.refetch(); },
  };
}
