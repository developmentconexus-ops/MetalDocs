import { Navigate, createBrowserRouter } from 'react-router-dom';
import { AppRoot } from '../features/shell/pages/AppRoot';
import { approvalRoutes } from '../features/approval/routes';
import { auditRoutes } from '../features/audit/routes';
import { contentBuilderRoutes } from '../features/content-builder/routes';
import { documentsRoutes } from '../features/documents/routes';
import { iamRoutes } from '../features/iam/routes';
import { notificationsRoutes } from '../features/notifications/routes';
import { operationsRoutes } from '../features/operations/routes';
import { passwordChangeRoutes } from '../features/password-change/routes';
import { registryRoutes } from '../features/registry/routes';
import { taxonomyRoutes } from '../features/taxonomy/routes';
import { templatesRoutes } from '../features/templates/routes';

export const router = createBrowserRouter([
  // Public routes — no Rail, no Toolbar
  {
    path: '/login',
    lazy: () =>
      import('../features/auth/pages/LoginPage').then((m) => ({ Component: m.LoginPage })),
  },
  // Protected routes — wrapped in AppRoot (auth guard) → AppShell (layout)
  {
    element: <AppRoot />,
    children: [
      {
        lazy: () =>
          import('../features/shell/components/AppShell').then((m) => ({ Component: m.AppShell })),
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
          ...passwordChangeRoutes,
          { path: '*', element: <Navigate to="/" replace /> },
        ],
      },
    ],
  },
]);
