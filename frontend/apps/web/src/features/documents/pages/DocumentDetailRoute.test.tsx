// FE-11: revision-initiate gating is capability-based (ADR 0022 — capabilities,
// never roles). This test drives the route-level hero action button (the only
// place `canInitiateRevision`/`canCreateRevision`/`canPublish` become observable
// UI) through `useHasCapability('document.edit')` instead of a role check.
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { DocumentDetailRoute } from './DocumentDetailRoute';

const hasCapMock = vi.fn();
vi.mock('../../iam/hooks/useHasCapability', () => ({
  useHasCapability: (cap: string) => hasCapMock(cap),
}));

vi.mock('../adapters/useDocumentArtifact', () => ({
  useDocumentArtifact: vi.fn(),
}));
vi.mock('../queries/useDocumentDetailQuery', () => ({
  useDocumentDetailQuery: vi.fn(),
}));
vi.mock('../queries/useApprovalInstanceQuery', () => ({
  useApprovalInstanceQuery: vi.fn(),
}));
vi.mock('../queries/useControlledDocumentActiveDocumentQuery', () => ({
  useControlledDocumentActiveDocumentQuery: vi.fn(),
}));
vi.mock('../queries/useDistributionSummaryQuery', () => ({
  useDistributionSummaryQuery: vi.fn(),
}));

import { useDocumentArtifact } from '../adapters/useDocumentArtifact';
import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
import { useApprovalInstanceQuery } from '../queries/useApprovalInstanceQuery';
import { useControlledDocumentActiveDocumentQuery } from '../queries/useControlledDocumentActiveDocumentQuery';
import { useDistributionSummaryQuery } from '../queries/useDistributionSummaryQuery';

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
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: BASE_DOC,
      refetch: vi.fn(),
    } as never);
    vi.mocked(useApprovalInstanceQuery).mockReturnValue({
      data: { completed_at: '2026-05-19T23:39:00.000Z', stages: [] },
      isLoading: false,
      refetch: vi.fn(),
    } as never);
    vi.mocked(useControlledDocumentActiveDocumentQuery).mockReturnValue({
      data: { document_id: 'doc-1', content_hash: 'hash-1', approval_state: 'published' },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as never);
    vi.mocked(useDistributionSummaryQuery).mockReturnValue({
      data: { total_targets: 3 },
      isError: false,
      isLoading: false,
    } as never);
    vi.mocked(useDocumentArtifact).mockReturnValue({
      model: baseModel('published'),
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as never);
  });

  it('enables "Iniciar revisão" when the user holds document.edit', async () => {
    hasCapMock.mockImplementation((cap: string) => cap === 'document.edit');

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

    renderRoute();

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Iniciar revisão/i })).toBeInTheDocument();
    });
    const btn = screen.getByRole('button', { name: /Iniciar revisão/i });
    expect(btn).toHaveAttribute('aria-disabled', 'true');
    expect(btn).toHaveAttribute('title', 'Sem permissão para iniciar revisão');
  });

  it('gates "Publicar / Agendar" on document.edit for approved documents', async () => {
    hasCapMock.mockReturnValue(false);
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: { ...BASE_DOC, status: 'approved' },
      refetch: vi.fn(),
    } as never);
    vi.mocked(useDocumentArtifact).mockReturnValue({
      model: baseModel('approved'),
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as never);

    renderRoute();

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Publicar \/ Agendar/i })).toBeInTheDocument();
    });
    const btn = screen.getByRole('button', { name: /Publicar \/ Agendar/i });
    expect(btn).toHaveAttribute('aria-disabled', 'true');
    expect(btn).toHaveAttribute('title', 'Sem permissão para publicar');
  });

  it('queries the document.edit capability (not roles) for the revision gate', async () => {
    hasCapMock.mockReturnValue(true);

    renderRoute();

    await waitFor(() => {
      expect(hasCapMock).toHaveBeenCalledWith('document.edit');
    });
  });
});
