import type { RouteObject } from "react-router-dom";

export const passwordChangeRoutes: RouteObject[] = [
  {
    path: "change-password",
    lazy: () => import("./pages/PasswordChangeRoutePage"),
  },
];
