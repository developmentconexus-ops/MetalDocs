import { useEffect } from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import * as authApi from '../../auth/api/auth';
import { onAuthExpired } from '../../../lib/api';
import { useAuthStore } from '../../../store/auth.store';
import { statusOf } from '../../shared/errors';

function FullPageSpinner() {
  return (
    <div
      role="status"
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        height: '100vh',
        background: 'var(--bg)',
        color: 'var(--text-muted)',
        fontSize: 13,
        fontFamily: 'var(--font-sans)',
      }}
    >
      Carregando…
    </div>
  );
}

export function AppRoot() {
  const authState = useAuthStore((s) => s.authState);
  const setAuthState = useAuthStore((s) => s.setAuthState);
  const setUser = useAuthStore((s) => s.setUser);

  // Bootstrap: call me() once on mount to hydrate auth state.
  useEffect(() => {
    async function bootstrap() {
      try {
        const user = await authApi.me();
        setUser(user);
        setAuthState('ready');
      } catch (err) {
        setAuthState(statusOf(err) === 401 ? 'idle' : 'error');
      }
    }
    void bootstrap();
  }, [setAuthState, setUser]);

  // Listen for 401 events from any API call (token expiry / forced logout).
  useEffect(() => {
    return onAuthExpired(({ returnTo }) => {
      if (returnTo && returnTo !== '/' && !returnTo.startsWith('/login')) {
        sessionStorage.setItem('auth:returnTo', returnTo);
      }
      setUser(null);
      setAuthState('idle');
    });
  }, [setAuthState, setUser]);

  if (authState === 'loading') return <FullPageSpinner />;
  if (authState === 'idle') return <Navigate to="/login" replace />;
  if (authState === 'error') {
    return (
      <div style={{ padding: 40, color: 'var(--danger)', fontFamily: 'var(--font-sans)' }}>
        Erro ao carregar sessão. <a href="/">Tentar novamente</a>
      </div>
    );
  }

  return <Outlet />;
}
