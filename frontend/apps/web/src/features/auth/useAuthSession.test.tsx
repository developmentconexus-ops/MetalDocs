import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import { useAuthSession } from './useAuthSession';
import { useAuthStore } from '../../store/auth.store';
import { useUiStore } from '../../store/ui.store';
import * as authApi from './api/auth';

vi.mock('./api/auth');

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient();
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

function submitEvent() {
  return { preventDefault: vi.fn() } as unknown as React.FormEvent<HTMLFormElement>;
}

describe('useAuthSession.handleChangePassword', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState(useAuthStore.getInitialState(), true);
    useUiStore.setState({ error: '', message: '' });
    useAuthStore.setState({
      passwordForm: { currentPassword: 'wrong', newPassword: 'NewPass123!', confirmPassword: 'NewPass123!' },
    });
  });

  it('maps AUTH_INVALID_CREDENTIALS (wrong current password) to a credential message', async () => {
    vi.mocked(authApi.changePassword).mockRejectedValue(
      Object.assign(new Error('Invalid username/email or password'), {
        code: 'AUTH_INVALID_CREDENTIALS',
        status: 401,
      }),
    );

    const { result } = renderHook(() => useAuthSession(), { wrapper });
    await act(async () => {
      await result.current.handleChangePassword(submitEvent());
    });

    expect(useUiStore.getState().error).toBe('Senha atual incorreta.');
  });

  it('does NOT map a genuine session-expiry (authn.expired) to the credential message', async () => {
    vi.mocked(authApi.changePassword).mockRejectedValue(
      Object.assign(new Error('Sessão expirada'), { code: 'authn.expired', status: 401 }),
    );

    const { result } = renderHook(() => useAuthSession(), { wrapper });
    await act(async () => {
      await result.current.handleChangePassword(submitEvent());
    });

    expect(useUiStore.getState().error).toBe('Sessão expirada');
  });

  it('blocks submission when confirmation does not match', async () => {
    useAuthStore.setState({
      passwordForm: { currentPassword: 'x', newPassword: 'NewPass123!', confirmPassword: 'different' },
    });

    const { result } = renderHook(() => useAuthSession(), { wrapper });
    await act(async () => {
      await result.current.handleChangePassword(submitEvent());
    });

    expect(useUiStore.getState().error).toBe('A confirmação da nova senha não confere.');
    expect(authApi.changePassword).not.toHaveBeenCalled();
  });
});
