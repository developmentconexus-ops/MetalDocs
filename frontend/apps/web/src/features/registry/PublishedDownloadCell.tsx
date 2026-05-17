import { useDocumentPdfStatus } from '../documents/hooks/editor/useDocumentPdfStatus';
import { PDFCell } from '../documents/components/PDFCell';

export function PublishedDownloadCell({ documentId }: { documentId: string }) {
  const pdf = useDocumentPdfStatus(documentId, true);
  return <PDFCell status={pdf.status} url={pdf.url} onRetry={pdf.retry} />;
}
