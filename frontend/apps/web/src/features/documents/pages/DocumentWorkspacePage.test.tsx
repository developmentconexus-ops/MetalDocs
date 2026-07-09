import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { DocumentWorkspacePage } from './DocumentWorkspacePage';
import * as documentsApi from '../api/documents';
import type { DocumentDetail } from '../api/documents';
import type { ApprovalInstance, StageInstance } from '../../approval/api/approvalTypes';

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

vi.mock('../../../store/auth.store', () => ({
  useAuthStore: (selector: (s: { user: { displayName: string; userId: string } }) => unknown) =>
    selector({ user: { displayName: 'Ana Revisora', userId: 'user-reviewer-1' } }),
}));

function makeDoc(overrides: Partial<DocumentDetail> = {}): DocumentDetail {
  return {
    id: 'doc-1',
    code: 'POP-QUA-0148',
    name: 'POP Limpeza de Linha',
    status: 'under_review',
    revision_version: 3,
    revision_number: 3,
    controlled_document_id: 'cd-1',
    current_revision_id: 'rev-1',
    created_by: 'user-author-1',
    created_at: '2026-05-01T10:00:00.000Z',
    ...overrides,
  } as DocumentDetail;
}

function makeStage(overrides: Partial<StageInstance> = {}): StageInstance {
  return {
    id: 'stage-1',
    stage_index: 0,
    label: 'Revisão técnica',
    status: 'active',
    stage_kind: 'review',
    signoffs: [],
    actors: [],
    due_at: null,
    ...overrides,
  } as StageInstance;
}

function makeInstance(overrides: Partial<ApprovalInstance> = {}): ApprovalInstance {
  return {
    id: 'inst-1',
    document_id: 'doc-1',
    route_id: 'r1',
    tenant_id: 't1',
    status: 'in_progress',
    submitted_by: 'maria',
    submitted_at: '2026-05-01T10:00:00.000Z',
    completed_at: null,
    stages: [makeStage()],
    etag: '"v1"',
    frozen_content_hash: null,
    viewer: {
      is_author: false,
      eligible_for_active_stage: true,
      has_signed_active_stage: false,
      via_delegation_from: null,
    },
    verdicts: [],
    ...overrides,
  } as ApprovalInstance;
}

