import type { ReactNode } from 'react';
import styles from './SelectableCard.module.css';

export type SelectableCardProps = {
  selected: boolean;
  onSelect: () => void;
  children: ReactNode;
  disabled?: boolean;
  className?: string;
  title?: string;
  ariaLabel?: string;
};

export function SelectableCard({
  selected,
  onSelect,
  children,
  disabled = false,
  className,
  title,
  ariaLabel,
}: SelectableCardProps): JSX.Element {
  const stateClass = disabled
    ? styles.disabled
    : selected
      ? styles.selected
      : styles.idle;
  const cls = [styles.root, stateClass, className].filter(Boolean).join(' ');

  return (
    <button
      type="button"
      role="radio"
      aria-checked={selected}
      aria-disabled={disabled || undefined}
      disabled={disabled}
      title={title}
      aria-label={ariaLabel}
      className={cls}
      onClick={disabled ? undefined : onSelect}
    >
      {children}
    </button>
  );
}
