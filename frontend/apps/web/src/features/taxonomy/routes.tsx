import type { RouteObject } from "react-router-dom";

export const taxonomyRoutes: RouteObject[] = [
  {
    path: "admin/taxonomy",
    handle: { workspaceView: "taxonomy-admin", requiresCapability: "taxonomy.view" },
    lazy: () => import("./pages/TaxonomyAdminRoutePage"),
  },
];
