import { useId, useMemo, useState } from "react";
import type { Ref } from "react";
import { Avatar } from "../../../components/ui/Avatar";
import { Icon } from "../../../components/ui/Icon";
import { useRovingRadioGroup } from "../../../components/ui/useRovingRadioGroup";
import type { ArtifactDecisionModel, ArtifactDecisionOption } from "./types";
import styles from "./ArtifactDecisionPanel.module.css";

interface ArtifactDecisionPanelProps {
  model: ArtifactDecisionModel;
}

/** Reads `.stale` (set by SignoffError) so the panel can show a refresh banner. */
function isStale(error: unknown): boolean {
  return typeof error === "object" && error !== null && "stale" in error && Boolean((error as { stale: unknown }).stale);
}

function errorMessage(error: unknown): string {
  if (error instanceof Error && error.message) return error.message;
  return "Não foi possível registrar a decisão. Tente novamente.";
}

/**
 * Stateful decision surface for the approval cockpit sidebar. Owns ALL local UI
 * state — selected option, justificativa, password, legal confirmation, in-flight
 * and error/stale — and calls `model.submit`. Kind-agnostic: documents supply a
 * password + legal block, templates supply neither; the panel renders whatever the
 * `ArtifactDecisionModel` declares.
 *
 * Mirrors the detalhe-signoff decision box: radio cards, justificativa (obligatory
 * for return/reject), an honest signer block (real identity + "gerado no ato" note,
 * no fabricated hash/IP), the optional legal-effect checkbox, and a submit footer
 * whose label follows the selected option.
 */
export function ArtifactDecisionPanel({ model }: ArtifactDecisionPanelProps) {
  const [selectedKey, setSelectedKey] = useState<string | null>(null);
  const [reason, setReason] = useState("");
  const [password, setPassword] = useState("");
  const [legalChecked, setLegalChecked] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [stale, setStale] = useState(false);

  const reasonFieldId = useId();
  const passwordFieldId = useId();

  const selectedIndex = model.options.findIndex((o) => o.key === selectedKey);
  const selected: ArtifactDecisionOption | null = selectedIndex >= 0 ? model.options[selectedIndex] : null;

  const roving = useRovingRadioGroup({
    count: model.options.length,
    selectedIndex,
    onSelect: (i) => setSelectedKey(model.options[i].key),
    orientation: "vertical",
    ariaLabel: model.heading,
  });

  const reasonRequired = selected?.requiresReason ?? false;
  const reasonMissing = reasonRequired && reason.trim().length === 0;
  const passwordMissing = model.password != null && password.length === 0;
  const legalMissing = model.legal != null && !legalChecked;

  const canSubmit = useMemo(
    () => selected != null && !reasonMissing && !passwordMissing && !legalMissing && !submitting,
    [selected, reasonMissing, passwordMissing, legalMissing, submitting],
  );

  const submitLabel = submitting ? "Registrando…" : selected != null ? selected.submitLabel : "Selecione uma decisão";

  async function handleSubmit() {
    if (!canSubmit || selected == null) return;
    setSubmitting(true);
    setError(null);
    setStale(false);
    try {
      await model.submit({ optionKey: selected.key, reason: reason.trim(), password });
      // Success: query invalidation re-renders the next lifecycle state. Clear the
      // secret regardless so it never lingers in memory across a state change.
      setPassword("");
    } catch (err) {
      if (isStale(err)) {
        setStale(true);
      } else {
        setError(errorMessage(err));
      }
      setPassword("");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className={styles.panel}>
      <div className={styles.scroll}>
        <div className={styles.kicker}>Decisão</div>

        {/* Decision radio cards */}
        <div className={styles.options} {...roving.groupProps}>
          {model.options.map((option, index) => {
            const isSelected = option.key === selectedKey;
            const itemProps = roving.getItemProps(index);
            return (
              <button
                key={option.key}
                type="button"
                role="radio"
                aria-checked={isSelected}
                data-option={option.key}
                data-tone={option.tone}
                className={`${styles.option} ${isSelected ? styles.optionSelected : ""}`}
                ref={itemProps.ref as Ref<HTMLButtonElement>}
                tabIndex={itemProps.tabIndex}
                onKeyDown={itemProps.onKeyDown}
                onClick={() => setSelectedKey(option.key)}
              >
                <span className={styles.radio} aria-hidden="true">
                  {isSelected && <span className={styles.radioInner} />}
                </span>
                <span className={styles.optionBody}>
                  <span className={styles.optionLabel}>{option.label}</span>
                  <span className={styles.optionDesc}>{option.description}</span>
                </span>
              </button>
            );
          })}
        </div>

        {/* Justificativa */}
        <label className={styles.field} htmlFor={reasonFieldId}>
          <span className={styles.kicker}>
            {model.reasonLabel}
            {reasonRequired && <span className={styles.required}> · obrigatória</span>}
          </span>
          <textarea
            id={reasonFieldId}
            className={styles.textarea}
            rows={3}
            value={reason}
            placeholder={model.reasonPlaceholder}
            onChange={(e) => setReason(e.target.value)}
            disabled={submitting}
            aria-invalid={reasonMissing ? "true" : "false"}
          />
        </label>

        {/* Signer block + optional legal e-signature password */}
        {model.signer != null && (
          <div className={styles.signer}>
            <div className={styles.kicker}>Assinatura</div>
            <div className={styles.signerIdentity}>
              <Avatar name={model.signer.name} size="md" />
              <div className={styles.signerText}>
                <div className={styles.signerName}>{model.signer.name}</div>
                {model.signer.detail != null && <div className={styles.signerDetail}>{model.signer.detail}</div>}
              </div>
            </div>
            {model.password != null && (
              <label className={styles.passwordField} htmlFor={passwordFieldId}>
                <span className={styles.passwordLabel}>{model.password.label}</span>
                <input
                  id={passwordFieldId}
                  className={styles.passwordInput}
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  autoComplete="current-password"
                  disabled={submitting}
                  required
                />
              </label>
            )}
            {model.signer.note != null && <div className={styles.signerNote}>{model.signer.note}</div>}
          </div>
        )}

        {/* Legal-effect confirmation */}
        {model.legal != null && (
          <label className={styles.legal}>
            <input
              type="checkbox"
              checked={legalChecked}
              onChange={(e) => setLegalChecked(e.target.checked)}
              disabled={submitting}
            />
            <span>{model.legal.text}</span>
          </label>
        )}

        {/* Stale / error banners */}
        {stale && (
          <div className={styles.stale} role="status">
            O documento foi alterado. Atualize a página antes de tentar novamente.
          </div>
        )}
        {error != null && (
          <div className={styles.error} role="alert">
            {error}
          </div>
        )}
      </div>

      {/* Submit footer */}
      <div className={styles.footer}>
        <button
          type="button"
          className={styles.submit}
          data-tone={selected?.tone ?? "none"}
          disabled={!canSubmit}
          onClick={() => void handleSubmit()}
        >
          <Icon name="lock" size={13} />
          {submitLabel}
        </button>
      </div>
    </div>
  );
}
