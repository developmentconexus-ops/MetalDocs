import '@testing-library/jest-dom/vitest';
import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import type { ReactNode } from 'react';

afterEach(cleanup);
import { MetalDocsEditor } from '../src/MetalDocsEditor';

vi.mock('@eigenpal/docx-js-editor', async () => {
  const React = await vi.importActual<typeof import('react')>('react');
  type MockProps = {
    documentBuffer?: ArrayBuffer;
    document?: unknown;
    mode?: 'editing' | 'viewing';
    renderTitleBarRight?: () => ReactNode;
  };

  return {
    templatePlugin: { id: 'template', name: 'template' },
    PluginHost: ({ children }: { children: ReactNode }) => <>{children}</>,
    DocxEditor: React.forwardRef<unknown, MockProps>(({
      documentBuffer,
      document,
      mode,
      renderTitleBarRight,
    }, _ref) => (
      <div
        data-testid="docx-editor-mock"
        data-has-buffer={documentBuffer ? 'yes' : 'no'}
        data-has-document={document ? 'yes' : 'no'}
      >
        {mode === 'editing' ? <div role="toolbar" data-testid="toolbar" /> : null}
        {renderTitleBarRight ? renderTitleBarRight() : null}
      </div>
    )),
  };
});

vi.mock('@eigenpal/docx-js-editor/core', () => ({
  createEmptyDocument: () => ({ type: 'empty-doc' }),
}));

describe('MetalDocsEditor', () => {
  it('shows toolbar in document-edit mode', () => {
    const buf = new ArrayBuffer(8);
    render(<MetalDocsEditor mode="document-edit" documentBuffer={buf} author="u1" />);
    const el = screen.getByTestId('docx-editor-mock');
    expect(el.getAttribute('data-has-buffer')).toBe('yes');
    expect(el.getAttribute('data-has-document')).toBe('no');
    expect(screen.queryByRole('toolbar')).not.toBeNull();
  });

  it('starts editable blank editors with an Eigenpal empty document', () => {
    render(<MetalDocsEditor mode="template-draft" author="u1" />);
    const el = screen.getByTestId('docx-editor-mock');
    expect(el.getAttribute('data-has-buffer')).toBe('no');
    expect(el.getAttribute('data-has-document')).toBe('yes');
    expect(screen.queryByRole('toolbar')).not.toBeNull();
  });

  it('hides toolbar in readonly mode', () => {
    render(<MetalDocsEditor mode="readonly" author="u1" />);
    const el = screen.getByTestId('docx-editor-mock');
    expect(el.getAttribute('data-has-buffer')).toBe('no');
    expect(el.getAttribute('data-has-document')).toBe('no');
    expect(screen.queryByRole('toolbar')).toBeNull();
  });

  it('renders renderTitleBarRight slot', () => {
    render(
      <MetalDocsEditor
        mode="document-edit"
        author="u1"
        renderTitleBarRight={() => <span data-testid="sentinel" />}
      />
    );
    expect(screen.queryByTestId('sentinel')).not.toBeNull();
  });
});
