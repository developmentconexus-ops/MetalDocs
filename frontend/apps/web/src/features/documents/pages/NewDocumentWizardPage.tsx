import { useEffect, useMemo, useReducer } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';
import { ApiError, resolveErrorMessage } from '../../../lib/api';
import { useAuthStore } from '../../../store/auth.store';
import { createControlledDocument } from '../../registry/api/controlledDocuments';
import { createDocument } from '../api/documentsV2';
import { useAreasQuery } from '../queries/useAreasQuery';
import { useProfilesQuery } from '../queries/useProfilesQuery';
import { useTemplatesByProfileQuery } from '../queries/useTemplatesByProfileQuery';
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

export type NewDocumentWizardPageProps = Record<string, never>;

function parseStepParam(raw: string | null): WizardStep {
  const n = Number(raw);
  if (n === 1 || n === 2 || n === 3 || n === 4) return n;
  return 1;
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

  const profiles = profilesQuery.data ?? [];
  const areas = areasQuery.data ?? [];
  const templates = templatesQuery.data?.templates ?? [];

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

  // If the URL pre-fills `?profile=X` but profile isn't in the loaded list,
  // wait until profilesQuery resolves; if not present after load, clear it
  // and clamp the wizard back to step 1 via clearProfile.
  useEffect(() => {
    if (!profilesQuery.data) return;
    if (state.profileCode && !profilesQuery.data.some((p) => p.code === state.profileCode)) {
      dispatch({ type: 'clearProfile' });
    }
  }, [profilesQuery.data, state.profileCode]);

  function goCancel() {
    navigate('/documents-v2');
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

  // Submit flow — two-call sequence (slot create → doc create) via useMutation.
  // TODO(novo-documento:slot-rollback): if doc-create fails after slot-create
  // succeeds, orphan slot persists + code is consumed. No compensation today.
  // See wiki/backlog/novo-documento.md#slot-rollback.
  const createMutation = useMutation({
    mutationFn: async (input: {
      profileCode: string;
      areaCode: string;
      title: string;
      templateVersionID: string;
      ownerUserId: string;
    }) => {
      const slot = await createControlledDocument({
        profileCode: input.profileCode,
        processAreaCode: input.areaCode,
        title: input.title,
        ownerUserId: input.ownerUserId,
      });
      const doc = await createDocument({
        controlled_document_id: slot.id,
        template_version_id: input.templateVersionID,
        name: input.title,
        form_data: {},
      });
      return doc;
    },
    onMutate: () => {
      dispatch({ type: 'submitStart' });
    },
    onSuccess: (doc) => {
      dispatch({ type: 'submitSuccess' });
      navigate(`/documents-v2/${doc.document_id}`);
    },
    onError: (err) => {
      const message =
        err instanceof ApiError
          ? resolveErrorMessage(err.code, err.message)
          : err instanceof Error
            ? err.message
            : 'Falha ao criar o documento.';
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
    createMutation.mutate({
      profileCode: state.profileCode,
      areaCode: state.areaCode,
      title: state.title.trim(),
      templateVersionID: state.templateVersionID,
      ownerUserId: currentUser.userId,
    });
  }

  // Allow stepper clicks to navigate back-or-down to reachable steps only.
  function onStepClick(step: WizardStep) {
    const target = clampStep(step, state);
    dispatch({ type: 'goToStep', step: target });
  }

  return (
    <WizardShell currentStep={state.step} onStepClick={onStepClick}>
      {state.step === 1 && (
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
          invitees={state.invitees}
          external={state.external}
          onChangeProfile={() => dispatch({ type: 'goToStep', step: 1 })}
          onSetArea={(code) => dispatch({ type: 'setArea', code })}
          onSetTitle={(value) => dispatch({ type: 'setTitle', value })}
          onSetVisibility={(key) => dispatch({ type: 'setVisibility', key })}
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
          template={selectedTemplate}
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
