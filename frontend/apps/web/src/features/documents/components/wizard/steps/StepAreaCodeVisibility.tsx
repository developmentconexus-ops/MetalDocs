import { SelectableCard } from '../../../../../components/ui/SelectableCard';
import { SelectMenu, type SelectMenuOption } from '../../../../../components/ui/SelectMenu';
import { CodeChip } from '../../../../../components/ui/CodeChip';
import { Icon } from '../../../../../components/ui/Icon';
import { ApiError, resolveErrorMessage } from '../../../../../lib/api';
import type { DocumentProfile, ProcessArea } from '../../../../taxonomy/types';
import { VISIBILITY_KEYS, VISIBILITY_META, type VisibilityKey } from '../../../lib/visibilityMeta';
import { CodePreviewBanner } from '../CodePreviewBanner';
import { WizardFooter } from '../WizardShell';
import type { WizardExternalConfig, WizardInvitee } from '../../../state/wizard.reducer';
import styles from './StepAreaCodeVisibility.module.css';

export type StepAreaCodeVisibilityProps = {
  profile: DocumentProfile | null;
  areas: ProcessArea[];
  isAreasLoading: boolean;
  isAreasError: boolean;
  areasError: unknown;
  onAreasRetry: () => void;
  areaCode: string;
  title: string;
  visibility: VisibilityKey;
  invitees: WizardInvitee[];
  external: WizardExternalConfig;
  onChangeProfile: () => void;
  onSetArea: (code: string) => void;
  onSetTitle: (value: string) => void;
  onSetVisibility: (key: VisibilityKey) => void;
  onAddInvitee: (invitee: WizardInvitee) => void;
  onRemoveInvitee: (id: string) => void;
  onSetExternal: (patch: Partial<WizardExternalConfig>) => void;
  onAdvance: () => void;
  onBack: () => void;
  onCancel: () => void;
  advanceDisabled: boolean;
};

export function StepAreaCodeVisibility(props: StepAreaCodeVisibilityProps): JSX.Element {
  const {
    profile,
    areas,
    isAreasLoading,
    isAreasError,
    areasError,
    onAreasRetry,
    areaCode,
    title,
    visibility,
    invitees,
    external,
    onChangeProfile,
    onSetArea,
    onSetTitle,
    onSetVisibility,
    onAddInvitee,
    onRemoveInvitee,
    onSetExternal,
    onAdvance,
    onBack,
    onCancel,
    advanceDisabled,
  } = props;

  const areaOptions: SelectMenuOption[] = areas.map((area) => ({
    value: area.code,
    label: `${area.code} — ${area.name}`,
  }));

  return (
    <div className="card">
      <div className="kicker">Etapa 2 de 4</div>
      <h2 className="h2">Área, código e visibilidade</h2>

      <div className={styles.profileAreaRow}>
        <div>
          <label className="kicker">Perfil selecionado</label>
          <div className={styles.profileChip}>
            <CodeChip>{profile?.code ?? '—'}</CodeChip>
            <div>
              <div>{profile?.name ?? '—'}</div>
              <div className="caption">Família: {profile?.familyCode ?? '—'}</div>
            </div>
            <button type="button" className="btn btn-sm btn-ghost" onClick={onChangeProfile}>
              Trocar
            </button>
          </div>
        </div>
        <div>
          <label className="kicker" htmlFor="wizard-area">
            Área *
          </label>
          {isAreasLoading ? (
            <div className="caption" aria-busy="true">
              Carregando áreas…
            </div>
          ) : isAreasError ? (
            <div role="alert" aria-live="assertive">
              {resolveErrorMessage(
                areasError instanceof ApiError ? areasError.code : undefined,
                areasError instanceof Error ? areasError.message : undefined,
              )}{' '}
              <button type="button" className="btn btn-sm" onClick={onAreasRetry}>
                Tentar novamente
              </button>
            </div>
          ) : areas.length === 0 ? (
            <div>
              <div className="caption">Nenhuma área cadastrada.</div>
              <a className="btn btn-sm" href="/taxonomy/areas">
                Cadastrar área
              </a>
            </div>
          ) : (
            <SelectMenu
              id="wizard-area"
              value={areaCode}
              options={areaOptions}
              placeholder="Selecione uma área"
              onSelect={onSetArea}
            />
          )}
        </div>
      </div>

      <div className={styles.titleRow}>
        <label className="kicker" htmlFor="wizard-title">
          Título *
        </label>
        <input
          id="wizard-title"
          className="input"
          type="text"
          value={title}
          placeholder="Documento exemplo"
          onChange={(event) => onSetTitle(event.target.value)}
        />
      </div>

      <CodePreviewBanner profileCode={profile?.code ?? null} areaCode={areaCode} />

      {/* TODO(novo-documento:visibility): captured in form state but NOT submitted —
          controlled_documents has no visibility column today. See
          wiki/backlog/novo-documento.md#visibility for the backend prereq. */}
      <div className="kicker">Visibilidade *</div>
      <div className={styles.visibilityGrid}>
        {VISIBILITY_KEYS.map((key) => {
          const meta = VISIBILITY_META[key];
          const selected = visibility === key;
          return (
            <SelectableCard
              key={key}
              selected={selected}
              onSelect={() => onSetVisibility(key)}
            >
              <div className={styles.visibilityCardBody}>
                <span className={styles.visibilityIconTile}>
                  <Icon name={meta.icon} />
                </span>
                <div className={styles.visibilityText}>
                  <div className={styles.visibilityLabel}>{meta.label}</div>
                  <p className="caption">{meta.description}</p>
                </div>
                {selected ? <Icon name="check" /> : null}
              </div>
            </SelectableCard>
          );
        })}
      </div>

      {visibility === 'people' ? (
        <PeopleSubcontrols
          invitees={invitees}
          onAddInvitee={onAddInvitee}
          onRemoveInvitee={onRemoveInvitee}
        />
      ) : null}

      {visibility === 'external' ? (
        <ExternalSubcontrols external={external} onSetExternal={onSetExternal} />
      ) : null}

      <WizardFooter
        stepLabel={
          advanceDisabled
            ? 'Etapa 2 de 4 · Preencha área e título para continuar'
            : 'Etapa 2 de 4 · Avance para selecionar template'
        }
        primaryDisabled={advanceDisabled}
        onAdvance={onAdvance}
        onBack={onBack}
        onCancel={onCancel}
      />
    </div>
  );
}

