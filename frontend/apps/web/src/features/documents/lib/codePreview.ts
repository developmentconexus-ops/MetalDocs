// Shared, presentational code-preview formatter for the document-creation flow.
// Mirrors the lifecycle states of the preview-code query so neither the wizard
// banner nor the confirm step renders a bare "???" when profile+area are known.
// The backend assigns the final immutable code at confirm time; this is the
// optimistic next-in-sequence preview.
//
// `text` is ALWAYS code-shaped so callers can interpolate it into prose
// ("O slot X será reservado…") without the sentence collapsing. Callers that
// need to react to failure read `state` instead of sniffing the string —
// F-QA4-1: a failed preview must surface as an explicit error, never as the
// "???" placeholder that also means "nothing selected yet".
export const CODE_PLACEHOLDER = '???';

// Shown by the banner when the preview-code query fails. The preview route is
// tier-1 gated on `controlled_documents.create`, so a 403 here means the caller
// cannot create at all — still an error to surface, never a silent placeholder.
export const CODE_PREVIEW_ERROR_MESSAGE = 'Não foi possível gerar o código';

export type CodePreviewState =
  | 'unselected'   // profile and/or area not chosen yet
  | 'loading'      // preview-code query in flight
  | 'error'        // preview-code query failed
  | 'unavailable'  // query settled but returned no code
  | 'ready';       // code resolved

export type CodePreviewResult = {
  text: string;
  state: CodePreviewState;
};

export type CodePreviewInput = {
  ready: boolean;                  // profile AND area both selected
  isLoading: boolean;              // preview-code query in first fetch
  isError?: boolean;               // preview-code query failed
  code: string | null | undefined; // resolved preview code, when available
  profileCode: string | null;
  areaCode: string | null;
};

export function formatCodePreview({
  ready,
  isLoading,
  isError = false,
  code,
  profileCode,
  areaCode,
}: CodePreviewInput): CodePreviewResult {
  if (!ready) {
    return {
      text: `${CODE_PLACEHOLDER}-${CODE_PLACEHOLDER}-${CODE_PLACEHOLDER}`,
      state: 'unselected',
    };
  }
  if (isLoading) {
    return { text: `${profileCode}-${areaCode}-…`, state: 'loading' };
  }
  // Error wins over the "no code" fallback: both render the same code-shaped
  // text, but only one of them is a failure the user must be told about.
  if (isError) {
    return { text: `${profileCode}-${areaCode}-${CODE_PLACEHOLDER}`, state: 'error' };
  }
  if (code == null) {
    return { text: `${profileCode}-${areaCode}-${CODE_PLACEHOLDER}`, state: 'unavailable' };
  }
  return { text: code, state: 'ready' };
}
