import { Icon } from '../../../components/ui/Icon';
import styles from './InboxToolbar.module.css';

type Props = {
  view: 'stack' | 'timeline';
  onViewChange: (v: 'stack' | 'timeline') => void;
};

const StackIcon = () => (
  <svg width="14" height="14" viewBox="0 0 16 16" fill="none" aria-hidden="true">
    <rect x="3" y="2" width="10" height="11" rx="1.5" stroke="currentColor" strokeWidth="1.4" />
    <line x1="3" y1="6" x2="13" y2="6" stroke="currentColor" strokeWidth="1.4" />
    <rect x="4.5" y="14" width="7" height="0.5" fill="currentColor" opacity="0.5" />
  </svg>
);

const TimelineIcon = () => (
  <svg width="14" height="14" viewBox="0 0 16 16" fill="none" aria-hidden="true">
    <line x1="2" y1="8" x2="14" y2="8" stroke="currentColor" strokeWidth="1.4" />
    <circle cx="4" cy="8" r="1.6" fill="currentColor" />
    <circle cx="8" cy="8" r="1.6" fill="currentColor" />
    <circle cx="12" cy="8" r="1.6" fill="currentColor" opacity="0.5" />
  </svg>
);

export function InboxToolbar({ view, onViewChange }: Props) {
  return (
    <div className={styles.toolbar}>
      <span className={`${styles.kicker} kicker`}>APROVAÇÕES</span>
      <span className={styles.breadcrumbSep}>›</span>
      <span className={styles.breadcrumbTitle}>Caixa de entrada</span>
      <span className={styles.spacer} />
      <button type="button" className={`${styles.filtersBtn} btn btn-sm`} disabled aria-disabled="true" title="Em breve">
        <Icon name="filter" size={13} /> Filtros
      </button>
      <div className={styles.viewSwitcher}>
        <button
          type="button"
          className={styles.viewSwitcherBtn}
          data-active={view === 'stack'}
          onClick={() => onViewChange('stack')}
        >
          <StackIcon /> Foco
        </button>
        <button
          type="button"
          className={styles.viewSwitcherBtn}
          data-active={view === 'timeline'}
          onClick={() => onViewChange('timeline')}
        >
          <TimelineIcon /> Linha do tempo
        </button>
      </div>
    </div>
  );
}
