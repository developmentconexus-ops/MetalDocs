import { render, screen, waitFor } from '@testing-library/react';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { DocumentShell } from './DocumentShell';

vi.mock('@metaldocs/editor-ui', () => ({
  MetalDocsEditor: React.forwardRef((props: Record<string, unknown>, ref) => {
    React.useImperativeHandle(ref, () => ({
      async saveNow() { return null; },
      async getDocumentBuffer() { return null; },
      getPageCount() { return 0; },
      focus() {},
    }));
    return (
      <div
        data-testid="editor"
        data-mode={props.mode as string}
        data-has-autosave={props.onAutoSave ? 'true' : 'false'}
      />
    );
  }),
}));

vi.mock('../api/documents', () => ({
  signedRevisionURL: vi.fn().mockReturnValue('/revisions/rev-1/signed-url'),
}));

function makeFetchOk() {
  return vi.fn().mockImplementation((url: string) => {
    if (String(url).includes('signed-url')) {
      return Promise.resolve({ ok: true, json: () => Promise.resolve({ url: 'https://s3/doc.docx' }) });
    }
    return Promise.resolve({ ok: true, arrayBuffer: () => Promise.resolve(new ArrayBuffer(8)) });
  });
}

describe('DocumentShell', () => {
  beforeEach(() => {
    global.fetch = makeFetchOk();
  });

  it('shows the loading state before the buffer resolves', () => {
    global.fetch = vi.fn().mockImplementation(() => new Promise(() => {}));
    render(
      <DocumentShell
        documentId="doc-1"
        currentRevisionId="rev-1"
        editorMode="readonly"
        author="Ana"
      />,
    );
    expect(screen.queryByTestId('editor')).toBeNull();
  });

  it('mounts the editor once the buffer resolves (ready state)', async () => {
    render(
      <DocumentShell
        documentId="doc-1"
        currentRevisionId="rev-1"
        editorMode="readonly"
        author="Ana"
      />,
    );
    await waitFor(() => {
      expect(screen.getByTestId('editor')).toBeTruthy();
    });
    expect(screen.getByTestId('editor').getAttribute('data-mode')).toBe('readonly');
  });

  it('shows an inline alert when the buffer fetch fails', async () => {
    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (String(url).includes('signed-url')) {
        return Promise.resolve({ ok: true, json: () => Promise.resolve({ url: 'https://s3/doc.docx' }) });
      }
      return Promise.resolve({ ok: false, status: 500 });
    });
    render(
      <DocumentShell
        documentId="doc-1"
        currentRevisionId="rev-1"
        editorMode="readonly"
        author="Ana"
      />,
    );
    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeTruthy();
    });
    expect(screen.queryByTestId('editor')).toBeNull();
  });

  it('renders bare editor (no chrome header) when chrome is absent', async () => {
    render(
      <DocumentShell
        documentId="doc-1"
        currentRevisionId="rev-1"
        editorMode="readonly"
        author="Ana"
      />,
    );
    await waitFor(() => expect(screen.getByTestId('editor')).toBeTruthy());
    expect(screen.queryByText('Header Center')).toBeNull();
  });

  it('wraps the editor in EditorChrome when chrome is supplied', async () => {
    render(
      <DocumentShell
        documentId="doc-1"
        currentRevisionId="rev-1"
        editorMode="document-edit"
        author="Ana"
        chrome={{ center: <span>Header Center</span>, right: <span>Header Right</span> }}
      />,
    );
    await waitFor(() => expect(screen.getByTestId('editor')).toBeTruthy());
    expect(screen.getByText('Header Center')).toBeTruthy();
    expect(screen.getByText('Header Right')).toBeTruthy();
  });

  it('mounts fine and mounts editor without onAutoSave when the prop is omitted', async () => {
    render(
      <DocumentShell
        documentId="doc-1"
        currentRevisionId="rev-1"
        editorMode="review"
        author="Ana"
      />,
    );
    await waitFor(() => expect(screen.getByTestId('editor')).toBeTruthy());
    expect(screen.getByTestId('editor').getAttribute('data-has-autosave')).toBe('false');
  });
});
