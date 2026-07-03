import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { fetchActiveDocumentInstance } from '../../controlled-documents/api/controlledDocuments';
import type { InboxItem } from '../api/approvalTypes';
import { useInboxQuery } from '../queries/useInboxQuery';
import { InboxStack } from '../components/InboxStack';
import { InboxTimeline } from '../components/InboxTimeline';
import { InboxToolbar } from '../components/InboxToolbar';
import styles from './InboxPage.module.css';

type ViewType = 'stack' | 'timeline';

function readStoredView(): ViewType {
  const raw = localStorage.getItem('md.inbox.v');
  return raw === 'stack' || raw === 'timeline' ? raw : 'stack';
}

export function getNextSelectedIdx(prev: number, totalItems: number) {
  return Math.min(prev + 1, Math.max(totalItems - 1, 0));
}

export function InboxPage() {
  const navigate = useNavigate();
  const [view, setView] = useState<ViewType>(() => readStoredView());
  const [selectedIdx, setSelectedIdx] = useState(0);
  const [actionError, setActionError] = useState<string | null>(null);

  const { data, isLoading, isError } = useInboxQuery();
  const items = data?.items ?? [];

  function handleViewChange(v: ViewType) {
    setView(v);
    localStorage.setItem('md.inbox.v', v);
  }

  function handleSelect(idx: number) {
    setSelectedIdx(idx);
  }

  function handleNext() {
    setSelectedIdx((prev) => getNextSelectedIdx(prev, items.length));
  }

  function handlePrev() {
    setSelectedIdx((prev) => Math.max(prev - 1, 0));
  }

  async function openDocument(item: InboxItem) {
    setActionError(null);
    try {
      const active = await fetchActiveDocumentInstance(item.controlled_document_id);
      if (!active?.document_id) {
        setActionError('Documento indisponivel no editor moderno no momento.');
        return;
      }
      navigate(`/documents/${active.document_id}/edit`);
    } catch {
      setActionError('Documento indisponivel no editor moderno no momento.');
    }
  }

  async function openDecisionFlow(item: InboxItem, decision: 'approve' | 'reject') {
    setActionError(null);
    try {
      const active = await fetchActiveDocumentInstance(item.controlled_document_id);
      if (!active?.document_id) {
        setActionError('Fluxo de aprovação indisponível para este documento no momento.');
        return;
      }
      navigate(`/approvals/${active.document_id}?decision=${decision}`);
    } catch {
      setActionError('Fluxo de aprovação indisponível para este documento no momento.');
    }
  }

  useEffect(() => {
    if (selectedIdx > 0 && selectedIdx >= items.length) {
      setSelectedIdx(Math.max(items.length - 1, 0));
    }
  }, [items.length, selectedIdx]);

  return (
    <div className={styles.page}>
      <InboxToolbar view={view} onViewChange={handleViewChange} />
      {actionError ? (
        <div className={styles.actionError} role="alert">
          {actionError}
        </div>
      ) : null}
      {view === 'timeline' ? (
        <InboxTimeline
          items={items}
          onOpenDocument={(item) => {
            void openDocument(item);
          }}
        />
      ) : (
        <InboxStack
          items={items}
          selectedIdx={selectedIdx}
          onSelect={handleSelect}
          onNext={handleNext}
          onPrev={handlePrev}
          isLoading={isLoading}
          isError={isError}
          onOpenDocument={(item) => {
            void openDocument(item);
          }}
          onApprove={(item) => {
            void openDecisionFlow(item, 'approve');
          }}
          onReject={(item) => {
            void openDecisionFlow(item, 'reject');
          }}
        />
      )}
    </div>
  );
}
