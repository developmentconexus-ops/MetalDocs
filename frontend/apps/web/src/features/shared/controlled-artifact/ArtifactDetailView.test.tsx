import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { ArtifactDetailView } from "./ArtifactDetailView";
import type { ArtifactActionSet, ArtifactViewModel } from "./types";

const noopActions: ArtifactActionSet = {
  submit: { available: false, run: async () => {} },
  review: { available: false, run: async () => {} },
  approve: { available: false, run: async () => {} },
  reject: { available: false, run: async () => {} },
  publish: { available: false, run: async () => {} },
  createVersion: { available: false, run: async () => {} },
};

function buildModel(overrides: Partial<ArtifactViewModel> = {}): ArtifactViewModel {
  return {
    kind: "document",
    id: "doc-1",
    code: "POP-GENERAL-014",
    title: "Procedimento de exemplo",
    status: "published",
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
        { key: "status", label: "REV01 · publicado", variant: "status" },
        { key: "type", label: "Procedimento Operacional", variant: "type" },
      ],
      subtitle: "publicado em 19 de maio de 2026",
    },
    meta: {
      profileLabel: "Procedimento Operacional",
      areaLabel: "Geral",
      visibilityLabel: "Toda empresa",
      fileSizeBytes: 1024,
      pageCount: 1,
      createdAt: "2026-05-19T00:00:00.000Z",
      effectiveFrom: "2026-05-19T23:39:00.000Z",
      nextReviewAt: null,
      ownerName: "Administrator",
      ownerDescriptor: "publicado em 19 de maio de 2026",
    },
    kpis: [
      { key: "currentVersion", label: "Versão atual", value: "REV01", hint: "desde 19/05/2026" },
      { key: "coverage", label: "Cobertura", value: "12", hint: "destinatários obrigados · abrir fanout →", href: "/distribution" },
      { key: "nextReview", label: "Próxima revisão", value: "—", hint: "sem data de revisão definida" },
      { key: "pages", label: "Páginas", value: "1", hint: "metadado do arquivo" },
    ],
    approvalChain: [
      {
        stageIndex: 0,
        label: "Qualidade",
        status: "passed",
        actorUserId: "approver-1",
        actorDisplay: "approver-1",
        decision: "approve",
        signedAt: "2026-05-19T23:39:00.000Z",
      },
    ],
    lineage: [
      {
        versionNumber: 1,
        revisionNumber: 1,
        revisionLabel: "REV01",
        status: "published",
        title: "Procedimento de exemplo",
        createdAt: "2026-05-19T23:39:00.000Z",
        isCurrent: true,
      },
    ],
    tabs: [
      { key: "documento", label: "Documento", href: "." },
      { key: "distribuicao", label: "Distribuição", href: "distribution" },
    ],
    actions: noopActions,
    ...overrides,
  };
}

function renderView(node: React.ReactNode) {
  return render(<MemoryRouter>{node}</MemoryRouter>);
}

describe("ArtifactDetailView", () => {
  it("renders the hero title, code, and model badges", () => {
    renderView(<ArtifactDetailView model={buildModel()} />);

    // Title appears in the hero and in the About owner banner.
    expect(screen.getAllByText("Procedimento de exemplo").length).toBeGreaterThan(0);
    // Code is rendered both in the doc card and the hero badge chip.
    expect(screen.getAllByText("POP-GENERAL-014").length).toBeGreaterThan(0);
    expect(screen.getByText("REV01 · publicado")).toBeInTheDocument();
    // Subtitle appears in the hero and is echoed in the About owner banner meta.
    expect(screen.getAllByText("publicado em 19 de maio de 2026").length).toBeGreaterThan(0);
  });

  it("renders the KPI strip with the governed current version and page count", () => {
    renderView(<ArtifactDetailView model={buildModel()} />);

    const currentVersionCell = screen.getByText(/Vers.*o atual/i).parentElement;
    expect(currentVersionCell?.textContent).toContain("REV01");

    const paginasCell = screen.getByText("Páginas").parentElement;
    expect(paginasCell?.textContent).toContain("1");
  });

  it("renders the About facts section from the model meta", () => {
    renderView(<ArtifactDetailView model={buildModel()} />);

    expect(screen.getByText("Identificação e responsabilidade")).toBeInTheDocument();
    expect(screen.getByText("Tipo")).toBeInTheDocument();
    expect(screen.getAllByText("Procedimento Operacional").length).toBeGreaterThan(0);
    const areaCell = screen.getByText("Área").parentElement;
    expect(areaCell?.textContent).toContain("Geral");
    const tamanhoCell = screen.getByText("Tamanho").parentElement;
    expect(tamanhoCell?.textContent).toMatch(/1\s*KB/);
  });

  it("renders the approval chain from model.approvalChain", () => {
    renderView(<ArtifactDetailView model={buildModel()} />);

    expect(screen.getByText("Sign-offs desta versão")).toBeInTheDocument();
    expect(screen.getByText("Qualidade")).toBeInTheDocument();
    expect(screen.getAllByText("approver-1").length).toBeGreaterThan(0);
  });

  it("renders the empty approval state when approvalChain is null without crashing", () => {
    renderView(<ArtifactDetailView model={buildModel({ approvalChain: null })} />);

    expect(screen.getByText("Nenhum registro de aprovação para esta versão.")).toBeInTheDocument();
  });

  it("renders the lineage timeline from model.lineage", () => {
    renderView(<ArtifactDetailView model={buildModel()} />);

    expect(screen.getByText("Histórico de versões")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /REV01/ })).toBeInTheDocument();
    expect(screen.getByText("atual")).toBeInTheDocument();
  });

  it("renders the empty lineage state when there is no history", () => {
    renderView(<ArtifactDetailView model={buildModel({ lineage: [] })} />);

    expect(screen.getByText("Nenhuma revisão governada encontrada.")).toBeInTheDocument();
  });

  it("D4 — coverage KPI cell renders '12', links to /distribution, no '%', no progressbar", () => {
    renderView(<ArtifactDetailView model={buildModel()} />);

    // Coverage value from kpis
    expect(screen.getByText("12")).toBeInTheDocument();
    // The cell must be a link to /distribution
    const link = screen.getByRole("link", { name: /Cobertura|destinatários/ });
    expect(link).toHaveAttribute("href", "/distribution");
    // No fabricated percentage
    expect(screen.queryByText(/%/)).toBeNull();
    // No progress bar
    expect(screen.queryByRole("progressbar")).toBeNull();
  });

  it("D4 — owner banner shows ownerName and ownerDescriptor from model.meta", () => {
    renderView(<ArtifactDetailView model={buildModel()} />);

    expect(screen.getByText("Administrator")).toBeInTheDocument();
    // ownerDescriptor appears in the owner banner meta area
    const ownerMetaEls = screen.getAllByText("publicado em 19 de maio de 2026");
    expect(ownerMetaEls.length).toBeGreaterThan(0);
  });

  it("renders the heroActions, aside, and extras slots", () => {
    renderView(
      <ArtifactDetailView
        model={buildModel()}
        heroActions={<button type="button">Ação do documento</button>}
        aside={<div>Cartão de cobertura</div>}
        extras={<div>Banner extra</div>}
      />,
    );

    expect(screen.getByRole("button", { name: "Ação do documento" })).toBeInTheDocument();
    expect(screen.getByText("Cartão de cobertura")).toBeInTheDocument();
    expect(screen.getByText("Banner extra")).toBeInTheDocument();
  });
});
