import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { MemoryRouter } from "react-router-dom";
import { ArtifactApprovalScreen } from "./ArtifactApprovalScreen";
import type { ArtifactAction, ArtifactViewModel } from "./types";

// ---------------------------------------------------------------------------
// Minimal ArtifactViewModel factory
// ---------------------------------------------------------------------------

function buildModel(overrides: Partial<ArtifactViewModel> = {}): ArtifactViewModel {
  return {
    kind: "document",
    id: "doc-1",
    code: "POP-GENERAL-014",
    title: "Procedimento de exemplo",
    status: "under_review",
    versionNumber: 1,
    revisionLabel: "REV01",
    hero: {
      breadcrumb: [
        { label: "Biblioteca", href: "/documents" },
        { label: "Geral" },
        { label: "POP-GENERAL-014" },
      ],
      badges: [
        { key: "code", label: "POP-GENERAL-014", variant: "code" },
        { key: "status", label: "REV01 · em revisão", variant: "status" },
        { key: "type", label: "Procedimento Operacional", variant: "type" },
      ],
      subtitle: "Aguardando decisão de aprovação",
    },
    meta: {
      profileLabel: "Procedimento Operacional",
      areaLabel: "Geral",
      visibilityLabel: "Toda empresa",
      fileSizeBytes: 1024,
      pageCount: 1,
      createdAt: "2026-05-19T00:00:00.000Z",
      effectiveFrom: null,
      nextReviewAt: null,
      ownerName: "Administrator",
      ownerDescriptor: "em revisão",
    },
    kpis: [],
    approvalChain: [
      {
        stageIndex: 0,
        label: "Qualidade",
        status: "active",
        roleLabel: "Qualidade",
        flowState: "current",
        actorUserId: "approver-1",
        actorDisplay: "Maria Souza",
        decision: null,
        signedAt: null,
      },
    ],
    lineage: [],
    tabs: [
      { key: "documento", label: "Documento", href: "." },
      { key: "distribuicao", label: "Distribuição", href: "distribution" },
    ],
    actions: [],
    ...overrides,
  };
}

function buildActions(overrides: Partial<ArtifactAction>[] = []): ArtifactAction[] {
  const defaults: ArtifactAction[] = [
    { key: "submit", label: "Submeter para revisão", variant: "primary", available: true, run: vi.fn() },
    { key: "signoff", label: "Assinar", variant: "primary", available: false, reason: "Aguarda etapa anterior", run: vi.fn() },
    { key: "reject", label: "Rejeitar", variant: "danger", available: true, run: vi.fn() },
  ];
  return defaults.map((d, i) => ({ ...d, ...overrides[i] }));
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("ArtifactApprovalScreen", () => {
  it("renders one button per model.actions entry in order with the correct label", () => {
    const actions = buildActions();
    render(
      <MemoryRouter>
        <ArtifactApprovalScreen
          model={buildModel({ actions })}
          main={<div>main content</div>}
        />
      </MemoryRouter>,
    );

    const buttons = screen.getAllByRole("button");
    const actionButtons = buttons.filter((b) => b.dataset["action"]);
    expect(actionButtons).toHaveLength(3);
    expect(actionButtons[0]).toHaveAttribute("data-action", "submit");
    expect(actionButtons[0]).toHaveTextContent("Submeter para revisão");
    expect(actionButtons[1]).toHaveAttribute("data-action", "signoff");
    expect(actionButtons[1]).toHaveTextContent("Assinar");
    expect(actionButtons[2]).toHaveAttribute("data-action", "reject");
    expect(actionButtons[2]).toHaveTextContent("Rejeitar");
  });

  it("renders an unavailable action as a disabled button with the reason as title", () => {
    const actions = buildActions();
    render(
      <MemoryRouter>
        <ArtifactApprovalScreen
          model={buildModel({ actions })}
          main={<div>main content</div>}
        />
      </MemoryRouter>,
    );

    const signoffBtn = screen.getByRole("button", { name: "Assinar" });
    expect(signoffBtn).toBeDisabled();
    expect(signoffBtn).toHaveAttribute("title", "Aguarda etapa anterior");
  });

  it("invokes run() when an available action button is clicked", () => {
    const run = vi.fn();
    const actions: ArtifactAction[] = [
      { key: "submit", label: "Submeter", variant: "primary", available: true, run },
    ];
    render(
      <MemoryRouter>
        <ArtifactApprovalScreen
          model={buildModel({ actions })}
          main={<div>main content</div>}
        />
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Submeter" }));
    expect(run).toHaveBeenCalledOnce();
  });

  it("renders no action buttons and no 'Ações' heading when model.actions is empty", () => {
    render(
      <MemoryRouter>
        <ArtifactApprovalScreen
          model={buildModel({ actions: [] })}
          main={<div>main content</div>}
        />
      </MemoryRouter>,
    );

    expect(screen.queryByText("Ações")).not.toBeInTheDocument();
    const actionButtons = screen.queryAllByRole("button").filter((b) => b.dataset["action"]);
    expect(actionButtons).toHaveLength(0);
  });

  it("renders the main, decisionExtras, and dialogs slots", () => {
    render(
      <MemoryRouter>
        <ArtifactApprovalScreen
          model={buildModel()}
          main={<div>Conteúdo principal</div>}
          decisionExtras={<div>Painel de integridade</div>}
          dialogs={<div>Modal de assinatura</div>}
        />
      </MemoryRouter>,
    );

    expect(screen.getByText("Conteúdo principal")).toBeInTheDocument();
    expect(screen.getByText("Painel de integridade")).toBeInTheDocument();
    expect(screen.getByText("Modal de assinatura")).toBeInTheDocument();
  });

  it("renders nothing for the approval chain when model.approvalChain is null", () => {
    render(
      <MemoryRouter>
        <ArtifactApprovalScreen
          model={buildModel({ approvalChain: null })}
          main={<div>main content</div>}
        />
      </MemoryRouter>,
    );

    expect(screen.queryByLabelText("Cadeia de aprovação")).not.toBeInTheDocument();
    expect(screen.queryByText("Cadeia de aprovação")).not.toBeInTheDocument();
  });

  it("renders chain actor display names when approvalChain has entries", () => {
    const approvalChain = [
      {
        stageIndex: 0,
        label: "Qualidade",
        status: "active",
        roleLabel: "Qualidade",
        flowState: "current" as const,
        actorUserId: "user-1",
        actorDisplay: "Maria Souza",
        decision: null,
        signedAt: null,
      },
      {
        stageIndex: 1,
        label: "Diretoria",
        status: "waiting",
        roleLabel: "Diretoria",
        flowState: "pending" as const,
        actorUserId: "user-2",
        actorDisplay: "João Lima",
        decision: null,
        signedAt: null,
      },
    ];
    render(
      <MemoryRouter>
        <ArtifactApprovalScreen
          model={buildModel({ approvalChain })}
          main={<div>main content</div>}
        />
      </MemoryRouter>,
    );

    expect(screen.getByText("Maria Souza")).toBeInTheDocument();
    expect(screen.getByText("João Lima")).toBeInTheDocument();
    // Stage/role labels appear in the flow-viz meta line ("<role> · <time>").
    expect(screen.getByText(/Qualidade/)).toBeInTheDocument();
    expect(screen.getByText(/Diretoria/)).toBeInTheDocument();
  });
});
