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

/**
 * Discriminated hero-action "kind" for the template detail screen, one per lifecycle
 * status. Replaces a hand-rolled `switch (model.status)` that used to live in
 * `TemplateDetailRoute` — the adapter owns the status→hero-kind derivation (single
 * source of truth), the route only maps each kind to its button JSX + handlers.
 *
 * - `review`    → under_review: primary "Revisar versão" + secondary "Ver no editor".
 * - `publish`   → approved: primary "Publicar versão" + secondary "Ver no editor".
 * - `newVersion`→ published: primary "Criar nova versão" (route-owned mutation) +
 *                 secondary "Ver no editor".
 * - `editOnly`  → draft: primary "Editar template" + disabled "Criar nova versão" hint.
 * - `readOnly`  → obsolete / unknown: single "Ver no editor" entry.
 */
export type TemplateDetailHeroKind = 'review' | 'publish' | 'newVersion' | 'editOnly' | 'readOnly';

export function resolveTemplateDetailHeroKind(status: string): TemplateDetailHeroKind {
  switch (status) {
    case 'under_review':
      return 'review';
    case 'approved':
      return 'publish';
    case 'published':
      return 'newVersion';
    case 'draft':
      return 'editOnly';
    default:
      return 'readOnly';
  }
}

export interface TemplateArtifact {
  model: ArtifactViewModel;
  isLoading: boolean;
  isError: boolean;
  refetch: () => void;
  /** Status-driven hero-action kind — see `TemplateDetailHeroKind`. */
  heroKind: TemplateDetailHeroKind;
}

/**
 * Template-kind adapter for the shared controlled-artifact view layer. Composes the
 * template detail query into a kind-agnostic `ArtifactViewModel` and derives the
 * detail-screen hero-action kind (FE-02) so `TemplateDetailRoute` stays thin
 * slot-assembly instead of hand-rolling a `switch (status)` hero-machine.
 *
 * approvalChain is null — templates have no approval-instance/signoff model; only
 * pending reviewer/approver roles exist on the version DTO.
 *
 * actions is [] — the detail screen renders no actions; template approval actions
 * are owned by `useTemplateApprovalArtifact` (the approval-screen adapter).
 */
export function useTemplateArtifact(templateId: string): TemplateArtifact {
  const user = useAuthStore((s) => s.user);

  const query = useTemplateDetailQuery(templateId);

  const template = query.data?.template;
  const version = query.data?.latest_version;

  // No cast needed: the generated VersionDTO.status is already the narrow template
  // enum (draft | under_review | approved | published | obsolete), a subset of
  // LifecycleStatus. Null = no version loaded yet (honest absence, not '').
  const status: LifecycleStatus | null = version?.status ?? null;
  const revisionLabel = version ? formatRevisionCode(version.revision_number) : null;
  const versionNumber = version?.revision_number ?? version?.version_number ?? 0;
  const badgeLabel = templateStatusLabel(version?.status ?? '');

  // Owner name: prefer the API-resolved created_by_display_name (FE-08); if
  // absent, fall back to the shared created_by resolution (current-user
  // displayName override, else the raw id).
  const ownerName =
    template?.created_by_display_name ?? resolveOwnerDisplay(template?.created_by, user) ?? EM_DASH;

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
    { label: 'Templates', href: '/templates' },
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
      value: template?.published_version != null
        ? formatRevisionCode(template.published_version.revision_number)
        : EM_DASH,
      hint: template?.published_version != null ? 'vigente' : 'não publicada',
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
    heroKind: resolveTemplateDetailHeroKind(status ?? ''),
  };
}
