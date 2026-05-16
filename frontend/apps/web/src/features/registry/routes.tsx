import type { RouteObject } from "react-router-dom";

export const registryRoutes: RouteObject[] = [
  {
    path: "registry",
    handle: { workspaceView: "registry" },
    lazy: () => import("./pages/RegistryExplorerPage"),
  },
  {
    path: "registry/*",
    handle: { workspaceView: "registry" },
    lazy: () => import("./pages/RegistryExplorerPage"),
  },
  {
    path: "controlled-documents",
    handle: { workspaceView: "controlled-documents" },
    lazy: () => import("./pages/ControlledDocumentsPage"),
  },
  {
    path: "controlled-documents/:controlledDocumentId",
    handle: { workspaceView: "controlled-documents" },
    lazy: () => import("./pages/ControlledDocumentsPage"),
  },
];
