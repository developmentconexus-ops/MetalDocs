import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { RouteFlowPreview } from './RouteFlowPreview';

describe('RouteFlowPreview', () => {
  it('shows the empty state when there are no stages', () => {
    render(<RouteFlowPreview stages={[]} />);
    expect(screen.getByText('Adicione etapas para ver o fluxo.')).toBeInTheDocument();
  });

  it('renders stage chips in order with kind and quorum labels', () => {
    render(
      <RouteFlowPreview
        stages={[
          { uid: 'a', label: 'Revisão jurídica', kind: 'review', quorum: 'any_1_of' },
          { uid: 'b', label: 'Assinatura diretoria', kind: 'approval', quorum: 'all_of' },
        ]}
      />,
    );

    const labels = screen.getAllByText(/Revisão jurídica|Assinatura diretoria/);
    expect(labels.map((el) => el.textContent)).toEqual(['Revisão jurídica', 'Assinatura diretoria']);
    expect(screen.getByText('Revisão')).toBeInTheDocument();
    expect(screen.getByText('Assinatura')).toBeInTheDocument();
    expect(screen.getByText('Qualquer um (1 de N)')).toBeInTheDocument();
    expect(screen.getByText('Todos (N de N)')).toBeInTheDocument();
  });
});
