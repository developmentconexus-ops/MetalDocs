import { describe, expect, it } from 'vitest';

import { defaultSelector, type SelectorDraft, type StageDraft } from './StageCard';
import { defaultStage, toStageRequests, validateDraft, type RouteDraft } from './routeDraft';

function makeSelector(overrides: Partial<SelectorDraft> = {}): SelectorDraft {
  return { ...defaultSelector(), ...overrides };
}

function makeStage(overrides: Partial<StageDraft> = {}): StageDraft {
  return { ...defaultStage(), label: 'Etapa', ...overrides };
}

function makeDraft(overrides: Partial<RouteDraft> = {}): RouteDraft {
  return {
    name: 'Rota',
    profileCode: 'JUR',
    stages: [makeStage()],
    ...overrides,
  };
}

describe('routeDraft selectors', () => {
  it('validateDraft rejects a stage with an empty-role role_in_fixed_area selector', () => {
    const draft = makeDraft({
      stages: [
        makeStage({
          label: 'Revisão',
          selectors: [makeSelector({ kind: 'role_in_fixed_area', role: '', areaCode: '' })],
        }),
      ],
    });

    expect(validateDraft(draft, false, null)).toBe('Na etapa "Revisão", selecione uma role.');
  });

  it('validateDraft rejects a stage with an empty area for role_in_fixed_area', () => {
    const draft = makeDraft({
      stages: [
        makeStage({
          label: 'Revisão',
          selectors: [makeSelector({ kind: 'role_in_fixed_area', role: 'approver', areaCode: '' })],
        }),
      ],
    });

    expect(validateDraft(draft, false, null)).toBe('Na etapa "Revisão", selecione uma área.');
  });

  it('validateDraft rejects a named_user selector with no user selected', () => {
    const draft = makeDraft({
      stages: [
        makeStage({
          label: 'Revisão',
          selectors: [makeSelector({ kind: 'named_user', userId: '' })],
        }),
      ],
    });

    expect(validateDraft(draft, false, null)).toBe('Na etapa "Revisão", selecione um usuário.');
  });

  it('validateDraft rejects a stage with zero selectors', () => {
    const draft = makeDraft({
      stages: [makeStage({ label: 'Revisão', selectors: [] })],
    });

    expect(validateDraft(draft, false, null)).toBe(
      'A etapa "Revisão" deve ter ao menos um seletor.',
    );
  });

  it('validateDraft accepts a valid role_in_document_area selector (no area required)', () => {
    const draft = makeDraft({
      stages: [
        makeStage({
          label: 'Revisão',
          selectors: [makeSelector({ kind: 'role_in_document_area', role: 'approver', areaCode: '' })],
        }),
      ],
    });

    expect(validateDraft(draft, false, null)).toBeNull();
  });

  it('toStageRequests emits per-kind fields only for named_user', () => {
    const draft = makeDraft({
      stages: [
        makeStage({
          label: 'Revisão',
          selectors: [makeSelector({ kind: 'named_user', userId: '  user-1  ', role: 'stale', areaCode: 'stale' })],
        }),
      ],
    });

    const [stage] = toStageRequests(draft);
    expect(stage.selectors).toEqual([{ kind: 'named_user', user_id: 'user-1' }]);
  });

  it('toStageRequests emits role+area (trimmed, lowercased) for role_in_fixed_area/submit_choice', () => {
    const draft = makeDraft({
      stages: [
        makeStage({
          label: 'Etapa 1',
          selectors: [
            makeSelector({ kind: 'role_in_fixed_area', role: ' Approver ', areaCode: ' AREA-01 ' }),
          ],
        }),
        makeStage({
          label: 'Etapa 2',
          selectors: [
            makeSelector({ kind: 'submit_choice', role: ' Editor ', areaCode: ' AREA-02 ' }),
          ],
        }),
      ],
    });

    const [stage1, stage2] = toStageRequests(draft);
    expect(stage1.selectors).toEqual([
      { kind: 'role_in_fixed_area', role: 'approver', area_code: 'area-01' },
    ]);
    expect(stage2.selectors).toEqual([
      { kind: 'submit_choice', role: 'editor', area_code: 'area-02' },
    ]);
  });

  it('toStageRequests emits role only for role_in_document_area', () => {
    const draft = makeDraft({
      stages: [
        makeStage({
          label: 'Revisão',
          selectors: [
            makeSelector({ kind: 'role_in_document_area', role: ' Approver ', areaCode: 'ignored' }),
          ],
        }),
      ],
    });

    const [stage] = toStageRequests(draft);
    expect(stage.selectors).toEqual([{ kind: 'role_in_document_area', role: 'approver' }]);
  });
});
