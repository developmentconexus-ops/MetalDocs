import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { PropsWithChildren } from 'react';
import { forwardRef, useImperativeHandle } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { DocumentEditorPage } from '../pages/DocumentEditorPage';
import type { DocumentDetail } from '../api/documents';

const queueSpy = vi.fn();
const flushSpy = vi.fn();
const emittedBuffer = new Uint8Array([9]).buffer;

vi.mock('../api/documents', () => ({
  getDocument: vi.fn(),
  signedRevisionURL: vi.fn(),
  finalizeDocument: vi.fn(),
}));

vi.mock('../hooks/editor/useDocumentAutosave', () => ({
  useDocumentAutosave: () => ({
    status: 'idle',
    queue: queueSpy,
    flush: flushSpy,
  }),
}));

vi.mock('../hooks/editor/useDocumentSession', () => ({
  useDocumentSession: () => ({
    state: { phase: 'writer', sessionID: 'sess-1', lastAckRevisionID: 'rev-1' },
    setLastAck: vi.fn(),
    release: vi.fn().mockResolvedValue(undefined),
  }),
}));

vi.mock('../hooks/editor/useDocumentComments', () => ({
  useDocumentComments: () => ({
    comments: [],
    setComments: vi.fn(),
    add: vi.fn().mockResolvedValue(undefined),
    resolve: vi.fn().mockResolvedValue(undefined),
    reopen: vi.fn().mockResolvedValue(undefined),
    remove: vi.fn().mockResolvedValue(undefined),
    reply: vi.fn().mockResolvedValue(undefined),
  }),
}));

vi.mock('../ExportMenu', () => ({
  ExportMenu: ({ children }: PropsWithChildren) => <div data-testid="export-menu">{children}</div>,
}));

vi.mock('../CheckpointsPanel', () => ({
  CheckpointsPanel: () => <div data-testid="checkpoints-panel" />,
}));

vi.mock('@metaldocs/editor-ui', () => ({
  MetalDocsEditor: forwardRef<any, any>(function MockMetalDocsEditor(props, ref) {
    useImperativeHandle(ref, () => ({
      async saveNow() {
        return emittedBuffer;
      },
      async getDocumentBuffer() {
        throw new Error('DocumentEditorPage should not re-read the editor buffer during autosave');
      },
      getPageCount() {
        return 3;
      },
      focus() {},
    }), []);
    return (
      <div data-testid="metaldocs-editor">
        <button type="button" onClick={() => void props.onAutoSave?.(emittedBuffer)}>
          trigger-autosave
        </button>
      </div>
    );
  }),
}));

describe('DocumentEditorPage', () => {
  beforeEach(async () => {
    vi.clearAllMocks();
    const api = await import('../api/documents');
    const doc: DocumentDetail = {
      id: 'doc-1',
      tenant_id: 'tenant-1',
      template_version_id: 'tv-1',
      name: 'Quarterly Report',
      status: 'draft',
      form_data_json: { foo: 'bar' },
      current_revision_id: 'rev-1',
      revision_version: 1,
      active_session_id: '',
      created_at: '2026-05-18T10:00:00Z',
      updated_at: '2026-05-18T10:00:00Z',
      created_by: 'user-1',
      revision_number: 0,
      code: 'DOC-1',
      current_revision_file_size_bytes: 1304,
      current_revision_page_count: 3,
      current_revision_page_count_source: 'eigenpal_client',
      values_frozen_at: null,
      archived_at: null,
      controlled_document_id: null,
      revision_title: null,
      profile_code_snapshot: null,
      process_area_code_snapshot: null,
      effective_from: null,
      effective_to: null,
      review_due_at: null,
      review_surfaced_at: null,
      last_reviewed_at: null,
    };
    vi.mocked(api.getDocument).mockResolvedValue(doc);
    vi.mocked(api.signedRevisionURL).mockReturnValue('/api/v1/documents/doc-1/revisions/rev-1/url');

    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url === 'https://cdn.example.com/doc.docx') {
        return new Response(new Uint8Array([7, 8, 9]).buffer, { status: 200 });
      }
      if (url.includes('/api/v1/documents/doc-1/revisions/rev-1/url')) {
        return Response.json({ url: 'https://cdn.example.com/doc.docx' });
      }
      if (url.includes('/api/v1/taxonomy/profiles') || url.includes('/api/v1/taxonomy/areas')) {
        return Response.json({ items: [] });
      }
      return Response.json({});
    });

    vi.stubGlobal('fetch', fetchMock);
  });

  it('renders editor root and mounts editor after session acquisition', async () => {
    renderWithQueryClient(<DocumentEditorPage documentID="doc-1" onDone={vi.fn()} />);

    expect(document.querySelector('[data-editor-root]')).toBeTruthy();
    expect(screen.queryByTestId('metaldocs-editor')).toBeNull();

    await waitFor(() => expect(screen.getByTestId('metaldocs-editor')).toBeTruthy());
  });

  it('queues autosave from editor callback', async () => {
    renderWithQueryClient(<DocumentEditorPage documentID="doc-1" onDone={vi.fn()} />);

    await waitFor(() => expect(screen.getByTestId('metaldocs-editor')).toBeTruthy());
    fireEvent.click(screen.getByRole('button', { name: 'trigger-autosave' }));

    await waitFor(() =>
      expect(queueSpy).toHaveBeenCalledWith(emittedBuffer, { foo: 'bar' }, 3),
    );
  });
});

function renderWithQueryClient(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      {ui}
    </QueryClientProvider>,
  );
}
