import { describe, it, expect, vi, afterEach } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import { MetalDocsEditor } from '../src/MetalDocsEditor';

afterEach(cleanup);

vi.mock('@eigenpal/docx-js-editor', () => ({
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
  DocxEditor: () => <div data-testid="docx-editor-mock" />,
}));

describe('template plugin wiring', () => {
  it('includes templatePlugin only when mode is template-draft', () => {
    render(<MetalDocsEditor mode="template-draft" author="u1" />);
    const host = screen.getByTestId('plugin-host');
    expect(host.getAttribute('data-plugins')).toBe('1');
    expect(host.getAttribute('data-plugin-names')).toContain('template');
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

  it('adds the sidebar bridge plugin alongside templatePlugin in template-draft mode', () => {
    render(
      <MetalDocsEditor
        mode="template-draft"
        author="u1"
        sidebarModel={{
          used: ['a'],
          missing: [],
          orphans: [],
          bannerError: false,
          errorCategories: [],
        }}
      />
    );
    const host = screen.getByTestId('plugin-host');
    const names = host.getAttribute('data-plugin-names') ?? '';
    expect(host.getAttribute('data-plugins')).toBe('2');
    expect(names).toContain('template');
    expect(names).toContain('metaldocs-sidebar-model');
  });

  it('includes external plugins regardless of mode', () => {
    render(
      <MetalDocsEditor
        mode="document-edit"
        author="u1"
        externalPlugins={[{ id: 'custom', name: 'custom' } as never]}
      />
    );
    const host = screen.getByTestId('plugin-host');
    expect(host.getAttribute('data-plugins')).toBe('1');
    expect(host.getAttribute('data-plugin-names')).toContain('custom');
  });
});
