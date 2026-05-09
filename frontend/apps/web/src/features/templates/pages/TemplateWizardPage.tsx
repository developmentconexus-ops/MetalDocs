import { useEffect, useReducer } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useProfilesQuery } from '../../taxonomy/queries/useProfilesQuery';
import {
  templateWizardReducer,
  initialTemplateWizardState,
  type TemplateWizardStep,
  type ScopeType,
} from '../state/templateWizard.reducer';
import { WizardShell } from '../../shared/components/wizard/WizardShell';
import type { StepperStep } from '../../../components/ui/Stepper';
import { StepScope } from '../components/wizard/steps/StepScope';

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
  const [state, dispatch] = useReducer(
    templateWizardReducer,
    undefined,
    () => ({
      ...initialTemplateWizardState,
      step: parseStepParam(searchParams.get('step')),
    }),
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

  const profilesQuery = useProfilesQuery();

  function handleSelectScopeType(scopeType: ScopeType) {
    dispatch({ type: 'SET_SCOPE_TYPE', scopeType });
  }

  function handleSelectProfile(code: string) {
    dispatch({ type: 'SET_PROFILE', code });
  }

  function handleAdvance() {
    dispatch({ type: 'GO_TO_STEP', step: 2 });
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
          onAdvance={handleAdvance}
          onCancel={handleCancel}
          advanceDisabled={step1Disabled}
          onRetry={() => void profilesQuery.refetch()}
        />
      )}
      {state.step !== 1 && (
        <div className="card">
          <div className="kicker">Etapa {state.step} de 5</div>
          <p className="caption">Em construção — próximas etapas a implementar.</p>
        </div>
      )}
    </WizardShell>
  );
}

export { TemplateWizardPage as Component };
export default TemplateWizardPage;
