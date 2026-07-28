import { useEffect, useState } from 'react';

import { Icon } from '../../../components/ui/Icon';
import { resolveQueryError } from '../../../lib/api';
import { useDocumentPdfExportQuery } from '../queries/useDocumentPdfExportQuery';
import styles from './DocumentPdfViewerDialog.module.css';

export interface DocumentPdfViewerDialogProps {
  documentId: string;
  /** Human label (governed code, falling back to the name) shown in the header. */
  documentLabel: string;
  onClose: () => void;
}

/**
 * D4 (2026-07-28) — the embedded PDF viewer.
 *
 * Ratified product decision: viewing a document's PDF happens INSIDE the screen;
 * downloading is a separate, explicit action. The artifact is fetched through the
 * authenticated export call (`useDocumentPdfExportQuery` → `fetchExportedPDFBlob`)
 * and embedded from an object URL — never handed to an `<a download>` and never
 * opened in another tab.
 *
 * Modal chrome mirrors the existing dialog primitive (`CancelInstanceDialog` /
 * `SupersedePublishDialog`): fixed scrim + `role="dialog" aria-modal`, the dialog
 * owns its own read + error state, and the route only toggles visibility.
 *
 * Every state is explicit — loading (`role="status"`), failure (`role="alert"` +
 * retry), and the fail-closed "loaded but no bytes" case. Nothing fails silently.
 */
export function DocumentPdfViewerDialog({
  documentId,
  documentLabel,
  onClose,
}: DocumentPdfViewerDialogProps) {
  const query = useDocumentPdfExportQuery(documentId, true);
  const blob = query.data ?? null;
  const [objectUrl, setObjectUrl] = useState<string | null>(null);

  // Object-URL lifecycle: one URL per Blob, revoked when the Blob changes and on
  // unmount. Without the revoke the bytes stay pinned for the lifetime of the tab.
  useEffect(() => {
    if (!blob) {
      setObjectUrl(null);
      return;
    }
    const url = URL.createObjectURL(blob);
    setObjectUrl(url);
    return () => {
      URL.revokeObjectURL(url);
    };
  }, [blob]);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  const isLoading = query.isPending || query.isFetching;
  // Fail closed: a settled query with no bytes is an error, not an endless spinner.
  const missingArtifact = query.isSuccess && !objectUrl;
  const errorMessage = query.isError
    ? resolveQueryError(query.error, 'Não foi possível carregar o PDF deste documento.')
    : missingArtifact
      ? 'O PDF deste documento foi gerado, mas veio vazio. Tente novamente.'
      : null;

  return (
    <div className={styles.overlay} role='presentation'>
      <div
        className={styles.dialog}
        role='dialog'
        aria-modal='true'
        aria-label={`Visualizar PDF do documento ${documentLabel}`}
      >
        <header className={styles.header}>
          <h2 className={styles.title}>Visualizar PDF — {documentLabel}</h2>
          <button
            type='button'
            className={styles.closeBtn}
            aria-label='Fechar visualização do PDF'
            onClick={onClose}
          >
            <Icon name='x' size={16} />
          </button>
        </header>

        <div className={styles.body}>
          {errorMessage ? (
            <div className={styles.state} role='alert'>
              <p className={styles.stateTitle}>{errorMessage}</p>
              <button type='button' className={styles.retryBtn} onClick={() => void query.refetch()}>
                Tentar novamente
              </button>
            </div>
          ) : isLoading || !objectUrl ? (
            <div className={styles.state} role='status' aria-live='polite'>
              Carregando o PDF do documento…
            </div>
          ) : (
            <iframe
              className={styles.frame}
              title={`PDF do documento ${documentLabel}`}
              src={objectUrl}
            />
          )}
        </div>

        <footer className={styles.footer}>
          <button type='button' className='btn' onClick={onClose}>
            Fechar
          </button>
        </footer>
      </div>
    </div>
  );
}
