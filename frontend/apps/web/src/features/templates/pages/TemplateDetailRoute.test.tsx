import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

// ── Navigate spy — hoisted before module imports ──────────────────────────────
const navigateMock = vi.fn();

vi.mock("react-router-dom", () => ({
  useParams: () => ({ templateId: "tpl-1" }),
  useNavigate: () => navigateMock,
}));

// ── Shared view — render slots so heroActions / extras are visible ────────────
vi.mock("../../shared/controlled-artifact/ArtifactDetailView", () => ({
  ArtifactDetailView: ({
    heroActions,
    extras,
  }: {
    heroActions?: React.ReactNode;
    extras?: React.ReactNode;
  }) => (
    <div>
      <div data-testid="hero-actions">{heroActions}</div>
      <div data-testid="extras">{extras}</div>
    </div>
  ),
}));

// ── Minimal ArtifactViewModel literal ────────────────────────────────────────
const MOCK_MODEL = {
  kind: "template" as const,
  id: "tpl-1",
  code: null,
  title: "Modelo X",
  status: "published" as const,
  versionNumber: 2,
  revisionLabel: "REV002",
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
  actions: [],
};

// ── Adapter mock ──────────────────────────────────────────────────────────────
const mockUseTemplateArtifact = vi.fn();
vi.mock("../adapters/useTemplateArtifact", () => ({
  useTemplateArtifact: (...args: unknown[]) => mockUseTemplateArtifact(...args),
}));

// ── Detail query mock ─────────────────────────────────────────────────────────
const mockUseTemplateDetailQuery = vi.fn();
vi.mock("../queries/useTemplateDetailQuery", () => ({
  useTemplateDetailQuery: (...args: unknown[]) =>
    mockUseTemplateDetailQuery(...args),
}));

// ── API mock ──────────────────────────────────────────────────────────────────
const createNextVersionMock = vi.fn();
vi.mock("../api/templates", () => ({
  createNextVersion: (...args: unknown[]) => createNextVersionMock(...args),
}));

// ── Import after mocks ────────────────────────────────────────────────────────
import { TemplateDetailRoute } from "./TemplateDetailRoute";

// ── Helpers ───────────────────────────────────────────────────────────────────
function makeClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function renderRoute() {
  return render(
    <QueryClientProvider client={makeClient()}>
      <TemplateDetailRoute />
    </QueryClientProvider>,
  );
}

function setupPublished() {
  mockUseTemplateArtifact.mockReturnValue({
    model: { ...MOCK_MODEL, status: "published" },
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
    heroKind: "newVersion",
  });
  mockUseTemplateDetailQuery.mockReturnValue({
    data: {
      template: { id: "tpl-1", name: "Modelo X" },
      latest_version: { status: "published" },
    },
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
  });
}

function setupDraft() {
  mockUseTemplateArtifact.mockReturnValue({
    model: { ...MOCK_MODEL, status: "draft" },
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
    heroKind: "editOnly",
  });
  mockUseTemplateDetailQuery.mockReturnValue({
    data: {
      template: { id: "tpl-1", name: "Modelo X" },
      latest_version: { status: "draft" },
    },
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
  });
}

function setupUnderReview() {
  mockUseTemplateArtifact.mockReturnValue({
    model: { ...MOCK_MODEL, status: "under_review" },
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
    heroKind: "review",
  });
  mockUseTemplateDetailQuery.mockReturnValue({
    data: {
      template: { id: "tpl-1", name: "Modelo X" },
      latest_version: { status: "under_review" },
    },
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
  });
}

function setupApproved() {
  mockUseTemplateArtifact.mockReturnValue({
    model: { ...MOCK_MODEL, status: "approved" },
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
    heroKind: "publish",
  });
  mockUseTemplateDetailQuery.mockReturnValue({
    data: {
      template: { id: "tpl-1", name: "Modelo X" },
      latest_version: { status: "approved" },
    },
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
  });
}

// ── Tests ─────────────────────────────────────────────────────────────────────
describe("TemplateDetailRoute", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders 'Criar nova versão' enabled when latest_version.status is published", () => {
    setupPublished();
    renderRoute();

    const btn = screen.getByRole("button", { name: /criar nova versão/i });
    expect(btn).toBeInTheDocument();
    expect(btn).not.toBeDisabled();
  });

  it("calls createNextVersion('tpl-1') and navigates to /templates/tpl-1/edit on success", async () => {
    setupPublished();
    createNextVersionMock.mockResolvedValue({ id: "ver-3", version_number: 3 });

    renderRoute();

    fireEvent.click(screen.getByRole("button", { name: /criar nova versão/i }));

    await waitFor(() => {
      expect(createNextVersionMock).toHaveBeenCalledWith("tpl-1");
    });
    await waitFor(() => {
      expect(navigateMock).toHaveBeenCalledWith("/templates/tpl-1/edit");
    });
  });

  it("shows 'Editar modelo' primary and disabled 'Criar nova versão' when status is draft", () => {
    setupDraft();
    renderRoute();

    const editBtn = screen.getByRole("button", { name: /editar modelo/i });
    expect(editBtn).toBeInTheDocument();
    expect(editBtn).not.toBeDisabled();

    const createBtn = screen.getByRole("button", { name: /criar nova versão/i });
    expect(createBtn).toBeDisabled();
    expect(createBtn).toHaveAttribute("aria-disabled");
  });

  it("renders loading state when adapter returns isLoading: true", () => {
    mockUseTemplateArtifact.mockReturnValue({
      model: MOCK_MODEL,
      isLoading: true,
      isError: false,
      refetch: vi.fn(),
    });
    mockUseTemplateDetailQuery.mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      refetch: vi.fn(),
    });

    renderRoute();

    expect(screen.getByText(/carregando modelo/i)).toBeInTheDocument();
    expect(screen.getByRole("status")).toBeInTheDocument();
  });

  it("under_review shows 'Revisar versão' and navigates to the approval screen", () => {
    setupUnderReview();
    renderRoute();

    const reviewBtn = screen.getByRole("button", { name: /revisar versão/i });
    expect(reviewBtn).toBeInTheDocument();

    fireEvent.click(reviewBtn);
    expect(navigateMock).toHaveBeenCalledWith("/templates/tpl-1/approval");
  });

  it("approved shows 'Publicar versão' and navigates to the approval screen", () => {
    setupApproved();
    renderRoute();

    const publishBtn = screen.getByRole("button", { name: /publicar versão/i });
    expect(publishBtn).toBeInTheDocument();

    fireEvent.click(publishBtn);
    expect(navigateMock).toHaveBeenCalledWith("/templates/tpl-1/approval");
  });
});
