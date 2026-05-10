import { useRef } from 'react';
import { WizardFooter } from '../../../../shared/components/wizard/WizardFooter';
import { Icon } from '../../../../../components/ui/Icon';
import type { PermissionsMode } from '../../../state/templateWizard.reducer';
import styles from './StepPermissions.module.css';

// ─── Mock data ───────────────────────────────────────────────────────────────
// TODO(novo-template-wizard:permissions-roles-api): replace with real
// personnel-roles endpoint when API ships.
// Backlog: wiki/backlog/novo-template-wizard.md
const MOCK_ROLES = [
  { id: 'QUA-INSP', name: 'Inspetor de Qualidade', area: 'Qualidade', count: 14 },
  { id: 'QUA-ANA', name: 'Analista de Qualidade', area: 'Qualidade', count: 8 },
  { id: 'QUA-COORD', name: 'Coordenador de Qualidade', area: 'Qualidade', count: 2 },
  { id: 'PROD-OP', name: 'Operador de Produção', area: 'Produção', count: 47 },
  { id: 'PROD-LIDER', name: 'Líder de Produção', area: 'Produção', count: 6 },
  { id: 'MAN-TEC', name: 'Técnico de Manutenção', area: 'Manutenção', count: 11 },
] as const;

// TODO(novo-template-wizard:permissions-area-counts): replace area user counts
// with real aggregate when API ships.
const MOCK_AREAS = [
  { id: 'QUA', name: 'Qualidade', count: 28 },
  { id: 'PROD', name: 'Produção', count: 89 },
  { id: 'MAN', name: 'Manutenção', count: 34 },
  { id: 'RH', name: 'RH', count: 12 },
  { id: 'TI', name: 'TI & Segurança', count: 18 },
  { id: 'FIN', name: 'Financeiro', count: 9 },
] as const;

// TODO(novo-template-wizard:permissions-user-count): replace with real
// company-wide user count endpoint when API ships.
const COMPANY_USER_COUNT = 340;

const MODE_TABS = [
  { id: 'all' as const, label: 'Todos', desc: 'Empresa inteira' },
  { id: 'areas' as const, label: 'Por área', desc: 'Departamentos' },
  { id: 'roles' as const, label: 'Por funções', desc: 'Perfis específicos' },
];

// ─── Props ────────────────────────────────────────────────────────────────────

export type StepPermissionsProps = {
  permissionsMode: PermissionsMode;
  selectedRoleIds: string[];
  selectedAreaIds: string[];
  onSetMode: (mode: PermissionsMode) => void;
  onToggleRole: (id: string) => void;
  onToggleArea: (id: string) => void;
  onAdvance: () => void;
  onBack: () => void;
  advanceDisabled: boolean;
};

// ─── Component ────────────────────────────────────────────────────────────────

