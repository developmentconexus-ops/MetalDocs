import { useDocumentPdfStatus } from '../../hooks/editor/useDocumentPdfStatus';
import styles from './PdfCanvas.module.css';

/**
 * F2d.5b D1 — the official post-approval PDF canvas.
 *
 * Rendered by DocumentWorkspacePage ONLY for document statuses the backend
 * serves via GET /documents/:id/view (approved/scheduled/published —
 * view_service.go viewableStatuses). Reuses the useDocumentPdfStatus polling
 * hook (pending → 3s poll → ready/failed). The PDF is the official artifact;
 * in-approval viewing stays on the source canvas (ADR 0080 amendment,
 * design 2026-07-09-f5b-pdf-official-view-design.md).
 */
export function PdfCanvas({ documentId }: { documentId: string }) {
  const pdf = useDocumentPdfStatus(documentId, true);

  if (pdf.status === 'ready' && pdf.url) {
    return (
      <iframe
        className={styles.frame}
        title="Documento oficial (PDF)"
        src={pdf.url}
      />
    );
  }

  const missingReadyUrl = pdf.status === 'ready' && !pdf.url;
  if (pdf.status === 'failed' || missingReadyUrl) {
    return (
      <div role="alert" className={styles.state}>
        <p className={styles.stateTitle}>Não foi possível gerar o PDF oficial.</p>
        <button type="button" className={styles.retry} onClick={pdf.retry}>
          Tentar novamente
        </button>
      </div>
    );
  }

  return (
    <div role="status" aria-live="polite" className={styles.state}>
      Gerando o PDF oficial…
      {pdf.stalled ? (
        <button type="button" className={styles.retry} onClick={pdf.retry}>
          Tentar novamente
        </button>
      ) : null}
    </div>
  );
}
