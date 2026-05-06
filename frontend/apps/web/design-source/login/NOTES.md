# Login

**Owning feature:** `features/auth`
**Target route:** `/login` (public — outside AppShell)
**Page file:** `features/auth/pages/LoginPage.tsx + LoginPage.module.css`

## Layout

Full-page split: no AppShell, no Rail, no Toolbar.

- Left (flex 1.1): `var(--rail)` dark panel — grid overlay, logo, kicker, 44px headline with accent italic, ISO body text, mono footer row.
- Right (flex 1): `var(--surface)` white — kicker, 26px h2, caption, SSO button, divider, identifier input, password row with forgot link, remember checkbox, primary submit, support footer.

When `mustChangePassword === true`, right panel switches to 3-field password change form (currentPassword / newPassword / confirmPassword).

## Reused primitives

- `components/ui/Icon` — not used on this page (logo mark is a plain div)

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