export function StepPermissions({
  permissionsMode,
  selectedRoleIds,
  selectedAreaIds,
  onSetMode,
  onToggleRole,
  onToggleArea,
  onAdvance,
  onBack,
  advanceDisabled,
}: StepPermissionsProps): JSX.Element {
  // Roving tabindex refs for mode radiogroup (arrow-key navigation per ARIA APG)
  const modeTabRefs = useRef<Array<HTMLButtonElement | null>>([]);

  function handleModeKeyDown(e: React.KeyboardEvent<HTMLButtonElement>, idx: number) {
    if (e.key !== 'ArrowRight' && e.key !== 'ArrowLeft') return;
    e.preventDefault();
    const next = e.key === 'ArrowRight' ? (idx + 1) % MODE_TABS.length : (idx - 1 + MODE_TABS.length) % MODE_TABS.length;
    const nextTab = MODE_TABS[next];
    if (!nextTab) return;
    onSetMode(nextTab.id);
    modeTabRefs.current[next]?.focus();
  }

  // Derived: total users covered (mocked counts)
  const coveredUserCount =
    permissionsMode === 'roles'
      ? MOCK_ROLES.filter((r) => selectedRoleIds.includes(r.id)).reduce((s, r) => s + r.count, 0)
      : permissionsMode === 'areas'
        ? MOCK_AREAS.filter((a) => selectedAreaIds.includes(a.id)).reduce((s, a) => s + a.count, 0)
        : COMPANY_USER_COUNT;

  const selectionCount =
    permissionsMode === 'roles' ? selectedRoleIds.length : selectedAreaIds.length;

  return (
    <div className="card">
      <div className="kicker">Etapa 4 de 5</div>
      <h2 className={`h2 ${styles.stepTitle}`}>Quem pode usar este template?</h2>
      <p className={`caption ${styles.intro}`}>
        Defina o público que verá este template ao criar um novo documento. Você pode ajustar depois.
      </p>

      {/* ── Mode segmented control ─────────────────────────────────── */}
      <div className={styles.modeSegmented} role="radiogroup" aria-label="Modo de permissão">
        {MODE_TABS.map((tab, idx) => (
          <button
            key={tab.id}
            ref={(el) => { modeTabRefs.current[idx] = el; }}
            type="button"
            role="radio"
            aria-checked={permissionsMode === tab.id}
            tabIndex={permissionsMode === tab.id ? 0 : -1}
            className={`${styles.modeTab} ${permissionsMode === tab.id ? styles.modeTabActive : ''}`}
            onClick={() => onSetMode(tab.id)}
            onKeyDown={(e) => handleModeKeyDown(e, idx)}
          >
            <span className={styles.modeTabLabel}>{tab.label}</span>
            <span className={`tiny ${styles.modeTabDesc}`}>{tab.desc}</span>
          </button>
        ))}
      </div>

      {/* ── All — company-wide banner ───────────────────────────────── */}
      {permissionsMode === 'all' && (
        <div className={styles.allBanner}>
          <span className={styles.allIcon} aria-hidden="true">
            <Icon name="home" size={20} />
          </span>
          <div className={styles.allBody}>
            <div className={styles.allTitle}>Toda a empresa</div>
            <p className="caption">
              Todos os colaboradores autenticados verão este template no wizard de novos documentos.
              Cobertura:{' '}
              <span className={`mono ${styles.allCount}`}>~{COMPANY_USER_COUNT}</span> usuários ativos em 6 áreas.
            </p>
          </div>
        </div>
      )}

      {/* ── Areas grid ─────────────────────────────────────────────── */}
      {permissionsMode === 'areas' && (
        <div
          className={styles.areaGrid}
          role="group"
          aria-label="Selecione áreas"
        >
          {MOCK_AREAS.map((area) => {
            const checked = selectedAreaIds.includes(area.id);
            return (
              <button
                key={area.id}
                type="button"
                role="checkbox"
                aria-checked={checked}
                className={`${styles.areaCard} ${checked ? styles.cardSelected : ''}`}
                onClick={() => onToggleArea(area.id)}
              >
                <span
                  className={`${styles.cardIcon} ${checked ? styles.cardIconSelected : ''}`}
                  aria-hidden="true"
                >
                  <Icon name="taxonomy" size={16} />
                </span>
                <div className={styles.cardBody}>
                  <div className={`mono ${styles.cardCode} ${checked ? styles.cardCodeSelected : ''}`}>
                    {area.id}
                  </div>
                  <div className={styles.cardName}>{area.name}</div>
                  <div className={`tiny ${styles.cardCount}`}>
                    <span className="mono">{area.count}</span> usuários
                  </div>
                </div>
                {checked && (
                  <span className={styles.cardCheck} aria-hidden="true">
                    <Icon name="check" size={16} />
                  </span>
                )}
              </button>
            );
          })}
        </div>
      )}

      {/* ── Roles grid ─────────────────────────────────────────────── */}
      {permissionsMode === 'roles' && (
        <div
          className={styles.roleGrid}
          role="group"
          aria-label="Selecione funções"
        >
          {MOCK_ROLES.map((role) => {
            const checked = selectedRoleIds.includes(role.id);
            return (
              <button
                key={role.id}
                type="button"
                role="checkbox"
                aria-checked={checked}
                className={`${styles.roleCard} ${checked ? styles.cardSelected : ''}`}
                onClick={() => onToggleRole(role.id)}
              >
                <span
                  className={`${styles.cardIcon} ${checked ? styles.cardIconSelected : ''}`}
                  aria-hidden="true"
                >
                  <Icon name="users" size={16} />
                </span>
                <div className={styles.cardBody}>
                  <div className={styles.roleIdRow}>
                    <span className={`mono ${styles.cardCode} ${checked ? styles.cardCodeSelected : ''}`}>
                      {role.id}
                    </span>
                    <span className="tiny">{role.area}</span>
                  </div>
                  <div className={styles.cardName}>{role.name}</div>
                  <div className={`tiny ${styles.cardCount}`}>
                    <span className="mono">{role.count}</span> usuários ativos
                  </div>
                </div>
                {checked && (
                  <span className={styles.cardCheck} aria-hidden="true">
                    <Icon name="check" size={16} />
                  </span>
                )}
              </button>
            );
          })}
        </div>
      )}

      {/* ── Coverage summary (roles + areas modes only) ─────────────── */}
      {permissionsMode !== 'all' && (
        <div className={styles.coverageSummary}>
          <div className={`kicker ${styles.coverageKicker}`}>Cobertura estimada</div>
          <div className={styles.coverageRow}>
            <span className={`mono ${styles.coverageCount}`}>{selectionCount}</span>
            <span className="caption">
              {permissionsMode === 'roles' ? (
                <>
                  {selectionCount === 1 ? 'perfil selecionado' : 'perfis selecionados'} · cobre{' '}
                  <span className={`mono ${styles.coverageUserCount}`}>~{coveredUserCount}</span> usuários
                  ativos
                </>
              ) : (
                <>
                  {selectionCount === 1 ? 'área selecionada' : 'áreas selecionadas'} · cobre todos os
                  usuários das áreas marcadas
                </>
              )}
            </span>
          </div>
        </div>
      )}

      <WizardFooter
        stepLabel={
          advanceDisabled
            ? 'Etapa 4 de 5 · Selecione ao menos um perfil'
            : 'Etapa 4 de 5 · Avance para revisar e confirmar'
        }
        showBack
        onBack={onBack}
        onAdvance={onAdvance}
        primaryDisabled={advanceDisabled}
      />
    </div>
  );
}

export default StepPermissions;
