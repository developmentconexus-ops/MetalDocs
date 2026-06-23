import { act, render, screen, waitFor } from '@testing-library/react';
import { createRef } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

// ---------------------------------------------------------------------------
// Mock @metaldocs/editor-ui
// ---------------------------------------------------------------------------
const fakeEditorSaveNow = vi.fn<() => Promise<ArrayBuffer | null>>();
const fakeEditorGetPageCount = vi.fn<() => number | null>();

vi.mock('@metaldocs/editor-ui', () => {
  const { forwardRef } = require('react');
  const MetalDocsEditor = forwardRef(function MetalDocsEditor(
    props: Record<string, unknown>,
    ref: unknown,
  ) {
    // Expose imperative handle via the forwarded ref so tests can call saveNow
    const { useImperativeHandle } = require('react');
    useImperativeHandle(ref, () => ({
      saveNow: fakeEditorSaveNow,
      getPageCount: fakeEditorGetPageCount,
      getDocumentBuffer: vi.fn(),
      focus: vi.fn(),
    }));
    return (
      <div
        data-testid="metaldocs-editor"
        data-mode={props.mode as string}
        data-has-buffer={props.documentBuffer !== undefined ? 'true' : 'false'}
        data-has-comments={Array.isArray(props.comments) ? 'true' : 'false'}
      />
    );
  });
  return { MetalDocsEditor };
});

// ---------------------------------------------------------------------------
// Mock hooks
// ---------------------------------------------------------------------------
const mockSessionState = { phase: 'writer' as const, sessionID: 'session-123', lastAckRevisionID: 'rev-0' };
const mockSetLastAck = vi.fn();
const mockSessionAcquire = vi.fn();
const mockSessionRelease = vi.fn();

vi.mock('../../documents/hooks/editor/useDocumentSession', () => ({
  useDocumentSession: vi.fn(() => ({
    state: mockSessionState,
    acquire: mockSessionAcquire,
    release: mockSessionRelease,
    setLastAck: mockSetLastAck,
  })),
}));

const mockAutosaveQueue = vi.fn<(buf: ArrayBuffer, snapshot: unknown, pageCount?: number | null) => Promise<void>>();
const mockAutosaveFlush = vi.fn<() => Promise<boolean>>();

vi.mock('../../documents/hooks/editor/useDocumentAutosave', () => ({
  useDocumentAutosave: vi.fn(() => ({
    status: 'idle' as const,
    queue: mockAutosaveQueue,
    flush: mockAutosaveFlush,
  })),
}));

const fakeComments = [{ id: 1, text: 'comment A' }];
const mockSetComments = vi.fn();
const mockAddComment = vi.fn();
const mockResolveComment = vi.fn();
const mockRemoveComment = vi.fn();
const mockReplyComment = vi.fn();
const mockReopenComment = vi.fn();
const mockRetryComments = vi.fn();

vi.mock('../../documents/hooks/editor/useDocumentComments', () => ({
  useDocumentComments: vi.fn(() => ({
    comments: fakeComments,
    orphans: [],
    loading: false,
    loadError: null,
    add: mockAddComment,
    resolve: mockResolveComment,
    reopen: mockReopenComment,
    remove: mockRemoveComment,
    reply: mockReplyComment,
    retry: mockRetryComments,
    setComments: mockSetComments,
  })),
}));

// ---------------------------------------------------------------------------
// Mock signedRevisionURL + apiFetch
// ---------------------------------------------------------------------------
const FAKE_SIGNED_URL = 'https://s3.example.com/file.docx?sig=abc';
const FAKE_ARRAY_BUFFER = new ArrayBuffer(8);

vi.mock('../../documents/api/documents', () => ({
  signedRevisionURL: vi.fn((documentId: string, revisionId: string) =>
    `/api/v1/documents/${documentId}/revisions/${revisionId}/url`,
  ),
}));

vi.mock('../../../lib/api', () => ({
  apiFetch: vi.fn(async () => ({ url: FAKE_SIGNED_URL })),
}));

// Mock global fetch for the S3 download
const mockFetch = vi.fn(async () => ({
  ok: true,
  arrayBuffer: async () => FAKE_ARRAY_BUFFER,
}));

