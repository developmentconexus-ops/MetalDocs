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
import { useCreationContextQuery } from '../../controlled-documents/queries/useCreationContextQuery';
import { useAreasQuery } from '../queries/useAreasQuery';
import { useProfilesQuery } from '../../taxonomy/queries/useProfilesQuery';
import type { DocumentProfile } from '../../taxonomy/types';
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
    return { scope: 'company', area_codes: [], user_ids: [] };
  }
  if (state.visibility === 'area') {
    return {
      scope: 'restricted',
      area_codes: state.visibilityAreaCodes,
      user_ids: [],
    };
  }
  if (state.visibility === 'people') {
    return {
      scope: 'restricted',
      area_codes: state.visibilityAreaCodes,
      user_ids: state.invitees.map((row) => row.id),
    };
  }
  return { scope: 'company', area_codes: [], user_ids: [] };
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

  // Server state.
  //
  // D1+D2 — /controlled-documents/creation-context is the AUTHORITY for what
  // this wizard may offer: the selectable profiles (annotated with
  // approval-route readiness) and the process areas the caller may actually
  // create into (narrowed server-side from their capability grants). The raw
  // taxonomy lists below are never the eligibility source:
  //   * `catalogProfilesQuery` supplies display-only enrichment (description,
  //     família) that the lean creation-context projection does not carry;
  //     a missing entry degrades to "—", never to a selectable profile.
  //   * `areasQuery` (full active catalog) feeds ONLY the restricted-visibility
  //     picker, which answers "who may READ" — orthogonal to create grants.
  const creationContextQuery = useCreationContextQuery();
  const catalogProfilesQuery = useProfilesQuery();
  const areasQuery = useAreasQuery();
  const templatesQuery = useTemplatesByProfileQuery(state.profileCode);
  const blankTemplateQuery = useBlankTemplateQuery();
  const previewCodeQuery = usePreviewCodeQuery(state.profileCode, state.areaCode || null);

  const profiles = useMemo<DocumentProfile[]>(() => {
    const enrichment = new Map(
      (catalogProfilesQuery.data ?? []).map((profile) => [profile.code, profile]),
    );
    return (creationContextQuery.data?.profiles ?? []).map((profile) => {
      const rich = enrichment.get(profile.code);
      return {
        code: profile.code,
        name: profile.name,
        // Eligibility comes from creation-context only — never from the
        // enrichment row, so the two sources can never disagree here.
        hasActiveRoute: profile.has_active_route,
        familyCode: rich?.familyCode ?? '',
        description: rich?.description ?? '',
        reviewIntervalDays: rich?.reviewIntervalDays ?? 0,
        defaultTemplateVersionId: rich?.defaultTemplateVersionId ?? null,
        ownerUserId: rich?.ownerUserId ?? null,
        editableByRole: rich?.editableByRole ?? '',
        governanceClass: rich?.governanceClass ?? 'controlado',
        archivedAt: null,
        createdAt: rich?.createdAt ?? '',
      };
    });
  }, [creationContextQuery.data, catalogProfilesQuery.data]);

  const areas = creationContextQuery.data?.areas ?? [];
  const visibilityAreas = areasQuery.data ?? [];
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
    state.templateID === blankTemplate.template_id &&
    state.templateVersionID === blankTemplate.template_version_id;

  // Derived: URL pre-filled `?profile=X` but profile is not in the loaded list.
  // TanStack Query v5 keeps isSuccess=true after first successful fetch (background refetches
  // use isFetching, not a reset to isSuccess=false), so this condition is stable once true.
  const profileNotFound =
    creationContextQuery.isSuccess &&
    state.profileCode !== null &&
    selectedProfile === null;
  // D1 fail-closed: `?profile=X&step=3` would otherwise walk straight past the
  // disabled card in step 1. An ineligible profile blocks the whole wizard, not
  // just its own card — creating under it is a guaranteed 409.
  const profileNotSelectable = selectedProfile !== null && !selectedProfile.hasActiveRoute;
  const profileBlocked = profileNotFound || profileNotSelectable;

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
          profile_code: input.profileCode,
          process_area_code: input.areaCode,
          title: input.title,
          owner_user_id: input.ownerUserId,
          document_name: input.title,
          template_version_id: input.templateVersionID,
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
      navigate(`/documents/${result.document.id}`);
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
      {profileBlocked && (
        <div className="card" role="alert" aria-live="assertive" aria-atomic="true">
          {profileNotSelectable ? (
            <>
              O perfil <span className="mono">{state.profileCode}</span> não tem rota de aprovação
              ativa e não pode receber documentos.
            </>
          ) : (
            <>
              O perfil pré-selecionado <span className="mono">{state.profileCode}</span> não está
              mais disponível.
            </>
          )}{' '}
          <button
            type="button"
            className="btn btn-sm"
            onClick={() => dispatch({ type: 'clearProfile' })}
          >
            Limpar seleção
          </button>
        </div>
      )}
      {!profileBlocked && state.step === 1 && (
        <StepProfile
          profiles={profiles}
          isLoading={creationContextQuery.isLoading}
          isError={creationContextQuery.isError}
          error={creationContextQuery.error}
          selectedCode={state.profileCode}
          onSelect={(code) => dispatch({ type: 'selectProfile', code })}
          onAdvance={goAdvance}
          onCancel={goCancel}
          advanceDisabled={!canAdvance(state)}
          onRetry={() => void creationContextQuery.refetch()}
        />
      )}
      {!profileBlocked && state.step === 2 && (
        <StepAreaCodeVisibility
          profile={selectedProfile}
          areas={areas}
          visibilityAreas={visibilityAreas}
          isAreasLoading={creationContextQuery.isLoading}
          isAreasError={creationContextQuery.isError}
          areasError={creationContextQuery.error}
          onAreasRetry={() => void creationContextQuery.refetch()}
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
      {!profileBlocked && state.step === 3 && (
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
          blankTemplateID={blankTemplate?.template_id ?? null}
          blankTemplateVersionID={blankTemplate?.template_version_id ?? null}
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
      {!profileBlocked && state.step === 4 && (
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
          previewCodeLoading={previewCodeQuery.isLoading}
          previewCodeError={previewCodeQuery.isError}
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