type PeopleSubcontrolsProps = {
  invitees: WizardInvitee[];
  onAddInvitee: (invitee: WizardInvitee) => void;
  onRemoveInvitee: (id: string) => void;
};

function PeopleSubcontrols({ invitees, onAddInvitee: _onAddInvitee, onRemoveInvitee: _onRemoveInvitee }: PeopleSubcontrolsProps): JSX.Element {
  // Defer per NOTES.md audit — share/invite model not yet defined. Rendered
  // disabled (aria-disabled + grayed) so the UI does not imply unsupported
  // behavior. See wiki/backlog/novo-documento.md#sharing.
  return (
    <div
      className={`card ${styles.subcontrolsCard}`}
      role="group"
      aria-disabled="true"
      title="Em breve"
    >
      <div className="kicker">Pessoas convidadas</div>
      <div className={styles.subcontrolsRow}>
        {invitees.length === 0 ? (
          <span className="caption">Em breve — convite por pessoa.</span>
        ) : (
          invitees.map((row) => (
            <span key={row.id} className="pill">
              {row.label}{' '}
              <button
                type="button"
                className="btn btn-ghost btn-sm"
                aria-label={`Remover ${row.label}`}
                disabled
                aria-disabled="true"
              >
                <span aria-hidden="true">×</span>
              </button>
            </span>
          ))
        )}
      </div>
      <button
        type="button"
        className={`btn btn-sm ${styles.subcontrolsAddBtn}`}
        disabled
        aria-disabled="true"
        title="Em breve"
      >
        Adicionar pessoa
      </button>
    </div>
  );
}

type ExternalSubcontrolsProps = {
  external: WizardExternalConfig;
  onSetExternal: (patch: Partial<WizardExternalConfig>) => void;
};

function ExternalSubcontrols({ external, onSetExternal: _onSetExternal }: ExternalSubcontrolsProps): JSX.Element {
  // Defer per NOTES.md audit — external-share model not yet defined. Rendered
  // disabled. See wiki/backlog/novo-documento.md#sharing.
  return (
    <div
      className={`card ${styles.subcontrolsCard}`}
      role="group"
      aria-disabled="true"
      title="Em breve"
    >
      <div className="kicker">Compartilhamento externo</div>
      <div className={styles.subcontrolsCol}>
        <label>
          <input type="checkbox" checked={external.passwordRequired} disabled aria-disabled="true" readOnly />{' '}
          Exigir senha (em breve)
        </label>
        <label>
          <input type="checkbox" checked={external.watermark} disabled aria-disabled="true" readOnly />{' '}
          Aplicar marca d’água (em breve)
        </label>
        <label>
          Expira em (dias)
          <input
            className="input"
            type="number"
            min={0}
            value={external.expiresInDays ?? ''}
            disabled
            aria-disabled="true"
            readOnly
          />
        </label>
      </div>
    </div>
  );
}

export default StepAreaCodeVisibility;
