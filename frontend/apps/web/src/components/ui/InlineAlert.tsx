import type { ReactNode } from 'react';

export type InlineAlertTone = 'error' | 'warning' | 'info';

type InlineAlertProps = {
  /** Visual/semantic tone. `error` uses role="alert" + aria-live="assertive";
   *  `warning`/`info` use role="status" + aria-live="polite" so they don't
   *  interrupt assistive-tech users for non-blocking messages. */
  tone?: InlineAlertTone;
  message: ReactNode;
  /** Optional action rendered below the message (e.g. a "Tentar novamente" retry button). */
  action?: ReactNode;
  className?: string;
};

const ROLE_BY_TONE: Record<InlineAlertTone, 'alert' | 'status'> = {
  error: 'alert',
  warning: 'status',
  info: 'status',
};

const ARIA_LIVE_BY_TONE: Record<InlineAlertTone, 'assertive' | 'polite'> = {
  error: 'assertive',
  warning: 'polite',
  info: 'polite',
};

/**
 * InlineAlert — shared inline error/warning/info banner primitive.
 *
 * Consolidates the repeated `role="alert" aria-live="assertive" aria-atomic="true"`
 * inline banner shape used across wizard steps and document/template flows
 * (frontend-primitives-refactor.md backlog item). Renders on the `card`
 * global utility class by default so it matches existing call sites.
 *
 * @example
 * ```tsx
 * <InlineAlert
 *   message={resolveQueryError(error, 'Falha ao carregar perfis.')}
 *   action={<button type="button" className="btn btn-sm" onClick={onRetry}>Tentar novamente</button>}
 * />
 * ```
 */
export function InlineAlert({ tone = 'error', message, action, className }: InlineAlertProps) {
  const cls = `card${className ? ` ${className}` : ''}`;
  return (
    <div role={ROLE_BY_TONE[tone]} aria-live={ARIA_LIVE_BY_TONE[tone]} aria-atomic="true" className={cls}>
      {message}
      {action ? <div>{action}</div> : null}
    </div>
  );
}
