import type { RouteObject } from "react-router-dom";

export const taxonomyRoutes: RouteObject[] = [
  {
    path: "admin/taxonomy",
    handle: { workspaceView: "taxonomy-admin", requiresAdmin: true },
    lazy: () => import("./pages/TaxonomyAdminRoutePage"),
  },
];
