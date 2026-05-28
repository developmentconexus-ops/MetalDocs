import { describe, it, expect, beforeAll, afterAll, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { RouterProvider, createMemoryRouter } from 'react-router-dom';
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

function makeUser(roles: UserRole[]): CurrentUser {
  return {
    userId: '1',
    tenantId: 'ffffffff-ffff-ffff-ffff-ffffffffffff',
    tenantName: 'System Tenant',
    username: 'u',
    email: 'u@example.com',
    displayName: 'U',
    mustChangePassword: false,
    roles,
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
          {
            path: 'admin',
            element: <div>Admin Page</div>,
            handle: { requiresAdmin: true },
          },
        ],
      },
    ],
    { initialEntries: [path] },
  );
  return render(<RouterProvider router={router} />);
}

describe('AppShell admin route guard', () => {
  beforeEach(() => {
    useAuthStore.setState({ authState: 'ready' });
  });

  it('renders an admin route for a system_admin user', () => {
    useAuthStore.setState({ user: makeUser(['system_admin']) });
    renderAt('/admin');
    expect(screen.getByText('Admin Page')).toBeTruthy();
  });

  it('redirects a non-admin away from an admin route', async () => {
    useAuthStore.setState({ user: makeUser(['viewer']) });
    renderAt('/admin');
    await waitFor(() => expect(screen.getByText('Dashboard')).toBeTruthy());
    expect(screen.queryByText('Admin Page')).toBeNull();
  });

  it('renders a non-admin route for a non-admin user', () => {
    useAuthStore.setState({ user: makeUser(['viewer']) });
    renderAt('/');
    expect(screen.getByText('Dashboard')).toBeTruthy();
  });
});
