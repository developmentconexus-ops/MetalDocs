# Login Screen Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the LoginPage stub with a production-quality full-page split-screen login that matches the Wine design system, handles the mustChangePassword flow, and redirects already-authenticated users.

**Architecture:** `LoginPage` is a self-contained public route — no AppShell, no Rail. Left panel (45%, `var(--rail)` dark) is purely decorative. Right panel (55%, white) renders one of three views: redirect guard (if already authenticated), password-change form (if `mustChangePassword`), or the login form. All state comes from `auth.store` (session) and `ui.store` (error flash). Form actions call `useAuthSession()` hooks. Zero server queries on this page.

**Tech Stack:** React 18, TypeScript, CSS Modules, Zustand (`auth.store`, `ui.store`), `useAuthSession`, React Router v7 (`Navigate`, `useNavigate`), Vitest + React Testing Library

**Worktree:** `.worktrees/screen-redesign` (branch `feature/screen-redesign`)

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `src/features/auth/pages/LoginPage.tsx` | Replace stub | Full login + password-change page |
| `src/features/auth/pages/LoginPage.module.css` | Create | Split-screen layout + form styles |
| `src/features/auth/pages/LoginPage.test.tsx` | Create | Renders, redirect, form submit, error, loading, mustChangePassword |
| `src/features/auth/routes.tsx` | Modify | Remove duplicate `login` path (conflicts with public route) |
| `frontend/apps/web/design-source/login/NOTES.md` | Create | Design notes for this screen |

---

## Task 1: Remove duplicate login route

The `authRoutes` array has `path: "login"` as a protected child route. This conflicts with the new public `/login` in `AppRouter.tsx`. Remove it now before building the page.

**Files:**
- Modify: `frontend/apps/web/src/features/auth/routes.tsx`

- [ ] **Step 1: Remove login path from authRoutes**

Replace entire `frontend/apps/web/src/features/auth/routes.tsx`:

```tsx
import type { RouteObject } from "react-router-dom";

export const authRoutes: RouteObject[] = [
  {
    path: "auth",
    lazy: () => import("./pages/AuthRoutePage"),
  },
  // /login is handled as a public route in AppRouter.tsx — no entry here.
];
```

- [ ] **Step 2: Commit**

```bash
cd .worktrees/screen-redesign
git add frontend/apps/web/src/features/auth/routes.tsx
git commit -m "fix(auth): remove duplicate /login from protected authRoutes"
```

---

## Task 2: CSS Module — split-screen layout + form styles

**Files:**
- Create: `frontend/apps/web/src/features/auth/pages/LoginPage.module.css`

- [ ] **Step 1: Create LoginPage.module.css**

