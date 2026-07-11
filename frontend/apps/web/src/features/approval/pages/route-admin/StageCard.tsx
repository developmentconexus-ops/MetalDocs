import type { IamRole } from '../../../../lib/iam/roles';
import type { ProcessArea } from '../../../taxonomy/types';
import {
  DRIFT_POLICY_OPTIONS,
  describeDriftPolicy,
  labelForDriftPolicy,
  type DriftPolicy,
  type QuorumKind,
} from './routeAdminLabels';
import { QuorumPills } from './QuorumPills';
import type { StageKind } from './routeGovernance';
import { StageActorSlot } from './StageActorSlot';
import { StageKindControl } from './StageKindControl';
import styles from './RouteAdmin.module.css';

export interface StageDraft {
  /** Stable id per draft row so React keys stay constant across edits. */
  uid: string;
  label: string;
  requiredRole: string;
  requiredCapability: string;
  areaCode: string;
  quorumKind: QuorumKind;
  m: string;
  driftPolicy: DriftPolicy;
  stageKind: StageKind;
}

interface StageCardProps {
  stage: StageDraft;
  stageNumber: number;
  isOnly: boolean;
  disabled: boolean;
  roleOptions: IamRole[];
  roleOptionsLoading: boolean;
  areaOptions: ProcessArea[];
  areaOptionsLoading: boolean;
  updateStage: (uid: string, patch: Partial<StageDraft>) => void;
  removeStage: (uid: string) => void;
}

export function StageCard({
  stage,
  stageNumber,
  isOnly,
  disabled,
  roleOptions,
  roleOptionsLoading,
  areaOptions,
  areaOptionsLoading,
  updateStage,
  removeStage,
}: StageCardProps) {
  const kindLabelId = `stage-kind-label-${stage.uid}`;
  const quorumLabelId = `stage-quorum-label-${stage.uid}`;

  return (
    <article className={styles.stageCard}>
      <div className={styles.stageHeader}>
        <strong>Etapa {stageNumber}</strong>
        <button
          type="button"
          className={styles.linkButton}
          onClick={() => removeStage(stage.uid)}
          disabled={disabled || isOnly}
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
        disabled={disabled}
      />

      <span id={kindLabelId} className={styles.fieldLabel}>
        Tipo da etapa {stageNumber}
      </span>
      <StageKindControl
        id={`stage-kind-${stage.uid}`}
        value={stage.stageKind}
        onChange={(kind) => updateStage(stage.uid, { stageKind: kind })}
        disabled={disabled}
        ariaLabelledBy={kindLabelId}
      />

      <StageActorSlot heading={`Quem atua na etapa ${stageNumber}`}>
        <label className={styles.fieldLabel} htmlFor={`stage-role-${stage.uid}`}>
          Role requerida da etapa {stageNumber}
        </label>
        <select
          id={`stage-role-${stage.uid}`}
          className={styles.input}
          value={stage.requiredRole}
          onChange={(event) => updateStage(stage.uid, { requiredRole: event.target.value })}
          disabled={disabled || roleOptionsLoading}
        >
          <option value="" disabled>
            {roleOptionsLoading ? 'Carregando roles…' : 'Selecione a role'}
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
          disabled={disabled || areaOptionsLoading}
        >
          <option value="" disabled>
            {areaOptionsLoading ? 'Carregando áreas…' : 'Selecione a área'}
          </option>
          {areaOptions.map((area) => (
            <option key={area.code} value={area.code}>
              {area.name} ({area.code})
            </option>
          ))}
        </select>
      </StageActorSlot>

      <span id={quorumLabelId} className={styles.fieldLabel}>
        Quórum da etapa {stageNumber}
      </span>
      <QuorumPills
        id={`stage-quorum-${stage.uid}`}
        value={stage.quorumKind}
        onChange={(kind) => updateStage(stage.uid, { quorumKind: kind })}
        disabled={disabled}
        ariaLabelledBy={quorumLabelId}
      />

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
            disabled={disabled}
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
        disabled={disabled}
      >
        {DRIFT_POLICY_OPTIONS.map((driftPolicy) => (
          <option key={driftPolicy} value={driftPolicy}>
            {labelForDriftPolicy(driftPolicy)}
          </option>
        ))}
      </select>
      <small className={styles.helpText}>{describeDriftPolicy(stage.driftPolicy)}</small>
    </article>
  );
}
