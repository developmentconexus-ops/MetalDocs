import { Icon } from '../../../../components/ui/Icon';
import { MOCK_DISTRIBUTION } from '../../lib/distributionMeta';
import styles from './DistributionFacts.module.css';

const { publishedAtFull, deadlineFull, daysLeft, policy, channel, remindersSent, remindersScheduled, byArea, people } = MOCK_DISTRIBUTION;

const uniqueRoles = new Set(people.map(p => p.role)).size;

const FACTS = [
  { icon: 'calendar' as const, label: 'Distribuído em',    value: publishedAtFull,                      mono: false, hint: '',                                  tone: ''        },
  { icon: 'clock'    as const, label: 'Prazo',             value: deadlineFull,                          mono: false, hint: `${daysLeft} dias restantes`,        tone: 'warning' },
  { icon: 'shield'   as const, label: 'Política',          value: policy,                               mono: false, hint: '',                                  tone: ''        },
  { icon: 'mail'     as const, label: 'Canal',             value: channel,                              mono: false, hint: '',                                  tone: ''        },
  { icon: 'bell'     as const, label: 'Lembretes enviados',value: String(remindersSent),                mono: true,  hint: `próximo: ${remindersScheduled}`,    tone: ''        },
  { icon: 'users'    as const, label: 'Grupos',            value: `${byArea.length} áreas · ${uniqueRoles} cargos`, mono: false, hint: '',                      tone: ''        },
];

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
