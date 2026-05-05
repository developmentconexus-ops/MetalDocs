import type { RouteObject } from "react-router-dom";

export const operationsRoutes: RouteObject[] = [
  {
    index: true,
    handle: { workspaceView: "operations" },
    lazy: () => import("./pages/OperationsPage"),
  },
  {
    path: "operations",
    handle: { workspaceView: "operations" },
    lazy: () => import("./pages/OperationsPage"),
  },
];
