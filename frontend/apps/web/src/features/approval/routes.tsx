import type { RouteObject } from "react-router-dom";

export const approvalRoutes: RouteObject[] = [
  {
    path: "approvals",
    handle: { workspaceView: "approvals" },
    lazy: () => import("./pages/InboxPage").then((module) => ({ Component: module.InboxPage })),
  },
  {
    path: "approval-routes",
    handle: { workspaceView: "approval-routes", requiresAdmin: true },
    lazy: () => import("./pages/RouteAdminPage").then((module) => ({ Component: module.RouteAdminPage })),
  },
];
