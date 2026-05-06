import { Navigate } from 'react-router-dom';
import { useAuthStore } from '../../../store/auth.store';
import { useUiStore } from '../../../store/ui.store';
import { useAuthSession } from '../useAuthSession';
import styles from './LoginPage.module.css';

function LeftPanel() {
  return (
    <div className={styles.left}>
      <div className={styles.leftGrid} aria-hidden="true" />
      <div className={styles.leftBody}>
        <h1 className={styles.headline}>Auditoria</h1>
      </div>
    </div>
  );
}

function LoginForm() {
  const authState = useAuthStore((state) => state.authState);
  const error = useUiStore((state) => state.error);
  const { loginForm, setLoginForm, handleLogin } = useAuthSession();
  const isLoading = authState === 'loading';

  return (
    <div className={styles.formCard}>
      {error && (
        <div className={styles.error} role="alert">
          {error}
        </div>
      )}
      <form onSubmit={handleLogin} noValidate>
        <label className={styles.fieldLabel} htmlFor="identifier">
          Identificador
        </label>
        <input
          id="identifier"
          className={styles.input}
          type="text"
          autoComplete="username"
          value={loginForm.identifier}
          onChange={(event) => setLoginForm((current) => ({ ...current, identifier: event.target.value }))}
          disabled={isLoading}
          required
        />

        <label className={styles.fieldLabel} htmlFor="password">
          Senha
        </label>
        <input
          id="password"
          className={styles.input}
          type="password"
          autoComplete="current-password"
          value={loginForm.password}
          onChange={(event) => setLoginForm((current) => ({ ...current, password: event.target.value }))}
          disabled={isLoading}
          required
        />

        <button type="submit" className={styles.submitBtn} disabled={isLoading}>
          {isLoading ? 'Entrando' : 'Entrar'}
        </button>
      </form>
    </div>
  );
}

function PasswordChangeForm() {
  const error = useUiStore((state) => state.error);
  const { passwordForm, setPasswordForm, handleChangePassword } = useAuthSession();

  return (
    <div className={styles.formCard}>
      <h2 className={styles.formTitle}>Alterar senha</h2>
      {error && (
        <div className={styles.error} role="alert">
          {error}
        </div>
      )}
      <form onSubmit={handleChangePassword} noValidate>
        <label className={styles.fieldLabel} htmlFor="currentPassword">
          Senha atual
        </label>
        <input
          id="currentPassword"
          className={styles.input}
          type="password"
          value={passwordForm.currentPassword}
          onChange={(event) =>
            setPasswordForm((current) => ({ ...current, currentPassword: event.target.value }))
          }
          required
        />

        <label className={styles.fieldLabel} htmlFor="newPassword">
          Nova senha
        </label>
        <input
          id="newPassword"
          className={styles.input}
          type="password"
          value={passwordForm.newPassword}
          onChange={(event) => setPasswordForm((current) => ({ ...current, newPassword: event.target.value }))}
          required
        />

        <label className={styles.fieldLabel} htmlFor="confirmPassword">
          Confirmar nova senha
        </label>
        <input
          id="confirmPassword"
          className={styles.input}
          type="password"
          value={passwordForm.confirmPassword}
          onChange={(event) =>
            setPasswordForm((current) => ({ ...current, confirmPassword: event.target.value }))
          }
          required
        />

        <button type="submit" className={styles.submitBtn}>
          Atualizar senha
        </button>
      </form>
    </div>
  );
}

export function LoginPage() {
  const authState = useAuthStore((state) => state.authState);
  const user = useAuthStore((state) => state.user);

  if (authState === 'ready' && user && !user.mustChangePassword) {
    return <Navigate to="/" replace />;
  }

  return (
    <main className={styles.page}>
      <LeftPanel />
      <section className={styles.right}>
        {authState === 'ready' && user?.mustChangePassword ? <PasswordChangeForm /> : <LoginForm />}
      </section>
    </main>
  );
}
