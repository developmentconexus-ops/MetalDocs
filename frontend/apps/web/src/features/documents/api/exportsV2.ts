import { apiFetch } from '../../../lib/api';

export type ExportPDFResult = {
  storage_key: string;
  signed_url: string;
  composite_hash: string;
  size_bytes: number;
  cached: boolean;
  revision_id: string;
};

export type DocxURLResult = {
  signed_url: string;
  revision_id: string;
};

export async function exportPDF(
  documentID: string,
  opts: { paper_size?: 'A4' | 'Letter'; landscape?: boolean } = {},
): Promise<ExportPDFResult> {
  return apiFetch(`/api/v2/documents/${documentID}/export/pdf`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(opts),
  });
}

export async function getDocxSignedURL(documentID: string): Promise<DocxURLResult> {
  return apiFetch(`/api/v2/documents/${documentID}/export/docx-url`);
}
