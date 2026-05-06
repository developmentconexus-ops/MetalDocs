import { useState } from "react";
import type { Document } from "../../types";
import { PROFILES } from "../../data/mock";
import { StatusBadge, ProfileIcon } from "../shared";

interface DocRowProps { doc: Document }

const GRID = "18px 1fr 120px 105px 95px 90px 85px 30px";

export function DocTableHead() {
  const headers = ["", "Documento", "Tipo / family", "Status", "Owner", "Versão", "Próx. revisão", ""];
  return (
    <div style={{ display: "grid", gridTemplateColumns: GRID, padding: "0 14px", height: 34, alignItems: "center", background: "#FAF6F6", borderBottom: "1px solid #E8DEDE" }}>
      {headers.map((h, i) => (
        <div key={i} style={{ fontSize: 10.5, fontWeight: 500, color: "#A89090", letterSpacing: "0.4px", textTransform: "uppercase", cursor: h ? "pointer" : "default" }}>
          {h}
        </div>
      ))}
    </div>
  );
}

export function DocRow({ doc }: DocRowProps) {
  const [hov, setHov] = useState(false);
  const profileLabel  = PROFILES.find(p => p.code === doc.profile)?.label ?? "";

  return (
    <div
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      style={{ display: "grid", gridTemplateColumns: GRID, padding: "0 14px", height: 46, alignItems: "center", borderBottom: "1px solid #E8DEDE", cursor: "pointer", background: hov ? "#F9F3F3" : "#fff", transition: "background 0.1s" }}
    >
      <div style={{ width: 14, height: 14, border: "1.5px solid #D8CACA", borderRadius: 4 }} />

      <div style={{ display: "flex", alignItems: "center", gap: 10, minWidth: 0 }}>
        <ProfileIcon code={doc.profile} />
        <div style={{ minWidth: 0 }}>
          <div style={{ fontSize: 12.5, fontWeight: 500, color: "#1A0E0E", whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
            {doc.title}
          </div>
          <div style={{ fontSize: 10.5, color: "#A89090", fontFamily: "DM Mono, monospace", marginTop: 1 }}>
            {doc.code}
          </div>
        </div>
      </div>

      <div style={{ fontSize: 11.5, color: "#7A6060", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
        {profileLabel}
      </div>

      <div><StatusBadge status={doc.status} /></div>

      <div style={{ fontSize: 11.5, color: "#7A6060" }}>{doc.owner}</div>

      <div style={{ fontSize: 11, color: "#7A6060", fontFamily: "DM Mono, monospace" }}>{doc.version}</div>

      <div style={{ fontSize: 11.5, color: doc.urgent ? "#C8364A" : "#7A6060", fontWeight: doc.urgent ? 500 : 400 }}>
        {doc.urgent ? `⚠ ${doc.nextReview}` : doc.nextReview}
      </div>

      <div style={{ display: "flex", justifyContent: "flex-end", opacity: hov ? 1 : 0, transition: "opacity 0.1s" }}>
        <button style={{ width: 24, height: 24, borderRadius: 6, border: "none", background: "none", cursor: "pointer", display: "flex", alignItems: "center", justifyContent: "center", color: "#7A6060" }}>
          <svg width={13} height={13} viewBox="0 0 13 13" fill="#7A6060">
            <circle cx="6.5" cy="2.5" r="1.2" /><circle cx="6.5" cy="6.5" r="1.2" /><circle cx="6.5" cy="10.5" r="1.2" />
          </svg>
        </button>
      </div>
    </div>
  );
}
