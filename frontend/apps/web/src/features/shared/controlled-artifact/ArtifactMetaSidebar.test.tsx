import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ArtifactMetaSidebar } from "./ArtifactMetaSidebar";
import type { ApprovalChainItem, ArtifactMetaModel, VersionHistoryItem } from "./types";

const emptyMeta: ArtifactMetaModel = {
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
};

describe("ArtifactMetaSidebar", () => {
  it("renders loading state instead of partial technical metadata during hydration", () => {
    render(
      <ArtifactMetaSidebar
        open
        onToggle={() => {}}
        loading
        code="POP-GENERAL-011"
        meta={emptyMeta}
      />,
    );

    expect(screen.getByText("Carregando metadados")).toBeInTheDocument();
    expect(screen.queryByText("---")).not.toBeInTheDocument();
    expect(screen.queryByText("—")).not.toBeInTheDocument();
    expect(screen.queryByText("pop")).not.toBeInTheDocument();
    expect(screen.queryByText("general")).not.toBeInTheDocument();
  });

  it("uses caller-supplied kind labels for aria-label and loading copy (purity)", () => {
    render(
      <ArtifactMetaSidebar
        open
        onToggle={() => {}}
        loading
        code="MODELO-001"
        meta={emptyMeta}
        ariaLabel="Identificação do template"
        loadingLabel="Carregando metadados do template"
      />,
    );

    expect(screen.getByLabelText("Identificação do template")).toBeInTheDocument();
    expect(screen.getByText("Carregando metadados do template")).toBeInTheDocument();
    expect(screen.queryByLabelText("Identificação do artefato")).not.toBeInTheDocument();
  });

  it("renders real governed metadata and hides approval chain in draft", () => {
    const meta: ArtifactMetaModel = {
      profileLabel: "Procedimento Operacional",
      areaLabel: "Recursos Humanos",
      visibilityLabel: "Restrito a area Recursos Humanos",
      fileSizeBytes: 1304,
      pageCount: 3,
      createdAt: null,
      effectiveFrom: null,
      nextReviewAt: null,
      ownerName: null,
      ownerDescriptor: null,
    };

    const lineage: VersionHistoryItem[] = [
      {
        versionNumber: 1,
        revisionNumber: 1,
        revisionLabel: "REV01",
        status: "draft",
        title: "Ajuste operacional",
        createdAt: "2026-05-18T12:00:00Z",
        isCurrent: true,
      },
      {
        versionNumber: 0,
        revisionNumber: 0,
        revisionLabel: "REV00",
        status: "published",
        title: "Primeira versao",
        createdAt: "2026-05-10T12:00:00Z",
        isCurrent: false,
      },
    ];

    render(
      <ArtifactMetaSidebar
        open
        onToggle={() => {}}
        code="POP-RH-001"
        meta={meta}
        approvalChain={null}
        lineage={lineage}
      />,
    );

    expect(screen.getByLabelText("Identificação do artefato")).toBeInTheDocument();
    expect(screen.getByText("Identificação")).toBeInTheDocument();
    expect(screen.queryByText("Metadados")).not.toBeInTheDocument();
    expect(screen.getByText("Tipo")).toBeInTheDocument();
    expect(screen.getByText("Área responsável")).toBeInTheDocument();
    expect(screen.getByText("Procedimento Operacional")).toBeInTheDocument();
    expect(screen.getByText("Recursos Humanos")).toBeInTheDocument();
    expect(screen.getByText("Restrito a area Recursos Humanos")).toBeInTheDocument();
    expect(screen.getByText("Páginas")).toBeInTheDocument();
    expect(screen.getByText("3 páginas · 1,3 KB")).toBeInTheDocument();
    expect(screen.getByText("REV01")).toBeInTheDocument();
    expect(screen.getByText("Ajuste operacional")).toBeInTheDocument();
    expect(screen.queryByText(/Draft|Em revis/i)).not.toBeInTheDocument();
    expect(screen.queryByText("Próximos aprovadores")).not.toBeInTheDocument();

    const toggle = screen.getByRole("button", { name: /fechar painel/i });
    expect(toggle).toHaveAttribute("aria-expanded", "true");
  });

  it("renders revision code, title, and date without workflow status text", () => {
    const lineage: VersionHistoryItem[] = [
      {
        versionNumber: 0,
        revisionNumber: 0,
        revisionLabel: "REV00",
        status: "draft",
        title: "Criacao do documento",
        createdAt: "2026-05-19T10:00:00-03:00",
        isCurrent: true,
      },
    ];

    render(
      <ArtifactMetaSidebar
        open
        onToggle={() => {}}
        meta={emptyMeta}
        lineage={lineage}
      />,
    );

    expect(screen.getByText("REV00")).toBeInTheDocument();
    expect(screen.getByText("Criacao do documento")).toBeInTheDocument();
    expect(screen.getByText("19/05/2026")).toBeInTheDocument();
    expect(screen.queryByText(/Draft|Em revis/i)).not.toBeInTheDocument();
  });

  it("keeps a governed dossier fill area when sidebar content is short", () => {
    const lineage: VersionHistoryItem[] = [
      {
        versionNumber: 0,
        revisionNumber: 0,
        revisionLabel: "REV00",
        status: "draft",
        title: "Criacao do documento",
        createdAt: "2026-05-19T10:00:00-03:00",
        isCurrent: true,
      },
    ];

    render(
      <ArtifactMetaSidebar
        open
        onToggle={() => {}}
        meta={emptyMeta}
        lineage={lineage}
      />,
    );

    expect(screen.getByText("Dossiê governado")).toBeInTheDocument();
  });

  it("collapses long governed histories and can expand them", async () => {
    const lineage: VersionHistoryItem[] = Array.from({ length: 5 }, (_, index) => ({
      versionNumber: index,
      revisionNumber: index,
      revisionLabel: `REV0${index}`,
      status: "approved" as VersionHistoryItem["status"],
      title: `Revision ${index}`,
      createdAt: `2026-05-${String(10 + index).padStart(2, "0")}T10:00:00-03:00`,
      isCurrent: index === 4,
    }));

    render(<ArtifactMetaSidebar open onToggle={() => {}} meta={emptyMeta} lineage={lineage} />);

    expect(screen.getByText("REV04")).toBeInTheDocument();
    expect(screen.queryByText("REV00")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /ver todas as revis/i }));
    expect(screen.getByText("REV00")).toBeInTheDocument();
  });

  it("renders full approval chain only during under review", () => {
    const approvalChain: ApprovalChainItem[] = [
      {
        stageIndex: 0,
        label: "Qualidade",
        status: "approved",
        roleLabel: "Qualidade",
        flowState: "approved",
        actorUserId: "user-1",
        actorDisplay: "Maria Souza",
        decision: "approve",
        signedAt: null,
      },
      {
        stageIndex: 0,
        label: "Qualidade",
        status: "active",
        roleLabel: "Qualidade",
        flowState: "current",
        actorUserId: "user-2",
        actorDisplay: "Joao Lima",
        decision: null,
        signedAt: null,
      },
      {
        stageIndex: 1,
        label: "Diretoria",
        status: "waiting",
        roleLabel: "Diretoria",
        flowState: "pending",
        actorUserId: "user-3",
        actorDisplay: "Ana Costa",
        decision: null,
        signedAt: null,
      },
    ];

    render(
      <ArtifactMetaSidebar
        open
        onToggle={vi.fn()}
        meta={emptyMeta}
        lineage={[]}
        approvalChain={approvalChain}
      />,
    );

    expect(screen.getByText("Próximos aprovadores")).toBeInTheDocument();
    expect(screen.getByText("Maria Souza")).toBeInTheDocument();
    expect(screen.getByText("Joao Lima")).toBeInTheDocument();
    expect(screen.getByText("Ana Costa")).toBeInTheDocument();
    expect(screen.getByText("aprovou")).toBeInTheDocument();
    expect(screen.getByText("próximo")).toBeInTheDocument();
    expect(screen.getByText("aguarda")).toBeInTheDocument();
  });

  it("renders only the collapsed toggle (no content) when closed", () => {
    render(
      <ArtifactMetaSidebar
        open={false}
        onToggle={() => {}}
        code="POP-RH-001"
        meta={emptyMeta}
      />,
    );

    const toggle = screen.getByRole("button", { name: /abrir painel/i });
    expect(toggle).toHaveAttribute("aria-expanded", "false");
    expect(screen.queryByText("Identificação")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("Identificação do artefato")).not.toBeInTheDocument();
  });

  it("renders an explained em-dash (never the old `---`) for absent identification values", () => {
    const meta: ArtifactMetaModel = {
      ...emptyMeta,
      absenceReasons: {
        profileLabel: "Sem perfil no snapshot.",
        areaLabel: "Sem área no snapshot.",
        visibilityLabel: "Sem documento controlado vinculado.",
        pageCount: "PDF oficial ainda não renderizado.",
      },
    };

    render(<ArtifactMetaSidebar open onToggle={() => {}} meta={meta} lineage={[]} />);

    expect(screen.queryByText("---")).not.toBeInTheDocument();
    expect(screen.getByTitle("Sem perfil no snapshot.")).toHaveTextContent("—");
    expect(screen.getByTitle("Sem área no snapshot.")).toHaveTextContent("—");
    expect(screen.getByTitle("Sem documento controlado vinculado.")).toHaveTextContent("—");
    // The pages row only materializes because an absence reason was supplied.
    expect(screen.getByText("Páginas")).toBeInTheDocument();
    expect(screen.getByTitle("PDF oficial ainda não renderizado.")).toHaveTextContent("—");
  });

  it("keeps the pages row hidden for kinds with no page metadata and no absence reason", () => {
    render(<ArtifactMetaSidebar open onToggle={() => {}} meta={emptyMeta} lineage={[]} />);

    expect(screen.queryByText("Páginas")).not.toBeInTheDocument();
    expect(screen.queryByText("---")).not.toBeInTheDocument();
  });

  it("uses the singular page noun for a one-page artifact", () => {
    const meta: ArtifactMetaModel = { ...emptyMeta, fileSizeBytes: 1304, pageCount: 1 };

    render(<ArtifactMetaSidebar open onToggle={() => {}} meta={meta} lineage={[]} />);

    expect(screen.getByText("1 página · 1,3 KB")).toBeInTheDocument();
  });
});
