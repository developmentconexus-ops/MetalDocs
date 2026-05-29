import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Icon } from '../../../components/ui/Icon';
import { Avatar } from '../../../components/ui/Avatar';
import { CodeChip } from '../../../components/ui/CodeChip';
import { resolveQueryError } from '../../../lib/api';
import { DocumentHero } from '../components/DocumentHero';
import { DocumentVersionTimeline } from '../components/DocumentVersionTimeline';
import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
import { useApprovalInstanceQuery } from '../queries/useApprovalInstanceQuery';
import { useControlledDocumentActiveDocumentQuery } from '../queries/useControlledDocumentActiveDocumentQuery';
import { useDocumentRevisionHistoryQuery } from '../queries/useDocumentRevisionHistoryQuery';
import { createRevision } from '../../controlled-documents/api/controlledDocuments';
import { SupersedePublishDialog } from '../../approval/components/SupersedePublishDialog';
import { useAuthStore } from '../../../store/auth.store';
import { formatPublishedAt, formatRevisionCode, formatSignedAt, formatShortDate } from '../lib/documentDetailMeta';
import styles from './DocumentPublishedPage.module.css';

const ACTIVE_SIBLING_STATES = ['draft', 'under_review', 'approved', 'scheduled', 'rejected'] as const;
type ActiveSiblingState = (typeof ACTIVE_SIBLING_STATES)[number];

type DocumentStatusPresentation = {
  badgeLabel: string;
  subtitle: string | null;
  ownerMeta: string;
};