```css
/* Full-page split: left dark panel + right white form */
.page {
  display: flex;
  height: 100vh;
  overflow: hidden;
  font-family: var(--font-sans);
}

/* === LEFT PANEL === */
.left {
  width: 45%;
  flex-shrink: 0;
  background: var(--rail);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 40px 48px;
}

.leftTop {
  display: flex;
  align-items: center;
  gap: 10px;
}

.logoMark {
  width: 32px;
  height: 32px;
  border-radius: 7px;
  background: var(--brand);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.logoText {
  font-size: 17px;
  font-weight: 600;
  color: var(--rail-text);
  letter-spacing: -0.02em;
}

.leftBody {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding-bottom: 48px;
}

.tagline {
  font-size: 28px;
  font-weight: 600;
  color: var(--rail-text);
  line-height: 1.25;
  letter-spacing: -0.02em;
  max-width: 340px;
}

.taglineSub {
  margin-top: 16px;
  font-size: 14px;
  color: var(--rail-text-muted);
  line-height: 1.55;
  max-width: 300px;
}

.leftBottom {
  font-size: 11px;
  color: var(--rail-text-muted);
}

/* === RIGHT PANEL === */
.right {
  flex: 1;
  background: var(--surface);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
  overflow-y: auto;
}

.formCard {
  width: 100%;
  max-width: 360px;
}

.formHeader {
  margin-bottom: 32px;
}

.formLogoMark {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: var(--brand);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
}

.formTitle {
  font-size: 22px;
  font-weight: 600;
  color: var(--text);
  letter-spacing: -0.02em;
  margin: 0 0 6px;
}

.formSubtitle {
  font-size: 13px;
  color: var(--text-muted);
  margin: 0;
}

/* === FORM FIELDS === */
.field {
  margin-bottom: 14px;
}

.label {
  display: block;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-soft);
  margin-bottom: 5px;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.input {
  width: 100%;
  height: 40px;
  padding: 0 12px;
  border: 1px solid var(--border-strong);
  border-radius: var(--r-2);
  background: var(--surface);
  font-size: 14px;
  font-family: var(--font-sans);
  color: var(--text);
  box-sizing: border-box;
  transition: border-color 120ms, box-shadow 120ms;
}

.input:focus {
  outline: none;
  border-color: var(--brand);
  box-shadow: 0 0 0 3px rgba(107, 31, 42, 0.12);
}

.input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* === ERROR BANNER === */
.error {
  background: var(--danger-bg);
  border: 1px solid var(--danger);
  border-radius: var(--r-2);
  padding: 10px 14px;
  font-size: 13px;
  color: var(--danger);
  margin-bottom: 16px;
}

/* === SUCCESS BANNER === */
.success {
  background: var(--success-bg);
  border: 1px solid var(--success);
  border-radius: var(--r-2);
  padding: 10px 14px;
  font-size: 13px;
  color: var(--success);
  margin-bottom: 16px;
}

/* === SUBMIT BUTTON === */
.submitBtn {
  width: 100%;
  height: 40px;
  border: none;
  border-radius: var(--r-2);
  background: var(--brand);
  color: white;
  font-size: 14px;
  font-weight: 500;
  font-family: var(--font-sans);
  cursor: pointer;
  transition: background 120ms;
  margin-top: 6px;
}

.submitBtn:hover:not(:disabled) {
  background: var(--brand-deep);
}

.submitBtn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/apps/web/src/features/auth/pages/LoginPage.module.css
git commit -m "feat(login): CSS module — split-screen layout + Wine form styles"
```

---

## Task 3: LoginPage implementation + tests

**Files:**
- Replace: `frontend/apps/web/src/features/auth/pages/LoginPage.tsx`
- Create: `frontend/apps/web/src/features/auth/pages/LoginPage.test.tsx`

**State contract:**
- `useAuthStore`: `authState: LoadState`, `user: CurrentUser | null`, `loginForm`, `passwordForm`, `setLoginForm`, `setPasswordForm`
- `useUiStore`: `error: string`, `message: string`
- `useAuthSession()`: `handleLogin`, `handleChangePassword`
- `useNavigate()`: navigate to `/` on successful login (detected by `authState === 'ready'` transition)

---

- [ ] **Step 1: Write failing tests**

Create `frontend/apps/web/src/features/auth/pages/LoginPage.test.tsx`:

```tsx
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { LoginPage } from './LoginPage';
import { useAuthStore } from '../../../store/auth.store';
import { useUiStore } from '../../../store/ui.store';

// Prevent openapi-fetch from being resolved
vi.mock('../../../lib/api/client', () => ({ request: vi.fn(), default: vi.fn() }));
vi.mock('../../../lib/api', () => ({ onAuthExpired: vi.fn(() => () => {}) }));

// Mock useAuthSession — we test only the LoginPage wiring, not the hook logic
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
    expect(screen.getByText(/documentos controlados/i)).toBeTruthy();
  });

  it('renders email and password inputs', () => {
    renderPage();
    expect(screen.getByLabelText(/usuário ou e-mail/i)).toBeTruthy();
    expect(screen.getByLabelText(/senha/i)).toBeTruthy();
  });

  it('renders submit button', () => {
    renderPage();
    expect(screen.getByRole('button', { name: /entrar/i })).toBeTruthy();
  });

  it('disables submit button when loading', () => {
    useAuthStore.setState({ authState: 'loading', user: null });
    renderPage();
    expect(screen.getByRole('button', { name: /entrar/i })).toBeDisabled();
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
    expect(screen.queryByLabelText(/usuário ou e-mail/i)).toBeNull();
  });
});
```

