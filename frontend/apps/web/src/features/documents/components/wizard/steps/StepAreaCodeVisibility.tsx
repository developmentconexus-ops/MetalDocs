import { SelectableCard } from '../../../../../components/ui/SelectableCard';
import { SelectMenu } from '../../../../../components/ui/SelectMenu';
import { CodeChip } from '../../../../../components/ui/CodeChip';
import { Icon } from '../../../../../components/ui/Icon';
import { CodePreviewBanner } from '../CodePreviewBanner';
import { WizardFooter } from '../WizardShell';
import styles from './StepAreaCodeVisibility.module.css';

type VisibilitySample = {
  id: 'area' | 'people' | 'company' | 'external';
  icon: 'taxonomy' | 'users' | 'home' | 'link';
  label: string;
  desc: string;
  selected?: boolean;
};

// TODO(novo-documento): Phase 3c maps these against `visibilityMeta.ts`. The
// JSX reference uses {taxonomy, users, home, link} for the four cards; the
// SSOT in `visibilityMeta.ts` uses {users, user-plus, building, external-link}
// (icons not yet in Icon.tsx). Reconciliation deferred to Phase 3b/3c.
const VISIBILITY_SAMPLES: VisibilitySample[] = [
  { id: 'area', icon: 'taxonomy', label: 'Apenas a área', desc: 'Documento exemplo — descrição da visibilidade.' },
  { id: 'people', icon: 'users', label: 'Pessoas selecionadas', desc: 'Documento exemplo — descrição da visibilidade.' },
  { id: 'company', icon: 'home', label: 'Toda a empresa', desc: 'Documento exemplo — descrição da visibilidade.', selected: true },
  { id: 'external', icon: 'link', label: 'Externo / Cliente', desc: 'Documento exemplo — descrição da visibilidade.' },
];

const AREA_OPTIONS = [
  { value: 'QUA', label: 'QUA — Qualidade' },
  { value: 'PRD', label: 'PRD — Produção' },
];

export function StepAreaCodeVisibility(): JSX.Element {
  return (
    <div className="card">
      <div className="kicker">Etapa 2 de 4</div>
      <h2 className="h2">Área, código e visibilidade</h2>

      <div className={styles.profileAreaRow}>
        <div>
          <label className="kicker">Perfil selecionado</label>
          <div className={styles.profileChip}>
            <CodeChip>POP</CodeChip>
            <div>
              <div>Procedimento Operacional</div>
              <div className="caption">Família: Qualidade</div>
            </div>
            <button type="button" className="btn btn-sm btn-ghost">
              Trocar
            </button>
          </div>
        </div>
        <div>
          <label className="kicker">Área *</label>
          <SelectMenu
            id="wizard-area"
            value="QUA"
            options={AREA_OPTIONS}
            onSelect={() => {}}
          />
        </div>
      </div>

      <div className={styles.titleRow}>
        <label className="kicker">Título *</label>
        <input
          className="input"
          type="text"
          defaultValue="Documento exemplo"
          readOnly
        />
      </div>

      <CodePreviewBanner
        kicker="Código gerado · próximo em (POP, QUA)"
        code="POP-QUA-???"
        caption="≈ POP-QUA-??? · Código final atribuído ao confirmar."
      />

      <div className="kicker">Visibilidade *</div>
      <div className={styles.visibilityGrid}>
        {VISIBILITY_SAMPLES.map((option) => (
          <SelectableCard
            key={option.id}
            selected={Boolean(option.selected)}
            onSelect={() => {}}
          >
            <div className={styles.visibilityCardBody}>
              <span className={styles.visibilityIconTile}>
                <Icon name={option.icon} />
              </span>
              <div className={styles.visibilityText}>
                <div>{option.label}</div>
                <p className="caption">{option.desc}</p>
              </div>
              {option.selected ? <Icon name="check" /> : null}
            </div>
          </SelectableCard>
        ))}
      </div>

      {/* TODO(novo-documento): conditional sub-controls for visibility=people
          (invitee chips) and visibility=external (password / watermark / expiry)
          are deferred to Phase 3c. See worksheet §1.6. */}

      <WizardFooter stepLabel="Etapa 2 de 4 · Avance para selecionar template" />
    </div>
  );
}

export default StepAreaCodeVisibility;
