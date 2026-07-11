import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { QuorumPills } from './QuorumPills';

describe('QuorumPills', () => {
  it('renders the three quorum options', () => {
    render(<QuorumPills id="quorum-1" value="any_1_of" onChange={vi.fn()} />);
    expect(screen.getByRole('radio', { name: 'Qualquer um (1 de N)' })).toBeInTheDocument();
    expect(screen.getByRole('radio', { name: 'Todos (N de N)' })).toBeInTheDocument();
    expect(screen.getByRole('radio', { name: 'M de N' })).toBeInTheDocument();
  });

  it('reflects the current value', () => {
    render(<QuorumPills id="quorum-1" value="all_of" onChange={vi.fn()} />);
    expect(screen.getByRole('radio', { name: 'Todos (N de N)' })).toBeChecked();
  });

  it('fires onChange with the selected kind', () => {
    const onChange = vi.fn();
    render(<QuorumPills id="quorum-1" value="any_1_of" onChange={onChange} />);
    fireEvent.click(screen.getByRole('radio', { name: 'M de N' }));
    expect(onChange).toHaveBeenCalledWith('m_of_n');
  });

  it('shows help text for the current value', () => {
    render(<QuorumPills id="quorum-1" value="any_1_of" onChange={vi.fn()} />);
    expect(
      screen.getByText('Aprovado quando qualquer aprovador da etapa assina.'),
    ).toBeInTheDocument();
  });

  it('exposes an accessible name via ariaLabelledBy', () => {
    render(
      <>
        <span id="quorum-1-label">Quórum da etapa 1</span>
        <QuorumPills
          id="quorum-1"
          value="any_1_of"
          onChange={vi.fn()}
          ariaLabelledBy="quorum-1-label"
        />
      </>,
    );
    expect(screen.getByRole('radiogroup', { name: 'Quórum da etapa 1' })).toBeInTheDocument();
  });
});
