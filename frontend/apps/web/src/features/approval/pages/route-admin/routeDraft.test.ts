import { describe, expect, it } from 'vitest';

import type { RouteSummary } from '../../api/routeAdminApi';
import { defaultSelector, type SelectorDraft, type StageDraft } from './StageCard';
import {
  defaultStage,
  isUntouchedDefaultStages,
  toCreateRequest,
  toDraft,
  toStageRequests,
  validateDraft,
  type RouteDraft,
} from './routeDraft';

function makeSelector(overrides: Partial<SelectorDraft> = {}): SelectorDraft {
  return { ...defaultSelector(), ...overrides };
}

function makeStage(overrides: Partial<StageDraft> = {}): StageDraft {
  return { ...defaultStage(), label: 'Etapa', ...overrides };
}

function makeDraft(overrides: Partial<RouteDraft> = {}): RouteDraft {
  return {
    name: 'Rota',
    subjectKind: 'document',
    profileCode: 'JUR',
    stages: [makeStage()],
    ...overrides,
  };
}

function makeRouteSummary(overrides: Partial<RouteSummary> = {}): RouteSummary {
  return {
    id: 'route-1',
    name: 'Rota Modelo',
    tenant_id: 'tenant-1',
    profile_code: 'JUR',
    active: true,
    version: 1,
    stages: [],
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

describe('toDraft profile_code null-safety (F18 completion S6)', () => {
  it('normalizes a malformed null profile_code to an empty string', () => {
    // ADR 0086 made profile_code contract-required and non-nullable, so this
    // input is no longer type-expressible — the cast is deliberate. The guard
    // it pins is still live: the Go wire struct holds a *string, so a NULL
    // column on a malformed row would reach the client as JSON null.
    const draft = toDraft(
      makeRouteSummary({ name: 'Rota Modelo', profile_code: null as unknown as string }),
    );
    expect(draft.profileCode).toBe('');
  });

  it('carries a document route profile_code through unchanged', () => {
    const draft = toDraft(makeRouteSummary({ profile_code: 'JUR' }));
    expect(draft.profileCode).toBe('JUR');
  });
});

describe('toDraft subject kind (ADR 0086)', () => {
  it('maps an explicit template subject_kind onto the draft', () => {
    const draft = toDraft(makeRouteSummary({ subject_kind: 'template', profile_code: 'JUR' }));
    expect(draft.subjectKind).toBe('template');
  });

  it('maps an explicit document subject_kind onto the draft', () => {
    const draft = toDraft(makeRouteSummary({ subject_kind: 'document', profile_code: 'JUR' }));
    expect(draft.subjectKind).toBe('document');
  });

  it('defaults an absent subject_kind to document (legacy rows predate the field)', () => {
    const draft = toDraft(makeRouteSummary({ profile_code: 'JUR' }));
    expect(draft.subjectKind).toBe('document');
  });

  it('defaults a brand-new draft to document', () => {
    expect(toDraft(null).subjectKind).toBe('document');
  });
});

describe('toCreateRequest (ADR 0086 — routes are keyed by configuration)', () => {
  it('sends subject_kind + trimmed profile_code and omits subject_key for a document route', () => {
    const payload = toCreateRequest(
      makeDraft({ name: '  Rota Doc  ', subjectKind: 'document', profileCode: '  JUR  ' }),
    );

    expect(payload.name).toBe('Rota Doc');
    expect(payload.subject_kind).toBe('document');
    expect(payload.profile_code).toBe('JUR');
    // The server derives subject_key from profile_code; sending a divergent
    // key is a 422 (validation.template_subject_key_mismatch).
    expect('subject_key' in payload).toBe(false);
  });

  it('sends profile_code for a TEMPLATE route too, and still omits subject_key', () => {
    const payload = toCreateRequest(
      makeDraft({ name: 'Rota Tpl', subjectKind: 'template', profileCode: 'JUR' }),
    );

    expect(payload.subject_kind).toBe('template');
    expect(payload.profile_code).toBe('JUR');
    expect('subject_key' in payload).toBe(false);
  });

  it('carries the draft stages through toStageRequests', () => {
    const draft = makeDraft({ subjectKind: 'template', stages: [makeStage({ label: 'Revisão' })] });
    expect(toCreateRequest(draft).stages).toEqual(toStageRequests(draft));
  });
});

describe('validateDraft profile requirement (ADR 0086)', () => {
  it('requires a profile code on create for a document route', () => {
    expect(validateDraft(makeDraft({ subjectKind: 'document', profileCode: '' }), false, null)).toBe(
      'Informe o código do perfil.',
    );
  });

  it('requires a profile code on create for a TEMPLATE route as well', () => {
    expect(validateDraft(makeDraft({ subjectKind: 'template', profileCode: '' }), false, null)).toBe(
      'Informe o código do perfil.',
    );
  });

  it('accepts a template route that carries a profile code', () => {
    const draft = makeDraft({
      subjectKind: 'template',
      profileCode: 'JUR',
      stages: [
        makeStage({
          label: 'Aprovação',
          selectors: [makeSelector({ kind: 'role_in_document_area', role: 'approver' })],
        }),
      ],
    });

    expect(validateDraft(draft, false, null)).toBeNull();
  });
});

describe('validateDraft stage-count floor (ADR 0087)', () => {
  it('rejects zero stages for a non-livre governance class', () => {
    const draft = makeDraft({ stages: [] });
    expect(validateDraft(draft, false, 'controlado')).toBe('A rota deve possuir ao menos uma etapa.');
  });

  it('rejects zero stages when governance class is unknown (archived profile)', () => {
    const draft = makeDraft({ stages: [] });
    expect(validateDraft(draft, true, null)).toBe('A rota deve possuir ao menos uma etapa.');
  });

  it('accepts zero stages for livre — auto-approval, nothing to route', () => {
    const draft = makeDraft({ stages: [] });
    expect(validateDraft(draft, false, 'livre')).toBeNull();
  });

  it('rejects a livre draft that still carries a stage', () => {
    const draft = makeDraft({
      stages: [
        makeStage({
          label: 'Revisão',
          selectors: [makeSelector({ kind: 'role_in_document_area', role: 'approver' })],
        }),
      ],
    });
    expect(validateDraft(draft, false, 'livre')).toBe(
      'Perfil livre não admite etapas — a rota livre é aprovada automaticamente.',
    );
  });
});

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

describe('routeDraft stage deadline (due_in_days)', () => {
  function makeStageSummary(
    overrides: Partial<RouteSummary['stages'][number]> = {},
  ): RouteSummary['stages'][number] {
    return {
      order: 1,
      name: 'Etapa',
      required_capability: 'approval.signoff',
      quorum: 'any_1_of',
      quorum_m: null,
      drift_policy: 'reduce_quorum',
      stage_kind: 'approval',
      due_in_days: null,
      selectors: [],
      ...overrides,
    };
  }

  it('carries due_in_days from the wire into the draft and back', () => {
    const draft = toDraft(
      makeRouteSummary({
        stages: [
          makeStageSummary({ due_in_days: 30 }),
          makeStageSummary({ order: 2, due_in_days: null }),
        ],
      }),
    );

    expect(draft.stages[0].dueInDays).toBe('30');
    expect(draft.stages[1].dueInDays).toBe('');

    const wire = toStageRequests(draft);
    expect(wire[0].due_in_days).toBe(30);
    // Empty means NO deadline: the field must be ABSENT, not 0 and not null.
    expect('due_in_days' in wire[1]).toBe(false);
  });
});

describe('isUntouchedDefaultStages (ADR 0087 livre auto-drop)', () => {
  it('treats the pristine seed stage as untouched', () => {
    expect(isUntouchedDefaultStages([defaultStage()])).toBe(true);
  });

  it('does NOT treat a seed stage with only a deadline filled in as untouched', () => {
    // If dueInDays weren't compared here, this stage would be misclassified as
    // "untouched" and silently dropped the moment the user switches to a livre
    // profile (ADR 0087) — deleting the deadline the user just typed in without
    // any warning. That is exactly the data loss this function exists to prevent.
    const stage = { ...defaultStage(), dueInDays: '30' };
    expect(isUntouchedDefaultStages([stage])).toBe(false);
  });
});

describe('validateDraft stage deadline lower bound', () => {
  function makeValidStage(overrides: Partial<StageDraft> = {}): StageDraft {
    return makeStage({
      selectors: [makeSelector({ kind: 'role_in_document_area', role: 'approver', areaCode: '' })],
      ...overrides,
    });
  }

  it('rejects a stage due_in_days of 0 (mirrors the server floor)', () => {
    const draft = makeDraft({ stages: [makeValidStage({ dueInDays: '0' })] });
    expect(validateDraft(draft, false, null)).toBe(
      'Na etapa "Etapa", o prazo deve ser de pelo menos 1 dia.',
    );
  });

  it('rejects a negative stage due_in_days', () => {
    const draft = makeDraft({ stages: [makeValidStage({ dueInDays: '-5' })] });
    expect(validateDraft(draft, false, null)).toBe(
      'Na etapa "Etapa", o prazo deve ser de pelo menos 1 dia.',
    );
  });

  it('rejects a non-integer stage due_in_days (server field is *int, decode 400 otherwise)', () => {
    const draft = makeDraft({ stages: [makeValidStage({ dueInDays: '1.5' })] });
    expect(validateDraft(draft, false, null)).toBe(
      'Na etapa "Etapa", o prazo deve ser de pelo menos 1 dia.',
    );
  });

  it('accepts a valid positive due_in_days', () => {
    const draft = makeDraft({ stages: [makeValidStage({ dueInDays: '30' })] });
    expect(validateDraft(draft, false, null)).toBeNull();
  });

  it('accepts an empty due_in_days (no deadline)', () => {
    const draft = makeDraft({ stages: [makeValidStage({ dueInDays: '' })] });
    expect(validateDraft(draft, false, null)).toBeNull();
  });
});
