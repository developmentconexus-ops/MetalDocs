import { describe, expect, it } from 'vitest';
import {
  INITIAL_STATE,
  canAdvance,
  clampStep,
  maxReachableStep,
  wizardReducer,
  type WizardState,
} from '../wizard.reducer';

const baseFilled: WizardState = {
  ...INITIAL_STATE,
  step: 4,
  profileCode: 'POP',
  areaCode: 'MFG',
  title: 'Procedimento X',
  templateID: 't1',
  templateVersionID: 'tv1',
};

describe('wizardReducer', () => {
  describe('selectProfile', () => {
    it('rejects empty/whitespace code', () => {
      const next = wizardReducer(INITIAL_STATE, { type: 'selectProfile', code: '   ' });
      expect(next).toBe(INITIAL_STATE);
    });

    it('clears template selection when profile changes', () => {
      const next = wizardReducer(baseFilled, { type: 'selectProfile', code: 'IT' });
      expect(next.profileCode).toBe('IT');
      expect(next.templateID).toBeNull();
      expect(next.templateVersionID).toBeNull();
    });

    it('clamps step back to 2 when current step depends on template', () => {
      const next = wizardReducer(baseFilled, { type: 'selectProfile', code: 'IT' });
      expect(next.step).toBe(2);
    });

    it('preserves step ≤ 2 unchanged', () => {
      const onStep1 = wizardReducer(
        { ...INITIAL_STATE, step: 1 },
        { type: 'selectProfile', code: 'POP' },
      );
      expect(onStep1.step).toBe(1);
      const onStep2 = wizardReducer(
        { ...INITIAL_STATE, step: 2, profileCode: 'OLD' },
        { type: 'selectProfile', code: 'POP' },
      );
      expect(onStep2.step).toBe(2);
    });
  });

  describe('clearProfile', () => {
    it('resets to step 1 and clears profile + template', () => {
      const next = wizardReducer(baseFilled, { type: 'clearProfile' });
      expect(next.step).toBe(1);
      expect(next.profileCode).toBeNull();
      expect(next.templateID).toBeNull();
      expect(next.templateVersionID).toBeNull();
    });
  });

  describe('addInvitee', () => {
    it('rejects duplicate id', () => {
      const seeded: WizardState = {
        ...INITIAL_STATE,
        invitees: [{ id: 'u1', label: 'User 1' }],
      };
      const next = wizardReducer(seeded, {
        type: 'addInvitee',
        invitee: { id: 'u1', label: 'Same' },
      });
      expect(next).toBe(seeded);
    });

    it('appends new id', () => {
      const next = wizardReducer(INITIAL_STATE, {
        type: 'addInvitee',
        invitee: { id: 'u1', label: 'User 1' },
      });
      expect(next.invitees).toHaveLength(1);
    });
  });

  describe('removeInvitee', () => {
    it('drops the matching id only', () => {
      const seeded: WizardState = {
        ...INITIAL_STATE,
        invitees: [
          { id: 'u1', label: 'A' },
          { id: 'u2', label: 'B' },
        ],
      };
      const next = wizardReducer(seeded, { type: 'removeInvitee', id: 'u1' });
      expect(next.invitees).toEqual([{ id: 'u2', label: 'B' }]);
    });
  });

  describe('setExternal', () => {
    it('merges patch into config', () => {
      const next = wizardReducer(INITIAL_STATE, {
        type: 'setExternal',
        patch: { passwordRequired: true },
      });
      expect(next.external.passwordRequired).toBe(true);
      expect(next.external.watermark).toBe(false);
    });
  });

  describe('area visibility defaults', () => {
    it('anchors visibilityAreaCodes to selected area when visibility=area', () => {
      const withAreaVisibility = wizardReducer(
        { ...INITIAL_STATE, visibility: 'area' },
        { type: 'setArea', code: 'QA' },
      );
      expect(withAreaVisibility.visibilityAreaCodes).toEqual(['QA']);
    });

    it('sets current area as restricted area default when switching to area visibility', () => {
      const next = wizardReducer(
        { ...INITIAL_STATE, areaCode: 'MFG', visibility: 'company' },
        { type: 'setVisibility', key: 'area' },
      );
      expect(next.visibilityAreaCodes).toEqual(['MFG']);
    });

    it('keeps selected document area when setting additional visibility areas', () => {
      const seeded = {
        ...INITIAL_STATE,
        areaCode: 'QA',
        visibility: 'area' as const,
        visibilityAreaCodes: ['QA'],
      };
      const next = wizardReducer(seeded, {
        type: 'setVisibilityAreas',
        codes: ['RH'],
      });
      expect(next.visibilityAreaCodes).toEqual(['QA', 'RH']);
    });

    it('clears restricted grants when switching to company visibility', () => {
      const seeded = {
        ...INITIAL_STATE,
        areaCode: 'QA',
        visibility: 'area' as const,
        visibilityAreaCodes: ['QA', 'RH'],
        invitees: [{ id: 'user-1', label: 'User 1' }],
      };
      const next = wizardReducer(seeded, { type: 'setVisibility', key: 'company' });
      expect(next.visibilityAreaCodes).toEqual([]);
      expect(next.invitees).toEqual([]);
    });
  });

  describe('submit lifecycle', () => {
    it('submitStart sets submitting + clears error', () => {
      const next = wizardReducer(
        { ...INITIAL_STATE, error: 'old' },
        { type: 'submitStart' },
      );
      expect(next.submitting).toBe(true);
      expect(next.error).toBeNull();
    });
    it('submitError clears submitting + sets error', () => {
      const next = wizardReducer(
        { ...INITIAL_STATE, submitting: true },
        { type: 'submitError', message: 'boom' },
      );
      expect(next.submitting).toBe(false);
      expect(next.error).toBe('boom');
    });
  });
});

describe('maxReachableStep', () => {
  it('returns 1 without profile', () => {
    expect(maxReachableStep(INITIAL_STATE)).toBe(1);
  });
  it('returns 2 with profile but no area+title', () => {
    expect(maxReachableStep({ ...INITIAL_STATE, profileCode: 'POP' })).toBe(2);
  });
  it('returns 3 with area+title but no template version', () => {
    expect(
      maxReachableStep({
        ...INITIAL_STATE,
        profileCode: 'POP',
        areaCode: 'MFG',
        title: 'X',
      }),
    ).toBe(3);
  });
  it('returns 4 fully filled', () => {
    expect(maxReachableStep(baseFilled)).toBe(4);
  });
});

describe('clampStep', () => {
  it('caps requested step at maxReachableStep', () => {
    expect(clampStep(4, INITIAL_STATE)).toBe(1);
    expect(clampStep(3, baseFilled)).toBe(3);
  });
});

describe('canAdvance', () => {
  it('step 4 requires consent', () => {
    expect(canAdvance({ ...baseFilled, consent: false })).toBe(false);
    expect(canAdvance({ ...baseFilled, consent: true })).toBe(true);
  });
  it('step 4 blocks while submitting', () => {
    expect(canAdvance({ ...baseFilled, consent: true, submitting: true })).toBe(false);
  });
});
