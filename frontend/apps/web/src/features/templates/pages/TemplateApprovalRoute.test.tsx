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

const capturedDecisionSubmit: { current: ((accept: boolean, reason: string) => Promise<void>) | null } = {
  current: null,
};

vi.mock("../adapters/useTemplateApprovalArtifact", () => ({
  useTemplateApprovalArtifact: (
    _templateId: string,
    handlers: import("../adapters/useTemplateApprovalArtifact").TemplateApprovalHandlers,
    decisionSubmit: (accept: boolean, reason: string) => Promise<void>,
  ) => {
    capturedHandlers.current = handlers;
    capturedDecisionSubmit.current = decisionSubmit;
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
        // Mirrors buildTemplateApprovalDecision's shape for status=approved — the
        // route now only supplies `submit`; construction lives in the adapter/lib.
        decision: {
          kicker: "Decisão requerida",
          heading: "Registrar decisão",
          description: "Confirme para publicar esta versão do modelo.",
          options: [
            {
              key: "accept",
              label: "Publicar",
              description: "Publica esta versão do modelo.",
              tone: "approve" as const,
              submitLabel: "Publicar",
              requiresReason: false,
            },
            {
              key: "reject",
              label: "Rejeitar",
              description: "Devolve o modelo para rascunho · requer motivo.",
              tone: "reject" as const,
              submitLabel: "Rejeitar",
              requiresReason: true,
            },
          ],
          reasonLabel: "Motivo",
          reasonPlaceholder: "Comentário registrado na trilha do modelo…",
          password: null,
          legal: null,
          signer: null,
          submit: async ({ optionKey, reason }: { optionKey: string; reason: string }) => {
            await decisionSubmit(optionKey === "accept", reason);
          },
        },
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

  it("renders the decision panel radios and the review canvas stub", () => {
    renderRoute();

    expect(screen.getByTestId("review-canvas")).toBeTruthy();
    // status=approved + accept.available → the shared DecisionPanel owns accept/reject
    // as radio cards (no plain action buttons).
    expect(screen.getByRole("radio", { name: /Publicar/ })).toBeTruthy();
    expect(screen.getByRole("radio", { name: /Rejeitar/ })).toBeTruthy();
  });

  it("selecting Publicar and submitting calls approveVersion with (tpl-1, 2, true, <uuid>, '')", async () => {
    vi.mocked(templatesApi.approveVersion).mockResolvedValue({ version_number: 2, status: "published" } as never);

    renderRoute();

    fireEvent.click(screen.getByRole("radio", { name: /Publicar/ }));
    // Submit footer label mirrors the selected option's submitLabel.
    fireEvent.click(screen.getByRole("button", { name: /Publicar/ }));

    await waitFor(() => {
      expect(vi.mocked(templatesApi.approveVersion)).toHaveBeenCalledOnce();
    });

    const [calledId, calledVersion, calledAccept, , calledReason] =
      vi.mocked(templatesApi.approveVersion).mock.calls[0];
    expect(calledId).toBe("tpl-1");
    expect(calledVersion).toBe(2);
    expect(calledAccept).toBe(true);
    // No motivo typed → reason arg (index 4) is the trimmed empty string.
    expect(calledReason).toBe("");
  });

  it("selecting Rejeitar, typing a motivo, and submitting calls approveVersion accept=false with the reason", async () => {
    vi.mocked(templatesApi.approveVersion).mockResolvedValue({ version_number: 2, status: "draft" } as never);

    renderRoute();

    fireEvent.click(screen.getByRole("radio", { name: /Rejeitar/ }));
    // Reject requires a motivo — the label carries the " · obrigatória" suffix.
    const textarea = screen.getByLabelText(/Motivo/);
    fireEvent.change(textarea, { target: { value: "Conteúdo incorreto" } });

    fireEvent.click(screen.getByRole("button", { name: /Rejeitar/ }));

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

  it("shows an error alert in the panel when approveVersion rejects", async () => {
    vi.mocked(templatesApi.approveVersion).mockRejectedValue(new Error("Servidor indisponível"));

    renderRoute();

    fireEvent.click(screen.getByRole("radio", { name: /Publicar/ }));
    fireEvent.click(screen.getByRole("button", { name: /Publicar/ }));

    await waitFor(() => {
      expect(screen.getByRole("alert")).toBeTruthy();
    });
  });
});
