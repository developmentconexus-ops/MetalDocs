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
    subject_kind: 'document',
    subject_key: 'doc-1',
    subject_title: 'POP Limpeza',
    subject_ref: 'doc-1',
    controlled_document_id: 'cd-1',
    // Deliberately NOT the design-mock code ('POP-QUA-0148') that other
    // assertions here prove is no longer hardcoded in the timeline markup.
    controlled_document_code: 'POP-QUA-0777',
    area_code: 'QUA',
    submitted_by: 'maria',
    submitted_at: '2026-04-14T10:00:00.000Z',
    stage_label: 'Revisão L2',
    quorum_progress: '1/2',
    stage_kind: 'review',
    due_at: '2026-04-21T10:00:00.000Z',
    ...overrides,
  };
}

function makeTemplateItem(overrides: Partial<InboxItem> = {}): InboxItem {
  return makeItem({
    instance_id: 'inst-tpl-1',
    subject_kind: 'template',
    subject_key: 'tpl-version-1',
    subject_title: 'Modelo POP',
    subject_ref: 'tpl-1',
    controlled_document_id: null,
    controlled_document_code: null,
    ...overrides,
  });
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
      expect(
        screen.getByText(
          'Nenhuma aprovação pendente. Documentos submetidos a rotas onde você é revisor ou aprovador aparecem aqui.',
        ),
      ).toBeTruthy();
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
    const item1 = makeItem({ instance_id: 'i1', subject_title: 'POP Limpeza' });
    const item2 = makeItem({ instance_id: 'i2', subject_title: 'Manual Segurança' });

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
      makeItem({ instance_id: 'i1', subject_title: 'Doc 1' }),
      makeItem({ instance_id: 'i2', subject_title: 'Doc 2' }),
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
    // F-QA4-8: the row identifies the document by its canonical human code
    // (controlled_documents.code), never by the controlled-document uuid.
    expect(screen.getByText('POP-QUA-0777')).toBeTruthy();
    expect(screen.queryByText('cd-1')).toBeNull();
  });

  // NOTE: the three "Abrir documento navigates to the modern editor route" /
  // "keeps navigation in modern editor flow" / "shows modern-flow error..."
  // tests that lived here were deleted under F5 (M2c C3, single destination).
  // They asserted the exact bug C3 fixes — the primary worklist open landing
  // on `/documents/:id/edit` (the author editor) via fetchActiveDocumentInstance.
  // That vector is gone; see "primary open (Abrir documento) navigates to the
  // approval cockpit, not the editor" below, which supersedes them.

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
      expect(navigateMock).toHaveBeenCalledWith('/documents/doc-1?decision=approve');
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
      expect(navigateMock).toHaveBeenCalledWith('/documents/doc-9?decision=reject');
    });
  });

  // F5 (M2c C3): single-destination — primary item open must land on the
  // mode-adaptive workspace, never the author editor. This replaces the old
  // fetchActiveDocumentInstance -> /documents/:id/edit vector. F2d.5 S3
  // retired the standalone approval cockpit route in favor of the single
  // canonical artifact path (`/documents/:id`, ADR 0080).
  it('primary open (Abrir documento) navigates to the document workspace, not the editor', async () => {
    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [makeItem({ subject_ref: 'doc-cockpit', controlled_document_id: 'cd-cockpit' })], total: 1 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();
    fireEvent.click(screen.getByText('Abrir documento'));

    await waitFor(() => {
      expect(navigateMock).toHaveBeenCalledWith('/documents/doc-cockpit');
    });
    expect(fetchActiveDocumentInstance).not.toHaveBeenCalled();
    expect(navigateMock).not.toHaveBeenCalledWith(expect.stringContaining('/edit'));
  });

  // Unit 4.2 slice 2: a template row (subject_kind === 'template') opens the
  // template approval route instead of the document cockpit.
  it('template row open navigates to the template approval route', async () => {
    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [makeTemplateItem({ subject_ref: 'tpl-cockpit' })], total: 1 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();
    fireEvent.click(screen.getByText('Abrir revisão do modelo'));

    await waitFor(() => {
      expect(navigateMock).toHaveBeenCalledWith('/templates/tpl-cockpit/approval');
    });
    expect(fetchActiveDocumentInstance).not.toHaveBeenCalled();
  });

  // A template row's controlled_document_id is always null on the wire — the
  // document-only quick decision flow must never run for it.
  it('template row does not trigger the document quick-decision flow', async () => {
    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [makeTemplateItem()], total: 1 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();

    expect(screen.queryByRole('button', { name: /Aprovar e assinar/i })).toBeNull();
    expect(screen.queryByRole('button', { name: /Devolver/i })).toBeNull();
    expect(fetchActiveDocumentInstance).not.toHaveBeenCalled();
  });

  it('filter selection re-queries useInboxQuery with mapped params', async () => {
    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [], total: 0 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();

    fireEvent.change(screen.getByLabelText('Estágio'), { target: { value: 'review' } });

    await waitFor(() => {
      const lastCall = vi.mocked(useInboxQuery).mock.calls.at(-1)?.[0];
      expect(lastCall).toMatchObject({ stage_kind: 'review' });
    });
  });

  it('sorts items due-date ascending with null due_at last', async () => {
    const noDue = makeItem({ instance_id: 'i-none', subject_title: 'Sem prazo', due_at: null });
    const later = makeItem({ instance_id: 'i-later', subject_title: 'Prazo distante', due_at: '2026-08-01T00:00:00.000Z' });
    const sooner = makeItem({ instance_id: 'i-sooner', subject_title: 'Prazo próximo', due_at: '2026-04-15T00:00:00.000Z' });

    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [noDue, later, sooner], total: 3 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();
    fireEvent.click(screen.getByText('Linha do tempo'));

    await waitFor(() => {
      const titles = screen.getAllByText(/Prazo|Sem prazo/).map((el) => el.textContent);
      const sortedIdx = {
        sooner: titles.indexOf('Prazo próximo'),
        later: titles.indexOf('Prazo distante'),
        none: titles.indexOf('Sem prazo'),
      };
      expect(sortedIdx.sooner).toBeLessThan(sortedIdx.later);
      expect(sortedIdx.later).toBeLessThan(sortedIdx.none);
    });
  });

  it('shows the no-work teaching empty state when there is no filter active', async () => {
    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [], total: 0 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();

    await waitFor(() => {
      expect(
        screen.getByText(
          'Nenhuma aprovação pendente. Documentos submetidos a rotas onde você é revisor ou aprovador aparecem aqui.',
        ),
      ).toBeTruthy();
    });
  });

  it('shows the filtered-empty teaching state when a filter is active and yields nothing', async () => {
    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [], total: 0 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();

    fireEvent.change(screen.getByLabelText('Estágio'), { target: { value: 'review' } });

    await waitFor(() => {
      expect(screen.getByText('Nenhuma aprovação corresponde aos filtros.')).toBeTruthy();
    });
  });

  it('renders an overdue chip for a past due_at', async () => {
    const overdueItem = makeItem({
      instance_id: 'i-overdue',
      subject_title: 'Doc atrasado',
      due_at: '2020-01-01T00:00:00.000Z',
    });

    vi.mocked(useInboxQuery).mockReturnValue({
      data: { items: [overdueItem], total: 1 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as unknown as ReturnType<typeof useInboxQuery>);

    renderPage();

    await waitFor(() => {
      expect(screen.getByText(/atrasado há/)).toBeTruthy();
    });
  });

  it('reverts the oversee toggle and shows a note on a 403 while oversee scope is active', async () => {
    const { ApiError } = await import('../../../lib/api/errors');
    vi.mocked(useInboxQuery).mockImplementation((params) => {
      if (params?.scope === 'oversee') {
        return {
          data: undefined,
          isLoading: false,
          isError: true,
          error: new ApiError('forbidden', 403, 'Forbidden'),
          refetch: vi.fn(),
        } as unknown as ReturnType<typeof useInboxQuery>;
      }
      return {
        data: { items: [], total: 0 },
        isLoading: false,
        isError: false,
        refetch: vi.fn(),
      } as unknown as ReturnType<typeof useInboxQuery>;
    });

    renderPage();

    fireEvent.click(screen.getByLabelText(/Supervisão/i));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Você não tem permissão de supervisão.');
    });

    await waitFor(() => {
      expect(screen.getByLabelText(/Supervisão/i)).not.toBeChecked();
    });
  });

  it('clears the overseeDenied note when the user re-enables oversee', async () => {
    const { ApiError } = await import('../../../lib/api/errors');
    let overseeShouldFail = true;
    vi.mocked(useInboxQuery).mockImplementation((params) => {
      if (params?.scope === 'oversee') {
        if (overseeShouldFail) {
          return {
            data: undefined,
            isLoading: false,
            isError: true,
            error: new ApiError('forbidden', 403, 'Forbidden'),
            refetch: vi.fn(),
          } as unknown as ReturnType<typeof useInboxQuery>;
        }
        return {
          data: { items: [], total: 0 },
          isLoading: false,
          isError: false,
          refetch: vi.fn(),
        } as unknown as ReturnType<typeof useInboxQuery>;
      }
      return {
        data: { items: [], total: 0 },
        isLoading: false,
        isError: false,
        refetch: vi.fn(),
      } as unknown as ReturnType<typeof useInboxQuery>;
    });

    renderPage();

    // First toggle: oversee query 403s, note shows, checkbox reverts.
    fireEvent.click(screen.getByLabelText(/Supervisão/i));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Você não tem permissão de supervisão.');
    });
    await waitFor(() => {
      expect(screen.getByLabelText(/Supervisão/i)).not.toBeChecked();
    });

    // Re-enable oversee — this time the fetch succeeds. The stale denial note
    // must clear (F5 Minor #1 / F7): no permanent one-way ratchet.
    overseeShouldFail = false;
    fireEvent.click(screen.getByLabelText(/Supervisão/i));

    await waitFor(() => {
      expect(screen.queryByText('Você não tem permissão de supervisão.')).toBeNull();
    });
  });
});

