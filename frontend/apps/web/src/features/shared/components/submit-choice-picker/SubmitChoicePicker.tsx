import { useUsersQuery } from '../../../../lib/iam/useUsersQuery';
import type { components } from '../../../../lib/api-types';
import { isUserRole } from '../../../../lib/iam/role-vocabulary';
import styles from './SubmitChoicePicker.module.css';

export type ApprovalRoutePreview = components['schemas']['ApprovalRoutePreview'];
export type ApprovalRoutePreviewStage = components['schemas']['ApprovalRoutePreviewStage'];
export type ActorSelector = components['schemas']['ActorSelector'];
export type SubmitChosenActors = components['schemas']['SubmitChosenActors'];
// F-QA4-2: the role vocabulary and its exhaustiveness guard used to be
// duplicated here. Both now come from the one canonical runtime module, which
// derives from the same generated enum and fails the build on drift.
//
// ActorSelector.role is a free-form pattern-constrained string on the wire
// (api/openapi ActorSelector schema). The FE-only listUsers filter narrows to
// the closed UserRole union via isUserRole, so an unknown role degrades to
// `undefined` (graceful empty pool) rather than an unchecked cast.

/**
 * SLICE 8b (unit 3.2): the stages a submit-time chosen_actors picker must
 * cover — any ApprovalRoutePreviewStage carrying a `submit_choice` selector.
 * Both submit flows (document + template version) filter the preview through
 * this helper before deciding whether to render the picker at all; an empty
 * result means omit `chosen_actors` from the request entirely (no legacy
 * empty-array fallback — see SubmitDocumentRequest.chosen_actors wire docs).
 */
export function submitChoiceStagesOf(
  preview: ApprovalRoutePreview | null | undefined,
): ApprovalRoutePreviewStage[] {
  return (preview?.stages ?? []).filter((stage) =>
    stage.selectors.some((selector) => selector.kind === 'submit_choice'),
  );
}

/**
 * Friendly first line mirroring the backend's fail-closed 422 (a
 * submit_choice stage with no chosen_actors entry, or an entry with an empty
 * user_ids, is rejected server-side). Used by both submit flows to gate the
 * confirm button and to render the pt-BR validation message.
 */
export function isSubmitChoiceSelectionComplete(
  stages: ApprovalRoutePreviewStage[],
  value: SubmitChosenActors[],
): boolean {
  return stages.every((stage) => {
    const entry = value.find((v) => v.stage_order === stage.stage_order);
    return (entry?.user_ids.length ?? 0) > 0;
  });
}

interface SubmitChoicePickerProps {
  /** Pre-filtered via submitChoiceStagesOf — the caller decides whether to mount this at all. */
  stages: ApprovalRoutePreviewStage[];
  value: SubmitChosenActors[];
  onChange: (next: SubmitChosenActors[]) => void;
  /** Renders the per-stage empty-selection message once the caller attempted to submit. */
  showValidation?: boolean;
  disabled?: boolean;
}

/**
 * Multi-select picker for every submit_choice stage in a resolved
 * ApprovalRoutePreview. Renders nothing when `stages` is empty — the caller
 * (submit dialog) is expected to render this unconditionally and let it
 * collapse to null rather than branching itself, so the "no submit_choice
 * stage" path never sends a synthesized/empty chosen_actors entry.
 */
export function SubmitChoicePicker({
  stages,
  value,
  onChange,
  showValidation,
  disabled,
}: SubmitChoicePickerProps) {
  if (stages.length === 0) return null;

  const handleStageChange = (stageOrder: number, userIds: string[]) => {
    const next = value.filter((entry) => entry.stage_order !== stageOrder);
    if (userIds.length > 0) {
      next.push({ stage_order: stageOrder, user_ids: userIds });
    }
    next.sort((a, b) => a.stage_order - b.stage_order);
    onChange(next);
  };

  return (
    <div className={styles.picker} data-testid="submit-choice-picker">
      {stages.map((stage) => {
        // submit_choice → role+area_code is enforced server-side (domain + DB);
        // a stage matched by submitChoiceStagesOf always has such a selector.
        const selector = stage.selectors.find((s) => s.kind === 'submit_choice');
        if (!selector) return null;
        const entry = value.find((v) => v.stage_order === stage.stage_order);
        return (
          <SubmitChoiceStageRow
            key={stage.stage_order}
            stage={stage}
            selector={selector}
            selectedUserIds={entry?.user_ids ?? []}
            onChange={(userIds) => handleStageChange(stage.stage_order, userIds)}
            showValidation={showValidation}
            disabled={disabled}
          />
        );
      })}
    </div>
  );
}

interface SubmitChoiceStageRowProps {
  stage: ApprovalRoutePreviewStage;
  selector: ActorSelector;
  selectedUserIds: string[];
  onChange: (userIds: string[]) => void;
  showValidation?: boolean;
  disabled?: boolean;
}

function SubmitChoiceStageRow({
  stage,
  selector,
  selectedUserIds,
  onChange,
  showValidation,
  disabled,
}: SubmitChoiceStageRowProps) {
  // Eligible-user pool is the SAME role x area_code constraint the backend
  // validates chosen_actors against at submit time (422 on violation) — this
  // is the friendly first line, not a substitute for that server check.
  // Eligible-user pool is the SAME role x area_code constraint the backend
  // validates chosen_actors against at submit time (422 on violation) — this
  // is the friendly first line, not a substitute for that server check. An
  // unknown wire role degrades to `undefined` (graceful empty pool).
  const usersQuery = useUsersQuery({
    role: isUserRole(selector.role) ? selector.role : undefined,
    area_code: selector.area_code,
    is_active: true,
  });
  const users = usersQuery.data?.items ?? [];
  const hasSelection = selectedUserIds.length > 0;

  const toggle = (userId: string, checked: boolean) => {
    if (disabled) return;
    const next = checked
      ? [...selectedUserIds, userId]
      : selectedUserIds.filter((id) => id !== userId);
    onChange(next);
  };

  return (
    <fieldset className={styles.stageFieldset}>
      <legend className={styles.stageLegend}>
        Etapa {stage.stage_order} — {stage.label}: escolha o(s) aprovador(es)
      </legend>
      {usersQuery.isLoading ? (
        <p className={styles.hint}>Carregando usuários elegíveis…</p>
      ) : users.length === 0 ? (
        <p className={styles.hint} role="alert">
          Nenhum usuário elegível encontrado para esta etapa.
        </p>
      ) : (
        <ul className={styles.userList}>
          {users.map((user) => (
            <li key={user.user_id} className={styles.userItem}>
              <label className={styles.userLabel}>
                <input
                  type="checkbox"
                  checked={selectedUserIds.includes(user.user_id)}
                  onChange={(event) => toggle(user.user_id, event.target.checked)}
                  disabled={disabled}
                />
                <span>{user.display_name} ({user.username})</span>
              </label>
            </li>
          ))}
        </ul>
      )}
      {showValidation && !hasSelection ? (
        <p className={styles.error} role="alert">
          Selecione ao menos um aprovador para esta etapa antes de submeter.
        </p>
      ) : null}
    </fieldset>
  );
}
