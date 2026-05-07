import { SelectableCard } from '../../../../../components/ui/SelectableCard';
import { Icon } from '../../../../../components/ui/Icon';
import { resolveQueryError } from '../../../../../lib/api';
import type { TemplateDTO } from '../../../../templates/api/templatesV2';
import { DocPaperPreview } from '../DocPaperPreview';
import { WizardFooter } from '../WizardFooter';
import styles from './StepTemplate.module.css';

export type StepTemplateProps = {
  profileLabel: string;
  templates: TemplateDTO[];
  isLoading: boolean;
  isError: boolean;
  error: unknown;
  onRetry: () => void;
  selectedTemplateID: string | null;
  selectedVersionID: string | null;
  onSelect: (templateID: string, versionID: string | null) => void;
  onAdvance: () => void;
  onBack: () => void;
  onCancel: () => void;
  advanceDisabled: boolean;
};

export function StepTemplate(props: StepTemplateProps): JSX.Element {
  const {
    profileLabel,
    templates,
    isLoading,
    isError,
    error,
    onRetry,
    selectedTemplateID,
    selectedVersionID,
    onSelect,
    onAdvance,
    onBack,
    onCancel,
    advanceDisabled,
  } = props;

  return (
    <div className="card">
      <div className="kicker">Etapa 3 de 4 · {profileLabel}</div>
      <h2 className="h2">Selecione o template a ser clonado</h2>
      <p className="caption">
        O conteúdo do template será copiado para o documento. Tokens fixos serão resolvidos no freeze.
      </p>

      {isLoading ? (
        <div className={styles.templateStack} aria-busy="true">
          <div className="card">
            <div className="caption">Carregando templates…</div>
          </div>
        </div>
      ) : isError ? (
        <div role="alert" aria-live="assertive" className="card">
          {resolveQueryError(error, 'Falha ao carregar templates.')}
          <div>
            <button type="button" className="btn btn-sm" onClick={onRetry}>
              Tentar novamente
            </button>
          </div>
        </div>
      ) : templates.length === 0 ? (
        <div className="card">
          <div className="caption">Nenhum template publicado para este perfil.</div>
        </div>
      ) : (
        <div className={styles.templateStack}>
          {/* TODO(novo-documento:template-versions): only published version is
              selectable; older versions rendered disabled with "Em breve" tooltip.
              GET /api/v2/templates/:id/versions endpoint not yet shipped.
              See wiki/backlog/novo-documento.md#template-versions. */}
          {templates.map((template) => {
            const publishedID = template.published_version_id;
            const selectable = publishedID !== null;
            const selected = isTemplateSelected(template.id, publishedID, selectedTemplateID, selectedVersionID);
            return (
              <SelectableCard
                key={template.id}
                selected={selected}
                disabled={!selectable}
                title={selectable ? undefined : 'Em breve'}
                onSelect={() => {
                  if (!selectable) return;
                  onSelect(template.id, publishedID);
                }}
              >
                <div className={styles.templateRow}>
                  <DocPaperPreview lines={7} variant="template" />
                  <div className={styles.templateMain}>
                    <div className={styles.templateLabelRow}>
                      <span className={styles.templateLabel}>{template.name}</span>
                      <span className="pill mono">v{template.latest_version}</span>
                      {selectable ? (
                        <span className="pill pill-frozen">
                          <span className={styles.publishedDot} aria-hidden="true" />
                          publicada
                        </span>
                      ) : (
                        <span className="pill" title="Em breve">
                          sem versão publicada
                        </span>
                      )}
                    </div>
                    <p className="caption">{template.description ?? ''}</p>
                    <div className={styles.templateMeta}>
                      por {template.created_by} · atualizado {formatDate(template.created_at)}
                    </div>
                  </div>
                  {selected ? <Icon name="check" /> : null}
                </div>
              </SelectableCard>
            );
          })}

          {/* TODO(novo-documento:blank-template): "Em branco" rendered disabled —
              POST /api/v2/documents does not support template_version_id: null today.
              See wiki/backlog/novo-documento.md#blank-template. */}
          <SelectableCard
            selected={false}
            disabled
            title="Em breve"
            onSelect={() => {}}
          >
            <div className={styles.templateRow}>
              <DocPaperPreview lines={0} variant="template" />
              <div className={styles.templateMain}>
                <div className={styles.templateLabelRow}>
                  <span className={styles.templateLabel}>Em branco</span>
                  <span className="pill" title="Em breve">
                    em breve
                  </span>
                </div>
                <p className="caption">Começar sem template (em breve).</p>
              </div>
            </div>
          </SelectableCard>
        </div>
      )}

      <WizardFooter
        stepLabel={
          advanceDisabled
            ? 'Etapa 3 de 4 · Selecione um template para continuar'
            : 'Etapa 3 de 4 · Avance para revisar e confirmar'
        }
        primaryDisabled={advanceDisabled}
        onAdvance={onAdvance}
        onBack={onBack}
        onCancel={onCancel}
      />
    </div>
  );
}

function isTemplateSelected(
  templateID: string,
  publishedVersionID: string | null,
  selectedTemplateID: string | null,
  selectedVersionID: string | null,
): boolean {
  return (
    selectedTemplateID === templateID &&
    selectedVersionID !== null &&
    selectedVersionID === publishedVersionID
  );
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString('pt-BR');
  } catch {
    return iso;
  }
}

export default StepTemplate;
