import { SelectableCard } from '../../../../../components/ui/SelectableCard';
import { Icon } from '../../../../../components/ui/Icon';
import { ApiError, resolveErrorMessage } from '../../../../../lib/api';
import type { DocumentProfile } from '../../../../taxonomy/types';
import { WizardFooter } from '../WizardShell';
import styles from './StepProfile.module.css';

export type StepProfileProps = {
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

export function StepProfile({
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
}: StepProfileProps): JSX.Element {
  return (
    <div className="card">
      <div className="kicker">Etapa 1 de 4</div>
      <h2 className="h2">Que tipo de documento você está criando?</h2>
      <p className="caption">
        O perfil determina o template e a sequência de codificação. Esta escolha não pode ser alterada depois.
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
        <div role="alert" aria-live="assertive" className="card">
          {resolveErrorMessage(
            error instanceof ApiError ? error.code : undefined,
            error instanceof Error ? error.message : undefined,
          )}
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
            const selected = selectedCode === profile.code;
            return (
              <SelectableCard
                key={profile.code}
                selected={selected}
                onSelect={() => onSelect(profile.code)}
                className={styles.profileCard}
              >
                <div className={styles.profileHeader}>
                  <span className={`mono ${styles.profileCode}`}>{profile.code}</span>
                  <span className={styles.profileName}>{profile.name}</span>
                  {selected ? <Icon name="check" /> : null}
                </div>
                <p className="caption">{profile.description || '—'}</p>
                <div className={styles.profileMeta}>
                  <span>
                    Família: <span>{profile.familyCode}</span>
                  </span>
                  <span>·</span>
                  <span>
                    {/* TODO(novo-documento:profile-counts): aggregate doc count per profile —
                        no endpoint today. See wiki/backlog/novo-documento.md#profile-counts. */}
                    <span className="mono">—</span> existentes
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
            ? 'Etapa 1 de 4 · Selecione um perfil para continuar'
            : 'Etapa 1 de 4 · Pronto para avançar'
        }
        showBack={false}
        primaryDisabled={advanceDisabled}
        onAdvance={onAdvance}
        onCancel={onCancel}
      />
    </div>
  );
}

export default StepProfile;
