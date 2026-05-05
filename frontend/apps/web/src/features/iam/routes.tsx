import type { RouteObject } from "react-router-dom";

export const iamRoutes: RouteObject[] = [
  {
    path: "admin",
    handle: { workspaceView: "admin", requiresAdmin: true },
    lazy: () => import("./pages/AdminCenterPage"),
  },
  {
    path: "admin/memberships",
    handle: { workspaceView: "iam-memberships", requiresAdmin: true },
    lazy: () => import("./pages/AreaMembershipAdminRoutePage"),
  },
];
