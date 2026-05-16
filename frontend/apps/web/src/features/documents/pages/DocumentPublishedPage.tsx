import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Icon } from '../../../components/ui/Icon';
import { Avatar } from '../../../components/ui/Avatar';
import { CodeChip } from '../../../components/ui/CodeChip';
import { DocumentHero } from '../components/DocumentHero';
import { DocumentVersionTimeline } from '../components/DocumentVersionTimeline';
import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
import { useApprovalInstanceQuery } from '../queries/useApprovalInstanceQuery';
import { useAuthStore } from '../../../store/auth.store';
import { formatPublishedAt, formatSignedAt, formatShortDate } from '../lib/documentDetailMeta';
import styles from './DocumentPublishedPage.module.css';

// TODO(backlog): Replace with real revision list from GET /api/v1/documents/:id/revisions
// See wiki/backlog/documento-publicado.md
const PLACEHOLDER_VERSIONS = [
  { v: 'v2.3', when: '18 mar 2025', author: 'A. Tavares', current: false,
    summary: 'Ajustes de redação após feedback de Engenharia.' },
  { v: 'v2.4', when: '02 jul 2025', author: 'R. Souza', current: false,
    summary: 'Revisão geral pós-auditoria interna. Novas etiquetas padronizadas para todas as plantas.' },
  { v: 'v3.0', when: '14 nov 2025', author: 'R. Souza', current: false,
    summary: 'Reestruturação completa do procedimento. Inclusão de seção para fontes energéticas múltiplas.' },
  { v: 'v3.1', when: '08 jan 2026', author: 'C. Mendes', current: false,
    summary: 'Correções de conformidade pós-treinamento. Atualização dos responsáveis por área.' },
  { v: 'v3.2', when: '12 mar 2026', author: 'C. Mendes', current: true,
    summary: 'Inclusão de procedimento para painéis CCM-04. Atualização da matriz de risco para fontes hidráulicas.' },
];

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

  const docQuery = useDocumentDetailQuery(documentId);
  const approvalQuery = useApprovalInstanceQuery(documentId);

  const [linkCopied, setLinkCopied] = useState(false);

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
  const code            = doc.code    ?? doc.Code           ?? '—';
  const docName         = doc.name    ?? doc.Name;
  const status          = doc.status  ?? doc.Status         ?? '';
  const createdByRaw    = doc.created_by ?? doc.CreatedBy   ?? '—';
  // TODO(backlog): API should return created_by_display_name snapshot — wiki/backlog/documento-publicado.md
  // Fallback: created_by stores user_id; if creator is the current user, use displayName
  const createdBy = (user?.userId && createdByRaw === user.userId)
    ? (user.displayName ?? createdByRaw)
    : createdByRaw;
  const revisionVersion = doc.revision_version ?? doc.RevisionVersion;
  const versionLabel    = revisionVersion != null ? `v${revisionVersion}` : '—';

  // Published date: use approval instance updated_at (= when approval completed)
  const publishedAt = formatPublishedAt(approval?.updated_at);
  const sinceDateHint = approval?.updated_at ? formatShortDate(approval.updated_at) : '—';

  // ObsoleteBanner
  const isObsolete = status === 'obsolete';

  // RBAC: roles that can initiate a revision
  // TODO(backlog): refine with controlled_document owner_id once API exposes it
  // See wiki/backlog/documento-publicado.md
  const canInitiateRevision =
    user != null &&
    user.roles.some((r) => ['admin', 'system_admin', 'editor', 'qms_admin', 'area_admin'].includes(r));

  // Signoff stages from approval instance
  const signoffStages = approval?.stages ?? [];
  const stageCount = signoffStages.length;
  // Connector spans from first pin center to last pin center
  const connectorSide = stageCount > 1 ? `${(100 / (2 * stageCount)).toFixed(2)}%` : '50%';

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
              {versionLabel} · vigente
            </span>
            {/* TODO(backlog): resolve profile_code → human label once DocumentResponse includes profile_code */}
            <span className={styles.typeLabel}>—</span>
          </>
        }
        title={docName ?? code}
        subtitle={publishedAt ? <span>publicado em {publishedAt}</span> : null}
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
            {/* TODO(backlog): wire POST /api/v1/controlled-documents/:cdId/revisions */}
            <button
              className="btn"
              type="button"
              aria-disabled={!canInitiateRevision}
              title={canInitiateRevision ? undefined : 'Sem permissão para iniciar revisão'}
            >
              <Icon name="edit" size={13} />
              Iniciar revisão
            </button>
            <button className="btn btn-ghost" type="button" onClick={handleCopyLink}>
              <Icon name={linkCopied ? 'check' : 'link'} size={13} />
              {linkCopied ? 'Link copiado!' : 'Copiar link'}
            </button>
          </div>
        }
      />

      {/* Content area */}
      <div className={styles.content}>

        {/* KPI strip */}
        <div className={styles.kpiStrip}>
          <div className={styles.kpiCell}>
            <div className={styles.kpiLabel}>Versão atual</div>
            <div className={styles.kpiValue}>{versionLabel}</div>
            <div className={styles.kpiHint}>
              {sinceDateHint !== '—' ? `desde ${sinceDateHint}` : '—'}
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
                  <div className={styles.ownerMeta}>
                    {publishedAt ? `publicou em ${publishedAt}` : 'publicado'}
                  </div>
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
                    <div key={stage.stage_id} className={styles.signoffStage}>
                      <div className={styles.signoffPin}>
                        <Icon name="check" size={16} />
                      </div>
                      <div className={styles.signoffStageName}>{stage.name}</div>
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

        {/* Section: Histórico de versões — TODO(backlog): wire GET /api/v1/documents/:id/revisions */}
        <section className={styles.section}>
          <div className={styles.sectionHead}>
            <div>
              <div className={styles.sectionKicker}>03 · Linhagem</div>
              <h2 className={styles.sectionTitle}>Histórico de versões</h2>
            </div>
            <span className={styles.sectionAside}>passe o mouse para detalhar</span>
          </div>
          <DocumentVersionTimeline versions={PLACEHOLDER_VERSIONS} />
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
