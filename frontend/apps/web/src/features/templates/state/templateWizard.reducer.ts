export type TemplateWizardStep = 1 | 2 | 3 | 4 | 5;

export type TemplateWizardState = {
  step: TemplateWizardStep;
  profileCode: string | null;
};

export const initialTemplateWizardState: TemplateWizardState = {
  step: 1,
  profileCode: null,
};

export type TemplateWizardAction =
  | { type: 'SET_PROFILE'; code: string }
  | { type: 'GO_TO_STEP'; step: TemplateWizardStep };

export function templateWizardReducer(
  state: TemplateWizardState,
  action: TemplateWizardAction,
): TemplateWizardState {
  switch (action.type) {
    case 'SET_PROFILE':
      return { ...state, profileCode: action.code };
    case 'GO_TO_STEP':
      return { ...state, step: action.step };
    default:
      return state;
  }
}
