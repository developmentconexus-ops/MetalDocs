import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Icon } from '../../../components/ui/Icon';
import { CodeChip } from '../../../components/ui/CodeChip';
import { InlineAlert } from '../../../components/ui/InlineAlert';
import { resolveQueryError, ApiError } from '../../../lib/api';
import { ArtifactDetailView } from '../../shared/controlled-artifact/ArtifactDetailView';
import { useDocumentArtifact } from '../adapters/useDocumentArtifact';
import type { ReleaseNoticeTone } from '../lib/documentReleasePresentation';
import { createRevision } from '../../controlled-documents/api/controlledDocuments';
import { exportPDF } from '../api/exports';
import { DocumentPdfViewerDialog } from '../components/DocumentPdfViewerDialog';
import { useHasCapability } from '../../../lib/iam/useHasCapability';
import styles from './DocumentDetailRoute.module.css';

/**
 * Document-specific route wrapper for the shared ArtifactDetailView. Owns only
 * interactive/dialog UI state (revision composer, PDF viewer, PDF download, copy
 * link) and injects it as heroActions / aside / extras slots.
 *
 * D4 (2026-07-28): "Visualizar PDF" opens the embedded `DocumentPdfViewerDialog`
 * (the PDF is read inside the app); "Baixar PDF" stays a separate, explicit
 * download action. The two must never be collapsed back into one affordance.
 *
 * ADR 0085 (Stage B): there is no manual "Publicar / Agendar" affordance. Release
 * is an approval-driven coordinator outcome (the publication plan travels on
 * submit). ADR 0085 (Stage C): where that affordance used to sit, the hero now
 * REPORTS the coordinator's readiness-hold projection (`release`) — read-only, no
 * CTA in any state, including the anomaly ones (resolution is operator/backend
 * work). The lifecycle STATUS chip still comes from `doc.status` via the shared
 * view; this block only adds the release fact on top of it.
 *
 * All queries and
 * lifecycle/capability gating are owned by `useDocumentArtifact` (FE-02) — the route
 * no longer re-fetches the document/approval/active-document/distribution queries or
 * re-derives gating; it consumes `gating` + the raw `doc`/`activeDocument` the adapter
 * already resolved. No kind logic leaks into the shared view — it receives only a
 * composed ArtifactViewModel + ReactNode slots.
 */
/** Tone → route-local notice class. `anomaly` keeps the neutral-error treatment. */
const RELEASE_NOTICE_CLASS: Record<ReleaseNoticeTone, string> = {
  released: styles.releaseNoticeReleased,
  progress: styles.releaseNoticeProgress,
  scheduled: styles.releaseNoticeScheduled,
  anomaly: styles.releaseNoticeAnomaly,
};

