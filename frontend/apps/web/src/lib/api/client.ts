import { dispatchAuthExpired } from "./authBus";
import { ApiError } from "./errors";

export async function apiFetch<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, init);

  if (res.status === 401) {
    dispatchAuthExpired(window.location.pathname + window.location.search);
    throw new ApiError("authn.expired", 401, "Sessão expirada");
  }

  if (!res.ok) {
    let body: { error?: { code?: string; message?: string; details?: unknown } } | undefined;

    try {
      body = await res.json();
    } catch {
      body = undefined;
    }

    const code = body?.error?.code ?? `http_${res.status}`;
    const message = body?.error?.message ?? "Erro interno";

    throw new ApiError(code, res.status, message, body?.error?.details);
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return (await res.json()) as T;
}
