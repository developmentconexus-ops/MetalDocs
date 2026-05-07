import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { apiFetch } from "../client";
import { ApiError } from "../errors";
import { AUTH_EXPIRED_EVENT } from "../authBus";

function makeResponse(body: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: vi.fn().mockResolvedValue(body),
  } as unknown as Response;
}

describe("apiFetch", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    window.history.pushState({}, "", "/");
  });

  it("returns parsed JSON on 200", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(makeResponse({ id: "doc-1" }));

    await expect(apiFetch<{ id: string }>("/api/documents/doc-1")).resolves.toEqual({ id: "doc-1" });
  });

  it("returns undefined on 204", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(makeResponse(undefined, 204));

    await expect(apiFetch<undefined>("/api/documents/doc-1")).resolves.toBeUndefined();
  });

  it("throws ApiError with parsed code on 4xx", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      makeResponse({ error: { code: "validation.request_invalid", message: "Requisição inválida", details: { field: "name" } } }, 400),
    );

    await expect(apiFetch("/api/documents")).rejects.toMatchObject({
      name: "ApiError",
      code: "validation.request_invalid",
      status: 400,
      message: "Requisição inválida",
      details: { field: "name" },
    } satisfies Partial<ApiError>);
  });

  it("dispatches auth:expired event on 401 and throws ApiError", async () => {
    window.history.pushState({}, "", "/documents/abc?tab=foo");
    const handler = vi.fn();
    window.addEventListener(AUTH_EXPIRED_EVENT, handler);
    vi.mocked(fetch).mockResolvedValueOnce(makeResponse({ error: { code: "authn.expired" } }, 401));

    await expect(apiFetch("/api/me")).rejects.toMatchObject({
      name: "ApiError",
      code: "authn.expired",
      status: 401,
      message: "Sessão expirada",
    });
    expect(handler).toHaveBeenCalledOnce();
    expect((handler.mock.calls[0][0] as CustomEvent).detail).toEqual({ returnTo: "/documents/abc?tab=foo" });

    window.removeEventListener(AUTH_EXPIRED_EVENT, handler);
  });

  it("falls back to http_500 code when error envelope absent", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(makeResponse({ message: "broken" }, 500));

    await expect(apiFetch("/api/documents")).rejects.toMatchObject({
      code: "http_500",
      status: 500,
      message: "Erro interno",
    });
  });

  it("sends Idempotency-Key header when idempotencyKey option is supplied", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(makeResponse({}));

    await apiFetch("/api/documents", { method: "POST", idempotencyKey: "uuid-test-1" });

    const callHeaders = vi.mocked(fetch).mock.calls[0][1]?.headers;
    expect(callHeaders).toMatchObject({ "Idempotency-Key": "uuid-test-1" });
  });

  it("does not send Idempotency-Key header when idempotencyKey option is omitted", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(makeResponse({}));

    await apiFetch("/api/documents", { method: "POST" });

    const callHeaders = vi.mocked(fetch).mock.calls[0][1]?.headers as Record<string, string> | undefined;
    expect(callHeaders?.["Idempotency-Key"]).toBeUndefined();
  });
});
