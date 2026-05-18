import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { DocumentEditorPage } from './DocumentEditorPage';

const mockState = vi.hoisted(() => ({
  editorProps: [] as Record<string, unknown>[],
  addCommentSpy: vi.fn(),
  setCommentsSpy: vi.fn(),
  autosaveQueueSpy: vi.fn(),
  editorSaveNowSpy: vi.fn(),
  autosaveFlushSpy: vi.fn(),
  commentsLoadError: null as string | null,
  retryCommentsSpy: vi.fn(),
}));

// ── Mock heavy dependencies ────────────────────────────────────────────────

vi.mock('@metaldocs/editor-ui', () => ({
  MetalDocsEditor: React.forwardRef((props: Record<string, unknown>, ref) => {
    React.useImperativeHandle(ref, () => ({
      async saveNow() {
        return mockState.editorSaveNowSpy();
      },
      async getDocumentBuffer() {
        return null;
      },
      focus() {},
    }));
    return (
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      <div
        data-testid="editor"
        data-mode={props.mode as any}
        ref={() => {
          mockState.editorProps.push(props);
        }}
        onClick={() =>
          (props.onDocumentNameChange as ((name: string) => void) | undefined)?.('NewName')
        }
      />
    );
  }),
}));

vi.mock('../hooks/editor/useDocumentSession', () => ({
  useDocumentSession: () => ({
    state: { phase: 'writer', sessionID: 's1', lastAckRevisionID: 'r1' },
    setLastAck: vi.fn(),
    release: vi.fn().mockResolvedValue(undefined),
  }),
}));

vi.mock('../hooks/editor/useDocumentAutosave', () => ({
  useDocumentAutosave: () => ({
    status: 'idle',
    queue: mockState.autosaveQueueSpy,
    flush: mockState.autosaveFlushSpy,
  }),
}));

vi.mock('../hooks/editor/useDocumentComments', () => ({
  useDocumentComments: () => ({
    comments: [],
    setComments: mockState.setCommentsSpy,
    add: mockState.addCommentSpy,
    resolve: vi.fn().mockResolvedValue(undefined),
    reopen: vi.fn().mockResolvedValue(undefined),
    remove: vi.fn().mockResolvedValue(undefined),
    reply: vi.fn().mockResolvedValue(undefined),
    loadError: mockState.commentsLoadError,
    retry: mockState.retryCommentsSpy,
  }),
}));

vi.mock('./CheckpointsDialog', () => ({
  CheckpointsDialog: () => null,
}));

vi.mock('./ExportMenuButton', () => ({
  ExportMenuButton: () => null,
}));

// Default mock — polls return pending. Individual tests override via mockReturnValue.
vi.mock('../hooks/editor/useDocumentPdfStatus', () => ({
  useDocumentPdfStatus: vi.fn(() => ({
    status: 'pending' as string,
    url: undefined as string | undefined,
    retry: vi.fn(),
  })),
}));

vi.mock('./PDFCell', () => ({
  PDFCell: ({ status, url }: { status: string; url?: string }) => {
    if (status === 'ready' && url) return <a href={url} download>Baixar PDF</a>;
    return <span>{status}</span>;
  },
}));

vi.mock('../api/documents', () => ({
  getDocument: vi.fn(),
  finalizeDocument: vi.fn().mockResolvedValue(undefined),
  renameDocument: vi.fn().mockResolvedValue(undefined),
  signedRevisionURL: vi.fn().mockReturnValue('/revisions/r1/signed-url'),
}));

// ── Helpers ───────────────────────────────────────────────────────────────

import * as api from '../api/documents';
import * as pdfHook from '../hooks/editor/useDocumentPdfStatus';
import { toast } from 'sonner';

function renderPage(ui: React.ReactElement) {
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

function makeDoc(status: string, overrides: Record<string, unknown> = {}) {
  return {
    Status: status,
    status,
    CurrentRevisionID: 'r1',
    current_revision_id: 'r1',
    Name: 'Original.docx',
    name: 'Original.docx',
    Code: 'C-001',
    code: 'C-001',
    RevisionVersion: 1,
    revision_version: 1,
    ...overrides,
  };
}

function makeFetchForBuffer() {
  return vi.fn().mockImplementation((url: string) => {
    if (String(url).includes('signed-url')) {
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ url: 'https://s3/doc.docx' }),
      });
    }
    return Promise.resolve({
      ok: true,
      arrayBuffer: () => Promise.resolve(new ArrayBuffer(8)),
    });
  });
}

