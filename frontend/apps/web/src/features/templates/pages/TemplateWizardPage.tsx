import { useEffect, useMemo, useReducer } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useProfilesQuery } from '../../taxonomy/queries/useProfilesQuery';
import {
  templateWizardReducer,
  initialTemplateWizardState,
  type TemplateWizardStep,
  type ScopeType,
  type PermissionsMode,
} from '../state/templateWizard.reducer';
import { WizardShell } from '../../shared/components/wizard/WizardShell';
import type { StepperStep } from '../../../components/ui/Stepper';
import { StepScope } from '../components/wizard/steps/StepScope';
import { StepIdentity } from '../components/wizard/steps/StepIdentity';
import { StepStructure } from '../components/wizard/steps/StepStructure';
import { StepPermissions } from '../components/wizard/steps/StepPermissions';
import { StepConfirmation } from '../components/wizard/steps/StepConfirmation';

const TPL_STEPS: StepperStep[] = [
  { id: '1', label: 'Perfil' },
  { id: '2', label: 'Identidade' },
  { id: '3', label: 'Estrutura' },
  { id: '4', label: 'Permissões' },
  { id: '5', label: 'Confirmação' },
];

const NAME_MIN = 3;

function parseStepParam(raw: string | null): TemplateWizardStep {
  const n = Number(raw);
  if (n === 1 || n === 2 || n === 3 || n === 4 || n === 5) return n;
  return 1;
}

