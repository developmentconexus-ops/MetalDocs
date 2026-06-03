import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { apiFetch } from "../client";

// Regression guard for the "Failed to fetch" defect. openapi-fetch hands the
// custom fetch a `Request` object. The client's rewriteRequest() must forward a
// Request whose path already starts with /api/ UNCHANGED — it must NOT rebuild
// it via `new Request(url, origReq)`, because copying a Request body yields a
// ReadableStream that fetch rejects without duplex:'half' ("Failed to fetch").
// Adding `baseUrl: API_BASE_URL` to createClient makes openapi-fetch emit
// /api/v1-prefixed URLs, hitting exactly this pass-through branch.
describe("apiFetch — Request (openapi-fetch) path", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("forwards an /api/v1 POST Request with its body intact (no reconstruction)", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(new Response(null, { status: 201 }));

    const req = new Request("http://localhost/api/v1/iam/area-memberships", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ userId: "u1", areaCode: "qua", role: "viewer" }),
    });

    await apiFetch(req);

    expect(fetch).toHaveBeenCalledOnce();
    const passed = vi.mocked(fetch).mock.calls[0][0] as Request;
    expect(passed.method).toBe("POST");
    expect(new URL(passed.url).pathname).toBe("/api/v1/iam/area-memberships");
    // Body must survive — the reconstruction bug stripped it into a broken stream.
    await expect(passed.clone().text()).resolves.toContain('"userId":"u1"');
  });
});
