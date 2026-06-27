import React from 'react';
import { render } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';

const dispatch = vi.fn();
const focus = vi.fn();
const insertText = vi.fn(() => ({ __tr: true }));
const fakeState = { selection: { from: 3, to: 3 }, tr: { insertText } };
const fakeView = { state: fakeState, dispatch, focus };
const fakeEditorRef = { getView: () => fakeView, getState: () => fakeState };

let tags: Array<{ type: string; name: string }> = [];

vi.mock('@eigenpal/docx-editor-react', async () => {
  const ReactMod = await import('react');
  return {
    DocxEditor: ReactMod.forwardRef((_props: Record<string, unknown>, ref: React.Ref<unknown>) => {
      ReactMod.useImperativeHandle(ref, () => ({
        getEditorRef: () => fakeEditorRef,
        save: async () => new ArrayBuffer(0),
        getTotalPages: () => 1,
        focus,
      }));
      return null;
    }),
    createEmptyDocument: () => ({}),
  };
});

vi.mock('@eigenpal/docx-editor-react/plugin-api', () => ({
  PluginHost: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  templatePlugin: {},
  getTemplatePluginTags: () => tags,
}));

vi.mock('@eigenpal/docx-editor-react/styles.css', () => ({}));

import { MetalDocsEditor } from './MetalDocsEditor';
import type { MetalDocsEditorRef } from './types';

describe('MetalDocsEditor token capability', () => {
  it('insertToken inserts {key} at the current selection via the live view', () => {
    const ref = React.createRef<MetalDocsEditorRef>();
    render(<MetalDocsEditor ref={ref} mode="template-draft" documentBuffer={new ArrayBuffer(8)} />);
    ref.current!.insertToken('doc_code');
    expect(insertText).toHaveBeenCalledWith('{doc_code}', 3, 3);
    expect(dispatch).toHaveBeenCalledWith({ __tr: true });
  });

  it('getUsedTokens returns variable token keys only, without braces', () => {
    tags = [
      { type: 'variable', name: 'doc_code' },
      { type: 'sectionStart', name: 'loop' },
      { type: 'variable', name: 'author' },
    ];
    const ref = React.createRef<MetalDocsEditorRef>();
    render(<MetalDocsEditor ref={ref} mode="template-draft" documentBuffer={new ArrayBuffer(8)} />);
    expect(ref.current!.getUsedTokens()).toEqual(['doc_code', 'author']);
  });
});
