import { useCallback, useEffect, useMemo, useReducer, useRef, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { ApiError, resolveErrorMessage } from '../../../lib/api';
import { useProfilesQuery } from '../../taxonomy/queries/useProfilesQuery';
import {
  templateWizardReducer,
  selectMaxReachableStep,
  initialTemplateWizardState,
  slugifyTemplateName,
  type TemplateWizardStep,
} from '../state/templateWizard.reducer';
import { createTemplate, importTemplateDocx } from '../api/templates';
import { WizardShell } from '../../shared/components/wizard/WizardShell';
import type { StepperStep } from '../../../components/ui/Stepper';
import { StepScope } from '../components/wizard/steps/StepScope';
import { StepIdentity } from '../components/wizard/steps/StepIdentity';
import { StepStructure } from '../components/wizard/steps/StepStructure';
import { StepConfirmation } from '../components/wizard/steps/StepConfirmation';

const EMPTY_SLUG_ERROR = 'Informe um nome com letras ou números para gerar o identificador técnico.';

// Wizard-specific copy for the create-template failures the user can actually
// act on here. Everything else falls through to `resolveErrorMessage`, the
// canonical RFC 9457 code -> PT-BR map — matching on `err.code` rather than on
// the human-readable `message`, which never carries the code.
const CREATE_ERROR_COPY: Record<string, string> = {
  // templates MapErr maps domain.ErrKeyConflict -> problem.CodeAlreadyExists.
  ALREADY_EXISTS: 'Já existe um template com este identificador técnico. Escolha outro nome.',
  // ADR 0086 config-first gate: the declared profile has no active TEMPLATE
  // approval route, so no template can be created under it until one exists.
  APPROVAL_ROUTE_MISSING:
    'Este perfil não tem rota de aprovação de template ativa. Configure a rota em Rotas de Aprovação antes de criar o template.',
};

function mapTemplateCreateError(err: unknown): string {
  if (err instanceof ApiError) {
    return CREATE_ERROR_COPY[err.code] ?? resolveErrorMessage(err);
  }
  if (err instanceof Error && err.message) return err.message;
  return 'Não foi possível criar o template. Tente novamente.';
}

const TPL_STEPS: StepperStep[] = [
  { id: '1', label: 'Perfil' },
  { id: '2', label: 'Identidade' },
  { id: '3', label: 'Estrutura' },
  { id: '4', label: 'Confirmação' },
];

function parseStepParam(raw: string | null): TemplateWizardStep {
  const n = Number(raw);
  if (n === 1 || n === 2 || n === 3 || n === 4) return n;
  return 1;
}

export function TemplateWizardPage(): JSX.Element {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  // Stable idempotency key for the wizard's create-template call. Regenerating
  // per click would defeat the backend's idempotency guard and let double-clicks
  // create duplicate templates.
  const idempotencyKeyRef = useRef<string>(crypto.randomUUID());
  // Remembers the template+version produced by a successful POST so that retries
  // after a DOCX import failure do NOT re-create the template (which would 409
  // on key_conflict and orphan the first row).
  const createdRef = useRef<{ templateId: string; versionNumber: number } | null>(null);
  const [state, dispatch] = useReducer(templateWizardReducer, undefined, () => {
    // Initial state has no profile chosen. Clamp deep-links beyond Step 1
    // to Step 1 - avoids a one-frame paint of placeholder content before
    // the defensive effect runs.
    const parsed = parseStepParam(searchParams.get('step'));
    const safeStep = initialTemplateWizardState.profileCode === null ? 1 : parsed;
    return { ...initialTemplateWizardState, step: safeStep };
  });
  const goToStep = useCallback((step: TemplateWizardStep) => {
    dispatch({ type: 'GO_TO_STEP', step });
  }, []);
  const templateKey = useMemo(() => slugifyTemplateName(state.name.trim()), [state.name]);
  const hasValidTemplateKey = templateKey.length > 0;
  const keyError = state.name.trim().length > 0 && !hasValidTemplateKey ? EMPTY_SLUG_ERROR : null;
  const maxReachableStep = selectMaxReachableStep(state);
  const advanceDisabled = maxReachableStep <= state.step || (state.step === 2 && !hasValidTemplateKey);

  // Sync state.step -> URL ?step=N
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

  // Defensive: if URL says ?step=2 but no profile chosen yet, send back to Step 1.
  useEffect(() => {
    if (state.step !== 1 && state.profileCode === null) {
      dispatch({ type: 'GO_TO_STEP', step: 1 });
    }
  }, [state.step, state.profileCode]);

  const profilesQuery = useProfilesQuery();

  const selectedProfile = useMemo(() => {
    if (!state.profileCode) return null;
    return profilesQuery.data?.find((p) => p.code === state.profileCode) ?? null;
  }, [state.profileCode, profilesQuery.data]);

  function handleSelectProfile(code: string) {
    dispatch({ type: 'SET_PROFILE', code });
  }

  async function handleSubmit() {
    if (!hasValidTemplateKey) {
      setSubmitError(EMPTY_SLUG_ERROR);
      return;
    }
    // ADR 0086: doc_type_code is required (422 when blank). The wizard cannot
    // reach step 4 without a profile, so this is a fail-closed assertion, not a
    // user-facing branch.
    if (!state.profileCode) {
      setSubmitError('Selecione o perfil do template antes de criar.');
      return;
    }

    setIsSubmitting(true);
    setSubmitError(null);
    try {
      let templateId: string;
      let versionNumber: number;
      if (createdRef.current) {
        templateId = createdRef.current.templateId;
        versionNumber = createdRef.current.versionNumber;
      } else {
        const { template, version } = await createTemplate({
          key: templateKey,
          name: state.name,
          description: state.description.trim() || undefined,
          doc_type_code: state.profileCode,
          idempotencyKey: idempotencyKeyRef.current,
        });
        templateId = template.id;
        versionNumber = version.version_number;
        createdRef.current = { templateId, versionNumber };
      }
      if (state.startingPoint === 'docx') {
        try {
          if (!state.selectedDocxFile) {
            throw new Error('Selecione um arquivo .docx antes de criar o template.');
          }
          await importTemplateDocx(templateId, versionNumber, state.selectedDocxFile);
        } catch {
          setSubmitError('Template criado, mas a importação do DOCX falhou. Tente novamente antes de abrir o editor.');
          setIsSubmitting(false);
          return;
        }
      }
      navigate(`/templates/${templateId}/edit`);
    } catch (err) {
      setSubmitError(mapTemplateCreateError(err));
      setIsSubmitting(false);
    }
  }

  function handleCancel() {
    navigate('/templates');
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
          <span className="mono">{'{campo}'}</span> para campos dinâmicos.
        </>
      }
      steps={TPL_STEPS}
      currentStep={String(state.step)}
      onStepClick={handleStepClick}
    >
      {state.step === 1 && (
        <StepScope
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
      {state.step === 2 && state.profileCode !== null && (
        <StepIdentity
          selectedProfile={selectedProfile}
          templateKey={templateKey}
          name={state.name}
          description={state.description}
          onChangeName={(value) => dispatch({ type: 'SET_NAME', value })}
          onChangeDescription={(value) => dispatch({ type: 'SET_DESCRIPTION', value })}
          onAdvance={() => goToStep(3)}
          onBack={() => goToStep(1)}
          onChangeScope={() => goToStep(1)}
          advanceDisabled={advanceDisabled}
          keyError={keyError}
        />
      )}
      {state.step === 3 && state.profileCode !== null && (
        <StepStructure
          startingPoint={state.startingPoint}
          selectedDocxName={state.selectedDocxName}
          selectedDocxSize={state.selectedDocxSize}
          onSelectStartingPoint={(value) => dispatch({ type: 'SET_STARTING_POINT', value })}
          onSelectDocxFile={(file) => dispatch({ type: 'SET_SELECTED_DOCX', file })}
          onAdvance={() => goToStep(4)}
          onBack={() => goToStep(2)}
          advanceDisabled={advanceDisabled}
        />
      )}
      {state.step === 4 && state.profileCode !== null && (
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
          templateKey={templateKey}
          startingPoint={state.startingPoint}
          selectedDocxName={state.selectedDocxName}
          onBack={() => goToStep(3)}
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
