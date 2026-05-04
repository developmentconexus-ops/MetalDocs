import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { DocumentEditorPage } from './DocumentEditorPage';

// ── Mock heavy dependencies ────────────────────────────────────────────────

vi.mock('@metaldocs/editor-ui', () => ({
  MetalDocsEditor: (props: Record<string, unknown>) => (
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    <div data-testid="editor" data-mode={props.mode as any} />
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
vi.mock('./hooks/useDocumentPdfStatus', () => ({
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

vi.mock('./api/documentsV2', () => ({
  getDocument: vi.fn(),
  finalizeDocument: vi.fn().mockResolvedValue(undefined),
  renameDocument: vi.fn().mockResolvedValue(undefined),
  signedRevisionURL: vi.fn().mockReturnValue('/revisions/r1/signed-url'),
}));

// ── Helpers ───────────────────────────────────────────────────────────────

import * as api from './api/documentsV2';
import * as pdfHook from './hooks/useDocumentPdfStatus';

function makeDoc(status: string) {
  return {
    Status: status,
    status,
    CurrentRevisionID: 'r1',
    current_revision_id: 'r1',
    Name: 'Doc X',
    name: 'Doc X',
    Code: 'C-001',
    code: 'C-001',
    RevisionVersion: 1,
    revision_version: 1,
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
