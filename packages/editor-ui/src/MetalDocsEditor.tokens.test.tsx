import React from 'react';
import { render, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

const bodyDispatch = vi.fn();
const headerDispatch = vi.fn();
const footerDispatch = vi.fn();
const focus = vi.fn();

// Toggle to simulate the vendor handing back a null editor ref (no views yet).
let editorRefNull = false;

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
      const footerRef = ReactMod.useRef<HTMLDivElement>(null);
      ReactMod.useImperativeHandle(ref, () => ({
        getEditorRef: () =>
          editorRefNull
            ? null
            : {
                getView: () =>
                  bodyRef.current ? makeView(bodyRef.current, bodyDispatch, 'Body {doc_code} here') : null,
                getHfPmViews: () => {
                  const m = new Map<string, ReturnType<typeof makeView>>();
                  if (headerRef.current) m.set('rId2', makeView(headerRef.current, headerDispatch, 'Header {author}'));
                  if (footerRef.current) m.set('rId3', makeView(footerRef.current, footerDispatch, 'Footer {effective_date}'));
                  return m;
                },
              },
        save: async () => new ArrayBuffer(0),
        getTotalPages: () => 1,
        focus,
      }));
      return (
        <div>
          <div ref={bodyRef} className="ProseMirror" data-band="body" tabIndex={-1} />
          <div ref={headerRef} className="ProseMirror" data-band="header" tabIndex={-1} />
          <div ref={footerRef} className="ProseMirror" data-band="footer" tabIndex={-1} />
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
  footerDispatch.mockClear();
  editorRefNull = false;
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

  it('insertToken targets the footer band once it gains focus', () => {
    const ref = React.createRef<MetalDocsEditorRef>();
    const { container } = render(
      <MetalDocsEditor ref={ref} mode="template-draft" documentBuffer={new ArrayBuffer(8)} />,
    );
    fireEvent.focusIn(container.querySelector('[data-band="footer"]') as HTMLElement);
    ref.current!.insertToken('effective_date');
    expect(footerDispatch).toHaveBeenCalledWith({ t: '{effective_date}', from: 2, to: 2 });
    expect(bodyDispatch).not.toHaveBeenCalled();
    expect(headerDispatch).not.toHaveBeenCalled();
  });

  it('getUsedTokens unions {name} tokens across body + header/footer views', () => {
    const ref = React.createRef<MetalDocsEditorRef>();
    render(<MetalDocsEditor ref={ref} mode="template-draft" documentBuffer={new ArrayBuffer(8)} />);
    expect(ref.current!.getUsedTokens()).toEqual(['doc_code', 'author', 'effective_date']);
  });

  it('surfaces header/footer edits as a change signal (vendor onChange is body-only)', () => {
    const onChange = vi.fn();
    const onAutoSave = vi.fn().mockResolvedValue(undefined);
    const ref = React.createRef<MetalDocsEditorRef>();
    const { container } = render(
      <MetalDocsEditor
        ref={ref}
        mode="template-draft"
        documentBuffer={new ArrayBuffer(8)}
        onChange={onChange}
        onAutoSave={onAutoSave}
      />,
    );
    // A ProseMirror edit in the footer band bubbles an `input` event; the
    // delegated root listener must turn it into a consumer change signal so
    // token-usage refresh (and autosave) cover non-body views.
    fireEvent.input(container.querySelector('[data-band="footer"]') as HTMLElement);
    expect(onChange).toHaveBeenCalled();
  });

  it('no-ops safely when the vendor editor ref is null', () => {
    editorRefNull = true;
    const ref = React.createRef<MetalDocsEditorRef>();
    render(<MetalDocsEditor ref={ref} mode="template-draft" documentBuffer={new ArrayBuffer(8)} />);
    expect(ref.current!.getUsedTokens()).toEqual([]);
    expect(() => ref.current!.insertToken('x')).not.toThrow();
    expect(bodyDispatch).not.toHaveBeenCalled();
    expect(headerDispatch).not.toHaveBeenCalled();
    expect(footerDispatch).not.toHaveBeenCalled();
  });
});
