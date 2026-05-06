import type { DocStatus, ProfileCode } from "../../types";
import { STATUS_CONFIG, PROFILES } from "../../data/mock";

// ─── StatusBadge ────────────────────────────────────────────────────────────

interface StatusBadgeProps { status: DocStatus }

export function StatusBadge({ status }: StatusBadgeProps) {
  const cfg = STATUS_CONFIG[status];
  return (
    <span style={{
      display: "inline-flex", alignItems: "center", gap: 4,
      padding: "2px 7px", borderRadius: 20,
      fontSize: 10.5, fontWeight: 500,
      background: cfg.bg, color: cfg.color,
    }}>
      <span style={{ width: 5, height: 5, borderRadius: "50%", background: cfg.color, flexShrink: 0 }} />
      {cfg.label}
    </span>
  );
}

// ─── ProfileIcon ────────────────────────────────────────────────────────────

interface ProfileIconProps { code: ProfileCode }

export function ProfileIcon({ code }: ProfileIconProps) {
  const p = PROFILES.find(x => x.code === code) ?? PROFILES[0];
  return (
    <div style={{
      width: 28, height: 28, borderRadius: 7, flexShrink: 0,
      display: "flex", alignItems: "center", justifyContent: "center",
      background: p.bg, color: p.color,
      fontFamily: "DM Mono, monospace",
      fontSize: 8.5, fontWeight: 500,
      letterSpacing: "0.2px", textTransform: "uppercase",
    }}>
      {p.code.toUpperCase()}
    </div>
  );
}

// ─── Chip ───────────────────────────────────────────────────────────────────

interface ChipProps {
  label:    string;
  count?:   number;
  dot?:     string;
  active?:  boolean;
  onClick?: () => void;
  icon?:    React.ReactNode;
}

export function Chip({ label, count, dot, active, onClick, icon }: ChipProps) {
  return (
    <button
      onClick={onClick}
      style={{
        display: "flex", alignItems: "center", gap: 5,
        height: 26, padding: "0 10px", borderRadius: 20,
        fontSize: 11.5, cursor: onClick ? "pointer" : "default",
        whiteSpace: "nowrap",
        border: `1px solid ${active ? "#DFC8C8" : "#D8CACA"}`,
        background: active ? "#F9F3F3" : "#fff",
        color: active ? "#6B1F2A" : "#483030",
        fontWeight: active ? 500 : 400,
        fontFamily: "DM Sans, sans-serif",
        transition: "all 0.12s",
      }}
    >
      {icon}
      {dot && <span style={{ width: 6, height: 6, borderRadius: "50%", background: dot }} />}
      {label}
      {count !== undefined && (
        <span style={{ fontFamily: "DM Mono, monospace", fontSize: 10 }}>{count}</span>
      )}
    </button>
  );
}

// ─── StatCard ───────────────────────────────────────────────────────────────

interface StatCardProps {
  label: string;
  value: number;
  sub:   string;
  color: string;
  pct:   number;
}

export function StatCard({ label, value, sub, color, pct }: StatCardProps) {
  return (
    <div style={{ background: "#fff", border: "1px solid #E8DEDE", borderRadius: 10, padding: "13px 15px" }}>
      <div style={{ fontSize: 11, color: "#7A6060", marginBottom: 3 }}>{label}</div>
      <div style={{ fontSize: 21, fontWeight: 500, lineHeight: 1, color }}>{value}</div>
      <div style={{ fontSize: 11, color: "#A89090", marginTop: 3 }}>{sub}</div>
      <div style={{ height: 3, background: "#E8DEDE", borderRadius: 2, marginTop: 8, overflow: "hidden" }}>
        <div style={{ height: "100%", borderRadius: 2, background: color, width: `${pct}%` }} />
      </div>
    </div>
  );
}

// ─── Divider ────────────────────────────────────────────────────────────────

export function SbDivider() {
  return <div style={{ height: 1, background: "#E8DEDE", margin: "6px 4px" }} />;
}

// ─── SectionLabel ───────────────────────────────────────────────────────────

interface SbLabelProps {
  children: React.ReactNode;
  onAdd?:   () => void;
}

export function SbLabel({ children, onAdd }: SbLabelProps) {
  return (
    <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "12px 8px 4px", fontSize: 10, fontWeight: 500, color: "#A89090", letterSpacing: "0.9px", textTransform: "uppercase" }}>
      {children}
      {onAdd && (
        <button onClick={onAdd} style={{ width: 18, height: 18, borderRadius: 4, border: "none", background: "none", cursor: "pointer", display: "flex", alignItems: "center", justifyContent: "center", color: "#A89090" }}>
          <svg width={10} height={10} viewBox="0 0 10 10" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round">
            <path d="M5 1v8M1 5h8" />
          </svg>
        </button>
      )}
    </div>
  );
}