type GovernedRevisionHistoryItem = {
  documentId: string;
  revisionNumber: number;
  revisionTitle: string;
  status: string;
  createdAt: string;
  isCurrent: boolean;
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
// TODO(backlog): wire relationship model — GET /api/v1/documents/:id/relationships
// See wiki/backlog/documento-publicado.md
const PLACEHOLDER_RELATED = [
  { code: 'IT-EHS-021', title: 'Inspeção de cadeados e travas', type: 'Instrução', rel: 'referenciado por' },
  { code: 'FR-EHS-008', title: 'Ficha de bloqueio individual', type: 'Formulário', rel: 'anexo obrigatório' },
  { code: 'PR-MAN-103', title: 'Manutenção preventiva CCM-04', type: 'Procedimento', rel: 'invoca este PR' },
];

// TODO(backlog): wire GET /api/v1/documents/:id/display-comments — architecture brainstorm needed
// See wiki/backlog/documento-publicado.md
const PLACEHOLDER_COMMENTS = [
  { who: 'Marcos Lima', role: 'Gerente Industrial', when: '13 mar · 09:42',
    text: 'Equipes de Pomerode já alinhadas. Treinamento iniciado para CCM-04 na próxima segunda.' },
  { who: 'Renata Souza', role: 'Coord. SSMA', when: '14 mar · 11:08',
    text: 'Adicionei FR-EHS-008 como anexo de referência para os supervisores de turno.' },
];

export function DocumentPublishedPage() {
  const { documentId: rawDocumentId } = useParams<{ documentId: string }>();
  const navigate = useNavigate();
  const user = useAuthStore((s) => s.user);

  // Route declaration `/documents/:documentId` makes this non-empty in practice.
  // Coerce to string and let the early return below handle the unreachable empty case
  // without breaking the rules of hooks.
  const documentId = rawDocumentId ?? '';

  const scheduledLifecycleRefetchInterval = 5_000;
  const docQuery = useDocumentDetailQuery(documentId, { pollScheduledLifecycle: true });
  const shouldPollScheduledLifecycle = docQuery.data?.Status === 'scheduled';
  const approvalQuery = useApprovalInstanceQuery(documentId);
  const controlledDocumentId = docQuery.data?.ControlledDocumentID ?? null;
  const activeDocumentQuery = useControlledDocumentActiveDocumentQuery(controlledDocumentId, {
    refetchInterval: shouldPollScheduledLifecycle ? scheduledLifecycleRefetchInterval : false,
  });
  const revisionHistoryQuery = useDocumentRevisionHistoryQuery(documentId, {
    refetchInterval: shouldPollScheduledLifecycle ? scheduledLifecycleRefetchInterval : false,
  });
  const activeDocument = activeDocumentQuery.data ?? null;

  const [linkCopied, setLinkCopied] = useState(false);
  const [showRevisionForm, setShowRevisionForm] = useState(false);
  const [revisionName, setRevisionName] = useState('');
  const [revisionError, setRevisionError] = useState('');
  const [isCreatingRevision, setIsCreatingRevision] = useState(false);
  const [showPublishDialog, setShowPublishDialog] = useState(false);

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
  const code            = doc.Code ?? '—';
  const docName         = doc.Name;
  const status          = doc.Status ?? '';
  const createdByRaw    = doc.CreatedBy ?? '—';
  // TODO(backlog): API should return created_by_display_name snapshot — wiki/backlog/documento-publicado.md
  // Fallback: created_by stores user_id; if creator is the current user, use displayName
  const createdBy = (user?.userId && createdByRaw === user.userId)
    ? (user.displayName ?? createdByRaw)
    : createdByRaw;
  const versionLabel = formatRevisionCode(doc.RevisionNumber);

  // Approval completion is the only lifecycle timestamp currently exposed on
  // the canonical document detail payload.
  const publishedAt = formatPublishedAt(approval?.completed_at ?? undefined);
  const sinceDateHint = approval?.completed_at ? formatShortDate(approval.completed_at) : '—';
  const isApproved = status === 'approved';
  const isPublished = status === 'published';
  const statusPresentation = getDocumentStatusPresentation(status, publishedAt);
  const activeSiblingDocumentId =
    activeDocument?.documentId &&
    activeDocument.documentId !== documentId &&
    ACTIVE_SIBLING_STATES.includes(activeDocument.approvalState as ActiveSiblingState)
      ? activeDocument.documentId
      : null;
  const activeSiblingStateCandidate = activeDocument?.approvalState as ActiveSiblingState | undefined;
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
    user.roles.some((r) => ['admin', 'system_admin', 'editor', 'qms_admin', 'area_admin'].includes(r));
  const canCreateRevision =
    canInitiateRevision &&
    isPublished &&
    Boolean(doc.ControlledDocumentID) &&
    activeSiblingDocumentId == null;
  const canPublish = canInitiateRevision && isApproved && Boolean(activeDocument?.contentHash);
  const publishContextNotice =
    isApproved && !canInitiateRevision
      ? 'Seu perfil atual não pode publicar esta revisão.'
      : isApproved && activeDocumentQuery.isError
        ? 'Não foi possível confirmar o contexto ativo de publicação. Atualize a página e tente novamente antes de publicar.'
        : isApproved && !activeDocument?.contentHash
          ? 'A publicação está bloqueada porque o contexto ativo desta revisão ainda não foi confirmado.'
          : null;

  // Signoff stages from approval instance
  const signoffStages = approval?.stages ?? [];
  const stageCount = signoffStages.length;
  // Connector spans from first pin center to last pin center
  const connectorSide = stageCount > 1 ? `${(100 / (2 * stageCount)).toFixed(2)}%` : '50%';
  const revisionHistoryItems = (revisionHistoryQuery.data?.items ?? []) as GovernedRevisionHistoryItem[];
  const revisionTimeline = revisionHistoryItems.map((item) => ({
    v: `REV${String(item.revisionNumber ?? 0).padStart(2, '0')}`,
    when: formatShortDate(item.createdAt) || '—',
    author: item.status || '—',
    current: Boolean(item.isCurrent),
    summary: item.revisionTitle?.trim() || 'Sem título governado registrado.',
  }));

  const currentPublishedHistoryItem =
    status === 'scheduled' && activeDocument?.publishedDocumentId
      ? revisionHistoryItems.find((item) => item.documentId === activeDocument.publishedDocumentId) ?? null
      : null;
  const currentVersionLabel = currentPublishedHistoryItem
    ? formatRevisionCode(currentPublishedHistoryItem.revisionNumber)
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
    if (!doc.ControlledDocumentID || !revisionName.trim()) {
      return;
    }
    setIsCreatingRevision(true);
    setRevisionError('');
    try {
      const response = await createRevision(
        doc.ControlledDocumentID,
        { name: revisionName.trim(), formData: {}, templateVersionId: doc.TemplateVersionID },
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
    <div className={styles.root}>
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
          { label: '—' },
          { label: code },
        ]}
        docCard={
          <div className={styles.docCard}>
            {/* TODO(backlog): show area label once DocumentResponse includes area_code */}
            <div className={styles.docCardHeader}>—</div>
            <div className={styles.docCardBody}>
              <div className={styles.docCardCode}>{code}</div>
              {/* TODO(backlog): show profile label once DocumentResponse includes profile_code */}
              <div className={styles.docCardType}>—</div>
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
            <span className={styles.vigenteBadge}>
              <span className={styles.vigenteDot} />
              {versionLabel} · {statusPresentation.badgeLabel}
            </span>
            {/* TODO(backlog): resolve profile_code → human label once DocumentResponse includes profile_code */}
            <span className={styles.typeLabel}>—</span>
          </>
        }
        title={docName ?? code}
        subtitle={statusPresentation.subtitle ? <span>{statusPresentation.subtitle}</span> : null}
        actions={
          <div className={styles.heroActions}>
            <button className="btn btn-primary btn-lg" type="button" onClick={handleView}>
              <Icon name="eye" size={15} />
              Visualizar documento
            </button>
            {/* TODO(backlog): wire PDF download — GET /api/v1/documents/:id/pdf */}
            <button className="btn" type="button" aria-disabled="true" title="Em breve">
              <Icon name="download" size={13} />
              Baixar PDF
            </button>
            {isApproved ? (
              <button
                className="btn"
                type="button"
                aria-disabled={!canPublish}
                title={
                  !canInitiateRevision
                    ? 'Sem permissÃ£o para publicar'
                    : !activeDocument?.contentHash
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
            <button className="btn btn-ghost" type="button" onClick={handleCopyLink}>
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

      {showPublishDialog && activeDocument?.contentHash ? (
        <SupersedePublishDialog
          documentId={documentId}
          contentHash={activeDocument.contentHash}
          revisionVersion={activeDocument.revisionVersion}
          publishedDocumentId={activeDocument.publishedDocumentId ?? undefined}
          onClose={() => setShowPublishDialog(false)}
          onSuccess={handlePublishSuccess}
        />
      ) : null}

      {/* Content area */}
      <div className={styles.content}>

        {/* KPI strip */}
        <div className={styles.kpiStrip}>
          <div className={styles.kpiCell}>
            <div className={styles.kpiLabel}>Versão atual</div>
            <div className={styles.kpiValue}>{currentVersionLabel}</div>
            <div className={styles.kpiHint}>
              {currentVersionHint !== '—' ? `desde ${currentVersionHint}` : '—'}
            </div>
          </div>
          {/* TODO(backlog): wire fanout coverage API */}
          <div className={styles.kpiCell}>
            <div className={styles.kpiLabel}>Cobertura</div>
            <div className={styles.kpiValue}>—</div>
            <div className={styles.kpiHint}>em breve</div>
          </div>
          {/* "Próxima revisão" and "Páginas" KPIs intentionally omitted — CUT in NOTES.md (no review-date field, no page count). */}
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
                    {/* TODO(backlog): resolve from profile_code once DocumentResponse includes it */}
                    <div className={styles.factValue}>—</div>
                  </div>
                </div>
                <div className={styles.factCell}>
                  <div className={styles.factIcon}>
                    <Icon name="taxonomy" size={14} />
                  </div>
                  <div className={styles.factContent}>
                    <div className={styles.factLabel}>Área</div>
                    {/* TODO(backlog): resolve from area_code once DocumentResponse includes it */}
                    <div className={styles.factValue}>—</div>
                  </div>
                </div>
              </div>
            </div>
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

        {/* Section: Documentos relacionados — TODO(backlog): wire relationship model */}
        <section className={styles.section}>
          <div className={styles.sectionHead}>
            <div>
              <div className={styles.sectionKicker}>04 · Referências</div>
              <h2 className={styles.sectionTitle}>Documentos relacionados</h2>
            </div>
          </div>
          <div className={styles.relatedGrid}>
            {PLACEHOLDER_RELATED.map((r) => (
              <div key={r.code} className={styles.relatedCard}>
                <div className={styles.relatedCardTop}>
                  <div className={styles.relatedCardIcon}>
                    <Icon name="docs" size={14} />
                  </div>
                  <span className={styles.relatedCardRel}>{r.rel}</span>
                </div>
                <div className={styles.relatedCardTitle}>{r.title}</div>
                <div className={styles.relatedCardFooter}>
                  <span className={styles.relatedCardCode}>{r.code}</span>
                  <span className={styles.relatedCardType}>{r.type}</span>
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* Section: Comentários — TODO(backlog): wire display-side comments */}
        <section className={styles.section}>
          <div className={styles.sectionHead}>
            <div>
              <div className={styles.sectionKicker}>05 · Discussão interna</div>
              <h2 className={styles.sectionTitle}>Comentários ({PLACEHOLDER_COMMENTS.length})</h2>
            </div>
          </div>
          <div className={styles.commentsCard}>
            {PLACEHOLDER_COMMENTS.map((c, i) => (
              <div
                key={i}
                className={`${styles.commentRow} ${i < PLACEHOLDER_COMMENTS.length - 1 ? styles.commentRowBorder : ''}`}
              >
                <Avatar name={c.who} size="md" />
                <div className={styles.commentContent}>
                  <div className={styles.commentMeta}>
                    <span className={styles.commentAuthor}>{c.who}</span>
                    <span className={styles.commentRole}>· {c.role}</span>
                    <span className={styles.commentWhen}>{c.when}</span>
                  </div>
                  <div className={styles.commentText}>{c.text}</div>
                </div>
              </div>
            ))}
            {/* Reply box shell — TODO(backlog): wire comment submission */}
            <div className={styles.replyRow}>
              <Avatar name={user?.displayName ?? 'Você'} size="md" />
              <div className={styles.replyBox}>
                <span className={styles.replyPlaceholder}>Adicionar um comentário…</span>
                <button className="btn btn-sm btn-primary" type="button" aria-disabled="true" title="Em breve">
                  Comentar
                </button>
              </div>
            </div>
          </div>
        </section>

      </div>
    </div>
  );
}

