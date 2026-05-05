import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { dispatchAuthExpired } from "../../../lib/api";
import { useAuthSession } from "../useAuthSession";

vi.mock("../api/auth", () => ({
  me: vi.fn(),
  login: vi.fn(),
  logout: vi.fn(),
  changePassword: vi.fn(),
}));

describe("useAuthSession auth:expired returnTo", () => {
  afterEach(() => {
    sessionStorage.clear();
  });

  it("stores returnTo on auth:expired event", () => {
    renderHook(() => useAuthSession({ onAuthenticated: vi.fn() }));

    act(() => dispatchAuthExpired("/documents/abc"));

    expect(sessionStorage.getItem("auth:returnTo")).toBe("/documents/abc");
  });

  it("ignores root path and login paths", () => {
    renderHook(() => useAuthSession({ onAuthenticated: vi.fn() }));

    act(() => dispatchAuthExpired("/"));
    expect(sessionStorage.getItem("auth:returnTo")).toBeNull();

    act(() => dispatchAuthExpired("/login"));
    expect(sessionStorage.getItem("auth:returnTo")).toBeNull();
  });
});
