import { CodeChip } from '../../../../../components/ui/CodeChip';
import { StatusPill } from '../../../../../components/ui/StatusPill';
import type { DocumentProfile, ProcessArea } from '../../../../taxonomy/types';
import type { TemplateDTO } from '../../../../templates';
import { VISIBILITY_META, type VisibilityKey } from '../../../lib/visibilityMeta';
import { DocPaperPreview } from '../DocPaperPreview';
import { WizardFooter } from '../../../../shared/components/wizard/WizardFooter';
import styles from './StepConfirm.module.css';

type SummaryField = { label: string; value: string };

export type StepConfirmProps = {
  profile: DocumentProfile | null;
  area: ProcessArea | null;
  title: string;
  visibility: VisibilityKey;
  template: TemplateDTO | null;
  authorDisplayName: string;
  createdAt: Date;
  consent: boolean;
  submitting: boolean;
  error: string | null;
  onConsent: (value: boolean) => void;
  onSubmit: () => void;
  onBack: () => void;
  onCancel: () => void;
  submitDisabled: boolean;
};

export function StepConfirm(props: StepConfirmProps): JSX.Element {
  const {
    profile,
    area,
    title,
    visibility,
    template,
    authorDisplayName,
    createdAt,
    consent,
    submitting,
    error,
    onConsent,
    onSubmit,
    onBack,
    onCancel,
    submitDisabled,
  } = props;

  const codePreview = `${profile?.code ?? '???'}-${area?.code ?? '???'}-???`;
  const versionLabel = 'v1'; // first version of a new doc; literal until version-bumping flow exists
  const visibilityLabel = VISIBILITY_META[visibility].label;
  const profileLabel = profile ? `${profile.code} — ${profile.name}` : '—';
  const areaLabel = area ? `${area.code} — ${area.name}` : '—';
  const templateLabel = template ? `${template.name} v${template.latest_version} (publicada)` : '—';
  const createdAtLabel = formatDateTime(createdAt);

  const summaryFields: ReadonlyArray<SummaryField> = [
    { label: 'Perfil', value: profileLabel },
    { label: 'Família', value: profile?.familyCode ?? '—' },
    { label: 'Área', value: areaLabel },
    { label: 'Visibilidade', value: visibilityLabel },
    { label: 'Autor', value: authorDisplayName || '—' },
    { label: 'Criado em', value: createdAtLabel },
  ];

  return (
    <div className="card">
      <div className="kicker">Etapa 4 de 4</div>
      <h2 className="h2">Confirme a criação do documento</h2>
      <p className="caption">
        Os campos abaixo serão registrados no slot. O código é definitivo e não pode ser reutilizado.
      </p>

      <div className={styles.previewCard}>
        <DocPaperPreview lines={11} code={`${codePreview} ${versionLabel}`} variant="thumbnail" />
        <div>
          <div className={styles.summaryHeaderRow}>
            <CodeChip>{codePreview}</CodeChip>
            <StatusPill status="draft" />
            <span className="pill mono">v1</span>
          </div>
          <div className={styles.docTitle}>{title || '—'}</div>
          <div className={styles.fieldGrid}>
            {summaryFields.map(({ label, value }) => (
              <div key={label} className={styles.fieldRow}>
                <span className={styles.fieldLabel}>{label}</span>
                <span>{value}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className={styles.nextStepsCallout}>
        <div className={`${styles.nextStepsKicker} kicker`}>Ao confirmar</div>
        {/* TODO(novo-documento:slot-rollback): if doc-create fails after slot-create
            succeeds, orphan slot persists + code is consumed. No compensation today.
            See wiki/backlog/novo-documento.md#slot-rollback. */}
        <ol className={styles.nextStepsList}>
          <li>
            O slot <span className="mono">{codePreview}</span> será reservado permanentemente — códigos
            não são reutilizados mesmo após arquivamento.
          </li>
          <li>
            Uma cópia do template <span className="mono">{template?.name ?? '—'}</span> será clonada como
            rascunho.
          </li>
          <li>Você será direcionado para o editor para preencher o conteúdo.</li>
          <li>Tokens fixos (código, autor, vigência, aprovadores) serão resolvidos no momento do freeze.</li>
        </ol>
      </div>

      <label className={styles.consentRow} htmlFor="wizard-consent">
        <input
          id="wizard-consent"
          type="checkbox"
          checked={consent}
          onChange={(event) => onConsent(event.target.checked)}
          disabled={submitting}
          aria-describedby="wizard-consent-desc"
        />
        <span id="wizard-consent-desc">
          Confirmo que entendi que o código <span className="mono">{codePreview}</span> é definitivo e não
          pode ser reutilizado.
        </span>
      </label>

      {error ? (
        <div role="alert" aria-live="assertive" aria-atomic="true" className={`card ${styles.errorAlert}`}>
          {error}
        </div>
      ) : null}

      <WizardFooter
        stepLabel={
          submitting
            ? 'Criando documento…'
            : submitDisabled
              ? 'Etapa 4 de 4 · Confirme para criar'
              : 'Etapa 4 de 4 · Tudo pronto para criar'
        }
        primaryLabel={submitting ? 'Criando…' : 'Criar documento'}
        primaryDisabled={submitDisabled}
        primaryVariant="submit"
        onAdvance={onSubmit}
        onBack={onBack}
        onCancel={onCancel}
      />
    </div>
  );
}

const DATE_TIME_FMT = new Intl.DateTimeFormat('pt-BR', {
  dateStyle: 'short',
  timeStyle: 'short',
});

function formatDateTime(date: Date): string {
  try {
    return DATE_TIME_FMT.format(date);
  } catch {
    return date.toISOString();
  }
}

export default StepConfirm;
