export type TemplateWizardStep = 1 | 2 | 3 | 4;
export type ScopeType = 'generic' | 'profile';
export type StartingPoint = 'docx' | 'blank';

export type TemplateWizardState = {
  step: TemplateWizardStep;
  /** null = not yet chosen */
  scopeType: ScopeType | null;
  /** only meaningful when scopeType === 'profile' */
  profileCode: string | null;
  /** Step 2 - Identidade */
  name: string;
  description: string;
  /** Step 3 - Estrutura */
  startingPoint: StartingPoint | null;
  selectedDocxFile: File | null;
  selectedDocxName: string | null;
  selectedDocxSize: number | null;
};

export const initialTemplateWizardState: TemplateWizardState = {
  step: 1,
  scopeType: null,
  profileCode: null,
  name: '',
  description: '',
  startingPoint: null,
  selectedDocxFile: null,
  selectedDocxName: null,
  selectedDocxSize: null,
};

export function slugifyTemplateName(name: string): string {
  const combining = new RegExp('[\\u0300-\\u036f]', 'g');
  return name
    .toLowerCase()
    .normalize('NFD')
    .replace(combining, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 64);
}

export type TemplateWizardAction =
  | { type: 'SET_SCOPE_TYPE'; scopeType: ScopeType }
  | { type: 'SET_PROFILE'; code: string }
  | { type: 'SET_NAME'; value: string }
  | { type: 'SET_DESCRIPTION'; value: string }
  | { type: 'SET_STARTING_POINT'; value: StartingPoint }
  | { type: 'SET_SELECTED_DOCX'; file: File }
  | { type: 'CLEAR_SELECTED_DOCX' }
  | { type: 'GO_TO_STEP'; step: TemplateWizardStep };

export function templateWizardReducer(
  state: TemplateWizardState,
  action: TemplateWizardAction,
): TemplateWizardState {
  const next = reduceCore(state, action);
  const max = selectMaxReachableStep(next);
  return next.step > max ? { ...next, step: max } : next;
}

function reduceCore(
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
    case 'SET_NAME':
      return { ...state, name: action.value };
    case 'SET_DESCRIPTION':
      return { ...state, description: action.value };
    case 'SET_STARTING_POINT':
      return {
        ...state,
        startingPoint: action.value,
        selectedDocxFile: action.value === 'blank' ? null : state.selectedDocxFile,
        selectedDocxName: action.value === 'blank' ? null : state.selectedDocxName,
        selectedDocxSize: action.value === 'blank' ? null : state.selectedDocxSize,
      };
    case 'SET_SELECTED_DOCX':
      return {
        ...state,
        startingPoint: 'docx',
        selectedDocxFile: action.file,
        selectedDocxName: action.file.name,
        selectedDocxSize: action.file.size,
      };
    case 'CLEAR_SELECTED_DOCX':
      return {
        ...state,
        selectedDocxFile: null,
        selectedDocxName: null,
        selectedDocxSize: null,
      };
    case 'GO_TO_STEP':
      return { ...state, step: action.step };
    default:
      return state;
  }
}

export function selectMaxReachableStep(
  state: TemplateWizardState,
): TemplateWizardStep {
  if (state.scopeType === null) return 1;
  if (state.scopeType === 'profile' && state.profileCode === null) return 1;
  if (state.name.trim().length < 3) return 2;
  if (slugifyTemplateName(state.name.trim()).length === 0) return 2;
  if (state.startingPoint === null) return 3;
  if (state.startingPoint === 'docx' && state.selectedDocxFile === null) return 3;
  return 4;
}
