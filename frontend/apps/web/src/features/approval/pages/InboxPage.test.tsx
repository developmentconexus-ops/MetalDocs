// @vitest-environment jsdom
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { getNextSelectedIdx, InboxPage } from './InboxPage';
import type { InboxItem } from '../api/approvalTypes';
import { fetchActiveDocumentInstance } from '../../controlled-documents/api/controlledDocuments';

const navigateMock = vi.fn();

vi.mock('react-router-dom', () => ({
  useNavigate: () => navigateMock,
}));

vi.mock('../queries/useInboxQuery', () => ({
  useInboxQuery: vi.fn(),
}));

vi.mock('../../controlled-documents/api/controlledDocuments', async () => {
  const actual = await vi.importActual<typeof import('../../controlled-documents/api/controlledDocuments')>(
    '../../controlled-documents/api/controlledDocuments',
  );
  return {
    ...actual,
    fetchActiveDocumentInstance: vi.fn(),
  };
});

import { useInboxQuery } from '../queries/useInboxQuery';

function makeItem(overrides: Partial<InboxItem> = {}): InboxItem {
  return {
    instance_id: 'inst-1',
    document_id: 'doc-1',
    controlled_document_id: 'cd-1',
    document_title: 'POP Limpeza',
    area_code: 'QUA',
    submitted_by: 'maria',
    submitted_at: '2026-04-14T10:00:00.000Z',
    stage_label: 'Revisão L2',
    quorum_progress: '1/2',
    ...overrides,
  };
}

function renderPage() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <InboxPage />
    </QueryClientProvider>,
  );
}

