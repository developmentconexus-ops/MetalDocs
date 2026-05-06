import { useCallback, useRef, type KeyboardEvent } from 'react';
import styles from './Stepper.module.css';

export type StepperStep = {
  id: string;
  label: string;
};

export type StepperProps = {
  steps: StepperStep[];
  current: string;
  onStepClick?: (id: string) => void;
};

export function Stepper({ steps, current, onStepClick }: StepperProps): JSX.Element {
  const buttonsRef = useRef<Array<HTMLButtonElement | null>>([]);
  const currentIndex = steps.findIndex((step) => step.id === current);

  const handleKeyDown = useCallback(
    (event: KeyboardEvent<HTMLButtonElement>, index: number) => {
      if (!onStepClick) return;
      if (event.key !== 'ArrowRight' && event.key !== 'ArrowLeft') return;
      event.preventDefault();
      const delta = event.key === 'ArrowRight' ? 1 : -1;
      const next = (index + delta + steps.length) % steps.length;
      const nextStep = steps[next];
      if (!nextStep) return;
      buttonsRef.current[next]?.focus();
      onStepClick(nextStep.id);
    },
    [onStepClick, steps],
  );

  return (
    <ol className={styles.root} aria-label="Etapas">
      {steps.map((step, index) => {
        const isCompleted = currentIndex >= 0 && index < currentIndex;
        const isActive = step.id === current;
        const stateClass = isActive
          ? styles.itemActive
          : isCompleted
            ? styles.itemCompleted
            : styles.itemUpcoming;
        const interactive = Boolean(onStepClick);

        return (
          <li key={step.id} className={`${styles.item} ${stateClass}`}>
            <button
              ref={(node) => {
                buttonsRef.current[index] = node;
              }}
              type="button"
              className={styles.button}
              aria-current={isActive ? 'step' : undefined}
              onClick={interactive ? () => onStepClick?.(step.id) : undefined}
              onKeyDown={interactive ? (event) => handleKeyDown(event, index) : undefined}
              disabled={!interactive}
              tabIndex={interactive ? 0 : -1}
            >
              <span className={styles.dot} aria-hidden="true">
                {isCompleted ? <span className={styles.dotCheck}>✓</span> : index + 1}
              </span>
              <span className={styles.label}>{step.label}</span>
            </button>
            {index < steps.length - 1 ? <span className={styles.connector} aria-hidden="true" /> : null}
          </li>
        );
      })}
    </ol>
  );
}
