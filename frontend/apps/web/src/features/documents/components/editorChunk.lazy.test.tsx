import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

// Hoisted counter: how many times the editor-ui module factory has been evaluated.
const editorMod = vi.hoisted(() => ({ evalCount: 0 }));

vi.mock('@metaldocs/editor-ui', () => {
  editorMod.evalCount++;
  return {
    // Minimal stand-in; real one is the TipTap editor. data-testid lets us await its mount.
    MetalDocsEditor: () => <div data-testid="editor" />,
  };
});

// Deliver a ready buffer so DocumentShell reaches the editor-mount branch.
vi.mock('../../../lib/api', () => ({
  apiFetch: vi.fn(async () => ({ url: 'https://signed.example/doc' })),
}));

afterEach(() => {
  vi.restoreAllMocks();
});

describe('F2d.5b D2 — MetalDocsEditor is a lazy chunk', () => {
  it('importing DocumentShell does NOT evaluate the editor-ui module', async () => {
    const before = editorMod.evalCount;
    await import('./DocumentShell');
    // Static import graph of DocumentShell must not pull editor-ui.
    expect(editorMod.evalCount).toBe(before);
  });

  it('mounting the ready docx canvas lazily fetches + evaluates the editor chunk', async () => {
    // Stub the signed-file fetch so buffer resolves to a real ArrayBuffer.
    const fetchSpy = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValue(new Response(new ArrayBuffer(8), { status: 200 }));

    const { DocumentShell } = await import('./DocumentShell');
    render(
      <DocumentShell
        documentId="doc-1"
        currentRevisionId="rev-1"
        editorMode="readonly"
        author="QA"
      />,
    );

    // Editor mounts only after: buffer fetch resolves AND the lazy chunk resolves via Suspense.
    expect(await screen.findByTestId('editor')).toBeInTheDocument();
    expect(editorMod.evalCount).toBeGreaterThan(0);
    fetchSpy.mockRestore();
  });
});
