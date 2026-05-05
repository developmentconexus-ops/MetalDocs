import type { RouteObject } from "react-router-dom";

export const notificationsRoutes: RouteObject[] = [
  {
    path: "notifications",
    handle: { workspaceView: "notifications" },
    lazy: () => import("./pages/NotificationsPage"),
  },
];
