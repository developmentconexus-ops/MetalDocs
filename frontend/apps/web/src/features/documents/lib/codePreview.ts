// Shared, presentational code-preview formatter for the document-creation flow.
// Mirrors the three lifecycle states of the preview-code query so neither the
// wizard banner nor the confirm step renders a bare "???" when profile+area
// are known. The backend assigns the final immutable code at confirm time;
// this is the optimistic next-in-sequence preview.
export const CODE_PLACEHOLDER = '???';

export type CodePreviewInput = {
  ready: boolean;                  // profile AND area both selected
  isLoading: boolean;              // preview-code query in first fetch
  code: string | null | undefined; // resolved preview code, when available
  profileCode: string | null;
  areaCode: string | null;
};

export function formatCodePreview({
  ready,
  isLoading,
  code,
  profileCode,
  areaCode,
}: CodePreviewInput): string {
  if (!ready) return `${CODE_PLACEHOLDER}-${CODE_PLACEHOLDER}-${CODE_PLACEHOLDER}`;
  if (isLoading) return `${profileCode}-${areaCode}-…`;
  return code ?? `${profileCode}-${areaCode}-${CODE_PLACEHOLDER}`;
}
