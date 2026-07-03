import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { LibraryStatCards } from './LibraryStatCards';

// FE-19 N8: only the "Em revisão" card has a real backend-fed count
// (under_review from GET /documents/stats). The other three placeholder
// cards ("sem fonte de dados") were removed rather than shipped disabled.
describe('LibraryStatCards', () => {
  it('renders only the real "Em revisão" card', () => {
    render(<LibraryStatCards statsByStatus={{ under_review: 4 }} />);

    expect(screen.getByText('Em revisão')).toBeTruthy();
    expect(screen.getByText('4')).toBeTruthy();

    expect(screen.queryByText('Aprovação pendente')).toBeNull();
    expect(screen.queryByText('Frozen este mês')).toBeNull();
    expect(screen.queryByText('Próx. revisão')).toBeNull();
    expect(screen.queryByText('sem fonte de dados')).toBeNull();
  });

  it('falls back to 0 when under_review is absent from the stats map', () => {
    render(<LibraryStatCards statsByStatus={{}} />);
    expect(screen.getByText('0')).toBeTruthy();
  });
});
