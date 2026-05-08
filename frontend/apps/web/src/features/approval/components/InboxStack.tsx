import { useEffect } from 'react';
import type { RichInboxItem } from '../lib/mockInboxData';
import { InboxApprovalCard } from './InboxApprovalCard';
import styles from './InboxStack.module.css';

interface InboxStackProps {
  items: RichInboxItem[];
  selectedIdx: number;
  onSelect: (idx: number) => void;
  onNext: () => void;
  onPrev: () => void;
  isLoading?: boolean;
  isError?: boolean;
}

export function InboxStack({
  items,
  selectedIdx,
  onSelect,
  onNext,
  onPrev,
  isLoading,
  isError,
}: InboxStackProps) {
  const urgentCount = items.filter((i) => i.urgent).length;
  const selectedItem = items[selectedIdx];

  // Keyboard navigation: A=approve, D=return, ←/→=prev/next
  useEffect(() => {
    if (items.length === 0) return;

    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'ArrowLeft') {
        onPrev();
      } else if (e.key === 'ArrowRight') {
        onNext();
      } else if (e.key === 'a' || e.key === 'A') {
        // TODO [BACKLOG: caixa-aprovacao.md]: trigger approve flow
      } else if (e.key === 'd' || e.key === 'D') {
        // TODO [BACKLOG: caixa-aprovacao.md]: trigger return flow
      }
    }

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [items.length, onNext, onPrev]);

  function renderCardArea() {
    if (isLoading) {
      return <div className={styles.loading}>Carregando...</div>;
    }
    if (isError) {
      return (
        <div className={styles.error} role="alert">
          Erro ao carregar aprovações.
        </div>
      );
    }
    if (items.length === 0) {
      return <div className={styles.empty}>Nenhuma aprovação pendente.</div>;
    }
    return (
      <>
        <div className={styles.cardNav}>
          <span className={`${styles.cardCounter} mono`}>
            {String(selectedIdx + 1).padStart(2, '0')} / {String(items.length).padStart(2, '0')}
          </span>
          <span className={styles.spacer} />
          <button
            type="button"
            className="btn btn-sm"
            onClick={onPrev}
            disabled={selectedIdx === 0}
          >
            ← anterior
          </button>
          <button
            type="button"
            className="btn btn-sm"
            onClick={onNext}
            disabled={selectedIdx === items.length - 1}
          >
            próximo →
          </button>
        </div>

        <div className={styles.cardStack}>
          {selectedItem && <InboxApprovalCard item={selectedItem} />}
        </div>

        <div className={styles.keyboardHints}>
          <kbd>A</kbd> Aprovar · <kbd>D</kbd> Devolver · <kbd>←/→</kbd> Navegar
        </div>
      </>
    );
  }

  return (
    <div className={styles.root}>
      <aside className={styles.queueRail}>
        <div className={styles.queueHeader}>
          <div className="kicker">SUA FILA</div>
          <div className={styles.queueCount}>
            <span className={styles.queueCountNum}>{items.length}</span>
            <span className="caption">decisões pendentes</span>
          </div>
          {urgentCount > 0 && (
            <div className="caption">
              <span className={styles.urgentToday}>{urgentCount} vencem hoje</span>
            </div>
          )}
        </div>

        {items.map((item, idx) => (
          <button
            key={item.instance_id}
            type="button"
            className={`${styles.queueItem}${idx === selectedIdx ? ` ${styles.queueItemActive}` : ''}`}
            onClick={() => onSelect(idx)}
          >
            <div className={styles.queueItemNumber}>{String(idx + 1).padStart(2, '0')}</div>
            <div className={styles.queueItemMeta}>
              <div className={styles.queueItemTop}>
                <span className={`${styles.queueItemCode} mono`}>{item.code}</span>
                {item.urgent && <span className={styles.urgentDot} />}
              </div>
              <div className={styles.queueItemTitle}>{item.document_title}</div>
              <div className={styles.queueItemSub}>
                <span className={styles.queueItemDeadline}>{item.deadline}</span>
                <span className={styles.dot}>·</span>
                <span>{item.area_code}</span>
              </div>
            </div>
          </button>
        ))}
      </aside>

      <main className={styles.cardArea}>
        {renderCardArea()}
      </main>
    </div>
  );
}
