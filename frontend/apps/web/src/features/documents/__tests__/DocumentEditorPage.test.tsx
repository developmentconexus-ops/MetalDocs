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
      ID: 'doc-1',
      TenantID: 'tenant-1',
      TemplateVersionID: 'tv-1',
      Name: 'Quarterly Report',
      Status: 'draft',
      FormDataJSON: { foo: 'bar' },
      CurrentRevisionID: 'rev-1',
      RevisionVersion: 1,
      ActiveSessionID: '',
      CreatedAt: '2026-05-18T10:00:00Z',
      UpdatedAt: '2026-05-18T10:00:00Z',
      CreatedBy: 'user-1',
      Code: 'DOC-1',
      currentRevisionFileSizeBytes: 1304,
      currentRevisionPageCount: 3,
      currentRevisionPageCountSource: 'eigenpal_client',
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