// ---------------------------------------------------------------------------
// Component under test — imported AFTER mocks
// ---------------------------------------------------------------------------
import { ReviewDocumentCanvas, type ReviewDocumentCanvasRef } from './ReviewDocumentCanvas';
import { useDocumentSession } from '../../documents/hooks/editor/useDocumentSession';
import { useDocumentAutosave } from '../../documents/hooks/editor/useDocumentAutosave';
import { useDocumentComments } from '../../documents/hooks/editor/useDocumentComments';
import { apiFetch } from '../../../lib/api';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------
function renderCanvas(overrides: Partial<{
  documentId: string;
  currentRevisionId: string;
  status: string;
  approverDisplay: string;
}> = {}) {
  const ref = createRef<ReviewDocumentCanvasRef>();
  const result = render(
    <ReviewDocumentCanvas
      documentId="doc-1"
      currentRevisionId="rev-abc"
      status="under_review"
      approverDisplay="Alice"
      ref={ref}
      {...overrides}
    />,
  );
  return { ref, ...result };
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe('ReviewDocumentCanvas', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Reset global fetch mock
    vi.stubGlobal('fetch', mockFetch);
    fakeEditorSaveNow.mockResolvedValue(FAKE_ARRAY_BUFFER);
    fakeEditorGetPageCount.mockReturnValue(2);
    mockAutosaveQueue.mockResolvedValue(undefined);
    mockAutosaveFlush.mockResolvedValue(true);
  });

  // 1. mounts MetalDocsEditor with mode="review"
  it('mounts MetalDocsEditor with mode="review"', async () => {
    renderCanvas();

    await waitFor(() => {
      const editor = screen.queryByTestId('metaldocs-editor');
      expect(editor).not.toBeNull();
    });

    const editor = screen.getByTestId('metaldocs-editor');
    expect(editor.dataset.mode).toBe('review');
  });

  // 2. passes the fetched working-revision ArrayBuffer as documentBuffer
  it('fetches the signed URL and passes the ArrayBuffer as documentBuffer', async () => {
    renderCanvas({ documentId: 'doc-1', currentRevisionId: 'rev-abc' });

    await waitFor(() => {
      expect(vi.mocked(apiFetch)).toHaveBeenCalledWith(
        '/api/v1/documents/doc-1/revisions/rev-abc/url',
      );
    });

    await waitFor(() => {
      const editor = screen.queryByTestId('metaldocs-editor');
      expect(editor?.dataset.hasBuffer).toBe('true');
    });
  });

  // 3. wires comments from useDocumentComments (controlled comments prop present)
  it('passes comments from useDocumentComments to MetalDocsEditor', async () => {
    renderCanvas();

    // Verify the hook was called with correct args
    expect(vi.mocked(useDocumentComments)).toHaveBeenCalledWith('doc-1', 'Alice');

    await waitFor(() => {
      const editor = screen.queryByTestId('metaldocs-editor');
      expect(editor?.dataset.hasComments).toBe('true');
    });
  });

  // 4. calling the forwarded ref's flushSave() triggers editor saveNow() then autosave flush()
  it('flushSave() calls editor saveNow(), queues the buffer, then calls autosave flush()', async () => {
    const { ref } = renderCanvas();

    // Wait for mount/buffer fetch to settle
    await waitFor(() => {
      expect(screen.queryByTestId('metaldocs-editor')).not.toBeNull();
    });

    // The component may not yet have mounted the editor if buffer is still loading
    // Wait for editor to be in a state where ref can be used
    await act(async () => {
      await ref.current?.flushSave();
    });

    expect(fakeEditorSaveNow).toHaveBeenCalledTimes(1);
    expect(mockAutosaveFlush).toHaveBeenCalledTimes(1);
  });

  // 5. useDocumentSession is called with enabled: true when status is under_review
  it('enables session when status is under_review', () => {
    renderCanvas({ status: 'under_review' });
    expect(vi.mocked(useDocumentSession)).toHaveBeenCalledWith(
      'doc-1',
      expect.objectContaining({ enabled: true }),
    );
  });

  // 6. useDocumentSession is called with enabled: false for other statuses
  it('disables session when status is not under_review', () => {
    renderCanvas({ status: 'draft' });
    expect(vi.mocked(useDocumentSession)).toHaveBeenCalledWith(
      'doc-1',
      expect.objectContaining({ enabled: false }),
    );
  });

  // 7. useDocumentAutosave is wired with correct session info
  it('wires useDocumentAutosave with sessionID and baseRevisionID from session state', () => {
    renderCanvas();
    expect(vi.mocked(useDocumentAutosave)).toHaveBeenCalledWith(
      expect.objectContaining({
        documentID: 'doc-1',
        sessionID: 'session-123',
        baseRevisionID: 'rev-0',
      }),
    );
  });
});
