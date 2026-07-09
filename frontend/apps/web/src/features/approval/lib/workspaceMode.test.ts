import { describe, it, expect } from 'vitest';

import type { components } from '../../../lib/api-types';
import { deriveStageMode, deriveWorkspaceMode } from './workspaceMode';

type Instance = components['schemas']['ApprovalInstanceByDocumentResponse'];
type Viewer = Instance['viewer'];
type Stage = components['schemas']['ApprovalStageInstanceResponse'];

// --- factories (pure, minimal — only fields the selector reads) -------------

function makeViewer(over: Partial<Viewer> = {}): Viewer {
  return {
    is_author: false,
    eligible_for_active_stage: false,
    has_signed_active_stage: false,
    via_delegation_from: null,
    ...over,
  };
}

function makeStage(over: Partial<Stage> = {}): Stage {
  const base = { stage_kind: 'review', status: 'active' };
  return { ...base, ...over } as Stage;
}

function makeInstance(over: Partial<Instance> = {}): Instance {
  const base = { status: 'in_progress', stages: [], viewer: makeViewer(), verdicts: [] };
  return { ...base, ...over } as Instance;
}

const doc = (status: string) => ({ status });

// --- deriveStageMode: subject-agnostic (stage_kind × eligible only) ----------

describe('deriveStageMode', () => {
  it('maps review + eligible → reviewing', () => {
    expect(deriveStageMode('review', true)).toBe('reviewing');
  });
  it('maps approval + eligible → approving', () => {
    expect(deriveStageMode('approval', true)).toBe('approving');
  });
  it('maps either kind + not eligible → observing', () => {
    expect(deriveStageMode('review', false)).toBe('observing');
    expect(deriveStageMode('approval', false)).toBe('observing');
  });
  it('maps undefined stage kind → observing', () => {
    expect(deriveStageMode(undefined, true)).toBe('observing');
  });
});

// --- deriveWorkspaceMode -----------------------------------------------------

describe('deriveWorkspaceMode', () => {
  it('author-editing — is_author, draft, no instance', () => {
    const viewer = makeViewer({ is_author: true });
    expect(deriveWorkspaceMode(doc('draft'), null, viewer)).toBe('author-editing');
  });

  it('author-editing — is_author, draft, instance present (not changes_requested)', () => {
    const viewer = makeViewer({ is_author: true });
    const instance = makeInstance({ status: 'in_progress', viewer });
    expect(deriveWorkspaceMode(doc('draft'), instance, viewer)).toBe('author-editing');
  });

  it('author-waiting — is_author, under_review, instance in_progress', () => {
    const viewer = makeViewer({ is_author: true });
    const instance = makeInstance({ status: 'in_progress', viewer });
    expect(deriveWorkspaceMode(doc('under_review'), instance, viewer)).toBe('author-waiting');
  });

  it('author-changes-requested — is_author, instance.status changes_requested (doc reverted to draft)', () => {
    const viewer = makeViewer({ is_author: true });
    const instance = makeInstance({ status: 'changes_requested', viewer });
    expect(deriveWorkspaceMode(doc('draft'), instance, viewer)).toBe('author-changes-requested');
  });

  it('reviewing — non-author, active review stage, eligible', () => {
    const viewer = makeViewer({ eligible_for_active_stage: true });
    const instance = makeInstance({
      status: 'in_progress',
      stages: [makeStage({ stage_kind: 'review', status: 'active' })],
      viewer,
    });
    expect(deriveWorkspaceMode(doc('under_review'), instance, viewer)).toBe('reviewing');
  });

  it('approving — non-author, active approval stage, eligible', () => {
    const viewer = makeViewer({ eligible_for_active_stage: true });
    const instance = makeInstance({
      status: 'in_progress',
      stages: [makeStage({ stage_kind: 'approval', status: 'active' })],
      viewer,
    });
    expect(deriveWorkspaceMode(doc('under_review'), instance, viewer)).toBe('approving');
  });

  it('approving via delegation — eligible via via_delegation_from, mode still approving', () => {
    const viewer = makeViewer({
      eligible_for_active_stage: true,
      via_delegation_from: { user_id: 'u-deleg', display_name: 'Delegator' },
    });
    const instance = makeInstance({
      status: 'in_progress',
      stages: [makeStage({ stage_kind: 'approval', status: 'active' })],
      viewer,
    });
    expect(deriveWorkspaceMode(doc('under_review'), instance, viewer)).toBe('approving');
  });

  it('observing — non-author, active stage, NOT eligible', () => {
    const viewer = makeViewer({ eligible_for_active_stage: false });
    const instance = makeInstance({
      status: 'in_progress',
      stages: [makeStage({ stage_kind: 'approval', status: 'active' })],
      viewer,
    });
    expect(deriveWorkspaceMode(doc('under_review'), instance, viewer)).toBe('observing');
  });

  it('observing — non-author, draft / no active stage', () => {
    const viewer = makeViewer({});
    expect(deriveWorkspaceMode(doc('draft'), null, viewer)).toBe('observing');
  });

  it('lifecycle — approved terminal document status', () => {
    const viewer = makeViewer({});
    const instance = makeInstance({ status: 'approved', stages: [], viewer });
    expect(deriveWorkspaceMode(doc('approved'), instance, viewer)).toBe('lifecycle');
  });

  it('lifecycle — published terminal document status', () => {
    const viewer = makeViewer({ is_author: true });
    const instance = makeInstance({ status: 'approved', stages: [], viewer });
    expect(deriveWorkspaceMode(doc('published'), instance, viewer)).toBe('lifecycle');
  });

  it('author dominates over stage — SoD-excluded author never approving', () => {
    // Pathological: is_author true AND somehow flagged eligible on an approval stage.
    // SoD makes this impossible server-side, but the selector must still favour the
    // author lens (never render the signature panel to the author).
    const viewer = makeViewer({ is_author: true, eligible_for_active_stage: true });
    const instance = makeInstance({
      status: 'in_progress',
      stages: [makeStage({ stage_kind: 'approval', status: 'active' })],
      viewer,
    });
    expect(deriveWorkspaceMode(doc('under_review'), instance, viewer)).toBe('author-waiting');
  });

  it('null viewer — falls back to observing (non-author lens)', () => {
    const instance = makeInstance({ status: 'in_progress', stages: [], viewer: makeViewer() });
    expect(deriveWorkspaceMode(doc('under_review'), instance, null)).toBe('observing');
  });
});
