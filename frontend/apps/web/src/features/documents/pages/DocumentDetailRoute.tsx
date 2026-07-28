import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Icon } from '../../../components/ui/Icon';
import { CodeChip } from '../../../components/ui/CodeChip';
import { resolveQueryError, ApiError } from '../../../lib/api';
import { ArtifactDetailView } from '../../shared/controlled-artifact/ArtifactDetailView';
import { useDocumentArtifact } from '../adapters/useDocumentArtifact';
import { createRevision } from '../../controlled-documents/api/controlledDocuments';
import { exportPDF } from '../api/exports';
import { DocumentPdfViewerDialog } from '../components/DocumentPdfViewerDialog';
import { SupersedePublishDialog } from '../../approval/components/SupersedePublishDialog';
import { useHasCapability } from '../../iam/hooks/useHasCapability';
import styles from './DocumentDetailRoute.module.css';

/**
 * Document-specific route wrapper for the shared ArtifactDetailView. Owns only
 * interactive/dialog UI state (revision composer, publish dialog, PDF viewer, PDF
 * download, copy link) and injects it as heroActions / aside / extras slots.
 *
 * D4 (2026-07-28): "Visualizar PDF" opens the embedded `DocumentPdfViewerDialog`
 * (the PDF is read inside the app); "Baixar PDF" stays a separate, explicit
 * download action. The two must never be collapsed back into one affordance.
 *
 * All queries and
 * lifecycle/capability gating are owned by `useDocumentArtifact` (FE-02) — the route
 * no longer re-fetches the document/approval/active-document/distribution queries or
 * re-derives gating; it consumes `gating` + the raw `doc`/`activeDocument` the adapter
 * already resolved. No kind logic leaks into the shared view — it receives only a
 * composed ArtifactViewModel + ReactNode slots.
 */
export function DocumentDetailRoute() {
  const { documentId: rawDocumentId } = useParams<{ documentId: string }>();
  const navigate = useNavigate();
  const canViewObsolete = useHasCapability('document.obsolete');

  const documentId = rawDocumentId ?? '';

  const { model, isLoading, isError, refetch, doc, activeDocument, obligatedCount, gating, refetchAll } =
    useDocumentArtifact(documentId);

  const [linkCopied, setLinkCopied] = useState(false);
  const [showRevisionForm, setShowRevisionForm] = useState(false);
  const [revisionName, setRevisionName] = useState('');
  const [revisionError, setRevisionError] = useState('');
  const [isCreatingRevision, setIsCreatingRevision] = useState(false);
  const [showPublishDialog, setShowPublishDialog] = useState(false);
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
    isApproved,
    isPublished,
    canCreateRevision,
    canPublish,
    canInitiateRevision,
    activeSiblingDocumentId,
    activeSiblingCtaLabel,
    activeSiblingDestination,
    publishContextNotice,
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

  const handlePublishSuccess = () => {
    refetchAll();
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
        Soft-disable: aria-disabled (not native `disabled`) is deliberate — these
        actions keep keyboard focus so AT users can reach the `title` tooltip that
        explains WHY the action is unavailable (missing permission / active context).
        The onClick guards (`if (canPublish) …`) suppress the action itself.
      */}
      {isApproved ? (
        <button
          className='btn'
          type='button'
          aria-disabled={!canPublish}
          title={
            !canInitiateRevision
              ? 'Sem permissão para publicar'
              : !activeDocument?.content_hash
                ? 'Aguardando contexto ativo para publicar'
                : undefined
          }
          onClick={() => {
            if (canPublish) setShowPublishDialog(true);
          }}
        >
          <Icon name='check' size={13} />
          Publicar / Agendar
        </button>
      ) : (
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
      )}
      <button className='btn btn-ghost' type='button' onClick={handleCopyLink} disabled={isObsolete}>
        <Icon name={linkCopied ? 'check' : 'link'} size={13} />
        {linkCopied ? 'Link copiado!' : 'Copiar link'}
      </button>
      {publishContextNotice ? (
        <div className={styles.heroNotice} role='status' aria-live='polite'>
          <Icon name='shield' size={14} />
          <span>{publishContextNotice}</span>
        </div>
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

  // ── Extras slot: revision composer, publish dialog, obsolete banner ───────
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
      {showPublishDialog && activeDocument?.content_hash ? (
        <SupersedePublishDialog
          documentId={documentId}
          revisionVersion={activeDocument.revision_version}
          publishedDocumentId={activeDocument.published_document_id ?? undefined}
          onClose={() => setShowPublishDialog(false)}
          onSuccess={handlePublishSuccess}
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
