import { useState } from "react";
import type { DocumentProfile } from "../types";
import {
  useCreateProfileMutation,
  useSetDefaultTemplateMutation,
  useUpdateProfileMutation,
} from "../queries/useTaxonomyMutations";
import { useFamiliesQuery } from "../queries/useFamiliesQuery";
import { usePublishedTemplatesQuery } from "../queries/usePublishedTemplatesQuery";
import { resolveErrorMessage } from "../../../lib/api/problem";
import styles from "./TaxonomyDialog.module.css";

type Props = {
  mode: "create" | "edit";
  profile?: DocumentProfile;
  onClose: () => void;
};

const EDITABLE_ROLE_OPTIONS = ["viewer", "editor", "author", "approver", "system_admin"];

export function ProfileEditDialog({ mode, profile, onClose }: Props) {
  const [code, setCode] = useState(profile?.code ?? "");
  const [familyCode, setFamilyCode] = useState(profile?.familyCode ?? "");
  const [name, setName] = useState(profile?.name ?? "");
  const [description, setDescription] = useState(profile?.description ?? "");
  const [reviewIntervalDays, setReviewIntervalDays] = useState(String(profile?.reviewIntervalDays ?? 365));
  const [editableByRole, setEditableByRole] = useState(profile?.editableByRole ?? "admin");
  const [templateVersionId, setTemplateVersionId] = useState(profile?.defaultTemplateVersionId ?? "");
  const [error, setError] = useState("");
  const [templateError, setTemplateError] = useState("");

  const familiesQuery = useFamiliesQuery();
  const publishedTemplatesQuery = usePublishedTemplatesQuery(mode === "edit");
  const publishedTemplates = publishedTemplatesQuery.data ?? [];

  const createMutation = useCreateProfileMutation();
  const updateMutation = useUpdateProfileMutation();
  const setDefaultTemplateMutation = useSetDefaultTemplateMutation();
  const saving = createMutation.isPending || updateMutation.isPending;

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    try {
      if (mode === "create") {
        await createMutation.mutateAsync({
          code: code.trim(),
          familyCode: familyCode.trim(),
          name: name.trim(),
          description: description.trim() || undefined,
          reviewIntervalDays: Number(reviewIntervalDays),
          editableByRole: editableByRole.trim() || undefined,
        });
      } else {
        await updateMutation.mutateAsync({
          code: profile!.code,
          req: {
            familyCode: familyCode.trim(),
            name: name.trim(),
            description: description.trim() || undefined,
            editableByRole: editableByRole.trim() || undefined,
            reviewIntervalDays: Number(reviewIntervalDays),
          },
        });
      }
      onClose();
    } catch (err) {
      setError(resolveErrorMessage(err));
    }
  }

  async function handleSetTemplate(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setTemplateError("");
    try {
      await setDefaultTemplateMutation.mutateAsync({
        code: profile!.code,
        req: { templateVersionId: templateVersionId.trim() },
      });
      setTemplateVersionId("");
    } catch (err) {
      setTemplateError(resolveErrorMessage(err));
    }
  }

  return (
    <div className={styles.overlay} onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}>
      <div className={styles.panel}>
        <h2 className={styles.title}>
          {mode === "create" ? "Novo Perfil Documental" : "Editar Perfil Documental"}
        </h2>
        <form onSubmit={(e) => void handleSubmit(e)}>
          {mode === "create" ? (
            <div className={styles.field}>
              <label className={styles.label}>Código *</label>
              <input
                value={code}
                onChange={(e) => setCode(e.target.value.toLowerCase())}
                required
                className={styles.input}
              />
            </div>
          ) : (
            <div className={styles.field}>
              <label className={styles.label}>Código</label>
              <input value={profile?.code ?? ""} readOnly className={styles.inputReadOnly} />
            </div>
          )}
          <div className={styles.field}>
            <label className={styles.label}>Família *</label>
            {mode === "create" ? (
              <select
                value={familyCode}
                onChange={(e) => setFamilyCode(e.target.value)}
                required
                className={styles.select}
              >
                <option value="" disabled>Selecione a família</option>
                {(familiesQuery.data ?? []).map((f) => (
                  <option key={f.code} value={f.code}>{f.name}</option>
                ))}
              </select>
            ) : (
              <input value={profile?.familyCode ?? ""} readOnly className={styles.inputReadOnly} />
            )}
          </div>
          <div className={styles.field}>
            <label className={styles.label}>Nome *</label>
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              className={styles.input}
            />
          </div>
          <div className={styles.field}>
            <label className={styles.label}>Descrição</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={2}
              className={styles.input}
            />
          </div>
          <div className={styles.field}>
            <label className={styles.label}>Intervalo de revisão (dias) *</label>
            <input
              type="number"
              min={1}
              value={reviewIntervalDays}
              onChange={(e) => setReviewIntervalDays(e.target.value)}
              required
              className={styles.input}
            />
          </div>
          <div className={styles.fieldLast}>
            <label className={styles.label}>Role editora *</label>
            <select
              value={editableByRole}
              onChange={(e) => setEditableByRole(e.target.value)}
              required
              className={styles.select}
            >
              {EDITABLE_ROLE_OPTIONS.map((role) => (
                <option key={role} value={role}>{role}</option>
              ))}
            </select>
          </div>
          {error && <p className={styles.error} role="alert">{error}</p>}
          <div className={styles.actions}>
            <button type="button" onClick={onClose} className={styles.buttonSecondary}>Cancelar</button>
            <button type="submit" disabled={saving} className={styles.buttonPrimary}>
              {saving ? "Salvando..." : "Salvar"}
            </button>
          </div>
        </form>

        {mode === "edit" && (
          <div className={styles.templateSection}>
            <h3 className={styles.templateSectionTitle}>Template padrão</h3>
            <form onSubmit={(e) => void handleSetTemplate(e)}>
              <div className={styles.field}>
                <label className={styles.label}>Template padrão</label>
                <select
                  value={templateVersionId}
                  onChange={(e) => setTemplateVersionId(e.target.value)}
                  required
                  className={styles.select}
                >
                  <option value="" disabled>Selecione um template publicado</option>
                  {publishedTemplates.map((t) => (
                    <option key={t.id} value={t.published_version.id}>
                      {t.name}{t.doc_type_code ? ` [${t.doc_type_code}]` : ""}
                    </option>
                  ))}
                </select>
                {publishedTemplates.length === 0 && (
                  <p className={styles.hint}>Nenhum template publicado encontrado.</p>
                )}
              </div>
              {templateError && <p className={styles.error} role="alert">{templateError}</p>}
              <button type="submit" disabled={setDefaultTemplateMutation.isPending} className={styles.buttonPrimary}>
                {setDefaultTemplateMutation.isPending ? "Definindo..." : "Definir template padrão"}
              </button>
            </form>
          </div>
        )}
      </div>
    </div>
  );
}
