type DocumentsScopeView = "library" | "my-docs" | "recent";

const documentsBaseByView: Record<DocumentsScopeView, string> = {
  "library": "/documents",
  "my-docs": "/documents/mine",
  "recent": "/documents/recent",
};

function normalizePath(pathname: string): string {
  if (!pathname) return "/";
  if (pathname.length > 1 && pathname.endsWith("/")) {
    return pathname.slice(0, -1);
  }
  return pathname;
}

export function documentsBasePath(view: DocumentsScopeView): string {
  return documentsBaseByView[view];
}

type DocumentsRoute =
  | { view: "overview" }
  | { view: "collection"; areaCode?: string; profileCode?: string }
  | { view: "detail"; documentId: string };

export function parseDocumentsRoute(scopeView: DocumentsScopeView, pathname: string): DocumentsRoute {
  const basePath = documentsBasePath(scopeView);
  const path = normalizePath(pathname);

  if (!path.startsWith(basePath)) {
    return { view: "overview" };
  }

  const rest = path.slice(basePath.length).replace(/^\/+/, "");
  if (!rest) {
    return { view: "overview" };
  }

  if (rest === "all") {
    return { view: "collection" };
  }

  if (rest.startsWith("area/")) {
    return { view: "collection", areaCode: decodeURIComponent(rest.slice("area/".length)) };
  }

  if (rest.startsWith("type/")) {
    return { view: "collection", profileCode: decodeURIComponent(rest.slice("type/".length)) };
  }

  if (rest.startsWith("doc/")) {
    return { view: "detail", documentId: decodeURIComponent(rest.slice("doc/".length)) };
  }

  return { view: "overview" };
}

export function buildDocumentsPath(
  scopeView: DocumentsScopeView,
  target: { view: "overview" } | { view: "collection"; areaCode?: string; profileCode?: string } | { view: "detail"; documentId: string },
): string {
  const basePath = documentsBasePath(scopeView);

  if (target.view === "overview") {
    return basePath;
  }

  if (target.view === "detail") {
    return `${basePath}/doc/${encodeURIComponent(target.documentId)}`;
  }

  if (target.areaCode) {
    return `${basePath}/area/${encodeURIComponent(target.areaCode)}`;
  }

  if (target.profileCode) {
    return `${basePath}/type/${encodeURIComponent(target.profileCode)}`;
  }

  return `${basePath}/all`;
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

export type { DocumentsRoute, DocumentsScopeView };
