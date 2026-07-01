import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ArtifactViewModel } from "../../shared/controlled-artifact/types";

// ---------------------------------------------------------------------------
// Mocks — declared before the import under test (hoisting)
// ---------------------------------------------------------------------------

vi.mock("react-router-dom", () => ({
  useParams: () => ({ templateId: "tpl-1" }),
}));

vi.mock("../components/TemplateReviewCanvas", () => ({
  TemplateReviewCanvas: () => <div data-testid="review-canvas" />,
}));

vi.mock("../api/templates", () => ({
  submitForReview: vi.fn(),
  reviewVersion: vi.fn(),
  approveVersion: vi.fn(),
}));

vi.mock("../queries/useTemplateDetailQuery", () => ({
  useTemplateDetailQuery: vi.fn(),
}));

const capturedHandlers: { current: import("../adapters/useTemplateApprovalArtifact").TemplateApprovalHandlers | null } = {
  current: null,
};

const MOCK_MODEL: ArtifactViewModel = {
  kind: "template",
  id: "tpl-1",
  code: null,
  title: "Modelo X",
  status: "approved",
  versionNumber: 2,
  revisionLabel: null,
  hero: {
    breadcrumb: [],
    badges: [],
    subtitle: null,
  },
  meta: {
    profileLabel: null,
    areaLabel: null,
    visibilityLabel: null,
    fileSizeBytes: null,
    pageCount: null,
    createdAt: null,
    effectiveFrom: null,
    nextReviewAt: null,
    ownerName: null,
    ownerDescriptor: null,
  },
  kpis: [],
  approvalChain: null,
  lineage: [],
  tabs: [{ key: "documento", label: "Documento" }],
  actions: [],
};

vi.mock("../adapters/useTemplateApprovalArtifact", () => ({
  useTemplateApprovalArtifact: (
    _templateId: string,
    handlers: import("../adapters/useTemplateApprovalArtifact").TemplateApprovalHandlers,
  ) => {
    capturedHandlers.current = handlers;
    return {
      model: {
        ...MOCK_MODEL,
        actions: [
          {
            key: "accept",
            label: "Publicar",
            variant: "primary" as const,
            available: true,
            run: () => handlers.runApprove(true),
          },
          {
            key: "reject",
            label: "Rejeitar",
            variant: "danger" as const,
            available: true,
            run: () => handlers.runApprove(false),
          },
        ],
      },
      version: { version_number: 2, status: "approved" },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    };
  },
}));

// ---------------------------------------------------------------------------
// Imports (after mocks)
// ---------------------------------------------------------------------------

import { TemplateApprovalRoute } from "./TemplateApprovalRoute";
import * as templatesApi from "../api/templates";
import { useTemplateDetailQuery } from "../queries/useTemplateDetailQuery";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function makeDetailReturn(overrides: { status?: string; version_number?: number } = {}) {
  return {
    data: {
      template: { id: "tpl-1", name: "Modelo X" },
      latest_version: {
        version_number: overrides.version_number ?? 2,
        status: overrides.status ?? "approved",
      },
    },
    isLoading: false,
    isError: false,
  };
}

function renderRoute() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <TemplateApprovalRoute />
    </QueryClientProvider>,
  );
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("TemplateApprovalRoute", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    capturedHandlers.current = null;
    vi.mocked(useTemplateDetailQuery).mockReturnValue(makeDetailReturn() as ReturnType<typeof useTemplateDetailQuery>);
  });

  it("renders the action buttons and the review canvas stub", () => {
    renderRoute();

    expect(screen.getByTestId("review-canvas")).toBeTruthy();
    expect(screen.getByRole("button", { name: "Publicar" })).toBeTruthy();
    expect(screen.getByRole("button", { name: "Rejeitar" })).toBeTruthy();
  });

  it("clicking Publicar calls approveVersion with (tpl-1, 2, true, <uuid>, '') and shows success", async () => {
    vi.mocked(templatesApi.approveVersion).mockResolvedValue({ version_number: 2, status: "published" } as never);

    renderRoute();

    fireEvent.click(screen.getByRole("button", { name: "Publicar" }));

    await waitFor(() => {
      expect(screen.getByText(/Publicado/)).toBeTruthy();
    });

    expect(vi.mocked(templatesApi.approveVersion)).toHaveBeenCalledOnce();
    const [calledId, calledVersion, calledAccept] = vi.mocked(templatesApi.approveVersion).mock.calls[0];
    expect(calledId).toBe("tpl-1");
    expect(calledVersion).toBe(2);
    expect(calledAccept).toBe(true);
    // reason arg (index 4) defaults to empty string
    expect(vi.mocked(templatesApi.approveVersion).mock.calls[0][4]).toBe("");
  });

  it("typing a reason then clicking Rejeitar calls approveVersion with accept=false and the typed reason", async () => {
    vi.mocked(templatesApi.approveVersion).mockResolvedValue({ version_number: 2, status: "draft" } as never);

    renderRoute();

    // The textarea is labelled "Motivo (opcional)" — status=approved so showReason=true
    const textarea = screen.getByLabelText("Motivo (opcional)");
    fireEvent.change(textarea, { target: { value: "Conteúdo incorreto" } });

    fireEvent.click(screen.getByRole("button", { name: "Rejeitar" }));

    await waitFor(() => {
      expect(vi.mocked(templatesApi.approveVersion)).toHaveBeenCalledOnce();
    });

    const [calledId, calledVersion, calledAccept, , calledReason] =
      vi.mocked(templatesApi.approveVersion).mock.calls[0];
    expect(calledId).toBe("tpl-1");
    expect(calledVersion).toBe(2);
    expect(calledAccept).toBe(false);
    expect(calledReason).toBe("Conteúdo incorreto");
  });

  it("shows an error alert when approveVersion rejects", async () => {
    vi.mocked(templatesApi.approveVersion).mockRejectedValue(new Error("Servidor indisponível"));

    renderRoute();

    fireEvent.click(screen.getByRole("button", { name: "Publicar" }));

    await waitFor(() => {
      expect(screen.getByRole("alert")).toBeTruthy();
    });
  });
});
