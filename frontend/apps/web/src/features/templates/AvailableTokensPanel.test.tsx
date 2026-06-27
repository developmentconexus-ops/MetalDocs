import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { AvailableTokensPanel } from './AvailableTokensPanel';
import type { PlaceholderCatalogEntry } from './api/catalog';

const catalog: PlaceholderCatalogEntry[] = [
  { key: 'doc_code', label: 'Código do documento' } as PlaceholderCatalogEntry,
  { key: 'author', label: 'Autor' } as PlaceholderCatalogEntry,
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

  it('prevents default on token mousedown so the editor keeps focus/caret', () => {
    render(<AvailableTokensPanel catalog={catalog} usedKeys={new Set()} unknownTokens={[]} invalidTokens={[]} onInsert={() => {}} />);
    const btn = screen.getByTestId('token-doc_code').querySelector('button')!;
    const ev = new MouseEvent('mousedown', { bubbles: true, cancelable: true });
    const prevented = !btn.dispatchEvent(ev);
    expect(prevented).toBe(true);
  });
});
