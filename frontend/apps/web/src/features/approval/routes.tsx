import type { RouteObject } from "react-router-dom";

export const approvalRoutes: RouteObject[] = [
  {
    path: "approvals",
    handle: { workspaceView: "approvals" },
    lazy: () => import("./pages/InboxPage").then((module) => ({ Component: module.InboxPage })),
  },
  {
    path: "approvals/:documentId",
    handle: { workspaceView: "approvals" },
    lazy: () =>
      import("./pages/SignoffDetailPage").then((module) => ({
        Component: module.SignoffDetailPage,
      })),
  },
  {
    path: "approval-routes",
    handle: { workspaceView: "approval-routes", requiresCapability: "route.manage" },
    lazy: () =>
      import("./pages/route-admin/RouteAdminPage").then((module) => ({
        Component: module.RouteAdminPage,
      })),
  },
];
