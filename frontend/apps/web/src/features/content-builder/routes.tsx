import type { RouteObject } from "react-router-dom";

export const contentBuilderRoutes: RouteObject[] = [
  {
    path: "content-builder",
    handle: { workspaceView: "content-builder" },
    lazy: () => import("./pages/ContentBuilderPage"),
  },
];
