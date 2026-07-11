import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import type { ProcessArea } from '../../../taxonomy/types';
import { StageCard, type StageDraft } from './StageCard';

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

function makeStage(overrides: Partial<StageDraft> = {}): StageDraft {
  return {
    uid: 'stage-1',
    label: 'Revisão',
    requiredRole: 'approver',
    requiredCapability: 'doc.signoff',
    areaCode: 'AREA-01',
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
        updateStage={vi.fn()}
        removeStage={removeStage}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Remover' }));
    expect(removeStage).toHaveBeenCalledWith('stage-1');
  });
});
