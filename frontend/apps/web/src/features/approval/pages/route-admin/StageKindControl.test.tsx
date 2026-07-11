import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { StageKindControl } from './StageKindControl';

describe('StageKindControl', () => {
  it('renders both stage kind options', () => {
    render(<StageKindControl id="stage-1" value="review" onChange={vi.fn()} />);
    expect(screen.getByRole('radio', { name: 'Revisão' })).toBeInTheDocument();
    expect(screen.getByRole('radio', { name: 'Assinatura' })).toBeInTheDocument();
  });

  it('reflects the current value', () => {
    render(<StageKindControl id="stage-1" value="approval" onChange={vi.fn()} />);
    expect(screen.getByRole('radio', { name: 'Assinatura' })).toBeChecked();
    expect(screen.getByRole('radio', { name: 'Revisão' })).not.toBeChecked();
  });

  it('fires onChange with the selected kind', () => {
    const onChange = vi.fn();
    render(<StageKindControl id="stage-1" value="review" onChange={onChange} />);
    fireEvent.click(screen.getByRole('radio', { name: 'Assinatura' }));
    expect(onChange).toHaveBeenCalledWith('approval');
  });

  it('shows the help text for the current value', () => {
    render(<StageKindControl id="stage-1" value="review" onChange={vi.fn()} />);
    expect(screen.getByText('Revisores comentam e conferem; não assinam.')).toBeInTheDocument();
  });
});
