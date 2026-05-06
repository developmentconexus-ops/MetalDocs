import styles from './StatusPill.module.css';

export type DocumentStatus =
  | 'draft'
  | 'review'
  | 'under_review'
  | 'approved'
  | 'frozen'
  | 'rejected'
  | 'archived'
  | 'finalized'
  | 'scheduled'
  | 'published'
  | 'superseded'
  | 'obsolete';

const STATUS_CONFIG: Record<DocumentStatus, { label: string; pillClass: string }> = {
  draft: { label: 'Rascunho', pillClass: 'pill pill-draft' },
  review: { label: 'Em Revisão', pillClass: 'pill pill-review' },
  under_review: { label: 'Em Revisão', pillClass: 'pill pill-review' },
  approved: { label: 'Aprovado', pillClass: 'pill pill-approved' },
  frozen: { label: 'Congelado', pillClass: 'pill pill-frozen' },
  rejected: { label: 'Rejeitado', pillClass: 'pill pill-rejected' },
  archived: { label: 'Obsoleto', pillClass: 'pill pill-obsolete' },
  finalized: { label: 'Publicado', pillClass: 'pill pill-published' },
  scheduled: { label: 'Agendado', pillClass: 'pill pill-scheduled' },
  published: { label: 'Publicado', pillClass: 'pill pill-published' },
  superseded: { label: 'Substituído', pillClass: 'pill pill-superseded' },
  obsolete: { label: 'Obsoleto', pillClass: 'pill pill-obsolete' },
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
