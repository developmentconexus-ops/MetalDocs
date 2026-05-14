import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { getNextSelectedIdx, InboxPage } from './InboxPage';
import type { InboxItem } from '../api/approvalTypes';
import { getActiveDocumentContext } from '../api/approvalApi';

const navigateMock = vi.fn();

vi.mock('react-router-dom', () => ({
  useNavigate: () => navigateMock,
}));

vi.mock('../queries/useInboxQuery', () => ({
  useInboxQuery: vi.fn(),
}));

vi.mock('../api/approvalApi', async () => {
  const actual = await vi.importActual<typeof import('../api/approvalApi')>('../api/approvalApi');
  return {
    ...actual,
    getActiveDocumentContext: vi.fn(),
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
    // Default localStorage clear
    localStorage.removeItem('md.inbox.v');
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
    localStorage.setItem('md.inbox.v', 'deadline-mock');

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

    expect(localStorage.getItem('md.inbox.v')).toBe('timeline');
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

  it('Abrir documento navigates to the registry detail route', async () => {
    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [makeItem({ controlled_document_id: 'cd-123' })], total: 1 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();
    fireEvent.click(screen.getByText('Abrir documento'));

    expect(navigateMock).toHaveBeenCalledWith('/registry-v2/cd-123');
  });

  it('approve action opens signoff flow only when active-document context is complete', async () => {
    vi.mocked(getActiveDocumentContext).mockResolvedValue({
      documentId: 'doc-1',
      contentHash: 'hash-1',
      approvalInstanceId: 'inst-1',
    } as Awaited<ReturnType<typeof getActiveDocumentContext>>);
    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [makeItem({ controlled_document_id: 'cd-123' })], total: 1 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();
    fireEvent.click(screen.getByText('Aprovar e assinar →'));

    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeTruthy();
      expect(screen.getByText(/Assinar/)).toBeTruthy();
    });
  });
});
