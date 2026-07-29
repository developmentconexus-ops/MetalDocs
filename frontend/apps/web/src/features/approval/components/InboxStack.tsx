import { useEffect } from 'react';
import type { InboxItem } from '../api/approvalTypes';
import { formatDueRelative } from '../../../lib/format/dates';
import { InboxApprovalCard } from './InboxApprovalCard';
import styles from './InboxStack.module.css';

interface InboxStackProps {
  items: InboxItem[];
  selectedIdx: number;
  onSelect: (idx: number) => void;
  onNext: () => void;
  onPrev: () => void;
  isLoading?: boolean;
  isError?: boolean;
  // F5 (M2c C3): when a filter is active and yields no items, the empty state
  // must teach "no match for filters" instead of "no work at all" — otherwise
  // a filtered-empty result reads as "you have nothing to do".
  isFiltered?: boolean;
  onOpenDocument?: (item: InboxItem) => void;
  onApprove?: (item: InboxItem) => void;
  onReject?: (item: InboxItem) => void;
}

const NO_WORK_EMPTY =
  'Nenhuma aprovação pendente. Documentos submetidos a rotas onde você é revisor ou aprovador aparecem aqui.';
const FILTERED_EMPTY = 'Nenhuma aprovação corresponde aos filtros.';

export function InboxStack({
  items,
  selectedIdx,
  onSelect,
  onNext,
  onPrev,
  isLoading,
  isError,
  isFiltered,
  onOpenDocument,
  onApprove,
  onReject,
}: InboxStackProps) {
  const todayStart = new Date();
  todayStart.setHours(0, 0, 0, 0);
  const todayCount = items.filter((item) => new Date(item.submitted_at).getTime() >= todayStart.getTime()).length;
  const selectedItem = items[selectedIdx];

  // Keyboard navigation: ←/→=prev/next.
  // A/D approve-return shortcuts were removed (FE-19 N8) — the handlers were
  // empty TODOs advertised via keyboardHints below with no working listener.
  // Approve/reject remain available via the card buttons (onApprove/onReject).
  useEffect(() => {
    if (items.length === 0) return;

    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'ArrowLeft') {
        onPrev();
      } else if (e.key === 'ArrowRight') {
        onNext();
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
      return <div className={styles.empty}>{isFiltered ? FILTERED_EMPTY : NO_WORK_EMPTY}</div>;
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
          {selectedItem && (
            <InboxApprovalCard
              item={selectedItem}
              onOpenDocument={onOpenDocument ? () => onOpenDocument(selectedItem) : undefined}
              onApprove={onApprove ? () => onApprove(selectedItem) : undefined}
              onReject={onReject ? () => onReject(selectedItem) : undefined}
            />
          )}
        </div>

        <div className={styles.keyboardHints}>
          <kbd>←/→</kbd> Navegar
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
          {todayCount > 0 && (
            <div className="caption">
              <span className={styles.urgentToday}>{todayCount} enviados hoje</span>
            </div>
          )}
        </div>

        {items.map((item, idx) => {
          const due = formatDueRelative(item.due_at);
          return (
            <button
              key={item.instance_id}
              type="button"
              className={`${styles.queueItem}${idx === selectedIdx ? ` ${styles.queueItemActive}` : ''}`}
              onClick={() => onSelect(idx)}
            >
              <div className={styles.queueItemNumber}>{String(idx + 1).padStart(2, '0')}</div>
                <div className={styles.queueItemMeta}>
                  <div className={styles.queueItemTop}>
                    {/* F-QA4-8: canonical human code; uuid only as truthful fallback. */}
                    <span className={`${styles.queueItemCode} mono`}>
                      {item.subject_kind === 'document'
                        ? (item.controlled_document_code ?? item.controlled_document_id)
                        : item.subject_ref}
                    </span>
                  </div>
                  <div className={styles.queueItemTitle}>{item.subject_title}</div>
                  <div className={styles.queueItemSub}>
                    <span className={styles.queueItemDeadline}>
                      {new Date(item.submitted_at).toLocaleDateString('pt-BR')}
                    </span>
                    <span className={styles.dot}>·</span>
                    <span>{item.area_code}</span>
                    <span className={styles.dot}>·</span>
                    <span className={due.overdue ? styles.dueOverdue : undefined}>{due.label}</span>
                  </div>
              </div>
            </button>
          );
        })}
      </aside>

      <main className={styles.cardArea}>
        {renderCardArea()}
      </main>
    </div>
  );
}
