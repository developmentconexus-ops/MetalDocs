import type { RouteObject } from 'react-router-dom';

export const tokenRoutes: RouteObject[] = [
  {
    path: 'templates/tokens',
    handle: { workspaceView: 'templates', requiresCapability: 'token.view' },
    lazy: () => import('./pages/TokensRoutePage'),
  },
];
