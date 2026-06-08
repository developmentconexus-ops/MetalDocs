import { render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { DocumentDistributionPage } from './DocumentDistributionPage';

vi.mock('react-router-dom', () => ({
  useParams: () => ({ documentId: 'doc-dist-1' }),
}));

vi.mock('../queries/useDocumentDetailQuery', () => ({
  useDocumentDetailQuery: vi.fn(),
}));

import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';

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

    render(<DocumentDistributionPage />);

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

    render(<DocumentDistributionPage />);

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

    render(<DocumentDistributionPage />);

    // Real code + version label from the API payload
    expect(screen.getAllByText(/POP-GENERAL-014/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/REV03/).length).toBeGreaterThan(0);
    expect(screen.getByRole('heading', { name: /Distribuição & cobertura de leitura/ })).toBeInTheDocument();

    // The hardcoded LOTO identity must NOT appear in the live (non-illustrative)
    // hero / breadcrumb / DocRefCard. (It MAY still appear inside the
    // aria-hidden illustrative scaffolding — that's intentional design preview.)
    const liveBanner = screen.getByRole('note');
    expect(within(liveBanner).queryByText(/PR-EHS-014/)).toBeNull();
    expect(within(liveBanner).queryByText(/LOTO/)).toBeNull();
  });

  it('renders the em-breve banner above the illustrative scaffolding', () => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: baseDoc,
      refetch: vi.fn(),
    } as never);

    render(<DocumentDistributionPage />);

    const banner = screen.getByRole('note');
    expect(banner).toHaveTextContent(/em breve/i);
    expect(banner).toHaveTextContent(/ilustrativos/i);
    // Banner names the real document — proves identity flows into the message
    expect(banner).toHaveTextContent(/Procedimento Distribuído/);
  });

  it('marks every illustrative section with the "Dados ilustrativos · Em breve" watermark and aria-hidden', () => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: baseDoc,
      refetch: vi.fn(),
    } as never);

    render(<DocumentDistributionPage />);

    const watermarks = screen.getAllByText(/Dados ilustrativos · Em breve/);
    // 5 scaffolded blocks: KPIStrip, Donut+Facts, CoverageByArea, Timeline, Recipients
    expect(watermarks.length).toBe(5);
    for (const wm of watermarks) {
      const illustrative = wm.parentElement;
      expect(illustrative?.getAttribute('aria-hidden')).toBe('true');
    }
  });

  it('hero CTAs remain disabled with the "Em breve" affordance', () => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: baseDoc,
      refetch: vi.fn(),
    } as never);

    render(<DocumentDistributionPage />);

    const ctaLabels = ['Lembrete em massa', 'Exportar relatório', 'Adicionar destinatários', 'Política de fanout'];
    for (const label of ctaLabels) {
      const btn = screen.getByRole('button', { name: new RegExp(label) });
      expect(btn.getAttribute('aria-disabled')).toBe('true');
      expect(btn.getAttribute('title')).toBe('Em breve');
    }
  });
});
