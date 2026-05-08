import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { InboxPage } from './InboxPage';
import type { RichInboxItem } from '../lib/mockInboxData';

vi.mock('react-router-dom', () => ({
  useNavigate: () => vi.fn(),
}));

vi.mock('../queries/useInboxQuery', () => ({
  useInboxQuery: vi.fn(),
}));

import { useInboxQuery } from '../queries/useInboxQuery';

function makeItem(overrides: Partial<RichInboxItem> = {}): RichInboxItem {
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
    // RichInboxItem mock fields
    code: 'POP-QUA-0001',
    kind: 'POP',
    deadline: '3h 28min',
    urgent: false,
    summary: 'Resumo do documento.',
    changes: 5,
    version: 'v1.0 → v1.1',
    deadline_at: '2026-05-09T10:00:00.000Z',
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

  it('empty API response falls back to MOCK_INBOX_ITEMS (4 items)', async () => {
    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [], total: 0 },
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();
    // Mock fallback has 4 items — counter shows 01 / 04
    await waitFor(() => {
      expect(screen.getByText('01 / 04')).toBeTruthy();
    });
  });

  it('renders queue items from API data', async () => {
    const item1 = makeItem({ instance_id: 'i1', document_title: 'POP Limpeza', code: 'POP-QUA-0001' });
    const item2 = makeItem({ instance_id: 'i2', document_title: 'Manual Segurança', code: 'IT-PROD-0002' });

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    vi.mocked(useInboxQuery).mockReturnValue({ data: { items: [item1, item2], total: 2 }, isLoading: false, isError: false } as any);

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
});
