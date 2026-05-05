import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { DocumentEditorPage } from './DocumentEditorPage';

// ── Mock heavy dependencies ────────────────────────────────────────────────

vi.mock('@metaldocs/editor-ui', () => ({
  MetalDocsEditor: (props: Record<string, unknown>) => (
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    <div
      data-testid="editor"
      data-mode={props.mode as any}
      onClick={() =>
        (props.onTitleChange as ((name: string) => void) | undefined)?.('NewName')
      }
    />
  ),
}));

vi.mock('./hooks/useDocumentSession', () => ({
  useDocumentSession: () => ({
    state: { phase: 'writer', sessionID: 's1', lastAckRevisionID: 'r1' },
    setLastAck: vi.fn(),
    release: vi.fn().mockResolvedValue(undefined),
  }),
}));

vi.mock('./hooks/useDocumentAutosave', () => ({
  useDocumentAutosave: () => ({
    status: 'idle',
    queue: vi.fn().mockResolvedValue(undefined),
    flush: vi.fn().mockResolvedValue(undefined),
  }),
}));

vi.mock('./hooks/useDocumentComments', () => ({
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

vi.mock('./CheckpointsDialog', () => ({
  CheckpointsDialog: () => null,
}));

vi.mock('./ExportMenuButton', () => ({
  ExportMenuButton: () => null,
}));

// Default mock — polls return pending. Individual tests override via mockReturnValue.
vi.mock('../hooks/v2/useDocumentPdfStatus', () => ({
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

vi.mock('../api/documentsV2', () => ({
  getDocument: vi.fn(),
  finalizeDocument: vi.fn().mockResolvedValue(undefined),
  renameDocument: vi.fn().mockResolvedValue(undefined),
  signedRevisionURL: vi.fn().mockReturnValue('/revisions/r1/signed-url'),
}));

// ── Helpers ───────────────────────────────────────────────────────────────

import * as api from '../api/documentsV2';
import * as pdfHook from '../hooks/v2/useDocumentPdfStatus';
import { toast } from 'sonner';

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
});

afterEach(() => {
  vi.clearAllMocks();
});

// ── Tests ─────────────────────────────────────────────────────────────────

describe('DocumentEditorPage E1 gate', () => {
  it('renders document-edit when status=draft', async () => {
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('draft') as never);
    render(<DocumentEditorPage documentID="d1" onDone={() => {}} />);
    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('document-edit'),
    );
  });

  it('renders readonly when status=under_review', async () => {
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('under_review') as never);
    render(<DocumentEditorPage documentID="d1" onDone={() => {}} />);
    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('readonly'),
    );
  });

  it('refetches doc on window focus and updates mode', async () => {
    let callCount = 0;
    vi.mocked(api.getDocument).mockImplementation(async () => {
      callCount += 1;
      return makeDoc(callCount === 1 ? 'draft' : 'under_review') as never;
    });

    render(<DocumentEditorPage documentID="d1" onDone={() => {}} />);
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

    render(<DocumentEditorPage documentID="d1" onDone={() => {}} />);

    // Wait for editor to mount with draft mode
    await waitFor(() =>
      expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('document-edit'),
    );

    // Title: C-001-Original (docCode-displayName, strips .docx)
    expect(screen.getByText(/C-001-Original/)).toBeTruthy();

    // Click editor → onTitleChange('NewName') → handleRename('NewName')
    fireEvent.click(screen.getByTestId('editor'));

    // After reject + rollback, title should revert
    await waitFor(() =>
      expect(screen.getByText(/C-001-Original/)).toBeTruthy(),
    );

    expect(vi.mocked(api.renameDocument)).toHaveBeenCalledWith('d1', 'NewName');
    expect(toastSpy).toHaveBeenCalled();
  });
});

describe('DocumentEditorPage E11 PDF cell', () => {
  it('shows PDF download link when doc is published and pdf is ready', async () => {
    vi.mocked(api.getDocument).mockResolvedValue(makeDoc('published') as never);
    vi.mocked(pdfHook.useDocumentPdfStatus).mockReturnValue({
      status: 'ready',
      url: 'https://s3/p.pdf',
      retry: vi.fn(),
    });

    render(<DocumentEditorPage documentID="d1" onDone={() => {}} />);
    const link = await screen.findByRole('link', { name: /Baixar PDF/i });
    expect(link.getAttribute('href')).toBe('https://s3/p.pdf');
  });
});
