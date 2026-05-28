import { Navigate, type RouteObject } from "react-router-dom";

const RedirectToLibrary = () => <Navigate to="/documents" replace />;

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
    handle: { workspaceView: "document-editor" },
    lazy: () => import("./pages/DocumentEditorRoutePage"),
  },
  {
    path: 'documents/:documentId',
    handle: { workspaceView: 'library' },
    lazy: () => import('./pages/DocumentDetailLayout').then(m => ({ Component: m.DocumentDetailLayout })),
    children: [
      {
        index: true,
        lazy: () => import('./pages/DocumentPublishedPage').then(m => ({ Component: m.DocumentPublishedPage })),
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
