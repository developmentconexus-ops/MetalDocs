import { type FormEvent, useEffect, useMemo, useState } from 'react';

import { Dialog } from '../../../../components/ui';
import { resolveErrorMessage } from '../../../../lib/api/problem';
import { useAreasQuery } from '../../../documents/queries/useAreasQuery';
import { useProfilesQuery } from '../../../taxonomy/queries/useProfilesQuery';
import { useIamRolesQuery } from '../../queries/useIamRolesQuery';
import type { CreateRouteRequest, RouteSummary, UpdateRouteRequest } from '../../api/routeAdminApi';
import { ApprovalPolicyBadge } from './ApprovalPolicyBadge';
import {
  approveNowOverlapNote,
  authorExcludedNote,
  livreBlockedMessage,
  routePolicyFor,
  type GovernanceClass,
} from './routeGovernance';
import { defaultStage, toDraft, toStageRequests, validateDraft, type RouteDraft } from './routeDraft';
import { RouteFlowPreview } from './RouteFlowPreview';
import { StageCard, type StageDraft } from './StageCard';
import styles from './RouteAdmin.module.css';

export interface RouteEditorSubmission {
  /** Present when editing an existing route. */
  routeId?: string;
  /** Payload typed against the codegen create request. */
  create?: CreateRouteRequest;
  /** Payload typed against the codegen update request. */
  update?: UpdateRouteRequest;
}

interface RouteEditorDialogProps {
  open: boolean;
  /** Route under edit; `null` means "create new". */
  route: RouteSummary | null;
  isSubmitting: boolean;
  /** Pre-built submission lets the container decide create vs update. */
  onSubmit: (payload: RouteEditorSubmission) => Promise<void>;
  onClose: () => void;
}

