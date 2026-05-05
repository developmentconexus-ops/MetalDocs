import { useState } from "react";
import type { DocumentFamily } from "./types";
import { deactivateFamily } from './api/taxonomy';
import { FamilyEditDialog } from "./FamilyEditDialog";

type Props = {
  families: DocumentFamily[];
  includeInactive: boolean;
  onToggleInactive: (value: boolean) => void;
  onRefresh: () => void;
};

export function FamilyList({ families, includeInactive, onToggleInactive, onRefresh }: Props) {
  const [dialogMode, setDialogMode] = useState<"create" | "edit" | null>(null);
  const [selectedFamily, setSelectedFamily] = useState<DocumentFamily | undefined>(undefined);

  function openCreate() {
    setSelectedFamily(undefined);
    setDialogMode("create");
  }

  function openEdit(family: DocumentFamily) {
    setSelectedFamily(family);
    setDialogMode("edit");
  }

  function closeDialog() {
    setDialogMode(null);
    setSelectedFamily(undefined);
  }

  async function handleDeactivate(family: DocumentFamily) {
    if (!window.confirm(`Desativar família "${family.name}" (${family.code})?`)) return;
    try {
      await deactivateFamily(family.code);
      onRefresh();
    } catch (err) {
      window.alert(err instanceof Error ? err.message : "Falha ao desativar.");
    }
  }

  return (
    <div>
      <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 12 }}>
        <button type="button" onClick={openCreate} style={{ padding: "6px 14px" }}>
          + Nova Família
        </button>
        <label style={{ fontSize: 13, display: "flex", alignItems: "center", gap: 4 }}>
          <input
            type="checkbox"
            checked={includeInactive}
            onChange={(e) => onToggleInactive(e.target.checked)}
          />
          Mostrar inativas
        </label>
      </div>

      <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13 }}>
        <thead>
          <tr style={{ borderBottom: "2px solid #e0e0e0", textAlign: "left" }}>
            <th style={{ padding: "6px 8px" }}>Codigo</th>
            <th style={{ padding: "6px 8px" }}>Nome</th>
            <th style={{ padding: "6px 8px" }}>Descricao</th>
            <th style={{ padding: "6px 8px" }}>Status</th>
            <th style={{ padding: "6px 8px" }}>Acoes</th>
          </tr>
        </thead>
        <tbody>
          {families.length === 0 && (
            <tr>
              <td colSpan={5} style={{ padding: "12px 8px", color: "#888" }}>Nenhuma família encontrada.</td>
            </tr>
          )}
          {families.map((f) => (
            <tr key={f.code} style={{ borderBottom: "1px solid #f0f0f0" }}>
              <td style={{ padding: "6px 8px", fontFamily: "monospace" }}>{f.code}</td>
              <td style={{ padding: "6px 8px" }}>{f.name}</td>
              <td style={{ padding: "6px 8px", color: "#666" }}>{f.description || "—"}</td>
              <td style={{ padding: "6px 8px" }}>
                <span style={{ color: f.isActive ? "#080" : "#888" }}>
                  {f.isActive ? "Ativa" : "Inativa"}
                </span>
              </td>
              <td style={{ padding: "6px 8px" }}>
                <button type="button" onClick={() => openEdit(f)} style={{ marginRight: 8, padding: "3px 10px", fontSize: 12 }}>
                  Editar
                </button>
                {f.isActive && (
                  <button type="button" onClick={() => void handleDeactivate(f)} style={{ padding: "3px 10px", fontSize: 12, color: "#c00" }}>
                    Desativar
                  </button>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {dialogMode && (
        <FamilyEditDialog
          mode={dialogMode}
          family={selectedFamily}
          onClose={closeDialog}
          onSaved={onRefresh}
        />
      )}
    </div>
  );
}
