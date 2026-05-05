import type { RouteObject } from "react-router-dom";

export const authRoutes: RouteObject[] = [
  {
    path: "auth",
    lazy: () => import("./pages/AuthRoutePage"),
  },
  {
    path: "login",
    lazy: () => import("./pages/AuthRoutePage"),
  },
];
