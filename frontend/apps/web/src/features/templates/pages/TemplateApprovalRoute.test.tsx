import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ArtifactViewModel } from "../../shared/controlled-artifact/types";
import type { TemplateApprovalHandlers } from "../adapters/useTemplateApprovalArtifact";
import type { TemplateApprovalDecisionSubmit } from "../lib/templateApprovalDecision";

// ---------------------------------------------------------------------------
// Hoisted shared state — the vitest-safe way for multiple `vi.mock` factories
// (hoisted above this file's imports) to share mutable data without a TDZ
// crash. Drives which of the two kernel-truth shapes the fake adapter below
// returns: decision (under_review) vs. plain Publicar action (approved).
// ---------------------------------------------------------------------------

const shared = vi.hoisted(() => ({
  status: "under_review" as string,
  versionNumber: 2,
}));

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
  signoffVersion: vi.fn(),
  publishVersion: vi.fn(),
}));

vi.mock("../queries/useTemplateDetailQuery", () => ({
  useTemplateDetailQuery: () => ({
    data: {
      template: { id: "tpl-1", name: "Modelo X" },
      latest_version: { version_number: shared.versionNumber, status: shared.status },
    },
    isLoading: false,
    isError: false,
  }),
}));

const BASE_MODEL = {
  kind: "template" as const,
  id: "tpl-1",
  code: null,
  title: "Modelo X",
  revisionLabel: null,
  hero: { breadcrumb: [], badges: [], subtitle: null },
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
};

