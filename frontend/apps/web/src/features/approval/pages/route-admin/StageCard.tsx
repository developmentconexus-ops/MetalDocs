import { v4 as uuidv4 } from 'uuid';

import type { IamRole } from '../../../../lib/iam/roles';
import type { ProcessArea } from '../../../taxonomy/types';
import type { components } from '../../../../lib/api-types';
import {
  ACTOR_SELECTOR_KIND_OPTIONS,
  DRIFT_POLICY_OPTIONS,
  describeDriftPolicy,
  labelForDriftPolicy,
  labelForSelectorKind,
  type ActorSelectorKind,
  type DriftPolicy,
  type QuorumKind,
} from './routeAdminLabels';
import { QuorumPills } from './QuorumPills';
import type { StageKind } from './routeGovernance';
import { StageActorSlot } from './StageActorSlot';
import { StageKindControl } from './StageKindControl';
import styles from './RouteAdmin.module.css';

export type ManagedUserCore = components['schemas']['ManagedUserCore'];

export interface SelectorDraft {
  /** Stable id per draft row so React keys stay constant across edits. */
  uid: string;
  kind: ActorSelectorKind;
  userId: string;
  role: string;
  areaCode: string;
}

export interface StageDraft {
  /** Stable id per draft row so React keys stay constant across edits. */
  uid: string;
  label: string;
  requiredCapability: string;
  selectors: SelectorDraft[];
  quorumKind: QuorumKind;
  m: string;
  /**
   * '' means NO deadline. Deliberately not defaulted to a number: an empty
   * field must round-trip back to an ABSENT wire value, not to 0.
   */
  dueInDays: string;
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
  userOptions: ManagedUserCore[];
  userOptionsLoading: boolean;
  updateStage: (uid: string, patch: Partial<StageDraft>) => void;
  removeStage: (uid: string) => void;
}

