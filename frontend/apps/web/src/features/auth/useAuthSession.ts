import { useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import * as api from './api/auth';
import { useAuthStore } from '../../store/auth.store';
import { useUiStore } from '../../store/ui.store';
import { asMessage, statusOf } from '../shared/errors';

// useAuthSession — login / logout / change-password.
// Bootstrap (me() on mount) is handled by AppRoot.
// All server cache invalidation happens via queryClient.clear() on logout/expiry.
export function useAuthSession() {
  const queryClient = useQueryClient();
  const {
    loginForm, passwordForm,
    setAuthState, setUser, setLoginForm, setPasswordForm,
  } = useAuthStore();
  const { setError, setMessage } = useUiStore();

  const handleLogin = useCallback(
    async (e: React.FormEvent<HTMLFormElement>) => {
      e.preventDefault();
      setError('');
      try {
        setAuthState('loading');
        const response = await api.login(loginForm);
        setUser(response.user);
        setAuthState('ready');
        if (!response.user.mustChangePassword) {
          const returnTo = sessionStorage.getItem('auth:returnTo');
          if (returnTo) {
            sessionStorage.removeItem('auth:returnTo');
            window.history.pushState({}, '', returnTo);
            window.dispatchEvent(new PopStateEvent('popstate'));
          }
        }
      } catch (err) {
        setUser(null);
        setAuthState('idle');
        setError(statusOf(err) === 401 ? 'Usuário ou senha inválidos.' : asMessage(err));
      }
    },
    [loginForm, setAuthState, setError, setUser],
  );

  const handleLogout = useCallback(async () => {
    try {
      await api.logout();
    } catch {
      // best-effort
    } finally {
      setUser(null);
      setAuthState('idle');
      queryClient.clear();
    }
  }, [queryClient, setAuthState, setUser]);

  const handleChangePassword = useCallback(
    async (e: React.FormEvent<HTMLFormElement>) => {
      e.preventDefault();
      setError('');
      setMessage('');
      if (passwordForm.newPassword !== passwordForm.confirmPassword) {
        setError('A confirmação da nova senha não confere.');
        return;
      }
      try {
        const response = await api.changePassword(passwordForm);
        setPasswordForm({ currentPassword: '', newPassword: '', confirmPassword: '' });
        setUser(response.user);
        setAuthState('ready');
        setMessage('Senha alterada com sucesso.');
      } catch (err) {
        setError(statusOf(err) === 401 ? 'Senha atual incorreta.' : asMessage(err));
      }
    },
    [passwordForm, setAuthState, setError, setMessage, setPasswordForm, setUser],
  );

  return {
    loginForm,
    passwordForm,
    setLoginForm,
    setPasswordForm,
    handleLogin,
    handleLogout,
    handleChangePassword,
  };
}
