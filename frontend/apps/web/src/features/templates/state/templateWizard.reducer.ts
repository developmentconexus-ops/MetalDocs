export type TemplateWizardStep = 1 | 2 | 3 | 4 | 5;
export type ScopeType = 'generic' | 'profile';

export type TemplateWizardState = {
  step: TemplateWizardStep;
  /** null = not yet chosen */
  scopeType: ScopeType | null;
  /** only meaningful when scopeType === 'profile' */
  profileCode: string | null;
};

export const initialTemplateWizardState: TemplateWizardState = {
  step: 1,
  scopeType: null,
  profileCode: null,
};

export type TemplateWizardAction =
  | { type: 'SET_SCOPE_TYPE'; scopeType: ScopeType }
  | { type: 'SET_PROFILE'; code: string }
  | { type: 'GO_TO_STEP'; step: TemplateWizardStep };

export function templateWizardReducer(
  state: TemplateWizardState,
  action: TemplateWizardAction,
): TemplateWizardState {
  switch (action.type) {
    case 'SET_SCOPE_TYPE':
      return {
        ...state,
        scopeType: action.scopeType,
        // clear profile when switching to generic
        profileCode: action.scopeType === 'generic' ? null : state.profileCode,
      };
    case 'SET_PROFILE':
      return { ...state, profileCode: action.code };
    case 'GO_TO_STEP':
      return { ...state, step: action.step };
    default:
      return state;
  }
}
