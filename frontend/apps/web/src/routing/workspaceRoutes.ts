function normalizePath(pathname: string): string {
  if (!pathname) return "/";
  if (pathname.length > 1 && pathname.endsWith("/")) {
    return pathname.slice(0, -1);
  }
  return pathname;
}

// ---------------------------------------------------------------------------
// Template editor route helpers
// ---------------------------------------------------------------------------

export type TemplateEditorParams = {
  profileCode: string;
  templateKey: string;
};

/**
 * Returns true when the current pathname targets the template editor.
 * Pattern: /controlled-documents/profiles/:profileCode/templates/:templateKey/edit
 */
export function isTemplateEditorPath(pathname: string): boolean {
  return parseTemplateEditorPath(pathname) !== null;
}

/**
 * Parses `/controlled-documents/profiles/:profileCode/templates/:templateKey/edit`
 * and returns the params, or null if the path does not match.
 */
export function parseTemplateEditorPath(pathname: string): TemplateEditorParams | null {
  const path = normalizePath(pathname);
  const match = /^\/controlled-documents\/profiles\/([^/]+)\/templates\/([^/]+)\/edit$/.exec(path);
  if (!match) return null;
  return {
    profileCode: decodeURIComponent(match[1]),
    templateKey: decodeURIComponent(match[2]),
  };
}

/**
 * Builds the template editor path from params.
 */
export function buildTemplateEditorPath(params: TemplateEditorParams): string {
  return `/controlled-documents/profiles/${encodeURIComponent(params.profileCode)}/templates/${encodeURIComponent(params.templateKey)}/edit`;
}
