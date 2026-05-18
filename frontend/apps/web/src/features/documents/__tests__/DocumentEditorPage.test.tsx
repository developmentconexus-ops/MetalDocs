import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { PropsWithChildren } from 'react';
import { forwardRef, useImperativeHandle } from 'react';
import { DocumentEditorPage } from '../pages/DocumentEditorPage';

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
    vi.mocked(api.getDocument).mockResolvedValue({
      ID: 'doc-1',
      Name: 'Quarterly Report',
      CurrentRevisionID: 'rev-1',
      CreatedBy: 'user-1',
      FormDataJSON: { foo: 'bar' },
      Status: 'draft',
    });
    vi.mocked(api.signedRevisionURL).mockReturnValue('/signed/url');

    const fetchMock = vi.fn()
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ url: 'https://cdn.example.com/doc.docx' }),
      })
      .mockResolvedValueOnce({
        ok: true,
        arrayBuffer: async () => new Uint8Array([7, 8, 9]).buffer,
      });

    vi.stubGlobal('fetch', fetchMock);
  });

  it('renders editor root and mounts editor after session acquisition', async () => {
    render(<DocumentEditorPage documentID="doc-1" onDone={vi.fn()} />);

    expect(document.querySelector('[data-editor-root]')).toBeTruthy();
    expect(screen.queryByTestId('metaldocs-editor')).toBeNull();

    await waitFor(() => expect(screen.getByTestId('metaldocs-editor')).toBeTruthy());
  });

  it('queues autosave from editor callback', async () => {
    render(<DocumentEditorPage documentID="doc-1" onDone={vi.fn()} />);

    await waitFor(() => expect(screen.getByTestId('metaldocs-editor')).toBeTruthy());
    fireEvent.click(screen.getByRole('button', { name: 'trigger-autosave' }));

    await waitFor(() =>
      expect(queueSpy).toHaveBeenCalledWith(emittedBuffer, { foo: 'bar' }),
    );
  });
});
