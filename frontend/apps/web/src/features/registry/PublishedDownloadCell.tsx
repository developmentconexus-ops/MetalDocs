import { useDocumentPdfStatus } from '../documents/v2/hooks/useDocumentPdfStatus';
import { PDFCell } from '../documents/v2/PDFCell';

export function PublishedDownloadCell({ documentId }: { documentId: string }) {
  const pdf = useDocumentPdfStatus(documentId, true);
  return <PDFCell status={pdf.status} url={pdf.url} onRetry={pdf.retry} />;
}
