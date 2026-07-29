import { describe, it, expect } from 'vitest';
import { mapApprovalChain } from './approvalWorkflow';

interface TestSignoff {
  actor_user_id: string;
  decision: string;
  signed_at: string;
}

interface TestActor {
  user_id: string;
  display_name: string;
  status: string;
  decision?: string | null;
}

interface TestStage {
  stage_index: number;
  label: string;
  status: string;
  signoffs: TestSignoff[];
  actors?: TestActor[];
}

function makeStage(overrides: Partial<TestStage> = {}): TestStage {
  return {
    stage_index: 0,
    label: 'Revisão',
    status: 'active',
    signoffs: [],
    actors: [],
    ...overrides,
  };
}

describe('mapApprovalChain — actor roster (F-QA4-8)', () => {
  it('emits one item per eligible actor with the backend display name before any decision', () => {
    const stages: TestStage[] = [
      makeStage({
        stage_index: 0,
        label: 'Aprovação da qualidade',
        status: 'active',
        signoffs: [],
        actors: [
          { user_id: 'user-ana', display_name: 'Ana Souza', status: 'active', decision: null },
          { user_id: 'user-bruno', display_name: 'Bruno Lima', status: 'active', decision: null },
        ],
      }),
    ];

    const result = mapApprovalChain(stages);

    expect(result).toHaveLength(2);
    expect(result[0]).toEqual({
      stageIndex: 0,
      label: 'Aprovação da qualidade',
      status: 'active',
      roleLabel: 'Aprovação da qualidade',
      flowState: 'current',
      actorUserId: 'user-ana',
      actorDisplay: 'Ana Souza',
      decision: null,
      signedAt: null,
    });
    expect(result[1]).toMatchObject({ actorUserId: 'user-bruno', actorDisplay: 'Bruno Lima' });
    // The whole point of the defect: no chain row may fall back to the raw id.
    expect(result.map((item) => item.actorDisplay)).not.toContain('user-ana');
  });

  it('carries a decided actor display name, decision and signed_at', () => {
    const stages: TestStage[] = [
      makeStage({
        stage_index: 0,
        label: 'Revisão técnica',
        status: 'passed',
        signoffs: [
          {
            actor_user_id: 'user-ana',
            decision: 'approve',
            signed_at: '2026-06-01T10:00:00Z',
          },
        ],
        actors: [
          { user_id: 'user-ana', display_name: 'Ana Souza', status: 'approved', decision: 'approve' },
        ],
      }),
    ];

    const result = mapApprovalChain(stages);

    expect(result).toHaveLength(1);
    expect(result[0]).toEqual({
      stageIndex: 0,
      label: 'Revisão técnica',
      status: 'approved',
      roleLabel: 'Revisão técnica',
      flowState: 'approved',
      actorUserId: 'user-ana',
      actorDisplay: 'Ana Souza',
      decision: 'approve',
      signedAt: '2026-06-01T10:00:00Z',
    });
  });

  it('maps a rejected actor and a still-waiting future-stage actor', () => {
    const stages: TestStage[] = [
      makeStage({
        stage_index: 0,
        label: 'Revisão',
        status: 'failed',
        signoffs: [
          { actor_user_id: 'user-ana', decision: 'reject', signed_at: '2026-06-01T11:00:00Z' },
        ],
        actors: [
          { user_id: 'user-ana', display_name: 'Ana Souza', status: 'rejected', decision: 'reject' },
        ],
      }),
      makeStage({
        stage_index: 1,
        label: 'Aprovação',
        status: 'pending',
        signoffs: [],
        actors: [
          { user_id: 'user-carla', display_name: 'Carla Dias', status: 'waiting', decision: null },
        ],
      }),
    ];

    const result = mapApprovalChain(stages);

    expect(result).toHaveLength(2);
    expect(result[0]).toMatchObject({
      status: 'rejected',
      flowState: 'rejected',
      actorDisplay: 'Ana Souza',
      decision: 'reject',
      signedAt: '2026-06-01T11:00:00Z',
    });
    expect(result[1]).toMatchObject({
      stageIndex: 1,
      status: 'waiting',
      flowState: 'pending',
      actorDisplay: 'Carla Dias',
      decision: null,
      signedAt: null,
    });
  });

  it('keeps template-subject / roster-less stages on the honest stage-level fallback', () => {
    // A stage the backend skipped never gets an actor roster: one row, all nulls
    // — genuine absence, not a lost display-name lookup.
    const stages: TestStage[] = [
      makeStage({ stage_index: 2, label: 'Aprovação', status: 'cancelled', signoffs: [], actors: [] }),
    ];

    const result = mapApprovalChain(stages);

    expect(result).toHaveLength(1);
    expect(result[0]).toMatchObject({
      stageIndex: 2,
      status: 'cancelled',
      flowState: 'pending',
      actorUserId: null,
      actorDisplay: null,
    });
  });
});

// Roster-less stages (actors: []) keep the pre-F-QA4-8 stage-level mapping.
describe('mapApprovalChain — stage-level fallback', () => {
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
