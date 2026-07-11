import { useState } from 'react';

import { ArtifactDecisionPanel } from '../../../shared/controlled-artifact/ArtifactDecisionPanel';
import type { ArtifactAction, ArtifactDecisionModel } from '../../../shared/controlled-artifact/types';
import type { ApprovalInstance, StageInstance } from '../../api/approvalTypes';
import { deriveStageMode } from '../../lib/workspaceMode';
import { useReviewVerdictMutation } from '../../queries/useReviewVerdictMutation';
import styles from './DecisionFooter.module.css';

interface DecisionFooterProps {
  decision: ArtifactDecisionModel | null;
  actions: ArtifactAction[];
  instance: ApprovalInstance;
  activeStage: StageInstance | undefined;
  onRefetchInstance: () => Promise<void> | void;
}

function variantClass(action: ArtifactAction): string {
  switch (action.variant) {
    case 'primary':
      return styles.actionBtnPrimary;
    case 'danger':
      return styles.actionBtnDanger;
    default:
      return styles.actionBtnSecondary;
  }
}

function OtherActions({ actions }: { actions: ArtifactAction[] }) {
  if (actions.length === 0) return null;
  return (
    <div className={styles.otherActions}>
      <h3 className={styles.otherActionsKicker}>Outras ações</h3>
      <div className={styles.actionsRow}>
        {actions.map((action) => (
          <button
            key={action.key}
            type="button"
            data-action={action.key}
            className={`${styles.actionBtn} ${variantClass(action)}`}
            disabled={!action.available}
            title={action.available ? undefined : action.reason}
            onClick={() => {
              void action.run();
            }}
          >
            {action.label}
          </button>
        ))}
      </div>
    </div>
  );
}

/** §6 (conservative default): a decision-derived meaning-of-signature line,
 *  added ABOVE the reused ArtifactDecisionPanel. Does NOT touch the existing
 *  MP 2.200-2 `legal.text` — see spec §6 for the flagged 21 CFR/MP jurisdiction
 *  question (surfaced to the operator at HS-1). */
function MeaningOfSignatureLine({ tone }: { tone: 'approve' | 'reject' }) {
  const text =
    tone === 'approve'
      ? 'Ao assinar, você declara aprovação deste documento.'
      : 'Ao assinar, você declara rejeição deste documento.';
  return <p className={styles.meaningLine}>{text}</p>;
}

/**
 * Shared request-changes verdict unit (R3): the single place that submits a
 * `request_changes` verdict on the active stage via `useReviewVerdictMutation`.
 * Used by both `ReviewModeFooter` (review-kind stages) and `ApprovalModeFooter`
 * (approval-kind stages, per G2 — the guard allows `request_changes`, never
 * `ready`, on approval stages).
 */
function useRequestChangesVerdict({
  instance,
  activeStage,
  onRefetchInstance,
}: {
  instance: ApprovalInstance;
  activeStage: StageInstance | undefined;
  onRefetchInstance: () => Promise<void> | void;
}) {
  const verdictMutation = useReviewVerdictMutation();
  const [open, setOpen] = useState(false);
  const [comment, setComment] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const canSubmit = activeStage != null;

  async function submit() {
    if (!activeStage) return;
    setSubmitting(true);
    setError(null);
    try {
      await verdictMutation.mutateAsync({
        instanceId: instance.id,
        stageId: activeStage.id,
        etag: instance.etag,
        body: { verdict: 'request_changes', comment: comment.trim() },
      });
      setOpen(false);
      setComment('');
      await onRefetchInstance();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Não foi possível registrar a decisão. Tente novamente.');
    } finally {
      setSubmitting(false);
    }
  }

  function reset() {
    setOpen(false);
    setComment('');
  }

  return { open, setOpen, comment, setComment, submitting, error, canSubmit, submit, reset };
}

type RequestChangesUnit = ReturnType<typeof useRequestChangesVerdict>;

/** Presentational comment dialog shared by both footer modes. */
function RequestChangesDialog({ unit }: { unit: RequestChangesUnit }) {
  return (
    <div className={styles.requestChangesDialog} role="dialog" aria-label="Solicitar mudanças">
      <label className={styles.field}>
        <span className={styles.fieldLabel}>Comentário · obrigatório</span>
        <textarea
          className={styles.textarea}
          rows={3}
          value={unit.comment}
          onChange={(e) => unit.setComment(e.target.value)}
          disabled={unit.submitting}
        />
      </label>
      <div className={styles.dialogActions}>
        <button
          type="button"
          className={`${styles.actionBtn} ${styles.actionBtnSecondary}`}
          disabled={unit.submitting}
          onClick={unit.reset}
        >
          Cancelar
        </button>
        <button
          type="button"
          className={`${styles.actionBtn} ${styles.actionBtnPrimary}`}
          disabled={unit.comment.trim().length === 0 || unit.submitting}
          onClick={() => void unit.submit()}
        >
          Enviar solicitação
        </button>
      </div>
    </div>
  );
}

/** Wraps ArtifactDecisionPanel (approve-only signoff ceremony) to surface the
 *  selected option's tone for the meaning-of-signature line, plus a secondary
 *  "Solicitar mudanças" comment-only verdict (R3: exactly two actions). */
