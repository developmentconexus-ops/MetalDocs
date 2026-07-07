// @vitest-environment jsdom
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { InboxFilters, toInboxParams, type InboxFilterState } from './InboxFilters';

const baseState: InboxFilterState = { stageKind: undefined, due: 'all', oversee: false };

describe('InboxFilters', () => {
  it('maps stage-kind select to stage_kind param', () => {
    const onChange = vi.fn();
    render(<InboxFilters value={baseState} onChange={onChange} />);

    fireEvent.change(screen.getByLabelText('Estágio'), { target: { value: 'review' } });

    expect(onChange).toHaveBeenCalledWith({ ...baseState, stageKind: 'review' });
  });

  it('maps due select to due filter state', () => {
    const onChange = vi.fn();
    render(<InboxFilters value={baseState} onChange={onChange} />);

    fireEvent.change(screen.getByLabelText('Vencimento'), { target: { value: 'overdue' } });

    expect(onChange).toHaveBeenCalledWith({ ...baseState, due: 'overdue' });
  });

  it('maps oversee toggle to oversee flag', () => {
    const onChange = vi.fn();
    render(<InboxFilters value={baseState} onChange={onChange} />);

    fireEvent.click(screen.getByLabelText(/Supervisão/i));

    expect(onChange).toHaveBeenCalledWith({ ...baseState, oversee: true });
  });

  it('shows oversee-denied note when oversee403 is set', () => {
    render(<InboxFilters value={baseState} onChange={vi.fn()} oversee403 />);

    expect(screen.getByRole('alert')).toHaveTextContent('Você não tem permissão de supervisão.');
  });

  it('does not show the oversee-denied note by default', () => {
    render(<InboxFilters value={baseState} onChange={vi.fn()} />);

    expect(screen.queryByRole('alert')).toBeNull();
  });
});

describe('toInboxParams', () => {
  const NOW = new Date('2026-07-07T12:00:00.000Z').getTime();

  it('maps stageKind to stage_kind', () => {
    expect(toInboxParams({ stageKind: 'approval', due: 'all', oversee: false }, NOW)).toEqual({
      stage_kind: 'approval',
    });
  });

  it('maps due=next7 to a due_before 7 days out', () => {
    const params = toInboxParams({ due: 'next7', oversee: false }, NOW);
    expect(params.due_before).toBe(new Date(NOW + 7 * 24 * 60 * 60 * 1000).toISOString());
  });

  it('maps due=overdue to a due_before of now', () => {
    const params = toInboxParams({ due: 'overdue', oversee: false }, NOW);
    expect(params.due_before).toBe(new Date(NOW).toISOString());
  });

  it('maps due=all to no due_before', () => {
    const params = toInboxParams({ due: 'all', oversee: false }, NOW);
    expect(params.due_before).toBeUndefined();
  });

  it('maps oversee to scope=oversee', () => {
    const params = toInboxParams({ due: 'all', oversee: true }, NOW);
    expect(params.scope).toBe('oversee');
  });

  it('omits scope when oversee is false', () => {
    const params = toInboxParams({ due: 'all', oversee: false }, NOW);
    expect(params.scope).toBeUndefined();
  });
});