- [ ] **Step 2: Run — expect FAIL**

```bash
cd .worktrees/screen-redesign/frontend/apps/web
npx vitest run src/features/auth/pages/LoginPage.test.tsx --reporter=verbose
```

Expected: `LoginPage is not a module / Cannot find module`

---

- [ ] **Step 3: Implement LoginPage.tsx**

Replace entire `frontend/apps/web/src/features/auth/pages/LoginPage.tsx`:

```tsx
import { Navigate, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../../store/auth.store';
import { useUiStore } from '../../../store/ui.store';
import { useAuthSession } from '../useAuthSession';
import { Icon } from '../../../components/ui/Icon';
import styles from './LoginPage.module.css';

// ──────────────────────────────────────────────
// LEFT DECORATIVE PANEL
// ──────────────────────────────────────────────
function LeftPanel() {
  return (
    <div className={styles.left}>
      <div className={styles.leftTop}>
        <span className={styles.logoMark}>
          <Icon name="docs" size={16} style={{ color: 'white' }} />
        </span>
        <span className={styles.logoText}>MetalDocs</span>
      </div>

      <div className={styles.leftBody}>
        <p className={styles.tagline}>
          Documentos controlados para indústrias sérias.
        </p>
        <p className={styles.taglineSub}>
          Procedimentos, políticas e instruções técnicas — rastreados, aprovados e sempre atualizados.
        </p>
      </div>

      <div className={styles.leftBottom}>
        © {new Date().getFullYear()} MetalDocs
      </div>
    </div>
  );
}

// ──────────────────────────────────────────────
// LOGIN FORM
// ──────────────────────────────────────────────
function LoginForm() {
  const authState = useAuthStore((s) => s.authState);
  const loginForm = useAuthStore((s) => s.loginForm);
  const error = useUiStore((s) => s.error);
  const { handleLogin, setLoginForm } = useAuthSession();
  const isLoading = authState === 'loading';

  return (
    <div className={styles.formCard}>
      <div className={styles.formHeader}>
        <div className={styles.formLogoMark}>
          <Icon name="docs" size={18} style={{ color: 'white' }} />
        </div>
        <h1 className={styles.formTitle}>Bem-vindo de volta</h1>
        <p className={styles.formSubtitle}>Entre com suas credenciais para acessar o acervo.</p>
      </div>

      {error && <div className={styles.error} role="alert">{error}</div>}

      <form onSubmit={handleLogin} noValidate>
        <div className={styles.field}>
          <label className={styles.label} htmlFor="identifier">
            Usuário ou e-mail
          </label>
          <input
            id="identifier"
            className={styles.input}
            type="text"
            autoComplete="username"
            autoFocus
            value={loginForm.identifier}
            onChange={(e) =>
              setLoginForm((f) => ({ ...f, identifier: e.target.value }))
            }
            disabled={isLoading}
            required
          />
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="password">
            Senha
          </label>
          <input
            id="password"
            className={styles.input}
            type="password"
            autoComplete="current-password"
            value={loginForm.password}
            onChange={(e) =>
              setLoginForm((f) => ({ ...f, password: e.target.value }))
            }
            disabled={isLoading}
            required
          />
        </div>

        <button
          type="submit"
          className={styles.submitBtn}
          disabled={isLoading}
        >
          {isLoading ? 'Entrando…' : 'Entrar'}
        </button>
      </form>
    </div>
  );
}

// ──────────────────────────────────────────────
// PASSWORD CHANGE FORM (mustChangePassword flow)
// ──────────────────────────────────────────────
function PasswordChangeForm() {
  const authState = useAuthStore((s) => s.authState);
  const passwordForm = useAuthStore((s) => s.passwordForm);
  const error = useUiStore((s) => s.error);
  const message = useUiStore((s) => s.message);
  const { handleChangePassword, setPasswordForm } = useAuthSession();
  const isLoading = authState === 'loading';

  return (
    <div className={styles.formCard}>
      <div className={styles.formHeader}>
        <div className={styles.formLogoMark}>
          <Icon name="lock" size={18} style={{ color: 'white' }} />
        </div>
        <h1 className={styles.formTitle}>Alterar senha</h1>
        <p className={styles.formSubtitle}>
          Sua conta requer a definição de uma nova senha antes de continuar.
        </p>
      </div>

      {error && <div className={styles.error} role="alert">{error}</div>}
      {message && <div className={styles.success} role="status">{message}</div>}

      <form onSubmit={handleChangePassword} noValidate>
        <div className={styles.field}>
          <label className={styles.label} htmlFor="currentPassword">
            Senha atual
          </label>
          <input
            id="currentPassword"
            className={styles.input}
            type="password"
            autoComplete="current-password"
            value={passwordForm.currentPassword}
            onChange={(e) =>
              setPasswordForm((f) => ({ ...f, currentPassword: e.target.value }))
            }
            disabled={isLoading}
            required
          />
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="newPassword">
            Nova senha
          </label>
          <input
            id="newPassword"
            className={styles.input}
            type="password"
            autoComplete="new-password"
            value={passwordForm.newPassword}
            onChange={(e) =>
              setPasswordForm((f) => ({ ...f, newPassword: e.target.value }))
            }
            disabled={isLoading}
            required
          />
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="confirmPassword">
            Confirmar nova senha
          </label>
          <input
            id="confirmPassword"
            className={styles.input}
            type="password"
            autoComplete="new-password"
            value={passwordForm.confirmPassword}
            onChange={(e) =>
              setPasswordForm((f) => ({ ...f, confirmPassword: e.target.value }))
            }
            disabled={isLoading}
            required
          />
        </div>

        <button
          type="submit"
          className={styles.submitBtn}
          disabled={isLoading}
        >
          {isLoading ? 'Salvando…' : 'Definir nova senha'}
        </button>
      </form>
    </div>
  );
}

// ──────────────────────────────────────────────
// PAGE ROOT
// ──────────────────────────────────────────────
export function LoginPage() {
  const authState = useAuthStore((s) => s.authState);
  const user = useAuthStore((s) => s.user);
  const navigate = useNavigate();

  // Already authenticated + no pending password change → go to app
  if (authState === 'ready' && user && !user.mustChangePassword) {
    const returnTo = sessionStorage.getItem('auth:returnTo') || '/';
    return <Navigate to={returnTo} replace />;
  }

  const rightContent =
    authState === 'ready' && user?.mustChangePassword
      ? <PasswordChangeForm />
      : <LoginForm />;

  return (
    <div className={styles.page}>
      <LeftPanel />
      <div className={styles.right}>
        {rightContent}
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Run tests — expect PASS**

```bash
cd .worktrees/screen-redesign/frontend/apps/web
npx vitest run src/features/auth/pages/LoginPage.test.tsx --reporter=verbose
```

Expected:
```
✓ renders left panel with brand tagline
✓ renders email and password inputs
✓ renders submit button
✓ disables submit button when loading
✓ shows error message when error is set
✓ shows password change form when user mustChangePassword
```
6 tests passed

- [ ] **Step 5: Commit**

```bash
cd .worktrees/screen-redesign
git add frontend/apps/web/src/features/auth/pages/LoginPage.tsx \
         frontend/apps/web/src/features/auth/pages/LoginPage.test.tsx
