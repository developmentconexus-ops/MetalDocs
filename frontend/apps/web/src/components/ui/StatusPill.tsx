import styles from './StatusPill.module.css';

export type DocumentStatus =
  | 'draft'
  | 'under_review'
  | 'approved'
  | 'frozen'
  | 'archived'
  | 'scheduled'
  | 'published'
  | 'superseded'
  | 'obsolete';

const STATUS_CONFIG: Record<DocumentStatus, { label: string; pillClass: string }> = {
  draft: { label: 'Rascunho', pillClass: 'pill pill-draft' },
  under_review: { label: 'Em Revisão', pillClass: 'pill pill-review' },
  approved: { label: 'Aprovado', pillClass: 'pill pill-approved' },
  frozen: { label: 'Congelado', pillClass: 'pill pill-frozen' },
  archived: { label: 'Obsoleto', pillClass: 'pill pill-obsolete' },
  scheduled: { label: 'Agendado', pillClass: 'pill pill-scheduled' },
  published: { label: 'Publicado', pillClass: 'pill pill-published' },
  superseded: { label: 'Substituído', pillClass: 'pill pill-superseded' },
  obsolete: { label: 'Obsoleto', pillClass: 'pill pill-obsolete' },
};

type StatusPillProps = {
  status: DocumentStatus | (string & {});
  className?: string;
};

export function StatusPill({ status, className }: StatusPillProps) {
  const config = STATUS_CONFIG[status as DocumentStatus] ?? { label: status, pillClass: 'pill' };
  return (
    <span className={`${config.pillClass} ${styles.pill}${className ? ` ${className}` : ''}`}>
      <span className={styles.dot} />
      {config.label}
    </span>
  );
}
