import type { RouteObject } from "react-router-dom";

export const templatesRoutes: RouteObject[] = [
  {
    path: "templates",
    handle: { workspaceView: "templates" },
    lazy: () => import("./pages/TemplatesListRoutePage"),
  },
  {
    path: "templates/new",
    handle: { workspaceView: "templates" },
    lazy: () => import("./pages/TemplateWizardPage"),
  },
  {
    // ADR 0013 coherence: the editor URL mirrors documents (`documents/:documentId/edit`)
    // and never leaks the internal 1-based version_number. The working version is
    // resolved from the template's latest_version inside the route page. Humans only
    // ever see the REV{nn} label (revision_number); version_number stays internal
    // (storage keys, FK, audit, API endpoints) exactly as documents keeps its content hash.
    path: "templates/:templateId/edit",
    handle: { workspaceView: "templates", editMode: true },
    lazy: () => import("./pages/TemplateEditorRoutePage"),
  },
];
