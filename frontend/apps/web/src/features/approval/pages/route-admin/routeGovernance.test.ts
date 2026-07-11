import { describe, expect, it } from 'vitest';
import {
  approveNowOverlapNote,
  authorExcludedNote,
  describeStageKind,
  flowPreviewEmptyLabel,
  labelForStageKind,
  livreBlockedMessage,
  routePolicyFor,
  STAGE_KIND_OPTIONS,
  stageActorSlotDefaultHeading,
  stageOrderViolationMessage,
  validateSignaturePolicy,
  validateStageOrder,
  type GovernanceClass,
  type StageKind,
} from './routeGovernance';

describe('routePolicyFor', () => {
  it('controlado requires a signature and allows the route', () => {
    expect(routePolicyFor('controlado')).toEqual({
      routeAllowed: true,
      signatureRequired: true,
      badgeLabel: 'Obrigatório ≥1 assinatura',
      badgeTone: 'required',
    });
  });

  it('simples allows the route without requiring a signature', () => {
    expect(routePolicyFor('simples')).toEqual({
      routeAllowed: true,
      signatureRequired: false,
      badgeLabel: 'Assinatura opcional',
      badgeTone: 'optional',
    });
  });

  it('livre blocks the route entirely', () => {
    expect(routePolicyFor('livre')).toEqual({
      routeAllowed: false,
      signatureRequired: false,
      badgeLabel: 'Perfil livre — sem rota de aprovação',
      badgeTone: 'blocked',
    });
  });
});

describe('labelForStageKind', () => {
  it('labels review', () => {
    expect(labelForStageKind('review')).toBe('Revisão');
  });

  it('labels approval', () => {
    expect(labelForStageKind('approval')).toBe('Assinatura');
  });
});

describe('describeStageKind', () => {
  it('describes review', () => {
    expect(describeStageKind('review')).toBe(
      'Revisores comentam e conferem; não assinam.',
    );
  });

  it('describes approval', () => {
    expect(describeStageKind('approval')).toBe(
      'Aprovadores assinam (e podem conversar).',
    );
  });
});

describe('STAGE_KIND_OPTIONS', () => {
  it('lists review then approval', () => {
    expect(STAGE_KIND_OPTIONS).toEqual(['review', 'approval']);
  });
});

describe('flowPreviewEmptyLabel', () => {
  it('returns the empty-state copy', () => {
    expect(flowPreviewEmptyLabel()).toBe('Adicione etapas para ver o fluxo.');
  });
});

describe('stageActorSlotDefaultHeading', () => {
  it('returns the default heading copy', () => {
    expect(stageActorSlotDefaultHeading()).toBe('Quem atua nesta etapa');
  });
});

describe('livreBlockedMessage', () => {
  it('returns a non-empty PT-BR message', () => {
    const message = livreBlockedMessage();
    expect(typeof message).toBe('string');
    expect(message.length).toBeGreaterThan(0);
  });
});

describe('validateSignaturePolicy', () => {
  const cases: Array<{
    name: string;
    gc: GovernanceClass;
    stageKinds: StageKind[];
    expected: string | null;
  }> = [
    {
      name: 'controlado with review-only stages errors',
      gc: 'controlado',
      stageKinds: ['review'],
      expected: 'Perfil controlado exige ao menos uma etapa de assinatura (aprovação).',
    },
    {
      name: 'controlado with a review + approval stage is valid',
      gc: 'controlado',
      stageKinds: ['review', 'approval'],
      expected: null,
    },
    {
      name: 'simples with review-only stages is valid',
      gc: 'simples',
      stageKinds: ['review'],
      expected: null,
    },
    {
      name: 'livre is always blocked regardless of stage kinds',
      gc: 'livre',
      stageKinds: ['review', 'approval'],
      expected: livreBlockedMessage(),
    },
  ];

  it.each(cases)('$name', ({ gc, stageKinds, expected }) => {
    expect(validateSignaturePolicy(gc, stageKinds)).toBe(expected);
  });
});

describe('validateStageOrder', () => {
  it('allows review-only routes', () => {
    expect(validateStageOrder(['review', 'review'])).toBeNull();
  });

  it('allows approval-only routes', () => {
    expect(validateStageOrder(['approval', 'approval'])).toBeNull();
  });

  it('allows review followed by approval', () => {
    expect(validateStageOrder(['review', 'review', 'approval'])).toBeNull();
  });

  it('rejects a review stage after an approval stage', () => {
    expect(validateStageOrder(['approval', 'review'])).toBe(stageOrderViolationMessage());
  });
});

describe('authorExcludedNote', () => {
  it('returns a non-empty PT-BR message', () => {
    const message = authorExcludedNote();
    expect(typeof message).toBe('string');
    expect(message.length).toBeGreaterThan(0);
  });
});

describe('approveNowOverlapNote', () => {
  it('returns a non-empty PT-BR message', () => {
    const message = approveNowOverlapNote();
    expect(typeof message).toBe('string');
    expect(message.length).toBeGreaterThan(0);
  });
});
