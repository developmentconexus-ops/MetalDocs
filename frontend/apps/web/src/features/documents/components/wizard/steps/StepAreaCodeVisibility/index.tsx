import { SelectableCard } from '../../../../../../components/ui/SelectableCard';
import { SelectMenu, type SelectMenuOption } from '../../../../../../components/ui/SelectMenu';
import { CodeChip } from '../../../../../../components/ui/CodeChip';
import { Icon } from '../../../../../../components/ui/Icon';
import { resolveQueryError } from '../../../../../../lib/api';
import type { DocumentProfile, ProcessArea } from '../../../../../taxonomy/types';
import { VISIBILITY_KEYS, VISIBILITY_META, type VisibilityKey } from '../../../../lib/visibilityMeta';
import { CodePreviewBanner } from '../../CodePreviewBanner';
import { WizardFooter } from '../../WizardFooter';
import type { WizardExternalConfig, WizardInvitee } from '../../../../state/wizard.reducer';
import { PeopleSubcontrols } from './PeopleSubcontrols';
import { ExternalSubcontrols } from './ExternalSubcontrols';
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
      <h2 className={`h2 ${styles.stepH2}`}>Área, código e visibilidade</h2>

      <div className={styles.profileAreaRow}>
        <div>
          <label className="kicker">Perfil selecionado</label>
          <div className={styles.profileChip}>
            <CodeChip>{profile?.code ?? '—'}</CodeChip>
            <div className={styles.profileChipInfo}>
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
            <div role="alert" aria-live="assertive" className="card">
              {resolveQueryError(areasError, 'Falha ao carregar áreas de processo.')}{' '}
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
      <div className={`kicker ${styles.visibilityKicker}`}>Visibilidade *</div>
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

export default StepAreaCodeVisibility;
