import { useState } from "react";

interface SubItemProps {
  label:   string;
  count:   number;
  active:  boolean;
  onClick: () => void;
  pill?:   { bg: string; color: string; text: string };
  dot?:    string;
}

export function SubItem({ label, count, active, onClick, pill, dot }: SubItemProps) {
  const [hov, setHov] = useState(false);

  return (
    <div
      onClick={onClick}
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      style={{
        display: "flex", alignItems: "center", gap: 8,
        padding: "5px 8px 5px 36px", borderRadius: 6,
        cursor: "pointer", fontSize: 12,
        color: active ? "#6B1F2A" : "#7A6060",
        fontWeight: active ? 500 : 400,
        background: active || hov ? "#F9F3F3" : "transparent",
        transition: "all 0.1s", position: "relative",
      }}
    >
      {active && (
        <div style={{ position: "absolute", left: 0, top: "50%", transform: "translateY(-50%)", width: 3, height: 14, background: "#6B1F2A", borderRadius: "0 3px 3px 0" }} />
      )}
      {pill && (
        <span style={{ display: "inline-flex", alignItems: "center", justifyContent: "center", width: 22, height: 14, borderRadius: 4, background: pill.bg, color: pill.color, fontFamily: "DM Mono, monospace", fontSize: 8.5, fontWeight: 500, flexShrink: 0 }}>
          {pill.text}
        </span>
      )}
      {dot && (
        <span style={{ width: 7, height: 7, borderRadius: "50%", background: dot, flexShrink: 0 }} />
      )}
      {label}
      <span style={{ marginLeft: "auto", fontFamily: "DM Mono, monospace", fontSize: 10, color: "#A89090" }}>
        {count}
      </span>
    </div>
  );
}
