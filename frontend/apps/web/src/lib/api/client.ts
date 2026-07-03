import createClient from "openapi-fetch";
import type { paths } from '../api-types';
import { traceRequestEnd, traceRequestStart } from "../observability/apiTrace";
import { dispatchAuthExpired } from "./authBus";
import { ApiError } from "./errors";
import { parseProblem } from "./problem";

// TRANSPORT HIERARCHY (FE-13): this module exports more than one transport
// on purpose — pick from the top down, not by habit.
//   1. Generated `api` (below) — CANONICAL for every contracted route. It is
//      typed off api-types (oapi-codegen output) and gives compile-time
//      drift detection against api/openapi. Default choice for new call sites.
//   2. `apiFetch` — the sanctioned escape hatch. Use it for (a) the ETag
//      client (needs raw Response / header access the generated client
//      doesn't expose) and (b) non-generated or not-yet-contracted routes
//      (e.g. GET /feature-flags). It still gets credentials, tracing, and
//      problem+json handling — the only thing it gives up is compile-time
//      route/schema typing.
//   3. `request` / `requestRaw` / `requestBlob` — legacy thin wrappers kept
//      for existing call sites. Shrink-only: do not add new callers: use
//      `api` or `apiFetch` directly instead.
// Never call raw `fetch()` against `/api/...` from feature code — it skips
// credentials, tracing, and RFC 9457 problem+json handling (see FE-13).

// INVARIANT: API_BASE_URL must start with "/api/" (or be a full URL on another
// origin). createClient receives it as `baseUrl` so openapi-fetch emits
// /api/v1-prefixed paths; rewriteRequest then early-returns the Request
// unchanged. If this ever became a non-"/api/" path, rewriteRequest would
// rebuild the Request and re-break body POSTs (the "Failed to fetch" bug).
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "/api/v1";

export type ApiFetchOptions = RequestInit & { idempotencyKey?: string };

function apiUrl(path: string) {
  if (/^https?:\/\//i.test(path)) return path;
  if (path.startsWith("/api/")) return path;
  if (!path.startsWith("/")) return path;
  return `${API_BASE_URL}${path}`;
}

function rewriteRequest(req: Request): Request {
  let url: URL;
  try {
    url = new URL(req.url);
  } catch {
    return req;
  }
  if (url.origin !== window.location.origin) return req;
  if (url.pathname.startsWith("/api/")) return req;
  url.pathname = `${API_BASE_URL}${url.pathname}`;
  return new Request(url.toString(), req);
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

    // The whole API emits application/problem+json (api-contract-hardening
    // Phase D — RFC 9457 is the only error shape). A missing Problem body means
    // the failure did not originate from a MetalDocs handler (e.g. an
    // infrastructure 502/504 or a bodyless response), so synthesize a generic
    // error from the status alone — there is no legacy envelope to parse.
    const code = `http_${res.status}`;
    throw ApiError.fromLegacy(
      code,
      res.status,
      `Não foi possível concluir a ação. Código: ${code}`,
    );
  }
}

export async function apiFetch<T>(url: string, init?: ApiFetchOptions): Promise<T>;
export async function apiFetch(input: RequestInfo | URL, init?: ApiFetchOptions): Promise<Response>;
export async function apiFetch<T>(input: RequestInfo | URL, init?: ApiFetchOptions): Promise<T | Response> {
  const isRawFetch = input instanceof Request || input instanceof URL;
  const method = (init?.method ?? (input instanceof Request ? input.method : "GET")).toUpperCase();
  const path = input instanceof Request ? input.url : String(input);
  const traceItemId = traceRequestStart(method, path);
  const requestInput =
    typeof input === "string"
      ? apiUrl(input)
      : input instanceof Request
        ? rewriteRequest(input)
        : input;
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

export const api = createClient<paths>({ baseUrl: API_BASE_URL, fetch: apiFetch });

