import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import {
  SubmitChoicePicker,
  isSubmitChoiceSelectionComplete,
  submitChoiceStagesOf,
  type ApprovalRoutePreview,
  type SubmitChosenActors,
} from './SubmitChoicePicker';

const usersQueryMock = vi.fn();
vi.mock('../../../../lib/iam/useUsersQuery', () => ({
  useUsersQuery: (...args: unknown[]) => usersQueryMock(...args),
}));

const PREVIEW_WITH_SUBMIT_CHOICE: ApprovalRoutePreview = {
  route_id: 'route-1',
  route_name: 'Rota padrão',
  stages: [
    {
      stage_order: 1,
      label: 'Revisão',
      selectors: [{ kind: 'role_in_document_area', role: 'reviewer' }],
    },
    {
      stage_order: 2,
      label: 'Aprovação',
      selectors: [{ kind: 'submit_choice', role: 'approver', area_code: 'qa' }],
    },
  ],
};

const PREVIEW_NO_SUBMIT_CHOICE: ApprovalRoutePreview = {
  route_id: 'route-2',
  route_name: 'Rota simples',
  stages: [
    { stage_order: 1, label: 'Revisão', selectors: [{ kind: 'role_in_document_area', role: 'reviewer' }] },
  ],
};

describe('submitChoiceStagesOf', () => {
  it('returns only stages carrying a submit_choice selector', () => {
    expect(submitChoiceStagesOf(PREVIEW_WITH_SUBMIT_CHOICE)).toEqual([PREVIEW_WITH_SUBMIT_CHOICE.stages[1]]);
    expect(submitChoiceStagesOf(PREVIEW_NO_SUBMIT_CHOICE)).toEqual([]);
    expect(submitChoiceStagesOf(null)).toEqual([]);
    expect(submitChoiceStagesOf(undefined)).toEqual([]);
  });
});

describe('isSubmitChoiceSelectionComplete', () => {
  const stages = submitChoiceStagesOf(PREVIEW_WITH_SUBMIT_CHOICE);

  it('is false when a submit_choice stage has no entry', () => {
    expect(isSubmitChoiceSelectionComplete(stages, [])).toBe(false);
  });

  it('is false when the entry has an empty user_ids array', () => {
    const value: SubmitChosenActors[] = [{ stage_order: 2, user_ids: [] }];
    expect(isSubmitChoiceSelectionComplete(stages, value)).toBe(false);
  });

  it('is true once every submit_choice stage has >=1 chosen user', () => {
    const value: SubmitChosenActors[] = [{ stage_order: 2, user_ids: ['user-1'] }];
    expect(isSubmitChoiceSelectionComplete(stages, value)).toBe(true);
  });
});

describe('SubmitChoicePicker', () => {
  it('renders nothing when there are no submit_choice stages', () => {
    usersQueryMock.mockReturnValue({ data: { items: [] }, isLoading: false });
    const { container } = render(
      <SubmitChoicePicker stages={[]} value={[]} onChange={vi.fn()} />,
    );
    expect(container.firstChild).toBeNull();
  });

  it('emits chosen_actors keyed by stage_order when a user is checked', () => {
    usersQueryMock.mockReturnValue({
      data: {
        items: [
          { user_id: 'user-1', username: 'ana', display_name: 'Ana Silva' },
          { user_id: 'user-2', username: 'bruno', display_name: 'Bruno Alves' },
        ],
      },
      isLoading: false,
    });

    const stages = submitChoiceStagesOf(PREVIEW_WITH_SUBMIT_CHOICE);
    const onChange = vi.fn();
    render(<SubmitChoicePicker stages={stages} value={[]} onChange={onChange} />);

    fireEvent.click(screen.getByLabelText('Ana Silva (ana)'));

    expect(onChange).toHaveBeenCalledWith([{ stage_order: 2, user_ids: ['user-1'] }]);
  });

  it('shows the zero-selection validation message only when showValidation is true', () => {
    usersQueryMock.mockReturnValue({
      data: { items: [{ user_id: 'user-1', username: 'ana', display_name: 'Ana Silva' }] },
      isLoading: false,
    });
    const stages = submitChoiceStagesOf(PREVIEW_WITH_SUBMIT_CHOICE);

    const { rerender } = render(
      <SubmitChoicePicker stages={stages} value={[]} onChange={vi.fn()} />,
    );
    expect(screen.queryByText('Selecione ao menos um aprovador para esta etapa antes de submeter.')).toBeNull();

    rerender(<SubmitChoicePicker stages={stages} value={[]} onChange={vi.fn()} showValidation />);
    expect(screen.getByText('Selecione ao menos um aprovador para esta etapa antes de submeter.')).toBeTruthy();
  });
});