describe('InboxPage', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    navigateMock.mockReset();
    // Default localStorage clear — use window.localStorage to avoid Node v26 global shadowing
    window.localStorage.removeItem('md.inbox.v');
  });

  it('loading state shows Carregando', () => {
    vi.mocked(useInboxQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();
    expect(screen.getByText('Carregando...')).toBeTruthy();
  });

  it('error state shows error alert', async () => {
    vi.mocked(useInboxQuery).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();
    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeTruthy();
      expect(screen.getByText('Erro ao carregar aprovações.')).toBeTruthy();
    });
  });

  it('empty API response renders honest empty state instead of mock fallback', async () => {
    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [], total: 0 },
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();
    await waitFor(() => {
      expect(screen.getByText('Nenhuma aprovação pendente.')).toBeTruthy();
    });

    expect(screen.queryByText('POP-QUA-0148')).toBeNull();
    expect(screen.queryByText('01 / 04')).toBeNull();
  });

  it('invalid persisted view falls back to stack', async () => {
    window.localStorage.setItem('md.inbox.v', 'deadline-mock');

    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [makeItem()], total: 1 },
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();

    await waitFor(() => {
      expect(screen.getByText('01 / 01')).toBeTruthy();
    });

    expect(screen.queryByText('Suas 1 decisões na ordem que importam')).toBeNull();
  });

  it('renders queue items from API data', async () => {
    const item1 = makeItem({ instance_id: 'i1', document_title: 'POP Limpeza' });
    const item2 = makeItem({ instance_id: 'i2', document_title: 'Manual Segurança' });

    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [item1, item2], total: 2 },
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();

    await waitFor(() => {
      expect(screen.getAllByText('POP Limpeza').length).toBeGreaterThan(0);
      expect(screen.getByText('Manual Segurança')).toBeTruthy();
    });
  });

  it('view switcher persists to localStorage', async () => {
    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [], total: 0 },
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();

    const timelineBtn = screen.getByText('Linha do tempo');
    fireEvent.click(timelineBtn);

    expect(window.localStorage.getItem('md.inbox.v')).toBe('timeline');
  });

  it('next/prev navigation updates counter', async () => {
    const items = [
      makeItem({ instance_id: 'i1', document_title: 'Doc 1' }),
      makeItem({ instance_id: 'i2', document_title: 'Doc 2' }),
    ];

    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items, total: 2 },
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();

    await waitFor(() => expect(screen.getByText('01 / 02')).toBeTruthy());

    const nextBtn = screen.getByText('próximo →');
    fireEvent.click(nextBtn);

    expect(screen.getByText('02 / 02')).toBeTruthy();
  });

  it('next navigation does not underflow on empty list', () => {
    expect(getNextSelectedIdx(0, 0)).toBe(0);
  });

  it('timeline view renders submitted-time review stream and unavailable heatmap state', async () => {
    vi.mocked(useInboxQuery).mockReturnValue({
      data: {
        items: [
          makeItem({
            submitted_at: new Date().toISOString(),
            stage_label: 'Revisão L2',
            quorum_progress: '1/2',
          }),
        ],
        total: 1,
      },
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();
    fireEvent.click(screen.getByText('Linha do tempo'));

    await waitFor(() => {
      expect(screen.getByText('Histórico de decisões ainda não disponível')).toBeTruthy();
    });

    expect(screen.getByText('Revisão L2')).toBeTruthy();
    expect(screen.queryByText('3h 28min')).toBeNull();
    expect(screen.queryByText('POP-QUA-0148')).toBeNull();
  });

  it('Abrir documento navigates to the modern editor route', async () => {
    vi.mocked(fetchActiveDocumentInstance).mockResolvedValue({
      document_id: 'doc-123',
      content_hash: 'hash-123',
      approval_instance_id: 'inst-123',
    } as Awaited<ReturnType<typeof fetchActiveDocumentInstance>>);
    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [makeItem({ controlled_document_id: 'cd-123' })], total: 1 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();
    fireEvent.click(screen.getByText('Abrir documento'));

    await waitFor(() => {
      expect(fetchActiveDocumentInstance).toHaveBeenCalledWith('cd-123');
    });
    expect(navigateMock).toHaveBeenCalledWith('/documents/doc-123/edit');
  });

  it('Abrir documento keeps navigation in modern editor flow', async () => {
    vi.mocked(fetchActiveDocumentInstance).mockResolvedValue({
      document_id: 'doc-modern',
      content_hash: 'hash-modern',
      approval_instance_id: 'inst-modern',
    } as Awaited<ReturnType<typeof fetchActiveDocumentInstance>>);
    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [makeItem({ controlled_document_id: 'cd-modern' })], total: 1 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();
    fireEvent.click(screen.getByText('Abrir documento'));

    await waitFor(() => {
      expect(fetchActiveDocumentInstance).toHaveBeenCalledWith('cd-modern');
    });
    expect(navigateMock).toHaveBeenCalledWith('/documents/doc-modern/edit');
    expect(navigateMock).not.toHaveBeenCalledWith('/controlled-documents/cd-modern');
  });

  it('Abrir documento shows modern-flow error when active context is unavailable', async () => {
    vi.mocked(fetchActiveDocumentInstance).mockRejectedValue(new Error('context unavailable'));
    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [makeItem({ controlled_document_id: 'cd-fail' })], total: 1 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();
    fireEvent.click(screen.getByText('Abrir documento'));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Documento indisponivel no editor moderno no momento.');
    });
    expect(navigateMock).not.toHaveBeenCalledWith('/controlled-documents/cd-fail');
  });

  it('approve action navigates to the signoff cockpit with decision=approve', async () => {
    vi.mocked(fetchActiveDocumentInstance).mockResolvedValue({
      document_id: 'doc-1',
      content_hash: 'hash-1',
      approval_instance_id: 'inst-1',
      revision_version: 0,
    } as Awaited<ReturnType<typeof fetchActiveDocumentInstance>>);
    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [makeItem({ controlled_document_id: 'cd-123' })], total: 1 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /Aprovar e assinar/i }));

    await waitFor(() => {
      expect(fetchActiveDocumentInstance).toHaveBeenCalledWith('cd-123');
      expect(navigateMock).toHaveBeenCalledWith('/approvals/doc-1?decision=approve');
    });
    expect(screen.queryByRole('dialog')).toBeNull();
  });

  it('approve action shows alert when active-document lookup fails', async () => {
    vi.mocked(fetchActiveDocumentInstance).mockRejectedValue(new Error('network failure'));
    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [makeItem({ controlled_document_id: 'cd-123' })], total: 1 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /Aprovar e assinar/i }));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeTruthy();
      expect(screen.getByText('Fluxo de aprovação indisponível para este documento no momento.')).toBeTruthy();
    });
    expect(screen.queryByRole('dialog')).toBeNull();
  });

  it('reject action navigates to the cockpit with decision=reject', async () => {
    vi.mocked(fetchActiveDocumentInstance).mockResolvedValue({
      document_id: 'doc-9',
      content_hash: 'hash-9',
      approval_instance_id: 'inst-9',
      revision_version: 2,
    } as Awaited<ReturnType<typeof fetchActiveDocumentInstance>>);
    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [makeItem({ controlled_document_id: 'cd-9' })], total: 1 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();
    fireEvent.click(screen.getByRole('button', { name: /Devolver/i }));

    await waitFor(() => {
      expect(navigateMock).toHaveBeenCalledWith('/approvals/doc-9?decision=reject');
    });
  });
});

