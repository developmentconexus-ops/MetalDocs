import styles from './StatusPill.module.css';

export type DocumentStatus =
  | 'draft'
  | 'review'
  | 'approved'
  | 'frozen'
  | 'rejected'
  | 'archived'
  | 'finalized';

const STATUS_CONFIG: Record<DocumentStatus, { label: string; pillClass: string }> = {
  draft:     { label: 'Rascunho',   pillClass: 'pill pill-draft' },
  review:    { label: 'Em revisão', pillClass: 'pill pill-review' },
  approved:  { label: 'Aprovado',   pillClass: 'pill pill-approved' },
  frozen:    { label: 'Frozen',     pillClass: 'pill pill-frozen' },
  rejected:  { label: 'Rejeitado',  pillClass: 'pill pill-rejected' },
  archived:  { label: 'Arquivado',  pillClass: 'pill pill-archived' },
  finalized: { label: 'Finalizado', pillClass: 'pill pill-approved' },
};

type StatusPillProps = {
  status: DocumentStatus;
  className?: string;
};

export function StatusPill({ status, className }: StatusPillProps) {
  const config = STATUS_CONFIG[status] ?? { label: status, pillClass: 'pill' };
  return (
    <span className={`${config.pillClass} ${styles.pill}${className ? ` ${className}` : ''}`}>
      <span className="dot" />
      {config.label}
    </span>
  );
}
