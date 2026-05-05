import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { AppRoot } from './AppRoot';
import { useAuthStore } from '../../../store/auth.store';

// Prevent openapi-fetch from being resolved in the test environment.
vi.mock('../../../lib/api/client', () => ({ request: vi.fn(), default: vi.fn() }));

// Mock the auth API — the module under test calls me() on mount.
vi.mock('../../auth/api/auth', () => ({
  me: vi.fn(),
}));

// onAuthExpired is used in AppRoot; mock the api barrel so we control it.
vi.mock('../../../lib/api', () => ({
  onAuthExpired: vi.fn(() => () => {}),
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
});
