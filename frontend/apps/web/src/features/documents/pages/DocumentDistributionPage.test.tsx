import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

import { DocumentDistributionPage } from './DocumentDistributionPage';

vi.mock('react-router-dom', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-router-dom')>();
  return {
    ...actual,
    useParams: () => ({ documentId: 'doc-dist-1' }),
  };
});

vi.mock('../queries/useDocumentDetailQuery', () => ({
  useDocumentDetailQuery: vi.fn(),
}));

vi.mock('../api/distribution', () => ({
  getDistributionSummary: vi.fn().mockResolvedValue({ total_targets: 0 }),
  listDistributionRecipients: vi.fn().mockResolvedValue({
    items: [],
    page: { next_cursor: null, has_more: false },
  }),
  getDistributionCoverage: vi.fn().mockResolvedValue([]),
}));

import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
import { getDistributionSummary } from '../api/distribution';

const baseDoc = {
  id: 'doc-dist-1',
  tenant_id: 'tenant-1',
  template_version_id: 'template-1',
  name: 'Procedimento Distribuído',
  status: 'published',
  form_data_json: {},
  current_revision_id: 'rev-1',
  revision_version: 2,
  active_session_id: '',
  values_frozen_at: null,
  archived_at: null,
  created_at: '2026-05-19T00:00:00.000Z',
  updated_at: '2026-05-19T00:00:00.000Z',
  created_by: 'admin-user',
  revision_number: 3,
  controlled_document_id: 'cd-1',
  revision_title: 'Procedimento Distribuído',
  profile_code_snapshot: 'pop',
  process_area_code_snapshot: 'general',
  code: 'POP-GENERAL-014',
  current_revision_file_size_bytes: 1024,
  current_revision_page_count: 1,
  current_revision_page_count_source: 'server_renderer',
};

function wrap() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  // MemoryRouter needed because ArtifactHero now uses react-router Link
  const { MemoryRouter } = require('react-router-dom') as typeof import('react-router-dom');
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(
      MemoryRouter,
      null,
      React.createElement(QueryClientProvider, { client: qc }, children),
    );
}

describe('DocumentDistributionPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('shows loading state', () => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: true,
      isError: false,
      data: undefined,
      refetch: vi.fn(),
    } as never);

    render(<DocumentDistributionPage />, { wrapper: wrap() });

    expect(screen.getByText(/Carregando documento/)).toBeInTheDocument();
  });

  it('shows error state with retry', () => {
    const refetch = vi.fn();
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: true,
      data: undefined,
      refetch,
    } as never);

    render(<DocumentDistributionPage />, { wrapper: wrap() });

    expect(screen.getByRole('alert')).toHaveTextContent(/não encontrado/);
    screen.getByRole('button', { name: /Tentar novamente/ }).click();
    expect(refetch).toHaveBeenCalledOnce();
  });

  it('renders real document identity from the query (no hardcoded PR-EHS-014)', () => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: baseDoc,
      refetch: vi.fn(),
    } as never);

    render(<DocumentDistributionPage />, { wrapper: wrap() });

    // Real code + version label from the API payload
    expect(screen.getAllByText(/POP-GENERAL-014/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/REV03/).length).toBeGreaterThan(0);
    expect(screen.getByRole('heading', { name: /Distribuição & cobertura de leitura/ })).toBeInTheDocument();

    // No hardcoded LOTO/PR-EHS identity anywhere
    expect(screen.queryByText(/PR-EHS-014/)).toBeNull();
    expect(screen.queryByText(/LOTO/)).toBeNull();
  });

  it('has no illustrative watermark or Em-breve banner', () => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: baseDoc,
      refetch: vi.fn(),
    } as never);

    render(<DocumentDistributionPage />, { wrapper: wrap() });

    // The old watermarks must be gone
    expect(screen.queryByText(/Dados ilustrativos · Em breve/)).toBeNull();
    // No "ilustrativos" banner
    expect(screen.queryByText(/todos os números.*são.*ilustrativos/i)).toBeNull();
  });

  it('hero CTAs remain disabled with deferred trigger note', () => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: baseDoc,
      refetch: vi.fn(),
    } as never);

    render(<DocumentDistributionPage />, { wrapper: wrap() });

    const ctaLabels = ['Lembrete em massa', 'Exportar relatório', 'Adicionar destinatários', 'Política de fanout'];
    for (const label of ctaLabels) {
      const btn = screen.getByRole('button', { name: new RegExp(label) });
      expect(btn.getAttribute('aria-disabled')).toBe('true');
      // Title references the parked mission instead of "Em breve"
      expect(btn.getAttribute('title')).toContain('distribuição');
    }
  });

  it('shows em-dash (never a fabricated 0) when the summary query errors', async () => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: baseDoc,
      refetch: vi.fn(),
    } as never);
    // Only the denominator (total_targets) producer fails; doc identity still loads.
    vi.mocked(getDistributionSummary).mockRejectedValueOnce(new Error('boom'));

    render(<DocumentDistributionPage />, { wrapper: wrap() });

    // Honest "—", not a fake 0, on the "Alvos totais" KPI cell.
    expect(await screen.findAllByText('—')).not.toHaveLength(0);
    expect(screen.queryByText(/^0$/)).toBeNull();
  });

  it('renders the heading "Distribuição & cobertura de leitura"', () => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: baseDoc,
      refetch: vi.fn(),
    } as never);

    render(<DocumentDistributionPage />, { wrapper: wrap() });

    expect(screen.getByRole('heading', { name: /Distribuição & cobertura de leitura/ })).toBeInTheDocument();
  });
});
