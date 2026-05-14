import { useEffect, useState } from 'react';
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
  const [view, setView] = useState<ViewType>(() => readStoredView());
  const [selectedIdx, setSelectedIdx] = useState(0);

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

  useEffect(() => {
    if (selectedIdx > 0 && selectedIdx >= items.length) {
      setSelectedIdx(Math.max(items.length - 1, 0));
    }
  }, [items.length, selectedIdx]);

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