export function RouteEditorDialog({
  open,
  route,
  isSubmitting,
  onSubmit,
  onClose,
}: RouteEditorDialogProps) {
  const isEdit = route !== null;
  const dialogTitle = isEdit ? 'Editar rota' : 'Criar rota';

  const [draft, setDraft] = useState<RouteDraft>(() => toDraft(route));
  const [error, setError] = useState<string | null>(null);

  // Only re-seed when the dialog (re)opens or the targeted route id changes.
  // Depending on the full `route` object would regenerate stage uuids on every
  // useRoutesQuery refetch while the dialog is open, blowing away in-progress
  // form state.
  useEffect(() => {
    if (open) {
      setDraft(toDraft(route));
      setError(null);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, route?.id]);

  const rolesQuery = useIamRolesQuery();
  const areasQuery = useAreasQuery();
  const profilesQuery = useProfilesQuery();

  const roleOptions = rolesQuery.data ?? [];
  const areaOptions = areasQuery.data ?? [];
  const profileOptions = profilesQuery.data ?? [];
  const hasNoProfiles = !profilesQuery.isLoading && !profilesQuery.isError && profileOptions.length === 0;

  // Governance class is only known when `draft.profileCode` matches an active
  // profile option. On edit, an archived route's profile may not be in that
  // list — in that case client-side gating is skipped and the backend is the
  // sole enforcer (wiring req #2).
  const selectedProfile = profileOptions.find((profile) => profile.code === draft.profileCode);
  const governanceClass: GovernanceClass | null = selectedProfile?.governanceClass ?? null;
  const policy = governanceClass ? routePolicyFor(governanceClass) : null;
  const isRouteBlocked = policy !== null && !policy.routeAllowed;

  const updateStage = (uid: string, patch: Partial<StageDraft>) => {
    setDraft((prev) => ({
      ...prev,
      stages: prev.stages.map((stage) => (stage.uid === uid ? { ...stage, ...patch } : stage)),
    }));
  };

  const addStage = () => {
    setDraft((prev) => ({ ...prev, stages: [...prev.stages, defaultStage()] }));
  };

  const removeStage = (uid: string) => {
    setDraft((prev) => ({
      ...prev,
      stages: prev.stages.length === 1 ? prev.stages : prev.stages.filter((s) => s.uid !== uid),
    }));
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);

    const validationError = validateDraft(draft, isEdit, governanceClass);
    if (validationError) {
      setError(validationError);
      return;
    }

    const stages = toStageRequests(draft);
    const payload: RouteEditorSubmission = isEdit
      ? {
          routeId: route!.id,
          update: { name: draft.name.trim(), stages },
        }
      : {
          create: {
            name: draft.name.trim(),
            profile_code: draft.profileCode.trim(),
            stages,
          },
        };

    try {
      await onSubmit(payload);
    } catch (err) {
      setError(resolveErrorMessage(err));
    }
  };

  const description = useMemo(
    () =>
      isEdit
        ? 'Edite o nome e as etapas. O código do perfil não pode ser alterado em rotas existentes.'
        : 'Defina nome, perfil e as etapas de aprovação. Use "Adicionar etapa" para incluir mais etapas.',
    [isEdit],
  );

  const footer = (
    <>
      <button
        type="button"
        className={styles.secondaryButton}
        onClick={onClose}
        disabled={isSubmitting}
      >
        Cancelar
      </button>
      <button
        type="submit"
        form="route-editor-form"
        className={styles.primaryButton}
        disabled={isSubmitting || isRouteBlocked}
      >
        {isSubmitting ? 'Salvando…' : 'Salvar rota'}
      </button>
    </>
  );

  return (
    <Dialog
      open={open}
      title={dialogTitle}
      description={description}
      onClose={isSubmitting ? () => {} : onClose}
      footer={footer}
      // Always block backdrop dismissal: editor drafts (route name + stages)
      // are easy to lose to a stray click on the dialog padding. The explicit
      // Cancel / × buttons remain the only exit affordances.
      disableBackdropClose
    >
      {error ? (
        <div className={styles.errorBox} role="alert">
          {error}
        </div>
      ) : null}

      {rolesQuery.isError ? (
        <div className={styles.warningBox} role="alert">
          Não foi possível carregar a lista de roles. Selecione manualmente ou tente novamente.
        </div>
      ) : null}

      {areasQuery.isError ? (
        <div className={styles.warningBox} role="alert">
          Não foi possível carregar as áreas. Recarregue a página para tentar novamente.
        </div>
      ) : null}

      {!isEdit && profilesQuery.isError ? (
        <div className={styles.warningBox} role="alert">
          Não foi possível carregar os perfis. Recarregue a página para tentar novamente.
        </div>
      ) : null}

      <form id="route-editor-form" className={styles.form} onSubmit={handleSubmit} noValidate>
        <label className={styles.fieldLabel} htmlFor="route-name">
          Nome da rota
        </label>
        <input
          id="route-name"
          className={styles.input}
          value={draft.name}
          onChange={(event) => setDraft((prev) => ({ ...prev, name: event.target.value }))}
          disabled={isSubmitting}
          autoFocus
        />

        <label className={styles.fieldLabel} htmlFor="route-profile-code">
          Código do perfil
        </label>
        <select
          id="route-profile-code"
          className={styles.input}
          value={draft.profileCode}
          onChange={(event) => setDraft((prev) => ({ ...prev, profileCode: event.target.value }))}
          // Edit is immutable; an empty registry or in-flight/failed load leaves
          // nothing valid to pick. The registered options guarantee the value
          // satisfies the backend FK to `metaldocs.document_profiles`.
          disabled={isSubmitting || isEdit || profilesQuery.isLoading || hasNoProfiles}
          aria-describedby={isEdit || hasNoProfiles ? 'route-profile-code-help' : undefined}
        >
          <option value="" disabled>
            {profilesQuery.isLoading ? 'Carregando perfis…' : 'Selecione o perfil'}
          </option>
          {/* Keep the immutable current code selectable on edit even if it was
              archived out of the active profile list. */}
          {isEdit && draft.profileCode && !profileOptions.some((p) => p.code === draft.profileCode) ? (
            <option value={draft.profileCode}>{draft.profileCode}</option>
          ) : null}
          {profileOptions.map((profile) => (
            <option key={profile.code} value={profile.code}>
              {profile.name} ({profile.code})
            </option>
          ))}
        </select>
        {isEdit ? (
          <small id="route-profile-code-help" className={styles.helpText}>
            Código do perfil é imutável após criação.
          </small>
        ) : hasNoProfiles ? (
          <small id="route-profile-code-help" className={styles.helpText}>
            Nenhum perfil registrado. Cadastre um perfil antes de criar a rota.
          </small>
        ) : null}

        {policy ? (
          <ApprovalPolicyBadge policy={policy} />
        ) : draft.profileCode ? (
          <small className={styles.helpText}>Perfil não classificado.</small>
        ) : null}

        {isRouteBlocked ? (
          <p className={styles.warningBox} role="alert">
            {livreBlockedMessage()}
          </p>
        ) : null}

        <p className={styles.noteBox}>{authorExcludedNote()}</p>
        <p className={styles.noteBox}>{approveNowOverlapNote()}</p>

        <section className={styles.stageSection}>
          <div className={styles.stageSectionHeader}>
            <h3 className={styles.sectionTitle}>Etapas</h3>
            <button
              type="button"
              className={styles.ghostButton}
              onClick={addStage}
              disabled={isSubmitting || isRouteBlocked}
            >
              Adicionar etapa
            </button>
          </div>

          <RouteFlowPreview
            stages={draft.stages.map((stage) => ({
              uid: stage.uid,
              label: stage.label,
              kind: stage.stageKind,
              quorum: stage.quorumKind,
            }))}
          />

          {draft.stages.map((stage, index) => (
            <StageCard
              key={stage.uid}
              stage={stage}
              stageNumber={index + 1}
              isOnly={draft.stages.length === 1}
              disabled={isSubmitting || isRouteBlocked}
              roleOptions={roleOptions}
              roleOptionsLoading={rolesQuery.isLoading}
              areaOptions={areaOptions}
              areaOptionsLoading={areasQuery.isLoading}
              updateStage={updateStage}
              removeStage={removeStage}
            />
          ))}
        </section>
      </form>
    </Dialog>
  );
}
