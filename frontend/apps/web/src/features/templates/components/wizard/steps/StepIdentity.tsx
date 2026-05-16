import type { DocumentProfile } from '../../../../taxonomy/types';
import { WizardFooter } from '../../../../shared/components/wizard/WizardFooter';
import type { ScopeType } from '../../../state/templateWizard.reducer';
import styles from './StepIdentity.module.css';

export type StepIdentityProps = {
  scopeType: ScopeType;
  selectedProfile: DocumentProfile | null;
  templateKey: string;
  name: string;
  description: string;
  onChangeName: (value: string) => void;
  onChangeDescription: (value: string) => void;
  onAdvance: () => void;
  onBack: () => void;
  onChangeScope: () => void;
  advanceDisabled: boolean;
  keyError?: string | null;
};

export function StepIdentity({
  scopeType,
  selectedProfile,
  templateKey,
  name,
  description,
  onChangeName,
  onChangeDescription,
  onAdvance,
  onBack,
  onChangeScope,
  advanceDisabled,
  keyError = null,
}: StepIdentityProps): JSX.Element {
  const kicker =
    scopeType === 'generic'
      ? 'Etapa 2 de 5 · Template genérico'
      : selectedProfile
        ? `Etapa 2 de 5 · Perfil ${selectedProfile.code} - ${selectedProfile.name}`
        : 'Etapa 2 de 5';

  // TODO(novo-template-wizard:next-code-preview): replace the slug key preview
  // with a server sequence only when the next-code endpoint ships.
  return (
    <div className="card">
      <div className="kicker">{kicker}</div>
      <h2 className="h2">Identidade do template</h2>

      {/* Scope recap row (profile + version) */}
      {(() => {
        const recap =
          scopeType === 'profile' && selectedProfile
            ? {
                code: selectedProfile.code,
                name: selectedProfile.name,
                detail: `Família: ${selectedProfile.familyCode}`,
              }
            : {
                code: 'GEN',
                name: 'Genérico',
                detail: 'Aplicável a qualquer perfil de documento.',
              };
        return (
          <div className={styles.recapRow}>
            <div className={styles.recapField}>
              <span className="kicker">Escopo selecionado</span>
              <div className={styles.recapBox}>
                <span className="code-chip mono">{recap.code}</span>
                <div className={styles.recapMeta}>
                  <div className={styles.recapName}>{recap.name}</div>
                  <div className="caption">{recap.detail}</div>
                </div>
                <button
                  type="button"
                  className="btn btn-sm btn-ghost"
                  onClick={onChangeScope}
                  aria-label="Trocar escopo do template"
                >
                  Trocar
                </button>
              </div>
            </div>

            <div className={styles.recapField}>
              <span className="kicker">Versão inicial</span>
              <div className={styles.recapBox}>
                <span className={`mono ${styles.versionLabel}`}>v1.0</span>
                <span className={styles.versionHint}>
                  Atribuída automaticamente · incrementa em cada publicação
                </span>
              </div>
            </div>
          </div>
        );
      })()}

      {/* Name */}
      <div className={styles.field}>
        <label htmlFor="tpl-name" className="kicker">
          Nome do template *
        </label>
        <input
          id="tpl-name"
          type="text"
          className={`input ${styles.nameInput}`}
          value={name}
          onChange={(e) => onChangeName(e.target.value)}
          placeholder="Ex.: Inspeção de Recebimento de Matéria-Prima"
          autoComplete="off"
          maxLength={120}
          required
          aria-required="true"
          aria-invalid={keyError ? 'true' : 'false'}
          aria-describedby={keyError ? 'tpl-name-hint tpl-name-error' : 'tpl-name-hint'}
        />
        <span id="tpl-name-hint" className={styles.fieldHint}>
          Aparece para os autores que forem cloná-lo no wizard de novo documento.
        </span>
        {keyError && (
          <span id="tpl-name-error" role="alert" className={styles.fieldHint}>
            {keyError}
          </span>
        )}
      </div>

      {/* Description */}
      <div className={styles.field}>
        <label htmlFor="tpl-description" className="kicker">
          Descrição
        </label>
        <textarea
          id="tpl-description"
          className={`input ${styles.descriptionInput}`}
          rows={3}
          value={description}
          onChange={(e) => onChangeDescription(e.target.value)}
          placeholder="Resumo do que este template cobre."
          maxLength={500}
        />
      </div>

      {/* Honest key preview. Next-code sequencing remains deferred in the backlog. */}
      <div className={styles.codePreview}>
        <div className={`kicker ${styles.codePreviewKicker}`}>Identificador tecnico</div>
        <div className={`mono ${styles.codePreviewValue}`}>{templateKey || 'preencha-o-nome'}</div>
        <div className={styles.codePreviewHint}>
          Gerado a partir do nome nesta versão. A prévia de código sequencial permanece no backlog.
        </div>
      </div>

      <WizardFooter
        stepLabel={
          advanceDisabled
            ? 'Etapa 2 de 5 · Informe o nome para continuar'
            : 'Etapa 2 de 5 · Pronto para avançar'
        }
        showBack
        onBack={onBack}
        onAdvance={onAdvance}
        primaryDisabled={advanceDisabled}
      />
    </div>
  );
}

export default StepIdentity;
