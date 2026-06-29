import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TokenList } from './TokenList';
import type { TokenDictionaryEntry } from '../api/tokensTypes';

const ENTRY: TokenDictionaryEntry = {
  id: '1', name: 'company_slogan', value: 'Qualidade desde 1990',
  label: 'Slogan', description: 'Slogan institucional',
  created_at: '2026-06-01T00:00:00Z', updated_at: '2026-06-01T00:00:00Z',
};

describe('TokenList', () => {
  it('renders entry rows', () => {
    render(<TokenList entries={[ENTRY]} canManage={true} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText('company_slogan')).toBeInTheDocument();
    expect(screen.getByText('Slogan')).toBeInTheDocument();
  });

  it('shows edit/delete only when canManage', () => {
    const { rerender } = render(<TokenList entries={[ENTRY]} canManage={false} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.queryByRole('button', { name: 'Editar' })).not.toBeInTheDocument();
    rerender(<TokenList entries={[ENTRY]} canManage={true} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByRole('button', { name: 'Editar' })).toBeInTheDocument();
  });

  it('fires onEdit with the entry', () => {
    const onEdit = vi.fn();
    render(<TokenList entries={[ENTRY]} canManage={true} onEdit={onEdit} onDelete={vi.fn()} />);
    fireEvent.click(screen.getByRole('button', { name: 'Editar' }));
    expect(onEdit).toHaveBeenCalledWith(ENTRY);
  });

  it('renders an empty state', () => {
    render(<TokenList entries={[]} canManage={true} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText('Nenhum token cadastrado.')).toBeInTheDocument();
  });
});
