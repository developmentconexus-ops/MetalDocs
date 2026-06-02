import type { RouteObject } from "react-router-dom";

export const iamRoutes: RouteObject[] = [
  {
    path: "admin",
    handle: { workspaceView: "admin", requiresCapability: "user.view" },
    lazy: () => import("./pages/AdminCenterPage"),
  },
  {
    path: "admin/memberships",
    handle: { workspaceView: "iam-memberships", requiresCapability: "membership.view" },
    lazy: () => import("./pages/AreaMembershipAdminRoutePage"),
  },
];
