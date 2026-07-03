import { SelectableCard } from '../../../../../components/ui/SelectableCard';
import { Icon } from '../../../../../components/ui/Icon';
import { InlineAlert } from '../../../../../components/ui/InlineAlert';
import { resolveQueryError } from '../../../../../lib/api';
import { formatRevisionCode } from '../../../../../lib/labels/revisionCode';
import type { TemplateDTO } from '../../../../templates';
import { DocPaperPreview } from '../DocPaperPreview';
import { WizardFooter } from '../../../../shared/components/wizard/WizardFooter';
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
  blankTemplateID: string | null;
  blankTemplateVersionID: string | null;
  blankTemplateName: string;
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
    blankTemplateID,
    blankTemplateVersionID,
    blankTemplateName,
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
        <InlineAlert
          message={resolveQueryError(error, 'Falha ao carregar templates.')}
          action={
            <button type="button" className="btn btn-sm" onClick={onRetry}>
              Tentar novamente
            </button>
          }
        />
      ) : (
        <div className={styles.templateStack}>
          {templates.length === 0 ? (
            <div className="card">
              <div className="caption">Nenhum template publicado para este perfil.</div>
            </div>
          ) : null}
          {/* ADR 0065: selectability and badge text now gate on the nested
              published_version / latest_version VersionRef objects — the whole
              object is checked for null, never an inner field. */}
          {templates.map((template) => {
            const publishedVersion = template.published_version;
            const selectable = publishedVersion != null;
            const publishedID = publishedVersion?.id ?? null;
            const selected = isTemplateSelected(template.id, publishedID, selectedTemplateID, selectedVersionID);
            const unselectableBadgeText = badgeTextForStatus(template.latest_version.status);
            return (
              <SelectableCard
                key={template.id}
                selected={selected}
                disabled={!selectable}
                title={selectable ? undefined : 'Em breve'}
                onSelect={() => {
                  if (!selectable || !publishedVersion) return;
                  onSelect(template.id, publishedVersion.id);
                }}
              >
                <div className={styles.templateRow}>
                  <DocPaperPreview lines={7} variant="template" />
                  <div className={styles.templateMain}>
                    <div className={styles.templateLabelRow}>
                      <span className={styles.templateLabel}>{template.name}</span>
                      {/* ADR 0013: render REV{nn} from the canonical contract field. */}
                      <span className="pill mono">{formatRevisionCode(template.latest_version.revision_number)}</span>
                      {selectable ? (
                        <span className="pill pill-frozen">
                          <span className={styles.publishedDot} aria-hidden="true" />
                          publicada
                        </span>
                      ) : (
                        <span className="pill" title="Em breve">
                          {unselectableBadgeText}
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

          <SelectableCard
            selected={isTemplateSelected(
              blankTemplateID ?? '',
              blankTemplateVersionID,
              selectedTemplateID,
              selectedVersionID,
            )}
            disabled={blankTemplateID === null || blankTemplateVersionID === null}
            title={blankTemplateVersionID ? undefined : 'Em breve'}
            onSelect={() => {
              if (!blankTemplateID || !blankTemplateVersionID) return;
              onSelect(blankTemplateID, blankTemplateVersionID);
            }}
          >
            <div className={styles.templateRow}>
              <DocPaperPreview lines={0} variant="template" />
              <div className={styles.templateMain}>
                <div className={styles.templateLabelRow}>
                  <span className={styles.templateLabel}>{blankTemplateName}</span>
                  {blankTemplateVersionID ? (
                    <span className="pill pill-frozen">
                      <span className={styles.publishedDot} aria-hidden="true" />
                      disponivel
                    </span>
                  ) : (
                    <span className="pill" title="Em breve">
                      em breve
                    </span>
                  )}
                </div>
                <p className="caption">Começar sem um template base.</p>
              </div>
              {isTemplateSelected(
                blankTemplateID ?? '',
                blankTemplateVersionID,
                selectedTemplateID,
                selectedVersionID,
              ) ? (
                <Icon name="check" />
              ) : null}
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

function badgeTextForStatus(status: string): string {
  switch (status) {
    case 'under_review':
      return 'em revisão';
    case 'approved':
      return 'aguardando publicação';
    default:
      return 'sem versão publicada';
  }
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString('pt-BR');
  } catch {
    return iso;
  }
}

export default StepTemplate;
