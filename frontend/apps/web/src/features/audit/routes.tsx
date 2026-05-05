import type { RouteObject } from "react-router-dom";

export const auditRoutes: RouteObject[] = [
  {
    path: "audit",
    handle: { workspaceView: "audit" },
    lazy: () => import("./pages/AuditPage"),
  },
];
