import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import type { ProcessArea } from '../../../taxonomy/types';
import { StageCard, type ManagedUserCore, type SelectorDraft, type StageDraft } from './StageCard';

function makeArea(overrides: Partial<ProcessArea> = {}): ProcessArea {
  return {
    code: 'AREA-01',
    name: 'Área 01',
    description: '',
    parentCode: null,
    ownerUserId: null,
    defaultApproverRole: null,
    archivedAt: null,
    createdAt: '2026-01-01T00:00:00.000Z',
    ...overrides,
  };
}

function makeUser(overrides: Partial<ManagedUserCore> = {}): ManagedUserCore {
  return {
    user_id: 'user-1',
    username: 'jdoe',
    display_name: 'Jane Doe',
    is_active: true,
    must_change_password: false,
    failed_login_attempts: 0,
    tenant_role: 'author',
    area_memberships: [],
    created_at: '2026-01-01T00:00:00.000Z',
    updated_at: '2026-01-01T00:00:00.000Z',
    ...overrides,
  };
}

function makeSelector(overrides: Partial<SelectorDraft> = {}): SelectorDraft {
  return {
    uid: 'selector-1',
    kind: 'role_in_fixed_area',
    userId: '',
    role: 'approver',
    areaCode: 'AREA-01',
    ...overrides,
  };
}

function makeStage(overrides: Partial<StageDraft> = {}): StageDraft {
  return {
    uid: 'stage-1',
    label: 'Revisão',
    requiredCapability: 'doc.signoff',
    selectors: [makeSelector()],
    quorumKind: 'any_1_of',
    m: '1',
    driftPolicy: 'reduce_quorum',
    stageKind: 'review',
    ...overrides,
  };
}

