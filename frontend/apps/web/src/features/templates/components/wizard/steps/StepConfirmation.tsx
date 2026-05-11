import { useState } from 'react';
import { useAuthStore } from '../../../../../store/auth.store';
import { StatusPill } from '../../../../../components/ui/StatusPill';
import { WizardFooter } from '../../../../shared/components/wizard/WizardFooter';
import styles from './StepConfirmation.module.css';

export type StepConfirmationProps = {
  name: string;
  selectedProfile: { code: string; name: string; family: string } | null;
  profileCode: string | null;
  startingPoint: 'docx' | 'blank' | null;
  selectedDocxName: string | null;
  permissionsMode: 'roles' | 'areas' | 'all';
  selectedRoleIds: string[];
  selectedAreaIds: string[];
  onBack: () => void;
  onSubmit: () => void;
  isSubmitting?: boolean;
  submitError?: string | null;
};

// k % 3 === 1 pattern: indices 1, 4, 7, 10 — matches design reference
const HIGHLIGHTED_THUMB_LINES = new Set([1, 4, 7, 10]);

export function StepConfirmation({
  name,
  selectedProfile,
  profileCode,
  startingPoint,
  selectedDocxName,
  permissionsMode,
  selectedRoleIds,
  selectedAreaIds,
  onBack,
  onSubmit,
  isSubmitting = false,
  submitError = null,
}: StepConfirmationProps): JSX.Element {
  const [confirmed, setConfirmed] = useState(false);
  const user = useAuthStore((s) => s.user);

  // TODO(novo-template-wizard:next-code-preview) same backlog as Step 2 (wiki/backlog/novo-template-wizard.md)
  const mockCode = profileCode ? `TPL-${profileCode}-001` : 'TPL-GEN-001';
  const perfilValue = profileCode
    ? `${profileCode}—${selectedProfile?.name ?? ''}`
    : 'Generico';
  const familiaValue = selectedProfile?.family ?? '—';
  const origemValue =
    startingPoint === 'docx' ? selectedDocxName ?? '.docx' : 'Em branco';
  const permissoesValue =
    permissionsMode === 'all'
      ? 'Toda a empresa'
      : permissionsMode === 'areas'
        ? `${selectedAreaIds.length} area(s) selecionada(s)`
        : `${selectedRoleIds.length} perfil(s) selecionado(s)`;
  const autorValue = user?.displayName ?? '—';

  function handleSubmit() {
    onSubmit();
  }

  return (
    <div className="card">
      <div className="kicker">Etapa 5 de 5</div>
      <h2 className={`h2 ${styles.stepTitle}`}>Revise e confirme a criação</h2>
      <p className={`caption ${styles.intro}`}>
        O template será criado como rascunho. Você pode editar a estrutura no editor visual antes de submeter para revisão.
      </p>

      <div className={styles.previewCard}>
        <div className={styles.thumb}>
          <div className={styles.thumbHeader} />
          <div className={styles.thumbCode}>{mockCode} v1.0</div>
          {Array.from({ length: 11 }, (_, index) => (
            <div
              key={index}
              className={`${styles.thumbLine} ${
                HIGHLIGHTED_THUMB_LINES.has(index) ? styles.thumbLineHighlight : ''
              }`}
            />
          ))}
        </div>

        <div className={styles.previewBody}>
          <div className={styles.headerRow}>
            <span className="code-chip mono">{mockCode}</span>
            <StatusPill status="draft" />
            <span className="pill mono">v1.0</span>
          </div>
          <div className={styles.templateName}>{name}</div>
          <div className={styles.metaGrid}>
            <div className={styles.metaRow}>
              <span className={styles.metaLabel}>Perfil</span>
              <span className={styles.metaValue}>{perfilValue}</span>
            </div>
            <div className={styles.metaRow}>
              <span className={styles.metaLabel}>Familia</span>
              <span className={styles.metaValue}>{familiaValue}</span>
            </div>
            <div className={styles.metaRow}>
              <span className={styles.metaLabel}>Origem</span>
              <span className={styles.metaValue}>{origemValue}</span>
            </div>
            <div className={styles.metaRow}>
              <span className={styles.metaLabel}>Permissoes</span>
              <span className={styles.metaValue}>{permissoesValue}</span>
            </div>
            <div className={styles.metaRow}>
              <span className={styles.metaLabel}>Autor</span>
              <span className={styles.metaValue}>{autorValue}</span>
            </div>
          </div>
        </div>
      </div>

      <div className={styles.confirmBlock}>
        <div className={`kicker ${styles.confirmKicker}`}>Ao confirmar</div>
        <ol className={styles.confirmList}>
          <li>O template receberá o código acima, que é definitivo.</li>
          <li>Uma versão 1.0 em rascunho será criada automaticamente.</li>
          <li>O editor de documento será aberto para você montar a estrutura.</li>
          <li>O template ficará visível conforme as permissões definidas.</li>
        </ol>
      </div>

      <label className={styles.checkLabel}>
        <input
          type="checkbox"
          checked={confirmed}
          onChange={(e) => setConfirmed(e.target.checked)}
        />
        <span>
          Confirmo que entendi que o código do template é definitivo e não pode ser reutilizado.
        </span>
      </label>

      {submitError && (
        <div role="alert" className={styles.submitError}>
          {submitError}
        </div>
      )}

      <WizardFooter
        stepLabel={
          isSubmitting
            ? 'Etapa 5 de 5 · Criando template…'
            : confirmed
              ? 'Etapa 5 de 5 · Tudo pronto para criar'
              : 'Etapa 5 de 5 · Confirme para continuar'
        }
        showBack
        onBack={onBack}
        primaryLabel={isSubmitting ? 'Criando…' : 'Criar e abrir editor →'}
        primaryDisabled={!confirmed || isSubmitting}
        onAdvance={handleSubmit}
      />
    </div>
  );
}

export default StepConfirmation;