function renderAt(url = '/documents/doc-1/workspace') {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route path="/documents/:documentId/workspace" element={<DocumentWorkspacePage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

function mockFileFetch() {
  global.fetch = vi.fn().mockImplementation((url: string) => {
    if (String(url).includes('signed-url')) {
      return Promise.resolve({ ok: true, json: () => Promise.resolve({ url: 'https://s3/doc.docx' }) });
    }
    return Promise.resolve({ ok: true, arrayBuffer: () => Promise.resolve(new ArrayBuffer(8)) });
  }) as unknown as typeof fetch;
}

describe('DocumentWorkspacePage', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(documentsApi, 'signedRevisionURL').mockReturnValue('/revisions/rev-1/signed-url');
    mockFileFetch();
  });

  it('reviewing (eligible): shows the read canvas + verdict CTAs + timeline, and no signature panel', async () => {
    vi.spyOn(documentsApi, 'getDocument').mockResolvedValue(makeDoc());
    vi.spyOn(documentsApi, 'getApprovalInstance').mockResolvedValue(makeInstance());
    const { container } = renderAt();

    await waitFor(() => expect(screen.getByTestId('editor')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'Pronto para aprovação' })).toBeInTheDocument();
    expect(screen.getByLabelText('Timeline de aprovação')).toBeInTheDocument();
    expect(screen.getByText('Revisando')).toBeInTheDocument();
    expect(container.querySelector('input[type="password"]')).toBeNull();
  });

  it('observing: no verdict CTAs, but the read canvas and timeline still render', async () => {
    vi.spyOn(documentsApi, 'getDocument').mockResolvedValue(makeDoc());
    vi.spyOn(documentsApi, 'getApprovalInstance').mockResolvedValue(
      makeInstance({
        viewer: {
          is_author: false,
          eligible_for_active_stage: false,
          has_signed_active_stage: false,
          via_delegation_from: null,
        },
      }),
    );
    renderAt();

    await waitFor(() => expect(screen.getByTestId('editor')).toBeInTheDocument());
    expect(screen.getByText('Visualizando')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Pronto para aprovação' })).not.toBeInTheDocument();
    expect(screen.getByLabelText('Timeline de aprovação')).toBeInTheDocument();
  });

  it('author-waiting: no verdict CTAs, read canvas + timeline render', async () => {
    vi.spyOn(documentsApi, 'getDocument').mockResolvedValue(
      makeDoc({ status: 'under_review', created_by: 'user-reviewer-1' }),
    );
    vi.spyOn(documentsApi, 'getApprovalInstance').mockResolvedValue(
      makeInstance({
        viewer: {
          is_author: true,
          eligible_for_active_stage: false,
          has_signed_active_stage: false,
          via_delegation_from: null,
        },
      }),
    );
    renderAt();

    await waitFor(() => expect(screen.getByTestId('editor')).toBeInTheDocument());
    expect(screen.getByText('Aguardando revisão')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Pronto para aprovação' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('approval-sidebar-footer')).not.toBeInTheDocument();
    expect(screen.getByLabelText('Timeline de aprovação')).toBeInTheDocument();
  });

  it('lifecycle: read canvas + timeline render, no verdict CTAs (publish is S2b)', async () => {
    vi.spyOn(documentsApi, 'getDocument').mockResolvedValue(makeDoc({ status: 'approved' }));
    vi.spyOn(documentsApi, 'getApprovalInstance').mockResolvedValue(
      makeInstance({
        stages: [],
        status: 'approved',
        viewer: {
          is_author: false,
          eligible_for_active_stage: false,
          has_signed_active_stage: false,
          via_delegation_from: null,
        },
      }),
    );
    renderAt();

    await waitFor(() => expect(screen.getByTestId('editor')).toBeInTheDocument());
    expect(screen.getByText('Visualizando')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Pronto para aprovação' })).not.toBeInTheDocument();
    expect(screen.getByLabelText('Timeline de aprovação')).toBeInTheDocument();
  });

  it('§6 loading: shows the shell skeleton, no central spinner', async () => {
    vi.spyOn(documentsApi, 'getDocument').mockImplementation(() => new Promise(() => {}));
    vi.spyOn(documentsApi, 'getApprovalInstance').mockImplementation(() => new Promise(() => {}));
    renderAt();

    expect(screen.getByTestId('workspace-skeleton')).toBeInTheDocument();
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });

  it('§6 instance error: canvas stays readable, sidebar shows the error + retry', async () => {
    // status !== 404 so the hook's own retry() would normally back off twice;
    // 404 makes the failure deterministic and fast in tests while still
    // producing isError=true, the same signal the component reads for any
    // instance-fetch failure (mirrors useDocumentApprovalArtifact's gate).
    vi.spyOn(documentsApi, 'getDocument').mockResolvedValue(makeDoc());
    vi.spyOn(documentsApi, 'getApprovalInstance').mockRejectedValue(
      Object.assign(new Error('not_found'), { status: 404 }),
    );
    renderAt();

    await waitFor(() => expect(screen.getByTestId('editor')).toBeInTheDocument());
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
    expect(screen.getByRole('alert')).toHaveTextContent('Não foi possível carregar os dados de aprovação.');
    expect(screen.getByRole('button', { name: 'Tentar novamente' })).toBeInTheDocument();
  });

  it('doc-not-found: teaching-copy empty state', async () => {
    vi.spyOn(documentsApi, 'getDocument').mockRejectedValue(new Error('not_found'));
    renderAt();

    await waitFor(() =>
      expect(screen.getByText('Não foi possível localizar este documento.')).toBeInTheDocument(),
    );
  });
});
