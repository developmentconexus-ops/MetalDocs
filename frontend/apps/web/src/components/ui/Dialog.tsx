import { useEffect, useId, useRef, type MouseEvent, type ReactNode } from 'react';

import styles from './Dialog.module.css';

interface DialogProps {
  /**
   * Controls visibility. The `<dialog>` node stays mounted regardless of
   * `open`; `showModal()` / `close()` toggle the modal state so the browser
   * can restore focus to the previously focused element on close.
   */
  open: boolean;
  /** Accessible label rendered as the dialog heading. */
  title: string;
  /** Fires when the user dismisses via Escape, backdrop click, or close button. */
  onClose: () => void;
  /** Dialog body. */
  children: ReactNode;
  /** Optional action row rendered below the body (typically Cancel + primary). */
  footer?: ReactNode;
  /** Optional descriptive paragraph rendered between title and body. */
  description?: ReactNode;
  /** If true, clicking the backdrop is ignored (Esc and close button still work). */
  disableBackdropClose?: boolean;
}

/**
 * Accessible modal dialog built on the native `<dialog>` element.
 *
 * - `showModal()` traps focus to dialog descendants while open.
 * - Browser returns focus to the previously focused element when `close()`
 *   fires — this relies on the dialog node staying mounted across
 *   open/close cycles, so we do NOT unmount on `open=false`.
 * - Escape closes via the native `cancel` event.
 * - Backdrop click closes unless `disableBackdropClose` is set.
 *
 * Use this primitive instead of hand-rolling `role="dialog"` overlays.
 */
export function Dialog({
  open,
  title,
  onClose,
  children,
  footer,
  description,
  disableBackdropClose,
}: DialogProps) {
  const dialogRef = useRef<HTMLDialogElement | null>(null);
  const titleId = useId();
  const descriptionId = useId();

  useEffect(() => {
    const node = dialogRef.current;
    if (!node) return;

    if (open && !node.open) {
      if (typeof node.showModal === 'function') {
        try {
          node.showModal();
        } catch {
          // jsdom and older browsers can throw on showModal — fall back to
          // the boolean `open` attribute so the dialog still mounts.
          node.setAttribute('open', '');
        }
      } else {
        node.setAttribute('open', '');
      }
      return;
    }

    if (!open && node.open) {
      if (typeof node.close === 'function') {
        node.close();
      } else {
        node.removeAttribute('open');
      }
    }
  }, [open]);

  useEffect(() => {
    const node = dialogRef.current;
    if (!node) return;

    const handleCancel = (event: Event) => {
      event.preventDefault();
      onClose();
    };

    node.addEventListener('cancel', handleCancel);
    return () => node.removeEventListener('cancel', handleCancel);
  }, [onClose]);

  const handleBackdropClick = (event: MouseEvent<HTMLDialogElement>) => {
    if (disableBackdropClose) return;
    if (event.target === dialogRef.current) {
      onClose();
    }
  };

  return (
    <dialog
      ref={dialogRef}
      className={styles.dialog}
      aria-labelledby={titleId}
      aria-describedby={description ? descriptionId : undefined}
      onClick={handleBackdropClick}
    >
      {open ? (
        <div className={styles.content}>
          <header className={styles.header}>
            <h2 id={titleId} className={styles.title}>
              {title}
            </h2>
            <button
              type="button"
              className={styles.closeButton}
              aria-label="Fechar"
              onClick={onClose}
            >
              ×
            </button>
          </header>

          {description ? (
            <p id={descriptionId} className={styles.description}>
              {description}
            </p>
          ) : null}

          <div className={styles.body}>{children}</div>

          {footer ? <footer className={styles.footer}>{footer}</footer> : null}
        </div>
      ) : null}
    </dialog>
  );
}
