import { CodeChip } from '../../../../../components/ui/CodeChip';
import { StatusPill } from '../../../../../components/ui/StatusPill';
import { WizardFooter } from '../WizardShell';
import styles from './StepConfirm.module.css';

type SummaryField = readonly [label: string, value: string];

const SUMMARY_FIELDS: readonly SummaryField[] = [
  ['Perfil', 'POP — Procedimento Op.'],
  ['Família', 'Qualidade'],
  ['Área', 'QUA — Qualidade'],
  ['Template', 'POP v3.2 (publicada)'],
  ['Autor', 'Documento exemplo'],
  ['Criado em', '03/05/2026 · 14:48'],
];

export function StepConfirm(): JSX.Element {
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
          <div className="mono">POP-QUA-??? v1</div>
          {Array.from({ length: 11 }).map((_, idx) => (
            <div key={idx} className={styles.thumbnailLine} />
          ))}
        </div>
        <div>
          <div className={styles.summaryHeaderRow}>
            <CodeChip>POP-QUA-???</CodeChip>
            <StatusPill status="draft" />
            <span className="pill mono">v1</span>
          </div>
          <div className={styles.docTitle}>Documento exemplo</div>
          <div className={styles.fieldGrid}>
            {SUMMARY_FIELDS.map(([label, value]) => (
              <div key={label} className={styles.fieldRow}>
                <span className={styles.fieldLabel}>{label}</span>
                <span className={styles.fieldValue}>{value}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className={styles.nextStepsCallout}>
        <div className="kicker">Ao confirmar</div>
        <ol>
          <li>
            O slot <span className="mono">POP-QUA-???</span> será reservado permanentemente — códigos não são reutilizados mesmo após arquivamento.
          </li>
          <li>
            Uma cópia do template <span className="mono">POP v3.2</span> será clonada como rascunho.
          </li>
          <li>Você será direcionado para o editor para preencher o conteúdo.</li>
          <li>Tokens fixos (código, autor, vigência, aprovadores) serão resolvidos no momento do freeze.</li>
        </ol>
      </div>

      <label className={styles.consentRow}>
        <input type="checkbox" defaultChecked readOnly />
        Confirmo que entendi que o código <span className="mono">POP-QUA-???</span> é definitivo e não pode ser reutilizado.
      </label>

      <WizardFooter
        stepLabel="Etapa 4 de 4 · Tudo pronto para criar"
        primaryLabel="Criar documento"
      />
    </div>
  );
}

export default StepConfirm;
