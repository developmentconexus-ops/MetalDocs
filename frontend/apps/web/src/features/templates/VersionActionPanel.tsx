import { useState, type CSSProperties } from 'react';
import {
  type VersionDTO,
  type NextDraftRef,
  approveVersion,
  reviewVersion,
} from './api/templates';
import { useAuthStore } from '../../store/auth.store';
import {
  canApprove,
  canPublish,
  canReview,
  type ActorContext,
  type VersionActionGate,
} from './lib/canActOnVersion';

type ActResult = { version: VersionDTO; nextDraft?: NextDraftRef | null };

type Props = {
  version: VersionDTO;
  onVersionUpdate: (v: VersionDTO, nextDraft?: NextDraftRef | null) => void;
};

const EMPTY_ACTOR: ActorContext = { roles: [], capabilities: [] };
const PUBLISHED_WITH_NEXT_DRAFT_MSG = 'Publicado. Nova versão de rascunho criada.';

function tooltipFor(gate: VersionActionGate): string | undefined {
  return gate.allowed ? undefined : gate.reason;
}

export function VersionActionPanel({ version, onVersionUpdate }: Props) {
  const [reason, setReason] = useState('');
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const user = useAuthStore((s) => s.user);
  const actor: ActorContext = user
    ? { roles: user.roles ?? [], capabilities: user.capabilities ?? [] }
    : EMPTY_ACTOR;

  async function act(fn: () => Promise<ActResult>, successMsg: string) {
    setBusy(true);
    setErr(null);
    setSuccess(null);
    try {
      const res = await fn();
      const msg = res.nextDraft ? PUBLISHED_WITH_NEXT_DRAFT_MSG : successMsg;
      setSuccess(msg);
      onVersionUpdate(res.version, res.nextDraft ?? null);
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  }

  const approveAct = (accept: boolean) =>
    approveVersion(version.template_id, version.version_number, accept, reason);
  const reviewAct = async (accept: boolean): Promise<ActResult> => ({
    version: await reviewVersion(version.template_id, version.version_number, accept, reason),
    nextDraft: null,
  });

  if (version.status === 'in_review') {
    const hasReviewer = version.pending_reviewer_role != null;
    const acceptLabel = hasReviewer ? 'Approve Review' : 'Publish';
    const acceptGate = hasReviewer ? canReview(version, actor) : canPublish(version, actor);
    const rejectGate = acceptGate;
    const acceptAction = hasReviewer
      ? () => act(() => reviewAct(true), 'Review approved')
      : () => act(() => approveAct(true), 'Published');
    const rejectAction = hasReviewer
      ? () => act(() => reviewAct(false), 'Review rejected')
      : () => act(() => approveAct(false), 'Rejected - back to draft');
    return (
      <div style={panelStyle}>
        <strong style={{ fontSize: '0.875rem' }}>{hasReviewer ? 'Reviewer actions' : 'Approver actions'}</strong>
        <textarea
          placeholder="Reason (optional)"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          rows={2}
          style={textareaStyle}
          disabled={busy}
        />
        {err && <div role="alert" style={{ color: '#dc2626', fontSize: '0.8125rem' }}>{err}</div>}
        {success && <div style={{ color: '#065f46', fontSize: '0.8125rem' }}>{success}</div>}
        <div style={{ display: 'flex', gap: 8 }}>
          <button
            onClick={acceptAction}
            disabled={busy || !acceptGate.allowed}
            title={tooltipFor(acceptGate)}
            style={{ ...btnStyle, background: hasReviewer ? '#16a34a' : '#2563eb', color: '#fff' }}
          >
            {acceptLabel}
          </button>
          <button
            onClick={rejectAction}
            disabled={busy || !rejectGate.allowed}
            title={tooltipFor(rejectGate)}
            style={{ ...btnStyle, background: '#fff', color: '#dc2626', border: '1px solid #dc2626' }}
          >
            Reject
          </button>
        </div>
      </div>
    );
  }

  if (version.status === 'approved') {
    const approveGate = canApprove(version, actor);
    return (
      <div style={panelStyle}>
        <strong style={{ fontSize: '0.875rem' }}>Approver actions</strong>
        <textarea
          placeholder="Reason (optional)"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          rows={2}
          style={textareaStyle}
          disabled={busy}
        />
        {err && <div role="alert" style={{ color: '#dc2626', fontSize: '0.8125rem' }}>{err}</div>}
        {success && <div style={{ color: '#065f46', fontSize: '0.8125rem' }}>{success}</div>}
        <div style={{ display: 'flex', gap: 8 }}>
          <button
            onClick={() => act(() => approveAct(true), 'Published')}
            disabled={busy || !approveGate.allowed}
            title={tooltipFor(approveGate)}
            style={{ ...btnStyle, background: '#2563eb', color: '#fff' }}
          >
            Publish
          </button>
          <button
            onClick={() => act(() => approveAct(false), 'Rejected - back to draft')}
            disabled={busy || !approveGate.allowed}
            title={tooltipFor(approveGate)}
            style={{ ...btnStyle, background: '#fff', color: '#dc2626', border: '1px solid #dc2626' }}
          >
            Reject
          </button>
        </div>
      </div>
    );
  }

  if (version.status === 'published') {
    return (
      <div style={{ ...panelStyle, background: '#f0fdf4', borderColor: '#bbf7d0' }}>
        <span style={{ color: '#166534', fontSize: '0.875rem' }}>This version is published</span>
      </div>
    );
  }

  return null;
}

const panelStyle: CSSProperties = {
  padding: '16px 20px',
  borderTop: '1px solid #e5e7eb',
  background: '#f9fafb',
  display: 'flex',
  flexDirection: 'column',
  gap: 10,
  flexShrink: 0,
};

const textareaStyle: CSSProperties = {
  padding: '6px 10px',
  border: '1px solid #d1d5db',
  borderRadius: 6,
  fontSize: '0.875rem',
  fontFamily: 'inherit',
  resize: 'vertical',
};

const btnStyle: CSSProperties = {
  padding: '7px 16px',
  border: 'none',
  borderRadius: 6,
  cursor: 'pointer',
  fontSize: '0.875rem',
  fontWeight: 500,
};
