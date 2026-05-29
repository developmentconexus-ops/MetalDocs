import { render, screen } from '@testing-library/react';
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
  ID: 'doc-dist-1',
  TenantID: 'tenant-1',
  TemplateVersionID: 'template-1',
  Name: 'Procedimento Distribuído',
  Status: 'published',
  FormDataJSON: {},
  CurrentRevisionID: 'rev-1',
  RevisionVersion: 2,
  ActiveSessionID: '',
  ValuesFrozenAt: null,
  ArchivedAt: null,
  CreatedAt: '2026-05-19T00:00:00.000Z',
  UpdatedAt: '2026-05-19T00:00:00.000Z',
  CreatedBy: 'admin-user',
  RevisionNumber: 3,
  ControlledDocumentID: 'cd-1',
  RevisionTitle: 'Procedimento Distribuído',
  ProfileCodeSnapshot: 'pop',
  ProcessAreaCodeSnapshot: 'general',
  Code: 'POP-GENERAL-014',
  currentRevisionFileSizeBytes: 1024,
  currentRevisionPageCount: 1,
  currentRevisionPageCountSource: 'server_renderer',
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

    // No fabricated identity leaking through
    expect(screen.queryByText(/PR-EHS-014/)).toBeNull();
    expect(screen.queryByText(/LOTO/)).toBeNull();
    expect(screen.queryByText(/Vence em 8 dias/)).toBeNull();
    expect(screen.queryByText(/248 pessoas/)).toBeNull();
  });

  it('shows the "em breve" empty state instead of fabricated metrics', () => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: baseDoc,
      refetch: vi.fn(),
    } as never);

    render(<DocumentDistributionPage />);

    expect(
      screen.getByRole('heading', { name: /Distribuição e cobertura de leitura — em breve/ }),
    ).toBeInTheDocument();
    expect(screen.getByText(/Cobertura de leitura/)).toBeInTheDocument();
    expect(screen.getByText(/Curva de adoção/)).toBeInTheDocument();
    expect(screen.getByText(/Lista de destinatários/)).toBeInTheDocument();

    // None of the fabricated mock numbers should be present
    expect(screen.queryByText(/182\b/)).toBeNull();
    expect(screen.queryByText(/156\b/)).toBeNull();
    expect(screen.queryByText(/Manutenção/)).toBeNull();
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
