interface TopbarProps {
  search:    string;
  setSearch: (v: string) => void;
}

export function Topbar({ search, setSearch }: TopbarProps) {
  return (
    <div style={{ height: 50, background: "#3E1018", display: "flex", alignItems: "center", padding: "0 16px 0 0", gap: 12, flexShrink: 0 }}>

      {/* Brand */}
      <div style={{ width: 240, flexShrink: 0, display: "flex", alignItems: "center", gap: 10, padding: "0 16px", borderRight: "1px solid rgba(255,255,255,0.06)", height: "100%" }}>
        <div style={{ width: 30, height: 30, borderRadius: 8, background: "linear-gradient(135deg, #8B2E3A, #6B1F2A)", display: "flex", alignItems: "center", justifyContent: "center", flexShrink: 0 }}>
          <svg width={15} height={15} viewBox="0 0 15 15" fill="none" stroke="rgba(255,255,255,0.9)" strokeWidth={1.3} strokeLinecap="round" strokeLinejoin="round">
            <path d="M3 2h6.5L12 4.5V13H3V2z" /><path d="M9.5 2v2.5H12" /><path d="M5 7h5M5 9.5h5M5 12h3" />
          </svg>
        </div>
        <div>
          <div style={{ fontSize: 13, fontWeight: 500, color: "#fff", letterSpacing: "0.2px" }}>MetalDocs</div>
          <div style={{ fontSize: 10, color: "rgba(255,255,255,0.38)", fontFamily: "DM Mono, monospace", letterSpacing: "0.4px" }}>SGQ · Metal Nobre</div>
        </div>
      </div>

      {/* Search */}
      <div style={{ flex: 1, maxWidth: 440, display: "flex", alignItems: "center", gap: 8, background: "rgba(255,255,255,0.07)", border: "1px solid rgba(255,255,255,0.09)", borderRadius: 8, height: 32, padding: "0 11px" }}>
        <svg width={13} height={13} viewBox="0 0 13 13" fill="none" stroke="rgba(255,255,255,0.35)" strokeWidth={1.5} strokeLinecap="round">
          <circle cx="5.5" cy="5.5" r="4" /><path d="M9 9l2.5 2.5" />
        </svg>
        <input
          value={search}
          onChange={e => setSearch(e.target.value)}
          placeholder="Buscar por título, código, área…"
          style={{ background: "none", border: "none", outline: "none", fontFamily: "DM Sans, sans-serif", fontSize: 12.5, color: "#fff", width: "100%" }}
        />
        <kbd style={{ fontFamily: "DM Mono, monospace", fontSize: 10, color: "rgba(255,255,255,0.22)", whiteSpace: "nowrap", marginLeft: "auto" }}>⌘K</kbd>
      </div>

      <div style={{ display: "flex", alignItems: "center", gap: 8, marginLeft: "auto" }}>

        {/* Bell */}
        <button style={{ width: 32, height: 32, borderRadius: 7, border: "1px solid rgba(255,255,255,0.1)", background: "rgba(255,255,255,0.06)", display: "flex", alignItems: "center", justifyContent: "center", color: "rgba(255,255,255,0.65)", cursor: "pointer", position: "relative" }}>
          <svg width={14} height={14} viewBox="0 0 14 14" fill="none" stroke="currentColor" strokeWidth={1.4} strokeLinecap="round">
            <path d="M7 1.5a4 4 0 0 1 4 4v2.5l1 1.5H2L3 8V5.5a4 4 0 0 1 4-4z" />
            <path d="M5.5 11.5a1.5 1.5 0 0 0 3 0" />
          </svg>
          <div style={{ position: "absolute", top: 5, right: 5, width: 6, height: 6, borderRadius: "50%", background: "#C8364A", border: "1.5px solid #3E1018" }} />
        </button>

        {/* New document */}
        <button style={{ height: 32, padding: "0 14px", borderRadius: 8, background: "#C8364A", border: "none", color: "#fff", fontFamily: "DM Sans, sans-serif", fontSize: 12.5, fontWeight: 500, cursor: "pointer", display: "flex", alignItems: "center", gap: 6 }}>
          <svg width={12} height={12} viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round">
            <path d="M6 1v10M1 6h10" />
          </svg>
          Novo documento
        </button>

        {/* Avatar */}
        <div style={{ width: 30, height: 30, borderRadius: "50%", background: "#8B2E3A", border: "2px solid rgba(255,255,255,0.18)", display: "flex", alignItems: "center", justifyContent: "center", fontSize: 11, fontWeight: 500, color: "#fff", cursor: "pointer", letterSpacing: "0.3px" }}>
          LE
        </div>
      </div>
    </div>
  );
}