git commit -m "feat(login): LoginPage — split-screen layout, login form, password-change flow"
```

---

## Task 4: Design source notes

**Files:**
- Create: `frontend/apps/web/design-source/login/NOTES.md`

- [ ] **Step 1: Create design notes**

Create `frontend/apps/web/design-source/login/NOTES.md`:

```markdown
# Login

**Owning feature:** `features/auth`
**Target route:** `/login` (public — outside AppShell)
**Page file:** `features/auth/pages/LoginPage.tsx + LoginPage.module.css`

## Layout

Full-page split: no AppShell, no Rail, no Toolbar.

- Left 45%: `var(--rail)` dark panel — Logo mark + wordmark + tagline + sub-tagline + copyright.
- Right 55%: `var(--surface)` white — logo mark icon, form title, form subtitle, form fields, submit.

When `mustChangePassword === true`, the right panel switches to a 3-field password change form (currentPassword / newPassword / confirmPassword). Same layout, different inner component.

## Reused primitives

- `components/ui/Icon` — logo mark (docs icon) + lock icon in password change

## New primitives needed

None — all styles scoped to LoginPage.module.css.

## Data sources

- No server queries on load.
- `POST /auth/login` via `useAuthSession().handleLogin()`
- `POST /auth/change-password` via `useAuthSession().handleChangePassword()`
- `auth.store`: `authState`, `user`, `loginForm`, `passwordForm` + setters
- `ui.store`: `error`, `message`

