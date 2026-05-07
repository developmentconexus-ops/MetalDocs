import type { VisibilityKey } from '../lib/visibilityMeta';

export type WizardStep = 1 | 2 | 3 | 4;

export type WizardInvitee = {
  id: string;
  label: string;
};

export type WizardExternalConfig = {
  passwordRequired: boolean;
  watermark: boolean;
  expiresInDays: number | null;
};

export type WizardState = {
  step: WizardStep;
  profileCode: string | null;
  areaCode: string;
  title: string;
  visibility: VisibilityKey;
  invitees: WizardInvitee[];
  external: WizardExternalConfig;
  templateID: string | null;
  templateVersionID: string | null;
  consent: boolean;
  submitting: boolean;
  error: string | null;
};

export type WizardAction =
  | { type: 'goToStep'; step: WizardStep }
  | { type: 'selectProfile'; code: string }
  | { type: 'clearProfile' }
  | { type: 'setArea'; code: string }
  | { type: 'setTitle'; value: string }
  | { type: 'setVisibility'; key: VisibilityKey }
  | { type: 'addInvitee'; invitee: WizardInvitee }
  | { type: 'removeInvitee'; id: string }
  | { type: 'setExternal'; patch: Partial<WizardExternalConfig> }
  | { type: 'selectTemplate'; templateID: string; templateVersionID: string | null }
  | { type: 'setConsent'; value: boolean }
  | { type: 'submitStart' }
  | { type: 'submitError'; message: string }
  | { type: 'submitSuccess' };

export const INITIAL_STATE: WizardState = {
  step: 1,
  profileCode: null,
  areaCode: '',
  title: '',
  visibility: 'company',
  invitees: [],
  external: { passwordRequired: false, watermark: false, expiresInDays: null },
  templateID: null,
  templateVersionID: null,
  consent: false,
  submitting: false,
  error: null,
};

export function wizardReducer(state: WizardState, action: WizardAction): WizardState {
  switch (action.type) {
    case 'goToStep':
      return { ...state, step: action.step, error: null };
    case 'selectProfile': {
      // Reducer guard: empty/whitespace code is not a valid selection.
      // Use clearProfile for the unset path.
      if (!action.code.trim()) return state;
      // Changing profile invalidates the template choice (templates are filtered by profile).
      // Title + area are kept (area is global today).
      return {
        ...state,
        profileCode: action.code,
        templateID: null,
        templateVersionID: null,
        error: null,
      };
    }
    case 'clearProfile':
      return {
        ...state,
        step: 1,
        profileCode: null,
        templateID: null,
        templateVersionID: null,
        error: null,
      };
    case 'setArea':
      return { ...state, areaCode: action.code };
    case 'setTitle':
      return { ...state, title: action.value };
    case 'setVisibility':
      return { ...state, visibility: action.key };
    case 'addInvitee': {
      if (state.invitees.some((row) => row.id === action.invitee.id)) return state;
      return { ...state, invitees: [...state.invitees, action.invitee] };
    }
    case 'removeInvitee':
      return { ...state, invitees: state.invitees.filter((row) => row.id !== action.id) };
    case 'setExternal':
      return { ...state, external: { ...state.external, ...action.patch } };
    case 'selectTemplate':
      return {
        ...state,
        templateID: action.templateID,
        templateVersionID: action.templateVersionID,
      };
    case 'setConsent':
      return { ...state, consent: action.value };
    case 'submitStart':
      return { ...state, submitting: true, error: null };
    case 'submitError':
      return { ...state, submitting: false, error: action.message };
    case 'submitSuccess':
      return { ...state, submitting: false, error: null };
    default:
      return state;
  }
}

/** Highest step the user is allowed to reach given current form state. */
export function maxReachableStep(state: WizardState): WizardStep {
  if (!state.profileCode) return 1;
  if (state.areaCode === '' || state.title.trim() === '') return 2;
  if (!state.templateVersionID) return 3;
  return 4;
}

export function clampStep(requested: WizardStep, state: WizardState): WizardStep {
  const max = maxReachableStep(state);
  return Math.min(requested, max) as WizardStep;
}

export function canAdvance(state: WizardState): boolean {
  switch (state.step) {
    case 1:
      return state.profileCode !== null;
    case 2:
      return state.areaCode !== '' && state.title.trim() !== '';
    case 3:
      return state.templateVersionID !== null;
    case 4:
      return state.consent && !state.submitting;
    default:
      return false;
  }
}
