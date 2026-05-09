// Phase 3a: structure mirror only — no state, no handlers, no queries.
// Placeholder values used. Phase 3c wires to reducer state and useProfilesQuery.
import { SelectableCard } from '../../../../../components/ui/SelectableCard';
import { Icon } from '../../../../../components/ui/Icon';
import { WizardFooter } from '../../../../shared/components/wizard/WizardFooter';
import styles from './StepScope.module.css';

// Placeholder profiles — Phase 3c replaces with real data from useProfilesQuery.
const PLACEHOLDER_PROFILES = [
  { code: 'POP', name: 'Procedimento Operacional', description: 'Sequência detalhada de execução de uma rotina operacional.', familyCode: 'Qualidade', disabled: false },
  { code: 'IT',  name: 'Instrução de Trabalho',    description: 'Passo-a-passo enxuto para tarefa específica em estação.',   familyCode: 'Qualidade', disabled: false },
  { code: 'POL', name: 'Política',                 description: 'Diretrizes corporativas de alto nível com escopo organizacional.', familyCode: 'Governança', disabled: false },
  { code: 'DC',  name: 'Descrição de Cargo',       description: 'Identificação, responsabilidades e requisitos de uma posição.', familyCode: 'RH',         disabled: false },
  { code: 'CHK', name: 'Checklist Operacional',    description: 'Lista de verificação binária com observações por item.',    familyCode: 'Qualidade', disabled: true  },
  { code: 'FOR', name: 'Formulário',               description: 'Captura estruturada de dados em campo ou em workflow.',     familyCode: 'Geral',      disabled: false },
];

export function StepScope(): JSX.Element {
  const selectedCode = 'POP'; // placeholder — Phase 3c wires to reducer state

  return (
    <div className="card">
      <div className="kicker">Etapa 1 de 5</div>
      <h2 className="h2">Escopo do template</h2>
      <p className="caption">
        Templates podem ser genéricos para um perfil (POP, IT, etc.) ou derivar de um documento específico já publicado.
      </p>

      {/* Note: design had a TabBar with "A partir de um documento" tab (disabled/em breve),
          but phase0-audit.md decision: CUT entirely — wrong concept, only profile-based scope. */}

      <div className={styles.profileGrid}>
        {PLACEHOLDER_PROFILES.map((profile) => {
          const selected = selectedCode === profile.code;
          return (
            <SelectableCard
              key={profile.code}
              selected={selected}
              onSelect={() => { /* Phase 3c */ }}
              className={styles.profileCard}
              disabled={profile.disabled}
            >
              <div className={styles.profileHeader}>
                <span className={`mono ${styles.profileCode}`}>{profile.code}</span>
                <span className={styles.profileName}>{profile.name}</span>
                {profile.disabled && <span className="pill">em breve</span>}
                {selected && <Icon name="check" />}
              </div>
              <p className="caption">{profile.description}</p>
              <div className={styles.profileMeta}>
                <span>Família: <span>{profile.familyCode}</span></span>
                <span>·</span>
                <span>
                  {/* TODO(novo-template-wizard:profile-counts): show published template count per
                      profile — no summary endpoint today. See wiki/backlog/novo-template-wizard.md */}
                  <span className="mono">—</span> templates publicados
                </span>
              </div>
            </SelectableCard>
          );
        })}
      </div>

      <WizardFooter
        stepLabel="Etapa 1 de 5 · Selecione um perfil para continuar"
        showBack={false}
        primaryDisabled={true}
        onAdvance={() => { /* Phase 3c */ }}
        onCancel={() => { /* Phase 3c */ }}
      />
    </div>
  );
}

export default StepScope;
