import type { RouteObject } from "react-router-dom";

export const templatesRoutes: RouteObject[] = [
  {
    path: "templates",
    handle: { workspaceView: "templates" },
    lazy: () => import("./pages/TemplatesRedirectPage"),
  },
  {
    path: "templates/*",
    handle: { workspaceView: "templates" },
    lazy: () => import("./pages/TemplatesRedirectPage"),
  },
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
    path: "templates/:templateId/versions/:versionNum",
    handle: { workspaceView: "templates", editMode: true },
    lazy: () => import("./pages/TemplateEditorRoutePage"),
  },
];
