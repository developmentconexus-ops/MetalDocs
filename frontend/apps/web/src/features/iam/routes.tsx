import type { RouteObject } from "react-router-dom";
import { Navigate } from "react-router-dom";

export const iamRoutes: RouteObject[] = [
  {
    path: "admin",
    handle: {
      workspaceView: "admin",
      requiresAnyCapability: ["user.view", "membership.view", "metrics.view"],
    },
    lazy: () => import("./pages/AdminCenterPage"),
    children: [
      { index: true, element: <Navigate to="overview" replace /> },
      {
        path: "overview",
        handle: { requiresAnyCapability: ["user.view", "membership.view", "metrics.view"] },
        lazy: () => import("./tabs/OverviewTab.route"),
      },
      {
        path: "people",
        handle: { requiresCapability: "user.view" },
        lazy: () => import("./tabs/PeopleTab.route"),
        children: [
          {
            path: ":userId",
            handle: { requiresCapability: "user.view" },
            lazy: () => import("./tabs/PeopleDetailDrawer.route"),
          },
        ],
      },
      {
        path: "roles",
        handle: { requiresCapability: "membership.view" },
        lazy: () => import("./tabs/RolesCapsTab.route"),
      },
      {
        path: "audit",
        handle: { requiresCapability: "audit.read" },
        lazy: () => import("./tabs/AuditTab.route"),
      },
      {
        path: "sessions",
        handle: { requiresCapability: "user.view" },
        lazy: () => import("./tabs/SessionsTab.route"),
      },
      {
        path: "usage",
        handle: { requiresCapability: "metrics.view" },
        lazy: () => import("./tabs/UsageTab.route"),
      },
    ],
  },
  {
    path: "admin/memberships",
    handle: { workspaceView: "iam-memberships", requiresCapability: "membership.view" },
    lazy: () => import("./pages/AreaMembershipAdminRoutePage"),
  },
];