## Redirect logic

| Condition | Action |
|---|---|
| `authState === 'ready' && !mustChangePassword` | `<Navigate to={returnTo \|\| '/'} />` |
| `authState === 'ready' && mustChangePassword` | Show PasswordChangeForm |
| `authState === 'idle' \| 'loading' \| 'error'` | Show LoginForm |

`returnTo` is read from `sessionStorage.getItem('auth:returnTo')`.

## Open questions

- None.
```

- [ ] **Step 2: Run all tests — expect all pass**

```bash
cd .worktrees/screen-redesign/frontend/apps/web
npx vitest run src/lib/queryKeys.test.ts \
               src/components/ui/primitives.test.tsx \
               src/features/shell/pages/AppRoot.test.tsx \
               src/features/auth/pages/LoginPage.test.tsx \
               --reporter=verbose
```

Expected: `20 tests passed`

- [ ] **Step 3: Commit everything**

```bash
cd .worktrees/screen-redesign
git add frontend/apps/web/design-source/login/NOTES.md
git commit -m "docs(login): design source notes"
```

---

## Task 5: Update wiki tracker

**Files:**
- Modify: `wiki/implementation/screen-redesign-tracker.md`

- [ ] **Step 1: Mark Login complete, unblock Library**

In `wiki/implementation/screen-redesign-tracker.md`, update:

```markdown
| **Login** | Full-page split layout, auth form, no Rail | `wiki/implementation/plan-login.md` | ✅ Complete |
| **Library** | Dense document table, stat cards, filter tabs, collapsible activity sidebar, SectionPanel | — | 🔲 Not started |
```

- [ ] **Step 2: Commit**

```bash
git add wiki/implementation/screen-redesign-tracker.md
git commit -m "docs(tracker): Login complete — Library unblocked"
```

---

## Self-Review

**Spec coverage:**
- Full-page split layout, no AppShell ✅ (`.page` flex with `.left` + `.right`)
- Left dark panel with logo + tagline ✅ (`LeftPanel` component)
- Right white panel with form ✅ (`LoginForm` component)
- `useAuthSession().handleLogin()` on submit ✅
- Redirect to `/` on success ✅ (`Navigate` in `LoginPage` root)
- `mustChangePassword` flow ✅ (`PasswordChangeForm` component)
- Error message display ✅ (`error` from `useUiStore`)
- Loading state on button ✅ (`isLoading` disables + changes text)
- Wine design tokens throughout CSS ✅ (all `var(--*)`, zero hex)
- CSS Modules only ✅
- Auth route deduplication ✅ (Task 1)
- Design source NOTES.md ✅ (Task 4)

**Placeholder scan:** None found. All steps contain complete code.

**Type consistency:**
- `LoginPage` exports named `LoginPage` — matches `AppRouter.tsx` lazy import `m.LoginPage` ✅
- `useAuthSession()` returns `{ handleLogin, handleChangePassword, setLoginForm, setPasswordForm, loginForm, passwordForm }` — all used correctly ✅
- `useAuthStore` fields: `authState: LoadState`, `user: CurrentUser | null`, `loginForm`, `passwordForm` — all match `auth.store.ts` ✅
- `useUiStore` fields: `error: string`, `message: string` — match current `ui.store.ts` ✅