// Faithful-but-thin fake: branches the canned shape on `shared.status` (same
// value the mocked detail query above reads), mirroring how the real adapter
// only offers `decision` for under_review and falls back to the plain
// Publicar action for approved — WITHOUT re-implementing that gating logic,
// which is owned and unit-tested by canActOnVersion.test.ts /
// templateApprovalActions.test.ts / templateApprovalChain.test.ts.
vi.mock("../adapters/useTemplateApprovalArtifact", () => ({
  useTemplateApprovalArtifact: (
    _templateId: string,
    handlers: TemplateApprovalHandlers,
    decisionSubmit: TemplateApprovalDecisionSubmit,
  ) => {
    const model: ArtifactViewModel =
      shared.status === "under_review"
        ? {
            ...BASE_MODEL,
            status: "under_review",
            versionNumber: shared.versionNumber,
            actions: [],
            decision: {
              kicker: "Decisão requerida",
              heading: "Registrar decisão",
              description: "Revise o conteúdo do modelo e confirme sua senha para registrar a decisão.",
              options: [
                {
                  key: "accept",
                  label: "Assinar aprovação",
                  description: "Aprova a versão e avança para publicação.",
                  tone: "approve",
                  submitLabel: "Assinar aprovação",
                  requiresReason: false,
                },
                {
                  key: "reject",
                  label: "Rejeitar",
                  description: "Devolve o modelo para rascunho · requer motivo.",
                  tone: "reject",
                  submitLabel: "Rejeitar",
                  requiresReason: true,
                },
              ],
              reasonLabel: "Motivo",
              reasonPlaceholder: "Comentário registrado na trilha do modelo…",
              password: { label: "Senha" },
              legal: null,
              signer: {
                name: "Ana Aprovadora",
                detail: "ana@example.com",
                note: "Assinatura digital gerada no ato da confirmação.",
              },
              submit: async ({ optionKey, reason, password }) => {
                // Mirrors runtime: signoff moves status off under_review
                // (approved on accept, back to draft on reject). Set this
                // BEFORE awaiting decisionSubmit — its final step is the
                // route's own setMessage call, which triggers the re-render
                // that re-invokes this mocked hook; shared.status must
                // already be the new value by then so that render sees
                // decision:undefined and mounts TemplateApprovalExtras
                // (where the sidebar message renders). A real adapter would
                // instead see this via the post-invalidate query refetch.
                shared.status = optionKey === "accept" ? "approved" : "draft";
                await decisionSubmit(optionKey === "accept", reason, password);
              },
            },
          }
        : {
            ...BASE_MODEL,
            status: "approved",
            versionNumber: shared.versionNumber,
            actions: [
              {
                key: "publish",
                label: "Publicar",
                variant: "primary",
                available: true,
                run: () => handlers.runPublish(),
              },
            ],
            decision: undefined,
          };

    return {
      model,
      version: { version_number: shared.versionNumber, status: shared.status },
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

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

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
    shared.status = "under_review";
    shared.versionNumber = 2;
  });

  describe("under_review — signoff decision", () => {
    it("renders the decision panel radios, senha field, and the review canvas stub", () => {
      renderRoute();

      expect(screen.getByTestId("review-canvas")).toBeTruthy();
      expect(screen.getByRole("radio", { name: /Assinar aprovação/ })).toBeTruthy();
      expect(screen.getByRole("radio", { name: /Rejeitar/ })).toBeTruthy();
      expect(screen.getByLabelText(/Senha/)).toBeTruthy();
    });

    it("selecting Assinar aprovação + senha calls signoffVersion(tpl-1, 2, {decision:'approve', password}, <uuid>)", async () => {
      vi.mocked(templatesApi.signoffVersion).mockResolvedValue({
        signoff_id: "s1",
        was_replay: false,
        outcome: "approved",
      } as never);

      renderRoute();

      fireEvent.click(screen.getByRole("radio", { name: /Assinar aprovação/ }));
      fireEvent.change(screen.getByLabelText(/Senha/), { target: { value: "correcthorse" } });
      fireEvent.click(screen.getByRole("button", { name: /Assinar aprovação/ }));

      await waitFor(() => {
        expect(vi.mocked(templatesApi.signoffVersion)).toHaveBeenCalledOnce();
      });

      const [calledId, calledVersion, cmd] = vi.mocked(templatesApi.signoffVersion).mock.calls[0];
      expect(calledId).toBe("tpl-1");
      expect(calledVersion).toBe(2);
      expect(cmd).toEqual({ decision: "approve", reason: undefined, password: "correcthorse" });
      expect(await screen.findByText("Assinatura registrada.")).toBeTruthy();
    });

    it("selecting Rejeitar without a senha keeps submit disabled", () => {
      renderRoute();

      fireEvent.click(screen.getByRole("radio", { name: /Rejeitar/ }));
      fireEvent.change(screen.getByLabelText(/Motivo/), { target: { value: "Conteúdo incorreto" } });

      expect(screen.getByRole("button", { name: /Rejeitar/ })).toBeDisabled();
    });

    it("selecting Rejeitar with motivo + senha calls signoffVersion with decision:'reject' and the reason", async () => {
      vi.mocked(templatesApi.signoffVersion).mockResolvedValue({
        signoff_id: "s1",
        was_replay: false,
        outcome: "rejected",
      } as never);

      renderRoute();

      fireEvent.click(screen.getByRole("radio", { name: /Rejeitar/ }));
      fireEvent.change(screen.getByLabelText(/Motivo/), { target: { value: "Conteúdo incorreto" } });
      fireEvent.change(screen.getByLabelText(/Senha/), { target: { value: "correcthorse" } });
      fireEvent.click(screen.getByRole("button", { name: /Rejeitar/ }));

      await waitFor(() => {
        expect(vi.mocked(templatesApi.signoffVersion)).toHaveBeenCalledOnce();
      });

      const [, , cmd] = vi.mocked(templatesApi.signoffVersion).mock.calls[0];
      expect(cmd).toEqual({ decision: "reject", reason: "Conteúdo incorreto", password: "correcthorse" });
      expect(await screen.findByText("Rejeitado — volta para rascunho.")).toBeTruthy();
    });

    it("shows an error alert in the panel when signoffVersion rejects", async () => {
      vi.mocked(templatesApi.signoffVersion).mockRejectedValue(new Error("Servidor indisponível"));

      renderRoute();

      fireEvent.click(screen.getByRole("radio", { name: /Assinar aprovação/ }));
      fireEvent.change(screen.getByLabelText(/Senha/), { target: { value: "correcthorse" } });
      fireEvent.click(screen.getByRole("button", { name: /Assinar aprovação/ }));

      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeTruthy();
      });
    });
  });

  describe("approved — plain Publicar action", () => {
    beforeEach(() => {
      shared.status = "approved";
    });

    it("renders a single Publicar button (no decision radios)", () => {
      renderRoute();

      expect(screen.getByRole("button", { name: "Publicar" })).toBeTruthy();
      expect(screen.queryByRole("radio")).toBeNull();
    });

    it("clicking Publicar calls publishVersion(tpl-1, 2, <uuid>) and shows 'Publicado.'", async () => {
      vi.mocked(templatesApi.publishVersion).mockResolvedValue({ published_version_id: "v2" } as never);

      renderRoute();
      fireEvent.click(screen.getByRole("button", { name: "Publicar" }));

      await waitFor(() => {
        expect(vi.mocked(templatesApi.publishVersion)).toHaveBeenCalledOnce();
      });
      const [calledId, calledVersion] = vi.mocked(templatesApi.publishVersion).mock.calls[0];
      expect(calledId).toBe("tpl-1");
      expect(calledVersion).toBe(2);
      expect(await screen.findByText("Publicado.")).toBeTruthy();
    });
  });
});