beforeEach(() => {
  global.fetch = makeFetchForBuffer();
  mockState.editorProps = [];
  mockState.addCommentSpy.mockReset();
  mockState.addCommentSpy.mockResolvedValue(undefined);
  mockState.setCommentsSpy.mockReset();
  mockState.autosaveQueueSpy.mockReset();
  mockState.autosaveQueueSpy.mockResolvedValue(undefined);
  mockState.editorSaveNowSpy.mockReset();
  mockState.editorSaveNowSpy.mockResolvedValue(null);
  mockState.autosaveFlushSpy.mockReset();
  mockState.autosaveFlushSpy.mockResolvedValue(true);
  mockState.commentsLoadError = null;
  mockState.retryCommentsSpy.mockReset();
});

afterEach(() => {
  vi.clearAllMocks();
});

// ── Tests ─────────────────────────────────────────────────────────────────

describe('DocumentEditorPage E1 gate', () => {
  it('renders document-edit when status=draft', async () => {
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('draft') as never);
    renderPage(<DocumentEditorPage documentID="d1" onDone={() => {}} />);
    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('document-edit'),
    );
  });

  it('renders readonly when status=under_review', async () => {
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('under_review') as never);
    renderPage(<DocumentEditorPage documentID="d1" onDone={() => {}} />);
    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('readonly'),
    );
  });

  it('allows review comment callbacks when status=under_review without enabling content editing', async () => {
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('under_review') as never);
    renderPage(<DocumentEditorPage documentID="d1" onDone={() => {}} />);

    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('readonly'),
    );

    const props = mockState.editorProps.at(-1) as {
      onCommentAdd?: (comment: { id: number; content: unknown[] }) => void;
    };
    props.onCommentAdd?.({ id: 77, content: [] });

    expect(mockState.addCommentSpy).toHaveBeenCalledWith({ id: 77, content: [] });
  });

  it('refetches doc on window focus and updates mode', async () => {
    let callCount = 0;
    vi.mocked(api.getDocument).mockImplementation(async () => {
      callCount += 1;
      return makeDoc(callCount === 1 ? 'draft' : 'under_review') as never;
    });

    renderPage(<DocumentEditorPage documentID="d1" onDone={() => {}} />);
    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('document-edit'),
    );

    fireEvent(window, new FocusEvent('focus'));

    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('readonly'),
    );
  });
});

describe('DocumentEditorPage E9 rename rollback', () => {
  it('rolls back document name on rename failure and shows error toast', async () => {
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('draft') as never);
    vi.mocked(api.renameDocument).mockRejectedValueOnce(new Error('Server error'));
    const toastSpy = vi.spyOn(toast, 'error');

    renderPage(<DocumentEditorPage documentID="d1" onDone={() => {}} />);

    // Wait for editor to mount with draft mode
    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('document-edit'),
    );

    // Title uses separate code chip + title span.
    expect(screen.getByText('Original')).toBeTruthy();

    // Click editor → onTitleChange('NewName') → handleRename('NewName')
    fireEvent.click(screen.getByTestId('editor'));

    // After reject + rollback, title should revert
    await waitFor(() =>
      expect(screen.getByText('Original')).toBeTruthy(),
    );

    expect(vi.mocked(api.renameDocument)).toHaveBeenCalledWith('d1', 'NewName');
    expect(toastSpy).toHaveBeenCalled();
  });
});

describe('DocumentEditorPage E11 PDF polling', () => {
  it('keeps readonly editor mode for published documents while pdf status hook reports ready', async () => {
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('published') as never);
    vi.mocked(pdfHook.useDocumentPdfStatus).mockReturnValue({
      status: 'ready',
      url: 'https://s3/p.pdf',
      retry: vi.fn(),
    });

    renderPage(<DocumentEditorPage documentID="d1" onDone={() => {}} />);
    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('readonly'),
    );
  });
});

