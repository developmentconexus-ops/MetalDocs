import { useCallback, useRef } from 'react';
import type { KeyboardEventHandler, RefCallback } from 'react';

export type RovingRadioGroupConfig = {
  count: number;
  selectedIndex: number;
  onSelect: (newIndex: number) => void;
  orientation?: 'horizontal' | 'vertical' | 'both';
  ariaLabel: string;
};

export type RovingRadioGroupItemProps = {
  ref: RefCallback<HTMLElement>;
  tabIndex: number;
  onKeyDown: KeyboardEventHandler<HTMLElement>;
};

export type RovingRadioGroupResult = {
  groupProps: {
    role: 'radiogroup';
    'aria-label': string;
  };
  getItemProps: (index: number) => RovingRadioGroupItemProps;
};

export function useRovingRadioGroup({
  count,
  selectedIndex,
  onSelect,
  orientation = 'both',
  ariaLabel,
}: RovingRadioGroupConfig): RovingRadioGroupResult {
  const refs = useRef<Array<HTMLElement | null>>([]);

  const onKeyDown = useCallback<KeyboardEventHandler<HTMLElement>>(
    (event) => {
      if (count <= 0) return;

      const isHorizontal = orientation === 'horizontal' || orientation === 'both';
      const isVertical = orientation === 'vertical' || orientation === 'both';

      let nextIndex: number | null = null;
      const currentIndex = selectedIndex >= 0 ? selectedIndex : 0;

      if (event.key === 'ArrowRight' && isHorizontal) {
        nextIndex = (currentIndex + 1) % count;
      } else if (event.key === 'ArrowDown' && isVertical) {
        nextIndex = (currentIndex + 1) % count;
      } else if (event.key === 'ArrowLeft' && isHorizontal) {
        nextIndex = (currentIndex - 1 + count) % count;
      } else if (event.key === 'ArrowUp' && isVertical) {
        nextIndex = (currentIndex - 1 + count) % count;
      } else if (event.key === 'Home') {
        nextIndex = 0;
      } else if (event.key === 'End') {
        nextIndex = count - 1;
      }

      if (nextIndex === null) return;

      event.preventDefault();
      onSelect(nextIndex);
      refs.current[nextIndex]?.focus();
    },
    [count, onSelect, orientation, selectedIndex],
  );

  return {
    groupProps: {
      role: 'radiogroup',
      'aria-label': ariaLabel,
    },
    getItemProps: (index) => ({
      ref: (node) => {
        refs.current[index] = node;
      },
      tabIndex:
        index === selectedIndex || (selectedIndex === -1 && index === 0)
          ? 0
          : -1,
      onKeyDown,
    }),
  };
}
