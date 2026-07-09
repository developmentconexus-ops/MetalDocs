import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { createMemoryRouter, RouterProvider } from 'react-router-dom';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

// jsdom's AbortSignal is not an instance of undici's, so react-router's
// client-side navigation Request construction throws (see AppShell.test.tsx
// for the same workaround). Drop the signal in test.
const RealRequest = globalThis.Request;
beforeAll(() => {
  globalThis.Request = class extends RealRequest {
    constructor(input: RequestInfo | URL, init?: RequestInit) {
      if (init && 'signal' in init) {
        const { signal: _signal, ...rest } = init;
        super(input, rest);
      } else {
        super(input, init);
      }
    }
  } as typeof Request;
});
afterAll(() => {
  globalThis.Request = RealRequest;
});

// Heavy pages are stubbed so this suite exercises route *wiring* (path →
// element/redirect resolution) without pulling the editor/query stacks those
// pages own — matching the pattern used by AppRouter.test.tsx (route-config
// assertions, not full page mounts).
vi.mock('./pages/DocumentWorkspacePage', () => ({
  DocumentWorkspacePage: () => <div data-testid="workspace-page" />,
}));

vi.mock('./pages/DocumentDetailRoute', () => ({
  DocumentDetailRoute: () => <div data-testid="detail-route" />,
}));

vi.mock('./pages/DocumentDistributionPage', () => ({
  DocumentDistributionPage: () => <div data-testid="distribution-page" />,
}));

import { documentsRoutes } from './routes';
import { approvalRoutes } from '../approval/routes';

function renderRouter(initialPath: string) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const router = createMemoryRouter([...documentsRoutes, ...approvalRoutes], {
    initialEntries: [initialPath],
  });
  render(
    <QueryClientProvider client={qc}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
  return router;
}

describe('F2d.5 S3 — single-artifact route flip', () => {
  it('renders DocumentDetailLayout (record surface) at /documents/:id/details', async () => {
    renderRouter('/documents/doc-1/details');

    await waitFor(() => {
      expect(screen.getByTestId('detail-route')).toBeInTheDocument();
    });
    expect(screen.queryByTestId('workspace-page')).not.toBeInTheDocument();
  });

  it('renders DocumentWorkspacePage (not the detail layout) at /documents/:id', async () => {
    renderRouter('/documents/doc-1');

    await waitFor(() => {
      expect(screen.getByTestId('workspace-page')).toBeInTheDocument();
    });
    expect(screen.queryByTestId('detail-route')).not.toBeInTheDocument();
  });

  it('redirects /documents/:id/edit to /documents/:id, preserving the param', async () => {
    const router = renderRouter('/documents/doc-42/edit');

    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/documents/doc-42');
    });
  });

  it('redirects /approvals/:documentId to /documents/:documentId, preserving the query string', async () => {
    const router = renderRouter('/approvals/doc-7?decision=approve');

    await waitFor(() => {
      expect(router.state.location.pathname).toBe('/documents/doc-7');
    });
    expect(router.state.location.search).toBe('?decision=approve');
  });
});