export function TemplateWizardPage(): JSX.Element {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const [state, dispatch] = useReducer(
    templateWizardReducer,
    undefined,
    () => {
      // Initial state has no scope chosen. Clamp deep-links beyond Step 1
      // to Step 1 — avoids a one-frame paint of placeholder content before
      // the defensive effect runs.
      const parsed = parseStepParam(searchParams.get('step'));
      const safeStep = initialTemplateWizardState.scopeType === null ? 1 : parsed;
      return { ...initialTemplateWizardState, step: safeStep };
    },
  );

  // Sync state.step → URL ?step=N
  useEffect(() => {
    setSearchParams(
      (prev) => {
        if (prev.get('step') === String(state.step)) return prev;
        const next = new URLSearchParams(prev);
        next.set('step', String(state.step));
        return next;
      },
      { replace: true },
    );
  }, [state.step, setSearchParams]);

  // Defensive: if URL says ?step=2 but no scope chosen yet, send back to Step 1.
  useEffect(() => {
    if (state.step !== 1 && state.scopeType === null) {
      dispatch({ type: 'GO_TO_STEP', step: 1 });
    }
  }, [state.step, state.scopeType]);

  const profilesQuery = useProfilesQuery();

  const selectedProfile = useMemo(() => {
    if (state.scopeType !== 'profile' || !state.profileCode) return null;
    return profilesQuery.data?.find((p) => p.code === state.profileCode) ?? null;
  }, [state.scopeType, state.profileCode, profilesQuery.data]);

  function handleSelectScopeType(scopeType: ScopeType) {
    dispatch({ type: 'SET_SCOPE_TYPE', scopeType });
  }

  function handleSelectProfile(code: string) {
    dispatch({ type: 'SET_PROFILE', code });
  }

  function handleAdvanceFromStep1() {
    dispatch({ type: 'GO_TO_STEP', step: 2 });
  }

  function handleAdvanceFromStep2() {
    dispatch({ type: 'GO_TO_STEP', step: 3 });
  }

  function handleAdvanceFromStep3() {
    dispatch({ type: 'GO_TO_STEP', step: 4 });
  }

  function handleAdvanceFromStep4() {
    dispatch({ type: 'GO_TO_STEP', step: 5 });
  }

  function handleBackToStep1() {
    dispatch({ type: 'GO_TO_STEP', step: 1 });
  }

  function handleBackToStep2() {
    dispatch({ type: 'GO_TO_STEP', step: 2 });
  }

  function handleBackToStep3() {
    dispatch({ type: 'GO_TO_STEP', step: 3 });
  }

  function handleBackToStep4() {
    dispatch({ type: 'GO_TO_STEP', step: 4 });
  }

  function handleSubmit() {
    navigate('/templates-v2');
  }

  function handleCancel() {
    navigate('/templates-v2');
  }

  function handleStepClick(id: string) {
    const n = Number(id) as TemplateWizardStep;
    dispatch({ type: 'GO_TO_STEP', step: n });
  }

  // Step 1 advance requires:
  //   - scope type chosen AND
  //   - if profile scope: a profile selected
  const step1Disabled =
    state.scopeType === null ||
    (state.scopeType === 'profile' && state.profileCode === null);

  // Step 2 advance requires non-empty trimmed name (≥3 chars).
  const step2Disabled = state.name.trim().length < NAME_MIN;

  // Step 3 advance requires:
  //   - starting point chosen AND
  //   - if 'docx': a file selected (mocked).
  const step3Disabled =
    state.startingPoint === null ||
    (state.startingPoint === 'docx' && state.selectedDocxName === null);

  // Step 4 advance requires:
  //   - if mode='roles': at least one role selected
  //   - modes 'areas' and 'all' are always valid (empty areas = valid per design)
  const step4Disabled =
    state.permissionsMode === 'roles' && state.selectedRoleIds.length === 0;

  return (
    <WizardShell
      kicker="Templates / Novo"
      title="Novo template reutilizável"
      description={
        <>
          Templates publicados ficam disponíveis para autores criarem novos documentos. Use placeholders{' '}
          <span className="mono">{'{{campo}}'}</span> para campos dinâmicos.
        </>
      }
      steps={TPL_STEPS}
      currentStep={String(state.step)}
      onStepClick={handleStepClick}
    >
      {state.step === 1 && (
        <StepScope
          scopeType={state.scopeType}
          onSelectScopeType={handleSelectScopeType}
          profiles={profilesQuery.data ?? []}
          isLoading={profilesQuery.isLoading}
          isError={profilesQuery.isError}
          error={profilesQuery.error}
          selectedCode={state.profileCode}
          onSelect={handleSelectProfile}
          onAdvance={handleAdvanceFromStep1}
          onCancel={handleCancel}
          advanceDisabled={step1Disabled}
          onRetry={() => void profilesQuery.refetch()}
        />
      )}
      {state.step === 2 && state.scopeType !== null && (
        <StepIdentity
          scopeType={state.scopeType}
          selectedProfile={selectedProfile}
          name={state.name}
          description={state.description}
          onChangeName={(value) => dispatch({ type: 'SET_NAME', value })}
          onChangeDescription={(value) =>
            dispatch({ type: 'SET_DESCRIPTION', value })
          }
          onAdvance={handleAdvanceFromStep2}
          onBack={handleBackToStep1}
          onChangeScope={handleBackToStep1}
          advanceDisabled={step2Disabled}
        />
      )}
      {state.step === 3 && state.scopeType !== null && (
        <StepStructure
          startingPoint={state.startingPoint}
          selectedDocxName={state.selectedDocxName}
          selectedDocxSize={state.selectedDocxSize}
          onSelectStartingPoint={(value) =>
            dispatch({ type: 'SET_STARTING_POINT', value })
          }
          onSelectDocx={(name, size) =>
            dispatch({ type: 'SET_SELECTED_DOCX', name, size })
          }
          onClearDocx={() => dispatch({ type: 'CLEAR_SELECTED_DOCX' })}
          onAdvance={handleAdvanceFromStep3}
          onBack={handleBackToStep2}
          advanceDisabled={step3Disabled}
        />
      )}
      {state.step === 4 && state.scopeType !== null && (
        <StepPermissions
          permissionsMode={state.permissionsMode}
          selectedRoleIds={state.selectedRoleIds}
          selectedAreaIds={state.selectedAreaIds}
          onSetMode={(mode: PermissionsMode) =>
            dispatch({ type: 'SET_PERMISSIONS_MODE', mode })
          }
          onToggleRole={(id) => dispatch({ type: 'TOGGLE_ROLE_ID', id })}
          onToggleArea={(id) => dispatch({ type: 'TOGGLE_AREA_ID', id })}
          onAdvance={handleAdvanceFromStep4}
          onBack={handleBackToStep3}
          advanceDisabled={step4Disabled}
        />
      )}
      {state.step === 5 && state.scopeType !== null && (
        <StepConfirmation
          name={state.name}
          description={state.description}
          scopeType={state.scopeType}
          selectedProfile={
            selectedProfile
              ? {
                  code: selectedProfile.code,
                  name: selectedProfile.name,
                  family: selectedProfile.familyCode,
                }
              : null
          }
          profileCode={state.profileCode}
          startingPoint={state.startingPoint}
          selectedDocxName={state.selectedDocxName}
          permissionsMode={state.permissionsMode}
          selectedRoleIds={state.selectedRoleIds}
          selectedAreaIds={state.selectedAreaIds}
          onBack={handleBackToStep4}
          onSubmit={handleSubmit}
        />
      )}
    </WizardShell>
  );
}

export { TemplateWizardPage as Component };
export default TemplateWizardPage;
