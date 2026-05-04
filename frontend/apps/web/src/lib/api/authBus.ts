export const AUTH_EXPIRED_EVENT = "auth:expired";

export interface AuthExpiredDetail {
  returnTo: string;
}

export function dispatchAuthExpired(returnTo: string): void {
  window.dispatchEvent(new CustomEvent<AuthExpiredDetail>(AUTH_EXPIRED_EVENT, { detail: { returnTo } }));
}

export function onAuthExpired(handler: (detail: AuthExpiredDetail) => void): () => void {
  const listener = (event: Event) => {
    handler((event as CustomEvent<AuthExpiredDetail>).detail);
  };

  window.addEventListener(AUTH_EXPIRED_EVENT, listener);

  return () => {
    window.removeEventListener(AUTH_EXPIRED_EVENT, listener);
  };
}
