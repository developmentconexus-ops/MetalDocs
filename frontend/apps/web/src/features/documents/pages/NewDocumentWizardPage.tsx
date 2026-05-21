import { useEffect, useMemo, useReducer } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { resolveQueryError } from '../../../lib/api';
import { QK } from '../../../lib/queryKeys';
import { useAuthStore } from '../../../store/auth.store';
import { createControlledDocumentAtomic } from '../../controlled-documents/api/controlledDocuments';
import type { components } from '../../../lib/api-types';
import { usePreviewCodeQuery } from '../../controlled-documents/queries/usePreviewCodeQuery';
import { useAreasQuery } from '../queries/useAreasQuery';
import { useProfilesQuery } from '../../taxonomy/queries/useProfilesQuery';
import { useTemplatesByProfileQuery } from '../queries/useTemplatesByProfileQuery';
import { useBlankTemplateQuery } from '../queries/useBlankTemplateQuery';
import {
  INITIAL_STATE,
  canAdvance,
  clampStep,
  wizardReducer,
  type WizardState,
  type WizardStep,
} from '../state/wizard.reducer';
import { WizardShell } from '../components/wizard/WizardShell';
import { StepProfile } from '../components/wizard/steps/StepProfile';
import { StepAreaCodeVisibility } from '../components/wizard/steps/StepAreaCodeVisibility';
import { StepTemplate } from '../components/wizard/steps/StepTemplate';
import { StepConfirm } from '../components/wizard/steps/StepConfirm';

function parseStepParam(raw: string | null): WizardStep {
  const n = Number(raw);
  if (n === 1 || n === 2 || n === 3 || n === 4) return n;
  return 1;
}

export function buildVisibilityPayload(
  state: WizardState,
): components['schemas']['ControlledDocumentVisibility'] {
  if (state.visibility === 'company') {
    return { scope: 'company', areaCodes: [], userIds: [] };
  }
  if (state.visibility === 'area') {
    return {
      scope: 'restricted',
      areaCodes: state.visibilityAreaCodes,
      userIds: [],
    };
  }
  if (state.visibility === 'people') {
    return {
      scope: 'restricted',
      areaCodes: state.visibilityAreaCodes,
      userIds: state.invitees.map((row) => row.id),
    };
  }
  return { scope: 'company', areaCodes: [], userIds: [] };
}

function initialStateFromUrl(searchParams: URLSearchParams): WizardState {
  const profileParam = searchParams.get('profile');
  const stepParam = parseStepParam(searchParams.get('step'));
  const base: WizardState = {
    ...INITIAL_STATE,
    profileCode: profileParam || null,
  };
  // Don't allow URL-seeded step to exceed what the seeded form-state permits.
  return { ...base, step: clampStep(stepParam, base) };
}

