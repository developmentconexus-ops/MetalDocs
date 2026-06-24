import { describe, it, expect, vi, afterEach } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import * as React from 'react';
import type { ReactNode } from 'react';
import { MetalDocsEditor } from '../src/MetalDocsEditor';

afterEach(cleanup);

vi.mock('@eigenpal/docx-editor-react', () => ({
  createEmptyDocument: () => ({ type: 'empty-doc' }),
  DocxEditor: React.forwardRef((_props, _ref) => <div data-testid="docx-editor-mock" />),
}));

vi.mock('@eigenpal/docx-editor-react/plugin-api', () => ({
  templatePlugin: { name: 'template', id: 'template' },
  PluginHost: ({
    plugins,
    children,
  }: {
    plugins: Array<{ name?: string }>;
    children: ReactNode;
  }) => (
    <div
      data-testid="plugin-host"
      data-plugins={plugins.length}
      data-plugin-names={plugins.map((p) => p.name ?? '?').join(',')}
    >
      {children}
    </div>
  ),
}));

describe('template plugin wiring', () => {
  it('includes templatePlugin and filter-transaction-guard when mode is template-draft', () => {
    render(<MetalDocsEditor mode="template-draft" author="u1" />);
    const host = screen.getByTestId('plugin-host');
    // templatePlugin + filter-transaction-guard = 2 plugins in template-draft mode
    expect(host.getAttribute('data-plugins')).toBe('2');
    expect(host.getAttribute('data-plugin-names')).toContain('template');
    expect(host.getAttribute('data-plugin-names')).toContain('filter-transaction-guard');
  });

  it('omits templatePlugin in document-edit mode', () => {
    render(<MetalDocsEditor mode="document-edit" author="u1" />);
    const host = screen.getByTestId('plugin-host');
    expect(host.getAttribute('data-plugins')).toBe('0');
    expect(host.getAttribute('data-plugin-names') ?? '').not.toContain('template');
  });

  it('omits templatePlugin in readonly mode', () => {
    render(<MetalDocsEditor mode="readonly" author="u1" />);
    const host = screen.getByTestId('plugin-host');
    expect(host.getAttribute('data-plugins')).toBe('0');
  });
});
