import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi, beforeEach } from 'vitest';

import { EditorCanvas } from './EditorCanvas';
import * as documentsApi from '../../api/documents';

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

function mockFileFetch() {
  global.fetch = vi.fn().mockImplementation((url: string) => {
    if (String(url).includes('signed-url')) {
      return Promise.resolve({ ok: true, json: () => Promise.resolve({ url: 'https://s3/doc.docx' }) });
    }
    return Promise.resolve({ ok: true, arrayBuffer: () => Promise.resolve(new ArrayBuffer(8)) });
  }) as unknown as typeof fetch;
}

describe('EditorCanvas', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(documentsApi, 'signedRevisionURL').mockReturnValue('/revisions/rev-1/signed-url');
    mockFileFetch();
  });

  it('renders nothing when canMountEditor is false', () => {
    const { container } = render(
      <EditorCanvas
        canMountEditor={false}
        documentId="doc-1"
        currentRevisionId="rev-1"
        editorMode="document-edit"
        author="Ana Autora"
      />,
    );
    expect(container).toBeEmptyDOMElement();
  });

  it('mounts DocumentShell (writable) when canMountEditor is true (default)', async () => {
    render(
      <EditorCanvas
        documentId="doc-1"
        currentRevisionId="rev-1"
        editorMode="document-edit"
        author="Ana Autora"
      />,
    );
    await waitFor(() => expect(screen.getByTestId('editor')).toBeInTheDocument());
    expect(screen.getByTestId('editor')).toHaveAttribute('data-mode', 'document-edit');
  });

  it('passes chrome slots through to the underlying DocumentShell/EditorChrome', async () => {
    render(
      <EditorCanvas
        documentId="doc-1"
        currentRevisionId="rev-1"
        editorMode="readonly"
        author="Ana Autora"
        chrome={{ center: <span>Título</span> }}
      />,
    );
    await waitFor(() => expect(screen.getByTestId('editor')).toBeInTheDocument());
    expect(screen.getByText('Título')).toBeInTheDocument();
  });
});
