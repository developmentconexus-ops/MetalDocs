import type { RouteObject } from "react-router-dom";

export const templatesRoutes: RouteObject[] = [
  {
    path: "templates",
    handle: { workspaceView: "templates-v2" },
    lazy: () => import("./pages/TemplatesRedirectPage"),
  },
  {
    path: "templates/*",
    handle: { workspaceView: "templates-v2" },
    lazy: () => import("./pages/TemplatesRedirectPage"),
  },
  {
    path: "templates-v2",
    handle: { workspaceView: "templates-v2" },
    lazy: () => import("./pages/TemplatesListRoutePage"),
  },
  {
    path: "templates-v2/novo",
    handle: { workspaceView: "templates-v2" },
    lazy: () => import("./pages/TemplateWizardPage"),
  },
  {
    path: "templates-v2/:templateId/versions/:versionNum",
    handle: { workspaceView: "templates-v2", editMode: true },
    lazy: () => import("./pages/TemplateAuthorRoutePage"),
  },
];
