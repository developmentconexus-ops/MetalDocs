import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { act, render, screen, fireEvent, waitFor } from '@testing-library/react';
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
  autosaveStatus: 'idle',
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
      getPageCount() {
        return 3;
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
    status: mockState.autosaveStatus,
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

vi.mock('../../taxonomy/queries/useProfilesQuery', () => ({
  useProfilesQuery: () => ({
    data: [{ code: 'POP', name: 'Procedimento Operacional' }],
  }),
}));

vi.mock('../queries/useAreasQuery', () => ({
  useAreasQuery: () => ({
    data: [{ code: 'RH', name: 'Recursos Humanos' }],
  }),
}));

vi.mock('../../registry/queries/useControlledDocumentDetailQuery', () => ({
  useControlledDocumentDetailQuery: () => ({
    data: { visibility: { scope: 'restricted', areaCodes: ['RH'], userIds: [] } },
  }),
}));

vi.mock('../queries/useDocumentRevisionHistoryQuery', () => ({
  useDocumentRevisionHistoryQuery: () => ({
    data: {
      items: [
        {
          documentId: 'doc-2',
          revisionNumber: 2,
          revisionTitle: 'Ajuste operacional',
          status: 'draft',
          createdAt: '2026-05-18T12:00:00Z',
          isCurrent: true,
        },
      ],
    },
  }),
}));

vi.mock('./CheckpointsDialog', () => ({
  CheckpointsDialog: () => null,
}));

vi.mock('./ExportMenuButton', () => ({
  ExportMenuButton: () => null,
}));


vi.mock('../api/documents', () => ({
  getDocument: vi.fn(),
  finalizeDocument: vi.fn().mockResolvedValue(undefined),
  getApprovalInstance: vi.fn().mockResolvedValue({ stages: [] }),
  renameDocument: vi.fn().mockResolvedValue(undefined),
  signedRevisionURL: vi.fn().mockReturnValue('/revisions/r1/signed-url'),
}));

// ── Helpers ───────────────────────────────────────────────────────────────

import * as api from '../api/documents';
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
    ControlledDocumentID: 'cd-1',
    ProfileCodeSnapshot: 'POP',
    ProcessAreaCodeSnapshot: 'RH',
    RevisionVersion: 0,
    revision_version: 0,
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
  mockState.autosaveStatus = 'idle';
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
    await act(async () => {
      await props.onAutoSave?.(emittedBuffer);
    });

    expect(mockState.autosaveQueueSpy).toHaveBeenCalledWith(emittedBuffer, { foo: 'bar' }, 3);
  });

  it('saves dirty editor content before submitting the governed revision', async () => {
    const onDone = vi.fn();
    const latestBuffer = new Uint8Array([1, 2, 3]).buffer;
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('draft', {
      FormDataJSON: { foo: 'bar' },
      form_data: { foo: 'bar' },
      RevisionVersion: 1,
      revision_version: 1,
    }) as never);
    mockState.autosaveStatus = 'dirty';
    mockState.editorSaveNowSpy.mockResolvedValue(latestBuffer);
    vi.mocked(api.finalizeDocument).mockResolvedValue(undefined as never);

    renderPage(<DocumentEditorPage documentID="d1" onDone={onDone} />);

    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('document-edit'),
    );

    const dirtyPropsWithChanges = mockState.editorProps.at(-1) as { onChange?: () => void };
    act(() => {
      dirtyPropsWithChanges.onChange?.();
      dirtyPropsWithChanges.onChange?.();
    });

    fireEvent.click(screen.getByRole('button', { name: /Submeter para revis/i }));
    fireEvent.change(screen.getByLabelText('TÃ­tulo'), { target: { value: 'Atualizacao de procedimento' } });
    fireEvent.click(screen.getByRole('button', { name: /Confirmar submiss/i }));

    await waitFor(() => expect(mockState.editorSaveNowSpy).toHaveBeenCalledTimes(1));
    expect(mockState.autosaveQueueSpy).toHaveBeenCalledWith(latestBuffer, { foo: 'bar' }, 3);
    expect(mockState.autosaveFlushSpy).toHaveBeenCalledTimes(1);
    await waitFor(() => expect(vi.mocked(api.finalizeDocument)).toHaveBeenCalledWith('d1', { revisionTitle: 'Atualizacao de procedimento' }));
    await waitFor(() => expect(onDone).toHaveBeenCalledTimes(1));
  });

  it('does not block finalize on saveNow when the editor has no unsaved changes', async () => {
    const onDone = vi.fn();
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('draft') as never);
    mockState.editorSaveNowSpy.mockImplementation(() => new Promise(() => {}));
    vi.mocked(api.finalizeDocument).mockResolvedValue(undefined as never);

    renderPage(<DocumentEditorPage documentID="d1" onDone={onDone} />);

    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('document-edit'),
    );

    const initialChangeProps = mockState.editorProps.at(-1) as { onChange?: () => void };
    act(() => {
      initialChangeProps.onChange?.();
    });

    fireEvent.click(screen.getByRole('button', { name: /Submeter para revis/i }));

    await waitFor(() => expect(vi.mocked(api.finalizeDocument)).toHaveBeenCalledWith('d1', {}));
    expect(mockState.editorSaveNowSpy).not.toHaveBeenCalled();
    expect(mockState.autosaveFlushSpy).not.toHaveBeenCalled();
    await waitFor(() => expect(onDone).toHaveBeenCalledTimes(1));
  });

  it('does not block finalize when the load-time editor change leaves autosave idle', async () => {
    const onDone = vi.fn();
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('draft') as never);
    mockState.editorSaveNowSpy.mockImplementation(() => new Promise(() => {}));
    vi.mocked(api.finalizeDocument).mockResolvedValue(undefined as never);

    renderPage(<DocumentEditorPage documentID="d1" onDone={onDone} />);

    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('document-edit'),
    );

    const loadProps = mockState.editorProps.at(-1) as { onChange?: () => void };
    act(() => {
      loadProps.onChange?.();
    });

    fireEvent.click(screen.getByRole('button', { name: /Submeter para revis/i }));

    await waitFor(() => expect(vi.mocked(api.finalizeDocument)).toHaveBeenCalledWith('d1', {}));
    expect(mockState.editorSaveNowSpy).not.toHaveBeenCalled();
    expect(mockState.autosaveFlushSpy).not.toHaveBeenCalled();
    await waitFor(() => expect(onDone).toHaveBeenCalledTimes(1));
  });

  it('blocks governed revision submission when dirty content cannot be flushed', async () => {
    const onDone = vi.fn();
    const latestBuffer = new Uint8Array([4, 5, 6]).buffer;
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('draft', {
      FormDataJSON: { foo: 'bar' },
      form_data: { foo: 'bar' },
      RevisionVersion: 1,
      revision_version: 1,
    }) as never);
    mockState.autosaveStatus = 'dirty';
    mockState.editorSaveNowSpy.mockResolvedValue(latestBuffer);
    mockState.autosaveFlushSpy.mockResolvedValue(false);
    const toastSpy = vi.spyOn(toast, 'error');

    renderPage(<DocumentEditorPage documentID="d1" onDone={onDone} />);

    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('document-edit'),
    );

    const dirtyProps = mockState.editorProps.at(-1) as { onChange?: () => void };
    act(() => {
      dirtyProps.onChange?.();
      dirtyProps.onChange?.();
    });

    fireEvent.click(screen.getByRole('button', { name: /Submeter para revis/i }));
    fireEvent.change(screen.getByLabelText('TÃ­tulo'), { target: { value: 'Atualizacao de procedimento' } });
    fireEvent.click(screen.getByRole('button', { name: /Confirmar submiss/i }));

    await waitFor(() => expect(mockState.editorSaveNowSpy).toHaveBeenCalledTimes(1));
    expect(mockState.autosaveQueueSpy).toHaveBeenCalledWith(latestBuffer, { foo: 'bar' }, 3);
    expect(mockState.autosaveFlushSpy).toHaveBeenCalledTimes(1);
    expect(toastSpy).toHaveBeenCalledWith('Erro ao salvar documento antes da submissão.');
    expect(vi.mocked(api.finalizeDocument)).not.toHaveBeenCalled();
    expect(onDone).not.toHaveBeenCalled();
  });

  it('blocks submit confirmation when revision title is blank', async () => {
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('draft', {
      RevisionVersion: 1,
      revision_version: 1,
    }) as never);

    renderPage(<DocumentEditorPage documentID="d1" onDone={() => {}} />);

    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('document-edit'),
    );

    const dirtyProps = mockState.editorProps.at(-1) as { onChange?: () => void };
    act(() => {
      dirtyProps.onChange?.();
    });

    fireEvent.click(screen.getByRole('button', { name: /Submeter para revis/i }));
    fireEvent.click(screen.getByRole('button', { name: /Confirmar submiss/i }));

    expect(screen.getByRole('alert')).toHaveTextContent('Informe o tÃ­tulo da revisÃ£o para submeter.');
    expect(vi.mocked(api.finalizeDocument)).not.toHaveBeenCalled();
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
    expect(screen.queryByText('Identificacao')).toBeNull();
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
    expect(screen.getByText('Identificacao')).toBeTruthy();
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
