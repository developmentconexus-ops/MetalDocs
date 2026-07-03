import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { LibraryFilterTabs } from './LibraryFilterTabs';

// FE-19 N8: Filtros/Exportar were removed (not shipped permanently disabled) —
// no working control exists until server-side advanced filters + export job
// endpoints ship.
describe('LibraryFilterTabs', () => {
  it('does not render the disabled Filtros/Exportar controls', () => {
    render(
      <LibraryFilterTabs activeTab="todos" onTabChange={() => {}} statsByStatus={{}} total={0} />,
    );

    expect(screen.queryByText('Filtros')).toBeNull();
    expect(screen.queryByText('Exportar')).toBeNull();
  });

  it('still fires onTabChange for a real tab', () => {
    const onTabChange = vi.fn();
    render(
      <LibraryFilterTabs
        activeTab="todos"
        onTabChange={onTabChange}
        statsByStatus={{ under_review: 2 }}
        total={5}
      />,
    );

    fireEvent.click(screen.getByRole('tab', { name: /Todos/i }));
    expect(onTabChange).toHaveBeenCalledWith('todos');
  });
});
