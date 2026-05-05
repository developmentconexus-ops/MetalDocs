import type { ReactNode } from 'react';

type CodeChipProps = {
  children: ReactNode;
  className?: string;
};

export function CodeChip({ children, className }: CodeChipProps) {
  return (
    <span className={`code-chip mono${className ? ` ${className}` : ''}`}>
      {children}
    </span>
  );
}
