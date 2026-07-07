import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { fetchActiveDocumentInstance } from '../../controlled-documents/api/controlledDocuments';
import type { InboxItem } from '../api/approvalTypes';
import { useInboxQuery } from '../queries/useInboxQuery';
import { InboxStack } from '../components/InboxStack';
import { InboxTimeline } from '../components/InboxTimeline';
import { InboxToolbar } from '../components/InboxToolbar';
import {
  DEFAULT_INBOX_FILTER_STATE,
  InboxFilters,
  isInboxFilterActive,
  toInboxParams,
  type InboxFilterState,
} from '../components/InboxFilters';
import { sortByDueAsc } from '../../../lib/inbox/sortByDue';
import { formatDueRelative } from '../../../lib/format/dates';
import { ApiError } from '../../../lib/api/errors';
import styles from './InboxPage.module.css';

type ViewType = 'stack' | 'timeline';

function readStoredView(): ViewType {
  const raw = localStorage.getItem('md.inbox.v');
  return raw === 'stack' || raw === 'timeline' ? raw : 'stack';
}

export function getNextSelectedIdx(prev: number, totalItems: number) {
  return Math.min(prev + 1, Math.max(totalItems - 1, 0));
}

// F5 (M2c C3): client-side belt+braces guard alongside the server-side
// due_before filter. due_before is inclusive of "due later today" items, so
// the "Atrasadas" bucket needs this extra check to exclude not-yet-overdue
// items from leaking in.
function isOverdue(item: InboxItem, now: number): boolean {
  return formatDueRelative(item.due_at, now).overdue;
}

export function InboxPage() {
  const navigate = useNavigate();
  const [view, setView] = useState<ViewType>(() => readStoredView());
  const [selectedIdx, setSelectedIdx] = useState(0);
  const [actionError, setActionError] = useState<string | null>(null);
  const [filters, setFilters] = useState<InboxFilterState>(DEFAULT_INBOX_FILTER_STATE);
  const [filtersOpen, setFiltersOpen] = useState(true);
  const [overseeDenied, setOverseeDenied] = useState(false);

  const params = useMemo(() => toInboxParams(filters, Date.now()), [filters]);
  const { data, isLoading, isError, error } = useInboxQuery(params);

  // Reactive oversee 403: if the oversee-scoped query fails with 403, surface
  // an inline note and revert the toggle once. No preemptive capability probe
  // (spec D2) — this is intentionally the only place oversee gating happens.
  useEffect(() => {
    if (filters.oversee && isError && error instanceof ApiError && error.status === 403) {
      setOverseeDenied(true);
      setFilters((prev) => ({ ...prev, oversee: false }));
    }
  }, [filters.oversee, isError, error]);

  const rawItems = data?.items ?? [];
  const items = useMemo(() => {
    const now = Date.now();
    const filtered = filters.due === 'overdue' ? rawItems.filter((item) => isOverdue(item, now)) : rawItems;
    return sortByDueAsc(filtered);
  }, [rawItems, filters.due]);
  const isFiltered = isInboxFilterActive(filters);

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

  // F5 (M2c C3): single destination — the primary worklist open goes straight
  // to the approval cockpit, which resolves review/readonly/oversee mode from
  // document_id alone (F3/F4). This replaces the prior
  // fetchActiveDocumentInstance-based author-editor route, which sent
  // reviewers/approvers into the writable editor (the W2 vector F3 already
  // killed at the cockpit).
  function openDocument(item: InboxItem) {
    setActionError(null);
    navigate(`/approvals/${item.document_id}`);
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
      <InboxToolbar
        view={view}
        onViewChange={handleViewChange}
        filtersOpen={filtersOpen}
        onToggleFilters={() => setFiltersOpen((prev) => !prev)}
      />
      {filtersOpen ? <InboxFilters value={filters} onChange={setFilters} oversee403={overseeDenied} /> : null}
      {actionError ? (
        <div className={styles.actionError} role="alert">
          {actionError}
        </div>
      ) : null}
      {view === 'timeline' ? (
        <InboxTimeline
          items={items}
          isLoading={isLoading}
          isError={isError}
          isFiltered={isFiltered}
          onOpenDocument={openDocument}
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
          isFiltered={isFiltered}
          onOpenDocument={openDocument}
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
