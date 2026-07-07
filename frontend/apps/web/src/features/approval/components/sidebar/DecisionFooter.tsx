import { useState } from 'react';

import { ArtifactDecisionPanel } from '../../../shared/controlled-artifact/ArtifactDecisionPanel';
import type { ArtifactAction, ArtifactDecisionModel } from '../../../shared/controlled-artifact/types';
import type { ApprovalInstance, StageInstance } from '../../api/approvalTypes';
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

/** Wraps ArtifactDecisionPanel to surface the selected option's tone for the
 *  meaning-of-signature line without altering the panel itself. */
function ApprovalModeFooter({ decision, actions }: { decision: ArtifactDecisionModel; actions: ArtifactAction[] }) {
  const [selectedTone, setSelectedTone] = useState<'approve' | 'reject' | null>(
    decision.options.find((o) => o.key === decision.defaultOptionKey)?.tone ?? null,
  );

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
  const [showRequestChanges, setShowRequestChanges] = useState(false);
  const [comment, setComment] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const canSubmitVerdict = activeStage != null;

  async function submitVerdict(verdict: 'ready' | 'request_changes', commentValue?: string) {
    if (!activeStage) return;
    setSubmitting(true);
    setError(null);
    try {
      await verdictMutation.mutateAsync({
        instanceId: instance.id,
        stageId: activeStage.id,
        etag: instance.etag,
        body: commentValue != null ? { verdict, comment: commentValue } : { verdict },
      });
      setShowRequestChanges(false);
      setComment('');
      await onRefetchInstance();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Não foi possível registrar a decisão. Tente novamente.');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className={styles.footer} data-testid="approval-sidebar-footer" style={{ position: 'sticky', bottom: 0 }}>
      {!showRequestChanges ? (
        <div className={styles.verdictActions}>
          <button
            type="button"
            className={`${styles.actionBtn} ${styles.actionBtnPrimary}`}
            disabled={!canSubmitVerdict || submitting}
            onClick={() => void submitVerdict('ready')}
          >
            Pronto para aprovação
          </button>
          <button
            type="button"
            className={`${styles.actionBtn} ${styles.actionBtnSecondary}`}
            disabled={!canSubmitVerdict || submitting}
            onClick={() => setShowRequestChanges(true)}
          >
            Solicitar mudanças
          </button>
        </div>
      ) : (
        <div className={styles.requestChangesDialog} role="dialog" aria-label="Solicitar mudanças">
          <label className={styles.field}>
            <span className={styles.fieldLabel}>Comentário · obrigatório</span>
            <textarea
              className={styles.textarea}
              rows={3}
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              disabled={submitting}
            />
          </label>
          <div className={styles.dialogActions}>
            <button
              type="button"
              className={`${styles.actionBtn} ${styles.actionBtnSecondary}`}
              disabled={submitting}
              onClick={() => {
                setShowRequestChanges(false);
                setComment('');
              }}
            >
              Cancelar
            </button>
            <button
              type="button"
              className={`${styles.actionBtn} ${styles.actionBtnPrimary}`}
              disabled={comment.trim().length === 0 || submitting}
              onClick={() => void submitVerdict('request_changes', comment.trim())}
            >
              Enviar solicitação
            </button>
          </div>
        </div>
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
 * Mode-aware sticky footer (spec §1.5). Review mode renders the verdict CTAs
 * (no signoff button, no password); approval mode reuses `ArtifactDecisionPanel`
 * verbatim (the tested password re-auth + legal + options flow). Both variants
 * render `actions` (publish/cancel) as a secondary "Outras ações" group.
 *
 * The branch is keyed off `decision != null` — the adapter's actual signal
 * that a signoff is currently offered (`useDocumentApprovalArtifact`'s
 * `policy.actions.signoff` gate is document-status-based, not stage_kind-based;
 * `isApprovalStage` labels the stage but is not itself the offer condition).
 * When decision is present the cockpit is in the legal e-signature flow
 * regardless of `isApprovalStage`; otherwise it falls back to the review-verdict
 * CTAs (which require an active stage to submit against).
 */
export function DecisionFooter({
  decision,
  actions,
  instance,
  activeStage,
  onRefetchInstance,
}: DecisionFooterProps) {
  if (decision != null) {
    return <ApprovalModeFooter decision={decision} actions={actions} />;
  }

  return (
    <ReviewModeFooter
      instance={instance}
      activeStage={activeStage}
      actions={actions}
      onRefetchInstance={onRefetchInstance}
    />
  );
}
