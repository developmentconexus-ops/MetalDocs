import { useState } from 'react';

import type { ApprovalInstance, Signoff, StageInstance } from '../../api/approvalTypes';
import { formatDateTime as formatDateTimeValue } from '../../../../lib/formatDate';
import { useHasCapability } from '../../../../lib/iam/useHasCapability';
import { useExtendSLA } from '../../queries/useExtendSLA';
import { ExtendSlaDialog } from '../ExtendSlaDialog';
import styles from './ApprovalTimeline.module.css';

// Mirrors the Go capability registry `internal/modules/iam/domain/model.go`
// (CapApprovalSLAExtend). Gates the "Prorrogar prazo" action the same way
// `useHasCapability('document.edit')` gates the cancel-instance action in
// DocumentWorkspacePage — client-side UX gate only, tier-2 authz.Require
// stays authoritative (ADR 0022).
const SLA_EXTEND_CAPABILITY = 'approval.sla_extend';

// The extend-SLA dialog needs a real due_at to display/anchor the extension
// against (no-fallback: a nullable deadline means *no deadline*, never a
// synthesized default — see the `?? ''` finding this type exists to close).
// Narrowing `extendingStage`'s `due_at` to `string` here makes that guarantee
// structural: the only place this state is set (below) is required to supply
// a non-null value, so `ExtendSlaDialog` can never receive a stand-in.
type ExtendableStage = StageInstance & { due_at: string };

interface ApprovalTimelineProps {
  instance: ApprovalInstance | null;
  loading: boolean;
  error?: string | null;
  onRetry?: () => void;
}

const STAGE_STATUS_LABEL: Record<StageInstance['status'], string> = {
  pending: 'Pendente',
  active: 'Em andamento',
  passed: 'Aprovado',
  failed: 'Reprovado',
  cancelled: 'Cancelado',
};

const DECISION_LABEL: Record<Signoff['decision'], string> = {
  approve: 'Aprovou',
  reject: 'Rejeitou',
};

// F4: the label for what this signature legally attests (spec §1.2). Distinct
// from DECISION_LABEL — signature_meaning is the derived, server-controlled
// legal attestation (never client-writable), rendered alongside the decision.
const SIGNATURE_MEANING_LABEL: Record<Signoff['signature_meaning'], string> = {
  approval: 'Aprovação',
  rejection: 'Rejeição',
};

// Runtime instance status is `in_progress | approved | rejected | cancelled`
// (see mapInstanceResponse in doc_approval_handler.go). There is no
// `completed` value — CON-10 fixed a prior drift where this component
// checked for a status the backend never emits, silently hiding the
// "Status final" timeline node for every approved/rejected instance.
const FINAL_STATUS_LABEL: Partial<Record<ApprovalInstance['status'], string>> = {
  approved: 'Aprovado',
  rejected: 'Rejeitado',
  cancelled: 'Cancelado',
};

function formatDateTime(iso?: string | null): string {
  if (!iso) {
    return '-';
  }
  return formatDateTimeValue(iso);
}

function formatSignatureMethod(signatureMethod: Signoff['signature_method']): string {
  if (signatureMethod === 'password_reauth') {
    return 'Reautenticação por senha';
  }
  return 'ICP-Brasil';
}

/**
 * The SINGLE approval timeline rendered in the cockpit sidebar (spec §1.2).
 * Forked from the former `ApprovalTimelinePanel` (deleted — its only consumer,
 * the former sidebar-extras component, was deleted too) with one addition:
 * each signoff now shows its `signature_meaning` label ("Aprovação"/"Rejeição")
 * — the derived, server-controlled legal attestation distinct from the decision.
 */
