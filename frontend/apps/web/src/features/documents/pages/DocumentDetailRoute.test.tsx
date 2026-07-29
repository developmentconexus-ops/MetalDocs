// FE-11: revision-initiate gating is capability-based (ADR 0022 — capabilities,
// never roles). This test drives the route-level hero action button (the only
// place `canInitiateRevision`/`canCreateRevision` become observable UI) through
// `useHasCapability('document.edit')` instead of a role check.
//
// ADR 0085 (Stage B): the manual "Publicar / Agendar" affordance is gone — release
// is an approval-driven coordinator outcome, so an approved document exposes no
// publication action here at all.
//
// FE-02: DocumentDetailRoute no longer runs its own document/approval/active-document/
// distribution queries or re-derives gating — it consumes `useDocumentArtifact`'s
// `doc` / `activeDocument` / `obligatedCount` / `gating` exports directly. This test
// mocks only that one adapter hook.
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { DocumentDetailRoute } from './DocumentDetailRoute';
import type { DocumentDetailGating } from '../adapters/useDocumentArtifact';

const hasCapMock = vi.fn();
vi.mock('../../iam/hooks/useHasCapability', () => ({
  useHasCapability: (cap: string) => hasCapMock(cap),
}));

vi.mock('../adapters/useDocumentArtifact', () => ({
  useDocumentArtifact: vi.fn(),
}));

// D4: the embedded viewer owns its own fetch; stub it so this suite asserts the
// route's wiring (view vs. download separation), not the viewer's internals —
// those are covered by DocumentPdfViewerDialog.test.tsx.
vi.mock('../components/DocumentPdfViewerDialog', () => ({
  DocumentPdfViewerDialog: ({ documentLabel }: { documentLabel: string }) => (
    <div data-testid='pdf-viewer-dialog'>{documentLabel}</div>
  ),
}));

import { useDocumentArtifact } from '../adapters/useDocumentArtifact';

const BASE_DOC = {
  id: 'doc-1',
  code: 'POP-QUA-0148',
  name: 'POP Limpeza de Linha',
  status: 'published',
  revision_version: 1,
  revision_number: 1,
  controlled_document_id: 'cd-1',
  current_revision_id: 'rev-1',
  created_by: 'admin-user',
};

function baseModel(status: string) {
  return {
    kind: 'document' as const,
    id: 'doc-1',
    code: 'POP-QUA-0148',
    title: 'POP Limpeza de Linha',
    status,
    versionNumber: 1,
    revisionLabel: 'REV01',
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
    tabs: [],
    actions: [],
  };
}

function baseGating(overrides: Partial<DocumentDetailGating> = {}): DocumentDetailGating {
  return {
    isObsolete: false,
    isApproved: false,
    isPublished: true,
    canInitiateRevision: true,
    canCreateRevision: true,
    activeSiblingDocumentId: null,
    activeSiblingState: null,
    activeSiblingCtaLabel: 'Iniciar revisão',
    activeSiblingDestination: null,
    ...overrides,
  };
}

function mockArtifact(overrides: {
  status?: string;
  gating?: Partial<DocumentDetailGating>;
  doc?: Record<string, unknown>;
} = {}) {
  const status = overrides.status ?? 'published';
  vi.mocked(useDocumentArtifact).mockReturnValue({
    model: baseModel(status),
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
    doc: { ...BASE_DOC, status, ...overrides.doc },
    obligatedCount: '3',
    gating: baseGating(overrides.gating),
  } as never);
}

function renderRoute() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={['/documents/doc-1']}>
        <Routes>
          <Route path="/documents/:documentId" element={<DocumentDetailRoute />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('DocumentDetailRoute — FE-11 capability gating', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('enables "Iniciar revisão" when the user holds document.edit', async () => {
    hasCapMock.mockImplementation((cap: string) => cap === 'document.edit');
    mockArtifact({
      status: 'published',
      gating: { canInitiateRevision: true, canCreateRevision: true },
    });

    renderRoute();

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Iniciar revisão/i })).toBeInTheDocument();
    });
    const btn = screen.getByRole('button', { name: /Iniciar revisão/i });
    expect(btn).toHaveAttribute('aria-disabled', 'false');
    expect(btn).not.toHaveAttribute('title', 'Sem permissão para iniciar revisão');
  });

  it('soft-disables "Iniciar revisão" without document.edit (capability, not role)', async () => {
    hasCapMock.mockReturnValue(false);
    mockArtifact({
      status: 'published',
      gating: { canInitiateRevision: false, canCreateRevision: false },
    });

    renderRoute();

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Iniciar revisão/i })).toBeInTheDocument();
    });
    const btn = screen.getByRole('button', { name: /Iniciar revisão/i });
    expect(btn).toHaveAttribute('aria-disabled', 'true');
    expect(btn).toHaveAttribute('title', 'Sem permissão para iniciar revisão');
  });

  it('ADR 0085 — an approved document exposes NO manual publication action', async () => {
    // Release is an approval-driven coordinator outcome; the detail screen reports
    // status only. A resurfaced publish/schedule affordance here is a regression.
    hasCapMock.mockReturnValue(true);
    mockArtifact({
      status: 'approved',
      gating: {
        isApproved: true,
        isPublished: false,
        canInitiateRevision: true,
        canCreateRevision: false,
      },
    });

    renderRoute();

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Iniciar revisão/i })).toBeInTheDocument();
    });
    expect(screen.queryByRole('button', { name: /Publicar/i })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Agendar/i })).not.toBeInTheDocument();
  });

  it('delegates the document.edit capability gate to useDocumentArtifact (not a route-level role check)', async () => {
    // FE-02: `document.edit` gating now lives entirely inside useDocumentArtifact
    // (see useDocumentArtifact.test.tsx for the capability-driven gating coverage).
    // The route only consumes `gating.canInitiateRevision`/`canCreateRevision` —
    // this test pins that the route renders the adapter's decision faithfully.
    hasCapMock.mockReturnValue(true);
    mockArtifact({
      status: 'published',
      gating: { canInitiateRevision: true, canCreateRevision: true },
    });

    renderRoute();

    await waitFor(() => {
      expect(useDocumentArtifact).toHaveBeenCalledWith('doc-1');
    });
    const btn = screen.getByRole('button', { name: /Iniciar revisão/i });
    expect(btn).toHaveAttribute('aria-disabled', 'false');
  });
});

describe('DocumentDetailRoute — D4 embedded PDF viewer', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    hasCapMock.mockReturnValue(true);
  });

  it('offers "Visualizar PDF" and "Baixar PDF" as two distinct actions', async () => {
    mockArtifact({ status: 'published' });

    renderRoute();

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Visualizar PDF' })).toBeInTheDocument();
    });
    expect(screen.getByRole('button', { name: 'Baixar PDF' })).toBeInTheDocument();
  });

  it('opens the embedded viewer instead of navigating away or downloading', async () => {
    mockArtifact({ status: 'published' });

    renderRoute();

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Visualizar PDF' })).toBeInTheDocument();
    });
    expect(screen.queryByTestId('pdf-viewer-dialog')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Visualizar PDF' }));

    expect(screen.getByTestId('pdf-viewer-dialog')).toHaveTextContent('POP-QUA-0148');
  });
});
