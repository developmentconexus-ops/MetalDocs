import { useState } from 'react';
import { MOCK_INBOX_ITEMS, enrichInboxItem } from '../lib/mockInboxData';
import { useInboxQuery } from '../queries/useInboxQuery';
import { InboxStack } from '../components/InboxStack';
import { InboxTimeline } from '../components/InboxTimeline';
import { InboxToolbar } from '../components/InboxToolbar';
import styles from './InboxPage.module.css';

export function InboxPage() {
  const [view, setView] = useState<string>(
    () => localStorage.getItem('md.inbox.v') || 'stack',
  );
  const [selectedIdx, setSelectedIdx] = useState(0);

  const { data, isLoading, isError } = useInboxQuery();

  // Enrich API items with mock extras; fall back to canonical mock set on error/empty.
  // TODO [BACKLOG: caixa-aprovacao.md]: Remove fallback when API returns all required fields.
  const items =
    data && data.items.length > 0
      ? data.items.map((item, idx) => enrichInboxItem(item, idx))
      : isLoading || isError
        ? []
        : MOCK_INBOX_ITEMS;

  function handleViewChange(v: string) {
    setView(v);
    localStorage.setItem('md.inbox.v', v);
  }

  function handleSelect(idx: number) {
    setSelectedIdx(idx);
  }

  function handleNext() {
    setSelectedIdx((prev) => Math.min(prev + 1, items.length - 1));
  }

  function handlePrev() {
    setSelectedIdx((prev) => Math.max(prev - 1, 0));
  }

  return (
    <div className={styles.page}>
      <InboxToolbar view={view} onViewChange={handleViewChange} />
      {view === 'timeline' ? (
        <InboxTimeline items={items} />
      ) : (
        <InboxStack
          items={items}
          selectedIdx={selectedIdx}
          onSelect={handleSelect}
          onNext={handleNext}
          onPrev={handlePrev}
          isLoading={isLoading}
          isError={isError}
        />
      )}
    </div>
  );
}
