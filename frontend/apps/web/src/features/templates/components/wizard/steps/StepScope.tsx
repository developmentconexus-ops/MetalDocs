import { SelectableCard } from '../../../../../components/ui/SelectableCard';
import { useRovingRadioGroup } from '../../../../../components/ui/useRovingRadioGroup';
import { Icon } from '../../../../../components/ui/Icon';
import { resolveQueryError } from '../../../../../lib/api';
import type { DocumentProfile } from '../../../../taxonomy/types';
import { WizardFooter } from '../../../../shared/components/wizard/WizardFooter';
import type { ScopeType } from '../../../state/templateWizard.reducer';
import styles from './StepScope.module.css';

// TODO(novo-template-wizard:chk-disabled): CHK disabled until Checklist feature ships.
// Remove this hardcode when taxonomy API exposes an `enabled` flag per profile.
const DISABLED_PROFILES = new Set(['CHK']);

export type StepScopeProps = {
  scopeType: ScopeType | null;
  onSelectScopeType: (scopeType: ScopeType) => void;
  profiles: DocumentProfile[];
  isLoading: boolean;
  isError: boolean;
  error: unknown;
  selectedCode: string | null;
  onSelect: (code: string) => void;
  onAdvance: () => void;
  onCancel: () => void;
  advanceDisabled: boolean;
  onRetry: () => void;
};

export function StepScope({
  scopeType,
  onSelectScopeType,
  profiles,
  isLoading,
  isError,
  error,
  selectedCode,
  onSelect,
  onAdvance,
  onCancel,
  advanceDisabled,
  onRetry,
}: StepScopeProps): JSX.Element {
  const scopeIndex = scopeType === 'generic' ? 0 : scopeType === 'profile' ? 1 : -1;
  const scopeGroup = useRovingRadioGroup({
    count: 2,
    selectedIndex: scopeIndex,
    onSelect: (newIndex) => onSelectScopeType(newIndex === 0 ? 'generic' : 'profile'),
    orientation: 'horizontal',
    ariaLabel: 'Tipo de escopo do template',
  });
  const profileIndex = profiles.findIndex((profile) => profile.code === selectedCode);
  const profileGroup = useRovingRadioGroup({
    count: profiles.length,
    selectedIndex: profileIndex,
    onSelect: (newIndex) => {
      const profile = profiles[newIndex];
      if (profile) onSelect(profile.code);
    },
    orientation: 'both',
    ariaLabel: 'Perfil do template',
  });

  return (
    <div className="card">
      <div className="kicker">Etapa 1 de 5</div>
      <h2 className="h2">Escopo do template</h2>
      <p className="caption">
        Templates genéricos aparecem para todos os perfis. Templates de perfil só aparecem ao criar documentos daquele perfil.
      </p>

      {/* Scope type choice */}
      <div className={styles.scopeGrid} {...scopeGroup.groupProps}>
        <SelectableCard
          {...scopeGroup.getItemProps(0)}
          selected={scopeType === 'generic'}
          onSelect={() => onSelectScopeType('generic')}
          className={styles.scopeCard}
        >
          <div className={styles.scopeHeader}>
            <span className={styles.scopeLabel}>Genérico</span>
            {scopeType === 'generic' ? <Icon name="check" /> : null}
          </div>
          <p className="caption">Aplicável a qualquer perfil de documento.</p>
        </SelectableCard>

        <SelectableCard
          {...scopeGroup.getItemProps(1)}
          selected={scopeType === 'profile'}
          onSelect={() => onSelectScopeType('profile')}
          className={styles.scopeCard}
        >
          <div className={styles.scopeHeader}>
            <span className={styles.scopeLabel}>Para um perfil</span>
            {scopeType === 'profile' ? <Icon name="check" /> : null}
          </div>
          <p className="caption">Restrito a um perfil específico (POP, DC, etc.).</p>
        </SelectableCard>
      </div>

      {/* Profile grid — only when profile scope selected */}
      {scopeType === 'profile' && (
        <div className={styles.profileSection}>
          <div className={`kicker ${styles.profileSectionKicker}`}>Selecione o perfil</div>

          {isLoading ? (
            <div className={styles.profileGrid} aria-busy="true">
              <div className="card">
                <div className="caption">Carregando perfis…</div>
              </div>
            </div>
          ) : isError ? (
            <div role="alert" aria-live="assertive" aria-atomic="true" className="card">
              {resolveQueryError(error, 'Falha ao carregar perfis.')}
              <div>
                <button type="button" className="btn btn-sm" onClick={onRetry}>
                  Tentar novamente
                </button>
              </div>
            </div>
          ) : profiles.length === 0 ? (
            <div className="card">
              <div className="caption">Nenhum perfil cadastrado.</div>
              <a className="btn btn-sm" href="/taxonomy/profiles">
                Cadastrar perfil
              </a>
            </div>
          ) : (
            <div className={styles.profileGrid} {...profileGroup.groupProps}>
              {profiles.map((profile, idx) => {
                const isDisabled = DISABLED_PROFILES.has(profile.code);
                const selected = selectedCode === profile.code;
                return (
                  <SelectableCard
                    key={profile.code}
                    {...profileGroup.getItemProps(idx)}
                    selected={selected}
                    disabled={isDisabled}
                    title={isDisabled ? 'Em breve' : undefined}
                    onSelect={() => onSelect(profile.code)}
                    className={styles.profileCard}
                  >
                    <div className={styles.profileHeader}>
                      <span className={`mono ${styles.profileCode}`}>{profile.code}</span>
                      <span className={styles.profileName}>{profile.name}</span>
                      {isDisabled ? (
                        <span className={styles.disabledBadge}>em breve</span>
                      ) : selected ? (
                        <Icon name="check" />
                      ) : null}
                    </div>
                    <p className="caption">{profile.description || '—'}</p>
                    <div className={styles.profileMeta}>
                      <span>
                        Família: <span>{profile.familyCode}</span>
                      </span>
                      <span>·</span>
                      <span>
                        {/* TODO(novo-template-wizard:template-counts): aggregate template count per
                            profile — no endpoint today. Remove placeholder when API ships. */}
                        <span className="mono">—</span> templates
                      </span>
                    </div>
                  </SelectableCard>
                );
              })}
            </div>
          )}
        </div>
      )}

      <WizardFooter
        stepLabel={stepLabel(scopeType, selectedCode)}
        showBack={false}
        primaryDisabled={advanceDisabled}
        onAdvance={onAdvance}
        onCancel={onCancel}
      />
    </div>
  );
}

function stepLabel(scopeType: ScopeType | null, selectedCode: string | null): string {
  if (scopeType === null) return 'Etapa 1 de 5 · Selecione o escopo para continuar';
  if (scopeType === 'generic') return 'Etapa 1 de 5 · Pronto para avançar';
  if (selectedCode === null) return 'Etapa 1 de 5 · Selecione um perfil para continuar';
  return 'Etapa 1 de 5 · Pronto para avançar';
}

export default StepScope;
