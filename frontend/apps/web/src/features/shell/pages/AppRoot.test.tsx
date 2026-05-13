import { describe, it, expect, vi, beforeEach } from 'vitest';
import { act, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { AppRoot } from './AppRoot';
import { useAuthStore } from '../../../store/auth.store';
import type { AuthExpiredDetail } from '../../../lib/api';

// Prevent openapi-fetch from being resolved in the test environment.
vi.mock('../../../lib/api/client', () => ({ request: vi.fn(), default: vi.fn() }));

// Mock the auth API — the module under test calls me() on mount.
vi.mock('../../auth/api/auth', () => ({
  me: vi.fn(),
}));

const authBusMock = vi.hoisted(() => ({
  handler: undefined as ((detail: AuthExpiredDetail) => void) | undefined,
}));

// onAuthExpired is used in AppRoot; mock the api barrel so we control it.
vi.mock('../../../lib/api', () => ({
  onAuthExpired: vi.fn((handler: (detail: AuthExpiredDetail) => void) => {
    authBusMock.handler = handler;
    return () => {
      authBusMock.handler = undefined;
    };
  }),
}));

import * as authApi from '../../auth/api/auth';

function wrapper({ children }: { children: React.ReactNode }) {
  return (
    <MemoryRouter>
      <Routes>
        <Route path="/login" element={<div>Login Page</div>} />
        <Route path="/" element={children} />
      </Routes>
    </MemoryRouter>
  );
}

describe('AppRoot', () => {
  beforeEach(() => {
    useAuthStore.setState({ authState: 'loading', user: null });
    authBusMock.handler = undefined;
    sessionStorage.clear();
  });

  it('shows spinner while loading', () => {
    vi.mocked(authApi.me).mockImplementation(() => new Promise(() => {}));
    render(<AppRoot />, { wrapper });
    expect(screen.getByRole('status')).toBeTruthy();
  });

  it('redirects to /login when me() returns 401', async () => {
    vi.mocked(authApi.me).mockRejectedValue(
      Object.assign(new Error('unauth'), { status: 401 }),
    );
    render(<AppRoot />, { wrapper });
    await waitFor(() => {
      expect(screen.getByText('Login Page')).toBeTruthy();
    });
  });

  it('stores returnTo on auth:expired event', async () => {
    vi.mocked(authApi.me).mockResolvedValue({
      userId: '1',
      username: 'admin',
      displayName: 'Admin',
      mustChangePassword: false,
      roles: [],
    });

    render(<AppRoot />, { wrapper });
    await waitFor(() => expect(authBusMock.handler).toBeTypeOf('function'));

    act(() => authBusMock.handler?.({ returnTo: '/documents/abc' }));

    expect(sessionStorage.getItem('auth:returnTo')).toBe('/documents/abc');
  });

  it('ignores root path and login paths on auth:expired event', async () => {
    vi.mocked(authApi.me).mockResolvedValue({
      userId: '1',
      username: 'admin',
      displayName: 'Admin',
      mustChangePassword: false,
      roles: [],
    });

    render(<AppRoot />, { wrapper });
    await waitFor(() => expect(authBusMock.handler).toBeTypeOf('function'));

    act(() => authBusMock.handler?.({ returnTo: '/' }));
    expect(sessionStorage.getItem('auth:returnTo')).toBeNull();

    act(() => authBusMock.handler?.({ returnTo: '/login' }));
    expect(sessionStorage.getItem('auth:returnTo')).toBeNull();
  });
});
