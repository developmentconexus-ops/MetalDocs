import { useCallback, useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import * as api from './api/auth';
import { onAuthExpired } from "../../lib/api";
import type { CurrentUser } from "../../lib/types";
import { useAuthStore } from "../../store/auth.store";
import { useUiStore } from "../../store/ui.store";
import { asMessage, statusOf } from "../shared/errors";

type UseAuthSessionOptions = {
  onAuthenticated: (user: CurrentUser) => Promise<void>;
};

export function useAuthSession({ onAuthenticated }: UseAuthSessionOptions) {
  const queryClient = useQueryClient();
  const {
    authState, user, loginForm, passwordForm,
    setAuthState, setUser, setLoginForm, setPasswordForm,
  } = useAuthStore();
  const { setMessage, setError } = useUiStore();

  // Clear session when token expires server-side.
  useEffect(() => {
    return onAuthExpired(({ returnTo }) => {
      if (returnTo && returnTo !== "/" && !returnTo.startsWith("/login")) {
        sessionStorage.setItem("auth:returnTo", returnTo);
      }
      setUser(null);
      setAuthState("idle");
      queryClient.clear();
    });
  }, [setAuthState, setUser, queryClient]);

  const bootstrap = useCallback(async () => {
    try {
      const currentUser = await api.me();
      setUser(currentUser);
      setAuthState("ready");
      if (!currentUser.mustChangePassword) {
        await onAuthenticated(currentUser);
      }
    } catch (err) {
      if (statusOf(err) === 401) {
        setAuthState("idle");
        return;
      }
      setAuthState("error");
      setError(asMessage(err));
    }
  }, [onAuthenticated, setAuthState, setError, setUser]);

  const handleLogin = useCallback(
    async (event: React.FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      setError("");
      setMessage("");
      try {
        setAuthState("loading");
        const response = await api.login(loginForm);
        setUser(response.user);
        setAuthState("ready");
        if (!response.user.mustChangePassword) {
          await onAuthenticated(response.user);
          const returnTo = sessionStorage.getItem("auth:returnTo");
          if (returnTo) {
            sessionStorage.removeItem("auth:returnTo");
            window.history.pushState({}, "", returnTo);
            window.dispatchEvent(new PopStateEvent("popstate"));
          }
        }
      } catch (err) {
        setUser(null);
        if (statusOf(err) === 401) {
          setError("Usuario ou senha invalidos.");
          setAuthState("idle");
          return;
        }
        setError(asMessage(err));
        setAuthState("error");
      }
    },
    [loginForm, onAuthenticated, setAuthState, setError, setMessage, setUser],
  );

  const handleLogout = useCallback(async () => {
    setError("");
    setMessage("");
    try {
      await api.logout();
    } catch {
      // Best-effort logout; still clear local state.
    } finally {
      setUser(null);
      setAuthState("idle");
      queryClient.clear();
    }
  }, [queryClient, setAuthState, setError, setMessage, setUser]);

  const handleChangePassword = useCallback(
    async (event: React.FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      setError("");
      setMessage("");
      try {
        if (passwordForm.newPassword !== passwordForm.confirmPassword) {
          setError("A confirmacao da nova senha nao confere.");
          return;
        }
        const response = await api.changePassword(passwordForm);
        setPasswordForm({ currentPassword: "", newPassword: "", confirmPassword: "" });
        setUser(response.user);
        setLoginForm((current: { identifier: string; password: string }) => ({ ...current, identifier: response.user.username, password: "" }));
        await onAuthenticated(response.user);
        setAuthState("ready");
        setMessage("Senha alterada com sucesso.");
      } catch (err) {
        setError(asMessage(err));
      }
    },
    [onAuthenticated, passwordForm, setAuthState, setError, setLoginForm, setMessage, setPasswordForm, setUser],
  );

  return {
    authState,
    user,
    loginForm,
    passwordForm,
    setLoginForm,
    setPasswordForm,
    bootstrap,
    handleLogin,
    handleLogout,
    handleChangePassword,
  };
}
