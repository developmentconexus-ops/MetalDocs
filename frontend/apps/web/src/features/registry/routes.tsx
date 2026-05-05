import type { RouteObject } from "react-router-dom";

export const registryRoutes: RouteObject[] = [
  {
    path: "registry",
    handle: { workspaceView: "registry" },
    lazy: () => import("./pages/RegistryExplorerPage"),
  },
  {
    path: "registry/*",
    handle: { workspaceView: "registry" },
    lazy: () => import("./pages/RegistryExplorerPage"),
  },
  {
    path: "registry-v2",
    handle: { workspaceView: "registry-v2" },
    lazy: () => import("./pages/RegistryV2Page"),
  },
  {
    path: "registry-v2/:controlledDocumentId",
    handle: { workspaceView: "registry-v2" },
    lazy: () => import("./pages/RegistryV2Page"),
  },
];
