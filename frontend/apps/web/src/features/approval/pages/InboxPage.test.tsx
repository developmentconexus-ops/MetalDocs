import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { InboxPage } from './InboxPage';

import * as approvalApi from '../api/approvalApi';
import * as taxonomyApi from '../../taxonomy/api';

const navigateMock = vi.fn();

vi.mock('../api/approvalApi');
vi.mock('../../taxonomy/api', () => ({
  fetchAreas: vi.fn().mockResolvedValue([]),
}));
vi.mock('react-router-dom', () => ({
  useNavigate: () => navigateMock,
}));

function createDeferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

describe('InboxPage', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    // Restore default: areas effect must not throw
    vi.mocked(taxonomyApi.fetchAreas).mockResolvedValue([]);
  });

  it('loading state', () => {
    const deferred = createDeferred<{ items: []; total: number }>();
    vi.mocked(approvalApi.listInbox).mockReturnValue(deferred.promise);

    render(<InboxPage />);

    expect(screen.getByText('Carregando...')).toBeTruthy();
  });

  it('empty state', async () => {
    vi.mocked(approvalApi.listInbox).mockResolvedValue({ items: [], total: 0 });

    render(<InboxPage />);

    await waitFor(() => {
      expect(screen.getByText('Nada pendente para revisão.')).toBeTruthy();
    });
  });

  it('renders 3 rows with correct columns', async () => {
    vi.mocked(approvalApi.listInbox).mockResolvedValue({
      total: 3,
      items: [
        {
          instance_id: 'inst-1',
          document_id: 'doc-1',
          controlled_document_id: 'cd-1',
          document_title: 'POP Limpeza',
          area_code: 'JUR',
          submitted_by: 'maria',
          submitted_at: '2026-04-14T10:00:00.000Z',
          stage_label: 'Revisão Jurídica',
          quorum_progress: '1/2',
        },
        {
          instance_id: 'inst-2',
          document_id: 'doc-2',
          controlled_document_id: 'cd-2',
          document_title: 'Manual Segurança',
          area_code: 'RH',
          submitted_by: 'ana',
          submitted_at: '2026-04-14T11:00:00.000Z',
          stage_label: 'Revisão RH',
          quorum_progress: '0/1',
        },
        {
          instance_id: 'inst-3',
          document_id: 'doc-3',
          controlled_document_id: 'cd-3',
          document_title: 'Instrução Técnica',
          area_code: 'ENG',
          submitted_by: 'joao',
          submitted_at: '2026-04-14T12:00:00.000Z',
          stage_label: 'Validação Engenharia',
          quorum_progress: '2/3',
        },
      ],
    });

    render(<InboxPage />);

    await waitFor(() => {
      expect(screen.getByText('POP Limpeza')).toBeTruthy();
      expect(screen.getByText('Manual Segurança')).toBeTruthy();
      expect(screen.getByText('Instrução Técnica')).toBeTruthy();
      expect(screen.getByText('Revisão Jurídica')).toBeTruthy();
    });
  });

  it('click row calls navigate to /documents/:id', async () => {
    vi.mocked(approvalApi.listInbox).mockResolvedValue({
      total: 1,
      items: [
        {
          instance_id: 'inst-1',
          document_id: 'doc-1',
          controlled_document_id: 'cd-1',
          document_title: 'Documento A',
          area_code: 'JUR',
          submitted_by: 'maria',
          submitted_at: '2026-04-14T10:00:00.000Z',
          stage_label: 'Revisão Jurídica',
          quorum_progress: '1/1',
        },
      ],
    });

    render(<InboxPage />);

    const rowCell = await screen.findByText('Documento A');
    fireEvent.click(rowCell.closest('tr') as HTMLElement);

    expect(navigateMock).toHaveBeenCalledWith('/registry-v2/cd-1');
  });

  it('filter by area re-fetches', async () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    vi.mocked(taxonomyApi.fetchAreas).mockResolvedValue([{ code: 'JUR', name: 'Jurídico', archived: false } as any]);
    vi.mocked(approvalApi.listInbox).mockResolvedValue({ items: [], total: 0 });
    render(<InboxPage />);

    await waitFor(() => {
      expect(vi.mocked(approvalApi.listInbox)).toHaveBeenCalledWith({
        area_code: undefined,
        offset: 0,
        limit: 20,
      });
    });

    // Wait for areas to load so 'JUR' option exists before firing change
    await waitFor(() => {
      expect(screen.getByRole('option', { name: /JUR/i })).toBeTruthy();
    });

    fireEvent.change(screen.getByLabelText('Área'), { target: { value: 'JUR' } });

    await waitFor(() => {
      expect(vi.mocked(approvalApi.listInbox)).toHaveBeenCalledWith({
        area_code: 'JUR',
        offset: 0,
        limit: 20,
      });
    });
  });

  it('populates area filter from fetchAreas, not hardcoded list', async () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    vi.mocked(taxonomyApi.fetchAreas).mockResolvedValue([
      { code: 'OPS', name: 'Operações', archived: false } as any,
      { code: 'QA', name: 'Qualidade', archived: false } as any,
    ]);
    vi.mocked(approvalApi.listInbox).mockResolvedValue({ items: [], total: 0 });

    render(<InboxPage />);

    await waitFor(() => {
      expect(screen.getByRole('option', { name: /Todas as áreas/i })).toBeTruthy();
    });
    expect(screen.getByRole('option', { name: /OPS — Operações/ })).toBeTruthy();
    expect(screen.getByRole('option', { name: /QA — Qualidade/ })).toBeTruthy();
    expect(screen.queryByRole('option', { name: /JUR/ })).toBeNull();
  });
});
