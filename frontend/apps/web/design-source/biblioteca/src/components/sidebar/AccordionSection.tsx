import { useState } from "react";
import type { Profile, ActiveFilter, AreaCode, ProfileCode } from "../../types";
import { AREAS, DOCUMENTS } from "../../data/mock";
import { SubItem } from "./SubItem";

interface AccordionSectionProps {
  profile:      Profile;
  activeFilter: ActiveFilter | null;
  onFilter:     (f: ActiveFilter) => void;
}

export function AccordionSection({ profile, activeFilter, onFilter }: AccordionSectionProps) {
  const [open, setOpen] = useState(profile.code === "po");
  const [hov, setHov]   = useState(false);

  return (
    <div style={{ marginBottom: 2 }}>
      <div
        onClick={() => setOpen(o => !o)}
        onMouseEnter={() => setHov(true)}
        onMouseLeave={() => setHov(false)}
        style={{
          display: "flex", alignItems: "center", gap: 7,
          padding: "6px 8px", borderRadius: 7, cursor: "pointer",
          color: open ? "#6B1F2A" : "#483030",
          fontSize: 12.5, fontWeight: 500,
          background: open || hov ? "#F9F3F3" : "transparent",
          transition: "background 0.1s",
        }}
      >
        <svg
          width={14} height={14} viewBox="0 0 14 14"
          fill="none" stroke="#A89090" strokeWidth={1.5} strokeLinecap="round"
          style={{ flexShrink: 0, transition: "transform 0.18s", transform: open ? "rotate(90deg)" : "rotate(0deg)" }}
        >
          <path d="M4 5.5l3 3 3-3" />
        </svg>

        <div style={{ width: 22, height: 22, borderRadius: 6, flexShrink: 0, display: "flex", alignItems: "center", justifyContent: "center", background: profile.bg }}>
          <svg width={12} height={12} viewBox="0 0 12 12" fill="none" stroke={profile.color} strokeWidth={1.3} strokeLinecap="round" strokeLinejoin="round">
            <path d="M2 1h6l3 3v7H2V1z" />
            <path d="M8 1v3h3" />
            <path d="M3.5 6h5M3.5 8h3" />
          </svg>
        </div>

        <span style={{ fontSize: 12.5, color: "#1A0E0E" }}>{profile.label}</span>
        <span style={{ marginLeft: "auto", fontFamily: "DM Mono, monospace", fontSize: 10, color: "#A89090" }}>
          {profile.count}
        </span>
      </div>

      <div style={{ overflow: "hidden", maxHeight: open ? 600 : 0, transition: "max-height 0.22s ease" }}>
        {/* All */}
        <SubItem
          label="Todos"
          count={profile.count}
          active={activeFilter?.type === profile.code && !activeFilter?.area}
          onClick={() => onFilter({ type: profile.code as ProfileCode })}
          pill={{ bg: profile.bg, color: profile.color, text: profile.code.toUpperCase() }}
        />
        {/* Per area */}
        {AREAS.map(a => {
          const count = DOCUMENTS.filter(d => d.profile === profile.code && d.area === a.code).length;
          if (count === 0) return null;
          return (
            <SubItem
              key={a.code}
              label={a.label}
              count={count}
              active={activeFilter?.type === profile.code && activeFilter?.area === a.code}
              onClick={() => onFilter({ type: profile.code as ProfileCode, area: a.code as AreaCode })}
              dot={a.color}
            />
          );
        })}
      </div>
    </div>
  );
}
