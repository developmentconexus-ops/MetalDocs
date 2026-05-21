import type { RouteObject } from "react-router-dom";

export const controlledDocumentsRoutes: RouteObject[] = [
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

