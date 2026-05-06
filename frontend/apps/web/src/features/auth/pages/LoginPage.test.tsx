import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { LoginPage } from './LoginPage';
import { useAuthStore } from '../../../store/auth.store';
import { useUiStore } from '../../../store/ui.store';

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

describe('LoginPage', () => {
  beforeEach(() => {
    useAuthStore.setState({ authState: 'idle', user: null });
    useUiStore.setState({ error: '', message: '' });
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
        userId: '1', username: 'admin', email: 'a@b.com',
        displayName: 'Admin', mustChangePassword: true, roles: [],
      },
    });
    renderPage();
    expect(screen.getByText(/alterar senha/i)).toBeTruthy();
    expect(screen.queryByLabelText(/identificador/i)).toBeNull();
  });
});