export function NewDocumentWizardPage(): JSX.Element {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const currentUser = useAuthStore((s) => s.user);
  const queryClient = useQueryClient();

  const [state, dispatch] = useReducer(wizardReducer, undefined, () => initialStateFromUrl(searchParams));

  // Sync state.step → URL ?step=N. Use updater form so we don't depend on the
  // referentially-unstable `searchParams` object (re-instantiated every render).
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

  // Server state
  const profilesQuery = useProfilesQuery();
  const areasQuery = useAreasQuery();
  const templatesQuery = useTemplatesByProfileQuery(state.profileCode);
  const blankTemplateQuery = useBlankTemplateQuery();
  const previewCodeQuery = usePreviewCodeQuery(state.profileCode, state.areaCode || null);

  const profiles = profilesQuery.data ?? [];
  const areas = areasQuery.data ?? [];
  const templates = templatesQuery.data?.templates ?? [];
  const blankTemplate = blankTemplateQuery.data ?? null;

  // Derived selections
  const selectedProfile = useMemo(
    () => profiles.find((p) => p.code === state.profileCode) ?? null,
    [profiles, state.profileCode],
  );
  const selectedArea = useMemo(
    () => areas.find((a) => a.code === state.areaCode) ?? null,
    [areas, state.areaCode],
  );
  const selectedTemplate = useMemo(
    () => templates.find((t) => t.id === state.templateID) ?? null,
    [templates, state.templateID],
  );
  const blankTemplateSelected =
    blankTemplate !== null &&
    state.templateID === blankTemplate.templateId &&
    state.templateVersionID === blankTemplate.templateVersionId;

  // Derived: URL pre-filled `?profile=X` but profile is not in the loaded list.
  // TanStack Query v5 keeps isSuccess=true after first successful fetch (background refetches
  // use isFetching, not a reset to isSuccess=false), so this condition is stable once true.
  const profileNotFound =
    profilesQuery.isSuccess &&
    state.profileCode !== null &&
    selectedProfile === null;

  function goCancel() {
    navigate('/documents');
  }

  function goBack() {
    if (state.step > 1) {
      dispatch({ type: 'goToStep', step: (state.step - 1) as WizardStep });
    }
  }

  function goAdvance() {
    if (!canAdvance(state)) return;
    if (state.step < 4) {
      dispatch({ type: 'goToStep', step: (state.step + 1) as WizardStep });
    }
  }

  const createMutation = useMutation({
    mutationFn: async (input: {
      profileCode: string;
      areaCode: string;
      title: string;
      templateVersionID: string;
      ownerUserId: string;
      idempotencyKey: string;
      visibility: components['schemas']['ControlledDocumentVisibility'];
    }) => {
      return createControlledDocumentAtomic(
        {
          profileCode: input.profileCode,
          processAreaCode: input.areaCode,
          title: input.title,
          ownerUserId: input.ownerUserId,
          documentName: input.title,
          templateVersionId: input.templateVersionID,
          visibility: input.visibility,
        },
        input.idempotencyKey,
      );
    },
    onMutate: () => {
      dispatch({ type: 'submitStart' });
    },
    onSuccess: (result) => {
      if (state.profileCode && state.areaCode) {
        void queryClient.invalidateQueries({
          queryKey: QK.controlledDocuments.preview(state.profileCode, state.areaCode),
        });
      }
      dispatch({ type: 'submitSuccess' });
      navigate(`/documents/${result.document.id}/edit`);
    },
    onError: (err) => {
      const message = resolveQueryError(err, 'Falha ao criar o documento.');
      dispatch({ type: 'submitError', message });
      toast.error(message);
    },
  });

  function handleCreate() {
    if (
      !state.profileCode ||
      !state.areaCode ||
      !state.title.trim() ||
      !state.templateVersionID
    )
      return;
    if (!currentUser?.userId) {
      const message = 'Aguardando autenticação. Tente novamente em alguns segundos.';
      dispatch({ type: 'submitError', message });
      toast.error(message);
      return;
    }
    const idempotencyKey = crypto.randomUUID();
    createMutation.mutate({
      profileCode: state.profileCode,
      areaCode: state.areaCode,
      title: state.title.trim(),
      templateVersionID: state.templateVersionID,
      ownerUserId: currentUser.userId,
      idempotencyKey,
      visibility: buildVisibilityPayload(state),
    });
  }

  // Allow stepper clicks to navigate back-or-down to reachable steps only.
  function onStepClick(step: WizardStep) {
    const target = clampStep(step, state);
    dispatch({ type: 'goToStep', step: target });
  }

  return (
    <WizardShell currentStep={state.step} onStepClick={onStepClick}>
      {state.step === 1 && profileNotFound && (
        <div className="card" role="alert" aria-live="assertive" aria-atomic="true">
          O perfil pré-selecionado <span className="mono">{state.profileCode}</span> não está mais disponível.
          {' '}
          <button
            type="button"
            className="btn btn-sm"
            onClick={() => dispatch({ type: 'clearProfile' })}
          >
            Limpar seleção
          </button>
        </div>
      )}
      {state.step === 1 && !profileNotFound && (
        <StepProfile
          profiles={profiles}
          isLoading={profilesQuery.isLoading}
          isError={profilesQuery.isError}
          error={profilesQuery.error}
          selectedCode={state.profileCode}
          onSelect={(code) => dispatch({ type: 'selectProfile', code })}
          onAdvance={goAdvance}
          onCancel={goCancel}
          advanceDisabled={!canAdvance(state)}
          onRetry={() => void profilesQuery.refetch()}
        />
      )}
      {state.step === 2 && (
        <StepAreaCodeVisibility
          profile={selectedProfile}
          areas={areas}
          isAreasLoading={areasQuery.isLoading}
          isAreasError={areasQuery.isError}
          areasError={areasQuery.error}
          onAreasRetry={() => void areasQuery.refetch()}
          areaCode={state.areaCode}
          title={state.title}
          visibility={state.visibility}
          visibilityAreaCodes={state.visibilityAreaCodes}
          invitees={state.invitees}
          external={state.external}
          onChangeProfile={() => dispatch({ type: 'goToStep', step: 1 })}
          onSetArea={(code) => dispatch({ type: 'setArea', code })}
          onSetTitle={(value) => dispatch({ type: 'setTitle', value })}
          onSetVisibility={(key) => dispatch({ type: 'setVisibility', key })}
          onSetVisibilityAreas={(codes) => dispatch({ type: 'setVisibilityAreas', codes })}
          onAddInvitee={(invitee) => dispatch({ type: 'addInvitee', invitee })}
          onRemoveInvitee={(id) => dispatch({ type: 'removeInvitee', id })}
          onSetExternal={(patch) => dispatch({ type: 'setExternal', patch })}
          onAdvance={goAdvance}
          onBack={goBack}
          onCancel={goCancel}
          advanceDisabled={!canAdvance(state)}
        />
      )}
      {state.step === 3 && (
        <StepTemplate
          profileLabel={
            selectedProfile ? `Perfil ${selectedProfile.code} — ${selectedProfile.name}` : 'Perfil'
          }
          templates={templates}
          isLoading={templatesQuery.isLoading}
          isError={templatesQuery.isError}
          error={templatesQuery.error}
          onRetry={() => void templatesQuery.refetch()}
          selectedTemplateID={state.templateID}
          selectedVersionID={state.templateVersionID}
          blankTemplateID={blankTemplate?.templateId ?? null}
          blankTemplateVersionID={blankTemplate?.templateVersionId ?? null}
          blankTemplateName={blankTemplate?.name ?? 'Em branco'}
          onSelect={(templateID, versionID) =>
            dispatch({ type: 'selectTemplate', templateID, templateVersionID: versionID })
          }
          onAdvance={goAdvance}
          onBack={goBack}
          onCancel={goCancel}
          advanceDisabled={!canAdvance(state)}
        />
      )}
      {state.step === 4 && (
        <StepConfirm
          profile={selectedProfile}
          area={selectedArea}
          title={state.title}
          visibility={state.visibility}
          visibilityAreaCodes={state.visibilityAreaCodes}
          inviteeCount={state.invitees.length}
          template={selectedTemplate}
          isBlankTemplateSelected={blankTemplateSelected}
          blankTemplateName={blankTemplate?.name ?? 'Em branco'}
          previewCode={previewCodeQuery.data?.code ?? null}
          authorDisplayName={currentUser?.displayName ?? ''}
          createdAt={new Date()}
          consent={state.consent}
          submitting={state.submitting}
          error={state.error}
          onConsent={(value) => dispatch({ type: 'setConsent', value })}
          onSubmit={() => void handleCreate()}
          onBack={goBack}
          onCancel={goCancel}
          submitDisabled={!canAdvance(state)}
        />
      )}
    </WizardShell>
  );
}

export default NewDocumentWizardPage;

