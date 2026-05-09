import type { RouteObject } from "react-router-dom";

export const documentsRoutes: RouteObject[] = [
  {
    path: 'documents',
    handle: { workspaceView: 'library' },
    lazy: () => import('./pages/LibraryPage').then((m) => ({ Component: m.default })),
  },
  {
    path: "documents/all",
    handle: { workspaceView: "library" },
    lazy: () => import("./pages/DocumentsHubPage").then((module) => ({ Component: module.LibraryDocumentsPage })),
  },
  {
    path: "documents/area/:areaCode",
    handle: { workspaceView: "library" },
    lazy: () => import("./pages/DocumentsHubPage").then((module) => ({ Component: module.LibraryDocumentsPage })),
  },
  {
    path: "documents/type/:profileCode",
    handle: { workspaceView: "library" },
    lazy: () => import("./pages/DocumentsHubPage").then((module) => ({ Component: module.LibraryDocumentsPage })),
  },
  {
    path: "documents/doc/:documentId",
    handle: { workspaceView: "library" },
    lazy: () => import("./pages/DocumentsHubPage").then((module) => ({ Component: module.LibraryDocumentsPage })),
  },
  {
    path: "documents/mine",
    handle: { workspaceView: "my-docs" },
    lazy: () => import("./pages/DocumentsHubPage").then((module) => ({ Component: module.MyDocumentsPage })),
  },
  {
    path: "documents/mine/*",
    handle: { workspaceView: "my-docs" },
    lazy: () => import("./pages/DocumentsHubPage").then((module) => ({ Component: module.MyDocumentsPage })),
  },
  {
    path: "documents/recent",
    handle: { workspaceView: "recent" },
    lazy: () => import("./pages/DocumentsHubPage").then((module) => ({ Component: module.RecentDocumentsPage })),
  },
  {
    path: "documents/recent/*",
    handle: { workspaceView: "recent" },
    lazy: () => import("./pages/DocumentsHubPage").then((module) => ({ Component: module.RecentDocumentsPage })),
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
    path: "documents-v2/new",
    handle: { workspaceView: "documents-v2" },
    lazy: () => import("./pages/NewDocumentWizardPage").then((m) => ({ Component: m.NewDocumentWizardPage })),
  },
  {
    path: "documents-v2/:documentID",
    handle: { workspaceView: "documents-v2" },
    lazy: () => import("./pages/DocumentEditorRoutePage"),
  },
];
