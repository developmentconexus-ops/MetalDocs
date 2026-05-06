import type { RouteObject } from "react-router-dom";

export const authRoutes: RouteObject[] = [
  {
    path: "auth",
    lazy: () => import("./pages/AuthRoutePage"),
  },
  // /login is handled as a public route in AppRouter.tsx — no entry here.
];
