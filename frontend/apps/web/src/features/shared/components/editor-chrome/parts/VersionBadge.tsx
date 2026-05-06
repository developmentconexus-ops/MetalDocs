import type { ReactNode } from 'react';
import styles from './VersionBadge.module.css';

type VersionBadgeProps = {
  children: ReactNode;
  className?: string;
};

/**
 * Brand-colored monospace chip for revision/version labels.
 * Examples: "REV05", "v5".
 */
export function VersionBadge({ children, className }: VersionBadgeProps) {
  return (
    <span className={`${styles.badge}${className ? ` ${className}` : ''}`}>
      {children}
    </span>
  );
}
