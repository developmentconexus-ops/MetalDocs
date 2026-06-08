import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, waitFor, within, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import MembershipsTab from "./MembershipsTab";
import { api } from "../../../lib/api/client";
import { toast } from "sonner";
import { useAuthStore } from "../../../store/auth.store";
import type { CurrentUser } from "../../../lib/types";

vi.mock("../../../lib/api/client", () => ({
  api: { GET: vi.fn(), POST: vi.fn(), DELETE: vi.fn() },
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}));

const mockGet = api.GET as unknown as ReturnType<typeof vi.fn>;
const mockPost = api.POST as unknown as ReturnType<typeof vi.fn>;

type Membership = {
  user_id: string;
  tenant_id: string;
  area_code: string;
  role: string;
  effective_from: string;
  granted_by?: string | null;
};

type ManagedUserLite = {
  user_id: string;
  username: string;
  display_name: string;
  area_memberships: ReadonlyArray<{ area_code: string }>;
};

function makeUser(capabilities: string[]): CurrentUser {
  return {
    userId: "actor-1",
    tenantId: "t-1",
    tenantName: "System Tenant",
    username: "admin",
    email: "admin@example.com",
    displayName: "Admin",
    mustChangePassword: false,
    roles: ["system_admin"],
    capabilities,
  };
}

function primeApi(memberships: Membership[], users: ManagedUserLite[]) {
  mockGet.mockImplementation((path: string) => {
    if (path === "/iam/area-memberships") {
      return Promise.resolve({ data: { items: memberships }, error: undefined });
    }
    if (path === "/iam/users") {
      return Promise.resolve({ data: { items: users }, error: undefined });
    }
    return Promise.resolve({ data: undefined, error: { status: 404 } });
  });
}

function renderTab() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={["/admin/memberships"]}>
        <MembershipsTab />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

const SAMPLE_USERS: ManagedUserLite[] = [
  { user_id: "u-1", username: "alice", display_name: "Alice", area_memberships: [{ area_code: "QUA" }] },
  { user_id: "u-2", username: "bob", display_name: "Bob", area_memberships: [] },
];

const SAMPLE_MEMBERSHIPS: Membership[] = [
  {
    user_id: "u-1",
    tenant_id: "t-1",
    area_code: "QUA",
    role: "editor",
    effective_from: "2026-01-01T00:00:00Z",
    granted_by: "actor-1",
  },
];

describe("MembershipsTab", () => {
  beforeEach(() => {
    mockGet.mockReset();
    mockPost.mockReset();
    (toast.success as ReturnType<typeof vi.fn>).mockReset();
    (toast.error as ReturnType<typeof vi.fn>).mockReset();
    useAuthStore.setState({ authState: "ready", user: makeUser([]) });
  });

  it("shows the grant action and row revoke for a user who can manage", async () => {
    useAuthStore.setState({
      user: makeUser(["membership.view", "membership.manage"]),
    });
    primeApi(SAMPLE_MEMBERSHIPS, SAMPLE_USERS);
    renderTab();

    await waitFor(() => expect(screen.getByText("Alice")).toBeInTheDocument());
    expect(screen.getByRole("button", { name: /conceder/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /revogar/i })).toBeInTheDocument();
  });

  it("hides manage affordances for a view-only user", async () => {
    useAuthStore.setState({ user: makeUser(["membership.view"]) });
    primeApi(SAMPLE_MEMBERSHIPS, SAMPLE_USERS);
    renderTab();

    await waitFor(() => expect(screen.getByText("Alice")).toBeInTheDocument());
    expect(screen.queryByRole("button", { name: /conceder/i })).toBeNull();
    expect(screen.queryByRole("button", { name: /revogar/i })).toBeNull();
  });

  it("renders the empty state when there are no memberships", async () => {
    useAuthStore.setState({ user: makeUser(["membership.view"]) });
    primeApi([], SAMPLE_USERS);
    renderTab();

    await waitFor(() =>
      expect(screen.getByText("Nenhum membership encontrado")).toBeInTheDocument(),
    );
  });

  it("grants a membership and surfaces a success toast", async () => {
    useAuthStore.setState({
      user: makeUser(["membership.view", "membership.manage"]),
    });
    primeApi(SAMPLE_MEMBERSHIPS, SAMPLE_USERS);
    mockPost.mockResolvedValue({
      data: { user_id: "u-2", tenant_id: "t-1", area_code: "PRD", role: "viewer" },
      error: undefined,
    });
    renderTab();

    await waitFor(() => expect(screen.getByText("Alice")).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: /conceder/i }));

    const dialog = await screen.findByRole("dialog");
    fireEvent.change(within(dialog).getByLabelText(/usuário/i), {
      target: { value: "u-2" },
    });
    fireEvent.change(within(dialog).getByLabelText(/^área/i), {
      target: { value: "PRD" },
    });
    fireEvent.click(within(dialog).getByRole("button", { name: /conceder/i }));

    await waitFor(() =>
      expect(mockPost).toHaveBeenCalledWith("/iam/area-memberships", {
        body: { user_id: "u-2", area_code: "PRD", role: "viewer" },
      }),
    );
    await waitFor(() =>
      expect(toast.success as ReturnType<typeof vi.fn>).toHaveBeenCalled(),
    );
  });
});
