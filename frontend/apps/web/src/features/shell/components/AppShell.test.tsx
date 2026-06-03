import { describe, it, expect, beforeAll, afterAll, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { Outlet, RouterProvider, createMemoryRouter } from 'react-router-dom';
import { AppShell } from './AppShell';
import { useAuthStore } from '../../../store/auth.store';
import type { CurrentUser, UserRole } from '../../../lib/types';

// Isolate the guard from the live chrome — children pull in heavy modules.
vi.mock('./Rail', () => ({ Rail: () => <nav data-testid="rail" /> }));
vi.mock('./AppToolbar', () => ({ AppToolbar: () => <div data-testid="toolbar" /> }));
vi.mock('./SectionPanel', () => ({ SectionPanel: () => <div data-testid="section-panel" /> }));

// jsdom's AbortSignal is not an instance of undici's, so react-router's
// client-side navigation Request construction throws. Drop the signal in test.
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

function makeUser(capabilities: string[], roles: UserRole[] = ['viewer']): CurrentUser {
  return {
    userId: '1',
    tenantId: 'ffffffff-ffff-ffff-ffff-ffffffffffff',
    tenantName: 'System Tenant',
    username: 'u',
    email: 'u@example.com',
    displayName: 'U',
    mustChangePassword: false,
    roles,
    capabilities,
  };
}

function renderAt(path: string) {
  const router = createMemoryRouter(
    [
      {
        path: '/',
        element: <AppShell />,
        children: [
          { index: true, element: <div>Dashboard</div> },
          // Mirrors features/iam/routes.tsx: the Admin Center parent carries a
          // wide requiresAnyCapability and the memberships child tightens it to
          // requiresCapability "membership.view". The orphan top-level
          // admin/memberships route was removed in PR-2 — memberships is now a
          // child of the admin parent, but the deep-link still resolves.
          {
            path: 'admin',
            element: (
              <>
                <div>Admin Page</div>
                <Outlet />
              </>
            ),
            handle: {
              requiresAnyCapability: ['user.view', 'membership.view', 'metrics.view'],
            },
            children: [
              {
                path: 'memberships',
                element: <div>Memberships Page</div>,
                handle: { requiresCapability: 'membership.view' },
              },
            ],
          },
          {
            path: 'multi',
            element: <div>Multi Page</div>,
            handle: { requiresAnyCapability: ['taxonomy.view', 'taxonomy.manage'] },
          },
          // Mirrors features/iam/routes.tsx: parent /center has a wide
          // requiresAnyCapability and child /center/audit tightens it with
          // requiresCapability "audit.read". A caller that satisfies only
          // the parent must NOT reach the child.
          {
            path: 'center',
            handle: {
              requiresAnyCapability: ['user.view', 'membership.view', 'metrics.view'],
            },
            children: [
              {
                path: 'audit',
                element: <div>Audit Page</div>,
                handle: { requiresCapability: 'audit.read' },
              },
            ],
          },
        ],
      },
    ],
    { initialEntries: [path] },
  );
  return render(<RouterProvider router={router} />);
}

describe('AppShell capability route guard', () => {
  beforeEach(() => {
    useAuthStore.setState({ authState: 'ready' });
  });

  it('renders /admin when user holds user.view', () => {
    useAuthStore.setState({ user: makeUser(['user.view']) });
    renderAt('/admin');
    expect(screen.getByText('Admin Page')).toBeTruthy();
  });

  it('redirects /admin to root when capability missing', async () => {
    useAuthStore.setState({ user: makeUser([]) });
    renderAt('/admin');
    await waitFor(() => expect(screen.getByText('Dashboard')).toBeTruthy());
    expect(screen.queryByText('Admin Page')).toBeNull();
  });

  it('redirects /admin/memberships when only user.view granted', async () => {
    useAuthStore.setState({ user: makeUser(['user.view']) });
    renderAt('/admin/memberships');
    await waitFor(() => expect(screen.getByText('Dashboard')).toBeTruthy());
    expect(screen.queryByText('Memberships Page')).toBeNull();
  });

  it('renders /admin/memberships when membership.view granted', () => {
    useAuthStore.setState({ user: makeUser(['membership.view']) });
    renderAt('/admin/memberships');
    expect(screen.getByText('Memberships Page')).toBeTruthy();
  });

  it('renders multi-cap route when any required cap is held', () => {
    useAuthStore.setState({ user: makeUser(['taxonomy.manage']) });
    renderAt('/multi');
    expect(screen.getByText('Multi Page')).toBeTruthy();
  });

  it('redirects multi-cap route when none of the required caps held', async () => {
    useAuthStore.setState({ user: makeUser(['user.view']) });
    renderAt('/multi');
    await waitFor(() => expect(screen.getByText('Dashboard')).toBeTruthy());
    expect(screen.queryByText('Multi Page')).toBeNull();
  });

  it('redirects nested /center/audit when viewer satisfies parent but not child cap', async () => {
    // Viewer-only caps: matches the parent's requiresAnyCapability
    // ("membership.view") but NOT the child's requiresCapability
    // ("audit.read"). Old gate returned on the first matched handle (parent)
    // and let the viewer through; new gate ANDs all collected handles.
    useAuthStore.setState({
      user: makeUser(['document.view', 'membership.view', 'taxonomy.view', 'template.view']),
    });
    renderAt('/center/audit');
    await waitFor(() => expect(screen.getByText('Dashboard')).toBeTruthy());
    expect(screen.queryByText('Audit Page')).toBeNull();
  });

  it('renders nested /center/audit when both parent and child caps held', () => {
    useAuthStore.setState({ user: makeUser(['membership.view', 'audit.read']) });
    renderAt('/center/audit');
    expect(screen.getByText('Audit Page')).toBeTruthy();
  });

  it('renders ungated routes regardless of capabilities', () => {
    useAuthStore.setState({ user: makeUser([]) });
    renderAt('/');
    expect(screen.getByText('Dashboard')).toBeTruthy();
  });
});
