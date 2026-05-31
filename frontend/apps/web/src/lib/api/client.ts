import createClient from "openapi-fetch";
import type { paths } from '../api-types';
import { traceRequestEnd, traceRequestStart } from "../observability/apiTrace";
import { dispatchAuthExpired } from "./authBus";
import { ApiError } from "./errors";
import { parseProblem } from "./problem";

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "/api/v1";

export type ApiFetchOptions = RequestInit & { idempotencyKey?: string };

function apiUrl(path: string) {
  if (/^https?:\/\//i.test(path)) return path;
  if (path.startsWith("/api/")) return path;
  if (!path.startsWith("/")) return path;
  return `${API_BASE_URL}${path}`;
}

function withDefaultHeaders(init?: ApiFetchOptions): RequestInit | undefined {
  if (!init) return undefined;
  const { idempotencyKey, ...rest } = init;
  return {
    credentials: "include",
    ...rest,
    headers: {
      ...(rest.body instanceof FormData ? {} : { "Content-Type": "application/json" }),
      ...(rest.headers ?? {}),
      ...(idempotencyKey ? { "Idempotency-Key": idempotencyKey } : {}),
    },
  };
}

async function assertApiResponse(res: Response) {
  if (!res.ok) {
    const problem = await parseProblem(res.clone());

    // Only a genuine session/authn failure (or a bodyless 401) means the session
    // expired. Domain 401s (e.g. AUTH_INVALID_CREDENTIALS on a wrong current
    // password) must keep their RFC 9457 code so callers can map them correctly.
    if (res.status === 401 && (!problem || problem.code === "AUTH_UNAUTHORIZED")) {
      dispatchAuthExpired(window.location.pathname + window.location.search);
      throw ApiError.fromLegacy("authn.expired", 401, "Sessão expirada");
    }

    if (problem) {
      throw new ApiError(problem);
    }

    let body: { error?: { code?: string; message?: string; details?: unknown } } | undefined;

    try {
      body = await res.json();
    } catch {
      body = undefined;
    }

    const code = body?.error?.code ?? `http_${res.status}`;
    const message = body?.error?.message ?? `Não foi possível concluir a ação. Código: ${code}`;

    console.warn("[api] legacy error envelope from", res.url, "code=", code);
    throw ApiError.fromLegacy(code, res.status, message, body?.error?.details);
  }
}

export async function apiFetch<T>(url: string, init?: ApiFetchOptions): Promise<T>;
export async function apiFetch(input: RequestInfo | URL, init?: ApiFetchOptions): Promise<Response>;
export async function apiFetch<T>(input: RequestInfo | URL, init?: ApiFetchOptions): Promise<T | Response> {
  const isRawFetch = input instanceof Request || input instanceof URL;
  const method = (init?.method ?? (input instanceof Request ? input.method : "GET")).toUpperCase();
  const path = input instanceof Request ? input.url : String(input);
  const traceItemId = traceRequestStart(method, path);
  const requestInput = typeof input === "string" ? apiUrl(input) : input;
  const res = await fetch(requestInput, withDefaultHeaders(init));
  traceRequestEnd(traceItemId);

  await assertApiResponse(res);

  if (isRawFetch) {
    return res;
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return (await res.json()) as T;
}

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  return apiFetch<T>(path, init ?? {});
}

export async function requestRaw(path: string, init?: RequestInit, allowedErrorStatuses: number[] = []): Promise<Response> {
  const method = (init?.method ?? "GET").toUpperCase();
  const traceItemId = traceRequestStart(method, path);
  const response = await fetch(apiUrl(path), withDefaultHeaders(init ?? {}));
  traceRequestEnd(traceItemId);
  if (!allowedErrorStatuses.includes(response.status)) {
    await assertApiResponse(response);
  }
  return response;
}

export async function requestBlob(path: string, init?: RequestInit): Promise<Blob> {
  const response = await requestRaw(path, init);
  return response.blob();
}

export const api = createClient<paths>({ fetch: apiFetch });

