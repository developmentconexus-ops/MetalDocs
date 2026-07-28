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
  return apiFetch(`/api/v1/documents/${documentID}/export/pdf`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(opts),
  });
}

export async function getDocxSignedURL(documentID: string): Promise<DocxURLResult> {
  return apiFetch(`/api/v1/documents/${documentID}/export/docx-url`);
}

/**
 * Product decision D4 (2026-07-28) — the document PDF must be VIEWABLE inside the
 * screen, not only downloaded. `POST /export/pdf` answers with a presigned storage
 * URL (JSON), so the artifact itself needs a second hop before it can be embedded.
 *
 * Hop 1 is the authenticated API call (`apiFetch` — session cookie, problem+json).
 * Hop 2 is a deliberate raw `fetch`: the presigned URL points at blob storage, not
 * at `/api/...`, so it must not carry the API session nor the apiFetch error
 * envelope. Same two-hop shape `DocumentShell` already uses for the signed revision
 * URL (`components/DocumentShell.tsx`).
 *
 * The blob is re-typed as `application/pdf` so the embed renders inline regardless
 * of the Content-Type the storage backend echoes back.
 *
 * Thrown messages are already user-facing PT-BR: `resolveQueryError` surfaces
 * `Error.message` verbatim, and no FE-synthetic problem code may be added to
 * `lib/api/errorMessages.ts` (it is pinned 1:1 against the backend snapshot).
 */
export async function fetchExportedPDFBlob(
  documentID: string,
  opts: { paper_size?: 'A4' | 'Letter'; landscape?: boolean } = {},
): Promise<Blob> {
  const res = await exportPDF(documentID, opts);
  if (!res.signed_url) {
    throw new Error('O PDF gerado não retornou um endereço válido para visualização.');
  }
  const fileRes = await fetch(res.signed_url);
  if (!fileRes.ok) {
    throw new Error('Não foi possível carregar o arquivo PDF gerado. Tente novamente.');
  }
  return new Blob([await fileRes.arrayBuffer()], { type: 'application/pdf' });
}
