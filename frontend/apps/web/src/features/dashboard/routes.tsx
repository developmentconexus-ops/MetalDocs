import type { RouteObject } from 'react-router-dom';

export const dashboardRoutes: RouteObject[] = [
  {
    index: true,
    lazy: () => import('./pages/DashboardPage'),
  },
];
