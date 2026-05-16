import type { PDFStatus } from '../hooks/editor/useDocumentPdfStatus';

type Props = { status: PDFStatus; url?: string; onRetry: () => void };

export function PDFCell({ status, url, onRetry }: Props) {
  if (status === 'ready' && url) {
    return <a href={url} download style={{ color: '#2a7a2a', fontWeight: 600 }}>Baixar PDF</a>;
  }
  if (status === 'failed') {
    return (
      <span style={{ color: '#c00' }}>
        Falha ao gerar PDF.
        <button type="button" onClick={onRetry} style={{ marginLeft: 4 }}>Tentar novamente</button>
      </span>
    );
  }
  return <span style={{ color: '#888' }}>Gerando PDF…</span>;
}
