import { useCallback, useEffect, useMemo, useReducer, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useProfilesQuery } from '../../taxonomy/queries/useProfilesQuery';
import {
  templateWizardReducer,
  selectMaxReachableStep,
  initialTemplateWizardState,
  type TemplateWizardStep,
  type ScopeType,
  type PermissionsMode,
} from '../state/templateWizard.reducer';
import { createTemplate } from '../api/templatesV2';
import { WizardShell } from '../../shared/components/wizard/WizardShell';
import type { StepperStep } from '../../../components/ui/Stepper';
import { StepScope } from '../components/wizard/steps/StepScope';
import { StepIdentity } from '../components/wizard/steps/StepIdentity';
import { StepStructure } from '../components/wizard/steps/StepStructure';
import { StepPermissions } from '../components/wizard/steps/StepPermissions';
import { StepConfirmation } from '../components/wizard/steps/StepConfirmation';

/** Converts a display name into a stable URL-safe key (max 64 chars). */
function slugifyName(name: string): string {
  // Combining diacritical marks U+0300–U+036F (written via RegExp to avoid
  // encoding-sensitive literal unicode in source).
  const COMBINING = new RegExp('[\\u0300-\\u036f]', 'g');
  return name
    .toLowerCase()
    .normalize('NFD')
    .replace(COMBINING, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 64);
}


const TPL_STEPS: StepperStep[] = [
  { id: '1', label: 'Perfil' },
  { id: '2', label: 'Identidade' },
  { id: '3', label: 'Estrutura' },
  { id: '4', label: 'Permissões' },
  { id: '5', label: 'Confirmação' },
];

function parseStepParam(raw: string | null): TemplateWizardStep {
  const n = Number(raw);
  if (n === 1 || n === 2 || n === 3 || n === 4 || n === 5) return n;
  return 1;
}

export function TemplateWizardPage(): JSX.Element {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
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
  const goToStep = useCallback((step: TemplateWizardStep) => {
    dispatch({ type: 'GO_TO_STEP', step });
  }, []);
  const maxReachableStep = selectMaxReachableStep(state);
  const advanceDisabled = maxReachableStep <= state.step;

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

  async function handleSubmit() {
    setIsSubmitting(true);
    setSubmitError(null);
    try {
      const { template, version } = await createTemplate({
        key: slugifyName(state.name),
        name: state.name,
        description: state.description.trim() || undefined,
      });
      navigate(`/templates-v2/${template.id}/versions/${version.version_number}`);
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : 'Erro ao criar template.');
      setIsSubmitting(false);
    }
  }

  function handleCancel() {
    navigate('/templates-v2');
  }

  function handleStepClick(id: string) {
    goToStep(parseStepParam(id));
  }

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
          onAdvance={() => goToStep(2)}
          onCancel={handleCancel}
          advanceDisabled={advanceDisabled}
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
          onAdvance={() => goToStep(3)}
          onBack={() => goToStep(1)}
          onChangeScope={() => goToStep(1)}
          advanceDisabled={advanceDisabled}
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
          onAdvance={() => goToStep(4)}
          onBack={() => goToStep(2)}
          advanceDisabled={advanceDisabled}
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
          onAdvance={() => goToStep(5)}
          onBack={() => goToStep(3)}
          advanceDisabled={advanceDisabled}
        />
      )}
      {state.step === 5 && state.scopeType !== null && (
        <StepConfirmation
          name={state.name}
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
          onBack={() => goToStep(4)}
          onSubmit={() => void handleSubmit()}
          isSubmitting={isSubmitting}
          submitError={submitError}
        />
      )}
    </WizardShell>
  );
}

export { TemplateWizardPage as Component };
export default TemplateWizardPage;
