import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Icon } from '../../../components/ui/Icon';
import { Avatar } from '../../../components/ui/Avatar';
import { CodeChip } from '../../../components/ui/CodeChip';
import { resolveQueryError, ApiError, resolveErrorMessage } from '../../../lib/api';
import { DocumentHero } from '../components/DocumentHero';
import { DocumentVersionTimeline } from '../components/DocumentVersionTimeline';
import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
import { useApprovalInstanceQuery } from '../queries/useApprovalInstanceQuery';
import { useControlledDocumentActiveDocumentQuery } from '../queries/useControlledDocumentActiveDocumentQuery';
import { useDocumentRevisionHistoryQuery } from '../queries/useDocumentRevisionHistoryQuery';
import { useDistributionSummaryQuery } from '../queries/useDistributionSummaryQuery';
import { useAreasQuery } from '../queries/useAreasQuery';
import { useProfilesQuery } from '../../taxonomy/queries/useProfilesQuery';
import { createRevision } from '../../controlled-documents/api/controlledDocuments';
import { exportPDF } from '../api/exports';
import { SupersedePublishDialog } from '../../approval/components/SupersedePublishDialog';
import { useAuthStore } from '../../../store/auth.store';
import { useHasCapability } from '../../iam/hooks/useHasCapability';
import {
  formatPublishedAt,
  formatRevisionCode,
  formatSignedAt,
  formatShortDate,
  formatFileSize,
  formatPageCount,
  resolveAreaLabel,
  resolveProfileLabel,
} from '../lib/documentDetailMeta';
import styles from './DocumentPublishedPage.module.css';

const ACTIVE_SIBLING_STATES = ['draft', 'under_review', 'approved', 'scheduled', 'rejected'] as const;
type ActiveSiblingState = (typeof ACTIVE_SIBLING_STATES)[number];

type DocumentStatusPresentation = {
  badgeLabel: string;
  subtitle: string | null;
  ownerMeta: string;
};

type GovernedRevisionHistoryItem = {
  document_id: string;
  revision_number: number;
  revision_title: string;
  status: string;
  created_at: string;
  is_current: boolean;
};

function getActiveSiblingCtaLabel(state: ActiveSiblingState): string {
  if (state === 'draft') return 'Continuar rascunho';
  if (state === 'under_review') return 'Acompanhar revis\u00e3o';
  if (state === 'approved') return 'Publicar revis\u00e3o aprovada';
  if (state === 'scheduled') return 'Ver publica\u00e7\u00e3o agendada';
  return 'Retomar revis\u00e3o rejeitada';
}

function getActiveSiblingDestination(documentId: string, state: ActiveSiblingState): string {
  if (state === 'draft' || state === 'rejected') {
    return `/documents/${documentId}/edit`;
  }
  return `/documents/${documentId}`;
}

