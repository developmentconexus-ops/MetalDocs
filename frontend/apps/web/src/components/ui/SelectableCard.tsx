import { forwardRef } from 'react';
import type { KeyboardEventHandler, ReactNode } from 'react';
import styles from './SelectableCard.module.css';

export type SelectableCardProps = {
  selected: boolean;
  onSelect: () => void;
  children: ReactNode;
  disabled?: boolean;
  className?: string;
  title?: string;
  ariaLabel?: string;
  tabIndex?: number;
  onKeyDown?: KeyboardEventHandler<HTMLButtonElement>;
};

export const SelectableCard = forwardRef<HTMLButtonElement, SelectableCardProps>(function SelectableCard(
  {
    selected,
    onSelect,
    children,
    disabled = false,
    className,
    title,
    ariaLabel,
    tabIndex,
    onKeyDown,
  },
  ref,
): JSX.Element {
  const stateClass = disabled
    ? styles.disabled
    : selected
      ? styles.selected
      : styles.idle;
  const cls = [styles.root, stateClass, className].filter(Boolean).join(' ');

  return (
    <button
      ref={ref}
      type="button"
      role="radio"
      aria-checked={selected}
      aria-disabled={disabled || undefined}
      disabled={disabled}
      title={title}
      aria-label={ariaLabel}
      tabIndex={tabIndex}
      onKeyDown={onKeyDown}
      className={cls}
      data-selected={selected ? 'true' : undefined}
      onClick={disabled ? undefined : onSelect}
    >
      {children}
    </button>
  );
});
