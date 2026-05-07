import { CodeChip } from '../../../../../components/ui/CodeChip';
import { StatusPill } from '../../../../../components/ui/StatusPill';
import type { DocumentProfile, ProcessArea } from '../../../../taxonomy/types';
import type { TemplateDTO } from '../../../../templates/api/templatesV2';
import { VISIBILITY_META, type VisibilityKey } from '../../../lib/visibilityMeta';
import { WizardFooter } from '../WizardShell';
import styles from './StepConfirm.module.css';

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
  const visibilityLabel = VISIBILITY_META[visibility].label;
  const profileLabel = profile ? `${profile.code} — ${profile.name}` : '—';
  const areaLabel = area ? `${area.code} — ${area.name}` : '—';
  const templateLabel = template ? `${template.name} v${template.latest_version} (publicada)` : '—';
  const createdAtLabel = formatDateTime(createdAt);

  const summaryFields: ReadonlyArray<readonly [string, string]> = [
    ['Perfil', profileLabel],
    ['Família', profile?.familyCode ?? '—'],
    ['Área', areaLabel],
    ['Visibilidade', visibilityLabel],
    ['Autor', authorDisplayName || '—'],
    ['Criado em', createdAtLabel],
  ];

  return (
    <div className="card">
      <div className="kicker">Etapa 4 de 4</div>
      <h2 className="h2">Confirme a criação do documento</h2>
      <p className="caption">
        Os campos abaixo serão registrados no slot. O código é definitivo e não pode ser reutilizado.
      </p>

      <div className={styles.previewCard}>
        <div className={styles.docThumbnail}>
          <div className={styles.thumbnailTitleBar} />
          <div className={styles.thumbnailCode}>{codePreview} v1</div>
          {Array.from({ length: 11 }).map((_, idx) => (
            <div
              key={idx}
              className={styles.thumbnailLine}
              style={{ width: `${55 + (idx * 11) % 38}%` }}
            />
          ))}
        </div>
        <div>
          <div className={styles.summaryHeaderRow}>
            <CodeChip>{codePreview}</CodeChip>
            <StatusPill status="draft" />
            <span className="pill mono">v1</span>
          </div>
          <div className={styles.docTitle}>{title || '—'}</div>
          <div className={styles.fieldGrid}>
            {summaryFields.map(([label, value]) => (
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
        <div role="alert" aria-live="assertive" className={`card ${styles.errorAlert}`}>
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

function formatDateTime(date: Date): string {
  try {
    const d = date.toLocaleDateString('pt-BR');
    const t = date.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' });
    return `${d} · ${t}`;
  } catch {
    return date.toISOString();
  }
}

export default StepConfirm;
