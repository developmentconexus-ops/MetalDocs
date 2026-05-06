import { SelectableCard } from '../../../../../components/ui/SelectableCard';
import { Icon } from '../../../../../components/ui/Icon';
import { WizardFooter } from '../WizardShell';
import styles from './StepTemplate.module.css';

type TemplateSample = {
  id: string;
  version: string;
  date: string;
  author: string;
  label: string;
  notes: string;
  current?: boolean;
  warning?: boolean;
  blank?: boolean;
  selected?: boolean;
};

// TODO(novo-documento): Phase 3c replaces this hardcoded sample with
// `useTemplatesByProfileQuery(profileCode)` results. "Em branco" remains
// disabled (no backend path); older versions remain disabled until
// `GET /api/v2/templates/:id/versions` ships. See worksheet §1.6.
const TEMPLATE_SAMPLES: TemplateSample[] = [
  {
    id: 'latest',
    version: 'v3.2',
    date: '12 dias atrás',
    author: 'Documento exemplo',
    label: 'Versão publicada (recomendada)',
    notes: 'Documento exemplo — notas da versão publicada.',
    current: true,
    selected: true,
  },
  {
    id: 'v31',
    version: 'v3.1',
    date: '3 meses atrás',
    author: 'Documento exemplo',
    label: 'Versão anterior',
    notes: 'Documento exemplo — notas da versão anterior.',
  },
  {
    id: 'blank',
    version: '—',
    date: '',
    author: '',
    label: 'Em branco',
    notes: 'Documento exemplo — começar sem template.',
    warning: true,
    blank: true,
  },
];

export function StepTemplate(): JSX.Element {
  return (
    <div className="card">
      <div className="kicker">Etapa 3 de 4 · Perfil POP — Procedimento Operacional</div>
      <h2 className="h2">Selecione o template a ser clonado</h2>
      <p className="caption">
        O conteúdo do template será copiado para o documento. Tokens fixos serão resolvidos no freeze.
      </p>
      <div className={styles.templateStack}>
        {TEMPLATE_SAMPLES.map((template) => (
          <SelectableCard
            key={template.id}
            selected={Boolean(template.selected)}
            onSelect={() => {}}
          >
            <div className={styles.templateRow}>
              <div className={styles.templatePreview}>
                <div className={styles.previewTitleBar} />
                {template.blank
                  ? null
                  : Array.from({ length: 7 }).map((_, idx) => (
                      <div key={idx} className={styles.previewLine} />
                    ))}
              </div>
              <div className={styles.templateMain}>
                <div className={styles.templateLabelRow}>
                  <span className={styles.templateLabel}>{template.label}</span>
                  {template.version !== '—' ? (
                    <span className="pill mono">{template.version}</span>
                  ) : null}
                  {template.current ? (
                    <span className="pill pill-frozen">
                      <span className="dot" />
                      publicada
                    </span>
                  ) : null}
                  {template.warning ? <span className="pill">cuidado</span> : null}
                </div>
                <p className="caption">{template.notes}</p>
                {template.author ? (
                  <div className={styles.templateMeta}>
                    por {template.author} · atualizado {template.date}
                  </div>
                ) : null}
              </div>
              {template.selected ? <Icon name="check" /> : null}
            </div>
          </SelectableCard>
        ))}
      </div>
      <WizardFooter stepLabel="Etapa 3 de 4 · Avance para revisar e confirmar" />
    </div>
  );
}

export default StepTemplate;