export function defaultSelector(): SelectorDraft {
  return {
    uid: uuidv4() as string,
    kind: 'role_in_fixed_area',
    userId: '',
    role: '',
    areaCode: '',
  };
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
  userOptions,
  userOptionsLoading,
  updateStage,
  removeStage,
}: StageCardProps) {
  const kindLabelId = `stage-kind-label-${stage.uid}`;
  const quorumLabelId = `stage-quorum-label-${stage.uid}`;

  const setSelectors = (selectors: SelectorDraft[]) => {
    updateStage(stage.uid, { selectors });
  };

  const updateSelector = (uid: string, patch: Partial<SelectorDraft>) => {
    setSelectors(
      stage.selectors.map((selector) => (selector.uid === uid ? { ...selector, ...patch } : selector)),
    );
  };

  const addSelector = () => {
    setSelectors([...stage.selectors, defaultSelector()]);
  };

  const removeSelector = (uid: string) => {
    if (stage.selectors.length === 1) return;
    setSelectors(stage.selectors.filter((selector) => selector.uid !== uid));
  };

  const moveSelector = (index: number, direction: -1 | 1) => {
    const target = index + direction;
    if (target < 0 || target >= stage.selectors.length) return;
    const next = [...stage.selectors];
    const [moved] = next.splice(index, 1);
    next.splice(target, 0, moved);
    setSelectors(next);
  };

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
        {stage.selectors.map((selector, index) => {
          const selectorNumber = index + 1;
          const isOnlySelector = stage.selectors.length === 1;
          return (
            <div key={selector.uid} className={styles.selectorRow}>
              <div className={styles.selectorRowHeader}>
                <span className={styles.fieldLabel}>
                  Seletor {selectorNumber} da etapa {stageNumber}
                </span>
                <div className={styles.selectorRowActions}>
                  <button
                    type="button"
                    className={styles.linkButton}
                    onClick={() => moveSelector(index, -1)}
                    disabled={disabled || index === 0}
                  >
                    Subir
                  </button>
                  <button
                    type="button"
                    className={styles.linkButton}
                    onClick={() => moveSelector(index, 1)}
                    disabled={disabled || index === stage.selectors.length - 1}
                  >
                    Descer
                  </button>
                  <button
                    type="button"
                    className={styles.linkButton}
                    onClick={() => removeSelector(selector.uid)}
                    disabled={disabled || isOnlySelector}
                    title={
                      isOnlySelector ? 'A etapa deve ter ao menos um seletor.' : undefined
                    }
                  >
                    Remover seletor
                  </button>
                </div>
              </div>

              <label
                className={styles.fieldLabel}
                htmlFor={`stage-selector-kind-${selector.uid}`}
              >
                Tipo do seletor {selectorNumber} da etapa {stageNumber}
              </label>
              <select
                id={`stage-selector-kind-${selector.uid}`}
                className={styles.input}
                value={selector.kind}
                onChange={(event) =>
                  updateSelector(selector.uid, { kind: event.target.value as ActorSelectorKind })
                }
                disabled={disabled}
              >
                {ACTOR_SELECTOR_KIND_OPTIONS.map((kind) => (
                  <option key={kind} value={kind}>
                    {labelForSelectorKind(kind)}
                  </option>
                ))}
              </select>

              {selector.kind === 'named_user' ? (
                <>
                  <label
                    className={styles.fieldLabel}
                    htmlFor={`stage-selector-user-${selector.uid}`}
                  >
                    Usuário do seletor {selectorNumber} da etapa {stageNumber}
                  </label>
                  <select
                    id={`stage-selector-user-${selector.uid}`}
                    className={styles.input}
                    value={selector.userId}
                    onChange={(event) => updateSelector(selector.uid, { userId: event.target.value })}
                    disabled={disabled || userOptionsLoading}
                  >
                    <option value="" disabled>
                      {userOptionsLoading ? 'Carregando usuários…' : 'Selecione o usuário'}
                    </option>
                    {userOptions.map((user) => (
                      <option key={user.user_id} value={user.user_id}>
                        {user.display_name} ({user.username})
                      </option>
                    ))}
                  </select>
                </>
              ) : null}

              {selector.kind === 'role_in_fixed_area' || selector.kind === 'submit_choice' ? (
                <>
                  <label
                    className={styles.fieldLabel}
                    htmlFor={`stage-selector-role-${selector.uid}`}
                  >
                    Role do seletor {selectorNumber} da etapa {stageNumber}
                  </label>
                  <select
                    id={`stage-selector-role-${selector.uid}`}
                    className={styles.input}
                    value={selector.role}
                    onChange={(event) => updateSelector(selector.uid, { role: event.target.value })}
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

                  <label
                    className={styles.fieldLabel}
                    htmlFor={`stage-selector-area-${selector.uid}`}
                  >
                    Área do seletor {selectorNumber} da etapa {stageNumber}
                  </label>
                  <select
                    id={`stage-selector-area-${selector.uid}`}
                    className={styles.input}
                    value={selector.areaCode}
                    onChange={(event) =>
                      updateSelector(selector.uid, { areaCode: event.target.value })
                    }
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
                </>
              ) : null}

              {selector.kind === 'role_in_document_area' ? (
                <>
                  <label
                    className={styles.fieldLabel}
                    htmlFor={`stage-selector-role-${selector.uid}`}
                  >
                    Role do seletor {selectorNumber} da etapa {stageNumber}
                  </label>
                  <select
                    id={`stage-selector-role-${selector.uid}`}
                    className={styles.input}
                    value={selector.role}
                    onChange={(event) => updateSelector(selector.uid, { role: event.target.value })}
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
                </>
              ) : null}
            </div>
          );
        })}

        <button
          type="button"
          className={styles.ghostButton}
          onClick={addSelector}
          disabled={disabled}
        >
          Adicionar seletor
        </button>
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

      <label className={styles.fieldLabel} htmlFor={`stage-due-${stage.uid}`}>
        Prazo da etapa {stageNumber} (dias)
      </label>
      <input
        id={`stage-due-${stage.uid}`}
        className={styles.input}
        type="number"
        min={1}
        step={1}
        placeholder="Sem prazo"
        value={stage.dueInDays}
        onChange={(event) => updateStage(stage.uid, { dueInDays: event.target.value })}
        disabled={disabled}
      />

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
