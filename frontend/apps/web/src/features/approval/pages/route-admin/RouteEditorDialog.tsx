import { type FormEvent, useEffect, useMemo, useState } from 'react';
// @ts-expect-error uuid package exists in workspace, no typings exposed in this app.
import { v4 as uuidv4 } from 'uuid';

import { Dialog } from '../../../../components/ui';
import { useAreasQuery } from '../../../documents/queries/useAreasQuery';
import { useIamRolesQuery } from '../../queries/useIamRolesQuery';
import { ApprovalError } from '../../api/mutationClient';
import type {
  CreateRouteRequest,
  RouteSummary,
  StageRequest,
  UpdateRouteRequest,
} from '../../api/routeAdminApi';
import {
  DRIFT_POLICY_OPTIONS,
  QUORUM_KIND_OPTIONS,
  describeDriftPolicy,
  describeQuorumKind,
  labelForDriftPolicy,
  labelForQuorumKind,
  messageForRouteError,
  type DriftPolicy,
  type QuorumKind,
} from './routeAdminLabels';
import styles from './RouteAdmin.module.css';

interface StageDraft {
  /** Stable id per draft row so React keys stay constant across edits. */
  uid: string;
  label: string;
  requiredRole: string;
  requiredCapability: string;
  areaCode: string;
  quorumKind: QuorumKind;
  m: string;
  driftPolicy: DriftPolicy;
}

interface RouteDraft {
  name: string;
  profileCode: string;
  stages: StageDraft[];
}

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

function defaultStage(): StageDraft {
  return {
    uid: uuidv4() as string,
    label: '',
    requiredRole: '',
    requiredCapability: 'doc.signoff',
    areaCode: '',
    quorumKind: 'any_1_of',
    m: '1',
    driftPolicy: 'reduce_quorum',
  };
}

function toDraft(route: RouteSummary | null): RouteDraft {
  if (!route) {
    return { name: '', profileCode: '', stages: [defaultStage()] };
  }
  return {
    name: route.name,
    profileCode: route.profile_code,
    stages: route.stages.map((stage) => ({
      uid: uuidv4() as string,
      label: stage.label,
      requiredRole: stage.required_role ?? '',
      requiredCapability: stage.required_capability || 'doc.signoff',
      areaCode: stage.area_code ?? '',
      quorumKind: stage.quorum_kind,
      m: String(stage.quorum_m ?? 1),
      driftPolicy: stage.drift_policy,
    })),
  };
}

function toStageRequests(draft: RouteDraft): StageRequest[] {
  return draft.stages.map((stage, index): StageRequest => {
    const payload: StageRequest = {
      order: index + 1,
      name: stage.label.trim(),
      required_role: stage.requiredRole.trim(),
      required_capability: stage.requiredCapability.trim() || 'doc.signoff',
      area_code: stage.areaCode.trim(),
      quorum: stage.quorumKind,
      drift_policy: stage.driftPolicy,
    };
    if (stage.quorumKind === 'm_of_n') {
      payload.quorum_m = Number(stage.m);
    }
    return payload;
  });
}