function ApprovalModeFooter({
  decision,
  actions,
  instance,
  activeStage,
  onRefetchInstance,
}: {
  decision: ArtifactDecisionModel;
  actions: ArtifactAction[];
  instance: ApprovalInstance;
  activeStage: StageInstance | undefined;
  onRefetchInstance: () => Promise<void> | void;
}) {
  const [selectedTone, setSelectedTone] = useState<'approve' | 'reject' | null>(null);
  const requestChanges = useRequestChangesVerdict({ instance, activeStage, onRefetchInstance });

  return (
    <div
      className={styles.footer}
      data-testid="approval-sidebar-footer"
      style={{ position: 'sticky', bottom: 0 }}
      onClickCapture={(e) => {
        const target = e.target as HTMLElement;
        const optionBtn = target.closest('[data-option]');
        if (optionBtn) {
          const key = optionBtn.getAttribute('data-option');
          const option = decision.options.find((o) => o.key === key);
          if (option) setSelectedTone(option.tone);
        }
      }}
    >
      {selectedTone != null && <MeaningOfSignatureLine tone={selectedTone} />}
      <ArtifactDecisionPanel model={decision} />

      {!requestChanges.open ? (
        <button
          type="button"
          className={`${styles.actionBtn} ${styles.actionBtnSecondary}`}
          disabled={!requestChanges.canSubmit}
          onClick={() => requestChanges.setOpen(true)}
        >
          Solicitar mudanças
        </button>
      ) : (
        <RequestChangesDialog unit={requestChanges} />
      )}

      {requestChanges.error != null && (
        <div className={styles.error} role="alert">
          {requestChanges.error}
        </div>
      )}

      <OtherActions actions={actions} />
    </div>
  );
}

function ReviewModeFooter({
  instance,
  activeStage,
  actions,
  onRefetchInstance,
}: {
  instance: ApprovalInstance;
  activeStage: StageInstance | undefined;
  actions: ArtifactAction[];
  onRefetchInstance: () => Promise<void> | void;
}) {
  const verdictMutation = useReviewVerdictMutation();
  const requestChanges = useRequestChangesVerdict({ instance, activeStage, onRefetchInstance });
  const [submittingReady, setSubmittingReady] = useState(false);
  const [readyError, setReadyError] = useState<string | null>(null);

  const canSubmitVerdict = activeStage != null;

  async function submitReady() {
    if (!activeStage) return;
    setSubmittingReady(true);
    setReadyError(null);
    try {
      await verdictMutation.mutateAsync({
        instanceId: instance.id,
        stageId: activeStage.id,
        etag: instance.etag,
        body: { verdict: 'ready' },
      });
      await onRefetchInstance();
    } catch (err) {
      setReadyError(err instanceof Error ? err.message : 'Não foi possível registrar a decisão. Tente novamente.');
    } finally {
      setSubmittingReady(false);
    }
  }

  const submitting = submittingReady || requestChanges.submitting;
  const error = readyError ?? requestChanges.error;

  return (
    <div className={styles.footer} data-testid="approval-sidebar-footer" style={{ position: 'sticky', bottom: 0 }}>
      {!requestChanges.open ? (
        <div className={styles.verdictActions}>
          <button
            type="button"
            className={`${styles.actionBtn} ${styles.actionBtnPrimary}`}
            disabled={!canSubmitVerdict || submitting}
            onClick={() => void submitReady()}
          >
            Pronto para aprovação
          </button>
          <button
            type="button"
            className={`${styles.actionBtn} ${styles.actionBtnSecondary}`}
            disabled={!canSubmitVerdict || submitting}
            onClick={() => requestChanges.setOpen(true)}
          >
            Solicitar mudanças
          </button>
        </div>
      ) : (
        <RequestChangesDialog unit={requestChanges} />
      )}

      {error != null && (
        <div className={styles.error} role="alert">
          {error}
        </div>
      )}

      <OtherActions actions={actions} />
    </div>
  );
}

/**
 * Mode-aware sticky footer (spec §1.5, F2d.5 S1 three-way gate). Three variants:
 *  1. approving — `decision != null` (the adapter's actual signal that a signoff
 *     is currently offered; `policy.actions.signoff` is document-status-based,
 *     not stage_kind-based). Reuses `ArtifactDecisionPanel` verbatim (the tested
 *     password re-auth + legal + options flow).
 *  2. reviewing — `decision == null` and `deriveStageMode` (subject-agnostic,
 *     stage_kind × `instance.viewer.eligible_for_active_stage`) resolves to
 *     'reviewing' with an active stage present. Renders the verdict CTAs (no
 *     signoff button, no password).
 *  3. observing/lifecycle — everything else (no active stage, or an ineligible
 *     viewer on an active stage): no decision surface at all. Only `actions`
 *     (publish/cancel) render, in the "Outras ações" group; if there are none,
 *     the footer renders nothing.
 *
 * All variants that do render surface `actions` as a secondary group.
 */
export function DecisionFooter({
  decision,
  actions,
  instance,
  activeStage,
  onRefetchInstance,
}: DecisionFooterProps) {
  if (decision != null) {
    return (
      <ApprovalModeFooter
        decision={decision}
        actions={actions}
        instance={instance}
        activeStage={activeStage}
        onRefetchInstance={onRefetchInstance}
      />
    );
  }

  const stageMode = deriveStageMode(
    activeStage?.stage_kind as 'review' | 'approval' | undefined,
    instance.viewer?.eligible_for_active_stage ?? false,
  );

  if (stageMode === 'reviewing' && activeStage != null) {
    return (
      <ReviewModeFooter
        instance={instance}
        activeStage={activeStage}
        actions={actions}
        onRefetchInstance={onRefetchInstance}
      />
    );
  }

  // observing / lifecycle / ineligible: no decision surface — only lifecycle
  // actions if any.
  if (actions.length === 0) return null;
  return (
    <div className={styles.footer} data-testid="approval-sidebar-footer" style={{ position: 'sticky', bottom: 0 }}>
      <OtherActions actions={actions} />
    </div>
  );
}
