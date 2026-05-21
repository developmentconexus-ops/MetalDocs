import { useEffect, useState } from "react";
import { fetchActiveDocumentInstance, fetchControlledDocument, obsoleteControlledDocument, createRevision, type ActiveDocumentResponse } from './api/controlledDocuments';
import { PublishedDownloadCell } from './PublishedDownloadCell';
import type { ControlledDocument } from "./types";
import { ControlledDocumentDetailPanel } from '../approval/components/ControlledDocumentDetailPanel';

type Props = {
  id: string;
  onBack: () => void;
  onOpenDocumentEditor?: (docId: string) => void;
};

function StatusBadge({ status }: { status: ControlledDocument["status"] }) {
  const colors: Record<ControlledDocument["status"], string> = {
    active: "#2a7a2a",
    obsolete: "#888",
    superseded: "#b87a00",
  };
  return (
    <span style={{ color: colors[status], fontWeight: 600, fontSize: 12 }}>
      {status.toUpperCase()}
    </span>
  );
}

export function ControlledDocumentDetailPage({ id, onBack, onOpenDocumentEditor }: Props) {
  const [doc, setDoc] = useState<ControlledDocument | null>(null);
  const [instance, setInstance] = useState<ActiveDocumentResponse | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [actioning, setActioning] = useState(false);

  const [showNewRevisionForm, setShowNewRevisionForm] = useState(false);
  const [newRevisionName, setNewRevisionName] = useState("");
  const [creating, setCreating] = useState(false);
  const [createError, setCreateError] = useState("");

  async function load() {
    setLoading(true);
    try {
      const d = await fetchControlledDocument(id);
      setDoc(d);
      try {
        const inst = await fetchActiveDocumentInstance(id);
        setInstance(inst);
      } catch {
        setInstance(null);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load.");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, [id]);

  async function handleObsolete() {
    if (!doc || !window.confirm(`Tornar obsoleto o documento "${doc.code}"?`)) return;
    setActioning(true);
    try {
      await obsoleteControlledDocument(id);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to obsolete.");
    } finally {
      setActioning(false);
    }
  }

  async function handleCreateNewRevision() {
    if (!doc || !newRevisionName.trim()) return;
    setCreating(true);
    setCreateError("");
    try {
      const idempotencyKey = crypto.randomUUID();
      const res = await createRevision(
        doc.id,
        { name: newRevisionName.trim(), formData: {} },
        idempotencyKey,
      );
      if (onOpenDocumentEditor) onOpenDocumentEditor(res.document.id);
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : "Falha ao criar revisão.");
    } finally {
      setCreating(false);
    }
  }

  if (loading) return <div style={{ padding: 24 }}>Carregando...</div>;
  if (error) return <div style={{ padding: 24, color: "#c00" }}>{error}</div>;
  if (!doc) return null;

  return (
    <div style={{ padding: 24 }}>
      <button type="button" onClick={onBack} style={{ marginBottom: 16, padding: "4px 12px" }}>
        &larr; Voltar
      </button>
      <h2 style={{ margin: "0 0 16px", fontSize: 18 }}>{doc.title}</h2>
      <table style={{ borderCollapse: "collapse", fontSize: 13, width: "100%", maxWidth: 600 }}>
        <tbody>
          {[
            ["ID", doc.id],
            ["Codigo", doc.code],
            ["Perfil", doc.profileCode],
            ["Area", doc.processAreaCode],
            ["Departamento", doc.departmentCode ?? "-"],
            ["Numero de sequencia", doc.sequenceNum != null ? String(doc.sequenceNum) : "-"],
            ["Dono", doc.ownerUserId],
            ["Override template", doc.overrideTemplateVersionId ?? "-"],
            ["Criado em", doc.createdAt],
            ["Atualizado em", doc.updatedAt],
          ].map(([label, value]) => (
            <tr key={label} style={{ borderBottom: "1px solid #f0f0f0" }}>
              <td style={{ padding: "6px 8px", fontWeight: 600, width: 180, color: "#555" }}>{label}</td>
              <td style={{ padding: "6px 8px", fontFamily: "monospace", fontSize: 12 }}>{value}</td>
            </tr>
          ))}
          <tr style={{ borderBottom: "1px solid #f0f0f0" }}>
            <td style={{ padding: "6px 8px", fontWeight: 600, color: "#555" }}>Status</td>
            <td style={{ padding: "6px 8px" }}><StatusBadge status={doc.status} /></td>
          </tr>
        </tbody>
      </table>

      {doc.status === "active" && (
        <div style={{ marginTop: 24 }}>
          <button
            type="button"
            onClick={() => void handleObsolete()}
            disabled={actioning}
            style={{ padding: "6px 14px", color: "#c00", borderColor: "#c00" }}
          >
            {actioning ? "Processando..." : "Tornar obsoleto"}
          </button>
        </div>
      )}
      {instance?.documentId && (
        <div style={{ marginTop: 32 }}>
          <ControlledDocumentDetailPanel
            documentId={instance.documentId}
            approvalState={instance.approvalState ?? 'draft'}
            contentHash={instance.contentHash ?? ''}
            revisionVersion={instance.revisionVersion ?? 0}
            publishedDocumentId={instance.publishedDocumentId}
            lockedByInstanceId={instance.approvalInstanceId}
          />
        </div>
      )}

      {instance?.publishedDocumentId && (
        <div style={{ marginTop: 32, padding: 16, border: '1px solid #2a7a2a', borderRadius: 4, background: '#f4faf4' }}>
          <p style={{ margin: '0 0 8px', fontWeight: 600, color: '#2a7a2a' }}>Revisão publicada</p>
          <PublishedDownloadCell documentId={instance.publishedDocumentId} />
        </div>
      )}

      {!instance?.documentId && doc.status === 'active' && (
        <div style={{ marginTop: 32, padding: 16, border: '1px dashed #ccc', borderRadius: 4, color: '#888' }}>
          {!showNewRevisionForm ? (
            <>
              <p style={{ margin: "0 0 12px" }}>Nenhuma revisão ativa para este registro.</p>
              <button
                type="button"
                onClick={() => setShowNewRevisionForm(true)}
                style={{ padding: "6px 14px", fontWeight: 600 }}
              >
                Nova Revisão
              </button>
            </>
          ) : (
            <div>
              <p style={{ margin: "0 0 12px", fontWeight: 600, color: "#333" }}>Nova Revisão — {doc.code}</p>
              <p style={{ margin: "0 0 8px", fontSize: 12, color: "#666" }}>
                Template resolvido pelo servidor via perfil <strong>{doc.profileCode}</strong>
                {doc.overrideTemplateVersionId ? ` (override: ${doc.overrideTemplateVersionId})` : ""}
              </p>
              <label style={{ display: "block", marginBottom: 8, fontSize: 13, color: "#333" }}>
                Nome do documento
                <input
                  type="text"
                  value={newRevisionName}
                  onChange={(e) => setNewRevisionName(e.target.value)}
                  disabled={creating}
                  style={{ display: "block", marginTop: 4, padding: "6px 8px", width: 320, fontSize: 13 }}
                  placeholder={`${doc.code} — nova revisão`}
                />
              </label>
              {createError && (
                <p role="alert" style={{ color: "#c00", fontSize: 12, margin: "4px 0 8px" }}>{createError}</p>
              )}
              <div style={{ display: "flex", gap: 8, marginTop: 8 }}>
                <button
                  type="button"
                  onClick={() => void handleCreateNewRevision()}
                  disabled={creating || !newRevisionName.trim()}
                  style={{ padding: "6px 14px", fontWeight: 600 }}
                >
                  {creating ? "Criando..." : "Gerar Documento"}
                </button>
                <button
                  type="button"
                  onClick={() => { setShowNewRevisionForm(false); setNewRevisionName(""); setCreateError(""); }}
                  disabled={creating}
                  style={{ padding: "6px 14px" }}
                >
                  Cancelar
                </button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
