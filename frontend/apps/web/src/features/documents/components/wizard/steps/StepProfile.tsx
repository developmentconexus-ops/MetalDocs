import { Link } from 'react-router-dom';
import { SelectableCard } from '../../../../../components/ui/SelectableCard';
import { Icon } from '../../../../../components/ui/Icon';
import { InlineAlert } from '../../../../../components/ui/InlineAlert';
import { resolveQueryError } from '../../../../../lib/api';
import type { DocumentProfile } from '../../../../taxonomy/types';
import { useHasCapability } from '../../../../../lib/iam/useHasCapability';
import { WizardFooter } from '../../../../shared/components/wizard/WizardFooter';
import styles from './StepProfile.module.css';

const NO_ROUTE_MARKER = 'Incompleto — sem rota de aprovação';

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
  // Affordance gate only — "Configurar rota" points at /approval-routes, which
  // the AppShell guard bounces without this capability. Capabilities, never
  // roles (ADR 0022); the backend remains the sole authz enforcer.
  const canManageRoutes = useHasCapability('route.manage');

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
        <InlineAlert
          message={resolveQueryError(error, 'Falha ao carregar perfis.')}
          action={
            <button type="button" className="btn btn-sm" onClick={onRetry}>
              Tentar novamente
            </button>
          }
        />
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
            // D1: a profile with no active approval route cannot receive a
            // document at all — creating under it is rejected with 409
            // `state.approval_route_missing`. Fail closed in the UI so the user
            // never walks four wizard steps into a guaranteed rejection; the
            // backend 409 stays the backstop, not the first line.
            const eligible = profile.hasActiveRoute;
            const card = (
              <SelectableCard
                key={profile.code}
                selected={selected && eligible}
                disabled={!eligible}
                title={!eligible ? 'Sem rota de aprovação ativa' : undefined}
                onSelect={() => {
                  if (!eligible) return;
                  onSelect(profile.code);
                }}
                className={styles.profileCard}
              >
                <div className={styles.profileHeader}>
                  <span className={`mono ${styles.profileCode}`}>{profile.code}</span>
                  <span className={styles.profileName}>{profile.name}</span>
                  {selected && eligible ? <Icon name="check" /> : null}
                </div>
                <p className="caption">{profile.description || '—'}</p>
                {!eligible && <p className={styles.profileBlocked}>{NO_ROUTE_MARKER}</p>}
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

            if (eligible) return card;

            // The remediation link lives OUTSIDE the card: SelectableCard is a
            // <button> and a nested <a> is invalid (and unreachable while
            // disabled). Without route.manage the user gets the marker only.
            return (
              <div key={profile.code} className={styles.profileCardBlockedCell}>
                {card}
                {canManageRoutes && (
                  <Link to="/approval-routes" className={styles.profileBlockedLink}>
                    Configurar rota
                  </Link>
                )}
              </div>
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
