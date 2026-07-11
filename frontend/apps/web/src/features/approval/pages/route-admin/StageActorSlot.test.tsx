import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { StageActorSlot } from './StageActorSlot';

describe('StageActorSlot', () => {
  it('renders the default heading and children', () => {
    render(
      <StageActorSlot>
        <div>role + area selects</div>
      </StageActorSlot>,
    );
    expect(screen.getByText('Quem atua nesta etapa')).toBeInTheDocument();
    expect(screen.getByText('role + area selects')).toBeInTheDocument();
  });

  it('renders a custom heading when provided', () => {
    render(
      <StageActorSlot heading="Atores da etapa 2">
        <div>content</div>
      </StageActorSlot>,
    );
    expect(screen.getByText('Atores da etapa 2')).toBeInTheDocument();
  });
});
