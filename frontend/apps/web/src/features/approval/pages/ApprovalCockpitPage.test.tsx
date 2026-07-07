import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ApprovalCockpitPage } from './ApprovalCockpitPage';
import * as approvalApi from '../api/approvalApi';
import * as controlledDocumentsApi from '../../controlled-documents/api/controlledDocuments';
import * as documentsApi from '../../documents/api/documents';

const mockState = vi.hoisted(() => ({
  sessionSpy: vi.fn(),
  autosaveSpy: vi.fn(),
}));

vi.mock('@metaldocs/editor-ui', () => ({
  MetalDocsEditor: React.forwardRef((props: Record<string, unknown>, ref) => {
    React.useImperativeHandle(ref, () => ({
      async saveNow() { return null; },
      async getDocumentBuffer() { return null; },
      getPageCount() { return 0; },
      focus() {},
    }));
    return <div data-testid="editor" data-mode={props.mode as string} />;
  }),
}));

// W2 fix: the approval feature must never mount a writer session or autosave.
// These mocks assert call-count 0 across all cases below.
vi.mock('../../documents/hooks/editor/useDocumentSession', () => ({
  useDocumentSession: mockState.sessionSpy,
}));
vi.mock('../../documents/hooks/editor/useDocumentAutosave', () => ({
  useDocumentAutosave: mockState.autosaveSpy,
}));

vi.mock('../../../store/auth.store', () => ({
  useAuthStore: (selector: (s: { user: { displayName: string; userId: string } }) => unknown) =>
    selector({ user: { displayName: 'Ana Revisora', userId: 'user-approver-1' } }),
}));

function makeDoc(overrides: Partial<documentsApi.DocumentDetail> = {}) {
  return {
    id: 'doc-1',
    code: 'POP-QUA-0148',
    name: 'POP Limpeza de Linha',
    status: 'under_review',
    revision_version: 3,
    revision_number: 3,
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
  } as Awaited<ReturnType<typeof controlledDocumentsApi.fetchActiveDocumentInstance>>;
}

function makeInstance(
  stages: Awaited<ReturnType<typeof approvalApi.getInstance>>['stages'] = [],
) {
  return {
    id: 'inst-1',
    document_id: 'doc-1',
    route_id: 'r1',
    tenant_id: 't1',
    status: 'in_progress',
    submitted_by: 'maria',
    submitted_at: '2026-04-14T10:00:00.000Z',
    completed_at: null,
    stages,
    etag: '"v3"',
    frozen_content_hash: 'a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90',
  } as Awaited<ReturnType<typeof approvalApi.getInstance>>;
}

