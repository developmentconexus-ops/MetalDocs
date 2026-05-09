import { Icon } from '../../../../components/ui/Icon';
import styles from './DistributionFacts.module.css';

const FACTS = [
  { icon: 'calendar' as const, label: 'Distribuído em',    value: '12 mar 2026 · 14:32', mono: false, hint: '',                     tone: ''        },
  { icon: 'clock'    as const, label: 'Prazo',             value: '22 mar 2026 · 23:59', mono: false, hint: '8 dias restantes',      tone: 'warning' },
  { icon: 'shield'   as const, label: 'Política',          value: 'Reconhecimento obrigatório (assinatura)', mono: false, hint: '', tone: ''        },
  { icon: 'mail'     as const, label: 'Canal',             value: 'E-mail + portal interno', mono: false, hint: '',                 tone: ''        },
  { icon: 'bell'     as const, label: 'Lembretes enviados',value: '2',                   mono: true,  hint: 'próximo: 20 mar · 08:00', tone: ''    },
  { icon: 'users'    as const, label: 'Grupos',            value: '8 áreas · 11 cargos', mono: false, hint: '',                     tone: ''        },
] as const;

export function DistributionFacts() {
  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <div className={styles.kicker}>Detalhes da distribuição</div>
        <button type="button" aria-disabled="true" title="Em breve" className={styles.ctaDisabled}>
          <Icon name="edit" size={11} />
          Editar política
        </button>
      </div>
      <div className={styles.list}>
        {FACTS.map((f, i) => (
          <div
            key={f.label}
            className={`${styles.factRow} ${i < FACTS.length - 1 ? styles.factRowBorder : ''}`}
          >
            <div className={`${styles.iconBox} ${f.tone === 'warning' ? styles.iconBoxWarning : ''}`}>
              <Icon name={f.icon} size={14} />
            </div>
            <div className={styles.factBody}>
              <div className={styles.factLabel}>{f.label}</div>
              <div className={`${styles.factValue} ${f.mono ? styles.factValueMono : ''}`}>
                {f.value}
              </div>
            </div>
            {f.hint ? (
              <div className={styles.factHint}>{f.hint}</div>
            ) : null}
          </div>
        ))}
      </div>
    </div>
  );
}
