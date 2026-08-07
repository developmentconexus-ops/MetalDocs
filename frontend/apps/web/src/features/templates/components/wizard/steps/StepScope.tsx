import { Link } from 'react-router-dom';
import { SelectableCard } from '../../../../../components/ui/SelectableCard';
import { useRovingRadioGroup } from '../../../../../components/ui/useRovingRadioGroup';
import { Icon } from '../../../../../components/ui/Icon';
import { resolveQueryError } from '../../../../../lib/api';
import type { DocumentProfile } from '../../../../taxonomy/types';
import { useHasCapability } from '../../../../../lib/iam/useHasCapability';
import { WizardFooter } from '../../../../shared/components/wizard/WizardFooter';
import styles from './StepScope.module.css';

// TODO(novo-template-wizard:chk-disabled): CHK disabled until Checklist feature ships.
// Remove this hardcode when taxonomy API exposes an `enabled` flag per profile.
const DISABLED_PROFILES = new Set(['CHK']);

const NO_TEMPLATE_ROUTE_MARKER = 'Incompleto — sem rota de aprovação de template';

// Single eligibility predicate for BOTH input paths (pointer + roving-radio
// keyboard). ADR 0086 config-first gate: a profile with no active TEMPLATE
// approval route cannot receive a template at all — POST /templates is rejected
// with 409 APPROVAL_ROUTE_MISSING. Fail closed in the UI so the user never walks
// four wizard steps into a guaranteed rejection; the backend 409 stays the
// backstop, not the first line.
function isProfileDisabled(profile: DocumentProfile): boolean {
  return DISABLED_PROFILES.has(profile.code) || !profile.hasActiveTemplateRoute;
}

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
  // Affordance gate only — "Configurar rota" points at /approval-routes, which
  // the AppShell guard bounces without this capability. Capabilities, never
  // roles (ADR 0022); the backend remains the sole authz enforcer.
  const canManageRoutes = useHasCapability('route.manage');

  const profileIndex = profiles.findIndex((profile) => profile.code === selectedCode);
  const profileGroup = useRovingRadioGroup({
    count: profiles.length,
    selectedIndex: profileIndex,
    onSelect: (newIndex) => {
      const profile = profiles[newIndex];
      // Fail closed: Arrow/Home/End must never land selection on a profile the
      // click path refuses. Without this the reducer would unlock the remaining
      // steps and submit would eat a guaranteed 409 APPROVAL_ROUTE_MISSING.
      if (!profile || isProfileDisabled(profile)) return;
      onSelect(profile.code);
    },
    orientation: 'both',
    ariaLabel: 'Perfil do template',
  });

  return (
    <div className="card">
      <div className="kicker">Etapa 1 de 4</div>
      <h2 className="h2">Perfil do template</h2>
      <p className="caption">
        Todo template pertence a um perfil documental e só aparece ao criar documentos daquele
        perfil. Esta escolha não pode ser alterada depois.
      </p>

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
              const comingSoon = DISABLED_PROFILES.has(profile.code);
              const missingRoute = !profile.hasActiveTemplateRoute;
              const isDisabled = isProfileDisabled(profile);
              const selected = selectedCode === profile.code;
              const card = (
                <SelectableCard
                  key={profile.code}
                  {...profileGroup.getItemProps(idx)}
                  {...(isDisabled ? { tabIndex: -1 } : {})}
                  selected={selected && !isDisabled}
                  disabled={isDisabled}
                  title={
                    comingSoon
                      ? 'Em breve'
                      : missingRoute
                        ? 'Sem rota de aprovação de template ativa'
                        : undefined
                  }
                  onSelect={() => {
                    if (isDisabled) return;
                    onSelect(profile.code);
                  }}
                  className={styles.profileCard}
                >
                  <div className={styles.profileHeader}>
                    <span className={`mono ${styles.profileCode}`}>{profile.code}</span>
                    <span className={styles.profileName}>{profile.name}</span>
                    {comingSoon ? (
                      <span className={styles.disabledBadge}>em breve</span>
                    ) : selected && !isDisabled ? (
                      <Icon name="check" />
                    ) : null}
                  </div>
                  <p className="caption">{profile.description || '—'}</p>
                  {!comingSoon && missingRoute && (
                    <p className={styles.profileBlocked}>{NO_TEMPLATE_ROUTE_MARKER}</p>
                  )}
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

              if (comingSoon || !missingRoute) return card;

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
      </div>

      <WizardFooter
        stepLabel={stepLabel(selectedCode)}
        showBack={false}
        primaryDisabled={advanceDisabled}
        onAdvance={onAdvance}
        onCancel={onCancel}
      />
    </div>
  );
}

function stepLabel(selectedCode: string | null): string {
  if (selectedCode === null) return 'Etapa 1 de 4 · Selecione um perfil para continuar';
  return 'Etapa 1 de 4 · Pronto para avançar';
}

export default StepScope;