export function ApprovalTimeline({ instance, loading, error, onRetry }: ApprovalTimelineProps) {
  const canExtendSla = useHasCapability(SLA_EXTEND_CAPABILITY);
  const extendMutation = useExtendSLA();
  const [extendingStage, setExtendingStage] = useState<ExtendableStage | null>(null);

  if (loading) {
    return <div className={styles.state}>Carregando timeline...</div>;
  }

  if (error) {
    return (
      <div className={styles.state} role="alert">
        <p>{error}</p>
        <button type="button" onClick={onRetry} disabled={!onRetry}>
          Tentar novamente
        </button>
      </div>
    );
  }

  if (!instance) {
    return <div className={styles.state}>Nenhum evento de aprovação registrado.</div>;
  }

  return (
    <section className={styles.panel} aria-label="Timeline de aprovação">
      <ol className={styles.timeline}>
        <li className={styles.node}>
          <div className={styles.dot} aria-hidden="true" />
          <div className={styles.content}>
            <h3 className={styles.title}>Submetido</h3>
            <p className={styles.meta}>
              Por <strong>{instance.submitted_by}</strong> em {formatDateTime(instance.submitted_at)}
            </p>
          </div>
        </li>

        {(instance.stages ?? [])
          .slice()
          .sort((a, b) => a.stage_index - b.stage_index)
          .map((stage) => (
            <li className={styles.node} key={stage.id}>
              <div className={styles.dot} aria-hidden="true" />
              <div className={styles.content}>
                <div className={styles.stageHeader}>
                  <h3 className={styles.title}>{stage.label}</h3>
                  <span className={`${styles.stageStatus} ${styles[`stage_${stage.status}`]}`}>
                    {STAGE_STATUS_LABEL[stage.status]}
                  </span>
                </div>
                {/* F8/Task 8 no-fallback: due_at is null while pending or when the
                    stage has no SLA configured — never a synthesized default.
                    `dueAt` is captured as a local const so the ternary's
                    narrowing to `string` survives into the button's closure
                    (see ExtendableStage above). */}
                {(() => {
                  const dueAt = stage.due_at;
                  return dueAt ? (
                    <div className={styles.dueRow}>
                      <p className={styles.meta}>Prazo: {formatDateTime(dueAt)}</p>
                      {stage.status === 'active' && canExtendSla ? (
                        <button
                          type="button"
                          className={styles.extendButton}
                          onClick={() => setExtendingStage({ ...stage, due_at: dueAt })}
                        >
                          Prorrogar prazo
                        </button>
                      ) : null}
                    </div>
                  ) : null;
                })()}
                {(stage.signoffs ?? []).length === 0 ? (
                  <p className={styles.meta}>Sem assinaturas registradas.</p>
                ) : (
                  <ul className={styles.signoffs}>
                    {(stage.signoffs ?? []).map((signoff) => (
                      <li className={styles.signoff} key={signoff.id}>
                        <p>
                          <strong>{signoff.actor_user_id}</strong> - {DECISION_LABEL[signoff.decision]}
                          {' · '}
                          <span className={styles.meaning}>
                            {SIGNATURE_MEANING_LABEL[signoff.signature_meaning]}
                          </span>
                        </p>
                        <p className={styles.meta}>
                          Assinatura: {formatSignatureMethod(signoff.signature_method)} | Em:{' '}
                          {formatDateTime(signoff.signed_at)}
                        </p>
                        {signoff.reason ? <p className={styles.meta}>Motivo: {signoff.reason}</p> : null}
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            </li>
          ))}

        {FINAL_STATUS_LABEL[instance.status] ? (
          <li className={styles.node}>
            <div className={styles.dot} aria-hidden="true" />
            <div className={styles.content}>
              <h3 className={styles.title}>Status final</h3>
              <p className={styles.meta}>
                {FINAL_STATUS_LABEL[instance.status]} em {formatDateTime(instance.completed_at)}
              </p>
            </div>
          </li>
        ) : null}
      </ol>

      {extendingStage ? (
        <ExtendSlaDialog
          open
          stageLabel={extendingStage.label}
          currentDueAt={extendingStage.due_at}
          isSubmitting={extendMutation.isPending}
          onConfirm={async (dueAt, reason) => {
            await extendMutation.mutateAsync({ instanceId: instance.id, dueAt, reason });
            setExtendingStage(null);
          }}
          onClose={() => setExtendingStage(null)}
        />
      ) : null}
    </section>
  );
}
