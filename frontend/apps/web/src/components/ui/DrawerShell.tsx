import { useCallback, useEffect, useRef, type MouseEvent, type ReactNode } from "react";
import styles from "./DrawerShell.module.css";

interface DrawerShellProps {
  open: boolean;
  onClose: () => void;
  ariaLabel: string;
  children: ReactNode;
  testId?: string;
}

/**
 * Right-rail modal drawer built on the native `<dialog>` element.
 *
 * Encapsulates `showModal()` lifecycle, Escape (`cancel`) handling, backdrop
 * click-to-close, and the close button slot. Children render in the scrollable
 * body below a fixed top bar with a close button.
 */
export function DrawerShell({
  open,
  onClose,
  ariaLabel,
  children,
  testId,
}: DrawerShellProps) {
  const dialogRef = useRef<HTMLDialogElement | null>(null);

  useEffect(() => {
    const node = dialogRef.current;
    if (!node) return;
    if (open && !node.open) {
      try {
        node.showModal();
      } catch {
        node.setAttribute("open", "");
      }
    }
    if (!open && node.open) {
      try {
        node.close();
      } catch {
        node.removeAttribute("open");
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
    node.addEventListener("cancel", handleCancel);
    return () => node.removeEventListener("cancel", handleCancel);
  }, [onClose]);

  const handleBackdropClick = useCallback(
    (event: MouseEvent<HTMLDialogElement>) => {
      if (event.target === dialogRef.current) onClose();
    },
    [onClose],
  );

  return (
    <dialog
      ref={dialogRef}
      className={styles.dialog}
      aria-label={ariaLabel}
      onClick={handleBackdropClick}
      data-testid={testId}
    >
      <div className={styles.body}>
        <div className={styles.topRow}>
          <button
            type="button"
            className={styles.closeBtn}
            aria-label="Fechar"
            onClick={onClose}
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              aria-hidden="true"
            >
              <path d="M18 6 6 18" />
              <path d="m6 6 12 12" />
            </svg>
          </button>
        </div>
        <div className={styles.content}>{children}</div>
      </div>
    </dialog>
  );
}
