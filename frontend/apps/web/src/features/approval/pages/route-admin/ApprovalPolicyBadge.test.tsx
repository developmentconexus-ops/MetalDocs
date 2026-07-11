import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { routePolicyFor } from './routeGovernance';
import { ApprovalPolicyBadge } from './ApprovalPolicyBadge';

describe('ApprovalPolicyBadge', () => {
  it('renders the required-signature badge for controlado', () => {
    render(<ApprovalPolicyBadge policy={routePolicyFor('controlado')} />);
    expect(screen.getByText('Obrigatório ≥1 assinatura')).toBeInTheDocument();
  });

  it('renders the optional-signature badge for simples', () => {
    render(<ApprovalPolicyBadge policy={routePolicyFor('simples')} />);
    expect(screen.getByText('Assinatura opcional')).toBeInTheDocument();
  });

  it('renders the blocked badge for livre', () => {
    render(<ApprovalPolicyBadge policy={routePolicyFor('livre')} />);
    expect(screen.getByText('Perfil livre — sem rota de aprovação')).toBeInTheDocument();
  });
});
