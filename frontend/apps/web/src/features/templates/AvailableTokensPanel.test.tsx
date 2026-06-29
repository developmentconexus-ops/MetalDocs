import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { AvailableTokensPanel } from './AvailableTokensPanel';
import type { Token } from './tokens/tokenCatalog';

const catalog: Token[] = [
  { key: 'doc_code', label: 'Código do documento', kind: 'computed' },
  { key: 'author', label: 'Autor', kind: 'computed' },
  { key: 'COMPANY_NAME', label: 'Empresa', kind: 'dictionary' },
];

describe('AvailableTokensPanel', () => {
  it('renders catalog tokens and calls onInsert with the key on click', () => {
    const onInsert = vi.fn();
    render(<AvailableTokensPanel catalog={catalog} usedKeys={new Set()} unknownTokens={[]} invalidTokens={[]} onInsert={onInsert} />);
    fireEvent.click(screen.getByTestId('token-doc_code').querySelector('button')!);
    expect(onInsert).toHaveBeenCalledWith('doc_code');
  });

  it('marks tokens present in the document as used', () => {
    render(<AvailableTokensPanel catalog={catalog} usedKeys={new Set(['author'])} unknownTokens={[]} invalidTokens={[]} onInsert={() => {}} />);
    expect(screen.getByTestId('token-author').getAttribute('data-used')).toBe('true');
    expect(screen.getByTestId('token-doc_code').getAttribute('data-used')).toBe('false');
  });

  it('warns about unknown tokens', () => {
    render(<AvailableTokensPanel catalog={catalog} usedKeys={new Set()} unknownTokens={['nope']} invalidTokens={[]} onInsert={() => {}} />);
    expect(screen.getByTestId('unknown-nope')).toBeTruthy();
  });

  it('warns about invalid-format tokens', () => {
    render(<AvailableTokensPanel catalog={catalog} usedKeys={new Set()} unknownTokens={[]} invalidTokens={['a.b']} onInsert={() => {}} />);
    expect(screen.getByTestId('invalid-a.b')).toBeTruthy();
  });

  it('warns on catalog load failure and suppresses the misleading unknown-token list', () => {
    render(
      <AvailableTokensPanel
        catalog={[]}
        catalogError
        usedKeys={new Set()}
        unknownTokens={['doc_code', 'author']}
        invalidTokens={[]}
        onInsert={() => {}}
      />,
    );
    expect(screen.getByTestId('catalog-error')).toBeTruthy();
    // Without a catalog every token would falsely classify as unknown - don't show that list.
    expect(screen.queryByTestId('unknown-doc_code')).toBeNull();
  });

  it('prevents default on token mousedown so the editor keeps focus/caret', () => {
    render(<AvailableTokensPanel catalog={catalog} usedKeys={new Set()} unknownTokens={[]} invalidTokens={[]} onInsert={() => {}} />);
    const btn = screen.getByTestId('token-doc_code').querySelector('button')!;
    const ev = new MouseEvent('mousedown', { bubbles: true, cancelable: true });
    const prevented = !btn.dispatchEvent(ev);
    expect(prevented).toBe(true);
  });

  it('renders computed section heading', () => {
    render(<AvailableTokensPanel catalog={catalog} usedKeys={new Set()} unknownTokens={[]} invalidTokens={[]} onInsert={() => {}} />);
    expect(screen.getByText('Preenchido pelo sistema (seguro)')).toBeTruthy();
    expect(screen.getByTestId('token-doc_code')).toBeTruthy();
  });

  it('renders dictionary section with its heading and tokens', () => {
    render(<AvailableTokensPanel catalog={catalog} usedKeys={new Set()} unknownTokens={[]} invalidTokens={[]} onInsert={() => {}} />);
    expect(screen.getByText('Definido pela sua organização')).toBeTruthy();
    expect(screen.getByTestId('token-COMPANY_NAME')).toBeTruthy();
  });

  it('calls onInsert with the dictionary token key on click', () => {
    const onInsert = vi.fn();
    render(<AvailableTokensPanel catalog={catalog} usedKeys={new Set()} unknownTokens={[]} invalidTokens={[]} onInsert={onInsert} />);
    fireEvent.click(screen.getByTestId('token-COMPANY_NAME').querySelector('button')!);
    expect(onInsert).toHaveBeenCalledWith('COMPANY_NAME');
  });

  it('does not render dictionary section when no dictionary tokens exist', () => {
    const computedOnly: Token[] = [
      { key: 'doc_code', label: 'Código', kind: 'computed' },
    ];
    render(<AvailableTokensPanel catalog={computedOnly} usedKeys={new Set()} unknownTokens={[]} invalidTokens={[]} onInsert={() => {}} />);
    expect(screen.queryByText('Definido pela sua organização')).toBeNull();
  });
});