function validateDraft(draft: RouteDraft, isEdit: boolean): string | null {
  if (!draft.name.trim()) {
    return 'Informe o nome da rota.';
  }
  if (!isEdit && !draft.profileCode.trim()) {
    return 'Informe o código do perfil.';
  }
  if (draft.stages.length === 0) {
    return 'A rota deve possuir ao menos uma etapa.';
  }

  const labels = new Set<string>();
  for (const stage of draft.stages) {
    const label = stage.label.trim();
    if (!label) {
      return 'Toda etapa deve ter nome.';
    }
    const normalized = label.toLocaleLowerCase('pt-BR');
    if (labels.has(normalized)) {
      return 'Nomes de etapa devem ser distintos.';
    }
    labels.add(normalized);

    if (!stage.requiredRole) {
      return `A etapa "${label}" deve ter uma role definida.`;
    }
    if (!stage.areaCode) {
      return `A etapa "${label}" deve ter uma área definida.`;
    }
    if (stage.quorumKind === 'm_of_n') {
      const mValue = Number(stage.m);
      if (!Number.isFinite(mValue) || mValue < 1) {
        return `Na etapa "${label}", informe um valor de M válido.`;
      }
    }
  }
  return null;
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

  const roleOptions = rolesQuery.data ?? [];
  const areaOptions = areasQuery.data ?? [];

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

    const validationError = validateDraft(draft, isEdit);
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
      if (err instanceof ApprovalError) {
        setError(messageForRouteError(err.code, err.status, err.message));
      } else if (err instanceof Error) {
        setError(messageForRouteError(undefined, undefined, err.message));
      } else {
        setError(messageForRouteError(undefined, undefined));
      }
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
        disabled={isSubmitting}
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
      disableBackdropClose={isSubmitting}
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
        <input
          id="route-profile-code"
          className={styles.input}
          value={draft.profileCode}
          onChange={(event) =>
            setDraft((prev) => ({ ...prev, profileCode: event.target.value.toUpperCase() }))
          }
          disabled={isSubmitting || isEdit}
          aria-describedby={isEdit ? 'route-profile-code-help' : undefined}
        />
        {isEdit ? (
          <small id="route-profile-code-help" className={styles.helpText}>
            Código do perfil é imutável após criação.
          </small>
        ) : null}

        <section className={styles.stageSection}>
          <div className={styles.stageSectionHeader}>
            <h3 className={styles.sectionTitle}>Etapas</h3>
            <button
              type="button"
              className={styles.ghostButton}
              onClick={addStage}
              disabled={isSubmitting}
            >
              Adicionar etapa
            </button>
          </div>

          {draft.stages.map((stage, index) => {
            const stageNumber = index + 1;
            const isOnly = draft.stages.length === 1;
            return (
              <article className={styles.stageCard} key={stage.uid}>
                <div className={styles.stageHeader}>
                  <strong>Etapa {stageNumber}</strong>
                  <button
                    type="button"
                    className={styles.linkButton}
                    onClick={() => removeStage(stage.uid)}
                    disabled={isSubmitting || isOnly}
                    title={isOnly ? 'A rota deve ter ao menos uma etapa.' : undefined}
                  >
                    Remover
                  </button>
                </div>

                <label className={styles.fieldLabel} htmlFor={`stage-label-${stage.uid}`}>
                  Nome da etapa {stageNumber}
                </label>
                <input
                  id={`stage-label-${stage.uid}`}
                  className={styles.input}
                  value={stage.label}
                  onChange={(event) => updateStage(stage.uid, { label: event.target.value })}
                  disabled={isSubmitting}
                />

                <label className={styles.fieldLabel} htmlFor={`stage-role-${stage.uid}`}>
                  Role requerida da etapa {stageNumber}
                </label>
                <select
                  id={`stage-role-${stage.uid}`}
                  className={styles.input}
                  value={stage.requiredRole}
                  onChange={(event) => updateStage(stage.uid, { requiredRole: event.target.value })}
                  disabled={isSubmitting || rolesQuery.isLoading}
                >
                  <option value="" disabled>
                    {rolesQuery.isLoading ? 'Carregando roles…' : 'Selecione a role'}
                  </option>
                  {roleOptions.map((role) => (
                    <option key={role.code} value={role.code} title={role.description}>
                      {role.label}
                    </option>
                  ))}
                </select>

                <label className={styles.fieldLabel} htmlFor={`stage-area-${stage.uid}`}>
                  Área da etapa {stageNumber}
                </label>
                <select
                  id={`stage-area-${stage.uid}`}
                  className={styles.input}
                  value={stage.areaCode}
                  onChange={(event) => updateStage(stage.uid, { areaCode: event.target.value })}
                  disabled={isSubmitting || areasQuery.isLoading}
                >
                  <option value="" disabled>
                    {areasQuery.isLoading ? 'Carregando áreas…' : 'Selecione a área'}
                  </option>
                  {areaOptions.map((area) => (
                    <option key={area.code} value={area.code}>
                      {area.name} ({area.code})
                    </option>
                  ))}
                </select>

                <label className={styles.fieldLabel} htmlFor={`stage-quorum-${stage.uid}`}>
                  Quórum da etapa {stageNumber}
                </label>
                <select
                  id={`stage-quorum-${stage.uid}`}
                  className={styles.input}
                  value={stage.quorumKind}
                  onChange={(event) =>
                    updateStage(stage.uid, { quorumKind: event.target.value as QuorumKind })
                  }
                  disabled={isSubmitting}
                >
                  {QUORUM_KIND_OPTIONS.map((kind) => (
                    <option key={kind} value={kind}>
                      {labelForQuorumKind(kind)}
                    </option>
                  ))}
                </select>
                <small className={styles.helpText}>{describeQuorumKind(stage.quorumKind)}</small>

                {stage.quorumKind === 'm_of_n' ? (
                  <>
                    <label className={styles.fieldLabel} htmlFor={`stage-m-${stage.uid}`}>
                      M da etapa {stageNumber}
                    </label>
                    <input
                      id={`stage-m-${stage.uid}`}
                      className={styles.input}
                      type="number"
                      min={1}
                      step={1}
                      value={stage.m}
                      onChange={(event) => updateStage(stage.uid, { m: event.target.value })}
                      disabled={isSubmitting}
                    />
                  </>
                ) : null}

                <label className={styles.fieldLabel} htmlFor={`stage-drift-${stage.uid}`}>
                  Política de drift da etapa {stageNumber}
                </label>
                <select
                  id={`stage-drift-${stage.uid}`}
                  className={styles.input}
                  value={stage.driftPolicy}
                  onChange={(event) =>
                    updateStage(stage.uid, { driftPolicy: event.target.value as DriftPolicy })
                  }
                  disabled={isSubmitting}
                >
                  {DRIFT_POLICY_OPTIONS.map((policy) => (
                    <option key={policy} value={policy}>
                      {labelForDriftPolicy(policy)}
                    </option>
                  ))}
                </select>
                <small className={styles.helpText}>{describeDriftPolicy(stage.driftPolicy)}</small>
              </article>
            );
          })}
        </section>
      </form>
    </Dialog>
  );
}
