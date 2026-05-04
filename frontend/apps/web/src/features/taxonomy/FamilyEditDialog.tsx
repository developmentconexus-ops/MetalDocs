import { useState } from "react";
import type { DocumentFamily, CreateFamilyRequest, UpdateFamilyRequest } from "./types";
import { createFamily, updateFamily } from "./api";

type Props = {
  mode: "create" | "edit";
  family?: DocumentFamily;
  onClose: () => void;
  onSaved: () => void;
};

export function FamilyEditDialog({ mode, family, onClose, onSaved }: Props) {
  const [code, setCode] = useState(family?.code ?? "");
  const [name, setName] = useState(family?.name ?? "");
  const [description, setDescription] = useState(family?.description ?? "");
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setSaving(true);
    try {
      if (mode === "create") {
        const req: CreateFamilyRequest = {
          code: code.trim(),
          name: name.trim(),
          description: description.trim() || undefined,
        };
        await createFamily(req);
      } else {
        const req: UpdateFamilyRequest = {
          name: name.trim(),
          description: description.trim() || undefined,
        };
        await updateFamily(family!.code, req);
      }
      onSaved();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Falha ao salvar.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div
      style={{
        position: "fixed", inset: 0, background: "rgba(0,0,0,0.4)", zIndex: 1000,
        display: "flex", alignItems: "center", justifyContent: "center",
      }}
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
    >
      <div style={{ background: "#fff", borderRadius: 8, padding: 24, minWidth: 400, maxWidth: 520, width: "100%" }}>
        <h2 style={{ margin: "0 0 16px", fontSize: 16 }}>
          {mode === "create" ? "Nova Família Documental" : "Editar Família Documental"}
        </h2>
        <form onSubmit={(e) => void handleSubmit(e)}>
          {mode === "create" && (
            <div style={{ marginBottom: 12 }}>
              <label style={{ display: "block", fontSize: 12, marginBottom: 4 }}>Codigo *</label>
              <input
                value={code}
                onChange={(e) => setCode(e.target.value.toLowerCase())}
                required
                style={{ width: "100%", padding: "6px 8px", boxSizing: "border-box" }}
              />
            </div>
          )}
          {mode === "edit" && (
            <div style={{ marginBottom: 12 }}>
              <label style={{ display: "block", fontSize: 12, marginBottom: 4 }}>Codigo</label>
              <input value={family?.code ?? ""} readOnly style={{ width: "100%", padding: "6px 8px", boxSizing: "border-box", background: "#f5f5f5" }} />
            </div>
          )}
          <div style={{ marginBottom: 12 }}>
            <label style={{ display: "block", fontSize: 12, marginBottom: 4 }}>Nome *</label>
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              style={{ width: "100%", padding: "6px 8px", boxSizing: "border-box" }}
            />
          </div>
          <div style={{ marginBottom: 16 }}>
            <label style={{ display: "block", fontSize: 12, marginBottom: 4 }}>Descricao</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={2}
              style={{ width: "100%", padding: "6px 8px", boxSizing: "border-box" }}
            />
          </div>
          {error && <p style={{ color: "#c00", fontSize: 12, marginBottom: 8 }}>{error}</p>}
          <div style={{ display: "flex", gap: 8, justifyContent: "flex-end" }}>
            <button type="button" onClick={onClose} style={{ padding: "6px 14px" }}>Cancelar</button>
            <button type="submit" disabled={saving} style={{ padding: "6px 14px" }}>
              {saving ? "Salvando..." : "Salvar"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
