import { SelectableCard } from '../../../../../components/ui/SelectableCard';
import { Icon } from '../../../../../components/ui/Icon';
import { resolveQueryError } from '../../../../../lib/api';
import type { DocumentProfile } from '../../../../taxonomy/types';
import { WizardFooter } from '../../../../shared/components/wizard/WizardFooter';
import styles from './StepScope.module.css';

// TODO(novo-template-wizard:chk-disabled): CHK disabled until Checklist feature ships.
// Remove this hardcode when taxonomy API exposes an `enabled` flag per profile.
const DISABLED_PROFILES = new Set(['CHK']);

export type StepScopeProps = {
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
  return (
    <div className="card">
      <div className="kicker">Etapa 1 de 5</div>
      <h2 className="h2">Escopo do template</h2>
      <p className="caption">
        Templates podem ser genéricos para um perfil (POP, IT, etc.) ou derivar de um documento específico já publicado.
      </p>

      {isLoading ? (
        <div className={styles.profileGrid} aria-busy="true">
          <div className="card">
            <div className="caption">Carregando perfis…</div>
          </div>
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
        <div className={styles.profileGrid}>
          {profiles.map((profile) => {
            const isDisabled = DISABLED_PROFILES.has(profile.code);
            const selected = selectedCode === profile.code;
            return (
              <SelectableCard
                key={profile.code}
                selected={selected}
                disabled={isDisabled}
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

      <WizardFooter
        stepLabel={
          selectedCode === null
            ? 'Etapa 1 de 5 · Selecione um perfil para continuar'
            : 'Etapa 1 de 5 · Pronto para avançar'
        }
        showBack={false}
        primaryDisabled={advanceDisabled}
        onAdvance={onAdvance}
        onCancel={onCancel}
      />
    </div>
  );
}

export default StepScope;
