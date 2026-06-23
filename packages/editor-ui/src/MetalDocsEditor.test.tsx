import React from 'react';
import { render } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';

const received: Record<string, unknown> = {};

// Mock specifiers must match what MetalDocsEditor.tsx actually imports from.
// MetalDocsEditor imports: DocxEditor, createEmptyDocument, DocxEditorRef from '@eigenpal/docx-editor-react'
vi.mock('@eigenpal/docx-editor-react', () => ({
  DocxEditor: (props: Record<string, unknown>) => { Object.assign(received, props); return null; },
  createEmptyDocument: () => ({}),
}));

// MetalDocsEditor imports: PluginHost, templatePlugin, EditorPlugin from '@eigenpal/docx-editor-react/plugin-api'
vi.mock('@eigenpal/docx-editor-react/plugin-api', () => ({
  PluginHost: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  templatePlugin: {},
}));

vi.mock('@eigenpal/docx-editor-react/styles.css', () => ({}));

// Mock the sidebarModelBridge plugin used internally
vi.mock('./plugins/sidebarModelBridge', () => ({
  buildSidebarModelPlugin: () => ({}),
}));

import { MetalDocsEditor } from './MetalDocsEditor';

describe('MetalDocsEditor mode mapping', () => {
  it('maps review mode to eigenpal suggesting', () => {
    render(<MetalDocsEditor mode="review" documentBuffer={new ArrayBuffer(8)} author="Rev" />);
    expect(received.mode).toBe('suggesting');
  });
  it('maps readonly to viewing and document-edit to editing', () => {
    render(<MetalDocsEditor mode="readonly" documentBuffer={new ArrayBuffer(8)} />);
    expect(received.mode).toBe('viewing');
    render(<MetalDocsEditor mode="document-edit" documentBuffer={new ArrayBuffer(8)} />);
    expect(received.mode).toBe('editing');
  });
});