export function DocumentDetailRoute() {
  const { documentId: rawDocumentId } = useParams<{ documentId: string }>();
  const navigate = useNavigate();
  const canViewObsolete = useHasCapability('document.obsolete');

  const documentId = rawDocumentId ?? '';

  const { model, isLoading, isError, refetch, doc, obligatedCount, gating, release } =
    useDocumentArtifact(documentId);

  const [linkCopied, setLinkCopied] = useState(false);
  const [showRevisionForm, setShowRevisionForm] = useState(false);
  const [revisionName, setRevisionName] = useState('');
  const [revisionError, setRevisionError] = useState('');
  const [isCreatingRevision, setIsCreatingRevision] = useState(false);
  const [showPdfViewer, setShowPdfViewer] = useState(false);
  const [pdfStatus, setPdfStatus] = useState<
    | { kind: 'idle' }
    | { kind: 'pending' }
    | { kind: 'rate_limited'; retryAfterSec: number }
    | { kind: 'error'; message: string }
  >({ kind: 'idle' });

  // ── Loading state ──────────────────────────────────────────────────────────
  if (isLoading) {
    return (
      <div className={styles.stateLoading} role='status' aria-live='polite'>
        <Icon name='docs' size={24} className={styles.stateIcon} />
        <span>Carregando documento…</span>
      </div>
    );
  }

  // ── Error state ────────────────────────────────────────────────────────────
  if (isError) {
    return (
      <div className={styles.stateError} role='alert'>
        <Icon name='x' size={20} className={styles.stateIcon} />
        <span>Documento não encontrado ou sem permissão de acesso.</span>
        <button className='btn btn-sm' type='button' onClick={() => refetch()}>
          Tentar novamente
        </button>
      </div>
    );
  }

  const code = model.code ?? '—';
  const docName = doc?.name ?? code;
  const {
    isObsolete,
    isPublished,
    canCreateRevision,
    canInitiateRevision,
    activeSiblingDocumentId,
    activeSiblingCtaLabel,
    activeSiblingDestination,
  } = gating;

  const handleDownloadPDF = async () => {
    setPdfStatus({ kind: 'pending' });
    try {
      const res = await exportPDF(documentId, { paper_size: 'A4' });
      window.open(res.signed_url, '_blank', 'noopener');
      setPdfStatus({ kind: 'idle' });
    } catch (e: unknown) {
      const err = e instanceof ApiError ? e : null;
      if (err?.status === 429) {
        setPdfStatus({ kind: 'rate_limited', retryAfterSec: 60 });
        return;
      }
      setPdfStatus({ kind: 'error', message: resolveQueryError(e, 'Falha ao gerar PDF.') });
    }
  };

  const handleView = () => navigate(`/documents/${documentId}`);

  const handleCopyLink = () => {
    navigator.clipboard.writeText(window.location.href)
      .then(() => {
        setLinkCopied(true);
        setTimeout(() => setLinkCopied(false), 2000);
      })
      .catch(() => {/* silent */});
  };

  const handleStartRevision = () => {
    if (!canCreateRevision || isCreatingRevision) return;
    setRevisionError('');
    setRevisionName(docName);
    setShowRevisionForm(true);
  };

  const handleContinueActiveRevision = () => {
    if (!activeSiblingDestination) return;
    navigate(activeSiblingDestination);
  };

  const handleCancelRevision = () => {
    if (isCreatingRevision) return;
    setShowRevisionForm(false);
    setRevisionName('');
    setRevisionError('');
  };

  const handleCreateRevision = async () => {
    if (!doc?.controlled_document_id || !revisionName.trim()) return;
    setIsCreatingRevision(true);
    setRevisionError('');
    try {
      const response = await createRevision(
        doc.controlled_document_id,
        { name: revisionName.trim(), form_data: {}, template_version_id: doc.template_version_id },
        crypto.randomUUID(),
      );
      navigate(`/documents/${response.document.id}`);
    } catch (err) {
      setRevisionError(resolveQueryError(err, 'Falha ao iniciar nova revisão.'));
    } finally {
      setIsCreatingRevision(false);
    }
  };

  // ── Hero actions slot ──────────────────────────────────────────────────────
  const heroActions = (
    <div className={styles.heroActions}>
      <button
        className='btn btn-primary btn-lg'
        type='button'
        onClick={handleView}
        disabled={isObsolete && !canViewObsolete}
        title={
          isObsolete && !canViewObsolete
            ? 'Sua sessão não inclui a capacidade para visualizar documentos obsoletos.'
            : undefined
        }
      >
        <Icon name='eye' size={15} />
        Visualizar documento
      </button>
      {/*
        D4 — reading the PDF happens in the embedded viewer, never by leaving the
        screen. The download below is the separate, explicit alternative.
      */}
      <button
        className='btn'
        type='button'
        aria-label='Visualizar PDF'
        onClick={() => setShowPdfViewer(true)}
        disabled={isObsolete && !canViewObsolete}
        title={
          isObsolete && !canViewObsolete
            ? 'Sua sessão não inclui a capacidade para visualizar documentos obsoletos.'
            : 'Visualizar PDF'
        }
      >
        <Icon name='eye' size={13} />
        Visualizar PDF
      </button>
      <button
        className='btn'
        type='button'
        aria-label='Baixar PDF'
        onClick={() => void handleDownloadPDF()}
        disabled={isObsolete || pdfStatus.kind === 'pending' || pdfStatus.kind === 'rate_limited'}
        title={
          pdfStatus.kind === 'rate_limited'
            ? `Limite de exportações atingido — tente novamente em ${pdfStatus.retryAfterSec}s`
            : 'Baixar PDF'
        }
      >
        <Icon name='download' size={13} />
        {pdfStatus.kind === 'pending'
          ? 'Gerando PDF…'
          : pdfStatus.kind === 'rate_limited'
            ? 'Aguarde…'
            : 'Baixar PDF'}
      </button>
      {pdfStatus.kind === 'rate_limited' && (
        <span role='alert' className={styles.pdfAlert}>
          Limite de exportações atingido — tente novamente em {pdfStatus.retryAfterSec}s.
        </span>
      )}
      {pdfStatus.kind === 'error' && (
        <span role='alert' className={styles.pdfAlert}>{pdfStatus.message}</span>
      )}
      {/*
        Soft-disable: aria-disabled (not native `disabled`) is deliberate — the
        action keeps keyboard focus so AT users can reach the `title` tooltip that
        explains WHY it is unavailable (missing permission / active revision).
        The onClick handler is the guard for the action itself.
      */}
      <button
        className='btn'
        type='button'
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
        <Icon name='edit' size={13} />
        {activeSiblingCtaLabel}
      </button>
      <button className='btn btn-ghost' type='button' onClick={handleCopyLink} disabled={isObsolete}>
        <Icon name={linkCopied ? 'check' : 'link'} size={13} />
        {linkCopied ? 'Link copiado!' : 'Copiar link'}
      </button>
      {/*
        ADR 0085 Stage C — release readiness-hold projection, rendered exactly where
        the deleted "Publicar / Agendar" surface used to live. Report-only: every
        tone is informational, so it reuses the shared InlineAlert primitive
        (role=status / aria-live=polite) instead of an interrupting alert. Absent
        projection (`release === null`) renders nothing at all.
      */}
      {release ? (
        <InlineAlert
          tone={release.tone === 'anomaly' ? 'warning' : 'info'}
          className={`${styles.releaseNotice} ${RELEASE_NOTICE_CLASS[release.tone]}`}
          message={
            <>
              <span className={styles.releaseNoticeTitle}>{release.title}</span>
              {release.detail ? (
                <span className={styles.releaseNoticeDetail}>{release.detail}</span>
              ) : null}
            </>
          }
        />
      ) : null}
    </div>
  );

  // ── Coverage side-card slot (aside) ───────────────────────────────────────
  const aside = (
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
          type='button'
          className={`btn btn-sm ${styles.coverageCardAction}`}
          onClick={() => navigate('distribution')}
        >
          Abrir Fanout
          <Icon name='chevron-right' size={14} />
        </button>
      </div>
    </aside>
  );

  // ── Extras slot: revision composer, PDF viewer, obsolete banner ───────────
  const extras = (
    <>
      {isObsolete && (
        <div className={styles.obsoleteBanner} role='alert' aria-label='Documento obsoleto'>
          <div className={styles.obsoleteStamp}>OBSOLETO</div>
        </div>
      )}
      {showRevisionForm && (
        <div className={styles.revisionComposer} role='region' aria-label='Iniciar nova revisão'>
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
          <label className={styles.revisionComposerField} htmlFor='published-revision-name'>
            Nome do documento
          </label>
          <input
            id='published-revision-name'
            className={`input ${styles.revisionComposerInput}`}
            type='text'
            value={revisionName}
            onChange={(event) => setRevisionName(event.target.value)}
            placeholder={docName}
            disabled={isCreatingRevision}
          />
          {revisionError && (
            <p className={styles.revisionComposerError} role='alert'>
              {revisionError}
            </p>
          )}
          <div className={styles.revisionComposerActions}>
            <button
              className='btn btn-primary'
              type='button'
              onClick={() => void handleCreateRevision()}
              disabled={isCreatingRevision || revisionName.trim().length === 0}
            >
              {isCreatingRevision ? 'Criando...' : 'Gerar documento'}
            </button>
            <button className='btn' type='button' onClick={handleCancelRevision} disabled={isCreatingRevision}>
              Cancelar
            </button>
          </div>
        </div>
      )}
      {showPdfViewer ? (
        <DocumentPdfViewerDialog
          documentId={documentId}
          documentLabel={code !== '—' ? code : docName}
          onClose={() => setShowPdfViewer(false)}
        />
      ) : null}
    </>
  );

  // ── Root class: obsolete dim ───────────────────────────────────────────────
  // ArtifactDetailView.root does not carry rootObsolete; we wrap it here so the
  // shared view stays unaware of the obsolete variant.
  const rootClass = isObsolete ? `${styles.root} ${styles.rootObsolete}` : undefined;

  return (
    <div className={rootClass} style={rootClass ? undefined : { display: 'contents' }}>
      <ArtifactDetailView
        model={model}
        heroActions={heroActions}
        aside={aside}
        extras={extras}
      />
    </div>
  );
}
