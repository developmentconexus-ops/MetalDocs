import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { LoginPage } from './LoginPage';
import { useAuthStore } from '../../../store/auth.store';
import { useUiStore } from '../../../store/ui.store';

const READY_USER = {
  userId: '1',
  tenantId: 'ffffffff-ffff-ffff-ffff-ffffffffffff',
  tenantName: 'System Tenant',
  username: 'admin',
  email: 'a@b.com',
  displayName: 'Admin',
  mustChangePassword: false,
  roles: [],
};

vi.mock('../../../lib/api/client', () => ({ request: vi.fn(), default: vi.fn() }));
vi.mock('../../../lib/api', () => ({ onAuthExpired: vi.fn(() => () => {}) }));

vi.mock('../useAuthSession', () => ({
  useAuthSession: () => ({
    loginForm: { identifier: '', password: '' },
    passwordForm: { currentPassword: '', newPassword: '', confirmPassword: '' },
    setLoginForm: vi.fn(),
    setPasswordForm: vi.fn(),
    handleLogin: vi.fn((e: React.FormEvent) => e.preventDefault()),
    handleChangePassword: vi.fn((e: React.FormEvent) => e.preventDefault()),
    handleLogout: vi.fn(),
  }),
}));

function renderPage() {
  return render(
    <MemoryRouter initialEntries={['/login']}>
      <LoginPage />
    </MemoryRouter>,
  );
}

function renderAuthedAt(target: string) {
  return render(
    <MemoryRouter initialEntries={['/login']}>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="*" element={<div>landed:{target}</div>} />
      </Routes>
    </MemoryRouter>,
  );
}

describe('LoginPage', () => {
  beforeEach(() => {
    useAuthStore.setState({ authState: 'idle', user: null });
    useUiStore.setState({ error: '', message: '' });
    sessionStorage.clear();
  });

  it('renders left panel with brand tagline', () => {
    renderPage();
    expect(screen.getByText(/auditoria/i)).toBeTruthy();
  });

  it('renders identifier and password inputs', () => {
    renderPage();
    expect(screen.getByLabelText(/identificador/i)).toBeTruthy();
    expect(screen.getByLabelText(/senha/i)).toBeTruthy();
  });

  it('renders submit button', () => {
    renderPage();
    expect(screen.getByRole('button', { name: /entrar/i })).toBeTruthy();
  });

  it('keeps the public login form enabled before shell bootstrap runs', () => {
    useAuthStore.setState(useAuthStore.getInitialState(), true);
    renderPage();
    expect(screen.getByRole('button', { name: /entrar/i })).toBeEnabled();
  });

  it('disables submit button when loading', () => {
    useAuthStore.setState({ authState: 'loading', user: null });
    renderPage();
    expect(screen.getByRole('button', { name: /entrando/i })).toBeDisabled();
  });

  it('shows error message when error is set', () => {
    useUiStore.setState({ error: 'Usuário ou senha inválidos.' });
    renderPage();
    expect(screen.getByText('Usuário ou senha inválidos.')).toBeTruthy();
  });

  it('shows password change form when user mustChangePassword', () => {
    useAuthStore.setState({
      authState: 'ready',
      user: {
        userId: '1', tenantId: 'ffffffff-ffff-ffff-ffff-ffffffffffff', tenantName: 'System Tenant', username: 'admin', email: 'a@b.com',
        displayName: 'Admin', mustChangePassword: true, roles: [],
      },
    });
    renderPage();
    expect(screen.getByText(/alterar senha/i)).toBeTruthy();
    expect(screen.queryByLabelText(/identificador/i)).toBeNull();
  });

  it('redirects an authenticated user to the captured returnTo path', () => {
    sessionStorage.setItem('auth:returnTo', '/documents');
    useAuthStore.setState({ authState: 'ready', user: READY_USER });
    renderAuthedAt('/documents');
    expect(screen.getByText('landed:/documents')).toBeTruthy();
  });

  it('redirects to / when no returnTo was captured', () => {
    useAuthStore.setState({ authState: 'ready', user: READY_USER });
    renderAuthedAt('/');
    expect(screen.getByText('landed:/')).toBeTruthy();
  });

  it('ignores an unsafe (protocol-relative) returnTo and falls back to /', () => {
    sessionStorage.setItem('auth:returnTo', '//evil.example.com');
    useAuthStore.setState({ authState: 'ready', user: READY_USER });
    renderAuthedAt('/');
    expect(screen.getByText('landed:/')).toBeTruthy();
  });

  it('clears auth:returnTo from sessionStorage after consuming it', () => {
    sessionStorage.setItem('auth:returnTo', '/documents');
    useAuthStore.setState({ authState: 'ready', user: READY_USER });
    renderAuthedAt('/documents');
    expect(sessionStorage.getItem('auth:returnTo')).toBeNull();
  });
});