function getDocumentStatusPresentation(status: string, publishedAt: string): DocumentStatusPresentation {
  const hasPublishedAt = publishedAt !== '—';

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
      return {
        badgeLabel: 'rascunho',
        subtitle: 'Rascunho — ainda não publicado',
        ownerMeta: 'rascunho',
      };
    case 'under_review':
      return {
        badgeLabel: 'em revisão',
        subtitle: 'Aguardando decisão de aprovação',
        ownerMeta: 'em revisão',
      };
    case 'rejected':
      return {
        badgeLabel: 'rejeitado',
        subtitle: 'Revisão rejeitada',
        ownerMeta: 'rejeitado',
      };
    default:
      return {
        badgeLabel: status || 'sem status',
        subtitle: hasPublishedAt ? `publicado em ${publishedAt}` : null,
        ownerMeta: hasPublishedAt ? `publicado em ${publishedAt}` : (status || '—'),
      };
  }
}
export function DocumentPublishedPage() {
  const { documentId: rawDocumentId } = useParams<{ documentId: string }>();
  const navigate = useNavigate();
  const user = useAuthStore((s) => s.user);
  // Obsolete docs stay viewable only for users who manage obsoletion (UX hint;
  // backend document.view is the real boundary — wiki/concepts/authz-tiers.md).
  const canViewObsolete = useHasCapability('document.obsolete');

  // Route declaration `/documents/:documentId` makes this non-empty in practice.
  // Coerce to string and let the early return below handle the unreachable empty case
  // without breaking the rules of hooks.
  const documentId = rawDocumentId ?? '';

  const scheduledLifecycleRefetchInterval = 5_000;
  const docQuery = useDocumentDetailQuery(documentId, { pollScheduledLifecycle: true });
  const shouldPollScheduledLifecycle = docQuery.data?.status === 'scheduled';
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
  const distributionSummaryQuery = useDistributionSummaryQuery(documentId);
  const activeDocument = activeDocumentQuery.data ?? null;

  const [linkCopied, setLinkCopied] = useState(false);
  const [showRevisionForm, setShowRevisionForm] = useState(false);
  const [revisionName, setRevisionName] = useState('');
  const [revisionError, setRevisionError] = useState('');
  const [isCreatingRevision, setIsCreatingRevision] = useState(false);
  const [showPublishDialog, setShowPublishDialog] = useState(false);
  // PDF export: mirrors ExportMenu.handlePDF (POST → open signed_url; 429 = rate-limited).
  const [pdfStatus, setPdfStatus] = useState<
    | { kind: 'idle' }
    | { kind: 'pending' }
    | { kind: 'rate_limited'; retryAfterSec: number }
    | { kind: 'error'; message: string }
  >({ kind: 'idle' });

  const handleDownloadPDF = async () => {
    setPdfStatus({ kind: 'pending' });
    try {
      const res = await exportPDF(documentId, { paper_size: 'A4' });
      window.open(res.signed_url, '_blank', 'noopener');
      setPdfStatus({ kind: 'idle' });
    } catch (e: unknown) {
      // 429 is rate-limited (retry_after_seconds is not in the Problem schema → fallback).
      // Every other failure surfaces an inline alert; never swallowed silently.
      const err = e instanceof ApiError ? e : null;
      if (err?.status === 429) {
        setPdfStatus({ kind: 'rate_limited', retryAfterSec: 60 });
        return;
      }
      setPdfStatus({ kind: 'error', message: resolveErrorMessage(e) });
    }
  };

  // ── Loading state ──────────────────────────────────────────────────────────
  if (docQuery.isLoading) {
    return (
      <div className={styles.stateLoading} role="status" aria-live="polite">
        <Icon name="docs" size={24} className={styles.stateIcon} />
        <span>Carregando documento…</span>
      </div>
    );
  }

  // ── Error state ────────────────────────────────────────────────────────────
  if (docQuery.isError || !docQuery.data) {
    return (
      <div className={styles.stateError} role="alert">
        <Icon name="x" size={20} className={styles.stateIcon} />
        <span>Documento não encontrado ou sem permissão de acesso.</span>
        <button className="btn btn-sm" type="button" onClick={() => docQuery.refetch()}>
          Tentar novamente
        </button>
      </div>
    );
  }

  const doc = docQuery.data;
  const approval = approvalQuery.data ?? null;

  // Normalize response casing (API returns both PascalCase and snake_case)
  const code            = doc.code ?? '—';
  const docName         = doc.name;
  const status          = doc.status ?? '';
  const createdByRaw    = doc.created_by ?? '—';
  // TODO(backlog): API should return created_by_display_name snapshot — wiki/backlog/documento-publicado.md
  // Fallback: created_by stores user_id; if creator is the current user, use displayName
  const createdBy = (user?.userId && createdByRaw === user.userId)
    ? (user.displayName ?? createdByRaw)
    : createdByRaw;
  const versionLabel = formatRevisionCode(doc.revision_number);
  const areaCode = doc.process_area_code_snapshot ?? '';
  const profileCode = doc.profile_code_snapshot ?? '';
  const areaLabel = areaCode ? resolveAreaLabel(areaCode, areasQuery.data ?? []) : '—';
  const profileLabel = profileCode ? resolveProfileLabel(profileCode, profilesQuery.data ?? []) : '—';

  // Coverage denominator (obligated audience) — M2 distribution summary. Numerator
  // (% read) is parked (ADR-0042); show the count + parked label, never a fabricated %.
  // EM_DASH on error/missing so we never render a fabricated 0.
  const obligatedCount = distributionSummaryQuery.isError
    ? '—'
    : (distributionSummaryQuery.data?.total_targets ?? '—');
  // File metadata — already on DocumentResponse (nullable → honest em-dash when absent).
  const pageCountLabel = formatPageCount(doc.current_revision_page_count);
  const fileSizeLabel = formatFileSize(doc.current_revision_file_size_bytes);

  // Approval completion is the only lifecycle timestamp currently exposed on
  // the canonical document detail payload.
  const publishedAt = formatPublishedAt(approval?.completed_at ?? undefined);
  const sinceDateHint = approval?.completed_at ? formatShortDate(approval.completed_at) : '—';
  const isApproved = status === 'approved';
  const isPublished = status === 'published';
  const statusPresentation = getDocumentStatusPresentation(status, publishedAt);
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
  const activeSiblingCtaLabel = activeSiblingState ? getActiveSiblingCtaLabel(activeSiblingState) : 'Iniciar revis\u00e3o';
  const activeSiblingDestination =
    activeSiblingDocumentId && activeSiblingState
      ? getActiveSiblingDestination(activeSiblingDocumentId, activeSiblingState)
      : null;
  // ObsoleteBanner
  const isObsolete = status === 'obsolete';

  // RBAC: roles that can initiate a revision
  // TODO(backlog): refine with controlled_document owner_id once API exposes it
  // See wiki/backlog/documento-publicado.md
  const canInitiateRevision =
    user != null &&
    user.roles.some((r) => ['system_admin', 'editor', 'qms_admin', 'area_admin'].includes(r));
  const canCreateRevision =
    canInitiateRevision &&
    isPublished &&
    Boolean(doc.controlled_document_id) &&
    activeSiblingDocumentId == null;
  const canPublish = canInitiateRevision && isApproved && Boolean(activeDocument?.content_hash);
  const publishContextNotice =
    isApproved && !canInitiateRevision
      ? 'Seu perfil atual não pode publicar esta revisão.'
      : isApproved && activeDocumentQuery.isError
        ? 'Não foi possível confirmar o contexto ativo de publicação. Atualize a página e tente novamente antes de publicar.'
        : isApproved && !activeDocument?.content_hash
          ? 'A publicação está bloqueada porque o contexto ativo desta revisão ainda não foi confirmado.'
          : null;

  // Signoff stages from approval instance
  const signoffStages = approval?.stages ?? [];
  const stageCount = signoffStages.length;
  // Connector spans from first pin center to last pin center
  const connectorSide = stageCount > 1 ? `${(100 / (2 * stageCount)).toFixed(2)}%` : '50%';
  const revisionHistoryItems = (revisionHistoryQuery.data?.items ?? []) as GovernedRevisionHistoryItem[];
  const revisionTimeline = revisionHistoryItems.map((item) => ({
    v: `REV${String(item.revision_number ?? 0).padStart(2, '0')}`,
    when: formatShortDate(item.created_at) || '—',
    author: item.status || '—',
    current: Boolean(item.is_current),
    summary: item.revision_title?.trim() || 'Sem título governado registrado.',
  }));

  const currentPublishedHistoryItem =
    status === 'scheduled' && activeDocument?.published_document_id
      ? revisionHistoryItems.find((item) => item.document_id === activeDocument.published_document_id) ?? null
      : null;
  const currentVersionLabel = currentPublishedHistoryItem
    ? formatRevisionCode(currentPublishedHistoryItem.revision_number)
    : versionLabel;
  const currentVersionHint = currentPublishedHistoryItem
    ? '—'
    : sinceDateHint;

  // Handlers
  const handleView = () => navigate(`/documents/${documentId}/edit`);

  const handleCopyLink = () => {
    navigator.clipboard.writeText(window.location.href)
      .then(() => {
        setLinkCopied(true);
        setTimeout(() => setLinkCopied(false), 2000);
      })
      .catch(() => {/* silent — no false-positive UI */});
  };

  const handleStartRevision = () => {
    if (!canCreateRevision || isCreatingRevision) {
      return;
    }
    setRevisionError('');
    setRevisionName(docName ?? code);
    setShowRevisionForm(true);
  };

  const handleContinueActiveRevision = () => {
    if (!activeSiblingDestination) {
      return;
    }
    navigate(activeSiblingDestination);
  };

  const handleCancelRevision = () => {
    if (isCreatingRevision) {
      return;
    }
    setShowRevisionForm(false);
    setRevisionName('');
    setRevisionError('');
  };

  const handlePublishSuccess = () => {
    void docQuery.refetch();
    void approvalQuery.refetch();
    void activeDocumentQuery.refetch();
  };

  const handleCreateRevision = async () => {
    if (!doc.controlled_document_id || !revisionName.trim()) {
      return;
    }
    setIsCreatingRevision(true);
    setRevisionError('');
    try {
      const response = await createRevision(
        doc.controlled_document_id,
        { name: revisionName.trim(), form_data: {}, template_version_id: doc.template_version_id },
        crypto.randomUUID(),
      );
      navigate(`/documents/${response.document.id}/edit`);
    } catch (err) {
      setRevisionError(resolveQueryError(err, 'Falha ao iniciar nova revisão.'));
    } finally {
      setIsCreatingRevision(false);
    }
  };

  return (
    <div className={`${styles.root}${isObsolete ? ` ${styles.rootObsolete}` : ''}`}>
      {/* ObsoleteBanner — only rendered when status is obsolete */}
      {isObsolete && (
        <div className={styles.obsoleteBanner} role="alert" aria-label="Documento obsoleto">
          <div className={styles.obsoleteStamp}>OBSOLETO</div>
        </div>
      )}

      {/* Hero */}
      <DocumentHero
        breadcrumbItems={[
          { label: 'Biblioteca', href: '/documents' },
          ...(areaLabel !== '—' ? [{ label: areaLabel }] : []),
          { label: code },
        ]}
        docCard={
          <div className={styles.docCard}>
            <div className={styles.docCardHeader}>{areaLabel}</div>
            <div className={styles.docCardBody}>
              <div className={styles.docCardCode}>{code}</div>
              <div className={styles.docCardType}>{profileLabel}</div>
              <div className={styles.docCardSpacer} />
              <div className={styles.docCardDivider} />
              <div className={styles.docCardFooter}>
                <span className={styles.docCardVersion}>{versionLabel}</span>
                <span className={styles.docCardDot} />
              </div>
            </div>
          </div>
        }
        badges={
          <>
            <CodeChip className={styles.codeChip}>{code}</CodeChip>
            {!isObsolete && (
              <span className={styles.vigenteBadge}>
                <span className={styles.vigenteDot} />
                {versionLabel} · {statusPresentation.badgeLabel}
              </span>
            )}
            <span className={styles.typeLabel}>{profileLabel}</span>
          </>
        }
        title={docName ?? code}
        subtitle={statusPresentation.subtitle ? <span>{statusPresentation.subtitle}</span> : null}
        actions={
          <div className={styles.heroActions}>
            <button
              className="btn btn-primary btn-lg"
              type="button"
              onClick={handleView}
              disabled={isObsolete && !canViewObsolete}
              title={
                isObsolete && !canViewObsolete
                  ? 'Sua sessão não inclui a capacidade para visualizar documentos obsoletos.'
                  : undefined
              }
            >
              <Icon name="eye" size={15} />
              Visualizar documento
            </button>
            <button
              className="btn"
              type="button"
              aria-label="Baixar PDF"
              onClick={handleDownloadPDF}
              disabled={isObsolete || pdfStatus.kind === 'pending' || pdfStatus.kind === 'rate_limited'}
              title={
                pdfStatus.kind === 'rate_limited'
                  ? `Limite de exportações atingido — tente novamente em ${pdfStatus.retryAfterSec}s`
                  : 'Baixar PDF'
              }
            >
              <Icon name="download" size={13} />
              {pdfStatus.kind === 'pending'
                ? 'Gerando PDF…'
                : pdfStatus.kind === 'rate_limited'
                  ? 'Aguarde…'
                  : 'Baixar PDF'}
            </button>
            {pdfStatus.kind === 'rate_limited' && (
              <span role="alert" className={styles.pdfAlert}>
                Limite de exportações atingido — tente novamente em {pdfStatus.retryAfterSec}s.
              </span>
            )}
            {pdfStatus.kind === 'error' && (
              <span role="alert" className={styles.pdfAlert}>{pdfStatus.message}</span>
            )}
            {isApproved ? (
              <button
                className="btn"
                type="button"
                aria-disabled={!canPublish}
                title={
                  !canInitiateRevision
                    ? 'Sem permissÃ£o para publicar'
                    : !activeDocument?.content_hash
                      ? 'Aguardando contexto ativo para publicar'
                      : undefined
                }
                onClick={() => {
                  if (canPublish) {
                    setShowPublishDialog(true);
                  }
                }}
              >
                <Icon name="check" size={13} />
                Publicar / Agendar
              </button>
            ) : (
            <button
              className="btn"
              type="button"
              aria-disabled={!canCreateRevision && !activeSiblingDocumentId}
              title={
                !canInitiateRevision
                  ? 'Sem permissão para iniciar revisão'
                  : activeSiblingDocumentId
                    ? 'Já existe uma revisão ativa para este documento controlado'
                  : !isPublished
                    ? 'Apenas documentos publicados podem iniciar uma nova revisão'
                    : undefined
              }
              onClick={activeSiblingDocumentId ? handleContinueActiveRevision : handleStartRevision}
            >
              <Icon name="edit" size={13} />
              {activeSiblingCtaLabel}
            </button>
            )}
            <button className="btn btn-ghost" type="button" onClick={handleCopyLink} disabled={isObsolete}>
              <Icon name={linkCopied ? 'check' : 'link'} size={13} />
              {linkCopied ? 'Link copiado!' : 'Copiar link'}
            </button>
            {publishContextNotice ? (
              <div className={styles.heroNotice} role="status" aria-live="polite">
                <Icon name="shield" size={14} />
                <span>{publishContextNotice}</span>
              </div>
            ) : null}
          </div>
        }
      />

      {showRevisionForm && (
        <div className={styles.revisionComposer} role="region" aria-label="Iniciar nova revisão">
          <div className={styles.revisionComposerHeader}>
            <div>
              <div className={styles.revisionComposerKicker}>Nova revisão</div>
              <h2 className={styles.revisionComposerTitle}>Criar novo draft de revisão</h2>
            </div>
            <CodeChip>{code}</CodeChip>
          </div>
          <p className={styles.revisionComposerBody}>
            Defina o nome de trabalho do novo draft. O título governado da revisão será confirmado depois, no envio formal para aprovação.
          </p>
          <label className={styles.revisionComposerField} htmlFor="published-revision-name">
            Nome do documento
          </label>
          <input
            id="published-revision-name"
            className={`input ${styles.revisionComposerInput}`}
            type="text"
            value={revisionName}
            onChange={(event) => setRevisionName(event.target.value)}
            placeholder={docName ?? code}
            disabled={isCreatingRevision}
          />
          {revisionError && (
            <p className={styles.revisionComposerError} role="alert">
              {revisionError}
            </p>
          )}
          <div className={styles.revisionComposerActions}>
            <button
              className="btn btn-primary"
              type="button"
              onClick={() => void handleCreateRevision()}
              disabled={isCreatingRevision || revisionName.trim().length === 0}
            >
              {isCreatingRevision ? 'Criando...' : 'Gerar documento'}
            </button>
            <button className="btn" type="button" onClick={handleCancelRevision} disabled={isCreatingRevision}>
              Cancelar
            </button>
          </div>
        </div>
      )}

      {showPublishDialog && activeDocument?.content_hash ? (
        <SupersedePublishDialog
          documentId={documentId}
          contentHash={activeDocument.content_hash}
          revisionVersion={activeDocument.revision_version}
          publishedDocumentId={activeDocument.published_document_id ?? undefined}
          onClose={() => setShowPublishDialog(false)}
          onSuccess={handlePublishSuccess}
        />
      ) : null}

      {/* Content area */}
      <div className={styles.content}>

        {/* KPI strip — 4 cells. Cobertura + Páginas are wired to real data (M2 distribution
            denominator + DocumentResponse file metadata). Próxima revisão is a tracked defer
            (no backend field) — see wiki/backlog/documento-publicado.md. */}
        <div className={styles.kpiStrip}>
          <div className={styles.kpiCell}>
            <div className={styles.kpiLabel}>Versão atual</div>
            <div className={styles.kpiValue}>{currentVersionLabel}</div>
            <div className={styles.kpiHint}>
              {currentVersionHint !== '—'
                ? `desde ${currentVersionHint}`
                : statusPresentation.ownerMeta}
            </div>
          </div>
          <button
            type="button"
            className={`${styles.kpiCell} ${styles.kpiCellLink}`}
            onClick={() => navigate('/distribution')}
          >
            <div className={styles.kpiLabel}>Cobertura</div>
            <div className={styles.kpiValue}>{obligatedCount}</div>
            <div className={styles.kpiHint}>destinatários obrigados · abrir fanout →</div>
          </button>
          <div className={styles.kpiCell}>
            <div className={styles.kpiLabel}>Próxima revisão</div>
            <div className={styles.kpiValue}>—</div>
            <div className={styles.kpiHint}>sem data de revisão definida</div>
          </div>
          <div className={styles.kpiCell}>
            <div className={styles.kpiLabel}>Páginas</div>
            <div className={styles.kpiValue}>{pageCountLabel}</div>
            <div className={styles.kpiHint}>metadado do arquivo</div>
          </div>
        </div>

        {/* Section: Sobre */}
        <section className={styles.section}>
          <div className={styles.sectionHead}>
            <div>
              <div className={styles.sectionKicker}>01 · Sobre</div>
              <h2 className={styles.sectionTitle}>Identificação e responsabilidade</h2>
            </div>
          </div>
          <div className={styles.aboutLayout}>
            <div className={styles.aboutCard}>
              <div className={styles.ownerBanner}>
                <Avatar name={createdBy} size="sm" />
                <div className={styles.ownerInfo}>
                  <div className={styles.ownerName}>{createdBy}</div>
                  <div className={styles.ownerMeta}>{statusPresentation.ownerMeta}</div>
                </div>
              </div>
              <div className={styles.factsGrid}>
                <div className={styles.factCell}>
                  <div className={styles.factIcon}>
                    <Icon name="docs" size={14} />
                  </div>
                  <div className={styles.factContent}>
                    <div className={styles.factLabel}>Tipo</div>
                    <div className={styles.factValue}>{profileLabel}</div>
                  </div>
                </div>
                <div className={styles.factCell}>
                  <div className={styles.factIcon}>
                    <Icon name="taxonomy" size={14} />
                  </div>
                  <div className={styles.factContent}>
                    <div className={styles.factLabel}>Área</div>
                    <div className={styles.factValue}>{areaLabel}</div>
                  </div>
                </div>
                <div className={styles.factCell}>
                  <div className={styles.factIcon}>
                    <Icon name="calendar" size={14} />
                  </div>
                  <div className={styles.factContent}>
                    <div className={styles.factLabel}>Vigente desde</div>
                    <div className={styles.factValue}>{sinceDateHint}</div>
                  </div>
                </div>
                <div className={styles.factCell}>
                  <div className={styles.factIcon}>
                    <Icon name="calendar" size={14} />
                  </div>
                  <div className={styles.factContent}>
                    <div className={styles.factLabel}>Próxima revisão</div>
                    <div className={styles.factValue}>—</div>
                  </div>
                </div>
                <div className={styles.factCell}>
                  <div className={styles.factIcon}>
                    <Icon name="docs" size={14} />
                  </div>
                  <div className={styles.factContent}>
                    <div className={styles.factLabel}>Tamanho</div>
                    <div className={styles.factValue}>{fileSizeLabel}</div>
                  </div>
                </div>
                {/* DEFER: confidentiality classification — no field on DocumentResponse yet.
                    Tracked in wiki/backlog/documento-publicado.md (trigger: classification field). */}
                <div className={styles.factCell}>
                  <div className={styles.factIcon}>
                    <Icon name="shield" size={14} />
                  </div>
                  <div className={styles.factContent}>
                    <div className={styles.factLabel}>Classificação</div>
                    <div className={styles.factValue}>—</div>
                  </div>
                </div>
              </div>
            </div>

            {/* Coverage side card — obligated-audience denominator (M2 distribution summary).
                The read numerator (% lido) is parked per ADR-0042; show the count + an explicit
                parked label, never a fabricated %. See wiki/backlog/documento-publicado.md. */}
            <aside className={styles.coverageCard}>
              <div className={styles.coverageCardHeader}>
                <div className={styles.kpiLabel}>Cobertura</div>
              </div>
              <div className={styles.coverageCardBody}>
                <div className={styles.kpiValue}>{obligatedCount}</div>
                <div className={styles.kpiHint}>destinatários obrigados</div>
                <div className={styles.kpiHint}>leitura em acompanhamento (ADR-0042)</div>
              </div>
              <div className={styles.coverageCardFooter}>
                <button
                  type="button"
                  className={`btn btn-sm ${styles.coverageCardAction}`}
                  onClick={() => navigate('/distribution')}
                >
                  Abrir Fanout
                  <Icon name="chevron-right" size={14} />
                </button>
              </div>
            </aside>
          </div>
        </section>

        {/* Section: Cadeia de aprovação */}
        <section className={styles.section}>
          <div className={styles.sectionHead}>
            <div>
              <div className={styles.sectionKicker}>02 · Cadeia de aprovação</div>
              <h2 className={styles.sectionTitle}>Sign-offs desta versão</h2>
            </div>
          </div>

          {approvalQuery.isLoading && (
            <div className={styles.signoffLoading} role="status" aria-live="polite">
              Carregando cadeia de aprovação…
            </div>
          )}

          {!approvalQuery.isLoading && signoffStages.length === 0 && (
            <div className={styles.signoffEmpty}>
              Nenhum registro de aprovação para esta versão.
            </div>
          )}

          {signoffStages.length > 0 && (
            <div className={styles.signoffCard}>
              <div
                className={styles.signoffGrid}
                style={{ gridTemplateColumns: `repeat(${stageCount}, 1fr)` }}
              >
                <div
                  className={styles.signoffConnector}
                  style={{ left: connectorSide, right: connectorSide }}
                />
                {signoffStages.map((stage) => {
                  // Show first signoff per stage (one actor per stage in current workflow)
                  const signoff = stage.signoffs[0] ?? null;
                  return (
                    <div key={stage.id} className={styles.signoffStage}>
                      <div className={styles.signoffPin}>
                        <Icon name="check" size={16} />
                      </div>
                      <div className={styles.signoffStageName}>{stage.label}</div>
                      {signoff && (
                        <>
                          <div className={styles.signoffActor}>
                            <Avatar name={signoff.actor_user_id} size="sm" />
                            <span className={styles.signoffActorName}>{signoff.actor_user_id}</span>
                          </div>
                          <div className={styles.signoffWhen}>{formatSignedAt(signoff.signed_at)}</div>
                        </>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </section>

        {/* Section: Histórico de versões */}
        <section className={styles.section}>
          <div className={styles.sectionHead}>
            <div>
              <div className={styles.sectionKicker}>03 · Linhagem</div>
              <h2 className={styles.sectionTitle}>Histórico de versões</h2>
            </div>
            {revisionTimeline.length > 0 ? (
              <span className={styles.sectionAside}>passe o mouse para detalhar</span>
            ) : null}
          </div>
          {revisionHistoryQuery.isLoading ? (
            <div className={styles.signoffLoading} role="status" aria-live="polite">
              Carregando linhagem governada…
            </div>
          ) : revisionTimeline.length > 0 ? (
            <DocumentVersionTimeline versions={revisionTimeline} />
          ) : (
            <div className={styles.signoffEmpty}>Nenhuma revisão governada encontrada.</div>
          )}
        </section>

        {/* Section: Documentos relacionados — backend não disponível, defer rastreado em wiki/backlog/documento-publicado.md */}
        <section className={styles.section}>
          <div className={styles.sectionHead}>
            <div>
              <div className={styles.sectionKicker}>04 · Referências</div>
              <h2 className={styles.sectionTitle}>Documentos relacionados</h2>
            </div>
            <span className={styles.sectionAside}>não disponível</span>
          </div>
          <div className={styles.signoffEmpty}>
            O modelo de relacionamentos entre documentos ainda não está disponível.
          </div>
        </section>

        {/* Section: Comentários — backend não disponível, defer rastreado em wiki/backlog/documento-publicado.md */}
        <section className={styles.section}>
          <div className={styles.sectionHead}>
            <div>
              <div className={styles.sectionKicker}>05 · Discussão interna</div>
              <h2 className={styles.sectionTitle}>Comentários</h2>
            </div>
            <span className={styles.sectionAside}>não disponível</span>
          </div>
          <div className={styles.signoffEmpty}>
            Comentários de exibição ainda não estão disponíveis para este documento.
          </div>
        </section>

      </div>
    </div>
  );
}