describe('StageCard', () => {
  it('renders the stage fields with uniquely named radiogroups', () => {
    render(
      <StageCard
        stage={makeStage()}
        stageNumber={1}
        isOnly
        disabled={false}
        roleOptions={[{ code: 'approver', label: 'Aprovador', description: '' }]}
        roleOptionsLoading={false}
        areaOptions={[makeArea()]}
        areaOptionsLoading={false}
        userOptions={[makeUser()]}
        userOptionsLoading={false}
        updateStage={vi.fn()}
        removeStage={vi.fn()}
      />,
    );

    expect(screen.getByLabelText('Nome da etapa 1')).toBeInTheDocument();
    expect(
      screen.getByRole('radiogroup', { name: 'Tipo da etapa 1' }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole('radiogroup', { name: 'Quórum da etapa 1' }),
    ).toBeInTheDocument();
  });

  it('disables Remover when isOnly is true', () => {
    render(
      <StageCard
        stage={makeStage()}
        stageNumber={1}
        isOnly
        disabled={false}
        roleOptions={[]}
        roleOptionsLoading={false}
        areaOptions={[]}
        areaOptionsLoading={false}
        userOptions={[]}
        userOptionsLoading={false}
        updateStage={vi.fn()}
        removeStage={vi.fn()}
      />,
    );

    expect((screen.getByRole('button', { name: 'Remover' }) as HTMLButtonElement).disabled).toBe(
      true,
    );
  });

  it('calls updateStage with the new label on change', () => {
    const updateStage = vi.fn();
    render(
      <StageCard
        stage={makeStage()}
        stageNumber={1}
        isOnly={false}
        disabled={false}
        roleOptions={[]}
        roleOptionsLoading={false}
        areaOptions={[]}
        areaOptionsLoading={false}
        userOptions={[]}
        userOptionsLoading={false}
        updateStage={updateStage}
        removeStage={vi.fn()}
      />,
    );

    fireEvent.change(screen.getByLabelText('Nome da etapa 1'), {
      target: { value: 'Revisão inicial' },
    });
    expect(updateStage).toHaveBeenCalledWith('stage-1', { label: 'Revisão inicial' });
  });

  it('calls removeStage with the stage uid when Remover is clicked', () => {
    const removeStage = vi.fn();
    render(
      <StageCard
        stage={makeStage()}
        stageNumber={1}
        isOnly={false}
        disabled={false}
        roleOptions={[]}
        roleOptionsLoading={false}
        areaOptions={[]}
        areaOptionsLoading={false}
        userOptions={[]}
        userOptionsLoading={false}
        updateStage={vi.fn()}
        removeStage={removeStage}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Remover' }));
    expect(removeStage).toHaveBeenCalledWith('stage-1');
  });

  it('renders a kind picker per selector and add/remove selector controls', () => {
    const updateStage = vi.fn();
    render(
      <StageCard
        stage={makeStage({
          selectors: [makeSelector({ uid: 'selector-1' }), makeSelector({ uid: 'selector-2' })],
        })}
        stageNumber={1}
        isOnly={false}
        disabled={false}
        roleOptions={[{ code: 'approver', label: 'Aprovador', description: '' }]}
        roleOptionsLoading={false}
        areaOptions={[makeArea()]}
        areaOptionsLoading={false}
        userOptions={[makeUser()]}
        userOptionsLoading={false}
        updateStage={updateStage}
        removeStage={vi.fn()}
      />,
    );

    expect(
      screen.getByLabelText('Tipo do seletor 1 da etapa 1'),
    ).toBeInTheDocument();
    expect(
      screen.getByLabelText('Tipo do seletor 2 da etapa 1'),
    ).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Adicionar seletor' })).toBeInTheDocument();

    const removeButtons = screen.getAllByRole('button', { name: 'Remover seletor' });
    expect(removeButtons).toHaveLength(2);
    fireEvent.click(removeButtons[0]);
    expect(updateStage).toHaveBeenCalledWith('stage-1', {
      selectors: [expect.objectContaining({ uid: 'selector-2' })],
    });
  });

  it('disables Remover seletor when the stage has only one selector', () => {
    render(
      <StageCard
        stage={makeStage({ selectors: [makeSelector({ uid: 'only' })] })}
        stageNumber={1}
        isOnly={false}
        disabled={false}
        roleOptions={[]}
        roleOptionsLoading={false}
        areaOptions={[]}
        areaOptionsLoading={false}
        userOptions={[]}
        userOptionsLoading={false}
        updateStage={vi.fn()}
        removeStage={vi.fn()}
      />,
    );

    expect(
      (screen.getByRole('button', { name: 'Remover seletor' }) as HTMLButtonElement).disabled,
    ).toBe(true);
  });

  it('adding a selector appends a new default row', () => {
    const updateStage = vi.fn();
    render(
      <StageCard
        stage={makeStage({ selectors: [makeSelector({ uid: 'only' })] })}
        stageNumber={1}
        isOnly={false}
        disabled={false}
        roleOptions={[]}
        roleOptionsLoading={false}
        areaOptions={[]}
        areaOptionsLoading={false}
        userOptions={[]}
        userOptionsLoading={false}
        updateStage={updateStage}
        removeStage={vi.fn()}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Adicionar seletor' }));
    expect(updateStage).toHaveBeenCalledWith('stage-1', {
      selectors: [
        expect.objectContaining({ uid: 'only' }),
        expect.objectContaining({ kind: 'role_in_fixed_area' }),
      ],
    });
  });

  it('shows only the user select for a named_user selector', () => {
    render(
      <StageCard
        stage={makeStage({
          selectors: [makeSelector({ uid: 'nu', kind: 'named_user', role: '', areaCode: '' })],
        })}
        stageNumber={1}
        isOnly={false}
        disabled={false}
        roleOptions={[{ code: 'approver', label: 'Aprovador', description: '' }]}
        roleOptionsLoading={false}
        areaOptions={[makeArea()]}
        areaOptionsLoading={false}
        userOptions={[makeUser()]}
        userOptionsLoading={false}
        updateStage={vi.fn()}
        removeStage={vi.fn()}
      />,
    );

    expect(screen.getByLabelText('Usuário do seletor 1 da etapa 1')).toBeInTheDocument();
    expect(screen.queryByLabelText('Role do seletor 1 da etapa 1')).toBeNull();
    expect(screen.queryByLabelText('Área do seletor 1 da etapa 1')).toBeNull();
  });
});
