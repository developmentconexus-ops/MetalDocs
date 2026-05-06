import { useState } from "react";
import type { Area, Document } from "../../types";
import { DocRow, DocTableHead } from "./DocRow";

interface GroupSectionProps {
  area: Area;
  docs: Document[];
}

export function GroupSection({ area, docs }: GroupSectionProps) {
  const [open, setOpen] = useState(area.code === "ven");

  if (docs.length === 0) return null;

  return (
    <div style={{ marginBottom: 4 }}>
      {/* Header */}
      <div
        onClick={() => setOpen(o => !o)}
        style={{ display: "flex", alignItems: "center", gap: 8, padding: "6px 2px", cursor: "pointer", marginBottom: open ? 6 : 2 }}
      >
        <svg
          width={13} height={13} viewBox="0 0 13 13"
          fill="none" stroke="#A89090" strokeWidth={1.5} strokeLinecap="round"
          style={{ flexShrink: 0, transition: "transform 0.15s", transform: open ? "rotate(0deg)" : "rotate(-90deg)" }}
        >
          <path d="M2.5 4.5l4 4 4-4" />
        </svg>
        <span style={{ width: 8, height: 8, borderRadius: "50%", background: area.color, flexShrink: 0 }} />
        <span style={{ fontSize: 11.5, fontWeight: 500, color: "#7A6060", letterSpacing: "0.3px" }}>{area.label}</span>
        <span style={{ fontFamily: "DM Mono, monospace", fontSize: 10, color: "#A89090" }}>{docs.length} documentos</span>
        <div style={{ flex: 1, height: 1, background: "#E8DEDE", marginLeft: 4 }} />
      </div>

      {open && (
        <div style={{ background: "#fff", border: "1px solid #E8DEDE", borderRadius: 10, overflow: "hidden", marginBottom: 6 }}>
          <DocTableHead />
          {docs.map(doc => <DocRow key={doc.id} doc={doc} />)}
        </div>
      )}
    </div>
  );
}
