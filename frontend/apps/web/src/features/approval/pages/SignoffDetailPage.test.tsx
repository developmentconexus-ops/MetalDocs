import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { SignoffDetailPage } from './SignoffDetailPage';
import * as approvalApi from '../api/approvalApi';
import * as documentsApi from '../../documents/api/documents';

vi.mock('../components/ReviewDocumentCanvas', () => ({
  ReviewDocumentCanvas: (_props: unknown) => <div data-testid="review-canvas" />,
}));

function makeDoc(overrides: Partial<documentsApi.DocumentDetail> = {}) {
  return {
    id: 'doc-1',
    code: 'POP-QUA-0148',
    name: 'POP Limpeza de Linha',
    status: 'under_review',
    revision_version: 3,
    controlled_document_id: 'cd-1',
    current_revision_id: 'rev-1',
    created_by: 'user-approver-1',
    ...overrides,
  } as documentsApi.DocumentDetail;
}

function makeContext() {
  return {
    document_id: 'doc-1',
    approval_state: 'under_review',
    content_hash: 'hash-abc',
    revision_version: 3,
    approval_instance_id: 'inst-1',
  } as Awaited<ReturnType<typeof approvalApi.getActiveDocumentContext>>;
}

function renderAt(url = '/approvals/doc-1') {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route path="/approvals/:documentId" element={<SignoffDetailPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('SignoffDetailPage', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(documentsApi, 'getDocument').mockResolvedValue(makeDoc());
    vi.spyOn(approvalApi, 'getActiveDocumentContext').mockResolvedValue(makeContext());
    vi.spyOn(approvalApi, 'getInstance').mockResolvedValue({
      id: 'inst-1',
      document_id: 'doc-1',
      route_id: 'r1',
      tenant_id: 't1',
      status: 'in_progress',
      submitted_by: 'maria',
      submitted_at: '2026-04-14T10:00:00.000Z',
      stages: [],
      etag: '"v3"',
    } as Awaited<ReturnType<typeof approvalApi.getInstance>>);
    vi.spyOn(documentsApi, 'listComments').mockResolvedValue([]);
  });

  it('renders the document header from getDocument', async () => {
    renderAt();
    await waitFor(() => {
      expect(screen.getByText('POP Limpeza de Linha')).toBeTruthy();
      expect(screen.getByText('POP-QUA-0148')).toBeTruthy();
    });
  });

  it('mounts the decision panel (Assinar present for under_review)', async () => {
    renderAt();
    await waitFor(() => { expect(screen.getByText('Assinar')).toBeTruthy(); });
  });

  it('tab list has exactly two tabs: documento and comentarios', async () => {
    renderAt();
    await waitFor(() => {
      expect(screen.getByText('POP Limpeza de Linha')).toBeTruthy();
    });
    const tabs = screen.getAllByRole('tab');
    expect(tabs).toHaveLength(2);
    expect(tabs[0].textContent).toBe('Documento');
    expect(tabs[1].textContent).toBe('Comentários');
  });

  it('renders ReviewDocumentCanvas with documentId and currentRevisionId on documento tab', async () => {
    renderAt();
    await waitFor(() => {
      expect(screen.getByTestId('review-canvas')).toBeTruthy();
    });
  });

  it('shows the sidebar error state when the active-document-context query errors', async () => {
    vi.spyOn(approvalApi, 'getActiveDocumentContext').mockRejectedValue(
      new Error('context boom'),
    );
    renderAt();
    await waitFor(() => {
      expect(
        screen.getByText('Não foi possível carregar os dados de aprovação.'),
      ).toBeTruthy();
    });
  });

  it('renders flattened comment text on the Comentários tab', async () => {
    vi.spyOn(documentsApi, 'listComments').mockResolvedValue([
      {
        id: 'cmt-1',
        library_comment_id: 'lib-1',
        parent_library_id: null,
        author: 'João Silva',
        author_id: 'u-1',
        content: [
          { type: 'paragraph', content: [{ type: 'text', text: 'Revisar' }] },
          { type: 'paragraph', content: [{ type: 'text', text: 'a seção 3.' }] },
        ],
        done: false,
        created_at: '2026-04-14T10:00:00.000Z',
        updated_at: '2026-04-14T10:00:00.000Z',
        resolved_at: null,
      },
    ] as unknown as Awaited<ReturnType<typeof documentsApi.listComments>>);
    renderAt();
    await waitFor(() => {
      expect(screen.getByText('POP Limpeza de Linha')).toBeTruthy();
    });
    fireEvent.click(screen.getByRole('tab', { name: 'Comentários' }));
    await waitFor(() => {
      expect(screen.getByText('Revisar a seção 3.')).toBeTruthy();
    });
  });
});
