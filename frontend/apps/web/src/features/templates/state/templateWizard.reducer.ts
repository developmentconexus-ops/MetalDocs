export type TemplateWizardStep = 1 | 2 | 3 | 4 | 5;
export type ScopeType = 'generic' | 'profile';
export type StartingPoint = 'docx' | 'blank';
export type PermissionsMode = 'roles' | 'areas' | 'all';

export type TemplateWizardState = {
  step: TemplateWizardStep;
  /** null = not yet chosen */
  scopeType: ScopeType | null;
  /** only meaningful when scopeType === 'profile' */
  profileCode: string | null;
  /** Step 2 — Identidade */
  name: string;
  description: string;
  /** Step 3 — Estrutura */
  startingPoint: StartingPoint | null;
  /** mocked at Step 3 — real upload deferred to post-create editor handoff
   *  (backlog: step3-docx-upload, step3-editor-handoff). */
  selectedDocxName: string | null;
  selectedDocxSize: number | null;
  /** Step 4 — Permissões */
  permissionsMode: PermissionsMode;
  selectedRoleIds: string[];
  selectedAreaIds: string[];
};

export const initialTemplateWizardState: TemplateWizardState = {
  step: 1,
  scopeType: null,
  profileCode: null,
  name: '',
  description: '',
  startingPoint: null,
  selectedDocxName: null,
  selectedDocxSize: null,
  permissionsMode: 'all',
  selectedRoleIds: [],
  selectedAreaIds: [],
};

export type TemplateWizardAction =
  | { type: 'SET_SCOPE_TYPE'; scopeType: ScopeType }
  | { type: 'SET_PROFILE'; code: string }
  | { type: 'SET_NAME'; value: string }
  | { type: 'SET_DESCRIPTION'; value: string }
  | { type: 'SET_STARTING_POINT'; value: StartingPoint }
  | { type: 'SET_SELECTED_DOCX'; name: string; size: number }
  | { type: 'CLEAR_SELECTED_DOCX' }
  | { type: 'SET_PERMISSIONS_MODE'; mode: PermissionsMode }
  | { type: 'TOGGLE_ROLE_ID'; id: string }
  | { type: 'TOGGLE_AREA_ID'; id: string }
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
        // clear docx selection when switching to blank
        selectedDocxName: action.value === 'blank' ? null : state.selectedDocxName,
        selectedDocxSize: action.value === 'blank' ? null : state.selectedDocxSize,
      };
    case 'SET_SELECTED_DOCX':
      return { ...state, selectedDocxName: action.name, selectedDocxSize: action.size };
    case 'CLEAR_SELECTED_DOCX':
      return { ...state, selectedDocxName: null, selectedDocxSize: null };
    case 'SET_PERMISSIONS_MODE':
      return { ...state, permissionsMode: action.mode };
    case 'TOGGLE_ROLE_ID':
      return {
        ...state,
        selectedRoleIds: state.selectedRoleIds.includes(action.id)
          ? state.selectedRoleIds.filter((id) => id !== action.id)
          : [...state.selectedRoleIds, action.id],
      };
    case 'TOGGLE_AREA_ID':
      return {
        ...state,
        selectedAreaIds: state.selectedAreaIds.includes(action.id)
          ? state.selectedAreaIds.filter((id) => id !== action.id)
          : [...state.selectedAreaIds, action.id],
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
  if (state.startingPoint === null) return 3;
  if (state.startingPoint === 'docx' && state.selectedDocxName === null)
    return 3;
  if (state.permissionsMode === 'roles' && state.selectedRoleIds.length === 0)
    return 4;
  return 5;
}
