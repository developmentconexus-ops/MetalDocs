import { useQuery } from '@tanstack/react-query';

import { QK } from '../../../lib/queryKeys';
import { fetchExportedPDFBlob } from '../api/exports';

/**
 * D4 (2026-07-28) — server state for the embedded PDF viewer: the rendered
 * artifact as a `Blob`, ready for `URL.createObjectURL`.
 *
 * The hook deliberately caches the Blob, not an object URL: object URLs are a
 * document-scoped resource that must be revoked by whoever created them, so
 * their lifecycle stays inside the consuming component (`DocumentPdfViewerDialog`)
 * while the cache holds only inert bytes.
 */
export function useDocumentPdfExportQuery(documentId: string, enabled: boolean) {
  return useQuery({
    queryKey: QK.documents.pdfExport(documentId),
    queryFn: () => fetchExportedPDFBlob(documentId, { paper_size: 'A4' }),
    enabled: enabled && documentId !== '',
    // The export route renders on demand and is rate-limited (429). Never
    // auto-retry it, and keep the generated artifact warm for the viewing
    // session so re-opening the viewer does not re-render the document.
    retry: false,
    staleTime: 5 * 60_000,
    gcTime: 10 * 60_000,
    refetchOnWindowFocus: false,
  });
}
