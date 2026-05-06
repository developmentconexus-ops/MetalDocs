import { useState } from "react";

interface NavItemProps {
  icon:         React.ReactNode;
  label:        string;
  badge?:       string | number;
  badgeVariant?: "red" | "amber" | "default";
  active?:      boolean;
  dot?:         boolean;
  onClick?:     () => void;
}

const BADGE_STYLES: Record<string, { bg: string; color: string }> = {
  red:     { bg: "#FDECEC", color: "#A32020" },
  amber:   { bg: "#FDF5E0", color: "#8A5800" },
  default: { bg: "#F2EAEA", color: "#6B1F2A" },
};

export function NavItem({ icon, label, badge, badgeVariant = "default", active, dot, onClick }: NavItemProps) {
  const [hov, setHov] = useState(false);
  const bs = BADGE_STYLES[badgeVariant];

  return (
    <div
      onClick={onClick}
      onMouseEnter={() => setHov(true)}
      onMouseLeave={() => setHov(false)}
      style={{
        display: "flex", alignItems: "center", gap: 8,
        padding: "6px 8px", borderRadius: 7, cursor: "pointer",
        fontSize: 12.5, fontWeight: active ? 500 : 400,
        color: active ? "#6B1F2A" : "#483030",
        background: active || hov ? "#F9F3F3" : "transparent",
        transition: "all 0.1s", position: "relative",
      }}
    >
      {active && (
        <div style={{ position: "absolute", left: 0, top: "50%", transform: "translateY(-50%)", width: 3, height: 16, background: "#6B1F2A", borderRadius: "0 3px 3px 0" }} />
      )}
      <span style={{ width: 15, height: 15, flexShrink: 0, display: "flex", alignItems: "center", justifyContent: "center", color: active ? "#6B1F2A" : "#7A6060" }}>
        {icon}
      </span>
      {label}
      {badge && (
        <span style={{ marginLeft: "auto", fontFamily: "DM Mono, monospace", fontSize: 10, padding: "1px 6px", borderRadius: 10, background: bs.bg, color: bs.color }}>
          {badge}
        </span>
      )}
      {dot && (
        <span style={{ width: 6, height: 6, borderRadius: "50%", background: "#1A6B35", marginLeft: badge ? 0 : "auto" }} />
      )}
    </div>
  );
}
