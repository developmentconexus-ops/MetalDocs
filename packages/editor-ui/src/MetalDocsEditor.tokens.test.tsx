import React from 'react';
import { render, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

const bodyDispatch = vi.fn();
const headerDispatch = vi.fn();
const focus = vi.fn();

const makeView = (dom: HTMLElement, dispatch: typeof bodyDispatch, text: string) => ({
  dom,
  dispatch,
  focus: vi.fn(),
  state: {
    selection: { from: 2, to: 2 },
    tr: { insertText: (t: string, from: number, to: number) => ({ t, from, to }) },
    doc: { content: { size: text.length }, textBetween: () => text },
  },
});

vi.mock('@eigenpal/docx-editor-react', async () => {
  const ReactMod = await import('react');
  return {
    DocxEditor: ReactMod.forwardRef((_props: Record<string, unknown>, ref: React.Ref<unknown>) => {
      const bodyRef = ReactMod.useRef<HTMLDivElement>(null);
      const headerRef = ReactMod.useRef<HTMLDivElement>(null);
      ReactMod.useImperativeHandle(ref, () => ({
        getEditorRef: () => ({
          getView: () =>
            bodyRef.current ? makeView(bodyRef.current, bodyDispatch, 'Body {doc_code} here') : null,
          getHfPmViews: () => {
            const m = new Map<string, ReturnType<typeof makeView>>();
            if (headerRef.current) m.set('rId2', makeView(headerRef.current, headerDispatch, 'Header {author}'));
            return m;
          },
        }),
        save: async () => new ArrayBuffer(0),
        getTotalPages: () => 1,
        focus,
      }));
      return (
        <div>
          <div ref={bodyRef} className="ProseMirror" data-band="body" tabIndex={-1} />
          <div ref={headerRef} className="ProseMirror" data-band="header" tabIndex={-1} />
        </div>
      );
    }),
    createEmptyDocument: () => ({}),
  };
});

vi.mock('@eigenpal/docx-editor-react/plugin-api', () => ({
  PluginHost: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  templatePlugin: {},
}));

vi.mock('@eigenpal/docx-editor-react/styles.css', () => ({}));

import { MetalDocsEditor } from './MetalDocsEditor';
import type { MetalDocsEditorRef } from './types';

beforeEach(() => {
  bodyDispatch.mockClear();
  headerDispatch.mockClear();
});

describe('MetalDocsEditor section-aware tokens', () => {
  it('insertToken targets the body view when nothing has been focused', () => {
    const ref = React.createRef<MetalDocsEditorRef>();
    render(<MetalDocsEditor ref={ref} mode="template-draft" documentBuffer={new ArrayBuffer(8)} />);
    ref.current!.insertToken('doc_code');
    expect(bodyDispatch).toHaveBeenCalledWith({ t: '{doc_code}', from: 2, to: 2 });
    expect(headerDispatch).not.toHaveBeenCalled();
  });

  it('insertToken targets the header band once it gains focus', () => {
    const ref = React.createRef<MetalDocsEditorRef>();
    const { container } = render(
      <MetalDocsEditor ref={ref} mode="template-draft" documentBuffer={new ArrayBuffer(8)} />,
    );
    fireEvent.focusIn(container.querySelector('[data-band="header"]') as HTMLElement);
    ref.current!.insertToken('author');
    expect(headerDispatch).toHaveBeenCalledWith({ t: '{author}', from: 2, to: 2 });
    expect(bodyDispatch).not.toHaveBeenCalled();
  });

  it('getUsedTokens unions {name} tokens across body + header/footer views', () => {
    const ref = React.createRef<MetalDocsEditorRef>();
    render(<MetalDocsEditor ref={ref} mode="template-draft" documentBuffer={new ArrayBuffer(8)} />);
    expect(ref.current!.getUsedTokens()).toEqual(['doc_code', 'author']);
  });
});
