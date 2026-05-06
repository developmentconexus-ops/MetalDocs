import { SelectableCard } from '../../../../../components/ui/SelectableCard';
import { Icon } from '../../../../../components/ui/Icon';
import { WizardFooter } from '../WizardShell';
import styles from './StepProfile.module.css';

type ProfileSample = {
  id: string;
  name: string;
  desc: string;
  family: string;
  count: string;
  selected?: boolean;
};

// TODO(novo-documento): Phase 3c replaces this hardcoded sample with
// `useProfilesQuery()` results. See worksheet §1.5.
const PROFILE_SAMPLES: ProfileSample[] = [
  {
    id: 'POP',
    name: 'Procedimento Operacional',
    desc: 'Documento exemplo — descrição do perfil.',
    family: 'Qualidade',
    count: '—',
    selected: true,
  },
  {
    id: 'IT',
    name: 'Instrução de Trabalho',
    desc: 'Documento exemplo — descrição do perfil.',
    family: 'Operação',
    count: '—',
  },
];

export function StepProfile(): JSX.Element {
  return (
    <div className="card">
      <div className="kicker">Etapa 1 de 4</div>
      <h2 className="h2">Que tipo de documento você está criando?</h2>
      <p className="caption">
        O perfil determina o template e a sequência de codificação. Esta escolha não pode ser alterada depois.
      </p>
      <div className={styles.profileGrid}>
        {PROFILE_SAMPLES.map((profile) => (
          <SelectableCard
            key={profile.id}
            selected={Boolean(profile.selected)}
            onSelect={() => {}}
          >
            <div className={styles.profileHeader}>
              <span className="mono">{profile.id}</span>
              <span className={styles.profileName}>{profile.name}</span>
              {profile.selected ? <Icon name="check" /> : null}
            </div>
            <p className="caption">{profile.desc}</p>
            <div className={styles.profileMeta}>
              <span>
                Família: <span>{profile.family}</span>
              </span>
              <span>·</span>
              <span>
                <span className="mono">{profile.count}</span> existentes
              </span>
            </div>
          </SelectableCard>
        ))}
      </div>
      <WizardFooter stepLabel="Etapa 1 de 4 · Selecione um perfil para continuar" showBack={false} />
    </div>
  );
}

export default StepProfile;
