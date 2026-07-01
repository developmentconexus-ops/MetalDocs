import { describe, it, expect } from 'vitest';
import { mapApprovalChain } from './approvalWorkflow';

interface TestSignoff {
  actor_user_id: string;
  decision: string;
  signed_at: string;
}

interface TestStage {
  stage_index: number;
  label: string;
  status: string;
  signoffs: TestSignoff[];
}

function makeStage(overrides: Partial<TestStage> = {}): TestStage {
  return {
    stage_index: 0,
    label: 'Revisão',
    status: 'active',
    signoffs: [],
    ...overrides,
  };
}

describe('mapApprovalChain', () => {
  it('maps a stage with a signoff to an ApprovalChainItem', () => {
    const stages: TestStage[] = [
      makeStage({
        stage_index: 0,
        label: 'Revisão',
        status: 'passed',
        signoffs: [
          {
            actor_user_id: 'user-abc',
            decision: 'approve',
            signed_at: '2026-06-01T10:00:00Z',
          },
        ],
      }),
    ];

    const result = mapApprovalChain(stages);

    expect(result).toHaveLength(1);
    expect(result[0]).toEqual({
      stageIndex: 0,
      label: 'Revisão',
      status: 'passed',
      roleLabel: 'Revisão',
      flowState: 'approved',
      actorUserId: 'user-abc',
      actorDisplay: 'user-abc',
      decision: 'approve',
      signedAt: '2026-06-01T10:00:00Z',
    });
  });

  it('maps a stage with empty signoffs to nulls', () => {
    const stages: TestStage[] = [
      makeStage({
        stage_index: 1,
        label: 'Aprovação',
        status: 'pending',
        signoffs: [],
      }),
    ];

    const result = mapApprovalChain(stages);

    expect(result).toHaveLength(1);
    expect(result[0]).toEqual({
      stageIndex: 1,
      label: 'Aprovação',
      status: 'pending',
      roleLabel: 'Aprovação',
      flowState: 'pending',
      actorUserId: null,
      actorDisplay: null,
      decision: null,
      signedAt: null,
    });
  });

  it('preserves input order across multiple stages', () => {
    const stages: TestStage[] = [
      makeStage({ stage_index: 0, label: 'Etapa A', status: 'passed' }),
      makeStage({ stage_index: 1, label: 'Etapa B', status: 'active' }),
      makeStage({ stage_index: 2, label: 'Etapa C', status: 'pending' }),
    ];

    const result = mapApprovalChain(stages);

    expect(result).toHaveLength(3);
    expect(result[0].stageIndex).toBe(0);
    expect(result[0].label).toBe('Etapa A');
    expect(result[1].stageIndex).toBe(1);
    expect(result[1].label).toBe('Etapa B');
    expect(result[2].stageIndex).toBe(2);
    expect(result[2].label).toBe('Etapa C');
  });
});
