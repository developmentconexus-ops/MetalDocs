import styles from './EditorMetaSidebar.module.css';
import { CodeChip } from '../../../components/ui/CodeChip';
import { Avatar } from '../../../components/ui/Avatar';
import { TimelineRail } from '../../../components/ui/TimelineRail';

type EditorMetaSidebarProps = {
  open: boolean;
  onToggle: () => void;
  code?: string;
};

// TODO(editor:meta): replace MOCK_META with real data when GET /api/v1/documents/:id
// returns ProfileName, AreaName, NextReviewAt, Visibility.
// Backlog: wiki/backlog/editor.md (row 1)
const MOCK_META = [
  { label: 'Perfil', value: 'Procedimento Op.' },
  { label: 'Área', value: 'RH' },
  { label: 'Vigência atual', value: 'v4 · 15/04/2026' },
  { label: 'Próx. revisão', value: '15/04/2027' },
  { label: 'Visibilidade', value: 'Toda empresa' },
] as const;

// TODO(editor:revisions): replace MOCK_REVISIONS with real data when
// GET /api/v1/documents/:id/revisions ships.
// Backlog: wiki/backlog/editor.md (row 2)
const MOCK_REVISIONS = [
  { id: 'v5', title: 'v5', subtitle: 'Você · em edição', aside: 'agora', active: true },
  { id: 'v4', title: 'v4', subtitle: 'Anexos LGPD', aside: '15/04/2026' },
  { id: 'v3', title: 'v3', subtitle: 'Integração SAP', aside: '03/11/2025' },
  { id: 'rev-2', title: 'rev. 2', subtitle: 'Estágio', aside: '22/05/2024' },
  { id: 'v1', title: 'v1', subtitle: 'Versão inicial', aside: '08/01/2024' },
];

// TODO(editor:approvers): replace MOCK_APPROVERS with real data when
// GET /api/v1/documents/:id/signoffs ships.
// Backlog: wiki/backlog/editor.md (row 3)
const MOCK_APPROVERS = [
  { id: '1', name: 'Rafael Castro', role: 'Líder Qualidade', status: 'next' as const },
  { id: '2', name: 'Juliana Prado', role: 'Gerente Geral', status: 'wait' as const },
];

export function EditorMetaSidebar({ open, onToggle, code }: EditorMetaSidebarProps) {
  return (
    <div className={styles.sidebarOuter}>
      <button
        type="button"
        className={styles.toggleTab}
        onClick={onToggle}
        aria-label={open ? 'Fechar painel de metadados' : 'Abrir painel de metadados'}
        aria-expanded={open}
      >
        <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          {open
            ? <polyline points="9 18 15 12 9 6" />
            : <polyline points="15 18 9 12 15 6" />}
        </svg>
      </button>
      {open && (
        <aside className={styles.sidebar} aria-label="Metadados do documento">
          <section className={styles.section}>
            <div className={styles.sectionHeader}>Metadados</div>
            <div className={styles.metaRows}>
              {code && (
                <div className={styles.metaRow}>
                  <span className={styles.metaLabel}>Código</span>
                  <CodeChip>{code}</CodeChip>
                </div>
              )}
              {MOCK_META.map((item) => (
                <div key={item.label} className={styles.metaRow}>
                  <span className={styles.metaLabel}>{item.label}</span>
                  <span className={styles.metaValue}>{item.value}</span>
                </div>
              ))}
            </div>
          </section>
          <div className={styles.divider} />
          <section className={styles.section}>
            <div className={styles.sectionHeader}>Revisões</div>
            <TimelineRail items={MOCK_REVISIONS} ariaLabel="Histórico de revisões" variant="flat" />
          </section>
          <div className={styles.divider} />
          <section className={styles.section}>
            <div className={styles.sectionHeader}>Próximos aprovadores</div>
            <div className={styles.approverList}>
              {MOCK_APPROVERS.map((a) => (
                <div key={a.id} className={styles.approverRow}>
                  <Avatar name={a.name} size="sm" />
                  <div className={styles.approverInfo}>
                    <span className={styles.approverName}>{a.name}</span>
                    <span className={styles.approverRole}>{a.role}</span>
                  </div>
                  <span className={`${styles.approverBadge} ${a.status === 'next' ? styles.approverBadgeNext : styles.approverBadgeWait}`}>
                    {a.status === 'next' ? 'próximo' : 'aguarda'}
                  </span>
                </div>
              ))}
            </div>
          </section>
        </aside>
      )}
    </div>
  );
}
