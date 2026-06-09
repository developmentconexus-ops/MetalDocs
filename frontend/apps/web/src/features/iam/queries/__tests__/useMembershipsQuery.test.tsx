import { describe, it, expect, beforeEach, vi } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { useMembershipsQuery } from "../useMembershipsQuery";
import { QK } from "../../../../lib/queryKeys";
import { api } from "../../../../lib/api/client";

vi.mock("../../../../lib/api/client", () => ({
  api: { GET: vi.fn() },
}));

const mockGet = api.GET as unknown as ReturnType<typeof vi.fn>;

function wrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
}

describe("QK.iam.memberships.list", () => {
  it("maps params into a stable, distinct query key", () => {
    expect(QK.iam.memberships.list({})).toEqual([
      "iam",
      "admin",
      "memberships",
      "list",
      {},
    ]);
    expect(QK.iam.memberships.list({ areaCode: "QUA", role: "editor" })).toEqual([
      "iam",
      "admin",
      "memberships",
      "list",
      { areaCode: "QUA", role: "editor" },
    ]);
  });
});

describe("useMembershipsQuery", () => {
  beforeEach(() => mockGet.mockReset());

  it("calls the listing endpoint with the params as query and returns items", async () => {
    const items = [
      {
        user_id: "u-1",
        tenant_id: "t-1",
        area_code: "QUA",
        role: "editor",
        effective_from: "2026-01-01T00:00:00Z",
        granted_by: "admin",
      },
    ];
    mockGet.mockResolvedValue({ data: { items }, error: undefined });

    const { result } = renderHook(() => useMembershipsQuery({ area_code: "QUA" }), {
      wrapper: wrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.items).toHaveLength(1);
    expect(result.current.data?.items[0].area_code).toBe("QUA");
    expect(mockGet).toHaveBeenCalledWith("/iam/area-memberships", {
      params: { query: { area_code: "QUA" } },
    });
  });

  it("surfaces an error when the endpoint returns an RFC 9457 problem", async () => {
    mockGet.mockResolvedValue({
      data: undefined,
      error: { status: 400, code: "VALIDATION_ERROR", title: "Bad Request" },
    });

    const { result } = renderHook(() => useMembershipsQuery(), {
      wrapper: wrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});
