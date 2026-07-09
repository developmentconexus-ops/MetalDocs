import { Navigate, useParams, type RouteObject } from "react-router-dom";

const RedirectToLibrary = () => <Navigate to="/documents" replace />;

// F2d.5 S3 — the cockpit/editor split is retired (ADR 0080); `/edit` now
// forwards into the mode-adaptive workspace at the canonical artifact path.
// `<Navigate to="...">` does not interpolate `:documentId` itself, so the
// param is read and interpolated here.
function RedirectEditToWorkspace() {
  const { documentId } = useParams<{ documentId: string }>();
  return <Navigate to={`/documents/${documentId}`} replace />;
}

export const documentsRoutes: RouteObject[] = [
  {
    path: 'documents',
    handle: { workspaceView: 'library' },
    lazy: () => import('./pages/LibraryPage').then((m) => ({ Component: m.default })),
  },
  { path: "documents/all", element: <RedirectToLibrary /> },
  { path: "documents/area/:areaCode", element: <RedirectToLibrary /> },
  { path: "documents/type/:profileCode", element: <RedirectToLibrary /> },
  { path: "documents/doc/:documentId", element: <RedirectToLibrary /> },
  { path: "documents/mine", element: <RedirectToLibrary /> },
  { path: "documents/mine/*", element: <RedirectToLibrary /> },
  { path: "documents/recent", element: <RedirectToLibrary /> },
  { path: "documents/recent/*", element: <RedirectToLibrary /> },
  {
    path: "documents/:documentId/edit",
    element: <RedirectEditToWorkspace />,
  },
  {
    path: 'documents/:documentId',
    handle: { workspaceView: 'document-editor' },
    lazy: () => import('./pages/DocumentWorkspacePage').then((m) => ({ Component: m.DocumentWorkspacePage })),
  },
  {
    path: 'documents/:documentId/details',
    handle: { workspaceView: 'library' },
    lazy: () => import('./pages/DocumentDetailLayout').then(m => ({ Component: m.DocumentDetailLayout })),
    children: [
      {
        index: true,
        lazy: () => import('./pages/DocumentDetailRoute').then(m => ({ Component: m.DocumentDetailRoute })),
      },
      {
        path: 'distribution',
        lazy: () => import('./pages/DocumentDistributionPage').then(m => ({ Component: m.DocumentDistributionPage })),
      },
    ],
  },
  {
    path: "documents/new",
    handle: { workspaceView: "document-editor" },
    lazy: () => import("./pages/NewDocumentWizardPage").then((m) => ({ Component: m.NewDocumentWizardPage })),
  },
];