describe('DocumentEditorPage autosave wiring', () => {
  it('queues the buffer provided by MetalDocsEditor without re-reading the editor ref', async () => {
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('draft', {
      FormDataJSON: { foo: 'bar' },
      form_data: { foo: 'bar' },
    }) as never);

    renderPage(<DocumentEditorPage documentID="d1" onDone={() => {}} />);

    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('document-edit'),
    );

    const props = mockState.editorProps.at(-1) as {
      onAutoSave?: (buf: ArrayBuffer) => Promise<void>;
    };
    const emittedBuffer = new Uint8Array([9, 8, 7]).buffer;
    await props.onAutoSave?.(emittedBuffer);

    expect(mockState.autosaveQueueSpy).toHaveBeenCalledWith(emittedBuffer, { foo: 'bar' });
  });

  it('forces an immediate editor save before finalize so latest bytes are queued', async () => {
    const onDone = vi.fn();
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('draft', {
      FormDataJSON: { foo: 'bar' },
      form_data: { foo: 'bar' },
    }) as never);
    const finalBuf = new Uint8Array([5, 4, 3]).buffer;
    mockState.editorSaveNowSpy.mockResolvedValue(finalBuf);
    vi.mocked(api.finalizeDocument).mockResolvedValue(undefined as never);

    renderPage(<DocumentEditorPage documentID="d1" onDone={onDone} />);

    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('document-edit'),
    );

    fireEvent.click(screen.getByRole('button', { name: /Submeter para revis/i }));

    await waitFor(() => expect(mockState.editorSaveNowSpy).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(mockState.autosaveQueueSpy).toHaveBeenCalledWith(finalBuf, { foo: 'bar' }));
    await waitFor(() => expect(mockState.autosaveFlushSpy).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(vi.mocked(api.finalizeDocument)).toHaveBeenCalledWith('d1'));
    await waitFor(() => expect(onDone).toHaveBeenCalledTimes(1));
  });

  it('does not finalize when flush fails', async () => {
    const onDone = vi.fn();
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('draft', {
      FormDataJSON: { foo: 'bar' },
      form_data: { foo: 'bar' },
    }) as never);
    const finalBuf = new Uint8Array([1, 2, 3]).buffer;
    mockState.editorSaveNowSpy.mockResolvedValue(finalBuf);
    mockState.autosaveFlushSpy.mockResolvedValue(false);

    renderPage(<DocumentEditorPage documentID="d1" onDone={onDone} />);

    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('document-edit'),
    );

    fireEvent.click(screen.getByRole('button', { name: /Submeter para revis/i }));

    await waitFor(() => expect(mockState.editorSaveNowSpy).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(mockState.autosaveQueueSpy).toHaveBeenCalledWith(finalBuf, { foo: 'bar' }));
    await waitFor(() => expect(mockState.autosaveFlushSpy).toHaveBeenCalledTimes(1));
    expect(vi.mocked(api.finalizeDocument)).not.toHaveBeenCalled();
    expect(onDone).not.toHaveBeenCalled();
  });
});

describe('DocumentEditorPage load failure state', () => {
  it('shows an inline not-found style error and does not render metadata chrome when the document fetch fails', async () => {
    vi.mocked(api.getDocument).mockRejectedValueOnce({
      code: 'not_found',
      message: 'not_found',
    });

    renderPage(<DocumentEditorPage documentID="missing-doc" onDone={() => {}} />);

    await waitFor(() =>
      expect(screen.getByRole('alert')).toHaveTextContent('Documento não encontrado.'),
    );

    expect(screen.queryByTestId('editor')).toBeNull();
    expect(screen.queryByText('Metadados')).toBeNull();
    expect(screen.queryByText('Próximos aprovadores')).toBeNull();
  });

  it('keeps metadata chrome visible and shows a friendly editor error when revision blob fetch fails', async () => {
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('draft') as never);
    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (String(url).includes('signed-url')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ url: 'https://s3/doc.docx' }),
        });
      }
      return Promise.resolve({
        ok: false,
        status: 403,
      });
    }) as typeof fetch;

    renderPage(<DocumentEditorPage documentID="d1" onDone={() => {}} />);

    await waitFor(() =>
      expect(screen.getByRole('alert')).toHaveTextContent('Falha ao carregar o arquivo do documento. Tente novamente.'),
    );

    expect(screen.getAllByText('C-001').length).toBeGreaterThan(0);
    expect(screen.getByText('Original')).toBeTruthy();
    expect(screen.getByText('Metadados')).toBeTruthy();
    expect(screen.queryByTestId('editor')).toBeNull();
    expect(screen.queryByText('http_403')).toBeNull();
    expect(screen.queryByText('missing_signed_url')).toBeNull();
  });

  it('shows a persistent comments load banner and offers retry when comments fail to load', async () => {
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('under_review') as never);
    mockState.commentsLoadError = 'Falha ao carregar comentários.';

    renderPage(<DocumentEditorPage documentID="d1" onDone={() => {}} />);

    await waitFor(() =>
      expect(screen.getByRole('alert')).toHaveTextContent('Falha ao carregar comentários.'),
    );

    fireEvent.click(screen.getByRole('button', { name: 'Tentar novamente' }));
    expect(mockState.retryCommentsSpy).toHaveBeenCalledTimes(1);
  });
});
