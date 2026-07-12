import { describe, expect, it } from 'vitest';

import type { ApprovalInstance, StageActor, StageInstance } from '../../api/approvalTypes';
import { deriveFastForwardOffer } from '../fastForwardOffer';

const CURRENT_USER = 'user-1';

function makeActor(overrides: Partial<StageActor> = {}): StageActor {
  return {
    user_id: CURRENT_USER,
    display_name: 'Ana Revisora',
    status: 'waiting',
    ...overrides,
  } as StageActor;
}

function makeStage(overrides: Partial<StageInstance> = {}): StageInstance {
  return {
    id: 'stage-review',
    stage_index: 0,
    label: 'Revisão técnica',
    status: 'active',
    signoffs: [],
    actors: [],
    stage_kind: 'review',
    due_at: null,
    ...overrides,
  } as StageInstance;
}

function makeInstance(overrides: Partial<ApprovalInstance> = {}): ApprovalInstance {
  return {
    id: 'inst-1',
    document_id: 'doc-1',
    route_id: 'route-1',
    tenant_id: 'tenant-1',
    status: 'in_progress',
    submitted_by: 'maria',
    submitted_at: '2026-04-15T10:00:00.000Z',
    completed_at: null,
    stages: [
      makeStage(),
      makeStage({
        id: 'stage-approval',
        stage_index: 1,
        status: 'pending',
        stage_kind: 'approval',
        actors: [makeActor()],
      }),
    ],
    etag: '"v3"',
    frozen_content_hash: null,
    viewer: {
      is_author: false,
      eligible_for_active_stage: true,
      has_signed_active_stage: false,
      via_delegation_from: null,
    },
    ...overrides,
  } as ApprovalInstance;
}

describe('deriveFastForwardOffer', () => {
  it('returns the offer when every condition holds', () => {
    const offer = deriveFastForwardOffer(makeInstance(), CURRENT_USER);
    expect(offer).toEqual({ reviewStageId: 'stage-review', nextApprovalStageId: 'stage-approval' });
  });

  it('returns null when the instance is not in_progress', () => {
    const offer = deriveFastForwardOffer(makeInstance({ status: 'approved' }), CURRENT_USER);
    expect(offer).toBeNull();
  });

  it('returns null when the active stage is not review-kind', () => {
    const instance = makeInstance({
      stages: [
        makeStage({ stage_kind: 'approval' }),
        makeStage({ id: 'stage-approval', stage_index: 1, status: 'pending', stage_kind: 'approval', actors: [makeActor()] }),
      ],
    });
    expect(deriveFastForwardOffer(instance, CURRENT_USER)).toBeNull();
  });

  it('returns null when the viewer is not eligible for the active stage', () => {
    const instance = makeInstance({
      viewer: { is_author: false, eligible_for_active_stage: false, has_signed_active_stage: false, via_delegation_from: null },
    });
    expect(deriveFastForwardOffer(instance, CURRENT_USER)).toBeNull();
  });

  it('returns null when the viewer already signed the active stage', () => {
    const instance = makeInstance({
      viewer: { is_author: false, eligible_for_active_stage: true, has_signed_active_stage: true, via_delegation_from: null },
    });
    expect(deriveFastForwardOffer(instance, CURRENT_USER)).toBeNull();
  });

  it('returns null when there is no next stage', () => {
    const instance = makeInstance({ stages: [makeStage()] });
    expect(deriveFastForwardOffer(instance, CURRENT_USER)).toBeNull();
  });

  it('returns null when the next stage is not approval-kind', () => {
    const instance = makeInstance({
      stages: [
        makeStage(),
        makeStage({ id: 'stage-2', stage_index: 1, status: 'pending', stage_kind: 'review', actors: [makeActor()] }),
      ],
    });
    expect(deriveFastForwardOffer(instance, CURRENT_USER)).toBeNull();
  });

  it('returns null when the viewer is not among the next stage actors', () => {
    const instance = makeInstance({
      stages: [
        makeStage(),
        makeStage({
          id: 'stage-approval',
          stage_index: 1,
          status: 'pending',
          stage_kind: 'approval',
          actors: [makeActor({ user_id: 'someone-else' })],
        }),
      ],
    });
    expect(deriveFastForwardOffer(instance, CURRENT_USER)).toBeNull();
  });

  it('returns null when currentUserId is empty', () => {
    expect(deriveFastForwardOffer(makeInstance(), '')).toBeNull();
  });
});
