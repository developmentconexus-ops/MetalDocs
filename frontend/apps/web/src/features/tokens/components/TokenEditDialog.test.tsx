import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TokenEditDialog } from './TokenEditDialog';

const COMPUTED = ['author', 'doc_code'];

describe('TokenEditDialog', () => {
  it('blocks submit and shows a grammar error for an invalid name', () => {
    const onSubmit = vi.fn();
    render(
      <TokenEditDialog mode="create" computedKeys={COMPUTED} submitting={false} onSubmit={onSubmit} onClose={vi.fn()} />,
    );
    fireEvent.change(screen.getByLabelText('Nome'), { target: { value: '1bad' } });
    fireEvent.change(screen.getByLabelText('Valor'), { target: { value: 'v' } });
    fireEvent.change(screen.getByLabelText('Rótulo'), { target: { value: 'L' } });
    fireEvent.click(screen.getByRole('button', { name: 'Salvar' }));
    expect(onSubmit).not.toHaveBeenCalled();
    expect(screen.getByText(/Nome inválido/)).toBeInTheDocument();
  });

  it('blocks submit on a computed-catalog collision', () => {
    const onSubmit = vi.fn();
    render(
      <TokenEditDialog mode="create" computedKeys={COMPUTED} submitting={false} onSubmit={onSubmit} onClose={vi.fn()} />,
    );
    fireEvent.change(screen.getByLabelText('Nome'), { target: { value: 'author' } });
    fireEvent.change(screen.getByLabelText('Valor'), { target: { value: 'v' } });
    fireEvent.change(screen.getByLabelText('Rótulo'), { target: { value: 'L' } });
    fireEvent.click(screen.getByRole('button', { name: 'Salvar' }));
    expect(onSubmit).not.toHaveBeenCalled();
    expect(screen.getByText(/token do sistema/)).toBeInTheDocument();
  });

  it('submits a valid entry', () => {
    const onSubmit = vi.fn();
    render(
      <TokenEditDialog mode="create" computedKeys={COMPUTED} submitting={false} onSubmit={onSubmit} onClose={vi.fn()} />,
    );
    fireEvent.change(screen.getByLabelText('Nome'), { target: { value: 'company_slogan' } });
    fireEvent.change(screen.getByLabelText('Valor'), { target: { value: 'Qualidade' } });
    fireEvent.change(screen.getByLabelText('Rótulo'), { target: { value: 'Slogan' } });
    fireEvent.click(screen.getByRole('button', { name: 'Salvar' }));
    expect(onSubmit).toHaveBeenCalledWith({ name: 'company_slogan', value: 'Qualidade', label: 'Slogan', description: '' });
  });

  it('disables the name field in edit mode', () => {
    render(
      <TokenEditDialog
        mode="edit"
        computedKeys={COMPUTED}
        submitting={false}
        initial={{ name: 'company_slogan', value: 'v', label: 'L', description: '' }}
        onSubmit={vi.fn()}
        onClose={vi.fn()}
      />,
    );
    expect(screen.getByLabelText('Nome')).toBeDisabled();
  });
});
