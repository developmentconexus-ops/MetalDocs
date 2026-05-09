export type TemplateWizardStep = 1 | 2 | 3 | 4 | 5;

export type TemplateWizardState = {
  step: TemplateWizardStep;
  profileCode: string | null;
  // Steps 2–5 fields added when those steps are implemented
};

export type TemplateWizardAction =
  | { type: 'GO_TO_STEP'; step: TemplateWizardStep }
  | { type: 'SET_PROFILE'; code: string }
  | { type: 'RESET' };

export const initialTemplateWizardState: TemplateWizardState = {
  step: 1,
  profileCode: null,
};

export function templateWizardReducer(
  state: TemplateWizardState,
  action: TemplateWizardAction,
): TemplateWizardState {
  switch (action.type) {
    case 'GO_TO_STEP':
      return { ...state, step: action.step };
    case 'SET_PROFILE':
      return { ...state, profileCode: action.code };
    case 'RESET':
      return initialTemplateWizardState;
    default:
      return state;
  }
}
