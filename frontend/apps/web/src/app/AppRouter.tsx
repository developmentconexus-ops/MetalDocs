import { Navigate, createBrowserRouter } from "react-router-dom";
import { approvalRoutes } from "../features/approval/routes";
import { auditRoutes } from "../features/audit/routes";
import { authRoutes } from "../features/auth/routes";
import { contentBuilderRoutes } from "../features/content-builder/routes";
import { documentsRoutes } from "../features/documents/routes";
import { iamRoutes } from "../features/iam/routes";
import { notificationsRoutes } from "../features/notifications/routes";
import { operationsRoutes } from "../features/operations/routes";
import { passwordChangeRoutes } from "../features/password-change/routes";
import { registryRoutes } from "../features/registry/routes";
import { Component as WorkspaceRoot } from "../features/shell/pages/WorkspaceRoot";
import { taxonomyRoutes } from "../features/taxonomy/routes";
import { templatesRoutes } from "../features/templates/routes";

export const router = createBrowserRouter([
  {
    element: <WorkspaceRoot />,
    children: [
      ...operationsRoutes,
      ...documentsRoutes,
      ...templatesRoutes,
      ...registryRoutes,
      ...taxonomyRoutes,
      ...iamRoutes,
      ...approvalRoutes,
      ...notificationsRoutes,
      ...contentBuilderRoutes,
      ...auditRoutes,
      ...authRoutes,
      ...passwordChangeRoutes,
      {
        path: "*",
        element: <Navigate to="/" replace />,
      },
    ],
  },
]);