function renderAt(url = '/approvals/doc-1') {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route path="/approvals/:documentId" element={<ApprovalCockpitPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('ApprovalCockpitPage', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    mockState.sessionSpy.mockReset();
    mockState.autosaveSpy.mockReset();
    mockState.sessionSpy.mockReturnValue({
      state: { phase: 'idle' },
      setLastAck: vi.fn(),
      release: vi.fn(),
    });
    mockState.autosaveSpy.mockReturnValue({
      status: 'idle',
      queue: vi.fn(),
      flush: vi.fn(),
    });
    vi.spyOn(documentsApi, 'getDocument').mockResolvedValue(makeDoc());
    vi.spyOn(documentsApi, 'signedRevisionURL').mockReturnValue('/revisions/rev-1/signed-url');
    vi.spyOn(controlledDocumentsApi, 'fetchActiveDocumentInstance').mockResolvedValue(makeContext());
    vi.spyOn(approvalApi, 'getInstance').mockResolvedValue(makeInstance());
    vi.spyOn(documentsApi, 'listComments').mockResolvedValue([]);
    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (String(url).includes('signed-url')) {
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ url: 'https://s3/doc.docx' }) });
      }
      return Promise.resolve({ ok: true, arrayBuffer: () => Promise.resolve(new ArrayBuffer(8)) });
    }) as unknown as typeof fetch;
  });

  it('renders the document title and code from getDocument', async () => {
    renderAt();
    await waitFor(() => {
      expect(screen.getAllByText('POP Limpeza de Linha').length).toBeGreaterThan(0);
      expect(screen.getAllByText('POP-QUA-0148').length).toBeGreaterThan(0);
    });
  });

  it('tab list has exactly two tabs: documento and comentarios', async () => {
    renderAt();
    await waitFor(() => {
      expect(screen.getAllByText('POP Limpeza de Linha').length).toBeGreaterThan(0);
    });
    const tabs = screen.getAllByRole('tab');
    expect(tabs).toHaveLength(2);
    expect(tabs[0].textContent).toBe('Documento');
    expect(tabs[1].textContent).toBe('Comentários');
  });

  it('approval stage: editor mounts in readonly mode, no session/autosave mounted', async () => {
    vi.spyOn(approvalApi, 'getInstance').mockResolvedValue(
      makeInstance([
        {
          id: 'stage-1',
          stage_index: 0,
          label: 'Aprovação da Qualidade',
          status: 'active',
          signoffs: [],
          actors: [
            { user_id: 'user-approver-1', display_name: 'Ana Revisora', status: 'active', decision: null },
          ],
          stage_kind: 'approval',
          due_at: null,
        },
      ]),
    );
    renderAt();
    await waitFor(() => {
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('readonly');
    });
    expect(mockState.sessionSpy).not.toHaveBeenCalled();
    expect(mockState.autosaveSpy).not.toHaveBeenCalled();
  });

  it('review stage + eligible actor: editor mounts in review mode, no session/autosave mounted', async () => {
    vi.spyOn(approvalApi, 'getInstance').mockResolvedValue(
      makeInstance([
        {
          id: 'stage-1',
          stage_index: 0,
          label: 'Revisão técnica',
          status: 'active',
          signoffs: [],
          actors: [
            { user_id: 'user-approver-1', display_name: 'Ana Revisora', status: 'waiting', decision: null },
          ],
          stage_kind: 'review',
          due_at: null,
        },
      ]),
    );
    renderAt();
    await waitFor(() => {
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('review');
    });
    expect(mockState.sessionSpy).not.toHaveBeenCalled();
    expect(mockState.autosaveSpy).not.toHaveBeenCalled();
  });

  it('review stage, non-eligible actor: editor falls back to readonly', async () => {
    vi.spyOn(approvalApi, 'getInstance').mockResolvedValue(
      makeInstance([
        {
          id: 'stage-1',
          stage_index: 0,
          label: 'Revisão técnica',
          status: 'active',
          signoffs: [],
          actors: [
            { user_id: 'someone-else', display_name: 'Outro', status: 'active', decision: null },
          ],
          stage_kind: 'review',
          due_at: null,
        },
      ]),
    );
    renderAt();
    await waitFor(() => {
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('readonly');
    });
    expect(mockState.sessionSpy).not.toHaveBeenCalled();
    expect(mockState.autosaveSpy).not.toHaveBeenCalled();
  });

  it('review stage, oversee observer (actor status not active/waiting): readonly', async () => {
    vi.spyOn(approvalApi, 'getInstance').mockResolvedValue(
      makeInstance([
        {
          id: 'stage-1',
          stage_index: 0,
          label: 'Revisão técnica',
          status: 'active',
          signoffs: [],
          actors: [
            { user_id: 'user-approver-1', display_name: 'Ana Revisora', status: 'approved', decision: 'approve' },
          ],
          stage_kind: 'review',
          due_at: null,
        },
      ]),
    );
    renderAt();
    await waitFor(() => {
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('readonly');
    });
    expect(mockState.sessionSpy).not.toHaveBeenCalled();
    expect(mockState.autosaveSpy).not.toHaveBeenCalled();
  });

  it('no active stage: editor mounts in readonly mode', async () => {
    vi.spyOn(approvalApi, 'getInstance').mockResolvedValue(makeInstance([]));
    renderAt();
    await waitFor(() => {
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('readonly');
    });
    expect(mockState.sessionSpy).not.toHaveBeenCalled();
    expect(mockState.autosaveSpy).not.toHaveBeenCalled();
  });

  it('visibility 404 on document detail renders the inline not-found alert, not a toast', async () => {
    vi.spyOn(documentsApi, 'getDocument').mockRejectedValue({ code: 'NOT_FOUND', message: 'not found' });
    renderAt();
    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeTruthy();
    });
    expect(screen.queryByTestId('editor')).toBeNull();
  });

  it('shows the sidebar error state when the active-document-context query errors', async () => {
    vi.spyOn(controlledDocumentsApi, 'fetchActiveDocumentInstance').mockRejectedValue(
      new Error('context boom'),
    );
    renderAt();
    await waitFor(() => {
      expect(
        screen.getByText('Não foi possível carregar os dados de aprovação.'),
      ).toBeTruthy();
    });
  });

  it('shows the inactive-flow message when there is no active approval context', async () => {
    vi.spyOn(controlledDocumentsApi, 'fetchActiveDocumentInstance').mockResolvedValue(null);
    renderAt();
    await waitFor(() => {
      expect(
        screen.getByText('Este documento não está em um fluxo de aprovação ativo.'),
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
      expect(screen.getAllByText('POP Limpeza de Linha').length).toBeGreaterThan(0);
    });
    fireEvent.click(screen.getByRole('tab', { name: 'Comentários' }));
    await waitFor(() => {
      expect(screen.getByText('Revisar a seção 3.')).toBeTruthy();
    });
  });

  it('renders the integrity disclosure (collapsed) in the sidebar, revealing the hash + copy button on expand', async () => {
    renderAt();
    await waitFor(() => {
      expect(screen.getByText(/Conteúdo verificado/)).toBeTruthy();
    });
    // Collapsed: the frozen content hash is NOT in the DOM.
    const frozenHash = 'a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90';
    expect(screen.queryByText(frozenHash)).toBeNull();

    fireEvent.click(screen.getByText(/Conteúdo verificado/));
    await waitFor(() => {
      expect(screen.getByText(frozenHash)).toBeTruthy();
    });
    expect(screen.getAllByRole('button', { name: 'Copiar' }).length).toBeGreaterThan(0);
  });

  it('does not render the stale-data banner (F2/signoff invalidation replaces polling)', async () => {
    renderAt();
    await waitFor(() => {
      expect(screen.getAllByText('POP Limpeza de Linha').length).toBeGreaterThan(0);
    });
    expect(screen.queryByText(/dados podem estar desatualizados/i)).toBeNull();
    expect(screen.queryByRole('button', { name: 'Atualizar' })).toBeNull();
  });

  it('preselects the reject decision inline when ?decision=reject is present', async () => {
    renderAt('/approvals/doc-1?decision=reject');
    await waitFor(() => {
      expect(screen.getByRole('radio', { name: /Assinar e devolver/ })).toHaveAttribute(
        'aria-checked',
        'true',
      );
    });
    expect(screen.getByRole('radio', { name: /Assinar e aprovar/ })).toHaveAttribute(
      'aria-checked',
      'false',
    );
  });
});
